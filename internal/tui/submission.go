package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"codeshot/internal/judge"
	"codeshot/internal/store"
)

// submissionScreen shows one submission in full: the verdict, how each test
// went, and the code that was submitted.
type submissionScreen struct {
	sub      store.Submission
	body     viewport.Model
	rendered int
	message  string
	err      error
}

func newSubmission(a *app, id int64) *submissionScreen {
	s := &submissionScreen{body: viewport.New()}
	s.sub, s.err = a.store.Submission(id)
	return s
}

func (s *submissionScreen) update(a *app, msg tea.Msg) tea.Cmd {
	if n, ok := msg.(noticeMsg); ok {
		s.message = string(n)
		return nil
	}
	if k, ok := msg.(tea.KeyPressMsg); ok {
		s.message = ""
		switch k.String() {
		case "esc", "q":
			return a.pop()
		case "r":
			return s.restore(a)
		}
	}
	var cmd tea.Cmd
	s.body, cmd = s.body.Update(msg)
	return cmd
}

func (s *submissionScreen) restore(a *app) tea.Cmd { return restoreInto(a, s.sub) }

func (s *submissionScreen) view(a *app) string {
	if s.err != nil {
		return errorStyle.Render(s.err.Error())
	}
	sub := s.sub
	header := titleStyle.Render(fmt.Sprintf("Submission #%d", sub.ID)) +
		faint.Render(fmt.Sprintf("  %s  ·  %s  ·  %s", sub.Slug, sub.Lang, sub.CreatedAt.Format("2006-01-02 15:04")))
	footer := help("↑↓", "scroll", "r", "restore into a new attempt", "esc", "back")
	if s.message != "" {
		footer = s.message
	}
	height := a.height - 2
	if s.rendered != a.width {
		s.body.SetContent(s.content(a, a.width-4))
		s.rendered = a.width
	}
	s.body.SetWidth(a.width - 4)
	s.body.SetHeight(height - 2)
	return header + "\n" + box("", s.body.View(), a.width, height, true) + "\n" + footer
}

func (s *submissionScreen) content(a *app, width int) string {
	sub := s.sub
	var b strings.Builder
	badge := lipgloss.NewStyle().Bold(true).Padding(0, 1).Reverse(true).
		Foreground(verdictStyle[sub.Verdict].GetForeground()).Render(strings.ToUpper(judge.Names[sub.Verdict]))
	passed := 0
	for _, r := range sub.Results {
		if r.Status == judge.Accepted {
			passed++
		}
	}
	fmt.Fprintf(&b, "%s  %s\n\n", badge, faint.Render(fmt.Sprintf("%d/%d tests passed, %d ms at most", passed, len(sub.Results), sub.TimeMS)))

	if sub.CompileOutput != "" {
		b.WriteString(errorStyle.Render(strings.TrimRight(sub.CompileOutput, "\n")) + "\n\n")
	}
	// The tests, as many to a line as fit.
	var cells []string
	for _, r := range sub.Results {
		cells = append(cells, fmt.Sprintf("%s %s %s", faint.Render(r.Test), verdictStyle[r.Status].Render(pad(r.Status, 3)), faint.Render(pad(fmt.Sprintf("%d ms", r.TimeMS), 8))))
	}
	perLine := max(1, width/22)
	for i := 0; i < len(cells); i += perLine {
		b.WriteString(strings.Join(cells[i:min(len(cells), i+perLine)], "   ") + "\n")
	}
	b.WriteString("\n")
	b.WriteString(markdown(a, "```"+codeFence[sub.Lang]+"\n"+sub.Code+"\n```", width))
	return b.String()
}
