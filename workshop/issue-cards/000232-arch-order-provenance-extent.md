---
id: 000232
status: open
created: 2026-09-17
updated: 2026-09-17
estimate_hours:
github_issue:
---

# ARCH-ORDER: separate provenance from authority, and bound the modeled extent of external state

## Problem

**This is a touch-up to ARCH-ORDER, not a ninth principle.** The body already
carries most of the work; the framing sentence disowns it, and one clause is
missing outright.

### 1. ARCH-ORDER disclaims in its framing what it does in its body

The framing sentence sends the reader away:

> "The lens is *temporal*, not provenance-based — ARCH-SECURE makes invalid state
> unrepresentable at a single-shot parse of input the component did not produce;
> this makes it unrepresentable in the state a component carries between events."

But ARCH-SECURE's provenance is a **different axis**:

> "Trust is a property of provenance, not of location: an input crossing a
> process, session, or version boundary is untrusted even when this same program
> wrote it"

That is *integrity* — is this byte stream what it claims to be. It is not
*truth-origin* — does this component own the fact, or is it holding a proxy for
someone else's. "Don't believe it because you wrote it" and "this isn't yours to
know" are different claims with different remedies.

Meanwhile ARCH-ORDER's own **knowledge and uncertainty** clause does the
truth-origin work already, and does it well:

> "Model what the component knows about an external entity, distinguishing
> desired state, observed state, and operation outcome... External operations
> define confirmed success, confirmed failure, and unconfirmed outcome; each has
> permitted actions and a reconciliation policy... An observation is evidence
> about an exact entity at a point in time, not permanent authority to act on it."

So a reader asking *"how do I treat state that proxies an external system"* meets
a disclaimer pointing at the wrong entry, three paragraphs above the answer.

The missing distinction in one line: **authority is who may change the state;
provenance is what makes it true.** A component can hold authority over a proxy
whose truth lives elsewhere. That quadrant — local authority, remote truth — is
the one that needs a reconciliation policy, and it is the one the framing hides.

### 2. Nothing in the registry bounds the EXTENT of a model

All eight entries address how to hold state. None asks **how much** external state
to model. That is genuinely unowned, and it is the half with no home anywhere.

The rule the fleet keeps rediscovering: **a model should be closed under the
operations you can actually perform.** A modeled external attribute earns its
place only if you can set it, or must branch on it. Otherwise it is a liability
with no matching capability — a field that can diverge, that you will write
reconciliation for, and that buys nothing.

### The cost, measured in this fleet

- **pair#207** — couch's asserted mouse mode had no release path short of
  restarting couch. A belief that reached `unknown` with no terminal transition
  out: a repair state machine missing its own funeral.
- **#217 already records four rounds of this defect class** in couch alone
  (BR-16 the missing re-assert, BR-22 the write, BR-26 the observation's shape,
  BR-33 the belief), resolved by one sentence — *"Unknown means stand back."*
  Four rounds is the registry reporting an unwritten policy.
- **pair#262 (2026-09-16) — the live one, and the sharpest.** The parent terminal's
  mode state is a proxy. In ONE file, two different treatments of it coexist with
  neither justified against the available primitives: `render.go:35` re-asserts
  the whole preamble every frame, while `presenter.go:287` (`parentModeDelta`)
  deltas mouse modes from believed state. The investigation's first instinct was to
  extend the delta treatment to the preamble — and the extent rule would have
  stopped it at plan time, because:
  - confirming parent mode state requires an **asynchronous DECRQM round-trip**
    whose reply returns through the input stream, so the belief is permanently one
    round-trip stale; and
  - `hostty.Reservation` writes `SetRegion` and the scroll-region reset to the same
    terminal, so the belief is not exclusive either — while
    `hostty/control.go:27` asserts that sequence *"lives here and only here"* and
    `terminal/render.go:35` + `history_render.go:313` both emit it.

  Under the taxonomy below, parent terminal modes are at best *settable, confirmable
  only asynchronously* — which makes `render.go`'s convergent re-assert the
  CORRECT treatment for the class, and the delta an over-modeling. #255's plan
  calling `ParentPresenter` the *"exclusive typed parent-output door"* is the
  aspiration; the second writer is the fact.
