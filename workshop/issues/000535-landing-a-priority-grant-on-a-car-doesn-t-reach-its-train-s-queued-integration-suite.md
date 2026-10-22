---
id: 000535
status: codecomplete
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'e62563420ae75fe0d7f96eb41f6c86a7461dad22' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T19:18:10-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "64611f08", done: "a7b66150"}
actual_hours: 0.63
---

# landing: a priority grant on a car doesn't reach its train's queued integration suite

## Problem

## Spec

pair train 254 (cars #775, #774, #771), 10-21 ~17:25: the TL granted fleet-blocking priority to #775 so the train's
already-queued integration suite would jump the machine queue (the train held pair's landing turn ahead of an
operator-accepted car). The grant showed in `sdlc landing status`, but the train's retained suite (lease 57913550, on
the composed train prefix 9ff56af4, not an issue head) stayed in normal FIFO at position 7 across queue ticks.
Grants bind to an issue's own suite lease; a train suite carrying that issue as a car doesn't inherit it.

Fix: a fleet-blocking grant on any car of a composed train applies to the train's suite lease (queued or about to
queue), re-evaluated on each queue tick; revoking or expiring the grant drops it. No priority for unrelated trains.

## Done when

- A train whose car holds a grant gets suite-queue priority for its integration suite (a test with a queued train lease
  and a later grant), and loses it on revoke or expiry.

## Plan

- [x] Reproduce missing train-car identity at the production local-check/helper seam.
- [x] Carry the exact composed prefix's issue IDs into suite queue metadata; reuse the existing grant provider and refresh/cache path to prioritize when any included car has a valid grant.
- [x] Cover late grant, revoke, unchanged-log expiry, multiple grants, unrelated repository/train, conflict exclusion, split prefix, reintegration and diagnostic baseline; run counted queue projection checks at N and 2N.
- [x] Update atlas and shipped guidance; pass targeted tests, recorded pushed-head suite and close review.

Landing and the final hub report follow through sdlc; they are not yet claimed complete.

### Proposed bounded design

The conductor retains issue IDs aligned with its existing successfully composed car heads. A local-run dependency projects only a matching current prefix SHA to its issues. The ordinary recorded test-check path passes those IDs to the capacity helper. Unmatched heads, non-test checks and unrecorded main-baseline diagnostics keep ordinary FIFO. Split/reintegrated runs use their actual prefix; rejected or removed cars confer nothing.

Add an optional issue-ID list to the existing queue subject, retaining common_dir and issue. Train metadata uses issue zero plus its exact prefix list; old helpers safely treat that as FIFO. The Python helper validates the optional list, expands the requested car identities into the existing singular-issue provider requests, and uses the latest active included deadline for admission/display. No protocol version change, grant copying, inferred grant or new authority file. The Go grant provider remains the sole authority (ARCH-DRY); membership projection and any-active ordering stay pure (ARCH-PURE). Existing grant-log signature invalidation and temporal refresh handle later grants, revocation and expiry without requeueing.

Alternatives: granting the conductor's issue cannot represent a later car's grant or a split prefix. Expanding grants by reading train event history adds a second membership authority and invalidation surface. Explicit exact-prefix metadata is smaller and follows the already composed inputs.

Actions remain grant, revoke and wait. Queue classes and recovery stay unchanged:

| Internal observation | Visible state | Action | Possible loss |
| --- | --- | --- | --- |
| At least one included grant active | fleet-blocking with included car/deadline | revoke or wait | none |
| No included grant, all revoked/expired, or unrelated grant | normal FIFO at original ticket | grant or wait | none |
| Invalid metadata or unreadable grant evidence | priority unavailable; normal FIFO | repair evidence or wait | priority only; suite retained |
| Suite already owns capacity | running; no preemption | wait | none |

Queue wait metadata has its existing lease lifetime and cleanup; no new durable family or background owner. The immutable membership list is bounded by the composed train; total request/response still obeys the existing 64 KiB protocol ceiling, failing to FIFO on unavailable priority. Per tick, expand each waiting subject once and visit each included issue once; reuse the existing one grant-log stat per repository and cached provider snapshot. Count N/2N projections/provider requests to reject rescans per car. No new external dependency, credential or destructive effect. Persisted additive metadata keeps old issue/common_dir semantics; helper record rewrites retain the subject dictionary. Malformed optional fields confer no priority. Semantic exact-prefix isolation is enforced with production-seam and real-helper regressions rather than a name-based Go lint rule.

Private conformance proposal: existing real Python helper/process/flock harness plus production Go provider, with a queued train subject receiving a later car grant and then revoke/expiry. Ask TL whether this is sufficient without staging before live acceptance.


## Log

### 2026-10-21
- 2026-10-21: closed — Focused production regressions passed (53.001s); broader train/priority tests passed (322.308s), tracked Go catalog passed (0.766s), diff check clean. Real Go grant provider plus helper/process/flock conformance covers queued late car grant, revoke and clock expiry; exact-prefix conflict/split/reintegration isolation, baseline/non-test FIFO and counted N/2N projection pass. TL approved no stage. Recorded suite on pushed a2cd7f9a89d75fd61505186b991d61ed4dae533f is queued as lease-875822645, NOT yet passing; this close review overlaps it under the worker contract, and no landing until completed pass; code fixes require fresh suite. Last routine suite/close/landing Plan row remains deliberately pending, hence no-plan-check. TL owns project updates, hence no-project.; review verdict: SHIP

- Claimed and entered planning from refreshed main. Root-cause trace: conductor.build uses a detached train worktree; runLocalCheck calls suiteSubject on that tree and tested composed SHA. suiteSubject only assigns one issue for an exact checkout HEAD on an issue branch, so a train has issue zero. The helper currently orders by one (common_dir, issue) key. Existing grant refresh already observes late grant/revoke/expiry; membership is the missing input. No code changed yet.

- TL approved additive metadata and exact-prefix projection (Couch a4d22167-ccd7-497d-ae1c-ae73cce8a9d5); private production provider plus real helper/process/flock conformance is sufficient, no stage. TL will confirm on the next natural train with a grant.
- TDD: stable real-helper late-grant regression failed before implementation: the included car's later grant left the train at normal FIFO position 2. Pure counted train projection failed with zero provider calls instead of one. Initial combined subtest fixture leaked its own held capacity between failed cases; split it into isolated tests and repeated the single late-grant RED (11.286s). Two earlier test selector typos were harness errors, not behavior evidence.
- Implemented optional issues metadata with legacy issue zero, exact current prefix projection through localRunDeps, and any-included-grant ordering using the existing provider/cache. The Go provider still receives singular issue requests; no grant authority or protocol version added. Real conductor fixture verifies conflicting car 2 is excluded while kept cars 1/3 reach the helper. Split/reintegration tests reject obsolete, empty and main heads; recorded suite/affected bind while smoke/baselines remain FIFO. Malformed lists fail to normal FIFO.
- Targeted production selection passed (53.001s): TestLandingTrainPriority*, TestSuitePriority*, TestSuiteQueuePriorityRealHelper and TestLocalCheckPrioritySubjectBinding. First run found only a fixture's unregistered newly committed fake GitHub SHA; registered that head and reran successfully. Pure grant-policy tests also passed, including N=8/16 train subjects, one provider request and one repository stat per snapshot, two car requests per train and unchanged-cache reuse.

- Broader verification passed: go test ./pkg/golint ./cmd/sdlc with TestLandingTrain*, suite-priority and local-check-priority selection (cmd/sdlc 322.308s; tracked Go catalog 0.766s). All real train scenarios and private provider/process/flock conformance passed. git diff --check clean. Recorded pushed-head full suite and mandatory close review remain pending.

- Infrastructure block at publish: sdlc pr failed fetching origin twice, including outside-sandbox retry, with fatal: unable to access https://github.com/xianxu/ariadne.git/: Could not resolve host: github.com. No PR or recorded suite was started. Local implementation and targeted/conformance evidence are committed; no gate bypass. Resume with sdlc pr after GitHub DNS/network recovers, then recorded suite, close and landing.

- Resumed after TL confirmed operator DNS repair. sdlc pr succeeded: PR382. Recorded suite queued on pushed a2cd7f9a89d75fd61505186b991d61ed4dae533f, lease-875822645, first in FIFO behind ariadne:1. This is pending capacity, not a test result. Close review may overlap under the worker contract; no landing until a completed branch-suite pass. Any review code change requires a new final-head suite.

- Recorded branch suite passed on a2cd7f9a89d75fd61505186b991d61ed4dae533f, lease-875822645: 2048 cmd/sdlc tests across 8 shards plus 75 other packages, zero failures, wall 876.8s (14m37s); capacity wait 359s. Saved log under .git/sdlc/local-checks/a2cd7f9a89d75fd61505186b991d61ed4dae533f/suite.log, sha256 prefix bddf2b8f9871; posted sdlc-local/suite. Close round 1 returned SHIP with zero findings; measured actual 0.63h. Reviewer independently passed targeted tests (61.023s), seven Python unit tests, singular-projection negative control and ticket/queue/slot unknown-metadata preservation. No new lesson from review findings (none). Entering landing without manually chasing main; the gate owns any necessary integration suite.

## Revisions

### 2026-10-21 — checklist bookkeeping after branch acceptance

The last routine row now records only completed documentation/tests/close. Landing/reporting remain explicit pending prose rather than a prematurely checked completion claim; scope and acceptance are unchanged.
