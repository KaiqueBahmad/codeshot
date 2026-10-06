package cli

import (
	"fmt"

	"codeshot/internal/tui"
)

// runTUI opens the TUI, and prints the folder it was asked for, if one was,
// once it has closed.
func runTUI(args []string) int {
	s, err := openStore()
	if err != nil {
		return report(err)
	}
	defer s.Close()
	folder, err := tui.Run(s)
	if err != nil {
		return report(err)
	}
	if folder != "" {
		fmt.Println(folder)
	}
	return 0
}
