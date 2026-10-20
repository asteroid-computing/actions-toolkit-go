package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"uuid"

	"github.com/spf13/afero"
)

const (
	// GithubStateFilePathEnv points to the env var that stores the state file path.
	GithubStateFilePathEnv = "GITHUB_STATE"
	// EnvironmentVariableFilePathEnv points to the env var that stores the env file path.
	EnvironmentVariableFilePathEnv = "GITHUB_ENV"
	// OutputFilePathEnv points to the env var that stores the output file path.
	OutputFilePathEnv = "GITHUB_OUTPUT"
	// SystemPathFilePathEnv points to the env var that stores the PATH file path.
	SystemPathFilePathEnv = "GITHUB_PATH"
)

// fileCommand represents a GitHub Actions file command targeting a specific file.
// Use [SetEnv], [SetOutput], [SetState], or [AddToPath] to create and send file commands.
type fileCommand struct {
	Delimiter string
	Fs        afero.Fs
	FilePath  string
	Key       string
	Value     string
}

// Send writes the [fileCommand] to its target file.
// GitHub Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#about-workflow-commands
func (fc *fileCommand) Send() error {
	if strings.Contains(fc.Key, fc.Delimiter) || strings.Contains(fc.Value, fc.Delimiter) {
		return errors.New("unexpected input: key and value cannot contain " + fc.Delimiter)
	}

	var contents strings.Builder

	_, err := fmt.Fprintf(&contents, "%s<<%s%s%s%s%s%s", fc.Key, fc.Delimiter, OsSpecificNewline, fc.Value, OsSpecificNewline, fc.Delimiter, OsSpecificNewline)
	if err != nil {
		return err
	}

	return appendToFile(fc.Fs, fc.FilePath, contents.String())
}

// appendToFile appends contents to an existing runner command file.
func appendToFile(fs afero.Fs, filePath, contents string) error {
	f, err := fs.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer func(f afero.File) {
		err := f.Close()
		if err != nil {
			Debug(err.Error())
		}
	}(f)

	_, err = f.WriteString(contents)

	return err
}

// buildFileCommand constructs a [fileCommand] using the provided parameters.
func buildFileCommand(fs afero.Fs, lookupEnv lookupEnvFunc, uuidFunc uuidGeneratorFunc, envVar, key, value string) (*fileCommand, error) {
	filePath, ok := lookupEnv(envVar)

	if !ok {
		return nil, errors.New("unable to find environment variable for file command")
	}

	return &fileCommand{
		Fs:        fs,
		Delimiter: "ghadelimiter_" + uuidFunc().String(),
		FilePath:  filePath,
		Key:       key,
		Value:     value,
	}, nil
}

// SetEnv executes a GitHub Workflow command to set an environment variable in subsequent workflows steps,
// and sets it in the current process so code later in this step sees it too, as the JS toolkit's exportVariable does.
// Unlike exportVariable, the current process is changed only after the runner file is written, so a failed call changes nothing.
// GitHub Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-an-environment-variable
func SetEnv(key, value string) error {
	return setEnv(afero.NewOsFs(), os.LookupEnv, os.Setenv, uuid.NewV7, key, value)
}

// setEnv is the testable implementation of [SetEnv].
func setEnv(fs afero.Fs, lookupEnv lookupEnvFunc, setenv setenvFunc, uuidFunc uuidGeneratorFunc, key, value string) error {
	fc, err := buildFileCommand(fs, lookupEnv, uuidFunc, EnvironmentVariableFilePathEnv, key, value)
	if err != nil {
		return err
	}

	if err := fc.Send(); err != nil {
		return err
	}

	return setenv(key, value)
}

// SetOutput executes a GitHub Workflow command to set a step output.
// GitHub Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-an-output-parameter
func SetOutput(key, value string) error {
	fc, err := buildFileCommand(afero.NewOsFs(), os.LookupEnv, uuid.NewV7, OutputFilePathEnv, key, value)

	if err != nil {
		return err
	}

	return fc.Send()
}

// SetOutputJSON sets a step output to v encoded as JSON, for a workflow to read with fromJSON(steps.<id>.outputs.<name>), as the JS toolkit's setOutput does for a value that is not a string.
// v is encoded with encoding/json, so it follows the same struct tags as elsewhere; map keys are sorted, and <, > and & are written as \u003c, \u003e and \u0026, which fromJSON reads back unchanged.
// A string is encoded as a quoted JSON string; use [SetOutput] to set a plain string output.
// GitHub Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-an-output-parameter
func SetOutputJSON(key string, v any) error {
	value, err := jsonCommandValue(v)
	if err != nil {
		return err
	}

	return SetOutput(key, value)
}

// jsonCommandValue encodes v as the JSON text of a command value.
func jsonCommandValue(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encoding value as JSON: %w", err)
	}

	return string(b), nil
}

// SetState executes a GitHub Workflow command to set a state parameter which is available to post-actions
// GitHub Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#sending-values-to-the-pre-and-post-actions
func SetState(key, value string) error {
	fc, err := buildFileCommand(afero.NewOsFs(), os.LookupEnv, uuid.NewV7, GithubStateFilePathEnv, key, value)

	if err != nil {
		return err
	}

	return fc.Send()
}

// GetState returns the value saved with [SetState] under name, read from the STATE_<name> environment variable the runner sets for the action's pre and post steps.
// It returns an empty string when no such state was saved.
// GitHub Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#sending-values-to-the-pre-and-post-actions
func GetState(name string) string {
	return getState(os.LookupEnv, name)
}

// getState is the testable implementation of [GetState].
func getState(lookupEnv lookupEnvFunc, name string) string {
	v, _ := lookupEnv("STATE_" + name)
	return v
}

// AddToPath executes a GitHub Workflow command to add a directory to the system PATH for subsequent workflow steps,
// and prepends it to the current process's PATH so code later in this step finds its executables too, as the JS toolkit's addPath does.
// Unlike addPath, the current process is changed only after the runner file is written, so a failed call changes nothing.
// The runner reads GITHUB_PATH as one directory per line, so dir must not contain a line break.
// GitHub Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#adding-a-system-path
func AddToPath(dir string) error {
	return addToPath(afero.NewOsFs(), os.LookupEnv, os.Setenv, dir)
}

// addToPath is the testable implementation of [AddToPath]. The current PATH is read through lookupEnv.
func addToPath(fs afero.Fs, lookupEnv lookupEnvFunc, setenv setenvFunc, dir string) error {
	if strings.ContainsAny(dir, "\r\n") {
		return errors.New("unexpected input: path cannot contain a line break")
	}

	filePath, ok := lookupEnv(SystemPathFilePathEnv)
	if !ok {
		return errors.New("unable to find environment variable for file command")
	}

	if err := appendToFile(fs, filePath, dir+OsSpecificNewline); err != nil {
		return err
	}

	path := dir
	if current, _ := lookupEnv("PATH"); current != "" {
		path = dir + string(os.PathListSeparator) + current
	}

	return setenv("PATH", path)
}
