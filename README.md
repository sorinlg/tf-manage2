# tf-manage2

A complete rewrite of the [original tf-manage](https://github.com/sorinlg/tf-manage) bash tool in Go, providing better performance, cross-platform compatibility, and enhanced developer experience while maintaining 100% backward compatibility.

## 🎯 Overview

tf-manage2 is a modern Terraform wrapper that organizes your infrastructure code using a standardized folder structure and workspace naming conventions. It simplifies Terraform workflows by providing a consistent interface for managing multiple products, environments, and module instances.

This is tf-manage2 v1.0.0 - a new implementation of the proven tf-manage workflow (originally at v6.4.0) rebuilt in Go for superior performance and reliability.

**Key Benefits:**
- 🚀 **60-75% faster** than the original bash version
- 🌍 **Cross-platform** support (Windows, macOS, Linux)
- 🔧 **Enhanced interactive support** for terraform prompts
- 📦 **Single binary** with no external dependencies
- 🎨 **Beautiful colored output** with clear status indicators
- 🔄 **Drop-in replacement** for existing tf-manage projects

For a detailed breakdown of new features, performance improvements, and migration information, see [WHATS_NEW.md](WHATS_NEW.md).

## 🔄 Migration from Original tf-manage

tf-manage2 is a **drop-in replacement** for the original tf-manage (v6.4.0):

- ✅ **No configuration changes required**
- ✅ **Same command interface**
- ✅ **Works with existing projects**
- ✅ **Compatible with existing workspaces**
- ✅ **Reads existing .tfm.conf files**

Simply replace the `tf` binary and continue using your existing workflows.

## ✨ Features

### Core Functionality
- **Standardized Project Structure**: Enforces consistent terraform project layout
- **Workspace Management**: Automatic workspace creation and selection with naming convention `{product}.{repo}.{module}.{env}.{module_instance}`
- **Configuration Management**: Reads `.tfm.conf` files for project-specific settings
- **Path Resolution**: Intelligent path handling for modules and environments
- **Variable Integration**: Automatic injection of tf-manage context into terraform variables

### Terraform Operations
- **Complete Action Support**: All terraform commands including `init`, `plan`, `apply`, `destroy`, `output`, `import`, `workspace`, and more
- **Interactive Mode**: Proper handling of terraform prompts for confirmations
- **Unattended Mode**: Automatic approval for CI/CD environments
- **Plan File Management**: Support for plan generation and application workflows
- **State Management**: Advanced state operations and workspace handling

### Developer Experience
- **Fast Execution**: Native Go performance with optimized validation
- **Clear Feedback**: ANSI colored output with ✓/✗ status indicators
- **Error Handling**: Detailed error messages with helpful context
- **Debug Support**: Comprehensive logging for troubleshooting

## 🚀 Usage

### Basic Command Structure
```bash
tf <product> <module> <env> <module_instance> <action> [workspace]
```

### Common Workflows

#### Initialize and Plan
```bash
# Initialize terraform for a module instance
tf project1 sample_module dev instance_x init

# Generate an execution plan
tf project1 sample_module dev instance_x plan
```

#### Apply Changes
```bash
# Apply changes interactively (prompts for confirmation)
tf project1 sample_module dev instance_x apply

# Apply a pre-generated plan file
tf project1 sample_module dev instance_x apply_plan
```

#### Destroy Resources
```bash
# Destroy infrastructure (prompts for confirmation)
tf project1 sample_module dev instance_x destroy
```

#### Advanced Operations
```bash
# Import existing resources
tf project1 sample_module dev instance_x 'import aws_instance.example i-1234567890abcdef0'

# Manage terraform state
tf project1 sample_module dev instance_x 'state list'

# Format terraform files
tf project1 sample_module dev instance_x fmt

# Validate configuration
tf project1 sample_module dev instance_x validate
```

### Environment Variables

#### Execution Mode Control
```bash
# Force unattended mode (auto-approve)
export TF_EXEC_MODE_OVERRIDE=1
tf project1 sample_module dev instance_x apply
```

#### CI/CD Detection
tf-manage2 automatically detects CI/CD environments and switches to unattended mode for:

GitHub Actions, GitLab CI, CircleCI, Travis CI, Azure DevOps, Jenkins, Bamboo, TeamCity, Buildkite, Drone CI, AWS CodeBuild, and any system setting `CI=true`.

```bash
# Examples of automatic detection:
GITHUB_ACTIONS=true tf project1 sample_module dev instance_x apply  # Unattended
GITLAB_CI=true tf project1 sample_module dev instance_x apply       # Unattended
CI=true tf project1 sample_module dev instance_x apply              # Unattended
tf project1 sample_module dev instance_x apply                      # Interactive
```

For complete CI/CD environment variable reference, see [REFERENCE.md](REFERENCE.md#cicd-platform-detection).

#### Terraform Variables
tf-manage automatically injects context variables into terraform:
```bash
# These variables are automatically available in your terraform code:
# - tfm_product     = "project1"
# - tfm_repo        = "tfm-project"
# - tfm_module      = "sample_module"
# - tfm_env         = "dev"
# - tfm_module_instance = "instance_x"
```

For complete variable reference and usage examples, see [REFERENCE.md](REFERENCE.md#terraform-variables).

### Project Structure

tf-manage expects this standardized project structure:

```
your-project/
├── .tfm.conf                          # Project configuration
└── terraform/
    ├── environments/                   # Environment-specific configurations
    │   ├── project1/
    │   │   ├── dev/
    │   │   │   └── sample_module/
    │   │   │       ├── instance_x.tfvars
    │   │   │       └── instance_y.tfvars
    │   │   ├── staging/
    │   │   └── prod/
    │   └── project2/
    └── modules/                        # Reusable terraform modules
        ├── sample_module/
        │   ├── main.tf
        │   ├── variables.tf
        │   ├── outputs.tf
        │   └── tfm.tf                  # tf-manage integration variables
        └── another_module/
```

#### Configuration File (.tfm.conf)
```bash
#!/bin/bash
export __tfm_repo_name='your-repo-name'
export __tfm_env_rel_path='terraform/environments'
export __tfm_module_rel_path='terraform/modules'
```

## 📥 Installation

### Homebrew (macOS/Linux)
```bash
# Add the tap
brew tap sorinlg/tap

# Install tf-manage2
brew install tf-manage2

# Verify installation
tf --version
```

### Download Binary
```bash
# Download latest release for your platform
curl -L https://github.com/sorinlg/tf-manage2/releases/latest/download/tf-manage2_Linux_x86_64.tar.gz -o tf-manage2.tar.gz
tar xzf tf-manage2.tar.gz
chmod +x tf
sudo mv tf /usr/local/bin/

# Or for macOS ARM64
curl -L https://github.com/sorinlg/tf-manage2/releases/latest/download/tf-manage2_Darwin_arm64.tar.gz -o tf-manage2.tar.gz
tar xzf tf-manage2.tar.gz
chmod +x tf
sudo mv tf /usr/local/bin/
```

### Build from Source
```bash
git clone https://github.com/sorinlg/tf-manage2.git
cd tf-manage2
go build -o tf .
sudo mv tf /usr/local/bin/
```

### Verify Installation
```bash
tf --help
tf --version
```

## 🔧 Development

For local development and contribution guidelines, see [DEVELOPMENT.md](DEVELOPMENT.md).

Quick start for development:
```bash
git clone https://github.com/sorinlg/tf-manage2.git
cd tf-manage2/examples/tfm-project
go run ../../ project1 sample_module dev instance_x init
go run ../../ project1 sample_module dev instance_x plan
```

## 📚 Documentation

This documentation follows the [Diátaxis framework](https://diataxis.fr/) for optimal user experience:

- **[README.md](README.md)** - Tutorial & How-to Guide: Overview, installation, and practical usage examples
- **[WHATS_NEW.md](WHATS_NEW.md)** - Explanation: Feature overview, benefits, and migration guidance
- **[REFERENCE.md](REFERENCE.md)** - Technical Reference: Complete specification of environment variables, configuration, commands, and file structures
- **[DEVELOPMENT.md](DEVELOPMENT.md)** - How-to Guide: Development setup, testing procedures, and contribution guidelines

## 🤝 Contributing

Contributions are welcome! Please read [DEVELOPMENT.md](DEVELOPMENT.md) for guidelines on:
- Setting up your development environment
- Running tests and examples
- Code style and best practices
- Submitting pull requests

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [Original tf-manage (bash)](https://github.com/adobe/tf-manage) - The original bash implementation (v6.4.0)
- [Terraform](https://www.terraform.io/) - Infrastructure as Code tool
