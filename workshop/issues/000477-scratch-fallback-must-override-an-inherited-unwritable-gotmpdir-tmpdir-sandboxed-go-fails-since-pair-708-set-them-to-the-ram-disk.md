---
id: 000477
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '68599ddc08daa58af6ab996e4674f7ee994f96cf' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T17:59:41-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "3091c833", done: "37f0e58d"}
---

# Scratch fallback must override an inherited unwritable GOTMPDIR/TMPDIR (sandboxed go fails since pair#708 set them to the RAM disk)

## Problem

Since the 10-19 reboot, pair#708 gives every slot shell `GOTMPDIR=/Volumes/TestScratch/tmp/<repo>-slotN` (and the
RAM-disk temp folder), but the agent sandboxes deny writes to `/Volumes/TestScratch`. Every sandboxed `go build` /
`go test` fails with "go: creating work dir ... operation not permitted" (ops:0 reproduced: `touch` into its own
GOTMPDIR is denied). ariadne:4's sandboxed `sdlc merge` for #456 saw its integration suite and the main baseline both
fail in 0 s, so main looked red when it wasn't.

`prepareScratch` (cmd/sdlc/testscratch.go) already falls back to the normal temp folder when the scratch root isn't
writable, but on fallback it returns no env, so the child inherits the unwritable `GOTMPDIR` (and `TMPDIR` if
inherited) from the slot shell and fails anyway.

## Spec

- On fallback, `prepareScratch` returns env entries that point `TMPDIR` and `GOTMPDIR` at a writable normal temp
  folder, overriding inherited values, instead of returning none.
- More generally, before a suite or baseline runs, an inherited `GOTMPDIR`/`TMPDIR` that isn't writable is replaced
  (with a one-line note naming it), so an environment mismatch never shows up as a red suite in 0 s.
- A suite whose every package fails at "creating work dir" is classified as an environment failure, not a red main
  (merge says so and names the variable).

## Done when

- Test: inherited `GOTMPDIR` pointing at an unwritable dir + unwritable scratch root → the child runs with a writable
  `GOTMPDIR` and the suite runs.
- Test: an all-packages "creating work dir" failure is reported as an environment failure naming GOTMPDIR.

## Plan

- [x] Repair scratch and normal-temp fallback with explicit child environment and cleanup.
- [x] Classify proven startup environment errors without publishing code-failure evidence; cover ordinary and train baselines.
- [x] Verify focused regressions and publish the candidate for recorded-suite and close review.

Recorded full-suite pass, SHIP and landing remain mandatory workflow gates; their
progress and evidence are tracked in the Log.

Durable plan: [scratch environment plan](../plans/000477-scratch-environment-plan.md).

## Log

- The real fix for the sandbox side is the operator adding `/Volumes/TestScratch` to the agents' sandbox write
  allowlist (a local-gates change, Needs you since #468). This issue makes sdlc robust either way.


### 2026-10-19 — diagnosis and proposed boundary

prepareScratch returns nil env on rejected scratch or failed scratch mkdir, so
runUnderSlot inherits the bad variables. Detached-worktree creation also uses
os.TempDir before that point, so an invalid TMPDIR can fail earlier. The scratch
fix must repair both boundaries through shared normal fallback selection.

Read-only exploration found that an ordinary saved failed SuiteAttempt would
poison a later same-head pass through guardSuitePass. The proposal retains
recognized environment startup output in a separate bounded diagnostic log,
returns a could-not-run error before status/attempt/flake publication, and
propagates baseline execution errors through train diagnosis. No persisted
format change, sandbox setting change, live denial probe or #476 attribution
change is proposed. Full durable plan is ready for hub approval before code.

### 2026-10-19 — approved design and temp repair checkpoint

Hub approved the design (receipt 155114fd-74e4-40e9-9132-45d4878d20b6).
change-code succeeded with inferred quick flow; no plan-quality or estimate
was claimed. The eventual diff remains subject to close's full review.

Observed red: startup failures published ordinary failed-suite evidence,
ordinary baseline diagnosis called main red, and the child inherited unusable
temp variables. The temp slice now repairs detached-checkout creation and sets
both child variables to one owned writable directory, with explicit RAM
provenance and cleanup. Normal fallback also sweeps abandoned owned run folders.
Focused scratch/child tests passed (scratchdisk 0.131s; cmd/sdlc 1.918s).
Environment classification and baseline propagation remain pending; their
committed regression tests intentionally remain red until that slice is done.

### 2026-10-19 — implementation and focused verification

The startup classifier passed its direct, multi-process, captured Python/make
wrapper, mixed-code/test, pass/activity, quoted, malformed/truncated and long-line
controls. Counted N=8/16 fixtures visit exactly N lines. It runs before flake
classification/publication and uses the same bounded complete-log read as facts.
A known-flake-rerun regression first failed because it returned no environment
error; sharing the classifier/retention path fixed that caller too.

Environment output is retained separately, with no SuiteAttempt schema change,
status or comment. A legacy fallback uses a non-evidence pending path until
classification, so even diagnostic-retention failure cannot look like suite.log.
Retention failures explicitly preserve/name the raw log. Tests prove same-head
repair passes, an earlier real failure still blocks, direct errors skip baseline
diagnosis, both baseline paths keep main health unknown, and rerun errors post
nothing while preserving the genuine parent failure.

Focused verification passed: scratchdisk 0.132s, tracked Go lint catalog 0.676s,
and cmd/sdlc scratch/environment/local-check/known-flake/saved-result/landing-train
controls 101.929s. README, local-check help, atlas and lesson updated. Recorded
full suite and mandatory close review remain pending. No live sandbox changes,
shared staging, timeout changes or #476 attribution changes.

### 2026-10-19 — published candidate

PR https://github.com/xianxu/ariadne/pull/333 is open. The required full recorded
suite is queued through the normal central sdlc on the pushed implementation
head. Preserve that waiter; no staging, private runner, or manual main chasing.
Close/landing wait for its pass and the mandatory review.

## Revisions

### 2026-10-19 — overlap independent verification gates

Separate completed implementation/targeted verification from the remaining
workflow gates in the checklist. Per tl-worker, the binary-owned close review
may overlap the queued recorded suite; no full-suite pass is claimed until
it completes, and landing still requires the final code's recorded pass.
The approved design and acceptance criteria are unchanged.

### 2026-10-19 — close round 1, BR-1 fixed

Close selected the full review automatically (323 added production-code lines).
Round 1 returned FIX-THEN-SHIP with Important BR-1, but the ledger kept close
unfinalized, so the next step is an ordinary close review after the fix.
The finding is valid: checkout creation can repair into a normal root the stale
sweep did not enumerate. Searched every scratchdisk.Roots/MkdirTemp caller: only
the os.TempDir-based local-check family gains repair roots; explicit review
cache roots stay unchanged.

Observed red: the stale private per-user checkout survived with bad TMPDIR.
After shared root enumeration, old checkouts are collected and fresh ones
survive even after TMPDIR is healthy again. Focused normal-root/creation/cleanup
checks passed (scratchdisk 0.159s; cmd/sdlc 0.892s). Added the cleanup-parity lesson.
Cancelled only this checkout's old-head queued suite with local-check cancel;
it never started tests and recorded no result. The fixed head needs a newly
recorded suite and a second close review; no timeout was raised.

### 2026-10-19 — close round 2, BR-2 and recovery command fixed

Round 2 verified BR-1 addressed but found the same cleanup family for per-run
directories after fallback recovery (BR-2). Close remains unfinalized. I missed
that sibling family in the first repair. The rule is now shared across both
families: scratchdisk.Roots enumerates every possible placement, and both the
checkout sweep and prepareScratch consume it independently of today's selection.
Review-cache creation/cleanup remains a fixed custom family, unchanged.

Both new regressions failed first: a former fallback run directory survived
after TMPDIR recovered, and targeted-rerun recovery printed the internal
suite-rerun name instead of the declared suite command. The shared-root sweep
and recovery-name mapping make them pass. Old removal, live/fresh preservation,
unrelated-tool preservation and both checkout/temp families are covered.
Focused checks passed: scratchdisk 0.196s, tracked lint 0.945s, cmd/sdlc 12.759s.

Cancelled only the second old-head waiter before tests started; no full suite
has run or produced an outcome yet. The fixed candidate needs the recorded
suite and third review. If that review asks for another cleanup mechanism
change, stop and re-plan with the hub before any further implementation.
