---
id: 000204
status: open
created: 2026-08-25
updated: 2026-08-25
estimate_hours:
github_issue:
---

# Run each boundary reviewer in a disposable isolated checkout

## Problem

Boundary reviewers are instructed to be read-only and leave fixes to the main
agent, but the review protocol also asks them to prove fixes by reverting or
mutating them and running tests. The process currently inherits the real
checkout as its working directory. Claude retains Bash under
`bypassPermissions`; Codex and Gemini do not consume Claude's tool allowlist.
The read-only claim is therefore prose, not an isolation boundary.

An audit of pair#146's ten most recent couch reviews found tool use in all ten
and deliberate mutation or generated artifacts in eight. Some reviewers
voluntarily used scratch worktrees or `git archive` copies, while one Codex
review created and then removed a binary in the real checkout. Safety currently
depends on each reviewer independently inventing the same containment practice.
