---
id: 000515
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
---

# flake registry: attribute script/shell test failures, not only Go tests (pair#706 blocked 10-20 by an unattributable shell-test flake)

## Problem

The known-flake registry attributes only Go test failures. On 10-20 pair's shell-wrapped test
`tests/copy-on-select-test.sh` (a subprocess timeout, later traced to macOS AMFI assessing per-run binary copies,
pair#749) failed pair#706's suite; the rerun passed every test, yet `sdlc` refused: "an earlier failure on this head
is unexplained … judged unattributable-failure. Registering a flake cannot explain it". A shell/script test failure
can't be registered at all, so the only path is a new commit, even when the failure is a known, registered-by-name
flake with an issue.
