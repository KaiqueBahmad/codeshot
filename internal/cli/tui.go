package cli

import "codeshot/internal/tui"

// runTUI opens the TUI.
func runTUI(args []string) int {
	s, err := openStore()
	if err != nil {
		return report(err)
	}
	defer s.Close()
	return report(tui.Run(s))
}
