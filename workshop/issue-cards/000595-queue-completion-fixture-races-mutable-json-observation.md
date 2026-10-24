---
id: 000595
status: open
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
github_issue:
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestQueueRealHelperCompletion/test-lock-exit.test.py]
---

# Queue completion fixture races mutable JSON observation

## Problem

Recorded ariadne#589 suite hit JSONDecodeError in the queue-helper completion fixture while polling mutable operation bytes. The same top-level test then passed on its exact failed head with no source changes.
