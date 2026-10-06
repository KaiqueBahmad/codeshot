package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"

	"codeshot/internal/judge"
	"codeshot/internal/lang"
	"codeshot/internal/problem"
	"codeshot/internal/workspace"
)

func init() {
	register(command{name: "test", run: testCmd})
}

// testCmd runs an attempt's solution against the tests in its folder: the
// samples it came with, and any that were added next to them. Nothing is
// recorded.
func testCmd(args []string) int {
	dir, spec, l, code, err := loadAttempt(args)
	if err != nil {
		return report(err)
	}
	tests, err := folderTests(filepath.Join(dir, "tests"))
	if err != nil {
		return report(err)
	}
	if len(tests) == 0 {
		return report(fmt.Errorf("no tests in %s", filepath.Join(dir, "tests")))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	color := isTerminal(os.Stdout)
	limits := judge.Limits{TimeMS: spec.TimeLimitMS, MemoryMB: spec.MemoryMB}
	outcome, err := judge.Judge(ctx, l, code, tests, limits, judge.Options{
		Log:      func(s string) { fmt.Fprintln(os.Stderr, paintIf(color, gray, s+"…")) },
		OnResult: func(r judge.Result) { printResult(os.Stdout, r, true, color) },
	})
	if err != nil {
		return report(err)
	}
	printOutcome(os.Stdout, outcome, color)
	if outcome.Verdict != judge.Accepted {
		return 1
	}
	return 0
}

// loadAttempt finds the attempt folder args names, or the one the current
// directory is in, and reads its solution.
func loadAttempt(args []string) (string, workspace.Spec, lang.Lang, string, error) {
	where := "."
	if len(args) > 0 {
		where = args[0]
	}
	dir, spec, err := workspace.Find(where)
	if err != nil {
		return "", spec, lang.Lang{}, "", err
	}
	l, err := lang.ByID(spec.Lang)
	if err != nil {
		return "", spec, l, "", err
	}
	code, err := os.ReadFile(filepath.Join(dir, l.File))
	if err != nil {
		return "", spec, l, "", err
	}
	return dir, spec, l, string(code), nil
}

// folderTests reads the NN.in and NN.out pairs in dir. An input with no
// output next to it is left out.
func folderTests(dir string) ([]problem.Test, error) {
	ins, err := filepath.Glob(filepath.Join(dir, "*.in"))
	if err != nil {
		return nil, err
	}
	sort.Strings(ins)
	var tests []problem.Test
	for _, in := range ins {
		input, err := os.ReadFile(in)
		if err != nil {
			return nil, err
		}
		output, err := os.ReadFile(strings.TrimSuffix(in, ".in") + ".out")
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(in), ".in")
		tests = append(tests, problem.Test{Name: name, Input: string(input), Output: string(output), Sample: true})
	}
	return tests, nil
}

var verdictColor = map[string]string{
	judge.Accepted:          green,
	judge.WrongAnswer:       red,
	judge.TimeLimitExceeded: yellow,
	judge.MemoryLimit:       yellow,
	judge.RuntimeError:      red,
	judge.CompilationError:  red,
}

func paintIf(color bool, c, s string) string {
	if !color {
		return s
	}
	return c + s + reset
}

// printResult prints how one test went. A failed one shows what it was given
// and what came out, when details is set.
func printResult(w io.Writer, r judge.Result, details, color bool) {
	fmt.Fprintf(w, "  %s  %s  %s\n", r.Test,
		paintIf(color, verdictColor[r.Status], fmt.Sprintf("%-3s", r.Status)),
		paintIf(color, gray, fmt.Sprintf("%d ms", r.TimeMS)))
	if !details || r.Status == judge.Accepted {
		return
	}
	section := func(title, text string) {
		if text == "" {
			return
		}
		fmt.Fprintf(w, "    %s\n%s\n", paintIf(color, gray, title), indent(clip(text, 12, 200), "      "))
	}
	section("input", r.Input)
	if r.Status == judge.WrongAnswer {
		section("expected", r.Expected)
		got := r.Stdout
		if strings.TrimSpace(got) == "" {
			got = "(nothing)"
		}
		section("got", got)
	}
	if r.Status == judge.RuntimeError {
		section(fmt.Sprintf("exit code %d", r.ExitCode), r.Stderr)
	} else {
		section("stderr", r.Stderr)
	}
}

// printOutcome prints the verdict of a whole run.
func printOutcome(w io.Writer, o judge.Outcome, color bool) {
	if o.Verdict == judge.CompilationError {
		fmt.Fprintln(w, indent(clip(o.CompileOutput, 40, 300), "  "))
	}
	passed := 0
	for _, r := range o.Results {
		if r.Status == judge.Accepted {
			passed++
		}
	}
	verdict := fmt.Sprintf(" %s ", strings.ToUpper(judge.Names[o.Verdict]))
	if color {
		verdict = "\x1b[1;30;" + map[string]string{green: "42", yellow: "43", red: "41"}[verdictColor[o.Verdict]] + "m" + verdict + reset
	}
	fmt.Fprintf(w, "\n%s  %d/%d tests passed, %d ms at most\n", verdict, passed, len(o.Results), o.TimeMS)
}

// clip keeps the first lines of s, each cut to width, and says how much was
// left out.
func clip(s string, lines, width int) string {
	all := strings.Split(strings.TrimRight(s, "\n"), "\n")
	var kept []string
	for i, l := range all {
		if i == lines {
			kept = append(kept, fmt.Sprintf("… %d more lines", len(all)-lines))
			break
		}
		if len(l) > width {
			l = l[:width] + fmt.Sprintf("… (%d more characters)", len(l)-width)
		}
		kept = append(kept, l)
	}
	return strings.Join(kept, "\n")
}

func indent(s, prefix string) string {
	return prefix + strings.ReplaceAll(s, "\n", "\n"+prefix)
}
