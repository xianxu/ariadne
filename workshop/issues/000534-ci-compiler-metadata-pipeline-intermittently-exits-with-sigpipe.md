---
id: 000534
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'e2a022303e8e88e90e15a171c6698c12e9732801' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T17:14:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# CI compiler metadata pipeline intermittently exits with SIGPIPE

## Problem

While verifying echo#1 with the unchanged upstream portable-ci.test.sh harness,
the actual workflow Build first pinned compiler step intermittently returned
141 even though its metadata had the selected revision and modified=false.
The template pipes printf into grep -Eq under set -euo pipefail. Early grep
success can close the pipe while printf is writing, producing SIGPIPE.

## Spec

Remove early-consumer-close races from the pinned compiler workflow template's
metadata checks. Use a here-string or a grep that consumes the whole input
(without -q, redirect output to /dev/null). Sweep the analogous metadata
pipelines in the template and its test harness. Preserve exact revision and
modified=false validation and fail-closed behavior. Consumers should obtain
the change through weave; do not patch each generated workflow manually.

## Done when

- Exact revision and modified checks reliably succeed for valid metadata and
  reject mismatches under pipefail, with repeated actual-block tests.
- The metadata pipeline class is enumerated and fixed in source templates.
- portable-ci.test.sh still exercises identity failures and old PATH exclusion.

## Plan

- [ ] Reproduce the exact metadata pipelines under pipefail; add repeated actual-block coverage for the incident's three-line input and larger metadata, plus mismatch/dirty/missing-field controls.
- [ ] Replace both metadata quiet-consumer pipelines in the shared build template with full-input checks. Add a scoped mechanical shell-source guard and verify its negative controls; enumerate metadata consumers in source and harness.
- [ ] Run portable-ci identity/PATH controls, focused lint/script checks, and the recorded suite. Close and publish a reviewable PR explicitly requiring the operator's local-gates-verified label; notify the hub, never apply the label.

Quick-flow scope: shared workflow source and its normal tests; consumers derive via weave, no consumer patching. ARCH-DRY retains the one generated build.sh rule used by both compiler slots and tests its extracted checks. ARCH-PURE keeps lexical detection in the existing shelllint pure core; the Go AST-only golint catalog cannot inspect YAML run blocks, so enforce this shell mistake in shelllint's Go-suite repository guard. No new runtime state or durable format. CI batch cost is two linear full-input reads of metadata; test repetition is bounded, with no network or real compiler builds in the existing stateful harness.

## Log

2026-10-21: echo#1 observed actual workflow exit 141 locally after an earlier
pass with the same workflow. Minimal exact-pipeline reproduction on macOS Bash:
set -o pipefail; printf '%s\n' "$METADATA" | grep -Eq "build[[:space:]]+vcs.revision=$PIN$"
METADATA was three lines: binary: go1.26.3, two-space build vcs.revision equal
to 63d6bfd86509d97649e10d7a5dd1e9a06db17d21, and two-space build vcs.modified=false.
1000 fresh Bash subprocesses returned {0: 999, 141: 1}. This is a scheduling
race rather than an OS-specific contract. Echo TL authorized upstream filing
and logged reruns; no template modification was made in echo.

- Claimed and started from main after #528 completed. Dispatch requires operator verification for the protected workflow change; PR will name the label requirement. Source enumeration: two metadata printf-to-quiet-grep checks in .github/workflows/merge-check.yml, shared by both pinned compiler slots; portable-ci.test.sh executes that generated block and has no independent metadata-check pipeline. File-input greps and unrelated non-metadata checks are outside this fix.
