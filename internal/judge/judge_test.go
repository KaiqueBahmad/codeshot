package judge

import (
	"context"
	"os/exec"
	"testing"

	"codeshot/internal/lang"
	"codeshot/internal/problem"
)

func TestSame(t *testing.T) {
	cases := []struct {
		got, want string
		same      bool
	}{
		{"1 2\n", "1 2\n", true},
		{"1 2", "1 2\n", true},
		{"1 2  \n\n\n", "1 2\n", true},
		{"1 2\r\n3\r\n", "1 2\n3\n", true},
		{"1  2\n", "1 2\n", false},
		{"\n1 2\n", "1 2\n", false},
		{"2 1\n", "1 2\n", false},
	}
	for _, c := range cases {
		if Same(c.got, c.want) != c.same {
			t.Errorf("Same(%q, %q) = %v", c.got, c.want, !c.same)
		}
	}
}

// The verdicts, from a real run in docker: each solution reads a number and
// is meant to print it back.
func TestJudge(t *testing.T) {
	if testing.Short() {
		t.Skip("runs docker")
	}
	if exec.Command("docker", "info").Run() != nil {
		t.Skip("docker is not running")
	}
	c, _ := lang.ByID("c")
	tests := []problem.Test{{Name: "01", Input: "5\n", Output: "5\n"}}
	limits := Limits{TimeMS: 500, MemoryMB: 64}

	solutions := map[string]string{
		Accepted:          `#include <stdio.h>` + "\n" + `int main(void){int n;scanf("%d",&n);printf("%d\n",n);}`,
		WrongAnswer:       `#include <stdio.h>` + "\n" + `int main(void){printf("6\n");}`,
		TimeLimitExceeded: `int main(void){for(;;);}`,
		MemoryLimit:       `#include <stdlib.h>` + "\n" + `#include <string.h>` + "\n" + `int main(void){for(;;){char*p=malloc(1<<20);memset(p,1,1<<20);}}`,
		RuntimeError:      `int main(void){return 3;}`,
		CompilationError:  `int main(void){ nope }`,
	}
	for want, code := range solutions {
		t.Run(want, func(t *testing.T) {
			out, err := Judge(context.Background(), c, code, tests, limits, Options{})
			if err != nil {
				t.Fatal(err)
			}
			if out.Verdict != want {
				t.Fatalf("verdict %s, want %s: %+v", out.Verdict, want, out)
			}
		})
	}
}
