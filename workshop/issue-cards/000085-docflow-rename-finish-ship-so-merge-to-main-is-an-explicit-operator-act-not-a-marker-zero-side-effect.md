---
id: '000085'
status: done
created: 2026-06-04
updated: 2026-06-04
estimate_hours: 0.5
actual_hours: 0.5
---

# docflow: rename finish→ship so merge-to-main is an explicit operator act, not a marker-zero side effect

## Problem

The merge-to-main lives in `docflow finish`. Audit (this session) confirms the
*script* already gates the merge behind an explicit verb — marker-zero is a
**guard** (`cmd_finish` refuses while 🤖 > 0), not a **trigger**; nothing
auto-merges. But two couplings push the agent to treat marker-zero as the merge
cue:

1. **Verb name.** `finish` reads as "the review is finished" — which an agent
   naturally fires the moment markers hit zero. The name conflates *the review
   conversation is resolved* with *land this on main*.
2. **Skill prose.** `construct/local/fix/SKILL.md:109` ("a feedback session is
   **complete when no marker ending in `[]` remains**") sits adjacent to the
   finish instruction (`:128`), inviting the slide marker-zero → "complete" →
   `finish` → merged.

"No 🤖 left" answers "is the conversation done?" — it does **not** answer "do I
want this on main now?" The operator may want more rounds, more content, or to
sit on a marker-clean branch. Merge-to-main should be a *named, deliberate*
operator act.
