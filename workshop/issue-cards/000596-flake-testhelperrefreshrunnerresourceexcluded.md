---
id: 000596
status: open
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
github_issue:
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestHelperRefreshRunnerResourceExcluded]
---

# flake: TestHelperRefreshRunnerResourceExcluded

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestHelperRefreshRunnerResourceExcluded: Failed in pre-close attempt d5d9d108b935de10548b6452dd5c2396 on 6eb2f26b (ariadne#592, a test-fixture-only diff): runLocalCheck reported TimedOut after 338us with TimeoutMinutes 1, shard6 under full-suite load. Passed on 468ee2e4 an hour earlier, and passes 5/5 alone on 6eb2f26b.
