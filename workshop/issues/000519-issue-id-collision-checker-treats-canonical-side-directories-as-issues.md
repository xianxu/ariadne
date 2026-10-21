---
id: 000519
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ec90c04285f9f8f5a4daccbb39e5d88abe766b87' # card fields mirrored from issue-cards; edit via sdlc
---

# Issue ID collision checker treats canonical side directories as issues

## Problem

pair#731 PR503 CI run39542144320 failed its duplicate-issue-ID gate because a canonical sibling side-document directory was counted as a second issue: workshop/issues/000731-a-pane-that-fell-back-to-a-native-shell-still-classifies-as-a-pair-terminal.side alongside the corresponding .md. The directory contained only metadata-capture.json, not another issue. Current lint-ids does not reproduce after the capture is moved inline.

## Spec

Issue collision enumeration must distinguish actual issue Markdown documents from canonical .side directories and other non-issue entries while preserving real duplicate-ID refusal and degraded-read refusal. Reproduce through the same ranged lint-ids path CI uses, with an issue plus nested side documents and non-Markdown evidence; pin actual conflicting Markdown issues as the negative control.

## Done when

- A real issue and its canonical .side subtree do not introduce a duplicate ID.
- Two actual issue documents with one ID still refuse in a ranged CI check.
- The fix has a regression through the production issue enumeration seam.

## Plan

- [ ]

## Log

Filed unclaimed from pair:6 after CI on pair#731. That branch preserves its small capture inline to keep its own landing independent; no gate bypass or upstream implementation was attempted. Ariadne AGENTS.local.md read; MEMORY.md is absent in its :0 checkout.
