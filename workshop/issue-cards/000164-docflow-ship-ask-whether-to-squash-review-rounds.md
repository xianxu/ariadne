---
id: 000164
status: open
created: 2026-07-05
updated: 2026-07-05
estimate_hours:
github_issue:
---

# docflow ship: ask whether to squash review rounds

## Problem

Today `docflow ship` (`scripts/docflow.sh`, `cmd_ship`) does a **`--no-ff` merge**
of `review/<slug>` into base and deletes the branch — deliberately preserving
every human/agent round commit so `git log` shows step-by-step how the artifact
was constructed (the script header explicitly states "No squash"). That's the
right default for durable authoring history.

But sometimes the author wants the opposite: **land the finished doc as one clean
commit and keep the inner making-of private** — the round-by-round back-and-forth
(and the agent's rationale in each round body) is scaffolding they don't want in
the published history of a manuscript.

Split out of **pair#89** (which kept the concurrent-edit-reconciliation scope);
this is the "ask to squash on ship" piece, which is substantively `docflow` +
`xx-fix` skill work and deserves its own design.
