package workspace

import (
	"codeshot/internal/home"
	"codeshot/internal/lang"
	"codeshot/internal/store"
	"codeshot/internal/stub"
)

// Starter is the code a new attempt at p in l starts from: the problem's
// own template for the language when it has one, else the function it is
// solved in written out for the language, else the language's plain template.
func Starter(s *store.Store, p store.ProblemInfo, l lang.Lang) (string, error) {
	code, ok, err := s.Template(p.ID, l.File)
	if err != nil || ok {
		return code, err
	}
	if p.Function != nil {
		return stub.Generate(l.ID, *p.Function)
	}
	return l.Template, nil
}

// New records a new attempt at p in l, and makes its folder in codeshot's
// directory with code as the solution. It gives back the folder.
func New(s *store.Store, p store.ProblemInfo, l lang.Lang, code string) (string, error) {
	root, err := home.Path("attempts")
	if err != nil {
		return "", err
	}
	samples, err := s.Tests(p.ID, true)
	if err != nil {
		return "", err
	}
	id, err := s.AddAttempt(p.ID, l.ID, "")
	if err != nil {
		return "", err
	}
	dir, err := Make(root, p.Slug, id)
	if err != nil {
		return "", err
	}
	if err := s.SetAttemptDir(id, dir); err != nil {
		return "", err
	}
	spec := Spec{
		AttemptID: id, Problem: p.Slug, Title: p.Title, Lang: l.ID, File: l.File,
		TimeLimitMS: p.TimeLimitMS, MemoryMB: p.MemoryMB,
	}
	return dir, Fill(dir, spec, code, p.Statement, samples)
}
