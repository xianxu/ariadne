---
id: '000121'
status: done
started: 2026-06-20T10:43:05-07:00
created: 2026-06-19
updated: 2026-06-22
estimate_hours: 3.51
actual_hours: 2.80
---

# Writing-assistant skill — pair review workbench agent protocol

## Problem

The `xx-fix` skill (`construct/local/fix/`) is the **agent half** of the pair agentic
review workbench (pair **#000066**). M0–M3 built the pair-side document surface (the
record/apply/undo spine, projection/markers, the review window + toggle + poke + bar).
M4 needs the agent half: today the workbench pokes a bare `/xx-fix <path>` and the
dumb agent guesses (it ran `doc-review`/fact-check, edited nothing). The skill must
become a review-mode-aware, record-driven collaborative writing assistant.

This is the cross-repo counterpart of pair #000066 M4 — the contract both sides honor
is `pair/workshop/targets/review-protocol.md` (the agent↔nvim state machine). pair #66
M4 (the nvim consumer / seam / bar / menu) **depends on** this issue.

Note: `xx-fix` has outlived its name (it's no longer "fix small things from `🤖[]`" —
it's a collaborative editor). Rename is deferred to a follow-up; the likely
user-facing name is `review`. Keep the `xx-fix` name working until the rename lands
in lockstep with any downstream trigger changes.
