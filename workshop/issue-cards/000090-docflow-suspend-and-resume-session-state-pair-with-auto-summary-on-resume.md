---
id: 000090
status: open
created: 2026-06-10
updated: 2026-06-10
estimate_hours:
github_issue:
---

# docflow: suspend and resume session-state pair with auto-summary on resume

## Problem

A `docflow` review session already persists its *content* on a `review/<slug>`
branch (rounds journaled via `docflow round`), but it has **no first-class way
to suspend and later resume the session itself**. Today both ends are ad-hoc:

- **Suspend (park):** the operator manually commits in-progress work, pushes the
  branch, hand-writes a pointer issue on `main` so the parked work is findable,
  and switches away. (Concrete instance: xianxu.dev issue `#000001`
  "AI blogging-workflow meta post" — a hand-made resume pointer whose whole body
  is "the work lives on branch `review/a-blogging-workflow`, switch and continue.")
- **Resume:** nothing reconstructs state. On switching back the operator (or a
  fresh agent) has to re-derive where things stand by hand — what rounds
  happened, which markers are open, what's left. In the motivating session the
  agent reconstructed this manually (read the issue, diffed branch vs base,
  listed open `🤖` markers, found the two unresolved exact-quote items) before
  any work could continue. `docflow status` exists but is terse (branch, round
  count, in-scope files, 🤖 count) — not a "here's where we are" summary.

These are a **pair**: a suspend that captures state and a resume that restores +
summarizes it. Missing the pair makes parking lossy and resumption a manual
re-orientation tax every time.
