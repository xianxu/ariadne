---
id: 000424
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'e05d3017ca5d9a78309d89fe6e80362382ce2789' # card fields mirrored from issue-cards; edit via sdlc
---

# side docs S3: the discussions index, link removal, fold-in from the parent, and moves that carry the .side folder (skill, tl-lead, sdlc archiving)

## Problem

S2 of the universal agent input interface (ops `teams/devinfra/projects/universal-agent-input-interface/design.md`) built how a side doc starts. S3 covers the rest of its life, and the operator approved its design, now final: ops `design.side/07-s3-side-docs.md` (cf0f0ebf), with threads 08 and 09. ariadne holds three parts of it: the `xx-universal-review` skill, `tl-lead status`, and sdlc's archiving. parley's mover is a separate parley issue.

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

- [ ]

## Log

### 2026-10-17
