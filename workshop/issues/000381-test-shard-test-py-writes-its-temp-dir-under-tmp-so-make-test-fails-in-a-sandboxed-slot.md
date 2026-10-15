---
id: 000381
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
severity: low
card_mirror: '5e3370cee8bb75b2cd334ddddabab24ba6e7317d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T21:16:15-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# test-shard.test.py writes its temp dir under /tmp, so make test fails in a sandboxed slot

## Problem

`scripts/test-shard.test.py` (added by #369, a419c448) makes its temp dir with `tempfile.TemporaryDirectory(prefix="shard-report-", dir="/tmp")`. In a sandboxed agent slot `/tmp` isn't writable, so `cmd/sdlc TestShardedReportRetainsSubtestDiagnostics`, which runs the script, fails with `PermissionError: [Errno 1] Operation not permitted: '/tmp/shard-report-…'`, and `make test` exits 1. Seen at ariadne:2 on 2026-10-14 during #377's pre-close suite: every other shard and package passed. The AGENTS §10 temp-dir rule (#342) says temp dirs honour `TMPDIR`.

## Spec

Drop `dir="/tmp"`; Python's `tempfile` honours `TMPDIR`. Sweep `scripts/*.py` for other hardcoded `/tmp` temp dirs in the same change. `pkg/shelllint` covers shell `mktemp` only, so consider extending it to Python `dir="/tmp"`.

## Done when

- `TestShardedReportRetainsSubtestDiagnostics` passes in a sandboxed slot (TMPDIR set, `/tmp` unwritable), and no tracked script hardcodes `/tmp` as a temp root.

## Plan

- [ ]

## Log

### 2026-10-14
