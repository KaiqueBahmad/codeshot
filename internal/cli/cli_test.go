package cli

import "testing"

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
