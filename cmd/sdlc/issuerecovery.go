// issuerecovery.go — `sdlc issue recovery` (#252): the checkout-local receipts
// of tracker operations that stopped before finishing (an uncertain push, an
// interrupted process). Listing never mutates; reconcile resumes an operation
// through the same transition model, and releases one that published nothing.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func newIssueRecoveryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recovery",
		Short: "List or resume interrupted tracker operations in this checkout",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List unfinished operation receipts (read-only)",
		Args:  cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRecoveryList(cmd.Context(), cmd.OutOrStdout())
		},
	})
	var issueID int
	reconcile := markMutatingCommand(&cobra.Command{
		Use:   "reconcile",
		Short: "Resume (or release, when nothing was published) an issue's unfinished operations",
		Args:  cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRecoveryReconcile(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), issueID)
		},
	})
	reconcile.Flags().IntVar(&issueID, "issue", 0, "issue ID (required)")
	cmd.AddCommand(reconcile)
	return cmd
}

func runRecoveryList(ctx context.Context, stdout io.Writer) error {
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	receipts, err := env.receipts()
	if err != nil {
		return err
	}
	all, listErr := receipts.List()
	if len(all) == 0 && listErr == nil {
		fmt.Fprintln(stdout, "no unfinished tracker operations")
		return nil
	}
	here := env.branchRef()
	for _, r := range all {
		spec := r.Spec()
		where := "here"
		if r.NeedsSourceCheckout() && spec.SourceBranch != here {
			where = "finish from " + strings.TrimPrefix(spec.SourceBranch, "refs/heads/")
		}
		fmt.Fprintf(stdout, "#%s\t%s\t%s\t%s\t%s\t%s\n", spec.IssueID, r.Operation(), r.Stage(), r.Outcome(), spec.Token, where)
	}
	return listErr
}

// runRecoveryReconcile resumes each of an issue's receipts. Creation cannot
// re-render a draft after losing its ID, so a creation that published nothing
// is released instead (the operator reruns `issue new`); every other state is
// driven forward, probing an uncertain effect before anything is repeated.
func runRecoveryReconcile(ctx context.Context, stdout, stderr io.Writer, issueID int) error {
	if issueID <= 0 {
		return errors.New("--issue is required and must be positive")
	}
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	receipts, err := env.receipts()
	if err != nil {
		return err
	}
	mine, err := receiptsFor(receipts, fmt.Sprintf("%06d", issueID))
	if err != nil {
		return err
	}
	defer func() {
		// A close whose evidence already landed (a merge made elsewhere, or an
		// interrupted publish) is completed by re-derivation, receipt or not.
		if settled, serr := settleLandedCompletions(ctx, env, envOr("WF_ISSUES_DIR", "workshop/issues")); serr != nil {
			cwarn(stderr, fmt.Sprintf("landed closes not settled: %v", serr))
		} else {
			for _, id := range settled {
				cok(stderr, fmt.Sprintf("#%s landed; its card is now done", issue.CLIRef(id)))
			}
		}
	}()
	if len(mine) == 0 {
		cok(stderr, fmt.Sprintf("no unfinished operations for #%d here", issueID))
		return nil
	}
	mainView, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	here := env.branchRef()
	for _, r := range mine {
		// Receipts are visible from every linked worktree; a remaining local effect
		// belongs to the one on the source branch. Refuse before touching anything.
		if r.NeedsSourceCheckout() && r.Spec().SourceBranch != here && !(r.Discardable() && r.Operation() == "creation") {
			return fmt.Errorf("%w: #%d's %s must be finished from the checkout on %s (this one is on %q)",
				tracker.ErrForeignCheckout, issueID, r.Operation(), r.Spec().SourceBranch, env.branch)
		}
		if r.Discardable() && r.Operation() == "creation" {
			if err := receipts.Discard(r); err != nil {
				return err
			}
			cwarn(stderr, fmt.Sprintf("#%d's creation published nothing; released it — rerun `sdlc issue new`", issueID))
			continue
		}
		var step tracker.Stepper
		var adapter tracker.Adapter
		switch r.Operation() {
		case "creation":
			resumed, err := tracker.ResumeCreation(r)
			if err != nil {
				return err
			}
			r = resumed.Receipt()
			step = tracker.CreationStepper
			adapter = tracker.NewCreationOp(ctx, env.repo, env.checkout(mainView.Ref()), func(string) (tracker.Draft, error) {
				return tracker.Draft{}, errors.New("the reserved ID was lost; nothing was published — rerun `sdlc issue new`")
			})
		case "transfer":
			resumed, err := tracker.ResumeTransfer(r)
			if err != nil {
				return err
			}
			r = resumed.Receipt()
			step = tracker.TransferStepper
			adapter = tracker.NewTransferOp(ctx, env.repo, env.main, env.root, env.branchRef(), moveDetailRemover(env))
		case "completion":
			resumed, err := tracker.ResumeCompletion(r)
			if err != nil {
				return err
			}
			r = resumed.Receipt()
			step = tracker.CompletionStepper
			adapter = tracker.NewCompletionOp(ctx, env.repo, env.branchRef(), gitEvidence{env, stderr}, time.Now().Format("2006-01-02"), env.ancestorOf)
		default:
			return fmt.Errorf("receipt %s: %s operations are recovered by their own verb", r.Spec().Token, r.Operation())
		}
		final, err := tracker.Drive(r, step, adapter, receipts)
		invalidateIssueRecords(env.ctx)
		if errors.Is(err, tracker.ErrSupersededClose) && final.Discardable() {
			// A newer close owns the card; this one never recorded anything.
			if derr := receipts.Discard(final); derr != nil {
				return derr
			}
			cwarn(stderr, fmt.Sprintf("#%d: released close %s — %v", issueID, final.Spec().Token, err))
			continue
		}
		if err != nil {
			return fmt.Errorf("#%d %s (%s): %w", issueID, r.Operation(), final.Stage(), err)
		}
		// The mirror commit belongs on the close's own branch, never another checkout's.
		if r.Operation() == "completion" && final.Spec().SourceBranch == env.branchRef() {
			commitCloseMirror(env, stderr, final.Spec().IssueID, final.Spec().SourcePath)
		}
		cok(stderr, fmt.Sprintf("#%d %s finished", issueID, r.Operation()))
	}
	fmt.Fprintln(stdout, "reconciled")
	return nil
}
