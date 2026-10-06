// Package lang is every language a solution can be written in: the file it
// goes in, what it starts as, and how the judge builds and runs it.
package lang

import (
	"fmt"
	"strings"
)

// Lang is one language. Build and Run are shell commands run inside the
// language's docker image, where the solution is at /src/<File>, read-only,
// and /build is somewhere to put what Build makes. An empty Build means
// there is nothing to build.
type Lang struct {
	ID       string
	Name     string
	File     string
	Image    string
	Build    string
	Run      string
	Template string
}

// All lists the languages, in the order they are offered.
var All = []Lang{
	{
		ID: "c", Name: "C", File: "main.c", Image: "gcc:14",
		Build: "gcc -O2 -std=c17 -o /build/main /src/main.c -lm",
		Run:   "/build/main",
		Template: `#include <stdio.h>

int main(void) {
    // Read the input from stdin, and print the answer to stdout.
    return 0;
}
`,
	},
	{
		ID: "cpp", Name: "C++", File: "main.cpp", Image: "gcc:14",
		Build: "g++ -O2 -std=c++20 -o /build/main /src/main.cpp",
		Run:   "/build/main",
		Template: `#include <bits/stdc++.h>
using namespace std;

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);
    // Read the input from cin, and print the answer to cout.
    return 0;
}
`,
	},
	{
		ID: "java", Name: "Java", File: "Main.java", Image: "eclipse-temurin:21-jdk",
		Build: "javac -d /build /src/Main.java",
		Run:   "java -XX:+UseSerialGC -Xss64m -cp /build Main",
		Template: `import java.io.*;
import java.util.*;

public class Main {
    public static void main(String[] args) throws IOException {
        BufferedReader in = new BufferedReader(new InputStreamReader(System.in));
        // Read the input from in, and print the answer to System.out.
    }
}
`,
	},
	{
		ID: "python", Name: "Python", File: "main.py", Image: "python:3.13-slim",
		// Parsing it finds what would not compile, without writing a .pyc
		// next to the read-only source.
		Build: `python3 -c "import ast; ast.parse(open('/src/main.py').read(), 'main.py')"`,
		Run:   "python3 /src/main.py",
		Template: `import sys


def main():
    data = sys.stdin.read().split()
    # Read the input from data, and print the answer.


main()
`,
	},
	{
		ID: "go", Name: "Go", File: "main.go", Image: "golang:1.26",
		Build: "GOCACHE=/build/.cache GO111MODULE=off go build -o /build/main /src/main.go",
		Run:   "/build/main",
		Template: `package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	// Read the input from in, and print the answer to out.
	_, _ = in, fmt.Fprint
}
`,
	},
	{
		ID: "rust", Name: "Rust", File: "main.rs", Image: "rust:1-slim",
		Build: "rustc -O --edition 2021 -o /build/main /src/main.rs",
		Run:   "/build/main",
		Template: `use std::io::{self, Read, Write};

fn main() {
    let mut input = String::new();
    io::stdin().read_to_string(&mut input).unwrap();
    let mut out = io::BufWriter::new(io::stdout().lock());
    // Read the input from input, and print the answer to out.
    let _ = &mut out;
    out.flush().unwrap();
}
`,
	},
	{
		ID: "javascript", Name: "JavaScript", File: "main.js", Image: "node:22-slim",
		Build: "node --check /src/main.js",
		Run:   "node --stack-size=65500 /src/main.js",
		Template: `const data = require("fs").readFileSync(0, "utf8").split(/\s+/).filter(Boolean);
// Read the input from data, and print the answer with console.log.
`,
	},
}

// ByID looks up a language by its id or its name, in any case.
func ByID(id string) (Lang, error) {
	for _, l := range All {
		if strings.EqualFold(l.ID, id) || strings.EqualFold(l.Name, id) {
			return l, nil
		}
	}
	ids := make([]string, len(All))
	for i, l := range All {
		ids[i] = l.ID
	}
	return Lang{}, fmt.Errorf("no language %q: pick one of %s", id, strings.Join(ids, ", "))
}
