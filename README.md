# actions-toolkit-go

[![CI](https://github.com/asteroid-computing/actions-toolkit-go/actions/workflows/ci.yaml/badge.svg)](https://github.com/asteroid-computing/actions-toolkit-go/actions/workflows/ci.yaml) [![Go Reference](https://pkg.go.dev/badge/github.com/asteroid-computing/actions-toolkit-go.svg)](https://pkg.go.dev/github.com/asteroid-computing/actions-toolkit-go) [![Release](https://img.shields.io/github/v/release/asteroid-computing/actions-toolkit-go)](https://github.com/asteroid-computing/actions-toolkit-go/releases) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A toolkit for writing GitHub Actions in Go, patterned after the JavaScript [actions/toolkit](https://github.com/actions/toolkit). The API is Go-shaped rather than a line-for-line port. Requires Go 1.27 or later.

```bash
go get github.com/asteroid-computing/actions-toolkit-go
```

The API is not yet stable: while the version is below 1.0.0, a minor release may contain breaking changes, which [CHANGELOG.md](CHANGELOG.md) calls out.

## Packages

- **`core`** — inputs (`GetInput`, `GetBooleanInput`, `GetMultilineInput`), outputs and file commands (`SetOutput`, `SetEnv`, `SetState`, `AddToPath`), workflow commands (`Debug`, `Notice`, `Warning`, `Error`, `StartGroup`/`EndGroup`, `MaskValue`, `StopCommands`) and the job summary builder (`StepSummary`).
- **`actioncontext`** — the run's context, read by `actioncontext.Load()` from the [default environment variables](docs/github-workflow-variables.md), with the event payload decoded into the matching [go-github](https://github.com/google/go-github) type (`PullRequestEvent()`, `PushEvent()`, …), the raw payload (`RawEventPayload()`), and `NewGitHubClient` targeting the run's API host (github.com or GHES).

```go
package main

import (
	"fmt"
	"os"

	"github.com/asteroid-computing/actions-toolkit-go/actioncontext"
	"github.com/asteroid-computing/actions-toolkit-go/core"
)

func main() {
	name, err := core.GetInput("name", &core.InputOptions{Required: true})
	if err != nil {
		core.Error(err.Error(), nil)
		os.Exit(1)
	}
	ac, err := actioncontext.Load()
	if err != nil {
		core.Error(err.Error(), nil)
		os.Exit(1)
	}
	if pr, ok := ac.PullRequestEvent(); ok {
		fmt.Printf("PR #%d\n", pr.GetNumber())
	}
	core.SetOutput("greeting", "Hello, "+name)
}
```

[cmd/example](cmd/example/main.go) is a complete example.

## Developing

Tooling is plain Go behind the `./walle` task runner: `./walle --help` lists the tasks, and `./walle check` is the gate CI runs. [AGENTS.md](AGENTS.md) has the layout, rules and conventions.

Releases are cut by [release-please](https://github.com/googleapis/release-please) from [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) on main: merging its release PR tags `vX.Y.Z` and publishes the GitHub release. See [CHANGELOG.md](CHANGELOG.md).

Contributions are welcome; see [CONTRIBUTING.md](CONTRIBUTING.md). Report security issues as described in [SECURITY.md](SECURITY.md), not in public issues.

## License

[MIT](LICENSE)
