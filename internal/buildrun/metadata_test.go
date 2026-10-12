package buildrun

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCIMetadata(t *testing.T) {
	event := filepath.Join(t.TempDir(), "event.json")
	if err := os.WriteFile(event, []byte(`{"pull_request":{"number":45,"head":{"sha":"abc1234def"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"pull request uses head sha", map[string]string{"GITHUB_EVENT_NAME": "pull_request", "GITHUB_EVENT_PATH": event, "GITHUB_SHA": "fffffffmerge"}, "pr45.abc1234"},
		{"pull request without payload falls back", map[string]string{"GITHUB_EVENT_NAME": "pull_request", "GITHUB_SHA": "1234567890"}, "branch.1234567"},
		{"main", map[string]string{"GITHUB_EVENT_NAME": "push", "GITHUB_REF_NAME": "main", "GITHUB_SHA": "1234567890"}, "main.1234567"},
		{"other branch", map[string]string{"GITHUB_EVENT_NAME": "push", "GITHUB_REF_NAME": "feat", "GITHUB_SHA": "1234567890"}, "branch.1234567"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ciMetadata(func(k string) string { return tt.env[k] }); got != tt.want {
				t.Errorf("ciMetadata = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseActionPlatform(t *testing.T) {
	p, ok := parseActionPlatform("changed-scope", "changed-scope_linux_arm64")
	if !ok || p != (Platform{"linux", "arm64"}) {
		t.Errorf("got %v %v", p, ok)
	}
	if _, ok := parseActionPlatform("build", "changed-scope_linux_arm64"); ok {
		t.Error("matched another action's binary")
	}
}
