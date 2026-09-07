---
id: 000216
status: open
deps: []
github_issue:
created: 2026-09-06
updated: 2026-09-06
estimate_hours:
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

## Spec

Amend `ARCH-CONSTRAINTS` in `cmd/sdlc/internal/judge/architecture.md`. Two small
deltas; no new marker, no change to the principle's subject.

**1. at-plan — a `measured fact` basis must name its probe.** Where the basis is
measurement rather than requirement or choice, the plan states *how to take the
measurement again*: the command or test that produces it, and the environmental
facts that qualify it (for this fleet: co-tenancy — how many agents were working
— which `ARCH-CONSTRAINTS` already lists under "target environment and
co-tenancy" but which is routinely omitted from the reading itself).

**2. at-plan — prefer an invariant that does not vary with the host.** Where the
constraint can be expressed as a count, a ratio, or an asymptotic shape ("one
round-trip per submit", "one read per change, not per keystroke", "fan-out
bounded by cores, not by sessions"), state it that way *in addition to* any
timing. A counted invariant is machine-independent, cheap, and cannot flake; a
threshold is neither portable nor stable. Timings remain useful as a
regression signal — they are not the promise.

**3. at-review — flag a budget with no way to re-take it.** A stated number
whose basis is measurement, with no named procedure, is aspiration; say so and
ask for the probe. Equally, flag a *threshold* asserted as portable when the
underlying reading is host- and load-dependent.

Where the envelope should live is out of scope here and belongs to the consumer
repo — the `target` datatype already exists for "an invariant worth defending
from drift", and `pair#205` is the first instance. This issue changes only what
the gate asks for.

Explicitly not in scope: mandating a benchmark harness, a CI job, or a numeric
threshold anywhere. The point is that a number carries its procedure — not that
every plan must grow a benchmark. `N/A` remains available and the anti-ceremony
clause stands.

## Done when

- `ARCH-CONSTRAINTS` carries the three deltas above, and the registry still has
  **six** entries — no `ARCH-PERFORMANCE`, no seventh marker.
- The entry still satisfies the per-entry contract check
  (`architectureEntry` + the `principle:`/`at-plan:`/`at-review:` bullet shape).
- `TestArchitectureMarkers`' hand-written list is untouched — this change adds no
  marker, so a diff there means the change was done wrong.
- Goldens re-captured for the three prompts that embed the registry.
- `atlas/workflow/architecture-principles.md` records the amendment and the
  reason a separate performance principle was rejected, so the same proposal is
  not re-raised.

## Plan

- [ ] Amend the `ARCH-CONSTRAINTS` at-plan and at-review bullets per Spec.
- [ ] Re-capture goldens; confirm no marker list changed.
- [ ] Update `atlas/workflow/architecture-principles.md` with the amendment and
      the ARCH-PERFORMANCE rejection rationale.
- [ ] Verify: `sdlc arch-principles` still renders 6; `go test ./cmd/sdlc/...`.

## Log

### 2026-09-06

Originated in a `pair` performance debugging session (`pair#201`, `#202`,
`#203`). The operator proposed `ARCH-PERFORMANCE`; checking the registry first
showed `ARCH-CONSTRAINTS` already covers the subject and would have flagged all
three defects, so the gap is mechanism rather than principle.

The operator's framing is what set the amendment's shape: *"given different type
of machines and such, I think it matters more to be able to run things to
measure, than necessarily hard limits."* That is the difference between a basis
*label* and a basis *procedure*, and it is why delta 2 (prefer host-independent
invariants) sits alongside delta 1 rather than being replaced by a threshold
requirement.

The supporting evidence is unusually direct: every timing taken during that
session was wrong by 2–8×, including one confidently-reported null from a
controlled experiment whose load shape did not represent the real workload. The
counted facts from the same session held under every condition. That asymmetry
is the argument.
