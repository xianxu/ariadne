---
id: 000158
status: codecomplete
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-06
estimate_hours:
actual_hours: 0.15
---

# convention delivery: move review-convention.md into the xx-fix skill dir so derivatives can reach it (dead workshop/targets pointer)

## Problem

`AGENTS.base.md §1` pointed every repo at *"Full table in `workshop/targets/review-convention.md`"*, but that file lives **only in ariadne** — `workshop/` is scaffolded per-repo, not exported, and `workshop/targets/` isn't even created in derivatives. So the pointer is **dead in brain, metis, nous, …** (verified: the file is missing in all three), while the dead pointer itself *is* delivered to every derivative's composed `AGENTS.md`. The convention's canonical grammar was unreachable exactly where agents work on docs.

## Spec

Deliver the convention through an **exported** channel. The `xx-fix` skill dir is served to every derivative (`.claude/skills/xx-fix` + `.agents/skills/xx-fix` → ariadne `construct/local/fix`), and the Agent Skills standard bundles supporting files with `SKILL.md`.

- Move `review-convention.md` → `construct/local/fix/` (beside `SKILL.md`); it now rides the skill symlink into every repo.
- `AGENTS.base.md §1`: reference it agent-agnostically as `.agents/skills/xx-fix/review-convention.md` — a **real path in every repo** (the neutral Agent Skills path; readable by any harness, incl. Claude Code). Also add the resolve rule: a `🤖[H]` in any artifact is a question/instruction the agent notices, answers, and **resolves in place** that turn.
- `xx-fix SKILL.md`: point at its sibling `review-convention.md` (this skill dir).

## Done when

- In a derivative, the file resolves via **both** `.agents/skills/xx-fix/review-convention.md` and `.claude/skills/xx-fix/review-convention.md`.
- The resolve rule + `.agents/skills` reference are in the composed `AGENTS.md`; the dead `workshop/targets` pointer is gone.

## Plan

- [x] Move file into the skill dir; update `AGENTS.base.md §1` (rule + reference) + `xx-fix SKILL.md`; `weave` + verify in metis.

## Log


- 2026-07-06: closed — In a derivative (metis) and in ariadne, review-convention.md resolves via BOTH .agents/skills/xx-fix/ and .claude/skills/xx-fix/ (skill symlink → construct/local/fix/); AGENTS.base.md §1 carries the resolve-in-place 🤖[H] rule + the .agents/skills path; dead workshop/targets/review-convention.md pointer removed. Shipped on main (16e9f0b7, bd53071).; review verdict: FIX-THEN-SHIP
### 2026-07-01

Implemented + verified in metis: file resolves via both skill dirs (→ ariadne source), composed `AGENTS.md` carries the resolve rule + `.agents/skills/xx-fix/review-convention.md` reference, dead `workshop/targets/review-convention.md` pointer gone. The move landed via an sdlc issue-sync auto-commit (`bd53071`); the two doc edits are the completing commit. **Deferred (separate weave issue):** the `.claude/skills → .agents/skills` symlink unification — it rewrites weave's per-harness skill-lowering + bidirectional GC and rests on a new "Claude follows a symlinked `.claude/skills`" assumption that needs a harness-ledger live-test. Other repos pick up the composed-AGENTS change on their next `weave`/`make refresh`.

## Revisions

### 2026-07-06 — close-review FIX-THEN-SHIP: full consumer set of the move

The boundary review's shadow-sweep surfaced that the Plan's single item enumerated only `AGENTS.base.md §1` + `xx-fix SKILL.md` as consumers of the old `workshop/targets/review-convention.md` path — it missed a **third** consumer, `construct/datatype/target.md:78`, whose relative link was left dangling by the move. Because `target.md` is itself a served base-layer datatype doc (DAG-merged into derivatives), that dead pointer reintroduced the exact "grammar unreachable in derivatives" failure this issue set out to kill. **Delta (fixed at the gate, same close commit):**
- `construct/datatype/target.md:78` — repointed to `.agents/skills/xx-fix/review-convention.md` (agent-agnostic, mirrors the `AGENTS.base.md §1` choice so it resolves in every woven repo).
- `atlas/workflow/ledger-landscape.md` — the "where does the marker grammar live" ledger entry still named `workshop/targets/review-convention.md` as authoritative; repointed to the skill-dir location + neutral path.
- `construct/local/fix/review-convention.md §6` — added a one-clause carve-out reconciling the new "resolve `🤖[H]` in place that same turn" rule with §6's "resolution is always operator-initiated" (an operator-authored `🤖[H]` directed at the agent *is* the acknowledgment), now that `AGENTS.base.md` names this file as the authority. Removes the apparent contradiction the review flagged.
- Full-repo sweep confirms no remaining live references to the old path (only historical mentions in this issue file).
