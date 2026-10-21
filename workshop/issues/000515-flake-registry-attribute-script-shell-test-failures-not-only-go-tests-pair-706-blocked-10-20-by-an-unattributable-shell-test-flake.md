---
id: 000515
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c062ab24656aed47bc87277146a75b39d718d96a' # card fields mirrored from issue-cards; edit via sdlc
---

# flake registry: attribute script/shell test failures, not only Go tests (pair#706 blocked 10-20 by an unattributable shell-test flake)

## Problem

The known-flake registry attributes only Go test failures. On 10-20 pair's shell-wrapped test
`tests/copy-on-select-test.sh` (a subprocess timeout, later traced to macOS AMFI assessing per-run binary copies,
pair#749) failed pair#706's suite; the rerun passed every test, yet `sdlc` refused: "an earlier failure on this head
is unexplained … judged unattributable-failure. Registering a flake cannot explain it". A shell/script test failure
can't be registered at all, so the only path is a new commit, even when the failure is a known, registered-by-name
flake with an issue.

## Spec

- Let the flake registry name non-Go test units (a script path plus a failure signature, e.g. the script and its
  failing check), with the same rules as Go flakes: an issue is required, entries expire, and the same-head rerun
  must pass.
- The suite's failure attribution reads script-test failures in a structured way (the runner already knows which
  script failed), so a registered script flake explains a failed attempt the way a Go one does.

## Done when

- Tests pin: a registered script flake makes the earlier failure attributable and the passing rerun counts; an
  unregistered script failure still refuses.

## Plan

- [ ]

## Log

### 2026-10-20

## Log

- 10-20 20:10 (TL): related catch-22 seen on pair#706: after an unattributed (unregistrable) shell-test failure on a
  head, train admission refuses ("PR has failing checks on its head"), but the fix (pair#749) only reaches the PR
  through main integration, which happens after admission. Recovery was the slot integrating main into its branch by
  hand and running a fresh suite on the new head. Consider: let admission integrate main when the only red check is a
  suite whose failure predates a landed fix, or at least name the manual recovery in the refusal.
