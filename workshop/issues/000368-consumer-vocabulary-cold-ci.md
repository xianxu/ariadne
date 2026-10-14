---
id: 000368
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '72eec97b3650b7bac671c6dc600b501c442e6bac' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T09:58:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: parley.nvim:2
    worktree: /Users/xianxu/workspace/worktree/parley.nvim-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# Cold consumer CI cannot prove committed generated vocabulary ownership across substrate updates

## Problem

parley.nvim#326: main merge-check run 38601556765 at 62a35b2aeb4a849cee910062a35ea02f0b5dba29 fails during bootstrap: preserving authored replacement at construct/generated/vocabulary/issue.json. CI installs published weave and clones current substrate main, while Parley ships a committed vocabulary export for standalone use. Its ignored construct/generated/weave ownership inventory is absent in a fresh checkout. A staged output differing from that committed export therefore has no previous ownership proof and is treated as authored. On a warm checkout the inventory permits displacement/regeneration. This means upstream vocabulary changes can break unchanged consumer CI.

## Spec

Define a reproducible consumer generation contract that preserves authored-file safety and standalone shipped vocabulary. Evaluate pinning substrate/tool revisions with committed generated provenance, or validating committed exports against their recorded source before advancing generation. Do not blanket overwrite occupied generated paths or weaken authored-replacement protection. Upstream vocabulary changes should carry explicit consumer regeneration/propagation when consumers track that source.

## Done when

- A cold clone of a consumer with a committed vocabulary export boots at its recorded substrate revision, even after upstream vocabulary changes.
- Explicit substrate advancement regenerates consumer exports with reviewable drift.
- Authored/tampered generated replacements remain protected; automated coverage exercises cold and warm checkouts, upstream changes, and local edits.

-

## Plan

- [ ]

## Log

### 2026-10-14
