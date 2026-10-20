// Package version carries the project's version.
//
// The semantic version is the [Version] const and changes only in source (a release tool such as release-please bumps it). [BuildMetadata] is the semver "+build" suffix, injected at link time by internal/buildrun via -ldflags "-X". Every component that should report its version must import this package: an -X naming a package the binary does not link is silently discarded.
package version

import "reflect"

// Version is the semantic version, without a leading "v".
const Version = "0.2.0"

// BuildMetadata is set at link time (e.g. "pr45.abc1234", "local.abc1234.dirty"). Empty under `go run`.
var BuildMetadata string

type marker struct{}

// ImportPath is this package's import path — the -X target for [BuildMetadata]. Derived rather than written as a string literal so it survives a module rename (gonew rewrites imports, not string constants).
var ImportPath = reflect.TypeFor[marker]().PkgPath()

// String returns the bare semver, e.g. "0.1.0".
func String() string { return Version }

// Semver returns the version with a leading v, e.g. "v0.1.0" — the tag form.
func Semver() string { return "v" + Version }

// Full returns the version with build metadata when present, e.g. "0.1.0+pr45.abc1234".
func Full(metadata string) string {
	if metadata == "" {
		return Version
	}
	return Version + "+" + metadata
}
