package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codeshot/internal/home"
	"codeshot/internal/lang"
	"codeshot/internal/store"
	"codeshot/internal/workspace"
)

func init() {
	register(command{name: "clean", run: cleanCmd})
}

// cleanCmd deletes attempt folders, which the database makes it safe to: what
// was submitted from them stays in it. A folder whose code was never
// submitted, or changed after it was, is kept unless forced, since that work
// is nowhere else.
func cleanCmd(args []string) int {
	flags := flag.NewFlagSet("clean", flag.ContinueOnError)
	keep := flags.Int("keep-last", 0, "keep the newest this many attempts of each problem")
	force := flags.Bool("force", false, "delete folders with work that was not submitted too")
	if err := flags.Parse(reorder(args)); err != nil {
		return 2
	}
	s, err := openStore()
	if err != nil {
		return report(err)
	}
	defer s.Close()
	root, err := home.Path("attempts")
	if err != nil {
		return report(err)
	}

	var problemID int64
	if flags.NArg() > 0 {
		p, err := s.Problem(flags.Arg(0))
		if err != nil {
			return report(err)
		}
		problemID = p.ID
	}
	attempts, err := s.Attempts(problemID)
	if err != nil {
		return report(err)
	}

	seen := map[int64]int{}
	removed, kept := 0, 0
	for _, a := range attempts { // newest first
		seen[a.ProblemID]++
		if seen[a.ProblemID] <= *keep {
			continue
		}
		why, err := keepReason(s, a, root, *force)
		if err != nil {
			return report(err)
		}
		if why == "gone" {
			continue
		}
		if why != "" {
			fmt.Printf("kept    %s: %s\n", a.Dir, why)
			kept++
			continue
		}
		if err := os.RemoveAll(a.Dir); err != nil {
			return report(err)
		}
		os.Remove(filepath.Dir(a.Dir)) // the problem's folder, once it is empty
		fmt.Printf("deleted %s\n", a.Dir)
		removed++
	}
	fmt.Printf("%d %s deleted", removed, plural(removed, "folder"))
	if kept > 0 {
		fmt.Printf(", %d kept with work that was not submitted (--force deletes them too)", kept)
	}
	fmt.Println()
	return 0
}

// keepReason says why an attempt's folder must not be deleted, "gone" when
// there is no folder, or nothing when it can go.
func keepReason(s *store.Store, a store.Attempt, root string, force bool) (string, error) {
	if _, err := os.Stat(a.Dir); os.IsNotExist(err) {
		return "gone", nil
	}
	// Only ever a folder codeshot made, holding the attempt it says it does.
	if rel, err := filepath.Rel(root, a.Dir); err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return "not inside " + root, nil
	}
	if _, spec, err := workspace.Find(a.Dir); err != nil || spec.AttemptID != a.ID {
		return "it holds some other attempt now", nil
	}
	if force {
		return "", nil
	}
	l, err := lang.ByID(a.Lang)
	if err != nil {
		return "", err
	}
	code, err := os.ReadFile(filepath.Join(a.Dir, l.File))
	if err != nil {
		return "", nil
	}
	submitted, ok, err := s.LatestCode(a.ID)
	switch {
	case err != nil:
		return "", err
	case !ok:
		return "never submitted", nil
	case submitted != string(code):
		return "changed since it was last submitted", nil
	}
	return "", nil
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
