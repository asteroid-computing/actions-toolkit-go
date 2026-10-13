package core

import (
	"bytes"
	"errors"
	"testing"
)

func TestWorkflowCommandString(t *testing.T) {
	tests := []struct {
		name    string
		command string
		props   map[string]string
		message string
		want    string
	}{
		{name: "no properties", command: DebugCommandStr, message: "hello", want: "::debug::hello"},
		{name: "empty message", command: EndGroupCommandStr, want: "::endgroup::"},
		{name: "properties in sorted order", command: ErrorCommandStr, props: map[string]string{"title": "T", "file": "a.go", "line": "3", "col": "7"}, message: "boom", want: "::error col=7,file=a.go,line=3,title=T::boom"},
		{name: "escaped data", command: NoticeCommandStr, message: "100%\r\nnext", want: "::notice::100%25%0D%0Anext"},
		{name: "escaped property", command: WarningCommandStr, props: map[string]string{"title": "a:b,c\n"}, message: "m", want: "::warning title=a%3Ab%2Cc%0A::m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewWorkflowCommand(nil, tt.command, tt.props, tt.message).String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestWorkflowCommandStringIsDeterministic guards against rendering properties in map iteration order, which varies between runs.
func TestWorkflowCommandStringIsDeterministic(t *testing.T) {
	props := (&WorkflowCommandOptions{FileName: new("a.go"), StartLine: new(1), EndLine: new(2), Title: new("T"), StartColumn: new(3), EndColumn: new(4)}).CmdProperties()
	const want = "::error col=3,endColumn=4,endLine=2,file=a.go,line=1,title=T::m"
	for range 100 {
		if got := NewWorkflowCommand(nil, ErrorCommandStr, props, "m").String(); got != want {
			t.Fatalf("String() = %q, want %q", got, want)
		}
	}
}

func TestWorkflowCommandSend(t *testing.T) {
	var buf bytes.Buffer
	if err := NewWorkflowCommand(&buf, MaskCommandStr, nil, "secret").Send(); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if got, want := buf.String(), "::add-mask::secret\n"; got != want {
		t.Errorf("Send wrote %q, want %q", got, want)
	}
}

func TestCmdPropertiesNilOptions(t *testing.T) {
	var opts *WorkflowCommandOptions
	if got := opts.CmdProperties(); got != nil {
		t.Errorf("CmdProperties() on nil options = %v, want nil", got)
	}
}

func TestIsDebug(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{name: "debug enabled", env: map[string]string{"RUNNER_DEBUG": "1"}, want: true},
		{name: "unset", env: map[string]string{}, want: false},
		{name: "true is not 1", env: map[string]string{"RUNNER_DEBUG": "true"}, want: false},
		{name: "zero", env: map[string]string{"RUNNER_DEBUG": "0"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDebug(fakeLookup(tt.env)); got != tt.want {
				t.Errorf("isDebug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetFailed(t *testing.T) {
	var buf bytes.Buffer
	code := -1
	setFailed(&buf, func(c int) { code = c }, "deploy failed: 50%\ndone")

	if got, want := buf.String(), "::error::deploy failed: 50%25%0Adone\n"; got != want {
		t.Errorf("setFailed wrote %q, want %q", got, want)
	}
	if code != 1 {
		t.Errorf("setFailed exited with %d, want 1", code)
	}
}

func TestSetFailedExitsWhenWriteFails(t *testing.T) {
	code := -1
	setFailed(failingWriter{}, func(c int) { code = c }, "boom")
	if code != 1 {
		t.Errorf("setFailed exited with %d after a write error, want 1", code)
	}
}

// failingWriter is an io.Writer whose writes always fail.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
