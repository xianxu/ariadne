---
id: '000079'
status: done
created: 2026-06-03
updated: 2026-06-04
estimate_hours: 2.5
actual_hours: 1.5
---

# doc-review-flow: branch-scoped prose review with per-round git journaling

## Problem

`xx-fix` co-authoring (operator + agent trading 🤖 markers in a markdown doc)
works well, but a heavier multi-round review has no memory: each round
overwrites the last and the only record of how a paragraph evolved — and *why* —
lives in the ephemeral chat transcript, which is gone within days. We want the
full back-and-forth retained as durable, attributable, forensically-greppable
history, without inventing a bespoke versioning scheme (git already does this)
and without leaving review scaffolding (markers) in shipped prose.

Observed during a real review of a xianxu.dev blog post (the first heavy
`xx-fix` session). The agent's verbose reasoning — the actual rationale behind
each edit — is currently lost entirely; only the terse in-doc marker survives.
