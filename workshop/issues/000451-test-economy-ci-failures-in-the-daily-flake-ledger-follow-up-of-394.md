---
id: 000451
status: open
deps: [ariadne#394]
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'a14fe1b44391c4c7b17949e799bf43593cbcee7c' # card fields mirrored from issue-cards; edit via sdlc
---

# test economy: CI failures in the daily flake ledger (follow-up of #394)

## Problem

#394's `sdlc flake list` shows the daily flake view from local suite attempts only; it prints CI as `not covered`. #390 §3 and #394's Done-when 3 want CI failures in the TL's daily view too, kept as their own classes (branch regression, infrastructure/fixture fault, confirmed same-head flake), never merged into the local flake rate.

## Spec

## Done when

- `sdlc flake list` reads CI outcomes (failed GitHub check runs on PR heads and main) and lists them by class with run links, separate from local same-head flakes.
- Missing CI data is shown as unknown, never as zero; tests use a stateful fake of the checks seam.

## Plan

- [ ]

## Log

### 2026-10-18
