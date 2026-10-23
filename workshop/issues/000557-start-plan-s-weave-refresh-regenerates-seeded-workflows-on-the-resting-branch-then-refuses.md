---
id: 000557
status: codecomplete
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours: 1.43
collaboration-mode: collaboration
card_mirror: 'cec4b11f0f71b158c53b3297c78cd3cecde6baa3' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T17:29:49-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 0.77
---

# start-plan's weave refresh regenerates seeded workflows on the resting branch, then refuses

## Problem

## Spec

`sdlc start-plan`'s catch-up bumps a peer repo's ariadne pin and runs `weave refresh`, which regenerates seeded files
(here `.github/workflows/merge-check.yml`: ariadne's SIGPIPE/drain fix in the compiler-provenance step, two
`grep -Eq` lines becoming `grep -E ... >/dev/null` plus a drain comment) **on the slot's resting branch**. start-plan
then refuses, because the resting branch has uncommitted tracked changes beyond the issue's details. Every slot that
refreshes against the new ariadne hits this:
- pair:8 (pair#809, 10-22 13:1x): bump ariadne 2dc46703 → 5785330c; saved the patch, restored the file, re-ran.
- parley.nvim:3 (parley.nvim#409, 10-22): "weave-generated merge-check workflow drift" kept in a named stash.

Resting branches only fast-forward (#284), so the refresh must not leave tracked changes there, and seeded workflow
files are local-gates configuration (#309): their update has to reach main as its own change the operator verifies,
not ride along on an issue branch.

Fix: start-plan's catch-up never leaves regenerated tracked files on a resting branch. When a refresh would change a
tracked seeded file, it restores the file, reports "seeded file X is behind ariadne; file or land the weave refresh
as its own change", and continues (or the refresh runs where it can't dirty the resting branch). Plus a way for the
pending regeneration to reach each consumer's main as one local-gates change.

## Done when

- A start-plan whose catch-up would regenerate a tracked seeded workflow leaves the resting branch clean, proceeds,
  and names the pending seeded-file update (test).
- The pending update can be produced as one reviewable change for the consumer repo (documented, tested).

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only; calibration is stale, so this is provisional.

Extending existing Go/FS/refresh seams, no novel stack/library needed. Per-primitive design uses the thorough-spec ×0.2 discount; implementation is v2.1 ×0.4. Familiarity ×1.0, design buffer +15%. The four implementation units include their colocated/native tests; review and docs are separate.

| Unit | Primitive and source hours | Discounted design | Scaled impl |
|---|---|---:|---:|
| Pure policy/classifier | smaller Go 0.30 / 0.50 | 0.06 | 0.20 |
| FS currentness/report deferral | smaller Go 0.30 / 0.50 | 0.06 | 0.20 |
| One compile-policy wiring refactor | cross-cutting refactor 0.40 / 0.50 | 0.08 | 0.20 |
| Native refresh/start-plan/regeneration proof | smaller Go 0.30 / 0.50 | 0.06 | 0.20 |
| README/help/atlas | docs 0.10 / 0.20 | 0.02 | 0.08 |
| One mandatory close review | milestone review 0.10 / 0.50 | 0.02 | 0.20 |

Design 0.30 ×1.15 + impl 1.08 = 1.425h, rounded to 1.43h. Suite queue wait is not focused work and is recorded separately, never hidden as estimated implementation.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: cross-cutting-refactor design=0.08 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.02 impl=0.20
design-buffer: 0.15
total: 1.43
```

## Plan

Proposed design: [planning seed deferral](../plans/000557-planning-seed-deferral-plan.md). Hub approved the seeded-output scope; entering change-code before implementation.

- [x] Add a planning-only publication policy that defers every Seed / SeedOnce destination and existing protected settings before any output effects.
- [x] Wire start-plan catch-up to that policy; report stale seed paths and the separate consumer regeneration action.
- [x] Prove native catch-up cleanliness and explicit consumer-branch regeneration, and document the operator-verified PR route.

## Log

### 2026-10-22
- 2026-10-22: closed — Only post-close code delta aligns TestSDLCWeaveDependenciesCheckOnly fake/parser and expected argv with the already-reviewed refresh --planning contract; no production behavior/scope change. Original full suite completed 2303 SDLC tests and 76 packages with only the two stale-fixture cases failing; targeted red reproduced both, then corrected dependency context + installed argv + native start-plan controls PASS 14.306s. Fresh recorded suite queued on pushed fix head; pass still required before landing. Fleet merge pause remains active; no new merge or landing will start. (fixed-then-ship); review verdict: SHIP
- 2026-10-22: closed — Native installed-weave/start-plan catch-up creates the issue branch with clean resting workflow/index for Claude and Codex; captured pair#809 replay proves pre-fix dirty seed publication and planning preservation, plus ordinary compile yields the pending workflow-only diff on a separate consumer branch. Targeted SDLC PASS 19.711s; weave policy/state/evidence/errors/interrupted publication/counted growth and compatibility controls PASS 41.402s; tracked Go catalog PASS 1.171s. Hub approved isolated proof, no live stage. Recorded suite lease-1202838612 on pushed code head c9ee0844b3d9251a082eb821db4229d1195ed472 is queued and retained; close review overlaps it under tl-worker, with suite pass required before landing.; review verdict: SHIP

### 2026-10-22 — claim and design investigation

Claimed in ariadne:2 and started planning from fresh main. #531's merge runner completed with a passing authoritative suite (2302 SDLC tests, 76 other packages), PR396 merged, and this slot clean on main-slot2 before the new claim. This ariadne checkout did not reproduce consumer seed drift because its own start-plan succeeded.

Root cause verified: startPlanBranch calls catchUpBeforeWork while on the resting branch; installed weave runs refresh --defer-settings. Refresh fast-forwards captured host/substrate targets and then compilePreparedSettings publishes normal actions. Only .claude/settings.json is deferred. plan.Seed performs an atomic content/mode replacement and is consumer-owned, outside managed identity/displacement recovery. plan.SeedOnce can also adopt a tracked entrypoint. Restoring files after refresh is unsafe on partial failure or a kill; prevent these actions before publication instead (ARCH-ORDER / SECURE).

Read-only bounded code trace identified existing OutputCurrent, DeferredOutput, settings deferral and real-Git refresh fixtures for reuse (ARCH-DRY). Verified both source commits named by the issue with Git: 2dc46703996c42f4b0916515d4572cf439dd2a32 and 5785330cda6bf31d533e4282aecc8f3f9cd632a7. Their actual workflow diff is the two quiet grep compiler-identity checks replaced with full-input grep plus a drain comment; capture that diff in the native regression rather than inventing a similar incident.

Proposal: a new explicit weave refresh --planning policy defers all Seed and SeedOnce actions regardless of tracking, plus existing protected settings; normal compile and explicit refresh retain their current regeneration semantics. Compare before publication, retain DeferredOutput evidence, name each stale destination and direct regeneration onto its own consumer issue branch with weave compile. That separate local-gates PR requires the operator's label; no automatic label, consumer branch, commit or PR creation. Plan explains that this covers the whole seeded-output class, not an unspecified global rollback for other tracked writer families; requesting hub confirmation of that scope. No source edits or implementation tests yet.

### 2026-10-22 — plan gate and estimate

Plan-quality round 1 CLEAN, no findings. Only after that acceptance, derived provisional Method-A v3.1 estimate 1.43h from the shared grammar and local calibration; recording card estimate through its setter before retrying change-code.

### 2026-10-22 — implementation and isolated production proof

Second change-code admitted the full flow: plan-quality CLEAN; estimate-quality INFO (the native proof allowance is ambitious; colocated tests are included in the implementation units, and queue wait is separate). No gate bypass. Implemented invocation-local publication policy, pure Seed/SeedOnce classification and observation/reporting before CheckManaged/every ApplyManaged. Installed SDLC catch-up now uses refresh --planning; settings-only compatibility remains. README, command help and atlas explain the dedicated consumer issue/branch, ordinary weave compile, separate PR and operator-only local-gates-verified label. No actual workflow configuration, live consumer checkout or :0 stage changed.

Captured the actual before/after merge-check bytes in cmd/weave/testdata/planning-refresh. Red-before native regression failed for Claude/Codex because --planning did not exist; its legacy --defer-settings controls reproduced the dirty tracked workflow. After implementation, those controls still reproduce the original defect while planning mode fast-forwards captured Git targets, preserves the clean resting checkout/index, names the deferred workflow, and ordinary compile produces exactly that pending workflow diff on a separate temporary consumer branch without committing or changing HEAD.

Targeted proof returned PASS: native SDLC installed-weave/start-plan tests plus existing catch-up/argv controls (19.711s); weave mixed-action property, destination/evidence preservation, N=8/16 counted observations, failure/cancellation at real managed publication, captured refresh, settings-deferral and ordinary refresh controls (41.402s). Real private Git repositories and OSFS exercise both Claude and Codex. Source-unreadable no-op, missing/current/edited/mode/symlink and prior-owned recovery states preserve ordinary seed semantics and zero protected mutation attempts. Observation/writer errors propagate without publication. The ownership oracle normalizes absent versus empty inventories but compares every retained identity; no seed ownership is introduced.

No speculative generic tracked-writer issue filed: Ship, retirement and .gitignore remain the approved known follow-up only if observed. This fix is a publication-policy ordering invariant, defended by production-path fixtures rather than a lexical lint rule. Next: tracked-file catalog, commit/push, recorded suite, mandatory close review and landing. No recorded check or merge runner is active at this checkpoint.

### 2026-10-22 — published, recorded suite retained

Tracked Go catalog PASS (1.171s), staged diff whitespace check clean. PR405 opened; GitHub local-gates-guard and its verdict SUCCESS. Recorded suite lease-1202838612 retained on pushed code head c9ee0844b3d9251a082eb821db4229d1195ed472, initially FIFO position 15, not started yet. No duplicate runner or cancellation. Hub authenticated approval fdcf2149-b38f-4d13-8b20-cdada682c4a8 accepts the isolated native installed-weave/start-plan proof and pair#809 replay; no live :0 stage. Mandatory close review may overlap this queued suite under tl-worker; a recorded pass covering final code still precedes landing. If review changes code, run a new recorded suite on that pushed code head. Next: finish close review, read this retained suite's result, then enter sdlc merge --yes on SHIP. No continuation or restart while the check is retained.

### 2026-10-22 — fleet merge pause and full-suite fixture correction

Authenticated hub notice de214331-e3be-4d43-8717-158a900a78ee pauses all new merges/landings ahead of the planned 22:00–23:00 Couch shutdown and Go upgrade. No #557 merge runner was started. Keep the issue unmerged until the hub authorizes new landings; take no new item. Preserve this branch and PR405.

The first retained suite (attempt 156235e802caa27cb038ef76c568b30b, lease-1202838612, code head c9ee0844b3d9251a082eb821db4229d1195ed472) completed all 2303 SDLC tests and 76 other packages: seven shards and other packages passed; only TestSDLCWeaveDependenciesCheckOnly/inherited= and /inherited=0 failed. Raw log: Git common dir sdlc/test-stats/failed/156235e802caa27cb038ef76c568b30b.log. This is a reproducible branch fixture defect, not infrastructure or a registered flake. The test's fake weave still required --defer-settings and refused the intentional --planning argument before checking dependency policy. Targeted red reproduced both failures (0.815s).

Updated exactly that fake contract and its expected argv to --planning, keeping checks-only environment, unchanged parent environment, propagation sequence and stdout/stderr assertions intact. Swept all remaining --defer-settings references: they are intentional compatibility controls/help; no other stale SDLC adapter expectation remains. Targeted dependency context + installed argv + native start-plan tests PASS (14.306s). No production behavior or approved scope changed after SHIP. Next: push this fixture correction, queue a fresh full suite on the pushed head, and narrow re-close; retain that check without restart. After it passes, commit/push a final handoff and remain idle under the merge pause. After authenticated permission to resume new landings, rebuild tools following the Go upgrade, inspect state/PR/check evidence, then sdlc merge --yes. Never infer that permission from Couch merely being available.

### 2026-10-22 19:52 — durable freeze handoff; check still retained

Authenticated hub follow-up 0949d13f-9cbf-4fc1-be07-48bee4643338 requires no merge tonight, no new item, and committed/pushed state before 21:45. This supersedes any earlier instruction to enter merge immediately on SHIP. No new capacity error was observed during this resumed work; report one if it repeats.

Verified sdlc state: original issue branch retained, #557 codecomplete; working tree clean before this Log update. Fixed-then-ship review round 2 SHIP with zero findings; close evidence 089a7041d7013b627b808639d46dad949cb74e1a, mirrored by 20c851ac9e4a6df6c61af8fd8e8d60bde742a42e. PR405 remains OPEN, merge-check and local-gates-guard (including verdict) SUCCESS on that PR head. No #557 merge/landing runner exists.

Retained recorded suite: lease-2459060446, attempt b4fa6c947bd70ef4d22947d0e2252993, exact pushed fix head 63a9053009e44fb1b1fcffb2dfebcfd9ab88e7b4. Owner PID95656 incarnation 1792722500.451295, initiating checkout ariadne:2. Detached worktree /Volumes/TestScratch/local-checks/sdlc-local-4118944290. Latest queue observation position 2; no run.log yet under Git common dir sdlc/test-stats/inflight/b4fa6c947bd70ef4d22947d0e2252993. No duplicate runner, cancellation or blind retry. Preserve this waiter; read the completed result before assuming it passed. Do not restart/checkpoint a continuation while it is queued/running.

Exact next step: finish reading this retained suite; on pass append its result here, commit/push and remain idle for the night. On failure inspect every failure, preserving its raw evidence; no flake registration or rerun without classification. Before any restart verify the run/result on its exact SHA. After the planned Go upgrade, rebuild this slot's tools as directed and re-read tl-worker help, sdlc state, PR405 and queue/landing evidence. New landing requires fresh authenticated hub authorization after tonight's freeze; only then run sdlc merge --yes (never duplicate a surviving merge runner). A Couch-available notice alone is not authorization. No new issue is accepted while waiting.

### 2026-10-22 19:57 — queued suite cancelled by hub; stop for tonight

Authenticated hub instruction 8fa9c8ee-eca8-4029-adc0-b2888268b77e explicitly requested cancellation of #557's queued suite to let #554 (echo critical path) proceed, and requeue first thing tomorrow. This supersedes the earlier keep-waiting instruction. Ran sdlc local-check cancel --run lease-2459060446; it returned cancellation success on 63a90530/PID95656. Retained waiter exited with suite lock interrupted: context canceled, before any suite execution. The exact lease is gone. No code-failure result or passing evidence was created by that cancellation. No PID kill, sibling cancellation, duplicate runner, merge or landing was performed. Sent the requested cancellation report to the hub (receipt ed00356c-0234-4733-9ce4-6cea821eec49).

Stopped with implementation and fixture correction reviewed SHIP, PR405 OPEN, and no owned background suite/review/merge left. Full-suite verification on the corrected code is OUTSTANDING: the only completed full run was the earlier stale-fixture failure; targeted correction/native controls passed, and re-close was SHIP. Do not treat this as a suite pass or land tonight.

Exact next step for 2026-10-23: after resuming from this Log, read tl-worker --help and verify queue/landing/PR state and any retained checks; rebuild this slot's binaries after the operator's planned Go upgrade. Requeue sdlc local-check run --name suite once on the current pushed PR head, retain its exact new run ID, and read its completed result. There is no surviving #557 runner to resume or duplicate. Classify any failure before retrying. Only after a recorded pass and fresh authenticated hub authorization to lift the landing freeze, run sdlc merge --yes. A Couch-available notice alone grants neither new landing authority nor a suite pass. Take no new issue tonight; commit/push this Log and remain idle.

## Revisions

### 2026-10-22 — proposed prevention at publication boundary

Propose deferring seed / seed-once publication before effects, instead of post-refresh restoration. Preserve ordinary regeneration and route pending updates through a dedicated consumer issue and operator-verified PR. The existing Done when remains the acceptance contract; await hub review of the seeded-output scope.

### 2026-10-22 — scope approved

Hub approval received and authenticated (submitted delivery b2a58869-feeb-45f0-b41c-973fb0167e90). Suppress every Seed/SeedOnce before planning publication, retain protected-settings deferral, and use ordinary weave compile on a dedicated consumer issue branch for the operator-verified workflow PR. Wider Ship / retirement / .gitignore protection is a known follow-up, to file only on observed dirtying evidence. No stash, live consumer mutation, automatic PR or self-label. Proceeding to change-code with the full-flow proposal; no source edits yet.


### 2026-10-22 — implementation checklist complete

The issue checklist now tracks delivered implementation, proof and docs. Suite/review/landing execution remains outstanding in the Log, avoiding a checklist item that would require landing before the close gate can run. No scope or acceptance change.
