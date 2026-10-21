---
id: 000530
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1aa9c043cf777ec7c136ae941c4bd2451092b58d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T14:22:07-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# local-check: a suite that fails before any test on ENOSPC (full Go cache) is could-not-run, not an unattributable failure or main-red

## Problem

## Spec

10-21 ~12:40 the shared Go build cache `/Volumes/AgentScratch/go-build` (182 GB) filled its volume (186 GB, 100 MiB
free). Every recorded local-check suite on the machine then failed before running a single test, writing the cache
(`no space left on device`, typically in pair's runtimebundle generation): pair#768 (train 238 baseline), #771, #774
(integration + diagnostic main), #775, ariadne#526, #528, among others. sdlc classified these as **unattributable
failures** on the exact head, not as the narrow "could not run" (infra) class it uses for TMPDIR problems. Slots now
expect the same-head unexplained-failure guard to resist a plain rerun once space is restored, and train 238's
baseline turned a main-red verdict out of it.

Fix:
1. Classify a suite that fails **before any test ran** with a recognised infrastructure error (ENOSPC/EDQUOT on the
   Go cache, TMPDIR or the build output; also missing toolchain) as **could-not-run**, same as the TMPDIR class: no
   failure evidence, no main-red verdict, no same-head guard; the message names the full volume and path.
2. A main-baseline diagnostic that could not run must not mark main red or hold/split a train.
3. A supported recovery for runs already recorded as unattributable under this signature (one explicit
   `local-check` reclassification by the owner or :0, logged), never a flake registration.

### Proposed design (awaiting hub approval)

Extend the existing strict `suiteEnvironmentLog`/`suiteEnvironmentError` path,
not the flake classifier. A complete bounded recognized startup log produces a
typed cause/resource/path; any test activity, unrelated diagnostic, truncation,
timeout or unknown evidence remains blocking. New infrastructure failures retain
their diagnostic without publishing code failure evidence. Direct execution,
affected, known-flake rerun, main diagnosis and train consumers share this path.

For historical runs, propose `sdlc local-check reclassify --attempt ID --reason
TEXT`: one exact attempt, immutable log/hash proof, same startup classifier,
originating recorded slot or verified repository :0. No force flag, automatic
bulk sweep, flake registration, status rewrite or old-pass republication.
Preserve original record bytes and raw log identity in an immutable audit before
atomically publishing a nonpassing v3 could-not-run disposition. Existing v1/v2
readers remain supported; older binaries encountering v3 refuse safely. The
shared evidence projection prevents legacy resurrection, false flake proof and
saved-pass promotion. Interrupted correction stays blocking until the same
command reconciles it. Audit pruning follows the attempt's existing lifecycle.

Alternatives rejected: silently interpreting old logs changes evidence without
an audit; deleting records or adding an ignored v2 flag leaves old readers or
legacy logs with conflicting authority. The durable plan contains the full
state/action/recovery table, consumer inventory and counted scan/read checks.

## Done when

- A test drives a fake ENOSPC in the Go cache before any test output and gets could-not-run, rerun allowed,
  no main-red; a real test failure after tests start is still a failure.
- The reclassification path exists and is logged; tests cover refusal when tests had already run.
- Recorded live startup output replays in the normal suite for candidate and
  diagnostic main; no false main-red, train hold/split/blame or passing status.
- Only an exact eligible attempt's originating slot or verified repository :0
  can reclassify; mixed/unknown/incomplete output and changed/missing log proof
  refuse. Existing real failures on the head remain blocking.
- Original record/log evidence and actor/time/reason audit survive interruption
  and duplicate recovery. No legacy resurrection, flake proof or old-pass
  promotion; old-reader behavior and unknown-field preservation are exercised.

## Plan

Full flow, one atomic close boundary, plan approval required before code.
Durable plan: [suite startup evidence](../plans/000530-suite-startup-evidence-plan.md).

- [ ] Replay the recorded startup failure; extend the strict live execution path
  with positive/negative controls and candidate/main/train coverage.
- [ ] Implement exact-attempt, owner/:0, audited recovery and versioned evidence
  projection; exercise interruptions, consumers, old readers and retention.
- [ ] Document, verify the full recorded suite, close review and land through
  sdlc; report the outcome to the hub.

## Log

### 2026-10-21

- Claimed after #508 landed. TL explicitly requested plan-first gate-evidence
  work. No production changes or tests have been written for #530.
- Re-read the actual #508 retained ENOSPC log: Go list failed while writing
  /Volumes/AgentScratch/go-build, followed by the complete Python go-list
  traceback and make exit. The existing classifier accepts only a work-dir
  creation diagnostic and an older go-test-compile traceback, so this output
  falls through to unattributable failure. Baseline consumers then mistake
  an unparsed failed run for main-red. This establishes the live root cause.
- Read-only delegated mapping confirmed that same-head suite/affected guards,
  saved pass reuse, flake/stats projections and legacy per-head logs all consume
  the persisted attempt. A v2 flag would be ignored by older readers. The #524
  future-version identity envelope and queue reconciliation audit pattern are
  the existing seams to reuse (ARCH-DRY, ARCH-EVOLVE, ARCH-ORDER).
- Drafted the durable plan with explicit state/recovery and compatibility
  coverage. No estimate before plan-quality, no change-code before hub approval.
- TL approved the committed design, command/ownership, audited nonpassing v3
  disposition, genuine rerun and strict refusal controls. Explicit added proof:
  a real #524-era reader must scope the unsupported v3 record to its subject,
  rather than poisoning unrelated evidence; fleet shared sdlc includes #524.
  Proceeding to the full-flow change-code plan gate before implementation.
- Plan gate round 1 raised PQ-1 (Important): name risky functions and replace
  test-case inventories with adversarial strategy/mechanical-guard contracts.
  Added those function contracts and compressed the task inventories; no design
  mechanism or scope changed. Re-running plan-quality before implementation.

## Revisions

- 2026-10-21 — Added the proposed design and explicit acceptance controls after
  tracing the outage through existing consumers. Original requirements remain
  above; ownership, atomic audit and compatibility details await hub approval.
