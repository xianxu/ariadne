---
collaboration-mode: discussion
---

# Drop one composed train car implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution approach. Preserve source HEAD/porcelain during cmd/sdlc tests. This is one atomic landing-gate change with one close boundary; plain checkboxes, no Mx tags.

**Goal:** An owner or repository :0 can durably exclude a composed car before merge, stop the obsolete train safely, and let survivors re-form without that car.

**Architecture:** Keep landingline.Transition as the train state authority. Publish its live generation/phase/membership in one bounded control record under line.lock. Drop acceptance and entry into PhaseMerging use that same mutex; neither command output nor polling authorizes a merge. Interrupt and join the old conductor's work, then use the existing held/Aborted release/ClosePR path from #521. Survivors rebuild from their exact original heads; the dropped runner exits and stays out until explicit retry.

**Tech stack:** Go, existing landingline flock/strict JSON/atomic-file seams, context cancellation, real temporary Git repositories, trainFakeGH/ciFakeGH and native local-check children.

## Existing behavior and alternatives

#521 already releases survivors when their conductor is held, but phases currently live only in conductTrain. The line has a process-lifetime head.lock, a short line.lock, waiting tickets, readiness/outcomes and 30-day event logs. Its head identity and ticket sequence fence one conductor generation. There is no landing/train vocabulary CUE model; existing phases are defined in internal/landingline/trainstate.go, which remains their source.

Recommended: abort/re-form the obsolete train, preserving survivors and existing validation rules. In-place surgery of a running train PR duplicates build/bisect/checked-head accounting and keeps a non-riding conductor's suite ownership problem. Merely drafting/pushing the car's PR is an observation that does not revoke the pinned original head; rejecting at an unlocked poll leaves a merge race. Both alternatives are rejected.

Recomposition changes the train's checked head: rerun suite, selected landing checks and CI for that new head; retain all existing exact-head reuse rules when a record genuinely matches. Never reuse the old three-car suite for two cars. Closing the old train and re-forming is a real recomposition, not a rewrite claiming its previous evidence still applies.

## Core concepts

| Name | Kind | Lives in | Status | Purpose |
|---|---|---|---|---|
| CarsDropped event / Transition | PURE | cmd/sdlc/internal/landingline/trainstate.go | modified | Hold requested members, release survivors through shared abort effects; reject terminal/merge-started events |
| TrainControl and drop decision | PURE | cmd/sdlc/internal/landingline/traincontrol.go | new | Typed generation, membership, phase, irreversible merge-start marker and pinned drop requests |
| Train control storage / merge fence | INTEGRATION | cmd/sdlc/internal/landingline/traincontrol_io.go | new | line.lock, existing head.lock lifetime and one bounded atomic JSON record |
| landing drop command | INTEGRATION | cmd/sdlc/landingdrop.go; cmd/sdlc/landingcmd.go | new / modified | Resolve actual caller/target repository and owner, accept or refuse, show recovery |
| conductTrain | INTEGRATION | cmd/sdlc/landingtrain.go; cmd/sdlc/landing.go | modified | Publish state, watch requests, cancel/join effects, serialize merge start and perform held/abort cleanup; retry a released surviving conductor with a fresh normal ticket |
| Line admission and sidecar sweep | INTEGRATION | cmd/sdlc/internal/landingline/line.go | modified | Reconcile accepted exclusions before new admission/control replacement after conductor death |
| trainFakeGH / ciFakeGH / real Git setup | INTEGRATION | cmd/sdlc/landingtrain_test.go; cmd/sdlc/landingdrop_test.go | modified / new tests | Existing stateful PR/ref/status records and real merge reachability, controlled effect barriers |

New pure decisions have unit tests without IO. Storage/CLI/conductor use real portable temporary folders and existing injected GH seams; no production configuration or new daemon. Control-record owner is the current head.lock generation, not a PID-only observation.

## Protocol and authority

- One fixed-size-family record, train-control.json in the line directory, protocol version1; includes train ticket/generation, phase, current composed cars (ticket, issue, branch, exact head, PR, owning checkout), accepted requests keyed by that exact car identity, and whether merge has ever started. No per-poll journal or per-request file family. Snapshot membership comes from the actual Transition result, not the candidate banner.
- Register before composition, tied to the held head ticket. A drop only accepts a current composed member, never a merely waiting or discarded candidate. Duplicate issue/member ambiguity refuses. Car owner provenance comes from its live ticket's QueueRoot (self from conductor root); resolve the caller's actual checkout separately from --repo. Reuse workspace identity/canonical root rules: same repository plus same owning workspace (or exact unnumbered checkout). Missing owner provenance refuses workers and names :0 recovery. Repository :0 is additionally allowed, with the existing Couch-tree worker-entering-:0 restriction; unknown identities/foreign repositories refuse. No claiming or reassigning ownership as a side effect.
- CLI: sdlc landing drop --issue N [--reason WHY] [--repo DIR]. Default reason “requested by owner or steward” makes the requested minimal command valid. Resolve authority outside line.lock; under line.lock revalidate pinned train/member/head generation before committing. Cooperative same-machine trust, like grants; local JSON is not an adversarial-user security boundary. Never infer authority from the --repo argument.
- Under line.lock, accepting a drop atomically records its pinned exclusion only if merge has not started. This durable exclusion is the safety commit. Write the held outcome in the same locked operation where possible; any later write failure reports “accepted; cleanup pending”, never “nothing happened”. The merge fence always checks the exclusion record, so a partial notification cannot permit landing. Do not nest line.Setttle/Record (which reacquire line.lock); use existing internal locked write helpers.
- Retry of the same request reports the same accepted decision without a second effect. Several requests may coexist; each current member can be dropped once. Request cannot drift to a new ticket/head for the same issue. Missing, truncated, duplicate-key, oversized, unsupported-version or mismatched live control refuses, never guesses serial/no-train success.
- Existing head.lock is the lifetime resource. Under line.lock, validate it is held by the recorded head generation. A dead head makes the control stale, not active. Before admitting another head or replacing a stale record, materialize every accepted dropped waiter's held outcome and exclude it from admission; only retire requests once the corresponding ticket is gone. Failure to reconcile blocks admission with a named recovery, rather than re-board an accepted drop. New tickets from explicit retries are distinct identities and unaffected.
- Atomic merge boundary: when a green CIDone would enter PhaseMerging / enqueue its Merge effect, conductor acquires line.lock and either sees accepted drops (process them instead) or commits merge-started plus phase. This is the start of the merge operation, before its final fetch/API; status says “merging (merge operation started)” without claiming the API was sent or GitHub confirmed it. Once started, requests refuse even if the operation later reintegrates/probes; report current phase plus this irreversible fact. This conservative boundary is intentional and needs hub approval.
- No network, Git fetch, suite work, or slow cancellation/join while line.lock is held. Existing repo-lock → short line-lock ordering at the fence is allowed; command never holds line.lock while acquiring repo lock. No new lock held throughout a suite.

## Conductor extent, state and failures

After registration, one scoped watcher reads requests at the existing landing poll cadence (2s production; injected barriers in tests). All long effects use a child context. On a drop, cancel that child context and join the current suite/CI/build effect before cleanup and releasing head.lock. Watcher exits and is joined on every return. Cancellation is addressed through the owned effect/context, not foreign local-check cancellation or ps-selected signals.

Then feed CarsDropped into Transition: settle all dropped cars held with reason, use shared Aborted cleanup to release only surviving riders and close/delete the old train PR/ref. If the conductor's own car is dropped it exits held. If another car is dropped, after confirmed old-train cleanup the conductor returns a typed re-formation result. Add a nonrecursive wrapper around the existing runDurableMerge attempt: complete that attempt's existing deferred line/queue cleanup, then retry the ordinary attempt from its first observations, obtaining a fresh ticket/queue operation and revalidating normal gates. Other survivors remain queued and may become the new conductor first; the old conductor joins as an ordinary survivor. Only this typed result after an actual accepted drop authorizes an automatic retry; generic errors do not. Caller context bounds the loop; no stale PR/head/turn or admitted context is retained between attempts.

Interrupted effect completions never feed obsolete SuiteDone/CIDone into the old machine. Cleanup failures remain visible as drop accepted / cleanup pending; no old-head merge is allowed. Closing PR/ref must be confirmed before survivor formation can land. Existing closePR (landingtrain.go:647) and reclaimTrains (:715) warn and continue on failure, and the serial fallback (:819) skips reclamation: they do not implement this barrier. Add failure-aware closeDroppedTrain(ctx, control) and reconcileDroppedTrain(ctx, turn) instead of treating those best-effort routines as proof. Persist a cleanup obligation (train/branch, repository/remote, expected pushed tip, known PR number, publication-started flag) in the same control safety record at acceptance; publication-started/identity updates happen before external publication, under the same drop fence. After cancel/join, closure requires a positive GH CLOSED observation for the known/discovered exact train PR, or local proof publication never started, plus absence of the exact leased remote train ref. A failed/lost close reply is probed; unknown/ambiguous publication or close stays pending, never inferred absent from a lagging listing. If publication started with an unknown PR identity, repeatedly inspect the exact train branch through the bounded existing GH seam; no positive closure means stop with pending status and the exact original conductor merge command, without claiming recovery finished. No automatic bypass or administrative force-ack is added.

Every newly admitted runDurableMerge attempt invokes reconcileDroppedTrain before PR selection or serial/train dispatch, including one-car, trains-off and grant paths. It inherits head.lock, reads the prior generation's pending obligation, performs GH/ref cleanup outside line.lock, then commits confirmed completion under line.lock after rechecking the generation/obligation. Failure stops that attempt and frees the head through existing defer; a later attempt retries reconciliation. Resume-only already-merged cleanup issues no merge and does not bypass this gate for any unmerged PR. Stale record replacement/retirement refuses while cleanup is pending even if all old tickets died. An admission may reserve the head to reconcile, but cannot board/publish/merge a successor until confirmation; accepted dropped waiters are excluded before even that admission. Thus a dead conductor cannot erase either the exclusion or cleanup obligation. A dropped waiter's process can die: its pinned request remains until ticket retirement; no old drop affects its new explicit retry.

| State/event | Next state and effects |
|---|---|
| Composed pre-merge + authorized drop wins lock | Accepted exclusion; cancel/join effect; held dropped members; shared abort releases survivors; old PR closed; new generation composes survivors |
| Drop pending + green suite/CI finishes | Drop takes precedence; ignore old completion; no Merge effect |
| Green CI wins lock before drop | Merge-started committed; drop refuses with exact phase/start marker |
| Merge uncertain / reintegrating after start | Drop refuses; existing probe/reintegration rules continue, no claim of unmerged |
| Second valid member drop before cleanup | Accumulate pinned exclusions; hold both, survivors only re-form |
| Request retry after acceptance | Same accepted/pending/completed receipt, no duplicate cancellation |
| Conductor dies after accepted write before notification | Head lock dies; admission reconciles exclusions first, then existing orphan cleanup/re-formation |
| Missing/old/corrupt control | Refuse drop with unavailable/unsupported state; never claim exclusion |
| Control/status write fails before merge start | Stop train safely; no destructive fallback merge |

## Visible state and recovery

Actions: inspect status; drop a composed member; explicitly retry the dropped branch after resolving its reason; finish a merge already started; retry cleanup/re-formation through the conductor's printed merge command.

| Internal state/failure | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Accepted exclusion, effect stopping | “#N dropped from train T; cleanup/recomposition pending: WHY” with pinned branch | Inspect sdlc landing status | Old train suite progress can be discarded; code/history preserved |
| Dropped runner receives held result | “#N held: WHY” and exact sdlc merge --branch BRANCH --yes recovery after fixing/re-closing | Explicit owner retry | Its queued place ends; no code/history lost |
| Survivors released/new generation | Composed membership names only survivors and new head | Allow existing survivor runners to continue | Necessary new-head suite/check/CI cost |
| Merge started / unknown / landed | Refusal names train, current phase, started marker and known outcome; never promises it can be removed | Finish/inspect that merge, then repair separately if landed | Drop cannot prevent the committed merge operation |
| Acceptance persisted but cleanup failed/dead conductor | “drop accepted; cleanup pending” plus exact conductor recovery command | Resume cleanup/re-formation command | Discarded suite progress; exclusion retained |
| No active compatible train/member | Refusal names absence, unsupported or unreadable state and says no drop accepted | Inspect status / coordinate conductor restart on current binary | None from the refusal |
| Foreign/unknown caller | Ownership refusal names permitted owner or repository :0 | Have that owner/:0 run exact drop command | None |

## Implementation tasks and test strategy

### 1. Pure core and record contract

Files: internal/landingline/trainstate.go, traincontrol.go and colocated tests; codec.go only for shared nested preservation.

- [x] TDD CarsDropped through Transition and named pure decideTrainDrop / decideTrainMerge / validateTrainControl functions. Sequence/property tests generate interrupting drop/completion/merge events; invariant: accepted members never appear in a Merge effect, survivor FIFO/exact heads are preserved, and rejection leaves state unchanged.
- [x] TDD DecodeTrainControl / TrainControl.MarshalJSON through malformed-record fuzz seeds and round-trip properties; mechanical guards are strict known fields/duplicates/version/size/identity and unknown top-level/nested preservation.
- [x] Implement minimal pure transitions and codec; go test ./cmd/sdlc/internal/landingline -count=1; commit passing core.

### 2. Storage, authority and merge fence

Files: internal/landingline/traincontrol_io.go, line.go and colocated tests; cmd/sdlc/landingdrop.go, landingcmd.go and landingdrop_test.go.

- [x] TDD Line.BeginTrain / UpdateTrain / RequestTrainDrop / StartTrainMerge / FinishTrainControl using real temporary folders and deterministic two-actor barriers; adversarial crash/retry/generation sequences must preserve a single durable decision and forbid a drop/merge double winner.
- [x] TDD Line.reconcileTrainDrops and replacement/admission using fault injection after the safety write; the mechanical guard excludes pinned dropped tickets before candidate selection and preserves pending cleanup across dead heads and retries.
- [x] TDD authorizeTrainDrop and runLandingDrop with actual-caller/target repository separation and workspace provenance; caller/target mutations must never widen owner/:0 authority. Missing identity/control refuses without effects.
- [x] Implement methods under existing line/head locks, shared atomic codec and strict bound; targeted landingline/command checks; commit.

### 3. Production conductor and survivor landing

Files: landingtrain.go, landing.go, landingdrop_train_test.go and reusable trainFakeGH/ciFakeGH barriers.

- [x] TDD conductTrain through a deterministic real-Git three-car scenario and a native blocked old suite: actual command accepts drop, obsolete effect joins, old PR/ref closure is confirmed, survivors automatically re-form a new head and land; Git reachability proves dropped commit absent, survivors present and dropped PR OPEN. This is the operator-approved live conformance, including dropped-conductor behavior.
- [x] TDD conductor drop watcher / effect admission with forced completion-vs-drop orderings, native cancellation and foreign live leases; mechanical guard is context cancellation plus lexical join, followed by the atomic merge fence, never timing luck.
- [x] TDD closeDroppedTrain / reconcileDroppedTrain against stateful GH lost/failed/lagging observations and restarted real-folder state; enforce no successor publication/merge before positive cleanup proof through both serial and train paths.
- [x] TDD runDurableMerge attempt wrapper through the typed re-formation result; observe a fresh queue operation/ticket and re-read heads/gates, with no automatic retry on ordinary errors or dropped self.
- [x] Implement scoped watcher, control updates, failure-aware cleanup and retry wrapper; run focused train/lease/serial controls and commit. Retain all failed attempts; classify rather than blind-retry.

### 4. Documentation and acceptance evidence

Files: cmd/sdlc/helptext/landing.md, README.md, atlas/workflow/sdlc-binary.md, issue Log.

- [x] Document command/authority, merge boundary, pending-cleanup state/recovery, held explicit retry and unsupported old conductors; status renders actual current state and exclusions.
- [x] TDD counted control/membership operations at8/16cars: O(N+D) walk, one short atomic update per event, one scoped watcher and existing2s cadence; reject per-member IO/poll multiplication. Run targeted checks/lint and push.
- [x] Run and persist operator-approved deterministic real-Git conformance, raw red/green/race logs and recorded full suite in common test-stats/evidence/000526; no staged Pair train or configuration change.
- [x] Update Log/plan ticks and prepare close evidence; sdlc close owns review, hub owns project (--no-project); on SHIP run sdlc merge --yes and report.

## Operating envelope and architecture

ARCH-DRY: reuse line.lock/head.lock/strict codec/atomic writes, held outcomes, abort/release/close and normal formation; reuse workspace ownership. ARCH-PURE: decision and CarsDropped transitions contain no IO; storage/CLI/watcher remain injected shell. ARCH-PURPOSE: held car cannot re-board automatically; two surviving cars actually land, including dropping the original conductor. ARCH-MOCK: existing stateful GH fake and real Git/native process portable fixtures; staged live conformance selected by hub.

ARCH-CONSTRAINTS: interactive control accepts after one short local transaction; effect observes within the existing2s polling interval plus bounded native cancellation/join. No suite-duration lock, no network under line mutex, O(N+D) record work and <=N requests for Ncurrent members; record <=existing1MiB boundary. CPU/memory/disk bounded by current train, not past trains. No new parallel suite or machine setting. On storage contention/errors report pending/refusal, never admit merge with unavailable control evidence. Counted8/16 tests reject accidental per-car polling or repeated scans; timings are diagnostic.

ARCH-SECURE: cross-process control and ticket roots are untrusted typed data with exact generation/head/identity checks; actual caller differs from --repo target. No new credentials. Malformed/older control cannot authorize drop/merge; rollout refuses unsupported active conductors rather than assuming they implement fencing. Trust remains cooperative same-machine.

ARCH-ORDER: one durable safety decision under the same mutex as merge start; process lifetime is existing head.lock. The event table above defines cancellation, death, concurrent requests, duplicate/retry and uncertain merge outcomes. Watcher/effect extent is lexical and joined. Most likely missed interleaving: suite/CI completion racing an accepted drop; controlled tests run both orders.

ARCH-FUNERAL: one bounded control record per repository, owned by active head; terminal record retained until dropped outcomes reconciled/tickets retired AND pending cleanup positively confirmed, then cleared or replaced on next generation. No unbounded tombstones. Audit uses existing EventCar/EventTrain logs and30-day retention. Train PR/ref/worktree retain existing cleanup/reclaim. Temporary watcher/context/children die and join before return.

ARCH-EVOLVE: new protocol record only; no reinterpretation of old outcomes/tickets. Strict known fields and duplicates, preserved unknown top-level/nested fields, explicit unsupported-version refusal because an old conductor lacks the merge fence. Existing no-control conductors remain operational but cannot claim drop support. No version bump of unrelated formats.

## Approval and revisions

The merge start fence is admission of the Merge effect after green CI, before its final fetch/API. Once that operation starts, refusal is conservative even if it later observes main movement; this avoids pretending a request can undo an operation already committed to merge. Operator approved this boundary and the deterministic real-Git fixture as live conformance through hub relay2364fcf8-71f4-477d-804e-c5fa4d1a4d45; no staged Pair train.

- 2026-10-21: Replace the open live-check choice with the operator-approved fixture and record merge-start approval. No change to the protocol or scope.

- 2026-10-21: Plan-quality PQ-1 clarified the already-promised cleanup barrier: current close/reclaim is best-effort, so persist the obligation, confirm through a failure-aware adapter, and gate every successor attempt including serial paths before PR selection. PQ-2 replaces case inventories with named risky functions and adversarial strategies. These refine the approved pending-cleanup behavior; no new operator action or bypass.

- 2026-10-21 handoff discovery: deterministic real3runner fixture proves the approved release-before-requeue wrapper can split ready survivors into separate serial landings. Proposed revision, pending hub approval: outer wrapper retains old turn's head.lock while the old queue operation retires; new ordinary attempt enqueues/vouches its fresh ticket before releasing that retained turn, with lexical fallback release on cancellation/early failure. It then uses unchanged normal Wait/Board rules. No timing delay or new persisted hold. This replaces only the release timing in the surviving-conductor handoff; dropped conductors still exit held.

- 2026-10-21: Hub approved head-lock handoff revision (1139e7ac-1690-4829-bbaf-df866fed2fea). Outer wrapper retains the old turn across old queue retirement; next ordinary attempt enqueues and synchronously vouches before releasing it, with fallback release on every early error/cancel/return. Both real conformance variants continue to require the two survivors in one train.

- 2026-10-21 verification: both real three-car drop variants land two survivors in one replacement train. Broad targeted train/drop/grant/cancellation controls pass487.849s; raw red/green cleanup, destination and handoff evidence is recorded in the issue Log and SSD test-stats/evidence/000526. Final recorded suite and boundary review remain pending.

- 2026-10-21 final controls: reject ordinary branches before any cleanup GitHub effect and render repeated receipts from actual pending/completed state. Both reproduced red before correction. Final shell controls71.797s, core/lint/race/vet and deterministic deadline controls pass; recorded full suite remains pending.

- 2026-10-21 acceptance: recorded branch suite on957b61af passed all8shards/1940command tests/75other packages in793.3s with zero failures; persisted branch-suite.log. Implementation and acceptance preparation complete; boundary review and merge remain the next gated actions, with project bookkeeping owned by hub.
