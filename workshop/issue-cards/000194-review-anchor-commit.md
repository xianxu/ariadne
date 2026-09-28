---
id: '000194'
status: done
started: 2026-08-20T15:52:21-07:00
created: 2026-08-20
updated: 2026-08-21
estimate_hours: 4.40
actual_hours: 9.74
---

# boundary reviews: anchor to the reviewed commit, and remember across rounds

## Problem

> **Diagnosis superseded 2026-08-20 — see `## Revisions`.** The freeze described
> below is real but is a *symptom*. The measured cost was the same commits being
> re-reviewed from the branch point every round. Kept verbatim as the filing record.

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
