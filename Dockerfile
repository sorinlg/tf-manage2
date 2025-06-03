# Multi-stage build for minimal tf + Terraform image

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

# Install curl and unzip for downloading Terraform
RUN apk add --no-cache curl unzip

# Download and extract Terraform
ARG TERRAFORM_VERSION=1.12.1
RUN curl -LO "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    unzip "terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    chmod +x terraform

# Utility stage - prepare Jenkins utilities
FROM debian:12-slim AS utility-builder

# Install utilities needed by Jenkins
RUN apt-get update && \
    apt-get install -y --no-install-recommends procps && \
    rm -rf /var/lib/apt/lists/*

# Runtime stage - using distroless for security
FROM gcr.io/distroless/static-debian12

# Copy the binaries from previous stages
COPY --from=builder /app/tf /usr/local/bin/tf
COPY --from=terraform-downloader /terraform /usr/local/bin/terraform
# Copy process utilities from debian stage for Jenkins
COPY --from=utility-builder /usr/bin/top /usr/bin/top
COPY --from=utility-builder /usr/bin/ps /usr/bin/ps
COPY --from=utility-builder /usr/bin/cat /usr/bin/cat
COPY --from=utility-builder /bin/ls /bin/ls
COPY --from=utility-builder /bin/sh /bin/sh
COPY --from=utility-builder /lib/ /lib/
COPY --from=utility-builder /usr/lib/ /usr/lib/

# Set PATH to include /usr/local/bin
ENV PATH="/usr/local/bin:${PATH}"

# Set working directory
WORKDIR /workspace

# Run as non-root user
USER 1001:1001

# Set entrypoint
ENTRYPOINT ["/bin/sh", "-c"]
