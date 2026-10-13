package core

import (
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"uuid"
)

// WorkflowCommandString is the delimiter that wraps GitHub Actions workflow commands.
// https://docs.github.com/actions/using-workflows/workflow-commands-for-github-actions#about-workflow-commands
const WorkflowCommandString = "::"

const (
	DebugCommandStr        = "debug"
	NoticeCommandStr       = "notice"
	WarningCommandStr      = "warning"
	ErrorCommandStr        = "error"
	StartGroupCommandStr   = "group"
	EndGroupCommandStr     = "endgroup"
	MaskCommandStr         = "add-mask"
	StopCommandsCommandStr = "stop-commands"
)

// WorkflowCommand represents a formatted GitHub Actions workflow command
// using the [WorkflowCommandString] delimiter syntax.
//
// GitHub Workflow Command Reference: https://docs.github.com/actions/using-workflows/workflow-commands-for-github-actions#about-workflow-commands
type WorkflowCommand struct {
	w          io.Writer
	command    string
	properties map[string]string
	message    string
}

// NewWorkflowCommand creates a WorkflowCommand targeting the given writer.
func NewWorkflowCommand(w io.Writer, command string, properties map[string]string, message string) *WorkflowCommand {
	return &WorkflowCommand{
		w:          w,
		command:    command,
		properties: properties,
		message:    message,
	}
}

// String renders the workflow command into the GitHub Actions command syntax.
// Properties are rendered in sorted key order, so the same command always renders the same string.
// GitHub Workflow Command Reference: https://docs.github.com/actions/using-workflows/workflow-commands-for-github-actions#about-workflow-commands
func (wc *WorkflowCommand) String() string {
	var cmdStr strings.Builder
	cmdStr.WriteString(WorkflowCommandString)
	cmdStr.WriteString(wc.command)

	if len(wc.properties) > 0 {
		cmdStr.WriteString(" ")
		for i, k := range slices.Sorted(maps.Keys(wc.properties)) {
			if i > 0 {
				cmdStr.WriteString(",")
			}

			fmt.Fprintf(&cmdStr, "%s=%s", k, escapeCommandParameter(wc.properties[k]))
		}
	}

	cmdStr.WriteString(WorkflowCommandString)
	cmdStr.WriteString(escapeCommandData(wc.message))

	return cmdStr.String()
}

// Send writes the workflow command to the provided writer.
// GitHub Workflow Command Reference: https://docs.github.com/actions/using-workflows/workflow-commands-for-github-actions#about-workflow-commands
func (wc *WorkflowCommand) Send() error {
	_, err := fmt.Fprintln(wc.w, wc)

	return err
}

// Debug writes a debug message to stdout as a [WorkflowCommand].
// GitHub Debug Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-a-debug-message
func Debug(message string) error {
	cmd := NewWorkflowCommand(os.Stdout, DebugCommandStr, nil, message)

	return cmd.Send()
}

// WorkflowCommandOptions describes optional metadata for notice/warning/error commands.
type WorkflowCommandOptions struct {
	FileName    *string // source file path for the annotation
	StartLine   *int    // starting line number
	EndLine     *int    // ending line number
	Title       *string // annotation title
	StartColumn *int    // starting column number
	EndColumn   *int    // ending column number
}

// CmdProperties converts the options into a property map for a [WorkflowCommand].
func (wco *WorkflowCommandOptions) CmdProperties() map[string]string {
	if wco == nil {
		return nil
	}

	params := make(map[string]string)

	if wco.FileName != nil && *wco.FileName != "" {
		params["file"] = *wco.FileName
	}
	if wco.StartLine != nil {
		params["line"] = strconv.Itoa(*wco.StartLine)
	}
	if wco.EndLine != nil {
		params["endLine"] = strconv.Itoa(*wco.EndLine)
	}
	if wco.Title != nil {
		params["title"] = *wco.Title
	}
	if wco.StartColumn != nil {
		params["col"] = strconv.Itoa(*wco.StartColumn)
	}
	if wco.EndColumn != nil {
		params["endColumn"] = strconv.Itoa(*wco.EndColumn)
	}

	return params
}

// Notice writes a notice message to stdout as a [WorkflowCommand].
// GitHub Notice Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-a-notice-message
func Notice(message string, options *WorkflowCommandOptions) error {
	cmd := NewWorkflowCommand(os.Stdout, NoticeCommandStr, options.CmdProperties(), message)

	return cmd.Send()
}

// Warning writes a warning message to stdout as a [WorkflowCommand].
// GitHub Warning Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-a-warning-message
func Warning(message string, options *WorkflowCommandOptions) error {
	cmd := NewWorkflowCommand(os.Stdout, WarningCommandStr, options.CmdProperties(), message)

	return cmd.Send()
}

// Error writes an error message to stdout as a [WorkflowCommand].
// GitHub Error Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-an-error-message
func Error(message string, options *WorkflowCommandOptions) error {
	cmd := NewWorkflowCommand(os.Stdout, ErrorCommandStr, options.CmdProperties(), message)

	return cmd.Send()
}

// IsDebug reports whether the runner has step debug logging enabled, which GitHub signals by setting RUNNER_DEBUG to 1.
// Like the JS toolkit's isDebug, only the exact value "1" counts; actioncontext's RunnerDebug field also accepts other strconv.ParseBool true values, which GitHub does not send.
// GitHub Reference: https://docs.github.com/en/actions/how-tos/monitor-workflows/enable-debug-logging
func IsDebug() bool {
	return isDebug(os.LookupEnv)
}

// isDebug is the testable implementation of [IsDebug].
func isDebug(lookupEnv lookupEnvFunc) bool {
	v, _ := lookupEnv("RUNNER_DEBUG")
	return v == "1"
}

// SetFailed writes message as an error annotation and exits the process with status 1, failing the step.
//
// The JS toolkit's setFailed only sets the exit code and lets the action keep running; Go cannot set an exit code for later, so SetFailed exits immediately.
// Deferred functions do not run, so write anything buffered first, in particular [StepSummary] with [Summary.Write].
// To fail with more control, write the error with [Error] and return from main with os.Exit(1) yourself.
// GitHub Error Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-an-error-message
func SetFailed(message string) {
	setFailed(os.Stdout, os.Exit, message)
}

// setFailed is the testable implementation of [SetFailed]. A failure to write the annotation does not prevent the exit.
func setFailed(w io.Writer, exit func(int), message string) {
	_ = NewWorkflowCommand(w, ErrorCommandStr, nil, message).Send()
	exit(1)
}

// StartGroup begins a collapsible log group.
// GitHub Group Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#grouping-log-lines
func StartGroup(name string) error {
	cmd := NewWorkflowCommand(os.Stdout, StartGroupCommandStr, nil, name)

	return cmd.Send()
}

// EndGroup closes the most recent collapsible log group.
// GitHub Group Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#grouping-log-lines
func EndGroup() error {
	cmd := NewWorkflowCommand(os.Stdout, EndGroupCommandStr, nil, "")

	return cmd.Send()
}

// MaskValue instructs GitHub Actions to redact the provided value from logs.
// GitHub Mask Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#masking-a-value-in-a-log
func MaskValue(value string) error {
	cmd := NewWorkflowCommand(os.Stdout, MaskCommandStr, nil, value)

	return cmd.Send()
}

type ResumeCommandsFunc func() error

// StopCommands suspends workflow command processing and returns a [ResumeCommandsFunc] to re-enable it.
// GitHub Stop Commands Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#stopping-and-starting-workflow-commands
func StopCommands() (ResumeCommandsFunc, error) {
	delimUUID := uuid.NewV7()

	cmd := NewWorkflowCommand(os.Stdout, StopCommandsCommandStr, nil, delimUUID.String())
	err := cmd.Send()
	if err != nil {
		return nil, err
	}

	resumeCommands := func() error {
		resumeCmd := NewWorkflowCommand(os.Stdout, delimUUID.String(), nil, "")

		return resumeCmd.Send()
	}

	return resumeCommands, nil
}

// escapeCommandData escapes workflow command data according to GitHub Actions rules.
func escapeCommandData(s string) string {
	if s == "" {
		return s
	}

	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	s = strings.ReplaceAll(s, "\n", "%0A")
	return s
}

// escapeCommandParameter escapes workflow command parameters according to GitHub Actions rules.
func escapeCommandParameter(s string) string {
	if s == "" {
		return s
	}

	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	s = strings.ReplaceAll(s, "\n", "%0A")
	s = strings.ReplaceAll(s, ":", "%3A")
	s = strings.ReplaceAll(s, ",", "%2C")
	return s
}
