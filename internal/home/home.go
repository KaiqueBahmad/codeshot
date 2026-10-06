// Package home knows where codeshot keeps what it has: ~/.codeshot, or the
// directory CODESHOT_HOME names instead.
package home

import (
	"os"
	"path/filepath"
)

// Dir is codeshot's own directory. It is not created here.
func Dir() (string, error) {
	if dir := os.Getenv("CODESHOT_HOME"); dir != "" {
		return filepath.Abs(dir)
	}
	user, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(user, ".codeshot"), nil
}

// Path is name inside codeshot's directory.
func Path(name ...string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{dir}, name...)...), nil
}
