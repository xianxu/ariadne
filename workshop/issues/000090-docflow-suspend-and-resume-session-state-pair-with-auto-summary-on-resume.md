---
id: 000090
status: open
deps: []
github_issue:
created: 2026-06-10
updated: 2026-06-10
estimate_hours:
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

## Spec

Add a suspend ⇄ resume pair to the docflow lifecycle (`scripts/docflow.sh` +
the **xx-fix** SKILL.md triggers), composing with existing `start/round/ship/status`.

- **`docflow suspend`** (operator triggers: "park this", "suspend the docflow",
  "stop for now") — commit any dirty edits (operator side), push the
  `review/<slug>` branch, and persist a **findable resume pointer**. Mechanism
  TBD in Plan: branch alone is the state; the pointer is what makes a parked
  session discoverable from `main`. Options: (a) auto-create/update a pointer
  **issue** on `main` (generalize the `#000001` pattern — title, branch name,
  one-line "to resume" note, link to plan), (b) a lighter `docflow ls` that
  scans `review/*` branches so no issue is needed, or (c) both. Decide in Plan.
  Idempotent: suspending twice updates, not duplicates.
- **`docflow resume <slug|branch>`** (triggers: "resume the docflow",
  "continue review/<slug>", or auto on `docflow start` detecting an existing
  branch) — switch to the branch and **always emit a state summary before any
  edit**: rounds so far (from the journal), open `🤖` markers by file + section
  kind, reading frontier, base..HEAD diffstat, and the explicit "what's left."
  This is the auto version of the manual reconstruction done in the motivating
  session — it must be automatic, not improvised.
- The **resume summary is mandatory** — a resumed session never proceeds to edits
  without first showing the operator where things stand (mirrors the "review
  rounds are explicitly triggered" discipline: re-orient before acting).

## Done when

- `docflow suspend` / `docflow resume` exist in `scripts/docflow.sh`, compose
  with `start/round/ship`, and are documented as triggers in the xx-fix SKILL.md.
- Resume prints a state summary (rounds, open markers, frontier, diffstat,
  what's-left) and refuses to silently start editing without it.
- Suspend persists a findable pointer (issue and/or `docflow ls`) idempotently.
- A process-level test drives start → round → suspend → resume and asserts the
  resume summary reflects the journaled rounds + open markers.

## Plan

- [ ] Decide the pointer mechanism: pointer-issue on `main` vs `docflow ls`
      branch-scan vs both (weigh against the `#000001` precedent + brain/atlas)
- [ ] `docflow suspend`: commit + push + persist pointer (idempotent)
- [ ] `docflow resume`: switch + mandatory state summary (rounds / markers /
      frontier / diffstat / what's-left), reusing `docflow status` internals
- [ ] Wire triggers into xx-fix SKILL.md (suspend/resume phrases + the
      "summarize before editing on resume" rule)
- [ ] Process-level test: start → round → suspend → resume summary assertion

## Log

### 2026-06-10

- Filed from a live xianxu.dev docflow session (`review/a-blogging-workflow`,
  blog post about the space-data-center writing workflow). Resuming that branch
  required manual state reconstruction — the exact gap this issue closes. The
  hand-made pointer `xianxu.dev#000001` is the prototype for `suspend`'s output;
  the manual "summarize where things are" I did on resume is the prototype for
  `resume`'s mandatory summary.
