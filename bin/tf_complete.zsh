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

    # Check if we're completing config commands
    if [[ "$words[2]" == "config" ]]; then
        case $CURRENT in
            3)
                # Complete config subcommands
                local -a config_commands
                config_commands=($(_call_tf_completion "config"))
                if (( ${#config_commands[@]} > 0 )); then
                    _describe 'config commands' config_commands
                fi
                ;;
            4)
                # Complete config init formats
                if [[ "$words[3]" == "init" ]]; then
                    local -a init_formats
                    init_formats=($(_call_tf_completion "config_init"))
                    if (( ${#init_formats[@]} > 0 )); then
                        _describe 'config formats' init_formats
                    fi
                fi
                ;;
        esac
        return 0
    fi

    # Define completion states using _arguments for regular terraform commands
    _arguments -C \
        '1:product:_tf_products_or_config' \
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
}

# Completion functions for each argument position
_tf_products_or_config() {
    local -a options
    local -a products
    products=($(_call_tf_completion "products"))

    # Add products and config command
    options=("${products[@]}" "config")

    if (( ${#options[@]} > 0 )); then
        _describe 'products or config' options
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

# Register the completion function
_tf_manage2 "$@"
