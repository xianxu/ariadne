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

- [x] Pre-lock hook in `Makefile.workflow`: when `test` is a goal, run
  `$(WF_TEST_PREFLIGHT)` (default `tests/host-preflight.sh`; a repo may set it before the
  include) before the lock; on a non-zero exit, `$(error …)` without calling the lock
  helper. Output goes to stderr (make's `$(shell)` would flatten it); only the status is
  read. A repo without one is unaffected.
- [x] pair: wired by the convention alone (it ships `tests/host-preflight.sh`); its own
  `make test` prerequisite still runs it again after the lock, which is harmless.
- [x] ariadne: no preflight, by decision (see Log).
- [x] Tests in `scripts/test/test-lock.test.sh` (CI runs it); atlas.

## Log

### 2026-10-11

### 2026-10-13 — ariadne:3

Done-when 1, live: in this agent sandbox, with the full-suite lock held by another run,
`WF_TEST_PREFLIGHT=<pair>/tests/host-preflight.sh make -n test` failed in 0 seconds with
pair's own message (unix socket bind: Operation not permitted; pty open: out of pty
devices; exec ps: Operation not permitted) and "not taking the full-suite lock", and no
[test-lock] wait. Done-when 2: a passing preflight goes on to the lock (times out on the
held one, takes the free one). Five cases in test-lock.test.sh; against main's
Makefile.workflow the failing-preflight cases fail.

Decision, revising the Spec's "ariadne gets the same probe set if its suite also needs
sockets/PTYs/ps (it does: #330's processgroup tests)": #330 then replaced those `ps`
probes with an exec-free kernel query (`testfix.AwaitExit`) and fixed the EPERM handling,
so ariadne's suite now passes in the agent sandbox (every sandboxed full suite this slot
ran today). pair's probes would refuse here and block runs that work. ariadne keeps the
opt-in hook with no preflight; one belongs here only when a test needs what the sandbox
refuses.

## Revisions

- 2026-10-13 — Spec: "ariadne gets the same probe set if its suite also needs
  sockets/PTYs/ps (it does…)". It no longer does (#330), so ariadne gets no preflight.
