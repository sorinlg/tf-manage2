# tf-manage2 Shell Completion

This directory contains shell completion scripts for tf-manage2, which provide intelligent auto-completion for the `tf` command in both bash and zsh.

## Files

- `tf_complete.sh` - Bash completion script (also works in zsh compatibility mode)
- `tf_complete.zsh` - Native zsh completion script with enhanced features

## Overview

## Architecture

The completion system consists of two components:

1. **Go Completion Commands** (`internal/cli/completion.go`): Provides completion logic using the same configuration and validation as the main CLI
2. **Shell Completion Scripts**:
   - **Bash** (`bin/tf_complete.sh`): Compatible with bash and zsh
   - **Zsh** (`bin/tf_complete.zsh`): Native zsh completion with enhanced features

## Installation

### Manual Installation

#### Bash
```bash
# Source the completion script in your shell
source /path/to/tf-manage2/bin/tf_complete.sh

# Add to your shell profile for permanent installation
echo 'source /path/to/tf-manage2/bin/tf_complete.sh' >> ~/.bashrc
```

#### Zsh
```bash
# For native zsh completion (recommended)
fpath=(/path/to/tf-manage2/bin $fpath)
autoload -U compinit && compinit

# Alternative: Use bash-compatible script
echo 'source /path/to/tf-manage2/bin/tf_complete.sh' >> ~/.zshrc
```

### Package Manager Installation

When installing tf-manage2 via package managers, both completion scripts are typically installed automatically:

- **Homebrew**: Installs both bash and zsh completions automatically
- **Manual packages**: Include both completion scripts in the archive

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

# Source the completion script (bash)
source ../../bin/tf_complete.sh

# Or test zsh completion
cp ../../bin/tf_complete.zsh _tf
autoload -U compinit && compinit

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
3. Update both bash and zsh scripts if needed
4. Test thoroughly in both shells

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
