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

- [x] Reproduce denied settings publication and trace first-run/second-run behavior.
- [x] Make protected settings denial nonfatal with one actionable stale notice; retain ownership and unrelated error safeguards.
- [x] Verify settings recovery and start-plan behavior; run focused checks and the recorded suite.

After verification, close through SDLC and land on SHIP.

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

- Exact-head host native conformance on bf379af2945f027cfc838fa0008b5dda406edd79 passed in 56.151s; saved log /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne411-native-14mgpyl6.log. Both native write-denial controls held; real refresh and first start-plan succeeded, stale/current notice counts 1/0.
- Expanded start-plan assertions passed, but its first run's hermeticity guard correctly detected my checkpoint moving HEAD while it ran (0c5b2e11 to bf379af2). Reran without moving HEAD: go test -trimpath ./cmd/sdlc -run 'Test(StartPlan|RunInstalledWeave|GateSettings)' -count=1 -timeout 10m PASS 59.980s. No gate waiver or test weakening.
- PR #270 opened; merge-check and local-gates-guard green. Recorded suite pinned to bf379af2 is queued first behind a remaining suite; no custom gate override. Current main already integrated.

- Recorded suite on bf379af2 failed only in scripts/test TestConstructScripts/plain: runner hit its unchanged 120s ceiling while dev-aliases.test.sh was active. All 12 SDLC shards passed: 1522 tests; 66 other packages were attempted. Suite execution 1027.7s; log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/68c1f404e315cdbc41260d1dd80c0308.log (sha256 prefix c36ed9ace6ec). No suite bypass.
- Isolated candidate plain rerun reproduced the same 120.239s timeout. Runner, construct tests and emitter are byte-identical to current main. Current machine load observed 167.47, diagnostic only; no custom load gate introduced. A scratch main checkout control is running; its result will determine whether this is independent of #411.
- Hub timeout question send returned uncertain receipt 28fc3ab8-32bc-4d6c-8ffe-901b8df59270; receipt query then unavailable after a restart and couch intermittently absent on PATH. Did not blindly resend or use another channel.

- Main control at 87e0feeb63e41946936ccb316583629ce8eb6b8a passed the unchanged plain shell test in 66.861s; log /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne411-main-control-l273bo38.log. Scratch main worktree removed cleanly. This did not reproduce a main failure, so no claim of an established main bug.
- Candidate plain rerun then passed in 80.640s (all six construct scripts, zero failures). Together with the earlier 120s timeouts and unchanged scripts, this supports a load-sensitive timeout. No code/test-budget change or waiver; rerun the normal recorded full suite.

- Second recorded suite on 986653326c2242c56ef7a095a3ebab498f98918f is running; its plain construct-shell run passed in 93.84s, resolving the prior timeout in this run. A different unchanged package failed: TestConcurrentAcquireReclaimsDeadHolderOnce, repolock_test.go:415, sdlc.lock/meta.json ENOENT (0.03s).
- Filed independent ariadne#413 unassigned with that captured failure and the suspect ProcessAlive fake (returns false for original dead holder and both new live contenders). Test blob matches main exactly: 9f1e403fe782fd03b405c2f2f9adb47f54b183f0. A scratch main 100-run control passed in 4.158s; log /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne411-lock-main-gezq_l3f.log; checkout removed. Do not claim a production lock bug or reliable main reproduction from this evidence. Hub blocker report 6d733704-0bca-4f42-b9d2-d64b494acf2b was admitted, but receipt query hit couch unavailable on PATH. No #413 code folded into #411, and no suite waiver.

- Confirmed #413 on main 35ba24e6cad39344c347af981de7ac8fcc1d1c31: 18/1000 repetitions failed with the same missing metadata in 47.462s; log /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne411-lock-main-1000-60b_vs0t.log. Reproduction published in unassigned #413; hub receipt 62726fe8-0389-4ff5-9a4b-6b888976fe06 confirmed submitted. Scratch checkout removed.
- Second full suite finished: all 12 SDLC shards over 1522 tests and construct scripts passed; only independent #413 failed in the other-package job. Runtime 948.3s (15m49s), saved log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/5f54a534ecf74f0c3b0c7af2873769a8.log, hash prefix 7f4685833154. Preserve the failed status; run one further unchanged full-suite retry through the standard gate. No code change, budget change, test exclusion or validation waiver.

- Both required full-suite executions are recorded above with their failures, not represented as passing. The third unchanged run is queued on b0b8a2de. Start the mandatory close review with all feature/host evidence and the independent reproduced main failure explicitly disclosed; merge validation still requires a passing recorded suite, with no --no-validate or other waiver. The checked Plan row records execution and feature verification, not a clean whole-repository suite.

- Close review returned REWORK: BR-1 found declaration-removal retirement still mutates protected settings; generalize protection to the destination across all actions and ownership retirement. BR-2 moves filesystem-backed managed ownership to the plan Integration table. BR-3 requires revised-head native host evidence because the reviewer cannot nest sandbox-exec. The obsolete queued third suite was interrupted before execution (owned CLI PID 84391, exit 130); no peer job or lock touched.

- BR-1 fixed in 7c362a27c316ad14b3ba62889ae3510ff60480fb: always defer the protected destination; exclude alternate/generated writers; retain removed-declaration identity/backups, allow ancestor scaffolding, and detect retirement residue. New tests reproduced RED then passed. All go test -trimpath ./cmd/weave/... -count=1 -timeout 5m passed (main 67.806s, plan 33.485s, refresh 76.199s). BR-2 corrected in eb3923d4: ownership is an IO integration through weavefs.
- Expanded native fixture initially lacked a backup because first adoption does not displace; corrected it with a real edit and second compile (aa326d59db8cf05a962daa7dc9f5e40764305f27). At that exact head, ARIADNE_NATIVE_SANDBOX=1 go test -trimpath ./cmd/sdlc -run Test(StartPlan|RunInstalledWeave|GateSettings) -count=1 -v -timeout 10m PASS 122.390s; native stale/current/retired PASS 56.41s, denied-write controls, first planning success, notice counts 1/0/1. Retired settings, backup and inventory remain byte-identical. Log /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne411-revised-native-9iww01mg.log. BR-3 dispute records this host evidence; reviewer nested-fence failure is not a production retry.
- The 7c362a27 suite was canceled while queued (owned PID 12411, cwd verified, exit 130) after correcting the native fixture. Requeue the revised head through the normal gate; no failed or canceled suite is represented as passing.

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

### 2026-10-16 — separate validation from future landing

The Plan's last checkbox now covers the verifiable checks; close and landing remain subsequent SDLC actions rather than self-dependent checkbox claims. Acceptance and scope unchanged.

### 2026-10-16 — review while unchanged validation retry queues

The required suite was executed twice before close review; both outcomes remain recorded verbatim. Feature acceptance is proven by focused and native tests. Start fresh review while the unchanged full-suite retry waits, keeping merge's passing-suite requirement intact. No acceptance, scope or gate waiver.

### 2026-10-16 — close-review scope refinement

Protect the destination across declaration removal and alternate/generated writers, retaining prior identity and backup evidence until an ordinary operator refresh reconciles it. Correct the plan ownership IO classification; native retirement denial joins the existing stale/current controls. No permission inference, write retry, gate waiver or change to catchup HEAD safety.
