---
id: 000177
status: working
deps: []
github_issue:
created: 2026-07-14
updated: 2026-07-14
estimate_hours: 0.44
started: 2026-07-14T17:20:20-07:00
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

## Spec

In the close/milestone-close atlas gate: when `git diff --name-only <base>..HEAD`
contains no code paths (everything matches docs — `*.md`, `workshop/`,
`atlas/` itself), auto-satisfy the gate with a loud info line
("atlas gate: no code surface in window — auto-satisfied") instead of refusing.
Keep the refusal + `--no-atlas` hatch for windows that DO touch code. Define
the docs classifier in one place (the friction report's repo classifier in
`windowstat` used `*.md` + `workshop/ atlas/ docs/` — align).

## Done when

- A docs-only window closes without `--no-atlas` and without an atlas edit,
  with the info line stating why.
- A code-touching window still refuses without an atlas change (existing tests
  unchanged).
- The friction report's no-atlas counts exclude the auto-satisfied case (the
  info line must not match the bypass ACK signature — add the gatesig-catalog
  awareness if the wording overlaps).

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module    design=0.1  impl=0.2
item: milestone-review     design=0.0  impl=0.12
design-buffer: 0.15
total: 0.44
```

Σdesign 0.1 × 1.15 + Σimpl 0.32 × 1.0 = 0.44. One new arm in an existing gate
block + a pure predicate + tests + doc sweep; milestone-review = the close-time
boundary review. *Produced via
`brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

## Plan

Durable plan: `workshop/plans/000177-atlas-gate-no-code-autoskip-plan.md`.
Single review boundary (no Mx tags).

- [x] Task 1 — `hasCodePath` predicate + gate arm + gatesig no-collision test (TDD)
- [x] Task 2 — doc sweep (helptexts, AGENTS.base.md §5, atlas) + live verify + close

## Log

### 2026-07-14 — built (single pass)

`hasCodePath` (single docs classifier: *.md / workshop/ / atlas/ / docs/;
everything else conservatively code) + the auto-satisfy arm in computeClose
(both verbs; empty windows included). Code-touching windows keep the exact
refusal/bypass semantics. Gatesig no-collision test guards the #172
instrument. Live-verified both arms in a hermetic repo via milestone-close
--dry-run. Docs swept (helptexts, AGENTS.base.md §5, atlas). go
build/vet/test green.

### 2026-07-14

Filed from #172 M4 follow-on discussion (operator-approved).
