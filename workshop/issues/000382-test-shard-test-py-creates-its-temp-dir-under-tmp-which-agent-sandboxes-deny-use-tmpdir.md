---
id: 000382
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '8bf699775854d9ff2811cbb237df330ea144682f' # card fields mirrored from issue-cards; edit via sdlc
---

# test-shard.test.py creates its temp dir under /tmp, which agent sandboxes deny; use TMPDIR

## Problem

`scripts/test-shard.test.py:24` (from #369) creates its temp dir with
`tempfile.TemporaryDirectory(prefix="shard-report-", dir="/tmp")`. Agent sandboxes (Claude Code's) deny writes to
`/tmp`, so `TestShardedReportRetainsSubtestDiagnostics` fails with PermissionError in every sandboxed `make test`
(found by ariadne:4 on 10-14 while running #367's suite). Unsandboxed runs pass, so the recorded suite hides it.

## Spec

Drop `dir="/tmp"` so the temp dir follows `TMPDIR` (the #342 convention). Sweep the repo's scripts and tests for other
hardcoded `/tmp` temp locations and fix them the same way.

## Done when

- The test passes with `TMPDIR` pointing at a writable directory and `/tmp` unwritable (a test or a documented check).
- No hardcoded `/tmp` temp dirs remain in scripts or tests, or each remaining one is justified in a comment.

## Plan

- [ ]

## Log

### 2026-10-14
