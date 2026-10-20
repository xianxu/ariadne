---
id: 000508
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b4362861439cee38147be832e6865ed66426ee56' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-worker: verify agent-facing changes on every agent kind; live bugs land with a recorded replayed fixture

## Problem

Live-only bugs on 10-20 (pair#727, #735, #742, the drain) were each found by hand, and pair#717's live check passed
on the agent it was built for and failed on Codex. Operator decisions (ops side doc
`2026-10-20-19-conformance-structure.md`): every agent kind, every time; a live bug's fix lands with its recorded
capture as a replayed fixture.

## Spec

Add to `tl-worker --help` (verification section):
- An agent-facing change is verified against **every agent kind** (Claude and Codex; Qoder when it has credits), in
  its tests and in any live check.
- A bug found live is fixed together with a **recorded fixture** of the real capture (terminal bytes, agent screen,
  tool reply) replayed in the normal suite, so it can't regress unnoticed.
- Live tests use the shared disposable-Couch harness (pair#744), never the live fleet.

## Done when

- `tl-worker --help` carries the three rules; a test pins the section.

## Plan

- [ ]

## Log

### 2026-10-20
