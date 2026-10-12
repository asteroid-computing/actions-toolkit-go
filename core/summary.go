// Package core provides helpers for emitting GitHub Actions workflow commands
// and working with runner-provided files.
// Information on GitHub Workflow Commands are here: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands
package core

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/afero"
)

// SummaryEnvVar is the environment variable containing the summary file path.
const SummaryEnvVar = "GITHUB_STEP_SUMMARY"

// StepSummary is the process-wide [Summary] writer bound to the GitHub Actions
// job summary file.
//
// Details on the GitHub Actions job summary are here: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#adding-a-job-summary
var StepSummary *Summary

// Summary holds buffered HTML content for the GitHub Actions job summary.
// https://docs.github.com/actions/using-workflows/workflow-commands-for-github-actions#adding-a-job-summary
type Summary struct {
	buffer   strings.Builder
	fs       afero.Fs
	filePath string
}

func init() {
	StepSummary = NewSummary(os.LookupEnv, afero.NewOsFs())
}

// NewSummary creates a [Summary] with the provided environment lookup and
// filesystem. If the summary file path is available, it is cached for writes.
// Similar to the GitHub provided actions toolkit, this does not panic if the
// summary file path is not available, but [Summary.Write] attempts will return an error.
func NewSummary(lookupEnvFunc LookupEnvFunc, fs afero.Fs) *Summary {
	if pathFromEnv, ok := lookupEnvFunc(SummaryEnvVar); ok {
		return &Summary{filePath: pathFromEnv, fs: fs}
	}

	return &Summary{fs: fs}
}

// SummaryWriteOptions configures how [Summary.Write] flushes the buffer to disk.
type SummaryWriteOptions struct {
	Overwrite bool // replace existing content instead of appending
}

// Write flushes the buffer to the summary file, appending unless
// [SummaryWriteOptions.Overwrite] is set. The buffer is cleared after a successful write.
func (s *Summary) Write(options *SummaryWriteOptions) error {
	if s.filePath == "" {
		return fmt.Errorf("summary filepath is not valid")
	}

	flags := os.O_WRONLY | os.O_CREATE
	if options != nil && options.Overwrite {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_APPEND
	}

	f, err := s.fs.OpenFile(s.filePath, flags, 0644)
	if err != nil {
		return fmt.Errorf("unable to write to summary file, %w", err)
	}

	defer func(file afero.File) {
		_ = file.Close()
	}(f)

	if _, err := f.WriteString(s.String()); err != nil {
		return fmt.Errorf("unable to write to summary file, %w", err)
	}

	s.buffer.Reset()

	return nil
}

// String returns the current buffer contents.
func (s *Summary) String() string {
	return s.buffer.String()
}

// ClearSummary wipes both the buffer and the summary file contents.
func (s *Summary) ClearSummary() error {
	s.EmptyBuffer()

	return s.Write(&SummaryWriteOptions{Overwrite: true})
}

// EmptyBuffer resets the in-memory buffer without writing to disk.
func (s *Summary) EmptyBuffer() {
	s.buffer.Reset()
}

// IsEmptyBuffer reports whether the buffer is empty.
func (s *Summary) IsEmptyBuffer() bool {
	return s.buffer.Len() == 0
}

// AddRaw appends raw text to the buffer and optionally appends an end-of-line.
func (s *Summary) AddRaw(text string, addEOL bool) {
	s.buffer.WriteString(text)
	if addEOL {
		s.AddEOL()
	}
}

// AddEOL appends an OS-specific end-of-line marker to the buffer.
func (s *Summary) AddEOL() {
	if runtime.GOOS == "windows" {
		s.buffer.WriteString("\r\n")
	} else {
		s.buffer.WriteString("\n")
	}
}

// AddBreak appends an HTML line break to the summary buffer.
func (s *Summary) AddBreak() {
	s.AddRaw(wrap("br", nil, nil), true)
}

// AddCodeBlock appends a code block (code) to the summary buffer with optional language (lang).
func (s *Summary) AddCodeBlock(code, lang string) {
	attrs := map[string]string{}
	if lang != "" {
		attrs["lang"] = lang
	}

	inner := wrap("code", &code, nil)
	s.AddRaw(wrap("pre", &inner, attrs), true)
}

// AddDetails appends a details HTML element to the summary buffer with a
// summary label and content.
func (s *Summary) AddDetails(label, content string) {
	inner := wrap("summary", &label, nil) + content
	s.AddRaw(wrap("details", &inner, nil), true)
}

// AddHeading appends an HTML heading to the shared StepSummary buffer.
// Level must be between 1 and 6; values outside this range default to 1.
func (s *Summary) AddHeading(text string, level int) {
	if level < 1 || level > 6 {
		level = 1
	}

	tag := fmt.Sprintf("h%d", level)
	s.AddRaw(wrap(tag, &text, nil), true)
}

// SummaryImageOptions configures the attributes for an image added via [Summary.AddImage].
type SummaryImageOptions struct {
	Width  string // HTML width attribute
	Height string // HTML height attribute
}

// AddImage appends an image element to the summary buffer with optional attributes.
func (s *Summary) AddImage(src, alt string, options *SummaryImageOptions) {
	attrs := map[string]string{
		"src": src,
		"alt": alt,
	}

	if options != nil {
		if options.Width != "" {
			attrs["width"] = options.Width
		}
		if options.Height != "" {
			attrs["height"] = options.Height
		}
	}

	s.AddRaw(wrap("img", nil, attrs), true)
}

// AddLink appends a hyperlink to the summary buffer.
func (s *Summary) AddLink(text, href string) {
	attrs := map[string]string{
		"href": href,
	}

	s.AddRaw(wrap("a", &text, attrs), true)
}

// AddList appends an unordered or ordered list to the summary buffer.
func (s *Summary) AddList(items []string, ordered bool) {
	tag := "ul"
	if ordered {
		tag = "ol"
	}

	var listItems strings.Builder

	for _, item := range items {
		listItems.WriteString(wrap("li", &item, nil))
	}

	content := listItems.String()
	s.AddRaw(wrap(tag, &content, nil), true)
}

// AddQuote appends a blockquote to the summary buffer with optional citation.
func (s *Summary) AddQuote(text, cite string) {
	attrs := map[string]string{}
	if cite != "" {
		attrs["cite"] = cite
	}

	s.AddRaw(wrap("blockquote", &text, attrs), true)
}

// AddSeparator appends a horizontal rule to the summary buffer.
func (s *Summary) AddSeparator() {
	s.AddRaw(wrap("hr", nil, nil), true)
}

// SummaryTableCell represents a cell in a summary table built via [Summary.AddTable].
type SummaryTableCell struct {
	Data    string // cell content
	Header  bool   // render as <th> instead of <td>
	ColSpan string // HTML colspan attribute
	RowSpan string // HTML rowspan attribute
}

// SummaryTableRow represents a row of [SummaryTableCell] values in a summary table.
type SummaryTableRow []SummaryTableCell

// AddTable appends an HTML table to the summary buffer.
func (s *Summary) AddTable(rows []SummaryTableRow) {
	var rowStrings strings.Builder

	for _, row := range rows {
		var cellStrings strings.Builder
		for _, cell := range row {
			cellTag := "td"
			if cell.Header {
				cellTag = "th"
			}

			cellAttrs := map[string]string{}
			if cell.ColSpan != "" {
				cellAttrs["colspan"] = cell.ColSpan
			}
			if cell.RowSpan != "" {
				cellAttrs["rowspan"] = cell.RowSpan
			}

			dataCopy := cell.Data
			cellStrings.WriteString(wrap(cellTag, &dataCopy, cellAttrs))
		}

		rowContent := cellStrings.String()
		rowStrings.WriteString(wrap("tr", &rowContent, nil))
	}

	content := rowStrings.String()
	s.AddRaw(wrap("table", &content, nil), true)
}

// wrap creates an HTML element string with optional content and attributes.
// Similar to the GitHub Actions toolkit implementation this is not exported
func wrap(tag string, content *string, attrs map[string]string) string {
	var str strings.Builder

	fmt.Fprintf(&str, "<%s", tag)
	for key, val := range attrs {
		fmt.Fprintf(&str, ` %s="%s"`, key, val)
	}
	str.WriteByte('>')

	if content != nil {
		fmt.Fprintf(&str, "%s</%s>", *content, tag)
	}

	return str.String()
}
