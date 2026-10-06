package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"codeshot/internal/store"
)

func init() {
	register(command{name: "list", run: listCmd})
}

// listCmd prints the problems, with how far each one has got.
func listCmd(args []string) int {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	tag := flags.String("tag", "", "only problems with this tag")
	difficulty := flags.String("difficulty", "", "only problems of this difficulty: easy, medium or hard")
	status := flags.String("status", "", "only problems with this status: solved, tried or untouched")
	names := flags.Bool("names", false, "print only each problem's slug and title, a tab apart, as completion reads them")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	s, err := openStore()
	if err != nil {
		return report(err)
	}
	defer s.Close()

	all, err := s.Problems()
	if err != nil {
		return report(err)
	}
	if len(all) == 0 && !*names {
		fmt.Fprintln(os.Stderr, "no problems yet: run 'codeshot sync' to fetch them")
		return 1
	}
	var shown []store.ProblemInfo
	for _, p := range all {
		if (*tag == "" || slices.Contains(p.Tags, *tag)) &&
			(*difficulty == "" || p.Difficulty == *difficulty) &&
			(*status == "" || p.Status == *status) {
			shown = append(shown, p)
		}
	}
	if *names {
		for _, p := range shown {
			fmt.Printf("%s\t%s\n", p.Slug, p.Title)
		}
		return 0
	}
	printProblems(os.Stdout, shown, isTerminal(os.Stdout))
	return 0
}

// statusMark is the one character a problem's status is shown as.
var statusMark = map[string]string{
	store.Solved:    "✓",
	store.Tried:     "~",
	store.Untouched: "·",
}

// ANSI colors, used only when writing to a terminal.
const (
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	red    = "\x1b[31m"
	gray   = "\x1b[90m"
	reset  = "\x1b[0m"
)

var difficultyColor = map[string]string{"easy": green, "medium": yellow, "hard": red}

var statusColor = map[string]string{store.Solved: green, store.Tried: yellow, store.Untouched: gray}

func printProblems(w io.Writer, all []store.ProblemInfo, color bool) {
	paint := func(c, s string) string {
		if !color {
			return s
		}
		return c + s + reset
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, p := range all {
		// tabwriter counts escape codes as width. Every color is the same
		// length, so a column stays lined up as long as each of its cells is
		// painted the same number of times.
		fmt.Fprintf(tw, "%s %s\t%s\t%s\t%s\n",
			paint(statusColor[p.Status], statusMark[p.Status]),
			paint(difficultyColor[p.Difficulty], fmt.Sprintf("%-6s", p.Difficulty)),
			p.Slug, p.Title, paint(gray, strings.Join(p.Tags, ", ")))
	}
	tw.Flush()
}

// isTerminal says whether f is a terminal, where colors make sense.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
