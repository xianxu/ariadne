---
id: 000487
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
severity: high
collaboration-mode: collaboration
card_mirror: 'f449de7aa53a57595be05fa2ac47cd2a4a29cc6c' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

2026-10-19: Filed unclaimed by ariadne:3 while #476 waits behind #467 in the landing line. Checked the current tracker; #481/#482 are unrelated and no active title covers this test. The #467 report is from canonical submitted Couch receipt 914556c3-73fa-4dc0-a9bb-3af82a5a14e0. No process, machine permission, lock or #476 code changes were made.
