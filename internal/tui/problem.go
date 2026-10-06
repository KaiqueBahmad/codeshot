package tui

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"codeshot/internal/store"
)

// problemScreen shows a problem: its statement on the left, and on the right
// every attempt at it with the submissions made from each.
type problemScreen struct {
	p        store.ProblemInfo
	rows     []historyRow
	cursor   int
	focus    int // 0: the statement, 1: the history
	text     viewport.Model
	rendered int // the width the statement was last rendered at
	message  string
	err      error
}

// historyRow is a line of the history: an attempt, or a submission under it.
type historyRow struct {
	attempt    *store.Attempt
	submission *store.Submission
}

func newProblem(a *app, p store.ProblemInfo) *problemScreen {
	s := &problemScreen{p: p, text: viewport.New()}
	s.load(a)
	return s
}

func (s *problemScreen) load(a *app) {
	if p, err := a.store.Problem(s.p.Slug); err == nil {
		s.p = p
	}
	attempts, err := a.store.Attempts(s.p.ID)
	if err != nil {
		s.err = err
		return
	}
	subs, err := a.store.Submissions(s.p.ID)
	if err != nil {
		s.err = err
		return
	}
	s.rows = nil
	for i := range attempts {
		s.rows = append(s.rows, historyRow{attempt: &attempts[i]})
		for j := range subs {
			if subs[j].AttemptID == attempts[i].ID {
				s.rows = append(s.rows, historyRow{submission: &subs[j]})
			}
		}
	}
	s.cursor = min(s.cursor, max(0, len(s.rows)-1))
}

func (s *problemScreen) update(a *app, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case refreshMsg:
		s.load(a)
		return nil
	case noticeMsg:
		s.message = string(msg)
		s.load(a)
		return nil
	case tea.KeyPressMsg:
		s.message = ""
		switch msg.String() {
		case "esc", "q":
			return a.pop()
		case "tab":
			s.focus = 1 - s.focus
			return nil
		case "s":
			return s.solve(a)
		}
		if s.focus == 1 {
			return s.historyKey(a, msg.String())
		}
	}
	var cmd tea.Cmd
	s.text, cmd = s.text.Update(msg)
	return cmd
}

func (s *problemScreen) historyKey(a *app, k string) tea.Cmd {
	switch k {
	case "up", "k":
		s.cursor = max(0, s.cursor-1)
	case "down", "j":
		s.cursor = min(max(0, len(s.rows)-1), s.cursor+1)
	case "enter":
		if len(s.rows) == 0 {
			return nil
		}
		row := s.rows[s.cursor]
		if row.submission != nil {
			a.push(newSubmission(a, row.submission.ID))
			return nil
		}
		return s.reopen(a, *row.attempt)
	}
	return nil
}

func (s *problemScreen) solve(a *app) tea.Cmd { return pickLanguage(a, s.p) }

func (s *problemScreen) reopen(a *app, at store.Attempt) tea.Cmd { return reopenAttempt(a, at) }

func (s *problemScreen) view(a *app) string {
	if s.err != nil {
		return errorStyle.Render(s.err.Error())
	}
	header := s.header()
	footer := help("tab", "statement / history", "↑↓", "scroll", "s", "solve", "enter", "open", "esc", "back")
	if s.focus == 1 && len(s.rows) > 0 && s.rows[s.cursor].attempt != nil {
		footer = help("tab", "statement / history", "↑↓", "move", "s", "solve", "enter", "open the folder again", "esc", "back")
	}
	if s.message != "" {
		footer = s.message
	}
	height := a.height - lipgloss.Height(header) - 1

	historyWidth := min(48, a.width/3)
	textWidth := a.width - historyWidth
	if s.rendered != textWidth {
		s.text.SetContent(markdown(a, s.p.Statement, textWidth-4))
		s.rendered = textWidth
	}
	s.text.SetWidth(textWidth - 4)
	s.text.SetHeight(height - 2)

	left := box("Statement", s.text.View(), textWidth, height, s.focus == 0)
	right := box("History", s.historyRows(historyWidth-4, height-2), historyWidth, height, s.focus == 1)
	return header + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + footer
}

func (s *problemScreen) header() string {
	p := s.p
	return statusMark[p.Status] + " " + titleStyle.Render(p.Title) + "  " +
		difficultyStyle[p.Difficulty].Render(p.Difficulty) + faint.Render("  "+strings.Join(p.Tags, ", ")) +
		faint.Render(fmt.Sprintf("  ·  %d ms  ·  %d MB", p.TimeLimitMS, p.MemoryMB))
}

func (s *problemScreen) historyRows(width, height int) string {
	if len(s.rows) == 0 {
		return faint.Render("No attempts yet.") + "\n\n" + "Press " + bold.Render("s") + " to solve it."
	}
	var lines []string
	start, end := window(len(s.rows), s.cursor, height)
	for i := start; i < end; i++ {
		var line string
		if r := s.rows[i]; r.attempt != nil {
			line = fmt.Sprintf("attempt %d · %s", r.attempt.ID, r.attempt.Lang)
			when := faint.Render(" " + r.attempt.CreatedAt.Format("Jan 2 15:04"))
			if _, err := os.Stat(r.attempt.Dir); err != nil {
				when = faint.Render(" (folder deleted)")
			}
			line = pad(line+when, width-1)
		} else {
			sub := r.submission
			verdict := verdictStyle[sub.Verdict].Render(pad(sub.Verdict, 3))
			line = fmt.Sprintf("  #%-4d %s %s", sub.ID, verdict, faint.Render(fmt.Sprintf("%5d ms  %s", sub.TimeMS, sub.CreatedAt.Format("Jan 2 15:04"))))
		}
		if i == s.cursor && s.focus == 1 {
			line = selectedStyle.Render("▌") + line
		} else {
			line = " " + line
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
