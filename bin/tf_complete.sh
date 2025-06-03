#!/bin/bash
# tf-manage2 bash completion script
# This script provides bash completion for tf-manage2 using the Go binary for suggestions

_tf_manage2_complete() {
    local cur_word prev_word
    COMPREPLY=()


    # COMP_WORDS is an array of words in the current command line.
    # COMP_CWORD is the index of the current word (the one the cursor is in).
    cur_word="${COMP_WORDS[$COMP_CWORD]}"
    prev_word="${COMP_WORDS[$COMP_CWORD-1]}"

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

    if [[ -z "$tfm_binary" ]]; then
        # If tf binary not found, don't provide completion
        return 1
    fi

    # Allow overriding the tf binary path with an environment variable
    if [[ -n "$TFM_BINARY" ]]; then
        tfm_binary="$TFM_BINARY"
    fi
    if [[ ! -x "$tfm_binary" ]]; then
        echo "Error: tf binary not found or not executable at $tfm_binary"
        return 1
    fi

    # Function to safely call tf completion and handle errors
    _call_tf_completion() {
        local completion_cmd="$1"
        shift
        local suggestions

        # Call tf completion command
        suggestions=$("$tfm_binary" __complete "$completion_cmd" "$@" 2>/dev/null)
        if [[ $? -ne 0 ]]; then
            # non-zero exit code indicates an error
            return 1
        fi
        if
        # Output suggestions if any
        if [[ -n "$suggestions" ]]; then
            echo "$suggestions"
        fi

        return 0
    }

    case $COMP_CWORD in
        1)
            # Complete products
            local suggestions
            suggestions=$(_call_tf_completion "products")
            if [[ $? -eq 0 && -n "$suggestions" ]]; then
                COMPREPLY=($(compgen -W "$suggestions" -- "$cur_word"))
            fi
            ;;
        2)
            # Complete modules
            local suggestions
            suggestions=$(_call_tf_completion "modules")
            if [[ $? -eq 0 && -n "$suggestions" ]]; then
                COMPREPLY=($(compgen -W "$suggestions" -- "$cur_word"))
            fi
            ;;
        3)
            # Complete environments
            local product="${COMP_WORDS[1]}"
            local module="${COMP_WORDS[2]}"
            local suggestions
            suggestions=$(_call_tf_completion "environments" "$product" "$module")
            if [[ $? -eq 0 && -n "$suggestions" ]]; then
                COMPREPLY=($(compgen -W "$suggestions" -- "$cur_word"))
            fi
            ;;
        4)
            # Complete configs/instances
            local product="${COMP_WORDS[1]}"
            local module="${COMP_WORDS[2]}"
            local env="${COMP_WORDS[3]}"
            local suggestions
            suggestions=$(_call_tf_completion "configs" "$product" "$env" "$module")
            if [[ $? -eq 0 && -n "$suggestions" ]]; then
                COMPREPLY=($(compgen -W "$suggestions" -- "$cur_word"))
            fi
            ;;
        5)
            # Complete actions
            local suggestions
            suggestions=$(_call_tf_completion "actions")
            if [[ $? -eq 0 && -n "$suggestions" ]]; then
                COMPREPLY=($(compgen -W "$suggestions" -- "$cur_word"))
            fi
            ;;
        6)
            # Complete workspace override
            local suggestions
            suggestions=$(_call_tf_completion "workspace")
            if [[ $? -eq 0 && -n "$suggestions" ]]; then
                COMPREPLY=($(compgen -W "$suggestions" -- "$cur_word"))
            fi
            ;;
        *)
            # No completion for additional arguments
            COMPREPLY=()
            ;;
    esac

    return 0
}

# Register the completion function for the tf command
complete -F _tf_manage2_complete tf
