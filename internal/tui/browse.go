package tui

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"codeshot/internal/bank"
	"codeshot/internal/home"
	"codeshot/internal/store"
)

// browse is the first screen: the tags down the left, and the problems of the
// one picked on the right, narrowed by difficulty and by a search.
type browse struct {
	problems   []store.ProblemInfo
	categories []category
	category   int
	focus      int // 0: the categories, 1: the problems
	cursor     int
	difficulty string
	search     textinput.Model
	searching  bool
	syncing    bool
	message    string
	err        error
}

// category is a tag, or All, with how many of its problems there are and how
// many are solved.
type category struct {
	name          string
	total, solved int
}

func newBrowse(a *app) *browse {
	b := &browse{search: textinput.New(), focus: 1}
	b.search.Prompt = "/"
	b.search.Placeholder = "search"
	b.load(a)
	return b
}

// load reads the problems again, keeping the category picked when it is
// still there.
func (b *browse) load(a *app) {
	b.problems, b.err = a.store.Problems()
	picked := ""
	if b.category < len(b.categories) {
		picked = b.categories[b.category].name
	}
	counts := map[string]*category{}
	all := category{name: "All"}
	for _, p := range b.problems {
		all.total++
		for _, t := range p.Tags {
			if counts[t] == nil {
				counts[t] = &category{name: t}
			}
			counts[t].total++
			if p.Status == store.Solved {
				counts[t].solved++
			}
		}
		if p.Status == store.Solved {
			all.solved++
		}
	}
	b.categories = []category{all}
	for _, c := range counts {
		b.categories = append(b.categories, *c)
	}
	sort.Slice(b.categories[1:], func(i, j int) bool { return b.categories[1+i].name < b.categories[1+j].name })
	b.category = max(0, slices.IndexFunc(b.categories, func(c category) bool { return c.name == picked }))
	b.cursor = min(b.cursor, max(0, len(b.shown())-1))
}

// shown is the problems the category, the difficulty and the search let
// through.
func (b *browse) shown() []store.ProblemInfo {
	var out []store.ProblemInfo
	query := strings.ToLower(b.search.Value())
	for _, p := range b.problems {
		if b.category > 0 && !slices.Contains(p.Tags, b.categories[b.category].name) {
			continue
		}
		if b.difficulty != "" && p.Difficulty != b.difficulty {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(p.Title+" "+p.Slug), query) {
			continue
		}
		out = append(out, p)
	}
	return out
}

type syncedMsg struct {
	n   int
	err error
}

func (b *browse) update(a *app, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case refreshMsg:
		b.load(a)
		return nil
	case syncedMsg:
		b.syncing = false
		if msg.err != nil {
			b.message = errorStyle.Render(msg.err.Error())
		} else {
			b.message = fmt.Sprintf("%d problems synced", msg.n)
		}
		b.load(a)
		return nil
	case tea.KeyPressMsg:
		if b.searching {
			return b.typeSearch(msg)
		}
		return b.key(a, msg.String())
	}
	return nil
}

func (b *browse) typeSearch(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		b.searching = false
		b.search.Reset()
		b.search.Blur()
		return nil
	case "enter", "down", "up":
		b.searching = false
		b.search.Blur()
		return nil
	}
	var cmd tea.Cmd
	b.search, cmd = b.search.Update(msg)
	b.cursor = 0
	return cmd
}

func (b *browse) key(a *app, k string) tea.Cmd {
	b.message = ""
	shown := b.shown()
	switch k {
	case "q":
		return tea.Quit
	case "tab", "left", "right":
		b.focus = 1 - b.focus
	case "up", "k":
		if b.focus == 0 {
			b.category = max(0, b.category-1)
			b.cursor = 0
		} else {
			b.cursor = max(0, b.cursor-1)
		}
	case "down", "j":
		if b.focus == 0 {
			b.category = min(len(b.categories)-1, b.category+1)
			b.cursor = 0
		} else {
			b.cursor = min(max(0, len(shown)-1), b.cursor+1)
		}
	case "g", "home":
		b.cursor = 0
	case "G", "end":
		b.cursor = max(0, len(shown)-1)
	case "e", "m", "h":
		d := map[string]string{"e": "easy", "m": "medium", "h": "hard"}[k]
		if b.difficulty == d {
			d = ""
		}
		b.difficulty = d
		b.cursor = 0
	case "a":
		b.difficulty = ""
		b.cursor = 0
	case "/":
		b.searching = true
		b.focus = 1
		return b.search.Focus()
	case "esc":
		b.search.Reset()
		b.difficulty = ""
	case "S":
		if !b.syncing {
			b.syncing = true
			b.message = "syncing from " + bank.DefaultURL + "…"
			return syncCmd(a)
		}
	case "enter":
		if b.focus == 0 {
			b.focus = 1
		} else if len(shown) > 0 {
			return b.open(a, shown[b.cursor])
		}
	}
	return nil
}

// open shows a problem.
func (b *browse) open(a *app, p store.ProblemInfo) tea.Cmd {
	a.push(newProblem(a, p))
	return nil
}

// syncCmd imports the default bank in the background.
func syncCmd(a *app) tea.Cmd {
	return func() tea.Msg {
		cache, err := home.Path("bank")
		if err != nil {
			return syncedMsg{err: err}
		}
		n, err := bank.Import(a.store, bank.From("", cache))
		return syncedMsg{n: n, err: err}
	}
}

func (b *browse) view(a *app) string {
	header := b.header()
	footer := help("↑↓", "move", "tab", "tags", "enter", "open", "/", "search", "e m h", "difficulty", "S", "sync", "q", "quit")
	if b.message != "" {
		footer = b.message
	}
	height := a.height - 2

	if b.err != nil {
		return header + "\n" + errorStyle.Render(b.err.Error())
	}
	if len(b.problems) == 0 {
		msg := lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Render("No problems yet"),
			"",
			"Press "+bold.Render("S")+" to sync them from "+bank.DefaultURL+",",
			"or run "+bold.Render("codeshot sync --from <folder>")+" with a checkout of it.",
		)
		return header + "\n" + lipgloss.Place(a.width, height, lipgloss.Center, lipgloss.Center, msg) + "\n" + footer
	}

	leftWidth := min(28, a.width/3)
	left := box("Tags", b.categoryRows(leftWidth-4, height-2), leftWidth, height, b.focus == 0)
	right := box(b.listTitle(), b.problemRows(a.width-leftWidth-4, height-2), a.width-leftWidth, height, b.focus == 1)
	return header + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + footer
}

func (b *browse) header() string {
	solved, tried := 0, 0
	for _, p := range b.problems {
		switch p.Status {
		case store.Solved:
			solved++
		case store.Tried:
			tried++
		}
	}
	return titleStyle.Render("codeshot") + faint.Render(fmt.Sprintf("  %d problems  ", len(b.problems))) +
		statusMark[store.Solved] + faint.Render(fmt.Sprintf(" %d solved  ", solved)) +
		statusMark[store.Tried] + faint.Render(fmt.Sprintf(" %d tried", tried))
}

func (b *browse) listTitle() string {
	title := b.categories[b.category].name
	if b.difficulty != "" {
		title += " · " + b.difficulty
	}
	return title
}

func (b *browse) categoryRows(width, height int) string {
	var rows []string
	start, end := window(len(b.categories), b.category, height)
	for i := start; i < end; i++ {
		c := b.categories[i]
		count := fmt.Sprintf("%d/%d", c.solved, c.total)
		name := pad(c.name, width-lipgloss.Width(count)-2)
		row := " " + name + " " + faint.Render(count)
		if i == b.category {
			row = selectedStyle.Render("▌"+name) + " " + count
		}
		rows = append(rows, row)
	}
	return strings.Join(rows, "\n")
}

func (b *browse) problemRows(width, height int) string {
	var rows []string
	if b.searching || b.search.Value() != "" {
		rows = append(rows, b.search.View())
		height--
	}
	shown := b.shown()
	if len(shown) == 0 {
		return strings.Join(append(rows, faint.Render("nothing matches")), "\n")
	}
	tagsWidth := max(0, width-2-34-8)
	start, end := window(len(shown), b.cursor, height)
	for i := start; i < end; i++ {
		p := shown[i]
		title := pad(p.Title, 32)
		if i == b.cursor && b.focus == 1 {
			title = selectedStyle.Render(title)
		}
		row := statusMark[p.Status] + " " + title + "  " +
			difficultyStyle[p.Difficulty].Render(pad(p.Difficulty, 6)) + "  " +
			faint.Render(fit(strings.Join(p.Tags, ", "), tagsWidth))
		if i == b.cursor && b.focus == 1 {
			row = selectedStyle.Render("▌") + row
		} else {
			row = " " + row
		}
		rows = append(rows, row)
	}
	return strings.Join(rows, "\n")
}
