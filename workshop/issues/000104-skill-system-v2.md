---
id: 000104
status: done
deps: []
github_issue:
target: skill-system
created: 2026-06-15
updated: 2026-06-16
estimate_hours: 15
actual_hours: 3.69
---

# skill-system v2 — unified, visibility-aware, target-independent skill composition

> **Not to be confused with parley#128** ("skill system redesign: declarative
> modules over one engine"). #128 is parley's *runtime* skill execution (the Lua
> chat-loop engine, `read_skill`, `propose_edits`, per-turn manifest assembly).
> THIS issue is ariadne/weave's *build-time* skill **composition** across the layer
> DAG (the `skill <dir>` intents, the `.claude/skills` lowering + the menu,
> visibility, prefixes). Orthogonal layers, same word — they don't interact.

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

## Spec

The gap inventory (surfaced on the #95 nous/brain cutover):

**A. Declaration / discovery**
- **A1 — three mechanisms, no single source of truth.** Want ONE discovery →
  `SkillIndex` → two lowerings (symlinks | menu). Today the symlink path doesn't
  use the index at all.
- **A2 — menu is intent-blind** (was #102, subsumed here). `GatherSkills`
  hardcodes the two dirs → nous's `construct/skills` invisible to menu + serve.
- **A3 — plain-symlink skills can't do per-target lowering.** A `symlink …
  .claude/skills/X` is inherently claude-shaped; never a menu entry or served
  body → nous's skills are claude-only. Declaration mechanism (not the skill's
  nature) decides target-independence — backwards.
- **A4 — the two paths see different layer sets for the same skills.**
  `LowerSkillSymlinks` only processes layers that DECLARE `skill` rows (just
  ariadne); `GatherSkills` scans every layer's `construct/local` including the
  derivatives' symlinked copies → double-scans ariadne's skills (masked by
  `skill.Build` dedup).

**B. Visibility (the new capability)**
- **B1 — internal skills unimplemented.** No skill path gates on export/internal
  or leaf-position; ariadne's "construct skill is ariadne-internal" comment is
  unenforced. Fix: route skills through the SAME `intent.Selected`/`participates`
  filter prose uses (the algebra's `internal-skills(Lₙ)` term).
- **B2 — no suppression.** A derivative can't DROP an inherited exported skill
  (the target's "Not suppression" open question). Adjacent missing op.

**C. Identity / prefix** (subsumes #101)
- **C1 — prefix hardcoded to `construct/local`.** `LowerSkillSymlinks` applies
  the prefix only when `sourceRel == "construct/local"`; any other dir → bare.
  nous's skills are bare BY ACCIDENT, not design.
- **C2 — per-layer prefix is blocked.** Prefix comes from the layer's
  `config.json localPrefix`, but every derivative's `config.json` is a SYMLINK to
  ariadne's (`xx-`). nous can't declare `nous-` without owning its config.json.
  So moving nous skills into `construct/local` would mis-name them `xx-tools`.
- **C3 — identity is dir-name-derived,** not frontmatter; couples on-disk layout
  to skill name; no cross-layer collision validation.

**D. Inheritance symlinks**
- **D1 — `construct/local`/`construct/adapted` are symlinks to ariadne** in every
  derivative (old inheritance trick). Causes the A4 double-scan AND physically
  blocks a layer from owning its own `construct/local`. Right model: per-layer
  REAL dirs holding only that layer's own skills; weave's layer-walk aggregates.

**E. Serving**
- **E1 — `weave skill <name>` can't serve plain-symlink skills.** `weave skill
  nous-tools` → not in the index → "unknown skill"; the agent-agnostic serving
  path is broken for exactly nous's skills.

**F. Adjacent composition gaps (linked follow-ups, not core)**
- **F1 — file-op collision accumulation documented but NOT built.** The algebra
  target claims "conflict-accumulating error-monad"; `grep` finds no collision/
  warning logic in `plan/` — it's silent last-writer-wins. (Own issue.)
- **F2 — settings DAG fold** (#97). **F3 — settings round-trip** (#100). **F4 —
  one-target-per-invocation** (claude symlinks vs codex menu can't coexist in one
  AGENTS.md).

## Decisions (operator, 2026-06-15)

1. **Visibility = per-dir convention.** Three conventional dirs encode it:
   `construct/adapted` (external source → export), `construct/local` (locally
   constructed → export), **`construct/skill` (private to the layer → internal)**.
   The `construct` skill (already at `construct/skill/`) is the internal exemplar —
   internal-by-declaration, not internal-by-location. Per-skill frontmatter is the
   fallback only if mixed visibility within one dir becomes common.
2. **Identity = dir-name** (keep; frontmatter `name:` deferred). (C3)
3. **Kill the whole-dir inheritance symlinks** (D1) — drop `<derivative>/construct/
   {local,adapted} → ../ariadne/...`; the layer-walk reads each ancestor's dirs
   directly. KEEP the per-skill `.claude/skills/<name>` → source-layer links (the
   claude lowering — they straighten to point directly at the owning layer).
4. **Per-layer prefix via own `config.json`** (C2) — un-symlink `config.json` in
   derivatives; **default prefix = the layer's repo name** (`nous-`, `brain-`,
   `metis-`); ariadne overrides to `xx-`.
5. **Migrate nous `construct/skills` → `construct/local`** as a `skill` intent;
   drop the plain symlink rows (A3). They are EXPORT skills (consumed by descendant
   brains), so `construct/local` (the export dir) is their correct home — the
   off-convention `construct/skills` name was the mistake. `construct/local/tools`
   + `nous-` → `nous-tools` (uniform naming). Note: can't be done cleanly before
   decisions 3+4 (kill the `construct/local`→ariadne symlink so nous can own the
   dir; give nous its own `nous-` prefix) — else it mis-prefixes to `xx-nous-tools`.

Still open: **suppression** (can a derivative drop an inherited export?) and the
**`construct/skill` layout** (one SKILL.md vs `<name>/SKILL.md`, to scan
uniformly). See target [[skill-system]].

## Done when

- ONE intent-driven discovery feeds BOTH lowerings (claude symlinks + codex/agy
  menu); a skill declared by `skill <dir>` appears in both, wherever it lives.
- Skills honor export/internal + leaf-position (an internal skill stays in its
  layer; verified with a fixture where an ancestor's internal skill does NOT
  reach a consumer, and the leaf's internal DOES).
- nous's skills appear in `weave skills`, serve via `weave skill nous-tools`, and
  lower for both targets — declared as a `skill` intent, not plain symlinks.
- Per-layer prefix works (nous → `nous-`); no double-prefix.
- A golden/fixture binds the skill formula to the code (so the target can no
  longer claim HIGH without proof).

## Plan

Detail in `workshop/plans/000104-skill-system-v2-plan.md` (M1 task-detailed; M2–M3
sketched, per the #128 convention).

- [x] M1 — unified intent-driven discovery + visibility: ONE `GatherSkills`
      (intent-driven, carries `Visibility`+`LayerIndex`) → pure `skill.SelectVisible`
      (reuses `intent.Selected`, ARCH-DRY) → `skill.Build` (menu) + new pure
      `plan.SkillSymlinks` (claude); DELETE `walk.LowerSkillSymlinks`. Behavior-
      preserving for ariadne (closes §A1/A2/A4 + §B1).
- [x] M2 — per-layer prefix DEFAULT (`config.json localPrefix` else repo-name;
      ariadne pins `xx-`) + the internal-skill 𝒜(R) CAPABILITY proven end-to-end via
      `walk→buildSkillIndex` (closes §C1; the §C2 own-`config.json` MIGRATION + the
      real `internal skill construct/skill` are M3). Behavior-preserving for ariadne.
- [x] M3 — cross-repo migration: kill the whole-dir `construct/{local,adapted}`→ariadne
      inheritance symlinks; **each derivative owns its `config.json`** (so the repo-name
      prefix activates); **declare `internal skill construct/skill` + the `construct/skill`
      layout/name** (Task C2 — applies M2's capability to the real construct skill);
      nous `construct/skills`→`construct/local`; re-weave all 10 repos (closes §C2/§D1/
      §A3/§E1 + the construct-skill internal exemplar). Retire #101 + #102 (folded in);
      lift the [[skill-system]] target's "DESIGN-ONLY" banner (invariants now test-bound).

## Log

### 2026-06-16
- 2026-06-16: closed — #104 done-when met across M1-M3: ONE intent-driven discovery (walk.GatherSkills) feeds BOTH lowerings (claude .claude/skills symlinks + codex/agy menu) from the SAME SelectVisible set; skills honor export/internal + leaf-position (TestBuildSkillIndexExcludesAncestorInternalSkill + live: xx-construct present on ariadne self-walk, absent in all 9 derivatives); nous skills appear in `weave skills` + serve via `weave skill nous-tools`, declared as a skill intent not plain symlinks; per-layer prefix works (nous → nous-, no double-prefix); formula test-bound (TestSelectVisible/TestGatherSkills_*) + live-verified across all 10 re-wove repos (verify-complete 0-unplanned, ancestors byte-pristine). All 3 milestones reviewed (M1 SHIP, M2/M3 FIX-THEN-SHIP).; review verdict: SHIP
- 2026-06-16: closed M3 — All 10 ariadne-styled repos re-wove + verify-complete 0-unplanned + ancestors byte-pristine; whole-dir inheritance symlinks dropped (skills flow via the layer walk — pair/nous/brains confirmed .claude/skills/xx-* + nous-* resolve to the owning layer), xx-construct lowered leaf-internal in ariadne + absent in all 9 derivatives, nous-tools/nous-resolve now menu-listed + servable + inherited by the brains via the layer; sdlc actual owner-resolves active-time-v3.py (live: measured 0.30h in pair with no local construct/local); full weave + sdlc suites + go vet + gofmt green; review verdict: FIX-THEN-SHIP — no Critical; the one Important (doc-drift in the construct skill BODY: the Self-sync rule + construct/adapted-inheritance + .claude/skills/construct location) reconciled to the post-M3 layer-walk model before crossing the boundary (auto-dispatch hit a transient socket error → re-ran `sdlc judge milestone-review` for the real verdict)

### 2026-06-15
- 2026-06-15: closed M2 — skillPrefix defaults to the layer repo-name (config.json override wins; ariadne pinned xx- via its config — golden MATCH 25/0 unchanged); internal-skill 𝒜(R) proven end-to-end via walk→buildSkillIndex (ancestor-internal excluded from a consumer, present on its own self-walk); full weave suite + vet + gofmt clean; review verdict: FIX-THEN-SHIP
- 2026-06-15: closed M1 — one intent-driven GatherSkills (carries Visibility+LayerIndex) → pure SelectVisible (reuses intent.Selected) → {Build menu, plan.SkillSymlinks claude} from the IDENTICAL selected set; walk.LowerSkillSymlinks deleted; ariadne golden MATCH 25/0 UNEXPECTED unchanged (behavior-preserving); full weave suite + vet + gofmt clean; review verdict: SHIP

- Filed from the #95 cutover gap analysis (nous/brain). Subsumes #101 (no-prefix)
  and #102 (menu intent-blind) — close those as folded-in once this lands, or
  keep as sub-trackers. F1 (file-op collision) deserves its own issue.
- Meta-lesson (see workshop/lessons.md): the composition-algebra target asserted
  the skill formula at "clarity HIGH" by analogy to prose, never built/tested it;
  the target gave false confidence and HID the gap. Demoting that slice + creating
  the `skill-system` subsystem target.
