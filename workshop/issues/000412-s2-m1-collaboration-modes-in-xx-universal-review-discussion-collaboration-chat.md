---
id: 000412
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours: 4.33
card_mirror: '345ce7efbb3b365627b38cbc94769dc9faead8d7' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T15:42:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
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

Durable plan: [workshop/plans/000412-s2-collaboration-modes-plan.md](../plans/000412-s2-collaboration-modes-plan.md). Tasks 1 to 9 there, one boundary (no Mx).

- [ ] 1. collaboration-modes.md: the protocol (frontmatter, sections, the three modes, instructions that act, side docs, the state table)
- [ ] 2. round.md §4 routes by collaboration mode; §5 opening, §6 reply, SKILL.md pointer, the adapter's side-doc paragraph
- [ ] 3. examples/: ten cases, each with before/ and after/ trees and a case.md
- [ ] 4. examples_test.go: the side-doc contract checker (TDD against a broken copy) and nextSideDocName
- [ ] 5. export: the collaboration-modes member at Version 1, export.json regenerated
- [ ] 6. tl-lead help: the session file's collaboration-mode line, chat mode cited, answer docs in tl-status-<op>.side/; topic tests
- [ ] 7. atlas: universal-review.md (and tl-skills.md)
- [ ] 8. live check: rounds on the examples' before.md copies, and one tracked round through Couch
- [ ] 9. full suite via sdlc local-check, then sdlc close

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec              design=1.0 impl=0.04
item: skill-or-dispatcher     design=0.4 impl=0.2
item: skill-or-dispatcher     design=0.3 impl=0.2
item: smaller-go-module       design=0.2 impl=0.2
item: smaller-go-module       design=0.1 impl=0.1
item: atlas-docs              design=0.1 impl=0.08
item: ux-rename-iteration     design=0.3 impl=0.1
item: real-api-discovery      design=0.0 impl=0.25
item: milestone-review        design=0.0 impl=0.2
item: milestone-review        design=0.0 impl=0.2
design-buffer: 0.15
total: 4.33
```

- issue-spec: the S2 brainstorm is done (ops design, operator-decided); this is the issue, the plan and two plan-quality rounds, spent mostly before the code.
- skill-or-dispatcher ×2: collaboration-modes.md with round.md §4, SKILL.md and the adapter (design ×0.2-ish, the design doc pre-resolves the decisions); the ten example cases, written to the test list.
- smaller-go-module ×2: the examples checker with nextSideDocName and the export member; tl-lead help and its topic and link tests.
- ux-rename-iteration: the rounds with the design owner (O1 to O3 so far).
- real-api-discovery: the live check, real rounds on every case plus one tracked round through Couch.
- milestone-review ×2: the close review, expected to take two rounds.

## Log

### 2026-10-16

- Dispatched by the TL (ops:0, message 1cdf2024): S2 M1, the ariadne side; the design is ops db130ac0 01-s2-section-modes.md plus side docs 02 to 06. No existing issue, so this one was filed.
- Asked the design owner parley.nvim:4 four points (56ce9462). Answered (971f4119): export role `collaboration-modes` at Version 1 (parley's #319 sync is strict about both role and version, so a bump spares nothing); examples not exported; `process-instruction:` copied into a side doc only when the parent has one; a chat follow-up is a diff hunk overlapping the item's extent minus its link. All folded into the plan.
- Rejected a Go router (`sdlc review sections`): parley's standalone prompt has no binary, so the prose stays the one source (ARCH-DRY).
- Plan approved by the TL (19edb7f1): full flow pinned (it changes how every agent answers every document); O1 decided, a 🤖[fold in] by a link is an imperative and becomes an edit in M1, marking the link folded; the session file line and moving the QA answers are the TL's after landing, with the exact line in the landing report.
- parley.nvim:4 confirmed O1 (e2492347, folded into ops 7d0b425d): one fold-in with two triggers (an approval in the side doc, a 🤖[fold in] by the link in the parent), both marking the link folded. With no settled proposal to fold, the agent asks with a one-line clarifying turn, as for an ambiguous ok.
- plan-quality round 1: FAILURE on PQ-1 (the examples layout cannot hold side docs before and after a round). Fixed by per-case before/ and after/ trees, swept across the plan; the three Minors fixed (gitignore in the landing report, old qa/ links followed, Non-goals). O2 (deleted chat link against reuse) and O3 (a fold-in under propose-strike) sent to parley.nvim:4.
