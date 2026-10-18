---
id: 000444
status: working
deps: [ariadne#437]
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '31e5e7ed77735ffc5729cfef5d12bbc0e0767665' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T14:37:53-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "9a740603", done: "4223225d"}
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

- [x] Record the single-turn emission rule and legacy-reading rule in the canonical convention.
- [x] Update xx-fix and xx-universal-review so clarification, disagreement, and continued exchange use linked side docs rather than appended marker turns.
- [x] Add before/after scenarios for an ambiguous ask, an ambiguous approval in a side doc, and a legacy chain; verify the agent's result never adds a second turn.
- [x] Confirm parley's existing parser still reads legacy chains, run targeted tests and the recorded suite, then close and land.

## Log

### 2026-10-18

- Started on the landed #437 edit path. The current convention permits indefinite chains, and three agent instructions still append a clarification turn; the new rule changes emission while retaining legacy reading.
- TDD: `TestCheckCaseCatches` first failed on an appended second turn; after adding the fixture guard, `TestExamples` failed on the two old append outputs. Those examples now open indexed side docs. A new ready legacy-chain case resolves by its final human turn; an unchanged legacy chain remains allowed in the fixture test.
- `go test ./pkg/reviewexport -count=1` passed. Parley's `make test-spec SPEC=modes/review` passed, including its three-turn parser case (`tests/unit/review_spec.lua`). `make -k` returned the Makefile help successfully.
- Live scratch Pair apply under `presentation: propose` returned `Code: ok` and a derived round: the ready `🤖[add the numbers]` became a direct `→ [discussion](note.side/01-numbers.md)` link while the matching `discussions:` entry landed in the same batch; readback and `pair review round` reported the new hash and `Current: true`.
- Recorded full suite on `973e8a8825a2c94c0945601bfafbe610a903042d`: 12 `cmd/sdlc` shards and 70 other packages passed, zero failures; wall 550.2 seconds after 1,333 seconds in the machine queue. Log sha256 begins `032375436979`.
