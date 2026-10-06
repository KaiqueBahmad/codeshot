// Package editor finds the editors installed here and opens an attempt in one.
package editor

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Editor is an editor that can open a solution.
type Editor struct {
	ID   string // what --editor takes
	Name string // what a menu shows
	// Terminal editors take over the terminal until they exit; the others
	// open a window of their own and give it straight back.
	Terminal bool
	// args gives the arguments that open file inside dir.
	args func(dir, file string) []string
	bin  string
}

// known is every editor looked for, in the order they are offered.
var known = []Editor{
	{ID: "nvim", Name: "Open with NeoVim", Terminal: true, args: fileOnly},
	{ID: "vim", Name: "Open with Vim", Terminal: true, args: fileOnly},
	{ID: "hx", Name: "Open with Helix", Terminal: true, args: fileOnly},
	{ID: "nano", Name: "Open with nano", Terminal: true, args: fileOnly},
	{ID: "emacs", Name: "Open with Emacs", Terminal: true, args: func(dir, file string) []string {
		return []string{"-nw", filepath.Join(dir, file)}
	}},
	{ID: "code", Name: "Open with VS Code", args: folderAndFile},
	{ID: "codium", Name: "Open with VSCodium", args: folderAndFile},
	{ID: "cursor", Name: "Open with Cursor", args: folderAndFile},
	{ID: "zed", Name: "Open with Zed", args: folderAndFile},
	{ID: "subl", Name: "Open with Sublime Text", args: folderAndFile},
}

func fileOnly(dir, file string) []string { return []string{filepath.Join(dir, file)} }

func folderAndFile(dir, file string) []string { return []string{dir, filepath.Join(dir, file)} }

// Installed lists the editors found on the PATH. $VISUAL or $EDITOR, when
// set to something that is not already among them, comes first.
func Installed() []Editor {
	var found []Editor
	seen := map[string]bool{}
	for _, e := range known {
		if bin, err := exec.LookPath(e.ID); err == nil {
			e.bin = bin
			found = append(found, e)
			seen[e.ID] = true
		}
	}
	for _, env := range []string{"VISUAL", "EDITOR"} {
		fields := strings.Fields(os.Getenv(env))
		if len(fields) == 0 || seen[filepath.Base(fields[0])] {
			continue
		}
		bin, err := exec.LookPath(fields[0])
		if err != nil {
			continue
		}
		extra := fields[1:]
		found = append([]Editor{{
			ID: filepath.Base(fields[0]), Name: "Open with $" + env + " (" + filepath.Base(fields[0]) + ")",
			Terminal: true, bin: bin,
			args: func(dir, file string) []string { return append(append([]string{}, extra...), filepath.Join(dir, file)) },
		}}, found...)
		break
	}
	return found
}

// ByID finds an installed editor by its id.
func ByID(id string) (Editor, bool) {
	for _, e := range Installed() {
		if e.ID == id {
			return e, true
		}
	}
	return Editor{}, false
}

// Command is what opens file inside dir in e, run from dir. Running it, and
// with which terminal, is left to the caller.
func (e Editor) Command(dir, file string) *exec.Cmd {
	cmd := exec.Command(e.bin, e.args(dir, file)...)
	cmd.Dir = dir
	return cmd
}
