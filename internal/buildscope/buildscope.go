// Package buildscope is the registry of the project's release components and the engine that computes each one's build closure: the module-local packages and files that compile into its entry-point binary.
//
// It answers two questions that must agree, for two callers that must agree:
//
//   - change detection — which components a set of changed files affects ([Affected]), so CI builds and validates only what a change touched;
//   - scoped validation — which packages a component comprises ([ClosurePackages]), so `go vet` / `go test` run against just those.
//
// walle (locally) and the changed-scope action (in CI) both import this package, so the answers cannot drift. The closure comes from `go list -deps`, not path globbing, because components share packages under internal/: a path prefix cannot say which binary an internal package compiles into, but the import graph can.
package buildscope

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/asteroid-computing/actions-toolkit-go/internal/gocmd"
)

// Component is a releasable binary and the entry point whose build closure defines it. BuildTags are the tags it ships with; the closure is computed under the same tags so it matches the files that actually compile in.
type Component struct {
	Name      string
	EntryPkg  string
	BuildTags []string
}

// Components is the ordered set of release components. Order is the canonical order for reports and outputs. To add a component, add its cmd/<name> package and an entry here; walle, the changed-scope action and the build action all pick it up.
//
// walle is a component so that a change to the tooling itself is in CI's scope: its closure covers every internal/ engine, so a PR touching only buildscope or buildrun still runs their vet and tests rather than being scoped to nothing.
var Components = []Component{
	{Name: "example", EntryPkg: "./cmd/example"},
	{Name: "walle", EntryPkg: "./cmd/walle"},
}

// Names returns the component names in canonical order — the source for CLI enums and help text, so they cannot go stale.
func Names() []string {
	names := make([]string, len(Components))
	for i, c := range Components {
		names[i] = c.Name
	}
	return names
}

// Lookup returns the component with the given name.
func Lookup(name string) (Component, bool) {
	for _, c := range Components {
		if c.Name == name {
			return c, true
		}
	}
	return Component{}, false
}

// Closure is a component's build closure within this module.
type Closure struct {
	Packages []string // module-local import paths, sorted
	Files    []string // module-relative, forward-slashed paths, sorted; always includes go.mod and go.sum
}

// closureTemplate emits one "P <import path>" line per package and one "F <abs path>" line per file. IgnoredGoFiles is included on purpose: a file excluded on the host (foo_arm64.go on an amd64 runner, foo_linux.go on a Mac) can still compile into a cross-compiled artifact, and missing a change is worse than over-building. Test files are included so a test-only change still validates the component.
const closureTemplate = `P {{.ImportPath}}
{{range .GoFiles}}F {{$.Dir}}/{{.}}
{{end}}{{range .IgnoredGoFiles}}F {{$.Dir}}/{{.}}
{{end}}{{range .EmbedFiles}}F {{$.Dir}}/{{.}}
{{end}}{{range .TestGoFiles}}F {{$.Dir}}/{{.}}
{{end}}{{range .XTestGoFiles}}F {{$.Dir}}/{{.}}
{{end}}`

// ClosureOf resolves one component's closure with a single `go list -deps` call.
//
// No -e: a package that fails to load (a broken import) fails the call rather than silently dropping out of the closure, which would make a broken change look like it touched nothing.
func ClosureOf(ctx context.Context, name string) (Closure, error) {
	comp, ok := Lookup(name)
	if !ok {
		return Closure{}, fmt.Errorf("unknown component %q (known: %s)", name, strings.Join(Names(), ", "))
	}
	mod, err := module()
	if err != nil {
		return Closure{}, err
	}
	args := []string{"list", "-deps", "-f", closureTemplate}
	if len(comp.BuildTags) > 0 {
		args = append(args, "-tags", strings.Join(comp.BuildTags, ","))
	}
	out, err := gocmd.Output(ctx, nil, "go", append(args, mod.importPath(comp.EntryPkg))...)
	if err != nil {
		return Closure{}, fmt.Errorf("closure of %s: %w", name, err)
	}
	return parseClosure(out, mod), nil
}

// parseClosure turns closureTemplate output into a Closure, keeping only this module's packages and files. Split out so it is testable without the toolchain.
func parseClosure(out string, mod moduleInfo) Closure {
	pkgs := map[string]struct{}{}
	files := map[string]struct{}{"go.mod": {}, "go.sum": {}}
	for line := range strings.Lines(out) {
		kind, value, ok := strings.Cut(strings.TrimSpace(line), " ")
		if !ok {
			continue
		}
		switch kind {
		case "P":
			if value == mod.Path || strings.HasPrefix(value, mod.Path+"/") {
				pkgs[value] = struct{}{}
			}
		case "F":
			rel, err := filepath.Rel(mod.Dir, value)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				continue // stdlib or module cache: never in a diff
			}
			files[filepath.ToSlash(rel)] = struct{}{}
		}
	}
	return Closure{Packages: sortedKeys(pkgs), Files: sortedKeys(files)}
}

// Closures resolves the closures of the named components (all components when names is empty) concurrently. Each `go list -deps` costs roughly half a second to a second on a real module; they are independent and read-only, so running them in parallel makes Affected cost about one call instead of one per component.
func Closures(ctx context.Context, names ...string) (map[string]Closure, error) {
	if len(names) == 0 {
		names = Names()
	}
	results := make([]Closure, len(names))
	g, ctx := errgroup.WithContext(ctx)
	for i, name := range names {
		g.Go(func() error {
			c, err := ClosureOf(ctx, name)
			results[i] = c
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	out := make(map[string]Closure, len(names))
	for i, name := range names {
		out[name] = results[i]
	}
	return out, nil
}

// ClosurePackages returns the sorted union of the named components' module-local packages — the argument list for `go vet` / `go test` to validate exactly those components. No names means every component.
func ClosurePackages(ctx context.Context, names ...string) ([]string, error) {
	closures, err := Closures(ctx, names...)
	if err != nil {
		return nil, err
	}
	set := map[string]struct{}{}
	for _, c := range closures {
		for _, p := range c.Packages {
			set[p] = struct{}{}
		}
	}
	return sortedKeys(set), nil
}

// Affected reports, for every component, whether any changed file (module-relative) is in its closure.
func Affected(ctx context.Context, changed []string) (map[string]bool, error) {
	closures, err := Closures(ctx)
	if err != nil {
		return nil, err
	}
	files := make(map[string][]string, len(closures))
	for name, c := range closures {
		files[name] = c.Files
	}
	return AffectedIn(changed, files), nil
}

// AffectedIn is [Affected] with the closures supplied, so the decision is testable without the toolchain. Every component appears in the result; one missing from closures is reported unaffected.
func AffectedIn(changed []string, closures map[string][]string) map[string]bool {
	set := make(map[string]struct{}, len(changed))
	for _, f := range changed {
		if f = strings.TrimSpace(f); f != "" {
			set[filepath.ToSlash(f)] = struct{}{}
		}
	}
	result := make(map[string]bool, len(Components))
	for _, comp := range Components {
		result[comp.Name] = slices.ContainsFunc(closures[comp.Name], func(f string) bool {
			_, ok := set[f]
			return ok
		})
	}
	return result
}

// AffectedNames returns the affected component names in canonical order.
func AffectedNames(affected map[string]bool) []string {
	var names []string
	for _, comp := range Components {
		if affected[comp.Name] {
			names = append(names, comp.Name)
		}
	}
	return names
}

// GitChanged returns the module-relative paths changed between the merge base of base and HEAD, and HEAD — what a pull request changed, computed from the checkout rather than the GitHub API (no token, no pagination, no 3000-file API cap). The checkout needs enough history to find the merge base: actions/checkout with fetch-depth: 0.
//
// git reports paths relative to the repository root; they match module-relative paths when the module is the repository root, which is this template's layout.
func GitChanged(ctx context.Context, base string) ([]string, error) {
	out, err := gocmd.Output(ctx, nil, "git", "diff", "--name-only", "--no-renames", base+"...HEAD")
	if err != nil {
		return nil, fmt.Errorf("listing files changed since %s: %w", base, err)
	}
	var files []string
	for line := range strings.Lines(out) {
		if f := strings.TrimSpace(line); f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}

type moduleInfo struct{ Path, Dir string }

// importPath turns a module-relative package ("./cmd/example") into its import path, so go commands resolve it from any directory inside the module rather than only from the root.
func (m moduleInfo) importPath(rel string) string {
	return path.Join(m.Path, rel)
}

// ImportPath returns a component's entry-point import path.
func ImportPath(ctx context.Context, c Component) (string, error) {
	mod, err := module()
	if err != nil {
		return "", err
	}
	return mod.importPath(c.EntryPkg), nil
}

// module returns the main module's import path and root directory. It is resolved once per process: closures are computed concurrently and each needs it.
var module = sync.OnceValues(func() (moduleInfo, error) { return loadModule(context.Background()) })

func loadModule(ctx context.Context) (moduleInfo, error) {
	out, err := gocmd.Output(ctx, nil, "go", "list", "-m", "-f", "{{.Path}}\t{{.Dir}}")
	if err != nil {
		return moduleInfo{}, fmt.Errorf("locating module: %w", err)
	}
	path, dir, ok := strings.Cut(strings.TrimSpace(out), "\t")
	if !ok || path == "" || dir == "" {
		return moduleInfo{}, fmt.Errorf("locating module: unexpected `go list -m` output %q (run from inside the module)", out)
	}
	return moduleInfo{Path: path, Dir: dir}, nil
}

func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
