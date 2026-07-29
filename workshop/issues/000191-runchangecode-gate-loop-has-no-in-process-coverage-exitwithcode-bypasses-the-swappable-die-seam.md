---
id: 000191
status: open
deps: []
github_issue:
created: 2026-07-29
updated: 2026-07-29
estimate_hours:
---

# runChangeCode gate loop has no in-process coverage — exitWithCode bypasses the swappable die seam

## Problem

`runChangeCode`'s gate loop cannot be tested in-process. It iterates `changeCodeGates` and
calls `exitWithCode(1)` on any gate failure (`changecode.go` → `term.go`), which reaches
`os.Exit` — so a test that drives it dies rather than returning. `expectDie` does not help:
it swaps the `die` var, which this path bypasses.

Concretely untested: the `--force` continuation branch (`changecode.go:134-138`), the one
place where a gate failure is deliberately *not* fatal. It decides whether a forced
invocation proceeds, and nothing exercises it.

Raised at ariadne#187 M1's boundary review, where promoting `exitWithCode` to a swappable
var was the suggested fix. Deferred then for a good reason — changing a process-exit seam
inside a FIX-THEN-SHIP bundle is not a safe bundling — and the actual risk of that milestone
(the emitted ACK strings) was covered by `TestForceAckMatchesGateCatalog`. #187's close
review raised it again with the same conclusion: worth its own issue rather than a third
deferral note.

## Spec

- `exitWithCode` becomes swappable in the same style as `die`, so the gate loop is drivable
  in-process. One seam, matching the existing convention rather than inventing a second.
- Coverage for the two branches that currently have none: gate failure exits non-zero and
  stops the sequence; gate failure **with** `--force` records the bypass and continues to the
  next gate.
- The change is to the seam only — no gate's behavior moves. A test asserting the CURRENT
  order/behavior should pass unchanged after the refactor, which is how we know it was a
  refactor.

## Done when

- `exitWithCode` is swappable in the same style as `die`, so `runChangeCode`'s gate loop is
  drivable in-process.
- Two currently-uncovered branches have tests: an unforced failing gate stops the sequence
  and exits non-zero; a **forced** failing gate records the bypass and reaches the NEXT gate
  (`changecode.go:134-138`, the branch that decides whether a forced run proceeds).
- **It is verifiably a refactor:** `TestForceAckMatchesGateCatalog` and the gate-order guard
  pass UNCHANGED, and no gate's behavior moves.
- No new process-exit path: one seam, matching the existing convention rather than adding a
  second way to exit.

## Plan

- [ ] Promote `exitWithCode` to a swappable var mirroring `die`
- [ ] Failing test: a forced invocation past a failing gate reaches the NEXT gate
- [ ] Failing test: an unforced failing gate stops the sequence and exits non-zero
- [ ] Confirm `TestForceAckMatchesGateCatalog` and the gate-order guard still pass unchanged

## Log

### 2026-07-29
