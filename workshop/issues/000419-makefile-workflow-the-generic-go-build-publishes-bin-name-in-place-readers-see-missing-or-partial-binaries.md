---
id: 000419
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: 'b7ff775fc6b1e0cf2fdce4dfa76a6058f595c91c' # card fields mirrored from issue-cards; edit via sdlc
---

# Makefile.workflow: the generic Go build publishes bin/<name> in place; readers see missing or partial binaries

## Problem

## Spec

From pair#624, 2026-10-16. pair:0's `bin/couch` went missing for seconds at 15:40, 16:06 and 16:08 while Couch and the slots were calling it, so sends failed. pair:7 found that `Makefile.workflow`'s generic Go build writes each binary straight to `bin/<name>`, because Go's final step copies onto the target path. A private probe that replaced a live binary saw 1 missing-file and 418 partial-file observations. pair#624 fixes pair's own build; this is the substrate half, shared by every ariadne-style repo.

The generic build should compile to a temp file next to the target and `rename` it over `bin/<name>`, so a reader sees either the old binary or the new one, never a missing or partial file.

## Done when

- The generic build publishes each binary by building to a temp file and renaming it.
- A test with a paused or failing compiler shows `bin/<name>` stays the old, complete binary until the rename.
- A failed build leaves the old binary and no temp files.

## Plan

- [ ]

## Log

### 2026-10-16
