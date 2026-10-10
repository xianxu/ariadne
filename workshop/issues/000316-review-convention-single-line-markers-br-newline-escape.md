---
id: 000316
status: working
deps: []
github_issue:
target: review-convention
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: 'b6e58b731c46d5fc90aaf86b9c9b7b16a6a68189' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-09T21:20:31-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    worktree: /Users/xianxu/workspace/worktree/parley.nvim-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "8ebc71b7", done: "b55b6c6a"}
---

# review-convention: single-line markers + <br> newline escape

## Problem

parley.nvim#312 renders 🤖 markers compactly (chain concealed, anchor
highlighted) and opens a chain in a thread float. That rendering is line-local:
a marker that spans lines can't be concealed cleanly and the incremental
re-render depends on markers being line-local. Agents (xx-fix, Claude Code,
pair) also write markers, so the rule has to live in the shared grammar or
agent-written markers keep spanning lines.

## Spec

Revise `construct/local/fix/review-convention.md`:
- §3: a marker is a single line; a newline inside a `[]`/`{}`/`<>` block is
  written `<br>`. (Supersedes the #125-era multi-line allowance for writers;
  readers may still tolerate legacy multi-line markers.)
- §5: accept/reject decode `<br>` back to a newline in the resolved text.
- `## Revisions` entry naming parley.nvim#312.

## Done when

- The target carries the single-line rule, the `<br>` escape, the resolve
  decoding note and a Revisions entry; parley's weaved copy picks it up on the
  next weave.

## Plan

- [x] Edit §3 / §5 + Revisions in `construct/local/fix/review-convention.md`

## Log

### 2026-10-09
