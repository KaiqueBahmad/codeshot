package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"codeshot/internal/editor"
	"codeshot/internal/home"
	"codeshot/internal/lang"
	"codeshot/internal/store"
	"codeshot/internal/workspace"
)

func init() {
	register(command{name: "solve", run: solveCmd})
}

// solveCmd starts a new attempt at a problem and opens it.
func solveCmd(args []string) int {
	flags := flag.NewFlagSet("solve", flag.ContinueOnError)
	langID := flags.String("lang", "", "language to solve it in")
	editorID := flags.String("editor", "", "editor to open it with; with none, the folder is printed")
	if err := flags.Parse(reorder(args)); err != nil {
		return 2
	}
	if flags.NArg() != 1 || *langID == "" {
		fmt.Fprintf(os.Stderr, "usage: codeshot solve <problem> --lang <language> [--editor <editor>]\n")
		return 2
	}
	l, err := lang.ByID(*langID)
	if err != nil {
		return report(err)
	}
	s, err := openStore()
	if err != nil {
		return report(err)
	}
	defer s.Close()

	p, err := s.Problem(flags.Arg(0))
	if err != nil {
		return report(err)
	}
	dir, err := NewAttempt(s, p, l, l.Template)
	if err != nil {
		return report(err)
	}
	return report(openIn(*editorID, dir, l.File))
}

// NewAttempt records a new attempt at p in l, and makes its folder with code
// as the solution. It gives back the folder.
func NewAttempt(s *store.Store, p store.ProblemInfo, l lang.Lang, code string) (string, error) {
	root, err := home.Path("attempts")
	if err != nil {
		return "", err
	}
	samples, err := s.Tests(p.ID, true)
	if err != nil {
		return "", err
	}
	dir, err := workspace.Reserve(root, p.Slug)
	if err != nil {
		return "", err
	}
	id, err := s.AddAttempt(p.ID, l.ID, dir)
	if err != nil {
		os.Remove(dir)
		return "", err
	}
	spec := workspace.Spec{
		AttemptID: id, Problem: p.Slug, Title: p.Title, Lang: l.ID, File: l.File,
		TimeLimitMS: p.TimeLimitMS, MemoryMB: p.MemoryMB,
	}
	return dir, workspace.Fill(dir, spec, code, p.Statement, samples)
}

// openIn opens file in dir with the editor whose id is editorID, waiting for
// it when it runs in this terminal. With no editor, it prints the folder.
func openIn(editorID, dir, file string) error {
	if editorID == "" || editorID == "none" {
		fmt.Println(dir)
		return nil
	}
	e, ok := editor.ByID(editorID)
	if !ok {
		return fmt.Errorf("no editor %q found on the PATH", editorID)
	}
	cmd := e.Command(dir, file)
	if !e.Terminal {
		fmt.Println(dir)
		return cmd.Start()
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// reorder moves the flags in args ahead of the rest, so that they can come
// after the arguments too, as in codeshot solve two-sum --lang c. The flag
// package stops at the first argument that is not a flag.
func reorder(args []string) []string {
	var flags, rest []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--":
			rest = append(rest, args[i+1:]...)
			i = len(args)
		case len(a) > 1 && a[0] == '-':
			flags = append(flags, a)
			if !strings.Contains(a, "=") && i+1 < len(args) {
				flags = append(flags, args[i+1])
				i++
			}
		default:
			rest = append(rest, a)
		}
	}
	return append(flags, rest...)
}

