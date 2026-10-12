// Command walle is the project's developer task runner: build, check, inspect and maintain, in plain Go.
//
// Run it from anywhere in the repository through the ./walle wrapper (`./walle build`), or as `go run ./cmd/walle <command>` from the module root. It needs nothing but the Go toolchain.
//
// Logic that CI also needs does not live here: it lives in internal/ (buildscope, buildrun, cireport), imported both by these commands and by the Go GitHub Actions under cmd/github-actions. walle commands are thin shells that parse flags and print; see AGENTS.md.
package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/asteroid-computing/actions-toolkit-go/internal/buildscope"
)

// CLI is the root command tree. The group tags put commands under the headings declared in main; within a group, field order is display order, so arrange by how often a command is reached for.
type CLI struct {
	Check checkCmd `cmd:"" group:"develop" help:"Run the gate CI runs: format, tidy, vet and race tests. Subcommands check determinism and committed action binaries."`
	Test  testCmd  `cmd:"" group:"develop" help:"Run the tests, optionally scoped to components."`
	Fmt   fmtCmd   `cmd:"" group:"develop" help:"Format Go source (go fmt), or with --check list unformatted files and fail."`
	Vet   vetCmd   `cmd:"" group:"develop" help:"Run go vet, optionally scoped to components."`
	Tidy  tidyCmd  `cmd:"" group:"develop" help:"Tidy go.mod (go mod tidy), or with --check fail if it is not tidy."`

	Build buildCmd `cmd:"" group:"build" help:"Build release components, or the Go GitHub Actions."`

	Changed changedCmd `cmd:"" group:"inspect" help:"Report which components a set of changed files affects, as JSON."`
	Closure closureCmd `cmd:"" group:"inspect" help:"Print a component's build closure: the module files that compile into it."`
	Version versionCmd `cmd:"" group:"inspect" help:"Print the project version."`

	Update updateCmd `cmd:"" group:"maintain" help:"Upgrade every direct dependency that has an update, then tidy."`
	Cache  cacheCmd  `cmd:"" group:"maintain" help:"Go build-cache maintenance for CI (trim to a run's working set)."`
}

func main() {
	// Ctrl-C cancels the context, and every child process walle starts is bound to it (internal/gocmd uses exec.CommandContext), so an interrupted build does not leave a go build running behind the prompt.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var cli CLI
	k := kong.Parse(&cli,
		kong.Name("walle"),
		kong.Description("The project's developer task runner. Run as ./walle <command> from anywhere in the repository."),
		// kong renders grouped help natively; fenton's 300-line custom printer predates relying on this.
		kong.Groups{
			"develop":  "Develop:",
			"build":    "Build:",
			"inspect":  "Inspect this checkout:",
			"maintain": "Maintain:",
		},
		kong.Vars{"components": strings.Join(buildscope.Names(), ",")},
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.UsageOnError(),
	)
	k.FatalIfErrorf(k.Run())
}
