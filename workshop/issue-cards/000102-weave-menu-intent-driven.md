---
id: '000102'
status: done
created: 2026-06-15
updated: 2026-06-16
actual_hours: N/A
---

# weave skill menu discovery is intent-blind (construct/skills skills missing from the menu)

## Problem

weave has TWO skill renderings of the same skill set, and they disagree on
WHERE skills live:

- **Symlink lowering** (`walk.LowerSkillSymlinks`, the claude `.claude/skills/`
  backend) is INTENT-DRIVEN: it iterates each layer's `skill <source-dir>`
  manifest intents and scans `in.Source`. So a layer can declare skills anywhere.
- **Menu discovery** (`walk.GatherSkills`, the codex/agy `## Skills` menu backend
  + the `weave skills`/`weave skill` server) is NOT: it HARDCODES scanning
  `construct/local` + `construct/adapted` for every layer and ignores the
  manifest intents entirely.

Surfaced on the nous M5 cutover (#95): nous's own skills live in
`construct/skills/` (nous-tools, nous-resolve), declared via plain
`symlink construct/skills/<s> .claude/skills/<s>` rows. After `make weave`:

- `.claude/skills/nous-tools` + `nous-resolve` are present and correct (the
  claude target works — they lower via plan.Plan's Symlink case), so the **claude
  cutover is unaffected**; and a consumer (brain) inherits them via nous's
  exported symlink rows.
- but `weave skills` lists 23 entries with **ZERO nous-*** — nous's skills are
  invisible to the menu, so a codex/agy compile (and any downstream menu that
  should surface nous's skills) silently drops them.

This violates the composition-algebra target's skill formula
(`skills(R) = ⋃ᵢ export-skills(Lᵢ) ∪ internal-skills(Lₙ)`, "composition is
target-independent; only the lowering differs"): today the menu lowering and the
symlink lowering see DIFFERENT operand sets.
