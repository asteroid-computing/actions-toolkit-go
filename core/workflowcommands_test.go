package core

import (
	"bytes"
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
