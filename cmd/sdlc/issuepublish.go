package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func newIssuePublishCmd() *cobra.Command {
	var source, issuesDir string
	var issues []int
	var dryRun bool
	cmd := markMutatingCommand(&cobra.Command{
		Use:   "publish --issue N[,N…] | --commit SHA",
		Short: "Publish issue details to main: first publication or an owner's edits (#284)",
		Args:  cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			guardSpineRepo(cmd.ErrOrStderr())
			root, err := gitx.RepoTopLevel()
			if err != nil {
				return err
			}
			cut, err := tracker.CutOver(root)
			if err != nil {
				return err
			}
			if !cut {
				if len(issues) > 0 {
					return errors.New("--issue publishes in an issue tracker repository; here, publish one documentation commit with --commit SHA")
				}
				if source == "" {
					return fmt.Errorf("--commit SHA is required; select one coherent documentation commit")
				}
				return publishIssueCommit(cmd.OutOrStdout(), cmd.ErrOrStderr(), root, source,
					envOr("WF_ISSUES_DIR", "workshop/issues"), envOr("WF_HISTORY_DIR", "workshop/history"))
			}
			if source != "" {
				return errors.New("this repository uses the issue tracker: `--commit` (copying commits to main) is retired here; publish details with `sdlc issue publish --issue N`")
			}
			return runTrackedPublish(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), issues, issuesDir, dryRun)
		},
	})
	cmd.Flags().IntSliceVar(&issues, "issue", nil, "issue ID(s) whose details to publish (issue tracker repositories): --issue 284 or --issue 284,285")
	cmd.Flags().StringVar(&source, "commit", "", "one non-merge documentation commit to publish in full (legacy repositories)")
	cmd.Flags().StringVar(&issuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue details")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "check and describe; change nothing")
	return cmd
}

// runTrackedPublish publishes each issue's details to main (#284): the first
// publication — details not yet on main — is `move-detail`'s transfer, one
// issue at a time, needing no owner; details already on main are the owner's
// edits, published together by republishOwned.
func runTrackedPublish(ctx context.Context, stdout, stderr io.Writer, nums []int, issuesDir string, dryRun bool) error {
	nums = claimIssues(&claimFlags{Issues: nums})
	if len(nums) == 0 {
		return errors.New("--issue N (or a list: --issue 284,285) is required")
	}
	dirs, err := resolveIDDirs(issuesDir, envOr("WF_HISTORY_DIR", "workshop/history"))
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
	view, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	var first []int
	var republish []string
	for _, n := range nums {
		id := fmt.Sprintf("%06d", n)
		card, err := snap.Require(id)
		if err != nil {
			return err
		}
		onMain, err := view.Exists(path.Join(dirs.Rel[0], path.Base(card.Path)))
		if err != nil {
			return err
		}
		if onMain {
			republish = append(republish, id)
		} else {
			first = append(first, n)
		}
	}
	for _, n := range first {
		if err := runMoveDetail(ctx, stdout, stderr, &moveDetailFlags{Issue: n, IssuesDir: issuesDir, DryRun: dryRun}); err != nil {
			return err
		}
	}
	if len(republish) == 0 {
		return nil
	}
	if len(first) > 0 {
		if env, err = openTracker(ctx); err != nil { // a first publication may have moved this checkout
			return err
		}
	}
	return republishOwned(env, stdout, stderr, republish, dirs.Rel[0], dryRun)
}

func publishIssueCommit(stdout, stderr io.Writer, root, source, issuesDir, historyDir string) error {
	tf, err := gitx.NewTrunkFile(root, "origin", "main")
	if err != nil {
		return err
	}
	selected, err := tf.SelectCommit(source)
	if err != nil {
		return err
	}
	roots := []string{issuesDir, envOr("WF_PLANS_DIR", "workshop/plans"), historyDir, "workshop/projects"}
	for i, path := range roots {
		roots[i], err = gitx.InsideRoot(root, path)
		if err != nil {
			return err
		}
	}
	if err := validateCommitSelection(selected, roots); err != nil {
		return err
	}
	result, err := tf.PublishCommit(selected)
	if err != nil {
		return err
	}
	fmt.Fprintf(stderr, "Source commit %s: %s", result.Source, result.Outcome)
	if result.Candidate != "" {
		fmt.Fprintf(stderr, " (%s)", result.Candidate)
	}
	fmt.Fprintln(stderr)
	fmt.Fprintln(stdout, result.Outcome)
	return nil
}
