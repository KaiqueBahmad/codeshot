package stub

import (
	"fmt"
	"strings"
)

func pyType(t Type) string {
	base := map[string]string{"int": "int", "long": "int", "word": "str", "line": "str", "string": "str", "bool": "bool"}[t.Base]
	for range t.Dims {
		base = "list[" + base + "]"
	}
	return base
}

func pyZero(t Type) string {
	if t.Dims > 0 {
		return "[]"
	}
	return map[string]string{"int": "0", "long": "0", "string": `""`, "bool": "False"}[t.Base]
}

func genPython(s spec) string {
	var params []string
	for _, p := range s.params {
		params = append(params, p.Name+": "+pyType(p.Type))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "def %s(%s) -> %s:\n", s.name, strings.Join(params, ", "), pyType(s.ret))
	fmt.Fprintf(&b, "    # Write your solution here.\n    return %s\n\n\n", pyZero(s.ret))
	b.WriteString(banner("#", s.name) + "\n")
	b.WriteString(`import sys


class _Input:
    def __init__(self, text):
        self.lines = text.split("\n")
        self.row = 0
        self.tokens = []

    def token(self):
        while not self.tokens:
            self.tokens = self.lines[self.row].split()[::-1]
            self.row += 1
        return self.tokens.pop()

    def line(self):
        # The line after the one the last token was read from.
        self.tokens = []
        line = self.lines[self.row].rstrip("\r")
        self.row += 1
        return line


def _main():
    inp = _Input(sys.stdin.read())
`)
	read := func(base string) string {
		switch base {
		case "int", "long":
			return "int(inp.token())"
		case "word":
			return "inp.token()"
		}
		return "inp.line()"
	}
	for _, v := range s.input {
		t := v.Type
		switch t.Dims {
		case 0:
			fmt.Fprintf(&b, "    %s = %s\n", v.Name, read(t.Base))
		case 1:
			fmt.Fprintf(&b, "    %s = [%s for _ in range(%s)]\n", v.Name, read(t.Base), lenExpr(t, func(n string) string { return n }))
		case 2:
			fmt.Fprintf(&b, "    %s = [[%s for _ in range(%d)] for _ in range(%s)]\n", v.Name, read(t.Base), t.Width, lenExpr(t, func(n string) string { return n }))
		}
	}
	var args []string
	for _, p := range s.params {
		args = append(args, p.Name)
	}
	fmt.Fprintf(&b, "    answer = %s(%s)\n", s.name, strings.Join(args, ", "))
	switch {
	case s.ret.Dims == 0 && s.ret.Base == "bool":
		b.WriteString(`    print("true" if answer else "false")` + "\n")
	case s.ret.Dims == 0:
		b.WriteString("    print(answer)\n")
	case s.ret.Dims == 1 && s.output == "lines":
		b.WriteString(`    print("\n".join(map(str, answer)))` + "\n")
	case s.ret.Dims == 1:
		b.WriteString(`    print(" ".join(map(str, answer)))` + "\n")
	default:
		if s.output == "count" {
			b.WriteString("    print(len(answer))\n")
		}
		b.WriteString(`    print("\n".join(" ".join(map(str, row)) for row in answer))` + "\n")
	}
	b.WriteString("\n\n_main()\n")
	return b.String()
}
