// localcheck_cmd.go — `sdlc local-check run` (#345): run declared local checks
// on a pushed head and record each as a commit status, ahead of `sdlc merge`.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

type localCheckFlags struct {
	Names []string
	All   bool
	Rev   string
}

// NewLocalCheckCmd is `sdlc local-check`; its one verb is `run`.
func NewLocalCheckCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "local-check", Args: cobra.NoArgs}
	var f localCheckFlags
	run := &cobra.Command{
		Use:           "run",
		Short:         "Run declared local checks on a pushed head and record them as commit statuses",
		Long:          renderLong("local-check"),
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLocalCheckCmd(commandContext(cmd.Context()), cmd.ErrOrStderr(), f)
		},
	}
	run.Flags().StringSliceVar(&f.Names, "name", nil, "declared check(s) to run (repeatable)")
	run.Flags().BoolVar(&f.All, "all", false, "run every declared check, in declaration order")
	run.Flags().StringVar(&f.Rev, "rev", "HEAD", "the commit to run on; it must be the tip of a branch on origin (a pushed PR head)")
	cmd.AddCommand(run)
	return cmd
}

// selectLocalChecks picks the checks a run names, refusing an unknown name.
func selectLocalChecks(decl localCheckDecl, names []string, all bool) ([]localCheck, error) {
	if all == (len(names) > 0) {
		return nil, errors.New("local-check run: give --name (one or more) or --all")
	}
	if all {
		return decl.Checks, nil
	}
	var out []localCheck
	for _, n := range names {
		c, ok := decl.check(n)
		if !ok {
			var known []string
			for _, k := range decl.Checks {
				known = append(known, k.Name)
			}
			return nil, fmt.Errorf("local-check run: %q is not declared in %s on origin/main (declared: %s)", n, localChecksPath, strings.Join(known, ", "))
		}
		out = append(out, c)
	}
	return out, nil
}

func runLocalCheckCmd(ctx context.Context, stderr io.Writer, f localCheckFlags) error {
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	git := func(args ...string) (string, error) {
		out, err := mergeRunner.GitInDir(root, args...)
		return string(out), err
	}
	// The declaration and the pushed-ness check both read origin, freshly.
	if out, err := git("fetch", "--quiet", "origin"); err != nil {
		return fmt.Errorf("local-check run: git fetch origin: %w\n%s", err, out)
	}
	decl, err := readLocalChecks(git, "origin/main")
	if err != nil {
		return err
	}
	// #332: the default suite is runnable by name, so `sdlc local-check run
	// --name suite` is the close suite.
	if decl, err = withDefaultSuiteContext(ctx, git, "origin/main", decl); err != nil {
		return err
	}
	if len(decl.Checks) == 0 {
		return fmt.Errorf("local-check run: origin/main declares no local checks (%s)", localChecksPath)
	}
	checks, err := selectLocalChecks(decl, f.Names, f.All)
	if err != nil {
		return err
	}
	out, err := git("rev-parse", "--verify", f.Rev+"^{commit}")
	if err != nil {
		return fmt.Errorf("local-check run: %s is not a commit: %w", f.Rev, err)
	}
	sha := strings.TrimSpace(out)
	if tips, _ := git("for-each-ref", "--points-at", sha, "--format=%(refname)", "refs/remotes/origin/"); strings.TrimSpace(tips) == "" {
		return fmt.Errorf("local-check run: %s is not the tip of a branch on origin; push it first (merge reads the statuses on a PR's pushed head)", abbrevSHA(sha))
	}
	repo, err := detectRepo()
	if err != nil {
		return err
	}
	pr := 0
	if branch, _ := git("branch", "--show-current"); strings.TrimSpace(branch) != "" {
		if n, err := ghClient.PRListForBranch(repo, strings.TrimSpace(branch)); err == nil {
			pr, _ = strconv.Atoi(strings.TrimSpace(n))
		}
	}
	d := localRunDeps{GH: ghClient, Repo: repo, Root: root, Git: git, PR: pr, Stderr: stderr}
	var failed []string
	for _, c := range checks {
		rec, err := runLocalCheck(ctx, d, c, sha)
		if err != nil {
			return err
		}
		if !rec.passed() {
			failed = append(failed, c.Name+" (log "+rec.LogPath+")")
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("local-check run: failed on %s: %s", abbrevSHA(sha), strings.Join(failed, "; "))
	}
	return nil
}
