---
id: 000493
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
---

# golint: guard process-global allocation oracles in ordinary tests

## Problem

pair#724 reproduced a strict allocation oracle flake with the parser unchanged: testing.AllocsPerRun measured 6 small and 26 large allocations when a channel handshake scheduled exactly 20 unrelated allocations per large call. The API reads process-global runtime.MemStats.Mallocs. GOMAXPROCS(1) does not exclude other goroutines, so an ordinary package test can misattribute foreign work. The fix isolates sampling in a self-exec helper and verifies that the sample PID matches the exact measuring child, while retaining the strict small/1MiB/2MiB comparison and an unbounded-parser negative control.

The worker rules ask for a reusable lint defense for a mechanical API misuse. Current local inventory also finds AllocsPerRun in pair cmd/internal/threadactivity/idle_decode_invariant_test.go and cmd/internal/couchtty/menu_perf_test.go; no failure in those two is asserted here.
