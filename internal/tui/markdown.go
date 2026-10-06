package tui

import (
	"strings"

	"charm.land/glamour/v2"
)

// markdown renders md for the terminal, wrapped at width, falling back to
// the text as it is when it cannot.
func markdown(a *app, md string, width int) string {
	style := "light"
	if a.dark {
		style = "dark"
	}
	r, err := glamour.NewTermRenderer(glamour.WithStandardStyle(style), glamour.WithWordWrap(width))
	if err != nil {
		return md
	}
	out, err := r.Render(md)
	if err != nil {
		return md
	}
	return strings.Trim(out, "\n")
}

// codeFence is the name markdown highlights each language's code by.
var codeFence = map[string]string{
	"c": "c", "cpp": "cpp", "java": "java", "python": "python",
	"go": "go", "rust": "rust", "javascript": "javascript",
}
