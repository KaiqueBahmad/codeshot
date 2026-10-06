package lang

import "testing"

func TestByID(t *testing.T) {
	for _, id := range []string{"cpp", "C++", "PYTHON"} {
		if _, err := ByID(id); err != nil {
			t.Errorf("ByID(%q): %v", id, err)
		}
	}
	if _, err := ByID("cobol"); err == nil {
		t.Error("ByID(cobol) found a language")
	}
}
