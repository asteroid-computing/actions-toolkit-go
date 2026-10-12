// Package buildrun compiles the project's release components and its Go GitHub Actions.
//
// It is the single build engine behind walle (on a laptop) and the build action (in CI): both import it, so an artifact built in CI is built by the same code, with the same flags, as one built locally.
//
// # Reproducibility is load-bearing
//
// Every artifact is a pure function of source, Go toolchain version and the injected metadata — never of the machine or the moment. Two consumers depend on that: a deploy tool that diffs asset hashes (CDK, for one) redeploys only what changed, and `walle check actions` can prove a committed action binary matches its source by rebuilding it on a different host and comparing bytes. Hence the fixed flag set in [goBuild]:
//
//   - CGO_ENABLED=0 — a static binary, no host C toolchain in the output;
//   - -trimpath — no absolute host paths;
//   - -buildvcs=false — no vcs.revision / vcs.modified stamp, which would differ between a clean and a dirty tree and between commits that did not touch the component;
//   - -ldflags "-s -w" — no symbol table or DWARF: roughly a third smaller, stack traces still work;
//   - metadata only via -X, and never a timestamp or CI run id.
//
// The package execs `go build` rather than linking the toolchain: there is no library API that compiles and links a binary.
package buildrun

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/asteroid-computing/actions-toolkit-go/internal/buildscope"
	"github.com/asteroid-computing/actions-toolkit-go/internal/gocmd"
	"github.com/asteroid-computing/actions-toolkit-go/version"
)

const (
	// BinDir receives host builds: bin/<component>, directly runnable.
	BinDir = "bin"
	// ReleaseDir receives cross-compiled release assets: bin/release/<component>-<os>-<arch>.
	ReleaseDir = "bin/release"
	// ActionsSrcDir holds one main package per Go GitHub Action: cmd/github-actions/<name>.
	ActionsSrcDir = "cmd/github-actions"
	// ActionsDir holds the action definitions: .github/actions/<name>/{action.yaml,index.js,dist/}.
	ActionsDir = ".github/actions"

	dirMode = 0o750
)

// Platform is a GOOS/GOARCH pair.
type Platform struct{ OS, Arch string }

func (p Platform) String() string { return p.OS + "/" + p.Arch }

// Host is the platform walle is running on.
var Host = Platform{runtime.GOOS, runtime.GOARCH}

// ReleasePlatforms is what `build --all-platforms` cross-compiles for. Each platform is a full compile of the component's closure — cross-compilation shares no cached output across a GOOS/GOARCH boundary — so list only platforms something actually runs on.
var ReleasePlatforms = []Platform{{"linux", "amd64"}, {"linux", "arm64"}, {"darwin", "arm64"}}

// ActionPlatforms is what `build actions` cross-compiles each action for. GitHub-hosted Linux runners are amd64 and arm64; add {"darwin", "arm64"} if you run these actions on macOS runners. The shim falls back to building from source on any platform without a binary, so a missing entry costs a compile, not a failure.
var ActionPlatforms = []Platform{{"linux", "amd64"}, {"linux", "arm64"}}

// Result records one built artifact.
type Result struct {
	Component string
	Platform  Platform
	Output    string
}

// BuildHost builds a component for the host into bin/<name>.
func BuildHost(ctx context.Context, component, metadata string) (Result, error) {
	comp, pkg, err := entry(ctx, component)
	if err != nil {
		return Result{}, err
	}
	out := filepath.Join(BinDir, comp.Name)
	if err := goBuild(ctx, Host, pkg, out, comp.BuildTags, metadata); err != nil {
		return Result{}, err
	}
	return Result{Component: comp.Name, Platform: Host, Output: out}, nil
}

// BuildAll cross-compiles a component for every platform concurrently into bin/release/<name>-<os>-<arch>. Results come back in platform order regardless of completion order, so reports read the same on every run. The group fails fast: the first error cancels the other builds.
func BuildAll(ctx context.Context, component, metadata string, platforms []Platform) ([]Result, error) {
	comp, pkg, err := entry(ctx, component)
	if err != nil {
		return nil, err
	}
	results := make([]Result, len(platforms))
	g, ctx := errgroup.WithContext(ctx)
	for i, p := range platforms {
		g.Go(func() error {
			out := filepath.Join(ReleaseDir, fmt.Sprintf("%s-%s-%s", comp.Name, p.OS, p.Arch))
			results[i] = Result{Component: comp.Name, Platform: p, Output: out}
			return goBuild(ctx, p, pkg, out, comp.BuildTags, metadata)
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

// BuildInto builds a component for one platform to an explicit path. The determinism check uses it so it exercises exactly the flags the real build uses.
func BuildInto(ctx context.Context, component string, p Platform, out, metadata string) error {
	comp, pkg, err := entry(ctx, component)
	if err != nil {
		return err
	}
	return goBuild(ctx, p, pkg, out, comp.BuildTags, metadata)
}

// entry looks up a component and resolves its entry point to an import path.
func entry(ctx context.Context, component string) (buildscope.Component, string, error) {
	comp, ok := buildscope.Lookup(component)
	if !ok {
		return comp, "", fmt.Errorf("unknown component %q (known: %s)", component, strings.Join(buildscope.Names(), ", "))
	}
	pkg, err := buildscope.ImportPath(ctx, comp)
	return comp, pkg, err
}

// Actions lists the Go GitHub Actions by convention: every directory under cmd/github-actions is one, named after the directory, with its definition in .github/actions/<name>. Discovery replaces a hand-kept registry, so there is no list to forget to update.
func Actions() ([]string, error) {
	entries, err := os.ReadDir(ActionsSrcDir)
	if err != nil {
		return nil, fmt.Errorf("listing actions: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// ActionBinary is the committed path of an action's binary for a platform. The shim (.github/actions/_shim/shim.js) resolves the same name.
func ActionBinary(name string, p Platform) string {
	ext := ""
	if p.OS == "windows" {
		ext = ".exe"
	}
	return filepath.Join(ActionsDir, name, "dist", fmt.Sprintf("%s_%s_%s%s", name, p.OS, p.Arch, ext))
}

// BuildActions cross-compiles every action for every platform, concurrently. Action binaries carry no metadata: they are committed, and metadata would make every rebuild a diff.
func BuildActions(ctx context.Context, platforms []Platform) ([]Result, error) {
	names, err := Actions()
	if err != nil {
		return nil, err
	}
	var results []Result
	for _, name := range names {
		for _, p := range platforms {
			results = append(results, Result{Component: name, Platform: p, Output: ActionBinary(name, p)})
		}
	}
	g, ctx := errgroup.WithContext(ctx)
	for _, r := range results {
		g.Go(func() error { return goBuild(ctx, r.Platform, "./"+ActionsSrcDir+"/"+r.Component, r.Output, nil, "") })
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

// Stale describes a committed action binary that does not match its source.
type Stale struct {
	Path   string
	Reason string
}

// StaleActions rebuilds every committed action binary into a temporary directory and compares bytes. Only binaries present in dist/ are checked, so an action that ships no binaries (source mode) is never stale.
//
// Each rebuild uses the Go toolchain recorded in the committed binary (GOTOOLCHAIN=goX.Y.Z, which the go command downloads once if needed), so a developer on a newer patch release than CI does not produce a false positive.
func StaleActions(ctx context.Context) (checked int, stale []Stale, err error) {
	names, err := Actions()
	if err != nil {
		return 0, nil, err
	}
	tmp, err := os.MkdirTemp("", "walle-actions-*")
	if err != nil {
		return 0, nil, err
	}
	defer os.RemoveAll(tmp)

	for _, name := range names {
		bins, _ := filepath.Glob(filepath.Join(ActionsDir, name, "dist", name+"_*_*"))
		for _, committed := range bins {
			p, ok := parseActionPlatform(name, filepath.Base(committed))
			if !ok {
				continue
			}
			checked++
			goVersion, err := embeddedGoVersion(ctx, committed)
			if err != nil {
				stale = append(stale, Stale{committed, err.Error()})
				continue
			}
			rebuilt := filepath.Join(tmp, filepath.Base(committed))
			env := map[string]string{"GOTOOLCHAIN": goVersion}
			if err := goBuildEnv(ctx, env, p, "./"+ActionsSrcDir+"/"+name, rebuilt, nil, ""); err != nil {
				return checked, stale, err
			}
			same, err := sameFile(committed, rebuilt)
			if err != nil {
				return checked, stale, err
			}
			if !same {
				stale = append(stale, Stale{committed, "does not match a rebuild from source with " + goVersion})
			}
		}
	}
	return checked, stale, nil
}

func parseActionPlatform(name, base string) (Platform, bool) {
	rest, ok := strings.CutPrefix(strings.TrimSuffix(base, ".exe"), name+"_")
	if !ok {
		return Platform{}, false
	}
	goos, arch, ok := strings.Cut(rest, "_")
	return Platform{goos, arch}, ok && goos != "" && arch != ""
}

// embeddedGoVersion reads the toolchain a binary was built with from its build info (`go version <file>` prints "<file>: go1.x.y").
func embeddedGoVersion(ctx context.Context, path string) (string, error) {
	out, err := gocmd.Output(ctx, nil, "go", "version", path)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(out)
	if len(fields) < 2 || !strings.HasPrefix(fields[len(fields)-1], "go") {
		return "", fmt.Errorf("cannot read Go version from %s: %q", path, out)
	}
	return fields[len(fields)-1], nil
}

// Ldflags returns the linker flags for an artifact: stripped, with metadata injected into the version package when non-empty.
func Ldflags(metadata string) string {
	flags := "-s -w"
	if metadata != "" {
		flags += fmt.Sprintf(" -X %s.BuildMetadata=%s", version.ImportPath, metadata)
	}
	return flags
}

func goBuild(ctx context.Context, p Platform, pkg, out string, tags []string, metadata string) error {
	return goBuildEnv(ctx, nil, p, pkg, out, tags, metadata)
}

// goBuildEnv is the one place a `go build` command line is assembled. See the package doc for why each flag is here.
func goBuildEnv(ctx context.Context, extra map[string]string, p Platform, pkg, out string, tags []string, metadata string) error {
	if err := os.MkdirAll(filepath.Dir(out), dirMode); err != nil {
		return err
	}
	env := map[string]string{"GOOS": p.OS, "GOARCH": p.Arch, "CGO_ENABLED": "0"}
	maps.Copy(env, extra)
	args := []string{"build", "-trimpath", "-buildvcs=false", "-ldflags", Ldflags(metadata)}
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}
	args = append(args, "-o", out, pkg)
	if err := gocmd.Quiet(ctx, env, "go", args...); err != nil {
		return fmt.Errorf("building %s for %s: %w", pkg, p, err)
	}
	return nil
}

// SHA256 returns the hex digest of a file.
func SHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func sameFile(a, b string) (bool, error) {
	ha, err := SHA256(a)
	if err != nil {
		return false, err
	}
	hb, err := SHA256(b)
	if err != nil {
		return false, err
	}
	return ha == hb, nil
}
