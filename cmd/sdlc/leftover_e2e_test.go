package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// Branches left behind at a cutover (#252): they pass the dry run (or it never
// saw them), are locked for sdlc until caught up, then work normally. These
// tests walk each shape from lock to landing, through the real publish gate.

const leftoverIssue = "workshop/issues/000001-one.md"

// cutOver runs the dry run and the apply in dir, as the operator would.
func cutOver(t *testing.T, dir string) {
	t.Helper()
	out := mustSlotRun(t, dir, "issue", "migrate")
	m := digestLine.FindStringSubmatch(out)
	if m == nil {
		t.Fatalf("no digest:\n%s", out)
	}
	mustSlotRun(t, dir, "issue", "migrate", "--apply", "--expect", m[1])
}

// designOn adds a complete design above the Log, keeping everything else
// (frontmatter, mirror, title, and whatever the branch wrote before).
func designOn(t *testing.T, root string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, leftoverIssue))
	if err != nil {
		t.Fatal(err)
	}
	design := "## Spec\n\nA thing.\n\n## Done when\n\n- it works\n\n## Plan\n\n- [x] do it\n\n## Log\n"
	if !strings.Contains(string(raw), "## Log\n") {
		t.Fatalf("no Log section:\n%s", raw)
	}
	writeRepoFile(t, root, leftoverIssue, strings.Replace(string(raw), "## Log\n", design, 1))
}

// lockedVerbs asserts every sdlc reader and writer refuses on a checkout from
// before the cutover, each naming the way across.
func lockedVerbs(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"issue", "list"},
		{"issue", "show", "1"},
		{"claim", "--issue", "1"},
		{"start-plan", "--issue", "1"},
		{"change-code", "--issue", "1", "--worktree=no", "--no-judge", "--no-estimate", "--no-estimate-recon"},
		{"close", "--issue", "1", "--verified", "x", "--actual", "1", "--no-atlas"},
	} {
		out, err := slotRun(t, dir, args...)
		if err == nil || !strings.Contains(out+err.Error(), "reconcile") {
			t.Errorf("sdlc %s ran on a pre-cutover checkout (err=%v):\n%s", strings.Join(args, " "), err, out)
		}
	}
	t.Chdir(dir)
	if err := runPublishGate(t.Context(), "origin/main", "workshop/issues", os.Stderr); !errors.Is(err, tracker.ErrCutover) {
		t.Errorf("the publish gate (pr/merge/push) ran on a pre-cutover checkout: %v", err)
	}
}

// landAndPush merges branch into main in dir and publishes through `sdlc push`
// with the real publish gate.
func landAndPush(t *testing.T, dir, branch string) {
	t.Helper()
	testfix.Git(t, dir, "switch", "-q", "main")
	testfix.Git(t, dir, "pull", "-q", "--ff-only")
	if out, err := exec.Command("git", "-C", dir, "merge", "-q", "--no-ff", "--no-edit", branch).CombinedOutput(); err != nil {
		t.Fatalf("landing %s conflicts with main: %v\n%s\n%s\n--- main:\n%s\n--- branch:\n%s", branch, err, out,
			testfix.Capture(t, dir, "status", "--short"), testfix.Capture(t, dir, "show", "main:"+leftoverIssue), testfix.Capture(t, dir, "show", branch+":"+leftoverIssue))
	}
	mustSlotRun(t, dir, "push", "--yes", "--no-validate")
}

func TestLeftoverBranchIsLockedUntilCaughtUpThenWorks(t *testing.T) {
	for _, catchUp := range []string{"reconcile", "merge main"} {
		t.Run(catchUp, func(t *testing.T) {
			r := legacyRepo(t)
			r.git("switch", "-q", "-c", "000001-one")
			writeRepoFile(t, r.root, leftoverIssue, r.git("show", "HEAD:"+leftoverIssue)+"\n- design begun before the cutover\n")
			r.git("commit", "-qam", "#1: early design")
			r.git("push", "-q", "origin", "000001-one")
			r.git("switch", "-q", "main")
			cutOver(t, r.root)
			r.git("switch", "-q", "000001-one")

			lockedVerbs(t, r.root)

			switch catchUp {
			case "reconcile":
				mustSlotRun(t, r.root, "issue", "migrate", "--reconcile")
			default:
				r.git("fetch", "-q", "origin")
				r.git("merge", "-q", "--no-edit", "origin/main")
			}
			if got := mustSlotRun(t, r.root, "issue", "show", "1"); !strings.Contains(got, "status: open") {
				t.Fatalf("caught-up branch cannot read its card:\n%s", got)
			}
			// The rest of the lifecycle, from the branch that was left behind.
			mustSlotRun(t, r.root, "claim", "--issue", "1")
			mustSlotRun(t, r.root, "start-plan", "--issue", "1")
			designOn(t, r.root)
			mustSlotRun(t, r.root, "issue", "sync", "--issue", "1")
			mustSlotRun(t, r.root, "change-code", "--issue", "1", "--worktree=no", "--no-judge", "--no-estimate", "--no-estimate-recon")
			writeRepoFile(t, r.root, "cmd/one.go", "package one\n")
			r.git("add", "cmd/one.go")
			r.git("commit", "-qm", "#1: implement")
			stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
			mustSlotRun(t, r.root, "close", "--issue", "1", "--verified", "leftover", "--actual", "1", "--no-atlas")
			if !strings.Contains(r.git("show", "HEAD:"+leftoverIssue), "design begun before the cutover") {
				t.Fatal("the pre-cutover design was lost")
			}
			landAndPush(t, r.root, "000001-one")
			if c := r.card(tracker.CardPath("000001", "one")); !strings.Contains(c, "status: done") {
				t.Fatalf("the leftover branch did not land to done:\n%s", c)
			}
		})
	}
}

// A branch in another clone edited card fields the dry run never saw. After the
// cutover reconcile refuses naming the fields; the remedy — apply the change on
// the card from a caught-up checkout, revert it on the branch — then works.
func TestInvisibleBranchWithCardEditsRecovers(t *testing.T) {
	r := legacyRepo(t)
	other := filepath.Join(t.TempDir(), "other")
	testfix.Git(t, "", "clone", "-q", r.origin, other)
	testfix.Git(t, other, "config", "user.name", "o")
	testfix.Git(t, other, "config", "user.email", "o@o")
	testfix.Git(t, other, "switch", "-q", "-c", "000001-one")
	base := strings.TrimRight(testfix.Capture(t, other, "show", "HEAD:"+leftoverIssue), "\n")
	claimed := strings.Replace(base, "status: open", "status: working", 1) + "\n- claimed on the branch, never synced\n"
	writeRepoFile(t, other, leftoverIssue, claimed)
	testfix.Git(t, other, "commit", "-qam", "#1: legacy claim on the branch")

	cutOver(t, r.root) // the dry run cannot see the other clone's branch

	testfix.Git(t, other, "fetch", "-q", "origin")
	out, err := slotRun(t, other, "issue", "migrate", "--reconcile")
	if err == nil || !strings.Contains(out+err.Error(), "status") || !strings.Contains(out+err.Error(), "from a caught-up checkout") {
		t.Fatalf("reconcile did not name the edited field and where to apply it (err=%v):\n%s", err, out)
	}
	// Remedy: claim on the card from a caught-up checkout (here, main)...
	testfix.Git(t, other, "switch", "-q", "main")
	testfix.Git(t, other, "pull", "-q", "--ff-only")
	mustSlotRun(t, other, "claim", "--issue", "1")
	// ...then revert the field on the branch and reconcile.
	testfix.Git(t, other, "switch", "-q", "000001-one")
	writeRepoFile(t, other, leftoverIssue, strings.Replace(claimed, "status: working", "status: open", 1))
	testfix.Git(t, other, "commit", "-qam", "#1: card fields now live on the card")
	mustSlotRun(t, other, "issue", "migrate", "--reconcile")
	if got := mustSlotRun(t, other, "issue", "show", "1"); !strings.Contains(got, "status: working") {
		t.Fatalf("after the remedy the card is not working:\n%s", got)
	}
	raw, err := os.ReadFile(filepath.Join(other, leftoverIssue))
	if err != nil || !issue.HasMirror(raw) || !strings.Contains(string(raw), "claimed on the branch, never synced") {
		t.Fatalf("the branch's details were not brought across intact: %v\n%s", err, raw)
	}
	if out, err := exec.Command("git", "-C", other, "merge", "--no-edit", "-q", "origin/main").CombinedOutput(); err != nil {
		t.Fatalf("merge main after the remedy: %v\n%s", err, out)
	}
}

// An issue that exists only on a branch the dry run never saw has no card:
// reconcile refuses with the way forward, and CI's id check refuses to land it.
func TestInvisibleBranchOnlyIssueCannotLand(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "sdlc")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build sdlc: %v\n%s", err, out)
	}
	r := legacyRepo(t)
	other := filepath.Join(t.TempDir(), "other")
	testfix.Git(t, "", "clone", "-q", r.origin, other)
	testfix.Git(t, other, "config", "user.name", "o")
	testfix.Git(t, other, "config", "user.email", "o@o")
	testfix.Git(t, other, "switch", "-q", "-c", "000009-private")
	writeRepoFile(t, other, "workshop/issues/000009-private.md", "---\nid: 000009\nstatus: open\ncreated: 2026-09-01\n---\n\n# Private\n\n## Problem\n\nFiled on a branch only.\n")
	testfix.Git(t, other, "add", "-A")
	testfix.Git(t, other, "commit", "-qm", "#9: filed on a branch")

	cutOver(t, r.root)

	testfix.Git(t, other, "fetch", "-q", "origin")
	out, err := slotRun(t, other, "issue", "migrate", "--reconcile")
	if err == nil || !strings.Contains(out+err.Error(), "sdlc issue new") {
		t.Fatalf("reconcile accepted a cardless issue (err=%v):\n%s", err, out)
	}
	// Merging main (as a GitHub-UI merge would) brings the marker but not a card.
	testfix.Git(t, other, "merge", "-q", "--no-edit", "origin/main")
	lint := exec.Command(bin, "issue", "lint-ids", "--base", "origin/main", "--head", "HEAD")
	lint.Dir = other
	lint.Env = envWithTMPDIR(t)
	lout, lerr := lint.CombinedOutput()
	var ee *exec.ExitError
	if !errors.As(lerr, &ee) || ee.ExitCode() != 1 || !strings.Contains(string(lout), "000009-private.md") {
		t.Fatalf("CI's id check would land a cardless issue (%v):\n%s", lerr, lout)
	}
}

// Backing out a cutover before any card write: delete the tracker, revert the
// migration commit, and prune the fetched tracker in every clone — then the
// legacy workflow is whole again. Without the prune a clone stays locked.
func TestCutoverBackOutRestoresLegacy(t *testing.T) {
	r := legacyRepo(t)
	other := filepath.Join(t.TempDir(), "other")
	cutOver(t, r.root)
	testfix.Git(t, "", "clone", "-q", r.origin, other) // a clone that saw the cutover
	testfix.Git(t, other, "config", "user.name", "o")
	testfix.Git(t, other, "config", "user.email", "o@o")

	r.git("pull", "-q", "--ff-only")
	migration := r.git("log", "-1", "--format=%H", "--grep=^migrate: issue details onto the issue tracker")
	r.git("push", "-q", "origin", "--delete", "issue-tracker")
	r.git("revert", "--no-edit", migration)
	r.git("push", "-q", "origin", "main")
	r.git("fetch", "-q", "--prune", "origin")
	if got := mustSlotRun(t, r.root, "issue", "list"); !strings.Contains(got, "000001") {
		t.Fatalf("legacy listing after the back-out:\n%s", got)
	}
	if _, err := os.Stat(filepath.Join(r.root, tracker.CutoverMarkerPath)); !os.IsNotExist(err) {
		t.Fatal("the marker survived the revert")
	}
	if raw, _ := os.ReadFile(filepath.Join(r.root, leftoverIssue)); issue.HasMirror(raw) {
		t.Fatal("the details kept their mirror after the revert")
	}

	testfix.Git(t, other, "pull", "-q", "--ff-only") // pulled, but not pruned
	if cut, err := tracker.CutOver(other); err != nil || !cut {
		t.Fatalf("an unpruned clone should still read as cut over (cut=%v err=%v)", cut, err)
	}
	testfix.Git(t, other, "fetch", "-q", "--prune", "origin")
	if cut, err := tracker.CutOver(other); err != nil || cut {
		t.Fatalf("a pruned clone still reads as cut over (cut=%v err=%v)", cut, err)
	}
	if got := mustSlotRun(t, other, "issue", "list"); !strings.Contains(got, "000001") {
		t.Fatalf("legacy listing in the other clone:\n%s", got)
	}
}
