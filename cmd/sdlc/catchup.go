// catchup.go — #303: before new work, catch the slot and its substrate
// dependencies up to origin/main. start-plan probes each checkout on the
// dependency path; when any trails origin/main it runs `weave refresh`, which owns
// the catch-up transaction (it advances nothing unless every checkout is clean and
// fast-forwardable, but keeps the updates it completed if a later step fails). A failed refresh refuses only when a checkout lacks code main had by the
// freshness cutoff (#269), so start-plan never refuses what close would accept.
package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/freshness"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// weaveRefresh is the catch-up seam. Tests replace it; TestMain stubs it so no
// test runs a real weave.
var weaveRefresh = runInstalledWeave
var weaveSettingsCheck = runInstalledWeaveSettings

// runInstalledWeave runs gate refresh with seeds and protected settings deferred, writing its output
// to out.
func runInstalledWeave(ctx context.Context, root string, out io.Writer) error {
	return runInstalledWeaveArgs(ctx, root, out, "refresh", "--planning")
}

func runInstalledWeaveSettings(ctx context.Context, root string, out io.Writer) error {
	return runInstalledWeaveArgs(ctx, root, out, "settings")
}

func runInstalledWeaveArgs(ctx context.Context, root string, out io.Writer, args ...string) error {
	bin, err := exec.LookPath("weave")
	if err != nil {
		return fmt.Errorf("find installed weave: %w", err)
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = root, out, out
	cmd.Env = weaveDependencyCheckEnv(cmd)
	return cmd.Run()
}

// SDLC may refresh repository state, but machine dependencies belong to the
// operator. Keep this permission local to each child, including nested weave.
func weaveDependencyCheckEnv(cmd *exec.Cmd) []string {
	env := cmd.Environ()
	out := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if !strings.HasPrefix(entry, "WEAVE_DEPENDENCIES_CHECK_ONLY=") {
			out = append(out, entry)
		}
	}
	return append(out, "WEAVE_DEPENDENCIES_CHECK_ONLY=1")
}

// catchUpFetchTimeout bounds each probe's fetch: an unresponsive remote leaves
// that checkout unchecked rather than hanging start-plan.
var catchUpFetchTimeout = time.Minute

// catchUpBeforeWork brings root and its substrate chain up to origin/main. It
// returns nil when planning may go on. It refuses an obsolete running build, a
// partial refresh failure, or a checkout that cannot catch up and lacks old code.
func catchUpBeforeWork(ctx context.Context, stdout io.Writer, root string, issueID int) error {
	unchecked := func(dir, why string) {
		cwarn(stdout, fmt.Sprintf("catch-up: could not check %s against origin/main (%s); NOT verified as current", dir, why))
	}
	present, absent := substrateWalk(root)
	for _, dir := range absent {
		unchecked(dir, "declared in construct/deps but not checked out")
	}
	repos := append([]string{root}, present...)
	var behind []string
	heads := map[string]string{}
	for _, dir := range repos {
		n, why := behindOrigin(ctx, dir)
		switch {
		case why != "":
			unchecked(dir, why)
		case n > 0:
			behind = append(behind, dir)
		}
		heads[dir] = headOf(ctx, dir)
	}
	if len(behind) == 0 {
		if err := weaveSettingsCheck(ctx, root, stdout); err != nil {
			cwarn(stdout, fmt.Sprintf("settings check unavailable (%s): run weave refresh --substrate from your shell", firstLine(err.Error())))
		}
		return nil
	}
	rerun := fmt.Sprintf("sdlc start-plan --issue %d", issueID)
	cinfo(stdout, fmt.Sprintf("catch-up: behind origin/main: %s — running `weave refresh`", strings.Join(behind, ", ")))
	binary := runningSDLC()
	sourceBefore := ""
	if binary.Source != "" {
		sourceBefore = headOf(ctx, binary.Source)
	}
	refreshErr := weaveRefresh(ctx, root, stdout)
	// Refresh keeps the updates it completed before an error, so what moved is
	// read from the checkouts, not from its exit status.
	var moved []string
	for _, dir := range repos {
		if headOf(ctx, dir) != heads[dir] {
			moved = append(moved, dir)
		}
	}
	sourceAfter := ""
	if sourceBefore != "" {
		sourceAfter = headOf(ctx, binary.Source)
	}
	if len(moved) > 0 || sourceAfter != sourceBefore {
		if refreshErr != nil {
			return fmt.Errorf("caught up %s before `weave refresh` failed (%v); nothing was branched or started. Repair refresh, then re-run `%s`", strings.Join(moved, ", "), refreshErr, rerun)
		}
		if sourceAfter != "" && sourceAfter != sourceBefore && binary.Revision != "" && binary.Revision != sourceAfter {
			return fmt.Errorf("caught up %s to origin/main; this sdlc was built before that (source %s, revision %s, now %s), so nothing was branched or started.\n      Re-run `%s` with a refreshed build", strings.Join(moved, ", "), binary.Source, binary.Revision, sourceAfter, rerun)
		}
		if binary.Revision == "" || sourceBefore == "" || sourceAfter == "" {
			cwarn(stdout, "catch-up: running sdlc source/revision unavailable; "+binary.version())
		}
		return nil
	}
	if refreshErr == nil {
		return fmt.Errorf("`weave refresh` succeeded but left %s behind origin/main; nothing was branched or started. Inspect them, then re-run `%s`", strings.Join(behind, ", "), rerun)
	}
	loc, zone, err := freshness.LoadLocation(root)
	if err != nil {
		return fmt.Errorf("`weave refresh` failed (%v), and the freshness zone is unreadable: %w", refreshErr, err)
	}
	cutoff := freshness.Cutoff(integrateNow(), loc)
	var stale []string
	for _, dir := range behind {
		r := freshness.Check(checkoutGit(ctx, dir), "refs/remotes/origin/main", "HEAD", cutoff, publishGateHasCodeSurface)
		switch r.State {
		case freshness.Stale:
			stale = append(stale, fmt.Sprintf("%s (%s)", dir, strings.Join(r.Code, ", ")))
		case freshness.Inconclusive:
			cwarn(stdout, fmt.Sprintf("catch-up: could not tell what %s lacks (%s); NOT verified as current", dir, r.Why))
		}
	}
	if len(stale) > 0 {
		return fmt.Errorf("`weave refresh` could not catch up (%v); its output above names what is dirty or diverged.\n"+
			"      These lack code origin/main had by %s (%s): %s.\n"+
			"      Commit, publish or reconcile that, run `weave refresh` in %s, then re-run `%s`. Nothing was branched or started",
			refreshErr, cutoff.Format("Mon Jan 2 15:04 MST"), zone, strings.Join(stale, "; "), root, rerun)
	}
	cwarn(stdout, fmt.Sprintf("catch-up: `weave refresh` failed (%v), but what these lack landed after %s or is docs only; planning on the current checkouts — run `weave refresh` when they are clean",
		refreshErr, cutoff.Format("Mon Jan 2 15:04 MST")))
	return nil
}

// behindOrigin fetches dir's origin main and counts the commits HEAD lacks. why
// is non-empty when the probe could not decide, a checkout with no origin
// included: it is reported unchecked, never current.
func behindOrigin(ctx context.Context, dir string) (n int, why string) {
	git := checkoutGit(ctx, dir)
	if _, err := git("remote", "get-url", "origin"); err != nil {
		return 0, "no origin remote"
	}
	fetchCtx, cancel := context.WithTimeout(ctx, catchUpFetchTimeout)
	defer cancel()
	if _, err := checkoutGit(fetchCtx, dir)("fetch", "--quiet", "origin", "+refs/heads/main:refs/remotes/origin/main"); err != nil {
		return 0, "fetch origin main: " + err.Error()
	}
	out, err := git("rev-list", "--count", "HEAD..refs/remotes/origin/main", "--")
	if err != nil {
		return 0, "compare HEAD with origin/main: " + err.Error()
	}
	n, err = strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, "compare HEAD with origin/main: " + err.Error()
	}
	return n, ""
}

// headOf is dir's HEAD commit, or "" when it cannot be read.
func headOf(ctx context.Context, dir string) string {
	out, err := checkoutGit(ctx, dir)("rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// checkoutGit adapts gitx's cancellable runner to freshness.Git for one checkout,
// folding git's stderr into the error so a refusal names the cause.
func checkoutGit(ctx context.Context, dir string) freshness.Git {
	return func(args ...string) ([]byte, error) {
		out, errOut, err := gitx.RunGitContext(ctx, append([]string{"-C", dir}, args...)...)
		if err != nil {
			if msg := strings.TrimSpace(string(errOut)); msg != "" {
				err = fmt.Errorf("%w: %s", err, firstLine(msg))
			}
			return out, err
		}
		return out, nil
	}
}
