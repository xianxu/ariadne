---
id: 000640
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: 'f2e6ad889cff8a8650191dab8b85bb68029f0dfd' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc queue status --json and sdlc landing status --json expose ages and holder pids

## Problem

## Spec

**Machine-readable ages for the queue and the landing line (via ops#10, for tl-lead sweep).** `sdlc queue status --json` (schema_version 1) lists running operations as id, workspace, kind, state and head: no start time or holder pid. Waiting suite tickets and their ages appear only in the human text ('ticket 55: position 1; age 5392s; ...'). `sdlc landing status` has no --json; the landing now (issue, pid) and each waiter's 'waiting since' and readiness are text only. A sweep that flags a stuck operation or a long wait has to parse human output, which drifts silently.

Add, additively (ARCH-EVOLVE: new fields; no meaning changes under schema_version 1):
- queue status --json: per active operation `started_at` (RFC 3339) and `pid` (the holder; 0 when unknown), plus `holder_alive` where process inspection is complete. A `waiters` array mirroring the human ticket lines: ticket, position, age_seconds, class, protected, granted, repo, worktree.
- landing status --json (new, schema_version 1): trains state, `landing_now` {issue, slug, pid, started_at} or null, and `waiting` in order {issue, slug, pid, class, waiting_since, ready, ready_head, pr, not_counted_reason}. Plus the delay p50/p90 already printed.

## Done when

- Golden tests for both JSON forms, including empty, a landing in progress, and waiters, alongside the existing human-form tests. An older reader ignores the new fields.
- The help for each verb names the schema.

## Plan

- [ ]

## Log

### 2026-10-25
