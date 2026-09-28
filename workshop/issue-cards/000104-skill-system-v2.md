---
id: '000104'
status: done
created: 2026-06-15
updated: 2026-06-16
estimate_hours: 15
actual_hours: 3.69
---

# skill-system v2 — unified, visibility-aware, target-independent skill composition

## Problem

Skills have **three discovery mechanisms that disagree, and none consult
visibility** (confirmed: `grep` finds no `Visibility`/`Selected`/leaf check in
`walk/skills.go`, `walk/skill_symlinks.go`, `skill/skill.go`):

1. **Intent-driven symlink lowering** (`LowerSkillSymlinks`) — reads each layer's
   `skill <dir>` rows; produces the claude `.claude/skills/<name>` links.
2. **Dir-hardcoded menu/serve** (`GatherSkills`) — IGNORES intents; scans
   `<layer>/construct/local`+`construct/adapted` for *every* layer; feeds the
   codex/agy `## Skills` menu + `weave skills`/`weave skill`.
3. **Ad-hoc plain symlinks** (nous's `symlink construct/skills/X .claude/skills/X`)
   — bypasses the skill subsystem entirely; just a file-op.

The composition-algebra target asserts `skills(R) = ⋃ export-skills(Lᵢ) ∪
internal-skills(Lₙ)` at "clarity HIGH", but that formula was generalized from the
PROSE fix (#99) by analogy and never built or tested for skills — only the claude
target is ever exercised, and claude routes around every gap. This issue is the
real skill subsystem: declare → identify → compose → lower (per harness) → serve
→ inherit. See target `skill-system` for the invariant; this issue is its build.
