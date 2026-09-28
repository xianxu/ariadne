---
id: '000177'
status: done
started: 2026-07-14T17:20:20-07:00
created: 2026-07-14
updated: 2026-07-14
estimate_hours: 0.44
actual_hours: 0.23
---

# atlas gate: auto-satisfy when the close window contains no code changes

## Problem

The close-time atlas gate demands an `atlas/` change in the review window (or
`--no-atlas` + rationale) regardless of what the window contains. A window with
NO code changes — docs/workshop-only closes, analysis milestones — has no new
code surface to map, so the demand is incoherent there and forces a pointless
`--no-atlas` acknowledgment.

Sizing honesty (#172 M4 window-diffstat study over 71 trailer-bearing closes):
this is a **correctness fix, not a volume fix** — 10 of the 11 observed
`--no-atlas` closes DID change >50 code lines (they were review-fix re-close
windows with no *new* surface; that volume belongs to #174). Only ~1 was
actually code-free. Fix it because a docs-only close demanding atlas is wrong,
not because it will move the bypass counts much.
