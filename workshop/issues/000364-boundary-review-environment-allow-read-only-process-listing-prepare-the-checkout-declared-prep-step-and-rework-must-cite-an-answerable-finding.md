---
id: 000364
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '5b377cf15afdf15a34109a5d554a8ee9d1da1751' # card fields mirrored from issue-cards; edit via sdlc
---

# Boundary review environment: allow read-only process listing, prepare the checkout (declared prep step), and REWORK must cite an answerable finding

## Problem

pair#528 M1's mandatory boundary review returned REWORK ten rounds running with zero open findings (2026-10-13). The reviewer stated the revised M1 contract was implemented, but would not pass it without verifying process behaviour itself, and couldn't:
1. its sandbox denies `/bin/ps` (the change and its tests inspect live processes);
2. its disposable checkout lacks generated, untracked build outputs (pair's `cmd/internal/runtimebundle/assets/runtime/files`), so some packages don't build or test there.

Host evidence (full suites, the blocked checks) was supplied in the Log, but a REWORK with no finding ID can't be disputed (`sdlc finding dispute` refuses never-issued IDs), so the author had no action: a dead end. The operator approved a `--no-judge` close for M1 as a one-off.

## Spec

1. **Read-only process listing in the review sandbox:** allow `ps` (and equivalent read-only process inspection), still denying signals, writes outside the checkout, and network beyond what the review needs.
2. **Prepared checkout:** a repo declares a preparation step (next to `.sdlc/local-checks.json`, e.g. `prepare` running `make build` or its asset generation); the review harness runs it in the disposable checkout before the reviewer starts, and records its result. Absent a declaration, behaviour is unchanged.
3. **REWORK must cite an answerable finding:** a verdict that blocks must carry at least one finding ID. "Could not verify X" becomes a named finding (kind: unverifiable) that the author can answer by fixing the environment or by `sdlc finding dispute` with host evidence, and the next round must rule on that evidence. A REWORK with no findings is rejected by the harness and re-run (bounded) rather than recorded.

## Done when

- A review of a change whose tests use `ps` can run them in the sandbox.
- With pair's declared prepare step, the reviewer's checkout builds and tests the runtimebundle packages.
- A reviewer response of REWORK with no findings is refused by the harness; an "unverifiable" finding can be disputed, and the next round answers it (tests for each).

## Plan

- [ ]

## Log

### 2026-10-13
