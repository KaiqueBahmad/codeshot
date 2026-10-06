package workspace

import (
	"path/filepath"
	"strings"
	"testing"

	"codeshot/internal/lang"
	"codeshot/internal/problem"
	"codeshot/internal/store"
	"codeshot/internal/stub"
)

// A problem's own template wins over its function, which wins over the
// language's plain template.
func TestStarter(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "codeshot.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p := problem.Problem{
		Slug: "echo",
		Meta: problem.Meta{
			Title: "Echo", Difficulty: "easy", TimeLimitMS: 1000, MemoryMB: 64,
			Function: &stub.Function{Name: "echo", Input: []string{"s: line"}, Params: []string{"s"}, Returns: "string"},
		},
		Tests:     []problem.Test{{Name: "01", Input: "a\n", Output: "a\n"}},
		Templates: map[string]string{"main.c": "/* the problem's own */\n"},
	}
	if err := s.SaveProblem(p, "test"); err != nil {
		t.Fatal(err)
	}
	info, err := s.Problem("echo")
	if err != nil {
		t.Fatal(err)
	}

	c, _ := lang.ByID("c")
	py, _ := lang.ByID("python")
	if code, _ := Starter(s, info, c); code != "/* the problem's own */\n" {
		t.Errorf("C starts from %q, want the problem's template", code)
	}
	if code, _ := Starter(s, info, py); !strings.Contains(code, "def echo(s: str) -> str:") {
		t.Errorf("Python starts from\n%s\nwant the function written out", code)
	}
	info.Function = nil
	if code, _ := Starter(s, info, py); code != py.Template {
		t.Errorf("with no function, Python starts from\n%s\nwant its plain template", code)
	}
}
