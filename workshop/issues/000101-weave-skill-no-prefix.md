---
id: 000101
status: open
deps: []
github_issue:
target: weave-composition-algebra
created: 2026-06-15
updated: 2026-06-15
estimate_hours: 1
---

# weave: support no-prefix skill lowering (bare names when localPrefix is empty)

## Problem

The local-skill namespace prefix is configured in a layer's `construct/config.json` as `{"localPrefix": "xx-"}`. weave's skill lowering (`walk/skill_symlinks.go`, `walk/skills.go`) reads `localPrefix(fs, layer.Path)` and lowers `construct/local/<dir>` → `.claude/skills/<prefix><dir>` (and `construct/adapted/<dir>` → bare). The default is `"xx-"`, mirroring the retired `sync-local-skills.sh`'s `PREFIX="${PREFIX:-xx-}"` fallback — which treats an **empty** prefix the same as **absent** (both → `xx-`). So there is currently **no way to lower local skills with bare names** (`fix`, `pensive`, `voice-apply`) instead of prefixed (`xx-fix`, …).

## Spec

Let a layer opt out of the prefix: an **explicit** empty/null `localPrefix` in `construct/config.json` → lower local skills with **bare** names (`construct/local/fix` → `.claude/skills/fix`), while an **absent** `localPrefix` keeps the `xx-` default (backward-compatible). I.e. distinguish "field absent → default `xx-`" from "field present-and-empty → no prefix" (so the bash `${PREFIX:-xx-}` collapse is the thing to fix). Applies uniformly to both faces: the `.claude/skills` symlink lowering (claude target) AND the M3 menu naming, so a skill's served name matches its lowered name.

Decide the exact config encoding (`"localPrefix": ""` vs `null` vs a separate `"prefixLocal": false`) when implementing — `""`-means-bare is the lightest.

## Done when

- A layer with `construct/config.json` `localPrefix: ""` lowers its local skills to bare `.claude/skills/<dir>` and serves them under bare names; a layer with the field absent still gets `xx-`.
- Both the symlink lowering and the menu use the same name.

## Plan

- [ ]

## Log

### 2026-06-15
- Filed from the parley tart-prep discussion: operator wants bare-name lowering as an option. Confirmed the prefix lives in `construct/config.json` `localPrefix` (default `xx-`); today empty collapses to `xx-` (bash fallback parity), so bare names aren't reachable.

- **Subsumed by ariadne#104** (skill-system v2): the no-prefix lowering is one
  facet of the unified prefix/identity model (C1/C2). Keep as a sub-tracker; close
  folded-in when #104 lands.
