---
id: 000588
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
card_mirror: 'f62964d42816937d51f7fd45acd888396607b2de' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead instructions link: trailing space trips the chat guard, and a refusal leaves a half-done write

## Problem

## Spec

**`tl-lead status instructions link` breaks on an item ending in a space, and leaves a half-done write (ops-echo:0,
10-23).** On a `@@chat@@` numbered item whose text ends with a trailing space, `link` trims the space before appending
` → [answer](…)`. Couch's chat guard (pair#838's) reads that as editing the operator's ask and refuses:
`editorial-chat at bytes 808:808: preserve each chat ask; only append its answer link`. By then `link` had already
written the `discussions:` index entry, so the failure left the document half-updated. Likely from ariadne#576's move to
the shared `sidedoc.ChatItems` grammar.

Fix: (1) append the link without changing any byte of the ask (trailing whitespace included); (2) make the operation
all-or-nothing: compose the item edit and the index entry, then write them in one session-sync write (or roll back the
first on failure).

## Done when

- Linking an item that ends in a space (and one ending in a tab) succeeds, leaving the ask's bytes intact (test).
- A refused link leaves the document unchanged, including the `discussions:` index (test).

## Plan

- [ ]

## Log

### 2026-10-23
