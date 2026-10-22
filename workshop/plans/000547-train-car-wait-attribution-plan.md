---
collaboration-mode: discussion
---

# Train car wait attribution implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for bounded delegation versus warm-context execution. Use checkbox tracking and TDD. This is a full-flow design with one close boundary, TL approved on 2026-10-22. The Revisions section records the governing policy; change-code must clear before implementation.

**Current scope:** The 2026-10-22 visibility revision at the end supersedes subtraction, idle-proof, new method and correction tasks below. Historical design is retained to explain the narrowed contract. Implement only the revised tasks.

**Goal:** Stop periodic status polls from turning a proven retained train-car wait into focused work, without attributing another conductor's suite to the car or guessing historical hours.

**Architecture:** Keep #548's immutable tested-head membership and existing solo-suite owner-slot selection intact. Record a car's own blocked landing wait at the existing OS-held ticket boundary, join it to positively observed train/conductor membership, and pin the validated interval in measurement evidence. Feed source-bound idle intervals into the existing subtraction engine. A conductor's queue, baseline and split are phases within that waiter lifetime, rather than independent guesses about the car's time.

**Tech stack:** Go, existing landingline flock/clock/codec and 30-day audit retention, tracker ownership/transcript provenance, gatestate measurement snapshots, native active-time engine and portable real-Git/stateful fixtures.

## Evidence and scope

- #548 at eb0e3d8a8396adb9cb7af9dba44987daef66ed3d admits suite attempts only after immutable SHA membership proves the measured issue; neutral heads inherit first-parent membership and exact train compositions confer only their own cars. Preserve that order and its unknown/foreign distinction. Do not globally relax conductor slot ownership or admit main-baseline attempts as an issue's own queue.
- Current attachOwnedQueueEvidence selects only the issue owner's suite queue interval. MainBaseline is deliberately skipped; a later split prefix can omit the waiting car entirely. Reading a train suite SHA cannot establish the car's complete blocked wait.
- pair#778 retained events prove ticket262/issue778/PR536 in train261. Enqueue 2026-10-21T20:37:15.152800-07:00 to conductor refusal 23:43:56.723901-07:00 is an envelope of 186m41.571101s, not a proven waiter lifetime. The old event schema omits waiter slot, process incarnation, head and actual Wait return; ticket/readiness/outcome files are gone.
- Conductor pair:2 attempts remain: 7b86027ca92a83fe2d21d23d8c90fc4c (combined failure), fe30c82b2d661f4d810f192bcfe27b2d (main baseline), b70b199a850719de2da728cf829fd0d6 (prefix). They establish phases but cannot supply missing car endpoints. Preserve the original 4.41h and 1.11h measurements; do not subtract the entire 213.9-minute dominant segment or hand-type replacement hours.
- The pair#778 close ledger retains tracker/code pins and transcript source prefix identity/digest. Its 79 MiB transcript will be processed by bounded extraction tooling, never pasted into the agent context. Record either an evidence-backed correction preview or the precise historical limitation; correction publication needs separate TL/operator disposition.

## Alternatives and decisions for approval

1. **Recommended: producer-owned waiter lifetime plus train membership.** The existing held landing ticket is the single resource proving that invocation remains blocked. Three bounded event kinds (start, train binding, end) retain the identity that existing cleanup removes. No per-poll history or conductor-phase journal is needed. ARCH-ORDER/DRY.
2. **Reject: treat the conductor's suite as the car's own suite.** This breaks #548 and the owner-slot guard, misses main-baseline time and split prefixes that no longer contain the car, and can subtract another issue's work.
3. **Reject: infer the wait from prose, polling frequency, enqueue/refusal timestamps or train completion.** Those are observations, not exact return/liveness boundaries. They cannot justify a trusted historical number.

No new CLI command, machine setting, landing priority or automatic actual correction. The visible change is a named train-car wait exclusion or an explicit incomplete-evidence refusal. Existing saved measurement generations replay unchanged.

## Core concepts (PURE)

| Name | Location | Status |
|---|---|---|
| LandingWaitIdentity, start/binding/end reduction and validation | cmd/sdlc/internal/landingline/wait_evidence.go | new |
| SelectedCarWait and source/engagement clipping | cmd/sdlc/actualtrainwait.go | new |
| MeasurementCarWait plus train-aware method identity | cmd/sdlc/internal/gatestate/measurement.go and measurement_unknown.go | modified |
| source-bound idle union/subtraction | cmd/sdlc/internal/activetime/queue.go | reused; modify only if a typed common idle input is needed |

One waiter lifetime belongs to one repository, selected issue, root/slot, full head, ticket sequence and OS process incarnation. It may encounter more than one train after release/reformation; bindings are ordered non-overlapping episodes, not membership inferred from a later prefix. A train binding includes the conductor ticket/process identity and the independently verified car head/PR. The selected car is never its own conductor.

The reducer consumes supplied facts only. Missing starts/ends, conflicting duplicate identities, head/slot/root mismatch, torn records, reversed clocks or incomplete train bindings are unknown, not zero-length waits. Exact duplicate immutable events are idempotent; contradictory duplicates refuse. A closed unboarded solo waiter confers no new car exclusion.

## Integration points

| Seam | Location | Status | Wraps |
|---|---|---|---|
| ticket-held wait evidence writes | cmd/sdlc/landingqueue.go; internal/landingline/line.go and outcome.go | modified | existing Handle lifetime, injected clock and append-only audit |
| train binding at verified boarding/outcome observation | cmd/sdlc/landingtrain.go; internal/landingline/outcome.go | modified | live ticket/readiness, verified member head and conductor identity |
| bounded wait-evidence reader | cmd/sdlc/actualtrainwait_io.go | new | one read of existing landing events log, exact raw record digests |
| measurement attachment/replay | cmd/sdlc/actualinputs.go and actualmeasurement.go | modified | tracker windows, source slots, #548 head membership and saved snapshots |
| stateful fixtures | internal/landingline/*_test.go; actualtrainwait_test.go; actualinputs_test.go | modified/new | portable queue files, deterministic clock, real Git heads and transcript timelines |

No new Git/GitHub service abstraction. Reuse #548's reader and immutable commit-membership proof for the pinned car head, once per distinct head. Reuse existing landingline and tracker fixtures. Read-only historical evidence extraction uses the existing Pair Git objects/local records, never edits Pair's actual or calibration row.

## Producer contract

Use the existing landing events file with additive versioned wait payloads; keep public landing/issue commands and ticket admission policy unchanged. Retire no landing locks or cleanup obligations.

1. After allocating a ticket and before entering the blocking Wait scope, write a start containing canonical repository/root/slot, issue, branch/full head, ticket, queue operation identity and PID incarnation plus exact start time. Obtain immutable launch identity from the existing queue runtime/selected target, not a branch-name guess. An initial write failure refuses before waiting and releases the owned ticket normally.
2. At positively verified boarding, append a binding between this live waiter and the conductor generation, carrying the verified car head/PR and both identities. Do not trust current train-control contents after cleanup or merely a transcript assertion. Preserve release/reform/drop transitions through the existing reducer and actual lock authority.
3. The car owner records end at the exit of its blocking wait scope, before its ticket is released, for every return path: admission, final car outcome, cancellation, known refusal or external landing detection. Stop/join readiness polling before reporting the end. A crash leaves an incomplete start; never invent an endpoint from later discovery that the process died.
4. Later evidence-write failures preserve the partial start/binding and name that calibration is unavailable. They do not undo a possible external landing or erase cleanup obligations. The exact failure disposition must remain visible to measurement readers; do not silently fall back to a trusted ordinary actual for a relevant incomplete episode.

The records describe an existing resource lifetime, not a new controller for landing effects. One current invocation owns each start/end; the existing conductor supplies the verified train relationship. Inject failures and races at writes and Wait returns to establish those guarantees.

## Selection and measurement contract

- Select canonical repository/issue and immutable car-head membership before using wait time. Require exactly one pinned owner engagement covering the selected portion and canonical root/slot identity; bind only admitted sources in that car slot. Keep conductor sources and other slots untouched.
- A complete proven blocked waiter suppresses polling engagement within its selected interval, including while another slot is queued, running a combined suite, checking main baseline, or splitting/testing a prefix without that car. Neither code inclusion in the later prefix nor a suite's duration creates the car wait.
- Preserve work before/after the interval and concurrent unrelated admitted sources. Reuse the existing queue policy for a proven blocking wait; do not introduce text heuristics for 'looks like polling' or treat the full enclosing activity segment as idle. If substantive same-source activity makes the wait-only claim ambiguous, surface incomplete calibration coverage rather than guessing replacement credit. The exact testable treatment of this ambiguity is part of TL approval.
- Merge overlapping own-suite and car-wait cuts before subtraction so minutes are not excluded twice. Diagnostic raw car-wait totals remain distinct from raw solo-suite queue totals; do not put train duration into MeasurementQueueWait's queue_seconds fields.
- Add a separate typed car-wait snapshot collection, exact evidence digests and explicit train-aware measurement method names (including hybrid legacy-prefix form). Older readers must refuse the new methods rather than silently ignore exclusions and produce a trusted larger value. Existing method meanings and stored snapshots remain unchanged. Preserve unknown top-level and nested fields; keep required fields/types/duplicates strict.
- Historical replay reads pinned snapshot facts without live log dependence. Fresh close/re-close/correction candidates may attach new producer facts. Corrections do not mutate the original adopted measurement or choose an ambiguous calibration row without TL/operator authority.

## User-visible states and recovery

Actions: measure/preview, restore evidence and retry, cancel a waiter, use the existing explicit calibration-exclusion path after steward review.

| Internal evidence | Visible outcome | Recovery action | Possible loss |
|---|---|---|---|
| complete bound start/binding/end | measured actual names proven car idle and preserves surrounding work | measure/preview | none beyond existing queue policy |
| positively foreign/unboarded complete episode | not this issue's car wait; ordinary rules unchanged | measure/preview | none |
| relevant start/binding without a trustworthy end, conflicting identity or torn bytes | actual not calibration evidence, names ticket/train/reason | restore evidence and retry | exact old endpoint may be unrecoverable |
| known old train-car episode lacking required provenance | historical limitation, no replacement hours | steward-reviewed existing exclusion path | historical sample may remain excluded |
| cancellation with owner-recorded end | ended bounded wait, no post-return exclusion | measure/preview | no inferred work after end |
| crashed/interrupted owner without completed receipt | unknown end, never extend to now or next poll | restore evidence and retry | unavailable historical time cannot be recreated |
| saved old close measurement | exact existing generation replay | measure/preview | original known defect remains historical until explicit correction |

## Complexity, bounds and lifecycle

The hot path does not write on a status poll. Each ticket writes one start, one end and one bounded binding per actual train episode; event count grows with tickets/reformations, not poll count or elapsed waiting. Keep each versioned record within the existing maxTicket bound and validate at the writer. Reuse the existing 30-day landing audit retention/raw-byte-preserving sweep; completed measurement snapshots pin the facts needed after audit expiry. Incomplete records stay diagnostically incomplete until their evidence ages out; never synthesize a completion.

Read the audit once per measurement with explicit total-entry/byte bounds and cancellation, using existing evidence-reader bound conventions. Index by ticket/train/owner and read each distinct immutable head once through #548's bounded batched reader. Expected local work is O(E + T + B) plus sorting/merging selected intervals; Git graph work retains #548 limits. Count actual file reads, reduced events and head requests at N=8/16 tickets/episodes. Extend poll duration N/2N in controlled timelines and assert focused work stays the same, not merely a similar runtime.

## Risky-function test strategies

| Function/seam | Independent regression strategy |
|---|---|
| wait-event reducer | adversarial event tables: duplicate/reordered start/end/bind, train self-membership, foreign repo/head/root, partial/torn/future/invalid timestamps; output invariant excludes only proven intervals |
| ticket-scope writer | real portable flock fixture with injected clock and readiness barrier; cancel/drop/admit/return races, start/end write failures and abrupt owner loss; assert no open-ended completion fabricated |
| selectCarWaits | actual pinned tracker engagements plus #548 head proof; one issue versus another conductor slot, split prefix omitting car, source ambiguity, crossing ownership windows; keep solo suite controls |
| measurement encode/validate/replay | unknown-field roundtrips including nested facts, strict known fields, new-method older-reader refusal, unchanged old snapshot replay |
| active-time consumer | controlled Claude/Codex timelines with substantive work before/after wait, periodic polls during queue/run/baseline/split, concurrent other-source work; compare N/2N and count exclusions only once |
| actual/correction integration | real Git/portable tracker + captured train261 metadata; prove complete modern evidence correction preview or a named historical limitation, without editing adopted hours |

## Implementation tasks (one close boundary)

### Task 1 — prove the failure and producer lifetime

- [ ] After #548 lands, begin from main containing it (no stacked branch); retain the read-first comparison and integrate the landed dependency before implementation through the appropriate SDLC gate.
- [ ] Add controlled transcript/producer fixtures and observe RED: pre/post issue work plus car waiter on another conductor, polls during combined queue/run/baseline/prefix, and N/2N wait lengths.
- [ ] Add pure identity/interval reducer and portable start/binding/end writer tests, including interrupted/missing boundaries and write errors.
- [ ] Implement the minimum producer facts around the existing held ticket and verified train relationship; no per-poll writes or conductor-phase journal. Run targeted landingline/train/queue regressions and checkpoint.

### Task 2 — bind fresh measurement and preserve replay

- [ ] Add failing ownership/head/source/clip tests and snapshot compatibility tests, preserving #548's foreign/unknown controls and existing solo-suite behavior.
- [ ] Add typed car-wait snapshots and explicit new method identities; attach validated facts through the existing measurement pipeline and reuse source-bound interval subtraction.
- [ ] Verify no double exclusion, concurrent unrelated activity, same-source ambiguity, ended waiters and old incomplete evidence. Add counted 8/16 cases and Claude/Codex transcript controls; include Qoder controls when compatible captured fixtures are available.
- [ ] Run targeted actual, close/re-close, correction and saved-replay tests; checkpoint the coherent slice and update atlas/README/help for actual provenance/refusal behavior.

### Task 3 — real evidence and acceptance

- [ ] Replay train261's bounded pinned metadata/source extraction through the production selection path. Record an exact correction preview only if all required historical identities/endpoints can be proved; otherwise record the limitation against pair#778 and propose the existing exclusion path to TL. Never type replacement hours or publish a correction without steward disposition.
- [ ] Run focused packages and relevant race checks, push the final code, then record the full suite through sdlc local-check run --name suite.
- [ ] Close through the binary-owned fresh-context review with completed evidence. No extra ad-hoc boundary review. Landing and any steward-requested live correction follow the accepted close; they are subsequent SDLC receipts, not claims of pre-close implementation completion.

## Approval requested

Approve the bounded producer-owned wait record, conservative historical limitation, explicit measurement-method boundary and preservation of #548's slot/head authority. Resolve the substantive same-source-activity ambiguity at the policy boundary before implementation; the recommended approach is conservative unverified coverage rather than heuristically classifying commands. No estimate is derived until the approved plan clears change-code plan-quality.

## Revisions

### 2026-10-22T05:18-07:00 — approval and counted unverified coverage

TL approval receipt 898cfb60-b77a-42a4-81f4-11b13628ed6d accepts the proposed bounded ticket-owned facts, existing retention, typed source-bound evidence and compatibility boundary. The same-source policy is now explicit: substantive work during a retained wait stays counted and is labelled **unverified coverage**. Only proven idle waiting is excluded. Do not use command-text heuristics, guessed credit, per-poll history or automatic historical actual rewrites. This supersedes the earlier pending-policy wording; ambiguous substantive activity does not itself become a calibration refusal. Malformed or contradictory provenance remains a separate evidence error.

Add a same-source substantive-work control proving the activity is counted and visibly labelled, alongside the poll-only N/2N controls. A held ticket proves the blocked command's lifetime; it does not by itself prove the entire transcript source was idle. Before implementation, identify the structural evidence that permits an idle cut; unsupported or ambiguous source activity stays counted. If this needs a mechanism beyond the approved producer facts, take that question to TL before code.

#548 landed at 80602233df36a670023b60af8dabb0b6f3c2c071 (tracker plus Git ancestry verified). Integrate landed main before change-code; retain the initial read-first digest as historical design evidence. No estimate or implementation has begun. The next workflow action is the full-flow change-code gate once the structural idle-proof detail is concrete.

### 2026-10-22T05:23-07:00 — visibility instead of subtraction (governing scope)

TL receipt bd4da1f3-415c-467a-a065-e0d099edd3d2 approves a narrower delivery: record bounded held-ticket start/bind/end facts as already designed, and report overlapping retained train waits beside measured hours as **unverified coverage: N min in a retained train wait**. All those minutes stay counted by existing attribution. No source-idle inference, interval subtraction, correction preview, new measurement method, runtime binding prerequisite or cross-repo implementation belongs in #547. File the structural runtime-job/invocation binding separately at low priority. This revision supersedes the original goal's exclusion claim, selection/subtraction policy, typed exclusion method and original Task 1–3 checklist wherever inconsistent.

**Revised goal:** Expose trustworthy retained-train-wait coverage without claiming that the waiting source was idle or changing measured hours.

**Revised pure entities and integration:** Keep LandingWaitIdentity and bounded start/bind/end reduction in internal/landingline/wait_evidence.go. SelectedCarWait in actualtrainwait.go selects exact complete intervals using issue/repo/head/owner engagement authority and unions overlapping intervals per source. It returns descriptive coverage, never QueueIntervals. Add an optional typed MeasurementCarWait collection to the existing measurement snapshot codec/unknown-field handling, with evidence digests and clipped source bindings. Existing method identities and all actual/calibration arithmetic stay unchanged: these facts are diagnostics only, safe for old readers to ignore. No change to internal/activetime/queue.go or activity event classification. Fresh attachment uses actualtrainwait_io.go's one bounded audit scan plus #548's immutable-head reader; replay reads pinned snapshot facts. Source provenance identifies whose overlap is reported, not exclusive idle.

**Visible states:** A complete selected start/binding/end reports the union of covered retained train-wait minutes. Partial/malformed relevant evidence reports 'unverified coverage: retained train wait evidence incomplete' with ticket/reason and no invented duration; ordinary measured hours still print. Positively foreign/unboarded complete episodes add no coverage. Saved snapshots without this optional field replay exactly as before. Crash/retention loss may permanently prevent an exact descriptive duration; restoring evidence and remeasuring is the only recovery. No new calibration refusal or automatic historical rewrite. Diagnostics explicitly say this coverage remains counted and does not prove idle.

**Producer details retained:** The existing locked ticket owns start/end, with process incarnation, queue operation, immutable head/root/slot/repo/issue. Binding comes from verified car boarding and conductor generation under existing lock authority; releases/reforms can give multiple bounded bindings. Start is written before Wait, end after readiness polling joins and before Release, on all normal/cancel/refusal/settled returns. Initial write failure refuses before entering Wait and cleans up ticket; later write failure warns while preserving normal external landing and cleanup outcomes. Missing ends are incomplete, never inferred. Existing 30-day raw-preserving audit retention remains. Start/end fact growth is O(tickets); binding growth O(actual train episodes), never polls.

**Revised Done when and checkable tasks (one close boundary):**

- [ ] RED/GREEN producer tests: complete start/bind/end under real portable ticket locks and injected clock; cancellation, train reformation/release, final outcome versus readiness race, abrupt loss and each write failure. Reducer rejects contradictory identity/timestamps and handles exact duplicate facts idempotently. Integrate into landingqueue.go/landingtrain.go and internal/landingline at existing held-resource boundaries.
- [ ] RED/GREEN selection tests: own issue on another conductor, baseline and split-prefix phases including omitted car; exact issue-head/repo/root/slot/engagement/source binding; foreign and unboarded controls; incomplete/torn/missing/conflicting records produce named diagnostic without guessing duration. Read audit once, retain exact digests, union overlaps, and reuse bounded #548 head proofs rather than walking Git per event.
- [ ] RED/GREEN snapshot and actual tests: optional typed coverage fields preserve unknown members, save exact clipped bindings and replay after audit expiry; old snapshots and method identities unchanged. Print the requested coverage phrase beside hours, distinguish incomplete evidence, explicitly keep coverage counted. Actual, close/re-close and calibration consumers must produce identical hours/status with and without descriptive coverage.
- [ ] Controlled Claude/Codex timelines contain pre/post work, retained car wait, status polls and substantive same-source work. Assert all existing measured hours remain unchanged, and coverage N/2N doubles exactly as facts extend. No promise that focused work remains constant under longer polling: that old Done-when is deferred to the separate runtime-binding issue. Other-source work and existing solo queues remain unchanged. Use Qoder fixtures when supported/available.
- [ ] Count reader scans/event reductions/head requests at 8/16 tickets; extend poll count without increasing producer facts. Explicit bounded total log bytes/entries, per-record bytes, cancellation and one-pass indexes retain the original complexity defense.
- [ ] Record pair#778's historical limitation (existing facts cannot provide full new identity/endpoints); do not change adopted hours. File the low-priority Pair follow-up for structural runtime-job/invocation binding. Update atlas/README/help to describe coverage, retained bounds and inability to infer idle.
- [ ] Run focused checks/race tests, checkpoint and push final code; recorded suite through sdlc local-check run --name suite; close with the binary-owned fresh-context review, then prompt landing on SHIP. No extra boundary review.

No estimate until this revised plan clears change-code's plan-quality gate. The dependency is already integrated from landed main.

### 2026-10-22 — executable visibility semantics (PQ-1 through PQ-3)

**PQ-1 / interval meaning:** Coverage describes the **whole retained waiter invocation that demonstrably served as another conductor's car**, not time aboard a specific train. ReduceWaitEvidence groups by immutable ticket/operation identity: no start -> incomplete; one start -> open; zero or more distinct verified bindings -> open; one owner end -> closed. A closed invocation with at least one valid binding yields [start,end); a closed invocation with none yields no train coverage. Exact duplicate facts are idempotent; conflicting starts/ends/identities or bindings outside endpoints make the invocation incomplete. A binding must match the waiter's full head and carry the verified car PR and a distinct conductor ticket/process identity. Bindings are a membership set, not episodes needing separate closure. Release/reformation leaves the same invocation open; the whole interval, including the initially unbound queue and released gaps, remains 'retained train wait' coverage after a valid binding exists. Cancellation/terminal outcomes do not supply an endpoint; only the owner's actual Wait-scope return (after polling join, before releasing remaining held resource) does. This removes the earlier promise of bounded binding episodes: no extra release fact or phase inference is needed. Reported coverage explicitly is not 'time aboard' or 'proven idle'.

**PQ-2 / diagnostic budgets:** ReadWaitEvidence uses a 32 MiB total audit budget, 100,000 nonblank records and 1 MiB per record (existing landingline maxTicket ceiling). These conservative ceilings bound a local 30-day diagnostic scan; tune only from measured fleet evidence, never by unbounded allocation. Exceeding a ceiling or encountering unreadable/torn relevant records returns a bounded incomplete-coverage warning, no numeric result and no accounting error. Cancellation follows the existing command context. Reduce at most the admitted records in one pass, index once by ticket, and batch distinct heads under #548's unchanged reader limits.

AttachCarWaitCoverage admits at most 16 selected waits and 8 KiB of optional serialized coverage facts, comfortably within the existing 64 KiB gatestate.SnapshotLimitBytes. It builds a copy of the otherwise valid accounting snapshot and validates the whole candidate. If count/bytes or final snapshot limit fails, keep the original snapshot exactly and emit a bounded runtime incomplete-coverage warning; do not truncate into a misleading numeric total or discard existing accounting/unknown fields. A compact persisted incomplete note may be retained only if the candidate still validates; if even it cannot fit, replay retains original accounting without a coverage claim (document this limitation). Fresh optional attachment never changes measurement validity/hours/status; malformed known fields on an externally edited saved snapshot still follow the existing strict decoder contract. The 16-wait cap is a snapshot budget, not a per-ticket fact cap; reduction/indexing remains separately bounded by the audit budget.

**PQ-3 / governing risky-function strategies:** This table replaces the superseded test strategies/case inventories above. Concrete test tables belong in code; the revised checklist identifies delivery slices, not a required prose inventory.

| Function | Adversarial input class and mechanical guard |
|---|---|
| ReduceWaitEvidence | arbitrary ordered/duplicated/partial identity fact streams; pure state reduction yields only complete bound invocation intervals, otherwise named incomplete coverage |
| Line.BeginWait / Line.EndWait | held-resource lifecycle and IO failures under portable locks/injected clock; no start after release, no invented crash end, exact single owner endpoint |
| Line.BindWait | concurrent verified boarding versus ticket removal/reformation; existing line mutex binds exact live identities, never a later process reusing a PID |
| ReadWaitEvidence | arbitrary/truncated/oversized audit bytes and cancellation; bounded single scan plus strict relevant-record decoder, no optional IO failure invalidates accounting |
| selectCarWaits | adversarial repo/head/owner/source/measurement windows; #548 immutable membership then exact pinned authority, union descriptive overlap without arithmetic inputs |
| AttachCarWaitCoverage | capacity boundaries and pre-existing unknown/large accounting snapshots; candidate-copy validation or preserve original plus bounded warning |
| MeasurementCarWait.Validate and codec methods | malformed known fields and unknown nested members; strict known types, preserving roundtrip, no new method meaning |
| carWaitCoverageWarnings / computeOwnedActualFor | supported agent timelines and saved replay with optional diagnostics; unchanged hours/status/calibration and explicit counted-unverified label, union coverage N/2N |

The initial audit and snapshot budgets are deliberately conservative diagnostic ceilings; incomplete coverage is useful and cannot justify changing an actual. Pair follow-up filed as **pair#800**, unclaimed, low severity.
