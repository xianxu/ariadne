---
id: 000426
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c89951ea05adcf38721e4b79c76467ba473e8966' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead status: write the discussions index through couch's frontmatter read/write (pair#643), dropping the printed-diff workaround

## Problem

ariadne#424 (M2) writes the TL session file's `discussions:` index through couch's whole-file `read --all` / `write --all`, which reset the diff baseline for the whole file. To avoid consuming the operator's unread edits, `tl-lead`'s `syncIndex` decides no-ops by `Peek` and prints the pending diff (`readAllShowingUnread`) before every whole-file read: correct, but noisy. pair#643 (landed, pair 43313c59) added `couch --session-sync read --frontmatter` and `write --frontmatter … --if-hash H`, which move the baseline only for the frontmatter.

## Spec

- `SessionSync` gains `ReadFrontmatter(file) (text, hash, err)` and `WriteFrontmatter(file, bodyPath, hash)` over the new verbs; the fake models a frontmatter-only baseline (a frontmatter read or write resets only the frontmatter's shadow).
- `syncIndex` reads and writes only the frontmatter: `sidedoc.ReconcileIndex` needs the body's links too, so it still reads the body without moving the baseline (`Peek`), but writes only the frontmatter through the CAS; the printed-diff workaround (`readAllShowingUnread`, `windowChanges`, `syncIndexAfter`) is removed.
- The live `TestSessionSyncConformance` checks that a frontmatter read or write leaves an operator's section edit in the diff.
- Precondition: the operator's running Couch has restarted with pair#643 (the TL confirms); until then the verbs refuse.

## Done when

- `instructions link`, `status submit` and `status index` never consume or print the operator's unread section edits, with tests against the fake and the live conformance.
- The printed-diff workaround is gone, and help/atlas describe the frontmatter write.

## Plan

- [ ]

## Log

### 2026-10-17
