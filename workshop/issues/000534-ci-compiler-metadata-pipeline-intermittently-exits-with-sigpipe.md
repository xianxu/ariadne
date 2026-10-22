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
flow: {kind: quick, provenance: inferred, spec: "0f49607d", done: "adaf0bd0"}
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

- [x] Reproduce the exact metadata pipelines under pipefail; add repeated actual-block coverage for the incident's three-line input and larger metadata, plus mismatch/dirty/missing-field controls.
- [x] Replace both metadata quiet-consumer pipelines in the shared build template with full-input checks. Add a scoped mechanical shell-source guard and verify its negative controls; enumerate metadata consumers in source and harness.
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
- Root-cause proof: exact unchanged template checks in 1,000 fresh Bash processes produced 998 passes and two exit-141 failures; ten larger valid-metadata attempts all exited 141 (30.279s total). New actual-check loop failed red with incident exit 141 before the fix.
- Both shared metadata checks now use non-quiet grep with output redirected to /dev/null, preserving pipefail and match semantics. Source/harness enumeration found no independent metadata checks to patch; both pinned compiler slots use the generated build.sh.
- Mechanical guard: pure MetadataPipeFindings in pkg/shelllint reuses existing shell lexing; its unit test failed against a no-op implementation, then the repository scan failed on exactly workflow lines 89 and 90. Applied to tracked shell scripts and workflow sources. Safe controls include file/here-string input, quoted examples/comments, unrelated pipes, pattern arguments, and option terminators. Shelllint package passed (0.279s). The Go AST-only catalog cannot inspect YAML; this shell-source rule is enforced by the existing Go-suite guard in its proper owner.
- Fixture correction: env delivery initially left the large metadata variable exported, causing exec environment-size refusal. Clear its export attribute to match the workflow's local assignment before expansion. Final actual-block regression passes on Bash 3.2.57: 1,000 incident iterations, 20 one-MiB valid iterations, and wrong/dirty/missing/empty refusals; all existing source identity, old PATH exclusion, failure propagation and range controls pass. A quiet-consumer mutation of the extracted large-input checks exits 141. Infrastructure/fixture errors are not test evidence for the fix.
- Targeted verification complete: final portable-ci actual-block harness passed; shelllint and golint full packages passed (0.165s / 0.615s). Recorded suite submitted on pushed 5a855c53a3efef734b709ed95a337607a6c859d1 as lease-2835200300 (terminal retained), initially FIFO position 10. Pending, not passing evidence. PR #380 explicitly requires operator-applied local-gates-verified; guard currently red while label absent, no label or bypass applied.
