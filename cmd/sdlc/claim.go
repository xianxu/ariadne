// claim.go — remote issue reservation and narrow local synchronization.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// claimFlags holds the parsed flag values for the claim subcommand.
type claimFlags struct {
	Issue     int
	IssuesDir string
	DryRun    bool
	NoStart   bool
	// Adopt records this workspace as the owner of a working (or blocked) card
	// that has none — one claimed before ownership existed (#277). It never
	// takes an owned card: reassignment is operator-directed `sdlc reclaim` (#278).
	Adopt bool
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
	// Issues is claim's --issue list (#284): the set claimed in one tracker
	// commit. Issue stays the single ID that `issue new` and sync build with.
	Issues []int
}

// NewClaimCmd returns the cobra command for `sdlc claim`.
func NewClaimCmd() *cobra.Command {
	f := claimFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:           "claim",
		Short:         "Claim an issue: record this workspace as its owner (the lock)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"lock\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			guardSpineRepo(cmd.ErrOrStderr()) // #176 lifecycle guard
			return runClaim(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().IntSliceVar(&f.Issues, "issue", nil, "issue ID(s) to claim, one tracker commit for the set (required): --issue 284 or --issue 284,285")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print what would happen; do not commit/push")
	cmd.Flags().BoolVar(&f.Adopt, "adopt", false, "record this workspace as owner of a working issue that has none (claimed before #277)")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", envOr("WF_HISTORY_DIR", "workshop/history"), "directory holding archived issues")
	cmd.Flags().BoolVar(&f.NoStart, "no-start", false, "retired; use issue publish --commit SHA for documentation")
	_ = cmd.Flags().MarkHidden("no-start") // retired: refused, kept only to explain itself
	return cmd
}

// claimRunner is the same gitRunner interface used by start; reused so
// tests can inject capture runners across both verbs.
var claimRunner gitRunner = execGitRunner{}

// runClaim reserves an open card on the tracker (#252). An issue is claimable
// only once its creation is complete — its details have landed on main — so a
// card-only or local-only issue can never be worked by a second thread.
func runClaim(ctx context.Context, stdout, stderr io.Writer, f *claimFlags) error {
	nums := claimIssues(f)
	if tracked, err := repositoryTracked(ctx, f.IssuesDir); err != nil {
		return err
	} else if !tracked {
		if len(nums) > 1 {
			return fmt.Errorf("a legacy repository claims one issue at a time")
		}
		if len(nums) == 1 {
			f.Issue = nums[0]
		}
		return runLegacyClaim(stdout, stderr, f)
	}
	if len(nums) == 0 || f.NoStart {
		return fmt.Errorf("claim requires --issue N (or a list: --issue 284,285) and an open issue card")
	}
	dirs, err := resolveIDDirs(f.IssuesDir, f.HistoryDir)
	if err != nil {
		return err
	}
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	ids := make([]string, len(nums))
	cards := map[string]tracker.Record{}
	detailPaths := map[string]string{}
	for i, n := range nums {
		id := fmt.Sprintf("%06d", n)
		card, err := snap.Require(id)
		if errors.Is(err, tracker.ErrNoCard) {
			return fmt.Errorf("no card #%s on %s; file issues with `sdlc issue new`", id, vocab.Issue().Discovery().Tracker)
		} else if err != nil {
			return err
		}
		ids[i], cards[id], detailPaths[id] = id, card, path.Join(dirs.Rel[0], path.Base(card.Path))
	}
	// Readiness — every issue's details on main — is checked now and again
	// against fresh main after the candidate is pinned, before the only mutation.
	ready := func() error {
		view, err := env.main.Snapshot()
		if err != nil {
			return err
		}
		for _, id := range ids {
			present, err := view.Exists(detailPaths[id])
			if err != nil {
				return err
			}
			if !present {
				return fmt.Errorf("#%s is not claimable yet: its details (%s) have not landed on main, so its creation is incomplete.\n"+
					"      The creator finishes it with `sdlc issue move-detail --issue %s` (or by merging the branch that filed it)", id, detailPaths[id], issue.CLIRef(id))
			}
		}
		return nil
	}
	if err := ready(); err != nil {
		return err
	}
	me, err := claimantIdentity(env)
	if err != nil {
		return err
	}
	if len(ids) == 1 {
		id := ids[0]
		if f.Adopt {
			return adoptClaim(stdout, stderr, env, cards[id], me, detailPaths[id], f.DryRun)
		}
		if done, err := finishRelocation(stdout, stderr, env, cards[id], me, detailPaths[id], f.DryRun); done || err != nil {
			return err
		}
	} else {
		if f.Adopt {
			return fmt.Errorf("--adopt takes one issue")
		}
		// Finishing an `sdlc move` reads this machine's worktrees: one issue alone.
		for _, id := range ids {
			if rec, err := readRelocation(env.root, id); err != nil {
				return err
			} else if rec != nil {
				return fmt.Errorf("#%s has an unfinished `sdlc move`; claim it alone (`sdlc claim --issue %s`) to finish it", id, issue.CLIRef(id))
			}
		}
	}
	today, started := time.Now().Format("2006-01-02"), startedClock()
	decide := func(current map[string]tracker.Record) (map[string][]byte, error) {
		return claimSetDecision(current, ids, today, started, me)
	}
	refresh := func() {
		refreshAfterClaim(env, stderr, ids, detailPaths)
		for _, id := range ids {
			if warn := refreshLocalMirror(env, detailPaths[id]); warn != "" {
				cwarn(stderr, warn)
			}
		}
	}
	if _, err := decide(cards); errors.Is(err, errAlreadyMine) {
		cok(stderr, fmt.Sprintf("%s already claimed by this workspace; nothing to do", claimRefs(ids)))
		if !f.DryRun {
			refresh()
			fmt.Fprintln(stdout, "claimed")
		}
		return nil
	} else if err != nil {
		return err
	}
	if f.DryRun {
		cinfo(stderr, fmt.Sprintf("dry-run — %s claimable (open and unowned, details on main); would claim in one tracker commit", claimRefs(ids)))
		return nil
	}
	err = cardsPublish(env, ids, operationToken("claim"), nil, decide, func(base, candidate string) error { return ready() })
	invalidateIssueRecords(env.ctx)
	if errors.Is(err, errAlreadyMine) {
		err = nil // every card became this workspace's while retrying: settled
	}
	if err = uncertainCardWrite(err, "sdlc claim --issue "+claimArg(ids)); err != nil {
		return err
	}
	for _, id := range ids {
		cok(stderr, fmt.Sprintf("Issue #%s claimed on %s: this workspace owns it; status stays open until `sdlc start-plan --issue %s` starts it.", id, vocab.Issue().Discovery().Tracker, issue.CLIRef(id)))
	}
	refresh()
	fmt.Fprintln(stdout, "claimed")
	return nil
}

// refreshAfterClaim brings this checkout up to what it now holds (#284), so the
// lock never protects a stale copy: a resting branch fast-forwards to the
// fetched main; elsewhere, details whose body differs from main's are named
// (the frontmatter carries the card mirror, which differs by design). It
// warns and never fails: the claim has already landed.
func refreshAfterClaim(env *trackerEnv, stderr io.Writer, ids []string, detailPaths map[string]string) {
	view, err := env.main.Snapshot()
	if err != nil {
		cwarn(stderr, fmt.Sprintf("checkout not refreshed: reading main failed: %v", err))
		return
	}
	if env.onRest() {
		head, err := env.git("rev-parse", "HEAD")
		if err != nil {
			cwarn(stderr, fmt.Sprintf("%s not fast-forwarded to main: reading HEAD failed: %v", env.resting, err))
			return
		}
		if head == view.Ref() {
			return
		}
		contained, err := env.gitTest("merge-base", "--is-ancestor", "HEAD", view.Ref())
		if err != nil {
			cwarn(stderr, fmt.Sprintf("%s not fast-forwarded to main: comparing it with main failed: %v", env.resting, err))
			return
		}
		if !contained {
			cwarn(stderr, fmt.Sprintf("%s not fast-forwarded to main: it has commits main lacks; reconcile them before shaping here", env.resting))
			return
		}
		if _, err := env.git("merge", "-q", "--ff-only", view.Ref()); err != nil {
			cwarn(stderr, fmt.Sprintf("%s not fast-forwarded to main (a local change is in the way): %v", env.resting, err))
		}
		return
	}
	for _, id := range ids {
		local, lerr := os.ReadFile(filepath.Join(env.root, filepath.FromSlash(detailPaths[id])))
		published, perr := view.Read(detailPaths[id])
		if lerr != nil || perr != nil {
			continue
		}
		_, localBody, lerr := issue.Parse(string(local))
		_, mainBody, perr := issue.Parse(string(published))
		if lerr == nil && perr == nil && localBody != mainBody {
			cwarn(stderr, fmt.Sprintf("#%s's details here differ from main's; publish them with `sdlc issue publish --issue %s`, or bring main in", issue.CLIRef(id), issue.CLIRef(id)))
		}
	}
}

// claimIssues is the issue set a claim names: the --issue list, else the
// single Issue a caller built the flags with. Duplicates collapse, order kept.
func claimIssues(f *claimFlags) []int {
	nums := f.Issues
	if len(nums) == 0 && f.Issue > 0 {
		nums = []int{f.Issue}
	}
	seen := map[int]bool{}
	var out []int
	for _, n := range nums {
		if n > 0 && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}

// claimRefs names a set for humans: "#284" or "#284, #285".
func claimRefs(ids []string) string { return issue.JoinRefs(ids, "#", ", ") }

// claimArg is the --issue value that reruns the same claim.
func claimArg(ids []string) string { return issue.JoinRefs(ids, "", ",") }

// finishRelocation is a repeat claim finishing the owner's own `sdlc move`
// whose owner update failed (#278); done reports it handled the claim.
func finishRelocation(stdout, stderr io.Writer, env *trackerEnv, card tracker.Record, me issue.Claimant, detailPath string, dryRun bool) (bool, error) {
	id := card.ID
	status, _ := issue.GetField(card.Card.Frontmatter, "status")
	if !slices.Contains(vocab.Issue().OwnershipEvent("move").Statuses, status) {
		return false, nil
	}
	own, recorded, _, err := ownership(env, card)
	if err != nil || own != issue.OwnershipForeign {
		return false, nil
	}
	ok, err := relocatable(env, card, recorded, me)
	if err != nil || !ok {
		return false, err
	}
	if dryRun {
		cinfo(stderr, fmt.Sprintf("dry-run — would record #%s's owner as this workspace (relocated from %s)", id, recorded.Worktree))
		return true, nil
	}
	if err := relocateClaimant(env, card, me); err != nil {
		return true, err
	}
	cok(stderr, fmt.Sprintf("#%s relocated: owner %s → %s", id, recorded.Worktree, me.Worktree))
	if warn := refreshLocalMirror(env, detailPath); warn != "" {
		cwarn(stderr, warn)
	}
	fmt.Fprintln(stdout, "claimed")
	return true, nil
}

// refreshLocalMirror brings this checkout's copy of an issue's card fields up
// to the tracker, when the checkout has the details. It never fails a verb that
// already published: a refusal (hand-edited mirrored field) is reported instead.
// The resting branch is never edited — a local change there is exactly the
// divergence the tracker exists to remove; its mirror refreshes when the issue
// branch is prepared (start-plan) or through main.
func refreshLocalMirror(env *trackerEnv, detailPath string) string {
	if env.onRest() {
		return ""
	}
	return refreshLocalMirrorAt(env, filepath.Join(env.root, filepath.FromSlash(detailPath)))
}

// archiveMirrors projects the done card into details files an archive just
// moved, before the archive commits them (#275). Unmirrored files (legacy
// details, plan artifacts) are left alone; a refusal — a hand-edited mirrored
// field — archives the file as it is, with a warning. The card stays the
// authority either way. The tracker is opened once per archive run.
type archiveMirrors struct {
	ctx    context.Context
	stderr io.Writer
	env    *trackerEnv
	err    error
}

func newArchiveMirrors(ctx context.Context, stderr io.Writer) *archiveMirrors {
	return &archiveMirrors{ctx: ctx, stderr: stderr}
}

func (m *archiveMirrors) refresh(path string) {
	raw, err := os.ReadFile(path)
	if err != nil || !issue.HasMirror(raw) {
		return
	}
	if m.env == nil && m.err == nil {
		m.env, m.err = openTrackerAt(m.ctx, filepath.Dir(path))
	}
	env, err := m.env, m.err
	if err == nil {
		abs, aerr := filepath.Abs(path)
		if err = aerr; err == nil {
			if warn := refreshLocalMirrorAt(env, abs); warn != "" {
				err = errors.New(warn)
			}
		}
	}
	if err != nil {
		cwarn(m.stderr, fmt.Sprintf("%s archived with a stale mirror: %v", filepath.Base(path), err))
	}
}

// refreshLocalMirrorAt refreshes one details file wherever it is — for a
// caller that commits the result itself (an archive of a done issue).
func refreshLocalMirrorAt(env *trackerEnv, abs string) string {
	details, err := os.ReadFile(abs)
	if errors.Is(err, os.ErrNotExist) {
		return ""
	}
	if err != nil {
		return fmt.Sprintf("local details unreadable, mirror not refreshed: %v", err)
	}
	id, _, ok := issue.ParseFilename(filepath.Base(abs))
	if !ok {
		return fmt.Sprintf("%s: not an issue filename; mirror not refreshed", abs)
	}
	refreshed, err := refreshMirror(env, id, details)
	if err != nil {
		return fmt.Sprintf("%s: mirror not refreshed: %v", filepath.Base(abs), err)
	}
	if !bytes.Equal(refreshed, details) {
		if err := os.WriteFile(abs, refreshed, 0o644); err != nil {
			return fmt.Sprintf("%s: mirror not refreshed: %v", filepath.Base(abs), err)
		}
	}
	return ""
}

// refreshMirror projects the current card into details, proving first that
// the details' mirrored fields are the untouched projection of their baseline.
func refreshMirror(env *trackerEnv, id string, details []byte) ([]byte, error) {
	if !issue.HasMirror(details) {
		return nil, fmt.Errorf("%w: #%s's details have no card_mirror; run `sdlc issue migrate --reconcile` on this branch first", tracker.ErrLegacyDetails, issue.CLIRef(id))
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		return nil, err
	}
	current, err := snap.Require(id)
	if err != nil {
		return nil, err
	}
	return refreshMirrorFrom(env, current, details)
}

// refreshMirrorFrom is refreshMirror over a card already read — for a caller
// that judges the same card version first (change-code's ownership gate, #277).
func refreshMirrorFrom(env *trackerEnv, current tracker.Record, details []byte) ([]byte, error) {
	if !issue.HasMirror(details) {
		return nil, fmt.Errorf("%w: #%s's details have no card_mirror; run `sdlc issue migrate --reconcile` on this branch first", tracker.ErrLegacyDetails, issue.CLIRef(current.ID))
	}
	baselineOID, err := issue.MirrorBaselineOID(details)
	if err != nil {
		return nil, err
	}
	baseline, err := env.repo.ReadCardBlob(baselineOID)
	if err != nil {
		return nil, fmt.Errorf("mirror baseline %s unavailable: %w", baselineOID, err)
	}
	return issue.RefreshMirror(details, baseline, current.Raw)
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

// adoptClaim records this workspace as the owner of a working or blocked card
// with no recorded owner (#277), by compare-and-swap on the card it read. An
// owned card refuses — adoption never reassigns.
func adoptClaim(stdout, stderr io.Writer, env *trackerEnv, card tracker.Record, me issue.Claimant, detailPath string, dryRun bool) error {
	id := card.ID
	next, err := adoptDecision(card.Raw, id, me)
	if errors.Is(err, errAlreadyMine) {
		cok(stderr, fmt.Sprintf("#%s is already owned by this workspace; nothing to do", id))
		return nil
	}
	if err != nil {
		return err
	}
	if dryRun {
		cinfo(stderr, fmt.Sprintf("dry-run — would record this workspace (%s) as #%s's owner", me.Worktree, id))
		return nil
	}
	err = uncertainCardWrite(cardPublish(env, card, next, operationToken("adopt"), nil, nil), fmt.Sprintf("sdlc claim --issue %s --adopt", issue.CLIRef(id)))
	invalidateIssueRecords(env.ctx)
	if errors.Is(err, tracker.ErrCardChanged) {
		return fmt.Errorf("card #%s changed while adopting (a peer may have adopted it); `sdlc issue show --issue %s` and retry only if it still has no owner", id, issue.CLIRef(id))
	}
	if err != nil {
		return err
	}
	cok(stderr, fmt.Sprintf("#%s adopted: this workspace (%s) is its recorded owner", id, me.Worktree))
	if warn := refreshLocalMirror(env, detailPath); warn != "" {
		cwarn(stderr, warn)
	}
	fmt.Fprintln(stdout, "adopted")
	return nil
}
