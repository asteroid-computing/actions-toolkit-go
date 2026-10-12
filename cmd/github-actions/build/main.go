// Command build is the Go entry point of the build GitHub Action: it compiles release components in the caller's job with internal/buildrun — the code `./walle build` runs — and writes the outcome to the job summary.
//
// Running in the caller's job is the point. A job per component pays checkout, toolchain setup and cache restore each time, and compiles the shared dependency tree each time; one job pays once and every component after the first reuses a warm build cache. Components build serially for the same reason: the first compiles the shared dependencies the rest reuse, and buildrun already fans out across platforms within a component.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/asteroid-computing/actions-toolkit-go/core"
	"github.com/asteroid-computing/actions-toolkit-go/internal/buildrun"
	"github.com/asteroid-computing/actions-toolkit-go/internal/buildscope"
	"github.com/asteroid-computing/actions-toolkit-go/internal/cireport"
	"github.com/asteroid-computing/actions-toolkit-go/version"
)

func main() {
	if err := run(context.Background()); err != nil {
		core.Error(err.Error(), nil)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	rawComponents, err := core.GetInput("components", nil)
	if err != nil {
		return err
	}
	requested, err := parseComponents(rawComponents)
	if err != nil {
		return err
	}
	rawAllPlatforms, err := core.GetInput("all-platforms", nil)
	if err != nil {
		return err
	}
	allPlatforms := rawAllPlatforms == "true"
	metadata, err := core.GetInput("metadata", nil)
	if err != nil {
		return err
	}
	if metadata == "" {
		metadata = buildrun.Metadata(ctx)
	}

	report := cireport.Report{
		Title:   "Build",
		Empty:   "No component's build closure was touched, so nothing was built.",
		Version: version.Full(metadata),
		RunURL:  runURL(),
	}
	// Build everything requested even after a failure: whether one component or all of them broke is the first thing a reader wants to know.
	for _, name := range requested {
		start := time.Now()
		outputs, err := buildOne(ctx, name, metadata, allPlatforms)
		item := cireport.Item{Name: name, Duration: time.Since(start), Status: cireport.StatusOK, Detail: strings.Join(outputs, "<br>")}
		if err != nil {
			core.Error(fmt.Sprintf("building %s: %v", name, err), nil)
			item.Status, item.Detail = cireport.StatusFailed, err.Error()
		}
		report.Items = append(report.Items, item)
	}

	// Report before failing: a failed build is when the summary is most useful.
	core.StepSummary.AddRaw(report.Markdown(), true)
	if err := core.StepSummary.Write(nil); err != nil {
		// A missing summary must not hide the build result or skip the output below.
		core.Warning("writing job summary: "+err.Error(), nil)
	}
	var built []string
	for _, it := range report.Items {
		if it.Status == cireport.StatusOK {
			built = append(built, it.Name)
		}
	}
	if err := core.SetOutput("built", strings.Join(built, ",")); err != nil {
		return err
	}
	if report.Failed() {
		return fmt.Errorf("one or more components failed to build")
	}
	return nil
}

// parseComponents validates the comma-separated input and returns it in canonical order, so the report reads the same on every run. An unknown name is an error, not a skip: it means the workflow and buildscope disagree, and quietly building nothing would look like a change that touched nothing.
func parseComponents(raw string) ([]string, error) {
	want := map[string]bool{}
	for name := range strings.SplitSeq(raw, ",") {
		if name = strings.TrimSpace(name); name == "" {
			continue
		}
		if _, ok := buildscope.Lookup(name); !ok {
			return nil, fmt.Errorf("unknown component %q (known: %s)", name, strings.Join(buildscope.Names(), ", "))
		}
		want[name] = true
	}
	var names []string
	for _, name := range buildscope.Names() {
		if want[name] {
			names = append(names, name)
		}
	}
	return names, nil
}

func buildOne(ctx context.Context, name, metadata string, allPlatforms bool) ([]string, error) {
	if !allPlatforms {
		res, err := buildrun.BuildHost(ctx, name, metadata)
		return []string{res.Output}, err
	}
	results, err := buildrun.BuildAll(ctx, name, metadata, buildrun.ReleasePlatforms)
	var outputs []string
	for _, res := range results {
		outputs = append(outputs, res.Output)
	}
	return outputs, err
}

func runURL() string {
	server, repo, id := os.Getenv("GITHUB_SERVER_URL"), os.Getenv("GITHUB_REPOSITORY"), os.Getenv("GITHUB_RUN_ID")
	if server == "" || repo == "" || id == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/actions/runs/%s", server, repo, id)
}
