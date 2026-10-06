// Package workspace makes the folder an attempt is worked on in:
//
//	~/.codeshot/attempts/<slug>/<n>/
//	  main.c        the solution, from the language's template
//	  problem.md    the statement, to read next to it
//	  spec.json     which problem and language, and the limits
//	  tests/        the sample tests, NN.in and NN.out
//	  run.sh        codeshot test on this folder
//	  submit.sh     codeshot submit on this folder
//
// The folder is only somewhere to work. The database keeps the attempt and
// what was submitted from it, so the folder can be deleted and made again.
package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"codeshot/internal/problem"
)

// SpecName is the file that marks a folder as an attempt's.
const SpecName = "spec.json"

// Spec is what spec.json holds. The limits are there to read: a submission
// is judged by the problem's limits as the database has them.
type Spec struct {
	AttemptID   int64  `json:"attempt_id"`
	Problem     string `json:"problem"`
	Title       string `json:"title"`
	Lang        string `json:"lang"`
	File        string `json:"file"`
	TimeLimitMS int    `json:"time_limit_ms"`
	MemoryMB    int    `json:"memory_mb"`
}

// Reserve makes the next free numbered folder for an attempt at slug, under
// root, and gives back its path.
func Reserve(root, slug string) (string, error) {
	parent := filepath.Join(root, slug)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	for n := 1; ; n++ {
		dir := filepath.Join(parent, strconv.Itoa(n))
		err := os.Mkdir(dir, 0o755)
		if err == nil {
			return dir, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", err
		}
	}
}

// Fill writes an attempt's files into dir: code as the solution, the
// statement, the spec, the sample tests and the scripts.
func Fill(dir string, spec Spec, code, statement string, samples []problem.Test) error {
	if err := WriteSpec(dir, spec); err != nil {
		return err
	}
	files := map[string]string{
		spec.File:    code,
		"problem.md": statement,
	}
	for _, t := range samples {
		files[filepath.Join("tests", t.Name+".in")] = t.Input
		files[filepath.Join("tests", t.Name+".out")] = t.Output
	}
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		return err
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	self := executable()
	for name, cmd := range map[string]string{"run.sh": "test", "submit.sh": "submit"} {
		script := fmt.Sprintf("#!/bin/sh\nexec %s %s \"$(dirname \"$0\")\"\n", self, cmd)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// WriteSpec writes spec into the attempt folder dir.
func WriteSpec(dir string, spec Spec) error {
	raw, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, SpecName), append(raw, '\n'), 0o644)
}

// executable is how the scripts call codeshot: by name when it is on the
// PATH, so that they keep working when it is reinstalled somewhere else, and
// by the path of this binary when it is not.
func executable() string {
	if _, err := exec.LookPath("codeshot"); err == nil {
		return "codeshot"
	}
	if self, err := os.Executable(); err == nil {
		return strconv.Quote(self)
	}
	return "codeshot"
}

// Find gives back the attempt folder dir is in, and its spec: dir itself, or
// the closest directory above it with a spec.json.
func Find(dir string) (string, Spec, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", Spec{}, err
	}
	for d := dir; ; d = filepath.Dir(d) {
		raw, err := os.ReadFile(filepath.Join(d, SpecName))
		if err == nil {
			var spec Spec
			if err := json.Unmarshal(raw, &spec); err != nil {
				return "", spec, fmt.Errorf("%s: %w", filepath.Join(d, SpecName), err)
			}
			return d, spec, nil
		}
		if filepath.Dir(d) == d {
			return "", Spec{}, fmt.Errorf("%s is not inside an attempt folder: no %s here or above", dir, SpecName)
		}
	}
}
