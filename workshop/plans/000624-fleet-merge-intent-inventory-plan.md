# Fleet merge-intent inventory implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution strategy. Use superpowers-executing-plans for the existing shared observer and the consumer coordination. Steps use checkbox syntax; this is one atomic close boundary, not multiple milestones.

**Goal:** One read-only fleet inventory reports each physical merge intent once at its exact member checkout, with explicit read quality, so pair#848 can describe interrupted landings without reading SDLC files.

**Architecture:** Enrich the existing collected inventory in the command layer, grouping its already canonical Git-common identities. Reuse the existing merge-intent observation, classification and command construction owner; promote its report shape into the fleet wire types rather than copying it. Keep store IO separate from a pure root/identity placement function. No intent migration, sweep, queue mutation or recovery execution.

**Tech stack:** Go, existing fleet JSON, landingline intent codec/flock store, real disposable Git and stateful test doubles.

## Source and decisions

The dispatch requires a plan to ops:0 before change-code (authenticated receipt e9543a32-7d72-4df4-83a0-4c4ee9a25619). Claim and start-plan are complete. No implementation or tests have been changed.

Read pair#848's approved plan from its issue branch, since pair:5's current checkout is on other work. Pair:5 confirmed the wire and semantic choices in authenticated receipt bba6d841-4e31-400d-a24a-71d4b4648b3e: live/orphaned/stale/unreadable records; complete/partial/unreadable read quality; capabilities.merge_intents=1; physical record identity; Root-only placement; diagnostics once. Terminal removal is complete plus an empty array on the next observation, not a new terminal tombstone. The consumer remains parked until this producer lands.

Existing authorities:

- cmd/sdlc/mergeintent.go owns reportMergeIntent, observeMergeIntents and exact re-entry construction. state and landing status already call it.
- cmd/sdlc/internal/landingline/intent.go owns preserving v1 decoding, nonblocking shared-lock observation, exclusive owner liveness, interrupted preservation and terminal removal. Its reader bounds each record to maxTicket (1 MiB).
- cmd/sdlc/internal/fleet/inventory.go already groups Git worktrees by canonical common directory in RepoIdentity and includes declared dependency clones. Slot membership is already represented by inventory.Slots[].Members; do not derive a competing slot identity from a basename or issue number.
- runFleetInventory currently emits the collected Inventory unchanged. Enrichment belongs immediately after collection, before either JSON encoding or human rendering. Existing fact/claim/readiness classification stays intact.

Alternatives: per-slot state calls repeat Git/tracker/process work and cannot deduplicate shared stores; direct Couch file reads duplicate the intent lock/codec authority. Reusing the command-layer observer and adding one grouped adapter is the smallest design that supplies the approved consumer contract.

## Wire contract (schema_version remains 1)

- Add capabilities.merge_intents: 1 to inventories actually enriched by this producer. An absent capability in an older payload remains absent on decode/re-encode; never manufacture support in withDefaults.
- Every capable row emits merge_intents as a non-null array, merge_intents_state as complete, partial or unreadable, and merge_intents_error for either incomplete state. Complete has no error; complete plus [] means this row has no observed intents at this completed read.
- Each report retains state, branch, issue, head, root, workspace (optional legacy field), pid, started_at, reentry, note and path from the existing owner. Add identity, equal to the canonical physical record path under its canonical common-dir store. It is an observation/deduplication key, not a new lease or generation token.
- root is the validated record's canonical exact checkout. Placement requires the same common-dir identity and an exact canonical row root match. Slot ownership comes from that row's existing Slots/Members association; record workspace is supplementary and never redirects placement. A legacy record without workspace still assigns by Root.
- Re-entry is generated only by the shared validated classifier for an unlocked valid intent with an existing local branch. Its command remains sdlc merge --branch VALIDATED_BRANCH --yes, to be run in root. Live and stale carry no command. Branch-inspection failure is unreadable/inspection-only, never stale or orphaned by assumption.
- A corrupt record without a trustworthy Root, a root outside its store's inventory rows, or conflicting placement becomes one repository diagnostic with stage merge_intents, exact path and reason. Such uncertainty makes the affected store's rows partial while retaining safely assigned observations. No guessed row or executable command is emitted for the diagnostic.
- If store enumeration is unavailable, all rows sharing it are unreadable unless valid partial observations were returned, in which case they are partial and retained. Other stores remain usable. One assigned record's branch-read failure remains inspection-only on its row and is reflected in quality.
- Queue pause is a carried note, not a second admission authority. Read pause owners once per inventory observation and retain the existing literal note. If that read fails, orphan reports say pause information is unavailable; do not silently turn the failure into proof of an unpaused queue. The report still executes nothing.
- Unknown future fields are tolerated; known types, required discriminators and duplicate keys stay strict. For capable payloads reject null arrays, missing row quality, invalid quality/error combinations and malformed identity/root/reentry fields. Legacy payloads without capability remain decodable without claiming merge evidence.
- JSON is fully encoded before writing stdout and capped at 1 MiB, matching pair#848's existing source limit. Overflow returns a diagnostic error and no partial JSON document. There is no silent truncation or fabricated complete empty result.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| MergeIntentReport and inventory capability/read-quality wire | cmd/sdlc/internal/fleet/mergeintents.go; types.go integration | new (promoted existing report shape) |
| reportMergeIntent | cmd/sdlc/mergeintent.go | modified (same classification/command owner, adds physical identity) |
| placeFleetMergeIntents | cmd/sdlc/fleet_mergeintents.go | new |

MergeIntentReport is one typed observation of a physical intent, not proof a recovery command may still execute. The existing main-package mergeIntentReport becomes an alias to the promoted fleet type, retaining one field owner for all status surfaces. A record belongs to at most one canonical row; a row may have several branch intents. Independent clone stores remain distinct even for identical branch names.

placeFleetMergeIntents receives canonical rows and observed store results. It indexes (common directory, root), places by exact key, deduplicates by physical identity, propagates read quality and returns deterministic diagnostics. It performs no Git, filesystem, queue or process calls. Conflicting observations are inspection-only, never first-writer-wins. Its unit tests need values, not IO mocks.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| collectFleetMergeIntents and injected observer | cmd/sdlc/fleet_mergeintents.go; fleet.go deps | new | shared merge-intent observer once per common store |
| observeMergeIntents | cmd/sdlc/mergeintent.go | modified | existing List, branch lookup, pause observation; preserves partial results/errors |
| observeIntent/List error handling | cmd/sdlc/internal/landingline/intent.go | modified | distinguish ENOENT disappearance from permission/IO failure |
| Fleet JSON encode/decode | cmd/sdlc/internal/fleet/types.go and mergeintents.go; cmd/sdlc/fleet.go | modified | additive wire fields, validation and bounded stdout |
| Stateful merge observation fake | cmd/sdlc/fleet_mergeintents_test.go | new | store ownership/interruption/terminal/corruption/pause transitions |

Fake state is keyed by common store and record identity; operations hold, interrupt, finish, delete branch, corrupt and deny reads. Repeated public command calls see current state through the same injected observer seam. Reuse fleet's existing FakeGit for counted Git behavior and real disposable Intents for lock conformance; no fake process-liveness substitute.

The observer currently drops every os.Open failure as if the record disappeared. Preserve only confirmed ENOENT as disappearance; return an unreadable observation for other failures. Preserve successful observations when a later directory or branch read fails. The shared reader remains read-only and does not take a repo transaction lock, sweep, rewrite or remove records.

## Actions and visible state

Actions: rerun inventory; run the exact merge in its recorded root through normal SDLC admission; inspect/repair the named record/source; upgrade the producer; resume the queue through its existing owner.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| valid held owner lock | live, no re-entry command | rerun inventory after owner finishes | none |
| unlocked valid intent, branch exists | orphaned, exact root/command | run the named merge in root | none from report; merge revalidates authority |
| orphan while queue paused or pause read unavailable | orphaned with pause/uncertainty note | resume/repair queue through owner, then rerun | recovery delayed, no record removed |
| unlocked valid intent, branch gone | stale, no re-entry | inspect through ordinary SDLC lifecycle | none from report |
| terminal outcome removed intent | complete + [] if no others remain | no merge recovery action | no new tombstone/history promised |
| mixed valid and unreadable/corrupt records | partial, valid reports retained, diagnostics | inspect named source and rerun | enumeration incomplete, explicitly stated |
| store cannot be read | unreadable with reason, no absence claim | inspect/repair source and rerun | delayed recovery |
| branch probe fails for one record | unreadable observation, no re-entry | inspect source and rerun | no guessed branch state |
| invalid/unmatched root or conflicting identity | repository diagnostic once; affected quality partial | inspect exact path/root and rerun | no guessed owner |
| capability missing in old inventory | unsupported merge evidence; old facts retained | upgrade producer and rerun | consumers must not infer no merges |
| JSON exceeds consumer bound or caller cancels before publication | command error, no JSON published | repair size/cancellation source and rerun | unavailable report, no store changes |
| stdout write fails after publication begins | transport error, report untrustworthy | repair transport and rerun | receiver may have incomplete bytes, never a success claim |

No new state machine is persisted: each command produces a fresh observation and returns with no background work. Existing intent lease ownership is the single lifetime resource proving a live merge; the report does not claim permanent quiescence. A merge can start after an inventory read; its normal Hold refuses duplicate execution when a recovery action actually runs.

## Implementation and verification

### Task 1 — pin the wire and shared report owner

Files: create cmd/sdlc/internal/fleet/mergeintents.go and mergeintents_test.go; modify types.go, json_test.go and cmd/sdlc/mergeintent.go; add cmd/sdlc/testdata/merge-intents-v1/ fixtures.

- [x] RED tests: legacy inventory missing capability remains unsupported; capable complete empty emits []; null/missing/malformed collections and quality/error combinations refuse; duplicates and known-type validation remain strict; additive unknown fields do not invalidate reads.
- [x] Pin literal pre-change v1 intent JSON, optional workspace absence and unknown nested members. Decode with the existing preserving codec, verify read-only bytes unchanged, and retain its existing decode/rewrite preservation tests.
- [x] Promote the report DTO with a main-package alias; retain the existing classifier and command builder. Add stable physical identity without changing live/orphaned/stale semantics or the v1 writer.
- [x] GREEN targeted commands: go test ./cmd/sdlc/internal/fleet -run 'Test.*MergeIntent|TestJSON' -count=1; go test ./cmd/sdlc/internal/landingline -run '^TestIntent' -count=1. Use final exact test names once written.

### Task 2 — preserve uncertainty and assemble one observation per store

Files: create cmd/sdlc/fleet_mergeintents.go and fleet_mergeintents_test.go; modify fleet.go, mergeintent.go, internal/landingline/intent.go and intent_test.go.

- [x] RED lower-boundary tests: permission/IO open failure is not absent; confirmed disappearance is omitted; held lock remains live even if content is temporarily unreadable; a later failure retains earlier valid observations; branch lookup failure cannot become stale. Inject errors deterministically rather than relying only on chmod under privileged test users.
- [x] RED pure placement cases: shared linked worktrees; declared independent dependency clone; identical branch in separate stores; several intents on one root; optional/mismatching workspace cannot redirect Root; unmatched/corrupt root diagnosed once; valid+corrupt mixed store retains reports with partial quality; conflicting identity never supplies executable guidance.
- [x] Add the grouped command-layer adapter and defaults. Read queue pause state once, call the shared observer once per physical common directory, normalize roots once, then pure placement. Test callers injecting inventory also inject the stateful observer so no fake root accesses live state.
- [x] Drive fake hold → live, interrupt → orphaned, pause/unpause, terminal removal → complete empty, branch deletion → stale, denial/corruption → partial/unreadable. Assert repeated calls never mutate intent bytes or queue ownership.
- [x] Count at N=8 and2N=16: one observer/store, one root index/row, one placement/record and at most one local branch probe/readable record. Doubling rows sharing one store must not double store reads; reject per-row rescans and per-slot SDLC subprocesses.
- [x] GREEN targeted command: go test ./cmd/sdlc -run '^(TestFleetMergeIntents|TestMergeIntent)' -count=1; then the same focused names under -race. One targeted runner at a time.

### Task 3 — qualify the public producer and document rollout

Files: modify cmd/sdlc/fleet_integration_test.go and fleet_test.go; cmd/sdlc/helptext/fleet-inventory.md; atlas/workflow/sdlc-binary.md (existing indexed fleet section).

- [x] RED/GREEN private real-executable fixture: linked worktree plus independent dependency store, real held lock then released interrupted intent, terminal removal and stale branch, corrupt record and private queue pause. Observe stdout through the real fleet inventory command, decode with strict typed contract, and assert no store bytes or directory entries changed. Private Git/tracker and WF_SDLC_QUEUE_ROOT only; never pause/read intent fixtures in the production queue/store.
- [x] Bound JSON to1MiB before stdout. Boundary test exact fit and overflow with zero partial output; check cancellation before publication and propagate stdout write errors without claiming atomic transport. Other command errors stay stderr/nonzero. Keep one inventory command for the whole fixture fleet.
- [x] Qualify existing state and landing-status surfaces against the promoted report and conservative reader so the shared owner does not regress their command text or liveness semantics.
- [x] Document capability, row quality, physical identity, Root/slot-member association, diagnostics, absence of terminal tombstones, legacy rollout, consumer byte limit and read-only/no-sweep behavior. No Pair source changes or activation in this issue.
- [x] Ask pair:5 to run its real decoder against a retained private new-binary capture; if its implementation is not ready, retain that fixture and state the consumer limit. Propose private executable conformance plus the hub's ordinary post-land :0 rebuild; no live staging, grants or merge re-entry for qualification.
- [ ] Push candidate; run recorded sdlc local-check run --pre-close; satisfy the single binary-owned close review; land through central sdlc merge --yes. Tell ops:0 and pair:5 the Git-verified merge and final schema so #848 can proceed from main.

## Operating envelope and architectural review

- ARCH-DRY: inventory, state and landing status use the same report/classifier; Git common-dir identity, slot membership, preserving codec and queue ownership remain their existing authorities.
- ARCH-PURE: placement/validation are deterministic; store/Git/queue observation stays in a small injected shell. The fake models transitions, not calls returning canned lists.
- ARCH-PURPOSE: all physical stores including dependency clones are represented once, with honest partial/unreadable evidence and diagnostics for roots that cannot be assigned.
- ARCH-MOCK: stateful observer fake plus existing FakeGit; real flock and executable conformance use private file-backed repositories and queue state. No GitHub or new network dependency is introduced by enrichment.
- ARCH-CONSTRAINTS: batch read in the existing inventory command, not a UI hot loop. Let R=rows, S=distinct common stores and I=intents: O(R+I) placement/index memory and work, S store enumerations, at most I branch probes, plus deterministic output sorting. Existing Git/tracker costs are unchanged. Record reads retain1MiB maxTicket; final stdout is capped at the consumer's1MiB requirement. Existing consumer owns its90s inventory deadline; collection propagates caller cancellation. N/2N counted tests defend growth, not noisy wall-time assertions. No extra concurrency or suite/VM capacity is introduced.
- ARCH-SECURE: persisted records and paths are untrusted until existing strict decode and canonical identity association. Never construct commands from arbitrary raw bytes or assign by pid/workspace text. Non-ENOENT errors and cancellation remain uncertainty. No credentials are created or logged; fixtures cannot reach production stores.
- ARCH-ORDER: lifetime owner lock is the liveness witness, not PID scans. Snapshot-to-action race is intentionally handled by existing merge admission, not a new reservation. Partial reads retain known observations. Terminal intent removal stays unchanged; report never sweeps.
- ARCH-FUNERAL: creates no new durable runtime state; report values die with the command. Existing intent owner still removes terminal records, retains interruptions and performs the existing30-day sweep only in its normal mutation path. Private test roots are owned by t.TempDir; diagnostic captures are evidence named in the issue Log.
- ARCH-EVOLVE: unchanged preserving v1 intent format and additive inventory schema1 fields. Older Couch can ignore them; new Couch must distinguish missing capability from proven empty. Inventory is a read-only observation, not a persisted read-modify-write store. Its new known fields are strictly validated without making legacy missing fields falsely supported.

## Not covered

No automatic recovery execution, intent writer migration, terminal tombstone, new liveness monitor, store sweep, scheduler/queue resume, Pair activation or actor-readiness changes. pair#848 owns Couch's joins and operator-facing recovery report after this producer lands.

## Revisions

### 2026-10-25 — Consumer contract confirmed before code

Pair:5 confirmed the exact capability, quality and terminal-removal semantics and requested controls for legacy missing capability, null/malformed fields and mixed valid/corrupt stores. These are included above. The issue's mention of terminal state refers to the existing terminal-removal lifecycle, not a new persisted state. Approval requested from ops:0 before change-code.

### 2026-10-25 — Hub approved implementation

Authenticated receipt565af516-d8bc-4ef8-b8c8-46042ce3b16e approves the plan, including the shared reader's lost-partial-result correction, consumer wire, bounds, legacy fixtures, counted growth and private conformance plus the hub's post-land primary rebuild. Proceed to change-code; keep that reader fix explicitly tested and named in the Log.

### 2026-10-25 — Qualification details and bounded partial diagnostics

Added source/read-quality controls for absent store observations and repositories whose discovery failed before any row could carry quality. They remain explicit merge_intents diagnostics, never complete empty. Nonregular records are inspection evidence; physical identity resolves the store directory without redirecting a rejected symlink's diagnostic to its target. The shared status hint now says inspect the named source and rerun, since a failed branch probe is not authority to remove a valid intent.

Counted growth includes D existing/new diagnostics and encoded bytes B: O(R+I+D+B), plus deterministic sorting. A RED8→16 mixed-read payload grew9820→28284 bytes because every row repeated every error. Details now occur once in repository diagnostics and each affected row has a short pointer; the fixed-path8/16 encoded-size control is linear. Existing diagnostic ordering is extracted unchanged as fleet.SortDiagnostics and reused, not restated by the adapter.

Private real-executable conformance and the frozen literal v1 schema fixture passed; the fixture is explicitly not a live-fleet capture. Broad fleet/intent race qualification passed33.470s; internal wire/intent race controls passed1.231s/1.280s. The final inspection-hint control passed4.820s after its RED proof. Pair:5's authenticated consumer reply says #848 remains parked pending this producer, so new-Couch decoder qualification is after this prerequisite lands; no consumer code or live fleet was used. Recorded pre-close, close review and landing are still outstanding.

### 2026-10-25 — Consumer qualification boundary confirmed

Authenticated pair:5 receipt b0834d96-7d4f-4709-8ad1-67bb48508d4e confirms that #848 has no new consumer decoder and remains plan-only until this producer lands. The private real-executable/store fixture and frozen v1 fixture are the agreed producer proof; no additional private capture is needed before landing. The handoff will name the landed schema/head and retained fixture paths, after which pair:5 integrates main and qualifies its one-inventory seam. This resolves the conditional consumer check without claiming consumer conformance.

### 2026-10-25 — Round-one correctness controls

BR-1 reproduced an existing prunable worktree directory without .git sorting before a healthy member and hiding its shared store. Observation now uses the captured canonical Git common directory, independent of member usability, still once per store. BR-2 reproduced cancellation failing to interrupt an active production Git probe. The inventory shared-reader entry now carries context and reuses the existing bounded landing Git runner (owned process group and pipe cleanup); cancellation stops subsequent probes. No new state or observation authority is introduced. BR-3 reproduced both old JSON algebra seed failures: the oracle now distinguishes additive Inventory/TreeRow from strict policy types and still rejects duplicates. BR-4 adds the README consumer summary. New production-path regression controls failed before their fixes and pass afterward.
