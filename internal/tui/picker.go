package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// picker asks a question with a list of answers. Typing narrows the list;
// enter picks the answer under the cursor.
type picker struct {
	question string
	note     string
	options  []string
	cursor   int
	filter   textinput.Model
	// choose is told which of options was picked, by its index.
	choose func(a *app, i int) tea.Cmd
}

func newPicker(question, note string, options []string, choose func(a *app, i int) tea.Cmd) (*picker, tea.Cmd) {
	p := &picker{question: question, note: note, options: options, choose: choose, filter: textinput.New()}
	p.filter.Prompt = "/"
	p.filter.Placeholder = "search"
	p.filter.SetWidth(40)
	return p, p.filter.Focus()
}

// shown is the indexes of the options the filter lets through.
func (p *picker) shown() []int {
	var out []int
	q := strings.ToLower(p.filter.Value())
	for i, o := range p.options {
		if strings.Contains(strings.ToLower(o), q) {
			out = append(out, i)
		}
	}
	return out
}

func (p *picker) update(a *app, msg tea.Msg) tea.Cmd {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	shown := p.shown()
	switch k.String() {
	case "esc":
		return a.pop()
	case "up", "ctrl+p", "ctrl+k":
		p.cursor = max(0, p.cursor-1)
		return nil
	case "down", "ctrl+n", "ctrl+j", "tab":
		p.cursor = min(max(0, len(shown)-1), p.cursor+1)
		return nil
	case "enter":
		if len(shown) == 0 {
			return nil
		}
		return p.choose(a, shown[p.cursor])
	}
	var cmd tea.Cmd
	p.filter, cmd = p.filter.Update(msg)
	p.cursor = min(p.cursor, max(0, len(p.shown())-1))
	return cmd
}

func (p *picker) view(a *app) string {
	var rows []string
	rows = append(rows, titleStyle.Render("$ "+p.question))
	if p.note != "" {
		rows = append(rows, faint.Render("  "+p.note))
	}
	rows = append(rows, "")
	shown := p.shown()
	for n, i := range shown {
		if n == p.cursor {
			rows = append(rows, selectedStyle.Render("> "+p.options[i]))
		} else {
			rows = append(rows, "  "+p.options[i])
		}
	}
	if len(shown) == 0 {
		rows = append(rows, faint.Render("  nothing matches"))
	}
	rows = append(rows, "", p.filter.View())
	body := lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(rows, "\n"))
	footer := help("↑↓", "move", "type", "search", "enter", "pick", "esc", "back")
	return lipgloss.PlaceVertical(a.height-1, lipgloss.Top, body) + "\n" + footer
}
