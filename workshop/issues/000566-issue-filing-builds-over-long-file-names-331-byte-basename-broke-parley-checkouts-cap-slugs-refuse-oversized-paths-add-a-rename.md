---
id: 000566
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '5f63bba8d10b0d9a2929aa8cecb8f3b6b0b30895' # card fields mirrored from issue-cards; edit via sdlc
---

# Issue filing builds over-long file names (331-byte basename broke parley checkouts); cap slugs, refuse oversized paths, add a rename

## Problem

## Spec

`sdlc flake register` (and anything else that files an issue from a generated title) builds the details/card file
name from the full title. On 10-22 it filed parley.nvim#421 from a title made of two long Lua test names; the
basename was 331 bytes, over macOS's 255-byte limit, and was pushed to `main` and `issue-tracker`, so every checkout
of parley main would fail with "File name too long". The TL renamed both files by hand with a plumbing commit
(main `26b27da7`). `sdlc issue set-title` keeps paths, and there's no rename verb.

Fix: cap the generated slug (e.g. to 80 bytes at a word boundary, keeping the ID prefix), in the one slug builder
every filer uses; refuse to publish any path whose basename exceeds the filesystem limit; and give `sdlc issue` a
supported rename for an existing card/details pair (both refs, content unchanged).

## Done when

- Filing from a 300-character title produces a basename within the cap (test), for both details and card.
- A path over the limit is refused before any push (test).
- `sdlc issue rename --issue N --slug S` renames both files on both refs (test).

## Plan

- [ ]

## Log

### 2026-10-22
