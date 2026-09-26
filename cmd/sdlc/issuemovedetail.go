// issuemovedetail.go — `sdlc issue move-detail` (#252): finish an issue's
// creation early by publishing its initial details to main, without shipping
// the branch that filed it. Details are published as a new main-native commit
// and the source is removed by a narrow commit on the source branch: relative
// to a base without the file that is add-then-remove, so the branch's eventual
// merge keeps main's details and every later edit to them.
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
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

type moveDetailFlags struct {
	Issue     int
	IssuesDir string
	DryRun    bool
}

func newIssueMoveDetailCmd() *cobra.Command {
	f := moveDetailFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:   "move-detail",
		Short: "Publish an issue's initial details to main so it becomes claimable",
		Long: `Complete an issue's creation by publishing its initial details to main
(#252), without shipping the branch that filed it.

With local details (the creator's checkout), their current bytes are published
as a new commit on main, then removed here: by a narrow commit on a feature
branch, or by fast-forwarding a resting branch to that publication. The
branch's later merge keeps main's copy and every edit made to it since.

Without local details, initial details are derived from the card.

Refuses when details already exist on main (creation is complete and someone
may own them), during a merge/rebase, when the source changed after it was
read, or when the source file is also on the branch's merge base with main.
An interrupted run resumes with ` + "`sdlc issue recovery reconcile --issue N`" + `.`,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMoveDetail(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "issue ID (required)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue details")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "check preconditions and print the plan; change nothing")
	return cmd
}

// gitOperationMarkers are the in-progress states during which the index and
// HEAD are not the user's settled intent.
var gitOperationMarkers = []string{"MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "REBASE_HEAD", "rebase-merge", "rebase-apply", "BISECT_LOG"}

func gitOperationInProgress(env *trackerEnv) (string, error) {
	for _, m := range gitOperationMarkers {
		p, err := env.git("rev-parse", "--git-path", m)
		if err != nil {
			return "", err
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(env.root, p)
		}
		if _, err := os.Lstat(p); err == nil {
			return m, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	return "", nil
}

func runMoveDetail(ctx context.Context, stdout, stderr io.Writer, f *moveDetailFlags) error {
	if f.Issue <= 0 {
		return errors.New("--issue is required and must be positive")
	}
	dirs, err := resolveIDDirs(f.IssuesDir, envOr("WF_HISTORY_DIR", "workshop/history"))
	if err != nil {
		return err
	}
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	if env.branch == "" {
		return errors.New("move-detail needs a checked-out branch (the source's home)")
	}
	if op, err := gitOperationInProgress(env); err != nil {
		return err
	} else if op != "" {
		return fmt.Errorf("a Git operation is in progress (%s); finish or abort it first — nothing was changed", op)
	}
	id := fmt.Sprintf("%06d", f.Issue)
	receipts, err := env.receipts()
	if err != nil {
		return err
	}
	if pending, err := receiptsFor(receipts, id); err != nil {
		return err
	} else if len(pending) > 0 {
		return fmt.Errorf("#%s has an unfinished %s operation here; resume it with `sdlc issue recovery reconcile --issue %s`", id, pending[0].Operation(), issue.CLIRef(id))
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	card, ok := snap.Card(id)
	if !ok {
		return fmt.Errorf("no card #%s on the tracker", id)
	}
	if h, ok, err := issue.CardHandoff(card.Raw); err != nil {
		return err
	} else if ok {
		return fmt.Errorf("#%s's initial details were already handed off by operation %s (from %s); nothing to move", id, h.Token, h.SourceBranch)
	}
	dest := path.Join(dirs.Rel[0], path.Base(card.Path))
	mainView, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	pinnedMain := mainView.Ref()
	if present, err := mainView.Exists(dest); err != nil {
		return err
	} else if present {
		return fmt.Errorf("#%s's details (%s) are already on main: creation is complete and another thread may own them; nothing was changed", id, dest)
	}

	spec := tracker.ReceiptSpec{Token: operationToken("move"), Repository: env.target.Repository, IssueID: id,
		CardPath: card.Path, SourcePath: dest, DestinationPath: dest, SourceBranch: env.branchRef(),
		SourceHEAD: env.head, CardOID: card.BlobOID, TrackerBase: snap.Ref(), MainBase: pinnedMain}
	abs := filepath.Join(env.root, filepath.FromSlash(dest))
	info, statErr := os.Lstat(abs)
	var source []byte
	staleMirror := false
	switch {
	case statErr == nil:
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not an ordinary file; refusing to publish it", dest)
		}
		if source, err = os.ReadFile(abs); err != nil {
			return fmt.Errorf("read %s: %w", dest, err)
		}
		if !issue.HasMirror(source) {
			return fmt.Errorf("%s has no card mirror; it predates the tracker and cannot be handed off", dest)
		}
		if err := checkMoveSource(env, dest, pinnedMain); err != nil {
			return err
		}
		// Publishing makes these bytes main's details: a hand-edited card field is
		// refused, and a merely stale mirror is refreshed — in memory here, and
		// on disk only after every check (below), so the later removal matches.
		refreshed, err := refreshMirror(env, id, source)
		if err != nil {
			return fmt.Errorf("%s: %w — nothing was changed", dest, err)
		}
		staleMirror = !bytes.Equal(refreshed, source)
		source = refreshed
		spec.Source = tracker.LocalSource
	case errors.Is(statErr, os.ErrNotExist):
		if source, err = issue.DetailsFromCard(card.Raw, env.format, time.Now().Format("2006-01-02")); err != nil {
			return err
		}
		spec.Source = tracker.CardSource
	default:
		return fmt.Errorf("cannot tell whether %s exists (%v); a read error is not absence — nothing was changed", dest, statErr)
	}
	if spec.SourceBase, err = moveSourceBase(env, dest, pinnedMain); err != nil {
		return err
	}
	if spec.SourceBlob, err = gitx.WriteBlob(ctx, env.root, source); err != nil {
		return err
	}
	if f.DryRun {
		from := "the card (no local details)"
		if spec.Source == tracker.LocalSource {
			from = dest + " on " + env.branch
		}
		cinfo(stderr, "dry-run — preconditions hold; nothing changed")
		fmt.Fprintf(stdout, "Would publish #%s's initial details from %s to main as %s\n", id, from, dest)
		return nil
	}
	t, err := tracker.NewTransfer(spec)
	if err != nil {
		return err
	}
	if staleMirror {
		// The first effect, after every fallible check including the receipt's
		// own validation: refresh the source file and, where Git tracks it,
		// its index entry too, so no staged/unstaged split is introduced.
		if err := os.WriteFile(abs, source, info.Mode().Perm()); err != nil {
			return err
		}
		staged, err := env.git("ls-files", "--cached", "--", dest)
		if err != nil {
			return err
		}
		if staged != "" {
			if _, err := env.git("add", "--", dest); err != nil {
				return err
			}
		}
	}
	op := tracker.NewTransferOp(ctx, env.repo, env.main, env.root, env.branchRef(), moveDetailRemover(env))
	final, err := tracker.Drive(t.Receipt(), tracker.TransferStepper, op, receipts)
	if err != nil {
		if errors.Is(err, tracker.ErrOperationUncertain) || final.ConfirmedStages() > 0 {
			return fmt.Errorf("%w\n      the operation is recorded; finish it with `sdlc issue recovery reconcile --issue %s`", err, issue.CLIRef(id))
		}
		if final.Discardable() {
			_ = receipts.Discard(final)
		}
		return err
	}
	cok(stderr, fmt.Sprintf("#%s's initial details are on main (%s); it is now claimable.", id, shortOID(tracker.MainCommit(final))))
	if spec.Source == tracker.LocalSource && !env.onRest() {
		cinfo(stderr, "the source was removed by a narrow commit on "+env.branch+"; its later merge keeps main's copy")
	}
	fmt.Fprintln(stdout, dest)
	return nil
}

func shortOID(oid string) string {
	if len(oid) > 12 {
		return oid[:12]
	}
	return oid
}

// checkMoveSource enforces the local-source preconditions: no staged/unstaged
// split, and on the resting branch no local-only commits (it will
// fast-forward to the publication).
func checkMoveSource(env *trackerEnv, dest, pinnedMain string) error {
	// A split is a staged version (index ≠ HEAD) with further unstaged edits
	// (worktree ≠ index): publishing either side would silently drop the other.
	// An ordinary uncommitted edit, or a fully staged one, is not a split.
	worktreeClean, err := env.gitTest("diff", "--quiet", "--", dest)
	if err != nil {
		return err
	}
	indexClean, err := env.gitTest("diff", "--cached", "--quiet", "--", dest)
	if err != nil {
		return err
	}
	if !worktreeClean && !indexClean {
		return fmt.Errorf("%s has unstaged changes over a staged version; stage or discard one side first — nothing was changed", dest)
	}
	if env.onRest() {
		if contained, err := env.gitTest("merge-base", "--is-ancestor", "HEAD", pinnedMain); err != nil {
			return err
		} else if !contained {
			return fmt.Errorf("%s has commits that are not on main; the handoff fast-forwards it, so reconcile them first — nothing was changed", env.resting)
		}
	}
	return nil
}

// moveSourceBase returns the merge base with main, after proving the details
// are absent at every merge base: only then is the source branch's
// add-then-remove net-zero against main's copy.
func moveSourceBase(env *trackerEnv, dest, pinnedMain string) (string, error) {
	out, err := env.git("merge-base", "--all", "HEAD", pinnedMain)
	if err != nil || out == "" {
		return "", fmt.Errorf("%s shares no history with main; cannot hand off from it", env.branch)
	}
	bases := strings.Fields(out)
	for _, base := range bases {
		if present, err := env.has(base, dest); err != nil {
			return "", err
		} else if present {
			return "", fmt.Errorf("%s already exists at merge base %s; this is not an initial handoff", dest, shortOID(base))
		}
	}
	return bases[0], nil
}

// moveDetailRemover relinquishes the verified source. On the resting branch it
// fast-forwards to the publication (the file then comes back as main's copy);
// elsewhere a tracked source is deleted by a narrow commit and an untracked one
// by removing exactly its verified bytes. Idempotent: an already-removed source
// is success.
func moveDetailRemover(env *trackerEnv) tracker.Remover {
	return func(spec tracker.ReceiptSpec, mainCommit string) error {
		abs := filepath.Join(env.root, filepath.FromSlash(spec.SourcePath))
		current, err := os.ReadFile(abs)
		switch {
		case errors.Is(err, os.ErrNotExist):
			current = nil
		case err != nil:
			return fmt.Errorf("read source %s: %w", spec.SourcePath, err)
		}
		if env.onRest() {
			if done, err := env.gitTest("merge-base", "--is-ancestor", mainCommit, "HEAD"); err != nil {
				return err
			} else if done {
				return nil // already fast-forwarded
			}
		}
		if current != nil {
			oid, err := gitx.WriteBlob(env.ctx, env.root, current)
			if err != nil {
				return err
			}
			if oid != spec.SourceBlob {
				return fmt.Errorf("%s changed after it was published; it is left untouched. Main has the published copy (%s); reconcile your edits into it, then `sdlc issue recovery reconcile --issue %s`", spec.SourcePath, shortOID(mainCommit), issue.CLIRef(spec.IssueID))
			}
		}
		tracked, err := env.has("HEAD", spec.SourcePath)
		if err != nil {
			return err
		}
		staged, err := env.git("ls-files", "--cached", "--", spec.SourcePath)
		if err != nil {
			return err
		}
		if staged != "" {
			if _, err := env.git("rm", "-q", "--cached", "--", spec.SourcePath); err != nil {
				return err
			}
		}
		if current != nil {
			if err := os.Remove(abs); err != nil {
				return err
			}
		}
		if env.onRest() {
			if _, err := env.git("merge", "-q", "--ff-only", mainCommit); err != nil {
				return fmt.Errorf("fast-forward %s to the publication: %w", env.resting, err)
			}
			return nil
		}
		if !tracked {
			return nil
		}
		msg := fmt.Sprintf("#%s: issue: hand off initial details to main\n\nTracker-Operation: %s\nDetail-Handoff: #%s", issue.CLIRef(spec.IssueID), spec.Token, spec.IssueID)
		if _, err := env.git("commit", "-q", "--no-verify", "-m", msg, "--only", "--", spec.SourcePath); err != nil {
			return fmt.Errorf("commit source removal: %w", err)
		}
		return nil
	}
}

// receiptsFor lists this checkout's unfinished operations for one issue.
func receiptsFor(receipts *tracker.RecoveryReceipts, id string) ([]tracker.Receipt, error) {
	all, err := receipts.List()
	var mine []tracker.Receipt
	for _, r := range all {
		if r.Spec().IssueID == id {
			mine = append(mine, r)
		}
	}
	return mine, err
}
