---
id: 000579
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '8736228a3a72ab9975e8e25a57238031fabffba9' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T16:38:06-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
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

Durable plan: [000579 plan](../plans/000579-unified-marker-rule-in-agents-1-process-artifact-side-docs-committed-sdlc-plan.md).

- [ ] The rule in prose: AGENTS.base.md §1, review convention §6, collaboration-modes §4
- [ ] `sidedoc.LocalLinks` extracted; `RebaseLinks` over it
- [ ] `linkguard` census; the side-links gate in change-code, close and issue publish
- [ ] (D1) publish carries the side docs the details link to
- [ ] atlas, targeted runs, recorded suite

## Log

### 2026-10-23
- Claimed and started from ariadne:5 after #565 landed. Survey: publish carries only details and `assets/`; change-code's checkpoint commits issue, plan and gate ledger; nothing commits `.side/` before the archive. So a refusal at publish from a resting branch has no "commit it" fix unless publish carries linked side docs: decision D1 to the TL.
- The link grammar already lives in `sidedoc.RebaseLinks`; the guard reuses it via an extracted `LocalLinks` (ARCH-DRY).
