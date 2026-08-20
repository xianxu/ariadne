---
id: 000194
status: open
deps: []
github_issue:
created: 2026-08-20
updated: 2026-08-20
estimate_hours:
---

# decouple boundary review execution from HEAD: finalize against the reviewed commit

## Problem

A boundary review (`sdlc close`, `sdlc milestone-close`) computes its window as
`BASE_SHA..HEAD`, dispatches a fresh-context reviewer, and on return refuses to
finalize if `HEAD` moved meanwhile — `closeReviewSnapshot.validate()`
(`cmd/sdlc/close.go:1198-1206`): *"HEAD changed from X to Y"*.

The review takes ~20 minutes of wall clock. For that whole window the working
tree is frozen: the agent cannot commit anything, on any branch, without
invalidating a run that is already most of the way through.

**This is a stop-the-world barrier, and it is the single largest cost in the
process.** Measured over one session (tools#1 + tools#2, 2026-08-20):

- 8 boundary-review runs, ~20 min each
- 4 of them were dead time — no other work was possible
- 1 run was killed outright, mid-flight, because the operator sent a change
  request and stalling them was the worse option
- 1 run completed and was then discarded, because a follow-up change had to land
  before the fixes could be applied

The failure mode is not theoretical: it converts an *asynchronous quality check*
into a *synchronous lock on the repository*, and the natural response — killing
the review to stay responsive — is exactly the behaviour the gate exists to
prevent.

## Spec

Anchor the review to the **commit it reviewed**, not to "HEAD has not moved".

`BASE_SHA..HEAD` is already resolved to a concrete SHA at dispatch. Record that
SHA as the review's anchor, and finalize against it:

- The verdict, the sidecar, and the `Review-Verdict:` trailer all refer to
  `REVIEWED_SHA`.
- Commits landing after `REVIEWED_SHA` are a **delta**, not an invalidation.
- On finalize, if `HEAD != REVIEWED_SHA`, the gate reports what the delta
  contains and decides — rather than refusing outright:
  - **doc-only delta** → finalize; the code surface reviewed is unchanged.
    `publishgate.go:186-193` already implements exactly this judgement for
    `sdlc merge` (#174, "doc-only commit(s) since close … reviewed-HEAD-unchanged
    holds for code"). The same rule should apply one stage earlier.
  - **code delta** → finalize the reviewed portion and report the unreviewed
    delta explicitly, so the next close reviews it; or refuse, but say
    *which commits* are unreviewed rather than only that HEAD moved.

The invariant worth preserving is **"every code commit was reviewed before it
shipped"** — not "no commit happened during the review". `merge`/`push` already
enforce the former at publish time via the anchor mechanism, which is why this is
a narrowing of the guard, not a loosening of the contract.

### Why the current guard is stricter than its own purpose

`validate()` also checks the issue file is unchanged, which is right — the review
read that text. But `HEAD` is checked as an identity rather than as a question
about code, and the machinery to ask the better question already exists in
`publishgate.go`. This issue is largely about **reusing that classifier one stage
earlier** (ARCH-DRY) rather than inventing a mechanism.

### Second-order effect worth having

Once a review is anchored to a SHA rather than to "now", running two reviews
concurrently on different commits becomes coherent, and a review can be
re-attached to its result later. Neither is required here; both are closed off by
the current design.

## Done when

- [ ] A boundary review records the SHA it reviewed, and finalizes against it.
- [ ] A commit landing during the review does not invalidate it; the gate
      classifies the delta and says what it decided.
- [ ] A doc-only delta finalizes, using the same classifier `sdlc merge` uses at
      `publishgate.go:186-193` — not a second implementation of it.
- [ ] A code delta is reported by commit, not as a bare "HEAD changed".
- [ ] The publish-time invariant is unchanged: no code ships unreviewed.
- [ ] A test covers the interleaving that motivated this — commit lands mid-
      review, review returns, finalize proceeds or refuses with the delta named.

## Plan

- [ ] Design via `sdlc start-plan` before implementing.

## Log

### 2026-08-20

Filed after a session that ran 8 boundary reviews across two issues in the
`tools` repo. The guard never produced a wrong verdict — every one of those
reviews found real, live-verified defects. The cost was entirely in
serialization, and the sharpest evidence is that the pragmatic response to an
operator request mid-review was to *kill the review*.

`publishgate.go` already contains the doc-only/code-delta classifier this needs;
the guard at `close.go:1198` predates it and asks a blunter question.
