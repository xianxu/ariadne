# Fleet resume after full shutdown — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Wake each hub's explicitly asleep workers, pending work first, with bounded concurrent resume operations and confirmed per-slot outcomes, after the shared tools have been rebuilt, then hand completion to the machine steward for the separate scheduled-task wave.

**Architecture:** Consume one typed Couch recovery report, retain the existing hub ownership and proven-conversation checks, and execute a deterministic plan through a bounded scheduler. Couch owns asleep marks, admission and durable operation receipts; tl-lead adds no persistent scheduler state. Rebuild guidance is an operator/TL procedure, not an automatic mutation of the primary checkout.

**Tech Stack:** Go, Cobra, Couch JSON CLI, stateful fake Couch, existing private real-CLI conformance harness.

## Approval surface and options

This is a restart-recovery workflow coordinating multiple processes. Per `cmd/tl-lead/help.md`, operator review is required before code; TL approval alone is insufficient. This document is the options/state/recovery side document for that review. Approval is now relayed in submitted receipts e44fc3b4-971d-4160-9bb4-269af41d0540 and 25d26f1d-9e80-483b-829f-7499e93e9f04; the Revisions below record the added third wave.

Recommended choices:

1. Default `--concurrency 3`; explicit positive N selects N simultaneous operations per invocation. N=1 retains serial execution. The limit is per hub command, not a machine-wide budget. A machine-wide scheduler would add a new ownership mechanism without a requirement; retaining serial execution leaves the measured 25-minute recovery unchanged.
2. Only explicit, valid asleep marks are eligible. Pending-work workers finish their phase before the remaining asleep workers start. This gives pending-work-first an observable meaning even when operations finish out of order. Keep dry run as the default; `--apply` still requires an explicit recovery request. Do not add a second legacy automatic-wake mode.
3. A confirmed terminal refusal/failure with stable Couch Code or Diagnostic may continue; a second matching `(Status, Code, Diagnostic)` stops new dispatch. Count across the invocation, not only consecutive results. Any uncertain outcome, unstarted command, or failure without a stable code stops dispatch immediately. Do not compare receipt IDs, addresses or prose to discover a shared cause.
4. Stop means no further launch, followed by collecting all already-started calls within their existing six-minute terminal-outcome budgets. Do not cancel a remote operation as proof it failed. Preserve successes and report the exact untouched set; any failed/uncertain outcome makes the command nonzero. Caller cancellation stops dispatch and collects in-flight outcomes; interruption or process death leaves Couch receipts as the recovery authority.
5. Rebuild ariadne:0 once, by the machine steward, before its fleet notice authorizes hub recovery. Each hub waits for that notice and wakes its workers; no hub independently rebuilds the shared checkout. No staging, live wake, or new durable lock is proposed for feature qualification.

## Contract verified with pair:6

Initial producer inspection was before pair#830 landed; pair:6 subsequently confirmed its main landing (PR 589), and the hub confirmed pair#739/#863 are on main. Qualify the landed producer before close; the old pre-landing/skipped-native observation is historical, not current qualification evidence. Its producer keeps `schema_version: 1` and adds optional RFC3339 `asleep_at` to **`rows[].agent.rows[]`**, alongside `repo_scope`, `tag`, `state`, and optional `reason`. The mark exists only while the shutdown park nonce is still in force. Normal `couch --resume ADDR` works with or without a claim and clears the asleep mark when adopted.

No capability marker exists. Missing `asleep_at` means either not asleep or an old producer: neither permits a wake. An unsupported report, invalid present timestamp, duplicate/conflicting address or ambiguous marked thread produces a named unknown/skip, never inferred eligibility. Require one unambiguous parked marked thread and the existing proven resume candidate; incomplete candidate evidence cannot authorize a resume. Accept additive unknown report fields; this reader never rewrites the report.

The old issue text described a top-level marker and no wake verb. Current peer evidence supersedes that description: hub-only `couch --wake` wakes hubs, not workers. This consumer uses ordinary `--resume` only and never executes command strings embedded in the report.

## Approved startup and three-wave boundary

A full shutdown brings up hubs only, with automation off and workers asleep. Startup is a quiet hold, not recovery authorization. No worker, scheduled task or hub automation resumes merely because a hub starts, an operator sends an unrelated message, flags clear, or a worker reports. The operator explicitly asks a hub to resume work; the machine steward first rebuilds and verifies ariadne:0 and sends the fleet notice before worker recovery.

Wave 1 is each hub's eligible explicitly asleep workers with pending work. Wave 2 is its remaining eligible explicitly asleep workers, only after all wave-1 calls finish. Each hub reports the outcomes of both waves and an explicit completion handoff to the machine steward. Failure, uncertainty or unknown required evidence produces a blocked/incomplete handoff with named untouched work and the existing inspection recovery, never a scheduler-enable claim. Absent asleep markers still authorize no worker mutation and are labelled as not asleep or an older producer, never proof of awake state.

Wave 3 is machine-wide: scheduled tasks resume only after every hub has reported both worker waves complete, or the operator explicitly says go. The one Couch owns scheduling; the machine steward (currently ops:0, already the Couch restart owner) owns this decision. No individual hub enables it alone. `tl-lead fleet resume` executes only worker waves 1 and 2 and prints their handoff; it never calls automation-on or scheduler control, creates no machine-wide barrier store, and does not reinterpret a local completion as global readiness. This reuses the existing authority rather than proving global absence by per-hub observation (ARCH-DRY, ARCH-ORDER).

Pair#874's approved, not-yet-landed contract starts every new Couch with scheduled tasks held. The main hub releases that hold with the generation-bound `couch --schedule resume`, giving the reason `all-handoffs` or `operator-go`; its status read reconciles a lost reply. Per-hub automation remains a separate veto. Each hub reports its worker waves 1 and 2 complete; the machine steward (ops:0) calls the release only after all hub handoffs or an explicit operator go. Pair#874 waits on pair#737 and owns the scheduler capability, enforcement and exact public arguments. Ariadne documents this seam and replaces its first-operator-message automation-enable guidance; it adds no scheduler mutation or global barrier. If the external capability is unavailable, scheduled tasks stay held, with pair#874 named as the next action. Pair#874 landing is not a prerequisite for #564 close. No live scheduling/recovery control is exercised for this issue.

## Core concepts

### PURE entities

| Name | Kind | Lives in | Status |
|------|------|----------|--------|
| RecoverySnapshot | PURE | `cmd/tl-lead/fleet_recovery.go` | new |
| ResumePlan / planResume | PURE | `cmd/tl-lead/fleet.go` | modified |
| ResumeBatch | PURE | `cmd/tl-lead/fleet_batch.go` | new |
| resumeReceipt / classifyResumeResult | PURE | `cmd/tl-lead/resume_result.go` | modified |

RecoverySnapshot projects one report into addressed workers, explicit asleep evidence and pending-work evidence. A row is pending when it has an active claim, a non-resting issue branch, unlanded work or dirty work; unavailable/conflicting evidence is unknown, not idle. Inactive claims alone do not imply pending work. Preserve producer quality/error fields that determine whether absence can be trusted. No separate SDLC fleet read re-derives claim truth for this operation (ARCH-DRY).

ResumePlan owns ordered rows, scope exclusions and reasons. Preserve existing `.hub` exact ownership/tie rules and `--all` visibility without increased authority; exclude all slot-zero hubs. A live worker is skipped even with a stale marker. Only eligible owned workers reach Candidates; only complete proven candidates reach Resume. Pending rows and remaining rows each use stable address order. Unknown rows are visible and unattempted.

ResumeBatch owns one finite plan and its launch/completion state. A permit spans launch through the classified terminal result, rather than just process creation (ARCH-CONSTRAINTS). The coordinator alone mutates state and writes output; workers return immutable results. Drain ready completions before admitting more work, apply stop decisions, and never launch after the coordinator has observed a stop. At most N operations may already be in flight when failure arrives. No new persistent queue or recovery ledger exists.

resumeReceipt projects Code and Diagnostic in addition to identity and status. Successful classification still requires matching ID/op/target, normal exit and succeeded status; confirmed failure requires consistent nonzero exit and terminal status. Error wrapping retains structured failure identity and original diagnostics. Actor presence never upgrades an uncertain operation to success.

### Integration points

| Name | Kind | Lives in | Status | Wraps |
|------|------|----------|--------|-------|
| CouchRecovery / execRecovery | INTEGRATION | `cmd/tl-lead/fleet_recovery.go` | new | read-only `couch --recover-plan-from-sdlc` |
| CouchResume / execResume | INTEGRATION | `cmd/tl-lead/fleet.go`, `cmd/tl-lead/resume_result.go` | modified | candidate read and exact-address resume receipt |
| fleetResume | INTEGRATION | `cmd/tl-lead/fleet.go` | modified | scope, report, actors, scheduler and command output |
| fakeCouch | INTEGRATION | `cmd/tl-lead/fleet_test.go` | modified | stateful implementation of the same Couch seams |

Inject recovery into `deps` in `cmd/tl-lead/main.go`. Bound report acquisition with an explicit timeout and output limit using the existing subprocess conventions; an unavailable report returns a repair instruction before any mutation. Resume retains its existing six-minute bound and bounded diagnostic actor probe. The fake holds asleep state, pending evidence, live actors, receipt outcomes and read counters under synchronization; channel barriers control starts/completions without timing sleeps. It clears successful asleep marks like the producer. Future producer changes widen this projection, not the scheduler.

## States, failures, restart and loss

| Internal condition | Visible state | One recovery action / loss boundary |
|------|------|------|
| Full-shutdown hubs started, operator has not requested recovery | quiet hold; workers asleep, automation/schedules off | Wait for explicit operator recovery request; no inferred permission |
| Shared tools not rebuilt or verification failed | recovery not started | Steward repairs/rebuilds ariadne:0, verifies, then issues notice; no workers launched |
| No explicit mark, foreign scope, hub or live worker | skipped, with reason | No action; request an explicit manual recovery if appropriate |
| Report/candidate evidence unavailable, malformed or ambiguous | unknown, not attempted | Inspect the named evidence/slot before a new request |
| Eligible phase queued / active call | waiting / resuming | Wait for that call's terminal receipt; no duplicate launch in this invocation |
| Matching successful terminal receipt | resumed | No action; Couch has cleared the mark |
| Confirmed isolated refusal/failure | failed, with receipt/code | Repair the named cause before explicitly retrying that slot |
| Repeated shared failure or uncertain/unclassified outcome | stopped; active results still collected, untouched listed | Inspect printed receipts and peek evidence; do not automatically resend |
| Local worker waves complete, other hub handoffs absent or scheduler capability unavailable | wave-three handoff; scheduled tasks held | Report to the machine steward; only it acts after every hub reports or operator says go; no local toggle |
| Command dies during calls | local summary incomplete | Inspect Couch operation receipts and capture a new report; durable Couch state survives, local schedule does not |

A second concurrent invocation is not coordinated by this command: Couch's per-slot admission remains authoritative. Do not claim a local semaphore fences other hubs or commands. Guidance prohibits overlapping recovery invocations for one hub; tests retain concurrent broker-refusal behavior as non-success.

## Implementation tasks and adversarial tests

One atomic issue close/review boundary; no artificial milestone tags.

### 1. Recovery projection and ordered plan

Files: create `cmd/tl-lead/fleet_recovery.go` and `_test.go`; modify `fleet.go`, `main.go`, `fleet_test.go`.

- [x] Write decoder/planner red tests for the actual nested wire; absent marker on old schema-1 reports; explicit marked no-claim worker; invalid timestamp; unknown fields; duplicate keys/addresses; wrong types; trailing JSON; unsupported schema; conflicting thread rows; incomplete candidates; foreign/tied ownership; slot-zero hubs; and stale marked live actors.
- [x] Run focused tests and retain the expected failures. Implement one typed projection and one classifier, with pending-first phases and deterministic reasons. Preserve strict types/required fields/duplicates using the repository's JSON facilities.
- [x] Count 8 and 16 workers: one recovery capture, one initial actor capture, at most one candidate read per eligible worker, and no per-completion fleet/report rescan. Assert untouched absent-marker rows never invoke a mutation.
- [x] Run the focused controls and commit implementation plus issue Log.

### 2. Bounded execution and receipt failure classes

Files: create `cmd/tl-lead/fleet_batch.go` and `_test.go`; modify `fleet.go`, `resume_result.go`, `resume_result_test.go`, `fleet_test.go`.

- [x] Write red pure transition tests for positive N, N=1 and N=3, ordered phase admission, out-of-order completion, repeated matching codes with different IDs/targets, distinct coded failures, unclassified failures, uncertainty, cancellation and exact not-attempted accounting.
- [x] Extend the stateful fake with barrier-controlled calls. Hold N calls active and prove N+1 cannot start; release one and prove exactly one starts; prove idle-phase calls wait for all pending-phase outcomes; stop on the second shared failure while retaining and joining every already-started call. Assert every address has one final outcome and all workers return before the command does.
- [x] Implement the coordinator and `--concurrency` default 3. Validate N before side effects. Preserve dry-run, --all authority, per-slot receipts, bounded actor diagnostics and nonzero partial failure. Update serial assumptions deliberately, keeping N=1 controls.
- [x] Extend receipt tests for coded failures, mismatched receipts, timeouts, malformed replies, exit/status disagreement and actor-present uncertainty. Run focused tests and race controls, then commit.

### Function-level adversarial contracts

- `decodeRecovery` (new pure decoder) must reject malformed/duplicate/type-conflicting evidence while preserving the distinction between absent marks and positive asleep proof. Recorded producer-shaped fixtures plus a malformed-input table are the independent guard.
- `planResume` must never widen hub authority or infer a mark from claims, status or neighboring rows. A truth table crosses live actors, explicit marks, pending evidence and scope conflicts; the 8/16 counted seam test rejects rescans and mutation of ineligible rows.
- `advanceResumeBatch` (new pure transition owner for ResumeBatch) handles start, terminal result and caller cancellation. Pending/remaining/stopped/complete states declare permitted launch effects; only the coordinator applies transitions. Deterministic event permutations assert concurrency, phase barriers, failure-key stopping, no dispatch after observed stop, exact untouched accounting and no scheduler effect.
- `classifyResumeResult` must keep failed, successful and uncertain receipts distinct under ID/target/op mismatches, inconsistent exits and missing stable codes. Table tests at the production receipt decoder plus the stateful executable fixture guard this authority boundary.
- `fleetResume` joins every started operation under its existing bound and reports scoped worker completion without a hidden wave-three toggle. Barrier-controlled fake calls and executable argv logs assert phase handoff, zero automation/scheduler commands, bounded in-flight calls and joined cancellation; help guards pin steward-only machine-wide wave three.

### 3. Recovery guidance and conformance

Files: modify `cmd/tl-lead/help.md`, `cmd/tl-lead/drain_test.go`, `cmd/tl-lead/main_test.go`, `README.md`, `atlas/workflow/tl-skills.md`, and `cmd/tl-lead/resume_conformance_test.go` as needed.

- [x] Replace serial/first-failure wording with the approved bound and stop rule. Put steward `make tools` in ariadne:0 and `go version -m bin/sdlc` verification before the fleet notice; build/verification failure stops the procedure. Preserve explicit operator start-recovery authorization and the hubs-only quiet hold. Replace the first-operator-message automation toggle with the machine-steward-only third-wave handoff, after every hub reports both worker waves complete (or explicit operator go).
- [x] Phrase-guard ordering and authority, pending-before-idle, absent marker no-op, per-invocation concurrency, partial outcomes, scoped completion handoff, machine-wide wave three, zero per-hub scheduler toggles, and exact recovery instructions. Add stateful/executable controls proving no automation or schedule control on success, partial failure, cancellation or dry run. Update the existing linked atlas page; no new standalone map needed.
- [x] Run `go test ./cmd/tl-lead -count=1` and focused `-race` scheduler/receipt tests. Run existing opt-in private CLI conformance with `TL_LEAD_RESUME_CONFORMANCE=1`, extending its fake executable/report seam to cover nested asleep markers and bounded calls. No production slots are resumed.
- [x] Obtain pair:6's landed #830 evidence and qualify the reader against its real report fixture. The earlier skipped-native gap predates pair#863 landing: inspect the current disposable-Couch harness and use the existing supported private qualification. Report only tests actually executed; never equate producer landing with consumer conformance. If unavailable, ask the hub to approve the explicit private conformance limit before close.
- [ ] Push, run `sdlc local-check run --pre-close`, and use normal close/PR/merge gates. After landing ask the hub to rebuild ariadne:0. No staging or live fleet recovery is part of this change without separate operator authorization.

## Growth and constraints

Planning is linear in report/thread rows plus address sorting; dispatch creates at most N active calls and one result per eligible slot. Report bytes, candidate calls and attempted addresses are counted at 8/16 scale. Default 3 comes from the issue's proposed small recovery bound and is configurable; six minutes remains Couch's existing five-minute receipt poll plus teardown allowance, not a new retry deadline. Re-measure real recovery duration only during an operator-authorized restart. Unknown outcomes spend no automatic retry budget because no retry is attempted.

## Revisions

- 2026-10-24: Initial operator-review proposal, incorporating pair:6's current nested asleep wire and explicitly distinguishing private conformance from the pending native wake qualification. No implementation authorized yet.

- 2026-10-24: Operator approval relayed by ops:0 in submitted receipt e44fc3b4-971d-4160-9bb4-269af41d0540; global third-wave ownership clarified in 25d26f1d-9e80-483b-829f-7499e93e9f04. Add hubs-only quiet startup, two per-hub worker waves and explicit completion handoff, then steward-owned machine-wide scheduled-task resumption after all hubs report or operator says go. Filed external scheduler part as pair#874; no scheduler control or machine barrier state is added to tl-lead. Add function-named adversarial strategies, explicit PURE/INTEGRATION classifications and the canonical plan filename. Source and qualification status for pair#830/#863 must be refreshed before close. Implementation remains behind change-code; this revision is authored in a private worktree while #594's existing merge owns the slot checkout.

- 2026-10-24 — qualification seam clarified by ops:0 receiptfee9e7f5-e79d-4403-8e78-d3cae12045a6: pair#870 owns lifting the native wake skip. Private executable conformance is approved while it waits; it does not block implementation. Recheck before close and run the native test if #870 has landed, otherwise state the explicit limit. No duplicate peer issue or live recovery authorized.

- 2026-10-24 — ops:0 supplied the approved pair#874 scheduler seam in submitted receipt dd147d9c-989b-4a5b-9420-412cbbdf3ce6: fresh Couch holds scheduled tasks; the main hub releases a generation-bound hold with couch --schedule resume and reason all-handoffs or operator-go, reconciling lost replies through status. Keep per-hub automation separate. Record the precise steward handoff without invoking or depending on this unlanded capability for #564 close; pair#874 waits on pair#737. #589 still lands before this issue's implementation.

- 2026-10-24 — implementation inspection: cmd/sdlc/internal/jsonstrict cannot be imported by tl-lead. Extract its unchanged implementation/tests to cmd/internal/jsonstrict, retaining a forwarding package at the old path, so both command consumers reuse one strict authority (ARCH-DRY). No key semantics or report schema changes. The initial claim-without-asleep regression is observed red; the approved recovery projection and scheduler remain the implementation scope.

- 2026-10-24 — private CLI qualification exposed SIGTERM bypassing the existing caller-cancellation coordinator. A context scoped only to fleet resume now handles SIGINT/SIGTERM and joins admitted calls; the actual executable red/green control verifies it. Updated the independent real-Couch fixture to the current isolated socket contract and a bounded owned short root. Pair#870 is still not landed, so native wake remains explicitly unqualified under the approved private-conformance limit. Package/race/catalog and final SDLC evidence follow in the issue Log.
