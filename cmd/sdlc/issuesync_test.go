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

// issuePath206 is the file syncRepo seeds, as change-code would resolve it.
const issuePath206 = syncIssuesDir + "/000206-issue-sync-verb.md"

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
	// `git rm` of the last file in a directory removes the directory too, which
	// the untracked-file fixture relies on doing.
	if err := os.MkdirAll(filepath.Join(repo, syncIssuesDir), 0o755); err != nil {
		t.Fatal(err)
	}
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
		gitInDirResponses: map[string][]byte{
			"branch --show-current": []byte("main\n"),
			// The copy staged the file, so the commit runs. (The narrower
			// `-- <issuesDir>/` key below is mainHasUncommittedIssueChanges'
			// question, which must stay empty or the sync refuses.)
			"diff --cached --name-only -- " + issueFile: []byte(issueFile + "\n"),
		},
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
	got, err := archiveCommitArgs("archive completed issues to history", moves)
	if err != nil {
		t.Fatalf("archiveCommitArgs: %v", err)
	}
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
	commitArgs, err := archiveCommitArgs(archiveCommitMessage, moves)
	if err != nil {
		t.Fatal(err)
	}
	if out, err := r.Git(commitArgs...); err != nil {
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

// TestChangeCodeSyncIssue_ModeMatrix runs the full cross-product syncIssue
// promises over, rather than the one cell a reviewer happened to probe.
//
// Both close-review rounds found the same defect: a swapped helper covering
// fewer cells than the one it replaced. Round 1 was change-code's auto-detect
// name mode; round 2 was the feature-worktree location, where routing through
// the publish arm put a half-written Spec on origin/main and left the branch's
// own copy dirty. Patching each cell as it is found is how a second round
// happens, so the table is the test.
//
// The invariant every cell asserts: the issue file ends up COMMITTED IN THIS
// WORKTREE, on the branch about to carry the work. Publishing is conditioned on
// already being on main.
func TestChangeCodeSyncIssue_ModeMatrix(t *testing.T) {
	// location prepares a worktree to run in and returns (cwd, mainWorktree).
	type location struct {
		name      string
		publishes bool
		prepare   func(t *testing.T, repo string) string
	}
	locations := []location{
		{"on main", true, func(t *testing.T, repo string) string { return repo }},
		{"in-place feature branch", false, func(t *testing.T, repo string) string {
			git(t, repo, "switch", "-q", "-c", "000206-issue-sync-verb")
			return repo
		}},
		{"feature worktree", false, func(t *testing.T, repo string) string {
			wt := filepath.Join(t.TempDir(), "feature")
			git(t, repo, "worktree", "add", "-b", "000206-issue-sync-verb", wt)
			return wt
		}},
	}
	// The three name modes feed syncIssue inputs it never reads — it consults
	// neither f.Issue nor f.Name, only the resolved path — so 12 of the 18 cells
	// duplicate 6. That IS what they buy: reverting to `id := f.Issue` reds
	// exactly those 12, which is the BR-1 regression. Worth ~7s of suite time to
	// pin "gate on the resolved path, not the flag" as a property of the table
	// rather than a comment.
	nameModes := []struct {
		name  string
		flags func(dir string) *changeCodeFlags
	}{
		{"--issue", func(string) *changeCodeFlags {
			return &changeCodeFlags{Issue: 206, IssuesDir: syncIssuesDir}
		}},
		{"--name", func(string) *changeCodeFlags {
			return &changeCodeFlags{Name: "000206-issue-sync-verb", IssuesDir: syncIssuesDir}
		}},
		{"auto-detect", func(string) *changeCodeFlags {
			return &changeCodeFlags{IssuesDir: syncIssuesDir}
		}},
	}
	fileStates := []struct {
		name  string
		setup func(t *testing.T, dir string)
	}{
		{"tracked+edited", func(t *testing.T, dir string) {
			writeSyncIssue(t, dir, "000206-issue-sync-verb.md", "## Plan\n\nedited\n")
		}},
		{"untracked", func(t *testing.T, dir string) {
			// syncRepo commits 000206, so the brand-new-file case is built by
			// removing it from history and re-creating it. `git rm` (not
			// --cached) is deliberate: the commit that records the removal is
			// itself pathspec'd, and a partial commit takes the WORKING TREE
			// content of the named path — leaving the file on disk would make
			// that commit quietly re-add it.
			git(t, dir, "rm", "-q", "--", issuePath206)
			git(t, dir, "commit", "-q", "-m", "untrack the issue file", "--", issuePath206)
			writeSyncIssue(t, dir, "000206-issue-sync-verb.md", "## Plan\n\nbrand new\n")
		}},
	}

	for _, loc := range locations {
		for _, nm := range nameModes {
			for _, fsx := range fileStates {
				t.Run(loc.name+"/"+nm.name+"/"+fsx.name, func(t *testing.T) {
					repo, _ := syncRepo(t)
					dir := loc.prepare(t, repo)
					chdirTo(t, dir)
					fsx.setup(t, dir)
					originBefore := strings.TrimSpace(git(t, repo, "rev-parse", "origin/main"))

					var stderr bytes.Buffer
					syncIssue(&stderr, nm.flags(dir), issuePath206)

					// The invariant: committed HERE, on the branch cut for the work.
					if !strings.Contains(headFiles(t, dir), issuePath206) {
						t.Errorf("issue file not committed in this worktree; stderr:\n%s", stderr.String())
					}
					if status := git(t, dir, "status", "--porcelain", "--", syncIssuesDir); strings.TrimSpace(status) != "" {
						t.Errorf("issue file left dirty — the branch must start tracked; status:\n%s", status)
					}
					if got := headSubject(t, dir); got != issueSyncMessage(206, "spec/plan at change-code") {
						t.Errorf("subject = %q, want the change-code sync message", got)
					}

					originAfter := strings.TrimSpace(git(t, repo, "rev-parse", "origin/main"))
					if loc.publishes && originAfter == originBefore {
						t.Error("on main this is the milestone publish — origin/main should have moved")
					}
					if !loc.publishes && originAfter != originBefore {
						t.Errorf("from a branch the body must NOT reach origin/main (%s → %s): "+
							"publishing a half-written Spec is what `pr`/`merge`/`close` are for",
							originBefore, originAfter)
					}
				})
			}
		}
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
	syncIssue(&stderr, &changeCodeFlags{Issue: 206, IssuesDir: syncIssuesDir}, issuePath206)

	if !strings.Contains(stderr.String(), "issue file not synced") {
		t.Errorf("expected a warning naming the failure; stderr:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "sdlc issue sync --issue 206 --push") {
		t.Errorf("the warning must name the retry; stderr:\n%s", stderr.String())
	}
	// Durability still happened: only the push failed, so the commit is local.
	if !strings.Contains(headFiles(t, repo), issuePath206) {
		t.Error("the commit should still have landed locally before the push failed")
	}
}

// TestChangeCodeSyncIssue_NonIssuePathIsANoop: a --name branch can point at a
// file outside the NNNNNN- convention, so there is no id to name in the subject
// and nothing to sync. This is the ONLY case the guard should catch.
func TestChangeCodeSyncIssue_NonIssuePathIsANoop(t *testing.T) {
	repo, _ := syncRepo(t)
	before := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD"))
	notAnIssue := writeSyncIssue(t, repo, "freeform-branch.md", "no id here\n")

	var stderr bytes.Buffer
	syncIssue(&stderr, &changeCodeFlags{IssuesDir: syncIssuesDir}, notAnIssue)

	if after := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD")); after != before {
		t.Errorf("an id-less path should not commit (%s → %s)", before, after)
	}
	if stderr.String() != "" {
		t.Errorf("an id-less path should be silent, got: %s", stderr.String())
	}
}

// TestChangeCodeSyncIssue_DryRunCommitsNothing pins M2: the helper threads
// DryRun rather than relying on runChangeCode returning before it. A helper that
// commits must not depend on a caller's early return for dry-run correctness.
func TestChangeCodeSyncIssue_DryRunCommitsNothing(t *testing.T) {
	repo, _ := syncRepo(t)
	before := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD"))
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Plan\n\nnot yet\n")

	var stderr bytes.Buffer
	syncIssue(&stderr, &changeCodeFlags{Issue: 206, IssuesDir: syncIssuesDir, DryRun: true}, issuePath206)

	if after := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD")); after != before {
		t.Errorf("--dry-run committed (%s → %s)", before, after)
	}
}

// TestIssueSync_DryRunCommitsNothing is the same guarantee at the verb.
func TestIssueSync_DryRunCommitsNothing(t *testing.T) {
	repo, _ := syncRepo(t)
	before := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD"))
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Spec\n\nnot yet\n")

	var stdout, stderr bytes.Buffer
	f := &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir, DryRun: true}
	if err := runIssueSync(&stdout, &stderr, f); err != nil {
		t.Fatalf("runIssueSync --dry-run: %v", err)
	}
	if after := strings.TrimSpace(git(t, repo, "rev-parse", "HEAD")); after != before {
		t.Errorf("--dry-run committed (%s → %s)", before, after)
	}
	if !strings.Contains(stderr.String(), "dry-run") {
		t.Errorf("--dry-run should say so; stderr:\n%s", stderr.String())
	}
	if status := git(t, repo, "status", "--porcelain", "--", syncIssuesDir); strings.TrimSpace(status) == "" {
		t.Error("--dry-run must leave the edit uncommitted in the working tree")
	}
}

// TestArchiveCommitArgs_RefusesEmptyMoves pins the degenerate input. `git commit
// -m x --` with nothing after the separator is NOT a scoped commit — git reads
// it as no pathspec at all and records the whole index, which is precisely what
// this helper exists to prevent. Every call site guards on len(moves) > 0, so
// the error is unreachable today; it is here so the guard cannot be dropped
// silently later.
func TestArchiveCommitArgs_RefusesEmptyMoves(t *testing.T) {
	got, err := archiveCommitArgs(archiveCommitMessage, nil)
	if err == nil {
		t.Fatalf("archiveCommitArgs(nil) = %v, want an error — an empty pathspec commits the whole index", got)
	}
	if got != nil {
		t.Errorf("argv should be nil alongside the error, got %v", got)
	}
}

// ── migrate: the sixth member of the class ───────────────────────────────────

// TestMigrateCommit_LeavesForeignStagedFileAlone covers the source side, which
// is where nothing guards it: migrate's step (1) cleanliness check is scoped to
// the migrated path (`status --porcelain -- relPath`), so a peer's staged work
// elsewhere in the source repo sails past it. Before the pathspec, that work was
// swept into a commit reading "migrate: move X to Y".
func TestMigrateCommit_LeavesForeignStagedFileAlone(t *testing.T) {
	repo := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	moved := "docs/moved.md"
	if err := os.MkdirAll(filepath.Join(repo, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, moved), []byte("body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-m", "seed")

	// migrate's step (6) source side: remove the file, stage that one path.
	if err := os.Remove(filepath.Join(repo, moved)); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "add", "--", moved)
	foreign := stageForeignFile(t, repo, "peer-work.go")

	r := execGitRunner{}
	if out, err := r.Git("commit", "-q", "-m", "migrate: move "+moved+" to peer", "--", moved); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}

	committed := headFiles(t, repo)
	if strings.Contains(committed, foreign) {
		t.Errorf("foreign staged file swept into the migrate commit; files:\n%s", committed)
	}
	if !strings.Contains(committed, moved) {
		t.Errorf("migrate commit should record the removal of %s; files:\n%s", moved, committed)
	}
	if status := git(t, repo, "status", "--porcelain"); !strings.Contains(status, "A  "+foreign) {
		t.Errorf("foreign file should still be staged, status:\n%s", status)
	}
}

// TestMigrateCommitArgs_HintAndCommandAreTheSameBuilder pins the ARCH-DRY point
// of migrateCommitArgs: `migrate --no-commit` prints the command for the
// operator to run, and that printed command must be the one migrate would have
// run. Before #206 both were hand-written inline, four times, and the printed
// pair could have lost its pathspec without a test noticing.
func TestMigrateCommitArgs_HintAndCommandAreTheSameBuilder(t *testing.T) {
	args := migrateCommitArgs("migrate: move a/b.md to peer", "a/b.md")
	want := []string{"commit", "-q", "-m", "migrate: move a/b.md to peer", "--", "a/b.md"}
	if strings.Join(args, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("migrateCommitArgs = %v, want %v", args, want)
	}
	// The hint renders the same argv, with only the subject quoted — so a reader
	// can paste it, and a lost `--` would change both at once or neither.
	hint := strings.Join(quoteMigrateMsg(args), " ")
	if !strings.Contains(hint, `-m "migrate: move a/b.md to peer"`) {
		t.Errorf("hint must quote the subject (it contains spaces): %s", hint)
	}
	if !strings.HasSuffix(hint, "-- a/b.md") {
		t.Errorf("hint must carry the pathspec the executed commit carries: %s", hint)
	}
}

// TestIssueSync_PushPublishesAnAlreadyCommittedBody is the C1 regression from
// close-review round 3, and the sequence the whole design makes normal: sync
// locally (no push), then publish.
//
// syncInPlace's "no working-tree changes" early return predates #206, when
// committing and publishing were one act. Splitting them made "committed
// locally, not yet pushed" a state the code deliberately creates — and in it
// changedIssueFiles is empty, so `--push` short-circuited and reported success
// in green while origin/main never moved. Both of the warnings this issue added
// name that exact command as the recovery, so the no-op was load-bearing.
func TestIssueSync_PushPublishesAnAlreadyCommittedBody(t *testing.T) {
	repo, _ := syncRepo(t)
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Spec\n\nthe design\n")

	var stdout, stderr bytes.Buffer
	local := &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir}
	if err := runIssueSync(&stdout, &stderr, local); err != nil {
		t.Fatalf("local sync: %v", err)
	}
	head := strings.TrimSpace(git(t, repo, "rev-parse", "main"))
	if origin := strings.TrimSpace(git(t, repo, "rev-parse", "origin/main")); origin == head {
		t.Fatal("fixture broken: the local sync should NOT have published")
	}

	// Nothing has changed in the working tree now — the body is committed. This
	// is precisely the state --push exists to finish.
	stdout.Reset()
	stderr.Reset()
	publish := &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir, Push: true}
	if err := runIssueSync(&stdout, &stderr, publish); err != nil {
		t.Fatalf("publish sync: %v (stderr: %s)", err, stderr.String())
	}
	if origin := strings.TrimSpace(git(t, repo, "rev-parse", "origin/main")); origin != head {
		t.Errorf("--push did not publish the already-committed body: origin/main %s, main %s\nstderr:\n%s",
			origin, head, stderr.String())
	}
}

// TestChangeCodeSyncIssue_RetryAfterAFailedPushPublishes walks the other
// consumer's advertised recovery end to end: change-code's commit lands, its
// push fails (no origin), the operator adds the remote and re-runs the command
// the warning printed. If that command can't finish the publish, the warning is
// a dead end.
func TestChangeCodeSyncIssue_RetryAfterAFailedPushPublishes(t *testing.T) {
	repo := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Plan\n\nthe accepted design\n")

	var stderr bytes.Buffer
	syncIssue(&stderr, &changeCodeFlags{Issue: 206, IssuesDir: syncIssuesDir}, issuePath206)
	if !strings.Contains(stderr.String(), "sdlc issue sync --issue 206 --push") {
		t.Fatalf("expected the retry advice; stderr:\n%s", stderr.String())
	}
	head := strings.TrimSpace(git(t, repo, "rev-parse", "main"))

	// Clear the cause the warning told us to clear.
	origin := filepath.Join(t.TempDir(), "origin.git")
	git(t, "", "init", "--bare", "-b", "main", origin)
	git(t, repo, "remote", "add", "origin", origin)

	var stdout, stderr2 bytes.Buffer
	if err := runIssueSync(&stdout, &stderr2, &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir, Push: true}); err != nil {
		t.Fatalf("advertised retry failed: %v (stderr: %s)", err, stderr2.String())
	}
	if got := strings.TrimSpace(git(t, repo, "rev-parse", "origin/main")); got != head {
		t.Errorf("the retry the warning names must publish: origin/main %s, main %s", got, head)
	}
}

// TestIssueSync_PublishMatrix is the table round 3 should have written. Round 3
// fixed "nothing to commit is not nothing to publish" on syncInPlace, measured
// it there, and shipped the OTHER arm inverted: syncViaMainWorktree pushed main
// without carrying the body across, so from a feature worktree `--push` printed
// green success while origin/main never moved. Naming a class and sweeping one
// of its two members is what a table prevents.
//
// The invariant, across every location × body state: after `--push`, the body
// that is in THIS worktree is the body on origin/main.
func TestIssueSync_PublishMatrix(t *testing.T) {
	const body = "## Spec\n\nthe published design\n"

	locations := []struct {
		name    string
		prepare func(t *testing.T, repo string) string
	}{
		{"on main", func(t *testing.T, repo string) string { return repo }},
		{"feature worktree", func(t *testing.T, repo string) string {
			wt := filepath.Join(t.TempDir(), "feature")
			git(t, repo, "worktree", "add", "-b", "000206-issue-sync-verb", wt)
			return wt
		}},
	}
	bodyStates := []struct {
		name  string
		setup func(t *testing.T, dir string)
	}{
		{"dirty", func(t *testing.T, dir string) {
			writeSyncIssue(t, dir, "000206-issue-sync-verb.md", body)
		}},
		{"already committed locally", func(t *testing.T, dir string) {
			writeSyncIssue(t, dir, "000206-issue-sync-verb.md", body)
			var o, e bytes.Buffer
			if err := runIssueSync(&o, &e, &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir}); err != nil {
				t.Fatalf("local pre-sync: %v (stderr: %s)", err, e.String())
			}
		}},
	}

	for _, loc := range locations {
		for _, bs := range bodyStates {
			t.Run(loc.name+"/"+bs.name, func(t *testing.T) {
				repo, _ := syncRepo(t)
				dir := loc.prepare(t, repo)
				chdirTo(t, dir)
				bs.setup(t, dir)

				var stdout, stderr bytes.Buffer
				err := runIssueSync(&stdout, &stderr, &issueSyncFlags{
					Issue: 206, IssuesDir: syncIssuesDir, Push: true,
				})
				if err != nil {
					t.Fatalf("--push: %v (stderr: %s)", err, stderr.String())
				}

				// The published body must match what this worktree holds. Reading
				// it out of origin is what catches a push that moved nothing —
				// comparing SHAs would not, since the two worktrees' mains differ
				// legitimately in the branch case.
				published := git(t, repo, "show", "origin/main:"+issuePath206)
				if strings.TrimSpace(published) != strings.TrimSpace(body) {
					t.Errorf("origin/main does not carry this worktree's body.\n got: %q\nwant: %q\nstderr:\n%s",
						published, body, stderr.String())
				}
			})
		}
	}
}

// TestSyncInPlace_CleanAndPublishedDoesNotTouchTheNetwork pins the other half of
// round 3's over-correction. `sdlc claim` is idempotent by design and die()s on
// a sync error, so making the no-changes path push unconditionally turned an
// offline re-run — previously `[ok] No issue changes to sync.`, exit 0 — into a
// fatal error. Publication is the gap between origin/main and local main; when
// that gap is empty there is nothing to publish and no reason to reach out.
func TestSyncInPlace_CleanAndPublishedDoesNotTouchTheNetwork(t *testing.T) {
	repo, _ := syncRepo(t)
	// Point origin at a path that no longer exists: any network attempt fails.
	git(t, repo, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))

	var stdout, stderr bytes.Buffer
	// Publishing caller (NoPush unset), clean tree, main == origin/main.
	if err := syncIssuesToMain(&stdout, &stderr, &claimFlags{Issue: 206, IssuesDir: syncIssuesDir}, execGitRunner{}, ""); err != nil {
		t.Fatalf("a clean, already-published tree must be a no-op, got: %v (stderr: %s)", err, stderr.String())
	}
	if !strings.Contains(stderr.String(), "No issue changes to sync") {
		t.Errorf("expected the idempotent no-op; stderr:\n%s", stderr.String())
	}
}

// TestPublishIsIdempotent walks the workflow this issue's own docs recommend,
// twice, from both locations. Round 4 made the publish arm re-seed its file list
// from the issue — correct about *what* to publish — and fed that file into a
// conflict detector built for genuinely-dirty files. After the first publish,
// main legitimately carries the body, so "changed on both sides since
// merge-base" is true and the second run died with `Conflict detected!` and a
// manual-merge guide. `sdlc claim` had been idempotent since it existed.
//
// The rule: a body main already carries byte-for-byte is nothing to route, so
// the conflict detector is never asked about it.
func TestPublishIsIdempotent(t *testing.T) {
	for _, loc := range []struct {
		name    string
		prepare func(t *testing.T, repo string) string
	}{
		{"on main", func(t *testing.T, repo string) string { return repo }},
		{"feature worktree", func(t *testing.T, repo string) string {
			wt := filepath.Join(t.TempDir(), "feature")
			git(t, repo, "worktree", "add", "-b", "000206-issue-sync-verb", wt)
			return wt
		}},
	} {
		t.Run(loc.name, func(t *testing.T) {
			repo, _ := syncRepo(t)
			dir := loc.prepare(t, repo)
			chdirTo(t, dir)
			writeSyncIssue(t, dir, "000206-issue-sync-verb.md", "## Spec\n\nthe design\n")

			run := func(label string, f *issueSyncFlags) {
				t.Helper()
				var stdout, stderr bytes.Buffer
				if err := runIssueSync(&stdout, &stderr, f); err != nil {
					t.Fatalf("%s: %v\nstderr:\n%s", label, err, stderr.String())
				}
			}
			// The documented workflow: checkpoint locally, then publish. Then
			// publish again — an agent re-running a verb is the normal case, not
			// an edge one.
			run("local sync", &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir})
			run("first publish", &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir, Push: true})
			run("second publish", &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir, Push: true})
			// And a third time with no --issue at all, which is how `sdlc claim`
			// re-syncs: it must stay the clean no-op it has always been.
			run("bare re-sync", &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir})
		})
	}
}

// TestClaimStaysIdempotentOffline pins the other half. `sdlc claim` die()s on a
// sync error and is re-run constantly, so a clean-tree claim must not reach for
// the network at all. An earlier cut inferred "is there anything to publish"
// from `origin/main..main`, which made every clean claim a wholesale push of
// local main — publishing bodies a no-push sync had deliberately kept local, and
// failing outright when offline. Only `issue sync --push` asks for that now.
func TestClaimStaysIdempotentOffline(t *testing.T) {
	repo, _ := syncRepo(t)
	writeSyncIssue(t, repo, "000206-issue-sync-verb.md", "## Spec\n\nkept local on purpose\n")

	var stdout, stderr bytes.Buffer
	if err := runIssueSync(&stdout, &stderr, &issueSyncFlags{Issue: 206, IssuesDir: syncIssuesDir}); err != nil {
		t.Fatalf("local sync: %v", err)
	}
	localHead := strings.TrimSpace(git(t, repo, "rev-parse", "main"))

	// Origin is now unreachable: any network attempt fails loudly.
	git(t, repo, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))

	stdout.Reset()
	stderr.Reset()
	// A publishing caller (claim's flag shape), clean tree, body committed locally.
	if err := syncIssuesToMain(&stdout, &stderr, &claimFlags{Issue: 206, IssuesDir: syncIssuesDir}, execGitRunner{}, ""); err != nil {
		t.Fatalf("a clean-tree claim must be an offline no-op, got: %v\nstderr:\n%s", err, stderr.String())
	}
	if !strings.Contains(stderr.String(), "No issue changes to sync") {
		t.Errorf("expected the idempotent no-op; stderr:\n%s", stderr.String())
	}
	if got := strings.TrimSpace(git(t, repo, "rev-parse", "main")); got != localHead {
		t.Errorf("claim must not have moved main (%s → %s)", localHead, got)
	}
}
