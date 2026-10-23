---
id: 000577
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '0e9a6f883227bf796450c42f7c0003f5d987a2cb' # card fields mirrored from issue-cards; edit via sdlc
---

# model-bench M1: bench runner (leak-free checkout, run, measure, score, ledger) and a tl-lead bench mode

## Problem

## Spec

**model-bench M1: the bench runner (observer tool) and a bench mode for tl-lead**, per the operator-approved design
in ops `teams/devinfra/projects/model-bench/project.md` (10-23). Three roles on a separate computer: couch-bench (the
observer, mostly this tool), ops-bench:0 (the TL hub, one fixed model, normal tl-lead rules plus a bench mode), pair:0
(the worker, the model under test). M1 runs **one autonomous corpus issue end to end with one worker model**.

The runner (a CLI in ariadne, run by the observer):
- **Prepare** a leak-free task checkout: the issue's parent commit, later history, remote and the archived issue
  removed; the issue as originally filed (Spec and Done-when only).
- **Start and watch** a run: hand the issue to the bench TL hub, poll to a terminal outcome ("would land": full suite
  and close review pass, nothing merges; or stopped), with a wall-clock budget.
- **Measure:** wall time, operator-wait time (notification to reply), worker and TL tokens (from the agents'
  transcripts), review rounds and findings, gate refusals, and the TL's own decisions (re-plans, rulings, stops).
- **Score:** apply the landing's own added tests (hidden acceptance) to the result.
- **Ledger:** append one row per run to a single-writer ledger in ops (keyed by repo, task type, worker model, TL
  model), plus the raw run record.

The bench mode for tl-lead: the TL never codes, approves or re-plans by its normal rules, logs each decision in a
form the runner collects, and never lands.

Out of M1: the second computer's setup, more than one worker model, the report.

## Done when

- One corpus issue (an autonomous one, e.g. parley.nvim#403) runs end to end on this Mac in a disposable setup,
  producing a ledger row with every field above and a hidden-acceptance score (test or recorded run).
- The prepared checkout provably contains no trace of the landed solution (test: later commits, the archived issue
  and the remote are absent).

## Plan

- [ ]

## Log

### 2026-10-23
