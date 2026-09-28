---
id: 000120
status: open
created: 2026-06-19
updated: 2026-06-19
estimate_hours:
github_issue:
---

# Subagent strategy: add the fork-for-implementation pattern (context inheritance)

## Problem

The subagent strategy — AGENTS.md §3 and the adapted superpowers skills
(`construct/adapted/superpowers-dispatching-parallel-agents/SKILL.md`,
`construct/adapted/superpowers-subagent-driven-development/SKILL.md`) — frames
subagent use around one axis: *"is the context I need capturable as a prompt?
→ subagent."* That implicitly assumes a **fresh** subagent (starts blank, sees
only the prompt). It never mentions the **fork** option, where the subagent
inherits the parent's *full conversation transcript* and runs on the parent
model.

That omission misses a high-value pattern. Forking substantial implementation
work lets it run with the entire design arc in context, in its **own** context
budget, and return only a digest — so the main thread keeps the design +
conclusion without accumulating the implementation exhaust (every edit, test
run, debug cycle). The user's framing: *"a stack of context"* — push full
context in, pop a digest out; the exhaust stays on the fork's frame and is
discarded on return. Validated this session (pair #66): M2 was forked and spent
~680k tokens on the fork while adding near-zero to the main thread.
