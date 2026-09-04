// claim.go — `sdlc claim [--issue N]` subcommand.
//
// Implements the issue-file synchronizer that commits + pushes
// workshop/issues/ changes to origin/main even when
// the operator is on a feature branch. Used as the workstream claim
// primitive: agents claim work by flipping status to `working` and
// running `sdlc claim` to broadcast that claim to origin/main.
//
// Two synchronization paths. The discriminator is NOT "main vs feature branch"
// (#206) — it is "commit here" vs "publish to origin/main from somewhere else":
//
//  1. syncInPlace: add + commit in THIS worktree, on THIS branch, then push
//     origin main when publishing. Reached whenever the caller isn't
//     publishing (NoPush) — offline-safe, no worktree hunt — or when this
//     worktree is already on main, where "here" and "main" coincide.
//  2. syncViaMainWorktree: the publish-from-a-feature-branch route.
//     - locate the main worktree via `git worktree list --porcelain -z`
//     - check main worktree has no uncommitted issue changes
//     - pull --rebase origin main on the main worktree
//     - detect conflicts (files changed on both branches since merge-base)
//     - copy changed issue files from feature worktree → main worktree
//     - commit + push on main worktree
//
// Every step of (2) exists to publish, which is why suppressing the push
// doesn't just skip the last line — it selects the other arm entirely.
//
// The command supports --issue (filter the sync to one issue file),
// --issues-dir (env override), and --dry-run.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// claimFlags holds the parsed flag values for the claim subcommand.
type claimFlags struct {
	Issue     int
	IssuesDir string
	DryRun    bool
	NoStart   bool
	// NoPush suppresses publication: commit in the current worktree and stop
	// (#206). Spelled negatively on purpose — `issue new` builds this struct as
	// a literal (issue.go), so a positive `Push` would zero-value to false
	// there and silently kill the reservation broadcast #82 M1 added. The zero
	// value is today's behavior for every existing caller.
	NoPush bool
	// PublishExisting asks for a publish even when nothing is dirty — the state
	// a prior no-push sync leaves behind, and the one `sdlc issue sync --push`
	// exists to finish. ONLY that verb sets it.
	//
	// Without it, "nothing to commit" stays the pre-#206 no-op for `claim` and
	// `issue new`, which is what they have always done and what makes `claim`
	// idempotent and usable offline. An earlier cut inferred this from
	// `origin/main..main` instead, which quietly turned every clean-tree claim
	// into a wholesale push of local main — publishing bodies a no-push sync had
	// deliberately kept local.
	PublishExisting bool
}

// NewClaimCmd returns the cobra command for `sdlc claim`.
func NewClaimCmd() *cobra.Command {
	f := claimFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:           "claim",
		Short:         "Sync workshop/issues/ changes to origin/main (workstream-claim primitive)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"lock\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			guardSpineRepo(cmd.ErrOrStderr()) // #176 lifecycle guard
			return runClaim(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "sync only this issue's file (default: all changed issue files)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print what would happen; do not commit/push")
	cmd.Flags().BoolVar(&f.NoStart, "no-start", false, "do not auto-flip an open --issue to working before syncing")
	return cmd
}

// claimRunner is the same gitRunner interface used by start; reused so
// tests can inject capture runners across both verbs.
var claimRunner gitRunner = execGitRunner{}

// runClaim dispatches to sync-on-main or sync-on-branch based on the
// current branch, exactly like the shell source. Before syncing it folds
// in the "start work" status flip (startOnClaim) so claiming an open issue
// is a single command (AGENTS.md §0).
func runClaim(stdout, stderr io.Writer, f *claimFlags) error {
	if f.Issue > 0 && !f.NoStart {
		if err := startOnClaim(stdout, stderr, f); err != nil {
			die(stderr, err.Error())
		}
	}
	// claim's whole job is the sync, so a failure is fatal (preserves the prior
	// die()-on-error UX now that the sync helpers return errors instead).
	if err := syncIssuesToMain(stdout, stderr, f, claimRunner, ""); err != nil {
		die(stderr, err.Error())
	}
	return nil
}

// syncIssuesToMain is the one sync dispatch, shared by `sdlc claim`, `sdlc issue
// new` (#82 M1), `sdlc issue sync` and `sdlc change-code` (#206). Extracted from
// runClaim so every caller broadcasts through the exact same machinery
// (ARCH-DRY) — the `--issue` filter on f narrows the sync to one file. The
// runner is threaded (not hard-wired to claimRunner) so callers and tests inject
// their own.
//
// msg is the commit subject; "" means "each arm's own default", which is how
// existing callers keep their historical messages verbatim (the on-branch
// default names the branch, which no caller could supply). `sdlc issue sync`
// passes a subject naming the issue.
//
// The NoPush arm comes FIRST, before the branch test. A caller that isn't
// publishing wants a durable local commit where the work is; the whole
// main-worktree route below exists to publish, so running it with the push
// removed would spend a worktree hunt and a network pull to land a commit on a
// branch the caller isn't on.
func syncIssuesToMain(stdout, stderr io.Writer, f *claimFlags, r gitRunner, msg string) error {
	if f.NoPush {
		return syncInPlace(stdout, stderr, f, r, msg)
	}
	branch := gitx.Capture("branch", "--show-current")
	if branch == "main" {
		return syncInPlace(stdout, stderr, f, r, msg)
	}
	return syncViaMainWorktree(stdout, stderr, f, branch, r, msg)
}

// startOnClaim folds the "start work" status flip into `sdlc claim`: an
// `--issue` claim on an *open* issue is the start-of-work gesture, so flip
// it to `working` before the sync broadcasts it to origin/main. Collapses the
// old two-step (`set-status … working` then `claim`) into one.
//
// (#113) Claim is a *cheap lock*: it demands no estimate. The estimate gate
// lives at `sdlc change-code`, so you claim early (the moment an idea
// crystallizes) and the claim commit's timestamp anchors the active-time
// window at engagement start — capturing design attention `sdlc actual` used
// to miss.
//
// Only the open→working transition is automatic. Claim doubles as the
// generic issue-file re-sync primitive, so an issue already in a
// deliberate state (working/blocked/punt/wontfix/done) is left untouched —
// claim never clobbers a status the operator set on purpose. `--no-start`
// suppresses the flip entirely.
func startOnClaim(stdout, stderr io.Writer, f *claimFlags) error {
	prev, err := issueStatus(f.IssuesDir, f.Issue)
	if err != nil {
		return err
	}
	if !vocab.Issue().IsOpen(prev) {
		return nil
	}
	// "working" is the claim target written here (a value-specific write, like
	// close's "done" write — not a category test), so it stays a literal (#122).
	path, _, _, err := applyStatus(f.IssuesDir, f.Issue, "working", false, f.DryRun)
	if err != nil {
		return err
	}
	if f.DryRun {
		cinfo(stderr, fmt.Sprintf("dry-run — would flip %s: status open → working", filepath.Base(path)))
		return nil
	}
	cok(stderr, fmt.Sprintf("%s: status open → working", filepath.Base(path)))
	return nil
}

// ── commit-here path ─────────────────────────────────────────────────────────

// syncInPlace stages + commits the changed issue files in the CURRENT worktree,
// on the current branch, and pushes origin/main unless f.NoPush. Named for what
// it does rather than where it runs (it was syncOnMain until #206): with NoPush
// it is the durable local commit `sdlc issue sync` wants from any branch, and
// with the push it is the on-main publish `claim` has always done.
//
// Returns an error rather than calling die() directly, so callers decide the
// severity: `claim` dies on it (its whole job is the sync), while `issue new`
// and `change-code` treat it as best-effort (the file is already written — a
// failed push must not abort creation or block entering implementation, e.g.
// offline or with no reachable origin).
func syncInPlace(stdout, stderr io.Writer, f *claimFlags, r gitRunner, msg string) error {
	changed, err := changedIssueFiles(f, r)
	if err != nil {
		return err
	}
	// Nothing in the working tree does NOT mean nothing to publish. Since #206
	// split durability from publication, "committed locally, not yet pushed" is a
	// state the code deliberately creates — a no-push sync, or a sync whose commit
	// landed and whose push failed. changedIssueFiles is empty in exactly that
	// state, so returning here would make `--push` a silent no-op and strand every
	// warning that names it as the recovery. Skip the COMMIT, never the publish.
	if len(changed) == 0 {
		if !f.PublishExisting {
			// The pre-#206 no-op, kept deliberately for every caller that did not
			// ask to publish already-committed work. `sdlc claim` is idempotent by
			// design and die()s on a sync error, so reaching for the network here
			// makes an offline re-run fatal where it has always exited 0.
			cok(stderr, "No issue changes to sync.")
			return nil
		}
		if f.DryRun {
			cinfo(stderr, "dry-run — no new issue changes; would publish existing commits")
			return nil
		}
		cinfo(stderr, "No new issue changes — publishing existing commits...")
		return pushMain(stdout, stderr, r)
	}
	cinfo(stderr, "Syncing issue changes...")
	for _, c := range changed {
		fmt.Fprintf(stderr, "  %s\n", c)
	}
	if f.DryRun {
		cinfo(stderr, "dry-run — no commit/push performed")
		return nil
	}
	pathspec, err := syncPathspec(f)
	if err != nil {
		return err
	}
	if out, err := r.Git(append([]string{"add", "--"}, pathspec...)...); err != nil {
		return fmt.Errorf("git add: %v\n%s", err, out)
	}
	// The commit carries the SAME pathspec as the add (#206). A bare `git
	// commit` records the whole index, so anything a peer agent had staged in
	// this checkout was swept into a commit that misdescribes it — the repo
	// transaction lock serializes sdlc verbs against each other, but nothing
	// stops a peer running plain `git add`. A pathspec implies --only, leaving
	// the rest of the index untouched.
	commitArgs := append([]string{"commit", "-m", syncMessage(msg, "issue-sync: update issues"), "--"}, pathspec...)
	if out, err := r.Git(commitArgs...); err != nil {
		return fmt.Errorf("commit failed: %v\n%s", err, out)
	}
	if f.NoPush {
		cok(stderr, "Issue changes committed locally (not pushed).")
		fmt.Fprintln(stdout, "synced")
		return nil
	}
	return pushMain(stdout, stderr, r)
}

// pushMain publishes the current worktree's main to origin. One source for both
// of syncInPlace's exits — the just-committed path and the nothing-new-to-commit
// path — so `--push` cannot mean "publish" in one and "no-op" in the other.
func pushMain(stdout, stderr io.Writer, r gitRunner) error {
	if out, err := r.Git("push", "origin", "main"); err != nil {
		return fmt.Errorf("push failed: %v\n%s", err, out)
	}
	cok(stderr, "Issues synced and pushed to origin/main.")
	fmt.Fprintln(stdout, "synced")
	return nil
}

// syncPathspec returns the paths a sync should stage AND commit: the whole
// issues dir, or just the --issue file when the sync is filtered. One source for
// both argv lists, so the add and the commit can't drift apart — a commit whose
// pathspec is wider than its add is exactly the bug #206 fixes.
//
// NOT pure despite reading like an argv builder: the --issue branch globs the
// working tree through issueFilesForID, so this is a thin IO helper and its
// filtered path needs a real directory to test against (ARCH-PURE).
//
// The "no file matches" error stands on its own rather than behind a claim that
// it is unreachable. An earlier version argued callers run changedIssueFiles
// first — which does not cover a DELETED issue file, where changedIssueFiles is
// non-empty (git reports the deletion) while the glob finds nothing.
func syncPathspec(f *claimFlags) ([]string, error) {
	if f.Issue <= 0 {
		return []string{f.IssuesDir + "/"}, nil
	}
	matches := issueFilesForID(f.IssuesDir, f.Issue)
	if len(matches) == 0 {
		return nil, fmt.Errorf("--issue %d: no file matches %s/%06d-*.md", f.Issue, f.IssuesDir, f.Issue)
	}
	return matches, nil
}

// syncMessage picks the commit subject: the caller's, or the arm's default when
// the caller passed none. Keeps the message a parameter of the helper rather
// than a branch inside it (#206) while letting each arm keep the exact wording
// it shipped with — the on-branch default names the branch, which no caller is
// in a position to supply.
func syncMessage(msg, fallback string) string {
	if msg == "" {
		return fallback
	}
	return msg
}

// ── publish-from-a-branch path ───────────────────────────────────────────────

// syncViaMainWorktree publishes the changed issue files to origin/main from a
// feature branch, by routing them through the worktree that has main checked
// out. Named for the route it takes (it was syncOnBranch until #206): every step
// — the worktree hunt, the cleanliness refusal, the network rebase, the copy —
// exists to publish, which is why a no-push caller takes syncInPlace instead.
//
// Mirrors syncInPlace's error contract: it returns errors instead of calling
// die() so `claim` (fatal) and `issue new` / `change-code` (best-effort) can
// choose.
func syncViaMainWorktree(stdout, stderr io.Writer, f *claimFlags, branch string, r gitRunner, msg string) error {
	changed, err := changedIssueFiles(f, r)
	if err != nil {
		return err
	}
	// Same rule as syncInPlace (#206): nothing to COPY is not nothing to publish.
	// The body may already be committed here — by a prior no-push sync, or by a
	// run whose push failed after the commit landed — and this arm's whole job is
	// routing the body to main, so it re-seeds the file list from the ISSUE
	// rather than from working-tree dirtiness and continues through the normal
	// copy → commit → push flow below.
	//
	// A first attempt at this pushed `origin main` from the main worktree without
	// carrying anything across, which is worse than the bug it fixed: main has no
	// new commits (the body is on the BRANCH), so it printed success while
	// origin/main never moved. Publication is the gap between origin/main and
	// this worktree's body, never the gap between the working tree and HEAD.
	if len(changed) == 0 {
		if !f.PublishExisting || f.Issue <= 0 {
			// Nobody asked to publish already-committed work, or there is no
			// --issue naming an identifiable body to route. Either way the
			// pre-#206 no-op is the honest answer.
			cok(stderr, "No issue changes to sync.")
			return nil
		}
		changed = issueFilesForID(f.IssuesDir, f.Issue)
		if len(changed) == 0 {
			cok(stderr, "No issue changes to sync.")
			return nil
		}
	}
	cinfo(stderr, fmt.Sprintf("Issue files changed on branch '%s':", branch))
	for _, c := range changed {
		fmt.Fprintf(stderr, "  %s\n", c)
	}

	// 1. Find the main worktree.
	mainPath, err := findMainWorktree(r)
	if err != nil {
		return err
	}

	// 2. Verify main worktree is on main.
	mainBranchOut, err := r.GitInDir(mainPath, "branch", "--show-current")
	if err != nil {
		return fmt.Errorf("git -C %s branch --show-current: %v\n%s", mainPath, err, mainBranchOut)
	}
	mainBranch := strings.TrimSpace(string(mainBranchOut))
	if mainBranch != "main" {
		return fmt.Errorf("expected main worktree to be on 'main', but it's on '%s'", mainBranch)
	}

	// 3. Check main worktree has no uncommitted issue changes.
	mainDirty, err := mainHasUncommittedIssueChanges(mainPath, f.IssuesDir, r)
	if err != nil {
		return err
	}
	if len(mainDirty) > 0 {
		return fmt.Errorf("main worktree has uncommitted issue changes. Commit or stash them first:\n  %s",
			strings.Join(mainDirty, "\n  "))
	}

	cok(stderr, fmt.Sprintf("Main worktree found at: %s", mainPath))

	if f.DryRun {
		cinfo(stderr, "dry-run — skipping pull/copy/commit/push")
		return nil
	}

	// 4. Pull --rebase origin main on main worktree.
	cinfo(stderr, "Pulling latest main from origin...")
	if out, err := r.GitInDir(mainPath, "pull", "--rebase", "origin", "main"); err != nil {
		return fmt.Errorf("failed to pull main from origin: %v\n%s", err, out)
	}

	// 4.5 Drop files whose main-worktree copy is ALREADY byte-identical to this
	//     one. Runs after the pull, so it compares against main's current state.
	//
	//     This is not conflict detection; it is declining to invoke it. The
	//     detector below asks "did both sides touch this file since merge-base",
	//     which is the right question for a genuinely-dirty file and the wrong
	//     one for a body that main already carries because an earlier run of THIS
	//     verb put it there. Without the filter, the recommended workflow — local
	//     sync, publish, publish again — dies on a false `Conflict detected!`,
	//     and `claim` stops being idempotent.
	//
	//     Nothing left to route means nothing to copy or commit; the publish
	//     below still runs, which is the whole reason a caller reached this arm.
	wtRoot, _ := gitx.RepoTopLevel()
	changed = filesDifferingFrom(mainPath, wtRoot, changed)
	if len(changed) == 0 {
		cinfo(stderr, "main worktree already carries this body — nothing to copy; publishing only.")
		if out, err := r.GitInDir(mainPath, "push", "origin", "main"); err != nil {
			return fmt.Errorf("push failed: %v\n%s", err, out)
		}
		cok(stderr, "Issues synced to main and pushed to origin.")
		fmt.Fprintln(stdout, "synced")
		return nil
	}

	// 5. Compute merge base and detect conflicts.
	mergeBase := strings.TrimSpace(string(mustGitOutput(r, "merge-base", "main", "HEAD")))
	if mergeBase == "" {
		return fmt.Errorf("cannot find merge base between main and HEAD")
	}
	mainChangedOut, _ := r.Git("diff", "--name-only", mergeBase, "main", "--", f.IssuesDir+"/")
	mainChanged := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(mainChangedOut)), "\n") {
		if line != "" {
			mainChanged[line] = true
		}
	}
	var conflicts []string
	for _, c := range changed {
		if mainChanged[c] {
			conflicts = append(conflicts, c)
		}
	}
	if len(conflicts) > 0 {
		// Print the multi-line resolution guide here (it's specific to this
		// state), then return a short sentinel — the caller decides fatal vs warn.
		fmt.Fprintf(stderr, "%sConflict detected!%s\n", ansiRed, ansiReset)
		fmt.Fprintln(stderr, "These issue files were changed on both your branch and main:")
		for _, c := range conflicts {
			fmt.Fprintf(stderr, "  %s\n", c)
		}
		fmt.Fprintf(stderr, "\nTo resolve:\n")
		fmt.Fprintf(stderr, "  1. cd %s\n", mainPath)
		fmt.Fprintf(stderr, "  2. For each file above, open it and manually merge your changes.\n")
		wtRoot, _ := gitx.RepoTopLevel()
		fmt.Fprintf(stderr, "     Your branch versions are at: %s\n", wtRoot)
		fmt.Fprintf(stderr, "  3. git add %s/\n", f.IssuesDir)
		fmt.Fprintf(stderr, "  4. git commit -m \"issue-sync: resolve conflicts\"\n")
		fmt.Fprintf(stderr, "  5. git push origin main\n")
		return fmt.Errorf("issue-sync conflict on %d file(s) — resolve as shown above", len(conflicts))
	}

	cok(stderr, "No conflicts detected.")

	// 6. Copy changed files to main worktree.
	cinfo(stderr, "Copying issue files to main worktree...")
	for _, c := range changed {
		src := filepath.Join(wtRoot, c)
		dest := filepath.Join(mainPath, c)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %v", filepath.Dir(dest), err)
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read %s: %v", src, err)
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %v", dest, err)
		}
		fmt.Fprintf(stderr, "  %s\n", c)
	}

	// 7. Commit + push on main worktree. The pathspec is the set of files we
	//    just copied across, and it goes on the COMMIT as well as the add
	//    (#206) — the main worktree's index can hold a peer agent's staged work
	//    that a bare commit would sweep, and the mainHasUncommittedIssueChanges
	//    precheck above only looks at issue files.
	cinfo(stderr, "Committing and pushing on main...")
	addArgs := append([]string{"add", "--"}, changed...)
	if out, err := r.GitInDir(mainPath, addArgs...); err != nil {
		return fmt.Errorf("git -C %s add: %v\n%s", mainPath, err, out)
	}
	// Commit only if the copy actually changed something. Re-seeding `changed`
	// from the issue means the files may be byte-identical to main's copy (a
	// re-run, or a sync whose commit landed and whose push failed) — and `git
	// commit` with a pathspec that stages nothing is an error, not a no-op. The
	// push below still runs: publishing an already-committed body is the whole
	// point of reaching here.
	stagedArgs := append([]string{"diff", "--cached", "--name-only", "--"}, changed...)
	staged, err := r.GitInDir(mainPath, stagedArgs...)
	if err != nil {
		return fmt.Errorf("git -C %s diff --cached: %v\n%s", mainPath, err, staged)
	}
	if strings.TrimSpace(string(staged)) == "" {
		cinfo(stderr, "the copy staged nothing — main was already current; publishing only.")
	} else {
		commitMsg := syncMessage(msg, fmt.Sprintf("issue-sync: update issues from branch '%s'", branch))
		commitArgs := append([]string{"commit", "-m", commitMsg, "--"}, changed...)
		if out, err := r.GitInDir(mainPath, commitArgs...); err != nil {
			return fmt.Errorf("commit failed: %v\n%s", err, out)
		}
	}
	if out, err := r.GitInDir(mainPath, "push", "origin", "main"); err != nil {
		return fmt.Errorf("push failed: %v\n%s", err, out)
	}
	cok(stderr, "Issues synced to main and pushed to origin.")
	fmt.Fprintln(stdout, "synced")
	return nil
}

// filesDifferingFrom returns the subset of paths whose content in srcRoot
// differs from destRoot's copy — the files a publish actually has to move. A
// missing destination counts as different (it needs the file); an unreadable
// source counts as different too, so the copy step reports the real IO error
// rather than this filter swallowing it.
//
// Pure-ish by design: reads only, no git, so the caller keeps the IO ordering
// decision (it must run after the pull) and this stays trivially testable.
func filesDifferingFrom(destRoot, srcRoot string, paths []string) []string {
	var out []string
	for _, p := range paths {
		src, err := os.ReadFile(filepath.Join(srcRoot, p))
		if err != nil {
			out = append(out, p)
			continue
		}
		dst, err := os.ReadFile(filepath.Join(destRoot, p))
		if err != nil || !bytes.Equal(src, dst) {
			out = append(out, p)
		}
	}
	return out
}

// ── helpers ──────────────────────────────────────────────────────────────────

// changedIssueFiles returns the union of:
//   - `git diff --name-only HEAD -- <issuesDir>/`   (working-tree + staged
//     relative to HEAD)
//   - `git diff --cached --name-only -- <issuesDir>/`  (staged-only)
//   - `git ls-files --others --exclude-standard -- <issuesDir>/`  (untracked)
//
// Sorted + deduped. If f.Issue is set, filter to only the matching
// NNNNNN-*.md file.
//
// The changed-file union includes
// "diff HEAD" (which already covers cached) plus "diff --cached" separately
// (redundant but preserved for parity); de-dup happens at the sort step.
func changedIssueFiles(f *claimFlags, r gitRunner) ([]string, error) {
	queries := [][]string{
		{"diff", "--name-only", "HEAD", "--", f.IssuesDir + "/"},
		{"diff", "--cached", "--name-only", "--", f.IssuesDir + "/"},
		{"ls-files", "--others", "--exclude-standard", "--", f.IssuesDir + "/"},
	}
	seen := map[string]struct{}{}
	var out []string
	for _, q := range queries {
		raw, err := r.Git(q...)
		if err != nil {
			// Mirror the shell `|| true` swallow: empty result, no error.
			continue
		}
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			if line == "" {
				continue
			}
			if _, ok := seen[line]; ok {
				continue
			}
			seen[line] = struct{}{}
			out = append(out, line)
		}
	}
	sort.Strings(out)

	if f.Issue > 0 {
		id := fmt.Sprintf("%06d", f.Issue)
		var filtered []string
		for _, p := range out {
			if strings.HasPrefix(filepath.Base(p), id+"-") {
				filtered = append(filtered, p)
			}
		}
		out = filtered
	}
	return out, nil
}

// findMainWorktree parses `git worktree list --porcelain -z` and returns
// the path of the worktree on branch `main`. Empty + error if none.
func findMainWorktree(r gitRunner) (string, error) {
	out, err := r.Git("worktree", "list", "--porcelain", "-z")
	if err != nil {
		return "", fmt.Errorf("git worktree list: %v\n%s", err, out)
	}
	// Reuse the single-source porcelain parser (ARCH-DRY, #200) rather than
	// re-walking the grammar. The IO (r.Git) stays here; the parse is pure.
	worktrees, err := gitx.ParseWorktrees(out)
	if err != nil {
		return "", fmt.Errorf("parse git worktree list: %w", err)
	}
	if mainPath, ok := worktreeForBranch(worktrees, "main"); ok {
		return mainPath, nil
	}
	return "", fmt.Errorf("could not find a worktree on branch 'main'. Is main checked out somewhere?")
}

// mainHasUncommittedIssueChanges returns the list of issue files in the
// main worktree that have uncommitted changes (working + staged).
func mainHasUncommittedIssueChanges(mainPath, issuesDir string, r gitRunner) ([]string, error) {
	dirty := map[string]struct{}{}
	for _, q := range [][]string{
		{"diff", "--name-only", "--", issuesDir + "/"},
		{"diff", "--cached", "--name-only", "--", issuesDir + "/"},
	} {
		raw, err := r.GitInDir(mainPath, q...)
		if err != nil {
			// A failed diff is a NON-ANSWER, not a clean worktree (#213 BR-26).
			// Swallowing it reported the main worktree clean without having
			// looked, and the caller then committed over whatever was there —
			// the same "blind read reported as an empty result" that every
			// instance in this issue reduces to.
			return nil, fmt.Errorf("read issue changes in %s: %v\n%s", mainPath, err, strings.TrimSpace(string(raw)))
		}
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			if line != "" {
				dirty[line] = struct{}{}
			}
		}
	}
	var out []string
	for k := range dirty {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}

// mustGitOutput is a thin shim that returns r.Git's stdout but discards
// errors (the shell uses `|| die` for these — we let the empty result
// trigger our own die() upstream).
func mustGitOutput(r gitRunner, args ...string) []byte {
	out, _ := r.Git(args...)
	return out
}
