package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"codeshot/internal/judge"
	"codeshot/internal/store"
	"codeshot/internal/workspace"
)

func init() {
	register(command{name: "submit", run: submitCmd})
}

// submitCmd judges an attempt's solution against every test of its problem,
// hidden ones too, and records the submission whatever the verdict.
func submitCmd(args []string) int {
	dir, spec, l, code, err := loadAttempt(args)
	if err != nil {
		return report(err)
	}
	s, err := openStore()
	if err != nil {
		return report(err)
	}
	defer s.Close()

	p, err := s.Problem(spec.Problem)
	if err != nil {
		return report(err)
	}
	attemptID, err := ensureAttempt(s, p, dir, spec)
	if err != nil {
		return report(err)
	}
	tests, err := s.Tests(p.ID, false)
	if err != nil {
		return report(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	color := isTerminal(os.Stdout)
	// The limits are the problem's as the database has them, not the ones in
	// spec.json, which is only there to be read.
	limits := judge.Limits{TimeMS: p.TimeLimitMS, MemoryMB: p.MemoryMB}
	fmt.Printf("submitting %s in %s against %d tests\n", p.Title, l.Name, len(tests))
	outcome, err := judge.Judge(ctx, l, code, tests, limits, judge.Options{
		StopAtFailure: true,
		Log:           func(msg string) { fmt.Fprintln(os.Stderr, paintIf(color, gray, msg+"…")) },
		OnResult:      func(r judge.Result) { printResult(os.Stdout, r, true, color) },
	})
	if err != nil {
		return report(err)
	}

	sub := store.Submission{
		AttemptID: attemptID, Code: code, Verdict: outcome.Verdict,
		TimeMS: outcome.TimeMS, CompileOutput: outcome.CompileOutput,
	}
	for _, r := range outcome.Results {
		sub.Results = append(sub.Results, store.Result{Test: r.Test, Status: r.Status, TimeMS: r.TimeMS})
	}
	id, err := s.AddSubmission(sub)
	if err != nil {
		return report(fmt.Errorf("recording the submission: %w", err))
	}
	printOutcome(os.Stdout, outcome, color)
	fmt.Println(paintIf(color, gray, fmt.Sprintf("recorded as submission %d", id)))
	if outcome.Verdict != judge.Accepted {
		return 1
	}
	return 0
}

// ensureAttempt gives back the id of the attempt in dir as the database has
// it. A folder the database does not know, because it was copied from
// elsewhere or the database was started over, is recorded as a new attempt,
// and one that moved has its new place recorded.
func ensureAttempt(s *store.Store, p store.ProblemInfo, dir string, spec workspace.Spec) (int64, error) {
	a, err := s.Attempt(spec.AttemptID)
	switch {
	case errors.Is(err, store.ErrNotFound) || (err == nil && a.Slug != p.Slug):
		id, err := s.AddAttempt(p.ID, spec.Lang, dir)
		if err != nil {
			return 0, err
		}
		spec.AttemptID = id
		return id, workspace.WriteSpec(dir, spec)
	case err != nil:
		return 0, err
	case a.Dir != dir:
		return a.ID, s.SetAttemptDir(a.ID, dir)
	}
	return a.ID, nil
}
