// Package stub writes the code a solution starts from when its problem gives
// the function it is solved in, as LeetCode does: a function to fill in, and
// under it a main, not to be touched, that reads the input, calls the
// function and prints what it returns. Solutions are still judged on what
// they print, so a problem without a function works as it always has.
//
// A problem gives its function in meta.json:
//
//	"function": {
//	  "name": "two_sum",
//	  "input": ["n: int", "target: int", "nums: int[n]"],
//	  "params": ["nums", "target"],
//	  "returns": "int[]",
//	  "output": "space"
//	}
//
// input is what stdin holds, in order. params is what the function is given,
// out of it, and returns is what it gives back. The types are:
//
//	int, long      a number, 32 or 64 bits wide
//	word           a run of characters without spaces
//	line           a whole line, spaces and all
//	T[n]           n of int, long or word, n being a number read before it
//	T[n][k]        n rows of k int or long
//
// A function returns an int, a long, a bool, a string, an int[], long[] or
// string[], or rows of k int or long, as int[][k]. A list is printed on one
// line, space apart, or with output "lines" one to a line; rows are printed
// one to a line, and with output "count" behind a line with how many there are.
package stub

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Function is the function a problem is solved in, as meta.json gives it.
type Function struct {
	Name    string   `json:"name"`
	Input   []string `json:"input"`
	Params  []string `json:"params"`
	Returns string   `json:"returns"`
	Output  string   `json:"output,omitempty"`
}

// Type is the type of a value read or returned.
type Type struct {
	Base  string // int, long, word, line, bool or string
	Dims  int    // 0 for one value, 1 for a list, 2 for rows
	Len   string // how many there are, for a list or rows read: a number or a name
	Width int    // how many values each row has
}

// Var is a value read from the input.
type Var struct {
	Name string
	Type Type
}

// spec is a Function, checked and taken apart.
type spec struct {
	name   string
	input  []Var
	params []Var
	ret    Type
	output string
}

var (
	identifier = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	typeSyntax = regexp.MustCompile(`^(int|long|word|line|bool|string)((?:\[[a-z0-9_]*\]){0,2})$`)
	dimSyntax  = regexp.MustCompile(`\[([a-z0-9_]*)\]`)
)

// Check says what is wrong with f, if anything.
func (f Function) Check() error {
	_, err := f.parse()
	return err
}

func (f Function) parse() (spec, error) {
	s := spec{name: f.Name, output: f.Output}
	if !identifier.MatchString(f.Name) {
		return s, fmt.Errorf("function name %q is not a snake_case name", f.Name)
	}
	known := map[string]Var{}
	for _, in := range f.Input {
		name, typ, ok := strings.Cut(in, ":")
		name, typ = strings.TrimSpace(name), strings.TrimSpace(typ)
		if !ok || !identifier.MatchString(name) {
			return s, fmt.Errorf("input %q is not name: type", in)
		}
		if _, dup := known[name]; dup {
			return s, fmt.Errorf("input %s is read twice", name)
		}
		t, err := parseType(typ)
		if err != nil {
			return s, fmt.Errorf("input %s: %w", name, err)
		}
		if err := checkInput(t, known); err != nil {
			return s, fmt.Errorf("input %s: %w", name, err)
		}
		v := Var{Name: name, Type: t}
		known[name] = v
		s.input = append(s.input, v)
	}
	for _, p := range f.Params {
		v, ok := known[p]
		if !ok {
			return s, fmt.Errorf("param %s is not one of the inputs", p)
		}
		s.params = append(s.params, v)
	}
	ret, err := parseType(f.Returns)
	if err != nil {
		return s, fmt.Errorf("returns: %w", err)
	}
	if err := checkReturn(ret); err != nil {
		return s, fmt.Errorf("returns: %w", err)
	}
	s.ret = ret
	switch {
	case ret.Dims == 1 && s.output == "":
		s.output = "space"
	case ret.Dims == 2 && s.output == "":
		s.output = "rows"
	case ret.Dims == 0 && s.output != "",
		ret.Dims == 1 && s.output != "space" && s.output != "lines",
		ret.Dims == 2 && s.output != "rows" && s.output != "count":
		return s, fmt.Errorf("output %q does not go with returns %s", s.output, f.Returns)
	}
	return s, nil
}

func parseType(s string) (Type, error) {
	m := typeSyntax.FindStringSubmatch(s)
	if m == nil {
		return Type{}, fmt.Errorf("%q is not a type", s)
	}
	t := Type{Base: m[1]}
	dims := dimSyntax.FindAllStringSubmatch(m[2], -1)
	t.Dims = len(dims)
	if t.Dims >= 1 {
		t.Len = dims[0][1]
	}
	if t.Dims == 2 {
		w, err := strconv.Atoi(dims[1][1])
		if err != nil || w <= 0 {
			return t, fmt.Errorf("%q: the width of a row must be a number", s)
		}
		t.Width = w
	}
	return t, nil
}

func checkInput(t Type, known map[string]Var) error {
	switch {
	case t.Base == "bool" || t.Base == "string":
		return fmt.Errorf("%s is not read from the input: use word or line", t.Base)
	case t.Base == "line" && t.Dims > 0:
		return fmt.Errorf("lines cannot be read as a list")
	case t.Base == "word" && t.Dims > 1:
		return fmt.Errorf("words cannot be read as rows")
	}
	if t.Dims == 0 {
		return nil
	}
	if t.Len == "" {
		return fmt.Errorf("a list read needs its length, as in int[n]")
	}
	if n, err := strconv.Atoi(t.Len); err == nil {
		if n < 0 {
			return fmt.Errorf("a length cannot be negative")
		}
		return nil
	}
	v, ok := known[t.Len]
	if !ok || v.Type.Dims != 0 || (v.Type.Base != "int" && v.Type.Base != "long") {
		return fmt.Errorf("the length %s is not a number read before it", t.Len)
	}
	return nil
}

func checkReturn(t Type) error {
	switch {
	case t.Base == "word" || t.Base == "line":
		return fmt.Errorf("a function returns a string, not a %s", t.Base)
	case t.Len != "":
		return fmt.Errorf("a list returned has no length in its type: write %s[]", t.Base)
	case t.Dims > 0 && t.Base == "bool":
		return fmt.Errorf("lists of bool are not supported")
	case t.Dims == 2 && t.Base == "string":
		return fmt.Errorf("rows of string are not supported")
	}
	return nil
}

// Generate writes the code a solution in the language lang starts from.
func Generate(lang string, f Function) (string, error) {
	s, err := f.parse()
	if err != nil {
		return "", err
	}
	gen, ok := generators[lang]
	if !ok {
		return "", fmt.Errorf("no stub for language %q", lang)
	}
	return gen(s), nil
}

var generators = map[string]func(spec) string{
	"c":          genC,
	"cpp":        genCpp,
	"java":       genJava,
	"python":     genPython,
	"go":         genGo,
	"rust":       genRust,
	"javascript": genJS,
}

// camel writes a snake_case name in camelCase.
func camel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if parts[i] != "" {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// unpassed is the inputs that are read but not given to the function.
func (s spec) unpassed() []Var {
	var out []Var
	for _, v := range s.input {
		if !slices.ContainsFunc(s.params, func(p Var) bool { return p.Name == v.Name }) {
			out = append(out, v)
		}
	}
	return out
}

// lenExpr is the length of a list read, written with name for a variable.
func lenExpr(t Type, name func(string) string) string {
	if _, err := strconv.Atoi(t.Len); err == nil {
		return t.Len
	}
	return name(t.Len)
}

// banner is the line between the function and the code that runs it.
func banner(comment, fn string) string {
	return comment + " ---- Reads the input, calls " + fn + " and prints what it returns. No need to change it. ----"
}

// lines joins code lines, each indented by indent.
func block(indent string, lines ...string) string {
	var b strings.Builder
	for _, l := range lines {
		if l == "" {
			b.WriteString("\n")
			continue
		}
		b.WriteString(indent + l + "\n")
	}
	return b.String()
}
