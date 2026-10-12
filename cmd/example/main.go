// Command example is a minimal GitHub Action step built on this toolkit. It is the module's release component: importing both [actioncontext] and [core] puts the library in its build closure, so a change to either package is in CI's vet, test and build scope. Replace or extend it, but keep both imports.
//
// It reads the optional `who-to-greet` input, logs a greeting naming the run's actor and repository, sets a `greeting` output and adds a job summary.
package main

import (
	"fmt"
	"os"

	"github.com/asteroid-computing/actions-toolkit-go/actioncontext"
	"github.com/asteroid-computing/actions-toolkit-go/core"
	"github.com/asteroid-computing/actions-toolkit-go/version"
)

func main() {
	if err := run(); err != nil {
		core.Error(err.Error(), nil)
		os.Exit(1)
	}
}

func run() error {
	who, err := core.GetInput("who-to-greet", nil)
	if err != nil {
		return err
	}
	ac := actioncontext.CurrentActionContext
	greeting := greet(who, ac.Actor, ac.Repository)
	fmt.Println(greeting)

	if ac.OutputFilePath != "" {
		if err := core.SetOutput("greeting", greeting); err != nil {
			return err
		}
	}
	if ac.StepSummaryFilePath != "" {
		core.StepSummary.AddHeading("Example", 2)
		core.StepSummary.AddRaw(greeting, true)
		core.StepSummary.AddRaw("toolkit "+version.Full(version.BuildMetadata), true)
		if err := core.StepSummary.Write(nil); err != nil {
			return err
		}
	}
	return nil
}

// greet builds the greeting, falling back to neutral wording for values that are unset outside GitHub Actions.
func greet(who, actor, repo string) string {
	if who == "" {
		who = "world"
	}
	if actor == "" {
		actor = "someone"
	}
	if repo == "" {
		repo = "a local checkout"
	}
	return fmt.Sprintf("Hello, %s! Run by %s in %s.", who, actor, repo)
}
