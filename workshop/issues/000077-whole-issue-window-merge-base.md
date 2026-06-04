---
id: 000077
status: done
deps: []
github_issue:
created: 2026-06-03
updated: 2026-06-03
estimate_hours: 1.0
actual_hours: 1.0
---

# whole-issue close window: base on merge-base(main) not first #N commit

## Problem

`boundaryWindowBase` (added in #58) bases a **whole-issue** close window on
`firstCommitReferencing("#N")^` — the parent of the *first* commit whose subject
contains `#N`. For an issue filed early and implemented much later, that first
`#N` commit is the **"#N: file issue"** commit, not the first *implementation*
commit. So the end-of-issue boundary review window spans everything since the
issue was filed — including unrelated, already-reviewed-and-merged work.

Surfaced concretely in #58's own boundary review: the whole-issue window
resolved to `6be88ca^..HEAD` = **147 commits across ~12 unrelated issues**, when
the actual #58 deliverable was a single commit. The review noted it reviewed the
right commit by inspection, but the *window* was diluted. This is "over-cover,
not under-cover" (the safe direction, same as #58's missing-trailer fallback), so
it's not a correctness bug — but it makes end-of-issue reviews noisy and #58
entrenched the first-`#N`-commit base as the single window source.

## Spec

Base the **whole-issue** close window (milestone == "") on a tighter anchor:
- `merge-base(main, HEAD)` — the branch point — so the window is exactly the
  work this branch added; OR
- the most-recent prior `Review-Verdict:` boundary on the branch (the same
  `previousReviewBoundary` the milestone path uses), whichever is tighter.

Milestone windows already base on the prior boundary (#58) and are unaffected.
The change is localized to `boundaryWindowBase`'s `milestone == ""` branch.

Open question to resolve in design: working directly on `main` (no feature
branch) makes `merge-base(main, HEAD)` == HEAD, which would empty the window —
need a fallback (first-`#N`-commit, or prior boundary) for the on-main case.
This is why #58 used first-`#N`-commit rather than merge-base in the first place.

## Done when

- A whole-issue close on a branch filed-early/implemented-late windows only the
  branch's own commits, not unrelated merged history.
- The on-`main` (no branch) case still resolves a sane non-empty window.
- Milestone windows unchanged; `go test ./cmd/sdlc/...` green.

## Plan

**Open question resolved:** whole-issue close bases on **`merge-base(main, HEAD)`
only** — NOT "merge-base or prior-boundary, whichever tighter." Rationale: the
whole-issue close is the *end-of-issue integration review* (per #58's atlas note:
each milestone already reviewed its own slice; the whole-issue pass is the
integration look at the entire branch as a unit). Tightening to the last
milestone boundary would shrink it to just the final slice and defeat that
integration purpose. The branch point is the right base: it windows exactly this
branch's commits and nothing merged before it. On-main fallback = today's
`firstSHA^` behavior (merge-base == HEAD there).

Focused single-pass change (no milestone split). Localized to
`boundaryWindowBase`'s `milestone == ""` branch in `milestoneclose.go`.

- [x] Extract the existing `firstSHA^` logic into pure-ish helper
      `branchStartByIssue(issueStr) string` (first `#N` commit's parent, with the
      initial-commit fallback). Reused by BOTH the milestone no-prior-boundary
      fallback and the whole-issue on-main fallback (ARCH-DRY) — no behavior change
      to the milestone path.
- [x] Add `gitx.MergeBaseWithMain() string` (in `gitx/window.go`, next to
      `DiffBase` per plan-quality ARCH-DRY note): returns `git merge-base main
      HEAD`, but `""` when it equals `rev-parse HEAD` (on main / no divergence) or
      is unavailable — signalling the caller to fall back. Doc must cross-reference
      `DiffBase` and state WHY it's intentionally separate: `DiffBase` is the
      `sdlc judge` window (COMPARE-SHA override + generic `origin/main`/`HEAD~10`
      on-main fallbacks, never empty); `MergeBaseWithMain` is the close-window
      branch point with a deliberately empty no-divergence result so the caller
      picks the *issue-specific* branch start on main.
- [x] Rewire `boundaryWindowBase` `milestone == ""` branch: return
      `mergeBaseWithMain()` if non-empty, else `branchStartByIssue(issueStr)`.
      Update the doc comment (whole-issue = branch point, not first-`#N`-commit).
- [x] Tests (`milestonewindow_test.go` real-git pattern, extend `windowRepo` for a
      branch): (a) feature branch with unrelated main history before the branch
      point → whole-issue base == merge-base, and the unrelated pre-branch commits
      are NOT in `base..HEAD`; (b) on-main → falls back to `firstSHA^`; (c)
      existing milestone + whole-issue-on-main tests still green (no regression).
- [x] `go test ./cmd/sdlc/...` green; atlas `sdlc-binary.md` whole-issue-window
      wording updated (currently says "branch start … first `#N` commit").

## Log

### 2026-06-03
- 2026-06-03: closed — go test ./cmd/sdlc/... green; new TestBoundaryWindowBase_WholeIssueBasesOnBranchPoint asserts whole-issue window = merge-base(main,HEAD) and EXCLUDES filed-early + unrelated merged commits; on-main fallback + milestone paths unchanged. gitx.MergeBaseWithMain added next to DiffBase; review verdict: SHIP
Filed from #58's boundary review (Important, non-blocking) — the whole-issue
window over-captures when issue-file and implementation are far apart in history.
Pre-existing behavior #58 entrenched as the single window source. See [[000058]]
(`boundaryWindowBase`, `milestone == ""` branch).

### 2026-06-03 — implemented
Whole-issue close now bases on the branch point via new
`gitx.MergeBaseWithMain()` (`merge-base main HEAD`, returns `""` on no-divergence
so the caller falls back). Extracted the old `firstSHA^` logic into
`branchStartByIssue` (shared by the milestone no-prior-boundary fallback and the
whole-issue on-main fallback, ARCH-DRY). `boundaryWindowBase`'s `milestone == ""`
branch: branch point if diverged, else issue branch start.

Per plan-quality ARCH-DRY note, `MergeBaseWithMain` lives in `gitx` next to
`DiffBase` with a doc cross-referencing it and stating why the fallback contracts
differ (DiffBase = judge window, never empty; MergeBaseWithMain = close window,
empty-on-no-divergence). Resolved the spec's open question: branch point only,
NOT prior-boundary tightening — the whole-issue pass is the integration review of
the entire branch.

Tests: new `TestBoundaryWindowBase_WholeIssueBasesOnBranchPoint` builds a real-git
fixture (issue filed early on main → unrelated #99/#100 work → feature branch →
implement) and asserts the window = merge-base and EXCLUDES the file-issue +
unrelated commits (the 147-commit over-capture #58 surfaced). On-main fallback
pinned by the existing whole-issue test. `go build`/`go vet`/`go test
./cmd/sdlc/...` green; gofmt clean; atlas `sdlc-binary.md` updated.

### 2026-06-03 — boundary review: SHIP (high)
Dogfood: the close window resolved to `5b244ee..HEAD` (the branch point) — just
this branch's commits, NOT the 147-commit over-capture this issue fixed. No
Critical/Important findings.

Forward-awareness note (review recommendation, NOT fixed here — orthogonal,
pre-existing, safe to defer): **merge-main-into-feature** is an under-coverage
gap. If `main` is pulled *into* the feature branch mid-issue, `merge-base(main,
HEAD)` advances to the merged-in `main` tip, so feature commits made *before*
that merge fall outside `base..HEAD` and escape the whole-issue review. This is
under-coverage (the dangerous direction), but: (a) it's not introduced here —
`DiffBase`/`sdlc judge` already use identical merge-base semantics; (b) the
repo's flow is feature→main (`sdlc push`/`merge`), not main→feature. If it ever
bites, the fix lives in the same `boundaryWindowBase` decision point.
