package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"codeshot/internal/judge"
	"codeshot/internal/store"
)

// The terminal's own palette, so that codeshot looks at home in any theme.
var (
	green   = lipgloss.Color("2")
	yellow  = lipgloss.Color("3")
	red     = lipgloss.Color("1")
	accent  = lipgloss.Color("6")
	subtle  = lipgloss.Color("8")
	bright  = lipgloss.Color("15")
	reverse = lipgloss.NewStyle().Reverse(true)
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(accent)
	faint         = lipgloss.NewStyle().Foreground(subtle)
	bold          = lipgloss.NewStyle().Bold(true)
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(accent)
	errorStyle    = lipgloss.NewStyle().Foreground(red)
)

var difficultyStyle = map[string]lipgloss.Style{
	"easy":   lipgloss.NewStyle().Foreground(green),
	"medium": lipgloss.NewStyle().Foreground(yellow),
	"hard":   lipgloss.NewStyle().Foreground(red),
}

var statusMark = map[string]string{
	store.Solved:    lipgloss.NewStyle().Foreground(green).Render("✓"),
	store.Tried:     lipgloss.NewStyle().Foreground(yellow).Render("~"),
	store.Untouched: faint.Render("·"),
}

var verdictStyle = map[string]lipgloss.Style{
	judge.Accepted:          lipgloss.NewStyle().Foreground(green),
	judge.WrongAnswer:       lipgloss.NewStyle().Foreground(red),
	judge.TimeLimitExceeded: lipgloss.NewStyle().Foreground(yellow),
	judge.MemoryLimit:       lipgloss.NewStyle().Foreground(yellow),
	judge.RuntimeError:      lipgloss.NewStyle().Foreground(red),
	judge.CompilationError:  lipgloss.NewStyle().Foreground(red),
}

// box draws content in a rounded border of exactly width × height cells,
// highlighted when it has the focus.
func box(title, content string, width, height int, focused bool) string {
	border := subtle
	if focused {
		border = accent
	}
	lines := strings.Split(content, "\n")
	inner := max(0, height-2)
	if len(lines) > inner {
		lines = lines[:inner]
	}
	for i, l := range lines {
		lines[i] = fit(l, width-4)
	}
	body := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(border).
		Padding(0, 1).Width(width).Height(height).
		Render(strings.Join(lines, "\n"))
	if title == "" {
		return body
	}
	// Write the title over the top border.
	t := " " + title + " "
	if focused {
		t = selectedStyle.Render(t)
	} else {
		t = faint.Render(t)
	}
	first, rest, _ := strings.Cut(body, "\n")
	first = ansi.Truncate(first, 2, "") + t + ansi.TruncateLeft(first, 2+lipgloss.Width(t), "")
	return first + "\n" + rest
}

// fit cuts s to width cells, with an ellipsis when something was cut.
func fit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return ansi.Truncate(s, width, "…")
}

// pad fills s with spaces up to width cells, cutting it when it is longer.
func pad(s string, width int) string {
	s = fit(s, width)
	return s + strings.Repeat(" ", max(0, width-lipgloss.Width(s)))
}

// window is the part of n rows that shows, height at a time, keeping the
// cursor in sight: the rows from start up to end.
func window(n, cursor, height int) (start, end int) {
	if height <= 0 {
		return 0, 0
	}
	start = max(0, min(cursor-height/2, n-height))
	return start, min(n, start+height)
}

// help draws a line of key hints.
func help(pairs ...string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, bold.Render(pairs[i])+" "+faint.Render(pairs[i+1]))
	}
	return strings.Join(parts, faint.Render("  ·  "))
}
