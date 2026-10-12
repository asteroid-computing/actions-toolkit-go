package actioncontext

import (
	"fmt"
	"net/url"

	"github.com/google/go-github/v92/github"
)

const publicGitHubAPIHost = "api.github.com"

// NewGitHubClient creates a [github.Client] targeting the API host derived from
// the context's APIURL field (sourced from GITHUB_API_URL).
//
// When APIURL is empty or resolves to the public GitHub host (api.github.com),
// the client connects to the public API without any enterprise URL configuration.
//
// When APIURL points to a GitHub Enterprise Server host, WithEnterpriseURLs is
// prepended to opts so the client targets the GHES instance.
// The upload URL is derived from the same scheme and host as APIURL.
//
// Caller-supplied opts are applied after any derived enterprise option,
// so callers can override anything — including auth, timeout, or URLs.
// To authenticate, pass [github.WithAuthToken].
func (ac *ActionContext) NewGitHubClient(opts ...github.ClientOptionsFunc) (*github.Client, error) {
	if ac.APIURL == "" || isPublicGitHubAPI(ac.APIURL) {
		client, err := github.NewClient(opts...)
		if err != nil {
			return nil, fmt.Errorf("creating GitHub client: %w", err)
		}
		return client, nil
	}

	enterpriseOpt, err := enterpriseOption(ac.APIURL)
	if err != nil {
		return nil, err
	}

	allOpts := make([]github.ClientOptionsFunc, 0, 1+len(opts))
	allOpts = append(allOpts, enterpriseOpt)
	allOpts = append(allOpts, opts...)

	client, err := github.NewClient(allOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating GitHub Enterprise client: %w", err)
	}
	return client, nil
}

// isPublicGitHubAPI reports whether rawURL targets the public GitHub API host.
func isPublicGitHubAPI(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Hostname() == publicGitHubAPIHost || u.Hostname() == "github.com"
}

// enterpriseOption builds a WithEnterpriseURLs option from the GHES API URL.
// The upload URL reuses the same scheme and host with the path /api/uploads/.
func enterpriseOption(rawURL string) (github.ClientOptionsFunc, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing GITHUB_API_URL %q: %w", rawURL, err)
	}

	baseURL := rawURL
	uploadURL := u.Scheme + "://" + u.Host + "/api/uploads/"

	return github.WithEnterpriseURLs(baseURL, uploadURL), nil
}
