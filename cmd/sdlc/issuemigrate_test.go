package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// legacyRepo is a pre-cutover repository: full-frontmatter details on main
// (one without a Problem heading), archived history with a duplicate ID, and a
// bare origin with no issue tracker.
func legacyRepo(t *testing.T) *trackerRepo {
	t.Helper()
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	testfix.Git(t, root, "config", "core.hooksPath", t.TempDir())
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, root, "remote", "add", "origin", origin)
	files := map[string]string{
		"workshop/issues/000001-one.md":          "---\nid: 000001\nstatus: open\ncreated: 2026-09-01\n---\n\n# One\n\n## Problem\n\nFirst.\n\n## Log\n",
		"workshop/issues/000002-two.md":          "---\nid: 000002\nstatus: working\ncreated: 2026-09-01\n---\n\n# Two\n\nThe report, as a preamble.\n\n## Plan\n\n- [ ] x\n",
		"workshop/history/issues/000002-old.md":  "---\nid: 000002\nstatus: done\ncreated: 2026-05-01\n---\n\n# Old two\n",
		"workshop/history/issues/000003-gone.md": "---\nid: 000003\nstatus: done\nactual_hours: 1\n---\n\n# Gone\n\n## Problem\nx\n",
	}
	for p, body := range files {
		writeRepoFile(t, root, p, body)
	}
	testfix.Git(t, root, "add", "-A")
	testfix.Git(t, root, "commit", "-qm", "legacy issues")
	testfix.Git(t, root, "push", "-q", "-u", "origin", "main")
	return &trackerRepo{t: t, root: root, origin: origin}
}

var digestLine = regexp.MustCompile(`digest: ([0-9a-f]{16})`)

func migrateDryRun(t *testing.T) (digest, out string, err error) {
	t.Helper()
	stdout, stderr, err := executeSDLCTestCommand("issue", "migrate")
	m := digestLine.FindStringSubmatch(stdout)
	if m == nil {
		t.Fatalf("dry run printed no digest:\n%s\n%s", stdout, stderr)
	}
	return m[1], stdout + stderr, err
}

func TestIssueMigrateDryRunChangesNothingAndReportsRefusals(t *testing.T) {
	r := legacyRepo(t)
	r.git("switch", "-q", "-c", "000001-one")
	writeRepoFile(t, r.root, "workshop/issues/000001-one.md", "---\nid: 000001\nstatus: working\ncreated: 2026-09-01\n---\n\n# One\n\n## Problem\n\nFirst.\n\n## Log\n")
	r.git("commit", "-qam", "#1: claim on the branch, never synced")
	r.git("switch", "-q", "main")
	mainBefore := r.originMain()
	_, out, err := migrateDryRun(t)
	if err == nil || !strings.Contains(out, "workshop/issues/000001-one.md (on 000001-one)") || !strings.Contains(out, "unpublished card fields") {
		t.Fatalf("unsynced branch claim not refused: %v\n%s", err, out)
	}
	for _, want := range []string{"cards: 3", "inserted `## Problem`", "duplicate IDs: 1"} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q:\n%s", want, out)
		}
	}
	if r.originMain() != mainBefore || testfix.Capture(t, r.origin, "branch", "--list", "issue-tracker") != "" {
		t.Fatal("the dry run wrote to the remote")
	}
}

func TestIssueMigrateApplyCutsOverAndResumes(t *testing.T) {
	r := legacyRepo(t)
	digest, out, err := migrateDryRun(t)
	if err != nil {
		t.Fatalf("dry run: %v\n%s", err, out)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", "0000000000000000"); err == nil || !strings.Contains(err.Error()+stderr, "not the reviewed") {
		t.Fatalf("a wrong digest applied: %v", err)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest); err != nil {
		t.Fatalf("apply: %v\n%s", err, stderr)
	}
	r.git("pull", "-q")
	marker := r.git("show", "HEAD:"+tracker.CutoverMarkerPath)
	trackerRoot, err := tracker.ParseCutoverMarker([]byte(marker))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(testfix.Capture(t, r.origin, "rev-list", "--max-parents=0", "issue-tracker")); got != trackerRoot {
		t.Fatalf("marker root %s, tracker root %s", trackerRoot, got)
	}
	for _, id := range []string{"000001-one", "000002-two", "000003-gone"} {
		if r.card(tracker.CardPath(id[:6], id[7:])) == "" {
			t.Errorf("no card %s", id)
		}
	}
	two := r.git("show", "HEAD:workshop/issues/000002-two.md")
	if !issue.HasMirror([]byte(two)) || !strings.Contains(two, "## Problem\n\nThe report, as a preamble.") {
		t.Fatalf("details not converted:\n%s", two)
	}
	if r.git("show", "HEAD:workshop/history/issues/000002-old.md") != strings.TrimSpace("---\nid: 000002\nstatus: done\ncreated: 2026-05-01\n---\n\n# Old two\n") {
		t.Fatal("archived details were rewritten")
	}
	// The migrated repository works through the guarded verbs.
	if _, stderr, err := executeSDLCTestCommand("claim", "--issue", "1"); err != nil {
		t.Fatalf("claim after cutover: %v\n%s", err, stderr)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest); err != nil || !strings.Contains(stderr, "already migrated") {
		t.Fatalf("re-apply: %v\n%s", err, stderr)
	}
}

// An apply interrupted after the tracker exists resumes from it; a tracker that
// is not this plan's is never adopted.
func TestIssueMigrateResumesItsOwnTrackerOnly(t *testing.T) {
	r := legacyRepo(t)
	digest, _, _ := migrateDryRun(t)
	env, err := openMigrate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	mainView, err := env.main.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	in, err := migrationInventory(env, mainView)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := applyTrackerBootstrap(env, tracker.PlanTrackerMigration(in)); err != nil { // phase 1, then "crash"
		t.Fatal(err)
	}
	if _, stderr, err := executeSDLCTestCommand("claim", "--issue", "1"); err == nil || !strings.Contains(err.Error()+stderr, "cutover") {
		t.Fatalf("a verb ran between the phases: %v", err)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest); err != nil {
		t.Fatalf("resume: %v\n%s", err, stderr)
	}
	if _, err := tracker.ParseCutoverMarker([]byte(r.git("show", "origin/main:"+tracker.CutoverMarkerPath))); err != nil {
		r.git("fetch", "-q", "origin")
		if _, err := tracker.ParseCutoverMarker([]byte(r.git("show", "origin/main:"+tracker.CutoverMarkerPath))); err != nil {
			t.Fatalf("resumed apply did not publish the marker: %v", err)
		}
	}

	other := legacyRepo(t)
	cardPath, card, _, _ := seededIssue(t, "000009", "foreign")
	bootstrapTrackerOnly(t, other.root, map[string]string{cardPath: card})
	mainBefore := other.originMain()
	d, _, _ := migrateDryRun(t)
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", d); err == nil || !strings.Contains(err.Error()+stderr, "not this migration's") {
		t.Fatalf("a foreign tracker was adopted: %v", err)
	}
	if other.originMain() != mainBefore {
		t.Fatal("main moved despite the refusal")
	}
}

// A branch from before the cutover refuses guarded verbs until reconciled;
// reconcile keeps its detail edits, and merging main is then clean.
func TestIssueMigrateReconcileBringsABranchAcross(t *testing.T) {
	r := legacyRepo(t)
	r.git("switch", "-q", "-c", "000002-two")
	path := "workshop/issues/000002-two.md"
	body := r.git("show", "HEAD:"+path)
	writeRepoFile(t, r.root, path, body+"\n- [ ] a design step on the branch\n")
	r.git("commit", "-qam", "#2: design on the branch")
	r.git("switch", "-q", "main")
	digest, out, err := migrateDryRun(t)
	if err != nil {
		t.Fatalf("detail-only branch edits must not block: %v\n%s", err, out)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest); err != nil {
		t.Fatalf("apply: %v\n%s", err, stderr)
	}
	r.git("switch", "-q", "000002-two")
	var planErr string
	msg, died := expectDie(t, func() {
		_, stderr, err := executeSDLCTestCommand("start-plan", "--issue", "2")
		if err != nil {
			planErr = err.Error() + stderr
		}
	})
	if !strings.Contains(msg+planErr, "reconcile") {
		t.Fatalf("an unreconciled branch planned: died=%v %q %q", died, msg, planErr)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--reconcile"); err != nil {
		t.Fatalf("reconcile: %v\n%s", err, stderr)
	}
	reconciled := r.git("show", "HEAD:"+path)
	if !issue.HasMirror([]byte(reconciled)) || !strings.Contains(reconciled, "a design step on the branch") {
		t.Fatalf("reconcile lost the branch's edits or the mirror:\n%s", reconciled)
	}
	r.git("fetch", "-q", "origin")
	if out, err := exec.Command("git", "-C", r.root, "merge", "--no-edit", "-q", "origin/main").CombinedOutput(); err != nil {
		t.Fatalf("merging main after reconcile conflicts: %v\n%s", err, out)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--reconcile"); err != nil || !strings.Contains(stderr, "nothing to reconcile") {
		t.Fatalf("second reconcile: %v\n%s", err, stderr)
	}
}

// A legacy codecomplete whose close is provable imports bound to it, and the
// landing selector then owns it from the branch (so merge completes it); an
// unprovable one blocks the cutover.
func TestIssueMigrateImportsAProvableLegacyClose(t *testing.T) {
	r := legacyRepo(t)
	r.git("switch", "-q", "-c", "000001-one")
	writeRepoFile(t, r.root, "cmd/a.go", "package a\n")
	r.git("add", "cmd/a.go")
	r.git("commit", "-qm", "#1: implement")
	closed := "---\nid: 000001\nstatus: codecomplete\nactual_hours: 2\ncreated: 2026-09-01\n---\n\n# One\n\n## Problem\n\nFirst.\n\n## Log\n- closed\n"
	writeRepoFile(t, r.root, "workshop/issues/000001-one.md", closed)
	r.git("commit", "-qam", "#1: close\n\nReview-Verdict: SHIP")
	anchor := r.git("rev-parse", "HEAD")
	r.git("push", "-q", "origin", "000001-one")
	r.git("switch", "-q", "main") // the legacy sync published the close's issue file to main
	writeRepoFile(t, r.root, "workshop/issues/000001-one.md", closed)
	r.git("commit", "-qam", "#1: issue-sync")
	r.git("push", "-q", "origin", "main")
	digest, out, err := migrateDryRun(t)
	if err != nil {
		t.Fatalf("provable close refused: %v\n%s", err, out)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest); err != nil {
		t.Fatalf("apply: %v\n%s", err, stderr)
	}
	b, ok, err := issue.CardCompletion([]byte(r.card(tracker.CardPath("000001", "one"))))
	if err != nil || !ok || b.EvidenceCommit != anchor || b.Token != tracker.MigrationToken("000001") {
		t.Fatalf("binding %+v %v", b, err)
	}
	r.git("switch", "-q", "000001-one")
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--reconcile"); err != nil {
		t.Fatalf("reconcile: %v\n%s", err, stderr)
	}
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	rs, err := loadIssueRecords(context.Background(), "workshop/issues", tracker.Fresh)
	if err != nil {
		t.Fatal(err)
	}
	owned, err := ownedCompletions(env, rs, "HEAD", r.originMain(), false)
	if err != nil || len(owned) != 1 || owned[0].ID != "000001" {
		t.Fatalf("the imported close is not owned by its branch: %+v %v", owned, err)
	}
	// Main's conversion pins the blob the tracker actually holds (the bound card).
	r.git("fetch", "-q", "origin")
	pin, err := issue.MirrorBaselineOID([]byte(r.git("show", "origin/main:workshop/issues/000001-one.md") + "\n"))
	if err != nil || pin != r.git("rev-parse", "origin/issue-tracker:"+tracker.CardPath("000001", "one")) {
		t.Fatalf("main pins %s, the tracker holds %s (%v)", pin, r.git("rev-parse", "origin/issue-tracker:"+tracker.CardPath("000001", "one")), err)
	}
	// The reconciled branch merges migrated main cleanly, lands, and settles to done.
	if out, err := exec.Command("git", "-C", r.root, "merge", "--no-edit", "-q", "origin/main").CombinedOutput(); err != nil {
		t.Fatalf("the imported close's branch conflicts with migrated main: %v\n%s", err, out)
	}
	r.git("switch", "-q", "main")
	r.git("pull", "-q", "--ff-only")
	r.git("merge", "-q", "--no-ff", "--no-edit", "000001-one")
	r.git("push", "-q", "origin", "main")
	if _, stderr, err := executeSDLCTestCommand("issue", "recovery", "reconcile", "--issue", "1"); err != nil {
		t.Fatalf("settle: %v\n%s", err, stderr)
	}
	if c := r.card(tracker.CardPath("000001", "one")); !strings.Contains(c, "status: done") {
		t.Fatalf("the landed imported close did not complete:\n%s", c)
	}

	// Code after the close makes it unprovable: the cutover waits.
	u := legacyRepo(t)
	u.git("switch", "-q", "-c", "000001-one")
	writeRepoFile(t, u.root, "workshop/issues/000001-one.md", closed)
	u.git("commit", "-qam", "#1: close")
	writeRepoFile(t, u.root, "cmd/late.go", "package late\n")
	u.git("add", "cmd/late.go")
	u.git("commit", "-qm", "#1: code after the close")
	u.git("switch", "-q", "main")
	writeRepoFile(t, u.root, "workshop/issues/000001-one.md", closed)
	u.git("commit", "-qam", "#1: issue-sync")
	u.git("push", "-q", "origin", "main")
	if _, out, err := migrateDryRun(t); err == nil || !strings.Contains(out, "code after its close") {
		t.Fatalf("an unprovable close did not block: %v\n%s", err, out)
	}
}

// A legacy close that landed but was never marked done binds main's record,
// unless its branch still carries code main does not have (#252 M4).
func TestIssueMigrateLandedLegacyClose(t *testing.T) {
	r := legacyRepo(t)
	r.git("switch", "-q", "-c", "000001-one")
	writeRepoFile(t, r.root, "cmd/a.go", "package a\n")
	r.git("add", "cmd/a.go")
	r.git("commit", "-qm", "#1: implement")
	closed := "---\nid: 000001\nstatus: codecomplete\nactual_hours: 2\ncreated: 2026-09-01\n---\n\n# One\n\n## Problem\n\nFirst.\n\n## Log\n- closed\n"
	writeRepoFile(t, r.root, "workshop/issues/000001-one.md", closed)
	r.git("commit", "-qam", "#1: close")
	r.git("switch", "-q", "main")
	r.git("merge", "-q", "--no-ff", "--no-edit", "000001-one") // landed; the done flip never ran
	writeRepoFile(t, r.root, "cmd/later.go", "package later\n")
	r.git("add", "cmd/later.go")
	r.git("commit", "-qm", "other work on main")
	r.git("push", "-q", "origin", "main", "000001-one")
	if _, out, err := migrateDryRun(t); err != nil {
		t.Fatalf("a landed close refused: %v\n%s", err, out)
	}
	r.git("switch", "-q", "000001-one")
	writeRepoFile(t, r.root, "cmd/unlanded.go", "package unlanded\n")
	r.git("add", "cmd/unlanded.go")
	r.git("commit", "-qm", "#1: code after the close, never landed")
	r.git("push", "-q", "origin", "000001-one")
	r.git("switch", "-q", "main")
	if _, out, err := migrateDryRun(t); err == nil || !strings.Contains(out, "carries code after it") {
		t.Fatalf("unlanded code after a landed close did not block: %v\n%s", err, out)
	}
}

// A pre-cutover branch carries older copies of issues it never touched (main
// moved them on after it forked). Reconcile checks only what the branch
// changed, and its merge of main's migration commit brings main's versions of
// the rest — including #5's archive move — so the later merge is clean (#252,
// found by smoke test A2).
func TestIssueMigrateReconcileLeavesUntouchedIssuesToMain(t *testing.T) {
	r := legacyRepo(t)
	five := "workshop/issues/000005-five.md"
	writeRepoFile(t, r.root, five, "---\nid: 000005\nstatus: open\ncreated: 2026-09-01\n---\n\n# Five\n\n## Problem\n\nx\n")
	r.git("add", five)
	r.git("commit", "-qm", "#5: file")
	r.git("push", "-q", "origin", "main")
	r.git("switch", "-q", "-c", "000002-two")
	two := "workshop/issues/000002-two.md"
	writeRepoFile(t, r.root, two, r.git("show", "HEAD:"+two)+"\n- a design step on the branch\n")
	r.git("commit", "-qam", "#2: design")
	r.git("push", "-q", "origin", "000002-two")
	r.git("switch", "-q", "main") // main moves on: #1 blocked, #5 done and archived
	one := "workshop/issues/000001-one.md"
	writeRepoFile(t, r.root, one, strings.Replace(r.git("show", "HEAD:"+one)+"\n", "status: open", "status: blocked", 1))
	r.git("mv", five, "workshop/history/issues/000005-five.md")
	writeRepoFile(t, r.root, "workshop/history/issues/000005-five.md", "---\nid: 000005\nstatus: done\nactual_hours: 1\ncreated: 2026-09-01\n---\n\n# Five\n\n## Problem\n\nx\n")
	r.git("add", "-A")
	r.git("commit", "-qm", "main moves on")
	r.git("push", "-q", "origin", "main")
	digest, out, err := migrateDryRun(t)
	if err != nil {
		t.Fatalf("dry run: %v\n%s", err, out)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest); err != nil {
		t.Fatalf("apply: %v\n%s", err, stderr)
	}
	r.git("switch", "-q", "000002-two")
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--reconcile"); err != nil {
		t.Fatalf("reconcile refused untouched older copies: %v\n%s", err, stderr)
	}
	r.git("fetch", "-q", "origin")
	if got, want := r.git("show", "HEAD:"+one), r.git("show", "origin/main:"+one); got != want {
		t.Fatalf("#1 not taken from main:\n%s\n---\n%s", got, want)
	}
	if r.git("ls-tree", "--name-only", "HEAD", "--", five) != "" || r.git("ls-tree", "--name-only", "HEAD", "--", "workshop/history/issues/000005-five.md") == "" {
		t.Fatal("#5, archived on main, was not moved to history by the cutover merge")
	}
	if b := r.git("show", "HEAD:"+two); !issue.HasMirror([]byte(b)) || !strings.Contains(b, "a design step on the branch") {
		t.Fatalf("#2 not reconciled:\n%s", b)
	}
	if out, err := exec.Command("git", "-C", r.root, "merge", "--no-edit", "-q", "origin/main").CombinedOutput(); err != nil {
		t.Fatalf("merge after reconcile: %v\n%s\n%s", err, out, r.git("status", "--short"))
	}
}

// A stale worktree registration (its .git link gone — git reports it
// prunable) holds no checkout: the inventory skips it instead of dying on it
// (#252, found by the parley.nvim A4 pre-flight).
func TestIssueMigrateSkipsPrunableWorktrees(t *testing.T) {
	r := legacyRepo(t)
	stale := filepath.Join(t.TempDir(), "stale")
	r.git("worktree", "add", "-q", "-b", "scratch", stale)
	if err := os.Remove(filepath.Join(stale, ".git")); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.git("worktree", "list", "--porcelain"), "prunable") {
		t.Fatal("fixture: git does not report the stale worktree as prunable")
	}
	if _, out, err := migrateDryRun(t); err != nil {
		t.Fatalf("a prunable worktree broke the dry run: %v\n%s", err, out)
	}
}

// Uncommitted issue edits refuse the cutover, each named by its full path —
// including the first, whose porcelain line leads with a space the trimmed
// output drops (#252, "orkshop/…" in the ducks dry run).
func TestIssueMigrateNamesEveryUncommittedIssueEdit(t *testing.T) {
	r := legacyRepo(t)
	for _, p := range []string{"workshop/issues/000001-one.md", "workshop/issues/000002-two.md"} {
		writeRepoFile(t, r.root, p, "---\nid: x\n---\n\n# edited\n")
	}
	_, out, err := migrateDryRun(t)
	if err == nil {
		t.Fatalf("uncommitted issue edits must refuse the cutover:\n%s", out)
	}
	for _, p := range []string{": workshop/issues/000001-one.md\n", ": workshop/issues/000002-two.md\n"} {
		if !strings.Contains(out, p) {
			t.Errorf("refusal does not name %q:\n%s", strings.TrimSpace(p), out)
		}
	}
}
