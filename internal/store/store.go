// Package store keeps codeshot's records in a SQLite database: the problems
// imported from a bank, every attempt at one, and every submission with the
// code that was submitted and how each test went. It is the source of truth;
// an attempt's folder is only where the code is worked on, and can go.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store is an open database.
type Store struct {
	db *sql.DB
}

// migrations brings the schema up to date, one step per version: the database
// is at the version that is the number of steps already run, kept in
// PRAGMA user_version. A step is only ever added at the end.
var migrations = []string{
	`CREATE TABLE problems (
		id            INTEGER PRIMARY KEY,
		slug          TEXT NOT NULL UNIQUE,
		source        TEXT NOT NULL,
		title         TEXT NOT NULL,
		difficulty    TEXT NOT NULL,
		statement     TEXT NOT NULL,
		time_limit_ms INTEGER NOT NULL,
		memory_mb     INTEGER NOT NULL,
		imported_at   INTEGER NOT NULL
	);
	CREATE TABLE problem_tags (
		problem_id INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
		tag        TEXT NOT NULL,
		PRIMARY KEY (problem_id, tag)
	);
	CREATE TABLE tests (
		problem_id INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
		name       TEXT NOT NULL,
		input      TEXT NOT NULL,
		output     TEXT NOT NULL,
		is_sample  INTEGER NOT NULL,
		PRIMARY KEY (problem_id, name)
	);
	CREATE TABLE attempts (
		id         INTEGER PRIMARY KEY,
		problem_id INTEGER NOT NULL REFERENCES problems(id),
		lang       TEXT NOT NULL,
		dir        TEXT NOT NULL,
		created_at INTEGER NOT NULL
	);
	CREATE TABLE submissions (
		id             INTEGER PRIMARY KEY,
		attempt_id     INTEGER NOT NULL REFERENCES attempts(id),
		code           TEXT NOT NULL,
		verdict        TEXT NOT NULL,
		time_ms        INTEGER NOT NULL,
		compile_output TEXT NOT NULL,
		created_at     INTEGER NOT NULL
	);
	CREATE TABLE submission_results (
		submission_id INTEGER NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
		test_name     TEXT NOT NULL,
		status        TEXT NOT NULL,
		time_ms       INTEGER NOT NULL,
		PRIMARY KEY (submission_id, test_name)
	);
	CREATE INDEX attempts_problem ON attempts(problem_id);
	CREATE INDEX submissions_attempt ON submissions(attempt_id);`,
}

// Open opens the database at path, creating it and its directory if they are
// not there, and brings its schema up to date.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	// One connection: the pragmas hold per connection, and codeshot never
	// needs two at once.
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	return s, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	var version int
	if err := s.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return err
	}
	if version > len(migrations) {
		return fmt.Errorf("the database is at version %d, newer than this codeshot knows (%d)", version, len(migrations))
	}
	for i := version; i < len(migrations); i++ {
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(migrations[i]); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
		if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, i+1)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
