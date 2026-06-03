---
id: 000072
status: open
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
estimate_hours:
---

# promote adapted superpowers-writing-plans as ariadne's canonical plan path over builtin EnterPlanMode

## Problem

AGENTS.md §2 says "non-trivial task (>3 files or >100 lines) → plan mode, wait for approval,"
and the constitution wants detailed designs to live in `workshop/plans/NNNNNN-…`. But the
agent's default "plan mode" is the **Claude Code builtin `EnterPlanMode` tool**, which writes
the plan to **`~/.claude/plans/<name>.md`** — harness-controlled, ephemeral, **not
version-controlled**, and disconnected from `workshop/plans/`.

Observed 2026-06-02 (`nous#41`): the plan lived in `~/.claude/plans/…`; a milestone-review
judge even recommended adding a `## Revisions` entry to it — a file that won't survive. The
durable plan record had to be hand-carried into the issue's `## Spec`/`## Log` instead.

Meanwhile ariadne **already ships** an adapted planning skill:
`construct/adapted/superpowers-writing-plans/` (+ sibling `superpowers-executing-plans/`),
which lands plans wherever we tell it. So this is **promotion, not invention**.

## Spec

- Make the **adapted `superpowers-writing-plans`** the canonical planning path in the
  AGENTS.md stack: §2's "plan mode" means *this skill*, landing the plan in
  `workshop/plans/NNNNNN-<slug>-plan.md` (the constitutional location), not the builtin
  `EnterPlanMode`'s ephemeral file.
- Tighten the artifact hierarchy wording (AGENTS.md §1/§2) so "complex → add
  `workshop/plans/…`" and "plan mode" point at the same skill + location.
- Decide the fallback for when the builtin `EnterPlanMode` *is* used (harness default): either
  (a) discourage it in favor of the skill, or (b) have `sdlc change-code` ingest the
  `~/.claude/plans/…` file into `workshop/plans/` so the durable copy exists.
- Verify the adapted skill is actually surfaced/invocable in the AGENTS.md stack (promote it
  more prominently if it's currently buried).

## Done when

- AGENTS.md §1/§2 name the adapted `superpowers-writing-plans` as the planning path, with
  plans landing in `workshop/plans/` (version-controlled).
- A non-trivial task's plan is durable + in-repo by default (not stranded in
  `~/.claude/plans/`).

## Plan

- [ ] Confirm the adapted `superpowers-writing-plans` skill writes to `workshop/plans/` (or
      make it).
- [ ] Edit AGENTS.md §2 (and §1's plans bullet) to name the skill + location as canonical.
- [ ] Decide + wire the EnterPlanMode fallback (discourage, or `change-code` ingests).

## Log

### 2026-06-02

Filed from the sdlc tooling retro
(`workshop/pensive/2026-06-02-01-pensive-sdlc-tooling-retro.md`, finding F6). `EnterPlanMode`
is a Claude Code builtin tool (writes ephemeral `~/.claude/plans`); ariadne already has the
adapted superpowers planning skill — promote it.
