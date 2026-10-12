# AGENTS.md — actions-toolkit-go

This repository is two things:

1. **The library** — a toolkit for writing GitHub Actions in Go, patterned after (not a copy of) the JavaScript [actions/toolkit](https://github.com/actions/toolkit): `core` (inputs, outputs, env/path/state file commands, workflow commands, job summary) and `actioncontext` (the run's context from the default environment variables plus the typed event payload, and a GitHub API client for the run's host).
2. **The tooling** — the walle pattern from [go-walle-bootstrap](https://github.com/asteroid-computing/go-walle-bootstrap): developer tooling and CI logic written in Go, not in Makefiles, shell scripts or workflow YAML. The repository's own Go actions are built on the library, so CI dogfoods it.

These rules are binding for humans and agents working here.

## The library

- **Go 1.27 or later.** `encoding/json/v2` and the `uuid` package are standard library in 1.27, so the library needs no `GOEXPERIMENT` and no third-party UUID module.
- **Testable by injection.** Exported functions are thin wrappers over unexported implementations that take an `afero.Fs` and a `LookupEnvFunc` (e.g. `getInput(lookupEnv, …)`, `buildFileCommand(fs, lookupEnv, uuidFunc, …)`, `loadActionContext(fs, lookupEnv)`). Tests use `afero.NewMemMapFs()` and a map-backed lookup; never the real environment or disk.
- **Table-driven tests**, success and error paths both (unset env vars, invalid input, missing files).
- **Doc comments follow [Go Doc Comments](https://go.dev/doc/comment)** and link the GitHub reference for each workflow command. `docs/github-workflow-variables.md` lists the default environment variables `actioncontext` reads.
- **`actioncontext` populates itself in `init()`** and logs the context as a `::debug::` line. Import it only from action binaries, never from `cmd/walle` (it would print on every walle run and link go-github into the task runner); the repository's own actions use `core` only.
- **go-github is part of the public API** (`EventPayload` and the typed accessors return `github.com/google/go-github/v92` types). Bumping its major version is a breaking change; `RawEventPayload` is the escape hatch for consumers on another major.
- **`cmd/example` is the library's release component.** It imports both `core` and `actioncontext`, which puts the library in its build closure, so a change to either package is in CI's vet, test and build scope. Keep both imports.

## The idea in one paragraph

A Go project always has the Go toolchain, so its tasks can be a Go program run with `go run` — no Make, no task-runner install, no bash portability bugs, and the tasks are typed, testable and debuggable like the product. CI needs the same logic (what changed, how to build it), so that logic lives in shared `internal/` packages imported by _both_ the local task runner and small Go GitHub Actions. A laptop and CI then cannot disagree, because they run the same code.

## Layout

```
core/                         library: inputs, outputs, file commands, workflow commands, job summary
actioncontext/                library: run context from env + typed event payload, GitHub client
docs/                         reference notes for the library (default environment variables)
walle                         sh wrapper: cd to the repo root, exec `go run ./cmd/walle "$@"`
cmd/walle/                    the task runner (kong CLI). Thin: parse flags, call internal/, print
cmd/github-actions/<name>/    one main package per Go GitHub Action, built on core. Thin: read inputs, call internal/, set outputs
cmd/example/                  example action binary and the library's release component (see above)
internal/gocmd/               the only place child processes (go, git) are run; owns project-wide go env
internal/buildscope/          component registry + build-closure engine (go list -deps): what changed, what to vet/test
internal/buildrun/            build engine: the one `go build` command line, platforms, metadata, action binaries
internal/cireport/            job-summary markdown for the actions
internal/gocache/             trims the Go build cache to a CI run's working set before it is saved
version/                      Version const (bumped by release-please) + BuildMetadata (set via -ldflags -X)
.github/actions/_shim/shim.js shared node launcher for every Go action
.github/actions/<name>/       action.yaml + one-line index.js (+ optional committed dist/)
.github/workflows/ci.yaml     single-job CI: changed-scope → fmt/tidy → vet/test → build → check actions
.github/workflows/release-please.yaml  release PR, tag and GitHub release on push to main
release-please-config.json    release-please configuration; .release-please-manifest.json records the last released version
```

## Three layers, one direction

1. **Wrapper** (`./walle`) — convenience only. Contains no logic and no environment. CI's actions do not run it, so anything placed in it would silently differ between a laptop and CI.
2. **Entry points** (`cmd/walle`, `cmd/github-actions/*`) — adapters. A walle command parses flags and prints; an action reads `INPUT_*` and writes outputs and a step summary. Neither holds decisions.
3. **Engines** (`internal/*`) — every decision: which components exist, what a change affects, the build flags, the metadata string, how a report renders. Tested without the toolchain where possible (`AffectedIn`, `parseClosure`, `ciMetadata`, `Report.Markdown`).

Dependencies point down only: entry points import engines; engines never import entry points; walle never shells out to an action and an action never shells out to `./walle`.

## Why actions import `internal/` instead of running `./walle`

Running `./walle` in every CI job means `go run` compiles the task runner first in every job, and that cost grows with everything walle links. Go actions that import the same packages avoid that cost and keep the boundary typed (a function call returning an error, not an argv and an exit code). CI still runs `./walle` for cheap, non-logic steps (fmt, tidy, vet, test) — which is only fast because **walle is kept light** (see rules).

## How a Go action runs

`action.yaml` declares `runs: using: node24, main: index.js`. `index.js` is one line that calls the shared shim with its own directory; the action name is the directory name. The shim then:

- **binary mode** — if `dist/<name>_<os>_<arch>` exists for the runner, exec it. No compile in CI at all.
- **source mode** — otherwise, `go build ./cmd/github-actions/<name>` into `RUNNER_TEMP` and exec that. Needs `actions/setup-go` earlier in the job. Costs one compile of the action's closure, mostly served from the restored Go build cache.

This repository ships in **source mode** (no binaries). Trade-off:

|                                                                  | Source mode (default)                             | Binary mode                                           |
| ---------------------------------------------------------------- | ------------------------------------------------- | ----------------------------------------------------- |
| CI cost per action                                               | one cached compile (seconds)                      | none                                                  |
| Repository growth                                                | none                                              | ~5 MB per binary per platform, again on every rebuild |
| Drift risk                                                       | none — always built from the checked-out source | guarded by `./walle check actions` in CI              |
| Works without setup-go                                           | no                                                | yes                                                   |
| Usable from other repos (`uses: org/repo/.github/actions/x@ref`) | needs Go on the caller's runner                   | yes                                                   |

Switch an action to binary mode with `./walle build actions` and commit `dist/`. `.gitattributes` marks `dist/**` binary and generated. From then on CI's `./walle check actions` rebuilds each committed binary — with the Go toolchain recorded inside it — and fails if the bytes differ, so a stale binary cannot ship.

## Reproducibility is load-bearing

Every artifact must be a pure function of source, toolchain version and the metadata string. `internal/buildrun.goBuildEnv` is the single place a `go build` command line is assembled: `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false`, `-ldflags "-s -w [-X version.BuildMetadata=…]"`. Never add a timestamp, run id or host path. `./walle check determinism` builds each component twice and compares digests; `./walle check actions` compares across machines. Keep both passing.

Build metadata comes from `buildrun.Metadata`: `BUILD_METADATA` if set, else `pr<N>.<sha7>` / `main.<sha7>` / `branch.<sha7>` in Actions, else `local.<sha7>[.dirty]`. Note that a SHA in the metadata changes every binary on every commit; if a deploy tool dedupes by asset hash, pass a constant `metadata` for those artifacts.

## Rules

- **CI logic is Go.** A workflow step may call `./walle <cmd>` or a Go action. Anything beyond a one-line invocation (parsing, branching, loops, string assembly) goes into `internal/` with a test. Pass step outputs to `run:` through `env:`, never `${{ }}` inside the script.
- **Keep walle light.** walle is compiled on demand by `go run` and in CI. Do not import cloud SDKs or private modules into `cmd/walle`. Commands that operate deployed infrastructure belong in a separate binary (e.g. `cmd/ops`) or a release component. `go list -deps ./cmd/walle | wc -l` is the budget to watch.
- **One registry per concept.** Components live only in `buildscope.Components`; CLI enums and help derive from it (`${components}`). Actions are discovered from `cmd/github-actions/*`; there is no list. Child-process env lives only in `gocmd.Env` (this is where e.g. `GOEXPERIMENT` or `GOPRIVATE` goes, once, rather than in the wrapper, workflows and helpers separately).
- **Checks never mutate.** `./walle check`, `fmt --check`, `tidy --check` fail with a fix command; they do not rewrite the tree.
- **Child processes go through `internal/gocmd`** so they inherit `gocmd.Env` and the cancellation context (Ctrl-C stops the child).
- **Paths are repo-root-relative.** The wrapper `cd`s to the repo root, and actions run in the workspace root.
- **The library and the tooling are in scope.** `example` is registered as a component so a change to `core` or `actioncontext` flags it, and `walle` is registered so a change to `cmd/walle` or any `internal/` engine flags it; CI then vets and tests those packages. The action mains themselves are compiled by the shim on every CI run; their own small tests run in the full-module `./walle check`.
- **The CI cache is seeded from main, trimmed, and keyed on what determines it.** main runs every component and saves `go-<os>-<arch>-<go version>-<hash(go.sum)>-<run_id>` after `./walle cache trim`; PRs only restore. Do not scope main runs to changed components: the trim would discard the other components' compile output.
- Run `./walle check` before pushing. It is what CI runs.

## Commits

Use [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/): `<type>(<scope>): <subject>`. The history then says what kind of change each commit is, and release-please derives the version bump and changelog from it (see Releases). Pull requests are usually squash-merged, so the PR title is the commit release-please reads: it must follow these rules too.

- **Types:** `feat` (new capability), `fix` (bug fix), `perf`, `refactor` (no behaviour change), `test`, `docs`, `build` (go.mod, build flags, the wrapper), `ci` (workflows and actions plumbing), `chore` (anything else).
- **Scope:** the package, component or engine touched — `core`, `actioncontext`, `walle`, `buildscope`, `buildrun`, `gocache`, `gocmd`, `cireport`, `actions`, `ci`, or a component name such as `example`. Omit it when a change spans the repository.
- **Subject:** imperative, lower case, no trailing period, the whole line at most 72 characters: `fix(buildscope): include IgnoredGoFiles in closures`.
- **Body:** say why, not what the diff already shows. Wrap nothing; one paragraph per line.
- **Breaking changes:** `!` after the type or scope (`feat(core)!: return an error from Debug`) plus a `BREAKING CHANGE:` footer describing the migration. A changed exported library API (including a go-github major bump), or a renamed walle command, action input or output, is breaking.
- **One logical change per commit.** An engine change and the walle command that exposes it may share a commit; unrelated cleanups do not. Committed action binaries (`dist/`) go in their own `build(actions): rebuild action binaries` commit.

## Releases

[release-please](https://github.com/googleapis/release-please) runs from `.github/workflows/release-please.yaml` ([googleapis/release-please-action](https://github.com/googleapis/release-please-action)) on every push to main:

1. It keeps one release PR open (`chore(main): release X.Y.Z`) that sets `version.Version` in `version/version.go`, the `.` entry in `.release-please-manifest.json`, and prepends the release's section to `CHANGELOG.md`. Each push to main updates it.
2. Merging the release PR makes the next run tag the merge commit `vX.Y.Z` and publish a GitHub release with the changelog section as its notes. The tag is the release: Go resolves `go get github.com/asteroid-computing/actions-toolkit-go@vX.Y.Z` from it.

The version comes from the commits since the last release: `feat` bumps minor, `fix`/`perf`/`revert` bump patch, and the other types (`docs`, `build`, `ci`, `chore`, `refactor`, `test`, `style`) neither trigger a release nor appear in the changelog unless marked breaking. While the version is below 1.0.0, `bump-minor-pre-major` makes a breaking change bump minor rather than jump to 1.0.0; ship 1.0.0 deliberately with `Release-As: 1.0.0` in a commit body. Settings in `release-please-config.json` that are not defaults, and why:

- `release-type: go` with `version-file: version/version.go` — rewrites the `const Version = "X.Y.Z"` line; keep that line's form, or the bump silently stops. `TestVersionMatchesReleaseManifest` fails if the const and the manifest disagree.
- `include-component-in-tag: false` — tags are `vX.Y.Z`, the form Go module versions require (the manifest default would be `actions-toolkit-go-vX.Y.Z`).
- `initial-version: 0.1.0` — the first release, which the empty manifest means has not happened yet. release-please's own default is 1.0.0. It has no effect once the manifest records a version.

Never edit `version.Version`, the manifest or `CHANGELOG.md` by hand, and never push a `v*` tag yourself; release-please would lose track of the last release.

release-please acts as the org's `astrocompute-release-please` GitHub App, not with the built-in `GITHUB_TOKEN`: GitHub does not start workflows for events that `GITHUB_TOKEN` causes, so CI would never run on the release PR. `actions/create-github-app-token` mints an installation token from the org secrets `RELEASE_PLEASE_CLIENT_ID` and `RELEASE_PLEASE_PRIVATE_KEY`, scoped to this repository with contents, pull requests and issues write, and revoked when the job ends. Both secrets are visible only to selected repositories, so a repository using this setup must be added to their repository access list (the App itself is installed on every org repository).

## How to…

**Add to the library.** Put the exported function in `core` or `actioncontext` as a thin wrapper over an unexported, injectable implementation, with table-driven tests against `afero.NewMemMapFs()` and a map-backed `LookupEnvFunc`. Link the GitHub reference in its doc comment. Match the JS toolkit's behaviour where one exists, and say so in the comment when it deliberately differs.

**Add a walle command.** Add a struct with `Run(ctx context.Context) error` in `cmd/walle`, add a field to `CLI` in `main.go` with `cmd:"" group:"<develop|build|inspect|maintain>" help:"…"`. If the logic is something CI also needs, put it in `internal/` and keep the command a shell. Field order within a group is display order.

**Add a component.** Create `cmd/<name>/main.go` (import `version` so build metadata is linked — `TestEveryComponentLinksVersion` enforces it). Add `{Name: "<name>", EntryPkg: "./cmd/<name>", BuildTags: …}` to `buildscope.Components`. walle's enums, `changed`, `closure`, `build`, the changed-scope outputs and the build action all pick it up; no YAML changes.

**Add a Go action.** Create `cmd/github-actions/<name>/main.go` using this repository's `core` package for inputs, outputs, annotations and the step summary (not `actioncontext`, whose `init()` logs; read the few `GITHUB_*` variables an action needs directly); put the logic in `internal/` with tests. Create `.github/actions/<name>/action.yaml` (`using: node24`, `main: index.js`) and `.github/actions/<name>/index.js` containing `require("../_shim/shim.js")(__dirname);`. Prefer component-agnostic outputs (lists, JSON) so the action needs no edit when components change. Optionally `./walle build actions` and commit `dist/`.

**Add a project-wide go setting.** Add it to `gocmd.Env`. Add it to the workflow `env:` only if a step outside walle/actions needs it.

**Run an action locally.**

```bash
touch /tmp/out /tmp/summary
env INPUT_COMPONENTS=example GITHUB_OUTPUT=/tmp/out GITHUB_STEP_SUMMARY=/tmp/summary node .github/actions/build/index.js
```

`core.SetOutput` appends to `GITHUB_OUTPUT`, which must exist. The example component runs the same way:

```bash
./walle build example && env INPUT_WHO-TO-GREET=Mona GITHUB_OUTPUT=/tmp/out ./bin/example
```

## Commands

`./walle --help` is authoritative. Summary: `check` (gate: fmt/tidy/vet/race tests; `check determinism`, `check actions`), `test`, `fmt`, `vet`, `tidy`, `build [components] [--all-platforms]`, `build actions`, `changed --base origin/main`, `closure <component>`, `version [--full|--semver]`, `update`, `cache trim --since <unix>` (CI only: trims the build cache to the run's working set before it is saved).
