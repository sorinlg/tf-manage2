# Multi-stage build for minimal tf + Terraform image

# Global build arguments available to all stages
ARG USERNAME='tf'
ARG USER_UID='1001'
ARG USER_GID="${USER_UID}"

# Build stage
FROM golang:1.24.3 AS builder

# # Install git (needed for Go modules)
# RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -extldflags=-static" -a -installsuffix cgo -o tf .

# Terraform download stage
FROM alpine:3.22 AS terraform-downloader

# Image configuration
ARG AWS_CLI_VERSION='2.15.38'
ARG TERRAFORM_VERSION='1.12.1'


# Install curl and unzip for downloading Terraform
RUN apk add --no-cache curl unzip

# Download and extract Terraform
RUN curl -LO "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    unzip "terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    chmod +x terraform

# Utility stage - prepare Jenkins utilities
FROM debian:12-slim AS utility-builder

# Re-declare ARGs for this stage
ARG USERNAME
ARG USER_UID
ARG USER_GID

# Install utilities needed by Jenkins
RUN apt-get update && \
    apt-get install -y --no-install-recommends procps git findutils && \
    rm -rf /var/lib/apt/lists/*

# Create a non-root user for Jenkins in this stage
RUN groupadd --gid ${USER_GID} ${USERNAME} && \
    useradd --uid ${USER_UID} --gid ${USER_GID} --create-home --shell /bin/sh ${USERNAME}

# Runtime stage - using distroless for security
FROM gcr.io/distroless/static-debian12

# Re-declare ARGs for this stage
ARG USERNAME

# Copy the binaries from previous stages
COPY --from=builder /app/tf /usr/local/bin/tf
COPY --from=terraform-downloader /terraform /usr/local/bin/terraform
# Copy process utilities from debian stage for Jenkins
COPY --from=utility-builder /usr/bin/top /usr/bin/top
COPY --from=utility-builder /usr/bin/ps /usr/bin/ps
COPY --from=utility-builder /usr/bin/cat /usr/bin/cat
COPY --from=utility-builder /usr/bin/find /usr/bin/find
COPY --from=utility-builder /usr/bin/git /usr/bin/git
COPY --from=utility-builder /bin/ls /bin/ls
COPY --from=utility-builder /bin/sh /bin/sh
COPY --from=utility-builder /lib/ /lib/
COPY --from=utility-builder /usr/lib/ /usr/lib/

# Copy user and group information
COPY --from=utility-builder /etc/passwd /etc/passwd
COPY --from=utility-builder /etc/group /etc/group
COPY --from=utility-builder /etc/shadow /etc/shadow
COPY --from=utility-builder /home/${USERNAME} /home/${USERNAME}

# Set PATH to include /usr/local/bin
ENV PATH="/usr/local/bin:${PATH}"
ENV HOME="/home/${USERNAME}"

# Set working directory
WORKDIR /workspace

# Run as non-root user
USER ${USERNAME}

# Keep container running for Jenkins
ENTRYPOINT [ "/usr/local/bin/tf" ]
