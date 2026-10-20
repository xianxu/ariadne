---
id: 000493
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ae1d50a7b84e05bf6563067b908b513ba247f9e6' # card fields mirrored from issue-cards; edit via sdlc
---

# golint: guard process-global allocation oracles in ordinary tests

## Problem

pair#724 reproduced a strict allocation oracle flake with the parser unchanged: testing.AllocsPerRun measured 6 small and 26 large allocations when a channel handshake scheduled exactly 20 unrelated allocations per large call. The API reads process-global runtime.MemStats.Mallocs. GOMAXPROCS(1) does not exclude other goroutines, so an ordinary package test can misattribute foreign work. The fix isolates sampling in a self-exec helper and verifies that the sample PID matches the exact measuring child, while retaining the strict small/1MiB/2MiB comparison and an unbounded-parser negative control.

The worker rules ask for a reusable lint defense for a mechanical API misuse. Current local inventory also finds AllocsPerRun in pair cmd/internal/threadactivity/idle_decode_invariant_test.go and cmd/internal/couchtty/menu_perf_test.go; no failure in those two is asserted here.

## Spec

Add a focused pkg/golint test-oracle rule detecting ordinary in-process uses of testing.AllocsPerRun where the test claims isolation from unrelated package work. Require an explicit, reviewable isolated-helper contract rather than assuming GOMAXPROCS(1) isolates the counter. The pair#724 self-exec helper and child-PID proof provide the motivating safe shape; direct ordinary test calls and a trivial helper wrapper that still runs in the same process are negative controls. Do not add a broad path/package allowlist or force tolerance increases. Decide the smallest enforceable syntax/annotation contract at planning; do not pretend static analysis can prove arbitrary goroutine quiescence. Audit current callers and report any disposition needed before enabling downstream enforcement.

## Done when

- A counted lint fixture rejects the known unsafe ordinary-test shape and a trivial same-process wrapper.
- The documented isolated-helper contract is accepted and cannot be enabled by a filename/path allowlist alone.
- Existing callers are inventoried/dispositioned without loosening allocation bounds; rule tests and normal repository checks pass.

## Plan

- [ ]

## Log

Filed from pair:7 while fixing pair#724, unclaimed for the owning team. No peer source changes. The root probe and implementation live in pair#724; this is the shared lint defense only.
