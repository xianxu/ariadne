package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/pkg/workspace"
)

func newLandingDropCmd(repo *string, d landingDeps) *cobra.Command {
	var issue int
	var reason string
	cmd := &cobra.Command{Use: "drop", Short: "Hold one composed car before its train starts merging (owner or repository :0)", Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLandingDrop(cmd.OutOrStdout(), *repo, issue, reason, d)
		},
	}
	cmd.Flags().IntVar(&issue, "issue", 0, "the composed car's issue number")
	cmd.Flags().StringVar(&reason, "reason", "requested by owner or steward", "why this car must stay out")
	return cmd
}

// Authority comes from the actual caller, independently of --repo. The owning
// checkout was pinned when the car boarded, not supplied by the drop command.
func authorizeTrainDrop(caller, target, owner workspace.Identity, tree *int) error {
	if caller.RepoIdentity == "" || caller.RepoIdentity != target.RepoIdentity {
		return errors.New("landing drop: caller and target must be the same repository")
	}
	if caller.Slot != nil && *caller.Slot == 0 {
		if tree != nil && *tree != 0 {
			return errors.New("landing drop: a worker entering :0 cannot act as the steward")
		}
		return nil
	}
	if owner.WorktreeRoot != "" && owner.RepoIdentity == caller.RepoIdentity &&
		(canonicalRoot(caller.WorktreeRoot) == canonicalRoot(owner.WorktreeRoot) ||
			caller.HomeSlot() != "" && caller.HomeSlot() == owner.HomeSlot()) {
		return nil
	}
	return fmt.Errorf("landing drop: only the car's owner at %s or this repository's :0 may drop it", owner.WorktreeRoot)
}

func runLandingDrop(out io.Writer, repo string, issue int, reason string, d landingDeps) error {
	if issue <= 0 {
		return errors.New("--issue N is required")
	}
	if strings.TrimSpace(reason) == "" {
		return errors.New("--reason must name why the car is held")
	}
	line, err := openRepoLine(repo)
	if err != nil {
		return err
	}
	control, err := line.ReadTrainControl()
	if err != nil {
		return fmt.Errorf("landing drop: control unavailable; no drop accepted: %w", err)
	}
	if control == nil {
		return fmt.Errorf("landing drop: #%d has no active compatible composed train; no drop accepted", issue)
	}
	var member landingline.TrainMember
	count := 0
	for _, m := range control.Members {
		if m.Issue == issue {
			member = m
			count++
		}
	}
	if count == 0 { // A retry can observe an already accepted, pruned member.
		for _, drop := range control.Drops {
			if drop.Train == control.Train && drop.Member.Issue == issue {
				member = drop.Member
				count++
			}
		}
	}
	if count != 1 {
		return fmt.Errorf("landing drop: #%d is not one unambiguous composed member of train %d (%s); no drop accepted", issue, control.Train, control.Phase)
	}
	resolve := d.dropIdentity
	if resolve == nil {
		resolve = func(dir string) (workspace.Identity, error) { return workspace.Resolve(execGitRunner{}, dir, "") }
	}
	caller, err := resolve(".")
	if err != nil {
		return fmt.Errorf("landing drop: caller identity unavailable: %w", err)
	}
	target, err := resolve(repo)
	if err != nil {
		return fmt.Errorf("landing drop: target identity unavailable: %w", err)
	}
	var owner workspace.Identity
	if member.OwnerRoot != "" {
		owner, _ = resolve(member.OwnerRoot)
	}
	tree, _ := d.couchTree()
	if err := authorizeTrainDrop(caller, target, owner, tree); err != nil {
		return err
	}
	receipt, err := line.RequestTrainDrop(control.Train, member, reason)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "#%d dropped from train %d; cleanup/recomposition pending: %s\n", issue, receipt.Drop.Train, receipt.Drop.Reason)
	if receipt.NotificationError != nil {
		fmt.Fprintf(out, "drop accepted; waiter notification pending: %v\n", receipt.NotificationError)
	}
	fmt.Fprintf(out, "Recovery: resolve the reason, re-close if changed, then explicitly run sdlc merge --branch %s --yes\n", member.Branch)
	return nil
}
