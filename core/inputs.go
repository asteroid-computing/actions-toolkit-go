package core

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrInputRequired is returned by [GetInput], [GetBooleanInput], and [GetMultilineInput]
// when an input is declared required but the resolved value is empty.
// Use errors.Is to test for this sentinel.
var ErrInputRequired = errors.New("input required and not supplied")

// ErrInvalidBooleanInput is returned by [GetBooleanInput] when the input value
// does not match any YAML 1.2 "Core Schema" boolean string
// ("true", "True", "TRUE", "false", "False", "FALSE").
// Use errors.Is to test for this sentinel.
var ErrInvalidBooleanInput = errors.New("input does not meet YAML 1.2 \"Core Schema\" boolean specification")

// InputOptions controls how an input value is read.
type InputOptions struct {
	// Required reports whether the input must be set.
	// GetInput returns [ErrInputRequired] when Required is true and the input is empty.
	Required bool
	// TrimWhitespace controls whitespace trimming.
	// A nil value (the default) or a pointer to true trims surrounding whitespace from the value.
	// A pointer to false preserves the raw value exactly as set in the environment.
	TrimWhitespace *bool
}

// GetInput reads a GitHub Actions input by name, applying the options in opts.
// The input is resolved from the environment variable INPUT_<NAME>,
// where <NAME> is the input name with spaces replaced by underscores and converted to uppercase.
// Hyphens in the name are preserved.
// If opts is nil it is treated as a zero InputOptions (not required, trim whitespace).
// GetInput returns [ErrInputRequired] (wrapped) when opts.Required is true and the input is unset or empty.
// GitHub Actions Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/variables
func GetInput(name string, opts *InputOptions) (string, error) {
	return getInput(os.LookupEnv, name, opts)
}

// GetBooleanInput reads a GitHub Actions input by name and interprets it as a boolean.
// It accepts only the YAML 1.2 "Core Schema" boolean strings:
// "true", "True", "TRUE" evaluate to true;
// "false", "False", "FALSE" evaluate to false.
// Any other value returns [ErrInvalidBooleanInput] (wrapped).
// GetBooleanInput propagates [ErrInputRequired] from GetInput when opts.Required is true.
// GitHub Actions Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/variables
func GetBooleanInput(name string, opts *InputOptions) (bool, error) {
	return getBooleanInput(os.LookupEnv, name, opts)
}

// GetMultilineInput reads a GitHub Actions input by name and splits the value on newlines.
// Empty lines are dropped from the result.
// Each line is trimmed of surrounding whitespace unless opts.TrimWhitespace points to false.
// GetMultilineInput returns a nil slice when the input is unset or resolves to no non-empty lines.
// GetMultilineInput propagates [ErrInputRequired] from GetInput when opts.Required is true.
// GitHub Actions Reference: https://docs.github.com/en/actions/reference/workflows-and-actions/variables
func GetMultilineInput(name string, opts *InputOptions) ([]string, error) {
	return getMultilineInput(os.LookupEnv, name, opts)
}

// getInput is the testable implementation of [GetInput].
// lookupEnv is called with the resolved env var key INPUT_<NAME>.
func getInput(lookupEnv lookupEnvFunc, name string, opts *InputOptions) (string, error) {
	if opts == nil {
		opts = &InputOptions{}
	}
	envKey := "INPUT_" + strings.ToUpper(strings.ReplaceAll(name, " ", "_"))
	val, _ := lookupEnv(envKey)
	// Required check runs on the raw value, before any trimming —
	// matching upstream @actions/core which checks !val before the trim branch.
	if opts.Required && val == "" {
		return "", fmt.Errorf("%w: %s", ErrInputRequired, name)
	}
	if opts.TrimWhitespace != nil && !*opts.TrimWhitespace {
		return val, nil
	}
	return strings.TrimSpace(val), nil
}

// getBooleanInput is the testable implementation of [GetBooleanInput].
func getBooleanInput(lookupEnv lookupEnvFunc, name string, opts *InputOptions) (bool, error) {
	val, err := getInput(lookupEnv, name, opts)
	if err != nil {
		return false, err
	}
	switch val {
	case "true", "True", "TRUE":
		return true, nil
	case "false", "False", "FALSE":
		return false, nil
	default:
		return false, fmt.Errorf("%w: %s", ErrInvalidBooleanInput, name)
	}
}

// getMultilineInput is the testable implementation of [GetMultilineInput].
func getMultilineInput(lookupEnv lookupEnvFunc, name string, opts *InputOptions) ([]string, error) {
	if opts == nil {
		opts = &InputOptions{}
	}
	val, err := getInput(lookupEnv, name, opts)
	if err != nil {
		return nil, err
	}
	shouldTrim := opts.TrimWhitespace == nil || *opts.TrimWhitespace
	var out []string
	for _, line := range strings.Split(val, "\n") {
		if line == "" {
			continue
		}
		if shouldTrim {
			out = append(out, strings.TrimSpace(line))
		} else {
			out = append(out, line)
		}
	}
	return out, nil
}
