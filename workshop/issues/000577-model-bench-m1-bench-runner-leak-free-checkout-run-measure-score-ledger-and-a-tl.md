---
id: 000577
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours: 6.51
collaboration-mode: collaboration
card_mirror: '2d9e79546c407ab395bb3c62cfac9c233ce6823c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T12:35:23-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# model-bench M1: bench runner (leak-free checkout, run, measure, score, ledger) and a tl-lead bench mode

## Problem

## Spec

**model-bench M1: the bench runner (observer tool) and a bench mode for tl-lead**, per the operator-approved design
in ops `teams/devinfra/projects/model-bench/project.md` (10-23). Three roles on a separate computer: couch-bench (the
observer, mostly this tool), ops-bench:0 (the TL hub, one fixed model, normal tl-lead rules plus a bench mode), pair:0
(the worker, the model under test). M1 runs **one autonomous corpus issue end to end with one worker model**.

The runner (a CLI in ariadne, run by the observer):
- **Prepare** a leak-free task checkout: the issue's parent commit, later history, remote and the archived issue
  removed; the issue as originally filed (Spec and Done-when only).
- **Start and watch** a run: hand the issue to the bench TL hub, poll to a terminal outcome ("would land": full suite
  and close review pass, nothing merges; or stopped), with a wall-clock budget.
- **Measure:** wall time, operator-wait time (notification to reply), worker and TL tokens (from the agents'
  transcripts), review rounds and findings, gate refusals, and the TL's own decisions (re-plans, rulings, stops).
- **Score:** apply the landing's own added tests (hidden acceptance) to the result.
- **Ledger:** append one row per run to a single-writer ledger in ops (keyed by repo, task type, worker model, TL
  model), plus the raw run record.

The bench mode for tl-lead: the TL never codes, approves or re-plans by its normal rules, logs each decision in a
form the runner collects, and never lands.

Out of M1: the second computer's setup, more than one worker model, the report.

## Done when

- One autonomous corpus issue (e.g. parley.nvim#403) runs end to end on this Mac in disposable bench slots with
  real agents (one fixed TL model and one worker model), producing a real ledger row with every field above and
  a hidden-acceptance score before close. Synthetic end-to-end rows are automated test evidence only.
- The prepared checkout provably contains no trace of the landed solution (test: later commits, the archived task
  artifacts, source Git object access and original/source remote are absent). The sole remote is a new private
  bare origin per run containing only sanitized replay objects; normal tracker CAS/gates remain real.

## Plan

- [ ] Deliver the approved project M1 through the [durable implementation plan](../plans/000577-model-bench-runner-plan.md): sanitized preparation/private tracker, observer runtime and TL mode, measurement/leak diagnostics, hidden acceptance and ledger, disposable conformance, full suite and close.


## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Calibration is reported stale by `sdlc estimate-source`; treat this as provisional focused ship-time, excluding suite-queue and operator-provisioning idle time.

Thorough approved plan: design discount ×0.2 and +15% design buffer. Existing Git/tracker/Couch/usage/managed-test owners replace novel platform work (library-availability reduction applied to the reused portions). Implementation values below already use v3.1's 40% scale; familiarity 1.5 reflects the new but bounded multi-role observer orchestration. No new external library or service is assumed.

| Primitive rows, in block order | Scope |
|---|---|
| issue-spec | project interpretation, approvals and executable plan |
| typed-data-prototype | typed manifest/public projection/phase/result contracts |
| greenfield-go-module ×2 | pure event/admission core; sanitized preparation and portable store |
| smaller-go-module | bench-local dependency construction and guarded receipts |
| api-integration | Couch identity/receipt/watch/cancellation adapter and stateful fake |
| greenfield-go-module ×2 | attributed measurements/leak detector; acceptance scoring/ledger |
| skill-or-dispatcher | TL bench mode, decisions and finish proposals |
| real-api-discovery | native Couch bridge and actual agent/runtime conformance |
| smaller-go-module | end-to-end fixture and real-run acceptance integration |
| atlas-docs | help, README, atlas and recovery contracts |
| milestone-review | one whole-issue close review, not an extra milestone boundary |

```estimate
model: estimate-logic-v3.1
familiarity: 1.5
item: issue-spec design=0.15 impl=0.08
item: typed-data-prototype design=0.10 impl=0.20
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: smaller-go-module design=0.05 impl=0.20
item: api-integration design=0.20 impl=0.60
item: greenfield-go-module design=0.15 impl=0.32
item: greenfield-go-module design=0.15 impl=0.32
item: skill-or-dispatcher design=0.10 impl=0.20
item: real-api-discovery design=0.00 impl=0.24
item: smaller-go-module design=0.05 impl=0.20
item: atlas-docs design=0.03 impl=0.08
item: milestone-review design=0.00 impl=0.20
design-buffer: 0.15
total: 6.51
```

Design 1.38h ×1.15 + implementation 3.28h ×1.5 = 6.507h, rounded 6.51h. Main uncertainties are true-runtime provisioning/conformance and historic task/tool compatibility. A larger required mechanism goes back to the hub rather than being hidden in a new multiplier.

## Log

### 2026-10-23

After landing #572 (PR #412), read the ops model-bench project and local charter before claiming #577. `sdlc claim` and `sdlc start-plan` created this dedicated branch from current main. No implementation has started. Using brainstorming/writing-plans; bounded read-only explorations verified existing Git isolation, runtime and measurement seams. The existing review clone is unsuitable because it shares source objects and restores current main; prepare a new repository from pinned parent-tree bytes instead (ARCH-SECURE).

Hub approved the isolation boundary (receipt `31b1fff2-8aec-4ae6-8764-a97676415460`): M1 uses trusted agents and prepared-repository isolation, with post-run transcript/tool detection and a visible possibly-leaked ledger result; OS read/network isolation is deferred until before the bench machine also hosts development checkouts. Hub also approved a private sanitized bare origin (receipt `9606d6e4-388f-48b4-a1a0-3f8be1af8271`), preserving real claim/start-plan/change-code/close CAS and only adding a bench-local check status sink. Default production behavior stays unchanged (ARCH-DRY, ARCH-PURPOSE).

Draft durable plan is written, awaiting hub review before code. It names the state/effect owner, failure/recovery table, bounded lifecycle, fake/native conformance, risky-function strategies and isolation limitations. Pair's existing conformance bridge can qualify native Couch service/lifetime behavior but does not launch the TL/worker replay; end-to-end automated rows are synthetic fixture evidence, while a real paid replay requires a separately provisioned disposable bench setup. No production ledger, source repo, account setting or live slot was changed. Next: send the committed plan to the hub, resolve review questions, then run full-flow change-code only after approval.

Plan approved by ops:0 with a real-run acceptance condition (authenticated receipt `32918d0f-ab73-4a25-b4e7-2bc8a6633d47`). Recorded that condition in Done-when and appended a concrete disposable-slot provisioning checklist: sanitized worker/private origin, fresh operating hub, supervised Couch discovery/launch, exact hub claim, real agent profiles/models, candidate tools/transcript identities, acceptance evidence and retirement. Hub provisions together when code is ready. Next: full-flow change-code plan-quality gate; no implementation yet.

Plan-quality round 1 refused on PQ-1 (external-dependency-boundary-completeness): run/republish resolve a GitHub-only repository identity and discover a PR before the proposed local sink. Independently verified the cited production code. Refined the same approved adapter with one named dependency constructor shared by both commands: validated private publication identity, no GH client/PR discovery in bench mode, ordinary production unchanged. Added actual built-CLI run/republish tests with an all-call rejecting GitHub sentinel and unchanged same-head proof guards. Next: rerun plan quality; no implementation has started.

Plan-quality round 2 CLEAN: PQ-1 disposed, no remaining blockers. The gate now requires the estimate. Derived the itemized v3.1 estimate only after plan approval, using the named source and baseline; 6.51h with stale-calibration caveat. Next: record the card estimate and clear estimate quality/change-code; no code before admission.

Full-flow change-code admitted at `5a75ce5dc` after plan-quality CLEAN (PQ-1 disposed) and estimate-quality INFO. Advisory: the mandatory true replay may exceed the provisional live allowance; count its agent execution in actual ship-time and retain that uncertainty, rather than excluding it as operator wait. No gate bypass.

Implementation foundation delivered with test-first evidence: pure Task/PublicTask strict codecs and projection, private Run/Step event/effect owner and would-land admission, and diagnostic leak scanning/coverage. Bounded agents edited only their assigned pure modules; no integration or process state was shared. Initial missing-owner reds were observed; behavioral controls also caught invalid UTF8 repair, repeated cleanup authorization and an unsafe private-fetch exemption before fixes. Root independently ran the whole bench package under race: PASS (1.322s). Leak scanner fuzz independently passed 128,152 executions; its cross-harness normalized samples are explicitly synthetic, not claimed as live captures.

Added local-only Git tree observation at the existing snapshot owner: `gitx.ReadLocalTree` shares the bounded batch/parser helper with ordinary snapshots, but disables lazy fetch, replacement refs and transports and refuses inherited object-store redirection. Its new tests first failed with missing owner, then passed (1.641s); snapshot/selection controls PASS (7.117s). The promisor sentinel proves a missing source blob is not repaired by a fetch, and replacement/export-filter controls preserve the pinned parent's actual bytes. Tracked golint PASS (1.625s), staged whitespace check PASS. This is the approved sanitized materialization boundary, not a new clone mechanism.

Important integration rule: close writes evidence/mirror commits after reviewing a code head. Pin the product candidate to that real reviewed code head and verify current metadata ancestry using existing publication proof; do not forge same-head facts by treating a metadata commit as the reviewed code commit. Next: portable run store and sanitized prepare integration/CLI; then bench-local check constructor/receipts, runtime/TL adapter, real measurement/scoring/ledger, documented recovery, full verification and hub-provisioned true replay before close. No suite, stage, live bench actors or acceptance replay has started.

Portable store added in `cmd/sdlc/bench_store.go`: one bounded raw record, accepted-event replay through Step, nested future-field retention, separate nonblocking lifetime/transaction locks and durable write-before-effect ordering. Test-first controls passed (0.552s). Two additional behavioral reds exposed creation inside an aliased source root and an oversized enclosing record leaving partial setup state; fixed physical-source comparison and envelope preflight before writes, with owned-root cleanup on setup failure. Tracked golint PASS (1.846s), staged whitespace PASS. Next atomic step is sanitized prepare plus real temporary-Git/CLI integration; runtime/TL, measurement, score/ledger, recovery/docs and the mandatory hub-provisioned true replay are still outstanding. No full suite or live bench setup is running.


Sanitized preparation, public context guards, bench-local publication, result codec, append-only ledger and TL forwarding are implemented as bounded modules. Preparation runs real migration/CAS in an independent Git repository and verifies exported product bytes/modes after Git attributes. Behavioral reds caught missing/wrong-type pins and attribute conversion, alongside sibling-artifact leakage controls. Canonical private file identities now flow through the existing suite attempt/envelope validators; routing remains selected only by validated bench context. Receipt reads revalidate the newest native attempt, log digest, success guard and current known-flake policy, refusing stale prior successes.

Native built-CLI conformance (owned disposable binary, removed afterwards) passed real local-check run and republish with an all-call rejecting GitHub sentinel: zero calls, extension preservation and tampered-log refusal. TL forwarding retains the actual gate exit code and bounded managed subprocess lifecycle. Public command audit captures normal refusals and legacy fatal exits, preserves the underlying error, and prevents effects when audit persistence fails; aliases resolve before the publication guard. Decisions/proposals remain evidence, not gate authority.

Fresh root verification after all assigned mutations stopped: focused SDLC bench plus ordinary run/republish/store/flake controls PASS (36.107s); TL bench/root-help controls PASS (0.218s); tracked golint PASS (1.957s). Pure bench package race PASS (1.443s); local-only blob/tree controls PASS (1.106s). Ledger tests cover partial tails, conflicts, exact retries, capacity, concurrent writers, source-field preservation and counted N/2N scans. These are component/CLI checks, not the required real acceptance run.

Next: integrate the observer CLI and stateful Couch adapter, trusted launch-profile/session-log cross-check and explicit retirement-proof handoff, exact would-land admission, captured-source measurement/leak normalization, managed hidden scoring, recording/recovery/retention and atlas. The TL wrappers currently forward to the shared protocol that is still being connected; this checkpoint is not an end-to-end runnable M1. Then run full verification and provision the mandatory autonomous replay with ops:0 before close. No full suite, live bench slot, stage or acceptance replay is running.

Observer integration now connects prepare/status/run/decision/finish/cancel/score/record/clean. Exact Couch actor and native conversation bindings, dispatch receipt observation, original budget, root-bound public proposals and explicit trusted retirement keep uncertain outcomes visible. A lost reply never redispatches; a new conversation in the same UI cannot inherit candidate authority. Pre-dispatch partial bindings name explicit root retirement and a fresh run. Candidate proof uses the approved strict close-owner relation; no metadata head or TL assertion replaces actual suite/SHIP evidence.

Captured-source collection uses one bounded immutable prefix per explicit worker/TL/reviewer input, native identity/usage/active-time owners and native audit/gate/accounting metrics. Unknown counters stay unknown; actual profile mismatch makes the run invalid and skips hidden scoring. Immutable source pins and a sealed collection envelope support interrupted measurement publication. Scoring runs only after persisted effect authorization; unknown scorer cleanup retains the exact protected native bundle in `scoring`, without acceptance, recording or cleanup permission. Recording freezes exact evidence bytes/digest and a single row/destination; lost append acknowledgement retries only those facts. Cleanup and retention preserve source, raw evidence and the external ledger.

Fresh root combined focused checks PASS (76.700s): bench preparation/context/local-check/audit/runtime/control/admission/collector/measurement/scorer/ledger/cleanup plus recovery-contract controls. This includes the explicitly synthetic complete replay through real temporary Git/tracker gates, captured measurements, stateful scoring, immutable JSONL and cleanup. Pure bench and agent identity race checks PASS (1.711s and 1.784s); TL forwarding/help controls PASS (0.204s). Additional focused record controls PASS (0.854s), measurement controls PASS (15.403s). These component/synthetic checks do not qualify the mandatory real-agent replay.

Native disposable Couch service qualification attempted through Pair's approved shared bridge. Build succeeded, but the actual run failed before readiness because its replacement private PATH omits Apple Git's `git-upload-pack`, needed by local upstream fetch. Failed bridge reported joined cleanup; advertised fleet and candidate roots were independently absent. Filed unclaimed pair#845 via `sdlc issue new` from existing branched pair:4 (no checkout writes), and reported to ops:0. No live agents were launched or modified. The failed native run is an infrastructure/harness blocker, not a conformance pass. An earlier retained diagnostic directory remains untouched.

Additional root close-owner/native mirror and recovery controls PASS (76.007s); staged lint PASS (1.890s), whitespace PASS. The unavailable-acceptance terminality test first failed on a second score authorization, then passed with the approved narrower phase rule (whole pure package race PASS 1.757s). Recovery/scoring/partial-binding/command registration controls PASS (3.155s). Embedded help/placeholder wiring PASS (0.962s), TL Bench/root-help controls PASS (0.140s). Go reports 1.27.3 and the shared default SSD build cache; no per-run cache override.

Integrated checkpoint pushed at `d4ee07f11ec286b3ad61400b708bed96fef313d1`; own-slot SDLC/TL tools built with Go 1.27.3 and verified that exact revision with vcs.modified=false. The recorded suite queued as lease-3779395426, position 4, before execution. Final recovery inspection then found that a crash during a later proof could revisit an earlier rejected proposal. Added a behavioral regression (red: earlier rejected candidate verified); fixed selection by the existing pure Run's exact pending proof-operation identity, removing the historical scan. Runtime controls PASS (15.976s); the pending-proof plus metadata-reader-bound controls PASS (3.637s). Cancelled only the queued lease through `sdlc local-check cancel --run lease-3779395426`; its owner exited interrupted without a test result. No test failure or flake is inferred from that cancellation. A fresh pushed corrected-head recorded suite is next.

Read-only v1 corpus feasibility selected ariadne#491 over parley.nvim#403: Parley's pinned product tree omits ignored workflow/helper tooling, while Ariadne's pinned parent contains both without an overlay. Exact parent/landing/original-issue pins and a test-only Python lock-contention oracle were verified through Git. No source edits/fetches, live roots or paid agents were created. Next: push the bounded correction, rebuild exact tools, prepare the selected immutable task and request disposable-slot provisioning from ops:0. Mandatory real autonomous replay, native service qualification, recorded full suite, fresh close review and landing remain outstanding. No current suite lease, stage or real bench run is active.

Corrected implementation checkpoint `8d94d2c737e8376e4ccdfe4dff50e0c2136b014b` is pushed; candidate `bin/sdlc` and `bin/tl-lead` are both Go 1.27.3, that revision, vcs.modified=false. Recorded suite is retained in background terminal 72740, lease-4253409475 on that exact head, inflight attempt 2998515e7302390fe41823134ebff4f8; latest observation was queued position 2. Do not duplicate or restart it. If Couch's scheduled provisioning restart affects this agent, audit its exact live lease or saved outcome before any rerun.

Real v1 #491 preparation succeeded on private SSD storage, with no agents or autonomous work launched. Run ID: `4c0928ea83953f5837dd728e74af526e`. Root: `/Users/xianxu/.local/state/ariadne/model-bench/runs/v1-ariadne-491-r1-5f177ee8`; private task: `/Users/xianxu/.local/state/ariadne/model-bench/inputs/v1-ariadne-491-r1-5f177ee8.json`. Public hub/work/task paths are its `public/hub`, `public/work`, `public/task.json`; hidden patch and provenance remain outside public. Parent `d627a362755a1c2f13315492db52d1f4caeaefc4`, landing `bcd6510d7ce17c3cdd3aa17073abdac59e3357ad`, original issue blob `a671752359defdcf26c59fa59b1e8404131d2718`. Captured test-only path is `scripts/test/test-lock-pause.test.py`; 7,574 patch bytes, SHA256 `1bc5c841e998eea20bc74ad9841fa22830f9e1d92d3eecb10e735ed7870a3022`; literal acceptance argv is Python3 plus that path. Original measured hours 0.22; no solution or private source pin was projected to agents.

Ops approved TL/reviewer `claude/claude-opus-5-5/high`, worker `codex/gpt-6.1-sol/medium`, and real observer-only ledger `/Users/xianxu/workspace/ops/teams/devinfra/projects/model-bench/ledger.jsonl` (ops:0 commits after the run). Prepared hub `.hub` is exactly `work:0` plus newline, committed only within that disposable hub to keep it clean. Verified from Pair docs/source: append only public/hub to the operator's existing `couch . ../ops-echo` command, run from workspace/pair. Positional roots become self-owned hubs. `.hub` claims but does not launch the worker; the operator opens public/work through the switcher's global Start form as an existing primary, producing `work:0` under `hub:0`. No symlink or new launcher. Operator pins models through existing agent parameter forms and Fresh/reboot, whose saved launch profile precedes defaults; final actor session/generation must be collected after that. Explicit parameters replace argv, so normal approved permission flags must be retained.

The proposed global candidate PATH was rejected by ops:0 because it would affect every production actor. Corrected to command-local routing in observer preflight only (approval recorded below). A real candidate TL-to-SDLC routing probe outside any bench context safely refused the proposal and printed the exact 8d94d2c7 candidate SDK path/revision; it produced no runtime/status/proposal effect. Do not change global settings or live fleet PATH. The operator provisioning steps are reactivated; wait for ops:0's completion notice. Next: verify exact actors/roots/native sessions, send observer-only tool preflight and hold instructions, validate tool replies and trusted launch/claim evidence, then write the provisioning file and start the observer with the approved ledger and explicit budget. This preflight is not operator input during the autonomous interval. Record any bare-tool slip as a bound TL ruling/workflow finding. Real replay, native service proof (pair#845), full recorded suite result, close review and landing are still outstanding.

### 2026-10-23 — Metis re-plan, clean implementation hold

The implementation was clean and pushed at 75a9ffb87551bc194be8ec26eb7c8eb56f9d0c69 when the authenticated operator scope change arrived. No bench actor was dispatched and no hidden evaluation occurred. The old provisioning/preflight recipe is superseded: do not start that standalone run. Read the ops provisioning side doc's Second look and the actual Metis/kbench code; appended a replacement execution plan using Metis as sole record/cache/ledger owner and the new couch-bench repo for four ordinary steps. Human grading is a separate never-cached step, with human waiting excluded from automated time and retained separately. The hub owns edits to the operator's side doc/project; this worker did not change them.

The retained recorded suite on code head 8d94d2c737e8376e4ccdfe4dff50e0c2136b014b finished (attempt 2998515e7302390fe41823134ebff4f8, 841.0 s execution after 2700 s queued). Packages and seven shards passed; shard3 reported the top-level TestEveryCommitSiteStamps failure for bench_prepare.go:169 and bench_score.go:215, whose synthetic Git commits lack agentid.StampCurrent. This is a branch defect, not a flake. Saved native log: /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/2998515e7302390fe41823134ebff4f8.log. Fix both sites within the approved preparation/scoring refactor and rerun on its final pushed head; no suite is currently retained or requeued. Report to ops:0 adfa37ce-5c55-4f13-8c5b-cce2b1dcb7b9 was confirmed submitted.

Read-only Metis exploration found input-addressed downstream cache keys, sweep-only ledger emission, repeat-dedup/overwrite risks, no native Go read capture by the Python sensor, and generic artifacts that do not preserve live checkout symlinks/modes. The revision therefore uses an immutable archive template, explicit tool/source provenance, fresh mutable session roots, no duplicate run journal, conservative no-cache scoring after a stochastic session, and typed evidence beside numeric metrics. Filed metis#71–74 with full details using sdlc issue new --no-claim; all four published card/details and touched no Metis checkout files. Its pre-existing construct/local/ml-research/SKILL.md edit remains untouched. Ops provisioned an empty clean couch-bench repository; no files were seeded there before plan approval.

Stopped after writing the revised plan and filings. Next: send this committed/pushed plan to ops:0 for approval, then re-enter change-code, extract the independent backends and seed couch-bench. The four upstream Metis capabilities must land before the true Metis acceptance run. Original native disposable service proof remains blocked on pair#845; keep that limitation explicit until conformance is exercised or the hub approves an evidence substitution. No close, stage, merge or actor-control action has started.

## Revisions

- 2026-10-23, operator-directed scope replacement relayed by ops:0 (`e56c25ec-3157-4d2d-b909-fb833b517ecf`, `f28bfafd-75e4-4f81-993c-acf79910fa1a`, `5bb56e07-05fd-4b83-9d3f-763472a1246c`, verified canonical submitted receipts): use Metis directly. Replace the independent SDLC bench run record, JSONL ledger/cache/lifecycle with Metis's experiment, step execution, record.json and ledger. Keep preparation, session drive, measurement and hidden score as step backends; add never-cached optional couch/grade with attributable human result and separately excluded human waiting. Seed the operator-created private couch-bench as a Metis consumer with steps/couch/{prepare,session,score,grade} and an ariadne#491 experiment. The durable plan's appended replacement execution plan is the proposed implementation contract; older Spec/Plan text is historical where superseded. Done-when still requires a true autonomous replay through real gates, exact approved profiles, supported retirement and autonomy attestation, hidden score, and now a real Metis record/ledger row plus explicit grade or permitted skip. Trusted prepared-repo isolation, leak detection, strict close proof and no gate bypass remain unchanged. Four upstream gaps are filed unclaimed: metis#71 (never-cache), metis#72 (timeout/live log), metis#73 (repeat rows), metis#74 (human grading). No implementation under this revision starts before hub approval and change-code.

- 2026-10-23, hub approval `d5e4ef5c-b690-465c-831f-f7d38055f8d5`: candidate tool routing is command-local inside only the disposable TL/worker subprocesses, carried in observer preflight. Do not prepend candidate tools to the live supervisor's PATH. Use explicit candidate SDK paths plus command-local PATH so TL's bare SDLC child resolves to the same build. Any routing slip is retained as a workflow finding via the existing bound TL decision/audit evidence.

- 2026-10-23, hub approval `af9f0c58-1b5e-4c4c-9c02-e08286dabb30`: each five-second poll may replay up to 2,000 accepted transitions and read up to 1 MiB of proposal audit; transcripts are never polled. Pin these metadata bounds in a reader test; no cursor/cache mechanism. Collection keeps the previously approved one-capture/fixed-pass scope.

- 2026-10-23, hub approval `897cbcae-be1b-4147-93fb-f5a83168a0d0`: confirmed acceptance unavailable is terminal for that run and recorded as a bench finding. Keep its first acceptance evidence immutable; another hidden evaluation needs a fresh run. An unconfirmed scorer continues to block recording and cleanup. This narrows the earlier same-candidate score-retry recovery wording without adding per-score state.

- 2026-10-23, authenticated hub approval `74c78252-ff37-48a3-8e8e-47fc92fa6e1a`: trusted retirement attestation may explicitly confirm no operator input for the exact bound run interval. Only that evidence permits autonomous=true and zero operator wait; otherwise autonomy/wait coverage remains unverified. Existing Couch exposes exact known receipts but no operator-submission history query, so M1 adds no new runtime mechanism and never infers autonomy from missing messages.

- 2026-10-23, authenticated hub approval `2d92cdc9-b095-4f6c-8e69-316dbcfbd7d0`: exact reviewed-code normalization belongs to the close owner. Accept only its direct evidence child with exact task detail/review sidecar/gate paths and optional exact mirror refresh. Extra documents, project updates or intervening commits refuse as a visible bench finding; full suite and SHIP must still bind the reviewed code candidate.
- 2026-10-23, authenticated hub approval `c523bd98-2b52-41cf-b0c2-8b818434d9bd`: capture each bounded transcript prefix once, then perform a fixed number of parsing passes through existing measurement/usage and tool-normalization owners. Count each pass at N and 2N; existing aggregation sorting is outside the linear collection claim. No cross-package stream or replacement arithmetic.

- 2026-10-23, authenticated hub approval `c47f95a3-bd2d-4f42-bace-62bd4945391f`: provisioning supplies trusted exact model/effort evidence, cross-checked against the actual agent transcript/session log; mismatch makes the run invalid. A Couch interrupt certifies a UI turn, not descendant shutdown. Remain `needs-reconciliation` until the checklist's explicit supported park/archive retirement and confirmation that the slot processes are gone. No Pair mechanism is added.

- 2026-10-23, authenticated hub approval `9ba79fcc-0dfb-4f57-9290-b03dc307b75b`: strip all old `workshop/issues`, `workshop/plans`, `workshop/history` and `workshop/issue-tracker.json` process inventory from the parent-tree replay, then overlay only the target's original Spec and Done-when and use normal migration/CAS. This prevents archived sibling plans from leaking answers and avoids orphaned tracker mirrors. Product code at the pinned parent stays exact. The leak-freedom fixture must prove sibling artifacts absent.

- 2026-10-23, plan approval condition: Done-when requires a true autonomous replay with real agents in disposable bench slots on this Mac before close. Added provisioning checklist to the durable plan; ops:0 will provision together when the code is ready. Synthetic rows remain tests only.

- 2026-10-23, hub-approved isolation scope: M1 proves sanitized prepared-repository isolation with trusted agents; post-run detection flags outside reads/source/scorer access, fetch/clone and GitHub access as possibly leaked. OS read/network isolation is a later milestone required before co-locating bench and development checkouts.
- 2026-10-23, hub-approved Done-when clarification: original/source remote absent; allow exactly one fresh private bare origin holding sanitized replay objects, so real tracker CAS/gates work without production publication. Only local-check status publication uses an explicitly bench-local sink; no gate bypass or fake production success.
