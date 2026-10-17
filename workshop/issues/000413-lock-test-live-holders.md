---
id: 000413
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
severity: medium
card_mirror: '2262f798d369b457e1d72efa73ea0b238ee9c19e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T17:13:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "0bb87194", done: "7f1ba37e"}
---

# Concurrent dead-holder lock test marks its live contenders dead and flakes with missing metadata

## Problem

During ariadne#411 recorded suite on 986653326c2242c56ef7a095a3ebab498f98918f, cmd/sdlc/internal/repolock failed TestConcurrentAcquireReclaimsDeadHolderOnce (0.03s): repolock_test.go:415 reported write sdlc repo lock metadata: open .../.git/sdlc.lock/meta.json: no such file or directory. The package is unchanged from current main. The full suite is still collecting other results; run ID 5f54a534ecf74f0c3b0c7af2873769a8.

The fixture supplies ProcessAlive: func(int) bool { return false } for the original dead PID 555 and both newly acquiring holders 111/222. This makes the live holders appear dead as well. That is a likely invalid liveness model; this observation alone does not establish a production lock bug.

## Spec

Model the original dead holder separately from the two live contenders, and assert safe ownership/release. Preserve production lock safeguards; change production only if a faithful liveness model demonstrates a production defect. Add deterministic ordering where needed rather than extending sleeps or budgets.

## Done when

- A delayed stale observation cannot reclaim a replacement live holder; reclamation is serialized and revalidated against the current directory generation before mutation. Busy/changed-generation retries honor the existing timeout and cancellation, and successful reclamation leaves no new durable artifact.

- The concurrent-dead-holder test represents live contenders correctly and repeatedly passes while checking exclusive ownership and successful release.
- The observed ENOENT regression is addressed without weakening lock safety or merely increasing timeouts.

## Plan

- [x] Reproduce the existing failure, then replace sleeps with controlled contender ordering; assert exclusivity, metadata ownership and release with only PID 555 dead.
- [x] Guard reclamation with candidate-directory flock, generation/inode and liveness rechecks, and cancellation/deadline checks before rename.
- [x] Run repeated and race-enabled repolock tests, including delayed stale observations, busy guards, late cancellation, changed generations and initializing metadata.
- [ ] Update atlas and record the full suite on the pushed head.

Then close through SDLC, land on SHIP and report the hub.

## Log

- Found while verifying #411; file independently and leave unassigned.
- Scratch main control at 87e0feeb63e41946936ccb316583629ce8eb6b8a, go test -trimpath ./cmd/sdlc/internal/repolock -run '^TestConcurrentAcquireReclaimsDeadHolderOnce$' -count=100 -timeout 5m, passed in 4.158s. This did not reproduce the intermittent failure. Log /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne411-lock-main-gezq_l3f.log; scratch checkout removed.

- Confirmed main reproduction: at 35ba24e6cad39344c347af981de7ac8fcc1d1c31, the same command with -count=1000 failed 18/1000 repetitions in 47.462s, each with missing .git/sdlc.lock/meta.json at repolock_test.go:415. Log /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne411-lock-main-1000-60b_vs0t.log. Both test blobs are 9f1e403fe782fd03b405c2f2f9adb47f54b183f0. Scratch worktree removed cleanly. This establishes a main test failure; it does not establish a production lock bug because the fixture's liveness oracle marks the new live contenders dead. Hub notified with #413 reference.

### 2026-10-16 — plan

Claimed and started from current main on its own branch. ARCH-MOCK/ORDER: liveness is PID-specific, ordering is controlled through existing ProcessAlive/Sleep seams, and every worker is canceled/joined before fixture teardown. ARCH-PURE/DRY: reuse Observe/Acquire and real filesystem ownership rather than another lock model. No new persistent family or production surface is intended. Inspect the stale-observation-to-rename boundary too: fixing the oracle alone must not conceal a real reclamation race. If the faithful model fails, record the counterexample and revise scope with the TL before changing production.

### 2026-10-16 — faithful-model counterexample

The unchanged blanket-dead test passed count=1000 in 24.539s at the new main-based branch; this does not erase the earlier 18/1000 RED. A new deterministic test, TestAcquireDelayedDeadObservationPreservesNewLiveHolder, failed immediately (0.01s): delayed contender completed while live PID 111 still held the lock, with nil acquisition error. PID-specific ProcessAlive marks only original PID 555 dead. Contender 222 reads 555 and pauses its liveness answer; contender 111 reclaims 555 and acquires; resume 222, which renames the current lock based on its stale 555 observation, then acquires concurrently. Real filesystem, no time sleeps, workers canceled/joined on teardown. This is a production stale-observation-to-rename race, not solely an invalid fake. Stop before production edits and revise with TL.

- Guarded implementation passes the repolock package. RED controls confirmed initial-probe cancellation could mutate the lock and busy guards ignored cancellation/budget. Added guarded-probe cancellation/deadline, revived-holder and final-inode/descriptor-release controls. A partial replacement metadata test also failed before its fix: reclaim now returns to ordinary observation/grace rather than failing on an unproven new generation. Production addition is 80 lines, preserving quick-flow scope. Repeated and race-enabled checks are running.

- Verification on the guarded source (9a7e6898f5ad4699f39549e35d8fe1ca72b3870f): both ownership regressions count=1000 PASS 145.904s; go test -race -trimpath ./cmd/sdlc/internal/repolock -count=20 -timeout 5m PASS 30.084s; CLI lock integration selection PASS 14.047s. Full repolock package PASS. Final-inode check preserves a replacement created during the probe and releases the old descriptor guard. Late cancellation/deadline, revived holders and busy-guard budget are asserted. Atlas and lessons updated; next is PR plus recorded full suite before close.

## Revisions

### 2026-10-16 — faithful ordering disproves test-only hypothesis

The specified production-bug control fails under accurate liveness. The pending fix must guard reclamation against the lock generation changing after observation; changing only ProcessAlive would hide a reproducible safety defect. Scope proposal goes to the TL before production changes.

- Faithful delayed-observation RED repeated 20/20 in 1.260s. Corrected original fake to mark only 555 dead, replaced the 20ms sleep with a synchronized start plus exclusive-holder/metadata/release assertions. With that fake fix, original test passes and delayed-observation test still fails (nil acquisition while live 111 holds). Sent proposed inode/flock revalidation scope to TL, receipt cc1a6b5e-3370-407a-a0e6-484f60394873 submitted. Production remains unchanged while TL chooses scope.

- With the accurate fake and no delayed-observation seam, the original concurrent test failed 19/1000 in 62.377s: mismatched/disappearing live-holder metadata and two simultaneous holders. This independently confirms normal-scheduler exposure. Read-only guard design review requires final inode revalidation after liveness, changed-metadata rejection, cancellation before rename, bounded nonblocking contention retries, and descriptor closure on every path. The existing internal/flock helper fails closed outside Unix; no unsafe fallback is proposed.

### 2026-10-16 — TL-approved production scope

TL confirmed the same-mechanism expansion (receipt f16a10f7-e717-475d-a199-61b14f8f1993). Add a thin reclamation IO helper: open the candidate directory, take the existing nonblocking flock, verify inode identity, reread the original holder generation and recheck liveness, then verify inode and cancellation/deadline again before rename. Close the descriptor on every outcome, before retry/sleep. Preserve pure Observe classification and foreign-host/age-only refusal (ARCH-DRY/PURE/ORDER). Busy or changed candidates reuse the existing acquisition budget; the guard creates no durable artifact and dies with its descriptor/process (ARCH-CONSTRAINTS/FUNERAL). The faithful counterexample and 19/1000 normal-scheduler failures above are the reason; a fake-only fix is insufficient. Add cancellation/busy-budget controls, then repeated/race tests and recorded full suite before close.
