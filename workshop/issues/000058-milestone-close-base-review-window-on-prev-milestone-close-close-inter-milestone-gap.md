---
id: 000058
status: working
deps: []
github_issue:
created: 2026-06-01
updated: 2026-06-03
estimate_hours: 1.5
---

# milestone-close — base review window on prev milestone close (close inter-milestone gap)

## Problem

`sdlc milestone-close` computes its fresh-eyes review window as **"first commit
referencing `#<issue> Mx`"^..HEAD** — the parent of the first commit tagged with
that specific milestone. So a commit that references the issue (`#<issue>`) but
**not** a milestone — a side-quest, a fix, a plain `#N` commit made between two
milestone closes — can fall in the gap: it post-dates the previous milestone's
window and is *excluded* as the base of the next one. It escapes both reviews.

Surfaced in #57: `c52f482` ("#57: dev-aliases — build to owner bin/") referenced
`#57` but not `M2`. M1's window ended at its close (`76a2828`); M2's window based
at `699635d^` = `c52f482`, excluding it. The M2 judge happened to review it as
compensation (it was the base), but that's luck, not guarantee — a commit two or
more before the milestone's first `Mx` commit would simply vanish from review.

This defeats the milestone-review's purpose (every change reviewed before the
next milestone) precisely for the unplanned commits (`side-quest:`, fixes) that
most warrant a fresh eye.

## Spec

**Base each milestone's review window on the previous review boundary, not on
the first `#N Mx` commit.** Concretely:

- The window for `Mx` is `[prev-boundary .. HEAD]`, where `prev-boundary` is the
  commit at which the *previous* milestone-close's review ended — i.e. the prior
  milestone-close commit carrying this issue's `Review-Verdict` trailer (the
  `Review-Window: …..<head>` it recorded).
- For the **first** milestone (no prior boundary), base on the issue's branch
  start (the merge-base with `main`, or the first `#N` commit's parent) — same
  as today's first-commit behavior, just not keyed to the `Mx` tag.

This guarantees every commit since the last review boundary — including
inter-milestone `#N`-but-not-`Mx` commits — lands in exactly one milestone
window, so nothing slips the net.

Mechanism options for the implementer to weigh (don't over-specify):
- Read the prior milestone-close commit's `Review-Window` trailer head SHA and
  use it as the new base (authoritative — it's exactly where the last review
  stopped).
- Or `git log --grep "Review-Verdict:" -- <issue file>` for the most recent
  prior boundary on this branch.

**Non-goals:** don't change the verdict trailers, the judge prompt, or the
`#N Mx[: ]` subject matcher (that's the *detection* fix from #57's side-quest,
already landed). This is purely the *window base* computation.

## Done when

- A `Mx` milestone-review window includes inter-milestone `#N`-but-not-`Mx`
  commits made since the previous milestone close (the #57 `c52f482` shape).
- The first milestone still bases on the branch start (unchanged behavior).
- A regression test pins it: a fixture branch with `[M1 work][M1 close][#N
  side-quest, no Mx][M2 work][M2 close]` asserts the M2 window covers the
  side-quest commit.
- `helptext/milestone-close.md` "Diff window" wording updated;
  `go test ./cmd/sdlc/...` green.

## Plan

Focused single-pass change (one milestone — not split).

- [x] In `cmd/sdlc/milestoneclose.go`, replace the window base with the previous
      review boundary via a new `boundaryWindowBase` helper (prior milestone-close
      commit carrying a `Review-Verdict:` trailer, found by `previousReviewBoundary`;
      else branch start). Per Revision 1, the atlas gate in `close.go` shares the
      SAME helper (ARCH-DRY), so inter-milestone commits land under both checks.
- [x] Regression test (real-git fixture, `milestonewindow_test.go`): an
      inter-milestone `#N`-but-not-`Mx` commit between M1 close and M2's first
      commit falls inside M2's window; first milestone + whole-issue base on
      branch start.
- [x] Update `cmd/sdlc/helptext/milestone-close.md` ("Diff window") + atlas
      (`sdlc-binary.md`); `go test ./cmd/sdlc/...` green. (`sdlc` rebuilds from
      source per-invocation — no separate install step.)

## Revisions

### 2026-06-03 — plan-quality findings resolved (pre-implementation)

`sdlc change-code` plan-quality judge raised three coupling points the original
Spec/Plan missed. Resolutions (these supersede the conflicting Plan/Non-goal
lines above):

1. **ARCH-DRY — atlas-gate window shares the same scan.** `resolveReviewWindow`'s
   base is documented as the same source as close.go's atlas gate
   (`firstCommitReferencing(refSubject)`, `close.go:384/397`). Moving only the
   *review* window to the prior boundary would diverge the two and falsify that
   invariant. **Resolution: move BOTH** — extract one `boundaryWindowBase`
   helper used by the atlas gate (close.go) AND the review window
   (milestoneclose.go). This keeps them a single source *and* is desirable:
   inter-milestone commits now also fall under the atlas-coverage check, not
   just the review. (Original Spec "Non-goals" did not name this; it's now in
   scope.)
2. **Shared helper — `resolveReviewWindow` is also called by whole-issue close**
   (`close.go:613`, refSubject `#N`), which must keep its branch-start window.
   The prior-boundary base is **milestone-only** (`milestone != ""`); the
   whole-issue path stays branch-start. The helper branches on milestone.
3. **Mechanism — Spec option 1 is unworkable.** The recorded `Review-Window`
   head is the literal string `"HEAD"` (set by `resolveReviewWindow`), not a
   SHA. Use **option 2**: grep the most recent prior commit touching the issue
   file carrying a `Review-Verdict:` trailer and use *its commit SHA* as the
   base (exclusive in `base..HEAD`, so the prior close commit itself isn't
   re-reviewed). Reuses the existing `milestoneHasVerdictCommit` grep shape
   (`close.go:837`).

Timing note: `runClose` does not commit (`close.go:115`), so at both the atlas
gate and the review-window computation the current close commit does not yet
exist — `previousReviewBoundary` finds the genuine *prior* boundary at both.

## Log

### 2026-06-01
Filed from #57's M2 review, which caught `c52f482` slipping between the M1 and
M2 auto-windows. The judge reviewed it as the M2 base and found it correct, so
no harm landed — but the gap is real: inter-milestone `#N`-but-not-`Mx` commits
(side-quests, fixes) can escape review entirely. Fix is to base each window on
the previous review boundary so coverage is gap-free. Pairs with the #57
side-quest that hardened the `#N Mx close:` subject *matcher* (`b083bab`-era) —
that fixed detection of the close commit; this fixes the window it reviews.

### 2026-06-03 — implemented
Extracted `boundaryWindowBase(issueStr, milestone, issuePath)` in
`milestoneclose.go` as the one window source; `resolveReviewWindow` and the
`close.go` atlas gate both call it (ARCH-DRY, per Revision 1). Milestone path
bases on `previousReviewBoundary` (most recent prior `Review-Verdict:` commit
touching the issue file); milestone=="" and first-milestone fall back to branch
start. Also DRY'd the issue-file glob into `issueFilePath` (reused by
`annotateLogLineWithVerdict`) and simplified `firstCommitReferencing` to drop its
now-dead `count` return. Regression test `milestonewindow_test.go` pins all three
shapes (prior-boundary covers the side-quest; first-milestone + whole-issue =
branch start). `go build`/`go vet`/`go test ./cmd/sdlc/...` all green.
