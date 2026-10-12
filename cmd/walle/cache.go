package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/asteroid-computing/actions-toolkit-go/internal/gocache"
	"github.com/asteroid-computing/actions-toolkit-go/internal/gocmd"
)

type cacheCmd struct {
	Trim cacheTrimCmd `cmd:"" help:"Delete Go build-cache entries not used since a run started, so a saved CI cache holds only the working set."`
}

// cacheTrimCmd runs in CI just before the cache is saved. Locally it is harmless but rarely useful: Go's own five-day trim is the right policy for a long-lived machine.
type cacheTrimCmd struct {
	Since int64         `required:"" help:"Run start time, Unix seconds (CI records it with date +%s before restoring the cache)."`
	Slack time.Duration `default:"2h" help:"Keep entries last used this long before --since. Must be at least Go's 1h mtime refresh interval, or entries used this run but not re-touched would be deleted."`
	Dir   string        `help:"Build cache to trim. Default: go env GOCACHE."`
}

func (c cacheTrimCmd) Run(ctx context.Context) error {
	if c.Slack < gocache.MtimeInterval {
		return fmt.Errorf("--slack %s is shorter than Go's mtime refresh interval (%s); entries used this run would be deleted", c.Slack, gocache.MtimeInterval)
	}
	dir := c.Dir
	if dir == "" {
		out, err := gocmd.Output(ctx, nil, "go", "env", "GOCACHE")
		if err != nil {
			return err
		}
		dir = strings.TrimSpace(out)
	}
	cutoff := time.Unix(c.Since, 0).Add(-c.Slack)
	st, err := gocache.Trim(dir, cutoff)
	if err != nil {
		return err
	}
	fmt.Printf("GOCACHE trim (cutoff %s): %s\n", cutoff.UTC().Format(time.RFC3339), st)
	return nil
}
