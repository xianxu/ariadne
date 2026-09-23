package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

func newIssuePublishCmd() *cobra.Command {
	var source string
	cmd := markMutatingCommand(&cobra.Command{
		Use: "publish --commit SHA", Short: "Publish one selected documentation commit to origin/main",
		Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			guardSpineRepo(cmd.ErrOrStderr())
			if source == "" {
				return fmt.Errorf("--commit SHA is required; select one coherent documentation commit")
			}
			root, err := gitx.RepoTopLevel()
			if err != nil {
				return err
			}
			return publishIssueCommit(cmd.OutOrStdout(), cmd.ErrOrStderr(), root, source,
				envOr("WF_ISSUES_DIR", "workshop/issues"), envOr("WF_HISTORY_DIR", "workshop/history"))
		},
	})
	cmd.Flags().StringVar(&source, "commit", "", "one non-merge documentation commit to publish in full")
	return cmd
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
