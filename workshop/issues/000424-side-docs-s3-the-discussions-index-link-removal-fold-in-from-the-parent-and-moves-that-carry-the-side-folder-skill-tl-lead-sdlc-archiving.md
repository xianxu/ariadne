---
id: 000424
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours: 4.13
collaboration-mode: collaboration
card_mirror: '06f2f44606833961f144cddea840a623f3f1ded8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T03:02:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# side docs S3: the discussions index, link removal, fold-in from the parent, and moves that carry the .side folder (skill, tl-lead, sdlc archiving)

## Problem

S2 of the universal agent input interface (ops `teams/devinfra/projects/universal-agent-input-interface/design.md`) built how a side doc starts. S3 covers the rest of its life, and the operator approved its design, now final: ops `design.side/07-s3-side-docs.md` (cf0f0ebf), with threads 08 and 09. ariadne holds three parts of it: the `xx-universal-review` skill, `tl-lead status`, and sdlc's archiving. parley's mover is parley.nvim#373.

## Spec

From 07, sections 3 to 7 and 10. Decisions (operator, 10-16):
- **One attachment, two references.** The attachment is the side doc's `parent:` and `marker:`; the references are the link in the parent's text (`→ [discussion](…)` or `→ [answer](…)`) and the parent's `discussions:` index entry.
- **The index rule is one-way.** Deleting a link drops its index entry on the agent's next round; removing an entry means reviewed and leaves the link. The agent adds an entry back only when it next writes something for the operator in that side doc. The agent never removes a link itself.
- **The index lives in the immediate parent**, recursively. The entry rides the parent's single compare-and-swap write, in the same round as the link.
- **No closing.** A side doc is never deleted, marked or locked. S2's reuse by `parent:` and `marker:` is unchanged.
- **Needs you stays hand-written** free text, not a view of the index.
- **Fold-in is triggered in the parent** by a `🤖[fold in]` at or beside the link. The fold marks the link "(folded <date>)" and adds a closing note to the side doc. An approval inside a side doc only records a `Decided (operator): …` line there; this replaces S2's approval-folds rule (row 7 and its row in section 3). One level per fold. Answer docs never fold into their chat item.
- **Moves the agent makes** follow one contract: move the file and its `.side/` folder together; rewrite the side docs' `parent:` and backlinks; on a name change, rewrite the parent's own links and index; run `couch --session-sync rename OLD NEW` when the file is registered. When the target folder name is taken, the orphan folder takes the suffix (`spec_1.side/`, then `_2`, …), and the moving document keeps `spec.side/`.
- **One fixture set:** the move fixtures are shared with parley's projection specs and travel in `export.json`.

The work, by place:
- **`xx-universal-review`:** `collaboration-modes.md` gains 07's sections 3 to 7 and rows 18 to 28. `examples/` gains: a deleted link dropping its entry, an entry added back after review, fold-in from the parent, a chain fold, and a name-change move. The examples checker gains two invariants: every index entry names a side doc whose `parent:` is that document, and the move fixtures match parley's.
- **tl-lead:** `tl-lead status` writes the session file's `discussions:` index and drops an entry when its item is deleted.
- **sdlc:** archiving an issue or plan to `workshop/history/` moves its `.side/` folder with it, under the same move contract.

## Done when

- The skill's `collaboration-modes.md` and examples carry sections 3 to 7 and rows 18 to 28, and the examples checker enforces both new invariants.
- `tl-lead status` writes and prunes the session file's index by the one-way rule, with a test.
- `sdlc` archiving an issue or plan that has a `.side/` folder moves the folder and keeps every link and index entry resolving, with a test.
- On a real document: a question opens a side doc and its index entry; the operator's removal of the entry sticks until the agent next writes there; deleting the link drops the entry; a `🤖[fold in]` by the link folds the outcome and marks the link.

## Plan

Durable plan: [workshop/plans/000424-…-plan.md](../plans/000424-side-docs-s3-the-discussions-index-link-removal-fold-in-from-the-parent-and-moves-that-carry-the-side-folder-skill-tl-lead-sdlc-archiving-plan.md).

- [ ] M1 — the skill: `pkg/sidedoc` (links, index), `collaboration-modes.md` (index, no closing, fold-in from the parent, moves; rows 21 to 31), six new examples and the existing ones updated, checker invariants, move fixtures exported (version 2)
- [ ] M2 — tl-lead: the session file's `discussions:` index (`instructions link` adds, `status submit` and `status index` reconcile) through `write --all`
- [ ] M3 — sdlc: the three archive paths move a document's `.side/` folder with it

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.3 impl=0.24
item: skill-or-dispatcher    design=0.3 impl=0.16
item: skill-or-dispatcher    design=0.2 impl=0.2
item: smaller-go-module      design=0.1 impl=0.2
item: smaller-go-module      design=0.05 impl=0.12
item: smaller-go-module      design=0.2 impl=0.2
item: greenfield-go-module   design=0.5 impl=0.32
item: atlas-docs             design=0.1 impl=0.08
item: milestone-review       design=0.0 impl=0.2
item: milestone-review       design=0.0 impl=0.2
item: milestone-review       design=0.0 impl=0.2
design-buffer: 0.15
total: 4.13
```

Items in order: M1 `pkg/sidedoc` (links, index; greenfield, design mostly settled by 07); M1 `collaboration-modes.md` contract; M1 the six examples and the updates to existing ones; M1 checker invariants and `projectMove`; M1 export version 2; M2 tl-lead index over the existing seam; M3 the `.side/` counterparts of the four asset archive stages across paths (a) to (f), the riskiest part; atlas and help; three milestone reviews. impl at 40% of the v2 table (v3.1); design buffer 15% for the thorough plan.

> *Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log

### 2026-10-17
- Filed from the S3 design (ops 07, final at cf0f0ebf); parley's half is parley.nvim#373 (filed by the TL from this slot's draft).
- 07's rows 18 to 28 become 21 to 31: the skill's states table already runs to 20.
- sdlc never renames an issue or plan file (`set-title` keeps the slug), and archiving keeps the base name, so sdlc's part is a same-name move of the `.side/` folder, with no link rewriting.
- parley's `scripts/sync-review-protocol.py` refuses a role it doesn't know and any version but 1, so exporting the move fixtures bumps the export to version 2, which parley.nvim#373 accepts. Parley's committed snapshot is unaffected until it syncs.
- Plan approved by the TL; plan-quality cleared in 2 rounds (PQ-1: M3 now follows the asset archive's four stages; the Minors fixed). Estimate 4.13h (v3.1). The estimate-quality judge flagged M2 (CAS and conformance tests) and M3 (six archive paths) as thin and the smoke round as unitemized; kept as derived, to compare at close.
