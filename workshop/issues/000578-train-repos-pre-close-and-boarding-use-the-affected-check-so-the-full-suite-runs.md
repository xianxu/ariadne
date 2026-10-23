---
id: 000578
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '4e441df726e52428c2a8ba724ca14d31159efbc9' # card fields mirrored from issue-cards; edit via sdlc
---

# Train repos: pre-close and boarding use the affected check, so the full suite runs once, in the train

## Problem

## Spec

**Use affected boarding in train repos, so the full suite runs once (operator, 10-23).** ariadne#473 (affected
suite) and ariadne#561 (trains board on affected) landed, but slots still run a full recorded suite before close and
boarding, because the worker guidance says so: tl-worker ("a full suite before a close"; "Run full suites through
`sdlc local-check run --name suite`"; "Run the required recorded pre-close suite") and AGENTS §5 ("run the full suite
before a close"). In pair this week, 1 affected boarding vs. 18 trains: the full suite runs twice (pre-close and in the
train) and trains save nothing. Goal: cut machine CPU/SSD per landing so development goes faster at the same cost.
Source: ops `session.side/2026-10-23-04-landing-queue-health.md`.

Change, for repositories where trains are on and `.sdlc/affected.json` exists (pair today; ariadne stays off):
- The pre-close and boarding check is `sdlc local-check run --name affected` on the pushed head; the train's
  integration suite is the full-suite gate. A full selection still records `suite` as today.
- tl-worker and AGENTS §5 say so, conditioned on the repo's landing config; other repos keep the full pre-close suite.
- If `sdlc close` (or any gate) demands a recorded full suite in such a repo, it accepts the affected record instead.
- A failure fix after a train failure still follows the existing rerun rules.

## Done when

- In a train repo with affected.json, close and boarding pass on an affected record alone, and the train's suite is
  the only full run for that landing (test).
- tl-worker's help and AGENTS §5 state the per-repo rule; repos without trains are unchanged (test or doc check).

## Plan

- [ ]

## Log

### 2026-10-23
