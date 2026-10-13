// Package actioncontext provides the GitHub Actions runtime context of the current step.
//
// Call [Load] to read it: it reads the default environment variables GitHub Actions sets for every step into an [ActionContext], and decodes the event payload named by GITHUB_EVENT_NAME and GITHUB_EVENT_PATH into [ActionContext.EventPayload], available through typed accessors such as [ActionContext.PullRequestEvent]. Importing the package has no side effects.
//
// Variable reference: https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables
package actioncontext

import (
	stdjson "encoding/json"
	"encoding/json/v2"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v92/github"
	"github.com/spf13/afero"
)

// ActionContext is the GitHub Actions runtime context of the current step: the default environment variables and the decoded event payload. Use [Load] to populate one.
//
// A string field is empty and a *bool field nil when its variable is unset; an int field is 0 when its variable is unset or not an integer.
type ActionContext struct {
	// GithubCI is always true on GitHub Actions. Environment variable: CI.
	GithubCI *bool `json:"githubCI,omitzero"`
	// Action is the name of the running action, or the id of a step. Environment variable: GITHUB_ACTION.
	Action string `json:"action,omitzero"`
	// ActionPath is the directory of the running action. Set only in composite actions. Environment variable: GITHUB_ACTION_PATH.
	ActionPath string `json:"actionPath,omitzero"`
	// ActionRepository is the owner and repository name of the running action. Environment variable: GITHUB_ACTION_REPOSITORY.
	ActionRepository string `json:"actionRepository,omitzero"`
	// Actions is true when GitHub Actions is running the workflow, which distinguishes a CI run from a local one. Environment variable: GITHUB_ACTIONS.
	Actions *bool `json:"actions,omitzero"`
	// Actor is the name of the person or app that initiated the workflow. Environment variable: GITHUB_ACTOR.
	Actor string `json:"actor,omitzero"`
	// ActorID is the account ID of the person or app that triggered the initial workflow run. Environment variable: GITHUB_ACTOR_ID.
	ActorID string `json:"actorId,omitzero"`
	// APIURL is the GitHub REST API base URL; [ActionContext.NewGitHubClient] targets it. Environment variable: GITHUB_API_URL.
	APIURL string `json:"apiUrl,omitzero"`
	// BaseRef is the target branch of the pull request. Set only for pull_request and pull_request_target. Environment variable: GITHUB_BASE_REF.
	BaseRef string `json:"baseRef,omitzero"`
	// EnvFilePath is the step's file for setting environment variables (see core.SetEnv). Environment variable: GITHUB_ENV.
	EnvFilePath string `json:"envFilePath,omitzero"`
	// EventName is the name of the event that triggered the workflow; it selects the type of [ActionContext.EventPayload]. Environment variable: GITHUB_EVENT_NAME.
	EventName string `json:"eventName,omitzero"`
	// EventPayloadFilePath is the file holding the event's full webhook payload. Environment variable: GITHUB_EVENT_PATH.
	EventPayloadFilePath string `json:"eventPayloadFilePath,omitzero"`
	// GraphQLURL is the GitHub GraphQL API URL. Environment variable: GITHUB_GRAPHQL_URL.
	GraphQLURL string `json:"graphqlUrl,omitzero"`
	// HeadRef is the source branch of the pull request. Set only for pull_request and pull_request_target. Environment variable: GITHUB_HEAD_REF.
	HeadRef string `json:"headRef,omitzero"`
	// JobID is the job_id of the current job. Environment variable: GITHUB_JOB.
	JobID string `json:"jobId,omitzero"`
	// OutputFilePath is the step's file for setting outputs (see core.SetOutput). Environment variable: GITHUB_OUTPUT.
	OutputFilePath string `json:"outputFilePath,omitzero"`
	// PathFilePath is the step's file for adding to PATH (see core.AddToPath). Environment variable: GITHUB_PATH.
	PathFilePath string `json:"pathFilePath,omitzero"`
	// Ref is the fully formed ref that triggered the run, e.g. refs/heads/main, refs/tags/v1.0.0 or refs/pull/45/merge. Environment variable: GITHUB_REF.
	Ref string `json:"ref,omitzero"`
	// RefName is the short name of the branch or tag that triggered the run. Environment variable: GITHUB_REF_NAME.
	RefName string `json:"refName,omitzero"`
	// RefProtected is true if branch protections or rulesets apply to the ref that triggered the run. Environment variable: GITHUB_REF_PROTECTED.
	RefProtected *bool `json:"refProtected,omitzero"`
	// RefType is the type of ref that triggered the run: branch or tag. Environment variable: GITHUB_REF_TYPE.
	RefType string `json:"refType,omitzero"`
	// Repository is the owner and repository name, e.g. octocat/Hello-World. Environment variable: GITHUB_REPOSITORY.
	Repository string `json:"repository,omitzero"`
	// RepositoryID is the ID of the repository. Environment variable: GITHUB_REPOSITORY_ID.
	RepositoryID string `json:"repositoryId,omitzero"`
	// RepositoryOwner is the repository owner's name. Environment variable: GITHUB_REPOSITORY_OWNER.
	RepositoryOwner string `json:"repositoryOwner,omitzero"`
	// RepositoryOwnerID is the repository owner's account ID. Environment variable: GITHUB_REPOSITORY_OWNER_ID.
	RepositoryOwnerID string `json:"repositoryOwnerId,omitzero"`
	// RetentionDays is how many days workflow run logs and artifacts are kept. Environment variable: GITHUB_RETENTION_DAYS.
	RetentionDays int `json:"retentionDays,omitzero"`
	// RunAttempt is the attempt number of this workflow run, starting at 1 and incremented on each re-run. Environment variable: GITHUB_RUN_ATTEMPT.
	RunAttempt int `json:"runAttempt,omitzero"`
	// RunID is the unique number of this workflow run within the repository; it does not change on re-runs. Environment variable: GITHUB_RUN_ID.
	RunID int `json:"runId,omitzero"`
	// RunNumber is the run's number within its workflow, starting at 1; it does not change on re-runs. Environment variable: GITHUB_RUN_NUMBER.
	RunNumber int `json:"runNumber,omitzero"`
	// ServerURL is the URL of the GitHub server, e.g. https://github.com. Environment variable: GITHUB_SERVER_URL.
	ServerURL string `json:"serverUrl,omitzero"`
	// Sha is the commit SHA that triggered the workflow; which commit depends on the event. Environment variable: GITHUB_SHA.
	Sha string `json:"sha,omitzero"`
	// StepSummaryFilePath is the step's job summary file (see core.StepSummary). Environment variable: GITHUB_STEP_SUMMARY.
	StepSummaryFilePath string `json:"stepSummaryFilePath,omitzero"`
	// TriggeringActor is the user that initiated the run; on a re-run it may differ from [ActionContext.Actor], whose privileges the run uses. Environment variable: GITHUB_TRIGGERING_ACTOR.
	TriggeringActor string `json:"triggeringActor,omitzero"`
	// Workflow is the name of the workflow. Environment variable: GITHUB_WORKFLOW.
	Workflow string `json:"workflow,omitzero"`
	// WorkflowRef is the ref path to the workflow file. Environment variable: GITHUB_WORKFLOW_REF.
	WorkflowRef string `json:"workflowRef,omitzero"`
	// WorkflowSha is the commit SHA of the workflow file. Environment variable: GITHUB_WORKFLOW_SHA.
	WorkflowSha string `json:"workflowSha,omitzero"`
	// WorkspacePath is the default working directory for steps and where actions/checkout puts the repository. Environment variable: GITHUB_WORKSPACE.
	WorkspacePath string `json:"workspacePath,omitzero"`
	// RunnerArch is the runner's architecture: X86, X64, ARM or ARM64. Environment variable: RUNNER_ARCH.
	RunnerArch string `json:"runnerArch,omitzero"`
	// RunnerDebug is true when debug logging is enabled; the variable is unset otherwise. Environment variable: RUNNER_DEBUG.
	RunnerDebug *bool `json:"runnerDebug,omitzero"`
	// RunnerEnvironment is github-hosted or self-hosted. Environment variable: RUNNER_ENVIRONMENT.
	RunnerEnvironment string `json:"runnerEnvironment,omitzero"`
	// RunnerName is the name of the runner executing the job; it may not be unique. Environment variable: RUNNER_NAME.
	RunnerName string `json:"runnerName,omitzero"`
	// RunnerOS is the runner's operating system: Linux, Windows or macOS. Environment variable: RUNNER_OS.
	RunnerOS string `json:"runnerOs,omitzero"`
	// RunnerTempPath is a temporary directory emptied at the start and end of each job. Environment variable: RUNNER_TEMP.
	RunnerTempPath string `json:"runnerTempPath,omitzero"`
	// RunnerToolCachePath is the directory of preinstalled tools on GitHub-hosted runners. Environment variable: RUNNER_TOOL_CACHE.
	RunnerToolCachePath string `json:"runnerToolCachePath,omitzero"`
	// EventPayload is the decoded event payload: a pointer to the github.com/google/go-github/v92/github event type matching [ActionContext.EventName], a *map[string]any for an event without a matching type, or nil when there is no event file. Prefer the typed accessors, such as [ActionContext.PullRequestEvent].
	EventPayload any `json:"eventPayload,omitzero"`

	// rawEventPayload holds the unmodified JSON bytes read from the event file.
	// Stored separately from EventPayload so callers on a different go-github major can unmarshal it themselves.
	rawEventPayload stdjson.RawMessage
}

// String returns a pretty-printed representation of the ActionContext with all fields.
// String fields that are zero-valued are shown as empty, bool pointer fields that are nil are shown as "<nil>".
func (ac *ActionContext) String() string {
	if ac == nil {
		return "ActionContext: <nil>"
	}

	var sb strings.Builder
	sb.WriteString("ActionContext:\n")

	formatStr := func(name, val string) string {
		if val == "" {
			return fmt.Sprintf("  %-25s <not set>\n", name+":")
		}
		return fmt.Sprintf("  %-25s %s\n", name+":", val)
	}

	formatBool := func(name string, ptr *bool) string {
		if ptr == nil {
			return fmt.Sprintf("  %-25s <not set>\n", name+":")
		}
		return fmt.Sprintf("  %-25s %t\n", name+":", *ptr)
	}

	formatInt := func(name string, val int) string {
		if val == 0 {
			return fmt.Sprintf("  %-25s <not set>\n", name+":")
		}
		return fmt.Sprintf("  %-25s %d\n", name+":", val)
	}

	formatInterface := func(name string, val any) string {
		if val == nil {
			return fmt.Sprintf("  %-25s <not set>\n", name+":")
		}
		return fmt.Sprintf("  %-25s %T\n", name+":", val)
	}

	sb.WriteString(formatBool("GithubCI", ac.GithubCI))
	sb.WriteString(formatStr("Action", ac.Action))
	sb.WriteString(formatStr("ActionPath", ac.ActionPath))
	sb.WriteString(formatStr("ActionRepository", ac.ActionRepository))
	sb.WriteString(formatBool("Actions", ac.Actions))
	sb.WriteString(formatStr("Actor", ac.Actor))
	sb.WriteString(formatStr("ActorID", ac.ActorID))
	sb.WriteString(formatStr("APIURL", ac.APIURL))
	sb.WriteString(formatStr("BaseRef", ac.BaseRef))
	sb.WriteString(formatStr("EnvFilePath", ac.EnvFilePath))
	sb.WriteString(formatStr("EventName", ac.EventName))
	sb.WriteString(formatStr("EventPayloadFilePath", ac.EventPayloadFilePath))
	sb.WriteString(formatStr("GraphQLURL", ac.GraphQLURL))
	sb.WriteString(formatStr("HeadRef", ac.HeadRef))
	sb.WriteString(formatStr("JobID", ac.JobID))
	sb.WriteString(formatStr("OutputFilePath", ac.OutputFilePath))
	sb.WriteString(formatStr("PathFilePath", ac.PathFilePath))
	sb.WriteString(formatStr("Ref", ac.Ref))
	sb.WriteString(formatStr("RefName", ac.RefName))
	sb.WriteString(formatBool("RefProtected", ac.RefProtected))
	sb.WriteString(formatStr("RefType", ac.RefType))
	sb.WriteString(formatStr("Repository", ac.Repository))
	sb.WriteString(formatStr("RepositoryID", ac.RepositoryID))
	sb.WriteString(formatStr("RepositoryOwner", ac.RepositoryOwner))
	sb.WriteString(formatStr("RepositoryOwnerID", ac.RepositoryOwnerID))
	sb.WriteString(formatInt("RetentionDays", ac.RetentionDays))
	sb.WriteString(formatInt("RunAttempt", ac.RunAttempt))
	sb.WriteString(formatInt("RunID", ac.RunID))
	sb.WriteString(formatInt("RunNumber", ac.RunNumber))
	sb.WriteString(formatStr("ServerURL", ac.ServerURL))
	sb.WriteString(formatStr("Sha", ac.Sha))
	sb.WriteString(formatStr("StepSummaryFilePath", ac.StepSummaryFilePath))
	sb.WriteString(formatStr("TriggeringActor", ac.TriggeringActor))
	sb.WriteString(formatStr("Workflow", ac.Workflow))
	sb.WriteString(formatStr("WorkflowRef", ac.WorkflowRef))
	sb.WriteString(formatStr("WorkflowSha", ac.WorkflowSha))
	sb.WriteString(formatStr("WorkspacePath", ac.WorkspacePath))
	sb.WriteString(formatStr("RunnerArch", ac.RunnerArch))
	sb.WriteString(formatBool("RunnerDebug", ac.RunnerDebug))
	sb.WriteString(formatStr("RunnerEnvironment", ac.RunnerEnvironment))
	sb.WriteString(formatStr("RunnerName", ac.RunnerName))
	sb.WriteString(formatStr("RunnerOS", ac.RunnerOS))
	sb.WriteString(formatStr("RunnerTempPath", ac.RunnerTempPath))
	sb.WriteString(formatStr("RunnerToolCachePath", ac.RunnerToolCachePath))
	sb.WriteString(formatInterface("EventPayload", ac.EventPayload))

	return sb.String()
}

// RawEventPayload returns the unparsed JSON bytes of the event payload read from the event file.
// It returns nil if no event file was loaded.
// This escape hatch lets consumers on a different go-github major version — or any
// package that needs fields not covered by the typed accessors — unmarshal the raw JSON themselves.
func (ac *ActionContext) RawEventPayload() stdjson.RawMessage {
	if ac == nil {
		return nil
	}
	return ac.rawEventPayload
}

// Load reads the current step's context from the process environment and, when GITHUB_EVENT_NAME and GITHUB_EVENT_PATH are set, decodes the event payload file.
//
// Load reads the environment and the file on every call; call it once and pass the result around. Outside GitHub Actions it returns a context with every field unset, not an error. It returns an error if the event file cannot be read or decoded.
func Load() (*ActionContext, error) {
	return load(afero.NewOsFs(), os.LookupEnv)
}

type lookupEnvFunc func(key string) (string, bool)

// load is the testable implementation of [Load]: lookupEnv supplies the environment and fs holds the event payload file.
func load(fs afero.Fs, lookupEnv lookupEnvFunc) (*ActionContext, error) {
	ac := &ActionContext{
		GithubCI:             lookupToBoolPtr(lookupEnv("CI")),
		Action:               lookupToStr(lookupEnv("GITHUB_ACTION")),
		ActionPath:           lookupToStr(lookupEnv("GITHUB_ACTION_PATH")),
		ActionRepository:     lookupToStr(lookupEnv("GITHUB_ACTION_REPOSITORY")),
		Actions:              lookupToBoolPtr(lookupEnv("GITHUB_ACTIONS")),
		Actor:                lookupToStr(lookupEnv("GITHUB_ACTOR")),
		ActorID:              lookupToStr(lookupEnv("GITHUB_ACTOR_ID")),
		APIURL:               lookupToStr(lookupEnv("GITHUB_API_URL")),
		BaseRef:              lookupToStr(lookupEnv("GITHUB_BASE_REF")),
		EnvFilePath:          lookupToStr(lookupEnv("GITHUB_ENV")),
		EventName:            lookupToStr(lookupEnv("GITHUB_EVENT_NAME")),
		EventPayloadFilePath: lookupToStr(lookupEnv("GITHUB_EVENT_PATH")),
		GraphQLURL:           lookupToStr(lookupEnv("GITHUB_GRAPHQL_URL")),
		HeadRef:              lookupToStr(lookupEnv("GITHUB_HEAD_REF")),
		JobID:                lookupToStr(lookupEnv("GITHUB_JOB")),
		OutputFilePath:       lookupToStr(lookupEnv("GITHUB_OUTPUT")),
		PathFilePath:         lookupToStr(lookupEnv("GITHUB_PATH")),
		Ref:                  lookupToStr(lookupEnv("GITHUB_REF")),
		RefName:              lookupToStr(lookupEnv("GITHUB_REF_NAME")),
		RefProtected:         lookupToBoolPtr(lookupEnv("GITHUB_REF_PROTECTED")),
		RefType:              lookupToStr(lookupEnv("GITHUB_REF_TYPE")),
		Repository:           lookupToStr(lookupEnv("GITHUB_REPOSITORY")),
		RepositoryID:         lookupToStr(lookupEnv("GITHUB_REPOSITORY_ID")),
		RepositoryOwner:      lookupToStr(lookupEnv("GITHUB_REPOSITORY_OWNER")),
		RepositoryOwnerID:    lookupToStr(lookupEnv("GITHUB_REPOSITORY_OWNER_ID")),
		RetentionDays:        lookupToInt(lookupEnv("GITHUB_RETENTION_DAYS")),
		RunAttempt:           lookupToInt(lookupEnv("GITHUB_RUN_ATTEMPT")),
		RunID:                lookupToInt(lookupEnv("GITHUB_RUN_ID")),
		RunNumber:            lookupToInt(lookupEnv("GITHUB_RUN_NUMBER")),
		ServerURL:            lookupToStr(lookupEnv("GITHUB_SERVER_URL")),
		Sha:                  lookupToStr(lookupEnv("GITHUB_SHA")),
		StepSummaryFilePath:  lookupToStr(lookupEnv("GITHUB_STEP_SUMMARY")),
		TriggeringActor:      lookupToStr(lookupEnv("GITHUB_TRIGGERING_ACTOR")),
		Workflow:             lookupToStr(lookupEnv("GITHUB_WORKFLOW")),
		WorkflowRef:          lookupToStr(lookupEnv("GITHUB_WORKFLOW_REF")),
		WorkflowSha:          lookupToStr(lookupEnv("GITHUB_WORKFLOW_SHA")),
		WorkspacePath:        lookupToStr(lookupEnv("GITHUB_WORKSPACE")),
		RunnerArch:           lookupToStr(lookupEnv("RUNNER_ARCH")),
		RunnerDebug:          lookupToBoolPtr(lookupEnv("RUNNER_DEBUG")),
		RunnerEnvironment:    lookupToStr(lookupEnv("RUNNER_ENVIRONMENT")),
		RunnerName:           lookupToStr(lookupEnv("RUNNER_NAME")),
		RunnerOS:             lookupToStr(lookupEnv("RUNNER_OS")),
		RunnerTempPath:       lookupToStr(lookupEnv("RUNNER_TEMP")),
		RunnerToolCachePath:  lookupToStr(lookupEnv("RUNNER_TOOL_CACHE")),
	}

	if ac.EventName == "" || ac.EventPayloadFilePath == "" {
		return ac, nil
	}

	payloadBytes, err := afero.ReadFile(fs, ac.EventPayloadFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read event payload file: %w", err)
	}

	payload := newEventPayload(ac.EventName)
	if err := json.Unmarshal(payloadBytes, payload, json.MatchCaseInsensitiveNames(true)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event payload: %w", err)
	}
	ac.EventPayload = payload
	ac.rawEventPayload = stdjson.RawMessage(payloadBytes)

	return ac, nil
}

// newEventPayload returns a pointer to the go-github type for the named event, or to a map for an event without one.
func newEventPayload(eventName string) any {
	switch eventName {
	case "branch_protection_rule":
		return &github.BranchProtectionRuleEvent{}
	case "check_run":
		return &github.CheckRunEvent{}
	case "check_suite":
		return &github.CheckSuiteEvent{}
	case "create":
		return &github.CreateEvent{}
	case "delete":
		return &github.DeleteEvent{}
	case "deployment":
		return &github.DeploymentEvent{}
	case "deployment_status":
		return &github.DeploymentStatusEvent{}
	case "discussion":
		return &github.DiscussionEvent{}
	case "discussion_comment":
		return &github.DiscussionCommentEvent{}
	case "fork":
		return &github.ForkEvent{}
	case "gollum":
		return &github.GollumEvent{}
	case "package":
		return &github.PackageEvent{}
	case "issue_comment":
		return &github.IssueCommentEvent{}
	case "issues":
		return &github.IssuesEvent{}
	case "label":
		return &github.LabelEvent{}
	case "merge_group":
		return &github.MergeGroupEvent{}
	case "milestone":
		return &github.MilestoneEvent{}
	case "page_build":
		return &github.PageBuildEvent{}
	case "public":
		return &github.PublicEvent{}
	case "pull_request":
		return &github.PullRequestEvent{}
	case "pull_request_review":
		return &github.PullRequestReviewEvent{}
	case "pull_request_review_comment":
		return &github.PullRequestReviewCommentEvent{}
	case "pull_request_target":
		return &github.PullRequestTargetEvent{}
	case "push":
		return &github.PushEvent{}
	case "registry_package":
		return &github.RegistryPackageEvent{}
	case "release":
		return &github.ReleaseEvent{}
	case "repository":
		return &github.RepositoryEvent{}
	case "repository_dispatch":
		return &github.RepositoryDispatchEvent{}
	case "status":
		return &github.StatusEvent{}
	case "watch":
		return &github.WatchEvent{}
	case "workflow_dispatch":
		return &github.WorkflowDispatchEvent{}
	case "workflow_run":
		return &github.WorkflowRunEvent{}
	default:
		return &map[string]any{}
	}
}

// lookupToStr converts lookupEnv results to a string value.
// Returns the value if set, or an empty string if not.
func lookupToStr(v string, isSet bool) string {
	if isSet {
		return v
	}
	return ""
}

// lookupToInt converts lookupEnv results to an int value.
// Returns the parsed integer if set and valid, or 0 otherwise.
func lookupToInt(v string, isSet bool) int {
	if isSet {
		vI, err := strconv.Atoi(v)
		if err != nil {
			return 0
		}
		return vI
	}
	return 0
}

// lookupToBoolPtr converts lookupEnv results to a bool pointer.
// Returns a pointer to the parsed bool if set and valid, or nil otherwise.
func lookupToBoolPtr(v string, isSet bool) *bool {
	if isSet {
		vB, err := strconv.ParseBool(v)
		if err != nil {
			return nil
		}
		return &vB
	}
	return nil
}
