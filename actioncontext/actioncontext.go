// Package actioncontext provides the GitHub Actions runtime context,
// auto-populated from environment variables at package import time.
//
// The init function populates the package-level variables (such as [Actor], [Repository], and [EventName])
// and [CurrentActionContext] by reading the standard GitHub Actions default environment variables.
//
// The typed event payload in [ActionContext.EventPayload] and the per-type accessor methods
// (such as [ActionContext.PullRequestEvent]) are populated from the event file
// identified by GITHUB_EVENT_NAME and GITHUB_EVENT_PATH.
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

	"github.com/asteroid-computing/actions-toolkit-go/core"
)

// ActionContext represents the complete GitHub Actions runtime context
// with all environment variables and event payload.
// Use [CurrentActionContext] for the auto-populated instance.
type ActionContext struct {
	GithubCI             *bool  `json:"githubCI,omitzero"`
	Action               string `json:"action,omitzero"`
	ActionPath           string `json:"actionPath,omitzero"`
	ActionRepository     string `json:"actionRepository,omitzero"`
	Actions              *bool  `json:"actions,omitzero"`
	Actor                string `json:"actor,omitzero"`
	ActorID              string `json:"actorId,omitzero"`
	APIURL               string `json:"apiUrl,omitzero"`
	BaseRef              string `json:"baseRef,omitzero"`
	EnvFilePath          string `json:"envFilePath,omitzero"`
	EventName            string `json:"eventName,omitzero"`
	EventPayloadFilePath string `json:"eventPayloadFilePath,omitzero"`
	GraphQLURL           string `json:"graphqlUrl,omitzero"`
	HeadRef              string `json:"headRef,omitzero"`
	JobID                string `json:"jobId,omitzero"`
	OutputFilePath       string `json:"outputFilePath,omitzero"`
	PathFilePath         string `json:"pathFilePath,omitzero"`
	Ref                  string `json:"ref,omitzero"`
	RefName              string `json:"refName,omitzero"`
	RefProtected         *bool  `json:"refProtected,omitzero"`
	RefType              string `json:"refType,omitzero"`
	Repository           string `json:"repository,omitzero"`
	RepositoryID         string `json:"repositoryId,omitzero"`
	RepositoryOwner      string `json:"repositoryOwner,omitzero"`
	RepositoryOwnerID    string `json:"repositoryOwnerId,omitzero"`
	RetentionDays        int    `json:"retentionDays,omitzero"`
	RunAttempt           int    `json:"runAttempt,omitzero"`
	RunID                int    `json:"runId,omitzero"`
	RunNumber            int    `json:"runNumber,omitzero"`
	ServerURL            string `json:"serverUrl,omitzero"`
	Sha                  string `json:"sha,omitzero"`
	StepSummaryFilePath  string `json:"stepSummaryFilePath,omitzero"`
	TriggeringActor      string `json:"triggeringActor,omitzero"`
	Workflow             string `json:"workflow,omitzero"`
	WorkflowRef          string `json:"workflowRef,omitzero"`
	WorkflowSha          string `json:"workflowSha,omitzero"`
	WorkspacePath        string `json:"workspacePath,omitzero"`
	RunnerArch           string `json:"runnerArch,omitzero"`
	RunnerDebug          *bool  `json:"runnerDebug,omitzero"`
	RunnerEnvironment    string `json:"runnerEnvironment,omitzero"`
	RunnerName           string `json:"runnerName,omitzero"`
	RunnerOS             string `json:"runnerOs,omitzero"`
	RunnerTempPath       string `json:"runnerTempPath,omitzero"`
	RunnerToolCachePath  string `json:"runnerToolCachePath,omitzero"`
	EventPayload         any    `json:"eventPayload,omitzero"`

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

// The following exported variables are auto-populated at package import time from
// the GitHub Actions default environment variables.
// Each mirrors the same field on [CurrentActionContext].
// Details for each variable:
// https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables
var (
	// CurrentActionContext holds the populated [ActionContext] from GitHub Actions environment variables.
	CurrentActionContext *ActionContext

	// GithubCI indicates whether the runner is executing under GitHub Actions (CI).
	// During a GitHub Actions workflow run, this is always true.
	// Environment Variable: CI
	GithubCI *bool

	// Action is the name of the action currently running, or the id of a step. For example, for an action, __repo-owner_name-of-action-repo.
	// Environment Variable: GITHUB_ACTION
	Action string

	// ActionPath is the path where an action is located. This property is only supported in composite actions.
	// Environment Variable: GITHUB_ACTION_PATH
	ActionPath string

	// ActionRepository is, for a step executing an action, the owner and repository name of the action.
	// Environment Variable: GITHUB_ACTION_REPOSITORY
	ActionRepository string

	// Actions is always set to true when GitHub Actions is running the workflow.
	// You can use this variable to differentiate when tests are being run locally or by GitHub Actions.
	// Environment Variable: GITHUB_ACTIONS
	Actions *bool

	// Actor is the name of the person or app that initiated the workflow.
	// Environment Variable: GITHUB_ACTOR
	Actor string

	// ActorID is the account ID of the person or app that triggered the initial workflow run.
	// Environment Variable: GITHUB_ACTOR_ID
	ActorID string

	// APIURL is the GitHub REST API base URL
	// Environment Variable: GITHUB_API_URL
	APIURL string

	// BaseRef is the name of the base ref or target branch of the pull request in a workflow run.
	// This is only set when the event that triggers a workflow run is either pull_request or pull_request_target.
	// Environment Variable: GITHUB_BASE_REF
	BaseRef string

	// EnvFilePath is the path on the runner to the file that sets variables from workflow commands.
	// The path to this file is unique to the current step and changes for each step in a job.
	// Environment Variable: GITHUB_ENV
	EnvFilePath string

	// EventName is the name of the event that triggered the workflow.
	// Environment Variable: GITHUB_EVENT_NAME
	EventName string

	// EventPayloadFilePath is the path to the file on the runner that contains the full event webhook payload.
	// Environment Variable: GITHUB_EVENT_PATH
	EventPayloadFilePath string

	// GraphQLURL is the GitHub GraphQL API base URL.
	// Environment Variable: GITHUB_GRAPHQL_URL
	GraphQLURL string

	// HeadRef is the head ref or source branch of the pull request in a workflow run.
	// This property is only set when the event that triggers a workflow run is either pull_request or pull_request_target.
	// Environment Variable: GITHUB_HEAD_REF
	HeadRef string

	// JobID is the Job ID of the current job.
	// Environment Variable: GITHUB_JOB
	JobID string

	// OutputFilePath is the path on the runner to the file that sets the current step's outputs from workflow commands.
	// The path to this file is unique to the current step and changes for each step in a job.
	// Environment Variable: GITHUB_OUTPUT
	OutputFilePath string

	// PathFilePath is the path on the runner to the file that sets system PATH variables from workflow commands.
	// The path to this file is unique to the current step and changes for each step in a job.
	// Environment Variable: GITHUB_PATH
	PathFilePath string

	// Ref is the fully-formed ref of the branch or tag that triggered the workflow run.
	// For workflows triggered by push, this is the branch or tag ref that was pushed.
	// For workflows triggered by pull_request, this is the pull request merge branch.
	// For workflows triggered by release, this is the release tag created.
	// For other triggers, this is the branch or tag ref that triggered the workflow run.
	// This is only set if a branch or tag is available for the event type.
	// The ref given is fully-formed, meaning that for branches the format is refs/heads/<branch_name>.
	// For pull requests events except pull_request_target, it is refs/pull/<pr_number>/merge
	// pull_request_target events have the ref from the base branch. For tags it is refs/tags/<tag_name>.
	// Environment Variable: GITHUB_REF
	Ref string

	// RefName is the short ref name of the branch or tag that triggered the workflow run.
	// This value matches the branch or tag name shown on GitHub.
	// Environment Variable: GITHUB_REF_NAME
	RefName string

	// RefProtected is true if branch protections or rulesets are configured for the ref that triggered the workflow run.
	// Environment Variable: GITHUB_REF_PROTECTED
	RefProtected *bool

	// RefType is the type of ref that triggered the workflow run. Valid values are branch or tag.
	// Environment Variable: GITHUB_REF_TYPE
	RefType string

	// Repository is the owner and repository name.
	// Environment Variable: GITHUB_REPOSITORY
	Repository string

	// RepositoryID is the ID of the repository.
	// Environment Variable: GITHUB_REPOSITORY_ID
	RepositoryID string

	// RepositoryOwner is the repository owner's name.
	// Environment Variable: GITHUB_REPOSITORY_OWNER
	RepositoryOwner string

	// RepositoryOwnerID is the repository owner's account ID.
	// Environment Variable: GITHUB_REPOSITORY_OWNER_ID
	RepositoryOwnerID string

	// RetentionDays is the number of days that workflow run logs and artifacts are kept.
	// Environment Variable: GITHUB_RETENTION_DAYS
	RetentionDays int

	// RunAttempt is a unique number for each attempt of a particular workflow run in a repository.
	// This number begins at 1 for the workflow run's first attempt, and increments with each re-run.
	// Environment Variable: GITHUB_RUN_ATTEMPT
	RunAttempt int

	// RunID is a unique number for each workflow run within a repository.
	// This number does not change if you re-run the workflow run.
	// Environment Variable: GITHUB_RUN_ID
	RunID int

	// RunNumber is a unique number for each run of a particular workflow in a repository.
	// This number begins at 1 for the workflow's first run, and increments with each new run.
	// This number does not change if you re-run the workflow run.
	// Environment Variable: GITHUB_RUN_NUMBER
	RunNumber int

	// ServerURL is the URL of the GitHub server.
	// Environment Variable: GITHUB_SERVER_URL
	ServerURL string

	// Sha is the commit SHA that triggered the workflow.
	// The value of this commit SHA depends on the event that triggered the workflow.
	// Environment Variable: GITHUB_SHA
	Sha string

	// StepSummaryFilePath is the path on the runner to the file that contains job summaries from workflow commands.
	// The path to this file is unique to the current step and changes for each step in a job.
	// Environment Variable: GITHUB_STEP_SUMMARY
	StepSummaryFilePath string

	// TriggeringActor is the username of the user that initiated the workflow run.
	// If the workflow run is a re-run, this value may differ from github.actor.
	// Any workflow re-runs will use the privileges of github.actor,
	// even if the actor initiating the re-run (github.triggering_actor) has different privileges.
	// Environment Variable: GITHUB_TRIGGERING_ACTOR
	TriggeringActor string

	// Workflow is the name of the workflow.
	// Environment Variable: GITHUB_WORKFLOW
	Workflow string

	// WorkflowRef is the ref path to the workflow.
	// Environment Variable: GITHUB_WORKFLOW_REF
	WorkflowRef string

	// WorkflowSha is the commit SHA for the workflow file.
	// Environment Variable: GITHUB_WORKFLOW_SHA
	WorkflowSha string

	// WorkspacePath is the default working directory on the runner for steps,
	// and the default location of your repository when using the checkout action.
	// Environment Variable: GITHUB_WORKSPACE
	WorkspacePath string

	// RunnerArch is the architecture of the runner executing the job. Possible values are X86, X64, ARM, or ARM64.
	// Environment Variable: RUNNER_ARCH
	RunnerArch string

	// RunnerDebug is this is set only if debug logging is enabled, and always has the value of 1.
	// It can be useful as an indicator to enable additional debugging or verbose logging in your own job steps.
	// Environment Variable: RUNNER_DEBUG
	RunnerDebug *bool

	// RunnerEnvironment is the environment of the runner executing the job.
	// Possible values are: github-hosted for GitHub-hosted runners provided by GitHub,
	// and self-hosted for self-hosted runners configured by the repository owner.
	// Environment Variable: RUNNER_ENVIRONMENT
	RunnerEnvironment string

	// RunnerName is the name of the runner executing the job.
	// This name may not be unique in a workflow run as runners at the repository
	// and organization levels could use the same name.
	// Environment Variable: RUNNER_NAME
	RunnerName string

	// RunnerOS is the operating system of the runner executing the job.
	// Possible values are Linux, Windows, or macOS.
	// Environment Variable: RUNNER_OS
	RunnerOS string

	// RunnerTempPath is the path to a temporary directory on the runner.
	// This directory is emptied at the beginning and end of each job.
	// Note that files will not be removed if the runner's user account does not have permission to delete them.
	// Environment Variable: RUNNER_TEMP
	RunnerTempPath string

	// RunnerToolCachePath is the path to the directory containing preinstalled tools for GitHub-hosted runners.
	// Environment Variable: RUNNER_TOOL_CACHE
	RunnerToolCachePath string

	// EventPayload is the typed event payload object parsed from the event file.
	// The concrete type depends on [EventName] and is one of the
	// [github.com/google/go-github/v92/github] event types.
	EventPayload any
)

func init() {
	err := loadActionContext(afero.NewOsFs(), os.LookupEnv)
	if err != nil {
		core.Debug("Error loading action context: " + err.Error())
	}

	if CurrentActionContext != nil {
		core.Debug(CurrentActionContext.String())
	}
}

type lookupEnvFunc func(key string) (string, bool)

// loadActionContext loads action context from environment variables and is provided for testing purposes.
// The provided lookupEnv function is used to retrieve environment variable values.
// The provided afero.Fs is used to read the event payload file.
// Returns an error if loading the event payload fails.
func loadActionContext(fs afero.Fs, lookupEnv lookupEnvFunc) error {
	EventPayload = nil

	Action = lookupToStr(lookupEnv("GITHUB_ACTION"))
	ActionPath = lookupToStr(lookupEnv("GITHUB_ACTION_PATH"))
	ActionRepository = lookupToStr(lookupEnv("GITHUB_ACTION_REPOSITORY"))
	Actions = lookupToBoolPtr(lookupEnv("GITHUB_ACTIONS"))
	Actor = lookupToStr(lookupEnv("GITHUB_ACTOR"))
	ActorID = lookupToStr(lookupEnv("GITHUB_ACTOR_ID"))
	APIURL = lookupToStr(lookupEnv("GITHUB_API_URL"))
	BaseRef = lookupToStr(lookupEnv("GITHUB_BASE_REF"))
	EnvFilePath = lookupToStr(lookupEnv("GITHUB_ENV"))
	EventName = lookupToStr(lookupEnv("GITHUB_EVENT_NAME"))
	EventPayloadFilePath = lookupToStr(lookupEnv("GITHUB_EVENT_PATH"))
	GithubCI = lookupToBoolPtr(lookupEnv("CI"))
	GraphQLURL = lookupToStr(lookupEnv("GITHUB_GRAPHQL_URL"))
	HeadRef = lookupToStr(lookupEnv("GITHUB_HEAD_REF"))
	JobID = lookupToStr(lookupEnv("GITHUB_JOB"))
	OutputFilePath = lookupToStr(lookupEnv("GITHUB_OUTPUT"))
	PathFilePath = lookupToStr(lookupEnv("GITHUB_PATH"))
	Ref = lookupToStr(lookupEnv("GITHUB_REF"))
	RefName = lookupToStr(lookupEnv("GITHUB_REF_NAME"))
	RefProtected = lookupToBoolPtr(lookupEnv("GITHUB_REF_PROTECTED"))
	RefType = lookupToStr(lookupEnv("GITHUB_REF_TYPE"))
	Repository = lookupToStr(lookupEnv("GITHUB_REPOSITORY"))
	RepositoryID = lookupToStr(lookupEnv("GITHUB_REPOSITORY_ID"))
	RepositoryOwner = lookupToStr(lookupEnv("GITHUB_REPOSITORY_OWNER"))
	RepositoryOwnerID = lookupToStr(lookupEnv("GITHUB_REPOSITORY_OWNER_ID"))
	RetentionDays = lookupToInt(lookupEnv("GITHUB_RETENTION_DAYS"))
	RunAttempt = lookupToInt(lookupEnv("GITHUB_RUN_ATTEMPT"))
	RunID = lookupToInt(lookupEnv("GITHUB_RUN_ID"))
	RunNumber = lookupToInt(lookupEnv("GITHUB_RUN_NUMBER"))
	ServerURL = lookupToStr(lookupEnv("GITHUB_SERVER_URL"))
	Sha = lookupToStr(lookupEnv("GITHUB_SHA"))
	StepSummaryFilePath = lookupToStr(lookupEnv("GITHUB_STEP_SUMMARY"))
	TriggeringActor = lookupToStr(lookupEnv("GITHUB_TRIGGERING_ACTOR"))
	Workflow = lookupToStr(lookupEnv("GITHUB_WORKFLOW"))
	WorkflowRef = lookupToStr(lookupEnv("GITHUB_WORKFLOW_REF"))
	WorkflowSha = lookupToStr(lookupEnv("GITHUB_WORKFLOW_SHA"))
	WorkspacePath = lookupToStr(lookupEnv("GITHUB_WORKSPACE"))

	RunnerArch = lookupToStr(lookupEnv("RUNNER_ARCH"))
	RunnerDebug = lookupToBoolPtr(lookupEnv("RUNNER_DEBUG"))
	RunnerEnvironment = lookupToStr(lookupEnv("RUNNER_ENVIRONMENT"))
	RunnerName = lookupToStr(lookupEnv("RUNNER_NAME"))
	RunnerOS = lookupToStr(lookupEnv("RUNNER_OS"))
	RunnerTempPath = lookupToStr(lookupEnv("RUNNER_TEMP"))
	RunnerToolCachePath = lookupToStr(lookupEnv("RUNNER_TOOL_CACHE"))

	var rawPayload stdjson.RawMessage

	if EventName != "" && EventPayloadFilePath != "" {
		payloadBytes, err := afero.ReadFile(fs, EventPayloadFilePath)

		if err != nil {
			EventPayload = nil
			return fmt.Errorf("failed to read event payload file: %w", err)
		}

		rawPayload = stdjson.RawMessage(payloadBytes)

		switch EventName {
		case "branch_protection_rule":
			EventPayload = &github.BranchProtectionRuleEvent{}
		case "check_run":
			EventPayload = &github.CheckRunEvent{}
		case "check_suite":
			EventPayload = &github.CheckSuiteEvent{}
		case "create":
			EventPayload = &github.CreateEvent{}
		case "delete":
			EventPayload = &github.DeleteEvent{}
		case "deployment":
			EventPayload = &github.DeploymentEvent{}
		case "deployment_status":
			EventPayload = &github.DeploymentStatusEvent{}
		case "discussion":
			EventPayload = &github.DiscussionEvent{}
		case "discussion_comment":
			EventPayload = &github.DiscussionCommentEvent{}
		case "fork":
			EventPayload = &github.ForkEvent{}
		case "gollum":
			EventPayload = &github.GollumEvent{}
		case "package":
			EventPayload = &github.PackageEvent{}
		case "issue_comment":
			EventPayload = &github.IssueCommentEvent{}
		case "issues":
			EventPayload = &github.IssuesEvent{}
		case "label":
			EventPayload = &github.LabelEvent{}
		case "merge_group":
			EventPayload = &github.MergeGroupEvent{}
		case "milestone":
			EventPayload = &github.MilestoneEvent{}
		case "page_build":
			EventPayload = &github.PageBuildEvent{}
		case "public":
			EventPayload = &github.PublicEvent{}
		case "pull_request":
			EventPayload = &github.PullRequestEvent{}
		case "pull_request_comment":
			EventPayload = &github.IssueCommentEvent{}
		case "pull_request_review":
			EventPayload = &github.PullRequestReviewEvent{}
		case "pull_request_review_comment":
			EventPayload = &github.PullRequestReviewCommentEvent{}
		case "pull_request_target":
			EventPayload = &github.PullRequestTargetEvent{}
		case "push":
			EventPayload = &github.PushEvent{}
		case "registry_package":
			EventPayload = &github.RegistryPackageEvent{}
		case "release":
			EventPayload = &github.ReleaseEvent{}
		case "repository":
			EventPayload = &github.RepositoryEvent{}
		case "repository_dispatch":
			EventPayload = &github.RepositoryDispatchEvent{}
		case "status":
			EventPayload = &github.StatusEvent{}
		case "watch":
			EventPayload = &github.WatchEvent{}
		case "workflow_dispatch":
			EventPayload = &github.WorkflowDispatchEvent{}
		case "workflow_run":
			EventPayload = &github.WorkflowRunEvent{}
		default:
			EventPayload = &map[string]any{}
		}

		err = json.Unmarshal(payloadBytes, EventPayload, json.MatchCaseInsensitiveNames(true))

		if err != nil {
			EventPayload = nil
			return fmt.Errorf("failed to unmarshal event payload: %w", err)
		}
	}

	CurrentActionContext = &ActionContext{
		GithubCI:             GithubCI,
		Action:               Action,
		ActionPath:           ActionPath,
		ActionRepository:     ActionRepository,
		Actions:              Actions,
		Actor:                Actor,
		ActorID:              ActorID,
		APIURL:               APIURL,
		BaseRef:              BaseRef,
		EnvFilePath:          EnvFilePath,
		EventName:            EventName,
		EventPayloadFilePath: EventPayloadFilePath,
		GraphQLURL:           GraphQLURL,
		HeadRef:              HeadRef,
		JobID:                JobID,
		OutputFilePath:       OutputFilePath,
		PathFilePath:         PathFilePath,
		Ref:                  Ref,
		RefName:              RefName,
		RefProtected:         RefProtected,
		RefType:              RefType,
		Repository:           Repository,
		RepositoryID:         RepositoryID,
		RepositoryOwner:      RepositoryOwner,
		RepositoryOwnerID:    RepositoryOwnerID,
		RetentionDays:        RetentionDays,
		RunAttempt:           RunAttempt,
		RunID:                RunID,
		RunNumber:            RunNumber,
		ServerURL:            ServerURL,
		Sha:                  Sha,
		StepSummaryFilePath:  StepSummaryFilePath,
		TriggeringActor:      TriggeringActor,
		Workflow:             Workflow,
		WorkflowRef:          WorkflowRef,
		WorkflowSha:          WorkflowSha,
		WorkspacePath:        WorkspacePath,
		RunnerArch:           RunnerArch,
		RunnerDebug:          RunnerDebug,
		RunnerEnvironment:    RunnerEnvironment,
		RunnerName:           RunnerName,
		RunnerOS:             RunnerOS,
		RunnerTempPath:       RunnerTempPath,
		RunnerToolCachePath:  RunnerToolCachePath,
		EventPayload:         EventPayload,
		rawEventPayload:      rawPayload,
	}

	return nil
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
