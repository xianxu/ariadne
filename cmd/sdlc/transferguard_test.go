package main

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/recovery"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// handedOff runs the spin-off handoff, then lands the owner's edit E on main.
func handedOff(t *testing.T) *trackerRepo {
	t.Helper()
	r := spinOffOnCodeBranch(t)
	var out, errs bytes.Buffer
	if err := runMoveDetail(context.Background(), &out, &errs, moveFlags(8)); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	ownerEdit(t, r)
	return r
}

const spinOffCard = "workshop/issue-cards/000008-spin-off.md"

// The two workspaces of the guard fixtures: this checkout, and another slot.
var (
	thisSlot  = issue.Claimant{Operator: "me", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/this", Repository: "r"}
	otherSlot = issue.Claimant{Operator: "peer", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/other", Repository: "r"}
)

// ownSpinOff makes this checkout thisSlot and records owner as #8's claimant.
func ownSpinOff(t *testing.T, r *trackerRepo, owner issue.Claimant) {
	t.Helper()
	withClaimant(t, thisSlot)
	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	if err := env.repo.ChangeCard("000008", spinOffCard, "fixture owner", operationToken("set"), func(c []byte) ([]byte, error) {
		return issue.SetCardClaimant(c, owner)
	}); err != nil {
		t.Fatal(err)
	}
}

// The guard's two refusals. A foreign one names the owner (otherSlot's
// checkout); the table also asserts it never offers a claim the owner's hold
// would refuse, and offers restore except over this workspace's own record.
const (
	refuseForeign = "is owned by peer on box at /w/other"
	refuseBehind  = "Merge main"
)

// The card's claimant in a guard fixture (#274: unclaimed is editable).
type holder int

const (
	heldByOther holder = iota
	heldByThis
	unclaimed
)

// holdSpinOff makes this checkout thisSlot and gives #8's card holder h.
func holdSpinOff(t *testing.T, r *trackerRepo, h holder) {
	t.Helper()
	switch h {
	case heldByOther:
		ownSpinOff(t, r, otherSlot)
	case heldByThis:
		ownSpinOff(t, r, thisSlot)
	default:
		withClaimant(t, thisSlot)
	}
}

func TestTransferGuardOverBranchShapes(t *testing.T) {
	for _, c := range []struct {
		name   string
		owner  holder
		shape  func(t *testing.T, r *trackerRepo)
		refuse string
	}{
		{"source branch after net-zero handoff", heldByOther, func(*testing.T, *trackerRepo) {}, ""},
		{"source merged main then left details alone", heldByOther, func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("merge", "-q", "--no-edit", "origin/main")
		}, ""},
		{"non-owner source re-adds its stale copy", heldByOther, reAddStale, refuseForeign},
		{"owner's checkout re-adds a stale copy", heldByThis, reAddStale, refuseBehind},
		{"non-owner deletes details after merging main", heldByOther, func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("merge", "-q", "--no-edit", "origin/main")
			r.git("rm", "-q", spinOffDetails)
			r.git("commit", "-qm", "drop details")
		}, refuseForeign},
		{"non-owner rewrites details after merging main", heldByOther, func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("merge", "-q", "--no-edit", "origin/main")
			writeRepoFile(t, r.root, spinOffDetails, "rewritten\n")
			r.git("commit", "-qam", "rewrite")
		}, refuseForeign},
		{"owner's edit on the issue branch", heldByThis, func(t *testing.T, r *trackerRepo) {
			issueBranchEdit(t, r, "000008-spin-off")
		}, ""},
		// pair#365: the issue's branch name was retired; the owner lands anyway.
		{"owner's edit on a renamed branch", heldByThis, func(t *testing.T, r *trackerRepo) {
			issueBranchEdit(t, r, "000008-spin-off-close")
		}, ""},
		{"owner's local main merges the issue branch for a direct push", heldByThis, func(t *testing.T, r *trackerRepo) {
			issueBranchEdit(t, r, "000008-spin-off")
			r.git("switch", "-q", "-C", "main", "origin/main")
			r.git("merge", "-q", "--no-ff", "--no-edit", "000008-spin-off")
		}, ""},
		{"owner's local main edits the details after that merge", heldByThis, func(t *testing.T, r *trackerRepo) {
			issueBranchEdit(t, r, "000008-spin-off")
			r.git("switch", "-q", "-C", "main", "origin/main")
			r.git("merge", "-q", "--no-ff", "--no-edit", "000008-spin-off")
			writeRepoFile(t, r.root, spinOffDetails, "the owner again\n")
			r.git("commit", "-qam", "edit on main")
		}, ""},
		{"non-owner's local main merges its rewrite", heldByOther, func(t *testing.T, r *trackerRepo) {
			issueBranchEdit(t, r, "000013-meddler")
			r.git("switch", "-q", "-C", "main", "origin/main")
			r.git("merge", "-q", "--no-ff", "--no-edit", "000013-meddler")
		}, refuseForeign},
		// pair#300–#303: the owner's work committed on another issue's branch.
		{"owner's edit on another issue's branch", heldByThis, func(t *testing.T, r *trackerRepo) {
			issueBranchEdit(t, r, "000013-meddler")
		}, ""},
		// #274: an unclaimed issue has no owner to protect, but its edits still
		// land only against main's latest version.
		{"unclaimed: rewrite after merging main", unclaimed, func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("merge", "-q", "--no-edit", "origin/main")
			writeRepoFile(t, r.root, spinOffDetails, "a spec correction\n")
			r.git("commit", "-qam", "correct the spec")
		}, ""},
		{"unclaimed: local main edits the details for a direct push", unclaimed, func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("switch", "-q", "-C", "main", "origin/main")
			writeRepoFile(t, r.root, spinOffDetails, "a log line on main\n")
			r.git("commit", "-qam", "edit on main")
		}, ""},
		{"unclaimed: re-adds a stale copy", unclaimed, reAddStale, refuseBehind},
		{"unrelated branch", heldByOther, func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("switch", "-q", "-c", "000011-other", "origin/main")
			writeRepoFile(t, r.root, "other.go", "package other\n")
			r.git("add", "other.go")
			r.git("commit", "-qm", "other work")
		}, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := handedOff(t)
			holdSpinOff(t, r, c.owner)
			c.shape(t, r)
			err := guardTransferredDetails(context.Background())
			if c.refuse == "" {
				if err != nil {
					t.Fatalf("refused a safe landing: %v", err)
				}
				return
			}
			if !errors.Is(err, errTransferredDetails) || !strings.Contains(err.Error(), c.refuse) {
				t.Fatalf("want refusal %q, got %v", c.refuse, err)
			}
			if strings.Contains(err.Error(), "sdlc claim") {
				t.Fatalf("refusal recommends a claim: %v", err)
			}
			if offers := strings.Contains(err.Error(), "sdlc issue restore"); offers == (c.owner == heldByThis) {
				t.Fatalf("restore offered = %v for holder %v: %v", offers, c.owner, err)
			}
		})
	}
}

// reAddStale keeps the filing branch's pre-handoff copy of #8's details.
func reAddStale(t *testing.T, r *trackerRepo) {
	stale := r.git("show", "HEAD~1:"+spinOffDetails)
	writeRepoFile(t, r.root, spinOffDetails, stale+"\n")
	r.git("add", spinOffDetails)
	r.git("commit", "-qm", "re-add")
}

// An owner whose branch predates main's latest version of the details (another
// slot published edit E) must merge main first, whether or not git could merge
// the two edits cleanly; after the merge it lands.
func TestTransferGuardOwnerBehindMainLatest(t *testing.T) {
	for _, c := range []struct {
		name     string
		edit     func(before string) string
		conflict bool
	}{
		{"edits merge cleanly", func(before string) string { return before + "\nThe owner's closing note.\n" }, false},
		{"edits conflict", func(string) string { return "the owner's design\n" }, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := handedOff(t)
			ownSpinOff(t, r, thisSlot)
			r.git("fetch", "-q", "origin")
			r.git("switch", "-q", "-c", "000008-spin-off", "origin/main~1")
			writeRepoFile(t, r.root, spinOffDetails, c.edit(r.git("show", "HEAD:"+spinOffDetails)))
			r.git("commit", "-qam", "owner edits")
			err := guardTransferredDetails(context.Background())
			if !errors.Is(err, errTransferredDetails) || !strings.Contains(err.Error(), refuseBehind) {
				t.Fatalf("owner behind main's latest version: %v", err)
			}
			out, mergeErr := exec.Command("git", "-C", r.root, "merge", "-q", "--no-edit", "origin/main").CombinedOutput()
			if (mergeErr != nil) != c.conflict {
				t.Fatalf("fixture: merge conflict = %v, want %v: %s", mergeErr != nil, c.conflict, out)
			}
			if c.conflict {
				writeRepoFile(t, r.root, spinOffDetails, "the owner's design, after E\n")
				r.git("commit", "-qam", "merge main")
			}
			if err := guardTransferredDetails(context.Background()); err != nil {
				t.Fatalf("owner after merging main: %v", err)
			}
		})
	}
}

func TestTransferGuardInterruptedRemovalProtectsOwnerEdits(t *testing.T) {
	r := spinOffOnCodeBranch(t)
	// Interrupt before the source removal: the branch still carries B.
	hooks, state := t.TempDir(), t.TempDir()
	hook := "#!/bin/sh\nwhile read l ls rr rs; do\n  case \"$rr\" in refs/heads/issue-tracker)\n" +
		"    n=$(cat '" + state + "/n' 2>/dev/null || echo 0); n=$((n+1)); echo $n > '" + state + "/n'\n" +
		"    if [ $n -ge 2 ]; then exit 1; fi;;\n  esac\ndone\nexit 0\n"
	if err := os.WriteFile(filepath.Join(hooks, "pre-push"), []byte(hook), 0o755); err != nil {
		t.Fatal(err)
	}
	r.git("config", "core.hooksPath", hooks)
	var out, errs bytes.Buffer
	_ = runMoveDetail(context.Background(), &out, &errs, moveFlags(8))
	r.git("config", "--unset", "core.hooksPath")
	// D is on main but the card does not name it yet, and the source still has B.
	// The owner edits main meanwhile; reconcile must still converge to net-zero.
	ownerEdit(t, r)
	var out2, errs2 bytes.Buffer
	if err := runRecoveryReconcile(context.Background(), &out2, &errs2, 8); err != nil {
		t.Fatalf("reconcile: %v\n%s", err, errs2.String())
	}
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("after reconcile the source branch is net-zero: %v", err)
	}
}

func TestTransferGuardWithoutTrackerIsNotApplicable(t *testing.T) {
	legacyRepo(t)
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("legacy repository: %v", err)
	}
	// A migrated checkout whose tracker vanished is not "no tracker": the
	// guard refuses rather than skip the protection (#252 cutover guard).
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	r.git("push", "-q", "origin", "--delete", "issue-tracker")
	if err := guardTransferredDetails(context.Background()); !errors.Is(err, tracker.ErrCutover) {
		t.Fatalf("marker without tracker: %v", err)
	}
}

// The publish gate (push, merge) and PR creation run the guard for real.
func TestPublishGateAndPRRefuseChangedHandedOffDetails(t *testing.T) {
	r := handedOff(t)
	ownSpinOff(t, r, otherSlot)
	r.git("fetch", "-q", "origin")
	r.git("merge", "-q", "--no-edit", "origin/main")
	r.git("rm", "-q", spinOffDetails)
	r.git("commit", "-qm", "drop details")
	var stderr bytes.Buffer
	if err := runPublishGate(context.Background(), r.git("merge-base", "HEAD", "origin/main"), "workshop/issues", &stderr); !errors.Is(err, errTransferredDetails) {
		t.Fatalf("publish gate: %v", err)
	}
	var stdout bytes.Buffer
	if err := runPR(&stdout, &stderr, &prFlags{IssuesDir: "workshop/issues"}); !errors.Is(err, errTransferredDetails) {
		t.Fatalf("pr: %v", err)
	}
}

// A fresh clone has none of the originating checkout's recovery refs; the
// tracker's card alone must still protect another workspace's details.
func TestTransferGuardFromAFreshCloneUsesOnlyTrackerRecords(t *testing.T) {
	r := handedOff(t)
	ownSpinOff(t, r, otherSlot)
	r.git("push", "-q", "origin", "000007-seven")
	clone := filepath.Join(t.TempDir(), "fresh")
	git(t, "", "clone", "-q", r.origin, clone)
	git(t, clone, "config", "user.name", "c")
	git(t, clone, "config", "user.email", "c@c")
	git(t, clone, "switch", "-q", "000007-seven")
	chdirTo(t, clone)
	if refs := git(t, clone, "for-each-ref", "refs/sdlc/"); strings.TrimSpace(refs) != "" {
		t.Fatalf("fixture: clone has private refs %s", refs)
	}
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("net-zero source branch refused from a fresh clone: %v", err)
	}
	git(t, clone, "merge", "-q", "--no-edit", "origin/main")
	writeRepoFile(t, clone, spinOffDetails, "rewritten in a fresh clone\n")
	git(t, clone, "commit", "-qam", "rewrite")
	if err := guardTransferredDetails(context.Background()); !errors.Is(err, errTransferredDetails) {
		t.Fatalf("fresh clone rewrite not refused: %v", err)
	}
}

// BR-6: between transfer.main and transfer.record the card names no main
// commit, but main already holds the details and may already carry the new
// owner's edits. A landing of the unfinished source branch must be refused.
func TestTransferGuardProtectsAnUnrecordedPublication(t *testing.T) {
	r := spinOffOnCodeBranch(t)
	block := blockTrackerAfterFirstPush(t, r)
	var out, errs bytes.Buffer
	if err := runMoveDetail(context.Background(), &out, &errs, moveFlags(8)); err == nil {
		t.Fatal("interruption fixture did not interrupt")
	}
	_ = os.Remove(block)
	h, ok, _ := issue.CardHandoff([]byte(r.card("workshop/issue-cards/000008-spin-off.md")))
	if !ok || h.MainCommit != "" {
		t.Fatalf("fixture: want an unrecorded handoff, got %+v", h)
	}
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("identical unfinished copy loses nothing: %v", err)
	}
	ownerEdit(t, r)
	if err := guardTransferredDetails(context.Background()); !errors.Is(err, errTransferredDetails) {
		t.Fatalf("unrecorded publication left unprotected: %v", err)
	}
}

// issueBranchEdit is an edit to #8's details on a fresh branch from main.
func issueBranchEdit(t *testing.T, r *trackerRepo, branch string) {
	t.Helper()
	r.git("fetch", "-q", "origin")
	r.git("switch", "-q", "-c", branch, "origin/main")
	writeRepoFile(t, r.root, spinOffDetails, "an edit on "+branch+"\n")
	r.git("commit", "-qam", "edit on "+branch)
}

// #285: the owner is the card's claimant, never a branch. No command help and
// neither guard refusal may say otherwise.
func TestNoHelpCallsABranchTheOwner(t *testing.T) {
	branchOwner := regexp.MustCompile(`(?i)owner'?s? (issue )?branch|owner branch|branch (is|as) the owner`)
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			walk(sub)
		}
		if m := branchOwner.FindString(c.Short + "\n" + c.Long); m != "" {
			t.Errorf("`%s` help calls a branch the owner: %q", c.CommandPath(), m)
		}
	}
	walk(buildRoot())
	env := &trackerEnv{target: gitx.PublicationTarget{Remote: "origin"}}
	for _, v := range []verdict{verdictForeign, verdictBehind} {
		if m := branchOwner.FindString(detailsRefusal(env, changedDetail{ID: "000008", Path: spinOffDetails, Verdict: v}).Error()); m != "" {
			t.Errorf("refusal %v calls a branch the owner: %q", v, m)
		}
	}
}

// #274: published details are editable by mayNotEditDetails' policy, stated
// once ({{DETAILS_EDITORS}}). No help, recovery contract, map, README or sdlc
// source comment may say only the owner publishes or that restore is a
// non-owner's offer.
func TestNoTextSaysOnlyTheOwnerEditsDetails(t *testing.T) {
	ownerOnly := regexp.MustCompile(`(?i)only the owner (re)?publish|only its owner publish|(the|an) owner's (later )?edits|owned issues' (edited )?details|offer to a non-owner|publishes owned|the owner only|owner publishes later edits|this workspace owns each card|only from the owner's checkout|owned issue's publication`)
	check := func(where, text string) {
		if m := ownerOnly.FindString(text); m != "" {
			t.Errorf("%s states the pre-#274 owner-only policy: %q", where, m)
		}
	}
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			walk(sub)
		}
		check("`"+c.CommandPath()+"` help", c.Short+"\n"+c.Long)
	}
	walk(buildRoot())
	check("recovery table", recovery.Table())
	// The policy is stated, not merely unrefuted: every help that says who
	// publishes later edits names the unclaimed case.
	for _, args := range [][]string{{"issue"}, {"issue", "publish"}, {"issue", "restore"}} { // move-detail: a hidden alias since #308
		c, _, err := buildRoot().Find(args)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(c.Short+c.Long, "unclaimed") {
			t.Errorf("`%s` help does not state the unclaimed case (#274)", c.CommandPath())
		}
	}
	root := repoRootForTest(t)
	docs, err := filepath.Glob(filepath.Join(root, "atlas", "workflow", "*.md"))
	if err != nil || len(docs) == 0 {
		t.Fatalf("no atlas workflow docs under %s: %v", root, err)
	}
	if err := filepath.WalkDir(filepath.Join(root, "cmd", "sdlc"), func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && !strings.HasSuffix(p, "_test.go") && (strings.HasSuffix(p, ".go") || strings.HasSuffix(p, ".md")) {
			docs = append(docs, p)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	for _, p := range append(docs, filepath.Join(root, "AGENTS.base.md"), filepath.Join(root, "README.md")) {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		check(p, string(raw))
	}
}

// #307: the guard judges the owner by slot. The owning slot lands from a
// checkout other than the one the card names; another slot is refused even
// with the card's own path.
func TestTransferGuardFollowsTheSlot(t *testing.T) {
	here := thisSlot
	here.Workspace = "pair:1"
	for _, c := range []struct {
		name   string
		owner  issue.Claimant
		refuse string
	}{
		{"owning slot, another checkout", func() issue.Claimant { o := here; o.Worktree = "/w/pair-slot1/ariadne"; return o }(), ""},
		{"another slot, this path", func() issue.Claimant { o := here; o.Workspace = "ariadne:1"; return o }(), "is owned by me on box at ariadne:1"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := handedOff(t)
			ownSpinOff(t, r, c.owner)
			withClaimant(t, here)
			issueBranchEdit(t, r, "000008-spin-off")
			err := guardTransferredDetails(context.Background())
			if c.refuse == "" {
				if err != nil {
					t.Fatalf("refused the owning slot: %v", err)
				}
				return
			}
			if !errors.Is(err, errTransferredDetails) || !strings.Contains(err.Error(), c.refuse) {
				t.Fatalf("want refusal %q, got %v", c.refuse, err)
			}
		})
	}
}

// #308: issue new publishes details at creation, so no text may still say it
// writes them locally, call publishing them a first-publication step or an
// escape hatch, or offer `move-detail` (a hidden alias now) in visible help.
func TestNoTextSaysIssueNewWritesDetailsLocally(t *testing.T) {
	stale := regexp.MustCompile(`(?i)write[s]? the details locally|details stay local|first-publication case|escape hatch, consult the operator|\(first\s+publication\)|a first publication \(no owner`)
	check := func(where, text string) {
		t.Helper()
		if m := stale.FindString(text); m != "" {
			t.Errorf("%s states the pre-#308 local filing: %q", where, m)
		}
	}
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			walk(sub)
		}
		help := c.Short + "\n" + c.Long
		check("`"+c.CommandPath()+"` help", help)
		// `sdlc recovery` lists every verb's contract, hidden aliases included.
		if !c.Hidden && c.Name() != "recovery" && strings.Contains(help, "move-detail") && !strings.Contains(help, "deprecated alias") {
			t.Errorf("`%s` help offers the retired `move-detail`", c.CommandPath())
		}
	}
	walk(buildRoot())
	root := repoRootForTest(t)
	docs, _ := filepath.Glob(filepath.Join(root, "atlas", "workflow", "*.md"))
	docs = append(docs, filepath.Join(root, "AGENTS.base.md"), filepath.Join(root, "AGENTS.md"), filepath.Join(root, "README.md"))
	if err := filepath.WalkDir(filepath.Join(root, "cmd", "sdlc", "helptext"), func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".md") {
			docs = append(docs, p)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	for _, p := range docs {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		check(p, string(raw))
	}
}
