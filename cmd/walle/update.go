package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/asteroid-computing/actions-toolkit-go/internal/gocmd"
)

// updateCmd upgrades every direct dependency with a newer version, in one `go get`, then tidies.
//
// It does not run `go clean -modcache -cache` first: those caches are shared by every Go project on the machine, so clearing them costs every other checkout a full re-download and recompile to work around a rare `go get` verification failure. If that failure occurs, clear the caches by hand.
type updateCmd struct {
	Yes bool `short:"y" help:"Upgrade without asking."`
}

func (c updateCmd) Run(ctx context.Context) error {
	const tmpl = `{{if and (not .Indirect) (not .Main) .Update}}{{.Path}} {{.Version}} {{.Update.Version}}{{"\n"}}{{end}}`
	out, err := gocmd.Output(ctx, nil, "go", "list", "-m", "-u", "-f", tmpl, "all")
	if err != nil {
		return fmt.Errorf("listing available updates: %w", err)
	}
	var targets []string
	for line := range strings.Lines(out) {
		if f := strings.Fields(line); len(f) == 3 {
			fmt.Printf("  %s  %s -> %s\n", f[0], f[1], f[2])
			targets = append(targets, f[0]+"@"+f[2])
		}
	}
	if len(targets) == 0 {
		fmt.Println("All direct dependencies are up to date.")
		return nil
	}
	if !c.Yes && !confirm(fmt.Sprintf("Upgrade these %d dependencies?", len(targets))) {
		return nil
	}
	if err := gocmd.Run(ctx, nil, "go", append([]string{"get"}, targets...)...); err != nil {
		return err
	}
	if err := gocmd.Run(ctx, nil, "go", "mod", "tidy"); err != nil {
		return err
	}
	fmt.Println("Upgraded. Run ./walle check to verify.")
	return nil
}

// confirm asks a yes/no question on the terminal. Without a terminal it answers no rather than blocking; callers offer --yes for non-interactive use.
func confirm(prompt string) bool {
	if fi, err := os.Stdin.Stat(); err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		fmt.Fprintln(os.Stderr, "stdin is not a terminal; not prompting (pass --yes to proceed)")
		return false
	}
	fmt.Fprintf(os.Stderr, "%s [y/N] ", prompt)
	s := bufio.NewScanner(os.Stdin)
	if !s.Scan() {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(s.Text()))
	return answer == "y" || answer == "yes"
}
