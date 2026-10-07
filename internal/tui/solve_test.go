package tui

import (
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestCopyFolderKeepsTUIOpen(t *testing.T) {
	// No installed editors: the copy option is the only selection.
	t.Setenv("PATH", t.TempDir())
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")
	underlying := &picker{}
	a := &app{stack: []screen{underlying}}
	const dir = "/tmp/attempt with spaces"
	pickEditor(a, dir, "main.go")
	p := a.top().(*picker)
	if len(p.options) != 1 || p.options[0] != "Copy folder path" {
		t.Fatalf("options = %v", p.options)
	}
	cmd := p.choose(a, 0)
	if len(a.stack) != 1 || a.top() != underlying {
		t.Fatal("copy did not return to the underlying screen")
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatal("copy did not return clipboard and refresh commands")
	}
	var copied, refreshed, notified bool
	for _, c := range batch {
		msg := c()
		if _, quit := msg.(tea.QuitMsg); quit {
			t.Fatal("copy closed the TUI")
		}
		if reflect.DeepEqual(msg, tea.SetClipboard(dir)()) {
			copied = true
		}
		switch msg.(type) {
		case refreshMsg:
			refreshed = true
		case noticeMsg:
			notified = true
		}
	}
	if !copied || !refreshed || !notified {
		t.Fatalf("clipboard=%v refresh=%v notice=%v", copied, refreshed, notified)
	}
}
