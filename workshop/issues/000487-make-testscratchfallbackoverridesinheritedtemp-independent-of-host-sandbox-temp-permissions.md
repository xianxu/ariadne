---
id: 000487
status: codecomplete
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-20
estimate_hours:
severity: high
collaboration-mode: collaboration
card_mirror: 'a2a69b024ea3cd8a9a4d79bc642ae35eacf28259' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T00:23:30-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c31d9875", done: "136dc89f"}
actual_hours: 0.30
---

# Make TestScratchFallbackOverridesInheritedTemp independent of host sandbox temp permissions

## Problem

Follow-up to ariadne#477. ariadne:5 reports that #467 PR325's integrated recorded suite attempt c3ec77ef failed exactly cmd/sdlc.TestScratchFallbackOverridesInheritedTemp (suiteenvironment_test.go:100). The real normal-temp probes under /var/folders/.../T and /tmp both returned operation not permitted inside the agent sandbox. A main-baseline diagnostic is queued; its result is not yet known. #476's unsandboxed recorded full suite passed, so this is an environment-dependent test assumption, not evidence of a #467 weave defect.

The test sets inherited TMPDIR and GOTMPDIR below a regular file, then calls the production runLocalCheck path. That deliberately invalidates its fixture temp root too. prepareScratch subsequently calls scratchdisk.NormalTempDir, which searches the host user-temp directory and /tmp. The child-write assertions therefore depend on host sandbox permissions rather than a controlled fallback root.

## Spec

Make the inherited-temp fallback regression hermetic: prove replacement of unusable inherited TMPDIR and GOTMPDIR using writable private fixture roots and the existing real child command seam. Model denied candidate roots in tests, without requiring the live sandbox to allow writes outside its test root. Keep production candidate selection and write-access verification strict; when every candidate is genuinely unavailable, retain the environment failure and repair/rerun behavior. Do not skip the regression or weaken it into an unconditional pass.

## Done when

- The regression proves both child variables are replaced, writable, and visibly reported when inherited values are invalid, entirely within fixture-controlled roots.
- A deterministic control with all normal candidates denied returns the expected environment failure, with no successful child run or suite-success evidence.
- Targeted checks pass in the ordinary sandbox with only the configured test temp root writable, and outside it. No permission changes or full-suite deadline increases are needed.
- A recorded suite passes through the normal workflow before landing. Record the #467 baseline classification when it becomes available; do not call the environmental test failure an intermittent flake.

## Plan

- [x] Reproduce the existing test's mkdir denial using a real macOS sandbox
  that denies host normal-temp roots while allowing the owned test root.
- [x] Add one per-run normal-temp selector dependency to localRunDeps and
  prepareScratch; nil uses the existing strict scratchdisk.NormalTempDir.
  Keep candidate policy, filesystem checks, cleanup and child env rules unchanged.
- [x] Build the inherited-temp fixture's worktree before invalidating its env,
  then exercise the real runUnderSlot boundary with an injected writable private
  fallback. Assert both child paths/writes, replacement notes and cleanup.
  Keep a runLocalCheck denial control: a failed selector prevents child execution
  and produces no suite-success evidence. Existing scratchdisk candidate tests
  retain the all-candidates-denied and actual-probe controls.
- [x] Run targeted checks inside and outside the sandbox, including candidate
  policy, actual child writes, denied-selector evidence protection and lint.

Workflow gates remain: push, recorded full suite and close review; then report
ready to the hub for its priority grant. No full-suite pass is claimed yet.

The simpler host-getconf fake is platform-specific and leaves Linux relying on
/tmp; a global hook would couple concurrent tests. A per-run IO dependency is
portable and keeps the production selector authoritative (ARCH-PURE, ARCH-DRY).
No new user-visible state, format, policy, permission or deadline is introduced.
Only fixed startup IO is added: one existing selector call per fallback; no
new input scan, worker or durable artifact. Fixture directories die with tests.


## Log

2026-10-19: Filed unclaimed by ariadne:3 while #476 waits behind #467 in the landing line. Checked the current tracker; #481/#482 are unrelated and no active title covers this test. The #467 report is from canonical submitted Couch receipt 914556c3-73fa-4dc0-a9bb-3af82a5a14e0. No process, machine permission, lock or #476 code changes were made.


### 2026-10-20 — Start and diagnosis
- 2026-10-20: closed — BR-1 host evidence is committed at the issue .side/sandbox-host-evidence.txt: exact source HEAD, profile SHA/content, probe source, argv, output and exit statuses. Under the profile, owned fixture create/remove passed; Darwin user-temp and /tmp writes were denied with errno1; both fallback tests plus NormalTemp/MkdirTemp/UserTempDir controls passed, both commands exit0. This addresses the reviewer nested-sandbox launch restriction without code changes. Earlier broader sandbox controls and outside-targeted tests passed. Production nil still selects strict NormalTempDir, child writes/cleanup and denied-selector no-child/no-success evidence are covered. Recorded full suite is queued/running on the unchanged code head; no full pass claimed yet. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

The #467 main-baseline diagnostic reproduced the same named test and sandbox
mkdir denial; sdlc classified persistent main failure and refused that landing.
This is deterministic environment dependence, never a registered flake. #476
subsequently landed with its unsandboxed integrated full suite passed. The hub
assigned this follow-up next, requiring a hermetic candidate and sandbox proof.
The fixture invalidates both inherited temp variables before runLocalCheck, so
its checkout setup can also need the host fallback. Construct that worktree
before invalidating env and target runUnderSlot, where the child-env contract
lives; retain top-level error/status protection in a separate denial control.


### 2026-10-20 — Sandbox red/green and implementation checks

The original test failed under the real macOS sandbox with the reported
operation-not-permitted errors in both host temp roots. With only the dependency
type stubbed, both new controls failed: fallback still used host roots, and an
ignored selector error allowed child execution and a successful status. Wiring
the per-run selector made both pass under the same sandbox (0.985s). The committed
profile is cmd/sdlc/testdata/487-deny-host-temp.sb; parent TMPDIR and GOTMPDIR were
set to an owned mktemp fixture under TestScratch. No machine permissions changed.

Wider sandbox checks passed (32.296s): scratch/fallback/prepare, environment,
local-run, known-flake gate, detached-worktree and cleanup controls. Strict
scratchdisk candidate tests passed (0.371s), as did golint (0.696s). Both focused
fallback checks also passed outside the sandbox (0.578s). Real child output
proves both vars use the private candidate and the run directory is removed.
The denial control uses runLocalCheck and proves no child, status, PR comment or
SuiteAttempt is published. Production selector and write-access policy remain
unchanged; the nil dependency takes the existing strict selector.

## Revisions

### 2026-10-20 — Separate completed implementation evidence from workflow gates

The last Plan checkbox now names the completed targeted verification. The full
recorded suite, close and readiness report remain workflow prose, so checking
the implementation does not claim future close or landing outcomes. Scope and
acceptance are unchanged.


### 2026-10-20 — BR-1 host sandbox acceptance evidence

Round 1 found no code defect but could not launch a nested sandbox in its
review environment (sandbox_apply: Operation not permitted). The requested
measured host proof is [committed here](000487-make-testscratchfallbackoverridesinheritedtemp-independent-of-host-sandbox-temp-permissions.side/sandbox-host-evidence.txt).
It records source HEAD, toolchain, profile digest/content, probe source and
commands, actual stdout/stderr and exit statuses. With that exact profile, the
owned fixture create/remove succeeded while both Darwin user temp and /tmp
writes were denied with errno 1. The two fallback tests and all selected
NormalTemp/MkdirTemp/UserTempDir controls passed under that same profile; both
commands exited zero. The cmd/sdlc package duration includes a 32-second small
test-slot wait (32.965s reported); scratchdisk passed in 0.241s. This supplies
acceptance evidence without changing or weakening the implementation. The
existing recorded suite remains queued/running on the unchanged code head.
