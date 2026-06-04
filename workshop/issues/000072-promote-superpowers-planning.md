---
id: 000072
status: working
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-04
estimate_hours: 2
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

Detailed plan: [`workshop/plans/000072-promote-superpowers-planning-plan.md`](../plans/000072-promote-superpowers-planning-plan.md)
(authored via `superpowers-writing-plans` — dogfooding the path this issue promotes).

**Decisions (operator-confirmed):** fallback = **discourage (prose-only)**; do NOT
teach the binary `~/.claude/plans` (agent-agnosticism, §11). Add a non-blocking
`start-plan` pointer line. Single review boundary (atomic) — no `Mx`.

- Confirmed: the skill already writes to `workshop/plans/` (SKILL.md:18) → promotion, not invention.
- [x] **Task 1** — `planPointer` pure helper + `TestPlanPointer`; printed by `start-plan`
      between the architecture block and the contention read (`startplan.go`).
- [x] **Task 2** — AGENTS.md §2/§1 name the skill + demote `~/.claude/plans`; SKILL.md slug
      grammar aligned to `NNNNNN-`; start-plan helptext OUTPUT line.
- [x] **Task 3** — atlas: `issue-lifecycle.md` (flow + producer), `sdlc-binary.md` (pointer),
      `artifact-hierarchy.md` (producer).

## Log

### 2026-06-02

Filed from the sdlc tooling retro
(`workshop/pensive/2026-06-02-01-pensive-sdlc-tooling-retro.md`, finding F6). `EnterPlanMode`
is a Claude Code builtin tool (writes ephemeral `~/.claude/plans`); ariadne already has the
adapted superpowers planning skill — promote it.

### 2026-06-04

Shipped. **Promotion, not invention** — the skill already wrote to `workshop/plans/`
(SKILL.md:18), so the work was constitution + atlas prose + one pure helper.
**Decisions:** fallback = *discourage (prose-only)*; rejected teaching the binary
`~/.claude/plans` (agent-agnosticism, §11) and the `sdlc plan import` bridge verb
(YAGNI). The session itself dogfooded the bug — plan mode stranded my plan at
`~/.claude/plans/validated-puzzling-map.md` (a slug with zero tie to `000072`,
confirming why binary-ingest's slug→issue mapping is unrecoverable) — so I authored
the durable plan via `superpowers-writing-plans` →
`workshop/plans/000072-…-plan.md`. **Delivered:** `planPointer` pure helper printed
by `start-plan` between the architecture block and the contention read (ordering
*what → how/where → environment*); AGENTS.md §1/§2 reframe + `~/.claude/plans`
demotion; SKILL.md slug grammar aligned to `NNNNNN-`; helptext OUTPUT; 3 atlas
files. plan-quality judge: CLEAN (high). `go test ./cmd/sdlc/...` green
(`TestPlanPointer` added); `go vet` clean; live `start-plan --issue 72` ordering
verified. `ARCH-PURE` (id-only pure helper, table-tested), `ARCH-DRY` (one path
grammar across constitution/skill/binary, pinned by the literal-asserting test).
