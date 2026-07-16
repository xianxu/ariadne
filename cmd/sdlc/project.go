package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewProjectCmd returns the project-record authoring command group. The M3
// surface deliberately excludes status/retro/close: those derived and gated
// workflow verbs land together in M4.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "project",
		Short:         "Create and manage workshop projects",
		Long:          "Placeholder — replaced by renderLong(\"project\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newProjectNewCmd())
	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectShowCmd())
	cmd.AddCommand(newProjectSetStatusCmd())
	cmd.AddCommand(newProjectValidateCmd())
	return cmd
}

type projectNewFlags struct{}
type projectListFlags struct{}
type projectShowFlags struct{}
type projectSetStatusFlags struct{}
type projectValidateFlags struct{}

func deferredProjectCmd(use, short string, mutating bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           use,
		Short:         short,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(*cobra.Command, []string) error {
			return fmt.Errorf("sdlc project %s implementation lands later in M3", use)
		},
	}
	if mutating {
		return markMutatingCommand(cmd)
	}
	return cmd
}

func newProjectNewCmd() *cobra.Command {
	return deferredProjectCmd("new", "Create a project from the model-derived scaffold", true)
}

func newProjectListCmd() *cobra.Command {
	return deferredProjectCmd("list", "List live projects", false)
}

func newProjectShowCmd() *cobra.Command {
	return deferredProjectCmd("show", "Show one project and its task summary", false)
}

func newProjectSetStatusCmd() *cobra.Command {
	return deferredProjectCmd("set-status", "Move a project through its guarded lifecycle", true)
}

func newProjectValidateCmd() *cobra.Command {
	return deferredProjectCmd("validate", "Validate project records against #Project", false)
}
