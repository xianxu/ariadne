---
id: '000215'
status: done
started: 2026-09-04T21:49:32-07:00
created: 2026-09-04
updated: 2026-09-05
estimate_hours: 1.49
actual_hours: 0.61
---

# ARCH-ORDER architecture principle

## Problem

The registry has six entries and none of them asks *what are the orderings, and
where does nondeterminism enter*. The gap is real, not a restatement:

- `ARCH-PURE` names `clock` in its IO list, but its lens is **business logic vs.
  IO** — "don't bury logic in handlers." An agent reading it will separate a
  handler from a transform and never think to enumerate transitions. IO breaks
  purity by *doing* something; concurrency breaks it by making *the order of what
  already happened* unknowable from the text. Different failure, different tell.
- `ARCH-CONSTRAINTS` mentions concurrency only as a **resource budget**
  ("unbounded concurrency or fan-out") — a capacity question, not a correctness
  structure.
- `ARCH-MOCK` covers external-service determinism but is silent on the scheduler,
  the clock's *ordering* role, and event arrival.

Why it matters more for agents than for humans, and why a gate is the right
instrument:

1. **No local surface.** Generation is linear; a program's text has one reading
   order and its executions have many. The happens-before edges live in the
   scheduler, not on the page, so there is no position in the token stream where
   "and the user can cancel here" is representable. It never gets emitted.
2. **The prior is diffuse exactly where the stakes are high.** LLMs fill spec
   holes well when the answer is modal across all codebases ("what if the list is
   empty"). Interleaving policy is not modal: cancel vs. queue vs. preempt vs.
   ignore, roll back vs. journal-and-resume, is a domain-economics decision that
   differs per product. The model samples *a* mode silently and independently at
   each site, so the defect is an inconsistent policy scattered across the tree
   rather than a missing behavior anyone would notice.
3. **Conventional UX masks it.** "Disable the button, show a spinner" *is* modal,
   and the model applies it unprompted. But it is advisory and UI-local: it does
   not prevent process death, connection loss, session close, a retry, a second
   actor, or out-of-order completion. It removes the *common* interleaving —
   which is why everyone agrees on it, and why it is the cell nobody was going to
   get hurt by. It makes the residual bug rare rather than absent, and thereby
   destroys the smoke-test signal that would have caught it.
4. **The oracle is broken.** A human smoke test is a sampler over interleavings
   with sample size 1 and no coverage report. It cannot distinguish "correct"
   from "got the lucky schedule," so it does not merely fail to catch — it
   actively confirms whichever policy the model sampled. Biased-toward-passing
   feedback is worse than none, and it is why the normal agentic write→run→fix
   loop degenerates here.

Consequence for the entry's shape: the at-plan lens must **not** spend the
operator's attention on cells conventional UX already answers. It should point
only at the events the caller cannot block, where the prior is genuinely diffuse.

The argument above is about LLMs in general; the evidence is from this fleet. The
origin is the many rounds worked through against couch's models when starting up
pair, and `pair#182`/`#185` supply the cases: one feature designed correctly
*because* it was treated as an ordering problem (couch's park-then-resume, four
named outcomes each with its own recovery, no defects in that part), four defects
that were ordering defects found late, and — on the reviewer, not the reviewed —
`go test -race` recorded as close evidence from a single run of a test that fails
3 in 10. Point 4 above is a transcript, not a theory. Detail and per-defect
attribution: `## Revisions`, second round.
