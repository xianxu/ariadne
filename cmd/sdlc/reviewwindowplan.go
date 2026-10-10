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
	MainBaseErr   string   // merge-base(main, HEAD) is not single (criss-cross, unrelated)
	MainStale     string   // main could not be fetched; measured as last fetched
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
		w.Kind, w.Note = kind, joinNotes(note, f.MainStale)
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
	if f.MainBaseErr != "" {
		return branchPatch(windowBranchFallback, f.MainBaseErr)
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
		w.Note = joinNotes(w.Note, f.MainStale)
		return w
	case f.Reviewed != "":
		return branchPatch(windowBranchFallback, f.RebaseErr)
	case f.LegacyTrailerBase != "":
		w.Kind, w.Base, w.HumanBase = windowInterdiff, f.LegacyTrailerBase, f.LegacyTrailerBase
		w.Note = joinNotes("pre-#304 boundary from the Review-Verdict trailer", f.MainStale)
		return w
	default:
		return branchPatch(windowBranchPatch, "")
	}
}

func joinNotes(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	}
	return a + "; " + b
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
	base, stale, err := reviewMainBase(ctx, f.Head)
	f.MainBase, f.MainStale = base, stale
	if err != nil {
		f.MainBaseErr = err.Error()
	}
	if f.MainBase == "" {
		f.BranchStart = branchStartByIssue(issueStr)
	}
	if milestone == "" || f.MainBaseErr != "" {
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
	sha, boundary, ok = gatestate.LatestReviewed(l, milestone)
	if ok && !isResolvedSHA(sha) {
		// A hand-edited ledger must not hand git an option- or ref-like argument;
		// an unusable value reads as "not in this repository" (branch-patch fallback).
		return "", boundary, false
	}
	return sha, boundary, ok
}

// reviewMainBase is merge-base(main, HEAD) against the ONE main every branch-patch
// consumer measures from (#304 D11), or "" on main / with no divergence. In an issue
// tracker repository main is freshly fetched (the view the publish gate already uses),
// so a stale tracking ref cannot put the base behind a main the branch already merged.
// Elsewhere it is gitx.MergeBaseWithMain, which has no fetch source, as before. A
// criss-cross or unrelated history is an ERROR naming the bases, never a silent "" —
// the caller turns it into a named branch-patch fallback (D2). stale names a fetch that
// failed, so the window says it was measured against main as last fetched.
func reviewMainBase(ctx context.Context, head string) (base, stale string, err error) {
	ref, stale := reviewMainRef(ctx)
	if ref == "" {
		if trunk := gitx.TrunkRef(); trunk != "" {
			if _, err := gitx.SoleMergeBase(trunk, "HEAD"); err != nil {
				return "", stale, err
			}
		}
		return gitx.MergeBaseWithMain(), stale, nil
	}
	b, err := gitx.SoleMergeBase(ref, "HEAD")
	if err != nil {
		return "", stale, err
	}
	if b == head {
		return "", stale, nil // no divergence: on main
	}
	return b, stale, nil
}

// reviewMainRef is main as freshly fetched in an issue tracker repository, else "".
// A tracker repository whose fetch fails returns "" with a note, so the caller measures
// against the local tracking ref and SAYS so. Fetched at most once per process and
// checkout.
func reviewMainRef(ctx context.Context) (ref, stale string) {
	// Keyed by the checkout AND its local view of main: any fetch moves the tracking
	// tip and so misses, while repeated calls within one invocation share one fetch.
	// A root-only key returned the main of the process's first call even after a
	// later fetch, which pulled main's own changes into the branch patch.
	key := gitx.Capture("rev-parse", "--show-toplevel") + "@" + gitx.Capture("rev-parse", "-q", "--verify", gitx.TrunkRef())
	root := strings.SplitN(key, "@", 2)[0]
	reviewMainMu.Lock()
	defer reviewMainMu.Unlock()
	if c, ok := reviewMainCache[key]; ok {
		return c.ref, c.stale
	}
	var c reviewMainEntry
	if _, tracked, err := tracker.ReadCutoverMarker(root); err == nil && tracked {
		env, err := openTracker(ctx)
		if err == nil {
			var view *gitx.TrunkView
			if view, err = env.main.Snapshot(); err == nil {
				c.ref = view.Ref()
			}
		}
		if err != nil {
			c.stale = "main as last fetched (fetch failed: " + firstLine(err.Error()) + ")"
		}
	}
	// The snapshot's fetch just moved the tracking tip: cache under the new key too.
	reviewMainCache[key] = c
	reviewMainCache[root+"@"+gitx.Capture("rev-parse", "-q", "--verify", gitx.TrunkRef())] = c
	return c.ref, c.stale
}

type reviewMainEntry struct{ ref, stale string }

var (
	reviewMainMu    sync.Mutex
	reviewMainCache = map[string]reviewMainEntry{}
)
