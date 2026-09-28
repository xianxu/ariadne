package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestPreferFreshReadersSurfaceStaleness guards #252's read contract: a
// PreferFresh read may fall back to the last-fetched tracker, so every file
// that asks for one must surface that (read `.Stale` and label its output) —
// found by the slot-cycle e2e, where `issue list` printed stale statuses bare.
// Writes read Fresh and never fall back.
var staleField = regexp.MustCompile(`\.Stale\b`)

func TestPreferFreshReadersSurfaceStaleness(t *testing.T) {
	var files []string
	for _, pattern := range []string{"*.go", "internal/*/*.go"} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, matches...)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || strings.HasPrefix(f, filepath.Join("internal", "tracker")) {
			continue // the tracker package defines the mode and sets Stale
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		src := string(raw)
		if strings.Contains(src, "tracker.PreferFresh") && !staleField.MatchString(src) {
			t.Errorf("%s reads PreferFresh but never surfaces .Stale; label the stale view (staleTrackerNote) or read Fresh", f)
		}
	}
}
