package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// NewWorkspaceCmd exposes the shared, read-only identity contract to Couch.
func NewWorkspaceCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use: "workspace [address]", Args: cobra.MaximumNArgs(1),
		SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			address := ""
			if len(args) > 0 {
				address = args[0]
			}
			identity, err := workspace.Resolve(execGitRunner{}, ".", address)
			if err != nil {
				return err
			}
			if asJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(identity)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Repo: %s\nWorkspace: %s\nPath: %s\nBranch: %s\nResting branch: %s\n", identity.Repo, workspaceText(identity.Address, "(ordinary worktree)"), identity.WorktreeRoot, workspaceText(identity.Branch, "(detached)"), workspaceText(identity.RestingBranch, "(none)"))
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit workspace identity JSON v1")
	return cmd
}
func workspaceText(value *string, absent string) string {
	if value == nil {
		return absent
	}
	return *value
}
