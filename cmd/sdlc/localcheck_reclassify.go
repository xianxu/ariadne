package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/pkg/workspace"
)

func newLocalCheckReclassifyCmd() *cobra.Command {
	var id, reason string
	cmd := &cobra.Command{Use: "reclassify", Short: "Audit one startup infrastructure failure as could-not-run; requires a genuine rerun", Args: cobra.NoArgs,
		Long: `Correct one exact recorded suite or affected attempt after startup-only infrastructure failure.
Required: --attempt ID --reason TEXT. The recorded owner or this repository's :0
may act. The immutable complete log must match its recorded hash and prove a
recognized startup failure. Passes, timeouts, empty, mixed, test-started or
truncated output refuse. This is not a reason-only waiver.

Original bytes, actor, reason, UTC time and diagnosis are retained before an
atomic nonpassing v3 disposition replaces the attempt. Repeat the same attempt
after interruption. No GitHub status is rewritten and no saved pass is promoted:
repair the environment, then run the exact check again. Older sdlc readers need
an upgrade for this attempt; unrelated evidence remains usable on #524 readers.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(reason) == "" || id == "" {
				return errors.New("local-check reclassify: give --attempt ID and a nonempty --reason")
			}
			root, err := gitx.RepoTopLevel()
			if err != nil {
				return err
			}
			identity, err := workspace.Resolve(execGitRunner{}, root, "")
			if err != nil || identity.Address == nil {
				return fmt.Errorf("local-check reclassify: verified workspace authority unavailable: %v", err)
			}
			repo, err := detectRepo()
			if err != nil {
				return err
			}
			d := localRunDeps{Root: root, Git: func(args ...string) (string, error) {
				out, err := mergeRunner.GitInDir(root, args...)
				return string(out), err
			}}
			logDir, err := localCheckLogDir(d)
			if err != nil {
				return err
			}
			if err := cmd.Context().Err(); err != nil {
				return err
			}
			store := suiteStoreRoot(logDir)
			if err := publishSuiteReclassification(store, id, repo, *identity.Address, reason, time.Now().UTC(), nil); err != nil {
				return err
			}
			remaining := int64(suiteAttemptMaxJSON)
			raw, err := boundedSuiteProofFile(filepath.Join(store, "attempts", id+".json"), suiteAttemptMaxJSON, &remaining, os.Open)
			if err != nil {
				return fmt.Errorf("correction published; exact rerun advice unavailable, inspect audit %s: %w", id, err)
			}
			a, err := decodeSuiteAttempt(raw)
			if err != nil || a.ID != id || a.Repo != repo {
				return errors.New("correction published; attempt moved before rerun advice, inspect its audit")
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "attempt %s: could-not-run; audit %s; no status changed\nrepair the environment, then sdlc local-check run --name %s --rev %s\n", id, filepath.Join(store, "reclassified", id+".json"), a.Check, a.SHA)
			return err
		}}
	cmd.Flags().StringVar(&id, "attempt", "", "exact published attempt ID (required)")
	cmd.Flags().StringVar(&reason, "reason", "", "required explanation retained in the immutable recovery audit")
	return markMutatingCommand(cmd)
}
