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
| conductTrain | INTEGRATION | cmd/sdlc/landingtrain.go | modified | Publish state, watch requests, cancel/join effects, serialize merge start and perform held/abort cleanup |
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

Then feed CarsDropped into Transition: settle all dropped cars held with reason, use shared Aborted cleanup to release only surviving riders and close/delete the old train PR/ref. If the conductor's own car is dropped it exits held; if a different car is dropped the conductor exits released, freeing the head so survivors (including it only after explicit runner retry where existing behavior requires) may re-form. The integration must ensure survivors actually continue without manual operator rescue: for a surviving conductor, requeue/resume its own merge automatically after clean release, or conduct the new surviving train in a fresh generation with a new legitimate owning head. Choose the existing release/re-admission mechanism, not a fake self-settled ticket. This last wiring is part of the deliverable, not deferred.

Interrupted effect completions never feed obsolete SuiteDone/CIDone into the old machine. Cleanup failures remain visible as drop accepted / cleanup pending; no old-head merge is allowed. Closing PR/ref must be confirmed before survivor formation can land; unknown close retains the existing orphan recovery machinery. A dropped waiter's process can die: its pinned request remains until ticket retirement; no old drop affects its new explicit retry.

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

## Tests and implementation tasks

### 1. Pure transitions and strict control record

Files: internal/landingline/trainstate.go, traincontrol.go, trainstate_test.go, traincontrol_test.go, codec.go only if needed for reusable nested preservation.

- [ ] Write red sequence tests across composing/testing/bisecting/checking/landing, dropped conductor/non-conductor, multiple requests, no survivors, duplicate/stale/foreign member, terminal/merge-started rejection. Assert original exact heads excluded, survivor order retained, no Merge after accepted drop.
- [ ] Add strict versioned codec tests: truncated/oversized/duplicate/wrong types/missing fields/unknown versions, unknown top-level and nested members preserved on read-modify-write. Data invariants live in one pure decision, not duplicated in CLI and conductor.
- [ ] Implement CarsDropped and shared abort/release effects; no separate mini train state machine or new outcome kind required. Run go test ./cmd/sdlc/internal/landingline -count=1; commit passing pure core.

### 2. Shared storage, authority and atomic fence

Files: internal/landingline/traincontrol_io.go, line.go, new traincontrol_io_test.go; cmd/sdlc/landingdrop.go, landingcmd.go, landingdrop_test.go.

- [ ] Red real-folder lock tests force both orders of drop vs merge start (channels/barriers, no sleeps). Exactly one wins; accepted drop can never authorize merge. Exercise duplicate request, simultaneous member requests, stale old generation, lost response retry, crash after safety write before held notification, restart/admission reconciliation and malformed record refusal.
- [ ] Implement register/update/request/start/finish methods under the existing line mutex and live head capability. Prevent replacement/admission from forgetting accepted drops. Preserve unknown record fields through every update; use existing maxTicket/writeAtomic bounds; refuse over capacity rather than growing indefinitely.
- [ ] Add command tests for own car, same-repo :0, foreign worker, fake --repo authority, worker entered into :0, unknown owner, missing issue, absent/non-composed car and exact state refusals. Identity preparation is outside the mutex; commit pins revalidated inside.
- [ ] Run targeted landingline + command tests, commit.

### 3. Production conductor, cancellation and actual survivor landing

Files: landingtrain.go, landingqueue.go only for genuine self re-admission/resume; landingdrop_train_test.go plus reusable trainFakeGH fixture barriers.

- [ ] Red three-real-car test blocks the old train's native suite, drops the middle car through the actual command/service boundary, joins the obsolete effect, closes the old PR, and observes two survivors automatically re-form and merge on a new head. Git ancestry proves dropped commit is absent, both surviving commits present, dropped PR OPEN, survivors MERGED. Assert only new exact-head validation reused/rerun, no old three-car suite accepted.
- [ ] Repeat dropping conductor; surviving runners re-form without foreign suite ownership. Include two simultaneous dropped members / one survivor and all dropped. Capture exact dropped runner reason and branch recovery promptly.
- [ ] Force drop on last CI completion vs Merge effect admission in both orders. Pause inside LandingMerge to prove late drop refuses merging; lost merge reply/failed probe refuses merge-unconfirmed, confirmed landed refuses landed. No test depends on scheduler luck.
- [ ] Interrupt queued/running suite, CI and build through their contexts; prove cancellation joins owned children and keeps unrelated local-check leases alive. Crash after acceptance before notification and retry into a new ticket demonstrate durable exclusion without poisoning the retry.
- [ ] Wire scoped watcher + control updates + atomic effect-admission fence and shared abort/re-admission flow. Run targeted train/lease/single-car controls; commit. Observe all failures before classifying; do not blind-retry.

### 4. Documentation, conformance and close evidence

Files: cmd/sdlc/helptext/landing.md, README.md, atlas/workflow/sdlc-binary.md (and atlas index only if adding a new page), issue Log.

- [ ] Document command, authority, irreversible merge boundary, held retry/recovery, new bounded record and rolling-upgrade refusal. Update status output to show accepted exclusions, actual current phase and pending cleanup without changing existing human membership promises.
- [ ] Count control updates/reads and pure membership walks at8/16cars: one read/atomic rewrite per control event, O(cars+accepted requests) decision, no per-member Git/network under line.lock. Watcher uses existing2s cadence, at most one watcher; no tight polling. Run focused suites/lint then push and record full suite.
- [ ] Request hub's chosen live conformance: stage --build only with authorization, then an actual Pair three-car train drop before merge (or hub-approved equivalent controlled fixture); owner/:0 and late refusal exercised against the staged binary, same revision recorded. Ariadne trains are off; do not alter landing configuration just to stage. Hub owns live choice and next natural train follow-up.
- [ ] Persist raw red/green/controlled-race/full-suite logs outside RAM disk in common test-stats/evidence/000526, update Log and prepare sdlc close --no-project evidence. On SHIP enter sdlc merge --yes and report PR/outcome to hub.

## Operating envelope and architecture

ARCH-DRY: reuse line.lock/head.lock/strict codec/atomic writes, held outcomes, abort/release/close and normal formation; reuse workspace ownership. ARCH-PURE: decision and CarsDropped transitions contain no IO; storage/CLI/watcher remain injected shell. ARCH-PURPOSE: held car cannot re-board automatically; two surviving cars actually land, including dropping the original conductor. ARCH-MOCK: existing stateful GH fake and real Git/native process portable fixtures; staged live conformance selected by hub.

ARCH-CONSTRAINTS: interactive control accepts after one short local transaction; effect observes within the existing2s polling interval plus bounded native cancellation/join. No suite-duration lock, no network under line mutex, O(N+D) record work and <=N requests for Ncurrent members; record <=existing1MiB boundary. CPU/memory/disk bounded by current train, not past trains. No new parallel suite or machine setting. On storage contention/errors report pending/refusal, never admit merge with unavailable control evidence. Counted8/16 tests reject accidental per-car polling or repeated scans; timings are diagnostic.

ARCH-SECURE: cross-process control and ticket roots are untrusted typed data with exact generation/head/identity checks; actual caller differs from --repo target. No new credentials. Malformed/older control cannot authorize drop/merge; rollout refuses unsupported active conductors rather than assuming they implement fencing. Trust remains cooperative same-machine.

ARCH-ORDER: one durable safety decision under the same mutex as merge start; process lifetime is existing head.lock. The event table above defines cancellation, death, concurrent requests, duplicate/retry and uncertain merge outcomes. Watcher/effect extent is lexical and joined. Most likely missed interleaving: suite/CI completion racing an accepted drop; controlled tests run both orders.

ARCH-FUNERAL: one bounded control record per repository, owned by active head; terminal record retained only until dropped outcomes reconciled/tickets retired, then cleared or replaced on next generation. No unbounded tombstones. Audit uses existing EventCar/EventTrain logs and30-day retention. Train PR/ref/worktree retain existing cleanup/reclaim. Temporary watcher/context/children die and join before return.

ARCH-EVOLVE: new protocol record only; no reinterpretation of old outcomes/tickets. Strict known fields and duplicates, preserved unknown top-level/nested fields, explicit unsupported-version refusal because an old conductor lacks the merge fence. Existing no-control conductors remain operational but cannot claim drop support. No version bump of unrelated formats.

## Open approval point

The merge start fence is admission of the Merge effect after green CI, before its final fetch/API. Once that operation starts, refusal is conservative even if it later observes main movement; this avoids pretending a request can undo an operation already committed to merge. Hub must approve this precise boundary and the live conformance choice before implementation.
