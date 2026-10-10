package main

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// TestActualScope pins computeActual's boundary-scope wiring (#270): scoped at
// the branch point on a diverged branch and on the issue's own branch before
// its first commit (where the branch point is HEAD), unscoped on main.
func TestActualScope(t *testing.T) {
	dir := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	git := func(args ...string) string { return testfix.Git(t, dir, args...) }
	base := strings.TrimSpace(git("rev-parse", "HEAD"))

	if got := actualScope("9"); got.BranchPoint != "" || got.Issue != "9" {
		t.Errorf("on main: %+v, want unscoped", got)
	}
	git("switch", "-q", "-c", "000009-fix")
	if got := actualScope("9"); got.BranchPoint != base {
		t.Errorf("issue branch before its first commit: %+v, want branch point %s", got, base)
	}
	if got := actualScope("10"); got.BranchPoint != "" {
		t.Errorf("another issue's undiverged branch: %+v, want unscoped", got)
	}
	git("commit", "-q", "--allow-empty", "-m", "#9 work")
	if got := actualScope("9"); got.BranchPoint != base {
		t.Errorf("diverged issue branch: %+v, want branch point %s", got, base)
	}
}
