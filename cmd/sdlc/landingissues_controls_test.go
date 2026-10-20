package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

// Keep readiness fault injection local: the underlying fake still owns real
// remote refs, exact-head statuses, and the expected-head server merge.
type issueRecordControlGH struct {
	*landingFakeGH
	readiness func(landingReadiness) landingReadiness
}

func (g *issueRecordControlGH) LandingReadiness(ctx context.Context, repo string, n int) (landingReadiness, error) {
	r, err := g.landingFakeGH.LandingReadiness(ctx, repo, n)
	if err == nil && g.readiness != nil {
		r = g.readiness(r)
	}
	return r, err
}

func issueRecordControlPreserved(t *testing.T, root string, gh *landingFakeGH, head string) {
	t.Helper()
	if gh.merges != 0 || procedureHead(t, root) != head || git(t, gh.remote, "rev-parse", "refs/heads/"+landingTestBranch) != head {
		t.Fatalf("refusal changed heads or merged: merges=%d local=%s want=%s", gh.merges, procedureHead(t, root), head)
	}
	if got := git(t, root, "branch", "--show-current"); got != landingTestBranch {
		t.Fatalf("refusal left branch %q", got)
	}
	if err := landingClean(mergeRunner, root); err != nil {
		t.Fatalf("refusal left unclean branch: %v", err)
	}
}

func TestLandingIssueRecordControlCodeAfterCleanSnapshot(t *testing.T) {
	for _, mixed := range []bool{false, true} {
		name := "code"
		if mixed {
			name = "mixed"
		}
		t.Run(name, func(t *testing.T) {
			root, server, gh := integrationFixture(t)
			gh.readinessStatus = "CLEAN"
			var tested []string
			gh.afterStatus = func(head string) { tested = append(tested, head) }
			moved := ""
			gh.afterReadiness = func() {
				if moved != "" {
					return
				}
				if mixed {
					procedureWrite(t, server, "workshop/issues/000700-filed.md", strings.Replace(instanceGateGood, "id: 2", "id: 000700", 1))
				}
				moved = integrationAdvance(t, server, "late-code.go", "package late\n")
			}
			var out bytes.Buffer
			if err := runMerge(io.Discard, &out, landingFlags()); err != nil {
				t.Fatalf("%v\n%s", err, &out)
			}
			if moved == "" || len(tested) != 2 || tested[0] == tested[1] || gh.pr.HeadOID != tested[1] || gh.merges != 1 {
				t.Fatalf("stale CLEAN bypassed integration: moved=%s tested=%v head=%s merges=%d\n%s", moved, tested, gh.pr.HeadOID, gh.merges, &out)
			}
			if !gitSucceeds(root, "merge-base", "--is-ancestor", moved, tested[1]) || gh.ci.statuses[tested[1]][localStatusContext(suiteCheckName)] != "success" {
				t.Fatalf("latest code was not included in exact-head suite: moved=%s tested=%v statuses=%v", moved, tested, gh.ci.statuses)
			}
		})
	}
}

func TestLandingIssueRecordControlReadinessHeadMismatch(t *testing.T) {
	root, server, gh := integrationFixture(t)
	gh.readinessStatus = "CLEAN"
	var tested string
	gh.afterStatus = func(head string) {
		tested = head
		integrationAdvance(t, server, "workshop/issues/000700-filed.md", "issue record\n")
	}
	ghClient = &issueRecordControlGH{landingFakeGH: gh, readiness: func(r landingReadiness) landingReadiness {
		r.HeadOID = strings.Repeat("a", 40)
		return r
	}}
	var out bytes.Buffer
	err := runMerge(io.Discard, &out, landingFlags())
	if err == nil || !strings.Contains(err.Error(), "identity") {
		t.Fatalf("head mismatch was not refused: %v\n%s", err, &out)
	}
	if tested == "" {
		t.Fatal("readiness mismatch did not follow the exact-head suite")
	}
	issueRecordControlPreserved(t, root, gh, tested)
}

func TestLandingIssueRecordControlSameRecordConflict(t *testing.T) {
	root, server, gh := integrationFixture(t)
	gh.readinessStatus = "CLEAN"
	var tested string
	gh.afterStatus = func(head string) {
		tested = head
		integrationAdvance(t, server, procedureIssue, "---\nid: 000001\nstatus: working\n---\n# Main changed the same issue\n")
	}
	var out bytes.Buffer
	err := runMerge(io.Discard, &out, landingFlags())
	if err == nil || !strings.Contains(err.Error(), "conflict") || !strings.Contains(err.Error(), procedureIssue) {
		t.Fatalf("record conflict did not use ordinary conflict refusal: %v\n%s", err, &out)
	}
	issueRecordControlPreserved(t, root, gh, tested)
}

func TestLandingIssueRecordControlLateDuplicateID(t *testing.T) {
	root, server, gh := integrationFixture(t)
	gh.readinessStatus = "CLEAN"
	// These issues are open, valid section-bearing records. Leave the actual
	// instance/section and duplicate-ID gates enabled. As in the instance-gate
	// fixtures, stub only the external vocabulary executable: the disposable
	// worktree has no vocabulary installation or construct tree of its own.
	stubConformingFrontmatter(t)
	procedureWrite(t, root, procedureIssue, "---\nid: 000001\nstatus: open\nupdated: 2026-10-19\n---\n# Procedure\n\n## Spec\nLanding controls.\n\n## Done when\n- Landing preserves issue IDs.\n\n## Plan\n- [ ] Verify landing.\n\n## Log\n")
	procedureWrite(t, root, "workshop/issues/000700-branch.md", strings.Replace(instanceGateGood, "id: 2", "id: 000700", 1))
	git(t, root, "add", "workshop/issues")
	git(t, root, "commit", "-qm", "valid branch issue records")
	git(t, root, "push", "-q", "upstream", landingTestBranch)
	var tested string
	gh.afterStatus = func(head string) {
		tested = head
		integrationAdvance(t, server, "workshop/issues/000700-main.md", strings.Replace(instanceGateGood, "id: 2", "id: 000700", 1))
	}
	f := landingFlags()
	f.NoValidate = false
	var out bytes.Buffer
	err := runMerge(&out, &out, f)
	if err == nil || !strings.Contains(err.Error(), "duplicate issue ID") || !strings.Contains(err.Error(), "000700") {
		t.Fatalf("late duplicate did not reach duplicate-ID refusal: %v\n%s", err, &out)
	}
	if tested == "" {
		t.Fatalf("duplicate appeared before the initial gates passed:\n%s", &out)
	}
	issueRecordControlPreserved(t, root, gh, tested)
}

// Registration creates checks on the new head; green checks on the old head
// must not be copied across even when the local suite can be reused.
type issueRecordPendingGH struct {
	*landingFakeGH
	readyHead, pendingHead string
}

func (g *issueRecordPendingGH) LandingPRs(ctx context.Context, repo, branch string) ([]landingPR, error) {
	prs, err := g.landingFakeGH.LandingPRs(ctx, repo, branch)
	if err == nil && g.readyHead != "" && g.pr.HeadOID != g.readyHead && g.pendingHead == "" {
		g.pendingHead = g.pr.HeadOID
		g.ci.setHead(g.pr.Number, g.pendingHead, prCheck{Name: "build", Bucket: "pending"})
	}
	return prs, err
}
func TestLandingIssueRecordFallbackWaitsForNewHeadCI(t *testing.T) {
	_, server, base := integrationFixture(t)
	g := &issueRecordPendingGH{landingFakeGH: base}
	ghClient = g
	g.readinessStatus = "BEHIND"
	suites := 0
	g.afterStatus = func(head string) {
		suites++
		g.readyHead = head
		g.ci.setHead(g.pr.Number, head, prCheck{Name: "build", Bucket: "pass"})
		integrationAdvance(t, server, "workshop/plans/000700-plan.md", "record\n")
	}
	oldPause := ciWaitPause
	t.Cleanup(func() { ciWaitPause = oldPause })
	pauses := 0
	ciWaitPause = func(context.Context, time.Duration) error {
		pauses++
		if g.pendingHead == "" || g.pendingHead == g.readyHead || g.merges != 0 {
			t.Fatalf("did not wait on unmerged new head: ready=%s pending=%s merges=%d", g.readyHead, g.pendingHead, g.merges)
		}
		g.ci.setHead(g.pr.Number, g.pendingHead, prCheck{Name: "build", Bucket: "pass"})
		return nil
	}
	var out bytes.Buffer
	if err := runMerge(io.Discard, &out, landingFlags()); err != nil {
		t.Fatalf("%v\n%s", err, &out)
	}
	if pauses != 1 || suites != 1 || g.merges != 1 || g.pr.HeadOID != g.pendingHead {
		t.Fatalf("pauses=%d suites=%d merges=%d head=%s pending=%s\n%s", pauses, suites, g.merges, g.pr.HeadOID, g.pendingHead, &out)
	}
}
