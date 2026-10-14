---
id: 000363
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '061acdc1115eb39585db8d35752dad6e7b844873' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T23:21:46-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "fd79a0e3", done: "aee77c44"}
---

# Construct shell tests isolate inherited color and terminal settings

## Problem

The #333 full suite failed only in vm-log.test.sh: Codex supplies NO_COLOR=1,
which overrides the test's CLICOLOR_FORCE=1. Unsetting NO_COLOR made the unchanged
test pass. Inherited CLICOLOR_FORCE also breaks the plain dim assertions.

## Spec

Make vm-log.test.sh own its color environment by clearing inherited NO_COLOR and
CLICOLOR_FORCE before its explicit cases. Keep production color precedence unchanged.
Sweep all six construct shell tests and their invoked scripts for NO_COLOR, TERM,
and CI assumptions; exercise the existing runner under plain, no-color/CI, and
forced-color environments. Reuse its bounded process-group cleanup (ARCH-DRY).

## Done when

- The shell runner passes with NO_COLOR=1 and with inherited CLICOLOR_FORCE=1;
  regression cases fail against the previous vm-log test.
- The sweep records any TERM/CI assumptions and the full make -k test passes.

## Plan

- [x] Add environment regression cases and observe their failures.
- [x] Isolate vm-log test color inputs; verify all shell tests and the full suite.

## Log

### 2026-10-13
- 2026-10-13: closed — Full make -k test at 77230109f766933ffc042e1f018823b3d84f9de4 passed: all 1239 SDLC tests and 52 other packages, zero failures, 399.4s; /tmp/ariadne-363-full-suite.log. Regression-first environment matrix failed before fix on no_color_ci forced-color assertion and forced_color_terminal plain dim assertion, while all five sibling scripts passed; after clearing inherited flags all three full shell-runner cases passed in 24.798s. Read all six construct test scripts and swept invoked scripts for NO_COLOR, TERM, CI; only vm-log color assertions needed isolation. Production color precedence remains exercised unchanged. Atlas updated; git diff --check clean.; review verdict: SHIP

Hub explicitly requested this fix after #333 landed. All six .test.sh files read;
only vm-log.test.sh depends on these color inputs. The other five assert paths,
exit codes, generated commands, and fixture state; their called scripts do not
branch on NO_COLOR, TERM, or CI (sandbox.sh's TERM is a signal name).

Regression first: TestConstructScripts/no_color_ci failed on forced-color step,
and forced_color_terminal failed on plain dim output; all five siblings passed.
The test now clears both inherited color flags, preserving its explicit NO_COLOR
precedence assertion. The Go runner reuses the existing per-run deadline and
process-group cleanup for all three environment cases.

Targeted green: go test ./scripts/test -run '^TestConstructScripts$' -count=1
passed in 24.798s. Full make -k test at
77230109f766933ffc042e1f018823b3d84f9de4 passed in 399.4s after a 505s lock wait:
all 1239 SDLC tests and 52 other packages, zero failures, under the normal Codex
environment. Log: /tmp/ariadne-363-full-suite.log. Atlas documents the matrix.
