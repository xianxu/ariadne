---
collaboration-mode: discussion
---

# Concurrent claim attribution implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3. Execute this context-dependent measurement repair in the owning session; SDLC owns fresh-context boundary review.

**Goal:** A held reservation must not claim another issue's active branch time, including its post-close verification and landing tail.

**Architecture:** Retain tracker ownership and transcript lineage as prerequisites. Intersect admitted source windows with a single branch/claim classifier before activity allocation. Capture worktree-local HEAD history once, retain the resulting evidence with the measurement, and replay adopted snapshots without consulting mutable current branch state.

**Tech stack:** Go, existing tracker/Git seams, native disposable Git worktrees, existing active-time engine and snapshot codecs.

## Evidence and scope

The issue records parley.nvim:3's #418 branch continuing through its 14:29 landing while #322 was also claimed. #322's 14:42 close proposed 4.59h, dominated by 11:18–14:42. That proves wrong-issue attribution of #418's unadopted tail; it does not prove those exact minutes were adopted twice. Preserve that distinction in the fixture and output. No raw TTY capture is attached to the issue; the incident fixture reconstructs the documented ordering, not invented terminal bytes.

Current EligibleSources admits a persistent root session for an entire owned worktree window. It has no branch input. resolveOwnedActualInputsFor writes those broad source windows into the measurement; Compute then attributes runs using the measured issue's scoped commit boundaries. A second reservation therefore admits A's activity while measuring B. Existing adopted-credit exclusions protect earlier adopted build/converge but cannot establish who owns A's unadopted tail.

All three hour consumers must derive from the corrected resolver: actual preview, active-time's owned issue route, and close adoption. Token usage is outside this repair. Legacy no-tracker measurement and already-adopted snapshot semantics remain unchanged. No retrospective repair or handwritten totals.

## Proposed policy — requires TL approval before code

1. A witnessed canonical issue branch selects that issue, but only where tracker ownership and session lineage already permit it. A foreign issue branch excludes the interval from this candidate even if its work was never adopted elsewhere.
2. A sole owned claim may continue using the existing ownership-based rule for planning on a resting branch or where branch evidence is unavailable. Label that provenance; this preserves claim-before-start-plan design attention. Do not describe it as proven branch activity.
3. Where the same machine/repository/workspace has overlapping claims, resting/detached/missing/ambiguous branch evidence cannot choose between them. Refuse the numeric measurement with the exact source/slot/time and reason. The explicit close recovery is --no-actual with the evidence in --verified; do not fabricate a new owner or move the start time.
4. Exact half-open branch transitions partition activity, including an inter-event gap or subagent span that crosses A → B. Preserve original dispatch/engagement identity. Queue waits are excluded only from the selected issue activity through the existing immutable-head ownership rule. A's close does not transfer A's tail to B.
5. Historical adopted windows and their exclusions replay as originally saved. This patch does not silently reinterpret old cards, calibration rows or legacy prefixes.

## Alternatives

- **Recommended: ownership plus captured worktree HEAD history.** This is harness-independent and captures checkout transitions already made by Git. A native private-worktree control will qualify its format and boundaries before implementation expands.
- **Transcript mentions or nearby issue commits only.** Rejected: a dispatch can mention B while still on A; commits can cite other issues, and main integration introduces foreign history. These are the incident's existing weak signals.
- **One claim per slot.** Rejected: reservations are intentional and an old completed claim can remain held. A new lifetime lock would change scheduling rather than reconstruct historical activity, and cannot prove the past. No absence-by-process scan is proposed.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| BranchObservation / BranchInterval | cmd/sdlc/internal/activetime/branch.go | new |
| SelectBranchActivity | cmd/sdlc/internal/activetime/branch.go | new |
| MeasurementBranchEvidence | cmd/sdlc/internal/gatestate/measurement_branch.go | new |

BranchObservation is a typed, time-ordered checkout/unknown witness for one verified worktree. BranchInterval is its half-open interpretation. SelectBranchActivity joins branch intervals, tracker claim cardinality and already-admitted source windows; it returns selected windows, labelled exclusions and unresolved intervals. The selected windows feed existing gap/spans, queue-idle and protected-credit logic rather than creating a second hour calculator.

MeasurementBranchEvidence retains bounded capture identity and only the transition witnesses needed to explain selection, beside the existing Sources.Windows. The persisted windows remain the actual replay authority. Historical replay never silently reads today's reflog to replace them.

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| captureOwnedBranchEvidence | cmd/sdlc/actualbranch.go | new | existing readonly Git adapter and per-worktree HEAD log |
| resolveSlotClaimWindows | cmd/sdlc/actualbranch.go | new | pinned tracker card/EngagementHistory reads |
| resolveOwnedActualInputsFor | cmd/sdlc/actualinputs.go | modified | source admission, capture and retained measurement inputs |

Capture resolves the Git-verified owner worktree and its own HEAD log, not the common repository's other slot log. Read a bounded stable prefix once per distinct root, retaining its digest/length and final observation boundary. A changing/truncated/unreadable log is unknown, never continuity. Decode a narrowly qualified Git grammar; unsupported branch-affecting operations or equal-time contradictory transitions remain unknown. Do not treat a current symbolic ref, a session's initial branch or the next commit subject as a historical branch witness.

Claim windows derive from the same immutable tracker pin as the candidate. Use complete per-card engagement history and exact machine/repository/workspace identity, including releases, reclaims and relocations. Read each relevant history once in the measurement request; do not query a tracker tip per event. Missing proof cannot establish a sole claim.

## State and recovery

Actions: inspect evidence; explicitly close without an actual; continue normal measurement.

| Internal state / failure | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Proven owned issue branch | selected, branch-witnessed | continue normal measurement | none |
| Proven foreign branch | excluded, names issue and interval | continue normal measurement | this candidate loses foreign minutes |
| Sole claim, no issue-branch witness | selected, ownership-only provenance | continue normal measurement | no fabricated branch proof |
| Concurrent claims and unresolved branch; incomplete claim history | measurement unavailable, names slot/range/reason | explicitly close without an actual | no numeric calibration sample |
| Missing or contradictory retained evidence | measurement unavailable | inspect evidence | existing saved actual is not rewritten |

This is a read-only batch classifier, not a daemon or workflow state machine. Its input order is captured once; cancellation stops collection and publishes no candidate proof. Late source appends do not change an already pinned source prefix. Repeated adoption still uses the existing close generation and protected-credit rules.

## Compatibility, bounds and architecture

Sources.Windows already expresses exact source eligibility, so older readers replay the narrowed windows safely. New branch provenance is additive and must preserve unknown fields at every nested read/write, with existing strict known types and duplicate-key handling. No method/version bump merely for extra audit data: ignoring it must not widen the saved windows. Test the real previous codec on a new snapshot and the existing rewrite path. If implementation shows an old reader can widen them, stop and bring that unsafe interpretation to the TL before choosing a rollout/version boundary.

Keep existing 128-source / 64-KiB snapshot bounds. Initial capture envelope: at most 4 MiB or 20,000 HEAD records per distinct worktree, with a visible unavailable result before crossing either bound. These are conservative batch safety limits, qualified against native fixtures and the incident window; no silent partial history. Persist only required transition witnesses, not transcript bodies or the whole reflog. Existing ledger lifecycle owns these additive records; no new file family or retention service.

Growth drivers: source events E, claim transitions C, branch transitions B, distinct roots R and tracker cards I. One capture per R and one history read per I; sort/join intervals once, then linear intersection per source, rather than probing Git or scanning all claims for every event. Count IO and sweep operations at 8 and 16 roots/transitions; reject per-event IO and quadratic rescans. Existing active-time cancellation and source/snapshot limits remain in force. No network, credentials or primary mutation.

ARCH-DRY shares one classifier and the existing hour engine. ARCH-PURE isolates interval selection. ARCH-PURPOSE includes A's unadopted tail and all hour consumers. ARCH-MOCK uses existing tracker fixtures and real private Git worktrees, plus counted injected readers. ARCH-CONSTRAINTS bounds discovery and joins. ARCH-SECURE retains ownership/lineage and strict typed history. ARCH-ORDER partitions transitions and preserves dispatch identity. ARCH-FUNERAL adds no independent durable family. ARCH-EVOLVE keeps exact saved windows and additive preserved audit fields.

## Implementation and adversarial verification

Single atomic issue close; these checkboxes are tasks, not milestone boundaries.

- [ ] Qualify native HEAD-log capture on a private repository with two real worktrees. Switch A → resting → fresh main-based B under controlled timestamps; include merge/rebase/reset/rename/detach controls. Demonstrate which records prove continuity and which become unknown. Preserve the incident ordering as a small labelled fixture.
- [ ] RED through the production resolver: simultaneous claims A/B, persistent session on A through close/queued test/landing, then B. Existing code credits A's unadopted tail to B. Check A and B interval identities, not only totals; an adopted A prefix stays excluded exactly once.
- [ ] Implement pure SelectBranchActivity. Adversarial strategy: reorder, duplicate and corrupt boundaries; equality at transition; gap/span crossing; missing/expired prefix; foreign issue; multiple roots; relocation and reclaim. Independent guard: remove branch restriction and require the incident regression to fail, plus property checks that A/B selected intervals are disjoint and no selected interval exceeds prior ownership.
- [ ] Implement captureOwnedBranchEvidence and resolveSlotClaimWindows through existing seams. Adversarial strategy: different slot logs, aliases, removed worktrees, modified/truncated capture, capture races, missing card history, cancellation and limit exhaustion. Guard: real Git/tracker fixtures plus counted 8/16 captures/history reads; assert zero writes and no live-root fixtures.
- [ ] Wire the resolver before persistence and activity allocation; exercise actual, active-time and close through normal command fixtures with Claude, Codex and Qoder sources. Verify sole-claim planning compatibility, simultaneous claims with unknown branch refusal, no foreign-tail reassignment, and original TaskSpan dispatch boundaries.
- [ ] Retain additive provenance and narrowed existing windows. Adversarial strategy: append/expire live logs after adoption and replay unchanged; tamper retained known fields; round-trip unknown nested fields; old-codec read/rewrite. Guard: exact old/new replay interval equality and strict codec controls. Do not backfill historical samples.
- [ ] Update actual/active-time help, README and atlas/workflow/sdlc-binary.md with branch versus reservation, ownership-only planning and unavailable recovery. Add a lessons rule for this semantic attribution defect; behavioral interval invariants, not lexical matching, enforce it.
- [ ] Targeted pure/integration/race tests, then pushed-head central sdlc local-check run --pre-close. Close with measured actual and mandatory review, enter central sdlc merge promptly on SHIP, retain the runner through result and report Git-verified landing. Private exact-executable preview on isolated paths is the conformance proof; request TL acceptance of that plus post-land rebuild, no staging and no historical card mutation.

## Approval

Await TL decision on the proposed sole-claim compatibility rule, unknown concurrent-claim refusal, captured HEAD-history evidence and private conformance/post-land rebuild. No implementation or recorded suite has started.

## Revisions

### 2026-10-23 19:32 PDT — authenticated TL approval

TL approved all three policy choices: exclude work on another claimed issue's witnessed branch; retain labelled ownership-only attribution for sole-claim planning; refuse a numeric actual for concurrent claims with missing, resting or ambiguous branch evidence and name explicit --no-actual recovery. Approved one selector over captured per-worktree HEAD history, additive saved provenance and immutable adopted replays. Proceed to change-code. The Approval paragraph above records the original proposal state. No implementation has started at this checkpoint.

### 2026-10-23 — native capture and existing ownership vocabulary

A disposable native Git probe confirmed separate worktree HEAD logs and checkout/rename records. Worktree creation alone does not name its initial branch; require a consistent transition chain, not an assumed initial issue. Same-second records retain physical log order; contradictory or unsupported evidence stays unknown. The current issue vocabulary explicitly says terminal statuses retain attribution identity but hold no lock. Therefore claim cardinality must use holdable status history, while historical owner evidence may still exclude a witnessed foreign-branch tail. Codecomplete stays holdable. This derives an existing lifecycle rule rather than changing it; DeriveEngagements and adopted replay semantics must remain unchanged. The pure BranchClaim input now distinguishes terminal attribution from a live reservation, with a red/green control.

### 2026-10-23 — native grammar and approved standalone semantic change

The committed native worktree test exposed two real format details synthetic records lacked: the initial creation row can omit its tab/message, and rename emits an adjacent delete/create pair through a zero OID. Accept only initial zero-old creation without a message and exact same-time, same-message, identity-preserving rename pairs; incomplete or contradictory pairs stay unavailable. Native branch creation/rename/detach/reset/fresh-main-branch controls are green. Merge/rebase, expiry, race and broader negative controls remain.

The initial plan incorrectly assumed active-time had an ownership-aware route. Source inspection showed it is a standalone raw-dir/window/multiple-issue calculator. TL explicitly approved changing tracked-repository defaults: all issues use this same branch/claim classifier; explicit --dir, --since and --until narrow rather than bypass eligibility; multi-issue output partitions source intervals before allocation. No-tracker behavior remains unchanged. State this semantic change plainly in help and atlas. Do not retain raw tracked mode as the default. This supersedes the earlier wording about an existing owned issue route. Approval is authenticated and recorded before standalone wiring.

### 2026-10-23 20:12 PDT — approved snapshot compatibility boundary

The branch filter exposes an execution-span case: a task dispatched during the original ownership/lineage engagement can continue into the selected later branch interval. Persist original DispatchWindows separately from final selected Sources.Windows, and perform dispatch admission before final branch clipping. An older engine ignores the added dispatch input and silently drops that later span. This violates exact replay equivalence even though it cannot widen foreign time. TL approval (submitted receipt 7398f142-811e-4267-bc4d-30d5c623c953) chooses a NEW method/version for branch-classified snapshots; old readers must refuse, since silent undercount corrupts velocity calibration. Existing adopted v1/v2 snapshots and their algorithms remain unchanged. The hub rebuilds the primary binary immediately after landing; no staging or live adoption is authorized by this checkpoint. This supersedes the original no-version-bump proposal for new branch snapshots.

Implementation concepts now use BranchInterval, BranchClaim and BranchNote in internal/activetime/branch.go, plus the separate native decoder branchhistory.go. MeasurementBranchEvidence retains capture identity and labelled per-source attribution intervals. Missing competing card history is uncertainty, never absence: it prevents sole-reservation fallback, while a directly witnessed candidate branch still requires the candidate's complete ownership and admitted source lineage. Existing queue-head classification controls caught and enforce this distinction.

### 2026-10-23 20:35 PDT — completed implementation and verification boundary

New branch candidates now use schema version 3 with branch, branch-queue, branch-hybrid and branch-hybrid-queue method names. Existing v1/v2 replay remains on its original interpretation. Original dispatch windows are saved separately from final selected windows, and all consumers validate issue/witness/selected-window consistency; contributor pruning preserves the new evidence. The compatibility test now requires the frozen old reader to refuse v3, superseding original old/new acceptance equivalence. Unknown nested fields still survive current-reader rewrites. Explicit standalone limits are applied after original source admission and before branch classification. Existing authoritative tracker fetches remain part of that shared resolver; the classifier/capture itself adds no network or operator-worktree writes.

Native merge, rebase-return, resting and fresh main-based branches qualify the decoder. A proposed symbolic-ref-without-message counterexample was refuted: this Git writes a record; the retained control verifies that unsupported action never backfills B onto A’s earlier work. No parser re-plan or new mechanism resulted. Private append/replacement/rewrite races, malformed/truncated/oversized histories, cancellation, shared-session A/B partition, three-harness child/root lineage, missing evidence, explicit narrowing and source-append/live-log-expiry replay are covered. The planned private exact-executable control and disabling-selector overlay remain before the recorded suite. The full review at close may upgrade this beyond the quick diff shell; no artificial milestone is introduced.
