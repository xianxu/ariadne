// integrategate.go — the integrate gate (#269): before a boundary review, the branch
// must contain the code that landed on main before the most recent noon
// (internal/freshness, ariadne-robustness-1 requirement 6). Otherwise the review and
// the suite describe a tree that never ships. Integrating is cheap since #304: a
// merge or rebase of main never widens a review window.
package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/freshness"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// integrateNow is the clock the gate's cutoff reads; tests pin it.
var integrateNow = time.Now

// integrateFetchTimeout bounds the legacy-repository fetch: an unresponsive remote
// makes the check inconclusive rather than hanging the close.
var integrateFetchTimeout = time.Minute

// integrateGate runs the gate for computeClose: a stale branch refuses through die,
// a check that could not decide warns, and nothing else is printed for a fresh one.
func integrateGate(ctx context.Context, stderr io.Writer, f *closeFlags, issueStr string) {
	if f.skip("integrate") {
		cwarn(stderr, "--no-integrate (or --force): skipping the integrate gate — rationale in --verified")
		return
	}
	root := gitx.Capture("rev-parse", "--show-toplevel")
	loc, zone, err := freshness.LoadLocation(root)
	if err != nil {
		die(stderr, fmt.Sprintf("integrate gate: %v — fix it, or pass --no-integrate (or --force) with the reason in --verified", err))
	}
	mainRef, mainName, why := integrateMainRef(ctx, root)
	if mainRef == "" {
		cwarn(stderr, formatIntegrateUnknown(why))
		return
	}
	r := freshness.Check(gitx.RunGit, mainRef, "HEAD", freshness.Cutoff(integrateNow(), loc), publishGateHasCodeSurface)
	switch r.State {
	case freshness.Inconclusive:
		cwarn(stderr, formatIntegrateUnknown(r.Why))
	case freshness.DocsOnly:
		cinfo(stderr, fmt.Sprintf("integrate gate: %d doc-only file(s) on %s since this branch's base; no code surface to integrate", r.Docs, mainName))
	case freshness.Stale:
		die(stderr, formatIntegrateRefusal(r, mainName, zone, integrateBranchPushed(mainName), rerunCmd(issueStr, f.Milestone, "")))
	}
}

// integrateMainRef is main freshly fetched: the review window's main in a tracker
// repository (one cached fetch, so the gate and the window agree), else the trunk
// ref after fetching its remote. A local-only main needs no fetch. ref is what the
// probe reads (a pinned commit in a tracker repository); name is the ref a person
// merges or rebases onto, which that fetch just moved to ref. ref is "" with a
// reason when main could not be brought up to date.
func integrateMainRef(ctx context.Context, root string) (ref, name, why string) {
	trunk := gitx.TrunkRef()
	if _, tracked, err := tracker.ReadCutoverMarker(root); err == nil && tracked {
		ref, stale := reviewMainRef(ctx)
		switch {
		case stale != "":
			return "", "", stale
		case ref == "":
			return "", "", "the issue tracker gave no main"
		case trunk == "":
			return ref, ref, ""
		}
		return ref, trunk, ""
	}
	if trunk == "" {
		return "", "", "no main branch found"
	}
	remote, branch, ok := strings.Cut(trunk, "/")
	if !ok || gitx.Capture("remote", "get-url", remote) == "" {
		return trunk, trunk, "" // a local main is already current
	}
	fetchCtx, cancel := context.WithTimeout(ctx, integrateFetchTimeout)
	defer cancel()
	if _, errOut, err := gitx.RunGitContext(fetchCtx, "fetch", "--quiet", remote, branch); err != nil {
		return "", "", fmt.Sprintf("fetch %s %s failed: %v %s", remote, branch, err, firstLine(strings.TrimSpace(string(errOut))))
	}
	return trunk, trunk, ""
}

// integrateBranchPushed reports whether the current branch has a counterpart on
// main's remote (or an upstream other than main), so the remedy is a merge: a rebase would need a
// force-push, which fights the publish's local/remote/PR head check.
func integrateBranchPushed(mainRef string) bool {
	// An upstream that is main itself (a branch made with `-b x origin/main`) says
	// nothing about whether this branch was pushed.
	if up := gitx.Capture("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}"); up != "" && up != mainRef {
		return true
	}
	branch := gitx.Capture("symbolic-ref", "-q", "--short", "HEAD")
	remote, _, ok := strings.Cut(mainRef, "/")
	return ok && branch != "" && gitx.Capture("rev-parse", "-q", "--verify", "refs/remotes/"+remote+"/"+branch) != ""
}

// formatIntegrateUnknown is the inconclusive line: never phrased as up to date.
func formatIntegrateUnknown(why string) string {
	return "integrate gate: could not check whether this branch has main's recent code (" + why +
		"); NOT verified as up to date — integrate main before publishing"
}

const integrateListCap = 8

// formatIntegrateRefusal names what is missing, the cutoff it was measured at and
// the remedy by push state. Pure.
func formatIntegrateRefusal(r freshness.Result, mainRef, zone string, pushed bool, rerun string) string {
	code := r.Code
	more := ""
	if len(code) > integrateListCap {
		more = fmt.Sprintf(" (+%d more)", len(code)-integrateListCap)
		code = code[:integrateListCap]
	}
	remedy := "git rebase " + mainRef + "    (the branch is not pushed)"
	if pushed {
		remedy = "git merge " + mainRef + "    (the branch is pushed: a rebase would need a force-push)"
	}
	return fmt.Sprintf("integrate gate: this branch lacks code that landed on %s by %s (%s; %s): %s%s.\n"+
		"  A review and suite run now would describe a tree that never ships. Integrate first:\n"+
		"    %s\n"+
		"  Commit local edits first, re-run the test suite, then re-run `%s`.\n"+
		"  (Integrating main never widens the review window, #304.)\n"+
		"  Or pass --no-integrate (or --force) with the reason in --verified.",
		mainRef, r.Cutoff.Format("Mon Jan 2 15:04 MST"), shortSHA(r.Required), zone,
		strings.Join(code, ", "), more, remedy, rerun)
}
