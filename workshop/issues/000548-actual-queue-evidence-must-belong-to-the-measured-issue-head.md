---
id: 000548
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
severity: critical
collaboration-mode: collaboration
card_mirror: 'd472d2c4bb21cdcd871dfe967c8345e950119ab5' # card fields mirrored from issue-cards; edit via sdlc
---

# Actual queue evidence must belong to the measured issue head

## Problem

Echo#9 actual preview refuses because another issue's recorded suite attempt lies in its measurement period while echo:0 holds more than one claim. The suite attempt belongs to echo#3, not echo#9. This blocks echo#9's fix and therefore echo#10, #3 and #5 suites/landings.

## Spec

A suite attempt may affect actual measurement only for the issue whose candidate head it ran on. Do not adopt or demand engagement proof for another issue's attempt merely because repository, slot and clock overlap. Preserve strict refusal for a selected issue's missing or ambiguous pinned ownership; do not drop uncertain evidence that really belongs to it.

## Done when

- A real fixture holds two issues in one slot, records a suite for issue A, and measures issue B without adopting or refusing on A's attempt.
- A's actual still uses its own exact recorded queue endpoints, and missing/ambiguous ownership for a relevant attempt still refuses.
- Head identity and issue lineage, rather than current branch-name guesses or timestamp overlap alone, decide relevance.
- Echo#9 actual preview passes the owner-run live check before adopting hours.

## Plan

- [ ]

## Log

2026-10-22: Discovered in #544 live conformance; ops-echo:0 delivery f8174005-af47-42ac-af83-c57928432164. Read-only actual9 at its pushed branch head on clean #544 binary b1ea682 refused: suite attempt fcf15650129203dfd1408728d625962b has 0 matching pinned slot engagements. Echo owner identifies it as echo#3's failed recorded run. echo:0 had kept #3 blocked at close and taken #9 per ops:0, so both claims coexist legitimately. The #544 repair and mixed legacy/tracker measurement for #3 passed independently. TL delivery59ea794e-0b24-4d8f-85cf-c1b571d0117a keeps this outside #544; delivery7ebee200-0aa0-49ac-9606-ee9ac3af6f02 assigns it to ariadne:1 immediately after #544 lands, plan first, with parallel planning from main allowed while #544 queues.
