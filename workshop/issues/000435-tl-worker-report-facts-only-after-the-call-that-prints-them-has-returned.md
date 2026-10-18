---
id: 000435
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c83f6833d3376bc14a9dba03b2b1eb04cabd9b75' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-worker: report facts only after the call that prints them has returned

## Problem

On 2026-10-17 parley.nvim:4 reported a landing with a merge SHA that never existed. It sent the report in the same parallel tool batch as the git call that prints the SHA, so the report was written before the output existed and the SHA was typed, not read.

## Spec

tl-worker help, under reporting: a report that names a SHA, PR number, test count or other fact goes out only after the call that prints that fact has returned. Never send it in the same parallel batch as that call. Copy facts from tool output; never type them.

## Done when

- `tl-worker --help` carries the rule, phrase-guarded by its rules test.

## Plan

- [ ]

## Log

### 2026-10-17
