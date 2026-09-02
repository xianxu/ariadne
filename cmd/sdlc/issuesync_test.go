package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// ── fixtures ─────────────────────────────────────────────────────────────────
//
// Every test in this file runs against a REAL git repo. The thing under test is
// git's `--only` pathspec semantics — what a commit records when the index holds
// more than the caller staged — and an argv-recording gitRunner fake cannot
// observe that: it would assert the argument list and pass even if the semantics
// were wrong (ARCH-MOCK). These tests chdir the process, so none of them is
// parallel.

const syncIssuesDir = "workshop/issues"

// syncRepo builds a repo on main with a bare origin, an issues dir, and one
// committed issue file, then chdirs in. Returns (repo, origin).
func syncRepo(t *testing.T) (string, string) {
	t.Helper()
	origin := filepath.Join(t.TempDir(), "origin.git")
	git(t, "", "init", "--bare", "-b", "main", origin)
	repo := testfix.Repo(t, testfix.InitialCommit())
	if err := os.MkdirAll(filepath.Join(repo, syncIssuesDir), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Spec\n\nthe reservation only\n")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "seed issue")
	git(t, repo, "remote", "add", "origin", origin)
	git(t, repo, "push", "-u", "origin", "main")
	chdirTo(t, repo)
	return repo, origin
}

func writeSyncIssue(t *testing.T, repo, name, body string) string {
	t.Helper()
	path := filepath.Join(repo, syncIssuesDir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func chdirTo(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
}

// stageForeignFile writes and `git add`s a file that no sdlc verb staged —
// standing in for a peer agent working in the same checkout. The repo
// transaction lock serializes sdlc verbs against each other; it does nothing
// about a peer running plain `git add`.
func stageForeignFile(t *testing.T, dir, name string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("peer work in progress\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "--", name)
	return name
}

func headFiles(t *testing.T, dir string) string {
	t.Helper()
	return git(t, dir, "diff-tree", "--no-commit-id", "--name-only", "-r", "HEAD")
}

func headSubject(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(git(t, dir, "log", "-1", "--format=%s"))
}

// ── the swept-index regression, per sync arm ─────────────────────────────────

// TestSyncInPlace_LeavesForeignStagedFileAlone is #206's core regression. Before
// the fix, syncOnMain narrowed `git add` to one issue and then ran a BARE `git
// commit`, which records the whole index — so a peer agent's staged work was
// swept into a commit that said it was syncing issues.
func TestSyncInPlace_LeavesForeignStagedFileAlone(t *testing.T) {
	repo, _ := syncRepo(t)
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Spec\n\nthe design, at last\n")
	foreign := stageForeignFile(t, repo, "peer-work.go")

	var stdout, stderr bytes.Buffer
	f := &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir}
	if err := runIssueSync(&stdout, &stderr, f); err != nil {
		t.Fatalf("runIssueSync: %v (stderr: %s)", err, stderr.String())
	}

	committed := headFiles(t, repo)
	if strings.Contains(committed, foreign) {
		t.Errorf("the foreign staged file was swept into the sync commit; files:\n%s", committed)
	}
	if !strings.Contains(committed, syncIssuesDir+"/000206-issue-sync-verb.md") {
		t.Errorf("the issue file should be in the sync commit; files:\n%s", committed)
	}
	// Still staged, exactly as the peer left it: `--only` must not touch it.
	if status := git(t, repo, "status", "--porcelain"); !strings.Contains(status, "A  "+foreign) {
		t.Errorf("foreign file should still be staged (A), status:\n%s", status)
	}
}

// TestSyncViaMainWorktree_CommitsOnlyTheCopiedIssueFiles is the publish arm's
// first coverage, and it pins a narrower claim than the in-place arm's.
//
// The swept-index bug is NOT deterministically reachable here: step 4 runs `git
// pull --rebase origin main` in the main worktree, and git refuses that outright
// when the index is dirty, so a peer's staged file makes the whole sync fail
// loudly long before the commit. (An earlier draft of this issue asserted the
// arm carried "the identical defect" — it does not, and the fix here is
// defense-in-depth closing the pull→commit race rather than a live bug.)
//
// So this proves what IS reachable: the arm works end to end, the commit in the
// OTHER worktree records exactly the copied issue files under the caller's
// subject, and an untracked peer file sitting in the main worktree is left
// alone. TestSyncViaMainWorktree_CommitCarriesPathspec covers the wiring.
func TestSyncViaMainWorktree_CommitsOnlyTheCopiedIssueFiles(t *testing.T) {
	mainWT, _ := syncRepo(t)
	feature := filepath.Join(t.TempDir(), "feature")
	git(t, mainWT, "worktree", "add", "-b", "000206-issue-sync-verb", feature)
	chdirTo(t, feature)

	writeSyncIssue(t, feature, "000206-issue-sync-verb.md", "## Spec\n\nedited on the branch\n")
	if err := os.WriteFile(filepath.Join(mainWT, "peer-work.go"), []byte("peer wip\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	f := &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir, Push: true}
	if err := runIssueSync(&stdout, &stderr, f); err != nil {
		t.Fatalf("runIssueSync: %v (stderr: %s)", err, stderr.String())
	}

	committed := strings.Fields(headFiles(t, mainWT))
	want := []string{syncIssuesDir + "/000206-issue-sync-verb.md"}
	if len(committed) != 1 || committed[0] != want[0] {
		t.Errorf("main-worktree commit files = %v, want exactly %v", committed, want)
	}
	if got := headSubject(t, mainWT); got != issueSyncMessage(206, "spec/plan") {
		t.Errorf("main-worktree commit subject = %q, want the issue-naming message", got)
	}
	if status := git(t, mainWT, "status", "--porcelain"); !strings.Contains(status, "?? peer-work.go") {
		t.Errorf("untracked peer file should be untouched in the main worktree, status:\n%s", status)
	}
	local := strings.TrimSpace(git(t, mainWT, "rev-parse", "main"))
	if remote := strings.TrimSpace(git(t, mainWT, "rev-parse", "origin/main")); remote != local {
		t.Errorf("publish arm should push; local main %s, origin/main %s", local, remote)
	}
}

// TestSyncViaMainWorktree_CommitCarriesPathspec is the wiring half. The
// SEMANTICS of a pathspec'd commit are proven once against real git by
// TestSyncInPlace_LeavesForeignStagedFileAlone; what is left to prove for this
// arm is that it passes one at all, which is an argv question — and the only
// question a recording runner is entitled to answer.
func TestSyncViaMainWorktree_CommitCarriesPathspec(t *testing.T) {
	const issueFile = "workshop/issues/000206-issue-sync-verb.md"
	// The copy step (step 6) touches the real filesystem rather than the runner,
	// so the source repo and the stand-in main worktree are real directories
	// even though every git call is stubbed.
	repo := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	if err := os.MkdirAll(filepath.Join(repo, syncIssuesDir), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Spec\n\nx\n")
	fakeMain := t.TempDir()
	r := &claimRunnerStub{
		responses: map[string][]byte{
			"diff --name-only HEAD": []byte(issueFile + "\n"),
			"worktree list":         []byte(worktreePorcelainZ([]string{"worktree " + fakeMain, "HEAD abc", "branch refs/heads/main"})),
			"merge-base":            []byte("abc123\n"),
		},
		gitInDirResponses: map[string][]byte{"branch --show-current": []byte("main\n")},
	}
	var stdout, stderr bytes.Buffer
	f := &claimFlags{Issue: 206, IssuesDir: syncIssuesDir}
	if err := syncViaMainWorktree(&stdout, &stderr, f, "000206-issue-sync-verb", r, "#206: issue-sync: spec/plan"); err != nil {
		t.Fatalf("syncViaMainWorktree: %v (stderr: %s)", err, stderr.String())
	}
	var commit []string
	for _, c := range r.gitInDirCalls {
		if len(c.Args) > 0 && c.Args[0] == "commit" {
			commit = c.Args
		}
	}
	if commit == nil {
		t.Fatalf("no commit issued; calls: %v", r.gitInDirCalls)
	}
	sep := -1
	for i, a := range commit {
		if a == "--" {
			sep = i
		}
	}
	if sep < 0 {
		t.Fatalf("commit has no `--` pathspec separator: %v — a bare commit records the whole index", commit)
	}
	if got := commit[sep+1:]; len(got) != 1 || got[0] != issueFile {
		t.Errorf("commit pathspec = %v, want [%s] (the same paths the add staged)", got, issueFile)
	}
}

// ── the verb's publish contract ──────────────────────────────────────────────

// TestIssueSync_DefaultDoesNotPush pins the design decision the verb exists to
// make safe: durability and publication are separable, and a mid-planning sync
// buys only durability. A half-written Spec must not be able to escape.
func TestIssueSync_DefaultDoesNotPush(t *testing.T) {
	repo, _ := syncRepo(t)
	before := strings.TrimSpace(git(t, repo, "rev-parse", "origin/main"))
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Spec\n\nhalf a thought\n")

	var stdout, stderr bytes.Buffer
	f := &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir}
	if err := runIssueSync(&stdout, &stderr, f); err != nil {
		t.Fatalf("runIssueSync: %v", err)
	}

	if after := strings.TrimSpace(git(t, repo, "rev-parse", "origin/main")); after != before {
		t.Errorf("default sync moved origin/main (%s → %s); it must not publish", before, after)
	}
	if local := strings.TrimSpace(git(t, repo, "rev-parse", "main")); local == before {
		t.Error("default sync did not commit locally; durability is the whole point")
	}
	if got := headSubject(t, repo); got != issueSyncMessage(206, "spec/plan") {
		t.Errorf("commit subject = %q, want %q", got, issueSyncMessage(206, "spec/plan"))
	}
}

// TestIssueSync_PushPublishes is the opt-in half.
func TestIssueSync_PushPublishes(t *testing.T) {
	repo, _ := syncRepo(t)
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Spec\n\nready to publish\n")

	var stdout, stderr bytes.Buffer
	f := &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir, Push: true}
	if err := runIssueSync(&stdout, &stderr, f); err != nil {
		t.Fatalf("runIssueSync: %v", err)
	}
	local := strings.TrimSpace(git(t, repo, "rev-parse", "main"))
	if remote := strings.TrimSpace(git(t, repo, "rev-parse", "origin/main")); remote != local {
		t.Errorf("--push should leave origin/main at the sync commit; local %s remote %s", local, remote)
	}
}

// TestIssueSync_FromFeatureBranchCommitsOnThatBranch pins where the durable
// commit lands. The no-push arm is selected BEFORE the branch test precisely so
// this case never reaches the publish route: that route hunts for a worktree on
// main (which in-place branching never creates), pulls over the network, and
// commits somewhere the caller isn't — none of which is durability.
func TestIssueSync_FromFeatureBranchCommitsOnThatBranch(t *testing.T) {
	repo, _ := syncRepo(t)
	git(t, repo, "switch", "-c", "000206-issue-sync-verb")
	mainBefore := strings.TrimSpace(git(t, repo, "rev-parse", "main"))
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Log\n\nmid-implementation note\n")

	var stdout, stderr bytes.Buffer
	f := &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir}
	if err := runIssueSync(&stdout, &stderr, f); err != nil {
		t.Fatalf("runIssueSync from a feature branch: %v (stderr: %s)", err, stderr.String())
	}

	if !strings.Contains(headFiles(t, repo), syncIssuesDir+"/000206-issue-sync-verb.md") {
		t.Error("the commit should be on the feature branch's HEAD")
	}
	if got := strings.TrimSpace(git(t, repo, "rev-parse", "main")); got != mainBefore {
		t.Errorf("a no-push sync must not move main (%s → %s)", mainBefore, got)
	}
	if status := git(t, repo, "status", "--porcelain", "--", syncIssuesDir); strings.TrimSpace(status) != "" {
		t.Errorf("issue file should be committed, not left dirty; status:\n%s", status)
	}
}

// TestIssueSync_RequiresIssue: the commit message names the issue, so there is
// no issue-less form of this verb.
func TestIssueSync_RequiresIssue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	msg, died := expectDie(t, func() {
		_ = runIssueSync(&stdout, &stderr, &issueSyncFlags{IssuesDir: syncIssuesDir})
	})
	if !died {
		t.Fatal("expected a refusal without --issue")
	}
	if !strings.Contains(msg, "--issue N is required") {
		t.Errorf("refusal should name the missing flag, got %q", msg)
	}
}

// TestIssueSyncMessage_IsClassifiedAsBookkeeping guards the coupling that is
// easy to break from the sdlc side: the subject this verb writes anchors #N, so
// it is only kept out of shipped-work windows by gitx's bookkeeping list. Change
// the prefix here without changing that list and drift detection, milestone
// review windows and active-time attribution all start counting tracker commits
// as implementation — silently.
func TestIssueSyncMessage_IsClassifiedAsBookkeeping(t *testing.T) {
	if got := issueSyncMessage(206, "spec/plan"); got != "#206: issue-sync: spec/plan" {
		t.Errorf("issueSyncMessage = %q, want the tree's #N: <area>: <subject> shape", got)
	}
	for _, what := range []string{"spec/plan", "spec/plan at change-code"} {
		subject := issueSyncMessage(206, what)
		if gitx.IsShippedWorkSubject("206", subject) {
			t.Errorf("%q is classified as shipped work; add its lead-in to gitx.bookkeepingVerbs", subject)
		}
	}
}

// ── the archive commits: the rest of the narrowed-add/bare-commit class ──────

// TestArchiveCommitArgs_MirrorsTheAdd is the pure half. archiveAddArgs stages
// exactly the paths an archive touched; before #206 the commit that followed
// carried no pathspec at all and recorded the whole index. Deriving the commit
// list FROM the add list is what keeps the two from drifting — a commit pathspec
// wider than its add is the bug itself (ARCH-DRY).
func TestArchiveCommitArgs_MirrorsTheAdd(t *testing.T) {
	moves := []preparedArchiveMove{
		{IssuePath: "workshop/issues/000206-a.md", HistoryPath: "workshop/history/000206-a.md"},
		{IssuePath: "workshop/issues/000207-b.md", HistoryPath: "workshop/history/000207-b.md", SourceUntracked: true},
	}
	got := archiveCommitArgs("archive completed issues to history", moves)
	want := []string{
		"commit", "-m", "archive completed issues to history", "--",
		"workshop/issues/000206-a.md", "workshop/history/000206-a.md",
		// 000207's source was untracked: it vanished at the rename, so only the
		// destination is named — exactly as archiveAddArgs decided.
		"workshop/history/000207-b.md",
	}
	if len(got) != len(want) {
		t.Fatalf("archiveCommitArgs = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("archiveCommitArgs = %v, want %v", got, want)
		}
	}
	// The pathspec must be the add's path list verbatim, derived not restated.
	if add := archiveAddArgs(moves); len(add)-2 != len(got)-4 {
		t.Errorf("commit pathspec (%d paths) and add pathspec (%d paths) disagree", len(got)-4, len(add)-2)
	}
}

// TestArchiveCommit_LeavesForeignStagedFileAlone is the real-git half, on the
// call site where the sweep is reachable: `sdlc push` auto-commits tracked
// changes but does not refuse a dirty index, so a peer's staged file sits in the
// index when the archive commit runs.
func TestArchiveCommit_LeavesForeignStagedFileAlone(t *testing.T) {
	repo := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	for _, d := range []string{"workshop/issues", "workshop/history"} {
		if err := os.MkdirAll(filepath.Join(repo, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	issuePath := "workshop/issues/000206-done.md"
	historyPath := "workshop/history/000206-done.md"
	if err := os.WriteFile(filepath.Join(repo, issuePath), []byte("---\nid: 000206\nstatus: done\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "seed")

	// The archive move, performed the way archiveDoneIssues would.
	if err := os.Rename(filepath.Join(repo, issuePath), filepath.Join(repo, historyPath)); err != nil {
		t.Fatal(err)
	}
	moves := []preparedArchiveMove{{IssuePath: issuePath, HistoryPath: historyPath}}
	foreign := stageForeignFile(t, repo, "peer-work.go")

	r := execGitRunner{}
	if out, err := r.Git(archiveAddArgs(moves)...); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	if out, err := r.Git(archiveCommitArgs(archiveCommitMessage, moves)...); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}

	committed := headFiles(t, repo)
	if strings.Contains(committed, foreign) {
		t.Errorf("foreign staged file swept into the archive commit; files:\n%s", committed)
	}
	if !strings.Contains(committed, historyPath) || !strings.Contains(committed, issuePath) {
		t.Errorf("archive commit should record the move (delete %s, add %s); files:\n%s", issuePath, historyPath, committed)
	}
	if status := git(t, repo, "status", "--porcelain"); !strings.Contains(status, "A  "+foreign) {
		t.Errorf("foreign file should still be staged, status:\n%s", status)
	}
}

// ── change-code's call ───────────────────────────────────────────────────────

// TestChangeCodeSyncIssue_CommitsAndPushes: the milestone case. Planning ends on
// main, plan-quality has accepted the design, so the issue file is committed AND
// published before any code is written — and the branch about to be cut starts
// from a tracked state, the property commitUntrackedIssueFile existed for.
func TestChangeCodeSyncIssue_CommitsAndPushes(t *testing.T) {
	repo, _ := syncRepo(t)
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Plan\n\nthe accepted design\n")

	var stderr bytes.Buffer
	syncIssue(&stderr, &changeCodeFlags{Issue: 206, IssuesDir: syncIssuesDir})

	if got := headSubject(t, repo); got != issueSyncMessage(206, "spec/plan at change-code") {
		t.Errorf("commit subject = %q, want the change-code sync message", got)
	}
	local := strings.TrimSpace(git(t, repo, "rev-parse", "main"))
	if remote := strings.TrimSpace(git(t, repo, "rev-parse", "origin/main")); remote != local {
		t.Errorf("change-code's sync publishes; local %s, origin/main %s", local, remote)
	}
	if status := git(t, repo, "status", "--porcelain", "--", syncIssuesDir); strings.TrimSpace(status) != "" {
		t.Errorf("branch must start from a tracked issue file; status:\n%s", status)
	}
}

// TestChangeCodeSyncIssue_WarnsRatherThanDying pins the posture. The helper this
// replaced warned on a failed push rather than dying; extending that to the
// whole sync is deliberate, because change-code's job is to OPEN implementation
// and a tracker commit that could not land must not stand between the operator
// and starting work. A die() here would surface as a test-process exit, so
// reaching the assertions at all is part of the assertion.
func TestChangeCodeSyncIssue_WarnsRatherThanDying(t *testing.T) {
	repo := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	if err := os.MkdirAll(filepath.Join(repo, syncIssuesDir), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Plan\n\nno origin to publish to\n")
	// NOTE: no `git remote add origin` — the push has nowhere to go.

	var stderr bytes.Buffer
	syncIssue(&stderr, &changeCodeFlags{Issue: 206, IssuesDir: syncIssuesDir})

	if !strings.Contains(stderr.String(), "issue file not synced") {
		t.Errorf("expected a warning naming the failure; stderr:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "sdlc issue sync --issue 206 --push") {
		t.Errorf("the warning must name the retry; stderr:\n%s", stderr.String())
	}
	// Durability still happened: only the push failed, so the commit is local.
	if !strings.Contains(headFiles(t, repo), syncIssuesDir+"/000206-issue-sync-verb.md") {
		t.Error("the commit should still have landed locally before the push failed")
	}
}

// TestChangeCodeSyncIssue_NoIssueIsANoop: `--name` mode has no issue ID to name
// in the commit message, so there is nothing to sync.
func TestChangeCodeSyncIssue_NoIssueIsANoop(t *testing.T) {
	repo, _ := syncRepo(t)
	before := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD"))
	var stderr bytes.Buffer
	syncIssue(&stderr, &changeCodeFlags{IssuesDir: syncIssuesDir})
	if after := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD")); after != before {
		t.Errorf("--name mode should not commit (%s → %s)", before, after)
	}
	if stderr.String() != "" {
		t.Errorf("--name mode should be silent, got: %s", stderr.String())
	}
}
