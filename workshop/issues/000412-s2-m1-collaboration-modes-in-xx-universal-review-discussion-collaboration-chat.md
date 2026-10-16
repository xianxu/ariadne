---
id: 000412
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: 'bd4617a0cd42d8dff6863449893376f7c0f061bf' # card fields mirrored from issue-cards; edit via sdlc
---

# S2 M1: collaboration modes in xx-universal-review (discussion, collaboration, chat)

## Problem

One review mode set (`process-instruction:`) governs a whole document, and every 🤖 question is answered in place. The
operator wants each section to say where the answer to an ask goes. The final design is ops
`teams/devinfra/projects/universal-agent-input-interface/design.side/01-s2-section-modes.md` (ops db130ac0), with
its side docs 02 to 06; design owner parley.nvim:4. This is S2 milestone 1 on the ariadne side. M2 (the per-type
audit) is a later issue.

## Spec

The shared skill `xx-universal-review` (`construct/local/universal-review/`) gains **collaboration modes**,
frontmatter `collaboration-mode:`, a separate axis from the review modes in `process-instruction:`:

- **discussion** (the default, side docs included): an imperative becomes an edit; an instruction that only acts
  (approve, file, tell) is honored, and an approval in a side doc folds the proposal into the parent through the
  parent's own write; a question opens a linked side doc in the document's flat `<name>.side/` folder.
- **collaboration**: today's round, answers in place.
- **chat**: each list item is an ask, answered by a dated answer doc in `<name>.side/`, linked from the item.

Deliverables:
- `collaboration-modes.md`: the frontmatter, heading matching and its error cases, the three protocols, the
  instructions-that-act table, the side-doc creation contract (path, names, frontmatter, body, one write per round,
  orphans), and the state table (design §5).
- `round.md` §4: resolve the collaboration modes and route each ask, pointing at `collaboration-modes.md`.
- `examples/`: one before/after document per mode.
- `pkg/reviewexport`: `collaboration-modes.md` joins `export.json`.
- `tl-lead`: the session file declares `collaboration-mode: {chat: ["Operator instructions"]}`; `tl-lead --help`
  cites the skill's chat mode instead of restating the Operator-instructions rules; answer docs go to
  `tl-status-<operator>.side/`. The TL keeps writing through `tl-lead status`.
- The atlas page `atlas/workflow/universal-review.md`.

## Done when

- On a real document, a 🤖 imperative becomes an edit; a 🤖 question becomes a linked side doc, where a follow-up
  question branches into a further side doc in the same folder and an "ok" folds the proposal into the parent; and a
  chat section's new item gets a linked, dated answer doc.
- The session file runs on `collaboration-mode:` with no change in what the operator sees in Operator instructions.
- `export.json` lists `collaboration-modes.md`, and `TestExportCurrent` passes.
- Every row of the design's state table is covered by an example or a check in the plan's test list.

## Plan

- [ ]

## Log

### 2026-10-16
