package core

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

const testRequestToken = "request-token-secret"

// idTokenServer serves the OIDC token endpoint, replying with statuses in order (repeating the last) and the body {"value":"<value>"} on 200.
// It records each request so tests can check what was sent.
type idTokenServer struct {
	*httptest.Server
	statuses []int
	body     string
	calls    atomic.Int32
	requests []*http.Request
}

func newIDTokenServer(t *testing.T, body string, statuses ...int) *idTokenServer {
	t.Helper()
	s := &idTokenServer{statuses: statuses, body: body}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(s.calls.Add(1))
		s.requests = append(s.requests, r.Clone(context.Background()))
		status := s.statuses[min(n, len(s.statuses))-1]
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(s.body))
		} else {
			_, _ = w.Write([]byte("upstream says no"))
		}
	}))
	t.Cleanup(s.Close)
	return s
}

// env returns the runner environment pointing at the server, with the query string the real request URL always carries.
func (s *idTokenServer) env() map[string]string {
	return map[string]string{
		idTokenRequestURLEnv:   s.URL + "/idtoken?api-version=2.0",
		idTokenRequestTokenEnv: testRequestToken,
	}
}

func TestGetIDToken(t *testing.T) {
	s := newIDTokenServer(t, `{"value":"jwt.token.value","count":1}`, http.StatusOK)
	var log bytes.Buffer

	got, err := getIDToken(context.Background(), fakeLookup(s.env()), s.Client(), &log, 0, "api://AzureADTokenExchange")
	if err != nil {
		t.Fatalf("getIDToken returned error: %v", err)
	}
	if got != "jwt.token.value" {
		t.Errorf("token = %q, want %q", got, "jwt.token.value")
	}
	if want := "::add-mask::jwt.token.value\n"; log.String() != want {
		t.Errorf("log = %q, want %q", log.String(), want)
	}

	r := s.requests[0]
	if r.Method != http.MethodGet || r.URL.Path != "/idtoken" {
		t.Errorf("request = %s %s, want GET /idtoken", r.Method, r.URL.Path)
	}
	if got, want := r.Header.Get("Authorization"), "Bearer "+testRequestToken; got != want {
		t.Errorf("Authorization = %q, want %q", got, want)
	}
	if got := r.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q, want application/json", got)
	}
	q := r.URL.Query()
	if got := q.Get("api-version"); got != "2.0" {
		t.Errorf("api-version = %q, want 2.0 (the existing query must be kept)", got)
	}
	if got := q.Get("audience"); got != "api://AzureADTokenExchange" {
		t.Errorf("audience = %q, want api://AzureADTokenExchange", got)
	}
}

func TestGetIDTokenDefaultAudience(t *testing.T) {
	s := newIDTokenServer(t, `{"value":"jwt"}`, http.StatusOK)

	if _, err := getIDToken(context.Background(), fakeLookup(s.env()), s.Client(), &bytes.Buffer{}, 0, ""); err != nil {
		t.Fatalf("getIDToken returned error: %v", err)
	}
	if q := s.requests[0].URL.Query(); q.Has("audience") || q.Get("api-version") != "2.0" {
		t.Errorf("query = %v, want only api-version=2.0", q)
	}
}

func TestGetIDTokenRetries(t *testing.T) {
	tests := []struct {
		name      string
		statuses  []int
		wantCalls int32
		wantErr   bool
	}{
		{name: "503 then success", statuses: []int{http.StatusServiceUnavailable, http.StatusOK}, wantCalls: 2},
		{name: "502 and 504 then success", statuses: []int{http.StatusBadGateway, http.StatusGatewayTimeout, http.StatusOK}, wantCalls: 3},
		{name: "gives up after three attempts", statuses: []int{http.StatusServiceUnavailable}, wantCalls: 3, wantErr: true},
		{name: "401 is not retried", statuses: []int{http.StatusUnauthorized}, wantCalls: 1, wantErr: true},
		{name: "500 is not retried", statuses: []int{http.StatusInternalServerError}, wantCalls: 1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newIDTokenServer(t, `{"value":"jwt"}`, tt.statuses...)
			var log bytes.Buffer

			_, err := getIDToken(context.Background(), fakeLookup(s.env()), s.Client(), &log, 0, "aud")
			if (err != nil) != tt.wantErr {
				t.Fatalf("getIDToken error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := s.calls.Load(); got != tt.wantCalls {
				t.Errorf("requests = %d, want %d", got, tt.wantCalls)
			}
			if tt.wantErr {
				if log.Len() != 0 {
					t.Errorf("log = %q, want nothing written on failure", log.String())
				}
				if !strings.Contains(err.Error(), "upstream says no") {
					t.Errorf("error %q does not include the response body", err)
				}
			}
		})
	}
}

func TestGetIDTokenCancelledContextStopsRetrying(t *testing.T) {
	s := newIDTokenServer(t, `{"value":"jwt"}`, http.StatusServiceUnavailable)
	ctx, cancel := context.WithCancel(context.Background())
	s.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.calls.Add(1)
		cancel()
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	// A long retry delay: only cancellation can end the wait.
	_, err := getIDToken(ctx, fakeLookup(s.env()), s.Client(), &bytes.Buffer{}, 1<<62, "aud")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("getIDToken error = %v, want context.Canceled", err)
	}
	if got := s.calls.Load(); got != 1 {
		t.Errorf("requests = %d, want 1", got)
	}
}

func TestGetIDTokenBadResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "no value field", body: `{"count":1}`},
		{name: "empty value", body: `{"value":""}`},
		{name: "not JSON", body: `<html>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newIDTokenServer(t, tt.body, http.StatusOK)
			var log bytes.Buffer
			if _, err := getIDToken(context.Background(), fakeLookup(s.env()), s.Client(), &log, 0, "aud"); err == nil {
				t.Fatal("expected an error")
			}
			if log.Len() != 0 {
				t.Errorf("log = %q, want nothing written", log.String())
			}
		})
	}
}

func TestGetIDTokenUnavailable(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{name: "neither set", env: map[string]string{}},
		{name: "URL missing", env: map[string]string{idTokenRequestTokenEnv: testRequestToken}},
		{name: "token missing", env: map[string]string{idTokenRequestURLEnv: "https://example.invalid/idtoken?api-version=2.0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getIDToken(context.Background(), fakeLookup(tt.env), http.DefaultClient, &bytes.Buffer{}, 0, "aud")
			if !errors.Is(err, ErrIDTokenUnavailable) {
				t.Fatalf("getIDToken error = %v, want ErrIDTokenUnavailable", err)
			}
		})
	}
}

func TestGetIDTokenMaskFailure(t *testing.T) {
	s := newIDTokenServer(t, `{"value":"jwt"}`, http.StatusOK)
	got, err := getIDToken(context.Background(), fakeLookup(s.env()), s.Client(), failingWriter{}, 0, "aud")
	if err == nil || got != "" {
		t.Fatalf("getIDToken = %q, %v; want no token and an error when masking fails", got, err)
	}
}

// TestGetIDTokenErrorsNeverIncludeTheRequestToken covers each failure path that reaches the network.
func TestGetIDTokenErrorsNeverIncludeTheRequestToken(t *testing.T) {
	unavailable := httptest.NewServer(http.NotFoundHandler())
	unavailable.Close() // connection refused
	for name, env := range map[string]map[string]string{
		"transport error": {idTokenRequestURLEnv: unavailable.URL + "/idtoken?api-version=2.0", idTokenRequestTokenEnv: testRequestToken},
		"error status":    newIDTokenServer(t, "", http.StatusForbidden).env(),
		"bad body":        newIDTokenServer(t, "{", http.StatusOK).env(),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := getIDToken(context.Background(), fakeLookup(env), http.DefaultClient, &bytes.Buffer{}, 0, "aud")
			if err == nil {
				t.Fatal("expected an error")
			}
			if strings.Contains(err.Error(), testRequestToken) {
				t.Errorf("error leaks the request token: %v", err)
			}
		})
	}
}
