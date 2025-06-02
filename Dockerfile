# Multi-stage build for minimal tf + Terraform image

# Build stage
FROM golang:1.24.3-alpine AS builder

# Install git (needed for Go modules)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tf .

# Terraform download stage
FROM alpine:3.22 AS terraform-downloader

# Install curl and unzip for downloading Terraform
RUN apk add --no-cache curl unzip

# Download and extract Terraform
ARG TERRAFORM_VERSION=1.12.1
RUN curl -LO "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    unzip "terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    chmod +x terraform

# Runtime stage - using distroless for minimal attack surface
FROM gcr.io/distroless/static-debian11

# Copy the binaries from previous stages
COPY --from=builder /app/tf /usr/local/bin/tf
COPY --from=terraform-downloader /terraform /usr/local/bin/terraform

# Set PATH to include /usr/local/bin
ENV PATH="/usr/local/bin:${PATH}"

# Set working directory
WORKDIR /workspace

# Run as non-root user (distroless uses user ID 65532 by default)
USER 65532:65532

# Set entrypoint
ENTRYPOINT ["tf"]
