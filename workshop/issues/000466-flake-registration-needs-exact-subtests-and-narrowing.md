---
id: 000466
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'e06b4b3fe79a012f37cc23e114f23ce3cd8fd8f1' # card fields mirrored from issue-cards; edit via sdlc
---

# flake registration needs exact subtests and narrowing

## Problem

`sdlc flake register` accepts only a top-level Go test name. While triaging #465, the failing subtest was `github.com/xianxu/ariadne/scripts/test.TestConstructScripts/plain/dev-aliases.test.sh`, but the setter only accepted the broad `github.com/xianxu/ariadne/scripts/test.TestConstructScripts`. That registration covers unrelated construct-script failures. The setter can add tests but has no unregister or narrowing operation, so a mistaken broad registration cannot be corrected through a card verb before the issue closes. This is a #394 follow-up; the temporary broad #465 registration is tolerated only until its fix lands.

## Spec

Support exact subtest names in the known-flake registry and same-head evidence/rerun path. Add a compare-and-swap card setter that removes or replaces a registered name, preserving concurrent additions and validating the resulting list. The list and landing classifier must distinguish a registered subtest from other subtests of the same top-level test; a failure in another construct script remains unknown and blocks. Keep the existing strict exact-head and ownership rules.

## Done when

- A real nested Go test fixture registers one named subtest; its same-head failure may use the known-flake rerun, while a sibling subtest failure cannot.
- A card owner can narrow or remove one registered name through the verb without editing card files; concurrent card additions survive, and a repeated setter is idempotent.
- Help and recovery contracts state the new names, writes and retry behavior; targeted tests and the recorded suite pass.

## Plan

- [ ]

## Log

2026-10-19: Found during #458's full-suite triage. `flakes.ValidTest` and `SplitTest` accept only top-level names; `sdlc flake register` only adds. #465 temporarily registers the top-level `scripts/test.TestConstructScripts` while its test-cost fix is in progress.
