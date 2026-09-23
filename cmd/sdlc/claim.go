// claim.go — remote issue reservation and narrow local synchronization.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// claimFlags holds the parsed flag values for the claim subcommand.
type claimFlags struct {
	Issue     int
	IssuesDir string
	DryRun    bool
	NoStart   bool
	// FirstPublication says this id has never been on the trunk, so re-allocating
	// it is cheap. Only `issue new` sets it. Renumbering is safe ONLY before
	// anything references the id; by claim time it is in the branch name, and
	// after that in commit subjects, deps: and sidecar filenames (ariadne#188).
	// The caller declares this rather than the publisher inferring it, because
	// the inference is exactly what an earlier draft of #207 got wrong — it would
	// have renumbered every existing issue on every sync.
	FirstPublication bool
	// HistoryDir is the archived-issue directory. Carried here because the id
	// space is issues + history: the trunk arm named a LITERAL "workshop/history"
	// and so could not see archived ids under WF_HISTORY_DIR, re-allocating onto
	// one and publishing beside another without refusing (#207 BR-17).
	HistoryDir string
	// Reallocations is an OUTPUT field: the publisher records any id changes it
	// made, so `issue new` prints the path it actually published rather than the
	// one it planned (#207 BR-1).
	Reallocations []*reallocation
	// NoPush suppresses publication: commit in the current worktree and stop
	// (#206). Spelled negatively on purpose — `issue new` builds this struct as
	// a literal (issue.go), so a positive `Push` would zero-value to false
	// there and silently kill the reservation broadcast #82 M1 added. The zero
	// value is today's behavior for every existing caller.
	NoPush bool
	// AllowNoChanges lets automatic callers leave previously authored commits local.
	AllowNoChanges bool
}

// NewClaimCmd returns the cobra command for `sdlc claim`.
func NewClaimCmd() *cobra.Command {
	f := claimFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:           "claim",
		Short:         "Reserve an open issue on fresh origin/main",
		Long:          "Placeholder — replaced by helptext.MustGet(\"lock\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			guardSpineRepo(cmd.ErrOrStderr()) // #176 lifecycle guard
			return runClaim(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "issue ID to reserve (required)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print what would happen; do not commit/push")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", envOr("WF_HISTORY_DIR", "workshop/history"), "directory holding archived issues")
	cmd.Flags().BoolVar(&f.NoStart, "no-start", false, "retired; use issue publish --commit SHA for documentation")
	return cmd
}

// claimRunner is the same gitRunner interface used by start; reused so
// tests can inject capture runners across both verbs.
var claimRunner gitRunner = execGitRunner{}

// runClaim reserves from remote bytes before reconciling local metadata.
func runClaim(stdout, stderr io.Writer, f *claimFlags) error {
	if f.Issue <= 0 || f.NoStart {
		return fmt.Errorf("claim requires --issue N and an open remote issue; publish documentation with sdlc issue publish --commit SHA")
	}
	paths, err := resolveSyncPaths(f)
	if err != nil {
		return err
	}
	pub, err := newTrunkPublisher(paths.Root)
	if err != nil {
		return err
	}
	var localPath string
	var before []byte
	matches := issueFilesForID(paths.Root, f.IssuesDir, f.Issue)
	if len(matches) > 1 {
		return fmt.Errorf("multiple local issue files for #%d", f.Issue)
	}
	if len(matches) == 1 {
		localPath = matches[0]
		info, err := os.Lstat(localPath)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("local issue is not an ordinary file: %s", localPath)
		}
		before, err = os.ReadFile(localPath)
		if err != nil {
			return err
		}
	}
	var claimed []byte
	dry := errors.New("claim dry run")
	err = pub.UpdateMany(fmt.Sprintf("#%d: issue-sync: claim", f.Issue), func(v *gitx.TrunkView) (gitx.TrunkWrite, error) {
		space, err := refIDSpace(v.Ref(), paths.Dirs, claimRunner)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		remotePaths := space[f.Issue]
		if len(remotePaths) != 1 {
			return gitx.TrunkWrite{}, fmt.Errorf("origin/main must contain exactly one issue #%d; found %d", f.Issue, len(remotePaths))
		}
		if localPath != "" && filepath.Base(localPath) != filepath.Base(remotePaths[0]) {
			return gitx.TrunkWrite{}, fmt.Errorf("issue #%d has different local and remote names (%s, %s); reconcile identity before claiming", f.Issue, filepath.Base(localPath), remotePaths[0])
		}
		raw, err := v.Read(remotePaths[0])
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		claimed, err = claimDecision(raw, f.Issue, time.Now().Format("2006-01-02"), startedClock())
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		if f.DryRun {
			return gitx.TrunkWrite{}, dry
		}
		return gitx.TrunkWrite{Write: map[string][]byte{remotePaths[0]: claimed}}, nil
	})
	if errors.Is(err, dry) {
		cinfo(stderr, "dry-run — remote issue is open; would reserve it")
		return nil
	}
	if err != nil {
		return err
	}
	if localPath != "" {
		current, readErr := os.ReadFile(localPath)
		if readErr != nil || !bytes.Equal(current, before) {
			cwarn(stderr, "claim published; local issue changed meanwhile — reconcile status from origin/main")
		} else {
			fm, body, parseErr := issue.Parse(string(before))
			remoteFM, _, _ := issue.Parse(string(claimed))
			if parseErr != nil {
				cwarn(stderr, "claim published; local issue could not be parsed — reconcile status from origin/main")
			} else {
				for _, field := range []string{"status", "updated", "started"} {
					value, ok := issue.GetField(remoteFM, field)
					if ok {
						fm = issue.SetField(fm, field, value)
					}
				}
				if err := os.WriteFile(localPath, []byte(issue.Compose(fm, body)), 0644); err != nil {
					return fmt.Errorf("claim published, but local reconciliation failed: %w", err)
				}
			}
		}
	}
	cok(stderr, fmt.Sprintf("Issue #%d reserved on origin/main (open → working).", f.Issue))
	fmt.Fprintln(stdout, "claimed")
	return nil
}

// syncIssuesToMain commits locally, reserves a newly created issue, or publishes
// the exact narrow commit it just authored. No checkout gets a branch push.
func syncIssuesToMain(stdout, stderr io.Writer, f *claimFlags, r gitRunner, msg string) error {
	paths, err := resolveSyncPaths(f)
	if err != nil {
		return err
	}
	if f.NoPush {
		return syncInPlace(stdout, stderr, f, r, msg, paths)
	}
	if f.FirstPublication {
		if f.Issue <= 0 {
			return fmt.Errorf("new issue reservation requires exactly one issue ID")
		}
		pub, err := newTrunkPublisher(paths.Root)
		if err != nil {
			return err
		}
		return syncViaTrunk(stdout, stderr, f, r, msg, pub, paths)
	}
	changed, err := changedIssueFiles(f, r, paths)
	if err != nil {
		return err
	}
	if len(changed) == 0 {
		if f.AllowNoChanges || f.DryRun {
			cok(stderr, "No new issue commit to publish.")
			return nil
		}
		return fmt.Errorf("no new issue commit; select the intended commit with sdlc issue publish --commit SHA")
	}
	local := *f
	local.NoPush = true
	if err := syncInPlace(io.Discard, stderr, &local, r, msg, paths); err != nil {
		return err
	}
	if f.DryRun {
		return nil
	}
	raw, err := r.GitInDir(paths.Root, "rev-parse", "--verify", "HEAD")
	if err != nil {
		return fmt.Errorf("resolve new issue commit: %w", err)
	}
	source := strings.TrimSpace(string(raw))
	if err := publishIssueCommit(stdout, stderr, paths.Root, source, f.IssuesDir, f.historyDir()); err != nil {
		return fmt.Errorf("issue commit %s saved locally; publication failed: %w; retry with sdlc issue publish --commit %s", source, err, source)
	}
	return nil
}

// syncPaths is the ONE resolution of where this repo keeps ids and where its
// git calls must run. Threaded rather than re-derived: execGitRunner.Git runs in
// the PROCESS cwd, so an unpinned call from a subdirectory reads a different
// tree than the caller meant — silently, and reporting success (#207 BR-16).
type syncPaths struct {
	Root string
	Dirs idDirs
}

func resolveSyncPaths(f *claimFlags) (syncPaths, error) {
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return syncPaths{}, err
	}
	dirs, err := resolveIDDirs(f.IssuesDir, f.historyDir())
	if err != nil {
		return syncPaths{}, err
	}
	return syncPaths{Root: root, Dirs: dirs}, nil
}

// historyDir honours the configured value, falling back to the same env/default
// pair every other verb uses rather than to a literal.
func (f *claimFlags) historyDir() string {
	if f.HistoryDir != "" {
		return f.HistoryDir
	}
	return envOr("WF_HISTORY_DIR", "workshop/history")
}

// ── commit-here path ─────────────────────────────────────────────────────────

// syncInPlace authors one local issue commit, preserving foreign staged paths.
func syncInPlace(stdout, stderr io.Writer, f *claimFlags, r gitRunner, msg string, paths syncPaths) error {
	changed, err := changedIssueFiles(f, r, paths)
	if err != nil {
		return err
	}
	if len(changed) == 0 {
		cok(stderr, "No issue changes to sync.")
		return nil
	}
	cinfo(stderr, "Syncing issue changes...")
	for _, c := range changed {
		fmt.Fprintf(stderr, "  %s\n", c)
	}
	if f.DryRun {
		cinfo(stderr, "dry-run — no commit/push performed")
		return nil
	}
	pathspec, err := syncPathspec(f, paths)
	if err != nil {
		return err
	}
	if out, err := r.GitInDir(paths.Root, append([]string{"add", "--"}, pathspec...)...); err != nil {
		return fmt.Errorf("git add: %v\n%s", err, out)
	}
	// The commit carries the SAME pathspec as the add (#206). A bare `git
	// commit` records the whole index, so anything a peer agent had staged in
	// this checkout was swept into a commit that misdescribes it — the repo
	// transaction lock serializes sdlc verbs against each other, but nothing
	// stops a peer running plain `git add`. A pathspec implies --only, leaving
	// the rest of the index untouched.
	commitArgs := append([]string{"commit", "-m", syncMessage(msg, defaultSyncSubject), "--"}, pathspec...)
	if out, err := r.GitInDir(paths.Root, commitArgs...); err != nil {
		return fmt.Errorf("commit failed: %v\n%s", err, out)
	}
	cok(stderr, "Issue changes committed locally (not pushed).")
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
func syncPathspec(f *claimFlags, paths syncPaths) ([]string, error) {
	if f.Issue <= 0 {
		return []string{f.IssuesDir + "/"}, nil
	}
	matches := issueFilesForID(paths.Root, f.IssuesDir, f.Issue)
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
// defaultSyncSubject is the commit subject when a caller passes "" — "each
// arm's own default" in the msg contract. ONE constant for both arms: the
// in-place arm held it as a literal and the trunk arm had none, which is how
// the trunk arm came to publish an EMPTY subject (#207 BR-1).
const defaultSyncSubject = "issue-sync: update issues"

func syncMessage(msg, fallback string) string {
	if msg == "" {
		return fallback
	}
	return msg
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
func changedIssueFiles(f *claimFlags, r gitRunner, paths syncPaths) ([]string, error) {
	// -z on every query. Without it git QUOTES any path outside ASCII
	// (`"workshop/issues/000300-caf\303\251.md"`), and the quoted form does not
	// exist on disk: the read failed as not-exist, the publisher classified it
	// as a DELETION, and the arm reported success having published nothing
	// (#207 BR-9).
	// Pinned to the repo ROOT. execGitRunner.Git runs in the process cwd
	// (runner.go), and a pathspec — like the ls-tree pathspec in #207 BR-12 —
	// resolves against it: from any subdirectory these queries matched NOTHING,
	// so `issue new`, `claim` and `issue sync` reported "no issue changes" and
	// exited 0 having published nothing (#207 BR-16). ls-files also returns
	// cwd-relative paths, which the publisher then joined onto the root.
	// --no-renames on every diff. Rename detection answers "what CHANGED",
	// which is the wrong question here — a `git mv` was reported as the new
	// path alone, so the publish carried Write(new) with no Delete(old) and
	// left the old name published forever, a real duplicate id on the trunk.
	// We need the PATHS, not git's account of the edit (#207 BR-19).
	queries := [][]string{
		{"diff", "--name-only", "--no-renames", "-z", "HEAD", "--", f.IssuesDir + "/"},
		{"diff", "--cached", "--name-only", "--no-renames", "-z", "--", f.IssuesDir + "/"},
		{"ls-files", "-z", "--others", "--exclude-standard", "--", f.IssuesDir + "/"},
	}
	seen := map[string]struct{}{}
	var out []string
	for _, q := range queries {
		raw, err := r.GitInDir(paths.Root, q...)
		if err != nil {
			// Mirror the shell `|| true` swallow: empty result, no error.
			continue
		}
		for _, line := range strings.Split(string(raw), "\x00") {
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
