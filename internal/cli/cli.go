// Package cli is the command line: what the arguments ask for, the help, and
// carrying the answer out.
package cli

import (
	"fmt"
	"os"

	"codeshot/internal/home"
	"codeshot/internal/store"
)

const usage = "usage: codeshot [command] [arguments]"

// helpHint is printed instead of the help when the arguments are wrong.
const helpHint = "run 'codeshot --help' for usage"

// help is the full text printed for --help and -h.
const help = usage + `

Codeshot is LeetCode in the terminal, with everything kept on this machine.
With no command it opens the TUI to browse problems and solve one.

Commands:
  tui            open the TUI (the same as no command at all)
  sync [--from <git url | dir>]
                 import the problems of a bank, by default the codeshot
                 repository
  list [--tag t] [--difficulty d] [--status s]
                 print the problems: ✓ solved, ~ tried, · untouched
  solve <problem> --lang <language> [--editor <editor>]
                 start a new attempt in a folder of its own, and open it in
                 the editor, or print the folder when none is given
  test [folder]  run an attempt against the tests in its folder; nothing
                 is recorded
  submit [folder]
                 judge an attempt against every test of its problem, hidden
                 ones too, and record the submission

Options:
  -h, --help     print this help and exit
  -v, --version  print the version of codeshot and exit`

// command is one thing codeshot can be asked to do. run gets the arguments
// that come after the command's name and gives back the exit status.
type command struct {
	name string
	run  func(args []string) int
}

// commands is every command, looked up by name. It is filled in by the files
// that implement each one, so adding a command touches only its own file.
var commands = map[string]command{}

func register(c command) { commands[c.name] = c }

// Main carries out what the command line asked for, and gives back the status
// codeshot exits with. args is the arguments without the program name.
func Main(args []string) int {
	if len(args) == 0 {
		return runTUI(nil)
	}
	switch args[0] {
	case "-h", "--help", "help":
		fmt.Println(help)
		return 0
	case "-v", "--version":
		fmt.Println("codeshot", currentVersion())
		return 0
	case "tui":
		return runTUI(args[1:])
	}
	c, ok := commands[args[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "codeshot: unknown command %q\n%s\n", args[0], helpHint)
		return 2
	}
	return c.run(args[1:])
}

// runTUI opens the TUI. Until it exists, it says so.
func runTUI(args []string) int {
	fmt.Println(help)
	return 0
}

// report prints err, if there is one, and turns it into an exit status.
func report(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "codeshot: %v\n", err)
		return 1
	}
	return 0
}

// openStore opens the database in codeshot's directory.
func openStore() (*store.Store, error) {
	path, err := home.Path("codeshot.db")
	if err != nil {
		return nil, err
	}
	return store.Open(path)
}
