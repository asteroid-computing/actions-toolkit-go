package buildscope

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
)

func TestAffectedIn(t *testing.T) {
	closures := map[string][]string{"example": {"cmd/example/main.go", "go.mod", "go.sum", "version/version.go"}}
	tests := []struct {
		name    string
		changed []string
		want    bool
	}{
		{"entry file", []string{"cmd/example/main.go"}, true},
		{"shared package", []string{"version/version.go"}, true},
		{"go.sum", []string{"go.sum"}, true},
		{"unrelated", []string{"README.md", ".github/workflows/pull-request.yaml"}, false},
		{"blank and whitespace lines", []string{"", "  ", " cmd/example/main.go "}, true},
		{"nothing", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AffectedIn(tt.changed, closures)["example"]; got != tt.want {
				t.Errorf("AffectedIn(%q)[example] = %v, want %v", tt.changed, got, tt.want)
			}
		})
	}
}

func TestAffectedInReportsEveryComponent(t *testing.T) {
	got := AffectedIn([]string{"x.go"}, nil)
	for _, name := range Names() {
		if _, ok := got[name]; !ok {
			t.Errorf("component %q missing from result", name)
		}
	}
}

func TestParseClosure(t *testing.T) {
	root := filepath.FromSlash("/src/m")
	mod := moduleInfo{Path: "example.com/m", Dir: root}
	out := "P example.com/m/cmd/a\n" +
		"F " + filepath.Join(root, "cmd/a/main.go") + "\n" +
		"F " + filepath.Join(root, "cmd/a/main_arm64.go") + "\n" +
		"P fmt\n" +
		"F /usr/local/go/src/fmt/print.go\n" +
		"P example.com/mx\n" // a sibling path prefix, not this module
	got := parseClosure(out, mod)
	if want := []string{"example.com/m/cmd/a"}; !slices.Equal(got.Packages, want) {
		t.Errorf("Packages = %q, want %q", got.Packages, want)
	}
	if want := []string{"cmd/a/main.go", "cmd/a/main_arm64.go", "go.mod", "go.sum"}; !slices.Equal(got.Files, want) {
		t.Errorf("Files = %q, want %q", got.Files, want)
	}
}

// TestEveryComponentLinksVersion guards the -X link-time injection: a component that does not import the version package silently gets no build metadata.
func TestEveryComponentLinksVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("runs go list")
	}
	closures, err := Closures(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for name, c := range closures {
		if !slices.Contains(c.Files, "version/version.go") {
			t.Errorf("component %s does not import the version package; its build metadata would be silently dropped", name)
		}
	}
}
