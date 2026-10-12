// Package gocache trims the Go build cache to the working set of one CI run, so the cache CI saves stays the size of what a build actually uses instead of accreting a layer of stale compile output with every save.
//
// It relies on how cmd/go maintains the cache (src/cmd/go/internal/cache/cache.go): every use of an entry refreshes its mtime, but at most once per mtimeInterval (one hour). So after a run, every entry that run read or wrote has an mtime no earlier than an hour before the run started; anything older went unused. Go's own trim deletes entries unused for five days, which never fires in a cache CI restores and re-saves daily.
//
// [Trim] removes exactly what Go's own trim would remove — files and executable-entry directories named *-a or *-d inside the 256 two-hex-digit subdirectories — only with a much shorter cutoff. README, trim.txt, testexpire.txt and fuzz/ are never touched.
//
// The idea comes from cloudx-io/setup-go's trim-gocache post step; it is reimplemented here so the template takes no third-party action and needs no post hook (CI saves the cache in an explicit step, so the trim simply runs before it).
package gocache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MtimeInterval is cmd/go's mtimeInterval: an entry's mtime is refreshed on use only if it is at least this old. A cutoff must leave at least this much slack before the run's start.
const MtimeInterval = time.Hour

// Stats reports what a trim kept and removed.
type Stats struct {
	Kept, Removed           int
	KeptBytes, RemovedBytes int64
}

func (s Stats) String() string {
	mb := func(b int64) float64 { return float64(b) / (1 << 20) }
	return fmt.Sprintf("kept %d entries (%.0f MB), removed %d entries (%.0f MB)", s.Kept, mb(s.KeptBytes), s.Removed, mb(s.RemovedBytes))
}

// Trim deletes cache entries in dir whose mtime is before cutoff. Errors on individual entries are ignored, as cmd/go's own trim ignores them: a file that cannot be removed only means a slightly larger cache.
func Trim(dir string, cutoff time.Time) (Stats, error) {
	var st Stats
	if _, err := os.Stat(dir); err != nil {
		return st, fmt.Errorf("GOCACHE %s: %w", dir, err)
	}
	for i := range 256 {
		subdir := filepath.Join(dir, fmt.Sprintf("%02x", i))
		entries, err := os.ReadDir(subdir)
		if err != nil {
			continue // missing subdirectory: nothing cached under this prefix
		}
		for _, e := range entries {
			name := e.Name()
			if !strings.HasSuffix(name, "-a") && !strings.HasSuffix(name, "-d") {
				continue
			}
			path := filepath.Join(subdir, name)
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			size := entrySize(path, info)
			if !info.ModTime().Before(cutoff) {
				st.Kept++
				st.KeptBytes += size
				continue
			}
			if os.RemoveAll(path) == nil { // RemoveAll: executable entries (go run / go tool) are directories
				st.Removed++
				st.RemovedBytes += size
			}
		}
	}
	return st, nil
}

func entrySize(path string, info os.FileInfo) int64 {
	if !info.IsDir() {
		return info.Size()
	}
	var n int64
	filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if fi, err := d.Info(); err == nil {
				n += fi.Size()
			}
		}
		return nil
	})
	return n
}
