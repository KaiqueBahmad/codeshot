package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"codeshot/internal/editor"
	"codeshot/internal/lang"
	"codeshot/internal/store"
	"codeshot/internal/workspace"
)

// noticeMsg is a line for the screen under a picker to show once the picker
// is gone.
type noticeMsg string

// pickLanguage asks which language to solve p in, and starts the attempt.
func pickLanguage(a *app, p store.ProblemInfo) tea.Cmd {
	names := make([]string, len(lang.All))
	for i, l := range lang.All {
		names[i] = l.Name
	}
	picker, cmd := newPicker("which language do you want to use?", p.Title, names, func(a *app, i int) tea.Cmd {
		l := lang.All[i]
		dir, err := workspace.New(a.store, p, l, l.Template)
		a.pop()
		if err != nil {
			return notice(err.Error())
		}
		return pickEditor(a, dir, l.File)
	})
	a.push(picker)
	return cmd
}

// restoreInto starts a new attempt at a submission's problem from its code.
func restoreInto(a *app, sub store.Submission) tea.Cmd {
	p, err := a.store.Problem(sub.Slug)
	if err != nil {
		return notice(err.Error())
	}
	l, err := lang.ByID(sub.Lang)
	if err != nil {
		return notice(err.Error())
	}
	dir, err := workspace.New(a.store, p, l, sub.Code)
	if err != nil {
		return notice(err.Error())
	}
	return pickEditor(a, dir, l.File)
}

// reopenAttempt opens the folder of an earlier attempt again, when it is
// still there.
func reopenAttempt(a *app, at store.Attempt) tea.Cmd {
	if _, spec, err := workspace.Find(at.Dir); err != nil || spec.AttemptID != at.ID {
		return notice("that attempt's folder is gone; open one of its submissions and press r to restore it")
	}
	l, err := lang.ByID(at.Lang)
	if err != nil {
		return notice(err.Error())
	}
	return pickEditor(a, at.Dir, l.File)
}

// folderOption is the answer for someone who will open the folder themselves.
const folderOption = "Just tell me the folder"

// pickEditor asks which editor to open file in dir with, and opens it.
func pickEditor(a *app, dir, file string) tea.Cmd {
	editors := editor.Installed()
	names := make([]string, 0, len(editors)+1)
	for _, e := range editors {
		names = append(names, e.Name)
	}
	names = append(names, folderOption)
	picker, cmd := newPicker("lets go", dir, names, func(a *app, i int) tea.Cmd {
		if i == len(editors) {
			a.folder = dir
			return tea.Quit
		}
		e := editors[i]
		back := a.pop()
		run := e.Command(dir, file)
		hint := fmt.Sprintf("in %s: ./run.sh tries the samples, ./submit.sh submits", shortPath(dir))
		if !e.Terminal {
			if err := run.Start(); err != nil {
				return tea.Batch(back, notice(err.Error()))
			}
			go run.Wait() // reap it once the editor exits
			return tea.Batch(back, notice("opened in "+e.ID+" · "+hint))
		}
		return tea.Sequence(tea.ExecProcess(run, func(err error) tea.Msg {
			if err != nil {
				return noticeMsg(e.ID + ": " + err.Error())
			}
			return noticeMsg(hint)
		}), back)
	})
	a.push(picker)
	return cmd
}

func notice(s string) tea.Cmd { return func() tea.Msg { return noticeMsg(s) } }

// shortPath writes a path under the home directory with a ~.
func shortPath(p string) string {
	if h, err := os.UserHomeDir(); err == nil {
		if rel, err := filepath.Rel(h, p); err == nil && filepath.IsLocal(rel) {
			return "~/" + rel
		}
	}
	return p
}
