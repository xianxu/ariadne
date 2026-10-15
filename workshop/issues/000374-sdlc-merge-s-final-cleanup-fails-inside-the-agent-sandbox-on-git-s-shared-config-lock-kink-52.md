---
id: 000374
status: codecomplete
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-15
estimate_hours: 1.94
card_mirror: 'daab3d63e8048d4ac077ddf8ee69c62bef3a99e1' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T23:06:07-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 2.28
---

# sdlc merge's final cleanup fails inside the agent sandbox on git's shared config lock (kink 52)

## Problem

`sdlc merge`'s final cleanup (after landing) fails inside the agent's sandbox on git's shared config lock (`.git/config.lock`), seen 3+ times (TL kink 52); slots rerun the printed recovery command unsandboxed. Retro item (e); not covered by ariadne#370, which landed first.

## Spec

Find which cleanup step writes git's shared config (branch deletion with upstream config, worktree prune, `git config` edits) and either avoid the shared-config write in the sandboxed path (e.g. no branch config to remove, or defer it to the next resting-branch command), or make the step idempotent and resumable so the next sdlc run completes it without the operator. No sandbox widening.

## Done when

- The native protected-config full-merge regression passes without manual rerun, and interrupted cleanup converges on the next run (tests). This is sufficient for local close, per TL approval.
- Production default-agent sandbox cleanup: **verified by the TL at the next Claude-slot landing** after #374 merges (ariadne:2 #383 or ariadne:4 #367). The TL owns this post-merge check and will log its result; no separate scratch issue is required.

## Plan

- [x] Keep ref deletion and identity/occupancy/CAS guards mandatory; make only obsolete branch-config removal a non-blocking deferred cleanup, with the affected config path and retry owner reported.
- [x] Reuse the existing done-card merge/reconcile sweep to retry configuration cleanup for absent, unoccupied landed issue branches; preserve live/recreated refs and unrelated config (ARCH-DRY, ARCH-FUNERAL).
- [x] Add real-Git regressions for repeated config-lock denial, interruption before/after ref deletion, automatic later cleanup when writable, and live/recreated branch protection. Exercise the full merge caller with the existing stateful GitHub fake, plus a native denied-config sandbox check.
- [x] Update help/recovery/atlas and run focused verification for deferred metadata cleanup.
- [x] Run recorded full-suite verification on the pushed integrated code head.

After acceptance, use SDLC close and land on SHIP; those workflow transitions are recorded by their own evidence.

## Log

### 2026-10-14

### 2026-10-14 — root cause and proposed design

Claimed by ariadne:5 and started from current main. The durable path explicitly runs git config --local --remove-section branch.<issue> after CAS update-ref deletion (cmd/sdlc/landing.go:321-348). The command acquires the Git common directory's config.lock, even from a linked worktree. A native macOS sandbox scratch experiment denied writes only to shared config/config.lock: update-ref exited 0; remove-section exited 255 with could-not-lock-config; the branch was absent and its upstream config persisted. Current recovery repeats the same permanently denied metadata write. Existing interruption tests cover one transient failure, not permanent denial. Legacy merge uses git branch -D, whose config cleanup is already warning-only; its caller returns success rather than this fatal result.

Recommended design: distinguish mandatory branch retirement from ancillary obsolete-upstream metadata. A failed metadata write warns without invalidating a confirmed landing. The existing done-card sweep on merge/reconcile retries only obsolete configuration belonging to done issues with no local ref or checkout. Cards plus existing config entries supply retry discovery, so no new receipts, queue, or sandbox privileges are needed. A persistent sandbox denial retains only existing metadata (typically the remote/merge pair per landed issue); the next writable merge/reconcile removes it automatically. Ref CAS, occupancy, identity checks and operational failures in critical cleanup remain blocking. The thin Git IO helper stays behind the existing runner (ARCH-PURE); tests use real scratch Git and the stateful GitHub fake (ARCH-MOCK).

Alternatives considered: making all final cleanup fatal preserves the present retry loop under a persistent sandbox restriction; permanently dropping configuration cleanup accumulates residue with no owner. Writing the protected shared config by another path is neither safe nor a permitted sandbox workaround. The proposed deferred sweep supplies an explicit lifecycle for the residue (ARCH-FUNERAL).

The original live acceptance asks for a scratch issue in the default agent sandbox. The full merge regression will cover a scratch issue with real Git and the production merge path plus stateful fake GitHub, including a real denied-config sandbox. Ask the hub how to arrange the additional default-Claude sandbox live landing without overlapping another slot's active issue; do not silently replace that acceptance with a simulation.

### 2026-10-15 — validated implementation checkpoint
- 2026-10-15: closed — Recorded full suite passed on 046de3e2: 12 shards/1376 SDLC tests plus 59 packages, zero failures, 491.1s. Focused real-Git, recovery and guard regressions and mutation checks passed. BR-1 host counter-measurement on reviewed 7144ce9e passed native sandbox full merge with denial control, retired ref, unchanged config bytes, exit0 in 4.370s; complete output committed in issue Log. Supplemental default-Claude scratch fixture passed. TL owns production conformance at next Claude-slot landing per acceptance revision. README/help/atlas/recovery docs updated.; review verdict: SHIP

Full plan-quality cleared PQ-1/PQ-2; full-flow change-code passed with derived estimate 1.94h. Estimate-quality INFO notes the verification/shipping allowance is optimistic; keep that whole-issue derivation rather than back-fitting it to elapsed time. Producer and compatibility implementations are complete, with no issue upstream consumer requiring tracking; main/rest config remains required. Existing interruption tests now distinguish critical ref retirement from nonfatal metadata cleanup. Shared optional Git-output handling is reused by landing and cleanup.

Focused restored-head verification passed: the directly affected 12-test selection (new producer/cleanup cases, interruption/dirty-rest recovery, CAS and pushed freshness) completed in 161.642s. An overly broad serial selection timed out at 15m while TestLandingIntegratesAMovedMain had run only 9s, without an assertion failure; replace that selection with the targeted result plus the upcoming sharded full suite, not a claimed pass. Native macOS conformance (ARIADNE_NATIVE_SANDBOX=1 go test ./cmd/sdlc -run ^TestLandingNativeConfigSandbox$ -count=1 -v -timeout 4m) ran the full merge command path in a child process under a custom profile denying shared config/config.lock writes, with real scratch Git and stateful local GitHub fake. Denial control failed as intended, merge succeeded/ref retired, config bytes stayed unchanged (13.038s). This custom-profile check is not yet the original default-agent live acceptance.

Mutation checks /tmp/ariadne374-mutations.log confirmed failures when restoring fatal metadata cleanup, removing ref-presence guard, and removing occupancy guard; sources were restored. An initial test-edit anchor failed and yielded a no-tests-to-run mutation attempt, which was rejected rather than counted; corrected harness reran all three successfully. Final restored native/probe/upstream focused run passed in 24.094s, including failed-observation controls and late manual-ref preservation. No whole suite is claimed yet; original live acceptance remains pending the hub response. Documentation and the operational ownership boundary are updated in README, help, atlas and recovery catalog.

### 2026-10-15 — default Claude sandbox acceptance

Reproduced the actual default-agent boundary using a temporary Go overlay of the native conformance harness: python3 /tmp/ariadne374-default-sandbox.py. The overlay replaces only the parent launcher (sandbox-exec → headless Claude with this repository's unchanged .claude/settings.json); committed production and child merge code are unchanged. Claude started in the scratch linked worktree, permissionMode default, and issued exactly one Bash tool call with no dangerouslyDisableSandbox field. The child test starts outside Git for its hermeticity snapshot, switches to the prepared scratch issue, verifies a shared-config write is denied, and executes the merge command path to success/ref retirement. Parent verifies protected config bytes are unchanged. Real Git is used throughout, with the existing stateful local GitHub fake; fixture-only no-judge/no-validate flags are explicit. This qualifies the local cleanup boundary rather than claiming a live GitHub service conformance run or bypassing the production landing gates.

The first default-profile run refused fake-server unpack into the host's unrelated /var/folders temporary tree. Retried with the scratch fixture in Claude's existing writable TMPDIR, /Volumes/AgentScratch/tmp/claude-501, without modifying sandbox settings. The successful default-profile child completed in 7.34s; complete host harness passed in 18.359s. Saved stream /tmp/ariadne374-default-sandbox-stream.jsonl SHA-256 74793f8ae2bc669416c81ce17c32a169ad60c4fa3c76f9276feff3127662abc3; command log /tmp/ariadne374-default-sandbox.log SHA-256 10842d0f55514c6a20ae2faadcaca7c85820422d83adf6a40fa22bc5aba0f0d2. Directly inspected the stream's system initialization and Bash input; this conclusion does not rely only on the model's summary. Sent these qualifications to hub, superseding the earlier unresolved environment question. No separate live GitHub issue was created.

Recorded full suite is queued on pushed integrated code head 046de3e2eb19f0072f28e5055fbef9b7bf42337d (session 84920), log /Users/xianxu/workspace/ariadne/.git/sdlc/local-checks/046de3e2eb19f0072f28e5055fbef9b7bf42337d/suite.log, PR https://github.com/xianxu/ariadne/pull/252. Keep this existing run. Only documentation changes after its pinned code head.

### 2026-10-15 — recorded full suite passed

sdlc local-check run --name suite passed on 046de3e2eb19f0072f28e5055fbef9b7bf42337d: all 12 shards over 1,376 cmd/sdlc tests plus 59 other packages, zero failures, wall 491.1s (8m11s). Recorded as sdlc-local/suite; log SHA-256 prefix 29a64b61bad4 at the path above. Weighted queue wait was 2,671s. Only issue/plan documentation changed after that code head. Local acceptance is complete under the TL's explicit revision; production Claude-slot conformance remains TL-owned after landing.


### Close review BR-1: host counter-measurement

The review found no correctness defect, but its nested sandbox refused sandbox-exec with exit 71 before the native child started. Re-ran on the host at pinned HEAD `7144ce9e9da2794944597c720cea49d5e69b68ef` (no source changes):

```text
$ env ARIADNE_NATIVE_SANDBOX=1 go test ./cmd/sdlc -run '^TestLandingNativeConfigSandbox$' -count=1 -v -timeout 4m
=== RUN   TestLandingNativeConfigSandbox
    branchconfig_native_test.go:88: === RUN   TestLandingNativeConfigSandbox
            branchconfig_native_test.go:49: native config-write denial confirmed; full merge command completed and issue ref retired
        --- PASS: TestLandingNativeConfigSandbox (3.19s)
        PASS
--- PASS: TestLandingNativeConfigSandbox (4.03s)
PASS
ok  github.com/xianxu/ariadne/cmd/sdlc  4.370s
```

Exit 0 also confirms the parent compared the protected config bytes before/after and found them unchanged (lines 83–86). This supplies the exact host evidence requested by BR-1; it does not assert the review sandbox can nest another sandbox. TL-owned production conformance remains the next Claude-slot landing.


## Revisions

### 2026-10-14 — prevent unnecessary metadata at publication

TL approved plan d5751c25 with a producer-dependency check. Inventory found no hard dependency on issue upstream config: durable publication/integration use explicit remote/refspecs; legacy merge checks origin/<branch>; integrateBranchPushed optionally reads upstream but falls back to the issue's remote-tracking ref. Bare pushes/pulls requiring tracking execute on main/rest, whose tracking remains unchanged. Therefore remove --set-upstream from both issue publication paths, explicitly name refspecs, and prevent inherited tracking in issue branch creation. Remove obsolete PR warnings and upstream promises from help/output; keep the existing publication lease. Existing/manual upstream metadata still needs the approved deferred-cleanup compatibility path, so this producer correction supplements the recorded cleanup plan rather than leaving older branches broken. Add production publication/creation tests asserting no new issue config and unchanged resting config under Git auto-setup settings, alongside the denial/recovery regressions.

### 2026-10-14 — implementation checkpoint

Conditional approval cd6e0add verified. Producer tests observed unwanted upstream config on baseline, then passed after removing -u and preventing inheritance. Compatibility tests observed the fatal real config.lock failure and missing retry on baseline; combined focused tests now pass (31.218s). Concurrent editing during the initial RED triggered the real-repo hermeticity guard in addition to the intended failures; serialized later edits and tests, whose GREEN completed cleanly. The scope now has a durable implementation plan at workshop/plans/000374-sandbox-merge-cleanup-plan.md because producer, compatibility, tests and documentation can exceed the quick shell's diff limit. Existing interruption tests and final acceptance remain pending.

The first durable-plan write failed because workshop/plans did not exist on this fresh branch. Created it and wrote the plan; canceled the already-started full-flow review before accepting any result so the next dispatch reads a stable, committed plan. No tests or review were claimed from that interrupted dispatch.

### 2026-10-14 — plan review refinements

Full-flow plan-quality round 1 requested a precise concurrency guarantee (PQ-1) and function-specific test strategies (PQ-2). The revised durable plan explicitly limits metadata-preservation guarantees to cooperating SDLC commands under the shared repo lock; a manual Git recreation after the last probe may lose ancillary tracking config but never its ref/commits. This supersedes the earlier unqualified recreated-ref language. The plan now assigns an adversarial input/event class and mechanical regression guard to each risky integration function. Live acceptance question sent to hub as receipt 7ebef1e8-3e97-4a82-909f-3ad2fa2e16b3; original acceptance remains pending.

### 2026-10-15 — TL acceptance revision

TL receipt 8ee3e12f-7646-4b3d-9b23-fdd4ce34efae explicitly accepts the native protected-config full-merge regression as sufficient to close. The production live default-agent check moves to the next Claude-slot landing after this fix merges, owned and logged by the TL (#383 in ariadne:2 or #367 in ariadne:4). Updated current Done when to match that authorized acceptance, retaining the distinction between completed local verification and pending production evidence. Earlier custom/default-profile scratch fixture checks remain supplemental evidence; they do not claim that later production landing already happened. No separate GitHub scratch issue was created.

### 2026-10-15 — verification checkbox precision

Separated the last Plan checkbox into completed docs/focused verification and pending recorded suite; close/landing remain subsequent SDLC transitions rather than a checkbox that could be prematurely ticked before they happen. Scope and acceptance are unchanged.

## Estimate

*Produced via brain/data/life/42shots/velocity/estimate-logic-v3.1.md against baseline-v3.1.md. Method A only.* Calibration is marked stale by sdlc estimate-source, so these coefficients are provisional. Same Go/Git stack; existing Git runners and stateful GitHub fake avoid a novel-stack/library search. Familiarity 1.0. Issue authoring uses table design 0.5 / implementation 0.3; producer refactor and small cleanup module each use implementation 0.5, docs 0.2, review 0.5, Git sandbox discovery 0.6. Implementation values below apply v3.1's 40% scale. Settled plan/pattern design discounts: refactor 0.5 x0.2=0.10, module 0.3 x0.2=0.06, docs 0.1 x0.2=0.02; authoring/review remain 0.5/0.1. Thorough-plan buffer 15%. Total 0.78 x1.15 +1.04 =1.937h, rounded 1.94h.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.50 impl=0.12
item: cross-cutting-refactor design=0.10 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.10 impl=0.20
item: real-api-discovery design=0.00 impl=0.24
design-buffer: 0.15
total: 1.94
```
