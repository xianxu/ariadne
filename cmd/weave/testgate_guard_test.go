package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestGitDrivingPackagesTakeTheTestGate: every cmd/weave package whose tests spawn
// git calls testgate.Enter() from its TestMain, so its runs share the machine-wide
// test budget (#324), the rule TestGitDrivingPackagesPreferRealGit enforces for
// cmd/sdlc. Source guard.
func TestGitDrivingPackagesTakeTheTestGate(t *testing.T) {
	spawnsGit := regexp.MustCompile(`exec\.Command(Context)?\([^)]*"git"`)
	dirs := map[string][2]bool{} // dir → {spawns git, takes the gate}
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == "testdata" {
			return filepath.SkipDir
		}
		// Skip this guard's own file: it names the calls it looks for, and would
		// otherwise satisfy itself for its own directory.
		if d.IsDir() || !strings.HasSuffix(path, "_test.go") || path == "testgate_guard_test.go" {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		st := dirs[filepath.Dir(path)]
		st[0] = st[0] || spawnsGit.Match(src)
		st[1] = st[1] || strings.Contains(string(src), "testgate.Enter()")
		dirs[filepath.Dir(path)] = st
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for dir, st := range dirs {
		if st[0] && !st[1] {
			t.Errorf("%s: tests spawn git but no TestMain calls testgate.Enter() (#324)", dir)
		}
	}
}
