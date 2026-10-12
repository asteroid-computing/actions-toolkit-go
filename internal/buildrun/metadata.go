package buildrun

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/asteroid-computing/actions-toolkit-go/internal/gocmd"
)

// Metadata returns the build-metadata string (the semver "+build" suffix) for this invocation, in order of precedence:
//
//  1. BUILD_METADATA, when set — an explicit override, e.g. "release.abc1234";
//  2. in GitHub Actions: "pr<N>.<sha7>" for a pull request (the PR head, not the synthetic merge commit), "main.<sha7>" on main, "branch.<sha7>" otherwise;
//  3. locally: "local.<sha7>", with ".dirty" for an uncommitted tree, or "local.unknown" outside git.
//
// It is one implementation for walle and the build action, replacing fenton's separate bash build-metadata action and Go localMetadata. It never contains a timestamp or run id: rebuilding the same commit must produce the same bytes.
func Metadata(ctx context.Context) string {
	if m := os.Getenv("BUILD_METADATA"); m != "" {
		return m
	}
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return ciMetadata(os.Getenv)
	}
	return localMetadata(ctx)
}

func ciMetadata(getenv func(string) string) string {
	sha := getenv("GITHUB_SHA")
	if getenv("GITHUB_EVENT_NAME") == "pull_request" || getenv("GITHUB_EVENT_NAME") == "pull_request_target" {
		if n, head, ok := pullRequest(getenv("GITHUB_EVENT_PATH")); ok {
			return "pr" + n + "." + short(head)
		}
	}
	if getenv("GITHUB_REF_NAME") == "main" {
		return "main." + short(sha)
	}
	return "branch." + short(sha)
}

// pullRequest reads the PR number and head SHA from the event payload.
func pullRequest(eventPath string) (number, headSHA string, ok bool) {
	raw, err := os.ReadFile(eventPath)
	if err != nil {
		return "", "", false
	}
	var ev struct {
		PullRequest struct {
			Number json.Number `json:"number"`
			Head   struct {
				SHA string `json:"sha"`
			} `json:"head"`
		} `json:"pull_request"`
	}
	if json.Unmarshal(raw, &ev) != nil || ev.PullRequest.Number == "" || ev.PullRequest.Head.SHA == "" {
		return "", "", false
	}
	return ev.PullRequest.Number.String(), ev.PullRequest.Head.SHA, true
}

func localMetadata(ctx context.Context) string {
	sha, err := gocmd.Output(ctx, nil, "git", "rev-parse", "--short=7", "HEAD")
	if sha = strings.TrimSpace(sha); err != nil || sha == "" {
		return "local.unknown"
	}
	if dirty, err := gocmd.Output(ctx, nil, "git", "status", "--porcelain"); err == nil && strings.TrimSpace(dirty) != "" {
		return "local." + sha + ".dirty"
	}
	return "local." + sha
}

func short(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	if sha == "" {
		return "unknown"
	}
	return sha
}
