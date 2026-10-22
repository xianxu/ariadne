package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// A normal Git repair can preserve a previously stranded discussion without
// manufacturing historical archive proof. A later issue owns a separate family
// and must still be able to land after that explicit operator repair.
func TestLandingArchiveHistoricalNoncanonicalSideRepair(t *testing.T) {
	_, origin, _ := laFixture(t)
	root := laPeer(t, origin)
	const oldIssue = "workshop/issues/000768-darwin-pty-final-output.md"
	const oldPlan = "workshop/plans/000768-darwin-pty-final-output-plan.md"
	const side = "workshop/plans/000768-darwin-pty-final-output-plan.md.side/shared-owner-draft.md"
	const archivedSide = "workshop/history/plans/000768-darwin-pty-final-output-plan.md.side/shared-owner-draft.md"
	body := func(id, status, text string) string {
		return strings.Replace(laBody(status, text), "000246", id, 1)
	}
	laWrite(t, root, oldIssue, body("000768", "working", "old issue"))
	base := laCommit(t, root, "publish old issue")
	laWrite(t, root, oldIssue, body("000768", "codecomplete", "old reviewed close"))
	laWrite(t, root, oldPlan, "old plan\n")
	laWrite(t, root, side, "---\ntype: side-doc\nparent: ../000768-darwin-pty-final-output-plan.md\n---\n\nShared owner draft.\n\n")
	head := laCommit(t, root, "#768: close")
	oldPR := landingPR{Number: 768, State: "MERGED", Repo: "test/repo", HeadRef: "000768-darwin-pty-final-output", HeadOID: head, BaseRef: "main", BaseOID: base, MergeOID: head, LandedOID: head}
	originalBlob := laGit(t, root, "rev-parse", head+":"+side)

	// Reproduce the old parent-only archive through ordinary Git, rather than
	// asking the now-defensive production planner to create an invalid bundle.
	for _, dir := range []string{"workshop/history/issues", "workshop/history/plans"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	oldArchivedIssue := "workshop/history/issues/" + filepath.Base(oldIssue)
	testfix.Git(t, root, "mv", oldIssue, oldArchivedIssue)
	laWrite(t, root, oldArchivedIssue, body("000768", "done", "old reviewed close"))
	testfix.Git(t, root, "mv", oldPlan, "workshop/history/plans/"+filepath.Base(oldPlan))
	broken := laCommit(t, root, archiveCommitMessage+"\n\n"+archiveProvenance(oldPR.Repo, oldPR))
	testfix.Git(t, root, "push", "-q", "origin", "HEAD:main")
	refuseOldProof := func(tip string) {
		t.Helper()
		if complete, err := laProof(root, tip, oldPR); err == nil || complete {
			t.Fatalf("historically incomplete archive accepted at %s: complete=%v err=%v", tip, complete, err)
		}
	}
	refuseOldProof(broken)
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, archivedSide)), 0o755); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, root, "mv", side, archivedSide)
	repaired := laCommit(t, root, "operator: preserve stranded discussion in history")
	testfix.Git(t, root, "push", "-q", "origin", "HEAD:main")
	if got := laGit(t, root, "rev-parse", repaired+":"+archivedSide); got != originalBlob {
		t.Fatalf("repair changed side-document bytes: blob=%s want=%s", got, originalBlob)
	}
	if gitSucceeds(root, "cat-file", "-e", repaired+":"+side) {
		t.Fatal("repair retained the active side document")
	}
	refuseOldProof(repaired)

	const futureIssue = "workshop/issues/000769-future.md"
	laWrite(t, root, futureIssue, body("000769", "working", "future issue"))
	futureBase := laCommit(t, root, "publish independent future issue")
	laWrite(t, root, futureIssue, body("000769", "codecomplete", "future reviewed close"))
	laWrite(t, root, "workshop/plans/000769-future-plan.md", "future plan\n")
	futureHead := laCommit(t, root, "#769: close")
	futurePR := landingPR{Number: 769, State: "MERGED", Repo: "test/repo", HeadRef: "000769-future", HeadOID: futureHead, BaseRef: "main", BaseOID: futureBase, MergeOID: futureHead, LandedOID: futureHead}
	testfix.Git(t, root, "push", "-q", "origin", "HEAD:main")
	if err := laArchive(root, futurePR); err != nil {
		t.Fatalf("independent issue archive blocked by repaired old family: %v", err)
	}
	tip := laGit(t, origin, "rev-parse", "main")
	if complete, err := laProof(root, tip, futurePR); err != nil || !complete {
		t.Fatalf("independent issue archive proof: complete=%v err=%v", complete, err)
	}
	if got := laGit(t, origin, "rev-parse", "main:"+archivedSide); got != originalBlob {
		t.Fatalf("future archive changed repaired discussion: blob=%s want=%s", got, originalBlob)
	}
	refuseOldProof(tip)
}
