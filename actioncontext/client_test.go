package actioncontext_test

import (
	"testing"

	"github.com/google/go-github/v92/github"

	"github.com/asteroid-computing/actions-toolkit-go/actioncontext"
)

func TestNewGitHubClientURLs(t *testing.T) {
	tests := []struct {
		name       string
		apiURL     string
		wantBase   string
		wantUpload string
	}{
		{
			name:       "empty_apiurl_targets_public",
			apiURL:     "",
			wantBase:   "https://api.github.com/",
			wantUpload: "https://uploads.github.com/",
		},
		{
			name:       "public_apiurl_targets_public",
			apiURL:     "https://api.github.com",
			wantBase:   "https://api.github.com/",
			wantUpload: "https://uploads.github.com/",
		},
		{
			name:       "ghes_apiurl_derives_enterprise_urls",
			apiURL:     "https://ghe.example.com/api/v3",
			wantBase:   "https://ghe.example.com/api/v3/",
			wantUpload: "https://ghe.example.com/api/uploads/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &actioncontext.ActionContext{APIURL: tt.apiURL}
			client, err := ac.NewGitHubClient()
			if err != nil {
				t.Fatalf("NewGitHubClient() unexpected error: %v", err)
			}
			if got := client.BaseURL(); got != tt.wantBase {
				t.Errorf("NewGitHubClient().BaseURL() = %q, want %q", got, tt.wantBase)
			}
			if got := client.UploadURL(); got != tt.wantUpload {
				t.Errorf("NewGitHubClient().UploadURL() = %q, want %q", got, tt.wantUpload)
			}
		})
	}
}

func TestNewGitHubClientCallerOptsOverrideDerived(t *testing.T) {
	// A caller-supplied WithEnterpriseURLs is applied after the derived one,
	// so it wins — proving opts ordering.
	ac := &actioncontext.ActionContext{APIURL: "https://ghe.example.com/api/v3"}

	client, err := ac.NewGitHubClient(
		github.WithEnterpriseURLs("https://other.example.com/api/v3/", "https://other.example.com/api/uploads/"),
	)
	if err != nil {
		t.Fatalf("NewGitHubClient() unexpected error: %v", err)
	}

	const wantBase = "https://other.example.com/api/v3/"
	if got := client.BaseURL(); got != wantBase {
		t.Errorf("caller opt override: BaseURL() = %q, want %q", got, wantBase)
	}
}
