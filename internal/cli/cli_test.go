package cli

import (
	"strings"
	"testing"
)

func TestMain_UnknownCommand(t *testing.T) {
	if got := Main([]string{"nope"}); got != 2 {
		t.Fatalf("Main(nope) = %d, want 2", got)
	}
}

func TestMain_Version(t *testing.T) {
	if got := Main([]string{"--version"}); got != 0 {
		t.Fatalf("Main(--version) = %d, want 0", got)
	}
}

func TestReorder(t *testing.T) {
	got := reorder([]string{"two-sum", "--lang", "c", "--editor=nvim"})
	want := []string{"--lang", "c", "--editor=nvim", "two-sum"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("reorder = %q, want %q", got, want)
	}
}
