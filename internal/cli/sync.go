package cli

import (
	"flag"
	"fmt"

	"codeshot/internal/bank"
	"codeshot/internal/home"
)

func init() {
	register(command{name: "sync", run: syncCmd})
}

// syncCmd imports the problems of a bank into the database.
func syncCmd(args []string) int {
	flags := flag.NewFlagSet("sync", flag.ContinueOnError)
	from := flags.String("from", "", "git URL or local directory to take the problems from")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	s, err := openStore()
	if err != nil {
		return report(err)
	}
	defer s.Close()

	cache, err := home.Path("bank")
	if err != nil {
		return report(err)
	}
	src := bank.From(*from, cache)
	fmt.Printf("fetching problems from %s\n", src.Name())
	all, err := src.Fetch()
	if err != nil {
		return report(err)
	}
	for _, p := range all {
		if err := s.SaveProblem(p, src.Name()); err != nil {
			return report(err)
		}
	}
	fmt.Printf("%d problems synced\n", len(all))
	return 0
}
