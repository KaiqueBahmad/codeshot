# codeshot

LeetCode in the terminal, with everything kept on your machine. Browse a bank
of problems by tag and difficulty, solve one in your own editor, and submit it
to a judge that runs it in docker against hidden tests, with time and memory
limits. There is no account to log in to: everything lives in `~/.codeshot`.

## Installing

Download the Linux amd64 `.deb` or `.tar.gz` from
[Releases](https://github.com/kaiquebahmad/codeshot/releases).
You need docker and git to use codeshot.

Install the Debian package with `sudo apt install ./codeshot_0.1.0_amd64.deb`
(replace the version with the one downloaded). It includes bash and fish
completion. For the archive, extract it and install the executable:

```bash
tar -xzf codeshot_v0.1.0_linux_amd64.tar.gz
mkdir -p ~/.local/bin
install -m755 codeshot ~/.local/bin/codeshot
```

The release's `checksums.txt` lets you verify the downloads with
`sha256sum --ignore-missing -c checksums.txt`.

To build from source, you need Go as specified in `go.mod`:

```bash
go build -o ~/.local/bin/codeshot ./cmd/codeshot
codeshot sync
```

`codeshot sync` clones this repository into `~/.codeshot/bank` and imports
the problems in its `problems/` folder. To work from a checkout instead, run
`codeshot sync --from .`.

## Using it

```bash
codeshot                 # open the TUI
```

In the TUI, the tags are on the left and their problems on the right, each
marked ✓ solved, ~ tried or · not yet touched. `e`, `m` and `h` narrow the
list to one difficulty, and `/` searches it. Opening a problem shows its
statement next to its history: every attempt, with the submissions made from
it. Press `s` to solve it. codeshot asks which language you want, makes a
folder for the attempt, and asks how to open it:

```
$ lets go
> Open with NeoVim
  Open with Vim
  Open with VS Code
  Just tell me the folder
```

The folder has everything the attempt needs:

```
~/.codeshot/attempts/two-sum/7/
├── main.c        the solution, from the language's template
├── problem.md    the statement
├── spec.json     problem, language and limits
├── tests/        the sample tests, NN.in and NN.out
├── run.sh        runs the solution against tests/
└── submit.sh     submits it
```

`./run.sh` tries the solution against the samples, plus any `NN.in`/`NN.out`
pair you add to `tests/`. `./submit.sh` judges it against every test of the
problem, the hidden ones too, and records the submission, whatever the verdict.

Everything the TUI does can also be done from the command line:

| command | does |
|---|---|
| `codeshot sync [--from <git url \| dir>]` | import the problems of a bank |
| `codeshot list [--tag t] [--difficulty d] [--status s]` | print the problems |
| `codeshot solve <problem> --lang <language> [--editor <editor>]` | start a new attempt |
| `codeshot test [folder]` | run an attempt against the tests in its folder |
| `codeshot submit [folder]` | judge an attempt against every test, and record it |
| `codeshot history [problem \| submission id]` | list submissions, or print one in full |
| `codeshot restore <submission id>` | start a new attempt from an old submission's code |
| `codeshot clean [problem] [--keep-last n] [--force]` | delete attempt folders |

The languages are C, C++, Java, Python, Go, Rust and JavaScript.

## Where things are kept

`~/.codeshot/codeshot.db` is a SQLite database, and the source of truth: the
problems with their tests, every attempt, and every submission, with the code
that was submitted and how each test went. An attempt's folder is only where
the code is worked on. `codeshot clean` deletes those folders, and the history
stays. It keeps any folder whose code was never submitted, or has changed
since it was, unless you pass `--force`. `codeshot restore` brings a
submission's code back into a new folder.

Set `CODESHOT_HOME` to keep everything somewhere other than `~/.codeshot`.

## The judge

A submission is built once, in a container of its language's image, without
network. Each test then runs in a second container with the problem's limits:
its memory with no swap, one CPU, at most 128 processes, and no network. The
time is measured inside the container, and a process killed for going over
the memory limit is counted from the container's cgroup. The verdicts are
Accepted, Wrong Answer, Time Limit Exceeded, Memory Limit Exceeded, Runtime
Error and Compilation Error. Spaces at the end of a line, and empty lines at
the end, do not count when output is compared.

The first submission in a language pulls its image, which takes a while once.

## Developing

The commands for working on codeshot are in `runbook.yml`, for
[runbook](https://github.com/kaiquebahmad/runbook): `runbook list` shows them,
and `runbook run go/check` runs everything that must pass before a commit.

They build `bin/codeshot-dev`, which keeps its data in `~/.codeshot-dev`, so
trying it out never touches the history of an installed codeshot.
`runbook run app/sync` imports the problems of the checkout into it,
`runbook run app/start` opens its TUI, and `runbook run app/install-completion`
makes `codeshot-dev <Tab>` complete in bash, apart from `codeshot`'s own
completion. `runbook run app/link` links `bin/codeshot-dev` into `~/.local/bin`,
so every later build runs as `codeshot-dev` from anywhere, and
`runbook run app/unlink` takes the link away. `runbook run app/dev-install`
does the build, the completion and the link at once.

See [RELEASING.md](RELEASING.md) for the tag and changelog release process.

## Adding problems

A problem is a folder in `problems/`, named for its slug:

```
problems/two-sum/
├── problem.md     the statement, in markdown
├── meta.json      title, difficulty, tags, limits, samples
├── solution.py    a reference solution
├── templates/     optional: code to start from, by file name, such as main.c
└── tests/
    ├── 01.in
    └── 01.out
```

```json
{
  "title": "Two Sum",
  "difficulty": "easy",
  "tags": ["arrays", "hash-table"],
  "time_limit_ms": 1000,
  "memory_mb": 256,
  "samples": ["01", "02"],
  "function": {
    "name": "two_sum",
    "input": ["n: int", "target: int", "nums: int[n]"],
    "params": ["nums", "target"],
    "returns": "int[]"
  }
}
```

The difficulty is `easy`, `medium` or `hard`. The tests named in `samples` are
copied into every attempt's folder; the rest stay hidden. Solutions are judged
on what they print to stdout for what they read from stdin.

`function` is what a solution starts from, as on LeetCode: a function to fill
in, and under it a main that reads the input, calls the function and prints
what it returns, written out for every language. `input` is what stdin holds,
in order, `params` is what the function is given out of it, and `returns` is
what it gives back:

| type | is |
|---|---|
| `int`, `long` | a number, 32 or 64 bits wide |
| `word` | a run of characters without spaces |
| `line` | a whole line, spaces and all |
| `int[n]`, `long[n]`, `word[n]` | `n` of them, `n` being a number read before |
| `int[n][k]`, `long[n][k]` | `n` rows of `k` numbers |

A function returns an `int`, `long`, `bool` or `string`, a list of `int`,
`long` or `string` written as `int[]`, or rows of `k` numbers written as
`int[][k]`. A list is printed on one line, space apart, or one to a line with
`"output": "lines"`. Rows are printed one to a line, and with
`"output": "count"` behind a line saying how many there are. The function is
named in snake_case, and each language writes it its own way: `two_sum` in C,
Python and Rust, and `twoSum` in C++, Java, Go and JavaScript.

When a language needs something the function cannot say, put the code it
should start from in `templates/`, named as the language names its file, such
as `main.c` or `Main.java`. A problem without a function starts from a plain
program that reads stdin.

Write the inputs, then let the reference solution write the outputs:

```bash
scripts/outputs.sh two-sum
```

`go test ./...` checks that every problem in `problems/` loads.
