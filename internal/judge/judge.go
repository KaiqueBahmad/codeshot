// Package judge runs a solution against tests, inside docker, and says how
// each test went.
//
// The solution is built once, in a container of its language's image that has
// room to compile. Then a second container is started with the problem's
// limits — its memory, one CPU, a cap on processes and no network — and every
// test is run in it with docker exec, the time taken measured inside, and any
// process the memory limit killed counted from the container's cgroup.
package judge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"codeshot/internal/lang"
	"codeshot/internal/problem"
)

// Verdicts, for a test and for a whole run.
const (
	Accepted          = "AC"
	WrongAnswer       = "WA"
	TimeLimitExceeded = "TLE"
	MemoryLimit       = "MLE"
	RuntimeError      = "RE"
	CompilationError  = "CE"
)

// Names spells each verdict out.
var Names = map[string]string{
	Accepted:          "Accepted",
	WrongAnswer:       "Wrong Answer",
	TimeLimitExceeded: "Time Limit Exceeded",
	MemoryLimit:       "Memory Limit Exceeded",
	RuntimeError:      "Runtime Error",
	CompilationError:  "Compilation Error",
}

// Limits is what a solution may take on one test.
type Limits struct {
	TimeMS   int
	MemoryMB int
}

// Result is how one test went.
type Result struct {
	Test     string
	Status   string
	TimeMS   int64
	ExitCode int
	Input    string
	Expected string
	Stdout   string
	Stderr   string
}

// Outcome is how a whole run went.
type Outcome struct {
	// Verdict is the status of the first test that was not accepted, or
	// Accepted when they all were.
	Verdict       string
	CompileOutput string
	// TimeMS is the longest any test took.
	TimeMS  int64
	Results []Result
}

// Options tunes a run.
type Options struct {
	// StopAtFailure leaves the tests after the first one that fails unrun,
	// as a submission does.
	StopAtFailure bool
	// Log is told what is going on before a test runs, such as an image
	// being pulled. It may be nil.
	Log func(string)
	// OnResult is told how each test went as soon as it is known. It may be
	// nil.
	OnResult func(Result)
}

// outputCap is the most a solution's output is read up to; past it, the
// output is wrong anyway.
const outputCap = 64 << 20

// buildTimeout is how long building may take.
const buildTimeout = 2 * time.Minute

// Judge runs code, written in l, against tests within limits.
func Judge(ctx context.Context, l lang.Lang, code string, tests []problem.Test, limits Limits, opt Options) (Outcome, error) {
	log := opt.Log
	if log == nil {
		log = func(string) {}
	}
	if err := ensureImage(ctx, l.Image, log); err != nil {
		return Outcome{}, err
	}

	tmp, err := os.MkdirTemp("", "codeshot-judge-")
	if err != nil {
		return Outcome{}, err
	}
	defer os.RemoveAll(tmp)
	src, build := filepath.Join(tmp, "src"), filepath.Join(tmp, "build")
	for _, d := range []string{src, build} {
		if err := os.Mkdir(d, 0o755); err != nil {
			return Outcome{}, err
		}
	}
	if err := os.WriteFile(filepath.Join(src, l.File), []byte(code), 0o644); err != nil {
		return Outcome{}, err
	}

	// Both containers run as whoever runs codeshot, so that what they write
	// in build can be cleaned up afterwards.
	user := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
	common := []string{
		"--network", "none", "--user", user, "-e", "HOME=/tmp",
		"--tmpfs", "/tmp:exec,size=256m", "-v", src + ":/src:ro",
	}

	if l.Build != "" {
		log("building")
		bctx, cancel := context.WithTimeout(ctx, buildTimeout)
		args := append([]string{"run", "--rm", "--memory", "2g", "--pids-limit", "512"}, common...)
		args = append(args, "-v", build+":/build", l.Image, "sh", "-c", l.Build)
		out, err := exec.CommandContext(bctx, "docker", args...).CombinedOutput()
		cancel()
		if ctx.Err() != nil {
			return Outcome{}, ctx.Err()
		}
		if err != nil {
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				return Outcome{}, fmt.Errorf("building: %w", err)
			}
			if bctx.Err() != nil {
				out = append(out, "\nbuilding took longer than "+buildTimeout.String()...)
			}
			return Outcome{Verdict: CompilationError, CompileOutput: string(out)}, nil
		}
	}

	mem := strconv.Itoa(limits.MemoryMB) + "m"
	args := append([]string{"run", "-d", "--rm", "--memory", mem, "--memory-swap", mem, "--cpus", "1", "--pids-limit", "128"}, common...)
	args = append(args, "-v", build+":/build:ro", "--entrypoint", "sleep", l.Image, "infinity")
	out, err := exec.CommandContext(ctx, "docker", args...).Output()
	if err != nil {
		return Outcome{}, fmt.Errorf("starting the container to run in: %w", describe(err))
	}
	container := strings.TrimSpace(string(out))
	defer exec.Command("docker", "rm", "-f", container).Run()

	outcome := Outcome{Verdict: Accepted}
	for _, t := range tests {
		r, err := runTest(ctx, container, l.Run, t, limits)
		if err != nil {
			return outcome, fmt.Errorf("test %s: %w", t.Name, err)
		}
		outcome.Results = append(outcome.Results, r)
		outcome.TimeMS = max(outcome.TimeMS, r.TimeMS)
		if opt.OnResult != nil {
			opt.OnResult(r)
		}
		if r.Status != Accepted {
			if outcome.Verdict == Accepted {
				outcome.Verdict = r.Status
			}
			if opt.StopAtFailure {
				break
			}
		}
	}
	return outcome, nil
}

// marker starts the line runTest's script ends its stderr with: the exit
// code, the milliseconds taken, and the cgroup's count of processes killed
// for memory before and after.
const marker = "@@codeshot "

// script runs a solution with run, giving up a second past the time limit,
// and reports how it went behind marker.
const script = `oom() { while read k v; do [ "$k" = oom_kill ] && echo "$v"; done </sys/fs/cgroup/memory.events 2>/dev/null; }
o=$(oom); s=$(date +%%s%%N)
timeout -k 1 %s %s
c=$?; e=$(date +%%s%%N); p=$(oom)
printf '\n` + marker + `%%d %%d %%s %%s\n' "$c" $(( (e - s) / 1000000 )) "${o:-0}" "${p:-0}" >&2
`

func runTest(ctx context.Context, container, run string, t problem.Test, limits Limits) (Result, error) {
	r := Result{Test: t.Name, Input: t.Input, Expected: t.Output}
	wait := strconv.FormatFloat(float64(limits.TimeMS)/1000+1, 'f', 3, 64)
	cmd := exec.CommandContext(ctx, "docker", "exec", "-i", container, "sh", "-c", fmt.Sprintf(script, wait, run))
	cmd.Stdin = strings.NewReader(t.Input)
	var stdout, stderr capped
	stdout.limit, stderr.limit = outputCap, 1<<20
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		return r, ctx.Err()
	}

	errText := stderr.String()
	i := strings.LastIndex(errText, "\n"+marker)
	if i < 0 {
		return r, fmt.Errorf("the run did not report back: %v\n%s", describe(err), errText)
	}
	var oomBefore, oomAfter int
	fmt.Sscanf(errText[i+1+len(marker):], "%d %d %d %d", &r.ExitCode, &r.TimeMS, &oomBefore, &oomAfter)
	r.Stderr = errText[:i]
	r.Stdout = stdout.String()

	switch {
	case oomAfter > oomBefore:
		r.Status = MemoryLimit
	case r.TimeMS > int64(limits.TimeMS) || r.ExitCode == 124:
		r.Status = TimeLimitExceeded
	case r.ExitCode != 0:
		r.Status = RuntimeError
	case stdout.over || !Same(r.Stdout, t.Output):
		r.Status = WrongAnswer
	default:
		r.Status = Accepted
	}
	return r, nil
}

// Same says whether got is the output wanted. Spaces at the end of a line,
// and empty lines at the end, do not count.
func Same(got, want string) bool {
	return normalize(got) == normalize(want)
}

func normalize(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// ensureImage pulls image unless docker already has it.
func ensureImage(ctx context.Context, image string, log func(string)) error {
	if _, err := exec.LookPath("docker"); err != nil {
		return errors.New("docker is not installed, and the judge runs solutions in it")
	}
	if exec.CommandContext(ctx, "docker", "image", "inspect", image).Run() == nil {
		return nil
	}
	log("pulling " + image + ", only this once")
	if out, err := exec.CommandContext(ctx, "docker", "pull", "--quiet", image).CombinedOutput(); err != nil {
		return fmt.Errorf("pulling %s: %w\n%s", image, err, out)
	}
	return nil
}

// describe adds what a failed command said to its error.
func describe(err error) error {
	var exit *exec.ExitError
	if errors.As(err, &exit) && len(exit.Stderr) > 0 {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exit.Stderr)))
	}
	return err
}

// capped keeps what is written to it up to limit bytes, and notes whether
// there was more.
type capped struct {
	bytes.Buffer
	limit int
	over  bool
}

func (c *capped) Write(p []byte) (int, error) {
	if room := c.limit - c.Len(); len(p) > room {
		c.over = true
		if room > 0 {
			c.Buffer.Write(p[:room])
		}
		return len(p), nil
	}
	return c.Buffer.Write(p)
}
