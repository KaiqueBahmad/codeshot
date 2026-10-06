package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"codeshot/internal/problem"
	"codeshot/internal/stub"
)

// The status of a problem, as far as its history goes.
const (
	Untouched = "untouched" // never attempted
	Tried     = "tried"     // attempted, never accepted
	Solved    = "solved"    // accepted at least once
)

// ProblemInfo is a problem as a listing shows it: everything but its tests.
type ProblemInfo struct {
	ID          int64
	Slug        string
	Source      string
	Title       string
	Difficulty  string
	Tags        []string
	Statement   string
	TimeLimitMS int
	MemoryMB    int
	Status      string
	// Function is the function the problem is solved in, or nil.
	Function *stub.Function
}

// ErrNotFound is what a lookup gives back when there is nothing by that name.
var ErrNotFound = errors.New("not found")

// SaveProblem stores p as it came from source, replacing what an earlier
// import of the same slug stored. Attempts and submissions at it are kept.
func (s *Store) SaveProblem(p problem.Problem, source string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	function := ""
	if p.Meta.Function != nil {
		raw, err := json.Marshal(p.Meta.Function)
		if err != nil {
			return err
		}
		function = string(raw)
	}
	var id int64
	err = tx.QueryRow(`
		INSERT INTO problems (slug, source, title, difficulty, statement, time_limit_ms, memory_mb, imported_at, function)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (slug) DO UPDATE SET
			source = excluded.source, title = excluded.title, difficulty = excluded.difficulty,
			statement = excluded.statement, time_limit_ms = excluded.time_limit_ms,
			memory_mb = excluded.memory_mb, imported_at = excluded.imported_at,
			function = excluded.function
		RETURNING id`,
		p.Slug, source, p.Meta.Title, p.Meta.Difficulty, p.Statement,
		p.Meta.TimeLimitMS, p.Meta.MemoryMB, time.Now().UnixMilli(), function,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("saving problem %s: %w", p.Slug, err)
	}

	if _, err := tx.Exec(`DELETE FROM problem_tags WHERE problem_id = ?`, id); err != nil {
		return err
	}
	for _, tag := range p.Meta.Tags {
		if _, err := tx.Exec(`INSERT OR IGNORE INTO problem_tags (problem_id, tag) VALUES (?, ?)`, id, tag); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`DELETE FROM tests WHERE problem_id = ?`, id); err != nil {
		return err
	}
	for _, t := range p.Tests {
		if _, err := tx.Exec(`INSERT INTO tests (problem_id, name, input, output, is_sample) VALUES (?, ?, ?, ?, ?)`,
			id, t.Name, t.Input, t.Output, t.Sample); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`DELETE FROM problem_templates WHERE problem_id = ?`, id); err != nil {
		return err
	}
	for file, code := range p.Templates {
		if _, err := tx.Exec(`INSERT INTO problem_templates (problem_id, file, code) VALUES (?, ?, ?)`, id, file, code); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Template gives back the code a problem has a solution in file start from,
// and whether it has any.
func (s *Store) Template(problemID int64, file string) (string, bool, error) {
	var code string
	err := s.db.QueryRow(`SELECT code FROM problem_templates WHERE problem_id = ? AND file = ?`, problemID, file).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return code, err == nil, err
}

// problemColumns is what scanProblem reads, in its order. The status comes
// from the attempts and submissions at the problem.
const problemColumns = `
	p.id, p.slug, p.source, p.title, p.difficulty, p.statement, p.time_limit_ms, p.memory_mb, p.function,
	COALESCE((SELECT group_concat(tag, ',') FROM (SELECT tag FROM problem_tags WHERE problem_id = p.id ORDER BY tag)), ''),
	CASE
		WHEN EXISTS (SELECT 1 FROM submissions s JOIN attempts a ON a.id = s.attempt_id
		             WHERE a.problem_id = p.id AND s.verdict = 'AC') THEN 'solved'
		WHEN EXISTS (SELECT 1 FROM attempts a WHERE a.problem_id = p.id) THEN 'tried'
		ELSE 'untouched'
	END`

func scanProblem(row interface{ Scan(...any) error }) (ProblemInfo, error) {
	var p ProblemInfo
	var tags, function string
	err := row.Scan(&p.ID, &p.Slug, &p.Source, &p.Title, &p.Difficulty, &p.Statement,
		&p.TimeLimitMS, &p.MemoryMB, &function, &tags, &p.Status)
	if err != nil {
		return p, err
	}
	if tags != "" {
		p.Tags = strings.Split(tags, ",")
	}
	if function != "" {
		p.Function = new(stub.Function)
		if err := json.Unmarshal([]byte(function), p.Function); err != nil {
			return p, fmt.Errorf("problem %s: function: %w", p.Slug, err)
		}
	}
	return p, nil
}

// Problems lists every problem, easiest first and then by title.
func (s *Store) Problems() ([]ProblemInfo, error) {
	rows, err := s.db.Query(`SELECT ` + problemColumns + ` FROM problems p
		ORDER BY CASE p.difficulty WHEN 'easy' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END, p.title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var all []ProblemInfo
	for rows.Next() {
		p, err := scanProblem(rows)
		if err != nil {
			return nil, err
		}
		all = append(all, p)
	}
	return all, rows.Err()
}

// Problem looks up one problem by its slug.
func (s *Store) Problem(slug string) (ProblemInfo, error) {
	p, err := scanProblem(s.db.QueryRow(`SELECT `+problemColumns+` FROM problems p WHERE p.slug = ?`, slug))
	if errors.Is(err, sql.ErrNoRows) {
		return p, fmt.Errorf("problem %s: %w", slug, ErrNotFound)
	}
	return p, err
}

// Tests gives back a problem's tests, in order of name. With samplesOnly, it
// is only the ones a solution gets to see.
func (s *Store) Tests(problemID int64, samplesOnly bool) ([]problem.Test, error) {
	rows, err := s.db.Query(`SELECT name, input, output, is_sample FROM tests
		WHERE problem_id = ? AND (is_sample OR NOT ?) ORDER BY name`, problemID, samplesOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tests []problem.Test
	for rows.Next() {
		var t problem.Test
		if err := rows.Scan(&t.Name, &t.Input, &t.Output, &t.Sample); err != nil {
			return nil, err
		}
		tests = append(tests, t)
	}
	return tests, rows.Err()
}
