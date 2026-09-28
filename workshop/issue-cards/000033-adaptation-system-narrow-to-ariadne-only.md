---
id: '000033'
status: done
created: 2026-05-25
updated: 2026-05-28
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
