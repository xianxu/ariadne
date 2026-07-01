---
id: 000158
status: open
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-01
estimate_hours:
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

### 2026-07-01

Implemented + verified in metis: file resolves via both skill dirs (→ ariadne source), composed `AGENTS.md` carries the resolve rule + `.agents/skills/xx-fix/review-convention.md` reference, dead `workshop/targets/review-convention.md` pointer gone. The move landed via an sdlc issue-sync auto-commit (`bd53071`); the two doc edits are the completing commit. **Deferred (separate weave issue):** the `.claude/skills → .agents/skills` symlink unification — it rewrites weave's per-harness skill-lowering + bidirectional GC and rests on a new "Claude follows a symlinked `.claude/skills`" assumption that needs a harness-ledger live-test. Other repos pick up the composed-AGENTS change on their next `weave`/`make refresh`.
