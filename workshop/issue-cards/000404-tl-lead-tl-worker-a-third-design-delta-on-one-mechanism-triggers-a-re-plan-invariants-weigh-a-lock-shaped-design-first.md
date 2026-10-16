---
id: 000404
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
---

# tl-lead/tl-worker: a third design delta on one mechanism triggers a re-plan; invariants weigh a lock-shaped design first

## Problem

2026-10-15: the TL approved eight design deltas to pair#584, one at a time, each reasonable alone (argv for every OS
process, legacy rows, unbound launches, ledger caps, codec speed, concurrency), without asking why one guard needed
eight. The guard (pair#556) proves "no live writer" by forensics over historical state, about 1,700 lines with ~45
refusal reasons; a wrapper-held flock that dies with its process does the same in one step with two outcomes
(pair#595). The operator: "the same old over engineering." #556's plan never weighed a lock.
