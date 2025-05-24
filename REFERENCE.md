# tf-manage2 Reference

This document provides technical reference information for tf-manage2. For how-to guides, see [DEVELOPMENT.md](DEVELOPMENT.md). For an overview of new features, see [WHATS_NEW.md](WHATS_NEW.md).

## Environment Variables

### Execution Mode Control

tf-manage2 automatically detects the execution environment and switches between interactive and unattended modes.

#### Manual Override
- **`TF_EXEC_MODE_OVERRIDE`** - When set to any non-empty value, forces unattended mode regardless of environment

#### CI/CD Platform Detection

tf-manage2 automatically detects the following CI/CD platforms and switches to unattended mode:

| Platform         | Environment Variable | Value               |
| ---------------- | -------------------- | ------------------- |
| GitHub Actions   | `GITHUB_ACTIONS`     | `true`              |
| GitLab CI        | `GITLAB_CI`          | `true`              |
| CircleCI         | `CIRCLECI`           | `true`              |
| Travis CI        | `TRAVIS`             | `true`              |
| Azure DevOps     | `TF_BUILD`           | `True`              |
| Jenkins          | `JENKINS_URL`        | Any non-empty value |
| Jenkins          | `BUILD_NUMBER`       | Any non-empty value |
| Jenkins (legacy) | `USER`               | `jenkins`           |
| Bamboo           | `bamboo_buildKey`    | Any non-empty value |
| TeamCity         | `TEAMCITY_VERSION`   | Any non-empty value |
| Buildkite        | `BUILDKITE`          | `true`              |
| Drone CI         | `DRONE`              | `true`              |
| AWS CodeBuild    | `CODEBUILD_BUILD_ID` | Any non-empty value |
| Generic CI       | `CI`                 | `true` or `1`       |

#### Detection Priority

The detection follows this priority order:
1. `TF_EXEC_MODE_OVERRIDE` (highest priority)
2. CI/CD platform-specific variables
3. Generic `CI` variable
4. Legacy `USER=jenkins` check
5. Default to "operator" mode (interactive)

## Terraform Variables

tf-manage2 automatically injects context variables into all terraform commands that support variables.

### tf-manage Context Variables

These variables are automatically added to terraform commands using `-var` flags:

| Variable              | Description                       | Example Value   |
| --------------------- | --------------------------------- | --------------- |
| `tfm_product`         | Product name from command line    | `project1`      |
| `tfm_repo`            | Repository name from .tfm.conf    | `tfm-project`   |
| `tfm_module`          | Module name from command line     | `sample_module` |
| `tfm_env`             | Environment from command line     | `dev`           |
| `tfm_module_instance` | Module instance from command line | `instance_x`    |

### Usage in Terraform

You can use these variables in your terraform configurations:

```hcl
# variables.tf
variable "tfm_product" {
  description = "Product name from tf-manage"
  type        = string
  default     = ""
}

variable "tfm_repo" {
  description = "Repository name from tf-manage"
  type        = string
  default     = ""
}

variable "tfm_module" {
  description = "Module name from tf-manage"
  type        = string
  default     = ""
}

variable "tfm_env" {
  description = "Environment from tf-manage"
  type        = string
  default     = ""
}

variable "tfm_module_instance" {
  description = "Module instance from tf-manage"
  type        = string
  default     = ""
}

# main.tf - Example usage
resource "null_resource" "example" {
  triggers = {
    # Use tf-manage variables for resource recreation
    product         = var.tfm_product
    environment     = var.tfm_env
    module_instance = var.tfm_module_instance
    # Other triggers...
  }
}
```

### Commands with Variable Injection

The following terraform commands automatically receive tf-manage context variables:

- `plan` - Variables injected with `-var` flags
- `apply` - Variables injected with `-var` flags
- `destroy` - Variables injected with `-var` flags
- `refresh` - Variables injected with `-var` flags
- `import` - Variables injected with `-var` flags

## Configuration Files

### .tfm.conf Format

tf-manage2 reads configuration from `.tfm.conf` files in the project root.

```bash
# Required: Repository name used in workspace naming
repo_name="your-repo-name"

# Optional: Custom paths (defaults shown)
tf_modules_dir="terraform/modules"
tf_environments_dir="terraform/environments"
```

### Configuration Properties

| Property              | Type   | Required | Default                  | Description                                  |
| --------------------- | ------ | -------- | ------------------------ | -------------------------------------------- |
| `repo_name`           | string | Yes      | -                        | Repository name used in workspace generation |
| `tf_modules_dir`      | string | No       | `terraform/modules`      | Path to terraform modules directory          |
| `tf_environments_dir` | string | No       | `terraform/environments` | Path to environments directory               |

## Workspace Naming Convention

tf-manage2 generates terraform workspace names using this pattern:

```
{product}.{repo}.{module}.{env}.{module_instance}
```

### Components

- **product** - First command line argument
- **repo** - Repository name from .tfm.conf
- **module** - Second command line argument
- **env** - Third command line argument (sanitized: `/` → `__`)
- **module_instance** - Fourth command line argument

### Examples

```bash
# Command: tf project1 sample_module dev instance_x plan
# Workspace: project1.my-repo.sample_module.dev.instance_x

# Command: tf project1 sample_module staging/blue instance_x plan
# Workspace: project1.my-repo.sample_module.staging__blue.instance_x
```

## Command Structure

### Basic Syntax

```bash
tf <product> <module> <env> <module_instance> <action> [workspace]
```

### Parameters

| Parameter         | Position | Required | Description                                                |
| ----------------- | -------- | -------- | ---------------------------------------------------------- |
| `product`         | 1        | Yes      | Product name (must exist in environments directory)        |
| `module`          | 2        | Yes      | Module name (must exist in modules directory)              |
| `env`             | 3        | Yes      | Environment name (must exist under product directory)      |
| `module_instance` | 4        | Yes      | Instance identifier (must have corresponding .tfvars file) |
| `action`          | 5        | Yes      | Terraform action to execute                                |
| `workspace`       | 6        | No       | Override workspace name                                    |

### Supported Actions

| Action         | Description             | Interactive | Variables Injected |
| -------------- | ----------------------- | ----------- | ------------------ |
| `init`         | Initialize terraform    | No          | No                 |
| `plan`         | Generate execution plan | No          | Yes                |
| `apply`        | Apply changes           | Yes*        | Yes                |
| `apply_plan`   | Apply saved plan        | No          | No                 |
| `destroy`      | Destroy resources       | Yes*        | Yes                |
| `output`       | Show outputs            | No          | No                 |
| `get`          | Download modules        | No          | No                 |
| `workspace`    | Manage workspaces       | No          | No                 |
| `providers`    | Manage providers        | No          | No                 |
| `import`       | Import resources        | Yes*        | Yes                |
| `taint`        | Mark for recreation     | No          | No                 |
| `untaint`      | Unmark for recreation   | No          | No                 |
| `state`        | State management        | No          | No                 |
| `refresh`      | Refresh state           | No          | Yes                |
| `validate`     | Validate configuration  | No          | No                 |
| `fmt`/`format` | Format files            | No          | No                 |
| `show`         | Show state/plan         | No          | No                 |

*Interactive only in operator mode; auto-approved in unattended mode.

## File Structure Requirements

tf-manage2 expects this directory structure:

```
project-root/
├── .tfm.conf                          # Configuration file
└── terraform/
    ├── environments/                   # Environment configurations
    │   └── <product>/
    │       └── <env>/
    │           └── <module>/
    │               └── <instance>.tfvars
    └── modules/                        # Terraform modules
        └── <module>/
            ├── main.tf
            ├── variables.tf
            └── outputs.tf
```

### Path Resolution

| Path Type        | Location                                       | Purpose                      |
| ---------------- | ---------------------------------------------- | ---------------------------- |
| Module Path      | `{tf_modules_dir}/{module}`                    | Terraform module source code |
| Environment Path | `{tf_environments_dir}/{product}/{env}`        | Environment-specific configs |
| Variable File    | `{env_path}/{module}/{instance}.tfvars`        | Instance configuration       |
| Plan File        | `{env_path}/{module}/{instance}.tfvars.tfplan` | Saved execution plans        |

## Exit Codes

tf-manage2 uses standard exit codes:

| Code | Meaning                                                  |
| ---- | -------------------------------------------------------- |
| 0    | Success                                                  |
| 1    | General error (validation failed, terraform error, etc.) |

Specific terraform command exit codes are passed through unchanged.

## Validation Rules

tf-manage2 performs these validations before executing terraform commands:

1. **Product validation** - Directory `{tf_environments_dir}/{product}` must exist
2. **Repository validation** - `repo_name` in .tfm.conf must be non-empty
3. **Module validation** - Directory `{tf_modules_dir}/{module}` must exist
4. **Environment validation** - Directory `{tf_environments_dir}/{product}/{env}` must exist
5. **Configuration validation** - File `{env_path}/{module}/{instance}.tfvars` must exist

### Validation Bypass

These commands skip workspace validation:
- `workspace` - For workspace management operations
- `init` - For initial terraform setup
- `fmt`/`format` - For file formatting operations
