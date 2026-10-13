package core

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// fakeLookup returns a lookupEnvFunc that serves the provided key→value map.
func fakeLookup(env map[string]string) lookupEnvFunc {
	return func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	}
}

func TestGetInput(t *testing.T) {
	trimFalse := new(false)
	trimTrue := new(true)

	tests := []struct {
		name    string
		input   string
		env     map[string]string
		opts    *InputOptions
		want    string
		wantErr error
	}{
		{
			name:  "unset_input_returns_empty",
			input: "my-input",
			env:   map[string]string{},
			opts:  nil,
			want:  "",
		},
		{
			name:  "set_and_trim_default",
			input: "name",
			env:   map[string]string{"INPUT_NAME": "  hello  "},
			opts:  nil,
			want:  "hello",
		},
		{
			name:  "trim_explicit_true",
			input: "name",
			env:   map[string]string{"INPUT_NAME": "  hello  "},
			opts:  &InputOptions{TrimWhitespace: trimTrue},
			want:  "hello",
		},
		{
			name:  "trim_false_preserves_whitespace",
			input: "name",
			env:   map[string]string{"INPUT_NAME": "  hello  "},
			opts:  &InputOptions{TrimWhitespace: trimFalse},
			want:  "  hello  ",
		},
		{
			name:    "required_unset_returns_err",
			input:   "token",
			env:     map[string]string{},
			opts:    &InputOptions{Required: true},
			want:    "",
			wantErr: ErrInputRequired,
		},
		{
			name:  "required_set_returns_value",
			input: "token",
			env:   map[string]string{"INPUT_TOKEN": "abc123"},
			opts:  &InputOptions{Required: true},
			want:  "abc123",
		},
		{
			name:  "space_in_name_normalised_to_underscore",
			input: "my input",
			env:   map[string]string{"INPUT_MY_INPUT": "spaced"},
			opts:  nil,
			want:  "spaced",
		},
		{
			name:  "hyphen_in_name_preserved",
			input: "my-input",
			env:   map[string]string{"INPUT_MY-INPUT": "hyph"},
			opts:  nil,
			want:  "hyph",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getInput(fakeLookup(tt.env), tt.input, tt.opts)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetInput(%q) error = %v, want errors.Is(%v)", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetInput(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("GetInput(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetBooleanInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		env     map[string]string
		opts    *InputOptions
		want    bool
		wantErr error
	}{
		{name: "true_lowercase", input: "flag", env: map[string]string{"INPUT_FLAG": "true"}, want: true},
		{name: "true_titlecase", input: "flag", env: map[string]string{"INPUT_FLAG": "True"}, want: true},
		{name: "true_uppercase", input: "flag", env: map[string]string{"INPUT_FLAG": "TRUE"}, want: true},
		{name: "false_lowercase", input: "flag", env: map[string]string{"INPUT_FLAG": "false"}, want: false},
		{name: "false_titlecase", input: "flag", env: map[string]string{"INPUT_FLAG": "False"}, want: false},
		{name: "false_uppercase", input: "flag", env: map[string]string{"INPUT_FLAG": "FALSE"}, want: false},
		{
			name:    "invalid_value",
			input:   "flag",
			env:     map[string]string{"INPUT_FLAG": "yes"},
			wantErr: ErrInvalidBooleanInput,
		},
		{
			name:    "required_propagated",
			input:   "flag",
			env:     map[string]string{},
			opts:    &InputOptions{Required: true},
			wantErr: ErrInputRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBooleanInput(fakeLookup(tt.env), tt.input, tt.opts)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetBooleanInput(%q) error = %v, want errors.Is(%v)", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetBooleanInput(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("GetBooleanInput(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetMultilineInput(t *testing.T) {
	trimFalse := new(false)

	tests := []struct {
		name    string
		input   string
		env     map[string]string
		opts    *InputOptions
		want    []string
		wantErr error
	}{
		{
			name:  "unset_returns_nil",
			input: "lines",
			env:   map[string]string{},
			want:  nil,
		},
		{
			name:  "multiline_split_and_trim",
			input: "lines",
			env:   map[string]string{"INPUT_LINES": "  foo  \n  bar  \n  baz  "},
			want:  []string{"foo", "bar", "baz"},
		},
		{
			name:  "empty_lines_dropped",
			input: "lines",
			env:   map[string]string{"INPUT_LINES": "foo\n\nbar\n\n"},
			want:  []string{"foo", "bar"},
		},
		{
			name:  "no_trim_preserves_whitespace",
			input: "lines",
			env:   map[string]string{"INPUT_LINES": "  foo  \n  bar  "},
			opts:  &InputOptions{TrimWhitespace: trimFalse},
			want:  []string{"  foo  ", "  bar  "},
		},
		{
			name:    "required_propagated",
			input:   "lines",
			env:     map[string]string{},
			opts:    &InputOptions{Required: true},
			wantErr: ErrInputRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMultilineInput(fakeLookup(tt.env), tt.input, tt.opts)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetMultilineInput(%q) error = %v, want errors.Is(%v)", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetMultilineInput(%q) unexpected error: %v", tt.input, err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetMultilineInput(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func ExampleGetInput() {
	os.Setenv("INPUT_NAME", "  Mona  ")
	val, err := GetInput("name", nil)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(val)
	// Output: Mona
}
