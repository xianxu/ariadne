---
id: 000333
status: working
deps: []
github_issue:
created: 2026-10-11
updated: 2026-10-13
estimate_hours:
card_mirror: 'd55991b7c78cbc520927345cd76032ec4d43fabb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T22:34:42-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "83c3fee6", done: "09127adc"}
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
