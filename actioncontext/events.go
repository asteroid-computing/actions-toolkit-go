package actioncontext

import "github.com/google/go-github/v92/github"

// BranchProtectionRuleEvent returns the event payload as a [github.BranchProtectionRuleEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) BranchProtectionRuleEvent() (*github.BranchProtectionRuleEvent, bool) {
	e, ok := ac.EventPayload.(*github.BranchProtectionRuleEvent)
	return e, ok
}

// CheckRunEvent returns the event payload as a [github.CheckRunEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) CheckRunEvent() (*github.CheckRunEvent, bool) {
	e, ok := ac.EventPayload.(*github.CheckRunEvent)
	return e, ok
}

// CheckSuiteEvent returns the event payload as a [github.CheckSuiteEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) CheckSuiteEvent() (*github.CheckSuiteEvent, bool) {
	e, ok := ac.EventPayload.(*github.CheckSuiteEvent)
	return e, ok
}

// CreateEvent returns the event payload as a [github.CreateEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) CreateEvent() (*github.CreateEvent, bool) {
	e, ok := ac.EventPayload.(*github.CreateEvent)
	return e, ok
}

// DeleteEvent returns the event payload as a [github.DeleteEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) DeleteEvent() (*github.DeleteEvent, bool) {
	e, ok := ac.EventPayload.(*github.DeleteEvent)
	return e, ok
}

// DeploymentEvent returns the event payload as a [github.DeploymentEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) DeploymentEvent() (*github.DeploymentEvent, bool) {
	e, ok := ac.EventPayload.(*github.DeploymentEvent)
	return e, ok
}

// DeploymentStatusEvent returns the event payload as a [github.DeploymentStatusEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) DeploymentStatusEvent() (*github.DeploymentStatusEvent, bool) {
	e, ok := ac.EventPayload.(*github.DeploymentStatusEvent)
	return e, ok
}

// DiscussionEvent returns the event payload as a [github.DiscussionEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) DiscussionEvent() (*github.DiscussionEvent, bool) {
	e, ok := ac.EventPayload.(*github.DiscussionEvent)
	return e, ok
}

// DiscussionCommentEvent returns the event payload as a [github.DiscussionCommentEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) DiscussionCommentEvent() (*github.DiscussionCommentEvent, bool) {
	e, ok := ac.EventPayload.(*github.DiscussionCommentEvent)
	return e, ok
}

// ForkEvent returns the event payload as a [github.ForkEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) ForkEvent() (*github.ForkEvent, bool) {
	e, ok := ac.EventPayload.(*github.ForkEvent)
	return e, ok
}

// GollumEvent returns the event payload as a [github.GollumEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) GollumEvent() (*github.GollumEvent, bool) {
	e, ok := ac.EventPayload.(*github.GollumEvent)
	return e, ok
}

// PackageEvent returns the event payload as a [github.PackageEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) PackageEvent() (*github.PackageEvent, bool) {
	e, ok := ac.EventPayload.(*github.PackageEvent)
	return e, ok
}

// IssueCommentEvent returns the event payload as a [github.IssueCommentEvent]
// and reports whether the current event is of that type.
// GitHub sends comments on pull requests as issue_comment events too; check the payload's Issue.IsPullRequest to tell them apart.
func (ac *ActionContext) IssueCommentEvent() (*github.IssueCommentEvent, bool) {
	e, ok := ac.EventPayload.(*github.IssueCommentEvent)
	return e, ok
}

// IssuesEvent returns the event payload as a [github.IssuesEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) IssuesEvent() (*github.IssuesEvent, bool) {
	e, ok := ac.EventPayload.(*github.IssuesEvent)
	return e, ok
}

// LabelEvent returns the event payload as a [github.LabelEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) LabelEvent() (*github.LabelEvent, bool) {
	e, ok := ac.EventPayload.(*github.LabelEvent)
	return e, ok
}

// MergeGroupEvent returns the event payload as a [github.MergeGroupEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) MergeGroupEvent() (*github.MergeGroupEvent, bool) {
	e, ok := ac.EventPayload.(*github.MergeGroupEvent)
	return e, ok
}

// MilestoneEvent returns the event payload as a [github.MilestoneEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) MilestoneEvent() (*github.MilestoneEvent, bool) {
	e, ok := ac.EventPayload.(*github.MilestoneEvent)
	return e, ok
}

// PageBuildEvent returns the event payload as a [github.PageBuildEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) PageBuildEvent() (*github.PageBuildEvent, bool) {
	e, ok := ac.EventPayload.(*github.PageBuildEvent)
	return e, ok
}

// PublicEvent returns the event payload as a [github.PublicEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) PublicEvent() (*github.PublicEvent, bool) {
	e, ok := ac.EventPayload.(*github.PublicEvent)
	return e, ok
}

// PullRequestEvent returns the event payload as a [github.PullRequestEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) PullRequestEvent() (*github.PullRequestEvent, bool) {
	e, ok := ac.EventPayload.(*github.PullRequestEvent)
	return e, ok
}

// PullRequestReviewEvent returns the event payload as a [github.PullRequestReviewEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) PullRequestReviewEvent() (*github.PullRequestReviewEvent, bool) {
	e, ok := ac.EventPayload.(*github.PullRequestReviewEvent)
	return e, ok
}

// PullRequestReviewCommentEvent returns the event payload as a [github.PullRequestReviewCommentEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) PullRequestReviewCommentEvent() (*github.PullRequestReviewCommentEvent, bool) {
	e, ok := ac.EventPayload.(*github.PullRequestReviewCommentEvent)
	return e, ok
}

// PullRequestTargetEvent returns the event payload as a [github.PullRequestTargetEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) PullRequestTargetEvent() (*github.PullRequestTargetEvent, bool) {
	e, ok := ac.EventPayload.(*github.PullRequestTargetEvent)
	return e, ok
}

// PushEvent returns the event payload as a [github.PushEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) PushEvent() (*github.PushEvent, bool) {
	e, ok := ac.EventPayload.(*github.PushEvent)
	return e, ok
}

// RegistryPackageEvent returns the event payload as a [github.RegistryPackageEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) RegistryPackageEvent() (*github.RegistryPackageEvent, bool) {
	e, ok := ac.EventPayload.(*github.RegistryPackageEvent)
	return e, ok
}

// ReleaseEvent returns the event payload as a [github.ReleaseEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) ReleaseEvent() (*github.ReleaseEvent, bool) {
	e, ok := ac.EventPayload.(*github.ReleaseEvent)
	return e, ok
}

// RepositoryEvent returns the event payload as a [github.RepositoryEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) RepositoryEvent() (*github.RepositoryEvent, bool) {
	e, ok := ac.EventPayload.(*github.RepositoryEvent)
	return e, ok
}

// RepositoryDispatchEvent returns the event payload as a [github.RepositoryDispatchEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) RepositoryDispatchEvent() (*github.RepositoryDispatchEvent, bool) {
	e, ok := ac.EventPayload.(*github.RepositoryDispatchEvent)
	return e, ok
}

// StatusEvent returns the event payload as a [github.StatusEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) StatusEvent() (*github.StatusEvent, bool) {
	e, ok := ac.EventPayload.(*github.StatusEvent)
	return e, ok
}

// WatchEvent returns the event payload as a [github.WatchEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) WatchEvent() (*github.WatchEvent, bool) {
	e, ok := ac.EventPayload.(*github.WatchEvent)
	return e, ok
}

// WorkflowDispatchEvent returns the event payload as a [github.WorkflowDispatchEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) WorkflowDispatchEvent() (*github.WorkflowDispatchEvent, bool) {
	e, ok := ac.EventPayload.(*github.WorkflowDispatchEvent)
	return e, ok
}

// WorkflowRunEvent returns the event payload as a [github.WorkflowRunEvent]
// and reports whether the current event is of that type.
func (ac *ActionContext) WorkflowRunEvent() (*github.WorkflowRunEvent, bool) {
	e, ok := ac.EventPayload.(*github.WorkflowRunEvent)
	return e, ok
}
