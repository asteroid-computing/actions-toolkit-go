package core

import (
	"errors"
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

	fc := &fileCommand{
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

	fc1 := &fileCommand{Fs: fs, FilePath: path, Key: "FIRST", Value: "1", Delimiter: "d1"}
	fc2 := &fileCommand{Fs: fs, FilePath: path, Key: "SECOND", Value: "2", Delimiter: "d2"}

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

	fc := &fileCommand{
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
		want string
	}{
		{name: "set env", env: EnvironmentVariableFilePathEnv, fn: func(path string) error { return SetEnv("ABC", "123") }, want: "ABC<<"},
		{name: "set output", env: OutputFilePathEnv, fn: func(path string) error { return SetOutput("RESULT", "ok") }, want: "RESULT<<"},
		{name: "set state", env: GithubStateFilePathEnv, fn: func(path string) error { return SetState("STATE", "done") }, want: "STATE<<"},
		{name: "add to path", env: SystemPathFilePathEnv, fn: func(path string) error { return AddToPath("/tmp/bin") }, want: "/tmp/bin" + OsSpecificNewline},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			targetFile := tmpDir + "/cmd.txt"
			if err := os.WriteFile(targetFile, []byte{}, 0o644); err != nil {
				t.Fatalf("failed to create target file: %v", err)
			}
			t.Setenv(tt.env, targetFile)
			// SetEnv and AddToPath change the test process's environment; t.Setenv restores both afterwards.
			t.Setenv("ABC", "")
			t.Setenv("PATH", os.Getenv("PATH"))

			if err := tt.fn(targetFile); err != nil {
				t.Fatalf("command returned error: %v", err)
			}

			data, err := os.ReadFile(targetFile)
			if err != nil {
				t.Fatalf("failed to read target file: %v", err)
			}
			content := string(data)
			if !strings.HasPrefix(content, tt.want) {
				t.Fatalf("expected file content starting %q, got: %q", tt.want, content)
			}
			switch tt.env {
			case EnvironmentVariableFilePathEnv:
				if got := os.Getenv("ABC"); got != "123" {
					t.Errorf("SetEnv did not set ABC in the process: got %q", got)
				}
			case SystemPathFilePathEnv:
				if got := os.Getenv("PATH"); !strings.HasPrefix(got, "/tmp/bin"+string(os.PathListSeparator)) {
					t.Errorf("AddToPath did not prepend to the process PATH: got %q", got)
				}
			}
		})
	}
}

// setenvCall records one call to an injected setenvFunc.
type setenvCall struct{ Key, Value string }

// recordSetenv returns a setenvFunc that appends each call to calls instead of changing the process environment.
func recordSetenv(calls *[]setenvCall) setenvFunc {
	return func(key, value string) error {
		*calls = append(*calls, setenvCall{key, value})
		return nil
	}
}

func TestAddToPath(t *testing.T) {
	const pathFile = "/github_path"
	sep := string(os.PathListSeparator)
	withFile := func(extra map[string]string) map[string]string {
		env := map[string]string{SystemPathFilePathEnv: pathFile}
		for k, v := range extra {
			env[k] = v
		}
		return env
	}
	tests := []struct {
		name      string
		existing  string
		dir       string
		env       map[string]string
		want      string
		wantCalls []setenvCall
		wantErr   bool
	}{
		{name: "writes one directory per line and prepends to PATH", dir: "/opt/tool/bin", env: withFile(map[string]string{"PATH": "/usr/bin" + sep + "/bin"}), want: "/opt/tool/bin" + OsSpecificNewline, wantCalls: []setenvCall{{"PATH", "/opt/tool/bin" + sep + "/usr/bin" + sep + "/bin"}}},
		{name: "appends to earlier entries", existing: "/first" + OsSpecificNewline, dir: "/second", env: withFile(map[string]string{"PATH": "/usr/bin"}), want: "/first" + OsSpecificNewline + "/second" + OsSpecificNewline, wantCalls: []setenvCall{{"PATH", "/second" + sep + "/usr/bin"}}},
		{name: "PATH unset", dir: "/opt/tool/bin", env: withFile(nil), want: "/opt/tool/bin" + OsSpecificNewline, wantCalls: []setenvCall{{"PATH", "/opt/tool/bin"}}},
		{name: "PATH empty", dir: "/opt/tool/bin", env: withFile(map[string]string{"PATH": ""}), want: "/opt/tool/bin" + OsSpecificNewline, wantCalls: []setenvCall{{"PATH", "/opt/tool/bin"}}},
		{name: "rejects a line feed", dir: "/a\n/b", env: withFile(nil), wantErr: true},
		{name: "rejects a carriage return", dir: "/a\r/b", env: withFile(nil), wantErr: true},
		{name: "GITHUB_PATH unset", dir: "/opt/tool/bin", env: map[string]string{"PATH": "/usr/bin"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if err := afero.WriteFile(fs, pathFile, []byte(tt.existing), 0o644); err != nil {
				t.Fatalf("failed to create path file: %v", err)
			}

			var calls []setenvCall
			err := addToPath(fs, fakeLookup(tt.env), recordSetenv(&calls), tt.dir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("addToPath(%q) error = %v, wantErr %v", tt.dir, err, tt.wantErr)
			}

			data, _ := afero.ReadFile(fs, pathFile)
			want := tt.want
			if tt.wantErr {
				want = tt.existing
			}
			if diff := cmp.Diff(want, string(data)); diff != "" {
				t.Errorf("path file mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantCalls, calls); diff != "" {
				t.Errorf("process environment changes mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAddToPathWriteFailureLeavesProcessUnchanged(t *testing.T) {
	var calls []setenvCall
	// The memory filesystem has no file at this path, and runner command files are never created by the toolkit.
	err := addToPath(afero.NewMemMapFs(), fakeLookup(map[string]string{SystemPathFilePathEnv: "/missing", "PATH": "/usr/bin"}), recordSetenv(&calls), "/opt/tool/bin")
	if err == nil {
		t.Fatal("expected an error writing to a missing GITHUB_PATH file")
	}
	if len(calls) != 0 {
		t.Errorf("process environment changed after a failed write: %v", calls)
	}
}

func TestSetEnv(t *testing.T) {
	const envFile = "/github_env"
	u := uuid.MustParse("00000000-0000-7000-8000-000000000042")
	delim := "ghadelimiter_" + u.String()
	nl := OsSpecificNewline
	tests := []struct {
		name      string
		env       map[string]string
		createEnv bool
		setenvErr error
		wantFile  string
		wantCalls []setenvCall
		wantErr   bool
	}{
		{name: "writes the file and sets the process", env: map[string]string{EnvironmentVariableFilePathEnv: envFile}, createEnv: true, wantFile: "GREETING<<" + delim + nl + "hello" + nl + delim + nl, wantCalls: []setenvCall{{"GREETING", "hello"}}},
		{name: "GITHUB_ENV unset", env: map[string]string{}, createEnv: true, wantErr: true},
		{name: "file write fails", env: map[string]string{EnvironmentVariableFilePathEnv: envFile}, wantErr: true},
		{name: "setenv fails", env: map[string]string{EnvironmentVariableFilePathEnv: envFile}, createEnv: true, setenvErr: errors.New("setenv failed"), wantFile: "GREETING<<" + delim + nl + "hello" + nl + delim + nl, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.createEnv {
				if err := afero.WriteFile(fs, envFile, nil, 0o644); err != nil {
					t.Fatalf("failed to create env file: %v", err)
				}
			}

			var calls []setenvCall
			setenv := func(key, value string) error {
				if tt.setenvErr != nil {
					return tt.setenvErr
				}
				return recordSetenv(&calls)(key, value)
			}
			err := setEnv(fs, fakeLookup(tt.env), setenv, func() uuid.UUID { return u }, "GREETING", "hello")
			if (err != nil) != tt.wantErr {
				t.Fatalf("setEnv error = %v, wantErr %v", err, tt.wantErr)
			}

			data, _ := afero.ReadFile(fs, envFile)
			if diff := cmp.Diff(tt.wantFile, string(data)); diff != "" {
				t.Errorf("env file mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantCalls, calls); diff != "" {
				t.Errorf("process environment changes mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
