---
id: 000219
status: open
deps: []
github_issue:
created: 2026-09-09
updated: 2026-09-09
estimate_hours:
---

# close_test flag assertion takes the real repo lock, so the suite cannot run inside an sdlc transaction

## Problem

`TestCloseCmd_MilestoneFlagHidden` (`cmd/sdlc/close_test.go:131`) executes a real
`NewCloseCmd()` to assert that `--milestone` is hidden and that `close` redirects
to `milestone-close`. That is a **flag-visibility assertion**, but executing the
command takes the production repo transaction lock at `.git/sdlc.lock`.

**Consequence: `go test ./cmd/sdlc/` cannot report green from inside any sdlc
transaction.** Every mutating verb holds that lock, so a suite run launched from
within `change-code`, `close`, or `merge` blocks in `repolock.Acquire`
(`repolock.go:210`) until `DefaultWaitTimeout` — 30 minutes (`repolock.go:19`).

Two ways this bites, and the second is worse:

1. **The gates cannot verify what they ask for.** A close gate that demands
   `go test ./cmd/sdlc/...` as evidence runs it inside the transaction that
   forbids it succeeding.
2. **It presents as a hang, not a lock wait.** Found during ariadne#218's
   plan-quality review: the reviewer saw `go test` blow its 600s default with a
   goroutine parked in syscall for nine minutes and reasonably read it as a hung
   child process. From a plain shell the same suite passes in 115-136s, so the
   two observations look irreconcilable until the lock is suspected. A defect that
   masquerades as a different defect costs more than its size.

Note `-timeout 1800s` does not rescue it: 30m equals `DefaultWaitTimeout`
exactly, so the two expire together.

## Spec

Two candidate fixes; they are not exclusive and the second may be the real one.

**A — make the test hermetic.** The assertion needs no lock: it is about flag
registration and a refusal message. Either drive the refusal without
`cmd.Execute()`, or point the lock at a temp git dir. Smallest change, fixes the
suite.

**B — refuse before acquiring (likely the root cause).** A verb that will reject
its arguments should not first take a global lock that waits half an hour.
Cheap, local refusals belong ahead of expensive, shared acquisitions — the same
ordering ariadne#209 BR-34 established for `guardSpineRepo` preceding flag
validation, and the reason is identical: the operator should get the real
complaint, not a symptom of machinery they never reached.

If B is right the test bug is a symptom, and fixing only A leaves the product
behavior (a 30-minute wait to be told a flag is wrong) in place. **Decide which
before implementing** — this issue exists partly because the two were confusable.

**Out of scope:** the lock design itself, which is working as intended
(`sdlc --help`, LOCAL REPO TRANSACTION LOCK).

## Done when

- `go test ./cmd/sdlc/ -count=1` reports green (modulo the pre-existing
  ariadne#210 failure) **when run from inside an sdlc transaction** — the
  condition that fails today. A plain-shell run already passes and proves
  nothing here.
- No test in `cmd/sdlc/` acquires `.git/sdlc.lock`; a check that would have
  caught this, rather than only the one instance.
- If fix B is taken: `sdlc close --issue N --milestone Mx` refuses on the flag
  without waiting on the lock, asserted by a test that holds the lock and expects
  a prompt refusal.


-

## Plan

- [ ]

## Log

### 2026-09-09
