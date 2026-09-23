package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// defaultBrainDir resolves the fleet sibling through Git, never through cwd's
// spelling. Explicit flags and environment paths bypass this default entirely.
func defaultBrainDir() (string, error) {
	identity, err := resolveWorkspace(".")
	if err != nil {
		return "", err
	}
	return filepath.Join(identity.FleetRoot, "brain"), nil
}

// wrapBrainDefaults runs after Cobra parses flags and before a command mutates
// anything. Value.Set intentionally preserves Changed: an omitted default remains
// distinguishable from an explicitly supplied relative path.
func wrapBrainDefaults(root *cobra.Command) {
	for _, cmd := range root.Commands() {
		wrapBrainDefaults(cmd)
	}
	flag := root.Flags().Lookup("brain-dir")
	if flag == nil || (root.RunE == nil && root.Run == nil) {
		return
	}
	runE, run := root.RunE, root.Run
	root.Run = nil
	root.RunE = func(cmd *cobra.Command, args []string) error {
		if !flag.Changed {
			// The standalone estimator override is already a complete calibration path.
			// It remains usable outside Git and keeps cwd-relative interpretation.
			if cmd.Name() != "estimate-source" || os.Getenv("WF_ESTIMATOR_SRC") == "" {
				brain, err := defaultBrainDir()
				if err != nil {
					return fmt.Errorf("resolve default --brain-dir: %w", err)
				}
				if err := flag.Value.Set(brain); err != nil {
					return err
				}
			}
		}
		if runE != nil {
			return runE(cmd, args)
		}
		run(cmd, args)
		return nil
	}
}
