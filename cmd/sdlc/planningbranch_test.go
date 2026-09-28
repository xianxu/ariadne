package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// TestPreparePlanningBranchOverCheckoutRelations generates every relation of
// rest to main (equal, behind, ahead, diverged) × tracked dirtiness × the
// branch planning starts from. Only a clean rest that main contains, or the
// issue branch itself, may proceed; every refusal leaves refs and files intact.
func TestPreparePlanningBranchOverCheckoutRelations(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	const issueBranch = "000009-nine"
	for _, relation := range []string{"equal", "behind", "ahead", "diverged"} {
		for _, dirty := range []bool{false, true} {
			for _, start := range []string{"rest", "issue", "other"} {
				name := fmt.Sprintf("%s/dirty=%v/%s", relation, dirty, start)
				t.Run(name, func(t *testing.T) {
					r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
					if relation == "behind" || relation == "diverged" {
						peerCommit(t, r, "peer.md")
					}
					if relation == "ahead" || relation == "diverged" {
						writeRepoFile(t, r.root, "local.md", "local only\n")
						r.git("add", "local.md")
						r.git("commit", "-qm", "local only")
					}
					switch start {
					case "issue":
						r.git("switch", "-q", "-c", issueBranch)
					case "other":
						r.git("switch", "-q", "-c", "000003-unrelated")
					}
					if dirty {
						writeRepoFile(t, r.root, "README", "edited\n")
					}
					restBefore, headBefore := r.git("rev-parse", "refs/heads/main"), r.git("rev-parse", "HEAD")
					env, err := openTracker(context.Background())
					if err != nil {
						t.Fatal(err)
					}
					res, err := preparePlanningBranch(env, "000009", detailPath)
					accept := start == "issue" || (start == "rest" && !dirty && (relation == "equal" || relation == "behind"))
					if (err == nil) != accept {
						t.Fatalf("accepted=%v, want %v: %v", err == nil, accept, err)
					}
					if r.git("rev-parse", "refs/heads/main") != restBefore {
						t.Fatal("resting ref moved")
					}
					if !accept {
						if r.git("rev-parse", "HEAD") != headBefore {
							t.Fatal("refusal moved HEAD")
						}
						return
					}
					if got := r.git("branch", "--show-current"); got != issueBranch {
						t.Fatalf("on %s after preparation", got)
					}
					if start == "rest" {
						if res != planningCreatedBranch || r.git("rev-parse", "HEAD") != r.originMain() {
							t.Fatal("issue branch not created at pinned main")
						}
						if r.git("show", "HEAD:"+detailPath) != strings.TrimSpace(detail) {
							t.Fatal("issue branch lacks the details")
						}
					}
				})
			}
		}
	}
}

func TestPreparePlanningBranchRefusesMissingDetailsAndReusesBranch(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, nil)
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := preparePlanningBranch(env, "000009", detailPath); err == nil || !strings.Contains(err.Error(), "not on main") {
		t.Fatalf("prepared without details on main: %v", err)
	}
	if r.git("branch", "--list", "000009-nine") != "" {
		t.Fatal("created a branch for an incompletely created issue")
	}
	peerAdd(t, r, detailPath, detail)
	r.git("branch", "000009-nine", "HEAD")
	env, _ = openTracker(context.Background())
	if res, err := preparePlanningBranch(env, "000009", detailPath); err != nil || res != planningSwitchedBranch {
		t.Fatalf("existing branch: %v %v", res, err)
	}
}

func TestStartPlanRequiresClaimAndPreparesBranch(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	var out bytes.Buffer
	if err := startPlanBranch(context.Background(), &out, 9); err == nil || !strings.Contains(err.Error(), "sdlc claim --issue 9") {
		t.Fatalf("planned an unclaimed issue: %v", err)
	}
	var errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatal(err)
	}
	if err := startPlanBranch(context.Background(), &out, 9); err != nil {
		t.Fatal(err)
	}
	if r.git("branch", "--show-current") != "000009-nine" {
		t.Fatal("start-plan did not move design onto the issue branch")
	}
	if !strings.Contains(r.git("show", ":"+detailPath), "status: open") {
		t.Fatal("fixture: committed details should still carry the old mirror")
	}
	if local := r.git("diff", "--", detailPath); !strings.Contains(local, "+status: working") {
		t.Fatalf("mirror not refreshed on the issue branch:\n%s", local)
	}
}

// peerCommit advances origin/main from another clone.
func peerCommit(t *testing.T, r *trackerRepo, file string) {
	t.Helper()
	peerAdd(t, r, file, "peer\n")
}

func peerAdd(t *testing.T, r *trackerRepo, file, body string) {
	t.Helper()
	peer := t.TempDir()
	testfix.Git(t, "", "clone", "-q", r.origin, peer)
	testfix.Git(t, peer, "config", "user.name", "p")
	testfix.Git(t, peer, "config", "user.email", "p@p")
	writeRepoFile(t, peer, file, body)
	testfix.Git(t, peer, "add", "--", file)
	testfix.Git(t, peer, "commit", "-qm", "peer")
	testfix.Git(t, peer, "push", "-q", "origin", "main")
}
