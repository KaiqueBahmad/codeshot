package stub

import (
	"fmt"
	"strings"
)

func javaType(t Type) string {
	base := map[string]string{"int": "int", "long": "long", "word": "String", "line": "String", "string": "String", "bool": "boolean"}[t.Base]
	return base + strings.Repeat("[]", t.Dims)
}

func genJava(s spec) string {
	name := camel(s.name)
	var params []string
	for _, p := range s.params {
		params = append(params, javaType(p.Type)+" "+camel(p.Name))
	}
	zero := map[string]string{"int": "0", "long": "0", "bool": "false", "string": `""`}[s.ret.Base]
	switch s.ret.Dims {
	case 1:
		zero = "new " + strings.TrimSuffix(javaType(s.ret), "[]") + "[0]"
	case 2:
		zero = "new " + strings.TrimSuffix(javaType(s.ret), "[][]") + "[0][]"
	}

	var b strings.Builder
	b.WriteString("import java.io.*;\nimport java.nio.charset.StandardCharsets;\nimport java.util.*;\n\n")
	fmt.Fprintf(&b, "class Solution {\n    public %s %s(%s) {\n        // Write your solution here.\n        return %s;\n    }\n}\n\n",
		javaType(s.ret), name, strings.Join(params, ", "), zero)
	b.WriteString(banner("//", name) + "\n")
	b.WriteString(`class CodeshotInput {
    private final byte[] buf;
    private int pos;
    private boolean mid;

    CodeshotInput() throws IOException { buf = System.in.readAllBytes(); }

    private static boolean space(byte c) { return c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '\f' || c == 11; }

    private void skip() { while (pos < buf.length && space(buf[pos])) pos++; }

    long num() {
        skip();
        boolean neg = pos < buf.length && buf[pos] == '-';
        if (pos < buf.length && (buf[pos] == '-' || buf[pos] == '+')) pos++;
        long v = 0;
        while (pos < buf.length && buf[pos] >= '0' && buf[pos] <= '9') v = v * 10 + (buf[pos++] - '0');
        mid = true;
        return neg ? -v : v;
    }

    String word() {
        skip();
        int start = pos;
        while (pos < buf.length && !space(buf[pos])) pos++;
        mid = true;
        return new String(buf, start, pos - start, StandardCharsets.UTF_8);
    }

    // The line after the one the last token was read from.
    String line() {
        if (mid) {
            while (pos < buf.length && buf[pos] != '\n') pos++;
            if (pos < buf.length) pos++;
            mid = false;
        }
        int start = pos;
        while (pos < buf.length && buf[pos] != '\n') pos++;
        int end = pos;
        if (end > start && buf[end - 1] == '\r') end--;
        if (pos < buf.length) pos++;
        return new String(buf, start, end - start, StandardCharsets.UTF_8);
    }
}

public class Main {
    public static void main(String[] args) throws IOException {
        CodeshotInput codeshotIn = new CodeshotInput();
`)
	read := func(base string) string {
		return map[string]string{"int": "(int) codeshotIn.num()", "long": "codeshotIn.num()", "word": "codeshotIn.word()", "line": "codeshotIn.line()"}[base]
	}
	for _, v := range s.input {
		t := v.Type
		n := camel(v.Name)
		elem := javaType(Type{Base: t.Base})
		switch t.Dims {
		case 0:
			fmt.Fprintf(&b, "        %s %s = %s;\n", elem, n, read(t.Base))
		case 1:
			fmt.Fprintf(&b, "        %s[] %s = new %s[(int) %s];\n", elem, n, elem, lenExpr(t, camel))
			fmt.Fprintf(&b, "        for (int i = 0; i < %s.length; i++) %s[i] = %s;\n", n, n, read(t.Base))
		case 2:
			fmt.Fprintf(&b, "        %s[][] %s = new %s[(int) %s][%d];\n", elem, n, elem, lenExpr(t, camel), t.Width)
			fmt.Fprintf(&b, "        for (%s[] row : %s) for (int j = 0; j < row.length; j++) row[j] = %s;\n", elem, n, read(t.Base))
		}
	}
	var args []string
	for _, p := range s.params {
		args = append(args, camel(p.Name))
	}
	fmt.Fprintf(&b, "        %s answer = new Solution().%s(%s);\n", javaType(s.ret), name, strings.Join(args, ", "))
	b.WriteString("        StringBuilder out = new StringBuilder();\n")
	switch {
	case s.ret.Dims == 0:
		b.WriteString("        out.append(answer).append('\\n');\n")
	case s.ret.Dims == 1:
		sep := "' '"
		if s.output == "lines" {
			sep = "'\\n'"
		}
		fmt.Fprintf(&b, "        for (int i = 0; i < answer.length; i++) {\n            if (i > 0) out.append(%s);\n            out.append(answer[i]);\n        }\n        out.append('\\n');\n", sep)
	default:
		if s.output == "count" {
			b.WriteString("        out.append(answer.length).append('\\n');\n")
		}
		b.WriteString("        for (var row : answer) {\n            for (int j = 0; j < row.length; j++) {\n                if (j > 0) out.append(' ');\n                out.append(row[j]);\n            }\n            out.append('\\n');\n        }\n")
	}
	b.WriteString("        System.out.print(out);\n    }\n}\n")
	return b.String()
}
