package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

const spinOffDetails = "workshop/issues/000008-spin-off.md"

func moveFlags(id int) *moveDetailFlags {
	return &moveDetailFlags{Issue: id, IssuesDir: "workshop/issues"}
}

// spinOffOnCodeBranch builds the handoff scenario: a code branch with unshipped
// work A, where `issue new` filed #8 (commit B carries only its details).
func spinOffOnCodeBranch(t *testing.T) *trackerRepo {
	t.Helper()
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	r.git("switch", "-q", "-c", "000007-seven")
	writeRepoFile(t, r.root, "code.go", "package code\n")
	r.git("add", "code.go")
	r.git("commit", "-qm", "A: unshipped code")
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, newIssueFlags(), []string{"Spin Off"}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	return r
}

// ownerEdit is the new owner's later edit E, landed on main from another clone.
func ownerEdit(t *testing.T, r *trackerRepo) string {
	t.Helper()
	peer := t.TempDir()
	testfix.Git(t, "", "clone", "-q", r.origin, peer)
	testfix.Git(t, peer, "config", "user.name", "owner")
	testfix.Git(t, peer, "config", "user.email", "o@o")
	raw, err := os.ReadFile(filepath.Join(peer, spinOffDetails))
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(raw), "## Spec\n", "## Spec\n\nThe new owner's design (E).\n", 1)
	writeRepoFile(t, peer, spinOffDetails, edited)
	testfix.Git(t, peer, "commit", "-qam", "E: owner edits details")
	testfix.Git(t, peer, "push", "-q", "origin", "main")
	return edited
}

func TestMoveDetailFromCodeBranchIsNetZeroForTheEventualMerge(t *testing.T) {
	for _, landing := range []string{"direct-merge", "merge-from-main-first", "squash"} {
		t.Run(landing, func(t *testing.T) {
			r := spinOffOnCodeBranch(t)
			published := r.git("show", "HEAD:"+spinOffDetails)
			var out, errs bytes.Buffer
			if err := runMoveDetail(context.Background(), &out, &errs, moveFlags(8)); err != nil {
				t.Fatalf("%v\n%s", err, errs.String())
			}
			// Main got only the details; the code stays unshipped.
			if got := strings.TrimSpace(testfix.Capture(t, r.origin, "show", "main:"+spinOffDetails)); got != published {
				t.Fatal("main does not carry the published details")
			}
			if _, err := os.Stat(filepath.Join(r.origin, "code.go")); err == nil || strings.Contains(testfix.Capture(t, r.origin, "ls-tree", "-r", "--name-only", "main"), "code.go") {
				t.Fatal("handoff shipped unrelated code")
			}
			// R: a narrow removal commit on the branch, carrying the operation trailer.
			if files := r.git("show", "--name-only", "--format=", "HEAD"); files != spinOffDetails {
				t.Fatalf("removal commit touched %q", files)
			}
			if !strings.Contains(r.git("log", "-1", "--format=%B"), "Detail-Handoff: #000008") {
				t.Fatal("removal lacks provenance trailer")
			}
			h, ok, err := issue.CardHandoff([]byte(r.card("workshop/issue-cards/000008-spin-off.md")))
			if err != nil || !ok || h.MainCommit == "" || h.SourceBranch != "refs/heads/000007-seven" {
				t.Fatalf("card handoff %+v %v %v", h, ok, err)
			}
			edited := ownerEdit(t, r)

			r.git("fetch", "-q", "origin")
			switch landing {
			case "merge-from-main-first":
				r.git("merge", "-q", "--no-edit", "origin/main")
				if got := r.git("show", "HEAD:"+spinOffDetails); got != strings.TrimSpace(edited) {
					t.Fatal("merge from main did not bring the owner's details")
				}
				fallthrough
			case "direct-merge":
				r.git("switch", "-q", "main")
				r.git("merge", "-q", "--ff-only", "origin/main")
				r.git("merge", "-q", "--no-edit", "000007-seven")
			case "squash":
				r.git("switch", "-q", "main")
				r.git("merge", "-q", "--ff-only", "origin/main")
				r.git("merge", "-q", "--squash", "000007-seven")
				r.git("commit", "-qm", "squash 000007")
			}
			if got := r.git("show", "HEAD:"+spinOffDetails); got != strings.TrimSpace(edited) {
				t.Fatalf("landing lost or reverted the owner's edit E:\n%s", got)
			}
			if r.git("show", "HEAD:code.go") != "package code" {
				t.Fatal("landing lost the code")
			}
		})
	}
}

func TestMoveDetailOnRestFastForwardsToThePublication(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, newIssueFlags(), []string{"Spin Off"}); err != nil {
		t.Fatal(err)
	}
	local, _ := os.ReadFile(filepath.Join(r.root, spinOffDetails))
	if err := runMoveDetail(context.Background(), &out, &errs, moveFlags(8)); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	if r.git("rev-parse", "HEAD") != r.originMain() {
		t.Fatal("rest did not fast-forward to the publication")
	}
	if status := r.git("status", "--porcelain"); status != "" {
		t.Fatalf("rest left dirty: %s", status)
	}
	if got, _ := os.ReadFile(filepath.Join(r.root, spinOffDetails)); !bytes.Equal(got, local) {
		t.Fatal("rest copy differs from what was published")
	}
	// Creation is complete: the issue is now claimable.
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(8)); err != nil {
		t.Fatalf("claim after handoff: %v", err)
	}
}

func TestMoveDetailWithoutLocalDetailsDerivesThemFromTheCard(t *testing.T) {
	cardPath, card, detailPath, _ := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, nil)
	head := r.git("rev-parse", "HEAD")
	var out, errs bytes.Buffer
	if err := runMoveDetail(context.Background(), &out, &errs, moveFlags(9)); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	onMain := testfix.Capture(t, r.origin, "show", "main:"+detailPath)
	for _, want := range []string{"# Seeded nine", "## Spec", issue.MirrorField} {
		if !strings.Contains(onMain, want) {
			t.Errorf("derived details lack %q", want)
		}
	}
	if r.git("rev-parse", "HEAD") != head {
		t.Fatal("card-source handoff touched the local branch")
	}
	var out2, errs2 bytes.Buffer
	if err := runMoveDetail(context.Background(), &out2, &errs2, moveFlags(9)); err == nil || !strings.Contains(err.Error(), "already handed off") {
		t.Fatalf("second handoff: %v", err)
	}
}

func TestMoveDetailRefusesExistingDestinationAndUnreadableSource(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	cardBefore := r.card(cardPath)
	var out, errs bytes.Buffer
	if err := runMoveDetail(context.Background(), &out, &errs, moveFlags(9)); err == nil || !strings.Contains(err.Error(), "already on main") {
		t.Fatalf("existing destination: %v", err)
	}
	if r.card(cardPath) != cardBefore {
		t.Fatal("refusal changed the card")
	}

	r2 := spinOffOnCodeBranch(t)
	abs := filepath.Join(r2.root, spinOffDetails)
	if err := os.Chmod(abs, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(abs, 0o644) })
	mainBefore := r2.originMain()
	if err := runMoveDetail(context.Background(), &out, &errs, moveFlags(8)); err == nil {
		t.Fatal("unreadable source treated as absent")
	}
	if r2.originMain() != mainBefore {
		t.Fatal("unreadable source published something")
	}
}

func TestMoveDetailRemoverLeavesAChangedSourceUntouched(t *testing.T) {
	r := spinOffOnCodeBranch(t)
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	head := r.git("rev-parse", "HEAD")
	writeRepoFile(t, r.root, spinOffDetails, "edited after publication\n")
	spec := tracker.ReceiptSpec{Token: "move-x", IssueID: "000008", SourcePath: spinOffDetails, SourceBlob: strings.Repeat("a", 40)}
	if err := moveDetailRemover(env)(spec, head); err == nil || !strings.Contains(err.Error(), "changed after it was published") {
		t.Fatalf("remover: %v", err)
	}
	if got, _ := os.ReadFile(filepath.Join(r.root, spinOffDetails)); string(got) != "edited after publication\n" || r.git("rev-parse", "HEAD") != head {
		t.Fatal("changed source was removed or committed")
	}
}

func TestMoveDetailInterruptedRecordIsFinishedByReconcile(t *testing.T) {
	r := spinOffOnCodeBranch(t)
	hooks, state := t.TempDir(), t.TempDir()
	// Reject tracker pushes after the first (the handoff) while the block file
	// exists: the record stage then cannot be confirmed in this run.
	hook := "#!/bin/sh\nwhile read l ls rr rs; do\n  case \"$rr\" in refs/heads/issue-tracker)\n" +
		"    n=$(cat '" + state + "/n' 2>/dev/null || echo 0); n=$((n+1)); echo $n > '" + state + "/n'\n" +
		"    if [ -f '" + state + "/block' ] && [ $n -ge 2 ]; then exit 1; fi;;\n  esac\ndone\nexit 0\n"
	if err := os.WriteFile(filepath.Join(hooks, "pre-push"), []byte(hook), 0o755); err != nil {
		t.Fatal(err)
	}
	r.git("config", "core.hooksPath", hooks)
	if err := os.WriteFile(filepath.Join(state, "block"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errs bytes.Buffer
	err := runMoveDetail(context.Background(), &out, &errs, moveFlags(8))
	if err == nil || !strings.Contains(err.Error(), "recovery reconcile --issue 8") {
		t.Fatalf("interrupted handoff: %v", err)
	}
	mainAfterPublish := r.originMain()
	if _, err := os.Stat(filepath.Join(r.root, spinOffDetails)); err != nil {
		t.Fatal("source removed before the publication was recorded")
	}
	var list bytes.Buffer
	if err := runRecoveryList(context.Background(), &list); err != nil || !strings.Contains(list.String(), "transfer.record") {
		t.Fatalf("recovery list: %q %v", list.String(), err)
	}
	if err := runMoveDetail(context.Background(), &out, &errs, moveFlags(8)); err == nil || !strings.Contains(err.Error(), "unfinished transfer") {
		t.Fatalf("second move-detail must defer to reconcile: %v", err)
	}
	if err := os.Remove(filepath.Join(state, "block")); err != nil {
		t.Fatal(err)
	}
	if err := runRecoveryReconcile(context.Background(), &out, &errs, 8); err != nil {
		t.Fatalf("reconcile: %v\n%s", err, errs.String())
	}
	if r.originMain() != mainAfterPublish {
		t.Fatal("reconcile republished the details")
	}
	if h, ok, _ := issue.CardHandoff([]byte(r.card("workshop/issue-cards/000008-spin-off.md"))); !ok || h.MainCommit != mainAfterPublish {
		t.Fatalf("record not finished: %+v", h)
	}
	if _, err := os.Stat(filepath.Join(r.root, spinOffDetails)); !os.IsNotExist(err) {
		t.Fatal("source not removed after reconcile")
	}
	list.Reset()
	if err := runRecoveryList(context.Background(), &list); err != nil || !strings.Contains(list.String(), "no unfinished") {
		t.Fatalf("receipt retained: %q %v", list.String(), err)
	}
}
