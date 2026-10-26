---
id: 000643
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '9bf179701cdfa3646305656cd22683e59b35d924' # card fields mirrored from issue-cards; edit via sdlc
---

# A claim made from a removed temp worktree can't be released: unclaim refuses the slot itself

## Problem

## Spec

**A claim made from a temporary worktree can't be released once that worktree is gone (ops-echo:0, echo#45,
10-25).** echo:0 claimed echo#45 from a temp worktree under /Volumes/TestScratch (it planned there because start-plan
refused the extra worktree). The claim's owner records that worktree's path; after the worktree was removed, `sdlc
unclaim` from echo:0's slot refuses with "only its owner releases it", and nothing can act as that owner any more.
The owner should identify the slot (repo:N), not a checkout path that can disappear, or unclaim should accept a
claim whose recorded owner path no longer exists when run from the same slot.

## Done when

- A claim taken from a temp worktree of slot N is released by `sdlc unclaim` run from slot N after the worktree is
  removed, while a different live slot is still refused; tests for both.
- echo#45's stuck claim is releasable without editing the tracker by hand.

## Plan

- [ ]

## Log

### 2026-10-25
