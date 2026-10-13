# Contributing

Thanks for helping improve actions-toolkit-go. Bug reports, fixes and features are all welcome. For anything larger than a small fix, open an issue first so the approach can be agreed before you write the code.

By contributing you agree that your contributions are licensed under the [MIT License](LICENSE), and that you follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## Setup

You need Go 1.27 or later, and nothing else: the developer tasks are a Go program behind the `./walle` wrapper.

```bash
./walle --help
./walle check
```

`./walle check` runs what CI runs: format and tidy checks, `go vet`, and the tests with the race detector. Run it before you push.

## Making a change

- Read [AGENTS.md](AGENTS.md). It describes the layout, the library's conventions (injectable filesystem and environment lookups behind each exported function, table-driven tests) and the tooling rules.
- Add or update tests with every change. Library code is tested without touching the real environment or disk.
- Exported identifiers need [doc comments](https://go.dev/doc/comment), linking the GitHub reference for any workflow command they implement.
- Keep a pull request to one logical change.

## Pull request titles

Pull requests are squash-merged, so the PR title becomes the commit on main, and [release-please](https://github.com/googleapis/release-please) builds the version bump and changelog from it. The title must follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

```
<type>(<scope>): <subject>
```

- `feat` adds a capability (minor release) and `fix` fixes a bug (patch release). `docs`, `test`, `refactor`, `build`, `ci` and `chore` do not trigger a release.
- The scope is the package or area touched, e.g. `core`, `actioncontext`, `walle`, `ci`.
- Mark a breaking change with `!` after the type or scope (`feat(core)!: …`) and describe the migration in a `BREAKING CHANGE:` footer in the PR description.

Do not edit `version/version.go`, `.release-please-manifest.json` or `CHANGELOG.md`; release-please maintains them.

## Reporting security issues

Do not open a public issue for a vulnerability. Follow [SECURITY.md](SECURITY.md).
