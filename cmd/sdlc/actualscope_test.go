package main

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
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

// BR-3: `sdlc active-time --branch-point` applies the same scope as `sdlc
// actual`, measuring the first --issue, and refuses without an --issue.
func TestActiveTimeScope(t *testing.T) {
	got, err := activeTimeScope("abc123", []string{"9", "5"})
	if err != nil || got != (activetime.Scope{BranchPoint: "abc123", Issue: "9"}) {
		t.Errorf("activeTimeScope = %+v, %v; want branch point abc123, issue 9", got, err)
	}
	if got, err := activeTimeScope("", []string{"9"}); err != nil || got != (activetime.Scope{}) {
		t.Errorf("no branch point: %+v, %v; want unscoped", got, err)
	}
	if _, err := activeTimeScope("abc123", nil); err == nil || !strings.Contains(err.Error(), "--issue") {
		t.Errorf("branch point without --issue: err = %v, want a refusal naming --issue", err)
	}
	cmd := NewActiveTimeCmd()
	cmd.SetArgs([]string{"--git-repo", t.TempDir(), "--branch-point", "abc123"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--branch-point needs --issue") {
		t.Errorf("CLI with --branch-point and no --issue: err = %v, want the refusal", err)
	}
}
