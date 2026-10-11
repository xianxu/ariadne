---
id: 000333
status: open
deps: []
github_issue:
created: 2026-10-11
updated: 2026-10-11
estimate_hours:
card_mirror: 'b5b5b2ba716c4410df1a73494a140febcdb6cea9' # card fields mirrored from issue-cards; edit via sdlc
---

# Full-suite lock: a sandboxed make test should fail its preflight before queueing for the lock

## Problem

pair#450 added `tests/host-preflight.sh` as `make test`'s first prerequisite: it probes unix sockets, PTYs and `exec ps`, and fails naming the agent sandbox instead of letting ~150 tests fail looking like code bugs. But the base layer's full-suite lock (#319) is taken when make parses the Makefile, before any prerequisite runs. So a sandboxed `make test` first waits for the shared machine lock (177–1452 s per run on 10-11, ariadne#330's Log) and only then fails its preflight.

## Spec

Give the base layer a pre-lock hook: a repo-provided preflight (convention: `tests/host-preflight.sh` if present) that runs before the full-suite lock is taken; if it fails, `make test` exits with its message and never queues. ariadne gets the same probe set if its suite also needs sockets/PTYs/ps (it does: #330's processgroup tests).

## Done when

- A sandboxed `make test` in pair fails within seconds with the preflight's message, without taking or waiting for the lock (test or a scripted check).
- An unsandboxed run takes the lock as today.

## Plan

- [ ] Pre-lock hook in the base-layer make include; wire pair's preflight; add one for ariadne; check both.

## Log

### 2026-10-11
