package cli

import (
	"os/exec"
	"strings"
	"testing"
)

// complete runs the bash completion for words, the last of which is being
// typed, and gives back what it offers.
func complete(t *testing.T, words ...string) []string {
	t.Helper()
	var script strings.Builder
	if err := completionScripts["bash"].Execute(&script, completionData()); err != nil {
		t.Fatal(err)
	}
	quoted := make([]string, len(words))
	for i, w := range words {
		quoted[i] = "'" + w + "'"
	}
	cmd := script.String() + "\nCOMP_WORDS=(codeshot " + strings.Join(quoted, " ") + ")\n" +
		"COMP_CWORD=" + string(rune('0'+len(words))) + "\n_codeshot\necho \"${COMPREPLY[*]}\"\n"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		t.Fatalf("bash: %v", err)
	}
	return strings.Fields(string(out))
}

func TestBashCompletion(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("no bash")
	}
	cases := []struct {
		words []string
		want  string
	}{
		{[]string{"su"}, "submit"},
		{[]string{"up"}, "update"},
		{[]string{"solve", "two-sum", "--lang", "py"}, "python"},
		{[]string{"solve", "--editor", "nv"}, "nvim"},
		{[]string{"list", "--difficulty", "h"}, "hard"},
		{[]string{"clean", "--f"}, "--force"},
		{[]string{"completion", "fi"}, "fish"},
	}
	for _, c := range cases {
		got := complete(t, c.words...)
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("completing %q offers %q, want %q", c.words, got, c.want)
		}
	}
}

func TestCompletionScripts(t *testing.T) {
	for shell := range completionScripts {
		if got := Main([]string{"completion", shell}); got != 0 {
			t.Errorf("completion %s exited %d", shell, got)
		}
	}
	if got := Main([]string{"completion", "tcsh"}); got != 2 {
		t.Errorf("completion tcsh exited %d, want 2", got)
	}
}
