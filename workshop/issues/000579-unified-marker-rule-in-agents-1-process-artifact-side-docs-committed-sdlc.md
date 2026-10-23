---
id: 000579
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '70217de39837627aed7a79f69e9b8fa2b9ba8b7f' # card fields mirrored from issue-cards; edit via sdlc
---

# Unified marker rule in AGENTS §1; process-artifact side docs committed; sdlc refuses links to uncommitted files

## Problem

## Spec

**Side docs in process artifacts, and the unified marker rule (operator, 10-23).** Follows the single collaborative
mode (ariadne#575): in issues, plans and projects too, a `🤖[…]` question gets a side doc (the marker becomes
`→ [discussion](…)`), an instruction is acted on, and `🤖[fold in]` collapses a side doc back into the text.
Discussion: ops `session.side/2026-10-23-07-side-docs-in-process-artifacts.md`.

1. **AGENTS §1** (the shared constitution): replace "a 🤖[H] in any artifact … answer it, and resolve it in place
   (fold the answer into the prose)" with the unified rule above, plus: side docs of process artifacts live next to
   them (`<artifact>.side/…`), are committed with them, and archive with the issue. The operator's private session
   side docs (`session.side/`, git-ignored in ops) stay uncommitted: they're operational context.
2. **A guard in sdlc** where an issue's text reaches main (`issue publish`, `change-code`, `close`): refuse when a link
   in the issue, its plan or a referenced project file points at a file that isn't committed or is git-ignored,
   naming the link and the fix (commit the side doc, or fold it in).
3. Keep the review convention (`.agents/skills/xx-fix/review-convention.md`) consistent with 1.

## Done when

- AGENTS §1 and the review convention state the unified rule and commit-with-artifact (doc check).
- `sdlc issue publish`, `change-code` and `close` refuse a link to an uncommitted or ignored file and pass once it's
  committed or removed (tests).

## Plan

- [ ]

## Log

### 2026-10-23
