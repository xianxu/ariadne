---
id: 000444
status: open
deps: [ariadne#437]
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'a7c1bb425fc5cbcabf3c4a7f0e1a2a1b6c24638c' # card fields mirrored from issue-cards; edit via sdlc
---

# review convention: one turn per marker; a longer exchange moves to a side doc

## Problem

The operator finds marker chains confusing: 🤖[note]{reply}[reply]{reply} grows on one line, and parley's concealment of turn bodies left near-empty lines with lone brackets (operator, 2026-10-18, with screenshots in ops' session file). The operator's decision: a marker carries one turn, and a longer exchange moves to a side doc.

## Spec

Change the review convention (construct/local/fix/review-convention.md and the xx-fix / xx-universal-review skills): a marker is one turn, `🤖[operator note]` or `🤖{agent note}` (plus the reference forms `<X>` and `~X~` with at most one turn). An agent never appends a second turn to a marker. When an answer needs discussion beyond one turn, the agent answers by opening a side doc (S3: `→ [discussion](path)` replacing the marker) instead of chaining. Existing chains in documents stay readable: the agent resolves them in its next round, either answering in place or moving them to a side doc. Coordinate with ariadne#437 (the same skills move onto apply for S4): land after #437.

## Done when

- The convention and both skills state the one-turn rule and the side-doc escalation, with examples.
- The skill tests (or fixtures) refuse an agent edit that appends a second turn to a marker.
- parley's marker parser keeps reading old chains (no breakage), confirmed with parley.

## Plan

- [ ]

## Log

### 2026-10-18
