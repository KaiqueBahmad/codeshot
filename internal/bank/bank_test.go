package bank

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFrom_Directory(t *testing.T) {
	src := From("../..", t.TempDir())
	if _, ok := src.(Dir); !ok {
		t.Fatalf("From(a directory) = %T, want Dir", src)
	}
	all, err := src.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("no problems read from the repository")
	}
}

func TestFrom_URL(t *testing.T) {
	if src := From("https://example.com/x.git", "/tmp/x"); src.(Git).URL != "https://example.com/x.git" {
		t.Fatalf("got %+v", src)
	}
}

// Cloning the repository this test runs in, and fetching again, pulls.
func TestGit_Fetch(t *testing.T) {
	root, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Skip("not in a git checkout")
	}
	url := "file://" + string(trimNewline(root))
	g := Git{URL: url, Dir: filepath.Join(t.TempDir(), "bank")}
	for range 2 {
		all, err := g.Fetch()
		if err != nil {
			t.Fatal(err)
		}
		if len(all) == 0 {
			t.Fatal("no problems in the clone")
		}
	}
}
