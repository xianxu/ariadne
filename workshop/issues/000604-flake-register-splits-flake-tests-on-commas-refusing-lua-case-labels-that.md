---
id: 000604
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '011f0ba20c426382fc0cf64cd80ffd6bd63fb4e2' # card fields mirrored from issue-cards; edit via sdlc
---

# flake register splits flake_tests on commas, refusing Lua case labels that contain one

## Problem

## Spec

**`sdlc flake register` refuses Lua case labels that contain a comma (parley.nvim:1, 10-24).** Registering
`retries a clean release whose reply was lost, so the next turn is new (#432)` failed: the card's `flake_tests`
splits on commas, so it read `so the next turn is new (#432)` as a separate, invalid test identity, and the card was
left unchanged. Plenary (and other) case labels are free text and often contain commas.

Fix: store `flake_tests` entries in a form that preserves arbitrary label text (a list in the card, or a quoting or
escaping rule applied symmetrically on write and read), with existing comma-free entries still readable.

## Done when

- A label containing commas (and quotes) registers, round-trips through the card, and matches the failing case in a
  recorded suite; a test pins it.
- Existing cards with comma-separated entries read unchanged.

## Plan

- [ ]

## Log

### 2026-10-24
