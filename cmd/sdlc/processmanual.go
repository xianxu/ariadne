// processmanual.go — `sdlc process-manual`: unroll every always-on injection
// source (sdlc-injected prompts, embedded help text, skills, lessons, the AGENTS
// chain, persisted memories) into one linked markdown "process manual" a human
// can read to audit the process (#153). Deterministic regeneration, not a
// hand-maintained doc.
//
// The pure core (InjectionSource + renderManual) and the IO collectors live in
// cmd/sdlc/internal/processmanual; this file is the thin cobra glue mirroring the
// other read-only verbs (NewEstimateSourceCmd).
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processmanual"
)

// NewProcessManualCmd returns the cobra command for `sdlc process-manual`.
func NewProcessManualCmd() *cobra.Command {
	var out string
	var full bool
	cmd := &cobra.Command{
		Use:           "process-manual",
		Short:         "Unroll every injection source into a linked process manual",
		Long:          "Placeholder — replaced by helptext.MustGet(\"process-manual\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProcessManual(cmd.OutOrStdout(), cmd.ErrOrStderr(), out, full)
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "write the manual to a file (default: stdout)")
	cmd.Flags().BoolVar(&full, "full", false, "inline the complete judge prompts instead of a first-paragraph gist (outline unchanged)")
	return cmd
}

// runProcessManual resolves the repo root + home, builds the manual, and writes
// it to stdout or --out. For --out, links are prefixed with the out file's path
// back to the repo root so they stay clickable (assumes --out is within the repo).
func runProcessManual(stdout, stderr io.Writer, outPath string, full bool) error {
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	home, _ := os.UserHomeDir()
	opts := processmanual.CollectOptions{
		RepoRoot:  root,
		SkillsDir: filepath.Join(root, ".claude", "skills"),
		HomeDir:   home,
		Full:      full,
	}

	linkPrefix := ""
	if outPath != "" {
		if outAbs, aerr := filepath.Abs(outPath); aerr == nil {
			if rel, rerr := filepath.Rel(filepath.Dir(outAbs), root); rerr == nil && rel != "." {
				linkPrefix = rel + "/"
			}
		}
	}

	manual := processmanual.Manual(opts, linkPrefix)
	if outPath == "" {
		fmt.Fprint(stdout, manual)
		return nil
	}
	if werr := os.WriteFile(outPath, []byte(manual), 0o644); werr != nil {
		return werr
	}
	fmt.Fprintf(stderr, "wrote process manual: %s\n", outPath)
	return nil
}
