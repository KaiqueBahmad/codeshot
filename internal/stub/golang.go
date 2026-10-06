package stub

import (
	"fmt"
	"strings"
)

// goType writes a type in Go. Both int and long are int there, which is 64
// bits wide on every machine codeshot runs on.
func goType(t Type) string {
	base := map[string]string{"int": "int", "long": "int", "word": "string", "line": "string", "string": "string", "bool": "bool"}[t.Base]
	return strings.Repeat("[]", t.Dims) + base
}

func genGo(s spec) string {
	name := camel(s.name)
	var params []string
	for _, p := range s.params {
		params = append(params, camel(p.Name)+" "+goType(p.Type))
	}
	zero := map[string]string{"int": "0", "long": "0", "bool": "false", "string": `""`}[s.ret.Base]
	if s.ret.Dims > 0 {
		zero = "nil"
	}

	var b strings.Builder
	b.WriteString("package main\n\nimport (\n\t\"bufio\"\n\t\"io\"\n\t\"os\"\n\t\"strconv\"\n)\n\n")
	fmt.Fprintf(&b, "func %s(%s) %s {\n\t// Write your solution here.\n\treturn %s\n}\n\n", name, strings.Join(params, ", "), goType(s.ret), zero)
	b.WriteString(banner("//", name) + "\n")
	b.WriteString(`// strconv prints numbers, which not every function returns.
var _ = strconv.Itoa

type codeshotInput struct {
	buf []byte
	pos int
	mid bool
}

func codeshotSpace(c byte) bool {
	return c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '\f' || c == '\v'
}

func (r *codeshotInput) skip() {
	for r.pos < len(r.buf) && codeshotSpace(r.buf[r.pos]) {
		r.pos++
	}
}

func (r *codeshotInput) num() int {
	r.skip()
	neg := r.pos < len(r.buf) && r.buf[r.pos] == '-'
	if r.pos < len(r.buf) && (r.buf[r.pos] == '-' || r.buf[r.pos] == '+') {
		r.pos++
	}
	v := 0
	for r.pos < len(r.buf) && r.buf[r.pos] >= '0' && r.buf[r.pos] <= '9' {
		v = v*10 + int(r.buf[r.pos]-'0')
		r.pos++
	}
	r.mid = true
	if neg {
		return -v
	}
	return v
}

func (r *codeshotInput) word() string {
	r.skip()
	start := r.pos
	for r.pos < len(r.buf) && !codeshotSpace(r.buf[r.pos]) {
		r.pos++
	}
	r.mid = true
	return string(r.buf[start:r.pos])
}

// line reads the line after the one the last token was read from.
func (r *codeshotInput) line() string {
	if r.mid {
		for r.pos < len(r.buf) && r.buf[r.pos] != '\n' {
			r.pos++
		}
		if r.pos < len(r.buf) {
			r.pos++
		}
		r.mid = false
	}
	start := r.pos
	for r.pos < len(r.buf) && r.buf[r.pos] != '\n' {
		r.pos++
	}
	end := r.pos
	if end > start && r.buf[end-1] == '\r' {
		end--
	}
	if r.pos < len(r.buf) {
		r.pos++
	}
	return string(r.buf[start:end])
}

func main() {
	codeshotBuf, _ := io.ReadAll(os.Stdin)
	codeshotIn := &codeshotInput{buf: codeshotBuf}
	codeshotOut := bufio.NewWriter(os.Stdout)
	defer codeshotOut.Flush()
`)
	read := func(base string) string {
		return map[string]string{"int": "codeshotIn.num()", "long": "codeshotIn.num()", "word": "codeshotIn.word()", "line": "codeshotIn.line()"}[base]
	}
	for _, v := range s.input {
		t := v.Type
		n := camel(v.Name)
		switch t.Dims {
		case 0:
			fmt.Fprintf(&b, "\t%s := %s\n", n, read(t.Base))
		case 1:
			fmt.Fprintf(&b, "\t%s := make(%s, %s)\n", n, goType(t), lenExpr(t, camel))
			fmt.Fprintf(&b, "\tfor i := range %s {\n\t\t%s[i] = %s\n\t}\n", n, n, read(t.Base))
		case 2:
			fmt.Fprintf(&b, "\t%s := make(%s, %s)\n", n, goType(t), lenExpr(t, camel))
			fmt.Fprintf(&b, "\tfor i := range %s {\n\t\t%s[i] = make(%s, %d)\n\t\tfor j := range %s[i] {\n\t\t\t%s[i][j] = %s\n\t\t}\n\t}\n",
				n, n, goType(Type{Base: t.Base, Dims: 1}), t.Width, n, n, read(t.Base))
		}
	}
	for _, v := range s.unpassed() {
		fmt.Fprintf(&b, "\t_ = %s\n", camel(v.Name))
	}
	var args []string
	for _, p := range s.params {
		args = append(args, camel(p.Name))
	}
	fmt.Fprintf(&b, "\tanswer := %s(%s)\n", name, strings.Join(args, ", "))
	str := map[string]string{"int": "strconv.Itoa(%s)", "long": "strconv.Itoa(%s)", "string": "%s", "bool": "strconv.FormatBool(%s)"}[s.ret.Base]
	switch {
	case s.ret.Dims == 0:
		fmt.Fprintf(&b, "\tcodeshotOut.WriteString(%s)\n\tcodeshotOut.WriteByte('\\n')\n", fmt.Sprintf(str, "answer"))
	case s.ret.Dims == 1:
		sep := "' '"
		if s.output == "lines" {
			sep = "'\\n'"
		}
		fmt.Fprintf(&b, "\tfor i, x := range answer {\n\t\tif i > 0 {\n\t\t\tcodeshotOut.WriteByte(%s)\n\t\t}\n\t\tcodeshotOut.WriteString(%s)\n\t}\n\tcodeshotOut.WriteByte('\\n')\n", sep, fmt.Sprintf(str, "x"))
	default:
		if s.output == "count" {
			b.WriteString("\tcodeshotOut.WriteString(strconv.Itoa(len(answer)))\n\tcodeshotOut.WriteByte('\\n')\n")
		}
		fmt.Fprintf(&b, "\tfor _, row := range answer {\n\t\tfor j, x := range row {\n\t\t\tif j > 0 {\n\t\t\t\tcodeshotOut.WriteByte(' ')\n\t\t\t}\n\t\t\tcodeshotOut.WriteString(%s)\n\t\t}\n\t\tcodeshotOut.WriteByte('\\n')\n\t}\n", fmt.Sprintf(str, "x"))
	}
	b.WriteString("}\n")
	return b.String()
}
