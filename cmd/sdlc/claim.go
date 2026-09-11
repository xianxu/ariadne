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
//  2. syncViaTrunk (synctrunk.go): publish from anywhere else (#207). Builds
//     the commit in the object database and CAS-pushes it at origin/main; no
//     checkout is read or written, so it works from an in-place feature branch
//     with no worktree on main — which is change-code's default.
//
// Every step of (2) exists to publish, which is why suppressing the push
// doesn't just skip the last line — it selects the other arm entirely.
//
// The command supports --issue (filter the sync to one issue file),
// --issues-dir (env override), and --dry-run.
package main

import (
	"fmt"
	"io"
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
	// FirstPublication says this id has never been on the trunk, so re-allocating
	// it is cheap. Only `issue new` sets it. Renumbering is safe ONLY before
	// anything references the id; by claim time it is in the branch name, and
	// after that in commit subjects, deps: and sidecar filenames (ariadne#188).
	// The caller declares this rather than the publisher inferring it, because
	// the inference is exactly what an earlier draft of #207 got wrong — it would
	// have renumbered every existing issue on every sync.
	FirstPublication bool
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
	// #207: publish straight to the trunk. The route this replaced drove whatever
	// checkout had main out, which is unavailable exactly in the workflow's
	// default mode — change-code branches in place, so an actively-worked repo
	// has no worktree on main.
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	pub, err := newTrunkPublisher(root)
	if err != nil {
		return err
	}
	return syncViaTrunk(stdout, stderr, f, r, msg, pub)
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
	commitArgs := append([]string{"commit", "-m", syncMessage(msg, defaultSyncSubject), "--"}, pathspec...)
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

// ── publish-from-a-branch path ───────────────────────────────────────────────

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
	// -z on every query. Without it git QUOTES any path outside ASCII
	// (`"workshop/issues/000300-caf\303\251.md"`), and the quoted form does not
	// exist on disk: the read failed as not-exist, the publisher classified it
	// as a DELETION, and the arm reported success having published nothing
	// (#207 BR-9).
	queries := [][]string{
		{"diff", "--name-only", "-z", "HEAD", "--", f.IssuesDir + "/"},
		{"diff", "--cached", "--name-only", "-z", "--", f.IssuesDir + "/"},
		{"ls-files", "-z", "--others", "--exclude-standard", "--", f.IssuesDir + "/"},
	}
	seen := map[string]struct{}{}
	var out []string
	for _, q := range queries {
		raw, err := r.Git(q...)
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

// mustGitOutput is a thin shim that returns r.Git's stdout but discards
// errors (the shell uses `|| die` for these — we let the empty result
// trigger our own die() upstream).
func mustGitOutput(r gitRunner, args ...string) []byte {
	out, _ := r.Git(args...)
	return out
}
