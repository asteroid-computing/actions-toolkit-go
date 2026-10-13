package core

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestSummaryWriteAppendAndOverwrite(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/summary"
	if err := afero.WriteFile(fs, path, []byte("old\n"), 0o644); err != nil {
		t.Fatalf("failed to seed summary file: %v", err)
	}

	s := newSummary(func(key string) (string, bool) {
		if key == SummaryEnvVar {
			return path, true
		}
		return "", false
	}, fs)

	s.AddHeading("Title", 1)
	if err := s.Write(nil); err != nil {
		t.Fatalf("append Write returned error: %v", err)
	}

	data, _ := afero.ReadFile(fs, path)
	content := string(data)
	if !strings.Contains(content, "old") || !strings.Contains(content, "<h1>Title</h1>") {
		t.Fatalf("expected append content, got: %q", content)
	}

	s.AddRaw("fresh", false)
	if err := s.Write(&SummaryWriteOptions{Overwrite: true}); err != nil {
		t.Fatalf("overwrite Write returned error: %v", err)
	}

	data, _ = afero.ReadFile(fs, path)
	if got := string(data); got != "fresh" {
		t.Fatalf("expected overwrite to replace content, got: %q", got)
	}
}

func TestSummaryWriteRequiresPath(t *testing.T) {
	s := newSummary(func(string) (string, bool) { return "", false }, afero.NewMemMapFs())
	s.AddRaw("text", false)
	if err := s.Write(nil); err == nil {
		t.Fatalf("expected error when summary path missing")
	}
}

func TestSummaryClearAndBufferState(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/summary"
	_ = afero.WriteFile(fs, path, []byte{}, 0o644)

	s := newSummary(func(key string) (string, bool) {
		if key == SummaryEnvVar {
			return path, true
		}
		return "", false
	}, fs)

	if !s.IsEmptyBuffer() {
		t.Fatalf("expected empty buffer initially")
	}

	s.AddRaw("line", true)
	if s.IsEmptyBuffer() {
		t.Fatalf("expected buffer not empty after AddRaw")
	}

	if err := s.Write(nil); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if !s.IsEmptyBuffer() {
		t.Fatalf("expected buffer cleared after write")
	}

	s.AddRaw("another", false)
	if err := s.ClearSummary(); err != nil {
		t.Fatalf("ClearSummary failed: %v", err)
	}

	data, _ := afero.ReadFile(fs, path)
	if len(data) != 0 {
		t.Fatalf("expected summary file cleared, got: %q", string(data))
	}
	if !s.IsEmptyBuffer() {
		t.Fatalf("expected buffer empty after ClearSummary")
	}
}

func TestSummaryContentHelpers(t *testing.T) {
	fs := afero.NewMemMapFs()
	s := newSummary(func(string) (string, bool) { return "/noop", true }, fs)

	s.AddRaw("raw", true)
	s.AddEOL()
	s.AddBreak()
	s.AddCodeBlock("echo hi", "bash")
	s.AddDetails("More", "<p>content</p>")
	s.AddHeading("Heading", 10) // clamps to h1
	s.AddImage("img.png", "desc", &SummaryImageOptions{Width: "100", Height: "50"})
	s.AddLink("Go", "https://go.dev")
	s.AddList([]string{"a", "b"}, false)
	s.AddList([]string{"1", "2"}, true)
	s.AddQuote("quote", "source")
	s.AddSeparator()
	s.AddTable([]SummaryTableRow{
		{
			{Data: "h1", Header: true},
			{Data: "h2", Header: true, ColSpan: "2"},
		},
		{
			{Data: "c1"},
			{Data: "c2", RowSpan: "2"},
			{Data: "c3"},
		},
	})

	out := s.String()

	if !strings.Contains(out, "raw"+OsSpecificNewline) {
		t.Fatalf("expected raw text with EOL, got: %q", out)
	}
	if !strings.Contains(out, "<br>") {
		t.Fatalf("expected break tag, got: %q", out)
	}
	if !strings.Contains(out, "<pre lang=\"bash\"><code>echo hi</code></pre>") {
		t.Fatalf("expected code block, got: %q", out)
	}
	if !strings.Contains(out, "<details><summary>More</summary><p>content</p></details>") {
		t.Fatalf("expected details block, got: %q", out)
	}
	if !strings.Contains(out, "<h1>Heading</h1>") {
		t.Fatalf("expected h1 heading, got: %q", out)
	}
	if !strings.Contains(out, "<img") || !strings.Contains(out, `src="img.png"`) || !strings.Contains(out, `alt="desc"`) {
		t.Fatalf("expected image tag, got: %q", out)
	}
	if !strings.Contains(out, "<a href=\"https://go.dev\">Go</a>") {
		t.Fatalf("expected link, got: %q", out)
	}
	if !strings.Contains(out, "<ul><li>a</li><li>b</li></ul>") {
		t.Fatalf("expected unordered list, got: %q", out)
	}
	if !strings.Contains(out, "<ol><li>1</li><li>2</li></ol>") {
		t.Fatalf("expected ordered list, got: %q", out)
	}
	if !strings.Contains(out, "<blockquote cite=\"source\">quote</blockquote>") {
		t.Fatalf("expected quote, got: %q", out)
	}
	if !strings.Contains(out, "<hr>") {
		t.Fatalf("expected separator, got: %q", out)
	}
	if !strings.Contains(out, "<table>") || !strings.Contains(out, "<th>h1</th>") || !strings.Contains(out, `colspan="2"`) || !strings.Contains(out, `rowspan="2"`) {
		t.Fatalf("expected table with spans, got: %q", out)
	}
}
