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

## Revisions

- 2026-10-23, plan approval condition: Done-when requires a true autonomous replay with real agents in disposable bench slots on this Mac before close. Added provisioning checklist to the durable plan; ops:0 will provision together when the code is ready. Synthetic rows remain tests only.

- 2026-10-23, hub-approved isolation scope: M1 proves sanitized prepared-repository isolation with trusted agents; post-run detection flags outside reads/source/scorer access, fetch/clone and GitHub access as possibly leaked. OS read/network isolation is a later milestone required before co-locating bench and development checkouts.
- 2026-10-23, hub-approved Done-when clarification: original/source remote absent; allow exactly one fresh private bare origin holding sanitized replay objects, so real tracker CAS/gates work without production publication. Only local-check status publication uses an explicitly bench-local sink; no gate bypass or fake production success.
