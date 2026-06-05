---
id: 000085
status: working
deps: []
github_issue:
created: 2026-06-04
updated: 2026-06-04
estimate_hours: 0.5
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

## Spec

Rename the merge verb `finish` → `ship` (the merge IS the ship-to-main act).
Keep `finish` as a **deprecated alias** → `cmd_ship` (prints a one-line notice)
so downstream repos and muscle memory don't break (ARCH-DRY: one implementation,
two names). Behavior is otherwise byte-identical: clean-tree guard, 🤖 guard
(refuse unless `--force`), `--no-ff` merge to base, delete branch, drop meta dir.

Tighten the skill prose so marker-zero is explicitly *not* the ship cue: the
merge fires only on an explicit operator ship decision, never merely because the
markers cleared.

No new state, no `resolve` verb — `status` already reports marker-zero; adding a
verb that only asserts what `status` shows would violate Simplicity First.

## Done when

- `docflow ship` performs the merge; `docflow finish` still works (alias, with a
  deprecation notice) — both covered by `docflow.test.sh`.
- `docflow.sh` header VERBS, dispatch, usage, and the `--force` hint string all
  say `ship` (alias documented).
- `construct/local/fix/SKILL.md` renames the verb AND states marker-zero ≠ merge
  cue; `atlas/workflow/docflow.md` verb table updated.
- `scripts/docflow.test.sh` green.

## Plan

- [x] `docflow.sh`: `cmd_finish`→`cmd_ship`; dispatch `ship)` + `finish)` alias (warns → cmd_ship); header VERBS + usage + `--force` hint + unknown-verb error + HISTORY-MODEL comment reference `ship`.
- [x] `construct/local/fix/SKILL.md`: renamed verb at :118/:128; reworded :109 (marker-zero = conversation resolved, *not* a merge cue) + :128 (ship = explicit operator act).
- [x] `atlas/workflow/docflow.md`: all four `finish` refs (`:5/:33/:46/:59`) → `ship`; verb-table row gains the "not fired by marker-zero" note + alias line.
- [x] `docflow.test.sh`: primary assertions switched to `ship`; added an alias case (exits 0 / lands on main / warns deprecated). Full e2e green (`ALL PASS`).

## Log

### 2026-06-04
- Audit traced the only merge path to `cmd_finish` (explicit verb; marker-zero is a guard not a trigger). Fix is naming + prose, not control flow. Single-pass atomic — no Mx tags.
- Implemented: `ship` is the merge verb, `finish` a deprecated alias (warns, then `cmd_ship`). 5 files (docflow.sh, its test, SKILL.md, atlas, issue). e2e verified unsandboxed (sandbox can't `mktemp -d`, so the e2e tier SKIPs in-sandbox — ran with sandbox disabled): `ALL PASS`, incl. new alias coverage. `bash -n` clean.
- Downstream note: xianxu.dev consumes `docflow.sh`/`SKILL.md` via symlink into ariadne, so the rename is live there immediately — no manifest propagation step.
