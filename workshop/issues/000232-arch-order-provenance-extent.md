---
id: 000232
status: open
deps: []
github_issue:
created: 2026-09-17
updated: 2026-09-17
estimate_hours:
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

## Spec

Revise `cmd/sdlc/internal/judge/architecture.md`'s ARCH-ORDER entry. Two changes.

**A. Narrow the disclaimer; name the axis.** Stop sending truth-origin questions to
ARCH-SECURE. Distinguish the two provenances explicitly — ARCH-SECURE owns trust
(is this input what it claims), ARCH-ORDER owns truth-origin (does this component
own the fact). State the authority/provenance pair, and name local-authority +
remote-truth as the quadrant requiring a reconciliation policy. The existing
knowledge-and-uncertainty clause becomes the answer the framing points AT rather
than away from.

**B. Add an extent clause.** The modeled extent of an external entity is bounded by
the primitives that entity exposes. Carry the taxonomy:

| can set | can read | treatment |
|---|---|---|
| no | no | **do not model it** |
| no | yes | observation only — never cached across a decision; read at point of use, carry `asOf` |
| yes | yes | controlled proxy — desired + observed + reconcile loop |
| yes | no | write-only — writes must be idempotent, and *unconfirmed* is a permanent state, not a transient one |

with the async-confirm case called out: a read that returns out-of-band makes the
belief permanently stale by one round-trip, which is nearer the write-only row than
the controlled-proxy row. That distinction is what pair#262 turns on.

**Techniques the at-plan lens should ask for**, none of which the entry names today:

- **Level-triggered, not edge-triggered.** Reconcile against observed state, not
  against events. A missed event under edge-triggering is permanent divergence; a
  level-triggered pass repairs it. (The `spec`/`status` split is this distinction
  made structural — one field is authored, the other is written from reality, and
  nothing lets a reader confuse them.)
- **Convergent writes over deltas.** Express desired end state. Under uncertain
  belief "set X to 5" is safe to repeat; "toggle X" is not. This is exactly the
  `render.go` / `parentModeDelta` split above.
- **Epochs / fencing** so a stale observation cannot authorize an action. pair
  already has `mouseEpoch`, `GeometryEpoch`, `Generation` — the technique exists in
  the fleet and is unnamed in the registry.
- **`unknown` as a first-class state with an escape transition** — the missing
  piece in pair#207.

**Out of scope.** No new ARCH-* marker. No change to ARCH-SECURE beyond whatever
cross-reference keeps the two provenances legible from both sides.

## Done when

- ARCH-ORDER no longer disclaims the axis its own knowledge clause governs; a
  reader arriving with "this state proxies an external system" is answered in
  place, not redirected to ARCH-SECURE.
- The entry states the authority/provenance distinction and names local-authority
  + remote-truth as the quadrant requiring a reconciliation policy.
- An extent clause exists carrying the taxonomy, with the "do not model it" row
  explicit and the async-confirm case distinguished from the controlled proxy.
- The at-plan lens asks, per modeled external attribute, which row it is in and
  what primitive justifies it.
- The named techniques appear in the lens (level-triggered, convergent writes,
  fencing, first-class `unknown`) — as questions the plan must answer, not prose.
- **Single-source is preserved (ARCH-DRY / ARCH-PURPOSE).** The registry at
  `cmd/sdlc/internal/judge/architecture.md` remains the one source; verify
  `atlas/workflow/architecture-principles.md` and the AGENTS.md narrative DERIVE
  rather than restate, and fix any hand-maintained copy found — a restatement is a
  deferred consumer.
- `sdlc arch-principles` emits the revision, and the plan-quality + boundary-review
  judges embed it (they already embed the registry; assert it rather than assume).
- **Retroactive check as the worked example:** run the revised lens against
  pair#262's M3 and confirm it produces the finding — that deltaing parent mode
  state over-models a proxy whose confirm path is asynchronous and whose writer is
  not exclusive. If it does not produce that finding, the revision is too weak.

## Plan

- [ ] Draft the two edits against the registry entry; keep them inside ARCH-ORDER.
- [ ] Shadow-sweep the consumers: atlas page, AGENTS.md narrative, judge prompts.
- [ ] Retroactive check against pair#262 M3.

## Log

### 2026-09-17

From an operator design discussion in `brain`. The operator's framing: ARCH-ORDER
established that state needs an authority and a contract to change it, and the
missing distinction is between authority and *the source of the state* — whether it
models internal behavior or proxies an external system — with two consequences,
reconciliation (already covered) and modeled extent (not).

Checked before filing: point 1 IS in the body (knowledge and uncertainty) and is
contradicted by the framing sentence; point 2 is absent from all eight entries.
Hence a touch-up, not a new principle — the operator's call and the right one.

pair#262 is the live worked example and the reason this is worth doing now rather
than as tidy-up: that investigation was actively heading toward the over-modeling
the extent rule forbids, and nothing in the current registry would have stopped it.
