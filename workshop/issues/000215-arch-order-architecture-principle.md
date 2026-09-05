---
id: 000215
status: working
deps: []
github_issue:
created: 2026-09-04
updated: 2026-09-04
estimate_hours:
started: 2026-09-04T21:49:32-07:00
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

## Spec

Add a seventh registry entry to `cmd/sdlc/internal/judge/architecture.md`. Named
`ARCH-ORDER`, not `ARCH-FSM` — naming the *deliverable* invites the cargo cult of
state machines over map/filter; naming the *question* does not. The trigger is
narrow (durable state + events from outside), and the entry carries an explicit
`N/A` clause, matching the gating idiom already set by `ARCH-CONSTRAINTS` and
`ARCH-SECURE`.

Draft entry text (final wording is the implementer's; the clauses are the spec):

```markdown
## ARCH-ORDER — Make the ordering explicit

- **principle:** When a component holds state across events arriving from
  outside it, the legal states and the transitions between them are part of the
  design, not an emergent property of the code. Model it as an explicit
  `(state, event) -> (state, effects)` enumeration rather than a constellation
  of independent flags whose legal combinations are unwritten. Where ARCH-PURE
  separates logic from IO, this separates **legal** state from **representable**
  state: N boolean/nullable fields declare 2^N states and usually mean about
  five, and the bugs live in the difference. Concurrency does not need a new
  effect abstraction — it needs the interleaving space written down where a
  reader, a type checker, and a test generator can all see it.
- **at-plan:** For a component with durable state and external events, enumerate
  states x events -> (state, effects), including the interrupting ones. Do not
  re-litigate what conventional UX already settles (block input, show a
  spinner) — that policy is advisory and UI-local. Spend the enumeration on the
  events the caller **cannot** block: process death mid-transition, connection
  loss, session/tab close, retry of an already-applied step, a second actor on
  the same state, completions arriving out of order. For each, say which of
  cancel / queue / preempt / ignore applies and whether state rolls back. Name
  the extent of any concurrent work — who is still running when this returns,
  who is cancelled when this fails. Name where nondeterminism enters (clock,
  scheduler, IO completion order, arrival order) and how a failing ordering is
  reproduced. Mark `N/A` for pure transforms and request/response paths that
  hold no state between events; do not fill a ceremonial checklist.
- **at-review:** Flag state whose transition set cannot be read off the code —
  boolean/nullable constellations with unwritten legal combinations, where
  `isLoading && err != nil && data != nil` is reachable and undefined; name the
  tagged enum they should collapse into. Flag concurrent work whose extent is
  not lexically bounded (a spawned task outliving its scope, no cancellation
  path). Flag an error or cancellation path that unwinds the sequencing but
  drops the in-flight effect. Flag tests that can only observe one interleaving
  — no seam to inject ordering, no way to reproduce a reported failure.
```

Deliberately **not** in scope: prescribing an effect system. The discipline
(name the effect, keep it in the signature, push it out of the core) is what
transfers; the encoding (`>>=`, transformer towers) does not, because it makes
correctness non-local — in an `mtl` stack, whether an error discards accumulated
state is fixed by the *order of the transformer tower*, declared once and
invisible at the use site. That is the exact inverse of the locality property
this registry is trying to buy. `Result` + `?`, `async fn`, `#[must_use]`, a
tagged state enum, exhaustive match: all local at the call site, all cheap for an
agent to get right.

Adoption is not the argument for that split, and this issue does not claim it is
(see Log — the "async/await won" framing was withdrawn). Locality is.

## Done when

- `sdlc arch-principles` renders seven entries, `ARCH-ORDER` among them, and its
  header count reads 7.
- `ARCH-ORDER` passes the same per-entry contract check the other markers
  satisfy (`architectureEntry` + the `principle:` / `at-plan:` / `at-review:`
  bullet contract in `judge_test.go`).
- `TestArchitectureMarkers`' hand-written `want` list includes `ARCH-ORDER` in
  registry order, and `TestDeferredPrinciplesReachNoGate` stays green.
- The plan-quality, start-plan, and code-review prompts all carry the marker via
  `{{ARCH_STAR}}` (derived — no new hardcoded list), with goldens re-captured.
- `atlas/workflow/architecture-principles.md` documents the entry and the
  `ARCH-PURE` boundary (different purity-breaker, sibling not bullet).

## Plan

- [ ] Add the `ARCH-ORDER` section to `cmd/sdlc/internal/judge/architecture.md`
      after `ARCH-SECURE`; keep the `principle:`/`at-plan:`/`at-review:` bullet
      contract.
- [ ] Add `ARCH-ORDER` to the hand-written `want` list in
      `TestArchitectureMarkers` (the deliberate non-derived tripwire — see its
      comment; do not "fix" it by deriving).
- [ ] Re-capture goldens; confirm `{{ARCH_STAR}}` expansion carries the marker
      into all three prompts with no other hardcoded list needing an edit.
- [ ] Update `atlas/workflow/architecture-principles.md`: the entry, the count,
      and the ARCH-PURE/ARCH-ORDER boundary.
- [ ] Verify: `sdlc arch-principles` shows 7; `go test ./cmd/sdlc/...` green.

## Log

### 2026-09-04

Originated in a brain advisor session on why agentic coding degrades on
concurrent work. The chain, kept because the entry's *shape* is downstream of it:

- Opening theory (operator): concurrency's structure is absent from program
  structure, so next-token generation has nowhere to attach it. Sharpened to
  **linearizability of the text** — one reading order, many execution orders.
- Rejected cure: future monads / `flatMap` chains. They make a *dataflow DAG*
  manifest, which is not the failing case; the failing case is external events
  arriving mid-flight, which is an input you receive, not an effect you
  sequence. Also: monadic bind is inherently sequential, so concurrency lives in
  the **Applicative** structure, not the Monad (cf. `Concurrently`'s parallel
  Applicative + sequential Monad, `Haxl` needing `ApplicativeDo`). The reasoning
  path from "extract effects" to "use monads" is a natural wrong turn.
- Withdrawn claim: that `async/await` represents the monadic *shape* winning by
  syntax absorption. Adoption was used as evidence of correctness — the
  inference "worse is better" exists to refute. The operator's counter-case:
  https://xianxu.dev/2025/05/conversation_around_concurrent_programming_models/
  — `await` linearizes the text and thereby makes **dependency order** local,
  but erases **task extent**: which work could have run concurrently, who is
  still alive at return, who is cancelled on failure. The industry arc
  (callbacks -> promises -> await -> nurseries/TaskGroup/`coroutineScope`/
  `StructuredTaskScope`) is that extent being re-localized as lexical scope. So
  the locality criterion survives; it was applied too coarsely, to "sequencing"
  when the property that matters is "extent". The at-plan clause on task extent
  and cancellation propagation comes from this correction.
- Convergence worth keeping: `await` is optimized to make the modal case free,
  which pushes *all* residual entropy onto cancellation and error propagation —
  precisely the cells where the prior is diffuse. That is the mechanism behind
  "LLMs write good happy-path async and bad cancellation," and it is why the
  at-plan lens targets unblockable events specifically.

Also considered and rejected: folding this into `ARCH-PURE` as an extra bullet.
`ARCH-PURE`'s IO list is an enumeration of *members*; the gap here is a
difference in *kind*, so it earns a sibling.

## Revisions

### 2026-09-04 — fresh-eyes review of the Spec

Reason: spec review before implementation. Five wording deltas to the draft
entry; two raised findings withdrawn. The Plan's wiring steps are unchanged.

**Withdrawn — the entry is one principle, not two.** The review first read the
draft as bundling illegal-state-modeling with ordering under one marker, and the
`at-plan` `N/A` clause as mis-scoped to only the latter. It isn't: the principle
bullet's own first sentence names "the legal states **and** the transitions
between them," and those are two halves of one enumeration — a transition table
cannot be written over a 2^N implicit state space, so the tagged enum is the
*precondition* for the ordering work rather than a separate concern. The trigger
is consistent across all three bullets (`at-review` clause 1 already reads
"state whose transition set cannot be read off the code"), and the `N/A` clause
is correctly scoped as written. Residue: see delta 2.

**Withdrawn — no collision with the anti-state-machine stance.**
`atlas/workflow/sdlc-binary.md:1251` ("Do not formalize the workflow into a state
machine") was read as contradicting the principle's `(state, event) -> (state,
effects)` instruction at `start-plan`. It does not. That sentence guards against
prematurely modeling *the SDLC process* as an FSM — checkpoint gates over prose
stages, kept deliberately flexible — and ARCH-ORDER's trigger ("a component holds
state across events arriving from outside it") already scopes to components, not
to the process the org runs. The `.git/sdlc.lock` release/reacquire-with-HEAD-
recheck dance is a component and fires correctly; "formalize the SDLC" is not and
does not. No disambiguating clause needed.

**Deltas to apply:**

1. **Distinguish ARCH-ORDER from ARCH-SECURE in the principle bullet.** The
   Problem section's gap analysis rules out ARCH-PURE, ARCH-CONSTRAINTS and
   ARCH-MOCK but never mentions ARCH-SECURE, which already carries "prefer
   making invalid state unrepresentable — parse input into a typed value at the
   boundary" (`architecture.md:125`). Both entries will use the word
   "unrepresentable" and a reader needs to know which fires. Add one sentence:
   ARCH-SECURE is a **single-shot parse by provenance at a boundary**;
   ARCH-ORDER is the state **carried between events**. Temporal, not
   provenance-based.

2. **Scope the `2^N` sentence to event-bearing state.** "N boolean/nullable
   fields declare 2^N states and usually mean about five, and the bugs live in
   the difference" reads, lifted out of the bullet, as a general data-modeling
   claim with no event trigger — and it is the sentence a reviewer will quote
   when firing on a synchronous flag constellation the entry does not intend to
   cover. Tie it to state held across events.

   Accepted consequence: illegal-state modeling for state that is neither
   event-driven nor a boundary parse has no home in the registry. Left as a
   known gap (YAGNI) — coin an entry if it recurs.

3. **Record the in-fleet origin in `## Log`.** The Problem section argues from
   general agentic-coding failure, which is a claim about LLMs rather than about
   this fleet — the shape `atlas/workflow/architecture-principles.md` explicitly
   contrasts ARCH-SECURE's grounding against. The real origin is the **many
   rounds worked through against couch's various models when starting up pair**,
   and the distinct failure modes that surfaced there; that is what triggered the
   brain discussion this issue's Log already records. Land the pointer now so the
   pair-actor consult (post-draft, to add specifics) has an anchor rather than
   reconstructing it.

4. **Separate `at-review` clause 2 from ARCH-CONSTRAINTS.** "Concurrent work
   whose extent is not lexically bounded" sits next to ARCH-CONSTRAINTS'
   "unbounded concurrency or fan-out" and will be conflated. Half a sentence:
   ARCH-ORDER is **lifetime/extent** (who is still running at return, who is
   cancelled on failure), ARCH-CONSTRAINTS is **capacity** (how many at once).

5. **Correct the Done-when consumer enumeration.** It names "plan-quality,
   start-plan, and code-review." `judge_test.go:330` asserts four `BuildPrompt`
   categories embed the registry — `PlanQuality`, `MilestoneReview`, `DRY`,
   `PURE` — plus both `ArchitectureBlock` lenses (`start-plan`,
   `arch-principles`) and `CodeReviewBody`. Everything derives from
   `ArchitectureMarkers()`, so this is not a correctness risk; but an incomplete
   consumer list is the wrong thing to ship in the registry that defines the
   ARCH-PURPOSE shadow-sweep.

**Confirmed unchanged:** the Plan's five steps. The wiring surface really is
registry + the one hand-written marker list (`judge_test.go:363`) + goldens: the
block header count is derived (`architecture.go:53`), `{{ARCH_STAR}}` is derived
(`review.go:44`), and `TestDeferredPrinciplesReachNoGate` stays green because
ARCH-AUTHORITY keeps the deferred file's verdict at `guard`.

### 2026-09-04 — second fresh-eyes review, from a pair-actor session

Reason: the operator asked for a read from the session that had just spent a day
inside `pair#182`/`#185`, which is the in-fleet evidence delta 3 says this issue
is missing. Verdict, deltas, and that evidence, in that order.

**Verdict: the gap is real and the entry earns its place.** Verified the premises
rather than reading them back: the registry does have six entries, and
`architecture.md:125` does carry "prefer making invalid state unrepresentable"
under ARCH-SECURE — so delta 1's collision is real, not hypothetical. The
rejection of folding into `ARCH-PURE` holds for the stated reason: IO breaks
purity by *doing* something, concurrency by making the order of what already
happened unreadable from the text. Different tell, different lens. `ARCH-ORDER`
over `ARCH-FSM` is right — naming the question is what stops it becoming state
machine cargo cult.

**The strongest passage is Problem points 3–4, and it is stronger than the issue
claims for it.** "Conventional UX removes the *common* interleaving, making the
residual bug rare rather than absent, and thereby destroys the smoke-test
signal", together with "a human smoke test is a sampler over interleavings with
sample size 1 and no coverage report", is the actual mechanism. The failure is
not that agents write bad concurrent code — it is that the feedback loop confirms
whichever policy got sampled.

#### Three deltas

1. **`at-plan` risks being the ceremony it warns against.** It says do not
   re-litigate what conventional UX settles, then lists six unblockable events
   and asks for a cancel/queue/preempt/ignore + rollback decision "for each". Six
   cells per component is a checklist, and the `N/A` clause does not save it
   because the sweep happens before the N/A judgment. The entry's own best
   insight argues for targeting rather than sweeping: `await` made the modal case
   free and pushed all residual entropy onto cancellation and error propagation.
   Suggest asking which of the six actually applies here and why the rest do not
   — one sentence naming the event most likely to be mishandled beats six cells
   filled in.

2. **The `N/A` clause has an inverted failure mode.** Matching
   `ARCH-CONSTRAINTS`/`ARCH-SECURE` is right for consistency, but this
   principle's own premise is that the failure is INVISIBLE to the author —
   "there is no position in the token stream where 'and the user can cancel here'
   is representable". The author who cannot see the ordering is exactly the
   author who marks it `N/A`. Make `N/A` a claim that can be wrong: require
   naming what makes the component stateless ("holds no state between events
   because X"), not a bare marker.

3. **`at-review` clause 4 is buried and should not be.** "Flag tests that can
   only observe one interleaving" is the only clause attacking the ORACLE rather
   than the code, and it is the highest-leverage one in the entry. It currently
   sits last in a long bullet behind three clauses about code shape.

#### In-fleet evidence for delta 3

Delta 3 asks for the pair-actor consult to anchor the Problem section in this
fleet rather than in a claim about LLMs. From `pair#182` (relaunch) and
`pair#185` (status-row notices), 2026-09-04:

- **Designed right BECAUSE it was treated as ordering.** couch's relaunch is
  park-then-resume, and the whole design is the order of the checks: every
  refusal that can be seen is raised before the destructive step. Its four
  outcomes — `Relaunched` / `RefusedBeforePark` / `ParkIncomplete` /
  `ParkedNotResumed` — each name their own recovery. That is
  `(state, event) -> (state, effects)` with a rollback policy per cell, written
  before the code, and it is the part of the feature that had no defects.
- **Four defects that were ordering defects, all found late.** (a) A background
  inventory refresh landing mid-operation judged that operation's own
  confirmation stale — an unblockable event arriving mid-transition — and erased
  the recovery message on the one outcome that needed it. (b) An exemption keyed
  on thread address rather than the owning operation: wrong granularity for "who
  owns this transition". (c) An expected-exit marker applied by SLOT rather than
  by handle, so after a replace-in-place it marked the replacement and would have
  swallowed that child's first real death. (d) `pair#168`: a trailing `launch`
  ledger row with no `binding` shadows a good binding and strands a live session
  — completion arriving out of order, plus process death mid-transition.
- **The broken oracle, confirmed on the reviewer rather than the reviewed.** I
  recorded "`go test -race` passes" as close evidence. It came from ONE run of a
  test that fails 3 in 10. That is precisely "sample size 1, no coverage report,
  biased toward passing", and a boundary reviewer caught it, not the test loop.
  Point 4 of the Problem section is not a theory here; it is a transcript.

Each of (a)–(d) would have been caught by the `at-plan` clause as drafted, and
(d) by the `at-review` clause on completions arriving out of order. That is the
argument for the entry that the Problem section is currently making from
generalities about LLMs.

