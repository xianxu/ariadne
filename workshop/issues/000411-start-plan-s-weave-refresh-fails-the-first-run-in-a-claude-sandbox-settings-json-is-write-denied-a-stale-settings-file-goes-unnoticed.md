---
id: 000411
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours: 1.13
card_mirror: '6144aa383e23659eca6be546b18ccbc64dd88206' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T15:06:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# start-plan's weave refresh fails the first run in a Claude sandbox (settings.json is write-denied); a stale settings file goes unnoticed

## Problem

In a Claude slot (parley.nvim:4, 2026-10-16), every first `sdlc start-plan` fails: its weave refresh tries to write
`.claude/settings.json`, which Claude Code's sandbox write-denies (`operation not permitted`; settings files are in the
sandbox's deny-within-allow list by design, so an agent can't rewrite its own permissions). Re-running start-plan
then goes through, so the slot's `.claude/settings.json` may lag the woven one without anyone noticing.

## Spec

- Find why the first run fails and the second passes (the refresh's change detection, or a partial write).
- A gate must not fail on a write the agent's sandbox forbids by design. Either skip the protected settings file when
  unwritable and report "settings.json is stale: run `weave refresh` from your shell" as a one-line notice (one state,
  one action), or move the settings write out of the gates entirely into an operator-run step.
- Never retry the write unsandboxed.

## Done when

- A test with an unwritable `.claude/settings.json` shows start-plan succeeding on the first run with the stale-settings
  notice, and the notice clearing once the file is refreshed.

## Plan

- [ ] Reproduce denied settings publication and trace first-run/second-run behavior.
- [ ] Make protected settings denial nonfatal with one actionable stale notice; retain ownership and unrelated error safeguards.
- [ ] Verify settings recovery and start-plan behavior, run focused checks and the recorded suite, then close and land on SHIP.

## Log

### 2026-10-16

- Root cause: catchUpBeforeWork only invokes refresh for behind HEADs. Refresh advances HEADs before compiling; publication denial fails the first run after Git progress, then the retry skips weave because HEADs are current. Preserve the independent moved-HEAD rebuilt-binary rerun (#303); no sandbox bypass. ARCH-DRY/PURE: reuse settingsx comparison, weavefs and managed ownership rather than classify arbitrary gate failures. Plan posted to hub; TL authorized proceeding after posting.

- Native RED: installed weave compile under sandbox-exec exact settings write denial exited 1 with EPERM at atomic rename; touch denial control failed and original bytes survived.
- Root cause proved: catchup.go skips weave when HEAD is current; refresh advances HEAD before compile. Its independent rebuilt-binary rerun remains intact.
- Plan-quality passed after addressing PQ-1 (issue-branch-compatible operator recovery) and PQ-2 (function-level strategies). Full change-code passed; estimate quality INFO notes a clean-review assumption and verification/landing variability. Each module allowance includes its focused regressions; native conformance and suite execution/landing share the integration allowance, while suite queue time is idle rather than active effort. Calibration source marked stale/provisional.
- Ownership TDD RED unknown action/not-produced, then whole plan package GREEN. Root reviewed agent diff; no desired settings identity fabricated, old metadata/backups retained through repeated deferral.
- CLI TDD RED unknown settings command/defer flag, then GREEN. SDLC TDD RED missing notice and un-deferred argv, then GREEN (23.351s including unchanged HEAD-move safety tests).
- All weave packages passed (including refresh 211.384s); latest recovery regression passed (7.050s) through actual refresh --substrate on an ahead issue branch, preserving branch/HEAD and clearing notice.
- Native production-seam GREEN: ARIADNE_NATIVE_SANDBOX=1 go test -trimpath ./cmd/sdlc -run '^TestStartPlanProtectedSettingsNativeSandbox$' -count=1 -v -timeout 10m passed in 18.402s. Candidate weave built from this source; exact settings denial confirmed; real gate refresh and first start-plan succeeded with stale notice count 1, current count 0. Private fixture substrate manifests and generated attributes were corrected before this green run. No sandbox bypass or production retry.

## Estimate

*Produced via brain/data/life/42shots/velocity/estimate-logic-v3.1.md against baseline-v3.1.md. Method A only; source flagged stale, provisional.*

Three smaller Go extensions: managed deferred ownership, read-only settings/refresh CLI, and start-plan wiring. Existing planner, settingsx, weavefs and Git fixtures supply the implementation pattern; no novel library needed. Each uses v2's high-end 0.3 design x 0.2 thorough-spec discount = 0.06, and 0.5 implementation x 0.4 v3.1 = 0.20. Docs use 0.1 design x 0.2 = 0.02 and 0.2 implementation x 0.4 = 0.08. One boundary review uses 0.02 design and 0.5 implementation x 0.4 = 0.20. Familiarity 1.0, thorough-plan design buffer 15%; 0.22 x 1.15 + 0.88 = 1.133h, rounded 1.13h.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.02 impl=0.20
design-buffer: 0.15
total: 1.13
```

## Revisions

### 2026-10-16 — choose operator settings publication

Native denial reproduced before implementation. Simply swallowing publication errors is unsafe because managed displacement may already have removed an earlier backup. Choose the Spec's explicit operator-run settings write alternative: gate refresh defers settings and read-only observations report staleness even when Git is current. Preserve #303's independent rebuilt-binary rerun after HEAD advancement. Detailed implementation: workshop/plans/000411-protected-settings-plan.md. TL authorized posting the plan and proceeding.

### 2026-10-16 — plan gate refinements

PQ-1: use the single action weave refresh --substrate from your shell so operator correction works on an ahead issue branch; preserve issue branch/HEAD. PQ-2: plan names prepareManaged/commitManaged/ApplyManaged, deferredClaudeSettings/buildSettings and catchUpBeforeWork/startPlanBranch with adversarial strategies. The full plan gate was entered explicitly after the concrete code contour exceeded quick flow. Native red and CLI red recorded; current focused CLI and plan package tests pass; SDLC argv red shows refresh lacks --defer-settings.
