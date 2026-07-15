---
id: 000102
status: done
deps: []
github_issue:
target: base-layer-mechanics
created: 2026-06-15
updated: 2026-06-16
estimate_hours:
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

## Spec

Make menu discovery intent-driven, mirroring `LowerSkillSymlinks`: `GatherSkills`
should derive each layer's skill source dirs from its `skill <source-dir>`
manifest intents (not a hardcoded pair), so the menu and the symlinks compose the
SAME skill set. ariadne's `skill construct/local` / `skill construct/adapted`
rows keep its current menu unchanged; nous gains a `skill construct/skills`
(or equivalent) intent so its skills appear in both renderings.

Open design sub-questions (resolve in the plan):
- **Prefixing.** Local skills get the layer's `localPrefix` (xx-); adapted are
  bare. A third source dir (`construct/skills`) needs a prefix rule — bare?
  layer-named (`nous-`)? This overlaps [[000101-...]] (no-prefix lowering when
  localPrefix is empty); decide the two together. nous's skills are already named
  `nous-tools`/`nous-resolve` (bare dir names), so bare lowering is the natural fit.
- **Migration.** Once `skill construct/skills` exists, nous's plain
  `symlink construct/skills/<s> .claude/skills/<s>` rows become redundant (the
  skill intent lowers the same symlinks). Drop them in the same change.
- **Whether to keep two source dirs vs one.** Could fold construct/skills into
  construct/local with a per-skill prefix opt-out — but that's a bigger model
  change; the minimal fix is making GatherSkills honor the intents it's given.

## Done when

- `weave skills` (and a codex/agy compile's `## Skills` menu) includes a layer's
  skills wherever its manifest's `skill` intents point — verified on nous
  (nous-tools + nous-resolve appear in the menu).
- The menu and the `.claude/skills` symlink set are derived from the SAME
  per-layer source dirs (one discovery path; ARCH-DRY).
- nous's redundant `symlink construct/skills/*` rows are replaced by the skill
  intent.

## Plan

> Folded into #104 (skill-system v2) — these items were DELIVERED THERE, not worked
> independently in #102. Ticked as delivered (work + evidence live in #104's close;
> see this issue's Log for the subsumption note).

- [x] Make `GatherSkills` honor `skill <source-dir>` intents (reuse the
      `LowerSkillSymlinks` source-dir derivation; one shared helper). *(via #104 M1 — GatherSkills is intent-driven; LowerSkillSymlinks deleted; one shared SelectVisible feeds menu + symlinks)*
- [x] Decide + implement the prefix rule for a non-local/non-adapted source dir
      (coordinate with #101). *(via #104 M2 — skillPrefix: config.json localPrefix else repo-name basename; construct/adapted stays bare)*
- [x] Migrate nous: `skill construct/skills` intent; drop the plain symlink rows. *(via #104 M3 — nous → one `skill construct/local` export intent; plain symlink rows dropped)*
- [x] Test: a layer with a `skill <dir>` outside construct/local|adapted appears
      in both the menu and the .claude/skills symlinks. *(via #104 — TestGatherSkills_IntentDrivenVisibilityAndNonStandardDir + live xx-construct lowering)*

## Log


- 2026-06-16: closed — Subsumed by #104 — M1 made GatherSkills intent-driven (the menu reads each layer's `skill <dir>` intents, not hardcoded construct/local|adapted); M3 migrated nous to a `skill construct/local` intent so nous-tools/nous-resolve now appear in `weave skills` + serve via `weave skill`. Its 4 plan items were delivered under #104 (not ticked here) — folded in, no independent window.; review verdict: not-run
### 2026-06-15

- Filed during #95 M5 nous cutover. The claude cutover is NOT blocked by this
  (nous's skills lower correctly for claude via the plain symlink rows; brain
  inherits them through nous's exported symlink intents). This is a codex/agy +
  menu-composition gap. Deferred so the M5 cutover stays scoped; the critical
  export bug found on the same pass (WriteFile clobber-through-symlink) was fixed
  in ariadne a1d71bf.

- **Subsumed by ariadne#104** (skill-system v2): the intent-blind menu (A2) is one
  facet of the three-disagreeing-mechanisms root cause. Keep as a sub-tracker; close
  folded-in when #104 lands.
