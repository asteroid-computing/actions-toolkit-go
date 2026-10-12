// Package gocmd runs child processes — mostly the go toolchain — for walle and the Go GitHub Actions.
//
// It is the one place the project's child-process environment is decided. [Env] holds the overrides every go invocation needs. Put such a setting here once and every build, vet, test and `go list` inherits it.
//
// Every function takes a context so Ctrl-C (walle wires signal.NotifyContext) or a cancelled errgroup stops the child instead of orphaning it.
package gocmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Env is the project-wide environment overlay applied to every child process. Empty by default; add entries such as "GOEXPERIMENT": "jsonv2" or "GOPRIVATE": "github.com/your-org/" here and nowhere else.
var Env = map[string]string{}

// Run executes name with args, streaming stdout and stderr to the parent. extra is layered on top of [Env]; it may be nil.
func Run(ctx context.Context, extra map[string]string, name string, args ...string) error {
	cmd := command(ctx, extra, name, args...)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", describe(name, args), err)
	}
	return nil
}

// Output executes name with args and returns its stdout. Stderr streams to the parent so tool diagnostics stay visible.
func Output(ctx context.Context, extra map[string]string, name string, args ...string) (string, error) {
	out, err := command(ctx, extra, name, args...).Output()
	if err != nil {
		return "", fmt.Errorf("%s: %w", describe(name, args), err)
	}
	return string(out), nil
}

// Quiet executes name with args, keeping stdout out of the parent's output: on failure any stdout is folded into the error. Used where a caller renders its own report (the build action writes a step summary) and does not want console noise.
func Quiet(ctx context.Context, extra map[string]string, name string, args ...string) error {
	out, err := command(ctx, extra, name, args...).Output()
	if err != nil {
		if msg := strings.TrimSpace(string(out)); msg != "" {
			return fmt.Errorf("%s: %w: %s", describe(name, args), err, msg)
		}
		return fmt.Errorf("%s: %w", describe(name, args), err)
	}
	return nil
}

func command(ctx context.Context, extra map[string]string, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	for k, v := range Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	for k, v := range extra {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	return cmd
}

// describe renders a command line for an error message, cutting multi-line arguments (go list templates) to their first line.
func describe(name string, args []string) string {
	parts := []string{name}
	for _, a := range args {
		if first, _, multi := strings.Cut(a, "\n"); multi {
			a = first + "…"
		}
		parts = append(parts, a)
	}
	return strings.Join(parts, " ")
}
