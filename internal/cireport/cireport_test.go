package cireport

import (
	"strings"
	"testing"
	"time"
)

func TestMarkdown(t *testing.T) {
	r := Report{
		Title:   "Build",
		Items:   []Item{{Name: "hello", Status: StatusOK, Detail: "bin/hello", Duration: 1234 * time.Millisecond}, {Name: "other", Status: StatusFailed, Detail: "a | b"}},
		Version: "0.1.0+pr1.abc1234",
		RunURL:  "https://github.com/o/r/actions/runs/1",
	}
	got := r.Markdown()
	for _, want := range []string{
		"## Build\n",
		"| ✅ | `hello` | bin/hello | 1.2s |",
		"| ❌ | `other` | a \\| b | — |",
		"Version `0.1.0+pr1.abc1234` · [run log](https://github.com/o/r/actions/runs/1)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("markdown missing %q:\n%s", want, got)
		}
	}
	if !r.Failed() {
		t.Error("Failed() = false with a failed item")
	}
}

func TestMarkdownEmpty(t *testing.T) {
	got := Report{Title: "Build", Empty: "No component changed."}.Markdown()
	if !strings.Contains(got, "No component changed.") || strings.Contains(got, "|") {
		t.Errorf("empty report rendered a table or lost its message:\n%s", got)
	}
}
