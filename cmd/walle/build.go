package main

import (
	"context"
	"fmt"

	"github.com/asteroid-computing/actions-toolkit-go/internal/buildrun"
	"github.com/asteroid-computing/actions-toolkit-go/internal/buildscope"
)

type buildCmd struct {
	Components buildComponentsCmd `cmd:"" default:"withargs" help:"Build release components (default: all) for the host into bin/, or with --all-platforms into bin/release/."`
	Actions    buildActionsCmd    `cmd:"" help:"Cross-compile every Go GitHub Action into .github/actions/<name>/dist/ for committing."`
}

// buildComponentsCmd is a thin shell over internal/buildrun, which the build action also calls: flags, output paths and platforms live there so CI and a laptop build identically.
type buildComponentsCmd struct {
	Names        []string `arg:"" optional:"" enum:"${components}" help:"Components to build (${components}). Default: all."`
	AllPlatforms bool     `help:"Cross-compile for every release platform instead of the host."`
}

func (c buildComponentsCmd) Run(ctx context.Context) error {
	names := c.Names
	if len(names) == 0 {
		names = buildscope.Names()
	}
	meta := buildrun.Metadata(ctx)
	for _, name := range names {
		if !c.AllPlatforms {
			res, err := buildrun.BuildHost(ctx, name, meta)
			if err != nil {
				return err
			}
			report(res)
			continue
		}
		results, err := buildrun.BuildAll(ctx, name, meta, buildrun.ReleasePlatforms)
		if err != nil {
			return err
		}
		for _, res := range results {
			report(res)
		}
	}
	return nil
}

type buildActionsCmd struct{}

func (buildActionsCmd) Run(ctx context.Context) error {
	results, err := buildrun.BuildActions(ctx, buildrun.ActionPlatforms)
	if err != nil {
		return err
	}
	for _, res := range results {
		report(res)
	}
	fmt.Println("Commit the dist/ directories; ./walle check actions verifies them in CI.")
	return nil
}

func report(res buildrun.Result) {
	fmt.Printf("built %s (%s)\n", res.Output, res.Platform)
}
