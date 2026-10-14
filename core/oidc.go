package core

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// ErrIDTokenUnavailable is returned by [GetIDToken] when the runner has not provided the OIDC request credentials,
// which happens when the job lacks the id-token: write permission, on pull requests from forks, and outside GitHub Actions.
// Use errors.Is to test for this sentinel.
var ErrIDTokenUnavailable = errors.New("OIDC ID token unavailable: ACTIONS_ID_TOKEN_REQUEST_URL or ACTIONS_ID_TOKEN_REQUEST_TOKEN is not set; grant the job `permissions: id-token: write`")

const (
	idTokenRequestURLEnv   = "ACTIONS_ID_TOKEN_REQUEST_URL"
	idTokenRequestTokenEnv = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"

	idTokenAttempts     = 3
	idTokenRetryDelay   = time.Second
	idTokenMaxBodyBytes = 1 << 20
)

// idTokenClient has a timeout, unlike http.DefaultClient. Its nil Transport means http.DefaultTransport, which honours HTTPS_PROXY and NO_PROXY on self-hosted runners.
var idTokenClient = &http.Client{Timeout: 30 * time.Second}

// GetIDToken requests a GitHub OIDC ID token (a JWT) for audience, for exchanging with a cloud provider or another service that trusts GitHub's OIDC issuer.
// An empty audience requests GitHub's default, which is the URL of the repository owner.
// The token is masked in the log (as with [MaskValue]) before it is returned.
//
// The job needs the id-token: write permission:
//
//	permissions:
//	  id-token: write
//	  contents: read
//
// Without it, on pull requests from forks, and outside GitHub Actions, the runner provides no request credentials and GetIDToken returns [ErrIDTokenUnavailable].
//
// Differences from the JS toolkit's getIDToken: a request that fails with 502, 503 or 504 is retried at most twice, with a one- then two-second delay, where JS makes up to ten attempts;
// and if masking the token fails, GetIDToken returns that error rather than the unmasked token.
//
// GitHub Reference: https://docs.github.com/en/actions/reference/security/oidc
func GetIDToken(ctx context.Context, audience string) (string, error) {
	return getIDToken(ctx, os.LookupEnv, idTokenClient, os.Stdout, idTokenRetryDelay, audience)
}

// getIDToken is the testable implementation of [GetIDToken]. The delay before retry n (1-based) is n*retryDelay.
func getIDToken(ctx context.Context, lookupEnv lookupEnvFunc, client *http.Client, w io.Writer, retryDelay time.Duration, audience string) (string, error) {
	rawURL, _ := lookupEnv(idTokenRequestURLEnv)
	requestToken, _ := lookupEnv(idTokenRequestTokenEnv)
	if rawURL == "" || requestToken == "" {
		return "", ErrIDTokenUnavailable
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parsing %s: %w", idTokenRequestURLEnv, err)
	}
	if audience != "" {
		q := u.Query()
		q.Set("audience", audience)
		u.RawQuery = q.Encode()
	}

	var body []byte
	for attempt := 1; ; attempt++ {
		var status int
		status, body, err = requestIDToken(ctx, client, u.String(), requestToken)
		if err != nil {
			return "", err
		}
		if status == http.StatusOK {
			break
		}
		retryable := status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
		if !retryable || attempt == idTokenAttempts {
			return "", fmt.Errorf("requesting OIDC ID token: %d %s: %s", status, http.StatusText(status), truncate(body, 512))
		}
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("requesting OIDC ID token: %w", ctx.Err())
		case <-time.After(time.Duration(attempt) * retryDelay):
		}
	}

	var resp struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("decoding OIDC ID token response: %w", err)
	}
	if resp.Value == "" {
		return "", errors.New("OIDC ID token response has no value")
	}

	if err := NewWorkflowCommand(w, MaskCommandStr, nil, resp.Value).Send(); err != nil {
		return "", fmt.Errorf("masking OIDC ID token: %w", err)
	}

	return resp.Value, nil
}

// requestIDToken makes one request and returns the status and up to idTokenMaxBodyBytes of the body.
// The returned error never includes the request token: it comes from building the request, from the transport (whose *url.Error carries only the URL) or from reading the body.
func requestIDToken(ctx context.Context, client *http.Client, rawURL, requestToken string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("requesting OIDC ID token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+requestToken)
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("requesting OIDC ID token: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, idTokenMaxBodyBytes))
	if err != nil {
		return 0, nil, fmt.Errorf("reading OIDC ID token response: %w", err)
	}
	// Drain anything past the limit so the connection can be reused for a retry.
	_, _ = io.Copy(io.Discard, res.Body)

	return res.StatusCode, body, nil
}

// truncate returns b as a string, cut to at most n bytes.
func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
