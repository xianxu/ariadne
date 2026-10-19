---
id: 000460
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd0cc51888725481f72811627438a7fa79dd0faa2' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc card decoder refuses unknown fields: a merge started before a rebuild rejects new card fields (ARCH-EVOLVE)

## Problem

ariadne#394 added `flake_tests` to the tracker card. sdlc's card decoder refuses unknown fields, so any sdlc process started before ariadne:0 was rebuilt rejects every card that carries the new field. On 2026-10-19 pair:3's long-running `sdlc merge` for pair#687 (started on the old binary) refused the card after the server-side merge and needed manual recovery; other merges in flight at the time are exposed the same way. This is exactly what ARCH-EVOLVE (ariadne#454, landed the same night) forbids: an older reader must tolerate and preserve unknown fields.

## Spec

Make the tracker-card decoder (and any card rewrite path) tolerate and preserve unknown members, per ARCH-EVOLVE: capture on decode, write back unchanged; known fields, types and required fields stay strict. Audit sdlc's other persisted records that a long-lived process reads (landing events, test-stats records, readiness sidecars from #449) for the same strictness, and fix or list them. A regression: an old-shape decoder reads a card with an extra field, rewrites it, and the field survives.

## Done when

- A card with an unknown field round-trips through every card read-modify-write path; the regression proves it.
- The audit list is in the issue Log, each item fixed or filed.

## Plan

- [ ]

## Log

### 2026-10-19
