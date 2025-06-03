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
ARG TERRAFORM_VERSION='1.9.8'

# Install curl and unzip for downloading Terraform
RUN apk add --no-cache curl unzip

# Download and extract Terraform
RUN curl -LO "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    unzip "terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    chmod +x terraform

# always build for linux/amd64
FROM --platform=linux/amd64 oraclelinux:9-slim

# Image configuration
ARG AWS_CLI_VERSION='2.15.38'
ARG USERNAME='tf'
ARG USER_UID='1001'
ARG USER_GID="${USER_UID}"
ARG TFM_INSTALLER_DIR='/opt/tf-manage-installer'
ARG TFM_INSTALL_PATH='/opt/terraform/tf-manage'

# add kubectl yum repo
COPY <<EOF  /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://pkgs.k8s.io/core:/stable:/v1.30/rpm/
enabled=1
gpgcheck=1
gpgkey=https://pkgs.k8s.io/core:/stable:/v1.30/rpm/repodata/repomd.xml.key
EOF

RUN \
  # install the required packages
  microdnf -y update \
  && microdnf -y install wget unzip git bash-completion which curl vim procps jq kubectl findutils \
  #
  # create non-root user
  && groupadd --gid $USER_GID $USERNAME \
  && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
  #
  # install the aws cli to /usr/local (no sudo needed in container)
  && curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64-${AWS_CLI_VERSION}.zip" -o "awscliv2.zip" \
  && unzip awscliv2.zip \
  && ./aws/install --install-dir /usr/local/aws-cli --bin-dir /usr/local/bin \
  && rm -rf awscliv2.zip aws/ \
  #
  # clean cache
  && microdnf clean all \
  #
  # git config
  && git config --global --add safe.directory /app \
  && git config --global --add safe.directory ${TFM_INSTALLER_DIR}

# switch to non-root user
USER $USERNAME

# install tf-manage
RUN mkdir -p /home/$USERNAME/bin
COPY --from=terraform-downloader /terraform /usr/local/bin/terraform
COPY --from=builder /app/tf /home/$USERNAME/bin
ENV PATH="${PATH}:/home/$USERNAME/bin"

# set the working directory
WORKDIR /app
