package gocache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTrim(t *testing.T) {
	dir := t.TempDir()
	start := time.Now()
	old := start.Add(-3 * time.Hour)
	recent := start.Add(-30 * time.Minute) // used this run but not re-touched: within mtimeInterval

	write := func(rel string, mtime time.Time) string {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, mtime, mtime); err != nil {
			t.Fatal(err)
		}
		return p
	}
	staleAction := write("0a/aaaa-a", old)
	staleOutput := write("0a/bbbb-d", old)
	keptRecent := write("ff/cccc-d", recent)
	keptNew := write("ff/dddd-a", start.Add(time.Minute))
	notAnEntry := write("0a/something-else", old)
	readme := write("README", old)
	trimTxt := write("trim.txt", old)
	fuzz := write("fuzz/pkg/FuzzX/corpus-d", old)

	// An executable cache entry is a directory; its own mtime decides.
	exe := filepath.Join(dir, "1b/eeee-d")
	write("1b/eeee-d/hello", old)
	if err := os.Chtimes(exe, old, old); err != nil {
		t.Fatal(err)
	}

	st, err := Trim(dir, start.Add(-2*MtimeInterval))
	if err != nil {
		t.Fatal(err)
	}
	if st.Removed != 3 || st.Kept != 2 {
		t.Errorf("stats = %+v, want 3 removed, 2 kept", st)
	}
	for _, p := range []string{staleAction, staleOutput, exe} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s survived the trim", p)
		}
	}
	for _, p := range []string{keptRecent, keptNew, notAnEntry, readme, trimTxt, fuzz} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("%s was removed: %v", p, err)
		}
	}
}

func TestTrimMissingDir(t *testing.T) {
	if _, err := Trim(filepath.Join(t.TempDir(), "absent"), time.Now()); err == nil {
		t.Error("expected an error for a missing GOCACHE")
	}
}
