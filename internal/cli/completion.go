package cli

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"codeshot/internal/editor"
	"codeshot/internal/lang"
)

func init() {
	register(command{name: "completion", run: completionCmd})
}

// completionCmd prints the completion script for a shell.
func completionCmd(args []string) int {
	if len(args) != 1 || completionScripts[args[0]] == nil {
		fmt.Fprintln(os.Stderr, "usage: codeshot completion bash|zsh|fish")
		return 2
	}
	var b strings.Builder
	if err := completionScripts[args[0]].Execute(&b, completionData()); err != nil {
		return report(err)
	}
	fmt.Print(b.String())
	return 0
}

// summaries is what completion says each command does, in the order it
// offers them.
var summaries = [][2]string{
	{"tui", "open the TUI"},
	{"sync", "import the problems of a bank"},
	{"list", "print the problems"},
	{"solve", "start a new attempt at a problem"},
	{"test", "run an attempt against the tests in its folder"},
	{"submit", "judge an attempt and record the submission"},
	{"history", "list the submissions, or print one in full"},
	{"restore", "start a new attempt from an old submission"},
	{"clean", "delete attempt folders"},
	{"update", "install the latest release of codeshot"},
	{"completion", "print a completion script for bash, zsh or fish"},
}

// completionData is what the scripts are filled in with, taken from the
// tables codeshot itself works from, so that they never fall behind them.
func completionData() map[string]any {
	var names []string
	for _, c := range summaries {
		names = append(names, c[0])
	}
	var langs []string
	for _, l := range lang.All {
		langs = append(langs, l.ID)
	}
	return map[string]any{
		"Commands":  summaries,
		"Names":     strings.Join(names, " "),
		"Langs":     strings.Join(langs, " "),
		"Editors":   strings.Join(append(editor.IDs(), "none"), " "),
		"WithValue": "--from|--tag|--difficulty|--status|--lang|--editor|--keep-last",
	}
}

// The scripts complete the commands, their flags and the values the flags
// take, and the problem a command takes, which they ask of the very codeshot
// being completed with list --names.
var completionScripts = map[string]*template.Template{
	"bash": template.Must(template.New("bash").Parse(`# codeshot completion for bash
_codeshot() {
    local cur prev cmd i
    local -a args
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    case "$prev" in
        --lang) COMPREPLY=($(compgen -W '{{.Langs}}' -- "$cur")); return ;;
        --editor) COMPREPLY=($(compgen -W '{{.Editors}}' -- "$cur")); return ;;
        --difficulty) COMPREPLY=($(compgen -W 'easy medium hard' -- "$cur")); return ;;
        --status) COMPREPLY=($(compgen -W 'solved tried untouched' -- "$cur")); return ;;
        --from) COMPREPLY=($(compgen -d -- "$cur")); return ;;
        --tag|--keep-last) return ;;
    esac

    # The command, and the arguments after it, flags and their values left out.
    for ((i = 1; i < COMP_CWORD; i++)); do
        case "${COMP_WORDS[i]}" in
            {{.WithValue}}) i=$((i + 1)) ;;
            -*) ;;
            *) if [[ -z "$cmd" ]]; then cmd="${COMP_WORDS[i]}"; else args+=("${COMP_WORDS[i]}"); fi ;;
        esac
    done

    if [[ -z "$cmd" ]]; then
        COMPREPLY=($(compgen -W '{{.Names}} --help --version' -- "$cur"))
        return
    fi
    if [[ "$cur" == -* ]]; then
        local flags
        case "$cmd" in
            sync) flags='--from' ;;
            list) flags='--tag --difficulty --status --names' ;;
            solve) flags='--lang --editor' ;;
            restore) flags='--editor' ;;
            clean) flags='--keep-last --force' ;;
        esac
        COMPREPLY=($(compgen -W "$flags" -- "$cur"))
        return
    fi
    (( ${#args[@]} == 0 )) || return
    case "$cmd" in
        solve|history|clean)
            local IFS=$'\n'
            COMPREPLY=($(compgen -W "$("${COMP_WORDS[0]}" list --names 2>/dev/null | cut -f1)" -- "$cur"))
            ;;
        test|submit) COMPREPLY=($(compgen -d -- "$cur")) ;;
        completion) COMPREPLY=($(compgen -W 'bash zsh fish' -- "$cur")) ;;
    esac
}
complete -F _codeshot codeshot
`)),

	"zsh": template.Must(template.New("zsh").Parse(`# codeshot completion for zsh
_codeshot() {
    local prog=$words[1]
    local -a commands
    commands=({{range .Commands}}
        '{{index . 0}}:{{index . 1}}'{{end}}
    )
    if (( CURRENT == 2 )); then
        _describe 'command' commands
        return
    fi
    case $words[CURRENT-1] in
        --lang) _values 'language' {{.Langs}}; return ;;
        --editor) _values 'editor' {{.Editors}}; return ;;
        --difficulty) _values 'difficulty' easy medium hard; return ;;
        --status) _values 'status' solved tried untouched; return ;;
        --from) _files -/; return ;;
        --tag|--keep-last) return ;;
    esac
    if [[ $PREFIX == -* ]]; then
        case $words[2] in
            sync) compadd -- --from ;;
            list) compadd -- --tag --difficulty --status --names ;;
            solve) compadd -- --lang --editor ;;
            restore) compadd -- --editor ;;
            clean) compadd -- --keep-last --force ;;
        esac
        return
    fi
    case $words[2] in
        solve|history|clean)
            # _describe wants name:description, and list --names gives
            # name<tab>title, so the tab becomes a colon.
            local -a names
            names=(${${(f)"$($prog list --names 2>/dev/null)"}/$'\t'/:})
            _describe 'problem' names
            ;;
        test|submit) _files -/ ;;
        completion) _values 'shell' bash zsh fish ;;
    esac
}
compdef _codeshot codeshot
`)),

	"fish": template.Must(template.New("fish").Parse(`# codeshot completion for fish
complete -c codeshot -f
{{range .Commands}}complete -c codeshot -n __fish_use_subcommand -a {{index . 0}} -d '{{index . 1}}'
{{end}}
# list --names gives name<tab>title, which fish shows as it is.
complete -c codeshot -n '__fish_seen_subcommand_from solve history clean' -a '((commandline -opc)[1] list --names 2>/dev/null)'
complete -c codeshot -n '__fish_seen_subcommand_from test submit' -a '(__fish_complete_directories)'
complete -c codeshot -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'
complete -c codeshot -n '__fish_seen_subcommand_from sync' -l from -r -a '(__fish_complete_directories)'
complete -c codeshot -n '__fish_seen_subcommand_from list' -l tag -x
complete -c codeshot -n '__fish_seen_subcommand_from list' -l difficulty -x -a 'easy medium hard'
complete -c codeshot -n '__fish_seen_subcommand_from list' -l status -x -a 'solved tried untouched'
complete -c codeshot -n '__fish_seen_subcommand_from list' -l names
complete -c codeshot -n '__fish_seen_subcommand_from solve' -l lang -x -a '{{.Langs}}'
complete -c codeshot -n '__fish_seen_subcommand_from solve restore' -l editor -x -a '{{.Editors}}'
complete -c codeshot -n '__fish_seen_subcommand_from clean' -l keep-last -x
complete -c codeshot -n '__fish_seen_subcommand_from clean' -l force
`)),
}
