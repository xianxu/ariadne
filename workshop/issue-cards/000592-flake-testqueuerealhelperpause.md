---
id: 000592
status: open
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
github_issue:
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestQueueRealHelperPause]
---

# flake: TestQueueRealHelperPause

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestQueueRealHelperPause: failed in attempt c5fff316216ec03a781b4796f770a3e5, passed on the same head in attempt f3122a02c4fdbbaba2e5dd1f03b2b89f
