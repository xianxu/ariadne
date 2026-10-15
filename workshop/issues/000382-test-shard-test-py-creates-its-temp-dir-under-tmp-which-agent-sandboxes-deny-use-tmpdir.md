---
id: 000382
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '343f6315d1f97d8a8977e09912ab0b366ec0f010' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T21:42:00-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c44738ca", done: "e42711c6"}
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

- [ ] Remove hardcoded test temporary-directory roots found by the sweep.
- [ ] Verify the report with a writable TMPDIR and denied outside allocation; run the recorded full suite.

## Log

### 2026-10-14

Claimed ariadne:5 and started from fresh main. The authorized small design is to let Python tempfile select its environment-configured root, removing explicit /tmp overrides rather than adding another fallback policy (ARCH-DRY). Initial inventory found the report test and embedded Python in scripts/test/slot-dependencies.test.sh. Wider scratch-state findings were sent to the hub for scope guidance: brainstorming start/stop is a coupled managed-skill protocol, and close-issue.py uses a warmup counter file rather than a temporary directory. A documented allocation-denial check will execute the real report test with a custom writable TMPDIR and reject temporary allocations outside it; the issue explicitly permits a documented check.
