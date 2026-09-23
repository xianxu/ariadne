package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// wrapProjectDefaults anchors only the project group's omitted artifact paths.
// Explicit flags, environment overrides and positional validation files keep
// their caller-relative meaning; no command changes the process cwd.
func wrapProjectDefaults(root *cobra.Command) {
	for _, cmd := range root.Commands() {
		if cmd.Name() == "project" {
			wrapProjectArtifactDefaults(cmd)
		}
	}
}

func wrapProjectArtifactDefaults(cmd *cobra.Command) {
	for _, child := range cmd.Commands() {
		wrapProjectArtifactDefaults(child)
	}
	projects := cmd.Flags().Lookup("projects-dir")
	history := cmd.Flags().Lookup("history-dir")
	if projects == nil && history == nil || cmd.RunE == nil {
		return
	}
	run := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "validate" && len(args) > 0 {
			return run(cmd, args)
		}
		var omitted []*pflag.Flag
		if projects != nil && !projects.Changed && os.Getenv("WF_PROJECTS_DIR") == "" {
			omitted = append(omitted, projects)
		}
		if history != nil && !history.Changed {
			omitted = append(omitted, history)
		}
		if len(omitted) > 0 {
			identity, err := resolveWorkspace(".")
			if err != nil {
				return fmt.Errorf("resolve default project paths: %w", err)
			}
			for _, flag := range omitted {
				value := flag.DefValue
				if !filepath.IsAbs(value) {
					value = filepath.Join(identity.WorktreeRoot, value)
				}
				if err := flag.Value.Set(value); err != nil {
					return err
				}
			}
		}
		return run(cmd, args)
	}
}
