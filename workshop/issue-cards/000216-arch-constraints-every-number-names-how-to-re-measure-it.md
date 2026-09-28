---
id: 000216
status: open
created: 2026-09-06
updated: 2026-09-06
estimate_hours:
github_issue:
---

# ARCH-CONSTRAINTS: every number names how to re-measure it

## Problem

`ARCH-CONSTRAINTS` asks a plan to give each relevant constraint "a budget/range,
**basis** (measured fact, requirement, domain-informed assumption, or operator
choice), and bounded behavior when exceeded". At review it asks that
"representative measurements or tests exercise the relevant environment and
workload".

Both are right and neither survives contact with time, because the basis is a
**label**, not a **procedure**. "Measured fact" records that someone once
measured something; it does not say how to measure it again. So the envelope is
checkable exactly once — at the gate that produced it — and thereafter decays
into aspiration: a number in a document that nothing can confirm or refute.

Two structural reasons this is not fixable by applying the existing lens harder:

**1. Regressions frequently have no diff.** A gate judges a change. The
motivating defect (`pair#202`) was correct code: reading a ~700-byte file on a
keystroke is fine. It became a defect months later when an LRU reached its
1000-entry cap through *usage*, with no commit involved. No at-review pass can
catch a regression that never appears in a diff.

**2. An absolute number does not transfer between machines.** A budget stated as
"< 50 ms" is either wrong on some host or so loose it asserts nothing. What
transfers is the *procedure* and the *relative* facts it produces — counts, and
before/after on one host.

### Why this is an amendment and not a new principle

The proposal that opened this was a new `ARCH-PERFORMANCE`. It is not needed and
would violate `ARCH-DRY`: `ARCH-CONSTRAINTS` already names the motivating
defects almost verbatim in its at-review lens —

- "blocking optional work on a critical UI path" → `pair#201`
- "repeated expensive work that should be cached or incremental" → `pair#202`
- "unbounded concurrency or fan-out" → `pair#203`

— and its at-plan lens already names `keystroke` as an interaction path to
classify. The principle is present and sufficient. What is missing is that its
output has no durable home and no repeatable procedure.

### Evidence: a day of measurements that were all wrong

From the `pair` session that motivated this (2026-09-06), every wall-clock number
taken moved by 2–8× purely with ambient activity that was not being recorded:

| measurement | busy host | quiet host | ratio |
|---|---|---|---|
| `zellij action` round-trip | 145 ms (max 467) | 17.6 ms | 8× |
| `alt+Return` (two such calls) | ~290–930 ms | 35 ms | 8–26× |
| process wake-up delay | 4.92 ms | 2.50 ms | 2× |

A threshold written from any one of those readings would be wrong. Worse, a
*controlled* experiment in that session returned a confident null — 12 CPU
burners moved wake-up delay 2.51 → 2.21 ms, i.e. not at all — because the load
shape was unrepresentative (steady CPU burn, where the real workload was a
process-spawn storm). The number was reproducible and meaningless.

The facts from that session that *did* hold under every condition were counts:
`alt+Return` spawns two subprocesses; N keystrokes cause N file reads; N sessions
request N×12 build threads. None needed a stopwatch, and none varied.
