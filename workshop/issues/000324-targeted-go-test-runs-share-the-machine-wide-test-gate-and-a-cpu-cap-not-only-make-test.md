---
id: 000324
status: open
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: 'e3c0c5a1d3e99aaeb91d08e9121bbb337b17d75a' # card fields mirrored from issue-cards; edit via sdlc
---

# Targeted go test runs share the machine-wide test gate and a CPU cap, not only make test

## Problem

ariadne#319's machine-wide lock covers `make test` only. Agents iterate with targeted `go test ./cmd/sdlc/ -run …`, which bypass it. On 10-10 at ~17:40 (pair:4's diagnosis, ops learnings kinks 53–54), two such runs from two ariadne slots ran at ~140% CPU each for 14+ minutes, forking git heavily; load hit 72–87 on 12 cores. Effects: Couch exited (its terminal stopped reading output for 5s, pair#430), and ariadne tests with 2s subprocess deadlines failed 4/4 (#325). `GOFLAGS=-p=3` limits packages only; in-package `-parallel` still defaults to GOMAXPROCS=12 per run, per slot.

## Spec

Every Go test run an agent starts in a woven repo goes through one machine-level budget, not only `make test`:
1. Cap per-run parallelism in the slot environment that weave or the launcher sets (e.g. `GOFLAGS='-p=2 -parallel=2'`, and GOMAXPROCS of about NCPU divided by active slots). Pick the mechanism that reaches agents' direct `go test` calls.
2. Make targeted runs take the #319 lock or a counting semaphore (N concurrent runs machine-wide), with the same loud waiting and escape hatches as #319.
3. Measure before and after: load average and test wall time with 4 slots iterating.

## Done when

- With 4 slots each running targeted `go test` in a loop, load stays under about 2× NCPU, and Couch and the tests stay up.
- A direct `go test` from an agent shell honors the cap (test: env inspection or a fake runner).
- The escape hatch is documented and loud.

## Plan

- [ ]

## Log

### 2026-10-10
