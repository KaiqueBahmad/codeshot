package stub

import (
	"fmt"
	"strings"
)

func rustType(t Type) string {
	base := map[string]string{"int": "i32", "long": "i64", "word": "String", "line": "String", "string": "String", "bool": "bool"}[t.Base]
	for range t.Dims {
		base = "Vec<" + base + ">"
	}
	return base
}

func genRust(s spec) string {
	var params []string
	for _, p := range s.params {
		params = append(params, p.Name+": "+rustType(p.Type))
	}
	zero := map[string]string{"int": "0", "long": "0", "bool": "false", "string": "String::new()"}[s.ret.Base]
	if s.ret.Dims > 0 {
		zero = "vec![]"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "impl Solution {\n    pub fn %s(%s) -> %s {\n        // Write your solution here.\n        %s\n    }\n}\n\n",
		s.name, strings.Join(params, ", "), rustType(s.ret), zero)
	b.WriteString(banner("//", s.name) + "\n")
	b.WriteString(`use std::io::{self, Read, Write};

struct Solution;

struct CodeshotInput {
    buf: Vec<u8>,
    pos: usize,
    mid: bool,
}

impl CodeshotInput {
    fn new() -> Self {
        let mut buf = Vec::new();
        io::stdin().read_to_end(&mut buf).unwrap();
        CodeshotInput { buf, pos: 0, mid: false }
    }

    fn skip(&mut self) {
        while self.pos < self.buf.len() && self.buf[self.pos].is_ascii_whitespace() {
            self.pos += 1;
        }
    }

    fn num(&mut self) -> i64 {
        self.skip();
        let neg = self.pos < self.buf.len() && self.buf[self.pos] == b'-';
        if self.pos < self.buf.len() && (self.buf[self.pos] == b'-' || self.buf[self.pos] == b'+') {
            self.pos += 1;
        }
        let mut v: i64 = 0;
        while self.pos < self.buf.len() && self.buf[self.pos].is_ascii_digit() {
            v = v * 10 + (self.buf[self.pos] - b'0') as i64;
            self.pos += 1;
        }
        self.mid = true;
        if neg { -v } else { v }
    }

    fn word(&mut self) -> String {
        self.skip();
        let start = self.pos;
        while self.pos < self.buf.len() && !self.buf[self.pos].is_ascii_whitespace() {
            self.pos += 1;
        }
        self.mid = true;
        String::from_utf8_lossy(&self.buf[start..self.pos]).into_owned()
    }

    // The line after the one the last token was read from.
    fn line(&mut self) -> String {
        if self.mid {
            while self.pos < self.buf.len() && self.buf[self.pos] != b'\n' {
                self.pos += 1;
            }
            if self.pos < self.buf.len() {
                self.pos += 1;
            }
            self.mid = false;
        }
        let start = self.pos;
        while self.pos < self.buf.len() && self.buf[self.pos] != b'\n' {
            self.pos += 1;
        }
        let mut end = self.pos;
        if end > start && self.buf[end - 1] == b'\r' {
            end -= 1;
        }
        if self.pos < self.buf.len() {
            self.pos += 1;
        }
        String::from_utf8_lossy(&self.buf[start..end]).into_owned()
    }
}

#[allow(unused_variables)]
fn main() {
    let mut input = CodeshotInput::new();
`)
	read := func(base string) string {
		return map[string]string{"int": "input.num() as i32", "long": "input.num()", "word": "input.word()", "line": "input.line()"}[base]
	}
	length := func(t Type) string {
		n := lenExpr(t, func(n string) string { return n })
		return "(" + n + ") as usize"
	}
	for _, v := range s.input {
		t := v.Type
		switch t.Dims {
		case 0:
			fmt.Fprintf(&b, "    let %s: %s = %s;\n", v.Name, rustType(t), read(t.Base))
		case 1:
			fmt.Fprintf(&b, "    let mut %s: %s = Vec::new();\n", v.Name, rustType(t))
			fmt.Fprintf(&b, "    for _ in 0..%s {\n        %s.push(%s);\n    }\n", length(t), v.Name, read(t.Base))
		case 2:
			fmt.Fprintf(&b, "    let mut %s: %s = Vec::new();\n", v.Name, rustType(t))
			fmt.Fprintf(&b, "    for _ in 0..%s {\n        let mut row = Vec::new();\n        for _ in 0..%d {\n            row.push(%s);\n        }\n        %s.push(row);\n    }\n",
				length(t), t.Width, read(t.Base), v.Name)
		}
	}
	var args []string
	for _, p := range s.params {
		args = append(args, p.Name)
	}
	fmt.Fprintf(&b, "    let answer = Solution::%s(%s);\n", s.name, strings.Join(args, ", "))
	b.WriteString("    let mut out = String::new();\n")
	switch {
	case s.ret.Dims == 0:
		b.WriteString("    out.push_str(&answer.to_string());\n    out.push('\\n');\n")
	case s.ret.Dims == 1:
		sep := `" "`
		if s.output == "lines" {
			sep = `"\n"`
		}
		fmt.Fprintf(&b, "    out.push_str(&answer.iter().map(|x| x.to_string()).collect::<Vec<_>>().join(%s));\n    out.push('\\n');\n", sep)
	default:
		if s.output == "count" {
			b.WriteString("    out.push_str(&answer.len().to_string());\n    out.push('\\n');\n")
		}
		b.WriteString("    for row in &answer {\n        out.push_str(&row.iter().map(|x| x.to_string()).collect::<Vec<_>>().join(\" \"));\n        out.push('\\n');\n    }\n")
	}
	b.WriteString("    io::stdout().write_all(out.as_bytes()).unwrap();\n}\n")
	return b.String()
}
