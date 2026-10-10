---
id: 000318
status: open
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: 'fc1be9075c568ca1e8147ec5eca08f44873a47e1' # card fields mirrored from issue-cards; edit via sdlc
---

# sandbox: allow couch in base-layer settings so agents can message peers

## Problem

Agents can't send couch peer messages from inside the Claude Code sandbox. `couch --send-to` dials the
message socket under `/tmp/pair-message-<uid>/<hash>.sock` and fails with `connect: operation not
permitted`. The outcome is reported as "uncertain", and `couch --message-status` then confirms nothing was
delivered. The base-layer sandbox settings (`.claude/settings.json`, woven into every repo) exclude `sdlc`,
`git` and `gh`, but not `couch`, and the socket isn't in `network.allowUnixSockets`.

The impact grows with multi-slot work. Every TL dispatch and every worker reply needs a sandbox bypass,
which costs a permission prompt or an auto-mode verdict each time. A background TL monitor can't reach
couch at all (it can't detect an idle slot). Seen by the TL in ariadne:1 on 10-09, and again from the ops
slot on 10-10 (ops `teams/` learnings for ariadne-robustness-1, kink 32).

## Spec

Add `couch` to `excludedCommands` in the base-layer sandbox settings, the same treatment as `sdlc`.
couch keeps its own guards: delivery receipts, refusal while a recipient has a pending message, and
`--reboot` refusing live slots. A narrower alternative is to allow the message-socket directory in
`allowUnixSockets`, but the path embeds the uid and a hash, so it's brittle. Pick the exclusion unless
the review finds a reason not to.

## Done when

- A sandboxed `couch --send-to <slot> --message …` from a woven repo delivers, and `couch --message-status ID --json` shows a receipt.
- The setting lives in the ariadne base layer and reaches peer repos through `weave refresh` (checked in ops).

## Plan

- [ ]

## Log

### 2026-10-10
