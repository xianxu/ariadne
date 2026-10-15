package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// #367 (ARCH-PURPOSE): every commit sdlc makes carries the Agent trailer. A
// line passing "commit" or "commit-tree" to git must stamp its message through
// agentid.StampCurrent, or be one of the reviewed exceptions below, so a new
// commit site can't skip the trailer silently.
func TestEveryCommitSiteStamps(t *testing.T) {
	exempt := map[string]string{
		`internal/gitx/commitpublication.go|"commit-tree", "-S"}`:  "re-slices the args built (and stamped) on the line above",
		`abandon.go|"commit", "-q", "--no-verify", "--no-edit"`:    "concludes a merge; merges carry no work and the close report skips them",
		`internal/gitx/recoveryref.go|"commit-tree"`:               "a recovery ref snapshot, never history",
		`internal/gitx/rebasedpatch.go|"commit-tree"`:              "a throwaway commit for a patch-id",
		`issuepublish.go|StringVar(&source, "commit"`:              "a flag name, not a git call",
		`activetime.go|"commit", "issues"`:                         "a table header",
		`internal/testfix/testfix.go|"commit", "-q", "-m", "init"`: "test fixture helper",
	}
	site := regexp.MustCompile(`"commit(-tree)?",`)
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(body), "\n") {
			if !site.MatchString(line) || strings.Contains(line, "agentid.StampCurrent(") {
				continue
			}
			known := false
			for key := range exempt {
				file, snippet, _ := strings.Cut(key, "|")
				if filepath.ToSlash(path) == file && strings.Contains(line, snippet) {
					known = true
				}
			}
			if !known {
				t.Errorf("%s:%d makes a commit without agentid.StampCurrent: %s", path, i+1, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
