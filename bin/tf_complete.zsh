#compdef tf

# tf-manage2 zsh completion script
# This script provides zsh completion for tf-manage2 using the Go binary for suggestions

_tf_manage2() {
    local context state state_descr line
    local -A opt_args

    # Path to the tf binary - try to find it in PATH
    local tfm_binary
    tfm_binary=$(command -v tf 2>/dev/null)

    if [[ -z "$tfm_binary" ]]; then
        # Fallback: try common installation paths
        for path in "/usr/local/bin/tf" "$HOME/.local/bin/tf" "./tf"; do
            if [[ -x "$path" ]]; then
                tfm_binary="$path"
                break
            fi
        done
    fi

    # Allow overriding the tf binary path with an environment variable
    if [[ -n "$TFM_BINARY" ]]; then
        tfm_binary="$TFM_BINARY"
    fi

    if [[ -z "$tfm_binary" || ! -x "$tfm_binary" ]]; then
        _message "tf binary not found or not executable"
        return 1
    fi

    # Function to safely call tf completion and handle errors
    _call_tf_completion() {
        local completion_cmd="$1"
        shift
        local suggestions

        # Call tf completion command
        suggestions=$("$tfm_binary" __complete "$completion_cmd" "$@" 2>/dev/null)
        if [[ $? -eq 0 && -n "$suggestions" ]]; then
            echo ${(f)suggestions}  # Split on newlines into array
        fi
    }

    # Check if first argument is 'config' to handle config commands differently
    if [[ $words[2] == "config" ]]; then
        # Handle config subcommands
        case $CURRENT in
            3)
                # Second argument after 'config' - suggest config commands (convert, init, validate)
                _tf_config_commands
                ;;
            4)
                # Third argument - handle init formats for 'config init' command
                if [[ $words[3] == "init" ]]; then
                    _tf_config_init_formats
                fi
                # For other config commands (convert, validate), no additional arguments needed
                ;;
        esac
    else
        # Define completion states for regular tf commands using _arguments
        _arguments -C \
            '1:product or command:_tf_first_arg' \
            '2:module:_tf_modules' \
            '3:environment:_tf_environments' \
            '4:config:_tf_configs' \
            '5:action:_tf_actions' \
            '6:workspace:_tf_workspace' \
            '*::arg:->args'

        case $state in
            args)
                # Handle additional arguments if needed
                ;;
        esac
    fi
}

# Completion functions for each argument position
_tf_first_arg() {
    local -a first_args
    local -a products

    # Get products from completion
    products=($(_call_tf_completion "products"))

    # Add config command with description
    first_args=("config:manage tf-manage2 configuration")

    # Add products with generic description
    for product in "${products[@]}"; do
        first_args+=("$product")
    done

    if (( ${#first_args[@]} > 0 )); then
        _describe 'product or command' first_args
    fi
}

_tf_products() {
    local -a products
    products=($(_call_tf_completion "products"))
    if (( ${#products[@]} > 0 )); then
        _describe 'products' products
    fi
}

_tf_modules() {
    local -a modules
    modules=($(_call_tf_completion "modules"))
    if (( ${#modules[@]} > 0 )); then
        _describe 'modules' modules
    fi
}

_tf_environments() {
    local -a environments
    local product="$words[2]"
    local module="$words[3]"

    if [[ -n "$product" && -n "$module" ]]; then
        environments=($(_call_tf_completion "environments" "$product" "$module"))
        if (( ${#environments[@]} > 0 )); then
            _describe 'environments' environments
        fi
    fi
}

_tf_configs() {
    local -a configs
    local product="$words[2]"
    local module="$words[3]"
    local env="$words[4]"

    if [[ -n "$product" && -n "$module" && -n "$env" ]]; then
        configs=($(_call_tf_completion "configs" "$product" "$env" "$module"))
        if (( ${#configs[@]} > 0 )); then
            _describe 'configs' configs
        fi
    fi
}

_tf_actions() {
    local -a actions
    actions=($(_call_tf_completion "actions"))
    if (( ${#actions[@]} > 0 )); then
        _describe 'actions' actions
    fi
}

_tf_workspace() {
    local -a workspaces
    workspaces=($(_call_tf_completion "workspace"))
    if (( ${#workspaces[@]} > 0 )); then
        _describe 'workspace overrides' workspaces
    fi
}

_tf_config_commands() {
    local -a config_commands
    local -a raw_commands

    # Get raw commands from completion
    raw_commands=($(_call_tf_completion "config"))

    # Add descriptions for each command
    for cmd in "${raw_commands[@]}"; do
        case "$cmd" in
            "convert")
                config_commands+=("convert:convert legacy .tfm.conf to .tfm.yaml format")
                ;;
            "init")
                config_commands+=("init:create a new configuration file")
                ;;
            "validate")
                config_commands+=("validate:validate the current configuration")
                ;;
            *)
                config_commands+=("$cmd:config command")
                ;;
        esac
    done

    if (( ${#config_commands[@]} > 0 )); then
        _describe 'config commands' config_commands
    fi
}

_tf_config_init_formats() {
    local -a init_formats
    local -a raw_formats

    # Get raw formats from completion
    raw_formats=($(_call_tf_completion "config_init"))

    # Add descriptions for each format
    for format in "${raw_formats[@]}"; do
        case "$format" in
            "yaml")
                init_formats+=("yaml:create modern .tfm.yaml configuration (recommended)")
                ;;
            "legacy")
                init_formats+=("legacy:create legacy .tfm.conf configuration (deprecated)")
                ;;
            *)
                init_formats+=("$format:configuration format")
                ;;
        esac
    done

    if (( ${#init_formats[@]} > 0 )); then
        _describe 'config init formats' init_formats
    fi
}

# Register the completion function
_tf_manage2 "$@"
