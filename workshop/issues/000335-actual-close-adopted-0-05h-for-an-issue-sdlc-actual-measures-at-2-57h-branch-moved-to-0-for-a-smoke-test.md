---
id: 000335
status: codecomplete
deps: []
github_issue:
created: 2026-10-11
updated: 2026-10-14
estimate_hours: 4.174
card_mirror: 'f521a15ff9e658d5e73deee56d84a8faa7b81d4c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T01:44:37-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 7.76
---

# actual: close adopted 0.05h for an issue sdlc actual measures at 2.57h (branch moved to :0 for a smoke test)

## Problem

pair#454 closed with `Close-Actual: 0.05` (commit 3d8523e9) against an estimate of 4.69h. Right after the merge, `sdlc actual --issue 454` from the same slot (pair:1) measured 2.57h (window 2026-10-11 20:08 → HEAD; build 2.57h; one dominant attribution segment 20:08 → 22:22). The close-time measurement was wrong by 50×, and close adopts its measurement when `--actual` is omitted, so the velocity record now holds 0.05h. Context: before close the branch was moved to pair:0 (`sdlc move :0`) for the operator's smoke test and the work after that ran there; close ran in pair:1 afterwards. The slot reported the ratio as likely a measurement-window artifact.

## Spec

Find why close's measurement differs from `sdlc actual` for the same issue minutes later: the window close used (base, HEAD, which checkout), the session transcripts it attributed (pair:1's vs pair:0's), and whether `sdlc move` breaks the attribution. Fix at the root so close and `sdlc actual` agree, including across a move to :0 and back. Add a refusal or a loud warning when close's measurement is implausible against the estimate (e.g. under a tenth) instead of adopting it silently. Correct pair#454's recorded actual (2.57h, or what the fixed measurement says) through the verb that owns it.

## Done when

- The cause is reproduced in a test (a branch moved to :0 and back, then closed).
- Close and `sdlc actual` measure the same window and agree.
- An implausibly low measurement against the estimate warns or refuses at close.
- After implementation lands and the hub rebuilds the shared binary, measured correct-actual repairs pair#454, pair#514 and #335 itself. Local acceptance verifies the command and dry-runs; peer data writes are deliberately post-landing.

## Plan

Proposed durable plan: `workshop/plans/000335-actual-measurement-plan.md`.

- [x] Reproduce checkout source loss and release/resume inflation.
- [x] Derive ownership intervals/source provenance and unify consumer inputs.
- [x] Add adoption diagnostics, correction/recovery, and production regressions.
- [ ] Verify and land, then repair pair#454/#514 and #335 through the rebuilt shared binary’s measured correction.

## Log

### 2026-10-11

### 2026-10-14 — root-cause investigation
- 2026-10-14: closed — Recorded sdlc-local/suite passed on 25c8e1661a59c361139f9433a7ef3534081b746e:1336 command tests,54 packages,12 shards,346.7s. Prior four findings disposed addressed. BR-5 numeric correction preview/execution diagnostic regression RED3.453s before shared rendering; correction/printing checks41.529s and final preview/execution/no-write/refusal checks6.191s passed. no-plan-check acknowledges hub-required postlanding shared rebuild then measured corrections of #335/#454/#514; final land/repair item intentionally pending.; review verdict: SHIP

Claimed #335 and entered planning. Raw #454 session evidence corrects the
original caller assumption: close explicitly ran in pair:0 while its live agent
remained rooted in pair:1. Exact caller-cwd transcript selection excluded that
session. Retained close.log records Oct11 20:08:38 and 0.05h. A current-engine
same-interval replay selects 3.3 minutes from pair:0+brain versus 118.6 minutes
from pair:1; this proves source loss, not the original 2.57h as a corrected value.

#514's release retained started; re-claim reused that first start and selected
unrelated same-slot sessions over the released period. Independent raw-event
accounting reconstructs 23.755738h → 23.76h. Old child activity crossing the new
claim means ownership intervals need source lineage too. Suspected duplicated
histories were independent sibling agents; legitimate parallel work remains.
Exact commit/session evidence and proposed contracts are in the durable plan.

No existing actual correction verb was found. The plan proposes measured,
generation-bound correction and recovery, then data repair after landing.
No implementation or estimate; submitting direction to the hub.

## Revisions

- 2026-10-14 — Hub adds opposite inflation case pair#514. Historical evidence
  corrects #454's close cwd to pair:0. Proposed scope adds release/resume
  intervals, source lineage, pinned-input parity and repair of both records.
  Original Spec/Done when preserved; supplemental acceptance is below.

## Supplemental acceptance

- Preserve genuine early work across release/resume, excluding the released gap
  and unrelated old descendants; reproduce the #514 failure with deterministic fixtures.
- Close/actual use identical pinned inputs and report their actual right edge,
  source eligibility, and quality across moves and landing.
- Correct both pair records through the measured owning verb, or explicitly
  exclude a sample when historical evidence cannot support a trustworthy repair.

### 2026-10-14 — hub direction approved

Receipt e5efe89a-aad5-40dd-813a-d6559e2053b1 approves ownership intervals,
owner-session attribution, shared caller inputs, diagnostics, no incomplete
auto-adoption, ratio warnings and measured corrections. Plan trimmed to narrow
move-source retention and dispatch-interval lineage, each justified by the
measured #454/#514 counterexample. General task-context inference dropped.
Proceeding through change-code; estimate only after plan-quality.

### 2026-10-14 — plan-quality round 1

PQ-1–PQ-4 Important accepted: specify retained measurement inputs, correction
state/effects, authoritative calibration invalidation and named function strategies.
PQ-5/PQ-6 Minor retention/conformance notes incorporated too. Added executable
contracts to the plan; no production changes. Re-running change-code.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only. Source flagged stale; treat values as provisional.
Read v2/v2.1/v3/v3.1 method and baseline before derivation. Go JSON/time and existing
Git/tracker/receipt/ledger machinery cover the IO plumbing; no library supplies
issue ownership or lineage semantics, so no speculative library discount. Thorough
approved plan applies ×0.2 design discount and15% design buffer. Impl values are
40% of v2 ranges, familiarity1.0 for this existing Go stack.

Items in order: engagement model; source eligibility; activity clipping; durable
measurement snapshot; consumer input integration; correction transition; correction
transaction integration; calibration resolution; plausibility warning; docs;
post-land two-record data repair; one close review. Each implementation item includes
its targeted regression strategy. Design1.16h + impl2.84h; buffered total4.174h.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.10 impl=0.24
item: greenfield-go-module design=0.10 impl=0.24
item: smaller-go-module design=0.04 impl=0.16
item: greenfield-go-module design=0.10 impl=0.28
item: api-integration design=0.20 impl=0.48
item: greenfield-go-module design=0.10 impl=0.28
item: api-integration design=0.20 impl=0.48
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.02 impl=0.08
item: atlas-docs design=0.04 impl=0.08
item: cross-repo-refactor-small design=0.06 impl=0.12
item: milestone-review design=0.04 impl=0.20
design-buffer: 0.15
total: 4.174
```

### 2026-10-14 — implementation integration checkpoint

Ownership, lineage, clipping and retained-input components are committed. Shared
production measurement, close/milestone snapshot persistence, explicit brain
forwarding, ratio diagnostics and measured correction/recovery are integrated.
All calibration consumers resolve exact-row invalidation before eligibility;
ordinary and correction writers share a lock. Pure correction and engine context
checks remain separate from command IO (ARCH-PURE/ARCH-DRY/ARCH-ORDER).

Targeted root Actual/Calibration/ConvergeReport/Forecast checks passed in 55.977s.
Seven internal/help packages passed; active-time took 19.505s. Real fixtures cover
owner/caller worktrees, release/reclaim, completion/milestone replay, source append
and changed-prefix refusal, explicit brain exclusion, interrupted correction and
unchanged generation. Mutations/root production consumer conformance remain next.

Read-only historical probes exposed a further legacy scope gap: today's main made
old completed heads unscoped. Preview values are not correction evidence yet;
resolving concrete close Review-Window pins and rejecting incomplete multi-boundary
history before numeric repair. No peer record was changed. Full suite not run yet.

### 2026-10-14 — integrated verification

Expanded root Actual/Calibration/ConvergeReport/Forecast plus historical-window
checks passed in 83.463s. Restored production consumer conformance passed in
3.266s. Three independent mutations were killed by that real engine/worktree
fixture: bypassing actual's ownership resolver, dropping close's brain override,
and dropping milestone's brain forwarding. Recovery namespace checks passed
0.470s. Exact historical scope and correction-input replay regressions passed.
Exclusion diagnostics now retain counts and bounded examples instead of flooding
output with unrelated historical sessions. Full suite is the next checkpoint.

### 2026-10-14 — real repair preflight

Stopped queued make suite before it acquired a slot (exit130, no suite run). The
#514 dry-run found a missing canonical archive lookup; fixed via issue.IDDirs.
#454's explicit main-integration merge second parent proves the exact concrete
Review-Window base, allowing its multi-boundary historical scope without inference
from current main. Both real fixture regressions observed RED, then the correction
and ownership/historical/consumer suite passed27.895s. Peer records still unchanged.
Hub receipt32a83138-00fc-44cc-a4eb-7675539e024d announces #332 landed; next verification
uses sdlc local-check run --name suite on the pushed integrated commit so merge
can reuse the recorded pass.

### 2026-10-14 — main integration verification

Merged #332 and resolved Git history/window conflicts preserving pinned reads and
bounded process-group cleanup. Ownership/active-time Git reads now use the shared
bounded Git boundary too. The new context guard observed RED for the legacy window
wrapper's fresh Background context; preserved its existing context-free API through
the bounded nil fallback while explicit pinned reads retain caller cancellation.
Context/production/archive checks passed5.809s; history/window checks passed
3.706s(gitx)/1.238s(tracker); active-time passed4.419s. Archive/scope repair fixtures
had passed27.895s before integration. No peer records changed.

### 2026-10-14 — recorded suite and correction retry verification

The recorded suite on c525514a97d7e8f1c9e046f4113c47804bee49f5 failed after
348.8s: all 54 other packages passed, but cmd/sdlc reported a missing recovery
catalog contract and the existing slow-numstat cancellation fixture exceeded its
600ms wall-clock assertion under the sharded suite. Added the correction contract
with named command proofs. The cancellation fixture now sleeps30s and permits3s
for cancellation/process cleanup, preserving a large separation from natural
completion; its isolated run passed0.778s. No numstat production change.

Completed correction retries now require matching card generation, effective
calibration operation and current main mirror, returning the recorded result
without remeasurement or another row. Missing/superseded evidence refuses.
Ownership-proven replacement rows become window-trusted; legacy measurements do
not acquire trust. The legacy calibration header upgrades atomically under the
writer lock while preserving every sample byte/hash and file mode, including
header-only recovery of an already-appended operation. Header regression observed
RED before the fix. Final correction/catalog/cancellation checks passed39.436s;
recovery/issue/estimate packages passed0.181s/0.743s/0.166s.

Hub receipt804134bb-8925-425e-87f2-e409e6befc45 explicitly retains the shared-binary
landing rule. Close, PR and merge use shared sdlc; after landing and its hub rebuild,
correct-actual remeasures #335 itself as well as pair#454/#514. Read-only previews
were0.05→1.78h and23.76→0.35h; no peer data has been changed. The final Plan item
stays pending through local acceptance because it includes these authorized
post-landing actions. Close will acknowledge that ordering with no-plan-check.

## Revisions

### 2026-10-14 — acceptance ordering and correction confirmation

Per the hub's explicit direction, local acceptance precedes shared-binary rebuild
and data repair. Done when now states that ordering and includes remeasurement of
#335's own old-binary close. Complete repeats prove all authoritative effects;
legacy header migration preserves raw calibration identities (ARCH-ORDER).

### 2026-10-14 — passing recorded full suite

sdlc local-check run --name suite passed on
 dad811f5fd233dbb2db266f3929ee7fc27f78ce7: all1333 cmd/sdlc tests in12 shards
and54 other packages, wall349.2s, zero failures. Recorded sdlc-local/suite;
log SHA256 prefix bf0435c69975. Shared-binary close review is next; the pending
land/rebuild/three-record repair ordering is acknowledged with no-plan-check.

### 2026-10-14 — boundary round1 REWORK

Shared close adopted5.96h provisionally and reviewed ca9b0908..c4bdbe1b. REWORK
raised BR-1 dispatch-engagement-attribution (Claude spans merged before clipping),
BR-2 authoritative-calibration-exclusion (older rows could revive after exclusion),
and BR-3 production-adoption-warning-coverage (comparator alone did not prove
close wiring). Reproducing/fixing each class with independent regressions; the
issue remains working. The earlier suite pass remains evidence for its commit,
not these new code changes. A new recorded suite and close review are required.

### 2026-10-14 — round1 findings addressed and mutation-verified

BR-1: TaskSpan retains original Dispatched; AdmitTaskSpans binds it to its
continuous Began engagement before union. Source eligibility carries Began;
replay derives it from retained engagement slices and refuses unbound/ambiguous
windows. Continuous relocation survives; fresh reclaim does not. Claude fixtures
observed RED for four leak cases. Three internal mutations caught admission
bypass, lost dispatch timestamp and reclaim grouping loss. Existing capped normal
event-gap semantics are unchanged.

BR-2: validated exclusion removes all preceding issue samples before every
consumer's eligibility/selection; later legitimate closes/replacements survive.
Repeated identical exclusions retain their original ordering. Multiple-row pure
consumer regression observed RED, passed after fix; mutation caught. Exact-target
replacement fuzz passed45385 executions. Production convergence and forecast
fixtures now include an older eligible sample; removing the cutoff resurrected
it and failed both commands.

BR-3: production runCloseWithReview omitted-actual low/high cases assert the
warning and successful reviewed finalization; runMilestoneClose control omits the
whole-issue warning. Removing close's warning call failed both issue cases.

Integrated root correction/ownership/replay/warning/convergence/forecast checks
passed58.663s. Restored mutation checks passed2.673s; full internal active-time,
transcripts and estimate packages passed4.111s/0.068s/0.065s. Next: commit/push,
fresh recorded suite on changed code, then shared-binary round2 close.

### 2026-10-14 — recorded suite pass and boundary round2

Recorded suite on e7fe9c345a32e5d5577c622e888c7cb72e2d2b1c passed344.3s:
1334 command tests,54 packages,12 shards, zero failures (log6f5d01f45a40).
Round2 independently confirmed BR-1/BR-2/BR-3 addressed with removal/restoration
mutations. One new Critical BR-4 completion-bound-correction-lifecycle: a new
completion token retained its predecessor's correction, so resolver correctly
refused the mismatched binding. Fix the envelope transition once for all close
consumers; preserve correction on same-generation landing updates. No data repair
performed and no landing claimed.

### 2026-10-14 — BR-4 verified

SetCardCompletion now retires ActualCorrection atomically when the previous token
changes; same-token landing/repeat preserves it and rejected rewrites leave input
untouched. Internal numeric/exclusion sequence regression observed RED, then
passed; retirement-disabled and same-generation-retirement mutations both caught.
Issue package passed0.639s. Production CodecompleteCard→resolver regression
covers both dispositions plus same-generation landing; disabling retirement
reproduced the exact binding refusal in both cases (4.463s). Restored regression
passed4.345s; expanded correction/reclose/rebase checks passed56.280s. Historical
correction evidence remains in immutable card blobs. Fresh recorded suite and
shared-binary round3 review remain required.

### 2026-10-14 — round3 diagnostic gap and repair

Recorded suite e8db458421fcf144425095f1d10abfe89fe455e3 passed348.7s:
1335 command tests,54 packages,12 shards, zero failures (log9a5b41f48380).
Round3 confirmed BR-4 addressed and returned FIX-THEN-SHIP, but the ledger refused
finalization for new Important BR-5 measurement-diagnostics-preservation.
Successful correction output omitted the measurement window and quality warnings.

Command-path preview/execution regression observed RED3.453s for omitted window,
historical reconstruction basis and exclusion provenance. Extracted the existing
measured rendering into printActualMeasurement, shared by actual and numeric
corrections after complete preflight validation. No duplicate diagnostic renderer
and no unrelated close-adoption guidance in correction output (ARCH-DRY).
Correction/printing checks passed41.529s; final preview/execution, dry-run
no-effects/refusal and printing checks passed6.191s. The existing dry-run test
continues proving no tracker/main publication, ledger append or recovery ref.
Fresh recorded suite then shared round4 close are next; final SHIP still pending.
