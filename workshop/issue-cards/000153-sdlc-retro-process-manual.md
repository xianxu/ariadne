---
id: '000153'
status: done
started: 2026-07-01T11:43:57-07:00
created: 2026-07-01
updated: 2026-07-01
estimate_hours: 5.03
actual_hours: 5.85
---

# sdlc retro process manual

## Problem

The agentic SDLC process is encoded **implicitly**, distributed across layers that
are never documented together: `AGENTS.md`/`CLAUDE.md`, `atlas/`, the `sdlc` binary
(help text + prompts it injects at gates), `.claude/skills/*/SKILL.md`,
`workshop/lessons.md`, and the memories the agent chooses to persist.

Two concrete pains follow:

1. **No single readable process manual.** A human periodically needs to see what the
   process *actually is* — to apply their own judgment about whether it's good — but
   today that means reading source across all those layers. `atlas/` is explicitly
   advisory (a lagging first-glance), so it doesn't serve this.
2. **Unclear when the sdlc-embedded prompts actually fire.** The prompts are
   systematic (they live in the binary) but there's no view of *which* fired in a
   given session, in what order, or whether the agent followed them.

Note on scope: the target is **human process-audit**, not agent-drift policing. Where
"drift" appears below it means the user's sense — *an instruction was injected but the
agent ignored it* — not documentation staleness.
