---
id: 000033
status: done
deps: []
created: 2026-05-25
updated: 2026-05-28
estimate_hours:
actual_hours: 1
---

# adaptation system: narrow to ariadne-self adaptation only

## Problem

The construct adaptation pipeline (`/construct adapt <source> --to <relpath>`) is built to adapt a single source plugin (e.g., superpowers) into N different target repos with N different intent transcripts. From `construct/skill/SKILL.md`:

> The same source can have different intents for different target repos — superpowers adapted for parley.nvim differs from superpowers adapted for a future project.

This is genuinely flexible. But ariadne *itself* is an opinionated stack on how to leverage AI in software development. The premise of an opinionated stack is that downstream consumers inherit the opinions, they don't redo them. So the multi-target flexibility may be more machinery than the actual use case warrants.

### What the code shows

I checked how things are actually wired:

- `construct/intents/superpowers/` contains two files: `ariadne.md` (target: `.`) and `parley.nvim.md` (target: `../parley.nvim`). So the multi-target capability is real and has two real instances.
- `construct/adapted/superpowers-*/` contains the rendered output of adapting superpowers *for ariadne* (the `ariadne.md` intent). This is what `/construct promote --to .` deposits.
- `construct/base.manifest` has `symlink construct/adapted` — meaning every ariadne-derivative repo (via `setup.sh`) gets a symlink (or vendored copy) of ariadne's `construct/adapted/` directory.
- In nous (`/Users/xianxu/workspace/nous/`), `construct/adapted/` exists as a real directory in vendor mode, byte-identical to `ariadne/construct/adapted/` (verified with `diff -r`). nous has *no* `construct/intents/` directory, no `construct/sources/` directory, no `construct/staging/` directory.
- nous's `.claude/skills/superpowers-brainstorming` → symlink to `../../construct/adapted/superpowers-brainstorming` → which (in symlink mode) would chain to ariadne's adapted version.

So the in-practice flow for ariadne derivatives is:

1. Ariadne adapts superpowers for itself (one intent: `intents/superpowers/ariadne.md`).
2. Ariadne promotes to `construct/adapted/`.
3. Derivatives inherit `construct/adapted/` via `base.manifest`, transitively getting ariadne's adaptation. They never run `/construct adapt` themselves.

Historically, `parley.nvim` was the one non-derivative case — it predates ariadne and originally had its own conventions. But parley.nvim has since adopted the ariadne layout (AGENTS.md, Makefile.workflow, construct/adapted, construct/local) and its `construct/adapted/` is now byte-identical to ariadne's (verified with `diff -rq`). So the separate `parley.nvim.md` intent is effectively dead — parley.nvim has converged to inheriting ariadne's adaptation just like nous does. There is no remaining non-derivative consumer.

### Decision

Drop multi-target. `/construct adapt <source>` (no `--to`) always adapts for ariadne. Intent file is `intents/<source>.md`, not `intents/<source>/<repo>.md`. Promote always goes to `.claude/skills/` (ariadne's own) and `construct/adapted/`. Derivatives inherit via `construct/adapted/` (already wired through `base.manifest`).

Why:
- Matches the "ariadne is opinionated" framing — derivatives shouldn't be re-deciding skill behavior, and in practice they don't (nous and parley.nvim both inherit verbatim).
- Removes scope/target/relpath axes from a skill where every existing call site converges to the same answer.
- Keeps `intents/<source>.md` as a single audit trail per source.

What this does *not* change:
- The vendoring/staging/promote/version-snapshot pipeline still exists, just with one target.
- The intent-as-transcript principle still holds.
- Derivative inheritance via `base.manifest` is unchanged.
- Local skills (`construct/local/`) and their symlink-based deployment are unrelated and untouched.

### Sub-questions resolved (2026-05-25)

- **`--to personal` scope:** drop. Removes the last knob; adapt always means "for ariadne."
- **`intents/superpowers/parley.nvim.md`:** delete. The intent file is superseded; the historical divergence is preserved in git history (`git log -- construct/intents/superpowers/parley.nvim.md`) if anyone needs to reconstruct it.

### parley.nvim cleanup (related work)

`parley.nvim/.claude/skills/superpowers-*` is currently a set of *real directories*, not symlinks. They were promoted via the old `/construct adapt superpowers --to ../parley.nvim` flow (Apr 25) and have since drifted: their content reflects the parley.nvim-specific intent (Visual Companion removed, etc.), while `parley.nvim/construct/adapted/superpowers-*` was refreshed May 23 to mirror ariadne's adaptation (Visual Companion *retained*).

Switching parley.nvim to inherit means:
1. Replace each real directory at `parley.nvim/.claude/skills/superpowers-*` with a symlink to `../../construct/adapted/superpowers-*` (matching nous and ariadne).
2. **Semantic regression to flag:** Visual Companion (browser-based design tool) re-appears in parley.nvim's brainstorming/etc. Originally parley.nvim asked to remove it because it's a Neovim plugin. After inheritance, the option is either (a) accept that the inherited skills are slightly browser-tinted, since parley.nvim users will simply not use them, or (b) add a parley.nvim-shaped adaptation back at the ariadne level (i.e., update `intents/superpowers.md` to also drop Visual Companion universally — which probably isn't desirable for ariadne itself).

Recommend (a): accept inheritance verbatim. The Visual Companion section is dormant if no one uses it; the cost of carrying a slightly broader skill is much less than maintaining parallel adaptations.

Other differences (per skill, from `diff -rq`): 1–6 file differences each, mostly attributable to (i) cross-reference rewrites in the newer adaptation, (ii) Visual Companion files in brainstorming, and (iii) ariadne-specific `workshop/plans/` path conventions that parley.nvim already shares.

## Spec

- `construct/skill/SKILL.md` rewritten with `--to`, `--as`, `Scope`, and `Target` removed from every command. `/construct adapt <source>`, `/construct promote <source>`, `/construct diff <skill>` — single-target throughout. Promote step writes to both `.claude/skills/` (ariadne's live) and `construct/adapted/` (inheritable copy) in one shot.
- Intent layout: `intents/<source>.md` (one file per source). Frontmatter no longer carries `scope:` / `target:`.
- Staging: flat `staging/skills/` — only one source is ever staged at a time. Less ceremony than `staging/<source>/`.
- Version manifests trimmed to `Skill | Source | Source Version | Intent File`. Pre-simplification manifests (v0001, v0002) carry a header note that their original deployment context lives in git history.
- `rollback.sh` stripped of personal-skills branch *and* its pre-existing dead `$version_dir/skills/repo/*/` glob. Now iterates skill dirs at the version-dir top level and restores into both `.claude/skills/` and `construct/adapted/`.
- New atlas entry `atlas/workflow/construct-adaptation.md` making "derivatives inherit, only ariadne adapts" explicit. Linked from `atlas/workflow/index.md` and cross-referenced from `base-layer.md`.
- parley.nvim drift cleaned up: real-dir skills at `parley.nvim/.claude/skills/superpowers-*` replaced with symlinks into `../../construct/adapted/superpowers-*`, matching nous's posture. Visual Companion re-appears verbatim; accepted (Neovim users won't use the browser-based section).

## Plan

- [x] Rewrite `construct/skill/SKILL.md`: drop `--to`, scope, target, `--as`, and the `--to personal` path; flatten staging to `staging/skills/`
- [x] Migrate `intents/superpowers/ariadne.md` → `intents/superpowers.md`
- [x] Delete `intents/superpowers/parley.nvim.md` (history preserved in git)
- [x] Update `construct/manifest.md` and `construct/versions/*/manifest.md` schema (drop Target/Scope columns)
- [x] Add atlas entry under `atlas/workflow/` stating derivatives inherit and only ariadne adapts
- [x] **parley.nvim cleanup:** delete real directories at `parley.nvim/.claude/skills/superpowers-*` and replace with symlinks to `../../construct/adapted/superpowers-*` (matching nous and ariadne). Accept Visual Companion re-appearance.
- [x] Verify nous, parley.nvim, and ariadne all resolve their adapted skills correctly post-change
- [x] (side-quest) Strip personal-skills branch from `construct/rollback.sh` and fix pre-existing dead `$version_dir/skills/repo/*/` glob (skills sit at top level of a version dir). Rollback now restores into both `.claude/skills/` and `construct/adapted/`.

## Log


- 2026-05-28: closed — Three repos resolve .claude/skills/superpowers-* to byte-identical content (md5 verified on superpowers-brainstorming/SKILL.md across ariadne, nous, parley.nvim). atlas/workflow/construct-adaptation.md added in the same commit.
### 2026-05-25 — issue created

Spun out from a session where user noted the adaptation system feels overly flexible relative to ariadne's opinionated posture. Verified the actual usage: two intents exist (ariadne.md and parley.nvim.md), derivatives inherit ariadne's adapted/ via base.manifest, and nous (the canonical derivative) has no intents/sources/staging directories of its own.

### 2026-05-25 — posture confirmed

User confirmed: parley.nvim itself became an ariadne descendant (it predates ariadne but adopted the layout over time). No remaining non-derivative consumer. Decision: simplify to `/construct adapt <source>` with no `--to`.

### 2026-05-25 — sub-questions resolved + parley.nvim drift discovered

Sub-questions both decided: drop `--to personal`; delete `parley.nvim.md` intent. While verifying, discovered that `parley.nvim/.claude/skills/superpowers-*` are real directories (vendored Apr 25 via the old `--to ../parley.nvim` flow), not symlinks to `construct/adapted/` as in nous/ariadne. They are stale and drifted (every skill has 1–6 file differences from the inherited version). Cleanup steps added to Plan. Visual Companion regression flagged but accepted — Neovim users simply won't invoke the browser-based section.

### 2026-05-27 — implemented

Decisions taken during execution:
- **Staging layout:** flat `staging/skills/` (not `staging/<source>/`). Only one source is ever staged at a time in this single-target world; the extra dir was ceremony.
- **Historical version manifests:** rewritten to the new schema (drop `Scope`/`Target` columns). v0001 and v0002 kept a header note that they predate the simplification; v0002's intent-file column now points to the surviving unified intent rather than the deleted `parley.nvim.md`.
- **rollback.sh:** stripped the personal-scoped branch *and* fixed a pre-existing dead glob (`$version_dir/skills/repo/*/`) that no longer matched actual on-disk layout. Rollback now restores each skill to both `.claude/skills/` and `construct/adapted/`, which means a rollback in ariadne also rolls back every derivative through the symlink chain. Logged as side-quest in plan + commit.

Atlas: added `atlas/workflow/construct-adaptation.md`, linked from `atlas/workflow/index.md`. Updated stale references in `atlas/workflow/base-layer.md` and `construct/base.manifest` comment block. Updated stale intent path in `workshop/issues/000034-...md`.

Verification: all three repos (ariadne, nous, parley.nvim) now resolve `.claude/skills/superpowers-*` to the same `ariadne/construct/adapted/superpowers-*` content (md5-equal on the spot-checked `superpowers-brainstorming/SKILL.md`).
