---
id: 000408
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: '59c47dad70858f30b32467bfaf76c4403bc51b79' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T11:18:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "62c3380d", done: "f3f27de4"}
---

# issue publish of an archived issue creates an active twin under workshop/issues (pair#579)

## Problem

2026-10-16: pair:3 ran `sdlc issue publish --issue 579` to add a Log entry to pair#579, which is done and archived
(`workshop/history/issues/000579-*.md`). With the default issues dir, publish created a NEW active twin
`workshop/issues/000579-*.md` (47 lines) on pair main (commit 7af5c7b1). Re-running with
`--issues-dir workshop/history/issues` published correctly to the archive. No verb removes the stray active copy.

## Spec

- `issue publish` resolves the issue's details file wherever it lives (active or archived), the same way
  `issue show` does, and refuses to create an active file for an issue whose card is done or archived.
- If both an active and an archived file exist for one issue, publish refuses and names both.

## Done when

- Tests: publishing an archived issue updates the archive file and creates nothing under `workshop/issues/`; a
  done-card issue with no details refuses; the twin case refuses.

## Plan

- [ ] Reproduce archive updates, missing terminal details and duplicate-location
  refusals in isolated real-Git publication fixtures; confirm red.
- [ ] Resolve details by ID across active and archive locations on pinned main,
  preserve republish safeguards, and recheck location before publication.
- [ ] Run focused and recorded full-suite verification; update Atlas and prepare
  the close review, then land on SHIP.

## Log

### 2026-10-16

### 2026-10-16 — design

Root cause: the active-only existence probe classifies an archived record as an
incomplete first publication; move-detail can then synthesize a new active body
from the card. Resolve by ID against the three canonical issue directories
(active, legacy flat history, history/issues), not by inventing the active path.
Reuse existing pinned TrunkView, issue filename parser, republish transaction,
stateful real-Git fixture and existing ownership/asset/concurrent-body checks
(ARCH-DRY, ARCH-PURE, ARCH-MOCK, ARCH-PURPOSE). Re-resolve per candidate before
publication; a concurrent archive or twin refuses, rather than resurrecting an
active copy (ARCH-ORDER, ARCH-SECURE). A lifetime lock cannot reconcile stale
branch/main locations; the existing repo lock already serializes local verbs.
Metadata scans stay bounded to the existing three directories; fetch blobs only
for the selected ID (ARCH-CONSTRAINTS). No new durable artifact family, process
or handle; existing issue and review archive lifecycle applies (ARCH-FUNERAL).

Short plan posted to ops:0 under the explicitly authorized small-fix flow.
