---
id: 000191
status: open
created: 2026-07-29
updated: 2026-07-29
estimate_hours:
github_issue:
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
