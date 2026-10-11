package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestGitDrivingPackagesPreferRealGit: every cmd/sdlc package whose tests spawn
// git must call testfix.PreferRealGit() from its TestMain, or its git spawns go
// through the macOS xcrun shim at about 3× the cost (#253). Source guard.
func TestGitDrivingPackagesPreferRealGit(t *testing.T) {
	spawnsGit := regexp.MustCompile(`exec\.Command(Context)?\([^)]*"git"|testfix\.(Git|Capture|Repo)\(`)
	dirs := map[string][3]bool{} // dir → {spawns git, calls PreferRealGit, takes the test gate}
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == "testdata" {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		dir := filepath.Dir(path)
		st := dirs[dir]
		st[0] = st[0] || spawnsGit.Match(src)
		st[1] = st[1] || strings.Contains(string(src), "testfix.PreferRealGit()")
		st[2] = st[2] || strings.Contains(string(src), "testgate.Enter()")
		dirs[dir] = st
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for dir, st := range dirs {
		if dir == filepath.Join("internal", "testfix") {
			continue // defines the helper
		}
		if st[0] && !st[1] {
			t.Errorf("%s: tests spawn git but no TestMain calls testfix.PreferRealGit()", dir)
		}
		// #324: a git-driving package is a heavy one; it takes a machine-wide test slot.
		if st[0] && !st[2] {
			t.Errorf("%s: tests spawn git but no TestMain calls testgate.Enter() (#324)", dir)
		}
	}
}
