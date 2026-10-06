package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Attempt is one go at a problem, in one language, in one folder.
type Attempt struct {
	ID        int64
	ProblemID int64
	Slug      string
	Lang      string
	Dir       string
	CreatedAt time.Time
}

// Result is how one test went in a submission.
type Result struct {
	Test   string
	Status string
	TimeMS int64
}

// Submission is code that was judged against every test of its problem.
type Submission struct {
	ID            int64
	AttemptID     int64
	Slug          string
	Lang          string
	Code          string
	Verdict       string
	TimeMS        int64
	CompileOutput string
	CreatedAt     time.Time
	Results       []Result
}

// AddAttempt records a new attempt at a problem, and gives back its id.
func (s *Store) AddAttempt(problemID int64, lang, dir string) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO attempts (problem_id, lang, dir, created_at) VALUES (?, ?, ?, ?)`,
		problemID, lang, dir, time.Now().UnixMilli())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// SetAttemptDir moves where an attempt's folder is.
func (s *Store) SetAttemptDir(id int64, dir string) error {
	_, err := s.db.Exec(`UPDATE attempts SET dir = ? WHERE id = ?`, dir, id)
	return err
}

const attemptColumns = `a.id, a.problem_id, p.slug, a.lang, a.dir, a.created_at`

func scanAttempt(row interface{ Scan(...any) error }) (Attempt, error) {
	var a Attempt
	var created int64
	err := row.Scan(&a.ID, &a.ProblemID, &a.Slug, &a.Lang, &a.Dir, &created)
	a.CreatedAt = time.UnixMilli(created)
	return a, err
}

// Attempt looks up one attempt by its id.
func (s *Store) Attempt(id int64) (Attempt, error) {
	a, err := scanAttempt(s.db.QueryRow(`SELECT `+attemptColumns+`
		FROM attempts a JOIN problems p ON p.id = a.problem_id WHERE a.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return a, fmt.Errorf("attempt %d: %w", id, ErrNotFound)
	}
	return a, err
}

// Attempts lists the attempts at a problem, or at every problem when
// problemID is 0, newest first.
func (s *Store) Attempts(problemID int64) ([]Attempt, error) {
	rows, err := s.db.Query(`SELECT `+attemptColumns+`
		FROM attempts a JOIN problems p ON p.id = a.problem_id
		WHERE ? = 0 OR a.problem_id = ? ORDER BY a.id DESC`, problemID, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var all []Attempt
	for rows.Next() {
		a, err := scanAttempt(rows)
		if err != nil {
			return nil, err
		}
		all = append(all, a)
	}
	return all, rows.Err()
}

// AddSubmission records a judged submission with its results, and gives back
// its id.
func (s *Store) AddSubmission(sub Submission) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO submissions (attempt_id, code, verdict, time_ms, compile_output, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		sub.AttemptID, sub.Code, sub.Verdict, sub.TimeMS, sub.CompileOutput, time.Now().UnixMilli())
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	for _, r := range sub.Results {
		if _, err := tx.Exec(`INSERT INTO submission_results (submission_id, test_name, status, time_ms) VALUES (?, ?, ?, ?)`,
			id, r.Test, r.Status, r.TimeMS); err != nil {
			return 0, err
		}
	}
	return id, tx.Commit()
}

const submissionColumns = `s.id, s.attempt_id, p.slug, a.lang, s.code, s.verdict, s.time_ms, s.compile_output, s.created_at`

const submissionJoins = ` FROM submissions s
	JOIN attempts a ON a.id = s.attempt_id
	JOIN problems p ON p.id = a.problem_id`

func scanSubmission(row interface{ Scan(...any) error }) (Submission, error) {
	var sub Submission
	var created int64
	err := row.Scan(&sub.ID, &sub.AttemptID, &sub.Slug, &sub.Lang, &sub.Code, &sub.Verdict,
		&sub.TimeMS, &sub.CompileOutput, &created)
	sub.CreatedAt = time.UnixMilli(created)
	return sub, err
}

// Submission looks up one submission by its id, with how each test went.
func (s *Store) Submission(id int64) (Submission, error) {
	sub, err := scanSubmission(s.db.QueryRow(`SELECT `+submissionColumns+submissionJoins+` WHERE s.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return sub, fmt.Errorf("submission %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return sub, err
	}
	rows, err := s.db.Query(`SELECT test_name, status, time_ms FROM submission_results
		WHERE submission_id = ? ORDER BY test_name`, id)
	if err != nil {
		return sub, err
	}
	defer rows.Close()
	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.Test, &r.Status, &r.TimeMS); err != nil {
			return sub, err
		}
		sub.Results = append(sub.Results, r)
	}
	return sub, rows.Err()
}

// Submissions lists the submissions at a problem, or at every problem when
// problemID is 0, newest first. Their results are left out; Submission has
// them.
func (s *Store) Submissions(problemID int64) ([]Submission, error) {
	rows, err := s.db.Query(`SELECT `+submissionColumns+submissionJoins+`
		WHERE ? = 0 OR a.problem_id = ? ORDER BY s.id DESC`, problemID, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var all []Submission
	for rows.Next() {
		sub, err := scanSubmission(rows)
		if err != nil {
			return nil, err
		}
		all = append(all, sub)
	}
	return all, rows.Err()
}

// LatestCode gives back the code of the last submission from an attempt, and
// whether there was one.
func (s *Store) LatestCode(attemptID int64) (string, bool, error) {
	var code string
	err := s.db.QueryRow(`SELECT code FROM submissions WHERE attempt_id = ? ORDER BY id DESC LIMIT 1`, attemptID).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return code, err == nil, err
}
