// reviewwindowplan.go — #304: every boundary-review window is defined on the BRANCH
// PATCH, diff(merge-base(main, HEAD), HEAD), never on a commit range. Merging or
// rebasing main changes the commits, not the patch, so neither widens a window.
//
//   - Whole-issue close reviews the branch patch.
//   - A milestone reviews the INTERDIFF since the last finalized review: the reviewed
//     head H_r (stamped on its ledger round, D5) replayed onto today's main by
//     gitx.RebasedReviewedBase gives a synthetic base S, and diff(S, HEAD) is exactly
//     this branch's work since that review — conflict resolutions included, main's
//     changes excluded.
//
// planReviewWindow is the PURE decision over gathered facts (ARCH-PURE); the git and
// ledger reads live in gatherWindowFacts. The atlas gate, the review manifest, the
// trailer and the printed line all spend the one reviewWindow it returns (ARCH-DRY,
// the #58 "same window" property).
package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

type windowKind int

const (
	windowNone           windowKind = iota // no commit anchors a window (no #N commit yet)
	windowBranchPatch                      // whole issue, or a first milestone
	windowInterdiff                        // since the last finalized review
	windowBranchFallback                   // a stamped review exists but cannot be replayed
)

// windowFacts is everything the planner needs, gathered by gatherWindowFacts.
type windowFacts struct {
	Milestone   string
	MainBase    string // merge-base(main, HEAD); "" on main / no divergence
	BranchStart string // the issue's first-commit parent (on-main fallback)
	Head        string

	// The last finalized review outside this boundary (#304 D6), from the ledger.
	Reviewed      string
	PriorBoundary string
	Rebased       string   // S: H_r replayed onto MainBase (or H_r itself on main)
	Conflicted    []string // paths where main rewrote lines the review read
	RebaseErr     string   // why H_r could not be replayed ("" when it was)

	// A pre-#304 ledger has no stamped round: the trailer-grep boundary (D6 fallback).
	LegacyTrailerBase string
}

// reviewWindow is the planned window: Base is what diffs run from; HumanBase is what a
// person reads (never the synthetic S, which means nothing and is collected by gc).
type reviewWindow struct {
	Kind          windowKind
	Base          string
	HumanBase     string
	Head          string
	MainBase      string
	PriorBoundary string
	Conflicted    []string
	Note          string // shown with the window (fallback reason, legacy source)
	IssueCommits  int    // first-parent, merges excluded: the issue's own commits
	Files         int
}

// planReviewWindow decides the window. Pure and total: every fact combination maps to
// exactly one kind.
func planReviewWindow(f windowFacts) reviewWindow {
	w := reviewWindow{Head: f.Head, MainBase: f.MainBase}
	branchPatch := func(kind windowKind, note string) reviewWindow {
		w.Kind, w.Note = kind, note
		w.Base = f.MainBase
		if w.Base == "" {
			w.Base = f.BranchStart
		}
		if w.Base == "" {
			w.Kind = windowNone
		}
		w.HumanBase = w.Base
		return w
	}
	if f.Milestone == "" {
		return branchPatch(windowBranchPatch, "")
	}
	switch {
	case f.Reviewed != "" && f.RebaseErr == "" && f.Rebased != "":
		w.Kind, w.Base, w.HumanBase = windowInterdiff, f.Rebased, f.Reviewed
		w.PriorBoundary, w.Conflicted = f.PriorBoundary, f.Conflicted
		if len(f.Conflicted) > 0 {
			w.Note = "includes conflict resolutions in " + strings.Join(f.Conflicted, ", ")
		}
		return w
	case f.Reviewed != "":
		return branchPatch(windowBranchFallback, f.RebaseErr)
	case f.LegacyTrailerBase != "":
		w.Kind, w.Base, w.HumanBase = windowInterdiff, f.LegacyTrailerBase, f.LegacyTrailerBase
		w.Note = "pre-#304 boundary from the Review-Verdict trailer"
		return w
	default:
		return branchPatch(windowBranchPatch, "")
	}
}

// formatReviewWindow renders the line close prints (D10). It shares no vocabulary with
// any refusal: gatesig classifies transcripts by substring (#172). Pure.
func formatReviewWindow(w reviewWindow) string {
	files := fmt.Sprintf("%d file(s)", w.Files)
	var line string
	switch w.Kind {
	case windowNone:
		return "review window: no commit anchors it yet"
	case windowBranchPatch:
		line = fmt.Sprintf("branch patch vs main@%s: %d issue commit(s), %s", abbrevSHA(w.Base), w.IssueCommits, files)
	case windowInterdiff:
		since := "the previous review"
		if w.PriorBoundary != "" {
			since = w.PriorBoundary + " review"
		}
		line = fmt.Sprintf("interdiff since %s (%s)", since, abbrevSHA(w.HumanBase))
		if w.MainBase != "" && w.Base != w.HumanBase {
			line += fmt.Sprintf(", rebased onto main@%s", abbrevSHA(w.MainBase))
		}
		line += ": " + files
	case windowBranchFallback:
		line = fmt.Sprintf("branch patch vs main@%s (the last review cannot be replayed): %d issue commit(s), %s",
			abbrevSHA(w.Base), w.IssueCommits, files)
	}
	if w.Note != "" {
		line += " — " + w.Note
	}
	return line
}

// planBoundaryWindow is the IO shell: it gathers the facts, plans, and measures.
func planBoundaryWindow(ctx context.Context, issueStr, milestone, issuePath, plansDir string) reviewWindow {
	f := gatherWindowFacts(ctx, issueStr, milestone, issuePath, plansDir)
	w := planReviewWindow(f)
	if w.Kind != windowNone && isResolvedSHA(w.Head) {
		if names, err := gitx.DiffNames(w.Base, w.Head); err == nil {
			w.Files = len(names)
		}
		if w.Kind != windowInterdiff {
			if n, err := strconv.Atoi(gitx.Capture("rev-list", "--count", "--first-parent", "--no-merges", w.Base+".."+w.Head)); err == nil {
				w.IssueCommits = n
			}
		}
	}
	return w
}

func gatherWindowFacts(ctx context.Context, issueStr, milestone, issuePath, plansDir string) windowFacts {
	f := windowFacts{Milestone: milestone, Head: gitx.Capture("rev-parse", "HEAD")}
	if f.Head == "" {
		f.Head = "HEAD"
	}
	f.MainBase = reviewMainBase(ctx, f.Head)
	if f.MainBase == "" {
		f.BranchStart = branchStartByIssue(issueStr)
	}
	if milestone == "" {
		return f
	}
	if sha, boundary, ok := latestReviewedFor(issuePath, plansDir, milestone); ok {
		f.Reviewed, f.PriorBoundary = sha, boundary
		switch {
		case gitx.Capture("rev-parse", "--verify", "-q", sha+"^{commit}") == "":
			f.RebaseErr = fmt.Sprintf("reviewed head %s is not in this repository", abbrevSHA(sha))
		case f.MainBase == "":
			f.Rebased = sha // on main: no main to replay onto; diff from the reviewed head itself
		default:
			// MainBase is an ancestor of HEAD, so passing it as "main" gives
			// merge-base(main, HEAD) = MainBase and merge-base(main, H_r) = H_r's own
			// base (the branch only ever integrates main forward).
			s, conflicted, err := gitx.RebasedReviewedBase(f.MainBase, sha)
			if err != nil {
				f.RebaseErr = err.Error()
			} else {
				f.Rebased, f.Conflicted = s, conflicted
			}
		}
		return f
	}
	// D6 fallback, removable once no open issue predates #304 (no stamped round exists).
	f.LegacyTrailerBase = previousReviewBoundary(issuePath)
	return f
}

// latestReviewedFor reads where review last stopped outside this boundary. An unreadable
// ledger over-covers (branch patch) rather than refusing: the review is still worth
// running, matching boundaryPriorFindings.
func latestReviewedFor(issuePath, plansDir, milestone string) (sha, boundary string, ok bool) {
	if issuePath == "" || plansDir == "" {
		return "", "", false
	}
	l, err := readBoundaryGateLedger(plansDir, filepath.Base(issuePath), issueIDFromPath(issuePath))
	if err != nil {
		return "", "", false
	}
	return gatestate.LatestReviewed(l, milestone)
}

// reviewMainBase is merge-base(main, HEAD) against the ONE main every branch-patch
// consumer measures from (#304 D11), or "" on main / with no divergence. In an issue
// tracker repository main is freshly fetched (the view the publish gate already uses),
// so a stale tracking ref cannot put the base behind a main the branch already merged.
// Elsewhere it is gitx.MergeBaseWithMain, which has no fetch source — as before.
func reviewMainBase(ctx context.Context, head string) string {
	ref := reviewMainRef(ctx)
	if ref == "" {
		return gitx.MergeBaseWithMain()
	}
	out, err := gitx.RunGit("merge-base", "--all", ref, "HEAD")
	if err != nil {
		return ""
	}
	bases := strings.Fields(string(out))
	if len(bases) != 1 || bases[0] == head {
		return "" // criss-cross (RebasedReviewedBase names it) or no divergence
	}
	return bases[0]
}

// reviewMainRef is main as freshly fetched in an issue tracker repository, else "".
// Fetched at most once per process and checkout.
func reviewMainRef(ctx context.Context) string {
	root := gitx.Capture("rev-parse", "--show-toplevel")
	reviewMainMu.Lock()
	defer reviewMainMu.Unlock()
	if ref, ok := reviewMainCache[root]; ok {
		return ref
	}
	ref := ""
	if _, tracked, err := tracker.ReadCutoverMarker(root); err == nil && tracked {
		if env, err := openTracker(ctx); err == nil {
			if view, err := env.main.Snapshot(); err == nil {
				ref = view.Ref()
			}
		}
	}
	reviewMainCache[root] = ref
	return ref
}

var (
	reviewMainMu    sync.Mutex
	reviewMainCache = map[string]string{}
)
