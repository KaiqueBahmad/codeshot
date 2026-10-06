// Package problem reads a problem from the directory format the problems/
// folder of the codeshot repository uses:
//
//	<slug>/
//	  problem.md      the statement
//	  meta.json       title, difficulty, tags, limits, and which tests are samples
//	  tests/NN.in     what the solution reads
//	  tests/NN.out    what it must print
//	  templates/      optional: the code a solution starts from, by file name,
//	                  for a language the function in meta.json does not suit
//
// The samples are copied next to a solution for it to try against; the rest
// stay hidden, and only a submission runs them.
package problem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"codeshot/internal/stub"
)

// The difficulties a problem can have, from easiest.
const (
	Easy   = "easy"
	Medium = "medium"
	Hard   = "hard"
)

// Difficulties lists every difficulty, from easiest.
var Difficulties = []string{Easy, Medium, Hard}

// Meta is what meta.json holds.
type Meta struct {
	Title       string   `json:"title"`
	Difficulty  string   `json:"difficulty"`
	Tags        []string `json:"tags"`
	TimeLimitMS int      `json:"time_limit_ms"`
	MemoryMB    int      `json:"memory_mb"`
	Samples     []string `json:"samples"`
	// Function is the function the problem is solved in, from which the
	// code a solution starts from is written. Without one, a solution
	// starts from its language's plain template.
	Function *stub.Function `json:"function,omitempty"`
}

// Test is one input and the output it must give.
type Test struct {
	Name   string
	Input  string
	Output string
	Sample bool
}

// Problem is a whole problem, as read from its directory.
type Problem struct {
	Slug      string
	Meta      Meta
	Statement string
	Tests     []Test
	// Templates is the code a solution starts from in a language, by the
	// file name the language uses, overriding what Function would give.
	Templates map[string]string
}

// Load reads the problem in dir. Its slug is the directory's name.
func Load(dir string) (Problem, error) {
	p := Problem{Slug: filepath.Base(dir)}

	raw, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		return p, fmt.Errorf("problem %s: %w", p.Slug, err)
	}
	if err := json.Unmarshal(raw, &p.Meta); err != nil {
		return p, fmt.Errorf("problem %s: meta.json: %w", p.Slug, err)
	}

	statement, err := os.ReadFile(filepath.Join(dir, "problem.md"))
	if err != nil {
		return p, fmt.Errorf("problem %s: %w", p.Slug, err)
	}
	p.Statement = string(statement)

	ins, err := filepath.Glob(filepath.Join(dir, "tests", "*.in"))
	if err != nil {
		return p, err
	}
	sort.Strings(ins)
	for _, in := range ins {
		name := strings.TrimSuffix(filepath.Base(in), ".in")
		input, err := os.ReadFile(in)
		if err != nil {
			return p, fmt.Errorf("problem %s: %w", p.Slug, err)
		}
		output, err := os.ReadFile(strings.TrimSuffix(in, ".in") + ".out")
		if err != nil {
			return p, fmt.Errorf("problem %s: test %s has no output: %w", p.Slug, name, err)
		}
		p.Tests = append(p.Tests, Test{
			Name:   name,
			Input:  string(input),
			Output: string(output),
			Sample: slices.Contains(p.Meta.Samples, name),
		})
	}

	templates, err := os.ReadDir(filepath.Join(dir, "templates"))
	if err != nil && !os.IsNotExist(err) {
		return p, fmt.Errorf("problem %s: %w", p.Slug, err)
	}
	for _, t := range templates {
		if t.IsDir() {
			continue
		}
		code, err := os.ReadFile(filepath.Join(dir, "templates", t.Name()))
		if err != nil {
			return p, fmt.Errorf("problem %s: %w", p.Slug, err)
		}
		if p.Templates == nil {
			p.Templates = map[string]string{}
		}
		p.Templates[t.Name()] = string(code)
	}

	return p, p.check()
}

// LoadAll reads every problem in the directories under root, in order of slug.
func LoadAll(root string) ([]Problem, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var all []Problem
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		p, err := Load(filepath.Join(root, e.Name()))
		if err != nil {
			return nil, err
		}
		all = append(all, p)
	}
	return all, nil
}

// check says what is wrong with a problem that was read, if anything.
func (p Problem) check() error {
	m := p.Meta
	switch {
	case m.Title == "":
		return fmt.Errorf("problem %s: meta.json has no title", p.Slug)
	case !slices.Contains(Difficulties, m.Difficulty):
		return fmt.Errorf("problem %s: difficulty %q is not one of %s", p.Slug, m.Difficulty, strings.Join(Difficulties, ", "))
	case m.TimeLimitMS <= 0 || m.MemoryMB <= 0:
		return fmt.Errorf("problem %s: meta.json needs a time_limit_ms and a memory_mb", p.Slug)
	case len(p.Tests) == 0:
		return fmt.Errorf("problem %s: has no tests", p.Slug)
	}
	if m.Function != nil {
		if err := m.Function.Check(); err != nil {
			return fmt.Errorf("problem %s: function: %w", p.Slug, err)
		}
	}
	for _, s := range m.Samples {
		if !slices.ContainsFunc(p.Tests, func(t Test) bool { return t.Name == s }) {
			return fmt.Errorf("problem %s: sample %s is not one of its tests", p.Slug, s)
		}
	}
	return nil
}
