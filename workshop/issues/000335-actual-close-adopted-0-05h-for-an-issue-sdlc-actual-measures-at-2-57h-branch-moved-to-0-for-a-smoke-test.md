---
id: 000335
status: working
deps: []
github_issue:
created: 2026-10-11
updated: 2026-10-14
estimate_hours: 4.174
card_mirror: 'd64c47a033a3591cb3ed7317d8f6a98a52ba5064' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T01:44:37-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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
- pair#454's actual is corrected.

## Plan

Proposed durable plan: `workshop/plans/000335-actual-measurement-plan.md`.

- [ ] Reproduce checkout source loss and release/resume inflation.
- [ ] Derive ownership intervals/source provenance and unify consumer inputs.
- [ ] Add adoption diagnostics, correction/recovery, and production regressions.
- [ ] Verify and land, then repair pair#454/#514 through measured correction.

## Log

### 2026-10-11

### 2026-10-14 — root-cause investigation

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
