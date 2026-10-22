package main

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestIssueInitRefusesPreexistingLocalTrackerBranch(t *testing.T) {
	for _, root := range []string{"", "workshop/issues", "workshop/history", "workshop/history/issues"} {
		t.Run("root="+root, func(t *testing.T) {
			r := emptyInitRepo(t)
			r.git("switch", "-qc", "issue-tracker")
			if root != "" {
				writeRepoFile(t, r.root, root+"/000001-old.md", "---\nid: 000001\nstatus: done\n---\n\n# Old work\n")
				r.git("add", ".")
				r.git("commit", "-qm", "reserved branch with legacy evidence")
			}
			r.git("switch", "-q", "main")
			main, head, status := r.originMain(), r.git("rev-parse", "HEAD"), r.git("status", "--porcelain", "--untracked-files=all")
			_, stderr, err := executeSDLCTestCommand("issue", "init")
			if err == nil || !strings.Contains(err.Error()+stderr, "tracker already initialized") {
				t.Fatalf("unverified reserved branch accepted: %v %s", err, stderr)
			}
			if r.originMain() != main || testfix.Capture(t, r.origin, "branch", "--list", "issue-tracker") != "" {
				t.Fatal("reserved-branch refusal published")
			}
			assertCheckoutUntouched(t, r, head, status)
		})
	}
}

func TestIssueInitRefusesUnverifiedRemoteTrackerBranch(t *testing.T) {
	r := emptyInitRepo(t)
	r.git("switch", "-qc", "issue-tracker")
	writeRepoFile(t, r.root, "unknown-tracker.txt", "not a verified init bootstrap\n")
	r.git("add", ".")
	r.git("commit", "-qm", "unverified reserved branch")
	r.git("push", "-q", "origin", "issue-tracker")
	r.git("switch", "-q", "main")
	r.git("branch", "-D", "issue-tracker")
	main := r.originMain()
	refs := testfix.Capture(t, r.origin, "show-ref")
	_, stderr, err := executeSDLCTestCommand("issue", "init")
	if err == nil || !strings.Contains(err.Error()+stderr, "tracker already initialized") {
		t.Fatalf("unverified reserved branch accepted: %v %s", err, stderr)
	}
	if r.originMain() != main || testfix.Capture(t, r.origin, "show-ref") != refs {
		t.Fatal("unverified reserved branch changed")
	}
}
