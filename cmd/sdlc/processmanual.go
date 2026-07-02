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
	var out, session string
	var full, includeMemory bool
	cmd := &cobra.Command{
		Use:           "process-manual",
		Short:         "Unroll every injection source into a linked process manual",
		Long:          "Placeholder — replaced by helptext.MustGet(\"process-manual\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProcessManual(cmd.OutOrStdout(), cmd.ErrOrStderr(), out, session, full, includeMemory)
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "write the manual to a file (default: stdout)")
	cmd.Flags().StringVar(&session, "session", "", "reconstruct which injection points FIRED in a session transcript (a .jsonl path, or \"current\" for this repo's active session) instead of the static catalog (#157)")
	cmd.Flags().BoolVar(&full, "full", false, "inline the complete judge prompts instead of a first-paragraph gist (outline unchanged)")
	cmd.Flags().BoolVar(&includeMemory, "include-memory", false, "inline private, machine-local persisted memories (redacted by default; do NOT commit the output)")
	return cmd
}

// runProcessManual resolves the repo root + home, builds the content, and writes
// it to stdout or --out. Without --session it renders the STATIC catalog (Manual);
// with --session it reconstructs which injection points FIRED in a transcript
// (SessionReport, #157). For --out, links are prefixed with the out file's path
// back to the repo root so they stay clickable (assumes --out is within the repo).
func runProcessManual(stdout, stderr io.Writer, outPath, session string, full, includeMemory bool) error {
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	// Guard the private-data footgun: memories carry absolute home paths + personal
	// content, so writing them to a file (which tends to get shared/committed) is
	// refused. Inspect them on stdout instead.
	if includeMemory && outPath != "" {
		return fmt.Errorf("--include-memory cannot be combined with --out: memories are private + machine-local; view them on stdout, don't write them to a (committable) file")
	}
	home, _ := os.UserHomeDir()
	opts := processmanual.CollectOptions{
		RepoRoot:      root,
		SkillsDir:     filepath.Join(root, ".claude", "skills"),
		HomeDir:       home,
		Full:          full,
		IncludeMemory: includeMemory,
	}

	linkPrefix := ""
	if outPath != "" {
		if outAbs, aerr := filepath.Abs(outPath); aerr == nil {
			if rel, rerr := filepath.Rel(filepath.Dir(outAbs), root); rerr == nil && rel != "." {
				linkPrefix = rel + "/"
			}
		}
	}

	content := ""
	label := "process manual"
	if session != "" {
		content, err = processmanual.SessionReport(opts, session, linkPrefix)
		if err != nil {
			return err
		}
		label = "session reconstruction"
	} else {
		content = processmanual.Manual(opts, linkPrefix)
	}

	if outPath == "" {
		fmt.Fprint(stdout, content)
		return nil
	}
	if werr := os.WriteFile(outPath, []byte(content), 0o644); werr != nil {
		return werr
	}
	fmt.Fprintf(stderr, "wrote %s: %s\n", label, outPath)
	return nil
}
