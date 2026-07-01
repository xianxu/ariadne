// retro.go — `sdlc retro`: unroll every always-on injection source (sdlc-injected
// prompts, embedded help text, skills, lessons, the AGENTS chain, persisted
// memories) into one linked markdown "process manual" a human can read to audit
// the process (#153). Deterministic regeneration, not a hand-maintained doc.
//
// The pure core (InjectionSource + renderManual) and the IO collectors live in
// cmd/sdlc/internal/retro; this file is the thin cobra glue mirroring the other
// read-only verbs (NewEstimateSourceCmd).
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/retro"
)

// NewRetroCmd returns the cobra command for `sdlc retro`.
func NewRetroCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:           "retro",
		Short:         "Unroll every injection source into a linked process manual",
		Long:          "Placeholder — replaced by helptext.MustGet(\"retro\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRetro(cmd.OutOrStdout(), cmd.ErrOrStderr(), out)
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "write the manual to a file (default: stdout)")
	return cmd
}

// runRetro resolves the repo root + home, builds the manual, and writes it to
// stdout or --out. For --out, links are prefixed with the out file's path back
// to the repo root so they stay clickable (assumes --out is within the repo).
func runRetro(stdout, stderr io.Writer, outPath string) error {
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	home, _ := os.UserHomeDir()
	opts := retro.CollectOptions{
		RepoRoot:  root,
		SkillsDir: filepath.Join(root, ".claude", "skills"),
		HomeDir:   home,
	}

	linkPrefix := ""
	if outPath != "" {
		if outAbs, aerr := filepath.Abs(outPath); aerr == nil {
			if rel, rerr := filepath.Rel(filepath.Dir(outAbs), root); rerr == nil && rel != "." {
				linkPrefix = rel + "/"
			}
		}
	}

	manual := retro.Manual(opts, linkPrefix)
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
