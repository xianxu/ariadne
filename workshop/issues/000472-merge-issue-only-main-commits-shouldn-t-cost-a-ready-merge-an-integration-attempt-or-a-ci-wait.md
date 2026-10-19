---
id: 000472
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '6c73f619e0a34cc6b8119d6b2d46c3b489e7bc8b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T15:37:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "6f5a5a98", done: "c3b74cb6"}
---

# merge: issue-only main commits shouldn't cost a ready merge an integration attempt or a CI wait

## Problem

2026-10-19, pair#705 (pair:7): its integrated suite passed and CI passed, then two **issue-only** commits reached main (a
`sdlc issue new` publish and an `abandon` archive, touching only `workshop/`). Each forced `sdlc merge` to integrate main
again and wait for fresh CI, using up attempt 3 of 3. On a busy repo, issue filing alone can starve a ready merge, and the
hub had to ask every slot to stop publishing.

## Spec

When main has moved since the PR head's last integration, and every new main commit touches only paths that cannot affect
the build or tests (the issue-tracker details paths: `workshop/issues/`, `workshop/history/`, `workshop/plans/`, …, from one
declared list), `sdlc merge`:
- does not count it as a new integration attempt;
- does not rerun the recorded suite (already true when code didn't change; confirm);
- integrates and merges without waiting for a fresh CI run when the repo's CI ignores those paths, or, if the repo requires
  an up-to-date branch with CI, waits for CI but doesn't spend an attempt.

Where possible, avoid integrating at all: merge the PR as is when GitHub allows it (branch protection not requiring
up-to-date), since issue-only main commits can't conflict with code.

## Done when

- Fixture: PR ready; main gains 3 issue-only commits during the wait; merge lands on its first attempt, with no suite rerun.
- Fixture: a main commit touching code still counts as an integration attempt (unchanged).

## Plan

- [x] Implement and verify the shared issue-record classifier and integration-state accounting.
- [x] Add exact GitHub CLEAN readiness and wire conservative landing/recovery behavior.
- [x] Document, run recorded verification, close on SHIP and land.

Durable design: [issue-only main drift plan](../plans/000472-issue-only-main-drift-plan.md).

## Log

### 2026-10-19

### 2026-10-19 — approved design boundary

Claimed and started planning. Hub approved one narrow issue-record path policy,
unchanged-head landing only on fresh exact PR/head/base CLEAN evidence, and
normal integration/CI fallback with no issue-only attempt charge. No ops/ policy
now, workflow edit, status copying or publication freeze. Existing docs-only
ancestor suite reuse remains; it is broader than the direct-merge policy.

Read-only exploration found no existing protection-policy seam. Extend the
validated landingPR evidence instead of inventing a protection/ruleset reader.
The existing stateful GitHub fake, per-head CI fake and real Git integration
fixtures cover the required behavior. Full-flow durable plan written before code.

## Revisions

### 2026-10-19 15:43 PDT — exact-head CI boundary

Use the Spec's avoid-integration option only with fresh CLEAN evidence. Workflow
paths-ignore alone cannot justify a new integrated head inheriting server checks:
PR path filtering considers the full PR diff. Strict/unknown fallback integrates,
waits for exact-head CI and preserves the existing local-suite reuse rule, while
issue-only movement never consumes the code integration budget.

### 2026-10-19 — live readiness correction

Read-only API conformance before code found that landingPR.BaseOID (GitHub
baseRefOid) is historical for an open PR; GraphQL baseRef.target.oid matches
current main. The durable plan now keeps listing/recovery evidence unchanged and
adds a dedicated typed readiness query through the GitHub adapter. Only the live
base reference can bind CLEAN to the final fetched main. Reported to the hub.

### 2026-10-19 — resumed implementation checkpoint

After Couch/filesystem recovery, the original stalled commands were gone and
sdlc state confirmed the clean issue branch. Repeated change-code succeeded,
inferred quick from the current design size, and authorized implementation;
close owns any upgrade when the final code exceeds the shell. No estimate or
plan-quality pass is claimed. All SDLC gates continue through central sdlc.

Implemented the single narrow Markdown-record policy and per-commit Git
observation (one history call, merge-parent and rename-source coverage). Added
state-owned integration charges, metadata-only deferral and typed GitHub
readiness with live base target. The bounded API runner keeps old callers'
behavior and limits only this fixed scalar readiness query. Final direct merge
rechecks existing issue/instance/publish guards against current main.

Focused classifier/state/readiness tests passed. Real-Git/stateful-GitHub tests
passed in 54.804s: three issue-only publications keep the tested ready head;
strict/unknown fallback survives four publications without spending another
code integration attempt or rerunning the suite. Broader existing integration
and recovery controls are running. Early end-to-end fixtures were corrected to
valid issue frontmatter and the configured upstream/main reference.


### 2026-10-19 — targeted verification complete

The broad stable-checkout landing/state/readiness/recovery selection passed
(157.115s). Five real-Git controls passed (38.936s): late code and mixed movement
both force a second tested head; readiness head mismatch, same-record conflict
and a late duplicate ID refuse without merging. The duplicate case enables the
actual instance/section and duplicate gates, stubbing only the external vocabulary
executable as existing instance fixtures do.

The explicit strict-fallback pending-CI control and metadata reset accounting
passed (5.916s): old-head green checks do not authorize the integrated head;
its pending checks wait, while the suite runs once. Merge-parent histories and
interrupted integration observations passed (0.763s): recovery charges the actual
merge's second parent, not later main movement.

Read-only real adapter conformance on PR 325 passed (0.962s), returning its live
base target and UNSTABLE status. A deliberate mutation disabling direct readiness
made the ready-head test fail specifically because the tested head changed;
restored code and existing CI/refusal controls passed (15.563s). No workflow,
status-copying, shared staging or real PR mutation was used for this smoke.

One earlier wider run passed its assertions but failed TestMain's real-checkout
hermeticity guard because I committed while it ran. This was a deterministic
orchestration error, not a flake. The stable 157.115s run replaces that invalid
evidence, and the lesson now prohibits concurrent commits/new dirty paths during
these tests. Adapter, state, classifier, landing help, README and atlas are ready
for the recorded suite and mandatory close review.


### 2026-10-19 — delivery gates running

Implementation and documentation are ready for close. PR 332 is open on pushed
4a686788; the recorded suite is queued under normal FIFO (both machine slots,
position 6 at admission). Close review may overlap that queue per tl-worker;
a recorded suite pass on the final code head remains required before landing.
The Plan delivery row records entry into these binary-owned gates, not a claim
that the queued suite or landing has completed. The remaining delivery results
will be recorded by close and merge.
