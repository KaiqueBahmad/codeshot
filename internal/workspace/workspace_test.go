package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"codeshot/internal/problem"
)

func TestMake(t *testing.T) {
	root := t.TempDir()
	dir, err := Make(root, "two-sum", 12)
	if err != nil {
		t.Fatal(err)
	}
	if dir != filepath.Join(root, "two-sum", "12") {
		t.Fatalf("folder %s", dir)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("no folder made: %v", err)
	}
}

func TestFillAndFind(t *testing.T) {
	dir, err := Make(t.TempDir(), "echo", 7)
	if err != nil {
		t.Fatal(err)
	}
	spec := Spec{AttemptID: 7, Problem: "echo", Lang: "c", File: "main.c", TimeLimitMS: 1000, MemoryMB: 64}
	samples := []problem.Test{{Name: "01", Input: "a\n", Output: "a\n", Sample: true}}
	if err := Fill(dir, spec, "int main(){}", "# Echo", samples); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"main.c", "problem.md", "spec.json", "tests/01.in", "tests/01.out", "run.sh", "submit.sh"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}

	found, got, err := Find(filepath.Join(dir, "tests"))
	if err != nil {
		t.Fatal(err)
	}
	if found != dir || got != spec {
		t.Fatalf("Find = %s %+v, want %s %+v", found, got, dir, spec)
	}
}

func TestFind_Outside(t *testing.T) {
	if _, _, err := Find(t.TempDir()); err == nil {
		t.Fatal("found an attempt where there is none")
	}
}
