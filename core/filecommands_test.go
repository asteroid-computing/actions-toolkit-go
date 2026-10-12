package core

import (
	"os"
	"strings"
	"testing"
	"uuid"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
)

func TestBuildFileCommandMissingEnv(t *testing.T) {
	_, err := buildFileCommand(afero.NewMemMapFs(), func(string) (string, bool) {
		return "", false
	}, uuid.NewV7, EnvironmentVariableFilePathEnv, "KEY", "VALUE")
	if err == nil {
		t.Fatalf("expected error when env var missing")
	}
}

func TestBuildFileCommandPopulatesFields(t *testing.T) {
	fs := afero.NewMemMapFs()
	u := uuid.MustParse("00000000-0000-7000-8000-000000000042")
	fc, err := buildFileCommand(fs, func(string) (string, bool) {
		return "/tmp/testfile", true
	}, func() uuid.UUID {
		return u
	}, OutputFilePathEnv, "KEY", "VALUE")
	if err != nil {
		t.Fatalf("buildFileCommand returned error: %v", err)
	}

	wantDelimiter := "ghadelimiter_" + u.String()
	if diff := cmp.Diff(wantDelimiter, fc.Delimiter); diff != "" {
		t.Fatalf("delimiter mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff("/tmp/testfile", fc.FilePath); diff != "" {
		t.Fatalf("file path mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff("KEY", fc.Key); diff != "" {
		t.Fatalf("key mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff("VALUE", fc.Value); diff != "" {
		t.Fatalf("value mismatch (-want +got):\n%s", diff)
	}
	if fc.Fs != fs {
		t.Fatalf("expected fs to be the provided instance")
	}
}

func TestFileCommandSendWritesToFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/envfile"
	if err := afero.WriteFile(fs, path, []byte("initial\n"), 0o644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	fc := &FileCommand{
		Fs:        fs,
		FilePath:  path,
		Key:       "FOO",
		Value:     "bar",
		Delimiter: "delim",
	}

	if err := fc.Send(); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	data, err := afero.ReadFile(fs, path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	content := string(data)
	want := "initial\n" + "FOO<<delim" + OsSpecificNewline + "bar" + OsSpecificNewline + "delim" + OsSpecificNewline
	if diff := cmp.Diff(want, content); diff != "" {
		t.Fatalf("Send() content mismatch (-want +got):\n%s", diff)
	}
}

func TestFileCommandSendTwoAppends(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "/envfile2"
	if err := afero.WriteFile(fs, path, []byte{}, 0o644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	fc1 := &FileCommand{Fs: fs, FilePath: path, Key: "FIRST", Value: "1", Delimiter: "d1"}
	fc2 := &FileCommand{Fs: fs, FilePath: path, Key: "SECOND", Value: "2", Delimiter: "d2"}

	if err := fc1.Send(); err != nil {
		t.Fatalf("fc1.Send returned error: %v", err)
	}
	if err := fc2.Send(); err != nil {
		t.Fatalf("fc2.Send returned error: %v", err)
	}

	data, err := afero.ReadFile(fs, path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	want := "FIRST<<d1" + OsSpecificNewline + "1" + OsSpecificNewline + "d1" + OsSpecificNewline +
		"SECOND<<d2" + OsSpecificNewline + "2" + OsSpecificNewline + "d2" + OsSpecificNewline
	if diff := cmp.Diff(want, string(data)); diff != "" {
		t.Fatalf("Send() content mismatch (-want +got):\n%s", diff)
	}
}

func TestFileCommandSendRejectsDelimiterCollision(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "/file", nil, 0o644)

	fc := &FileCommand{
		Fs:        fs,
		FilePath:  "/file",
		Key:       "bad-delim",
		Value:     "value",
		Delimiter: "bad",
	}

	if err := fc.Send(); err == nil {
		t.Fatalf("expected error when key contains delimiter")
	}
}

func TestSetCommandsWriteUsingEnvPaths(t *testing.T) {
	tests := []struct {
		name string
		env  string
		fn   func(string) error
	}{
		{name: "set env", env: EnvironmentVariableFilePathEnv, fn: func(path string) error { return SetEnv("ABC", "123") }},
		{name: "set output", env: OutputFilePathEnv, fn: func(path string) error { return SetOutput("RESULT", "ok") }},
		{name: "set state", env: GithubStateFilePathEnv, fn: func(path string) error { return SetState("STATE", "done") }},
		{name: "add to path", env: SystemPathFilePathEnv, fn: func(path string) error { return AddToPath("/tmp/bin") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			targetFile := tmpDir + "/cmd.txt"
			if err := os.WriteFile(targetFile, []byte{}, 0o644); err != nil {
				t.Fatalf("failed to create target file: %v", err)
			}
			t.Setenv(tt.env, targetFile)

			if err := tt.fn(targetFile); err != nil {
				t.Fatalf("command returned error: %v", err)
			}

			data, err := os.ReadFile(targetFile)
			if err != nil {
				t.Fatalf("failed to read target file: %v", err)
			}
			content := string(data)
			if !strings.Contains(content, "<<") {
				t.Fatalf("expected workflow command content in file, got: %q", content)
			}
		})
	}
}
