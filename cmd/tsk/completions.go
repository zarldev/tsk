package main

import (
	"fmt"
	"os"
)

const bashCompletion = `_tsk() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    commands="add list ls done rm edit clear export config version completion"

    case "$prev" in
        tsk)
            COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
            return
            ;;
        done|rm|edit)
            local ids
            ids=$(tsk list 2>/dev/null | awk '{print $1}')
            COMPREPLY=( $(compgen -W "$ids" -- "$cur") )
            return
            ;;
        list|ls|export)
            COMPREPLY=( $(compgen -W "--done --pending" -- "$cur") )
            return
            ;;
        -p)
            COMPREPLY=( $(compgen -W "h m l high medium low" -- "$cur") )
            return
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
            return
            ;;
    esac

    if [[ "$COMP_CWORD" -ge 2 ]]; then
        case "${COMP_WORDS[1]}" in
            add)
                if [[ "$cur" == -* ]]; then
                    COMPREPLY=( $(compgen -W "-p" -- "$cur") )
                fi
                ;;
            list|ls|export)
                COMPREPLY=( $(compgen -W "--done --pending" -- "$cur") )
                ;;
        esac
    fi
}

complete -F _tsk tsk
`

const zshCompletion = `#compdef tsk

_tsk() {
    local -a commands
    commands=(add list ls done rm edit clear export config version completion)

    if (( CURRENT == 2 )); then
        compadd -a commands
        return
    fi

    case "$words[2]" in
        done|rm|edit)
            local -a ids
            ids=(${(f)"$(tsk list 2>/dev/null | awk '{print $1}')"})
            compadd -a ids
            ;;
        list|ls|export)
            compadd -- --done --pending
            ;;
        add)
            if [[ "$words[CURRENT-1]" == "-p" ]]; then
                compadd -- h m l high medium low
            elif [[ "$words[CURRENT]" == -* ]]; then
                compadd -- -p
            fi
            ;;
        completion)
            compadd -- bash zsh fish
            ;;
    esac
}

_tsk "$@"
`

const fishCompletion = `complete -c tsk -e
complete -c tsk -n __fish_use_subcommand -a "add list ls done rm edit clear export config version completion" -f
complete -c tsk -n "__fish_seen_subcommand_from done rm edit" -a "(tsk list 2>/dev/null | string match -r '^\s*\\d+' | string trim)" -f
complete -c tsk -n "__fish_seen_subcommand_from list ls export" -a "--done --pending" -f
complete -c tsk -n "__fish_seen_subcommand_from add" -a "-p" -f
complete -c tsk -n "__fish_seen_subcommand_from completion" -a "bash zsh fish" -f
`

func cmdCompletion() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: tsk completion <bash|zsh|fish>")
		os.Exit(1)
	}
	switch os.Args[2] {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	default:
		fmt.Fprintf(os.Stderr, "unsupported shell: %s\n", os.Args[2])
		os.Exit(1)
	}
}
