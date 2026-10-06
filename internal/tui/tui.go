// Package tui is codeshot's terminal interface: browse the problems by tag and
// difficulty, read one with its history, and start solving it.
//
// It is a stack of screens. The one on top gets every message and draws the
// whole terminal; opening something pushes a screen, and esc pops it.
package tui

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"codeshot/internal/store"
)

// screen is one page of the interface.
type screen interface {
	update(a *app, msg tea.Msg) tea.Cmd
	view(a *app) string
}

// app is what every screen shares.
type app struct {
	store         *store.Store
	width, height int
	dark          bool
	stack         []screen
	// folder is printed once the TUI has closed, for someone who asked to
	// be told where an attempt is rather than have it opened.
	folder string
}

// Run opens the TUI on s, and gives back a folder to print once it has
// closed, if one was asked for.
func Run(s *store.Store) (string, error) {
	a := &app{store: s, dark: lipgloss.HasDarkBackground(os.Stdin, os.Stdout)}
	a.push(newBrowse(a))
	_, err := tea.NewProgram(root{a}).Run()
	return a.folder, err
}

func (a *app) push(s screen) { a.stack = append(a.stack, s) }

// pop goes back to the screen under the top one, and tells it to refresh,
// since what it shows may have changed while it was covered.
func (a *app) pop() tea.Cmd {
	if len(a.stack) > 1 {
		a.stack = a.stack[:len(a.stack)-1]
	}
	return func() tea.Msg { return refreshMsg{} }
}

func (a *app) top() screen { return a.stack[len(a.stack)-1] }

// refreshMsg tells a screen to read again what it shows from the database.
type refreshMsg struct{}

// root is the tea.Model, handing each message to the screen on top.
type root struct{ a *app }

func (r root) Init() tea.Cmd { return nil }

func (r root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.a.width, r.a.height = msg.Width, msg.Height
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return r, tea.Quit
		}
	}
	return r, r.a.top().update(r.a, msg)
}

func (r root) View() tea.View {
	v := tea.NewView("")
	if r.a.width > 0 {
		v.SetContent(r.a.top().view(r.a))
	}
	v.AltScreen = true
	v.WindowTitle = "codeshot"
	return v
}
