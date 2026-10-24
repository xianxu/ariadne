---
id: 000581
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
severity: medium
card_mirror: '6aa79a32d2ffabc6ad5fa3050b5457ef58423f41' # card fields mirrored from issue-cards; edit via sdlc
---

# sidedoc ChatItems: an invalid answer-link destination reads as answered

## Problem

## Spec

Reported by pair:3 during pair#838's integration (2026-10-23). `pkg/sidedoc` `answeredItem`
(`→\s*(\[answer\]\([^)]*\)\s*$|done:)`, chat.go) treats a syntactically invalid Markdown destination as answered:
the legacy item `- Why? → [answer](broken path)` reads as answered, and so does an empty `→ [answer]()`. The item
is then skipped forever, so an operator question goes unanswered without any signal. The writer already refuses
such targets: tl-lead `linkBullet` rejects empty targets, whitespace and parentheses. Pair's earlier guard required a real
no-whitespace destination and let the first real answer be appended.

Fix: the reader accepts what the writer can produce, a nonempty destination with no whitespace or parentheses
(`\[answer\]\([^\s()]+\)`), and the rule is stated once, shared by `linkBullet`'s refusal and `answeredItem`
(ARCH-DRY). collaboration-modes.md §3 names the destination rule. pair keeps the shared oracle (no local fork) and
picks up the fix at its landed revision.

## Done when

- `TestChatItems` rows: `→ [answer](broken path)` and `→ [answer]()` are open; `→ [answer](a/b.md)` is answered;
  a link inside a fence still never answers.
- tl-lead's refusal and the reader use one definition of a valid answer target (a test pins that they agree).
- pair:3 is told the landed revision.

## Plan

- [ ]

## Log

### 2026-10-23
