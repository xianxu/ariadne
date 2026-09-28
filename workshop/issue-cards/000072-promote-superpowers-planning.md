---
id: '000072'
status: done
created: 2026-06-02
updated: 2026-06-04
estimate_hours: 2
actual_hours: 1.5
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
