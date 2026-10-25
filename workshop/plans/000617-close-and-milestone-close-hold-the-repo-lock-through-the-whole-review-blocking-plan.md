# Short close transactions implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach. Execute this cohesive change with superpowers-executing-plans; the SDLC close owns the fresh-context boundary review. Steps use checkbox syntax for tracking.

**Goal:** Close and milestone-close permit other worktrees' bookkeeping during expensive preparation and external review, without publishing stale close evidence.

**Architecture:** Prepare outside the repository transaction lock, validate the captured read-set at admission and again before committing, then perform local authoritative writes under that lock. Reuse the existing review snapshot, adoption filter, conditional tracker operations and post-publication recovery; do not add a second lock or durable transaction protocol.

**Tech Stack:** Go, Cobra, existing Git/tracker adapters, private linked Git worktrees and barrier-controlled reviewer fakes.

## Diagnosis and scope

`NewCloseCmd` and `NewMilestoneCloseCmd` already use manual lock ownership; #166 introduced this in July. `reviewThenFinalizeLocked` dispatches the reviewer outside the lock. Existing `TestCloseCommandsReleaseLockDuringBoundaryReview` and `TestReviewConcurrencySchedules` pass (94.650s for the targeted run).

The real #859 incident instead spent approximately eleven minutes in preparation: pid77562 started17:00:50, first measured-actual output and then boundary-dispatch output appeared17:12:10. Pair:5 confirms its sample17:06:47 was before reviewer execution; retained sample SHA256 e3eb754342f617d2d2c0a8e96711fa36dfeef5b7eaa853a7a21825a885fc1df5. `computeClose` calls measurement at close.go:454 while both locked wrappers hold the shared repository lock. Pair:7's #879 waiter independently confirms a long hold, but its capture does not identify the holder phase.

The change covers full close, milestone close, quick/fix review and explicit skip/dry-run paths. Expensive transcript/history measurement, project discovery and review preparation run unlocked. Churn preparation and non-authoritative peer/branch publication also leave the authoritative-write section. Local evidence commits, issue/ledger/calibration writes and conditional card/receipt transactions retain existing serialization. The tracker transaction's own bounded remote CAS remains part of that authoritative transaction; this is not a redesign of tracker publication.

No measurement algorithm, confidence policy, stored schema, command flags, review recipe, retry policy, queue policy or live-fleet behavior changes. No new persisted coordinator. A stale result refuses and names the same close command to rerun; it never silently recomputes or consumes another model review.

## Alternatives and chosen boundary

Adding another reviewer-unlock wrapper would not fix the reproduced preparation hold and would duplicate #166. Dropping the lock altogether would permit a changed read-set to overwrite newer issue/card/evidence state. Retain the existing write lock but move computation outside it and compare consumed authority before writes; this is the smallest boundary that serves both concurrency and safety. Holding the entire tracker generation constant would make unrelated claims stale, so only relevant adoption candidates and the close's own card are dependencies.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| ClosePreparationReadSet | cmd/sdlc/closeprepare.go | new |
| AdoptionReadSet | cmd/sdlc/actualexclusion.go | new |
| compareClosePreparation | cmd/sdlc/closeprepare.go | new |

ClosePreparationReadSet is one invocation's immutable authority: repository/worktree/branch/HEAD identity, the consumed issue/plan/gate/project bytes and presence, relevant tracker card identity, and the measurement's adoption candidates. Computed hours, split, saved inputs and #603 confidence are outputs carried together, never independently re-derived. It dies with the invocation.

AdoptionReadSet contains the exact numeric-adoption candidates consumed by `loadAdoptedExclusions`, including their identities and content authority. Extract its existing date/number/readability admission filter once and reuse it for capture and comparison. Additions or changes invalidate; conservatively treat removal as stale too. A title/claim change on a card outside this candidate set does not invalidate. Unreadable potentially relevant evidence refuses. Do not introduce a second overlap or attribution classifier.

compareClosePreparation only compares gathered values; it performs no Git, tracker, process or file IO. Existing `preparedReview` comparison and close's code-versus-doc anchor semantics remain the source for their current decisions. During preparation require a stable captured HEAD; during the subsequent review preserve today's permitted docs-only delta policy.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| prepareCloseTransaction | cmd/sdlc/closeprepare.go | new | existing computeClose and artifact/tracker reads |
| capture/observe adoption candidates | cmd/sdlc/actualexclusion.go, actual.go | modified | existing measurement tracker snapshot |
| runCloseWithReviewLocked / runMilestoneCloseLocked | cmd/sdlc/close.go, milestoneclose.go | modified | phased repo-lock ownership |
| finishCloseEffects | cmd/sdlc/closeprepare.go, close.go, boundarypush.go | new | existing warning-only peer/branch publication and diagnostic rendering |

Capture the measurement read-set at the read that supplies its candidates, not afterward: observing a later tracker snapshot would bless an intervening adoption. Carry this transient evidence through actualResult/closeFlags without persisting it. Reuse command-scoped issue records for internally consistent preparation; reacquisition MUST use a new records scope/fresh authority, not the cached preparation view. Verify the close's own card/owner/completion generation for both verbs before any authoritative write. Existing tracker CAS remains the final publication guard.

finishCloseEffects receives exact committed repository/branch/head and prepared peer edits, not floating HEAD. It runs after unlocking, preserves warning-only publication semantics and does not convert failed finalization into success. A lost branch push remains recoverable by the next boundary or sdlc pr. Existing peer publication re-derives edits on peer main; do not mutate peer checkouts. Split expensive churn collection from the small current-ledger/calibration update so output remains accurate without a history scan under the lock. Local authoritative effects already completed remain recorded if a later diagnostic/publication fails.

## Ownership and visible states

Actions: wait, rerun the named close command, reconcile the existing completion receipt.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| unlocked preparation/review in flight | preparing/reviewing | wait | in-memory result on process death |
| captured input changed or fresh evidence unavailable | stale; not finalized | rerun named close | current preparation/review cost; no newer input overwritten |
| caller canceled or reviewer failed | interrupted/failed; not finalized | rerun named close | incomplete review output |
| local evidence/card publication partially completed | existing completion pending | existing receipt reconcile command | existing protocol's recorded partial effects only |
| authoritative close finalized; optional push failed | closed, push warning | next boundary or sdlc pr | remote branch may lag; local close retained |

The repo transaction lock is the one resource authorizing local writes. Holding it for a model's lifetime cannot work because that serializes unrelated worktrees. Immutable inputs plus revalidation are necessary across the deliberately unlocked interval; the lock remains held from final validation through local mutation. There is no background task after the command returns. Cancellation before finalization writes nothing; after an authoritative partial effect, existing receipt recovery remains honest.

## Operating envelope and architecture

This is an infrequent batch checkpoint shared by8–9 active slots, with measured preparation taking8–11+minutes during load and reviews allowed their existing bounded duration. None of that latency should be spent monopolizing the repository lock. No timeout increase or timing-based acceptance is proposed. Relevant local bytes/cards grow with the actual read-set; capture and compare each once per phase, O(A+C), not per transcript event. Count at N and2N that validation reads scale linearly and never reruns transcript measurement. Existing remote-operation budgets continue to apply. UI/frame throughput is irrelevant because this command is a batch gate; memory holds only metadata and bytes already consumed by preparation.

ARCH-DRY: retain one comparison/filter per authority and existing lock/CAS/snapshot seams. ARCH-PURE: comparison is pure, capture/dispatch/publication are thin IO. ARCH-PURPOSE: reproduce preparation contention, not merely the already-tested unlocked reviewer. ARCH-MOCK: reuse tracker/Git fixtures and paused executable reviewer; real executables run only against private repos. ARCH-CONSTRAINTS: exclude long scans/dispatch from lock scope and count read-set growth. ARCH-SECURE: cached/old/unreadable authority cannot finalize; argv and exact refs retain existing rules. ARCH-STATE: the above phases retain partial receipts and reject stale completions. ARCH-FUNERAL: transient read-sets die at return and owned child processes are joined; no new durable artifacts. ARCH-EVOLVE: no format change or version bump; #603's measurement output semantics pass through unchanged.

## Implementation tasks (one close boundary)

- [ ] Reproduce the preparation hold with controlled existing computeActualForCloseFn and real lock admission. Retain the positive reviewer-unlock controls. Show independent bookkeeping cannot acquire during blocked preparation on current code; no wall-clock threshold or machine stress.
- [ ] Add closeprepare.go's captured read-set and phase coordinator, extracting the adoption candidate filter/read-set at its existing read seam. Wire preparation outside the lock in both verbs, then reacquire, freshly revalidate and admit the existing unlocked reviewer. Ensure skip/dry-run paths share the same expensive preparation boundary.
- [ ] Revalidate again before any sidecar/ledger/issue/evidence/card write, including new adoption candidates and own-card generation. Preserve stale/cancel refusal and the existing docs-only mid-review policy. Separate post-close optional effects and expensive churn from the local write section; pin any deferred branch push to its finalized identity.
- [ ] Add native real-executable/private tracker/linked-worktree conformance for both verbs, using barrier-controlled review and an unrelated card setter or issue publish. Add transient read-set/class comparison tests and update atlas/workflow/sdlc-binary.md, root/close/milestone help where their lock contract needs precision, and workshop/lessons.md.
- [ ] Run targeted verification, push, recorded sdlc local-check run --pre-close, then normal close/PR/merge gates; report landing for the hub's central rebuild. No primary staging or live lock experiments.

## Function-level adversarial verification

- `prepareCloseTransaction`, both command wrappers: barriers inside measurement and reviewer dispatch plus an independently acquiring command prove lock availability; positive finalization and canceled/stale schedules establish the permitted transition boundary. Restore the unsafe measurement placement and require red before accepting green.
- `compareClosePreparation`: mutate captured identity/artifact presence/content and feed changed current card authority; assert stale decisions preserve the newer files and do not write review sidecars or evidence. Exercise the real caller so a correct but unused helper cannot pass.
- adoption read-set capture/comparison: generate tracker histories with interleaved numeric adoption/correction while preparation is held; the shared candidate filter admits added/changed relevant authority and ignores unrelated non-adoption bookkeeping. Use #603 summary inputs after integration. No second scan of transcripts under finalization.
- `finishCloseEffects`: block/fail warning-only publication and verify another worktree can acquire the repo lock while exact finalized evidence remains intact; mutate branch identity to ensure deferred effects cannot publish a newer floating tip.
- Existing close review concurrency, stale artifacts, ledger/plan seed, cancellation/reacquire, tracker close/milestone evidence and branch push tests remain controls. Targeted normal/race runs are serial and bounded; full verification stays in the machine queue.

## Coordination and rollout

Ariadne:2 explicitly agreed to this seam: #617 wraps preparation/revalidation; #603 owns exact/partial/unmeasured confidence, field/trailer/calibration semantics and adopted summary handling. Whichever lands second integrates the other's main and revalidates the combined behavior. #617 must carry Actual, Split, Inputs and Confidence together when #603 is present. No stacking or edits to their branch.

Private linked-worktree/executable proof exercises the actual lock and tracker adapters without pausing or mutating the production fleet. On landing ops:0 rebuilds the central sdlc. Old clients keep their current behavior until rebuilt; no new stored data requires a coordinated reader rollout.

## Approval

Approved by ops:0 receipt25803def-0436-4cbb-ad57-7223be72b210 (2026-10-24T20:33:28-07:00). Enter change-code before implementation.

## Revisions

### 2026-10-24 — Evidence narrows the incident and identifies measurement authority

The filed Spec described a lock held during review. Owner evidence instead locates #859's observed hold in pre-review measurement, so this plan preserves the existing reviewer-unlock path and fixes the enclosing preparation/finalization phases. Ariadne:2 confirmed no existing validator detects newly adopted competing time; the shared adoption candidate read-set is necessary to preserve accounting while moving preparation outside the lock. No confidence/measurement algorithm change is proposed. Approval remains pending.

### 2026-10-24 — TL approval

Ops:0 approved unlocked expensive preparation, fresh read-set validation before admission and writes, unlocked review, short authoritative writes under the lock, peer/branch side effects outside it, and adoption candidates captured at their original read and rechecked. New or changed adoptions stale; unrelated claims do not. Private barrier conformance and no new state/schema are confirmed. Proceed through change-code.

### 2026-10-24 — Implementation naming and publication detail

The invocation guard is `closePreparationReadSet` in closeprepare.go and is an IO integration, since its validate method gathers fresh authority. Pure comparisons reuse `comparePreparedReview`, `closeCardAuthority` and `compareAdoptionReadSet` rather than adding the conceptual `compareClosePreparation` name. Preparation shares one fresh tracker view for own-card and adoption comparison. Optional branch publication pins the finalized commit; its short local tracking-ref update reacquires the existing lock after network IO. These are implementations of the approved boundary, with no new durable state or measurement policy.
