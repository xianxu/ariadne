---
id: 000507
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '7ae53112ab5d4b8bd6302e0a3caa9817be4a0878' # card fields mirrored from issue-cards; edit via sdlc
---

# workingdoc datatype and tl-lead doc settle/status: settle side docs into tracked copies, re-settle to update, drift visible

## Problem

Side docs (the operator–TL co-edited pages beside the session file) have become the main design medium, but they're
untracked, so settled decisions can be lost, and projects can't link them durably. Operator, 10-20: make a
`workingdoc` type; copy a settled side doc into a tracked place; projects link it as background. Concern raised:
once copied, the copy freezes while discussion continues in the side doc.

## Spec

- `construct/datatype/workingdoc.md`: a living, co-edited page on one topic, with owner, topic and a
  `settled-into: <path> @ <commit>` pointer.
- `tl-lead doc settle <side-doc> [--to PATH]`: copy with review markers stripped to `workshop/workingdoc/<date>-<slug>.md`
  (default) or the given path (e.g. a project's `design.md`), commit it, stamp the side doc's pointer. **Re-settling
  overwrites and commits the copy**, so git history is the copy's version history.
- `tl-lead doc status`: side docs edited since their last settle, so drift is visible.

## Done when

- The datatype, both verbs and their tests exist; `tl-lead --help` documents settling.

## Plan

- [ ]

## Log

### 2026-10-20
