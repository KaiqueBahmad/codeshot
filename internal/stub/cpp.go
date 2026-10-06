package stub

import (
	"fmt"
	"strings"
)

func cppType(t Type) string {
	base := map[string]string{"int": "int", "long": "long long", "word": "string", "line": "string", "string": "string", "bool": "bool"}[t.Base]
	for range t.Dims {
		base = "vector<" + base + ">"
	}
	return base
}

func genCpp(s spec) string {
	name := camel(s.name)
	var params []string
	for _, p := range s.params {
		typ := cppType(p.Type)
		if p.Type.Dims > 0 {
			typ += "&"
		}
		params = append(params, typ+" "+camel(p.Name))
	}
	zero := map[string]string{"int": "0", "long": "0", "bool": "false", "string": `""`}[s.ret.Base]
	if s.ret.Dims > 0 {
		zero = "{}"
	}

	var b strings.Builder
	b.WriteString("#include <bits/stdc++.h>\nusing namespace std;\n\n")
	fmt.Fprintf(&b, "%s %s(%s) {\n    // Write your solution here.\n    return %s;\n}\n\n", cppType(s.ret), name, strings.Join(params, ", "), zero)
	b.WriteString(banner("//", name) + "\n")
	b.WriteString(`struct CodeshotInput {
    string buf;
    size_t pos = 0;
    bool mid = false;

    CodeshotInput() { buf.assign(istreambuf_iterator<char>(cin), istreambuf_iterator<char>()); }

    void skip() {
        while (pos < buf.size() && isspace((unsigned char)buf[pos])) pos++;
    }

    long long num() {
        skip();
        bool neg = pos < buf.size() && buf[pos] == '-';
        if (pos < buf.size() && (buf[pos] == '-' || buf[pos] == '+')) pos++;
        long long v = 0;
        while (pos < buf.size() && isdigit((unsigned char)buf[pos])) v = v * 10 + (buf[pos++] - '0');
        mid = true;
        return neg ? -v : v;
    }

    string word() {
        skip();
        size_t start = pos;
        while (pos < buf.size() && !isspace((unsigned char)buf[pos])) pos++;
        mid = true;
        return buf.substr(start, pos - start);
    }

    // The line after the one the last token was read from.
    string line() {
        if (mid) {
            while (pos < buf.size() && buf[pos] != '\n') pos++;
            if (pos < buf.size()) pos++;
            mid = false;
        }
        size_t start = pos;
        while (pos < buf.size() && buf[pos] != '\n') pos++;
        size_t end = pos;
        if (end > start && buf[end - 1] == '\r') end--;
        if (pos < buf.size()) pos++;
        return buf.substr(start, end - start);
    }
};

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);
    CodeshotInput codeshotIn;
`)
	read := func(base string) string {
		return map[string]string{"int": "(int)codeshotIn.num()", "long": "codeshotIn.num()", "word": "codeshotIn.word()", "line": "codeshotIn.line()"}[base]
	}
	for _, v := range s.input {
		t := v.Type
		n := camel(v.Name)
		switch t.Dims {
		case 0:
			fmt.Fprintf(&b, "    %s %s = %s;\n", cppType(t), n, read(t.Base))
		case 1:
			fmt.Fprintf(&b, "    %s %s((size_t)%s);\n", cppType(t), n, lenExpr(t, camel))
			fmt.Fprintf(&b, "    for (auto& x : %s) x = %s;\n", n, read(t.Base))
		case 2:
			fmt.Fprintf(&b, "    %s %s((size_t)%s, %s(%d));\n", cppType(t), n, lenExpr(t, camel), cppType(Type{Base: t.Base, Dims: 1}), t.Width)
			fmt.Fprintf(&b, "    for (auto& row : %s) for (auto& x : row) x = %s;\n", n, read(t.Base))
		}
	}
	var args []string
	for _, p := range s.params {
		args = append(args, camel(p.Name))
	}
	fmt.Fprintf(&b, "    auto answer = %s(%s);\n", name, strings.Join(args, ", "))
	switch {
	case s.ret.Dims == 0 && s.ret.Base == "bool":
		b.WriteString(`    cout << (answer ? "true" : "false") << '\n';` + "\n")
	case s.ret.Dims == 0:
		b.WriteString("    cout << answer << '\\n';\n")
	case s.ret.Dims == 1:
		sep := "' '"
		if s.output == "lines" {
			sep = "'\\n'"
		}
		fmt.Fprintf(&b, "    for (size_t i = 0; i < answer.size(); i++) {\n        if (i) cout << %s;\n        cout << answer[i];\n    }\n", sep)
		b.WriteString("    cout << '\\n';\n")
	default:
		if s.output == "count" {
			b.WriteString("    cout << answer.size() << '\\n';\n")
		}
		b.WriteString("    for (auto& row : answer) {\n        for (size_t j = 0; j < row.size(); j++) {\n            if (j) cout << ' ';\n            cout << row[j];\n        }\n        cout << '\\n';\n    }\n")
	}
	b.WriteString("    return 0;\n}\n")
	return b.String()
}
