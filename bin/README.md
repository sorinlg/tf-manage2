# tf-manage2 Bash Completion

This directory contains the bash completion script for tf-manage2, which provides intelligent auto-completion for the `tf` command.

## Overview

## Architecture

The completion system consists of two components:

1. **Go Completion Commands** (`internal/cli/completion.go`): Provides completion logic using the same configuration and validation as the main CLI
2. **Bash Completion Script** (`bin/tf_complete.sh`): Thin wrapper that calls the Go binary for suggestions

## Installation

### Manual Installation

```bash
# Source the completion script in your shell
source /path/to/tf-manage2/bin/tf_complete.sh

# Or add to your shell profile for permanent installation
echo 'source /path/to/tf-manage2/bin/tf_complete.sh' >> ~/.bashrc
# For zsh users:
echo 'source /path/to/tf-manage2/bin/tf_complete.sh' >> ~/.zshrc
```

### Package Manager Installation

When installing tf-manage2 via package managers, the completion script is typically installed automatically.

## Usage

Once installed, tab completion works for all command positions:

```bash
tf <TAB>              # Shows available projects
tf project1 <TAB>     # Shows available modules
tf project1 sample_module <TAB>  # Shows available environments
tf project1 sample_module dev <TAB>  # Shows available configs
tf project1 sample_module dev instance_x <TAB>  # Shows available actions
tf project1 sample_module dev instance_x plan <TAB>  # Shows workspace options
```

## Completion Commands

The Go binary supports the following completion commands:

| Command                                       | Description                           | Example                       |
| --------------------------------------------- | ------------------------------------- | ----------------------------- |
| `__complete projects`                         | Lists available projects              | `project1`, `project2`        |
| `__complete modules`                          | Lists available modules               | `sample_module`               |
| `__complete environments <product> <module>`  | Lists environments for product/module | `dev`, `staging`, `prod`      |
| `__complete configs <product> <env> <module>` | Lists configuration instances         | `instance_x`, `instance_y`    |
| `__complete actions`                          | Lists terraform actions               | `init`, `plan`, `apply`, etc. |
| `__complete workspace`                        | Suggests workspace override           | `workspace=default`           |
| `__complete repo`                             | Shows repository name                 | `tfm-project`                 |


## Development

### Testing Completion

```bash
# Go to an example project directory
cd examples/tfm-project

# Get the latest build
go build -o tf ../../

# Set the TFM_BINARY environment variable to point to the tf binary
export TFM_BINARY='./tf'

# Source the completion script
source ../../bin/tf_complete.sh

# Test individual completion commands
${TFM_BINARY} __complete projects
${TFM_BINARY} __complete modules
${TFM_BINARY} __complete environments project1 sample_module

# Test bash completion
tf <TAB>
```

### Adding New Completions

1. Add completion logic to `internal/cli/completion.go`
2. Add command handler to `internal/cli/cli.go`
3. Update bash script if needed
4. Test thoroughly

## Troubleshooting

### Completion not working

1. Verify tf binary is in PATH: `which tf`
2. Check if completion script is sourced: `complete -p tf`
3. Ensure you're in a tf-manage project directory
4. Check for configuration errors: `tf __complete projects`

### Missing suggestions

1. Verify directory structure matches tf-manage requirements
2. Check `.tfm.conf` configuration
3. Ensure required directories exist
4. Test completion commands directly: `tf __complete <command>`
