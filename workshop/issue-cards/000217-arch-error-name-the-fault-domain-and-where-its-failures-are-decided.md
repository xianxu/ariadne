---
id: 000217
status: open
created: 2026-09-07
updated: 2026-09-07
estimate_hours:
github_issue:
---

# ARCH-ERROR: name the fault domain and where its failures are decided

## Problem

The registry has seven entries and **none asks which layer owns a failure or
what it does about it.** Error handling appears only incidentally:

- `ARCH-SECURE:143` — "the failure path degrades visibly rather than crashing or
  substituting a fabricated value", scoped to *parsing untrusted input*.
- `ARCH-ORDER:173,189` — who is cancelled when this fails; an error path that
  unwinds the sequencing but drops the in-flight effect. That is *state
  consistency within a domain*, not placement across domains.
- `ARCH-SECURE:140` — credentials must not reach an error message.

`AGENTS.md`'s "Root Cause — no temp fixes or lazy null checks" is adjacent but
is about not papering over a fault, not about where the fault is decided.

**The agent failure mode this targets is specific**, which is what makes it
registry-worthy rather than style advice: LLMs write `try`/`catch` at the
**detection** site, log, and continue — destroying the information the caller
needed and converting a loud failure into a silent wrong answer. In Go it
surfaces as `if err != nil { return err }` chains that add no context, so the
layer that could decide receives something it cannot act on. Both shapes pass
review and fail in production.

### The cost, measured in this fleet

`pair`'s couch has a supervision policy. It was never written down, so it was
rediscovered one defect at a time — **four rounds of the same defect class**
(BR-16 the missing re-assert, BR-22 the write, BR-26 the observation's shape,
BR-33 the belief, the last from an operator report as `pair#196`). The policy
that would have short-circuited all four is one sentence, and the fix that
finally landed states it almost verbatim:

> Unknown means stand back: couch forgoes a keyboard-reachable convenience,
> where the alternative costs the operator their editor's drag, which no
> keystroke recovers.

That is a supervisor choosing to lose its own feature rather than risk state it
cannot reconstruct. Stated as policy it is derivable; discovered per-defect it
cost four rounds. `pair#209` is the same shape a fifth time on a different
subsystem — a switch reconstructs the screen from a bounded ring and has no
designed behaviour for "the ring was insufficient".
