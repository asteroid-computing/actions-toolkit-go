package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/asteroid-computing/actions-toolkit-go/internal/buildrun"
	"github.com/asteroid-computing/actions-toolkit-go/internal/buildscope"
	"github.com/asteroid-computing/actions-toolkit-go/internal/gocmd"
)

// scope is the package list for vet and test: the whole module, or the union of the given components' closures. CI passes the components a change affected, so unrelated packages are not re-vetted or re-tested.
func scope(ctx context.Context, components []string) ([]string, error) {
	if len(components) == 0 {
		return []string{"./..."}, nil
	}
	return buildscope.ClosurePackages(ctx, components...)
}

type fmtCmd struct {
	Check bool `help:"List unformatted files and fail instead of rewriting them. Never modifies the tree, so it is safe in CI."`
}

func (c fmtCmd) Run(ctx context.Context) error {
	if !c.Check {
		return gocmd.Run(ctx, nil, "go", "fmt", "./...")
	}
	out, err := gocmd.Output(ctx, nil, "gofmt", "-l", ".")
	if err != nil {
		return err
	}
	if files := strings.Fields(out); len(files) > 0 {
		for _, f := range files {
			// ::error file=…:: annotates the file in a GitHub PR diff; it is an inert prefix anywhere else.
			fmt.Printf("::error file=%s::not gofmt-formatted\n", f)
		}
		return fmt.Errorf("%d file(s) are not formatted; run ./walle fmt", len(files))
	}
	return nil
}

type tidyCmd struct {
	Check bool `help:"Fail, printing the diff, if go.mod or go.sum is not tidy. Never modifies the tree."`
}

func (c tidyCmd) Run(ctx context.Context) error {
	if c.Check {
		if err := gocmd.Run(ctx, nil, "go", "mod", "tidy", "-diff"); err != nil {
			return errors.New("go.mod/go.sum are not tidy; run ./walle tidy")
		}
		return nil
	}
	return gocmd.Run(ctx, nil, "go", "mod", "tidy")
}

type vetCmd struct {
	Components []string `sep:"," enum:"${components}" help:"Vet only these components' package closures (${components})."`
}

func (c vetCmd) Run(ctx context.Context) error {
	pkgs, err := scope(ctx, c.Components)
	if err != nil {
		return err
	}
	return gocmd.Run(ctx, nil, "go", append([]string{"vet"}, pkgs...)...)
}

type testCmd struct {
	Race       bool     `help:"Enable the race detector."`
	Components []string `sep:"," enum:"${components}" help:"Test only these components' package closures (${components})."`
}

func (c testCmd) Run(ctx context.Context) error {
	pkgs, err := scope(ctx, c.Components)
	if err != nil {
		return err
	}
	args := []string{"test"}
	if c.Race {
		args = append(args, "-race")
	}
	return gocmd.Run(ctx, nil, "go", append(args, pkgs...)...)
}

type checkCmd struct {
	Gate        checkGateCmd        `cmd:"" default:"withargs" help:"The gate (default): fmt --check, tidy --check, vet, race tests."`
	Determinism checkDeterminismCmd `cmd:"" help:"Build every component twice and fail if the binaries differ."`
	Actions     checkActionsCmd     `cmd:"" help:"Fail if a committed action binary under .github/actions/*/dist does not match its source."`
}

// checkGateCmd is what CI runs and what to run before pushing. Nothing in it modifies the tree: fenton's gate ran a mutating `go mod tidy`, which passes locally and leaves the fix uncommitted.
type checkGateCmd struct {
	Components []string `sep:"," enum:"${components}" help:"Scope vet and tests to these components (${components}). Format and tidy checks always cover the whole module."`
}

func (c checkGateCmd) Run(ctx context.Context) error {
	steps := []struct {
		name string
		run  func(context.Context) error
	}{
		{"fmt", fmtCmd{Check: true}.Run},
		{"tidy", tidyCmd{Check: true}.Run},
		{"vet", vetCmd{Components: c.Components}.Run},
		{"test", testCmd{Race: true, Components: c.Components}.Run},
	}
	for _, s := range steps {
		fmt.Fprintf(os.Stderr, "==> %s\n", s.name)
		if err := s.run(ctx); err != nil {
			return fmt.Errorf("%s: %w", s.name, err)
		}
	}
	return nil
}

type checkDeterminismCmd struct{}

// Run builds each component twice into separate directories with fixed metadata and compares digests, using buildrun.BuildInto — the function the real build uses — so the check cannot pass against flags the shipped artifact does not use.
func (checkDeterminismCmd) Run(ctx context.Context) error {
	const meta = "local.determinism-check"
	var failed []string
	for _, comp := range buildscope.Components {
		var sums [2]string
		for i := range sums {
			dir, err := os.MkdirTemp("", "walle-det-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(dir)
			out := dir + "/" + comp.Name
			if err := buildrun.BuildInto(ctx, comp.Name, buildrun.Host, out, meta); err != nil {
				return err
			}
			if sums[i], err = buildrun.SHA256(out); err != nil {
				return err
			}
		}
		if sums[0] != sums[1] {
			fmt.Printf("FAIL %s: %s != %s\n", comp.Name, sums[0], sums[1])
			failed = append(failed, comp.Name)
			continue
		}
		fmt.Printf("ok   %s %s\n", comp.Name, sums[0])
	}
	if len(failed) > 0 {
		return fmt.Errorf("not reproducible: %s", strings.Join(failed, ", "))
	}
	return nil
}

type checkActionsCmd struct{}

func (checkActionsCmd) Run(ctx context.Context) error {
	checked, stale, err := buildrun.StaleActions(ctx)
	if err != nil {
		return err
	}
	if checked == 0 {
		fmt.Println("No committed action binaries; actions build from source (see AGENTS.md).")
		return nil
	}
	for _, s := range stale {
		fmt.Printf("::error file=%s::%s\n", s.Path, s.Reason)
	}
	if len(stale) > 0 {
		return fmt.Errorf("%d of %d committed action binaries are stale; run ./walle build actions and commit dist/", len(stale), checked)
	}
	fmt.Printf("%d committed action binaries match source.\n", checked)
	return nil
}
