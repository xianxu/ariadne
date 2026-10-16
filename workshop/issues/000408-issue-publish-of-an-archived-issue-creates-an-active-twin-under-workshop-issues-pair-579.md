---
id: 000408
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours: 0.94
card_mirror: 'bcacd854d9ebf6ca374fb6920ff4398f5e9e821f' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T11:18:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# issue publish of an archived issue creates an active twin under workshop/issues (pair#579)

## Problem

2026-10-16: pair:3 ran `sdlc issue publish --issue 579` to add a Log entry to pair#579, which is done and archived
(`workshop/history/issues/000579-*.md`). With the default issues dir, publish created a NEW active twin
`workshop/issues/000579-*.md` (47 lines) on pair main (commit 7af5c7b1). Re-running with
`--issues-dir workshop/history/issues` published correctly to the archive. No verb removes the stray active copy.

## Spec

- `issue publish` resolves the issue's details file wherever it lives (active or archived), the same way
  `issue show` does, and refuses to create an active file for an issue whose card is done or archived.
- If both an active and an archived file exist for one issue, publish refuses and names both.

## Done when

- Tests: publishing an archived issue updates the archive file and creates nothing under `workshop/issues/`; a
  done-card issue with no details refuses; the twin case refuses.

- A local-only archive with no published details refuses, preserving its bytes,
  leaving main unchanged and creating no active details.
- Initial publication, retries after moved main and saved-receipt recovery
  refuse archives introduced after preflight, for both archive layouts; no
  active write occurs and the source is preserved.
- Backup and non-Markdown entries are ignored by local details discovery.

## Plan

- [x] Reproduce archive updates, missing terminal details and duplicate-location
  refusals in isolated real-Git publication fixtures; confirm red.
- [x] Resolve details by ID across active and archive locations on pinned main,
  preserve republish safeguards, and recheck location before publication.
- [x] Run focused and recorded full-suite verification; update Atlas and prepare
  the close review.

## Log

### 2026-10-16 — close round 2 stopped for TL decision

Close returned REWORK, leaving the card working. The independent reviewer confirmed BR-1/BR-2/BR-3 addressed, including mutation evidence: removing each behavioral fix reproduced its six failures. New Critical BR-4 shows terminal-card admission can become stale between preflight and initial publication or saved-candidate recovery; the reviewer reproduced active creation and source removal after a card became done. No further code has been changed. Per the TL's explicit instruction to stop before a third mechanism change, the next correction waits for the TL decision.

Proposed correction for review: retain the same BeforePublish seam and reachability-first behavior. Read the current card through env.repo.Snapshot/Require, then call the existing pure requireInitialDetails before the existing location/asset checks. The shared unconfirmed-publication rule becomes current nonterminal card AND absent ID-wide details. Extend the current stateful fixture to change card status across first/retry/saved-receipt boundaries; assert refusal, no active main write and source retention. No new state, receipt or hook. Re-run targeted tests, recorded suite and close after approval.


### 2026-10-16 — final correction full suite

Recorded suite passed on 30acd23c1e70d785dacce6d86d09e851941d7145: 1,516 SDLC tests across 12 shards and 66 other packages, zero failures, 701.4s wall (11m42s recorded). Evidence: sdlc-local/suite success, log sha256:18c999319064. The final regression group passed in 84.718s and broader publication/transfer/recovery coverage in 207.171s. BR-1 is addressed through the existing BeforePublish hook shared by first/retry/recovery; BR-2 by filtering canonical filenames before family classification; BR-3 by the appended acceptance clauses. PQ-1 prose is compressed to function-specific strategies. No new receipt/state mechanism introduced.


### 2026-10-16 — recorded full-suite acceptance

Isolated `sdlc local-check run --name suite` passed: 1,513 SDLC tests across
12 shards plus 66 packages, zero failures, 685.4 seconds execution after 357
seconds queue wait. This supplies a clean broad verification run after the
earlier worktree-edit guard failure. Atlas maps archive resolution and refusal
behavior. Close review follows; publication happens after SHIP.


### 2026-10-16 — focused verification

Stable-worktree regression run passed: TestPublishArchivedDetails (both archive
layouts and mixed-set atomicity), TestPublishTerminalCardWithoutDetailsRefuses,
TestPublishDuplicateLocationsRefuses (local and published twins),
TestPublishRefusesTwinIntroducedBeforeCandidate, and
TestPublishUnpublishedLocalArchiveRefuses. Runtime 27.370s; diff check clean.
The production delta is 98 added lines, inside the quick-flow shell.


### 2026-10-16 — red reproduction and root correction

Real-Git/bare-origin regressions reproduced both archive layouts losing the
requested edit, done-card missing details synthesizing a new active file, and
local or published twins being accepted. The root correction resolves a pinned
main location before choosing creation versus republish, preserves mixed-set
atomic publication and rechecks candidate locations. Terminal-status membership
comes from vocab rather than a copied enum. Added a concurrent-twin regression
and a local-only archive edge case to defend the no-resurrection invariant.


### 2026-10-16

### 2026-10-16 — design

Root cause: the active-only existence probe classifies an archived record as an
incomplete first publication; move-detail can then synthesize a new active body
from the card. Resolve by ID against the three canonical issue directories
(active, legacy flat history, history/issues), not by inventing the active path.
Reuse existing pinned TrunkView, issue filename parser, republish transaction,
stateful real-Git fixture and existing ownership/asset/concurrent-body checks
(ARCH-DRY, ARCH-PURE, ARCH-MOCK, ARCH-PURPOSE). Re-resolve per candidate before
publication; a concurrent archive or twin refuses, rather than resurrecting an
active copy (ARCH-ORDER, ARCH-SECURE). A lifetime lock cannot reconcile stale
branch/main locations; the existing repo lock already serializes local verbs.
Metadata scans stay bounded to the existing three directories; fetch blobs only
for the selected ID (ARCH-CONSTRAINTS). No new durable artifact family, process
or handle; existing issue and review archive lifecycle applies (ARCH-FUNERAL).

Short plan posted to ops:0 under the explicitly authorized small-fix flow.

## Revisions

### 2026-10-16 — local-only archived source

The no-twin invariant also applies when a local archive has not reached main.
A regression reproduced first-creation recovery synthesizing an active twin in
that state. Refuse and name the local archive instead; do not migrate or invent
content. This is the same location-resolution root cause, not a new workflow.

### 2026-10-16 — verification discipline

The broad publish/move-detail assertions passed, but the real-repository
hermeticity guard failed because the agent edited Atlas and issue records while
the run was active. No fixture escaped its temporary repository. Finish edits,
then rerun against a stable worktree; do not count that invocation as passing.

### 2026-10-16 — checkpoint wording

The final Plan row records verification and readiness for close; publication
follows the review verdict rather than being ticked before landing. No scope or
behavior change.

### 2026-10-16 — BR-1/BR-2/BR-3 full-flow checkpoint

Complete location revalidation in the existing TransferOp.BeforePublish hook,
which is shared by initial publication, retry and receipt recovery. Filter local
entries by the same issue filename grammar as published entries. Append concrete
Done-when clauses for these and the local-only archive state. The expanded fix
is expected to leave the quick-flow diff shell; durable implementation plan:
workshop/plans/000408-archive-location-plan.md. Await hub approval before changing
implementation, then run change-code's full-flow plan gate and derive its estimate
only after plan quality clears.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Calibration is provisional (source marked stale).

Three smaller Go module extensions cover location resolution, republish and transfer admission. Each uses the upper implementation anchor 0.5h scaled to 40%, with 0.15h design discounted to 0.03h by the approved design. Existing Git/tracker facilities cover the stack; no greenfield library needed. Atlas uses 0.05h design ×0.2 and 0.1h implementation ×0.4; boundary review uses 0.1h design ×0.2 and 0.4h implementation ×0.4. Familiarity is 1.0 and the thorough-plan design buffer is 15%.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.03 impl=0.20
item: smaller-go-module design=0.03 impl=0.20
item: smaller-go-module design=0.03 impl=0.20
item: atlas-docs design=0.01 impl=0.04
item: milestone-review design=0.02 impl=0.16
design-buffer: 0.15
total: 0.94
```

### 2026-10-16 — full-flow implementation resumed

The TL approved the durable plan and the existing BeforePublish hook, with no new state and a stop before any third mechanism change. change-code passed plan-quality and estimate-quality; full flow is recorded. New regressions reproduced BR-1 in all six publication/retry/recovery cases across both archive layouts, and BR-2 in all six backup/non-Markdown discovery cases (48.956s run). The direct pure terminal-admission test passes. Implementing the approved hook and existing filename-parser filter.

Estimate scope clarification: the three module primitives cover the whole issue's resolution, republish and transfer changes, including their regression tests; module implementation anchors include focused and full-suite verification. The review primitive includes PR/landing checks. This is the original whole-issue model estimate, not a remaining-work forecast; full-suite and review latency may make the actual higher.

New regression run passed (84.718s): TestPublishIgnoresNonDetailEntries, TestInitialPublicationArchiveRevalidation and TestInitialDetailsTerminalAdmission. BR-1 uses only the existing BeforePublish seam; BR-2 uses the existing issueFilename grammar; BR-3 acceptance clauses are appended. PQ-1 test strategy prose compressed into function-specific lines.

Broader targeted verification passed: go test ./cmd/sdlc -run 'Test(Publish|MoveDetail|AssetRecovery|InitialPublicationArchiveRevalidation|InitialDetailsTerminalAdmission)' -count=1 -timeout 10m (207.171s). Stable worktree; no hermeticity failure. Existing confirmed-candidate recovery remains green. Starting the recorded full suite on the pushed correction.
