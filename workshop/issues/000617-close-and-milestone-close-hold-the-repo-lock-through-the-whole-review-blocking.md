---
id: 000617
status: working
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '1e02bc7596fd42d6b4ffb312b4bcc452e97e1f22' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T20:25:11-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "cc3bb5d2", done: "7faf0b5c"}
---

# close and milestone-close hold the repo lock through the whole review, blocking other slots' bookkeeping

## Problem

## Spec

**`sdlc close` and `sdlc milestone-close` hold the repo's shared transaction lock for their whole run, including the
boundary review, blocking every other slot's bookkeeping in that repo (TL observation, 10-24).** pair:5's
`sdlc milestone-close --issue 859 --milestone M1` (pid 77562) held pair's repo lock for 25+ minutes while its review
ran; meanwhile pair:4's `sdlc stage --clear` and other slots' doc-only pushes and issue publishes waited on it. The same
happened earlier with pair#851 waiting on another close's lock. With 8 to 9 active slots per repo, one long review
serializes everyone's bookkeeping.

Fix: take the lock only around the short critical sections (reading and writing tracker or card state, publishing
commits), not across the review dispatch and wait; revalidate state after the review before writing the verdict.

## Done when

- A close or milestone-close holds the repo lock only for its bookkeeping writes; a concurrent `stage --clear`,
  `issue publish` or another close's bookkeeping proceeds while a review runs; tested with a slow fake review.
- A state change during the review is detected at the post-review write (refuse or re-run), never overwritten.

## Plan

- [ ] Implement the approved phased preparation/read-set/revalidation plan in workshop/plans/000617-close-and-milestone-close-hold-the-repo-lock-through-the-whole-review-blocking-plan.md.
- [ ] Prove independent bookkeeping during preparation/review with private tracker worktrees, preserve stale/cancel refusal, and complete normal verification/landing.

## Log

### 2026-10-24

### 2026-10-24 — Diagnosis and proposed boundary

Claimed and entered start-plan from main. Current source has manually scoped close locks since #166; targeted TestCloseCommandsReleaseLockDuringBoundaryReview plus TestReviewConcurrencySchedules passed94.650s. Pair:5 confirms #859 pid77562 spent17:00:50 to17:12:10 before its measured-actual line and reviewer dispatch, then completed M1 SHIP normally. The sample at17:06:47 and exact transcript establish preparation, not reviewer execution, as the observed hold. Pair:7 supplied exact #879 waiter lines but no holder stack/phase; no stronger attribution is claimed.

Ops:0 approved the diagnosis/direction and asked for this plan. Ariadne:2 (#603) agreed on the seam: preserve measurement/confidence semantics and wrap preparation/revalidation; whichever lands second integrates. Their follow-up identifies the adoption candidate filter as the needed concurrency read-set: no existing write-time guard rejects another issue's newly adopted interval. The plan reuses that filter, captures at the original read, and validates with fresh records scope rather than a cached tracker snapshot. Durable plan written for TL approval; no production edits or change-code yet. One read-only test exploration agent returned existing coverage and the real executable/tracker/linked-worktree gap; it changed nothing.

## Revisions

### 2026-10-24 — Observed hold was preparation, not reviewer execution

Pair:5's retained owner output places #859's long shared lock hold before its measured-actual line and reviewer dispatch. Existing reviewer release/reacquire code and tests work. Preserve that path; fix expensive preparation under the lock, validate consumed local/card/adoption authority, and move warning-only post-close effects outside the authoritative section. Ops:0 approved this diagnosis/direction; detailed plan is pending approval. Ariadne:2 confirmed the #603 coordination seam and the currently missing write-time adoption check.

### 2026-10-24 — TL approved implementation

Receipt25803def-0436-4cbb-ad57-7223be72b210 approved the durable plan. Record the boundary before change-code. Pair:5 independently confirmed exact start/first-output times and normal SHIP completion; its sample hash matches the plan. No retained targeted/background runner remains; the pre-design concurrency test completed exit0.
