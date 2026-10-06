// Package bank is where problems come from. A Source gives back problems in
// codeshot's own format, wherever it gets them; sync stores what it gives in
// the database. The codeshot repository's problems/ folder is the first
// source, and an importer for some other site is only another Source.
package bank

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"codeshot/internal/problem"
	"codeshot/internal/store"
)

// DefaultURL is the repository sync clones when it is told no other place.
const DefaultURL = "https://github.com/kaiquebahmad/codeshot.git"

// Source is somewhere problems come from.
type Source interface {
	// Name says where the problems come from, as the database records it.
	Name() string
	// Fetch gives back every problem the source has.
	Fetch() ([]problem.Problem, error)
}

// From picks the source for where: a directory on this machine is read as it
// is, and anything else is taken for a git repository to clone into cache.
// An empty where is CODESHOT_BANK, and failing that DefaultURL.
func From(where, cache string) Source {
	if where == "" {
		where = os.Getenv("CODESHOT_BANK")
	}
	if where == "" {
		where = DefaultURL
	}
	if info, err := os.Stat(where); err == nil && info.IsDir() {
		abs, err := filepath.Abs(where)
		if err == nil {
			where = abs
		}
		return Dir{Path: where}
	}
	return Git{URL: where, Dir: cache}
}

// Dir is a directory of problems on this machine: a checkout of the codeshot
// repository, or a problems folder itself.
type Dir struct {
	Path string
}

func (d Dir) Name() string { return d.Path }

func (d Dir) Fetch() ([]problem.Problem, error) {
	return problem.LoadAll(problemsIn(d.Path))
}

// Git is a git repository laid out like codeshot's, cloned into Dir and pulled
// on every fetch after the first.
type Git struct {
	URL string
	Dir string
}

func (g Git) Name() string { return g.URL }

func (g Git) Fetch() ([]problem.Problem, error) {
	// A clone of some other repository is dropped and made again, so that
	// changing where problems come from never mixes two of them.
	if remote, err := output("git", "-C", g.Dir, "remote", "get-url", "origin"); err == nil && remote == g.URL {
		if err := run("git", "-C", g.Dir, "pull", "--ff-only", "--quiet"); err != nil {
			return nil, err
		}
	} else {
		if err := os.RemoveAll(g.Dir); err != nil {
			return nil, err
		}
		if err := run("git", "clone", "--depth", "1", "--quiet", g.URL, g.Dir); err != nil {
			return nil, err
		}
	}
	return problem.LoadAll(problemsIn(g.Dir))
}

// problemsIn is the problems folder of a repository at root, or root itself
// when it has none and is taken to be one.
func problemsIn(root string) string {
	if info, err := os.Stat(filepath.Join(root, "problems")); err == nil && info.IsDir() {
		return filepath.Join(root, "problems")
	}
	return root
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s %v: %w\n%s", name, args, err, out)
	}
	return nil
}

func output(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", err
	}
	return string(trimNewline(out)), nil
}

func trimNewline(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r') {
		b = b[:len(b)-1]
	}
	return b
}

// Import fetches the problems of src and stores them, and gives back how many
// there were.
func Import(s *store.Store, src Source) (int, error) {
	all, err := src.Fetch()
	if err != nil {
		return 0, err
	}
	for _, p := range all {
		if err := s.SaveProblem(p, src.Name()); err != nil {
			return 0, err
		}
	}
	return len(all), nil
}
