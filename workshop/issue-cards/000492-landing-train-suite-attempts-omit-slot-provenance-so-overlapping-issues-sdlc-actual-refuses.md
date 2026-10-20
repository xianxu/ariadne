---
id: 000492
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
started: 2026-10-20T08:41:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# Landing-train suite attempts omit slot provenance, so overlapping issues' sdlc actual refuses

## Problem

Landing-train suite attempts are saved without a `slot` field. pair's train 174 suite (attempt
f6a4337408d36adfd9d00ddb9d298d99, head c2c6b927, conductor pair:7) was recorded as version 2 with no `slot`, so
`sdlc actual` for an unrelated issue whose window overlapped it (pair#725) refused: "suite attempt … has unknown
slot provenance … ask the TL/hub to restore the named measurement evidence". The TL restored it by hand on 10-20
(added `"slot": "pair:7"`; original kept). Every train will do this to every overlapping issue's actuals.
