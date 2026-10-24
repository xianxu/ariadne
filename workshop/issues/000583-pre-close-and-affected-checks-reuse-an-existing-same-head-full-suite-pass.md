---
id: 000583
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
card_mirror: '0dcb85063d179cb12a612bf01ba9687da95a0d72' # card fields mirrored from issue-cards; edit via sdlc
---

# Pre-close and affected checks reuse an existing same-head full-suite pass instead of queueing a new full run

## Problem

## Spec

**`--pre-close` reruns a full suite on a head that already has a passing full-suite record (pair:7, ariadne#578
rollout, 10-23).** On pair#739's unchanged head 00758b61, which had a just-completed recorded full-suite pass (with a
known-flake rerun), `sdlc local-check run --pre-close` printed "affected", then the affected selector fell back to a
full selection (a change to `cmd/couch/main_test.go` selects everything) and queued a **new** full suite on the same
head. A full-suite pass on the exact head already covers every affected selection.

Fix: before running, `--pre-close` (and `--name affected`) accept an existing passing `sdlc-local/suite` record for
the exact pushed head (including a known-flake-qualified pass) as satisfying the check, print that it's reusing it,
and run nothing. A record for another head never counts.

## Done when

- With a passing suite record on the exact head, `--pre-close` and `--name affected` run nothing and report reuse;
  without one, behaviour is unchanged (tests).

## Plan

- [ ]

## Log

### 2026-10-23
