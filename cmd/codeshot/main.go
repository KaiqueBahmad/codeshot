// Codeshot brings practicing problems the way LeetCode does to the terminal:
// browse a bank of problems, solve one in the editor of your choice, and
// submit it to a judge that runs it in docker against hidden tests. Nothing
// leaves the machine; everything lives in ~/.codeshot.
package main

import (
	"os"

	"codeshot/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
