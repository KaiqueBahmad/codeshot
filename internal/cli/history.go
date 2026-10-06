package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"text/tabwriter"

	"codeshot/internal/judge"
	"codeshot/internal/lang"
	"codeshot/internal/store"
	"codeshot/internal/workspace"
)

func init() {
	register(command{name: "history", run: historyCmd})
	register(command{name: "restore", run: restoreCmd})
}

// historyCmd lists the submissions, of one problem or of all. Given the id of
// a submission instead, it prints that one in full, code and all.
func historyCmd(args []string) int {
	s, err := openStore()
	if err != nil {
		return report(err)
	}
	defer s.Close()
	color := isTerminal(os.Stdout)

	if len(args) > 0 {
		if id, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			sub, err := s.Submission(id)
			if err != nil {
				return report(err)
			}
			printSubmission(os.Stdout, sub, color)
			return 0
		}
	}

	var problemID int64
	if len(args) > 0 {
		p, err := s.Problem(args[0])
		if err != nil {
			return report(err)
		}
		problemID = p.ID
	}
	subs, err := s.Submissions(problemID)
	if err != nil {
		return report(err)
	}
	if len(subs) == 0 {
		fmt.Println("no submissions yet")
		return 0
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, sub := range subs {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%d ms\n", sub.ID, sub.CreatedAt.Format("2006-01-02 15:04"),
			sub.Slug, sub.Lang, paintIf(color, verdictColor[sub.Verdict], fmt.Sprintf("%-3s", sub.Verdict)), sub.TimeMS)
	}
	tw.Flush()
	fmt.Println(paintIf(color, gray, "codeshot history <id> shows one in full, codeshot restore <id> works on it again"))
	return 0
}

func printSubmission(w io.Writer, sub store.Submission, color bool) {
	fmt.Fprintf(w, "submission %d, %s in %s, %s\n", sub.ID, sub.Slug, sub.Lang, sub.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Fprintf(w, "%s, %d ms at most\n\n", paintIf(color, verdictColor[sub.Verdict], judge.Names[sub.Verdict]), sub.TimeMS)
	if sub.CompileOutput != "" {
		fmt.Fprintln(w, indent(clip(sub.CompileOutput, 40, 300), "  "))
		fmt.Fprintln(w)
	}
	for _, r := range sub.Results {
		fmt.Fprintf(w, "  %s  %s  %s\n", r.Test,
			paintIf(color, verdictColor[r.Status], fmt.Sprintf("%-3s", r.Status)),
			paintIf(color, gray, fmt.Sprintf("%d ms", r.TimeMS)))
	}
	fmt.Fprintf(w, "\n%s\n", sub.Code)
}

// restoreCmd starts a new attempt from the code of an old submission.
func restoreCmd(args []string) int {
	flags := flag.NewFlagSet("restore", flag.ContinueOnError)
	editorID := flags.String("editor", "", "editor to open it with; with none, the folder is printed")
	if err := flags.Parse(reorder(args)); err != nil {
		return 2
	}
	id, err := strconv.ParseInt(flags.Arg(0), 10, 64)
	if flags.NArg() != 1 || err != nil {
		fmt.Fprintln(os.Stderr, "usage: codeshot restore <submission id> [--editor <editor>]")
		return 2
	}
	s, err := openStore()
	if err != nil {
		return report(err)
	}
	defer s.Close()

	sub, err := s.Submission(id)
	if err != nil {
		return report(err)
	}
	p, err := s.Problem(sub.Slug)
	if err != nil {
		return report(err)
	}
	l, err := lang.ByID(sub.Lang)
	if err != nil {
		return report(err)
	}
	dir, err := workspace.New(s, p, l, sub.Code)
	if err != nil {
		return report(err)
	}
	return report(openIn(*editorID, dir, l.File))
}
