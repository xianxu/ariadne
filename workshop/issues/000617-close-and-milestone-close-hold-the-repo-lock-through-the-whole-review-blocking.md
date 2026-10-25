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

- [x] Implement the approved phased preparation/read-set/revalidation plan in workshop/plans/000617-close-and-milestone-close-hold-the-repo-lock-through-the-whole-review-blocking-plan.md.
- [ ] Prove independent bookkeeping during preparation/review with private tracker worktrees, preserve stale/cancel refusal, and complete normal verification/landing.

## Log

### 2026-10-24

### 2026-10-24 — Diagnosis and proposed boundary

Claimed and entered start-plan from main. Current source has manually scoped close locks since #166; targeted TestCloseCommandsReleaseLockDuringBoundaryReview plus TestReviewConcurrencySchedules passed94.650s. Pair:5 confirms #859 pid77562 spent17:00:50 to17:12:10 before its measured-actual line and reviewer dispatch, then completed M1 SHIP normally. The sample at17:06:47 and exact transcript establish preparation, not reviewer execution, as the observed hold. Pair:7 supplied exact #879 waiter lines but no holder stack/phase; no stronger attribution is claimed.

Ops:0 approved the diagnosis/direction and asked for this plan. Ariadne:2 (#603) agreed on the seam: preserve measurement/confidence semantics and wrap preparation/revalidation; whichever lands second integrates. Their follow-up identifies the adoption candidate filter as the needed concurrency read-set: no existing write-time guard rejects another issue's newly adopted interval. The plan reuses that filter, captures at the original read, and validates with fresh records scope rather than a cached tracker snapshot. Durable plan written for TL approval; no production edits or change-code yet. One read-only test exploration agent returned existing coverage and the real executable/tracker/linked-worktree gap; it changed nothing.

### 2026-10-24 — First implementation checkpoint

change-code inferred quick from the current design; no flow pin or gate bypass was used. The final diff will leave the code shell and close must supply its full review. The preparation regression first failed correctly for BOTH verbs: computeActualForCloseFn ran with held=1 (2.427s). The new phased wrapper captures initial input authority, computes unlocked, validates fresh own-card and adopted-candidate authority at admission and again before writes, and retains the existing reviewer snapshot. Metadata comparison reuses one fresh card view per phase. Measurement confidence/algorithms remain owned by #603.

Transient adoption read-set and existing parked/legacy exclusion controls passed43.015s, including a real private tracker with cached preparation records followed by an intervening numeric adoption. Core command regressions passed60.504s: measurement unlocked, stale issue/plan/HEAD refused before dispatch, and actual issue-branch pre-push hooks see the lock absent. Branch publication first reproduced locked for both verbs; its fixture needed an explicit private core.hooksPath, and later a destination-ref match because publication now pins a SHA. The exact finalized issue commit is deferred; peer publication and churn collection also leave the lock, while authoritative writes and the short local tracking-ref update stay serialized.

Real executable/private tracker/linked-worktree conformance passed33.970s, package exit0: both verbs permit a sibling title setter during a real churn-read preparation barrier and a private reviewer barrier; finalization preserves the unrelated title and emits valid evidence. An earlier body-pass run was correctly REJECTED by TestMain because I edited docs while it ran; it is not passing evidence. There were no live locks or fleet processes manipulated. No targeted runner remains after that accepted repeat. One final pin regression and the existing concurrency/publication controls remain, followed by race/catalog checks, queued pre-close suite and close/PR/merge. Source/docs are checkpointed as implementation in progress, not codecomplete.

## Revisions

### 2026-10-24 — Observed hold was preparation, not reviewer execution

Pair:5's retained owner output places #859's long shared lock hold before its measured-actual line and reviewer dispatch. Existing reviewer release/reacquire code and tests work. Preserve that path; fix expensive preparation under the lock, validate consumed local/card/adoption authority, and move warning-only post-close effects outside the authoritative section. Ops:0 approved this diagnosis/direction; detailed plan is pending approval. Ariadne:2 confirmed the #603 coordination seam and the currently missing write-time adoption check.

### 2026-10-24 — TL approved implementation

Receipt25803def-0436-4cbb-ad57-7223be72b210 approved the durable plan. Record the boundary before change-code. Pair:5 independently confirmed exact start/first-output times and normal SHIP completion; its sample hash matches the plan. No retained targeted/background runner remains; the pre-design concurrency test completed exit0.

### 2026-10-24 — Consolidated publication boundary approved

Ops:0 receipt d1246e0f-8e5c-4188-856f-be14df1d44bf approves one immutable publication record captured under the finalization lock: repository/transport, branch, finalized commit and expected remote tip. The unlocked push uses exactly that lease; changed remote authority refuses, and local tracking metadata updates with CAS under a short lock. No retry, persistent state or schema. This is the third design delta: a fourth cuts scope to retaining publication under the lock and moving measurement only.

The newer-authority regression reproduced both failures (5.115s): the deferred push rewound a newer remote tip, and reading a changed tracking ref late refused the intended finalized commit. Existing concurrency/publication controls passed237.702s; fresh adoption checks used19 Git calls for both8 and16 candidates (3.264s). The focused measurement/stale/reviewer-lock race controls passed63.847s; that regex did not include the adoption/publication families, which still need a current race run. Implementation remains in progress.

### 2026-10-24 — Publication boundary verified

The consolidated immutable record now captures finalized SHA and remote lease under the lock, pushes to that captured transport, and updates tracking metadata with CAS under a short lock. The real-Git newer-remote and newer-tracking controls went green. The first combined run found only a legacy warning-text assertion regression; restoring that existing warning preserved the unrelated boundary contract. The corrected combined race run passed100.780s, covering adoption read-sets, unlocked measurement, stale preparation, reviewer unlock, finalized-head/transport/newer-authority publication and existing push controls. The complete tracked Go lint catalog passed1.138s; git diff --check is clean. No fourth mechanism delta appeared. Next: PR, recorded queued pre-close suite, mandatory close review, then central merge; no full suite or live staging has run yet.

### 2026-10-24 — Recorded verification queued; review overlaps

PR437 is open. The central pre-close command resolved to suite on044aa40f and retained lease-495839293, private checkout /Volumes/TestScratch/local-checks/sdlc-local-3742177344. It is normal FIFO, initially position10, behind the existing fleet work; no duplicate or cancellation. Mandatory close review may overlap the queued check under tl-worker's rule; the final recorded pass still gates landing. The precise --no-plan-check acknowledgment covers the remaining process checkboxes (recorded suite and landing), not unfinished implementation. Targeted/race/native evidence above is complete; no full-suite pass is claimed. Previous recovery screen was requested and reported unavailable; recovery relies on the committed Log, plan and Git evidence.

### 2026-10-24 — BR-1 command wiring coverage

Round1 was FIX-THEN-SHIP, Important BR-1 (authority-invalidation-command-coverage), with no Critical findings. Helper comparison and --no-actual executable conformance did not prove actualResult.Adoptions → closeFlags → preparation guard wiring. Added TestCloseCommandsRefuseChangedAdoptions: both close verbs, new/changed numeric adoption, preparation/review interleavings; each refuses before issue/card/evidence or sidecar/ledger mutation. Private tracker only, deterministic callback ordering, no timing threshold.

The first run exposed a missing private plans directory (fixture setup, corrected). All8 schedules then passed69.858s. Removing only guard.adopted propagation made all4 new-adoption command/phase schedules wrongly finalize (RED46.096s); restoring it and removing only the adoption comparison made both review-time schedules wrongly finalize (RED24.287s). Both temporary source mutations were fully restored (closeprepare.go diff empty). Final all8 race run passed69.002s. This is the already-planned test coverage, not a fourth mechanism delta. The old queued044aa40f suite lacks these new tests and will be canceled by its exact owned lease before the updated head is queued.

Ariadne:2 confirms #603 integration must carry closeFlags.Confidence with Actual/Split/Inputs; verify TestLegacyCloseRecordsEveryConfidence and TestTrackerCloseCarriesConfidenceTrailer after whichever landing integrates the other. Its current pre-close suite is running; no #603 integration is claimed yet.
