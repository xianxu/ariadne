package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestTrackerGitResultsKeepTheirErrors guards the silent-error family behind
// #252 BR-14/BR-16: no trackerEnv.git result may discard its error. A probe goes
// through gitTest/has; everything else propagates the failure.
func TestTrackerGitResultsKeepTheirErrors(t *testing.T) {
	dropped := regexp.MustCompile(`(,\s*_\s*:?=|^\s*_\s*=)\s*(env|e)\.git\(`)
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range files {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(raw), "\n") {
			if dropped.MatchString(line) {
				t.Errorf("%s:%d discards a Git error: %s", p, i+1, strings.TrimSpace(line))
			}
		}
	}
}
