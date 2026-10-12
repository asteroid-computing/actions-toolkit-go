// Command changed-scope is the Go entry point of the changed-scope GitHub Action: it reports which release components a change touched, so the workflow builds and validates only those.
//
// It diffs the checkout against the merge base with the `base` input (internal/buildscope.GitChanged) and intersects the result with each component's build closure (buildscope.Affected) — the same engine `./walle changed --base <ref>` uses locally. Requires actions/setup-go (closures come from `go list`) and actions/checkout with fetch-depth: 0.
//
// Outputs are deliberately component-agnostic — a comma list, a JSON array and a boolean — so adding a component needs no change to action.yaml. Gate a job on one component with contains(fromJSON(steps.scope.outputs.components-json), 'name').
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/asteroid-computing/actions-toolkit-go/core"
	"github.com/asteroid-computing/actions-toolkit-go/internal/buildscope"
)

func main() {
	if err := run(context.Background()); err != nil {
		core.Error(err.Error(), nil)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	base, err := core.GetInput("base", nil)
	if err != nil {
		return err
	}

	var names []string
	if noBase(base) {
		// No usable base — a new branch's first push, or a manual run. Scope cannot be narrowed, so everything is in scope: over-building is safe, under-building is not.
		core.Notice("no base commit to diff against; treating every component as affected", nil)
		names = buildscope.Names()
	} else {
		changed, err := buildscope.GitChanged(ctx, base)
		if err != nil {
			return fmt.Errorf("%w (does the checkout use fetch-depth: 0?)", err)
		}
		affected, err := buildscope.Affected(ctx, changed)
		if err != nil {
			return err
		}
		names = buildscope.AffectedNames(affected)
	}

	asJSON, err := json.Marshal(append([]string{}, names...)) // [] rather than null when empty, so fromJSON yields an array
	if err != nil {
		return err
	}
	for k, v := range map[string]string{
		"components":      strings.Join(names, ","),
		"components-json": string(asJSON),
		"any":             fmt.Sprint(len(names) > 0),
	} {
		if err := core.SetOutput(k, v); err != nil {
			return fmt.Errorf("setting output %s: %w", k, err)
		}
	}
	fmt.Printf("affected components: %s\n", strings.Join(names, ", "))
	return nil
}

// noBase reports whether base cannot be diffed against: empty, or the all-zero SHA GitHub sends as `before` when a branch is created.
func noBase(base string) bool {
	return strings.Trim(base, "0") == ""
}
