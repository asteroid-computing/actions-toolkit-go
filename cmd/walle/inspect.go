package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/asteroid-computing/actions-toolkit-go/internal/buildrun"
	"github.com/asteroid-computing/actions-toolkit-go/internal/buildscope"
	"github.com/asteroid-computing/actions-toolkit-go/version"
)

type closureCmd struct {
	Component string `arg:"" enum:"${components}" help:"Component whose closure to print (${components})."`
	Packages  bool   `help:"Print packages instead of files."`
}

func (c closureCmd) Run(ctx context.Context) error {
	closure, err := buildscope.ClosureOf(ctx, c.Component)
	if err != nil {
		return err
	}
	lines := closure.Files
	if c.Packages {
		lines = closure.Packages
	}
	fmt.Println(strings.Join(lines, "\n"))
	return nil
}

// changedCmd is the local twin of the changed-scope action: same engine, same answer. `./walle changed --base origin/main` shows what CI will build for the current branch. One source is required; there is no implicit stdin read, which would sit silently waiting when run from a terminal.
type changedCmd struct {
	Base     string   `xor:"source" help:"Diff against the merge base with this ref (e.g. origin/main), as the changed-scope action does."`
	Files    []string `xor:"source" sep:"," help:"Comma-separated changed paths (module-relative)."`
	FromFile string   `xor:"source" help:"Read newline-separated changed paths from this file, or - for stdin."`
}

func (c changedCmd) Run(ctx context.Context) error {
	changed, err := c.changed(ctx)
	if err != nil {
		return err
	}
	affected, err := buildscope.Affected(ctx, changed)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(affected)
}

func (c changedCmd) changed(ctx context.Context) ([]string, error) {
	var raw []byte
	var err error
	switch {
	case c.Base != "":
		return buildscope.GitChanged(ctx, c.Base)
	case len(c.Files) > 0:
		return c.Files, nil
	case c.FromFile == "-":
		raw, err = io.ReadAll(os.Stdin)
	case c.FromFile != "":
		raw, err = os.ReadFile(c.FromFile)
	default:
		return nil, errors.New("no changed files given: pass --base <ref>, --files <a,b> or --from-file <path|->")
	}
	return strings.Split(string(raw), "\n"), err
}

type versionCmd struct {
	Full   bool `xor:"form" help:"Append the build metadata a build here would inject, e.g. 0.1.0+local.abc1234."`
	Semver bool `xor:"form" help:"Print with a leading v, e.g. v0.1.0 (the tag form)."`
}

func (c versionCmd) Run(ctx context.Context) error {
	switch {
	case c.Full:
		// The metadata a build would inject now, not walle's own (empty under go run).
		fmt.Println(version.Full(buildrun.Metadata(ctx)))
	case c.Semver:
		fmt.Println(version.Semver())
	default:
		fmt.Println(version.String())
	}
	return nil
}
