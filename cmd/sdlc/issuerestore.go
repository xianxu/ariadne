// issuerestore.go — `sdlc issue restore` (#285): put main's version of
// published details back on this branch, the mechanical resolution of a
// transfer-guard refusal from a checkout that does not own the issue.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func newIssueRestoreCmd() *cobra.Command {
	var issues []int
	cmd := markMutatingCommand(&cobra.Command{
		Use:   "restore --issue N[,N…]",
		Short: "Put main's version of published issue details back on this branch (#285)",
		Long: "Undo this branch's change to published issue details, as the transfer guard\n" +
			"offers when a landing changes details from a checkout that does not own the\n" +
			"issue. For each named issue whose details the prospective merge with main\n" +
			"would change, the file is set to main's version (or removed, when main has\n" +
			"archived it), in one commit `#N: issue: restore main's details`. A dirty\n" +
			"details file is refused; with nothing to restore, nothing is committed.\n" +
			"To keep the edit instead, `sdlc claim --issue N` and land it as the owner.\n" +
			"In the owner's own checkout restore refuses (the edit is the owner's): merge\n" +
			"main and resolve the details instead.",
		Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			guardSpineRepo(cmd.ErrOrStderr())
			return runIssueRestore(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), issues)
		},
	})
	cmd.Flags().IntSliceVar(&issues, "issue", nil, "issue ID(s) whose details to restore: --issue 8 or --issue 8,9")
	return cmd
}

func runIssueRestore(ctx context.Context, stdout, _ io.Writer, nums []int) error {
	nums = claimIssues(&claimFlags{Issues: nums})
	if len(nums) == 0 {
		return errors.New("--issue N (or a list: --issue 8,9) is required")
	}
	if tracked, err := repositoryTracked(ctx, "."); err != nil {
		return err
	} else if !tracked {
		return errors.New("`issue restore` applies to issue tracker repositories, where the transfer guard runs")
	}
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	mainTip := view.Ref()
	changed, err := changedDetails(env, mainTip)
	if err != nil {
		return err
	}
	want := map[string]bool{}
	for _, n := range nums {
		want[fmt.Sprintf("%06d", n)] = true
	}
	// The guard judges the merge, where git follows an archive's rename; the
	// restore makes every copy of a refused issue's details equal main's —
	// wherever it lives on either side — so the merge keeps main's.
	names := map[string]bool{}
	var refs []string
	for _, c := range changed {
		if want[c.ID] && c.Verdict == verdictBehind {
			// This checkout owns the issue: the edit is the owner's own work,
			// which merging main keeps; restoring would drop it.
			return detailsRefusal(env, c)
		}
	}
	for _, c := range changed {
		if want[c.ID] && !names[path.Base(c.Path)] {
			names[path.Base(c.Path)] = true
			refs = append(refs, "#"+issue.CLIRef(c.ID))
		}
	}
	if len(names) == 0 {
		fmt.Fprintln(stdout, "nothing to restore: the named issues' details land as main has them")
		return nil
	}
	diff, err := env.gitRaw(nil, "diff", "--name-only", "--no-renames", "-z", mainTip, "HEAD")
	if err != nil {
		return err
	}
	var paths []string
	for _, p := range strings.Split(string(diff), "\x00") {
		if p != "" && names[path.Base(p)] {
			paths = append(paths, p)
		}
	}
	// A dirty file is refused unless it already holds main's version: an
	// interrupted restore, which the rerun finishes.
	for _, p := range paths {
		if err := restorable(env, mainTip, p); err != nil {
			return err
		}
	}
	for _, p := range paths {
		onMain, err := env.has(mainTip, p)
		if err != nil {
			return err
		}
		if onMain {
			_, err = env.git("checkout", mainTip, "--", p)
		} else {
			_, err = env.git("rm", "-q", "--ignore-unmatch", "--", p)
		}
		if err != nil {
			return err
		}
	}
	msg := strings.Join(refs, ",") + ": issue: restore main's details"
	if _, err := env.git(append([]string{"commit", "-q", "-m", msg, "--"}, paths...)...); err != nil {
		return err
	}
	for _, p := range paths {
		fmt.Fprintf(stdout, "restored %s to main's version\n", p)
	}
	return nil
}

// restorable refuses a details file with uncommitted changes, unless the
// working copy already is main's version (present or absent).
func restorable(env *trackerEnv, mainTip, p string) error {
	dirty, err := env.git("status", "--porcelain", "--", p)
	if err != nil || dirty == "" {
		return err
	}
	onMain, err := env.has(mainTip, p)
	if err != nil {
		return err
	}
	abs := filepath.Join(env.root, filepath.FromSlash(p))
	_, statErr := os.Lstat(abs)
	switch {
	case !onMain && errors.Is(statErr, os.ErrNotExist):
		return nil
	case onMain && statErr == nil:
		want, err := env.git("rev-parse", mainTip+":"+p)
		if err != nil {
			return err
		}
		if got, err := env.git("hash-object", "--", p); err != nil || got == want {
			return err
		}
	}
	return fmt.Errorf("uncommitted changes to %s, which restore would overwrite; commit or discard them first", p)
}
