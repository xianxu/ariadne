// localcheck_cmd.go — `sdlc local-check run` (#345): run declared local checks
// on a pushed head and record each as a commit status, ahead of `sdlc merge`.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

type localCheckFlags struct {
	Names []string
	All   bool
	Rev   string
}

// NewLocalCheckCmd is `sdlc local-check`: `run`, and `cancel` (#400).
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
	run.Flags().BoolVar(&f.All, "all", false, "run every head-scoped check, in declaration order")
	run.Flags().StringVar(&f.Rev, "rev", "HEAD", "the commit to run on; it must be the tip of a branch on origin (a pushed PR head)")
	var cancelName, cancelRun string
	cancel := &cobra.Command{
		Use:           "cancel",
		Short:         "Stop this checkout's running local checks (never another slot's)",
		Long:          renderLong("local-check"),
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Flags().Changed("run") && cancelRun == "" {
				return errors.New("local-check cancel: --run requires a nonempty run ID")
			}
			return runLocalCheckCancel(cmd.ErrOrStderr(), cancelName, cancelRun)
		},
	}
	cancel.Flags().StringVar(&cancelName, "name", "", "stop all this checkout's runs of this check (default: every check)")
	cancel.Flags().StringVar(&cancelRun, "run", "", "stop exactly one leased run ID printed at run start (exclusive with --name)")
	cancel.MarkFlagsMutuallyExclusive("name", "run")
	var rp republishFlags
	republish := &cobra.Command{
		Use:           "republish",
		Short:         "Post a completed check's saved result on its exact commit, without rerunning it",
		Long:          renderLong("local-check"),
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLocalCheckRepublish(commandContext(cmd.Context()), cmd.ErrOrStderr(), rp)
		},
	}
	republish.Flags().StringVar(&rp.Name, "name", "", "the check whose saved result to post (required)")
	republish.Flags().StringVar(&rp.Rev, "rev", "HEAD", "the commit the saved attempt ran on")
	republish.Flags().StringVar(&rp.Attempt, "attempt", "", "a specific attempt id (default: the newest that verifies)")
	cmd.AddCommand(run, cancel, republish, newLocalCheckReclassifyCmd(), newLocalCheckResumeCmd(), newLocalCheckReconcileCmd(defaultLandingDeps()))
	return cmd
}

// localCheckCancelWait bounds how long cancel waits for a stopped run to unwind.
const localCheckCancelWait = 2 * time.Minute

// runLocalCheckCancel stops the runs this checkout started, found by their
// leases (#400): never a process picked from the global process table.
func runLocalCheckCancel(stderr io.Writer, name, runID string) error {
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	logDir, err := localCheckLogDir(localRunDeps{Root: root, Git: func(args ...string) (string, error) {
		out, err := mergeRunner.GitInDir(root, args...)
		return string(out), err
	}})
	if err != nil {
		return err
	}
	_, err = cancelLocalChecks(localCheckRunsDir(logDir), root, name, runID, stderr, localCheckCancelWait)
	return err
}

// splitAffectedName separates the affected suite (#473), which runs through
// its planner, from the declared checks a run names; --all never includes it
// and never combines with --name.
func splitAffectedName(names []string, all bool) (declared []string, affected bool, err error) {
	if all && len(names) > 0 {
		return nil, false, errors.New("local-check run: give --name (one or more) or --all")
	}
	declared = slices.DeleteFunc(slices.Clone(names), func(n string) bool { return n == affectedCheckName })
	return declared, len(declared) < len(names), nil
}

// selectLocalChecks picks the checks a run names, refusing an unknown name.
func selectLocalChecks(decl localCheckDecl, names []string, all bool) ([]localCheck, error) {
	if all == (len(names) > 0) {
		return nil, errors.New("local-check run: give --name (one or more) or --all")
	}
	if all {
		var out []localCheck
		for _, c := range decl.Checks {
			if c.RunScope != "landing" {
				out = append(out, c)
			}
		}
		return out, nil
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
	mainHead, err := git("rev-parse", "--verify", "origin/main^{commit}")
	if err != nil {
		return fmt.Errorf("local-check run: pin origin/main: %w", err)
	}
	mainHead = strings.TrimSpace(mainHead)
	out, err := git("rev-parse", "--verify", f.Rev+"^{commit}")
	if err != nil {
		return fmt.Errorf("local-check run: %s is not a commit: %w", f.Rev, err)
	}
	sha := strings.TrimSpace(out)
	tips, err := git("for-each-ref", "--points-at", sha, "--format=%(refname)", "refs/remotes/origin/")
	if err != nil {
		return fmt.Errorf("local-check run: read pushed tips: %w", err)
	}
	if strings.TrimSpace(tips) == "" {
		return fmt.Errorf("local-check run: %s is not the tip of a branch on origin; push it first (merge reads the statuses on a PR's pushed head)", abbrevSHA(sha))
	}
	decl, err := readLocalChecks(git, mainHead)
	if err != nil {
		return err
	}
	// #332: the default suite is runnable by name, so `sdlc local-check run
	// --name suite` is the close suite.
	var bootstrap bool
	if decl, bootstrap, err = withSubjectDefaultSuiteContext(ctx, git, mainHead, sha, decl); err != nil {
		return err
	}
	if len(decl.Checks) == 0 {
		if decl.Version != 0 {
			return fmt.Errorf("local-check run: origin/main declares no local checks (%s); bootstrap requires no declaration on main; establish its suite on main, then run sdlc local-check run --name suite", localChecksPath)
		}
		return fmt.Errorf("local-check run: origin/main declares no local checks (%s), and the pushed subject has no bootstrap test target; add a make test target, push, then run sdlc local-check run --name suite", localChecksPath)
	}
	if bootstrap {
		fmt.Fprintf(stderr, "local-check: bootstrap suite on %s: required make test; head declarations stay inactive until on main\n", abbrevSHA(sha))
	}
	// The affected suite (#473) is declared in its own file and runs through
	// its planner; every other name is a declared check.
	names, wantAffected, err := splitAffectedName(f.Names, f.All)
	if err != nil {
		return err
	}
	var checks []localCheck
	if len(names) > 0 || !wantAffected {
		if checks, err = selectLocalChecks(decl, names, f.All); err != nil {
			return err
		}
	}
	var affectedCfg affected.Config
	if wantAffected {
		cfg, ok, err := readAffectedConfig(git, mainHead)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("local-check run: origin/main declares no %s; run --name %s", affectedConfigPath, suiteCheckName)
		}
		affectedCfg = cfg
	}
	if err := validateManualLandingTarget(checks, sha, mainHead); err != nil {
		return err
	}
	repo, err := detectRepo()
	if err != nil {
		return err
	}
	pr := currentBranchPR(git, repo)
	d := localRunDeps{GH: ghClient, Repo: repo, Root: root, Git: git, CleanupGit: localCheckCleanupGit(ctx, mergeRunner, root), PR: pr, Stderr: stderr, FlakeRegistry: gateFlakeRegistry(root)}
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
	if wantAffected {
		suite, ok := decl.check(suiteCheckName)
		if !ok {
			return fmt.Errorf("local-check run: %s needs a %s check to fall back on for a full selection", affectedCheckName, suiteCheckName)
		}
		rec, err := runAffectedCheck(ctx, d, productionAffectedDeps(d, mainHead, affectedCfg), affectedCfg, suite, sha)
		if err != nil {
			return err
		}
		if !rec.passed() {
			failed = append(failed, rec.Name+" (log "+rec.LogPath+")")
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("local-check run: failed on %s: %s", abbrevSHA(sha), strings.Join(failed, "; "))
	}
	return nil
}

type republishFlags struct{ Name, Rev, Attempt string }

// runLocalCheckRepublish posts a saved attempt's outcome (#431): the recovery
// for a completed run whose status post could not reach GitHub.
func runLocalCheckRepublish(ctx context.Context, stderr io.Writer, f republishFlags) error {
	if f.Name == "" {
		return errors.New("local-check republish: give --name")
	}
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	git := func(args ...string) (string, error) {
		out, err := mergeRunner.GitInDir(root, args...)
		return string(out), err
	}
	out, err := git("rev-parse", "--verify", f.Rev+"^{commit}")
	if err != nil {
		return fmt.Errorf("local-check republish: %s is not a commit: %w", f.Rev, err)
	}
	sha := strings.TrimSpace(out)
	repo, err := detectRepo()
	if err != nil {
		return err
	}
	d := localRunDeps{GH: ghClient, Repo: repo, Root: root, Git: git, Stderr: stderr, FlakeRegistry: gateFlakeRegistry(root)}
	logDir, err := localCheckLogDir(d)
	if err != nil {
		return err
	}
	rec, _, err := loadSavedAttemptCurrent(ctx, d, suiteStoreRoot(logDir), sha, f.Name, f.Attempt, hostName())
	if err != nil {
		return fmt.Errorf("local-check republish: %w", err)
	}
	if rec.ResumeParent == "" && (rec.Outcome == flakes.KnownFlakePassed || rec.Outcome == flakes.KnownFlakeAccept) {
		// #394: a saved known-flake pass still needs its flake issue open and
		// the operator's policy on main to allow it.
		c, err := declaredCheck(ctx, git, f.Name, sha)
		if err != nil {
			return fmt.Errorf("local-check republish: %w", err)
		}
		if next := revalidateKnownFlake(ctx, d, c, rec.knownFlakeOutcome()); next.Kind != rec.Outcome {
			return fmt.Errorf("local-check republish: the saved known-flake pass on %s no longer holds: %s", abbrevSHA(sha), next.Reason)
		}
	}
	d.PR = currentBranchPR(git, repo)
	if err := postLocalCheckRecord(ctx, d, rec); err != nil {
		return fmt.Errorf("local-check republish: %s on %s was NOT recorded: %w", f.Name, abbrevSHA(sha), err)
	}
	fmt.Fprintf(stderr, "==> local check %s: saved result (%s) recorded as %s on %s\n", f.Name, rec.statusDescription(), localStatusContext(f.Name), abbrevSHA(sha))
	return nil
}

// currentBranchPR is the open PR for the checked-out branch, 0 when none is
// known (a failed check then posts no PR comment).
func currentBranchPR(git func(args ...string) (string, error), repo string) int {
	branch, _ := git("branch", "--show-current")
	if strings.TrimSpace(branch) == "" {
		return 0
	}
	n, err := ghClient.PRListForBranch(repo, strings.TrimSpace(branch))
	if err != nil {
		return 0
	}
	pr, _ := strconv.Atoi(strings.TrimSpace(n))
	return pr
}

// declaredCheck reads current main policy for one original saved subject.
// Only its default bootstrap target may come from that exact subject; the
// current checkout and the subject's declaration have no policy authority.
func declaredCheck(ctx context.Context, git func(args ...string) (string, error), name, subject string) (localCheck, error) {
	if out, err := git("fetch", "--quiet", "origin"); err != nil {
		return localCheck{}, fmt.Errorf("git fetch origin: %w\n%s", err, out)
	}
	out, err := git("rev-parse", "--verify", "origin/main^{commit}")
	if err != nil {
		return localCheck{}, fmt.Errorf("pin current origin/main: %w", err)
	}
	base := strings.TrimSpace(out)
	if base == "" {
		return localCheck{}, fmt.Errorf("pin current origin/main: git returned no commit")
	}
	decl, err := readLocalChecks(git, base)
	if err != nil {
		return localCheck{}, err
	}
	decl, _, err = withSubjectDefaultSuiteContext(ctx, git, base, subject, decl)
	if err != nil {
		return localCheck{}, err
	}
	if name == affectedCheckName {
		// #473 M2 BR-9: affected is declared in its own file, with the suite's
		// known-flake policy; resolving it here keeps a saved affected pass
		// republishable.
		cfg, ok, err := readAffectedConfig(git, base)
		if err != nil {
			return localCheck{}, err
		}
		suite, hasSuite := decl.check(suiteCheckName)
		if !ok || !hasSuite {
			return localCheck{}, fmt.Errorf("origin/main declares no %s with a %s to fall back on", affectedConfigPath, suiteCheckName)
		}
		return affectedCheckDecl(cfg, suite), nil
	}
	for _, c := range decl.Checks {
		if c.Name == name {
			return c, nil
		}
	}
	return localCheck{}, fmt.Errorf("origin/main declares no local check %q", name)
}
