---
id: 000215
status: codecomplete
deps: []
github_issue:
created: 2026-09-04
updated: 2026-09-04
estimate_hours: 1.49
started: 2026-09-04T21:49:32-07:00
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

## Spec

Add a seventh registry entry to `cmd/sdlc/internal/judge/architecture.md`. Named
`ARCH-ORDER`, not `ARCH-FSM` — naming the *deliverable* invites the cargo cult of
state machines over map/filter; naming the *question* does not. The trigger is
narrow (durable state + events from outside), and the entry carries an explicit
`N/A` clause, matching the gating idiom already set by `ARCH-CONSTRAINTS` and
`ARCH-SECURE`.

Entry text, post-revision — the eight deltas from `## Revisions` are folded in
here, so THIS block is the current version and Revisions is the changelog. Final
prose is the implementer's; the clauses are the spec.

```markdown
## ARCH-ORDER — Make the ordering explicit

- **principle:** When a component holds state across events arriving from
  outside it, the legal states and the transitions between them are part of the
  design, not an emergent property of the code. Model it as an explicit
  `(state, event) -> (state, effects)` enumeration rather than a constellation
  of independent flags whose legal combinations are unwritten. Where ARCH-PURE
  separates logic from IO, this separates **legal** state from **representable**
  state: N boolean/nullable fields carried across events declare 2^N states and
  usually mean about five, and the bugs live in the difference. The lens is
  *temporal*, not provenance-based — ARCH-SECURE makes invalid state
  unrepresentable at a single-shot parse of input the component did not produce;
  this makes it unrepresentable in the state a component carries **between
  events**. Concurrency does not need a new effect abstraction — it needs the
  interleaving space written down where a reader, a type checker, and a test
  generator can all see it.
- **at-plan:** For a component with durable state and external events, enumerate
  states x events -> (state, effects), including the interrupting ones. Do not
  re-litigate what conventional UX already settles (block input, show a spinner)
  — that policy is advisory and UI-local. Spend the enumeration on the events the
  caller **cannot** block: process death mid-transition, connection loss,
  session/tab close, retry of an already-applied step, a second actor on the same
  state, completions arriving out of order. Target rather than sweep — name which
  of those apply here and why the rest do not, and for each that applies say
  which of cancel / queue / preempt / ignore governs it and whether state rolls
  back. One sentence naming the event most likely to be mishandled beats six
  cells filled in. Name the extent of any concurrent work — who is still running
  when this returns, who is cancelled when this fails — where nondeterminism
  enters (clock, scheduler, IO completion order, arrival order), and how a
  failing ordering is reproduced. `N/A` is a claim that can be wrong: write it as
  "holds no state between events because X", never as a bare marker — a bare one
  is exactly what the author who cannot see the ordering will write.
- **at-review:** Flag tests that can only observe one interleaving — no seam to
  inject ordering, no way to reproduce a reported failure. A green run of such a
  test is a sample of size one that reports no coverage, so it confirms whichever
  ordering the author happened to get; this is the highest-leverage flag in the
  entry, because it attacks the oracle rather than the code. Flag state whose
  transition set cannot be read off the code — boolean/nullable constellations
  with unwritten legal combinations, where `isLoading && err != nil && data !=
  nil` is reachable and undefined; name the tagged enum they should collapse
  into. Flag concurrent work whose **extent** is not lexically bounded: a spawned
  task outliving its scope, no cancellation path. Extent is lifetime — who is
  still running at return; ARCH-CONSTRAINTS' "unbounded concurrency" is capacity
  — how many run at once. Flag an error or cancellation path that unwinds the
  sequencing but drops the in-flight effect.
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
  header count reads 7 — derived from `len(ArchitectureMarkers())`
  (`architecture.go:57`), so no count is edited by hand.
- `ARCH-ORDER` satisfies the generic per-entry contract:
  `TestArchitectureRegistry_Content` (`judge_test.go:119`), which derives its
  marker set from `ArchitectureMarkers()` and requires a `## ARCH-ORDER` heading
  plus all three of `**principle:**` / `**at-plan:**` / `**at-review:**`. (NOT
  `architectureEntry` — that helper has exactly one caller,
  `constraintsContractViolations` at `judge_test.go:222`, and is
  ARCH-CONSTRAINTS-specific machinery.)
- The entry's prose introduces no `ARCH-<NAME>` token lacking its own `##`
  heading. `markersIn` (`architecture.go:29`) scans the whole registry, so a bare
  mention becomes a phantom marker and trips
  `TestArchitectureRegistry_Content`'s "found it in prose only" branch. The
  entry cites ARCH-PURE, ARCH-SECURE and ARCH-CONSTRAINTS (all headed, all safe);
  the Spec's `ARCH-FSM` naming rationale stays in the issue and out of the entry.
- `TestArchitectureMarkers`' hand-written `want` list (`judge_test.go:363`)
  includes `ARCH-ORDER` last in registry order, and
  `TestDeferredPrinciplesReachNoGate` stays green — ARCH-AUTHORITY keeps the
  deferred file's verdict at `guard`.
- Every registry consumer carries the marker with no new hardcoded list. The
  mechanism is three paths, not one:
  - four `BuildPrompt` categories via `{{ARCH_BLOCK}}` -> `ArchitectureBlock`
    (`prompts.go:60`) — `PlanQuality`, `MilestoneReview`, `DRY`, `PURE`, pinned
    at `judge_test.go:328`;
  - `start-plan` and `arch-principles` via a direct `judge.ArchitectureBlock`
    call (`startplan.go:75`);
  - `code-review.md:112` via `{{ARCH_STAR}}` -> `ArchitectureMarkers()`
    (`review.go:44`) — the one and only template using that token.
- The four registry-bearing goldens are re-captured — `dry.prompt`,
  `milestone-review.prompt`, `plan-quality.prompt`, `pure.prompt` — and their
  diff inspected to contain only the ARCH-ORDER block plus the header count
  `6` -> `7`.
- `atlas/workflow/architecture-principles.md` documents the entry and both
  boundaries — ARCH-PURE (different purity-breaker — sibling, not bullet) and
  ARCH-SECURE (provenance vs. temporal) — by MAPPING to the registry, not
  restating its clause text. NOT the entry count: that is derived
  (`len(ArchitectureMarkers())`, `architecture.go:57`) and the atlas has never
  carried it.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec           design=0.8  impl=0.1
item: smaller-go-module    design=0.05 impl=0.12
item: atlas-docs           design=0.05 impl=0.05
item: milestone-review     design=0.0  impl=0.18
design-buffer: 0.15
total: 1.49
```

Σdesign 0.9 × 1.15 + Σimpl 0.45 × 1.0 = 1.49.

`issue-spec` is priced **undiscounted** and against the window `sdlc actual`
actually computes, which opens at `1a01f82` (the first `#215: issue-sync`), not
at the claim. So it covers both `## Revisions` fresh-eyes rounds, the eight-delta
fold, and the Done-when factual sweep. The ×0.2 spec-quality discount does not
apply to the primitive whose deliverable *is* the spec — nothing pre-resolved
that dialogue; it was the dialogue (precedent: ariadne#205 carried
`issue-spec design=1.0` undiscounted). 0.8 rather than the 1.0+ of a
from-scratch authoring, because the originating brainstorm predates the window's
first commit even though that commit lands its output.

The ×0.2 discount IS applied to the other three rows and is sound there: the
Spec holds the entry text verbatim and Done-when cites file:line for every
consumer, so implementation is transcription rather than decision.
`smaller-go-module` is the registry entry, the one hand-written `want` list, and
the four golden re-captures — `impl` above the floor because
`golden_test.go:35` makes the diff read mandatory attention, not autonomous
typing. `atlas-docs` is `architecture-principles.md` (two boundary paragraphs,
undrafted). `milestone-review` is the one close-time boundary review
(single-pass, no `Mx`), priced near the primitive ceiling: ariadne#208 — the same
shape, one registry entry ago — priced two rows at ceiling and still closed 1.72
against 1.13, and this deliverable is 40 lines of dense prose a reviewer will
want to wordsmith. *Produced via
`brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

## Plan

- [x] Add the `ARCH-ORDER` section to `cmd/sdlc/internal/judge/architecture.md`
      after `ARCH-SECURE`, from the Spec's entry block verbatim — that block is
      post-revision (all eight deltas folded in), so no replay of `## Revisions`
      is needed. Keep the `principle:`/`at-plan:`/`at-review:` bullet contract.
- [x] Confirm the entry introduces no unheaded `ARCH-<NAME>` token (no
      `ARCH-FSM`), so `markersIn` still finds exactly seven markers.
- [x] Add `ARCH-ORDER` to the hand-written `want` list in
      `TestArchitectureMarkers` (`judge_test.go:363`) — the deliberate
      non-derived tripwire; do not "fix" it by deriving.
- [x] Re-capture the four registry-bearing goldens (`-update-golden`) and READ
      the diff: only the ARCH-ORDER block and the header count `6` -> `7` may
      appear. `golden_test.go:35` forbids re-capture used to paper over drift, so
      an unrelated prompt edit riding along is a stop, not a pass.
- [x] Update `atlas/workflow/architecture-principles.md`: a MAP-level paragraph
      (boundaries, shaping choices, in-fleet origin) — never a restatement of the
      entry's clauses, and never the entry count, which is derived.
- [x] Verify: `sdlc arch-principles` shows 7; `go test ./cmd/sdlc/...` green.

## Log

### 2026-09-04 — implemented
- 2026-09-04: closed — sdlc arch-principles renders 7 entries, header count derived and reads 7. All 7 markers headed; registry token set == headed set (no phantom; ARCH-FSM stayed in the issue). Golden re-capture touched exactly the 4 predicted files (dry/milestone-review/plan-quality/pure); diff reconciles line-for-line: 189 insertions = 184 entry lines (46x4) + 4 header counts 6->7 + 1 ARCH_STAR list gaining ARCH-ORDER, 5 deletions their counterparts — nothing unrelated rode along per golden_test.go:35. go test ./cmd/sdlc/internal/judge/ green incl. TestArchitectureMarkers, TestArchitectureRegistry_Content, TestDeferredPrinciplesReachNoGate. Atlas updated and, after BR-1, MAPS to the registry rather than restating it: 12-word-window scan of the entry against the atlas finds 0 shared spans (was 6). BR-3 taken (Done-when corrected: the count is derived, the atlas never carried it). BR-2 recorded as a registry-level length question, not trimmed — the review scopes it as a note for the next entry. Pre-existing unrelated failure TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory (hardcoded path to an archived plan) confirmed identical with changes stashed; tracked as ariadne#210.; review verdict: FIX-THEN-SHIP

`ARCH-ORDER` landed as the seventh registry entry
(`cmd/sdlc/internal/judge/architecture.md:146`), taken verbatim from the
post-revision Spec block. Verification:

- `sdlc arch-principles` renders 7 entries and its header count reads 7 (derived
  from `len(ArchitectureMarkers())`, no hand-edited count anywhere).
- All seven markers carry `##` headings and the token set in the registry is
  exactly the headed set — no phantom marker, `ARCH-FSM` stayed in the issue.
- Golden re-capture touched exactly the four predicted files
  (`dry`/`milestone-review`/`plan-quality`/`pure`) and the diff reconciles to the
  line: 189 insertions = 184 entry lines (46 x 4) + 4 header counts `6` -> `7` +
  1 `{{ARCH_STAR}}` list gaining `ARCH-ORDER`; 5 deletions are their
  counterparts. Nothing unrelated rode along, per `golden_test.go:35`.
- `go test ./cmd/sdlc/internal/judge/` green, including
  `TestArchitectureMarkers`, `TestArchitectureRegistry_Content`, and
  `TestDeferredPrinciplesReachNoGate`.

**Pre-existing failure, not from this window:**
`TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` fails on a
hardcoded path to `workshop/plans/000200-...-plan.md`, which was archived.
Confirmed identical failure with this issue's changes stashed. Already tracked
as ariadne#210 (`fleet-plan-test-hardcoded-path`); untouched here.

**Gate history.** Plan-quality blocked at round 1 on two Importants and cleared
at round 2 (PQ-1..PQ-4 all disposed `addressed`). Both Importants were real:
PQ-1 caught that the eight `## Revisions` deltas had no landing place — the Spec
still held the pre-delta draft, so an implementer could tick every box and ship
the un-revised entry (ARCH-DRY: one current version, changelog separate). PQ-2
caught a factual error that had survived the original spec *and* two fresh-eyes
review rounds: `{{ARCH_STAR}}` is not the mechanism by which plan-quality and
start-plan receive the registry — they use `{{ARCH_BLOCK}}` -> `ArchitectureBlock`
(`prompts.go:60`) and a direct call (`startplan.go:75`); `{{ARCH_STAR}}` appears
in exactly one template (`code-review.md:112`). The same sweep corrected
`architectureEntry` (`judge_test.go:195`, one caller, ARCH-CONSTRAINTS-specific)
to `TestArchitectureRegistry_Content` (`judge_test.go:119`) as the generic
per-entry contract. That is ARCH-PURPOSE's instance-vs-class case landing inside
the issue that adds a principle about legibility — the delta had *named* the
incomplete list a round earlier and nobody swept the class.

The estimate was corrected before implementation after the estimate-quality
judge (INFO, non-blocking) identified a window-boundary error; see `## Revisions`.

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

### 2026-09-04 — estimate corrected before implementation

Reason: the estimate-quality judge returned INFO (non-blocking) on the first
`## Estimate` block, but its point 1 was a boundary error, not a sizing quibble,
and a knowingly-wrong estimate pollutes velocity calibration.

The first block priced `issue-spec` as "the in-window spec work, **not** the
original authoring" and applied the ×0.2 spec-quality discount to it, for a total
of 0.64h. Both were wrong. `sdlc actual --issue 215` opens the window at
`1a01f82` (2026-09-04 17:09), not at the 21:49 claim, and already measured 0.61h
with zero implementation done — so the row's stated scope was false for the
window the tool computes. And the ×0.2 discount is circular on the one primitive
whose deliverable is the spec itself: nothing pre-resolved that design, it *was*
the design.

Delta: `issue-spec` 0.15 -> 0.8 design (undiscounted, whole window);
`milestone-review` impl 0.1 -> 0.18 (near ceiling, per the ariadne#208
precedent that priced two rows at ceiling and still closed 1.72 against 1.13);
`smaller-go-module` impl 0.1 -> 0.12 (the mandated golden-diff read is
attention, not typing). Total 0.64 -> 1.49. The ×0.2 discount stands on the
other three rows. Corrected before any code changed, so the revision is a fix to
the derivation rather than a fit to a known outcome.

### 2026-09-04 — close-review corrections

Reason: close-gate boundary review, round 1 (verdict FIX-THEN-SHIP; BR-1
Important, BR-2/BR-3 Minor).

**BR-1 (Important, `atlas-restates-single-source`) — taken, swept as a class.**
The shipped atlas paragraph copied ARCH-ORDER clause wording near-verbatim
instead of mapping to it. Verified rather than assumed: the finding's line cite
had drifted, but a normalized comparison found six distinct clause spans shared
between `atlas/workflow/architecture-principles.md` and the registry entry —
including the whole oracle sentence, the `N/A`-as-falsifiable-claim clause, and
the principle opener. Nothing pinned the copy, so a later rewording of the
registry would have stranded it silently. That is ARCH-DRY, and it contradicts
the same atlas file's own route-don't-restate convention (`:24-27`), which every
sibling paragraph follows.

Swept the class, not the cited site: the entry paragraph is now map-level
throughout, keeping only what is genuinely atlas-only content — the two boundary
bullets (ARCH-PURE sibling-not-bullet, ARCH-SECURE provenance-vs-temporal), the
two shaping choices a future editor could undo by accident (at-plan targets
rather than sweeps; at-review leads with the oracle clause), and the
`pair#182`/`#185` origin. Verified by a 12-word-window scan of the registry entry
against the atlas: **0 shared spans**, down from six.

**BR-3 (Minor, `derived-fact-restated`) — taken.** The Done-when bullet asked
the atlas to document "the count." That was wrong in the criterion, not in the
implementation: the count derives from `len(ArchitectureMarkers())`
(`architecture.go:57`) and no "six"/"seven" has ever appeared in the atlas file —
writing one in would create exactly the hand-maintained restatement its own
"Adding an entry" section rules out. Bullet corrected in place (and it now also
states the map-don't-restate requirement BR-1 exposed as unstated).

**BR-2 (Minor, `registry-entry-length-unbudgeted`) — recorded, not acted on.**
ARCH-ORDER is 531 words against ARCH-PURPOSE's 371 and ARCH-DRY's 92, and grew
`architecture.md` from 9,510 to 12,868 bytes — roughly +3.3 KB on each of the six
registry-bearing prompt paths. The review itself scopes this as "a note for the
next entry, not a defect in this one," and shortening the entry would undo the
targeting and oracle-first clauses that two review rounds deliberately added. The
real observation is that the registry has no length norm while its cost is
multiplied across six prompts; that is a registry-level question worth its own
issue rather than a trim here.

### 2026-09-04 — close-review round 2 (advisory, both taken)

Round 2 disposed all three round-1 findings as addressed and raised two Minors,
both of the "you fixed the instance, now write the rule" shape. Both taken, and
both rules landed in `atlas/workflow/architecture-principles.md` "Adding an
entry" — the checklist the next entry-adder actually reads — rather than only
here, since this file archives to `workshop/history/`, which AGENTS.md marks
"don't read."

**BR-4 (`derived-fact-restated`, 2nd in family).** BR-3 fixed the Done-when
bullet; the enumerable sibling at Plan step 5 still demanded "the count" from the
atlas and was ticked `[x]`, asserting delivery of the very restatement Done-when
now forbids. Site fixed, and the rule written down: **a fact derived at runtime
from the registry — entry count, marker list, `{{ARCH_STAR}}` — may appear as a
verification criterion about a derived output, never as a hand-maintained
requirement in an artifact.** Prevalence re-measured after the fix: 0 demand
sites; the remaining "7"/"seven" hits are verification assertions the rule
permits.

**BR-5 (`atlas-restates-single-source`, 2nd in family).** BR-1's instances were
genuinely gone (0 shared 12-word spans), so the sweep was clean — what was
missing was the rule. The atlas "Adding an entry" checklist listed three touch
points (`architecture.md`, the hand-written list, the goldens) and omitted the
map-level paragraph in that same file, which every entry since ARCH-MOCK has in
fact required. A reader of that checklist would have repeated BR-1 exactly. Added
as the fourth touch point, with its shape stated.

**Calibration note for #127.** Closed at est 1.49 / actual 0.61 — 2.4× over. The
first derivation (0.64) was nearly exact on the total while being wrong on its
reasoning: it priced `issue-spec` against the claim rather than against the
window `sdlc actual` computes. Correcting that boundary error moved
`issue-spec design` 0.15 -> 0.8 and overshot, because the correction reasoned
from ariadne#208's *under*-estimate (closed 1.72 against 1.13) and over-applied
the lesson. The signal worth keeping: for a registry-content issue whose spec
work is review rounds rather than authoring, `issue-spec design` near the table
floor (~0.3–0.5) fits better than either 0.15 or 0.8, and one `milestone-review`
row at 0.1 was right — the second close-review round cost minutes, not the
near-ceiling 0.18 the correction assumed.
