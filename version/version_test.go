package version

import (
	"encoding/json"
	"os"
	"testing"
)

// TestVersionMatchesReleaseManifest guards the release-please wiring: release-please rewrites the `const Version = "X.Y.Z"` line in this package and the manifest entry in the same release PR. If the const is edited by hand, or its line no longer matches release-please's pattern, the two drift and this fails. Before the first release the manifest is empty and there is nothing to compare.
func TestVersionMatchesReleaseManifest(t *testing.T) {
	raw, err := os.ReadFile("../.release-please-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]string
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	released, ok := manifest["."]
	if !ok {
		t.Skip("no release yet")
	}
	if released != Version {
		t.Errorf("version.Version = %q, but .release-please-manifest.json records %q", Version, released)
	}
}
