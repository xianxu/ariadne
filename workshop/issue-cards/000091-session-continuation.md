---
id: '000091'
status: done
created: 2026-06-11
updated: 2026-06-11
estimate_hours: 2
actual_hours: 0.09
---

# Session continuation: datatype prototype

## Problem

A live coding session carries human-meaningful working state that the flattened
repo loses — even when working state was faithfully written to issues / `sdlc state`:

- back-and-forth **deliberation** on a topic that never landed in an artifact;
- **cross-issue cross-pollination** — a session often spans several related issues,
  and a decision in one silently informs another;
- **dead ends** and the reasoning behind decisions made — the "tried X, rejected
  for Y" that the final repo state erases;
- the **"you are here / what's the next concrete step"** sense a flattened tree
  can't convey.

There's no durable, portable way to materialize that so work can resume later
(after a break), by another person, on another machine, or under a different
agent. Native agent session stores (`~/.claude/...`, `.antigravitycli/...`, codex)
are machine-faithful but per-agent, locally-scoped, subject to recycling, and the
*wrong kind* of state for this purpose.
