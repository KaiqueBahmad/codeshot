package store

import (
	"errors"
	"path/filepath"
	"testing"

	"codeshot/internal/problem"
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "codeshot.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func echo(title string) problem.Problem {
	return problem.Problem{
		Slug:      "echo",
		Meta:      problem.Meta{Title: title, Difficulty: "easy", Tags: []string{"io", "strings"}, TimeLimitMS: 1000, MemoryMB: 64},
		Statement: "# Echo",
		Tests: []problem.Test{
			{Name: "01", Input: "a\n", Output: "a\n", Sample: true},
			{Name: "02", Input: "b\n", Output: "b\n"},
		},
	}
}

func TestSaveProblem(t *testing.T) {
	s := open(t)
	if err := s.SaveProblem(echo("Echo"), "test"); err != nil {
		t.Fatal(err)
	}
	p, err := s.Problem("echo")
	if err != nil {
		t.Fatal(err)
	}
	if p.Title != "Echo" || len(p.Tags) != 2 || p.Status != Untouched {
		t.Fatalf("got %+v", p)
	}
	samples, _ := s.Tests(p.ID, true)
	all, _ := s.Tests(p.ID, false)
	if len(samples) != 1 || len(all) != 2 {
		t.Fatalf("%d samples and %d tests, want 1 and 2", len(samples), len(all))
	}
}

func TestSaveProblem_AgainKeepsHistory(t *testing.T) {
	s := open(t)
	s.SaveProblem(echo("Echo"), "test")
	p, _ := s.Problem("echo")
	attempt, err := s.AddAttempt(p.ID, "c", "/tmp/x")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddSubmission(Submission{AttemptID: attempt, Code: "x", Verdict: "AC"}); err != nil {
		t.Fatal(err)
	}

	if err := s.SaveProblem(echo("Echo Again"), "test"); err != nil {
		t.Fatal(err)
	}
	all, _ := s.Problems()
	if len(all) != 1 || all[0].Title != "Echo Again" || all[0].Status != Solved {
		t.Fatalf("got %+v", all)
	}
}

func TestStatus(t *testing.T) {
	s := open(t)
	s.SaveProblem(echo("Echo"), "test")
	p, _ := s.Problem("echo")
	attempt, _ := s.AddAttempt(p.ID, "c", "/tmp/x")
	if p, _ = s.Problem("echo"); p.Status != Tried {
		t.Fatalf("status %s after an attempt, want tried", p.Status)
	}
	s.AddSubmission(Submission{AttemptID: attempt, Code: "x", Verdict: "WA"})
	if p, _ = s.Problem("echo"); p.Status != Tried {
		t.Fatalf("status %s after a WA, want tried", p.Status)
	}
	s.AddSubmission(Submission{AttemptID: attempt, Code: "x", Verdict: "AC"})
	if p, _ = s.Problem("echo"); p.Status != Solved {
		t.Fatalf("status %s after an AC, want solved", p.Status)
	}
}

func TestSubmission(t *testing.T) {
	s := open(t)
	s.SaveProblem(echo("Echo"), "test")
	p, _ := s.Problem("echo")
	attempt, _ := s.AddAttempt(p.ID, "python", "/tmp/x")
	id, err := s.AddSubmission(Submission{
		AttemptID: attempt, Code: "print(input())", Verdict: "WA", TimeMS: 12,
		Results: []Result{{Test: "01", Status: "AC", TimeMS: 10}, {Test: "02", Status: "WA", TimeMS: 12}},
	})
	if err != nil {
		t.Fatal(err)
	}
	sub, err := s.Submission(id)
	if err != nil {
		t.Fatal(err)
	}
	if sub.Slug != "echo" || sub.Lang != "python" || sub.Code != "print(input())" || len(sub.Results) != 2 {
		t.Fatalf("got %+v", sub)
	}
	if _, err := s.Submission(id + 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
