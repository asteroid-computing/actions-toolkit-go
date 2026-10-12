// Package cireport renders a CI step's outcome as markdown for the workflow job summary, so a reader sees what was built (and what failed, and how long it took) on the run page instead of expanding a log.
//
// Rendering ([Report.Markdown]) performs no I/O and is what the tests cover; delivery is left to the caller (the actions write it with githubactions.AddStepSummary). It is a package, not another action, because reporting is logic shared between Go callers in one module — an action would mean serialising results to JSON and parsing them back.
package cireport

import (
	"fmt"
	"strings"
	"time"
)

// Status is the outcome of one item.
type Status string

const (
	StatusOK      Status = "ok"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
)

func (s Status) icon() string {
	switch s {
	case StatusOK:
		return "✅"
	case StatusFailed:
		return "❌"
	default:
		return "⏭️"
	}
}

// Item is one row: a component that was built, skipped or failed.
type Item struct {
	Name     string
	Status   Status
	Detail   string        // artifacts, an error, or the reason for a skip
	Duration time.Duration // zero when not applicable
}

// Report is a step's outcome, ready to render.
type Report struct {
	Title   string
	Lede    string // one line of context above the table
	Empty   string // shown instead of a table when there are no items
	Items   []Item
	Version string
	RunURL  string
}

// Failed reports whether any item failed.
func (r Report) Failed() bool {
	for _, it := range r.Items {
		if it.Status == StatusFailed {
			return true
		}
	}
	return false
}

// Markdown renders the report. Durations are rounded to a tenth of a second: millisecond precision is noise that changes on every run.
func (r Report) Markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "## %s\n\n", r.Title)
	if r.Lede != "" {
		fmt.Fprintf(&b, "%s\n\n", r.Lede)
	}
	if len(r.Items) == 0 {
		b.WriteString(cmp(r.Empty, "Nothing to report.") + "\n")
	} else {
		b.WriteString("| | Name | Result | Time |\n|---|---|---|---|\n")
		for _, it := range r.Items {
			dur := "—"
			if it.Duration > 0 {
				dur = it.Duration.Round(100 * time.Millisecond).String()
			}
			detail := strings.ReplaceAll(cmp(it.Detail, string(it.Status)), "|", `\|`)
			fmt.Fprintf(&b, "| %s | `%s` | %s | %s |\n", it.Status.icon(), it.Name, detail, dur)
		}
	}
	var footer []string
	if r.Version != "" {
		footer = append(footer, fmt.Sprintf("Version `%s`", r.Version))
	}
	if r.RunURL != "" {
		footer = append(footer, fmt.Sprintf("[run log](%s)", r.RunURL))
	}
	if len(footer) > 0 {
		fmt.Fprintf(&b, "\n%s\n", strings.Join(footer, " · "))
	}
	return b.String()
}

func cmp(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
