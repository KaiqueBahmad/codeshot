package problem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const meta = `{"title":"Echo","difficulty":"easy","tags":["io"],"time_limit_ms":1000,"memory_mb":64,"samples":["01"]}`

func TestLoad(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "echo")
	write(t, filepath.Join(dir, "meta.json"), meta)
	write(t, filepath.Join(dir, "problem.md"), "# Echo")
	write(t, filepath.Join(dir, "tests", "01.in"), "a\n")
	write(t, filepath.Join(dir, "tests", "01.out"), "a\n")
	write(t, filepath.Join(dir, "tests", "02.in"), "b\n")
	write(t, filepath.Join(dir, "tests", "02.out"), "b\n")

	p, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if p.Slug != "echo" || p.Meta.Title != "Echo" || len(p.Tests) != 2 {
		t.Fatalf("got %+v", p)
	}
	if !p.Tests[0].Sample || p.Tests[1].Sample {
		t.Fatalf("only 01 should be a sample: %+v", p.Tests)
	}
}

func TestLoad_MissingOutput(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "echo")
	write(t, filepath.Join(dir, "meta.json"), meta)
	write(t, filepath.Join(dir, "problem.md"), "# Echo")
	write(t, filepath.Join(dir, "tests", "01.in"), "a\n")

	_, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), "no output") {
		t.Fatalf("err = %v, want one about the missing output", err)
	}
}

func TestLoad_BadDifficulty(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "echo")
	write(t, filepath.Join(dir, "meta.json"), strings.Replace(meta, "easy", "trivial", 1))
	write(t, filepath.Join(dir, "problem.md"), "# Echo")
	write(t, filepath.Join(dir, "tests", "01.in"), "a\n")
	write(t, filepath.Join(dir, "tests", "01.out"), "a\n")

	if _, err := Load(dir); err == nil {
		t.Fatal("a difficulty of trivial was accepted")
	}
}

// The problems shipped in the repository must all load.
func TestLoadAll_Repository(t *testing.T) {
	all, err := LoadAll("../../problems")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("no problems in problems/")
	}
}
