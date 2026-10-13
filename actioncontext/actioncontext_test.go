package actioncontext

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-github/v92/github"
	"github.com/spf13/afero"
)

func TestLookupToStr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		isSet  bool
		expect string
	}{
		{name: "value set", input: "foo", isSet: true, expect: "foo"},
		{name: "value not set", input: "", isSet: false, expect: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := lookupToStr(tt.input, tt.isSet)
			if diff := cmp.Diff(tt.expect, got); diff != "" {
				t.Fatalf("lookupToStr() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLookupToInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		isSet  bool
		expect int
	}{
		{name: "valid int", input: "42", isSet: true, expect: 42},
		{name: "invalid int", input: "x", isSet: true, expect: 0},
		{name: "value not set", input: "", isSet: false, expect: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := lookupToInt(tt.input, tt.isSet)
			if diff := cmp.Diff(tt.expect, got); diff != "" {
				t.Fatalf("lookupToInt() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLookupToBoolPtr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		isSet  bool
		expect *bool
	}{
		{name: "true", input: "true", isSet: true, expect: new(true)},
		{name: "false", input: "false", isSet: true, expect: new(false)},
		{name: "invalid bool", input: "yes", isSet: true, expect: nil},
		{name: "value not set", input: "", isSet: false, expect: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := lookupToBoolPtr(tt.input, tt.isSet)
			if diff := cmp.Diff(tt.expect, got); diff != "" {
				t.Fatalf("lookupToBoolPtr() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLoadPopulatesValuesAndEventPayload(t *testing.T) {
	fs := afero.NewMemMapFs()

	const payloadPath = "/event.json"
	err := afero.WriteFile(fs, payloadPath, []byte(`{"ref":"refs/heads/main"}`), 0o644)
	if err != nil {
		t.Fatalf("failed to write payload file: %v", err)
	}

	env := map[string]string{
		"GITHUB_ACTION":            "build",
		"GITHUB_ACTION_REPOSITORY": "owner/action",
		"GITHUB_ACTOR":             "octocat",
		"GITHUB_ACTIONS":           "true",
		"GITHUB_RUN_ID":            "123",
		"GITHUB_RUN_ATTEMPT":       "2",
		"GITHUB_REF_PROTECTED":     "false",
		"GITHUB_EVENT_NAME":        "push",
		"GITHUB_EVENT_PATH":        payloadPath,
		"RUNNER_DEBUG":             "1",
	}

	ac, err := load(fs, mockLookup(env))
	if err != nil {
		t.Fatalf("load returned error: %v", err)
	}

	want := &ActionContext{
		Action:               "build",
		ActionRepository:     "owner/action",
		Actor:                "octocat",
		Actions:              new(true),
		RunID:                123,
		RunAttempt:           2,
		RefProtected:         new(false),
		EventName:            "push",
		EventPayloadFilePath: payloadPath,
		RunnerDebug:          new(true),
		EventPayload:         &github.PushEvent{Ref: new("refs/heads/main")},
	}
	if diff := cmp.Diff(want, ac, cmpopts.IgnoreUnexported(ActionContext{})); diff != "" {
		t.Fatalf("load mismatch (-want +got):\n%s", diff)
	}
}

func TestLoadOutsideActions(t *testing.T) {
	ac, err := load(afero.NewMemMapFs(), mockLookup(map[string]string{}))
	if err != nil {
		t.Fatalf("load returned error: %v", err)
	}
	if diff := cmp.Diff(&ActionContext{}, ac, cmpopts.IgnoreUnexported(ActionContext{})); diff != "" {
		t.Fatalf("expected an empty context (-want +got):\n%s", diff)
	}
	if ac.EventPayload != nil {
		t.Fatalf("expected nil payload without an event file, got %T", ac.EventPayload)
	}
}

func TestLoadPayloadErrors(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "/invalid.json", []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("failed to write payload file: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{name: "invalid JSON", path: "/invalid.json"},
		{name: "missing file", path: "/missing.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac, err := load(fs, mockLookup(map[string]string{"GITHUB_EVENT_NAME": "push", "GITHUB_EVENT_PATH": tt.path}))
			if err == nil {
				t.Fatalf("expected an error, got context %v", ac)
			}
			if ac != nil {
				t.Fatalf("expected nil context on error, got %v", ac)
			}
		})
	}
}

func TestLoadUnknownEventDecodesToMap(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "/event.json", []byte(`{"answer":42}`), 0o644); err != nil {
		t.Fatalf("failed to write payload file: %v", err)
	}

	ac, err := load(fs, mockLookup(map[string]string{"GITHUB_EVENT_NAME": "some_new_event", "GITHUB_EVENT_PATH": "/event.json"}))
	if err != nil {
		t.Fatalf("load returned error: %v", err)
	}
	m, ok := ac.EventPayload.(*map[string]any)
	if !ok {
		t.Fatalf("expected *map[string]any payload, got %T", ac.EventPayload)
	}
	if diff := cmp.Diff(map[string]any{"answer": float64(42)}, *m); diff != "" {
		t.Fatalf("payload mismatch (-want +got):\n%s", diff)
	}
}

// TestLoadPullRequestPayload decodes a realistic payload: nested objects, a go-github Timestamp (custom unmarshaler) and members no struct field declares, which must be ignored rather than rejected.
func TestLoadPullRequestPayload(t *testing.T) {
	fs := afero.NewMemMapFs()

	const payloadPath = "/event.json"
	const raw = `{"action":"opened","number":45,"pull_request":{"number":45,"created_at":"2026-10-12T09:30:00Z","head":{"ref":"feat/x","sha":"abc1234def"},"base":{"ref":"main"},"not_a_go_github_field":true},"repository":{"full_name":"owner/repo"}}`
	if err := afero.WriteFile(fs, payloadPath, []byte(raw), 0o644); err != nil {
		t.Fatalf("failed to write payload file: %v", err)
	}

	env := map[string]string{
		"GITHUB_EVENT_NAME": "pull_request",
		"GITHUB_EVENT_PATH": payloadPath,
	}

	ac, err := load(fs, mockLookup(env))
	if err != nil {
		t.Fatalf("load returned error: %v", err)
	}

	pr, ok := ac.PullRequestEvent()
	if !ok {
		t.Fatalf("expected PullRequestEvent payload, got %T", ac.EventPayload)
	}
	got := []string{pr.GetAction(), pr.GetPullRequest().GetHead().GetSHA(), pr.GetPullRequest().GetBase().GetRef(), pr.GetRepo().GetFullName(), pr.GetPullRequest().GetCreatedAt().UTC().Format("2006-01-02T15:04:05Z")}
	want := []string{"opened", "abc1234def", "main", "owner/repo", "2026-10-12T09:30:00Z"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("payload mismatch (-want +got):\n%s", diff)
	}
	if pr.GetNumber() != 45 {
		t.Fatalf("expected number 45, got %d", pr.GetNumber())
	}
}

func TestRawEventPayload(t *testing.T) {
	fs := afero.NewMemMapFs()

	const payloadPath = "/event.json"
	const raw = `{"ref":"refs/heads/main"}`
	err := afero.WriteFile(fs, payloadPath, []byte(raw), 0o644)
	if err != nil {
		t.Fatalf("failed to write payload file: %v", err)
	}

	env := map[string]string{
		"GITHUB_EVENT_NAME": "push",
		"GITHUB_EVENT_PATH": payloadPath,
	}

	ac, err := load(fs, mockLookup(env))
	if err != nil {
		t.Fatalf("load returned error: %v", err)
	}

	if got := string(ac.RawEventPayload()); got != raw {
		t.Errorf("RawEventPayload() = %q, want %q", got, raw)
	}
}

func TestRawEventPayloadNoEventFile(t *testing.T) {
	ac, err := load(afero.NewMemMapFs(), mockLookup(map[string]string{}))
	if err != nil {
		t.Fatalf("load returned error: %v", err)
	}

	if got := ac.RawEventPayload(); got != nil {
		t.Errorf("RawEventPayload() = %q, want nil", got)
	}
}

func mockLookup(values map[string]string) lookupEnvFunc {
	return func(key string) (string, bool) {
		v, ok := values[key]
		return v, ok
	}
}

func TestActionContextString_NilContext(t *testing.T) {
	var ac *ActionContext
	got := ac.String()

	if got != "ActionContext: <nil>" {
		t.Fatalf("expected \"ActionContext: <nil>\", got %q", got)
	}
}

func TestActionContextString_EmptyContext(t *testing.T) {
	ac := &ActionContext{}
	got := ac.String()

	if !strings.Contains(got, "ActionContext:") {
		t.Fatalf("expected output to contain \"ActionContext:\", got %q", got)
	}

	expectedFields := []string{
		"GithubCI:",
		"Action:",
		"Actor:",
		"Repository:",
		"RunID:",
		"EventPayload:",
		"<not set>",
	}

	for _, field := range expectedFields {
		if !strings.Contains(got, field) {
			t.Fatalf("expected output to contain %q, got %q", field, got)
		}
	}
}

func TestActionContextString_PartiallyPopulated(t *testing.T) {
	ac := &ActionContext{
		Actor:      "octocat",
		Repository: "owner/repo",
		RunID:      12345,
		Actions:    new(true),
		RunnerOS:   "Linux",
	}

	got := ac.String()

	if !strings.Contains(got, "Actor:") || !strings.Contains(got, "octocat") {
		t.Fatalf("expected Actor to show \"octocat\", got %q", got)
	}

	if !strings.Contains(got, "Repository:") || !strings.Contains(got, "owner/repo") {
		t.Fatalf("expected Repository to show \"owner/repo\", got %q", got)
	}

	if !strings.Contains(got, "RunID:") || !strings.Contains(got, "12345") {
		t.Fatalf("expected RunID to show \"12345\", got %q", got)
	}

	if !strings.Contains(got, "Actions:") || !strings.Contains(got, "true") {
		t.Fatalf("expected Actions to show \"true\", got %q", got)
	}

	if !strings.Contains(got, "RunnerOS:") || !strings.Contains(got, "Linux") {
		t.Fatalf("expected RunnerOS to show \"Linux\", got %q", got)
	}

	if !strings.Contains(got, "Action:") {
		t.Fatalf("expected Action field to be present, got %q", got)
	}

	if !strings.Contains(got, "<not set>") {
		t.Fatalf("expected output to contain \"<not set>\" for unpopulated fields, got %q", got)
	}
}

func TestActionContextString_FullyPopulated(t *testing.T) {
	ac := &ActionContext{
		GithubCI:             new(true),
		Action:               "build",
		ActionPath:           "/path/to/action",
		ActionRepository:     "owner/action-repo",
		Actions:              new(true),
		Actor:                "octocat",
		ActorID:              "1234567",
		APIURL:               "https://api.github.com",
		BaseRef:              "main",
		EnvFilePath:          "/tmp/env",
		EventName:            "push",
		EventPayloadFilePath: "/tmp/event.json",
		GraphQLURL:           "https://api.github.com/graphql",
		HeadRef:              "feature-branch",
		JobID:                "job-123",
		OutputFilePath:       "/tmp/output",
		PathFilePath:         "/tmp/path",
		Ref:                  "refs/heads/main",
		RefName:              "main",
		RefProtected:         new(true),
		RefType:              "branch",
		Repository:           "owner/repo",
		RepositoryID:         "repo-123",
		RepositoryOwner:      "owner",
		RepositoryOwnerID:    "owner-123",
		RetentionDays:        90,
		RunAttempt:           1,
		RunID:                12345,
		RunNumber:            42,
		ServerURL:            "https://github.com",
		Sha:                  "abc123def456",
		StepSummaryFilePath:  "/tmp/summary",
		TriggeringActor:      "octocat",
		Workflow:             "CI",
		WorkflowRef:          "owner/repo/.github/workflows/ci.yml@main",
		WorkflowSha:          "workflow-sha",
		WorkspacePath:        "/workspace",
		RunnerArch:           "X64",
		RunnerDebug:          new(false),
		RunnerEnvironment:    "github-hosted",
		RunnerName:           "runner-1",
		RunnerOS:             "Linux",
		RunnerTempPath:       "/tmp",
		RunnerToolCachePath:  "/tool-cache",
		EventPayload:         &github.PushEvent{},
	}

	got := ac.String()

	if !strings.HasPrefix(got, "ActionContext:\n") {
		t.Fatalf("expected output to start with \"ActionContext:\\n\", got %q", got)
	}

	expectedValues := []string{
		"build",
		"octocat",
		"owner/repo",
		"https://api.github.com",
		"refs/heads/main",
		"main",
		"branch",
		"abc123def456",
		"CI",
		"Linux",
		"X64",
		"github-hosted",
	}

	for _, val := range expectedValues {
		if !strings.Contains(got, val) {
			t.Fatalf("expected output to contain %q, got %q", val, got)
		}
	}

	if !strings.Contains(got, "12345") {
		t.Fatalf("expected RunID 12345, got %q", got)
	}
	if !strings.Contains(got, "42") {
		t.Fatalf("expected RunNumber 42, got %q", got)
	}
	if !strings.Contains(got, "90") {
		t.Fatalf("expected RetentionDays 90, got %q", got)
	}

	if !strings.Contains(got, "true") {
		t.Fatalf("expected bool value true, got %q", got)
	}
	if !strings.Contains(got, "false") {
		t.Fatalf("expected bool value false, got %q", got)
	}

	if !strings.Contains(got, "github.PushEvent") {
		t.Fatalf("expected EventPayload type *github.PushEvent, got %q", got)
	}
}

func TestActionContextString_WithNilEventPayload(t *testing.T) {
	ac := &ActionContext{
		Actor:        "octocat",
		EventPayload: nil,
	}

	got := ac.String()

	if !strings.Contains(got, "EventPayload:") {
		t.Fatalf("expected output to contain \"EventPayload:\", got %q", got)
	}

	lines := strings.Split(got, "\n")
	foundEventPayloadNotSet := false
	for _, line := range lines {
		if strings.Contains(line, "EventPayload:") && strings.Contains(line, "<not set>") {
			foundEventPayloadNotSet = true
			break
		}
	}

	if !foundEventPayloadNotSet {
		t.Fatalf("expected EventPayload to show <not set>, got %q", got)
	}
}
