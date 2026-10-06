package stub

import (
	"fmt"
	"strings"
)

func cScalar(base string) string {
	return map[string]string{"int": "int", "long": "long long", "word": "char*", "line": "char*", "string": "char*", "bool": "bool"}[base]
}

// cParams is how a value is passed in C: a list as a pointer and its size,
// as LeetCode does.
func cParams(v Var) []string {
	t := v.Type
	switch t.Dims {
	case 0:
		return []string{cScalar(t.Base) + " " + v.Name}
	case 1:
		return []string{cScalar(t.Base) + "* " + v.Name, "int " + v.Name + "_size"}
	}
	return []string{cScalar(t.Base) + "** " + v.Name, "int " + v.Name + "_size"}
}

func cFormat(base string) string {
	return map[string]string{"int": "%d", "long": "%lld", "string": "%s"}[base]
}

func genC(s spec) string {
	var params []string
	for _, p := range s.params {
		params = append(params, cParams(p)...)
	}
	ret := cScalar(s.ret.Base)
	note := ""
	switch s.ret.Dims {
	case 1:
		ret += "*"
		params = append(params, "int* return_size")
		note = "\n    // Set *return_size to how many values the array returned has."
	case 2:
		ret += "**"
		params = append(params, "int* return_size")
		note = fmt.Sprintf("\n    // Return rows of %d values each, and set *return_size to how many rows.", s.ret.Width)
	}
	for _, p := range s.params {
		if p.Type.Dims == 2 {
			note += fmt.Sprintf("\n    // Each row of %s has %d values.", p.Name, p.Type.Width)
		}
	}
	zero := map[string]string{"int": "0", "long": "0", "bool": "false", "string": `calloc(1, 1)`}[s.ret.Base]
	body := "    return " + zero + ";\n"
	if s.ret.Dims > 0 {
		body = "    *return_size = 0;\n    return NULL;\n"
	}

	var b strings.Builder
	b.WriteString("#include <stdbool.h>\n#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n\n")
	fmt.Fprintf(&b, "%s %s(%s) {\n    // Write your solution here.%s\n%s}\n\n", ret, s.name, strings.Join(params, ", "), note, body)
	b.WriteString(banner("//", s.name) + "\n")
	b.WriteString(`#include <ctype.h>

static char* in_buf;
static size_t in_len, in_pos;
static int in_mid;

static void in_load(void) {
    size_t cap = 1 << 16;
    in_buf = malloc(cap);
    for (;;) {
        size_t got = fread(in_buf + in_len, 1, cap - in_len - 1, stdin);
        if (got == 0) break;
        in_len += got;
        if (in_len + 1 == cap) in_buf = realloc(in_buf, cap *= 2);
    }
    in_buf[in_len] = '\0';
}

static void in_skip(void) {
    while (in_pos < in_len && isspace((unsigned char)in_buf[in_pos])) in_pos++;
}

static long long in_long(void) {
    in_skip();
    char* end;
    long long v = strtoll(in_buf + in_pos, &end, 10);
    in_pos = end - in_buf;
    in_mid = 1;
    return v;
}

static char* in_take(size_t start, size_t end) {
    char* s = malloc(end - start + 1);
    memcpy(s, in_buf + start, end - start);
    s[end - start] = '\0';
    return s;
}

static char* in_word(void) {
    in_skip();
    size_t start = in_pos;
    while (in_pos < in_len && !isspace((unsigned char)in_buf[in_pos])) in_pos++;
    in_mid = 1;
    return in_take(start, in_pos);
}

// The line after the one the last token was read from.
static char* in_line(void) {
    if (in_mid) {
        while (in_pos < in_len && in_buf[in_pos] != '\n') in_pos++;
        if (in_pos < in_len) in_pos++;
        in_mid = 0;
    }
    size_t start = in_pos;
    while (in_pos < in_len && in_buf[in_pos] != '\n') in_pos++;
    size_t end = in_pos;
    if (end > start && in_buf[end - 1] == '\r') end--;
    if (in_pos < in_len) in_pos++;
    return in_take(start, end);
}

int main(void) {
    in_load();
`)
	read := func(base string) string {
		return map[string]string{"int": "(int)in_long()", "long": "in_long()", "word": "in_word()", "line": "in_line()"}[base]
	}
	id := func(n string) string { return n }
	for _, v := range s.input {
		t := v.Type
		elem := cScalar(t.Base)
		switch t.Dims {
		case 0:
			fmt.Fprintf(&b, "    %s %s = %s;\n", elem, v.Name, read(t.Base))
		case 1:
			n := lenExpr(t, id)
			fmt.Fprintf(&b, "    int %s_size = (int)%s;\n", v.Name, n)
			fmt.Fprintf(&b, "    %s* %s = malloc(sizeof(%s) * (%s_size + 1));\n", elem, v.Name, elem, v.Name)
			fmt.Fprintf(&b, "    for (int i = 0; i < %s_size; i++) %s[i] = %s;\n", v.Name, v.Name, read(t.Base))
		case 2:
			n := lenExpr(t, id)
			fmt.Fprintf(&b, "    int %s_size = (int)%s;\n", v.Name, n)
			fmt.Fprintf(&b, "    %s** %s = malloc(sizeof(%s*) * (%s_size + 1));\n", elem, v.Name, elem, v.Name)
			fmt.Fprintf(&b, "    for (int i = 0; i < %s_size; i++) {\n", v.Name)
			fmt.Fprintf(&b, "        %s[i] = malloc(sizeof(%s) * %d);\n", v.Name, elem, t.Width)
			fmt.Fprintf(&b, "        for (int j = 0; j < %d; j++) %s[i][j] = %s;\n    }\n", t.Width, v.Name, read(t.Base))
		}
	}
	var args []string
	for _, p := range s.params {
		args = append(args, p.Name)
		if p.Type.Dims > 0 {
			args = append(args, p.Name+"_size")
		}
	}
	retType := cScalar(s.ret.Base)
	if s.ret.Dims > 0 {
		b.WriteString("    int answer_size = 0;\n")
		args = append(args, "&answer_size")
		retType += strings.Repeat("*", s.ret.Dims)
	}
	fmt.Fprintf(&b, "    %s answer = %s(%s);\n", retType, s.name, strings.Join(args, ", "))
	f := cFormat(s.ret.Base)
	switch {
	case s.ret.Dims == 0 && s.ret.Base == "bool":
		b.WriteString(`    puts(answer ? "true" : "false");` + "\n")
	case s.ret.Dims == 0:
		fmt.Fprintf(&b, "    printf(\"%s\\n\", answer);\n", f)
	case s.ret.Dims == 1:
		sep := " "
		if s.output == "lines" {
			sep = "\\n"
		}
		fmt.Fprintf(&b, "    for (int i = 0; i < answer_size; i++) printf(i ? \"%s%s\" : \"%s\", answer[i]);\n", sep, f, f)
		b.WriteString(`    printf("\n");` + "\n")
	default:
		if s.output == "count" {
			b.WriteString(`    printf("%d\n", answer_size);` + "\n")
		}
		b.WriteString("    for (int i = 0; i < answer_size; i++) {\n")
		fmt.Fprintf(&b, "        for (int j = 0; j < %d; j++) printf(j ? \" %s\" : \"%s\", answer[i][j]);\n", s.ret.Width, f, f)
		b.WriteString(`        printf("\n");` + "\n    }\n")
	}
	b.WriteString("    return 0;\n}\n")
	return b.String()
}
