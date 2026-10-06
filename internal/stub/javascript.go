package stub

import (
	"fmt"
	"strings"
)

func jsType(t Type) string {
	base := map[string]string{"int": "number", "long": "number", "word": "string", "line": "string", "string": "string", "bool": "boolean"}[t.Base]
	return base + strings.Repeat("[]", t.Dims)
}

func genJS(s spec) string {
	name := camel(s.name)
	var b strings.Builder
	b.WriteString("/**\n")
	var args []string
	for _, p := range s.params {
		fmt.Fprintf(&b, " * @param {%s} %s\n", jsType(p.Type), camel(p.Name))
		args = append(args, camel(p.Name))
	}
	fmt.Fprintf(&b, " * @return {%s}\n */\n", jsType(s.ret))
	zero := map[string]string{"int": "0", "long": "0", "bool": "false", "string": `""`}[s.ret.Base]
	if s.ret.Dims > 0 {
		zero = "[]"
	}
	fmt.Fprintf(&b, "function %s(%s) {\n    // Write your solution here.\n    return %s;\n}\n\n", name, strings.Join(args, ", "), zero)
	b.WriteString(banner("//", name) + "\n")
	b.WriteString(`class CodeshotInput {
    constructor(text) {
        this.text = text;
        this.pos = 0;
        this.mid = false;
    }

    space(c) {
        return c === 32 || (c >= 9 && c <= 13);
    }

    skip() {
        while (this.pos < this.text.length && this.space(this.text.charCodeAt(this.pos))) this.pos++;
    }

    word() {
        this.skip();
        const start = this.pos;
        while (this.pos < this.text.length && !this.space(this.text.charCodeAt(this.pos))) this.pos++;
        this.mid = true;
        return this.text.slice(start, this.pos);
    }

    num() {
        return Number(this.word());
    }

    // The line after the one the last token was read from.
    line() {
        if (this.mid) {
            const nl = this.text.indexOf("\n", this.pos);
            this.pos = nl < 0 ? this.text.length : nl + 1;
            this.mid = false;
        }
        let end = this.text.indexOf("\n", this.pos);
        if (end < 0) end = this.text.length;
        let line = this.text.slice(this.pos, end);
        this.pos = Math.min(this.text.length, end + 1);
        return line.endsWith("\r") ? line.slice(0, -1) : line;
    }
}

const codeshotIn = new CodeshotInput(require("fs").readFileSync(0, "utf8"));
`)
	read := func(base string) string {
		return map[string]string{"int": "codeshotIn.num()", "long": "codeshotIn.num()", "word": "codeshotIn.word()", "line": "codeshotIn.line()"}[base]
	}
	for _, v := range s.input {
		t := v.Type
		n := camel(v.Name)
		switch t.Dims {
		case 0:
			fmt.Fprintf(&b, "const %s = %s;\n", n, read(t.Base))
		case 1:
			fmt.Fprintf(&b, "const %s = Array.from({ length: %s }, () => %s);\n", n, lenExpr(t, camel), read(t.Base))
		case 2:
			fmt.Fprintf(&b, "const %s = Array.from({ length: %s }, () => Array.from({ length: %d }, () => %s));\n", n, lenExpr(t, camel), t.Width, read(t.Base))
		}
	}
	fmt.Fprintf(&b, "const answer = %s(%s);\n", name, strings.Join(args, ", "))
	switch {
	case s.ret.Dims == 0:
		b.WriteString("process.stdout.write(String(answer) + \"\\n\");\n")
	case s.ret.Dims == 1:
		sep := `" "`
		if s.output == "lines" {
			sep = `"\n"`
		}
		fmt.Fprintf(&b, "process.stdout.write(answer.join(%s) + \"\\n\");\n", sep)
	default:
		if s.output == "count" {
			b.WriteString("process.stdout.write(answer.length + \"\\n\");\n")
		}
		b.WriteString("process.stdout.write(answer.map((row) => row.join(\" \")).join(\"\\n\") + (answer.length ? \"\\n\" : \"\"));\n")
	}
	return b.String()
}
