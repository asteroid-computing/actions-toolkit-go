# actions-toolkit-go

A toolkit for writing GitHub Actions in Go, patterned after the JavaScript [actions/toolkit](https://github.com/actions/toolkit). The API is Go-shaped rather than a line-for-line port. Requires Go 1.27 or later.

```bash
go get github.com/asteroid-computing/actions-toolkit-go
```

## Packages

- **`core`** — inputs (`GetInput`, `GetBooleanInput`, `GetMultilineInput`), outputs and file commands (`SetOutput`, `SetEnv`, `SetState`, `AddToPath`), workflow commands (`Debug`, `Notice`, `Warning`, `Error`, `StartGroup`/`EndGroup`, `MaskValue`, `StopCommands`) and the job summary builder (`StepSummary`).
- **`actioncontext`** — the run's context, populated at import from the [default environment variables](docs/github-workflow-variables.md), with the event payload decoded into the matching [go-github](https://github.com/google/go-github) type (`PullRequestEvent()`, `PushEvent()`, …), the raw payload (`RawEventPayload()`), and `NewGitHubClient` targeting the run's API host (github.com or GHES).

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
	if pr, ok := actioncontext.CurrentActionContext.PullRequestEvent(); ok {
		fmt.Printf("PR #%d\n", pr.GetNumber())
	}
	core.SetOutput("greeting", "Hello, "+name)
}
```

[cmd/example](cmd/example/main.go) is a complete example.

## Developing

Tooling follows the [walle pattern](https://github.com/asteroid-computing/go-walle-bootstrap): `./walle --help` lists the tasks, and `./walle check` is the gate CI runs. [AGENTS.md](AGENTS.md) has the layout, rules and conventions.
