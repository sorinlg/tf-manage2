# tf-manage2

A lean Terraform workspace manager with standardized folder structure and simplified CLI interface.

## Introduction

tf-manage2 provides:
- **Standard folder structure**: Organized `terraform/environments/{project}/{env}/` and `terraform/modules/` layout
- **Workspace naming**: Automatic `{project}.{repo}.{module}.{env}.{module_instance}` convention
- **Simplified CLI**: Single command interface to consume folder structure with built-in validation
- **Local developer experience**: Interactive mode with prompts and colored output
- **CI runtime support**: Auto-detects CI environments and enables unattended mode

## Installation

<details><summary>Homebrew (Stable)</summary>

```bash
brew install sorinlg/tap/tf-manage2
```

</details>

<details><summary>Homebrew (Development/Prerelease)</summary>

For testing prerelease versions with new features:

```bash
brew install sorinlg/dev-tap/tf-manage2-dev
```

This tap contains alpha, beta, and rc versions for early testing.

</details>

<details><summary>Go install</summary>

```bash
go install github.com/sorinlg/tf-manage2@latest
```
</details>
<details><summary>Download binary</summary>

```bash
curl -L https://github.com/sorinlg/tf-manage2/releases/latest/download/tf-manage2-linux-amd64 -o tf-manage2
chmod +x tf-manage2
sudo mv tf-manage2 /usr/local/bin/
```
</details>

<details><summary>Shell completion</summary>

After installing tf-manage2, enable shell completion:

**Bash:**
```bash
# Add to ~/.bashrc
if command -v tf &> /dev/null; then
  . $(brew --prefix)/etc/bash_completion.d/tf
fi
```

**Zsh:**
```bash
# Add to ~/.zshrc
if command -v tf &> /dev/null; then
  fpath=( $(brew --prefix)/share/tf-completion $fpath )
  autoload -U compinit && compinit
fi
```

Then reload your shell:
```bash
source ~/.bashrc  # or ~/.zshrc for zsh users
```

</details>

## Usage

```bash
tf <project> <module> <env> <module_instance> <action> [workspace]
```

**Examples:**
```bash
tf project1 sample_module dev instance_x init
tf project1 sample_module dev instance_x plan
tf project1 sample_module dev instance_x apply
tf project1 sample_module staging instance_y destroy workspace=custom
```

**Supported actions:** `init`, `plan`, `apply`, `destroy`, `output`, `workspace`, `validate`, and more.

## Configuration

tf-manage2 supports both modern YAML and legacy bash configuration formats:

### Modern YAML Format (Recommended)

Create a `.tfm.yaml` file in your project root:

```yaml
# tf-manage2 configuration file
config_version: "2.0"
repo_name: "your-project-name"
env_rel_path: "terraform/environments"
module_rel_path: "terraform/modules"
```

### Legacy Bash Format (Deprecated)

Create a `.tfm.conf` file in your project root:

```bash
#!/bin/bash
export __tfm_repo_name='your-project-name'
export __tfm_env_rel_path='terraform/environments'
export __tfm_module_rel_path='terraform/modules'
```

⚠️ **Deprecation Notice**: The legacy `.tfm.conf` format is deprecated and will be removed in v2.0. Use `tf config convert` to migrate to the YAML format.

### Configuration Management

```bash
# Create new YAML configuration
tf config init yaml

# Create new legacy configuration (deprecated)
tf config init legacy

# Convert legacy to YAML format
tf config convert

# Validate current configuration
tf config validate
```

The tool auto-detects git repository root and validates project structure.

## Legacy Support & Migration
tf-manage2 maintains full compatibility with existing [tf-manage](https://github.com/sorinlg/tf-manage) projects while introducing modern configuration management.

### ✨ Key Improvements
- **Native Execution**: Go binary execution is significantly faster than shell script interpretation
- **Reduced Startup Time**: Eliminates bash script parsing overhead
- **Optimized Validation**: Native Go functions for path and file validation instead of shell commands
- **No Shell Dependencies**: Eliminates dependency on specific bash versions or shell features
- **Drop-in Replacement**: Same command interface and behavior with support for existing project structures
- **Modern Configuration**: New YAML format with backward compatibility for existing `.tfm.conf` files
- **Migration Tools**: Built-in conversion tools and deprecation notices
- **Performance Improvements**: 60-75% faster execution compared to the original bash version

### Migration Path

1. **Immediate**: Existing `.tfm.conf` files work without changes
2. **v1.x**: Deprecation warnings guide users to migrate to YAML format
3. **v2.0**: Legacy format support will be removed

**Compatibility Promise**: All v1.x `.tfm.conf` configurations will be supported until v2.0 with clear migration guidance.
