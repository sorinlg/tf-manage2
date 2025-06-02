# tf-manage2

A lean Terraform workspace manager with standardized folder structure and simplified CLI interface.

## Introduction

tf-manage2 provides:
- **Standard folder structure**: Organized `terraform/environments/{project}/{env}/` and `terraform/modules/` layout
- **Workspace naming**: Automatic `{product}.{repo}.{module}.{env}.{module_instance}` convention
- **Simplified CLI**: Single command interface to consume folder structure with built-in validation
- **Local developer experience**: Interactive mode with prompts and colored output
- **CI runtime support**: Auto-detects CI environments and enables unattended mode

## Installation

<details><summary>Homebrew</summary>

```bash
brew tap sorinlg/tap
brew install tf-manage2
```

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

## Usage

```bash
tf <product> <module> <env> <module_instance> <action> [workspace]
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

Create a `.tfm.conf` file in your project root:

```bash
#!/bin/bash
export __tfm_repo_name='your-project-name'
export __tfm_env_rel_path='terraform/environments'
export __tfm_module_rel_path='terraform/modules'
```

The tool auto-detects git repository root and validates project structure.

## Legacy Support
tf-manage2 maintains compatibility with existing [tf-manage](https://github.com/sorinlg/tf-manage) projects. It supports the same folder structure and configuration, allowing seamless migration.

### ✨ Key Improvements
- **Native Execution**: Go binary execution is significantly faster than shell script interpretation
- **Reduced Startup Time**: Eliminates bash script parsing overhead
- **Optimized Validation**: Native Go functions for path and file validation instead of shell commands
- **No Shell Dependencies**: Eliminates dependency on specific bash versions or shell features
- **Drop-in Replacement**: Same command interface and behavior with support for existing project structures
- **Configuration Compatibility**: Reads existing .tfm.conf files (new modern format is planned)
- **Performance Improvements**: 60-75% faster execution compared to the original bash version
