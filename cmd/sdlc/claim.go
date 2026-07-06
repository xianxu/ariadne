// claim.go — `sdlc claim [--issue N]` subcommand.
//
// Ports scripts/issue-sync.sh — the issue-file synchronizer that
// commits + pushes workshop/issues/ changes to origin/main even when
// the operator is on a feature branch. Used as the workstream claim
// primitive: agents claim work by flipping status to `working` and
// running `sdlc claim` to broadcast that claim to origin/main.
//
// Two paths in the source script (preserved verbatim here):
//
//  1. On main:    add + commit + push directly.
//  2. On a feature branch:
//     - locate the main worktree via `git worktree list --porcelain`
//     - check main worktree has no uncommitted issue changes
//     - pull --rebase origin main on the main worktree
//     - detect conflicts (files changed on both branches since merge-base)
//     - copy changed issue files from feature worktree → main worktree
//     - commit + push on main worktree
//
// The shell script supports no flags. We add --issue (filter the sync to
// one issue file), --issues-dir (env override), --dry-run.
package main

import (
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
	if err := syncIssuesToMain(stdout, stderr, f, claimRunner); err != nil {
		die(stderr, err.Error())
	}
	return nil
}

// syncIssuesToMain is the branch-aware sync dispatch shared by `sdlc claim` and
// `sdlc issue new` (#82 M1): on main it commits + pushes the changed issue files
// directly; on a feature branch it routes them to the main worktree. Extracted
// from runClaim so issue creation can broadcast a freshly-scaffolded file to
// origin/main through the exact same machinery (ARCH-DRY) — the `--issue` filter
// on f narrows the sync to that one file. The runner is threaded (not hard-wired
// to claimRunner) so callers and tests inject their own.
func syncIssuesToMain(stdout, stderr io.Writer, f *claimFlags, r gitRunner) error {
	branch := gitx.Capture("branch", "--show-current")
	if branch == "main" {
		return syncOnMain(stdout, stderr, f, r)
	}
	return syncOnBranch(stdout, stderr, f, branch, r)
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

// ── on-main path ─────────────────────────────────────────────────────────────

// syncOnMain returns an error rather than calling die() directly, so callers
// decide the severity: `claim` dies on it (its whole job is the sync), while
// `issue new` treats it as best-effort (the file is already written — a failed
// push must not abort creation, e.g. offline or with no reachable origin).
func syncOnMain(stdout, stderr io.Writer, f *claimFlags, r gitRunner) error {
	changed, err := changedIssueFiles(f, r)
	if err != nil {
		return err
	}
	if len(changed) == 0 {
		cok(stderr, "No issue changes to sync.")
		return nil
	}
	cinfo(stderr, "Syncing issue changes on main...")
	for _, c := range changed {
		fmt.Fprintf(stderr, "  %s\n", c)
	}
	if f.DryRun {
		cinfo(stderr, "dry-run — no commit/push performed")
		return nil
	}
	// Match the shell exactly: git add issues/ (when no --issue filter).
	// With --issue, narrow to that issue's NNNNNN-*.md files.
	addArgs := []string{"add", f.IssuesDir + "/"}
	if f.Issue > 0 {
		id := fmt.Sprintf("%06d", f.Issue)
		matches, _ := filepath.Glob(filepath.Join(f.IssuesDir, id+"-*.md"))
		if len(matches) == 0 {
			return fmt.Errorf("--issue %d: no file matches %s/%s-*.md", f.Issue, f.IssuesDir, id)
		}
		addArgs = append([]string{"add"}, matches...)
	}
	if out, err := r.Git(addArgs...); err != nil {
		return fmt.Errorf("git add: %v\n%s", err, out)
	}
	if out, err := r.Git("commit", "-m", "issue-sync: update issues"); err != nil {
		return fmt.Errorf("commit failed: %v\n%s", err, out)
	}
	if out, err := r.Git("push", "origin", "main"); err != nil {
		return fmt.Errorf("push failed: %v\n%s", err, out)
	}
	cok(stderr, "Issues synced and pushed to origin/main.")
	fmt.Fprintln(stdout, "synced")
	return nil
}

// ── on-branch path ───────────────────────────────────────────────────────────

// syncOnBranch mirrors syncOnMain's error contract: it returns errors instead
// of calling die() so `claim` (fatal) and `issue new` (best-effort) can choose.
func syncOnBranch(stdout, stderr io.Writer, f *claimFlags, branch string, r gitRunner) error {
	changed, err := changedIssueFiles(f, r)
	if err != nil {
		return err
	}
	if len(changed) == 0 {
		cok(stderr, "No issue changes to sync.")
		return nil
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
	wtRoot, _ := gitx.RepoTopLevel()
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

	// 7. Commit + push on main worktree.
	cinfo(stderr, "Committing and pushing on main...")
	if out, err := r.GitInDir(mainPath, "add", f.IssuesDir+"/"); err != nil {
		return fmt.Errorf("git -C %s add: %v\n%s", mainPath, err, out)
	}
	commitMsg := fmt.Sprintf("issue-sync: update issues from branch '%s'", branch)
	if out, err := r.GitInDir(mainPath, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("commit failed: %v\n%s", err, out)
	}
	if out, err := r.GitInDir(mainPath, "push", "origin", "main"); err != nil {
		return fmt.Errorf("push failed: %v\n%s", err, out)
	}
	cok(stderr, "Issues synced to main and pushed to origin.")
	fmt.Fprintln(stdout, "synced")
	return nil
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
// Matches issue-sync.sh's changed_issue_files() — note the union includes
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

// findMainWorktree parses `git worktree list --porcelain` and returns
// the path of the worktree on branch `main`. Empty + error if none.
//
// Matches the awk pipeline in issue-sync.sh:
//
//	awk '/^worktree /{path=$2} /branch refs\/heads\/main$/{print path}'
func findMainWorktree(r gitRunner) (string, error) {
	out, err := r.Git("worktree", "list", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("git worktree list: %v\n%s", err, out)
	}
	// Reuse the single-source porcelain parser (ARCH-DRY, #156) rather than
	// re-walking the grammar. The IO (r.Git) stays here; the parse is pure.
	if mainPath, ok := worktreeForBranch(string(out), "main"); ok {
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
			continue // mirror shell `|| true`
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
