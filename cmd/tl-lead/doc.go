package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func newDocCmd(d deps) *cobra.Command {
	parent := &cobra.Command{Use: "doc", Short: "Settle a working document and inspect drift"}
	repo := d.documents
	if repo == nil {
		repo = execDocRepository{sdlcBin: "sdlc"}
	}
	var target string
	settle := &cobra.Command{
		Use:   "settle SIDE-DOC [--to PATH]",
		Short: "Publish the agreed side document and stamp its settlement pointer",
		Long:  "Publish one resolved side document into a tracked copy, then stamp its confirmed path and commit through Couch. Remaining review questions, proposals and malformed markers refuse. Use a quiet editing window. A published copy survives a failed source stamp; reconcile the named source and retry.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			root, err := git(cwd, "rev-parse", "--show-toplevel")
			if err != nil {
				return fmt.Errorf("doc settle requires a Git checkout: %w", err)
			}
			invocation := d
			switch c := d.ss.(type) {
			case couchSessionSync:
				c.ctx = cmd.Context()
				invocation.ss = c
			case *couchSessionSync:
				copy := *c
				copy.ctx = cmd.Context()
				invocation.ss = &copy
			}
			result, err := settleWorkingdoc(cmd.Context(), invocation, repo, root, args[0], target)
			for _, w := range result.Warnings {
				fmt.Fprintln(cmd.ErrOrStderr(), "warning:", w)
			}
			if err != nil {
				if errors.Is(err, ErrLocked) {
					return exitError{3, err}
				}
				if errors.Is(err, ErrStale) {
					return exitError{6, err}
				}
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "settled %s @ %s; live source retained\n", result.Target, result.Commit)
			return nil
		},
	}
	settle.Flags().StringVar(&target, "to", "", "repository-relative tracked Markdown destination (recorded target on re-settle)")
	status := &cobra.Command{
		Use: "status", Short: "Show side documents changed since their last settled edition", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			session, err := defaultStatusPath()
			if err != nil {
				return err
			}
			rows, err := readWorkingdocStatus(cmd.Context(), repo, filepath.Dir(session), readWorkingdoc)
			if len(rows) == 0 && err == nil {
				fmt.Fprintln(cmd.OutOrStdout(), "No side documents.")
			}
			for _, r := range rows {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s", r.State, r.Source)
				if r.Target != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "\t%s @ %s", r.Target, r.Commit)
				}
				if r.Detail != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "\t%s", r.Detail)
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
			return err
		},
	}
	parent.AddCommand(settle, status)
	return parent
}
