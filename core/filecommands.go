package core

import (
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

// SetEnv executes a GitHub Workflow command to set an environment variable in subsequent workflows steps.
// GitHub Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#setting-an-environment-variable
func SetEnv(key, value string) error {
	fc, err := buildFileCommand(afero.NewOsFs(), os.LookupEnv, uuid.NewV7, EnvironmentVariableFilePathEnv, key, value)

	if err != nil {
		return err
	}

	return fc.Send()
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

// SetState executes a GitHub Workflow command to set a state parameter which is available to post-actions
// GitHub Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#sending-values-to-the-pre-and-post-actions
func SetState(key, value string) error {
	fc, err := buildFileCommand(afero.NewOsFs(), os.LookupEnv, uuid.NewV7, GithubStateFilePathEnv, key, value)

	if err != nil {
		return err
	}

	return fc.Send()
}

// AddToPath executes a GitHub Workflow command to add a directory to the system PATH for subsequent workflow steps.
// The runner reads GITHUB_PATH as one directory per line, so dir must not contain a line break.
// GitHub Command Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#adding-a-system-path
func AddToPath(dir string) error {
	return addToPath(afero.NewOsFs(), os.LookupEnv, dir)
}

// addToPath is the testable implementation of [AddToPath].
func addToPath(fs afero.Fs, lookupEnv lookupEnvFunc, dir string) error {
	if strings.ContainsAny(dir, "\r\n") {
		return errors.New("unexpected input: path cannot contain a line break")
	}

	filePath, ok := lookupEnv(SystemPathFilePathEnv)
	if !ok {
		return errors.New("unable to find environment variable for file command")
	}

	return appendToFile(fs, filePath, dir+OsSpecificNewline)
}
