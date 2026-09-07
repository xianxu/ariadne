---
id: 000217
status: open
deps: []
github_issue:
created: 2026-09-07
updated: 2026-09-07
estimate_hours:
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

## Spec

Add `ARCH-ERROR` to `cmd/sdlc/internal/judge/architecture.md`. Its unique
question, which nothing else in the registry asks: **which layer owns this
failure, and which of retry / degrade / abort does it choose?**

### Trigger — narrow, and `N/A` must stay cheap

Applies to a component that **holds state across calls** — an "actor" in the
loose sense, a thing with a lifecycle that can fail and be restarted. A pure
function that cannot fail, and a leaf that genuinely just propagates, are `N/A`.
Follow `ARCH-ORDER`'s idiom: `N/A` is a claim that can be wrong, so it is
written as a claim rather than left blank — but it must remain a one-line
answer, or this becomes ceremony on every function (the `ARCH-PERFORMANCE`
lesson, `#216`).

### The clauses

**1. Name the fault domain.** The unit that can fail and be recovered
independently. In `pair`: a couch thread; a `pair term` tab; an `sdlc`
invocation. The domain usually already exists but is **anonymous** — implied by
a goroutine, a request scope, a `defer`, a subprocess — and because nobody named
it, nobody could attach policy to it, so policy scattered to the detection
sites. Naming it is the whole move.

**2. Is restart a valid recovery?** Only if the restarted thing can reconstruct
its state. This is where the famous answer is wrong and the entry earns its
keep — see the worked example below.

**3. Three dispositions, and only three:** retry (transient), degrade (lose a
feature, keep the system), abort the domain (restart it). A handler that does
none of these is hiding an error, which is a checkable test rather than a taste
claim.

**4. Retry belongs to exactly one layer — the one that owns the deadline.**
N levels retrying r times is `r^N` amplification arriving precisely when the
system can least absorb it, and multi-level backoff caps the *rate* while
multiplying *latency*, so the top caller burns its budget on retries beneath it
and times out having learned nothing.

The symmetry to state: **deadlines flow down as cancellation; errors flow up as
typed values; neither direction invents a budget it does not own.** A lower
layer imposing its own timeout is inventing a budget, and blocking indefinitely
is correct when nobody has asked it to hurry. What it must not do is retry or
swallow — it is *promptly honest*, not fast. "Transient vs permanent" is a type
at the boundary, not a string match at the call site.

**5. Concentrate the policy in a named thing.** Erlang's contribution is not
crashing; it is that the supervisor owns restart policy and the worker owns
none, so worker code stays about the happy path and the policy sits in one
declarative, testable place. Without an actor runtime the mechanisms are: a type
that owns a domain's lifecycle (couch's `Couch` and `Console` already are
supervisors — they were never called that, which is why their failure policy is
scattered across attach/switch/paint), a structured-concurrency scope, a typed
error taxonomy at the domain boundary, and reconstruct-from-durable-state as the
restart analogue.

**6. Decide with evidence, and mind the gain.** A stateful component should
expose its own health as readable local state (EWMA latency, error rate,
in-flight count) rather than burying it in a private breaker field: that is the
evidence the fault-domain boundary consults, it is assertable in tests, and it
is available without a telemetry round trip.

The stability clause, which agents get wrong by default because the obvious code
is a per-call `if slow { skip }`: **a decision that redirects load needs
aggregate evidence and hysteresis; a decision that only spends your own budget
does not.** Per-request adaptation is a high-gain, undamped controller acting on
n=1 — correlated skips synchronise into oscillation, and per-request evidence
cannot distinguish "overloaded, shedding helps" from "cold cache, shedding
prevents warming". Per-request *budgeting* ("200ms left, the optional call costs
300ms, skip it") is not a feedback loop at all and is safe. So is attribution.

### Worked example: couch inverts "let it crash"

couch has **two nested fault domains with opposite restart validity**:

- the **pane/attachment** — cheap to restart, because its state is
  reconstructible from `ThreadStore` (that reconstruction is `pair#206`);
- the **agent process** — *not* restartable, because hours of context are
  unreconstructible, and preserving exactly that state is why couch exists.

So couch's policy is one line: **absorb faults at the pane boundary; never let
one propagate to the agent.** "Let it crash" is precisely wrong here, and the
entry should carry this because the case where the famous answer inverts is what
makes the principle load-bearing rather than quotable.

### Boundaries with neighbouring entries

- `ARCH-ORDER` is *within* a domain: what happens to state across events,
  including a retry arriving at an already-applied step. `ARCH-ERROR` is *where
  the domain boundaries are* and who decides across them. Clean seam: ORDER says
  a retry can arrive at an applied step; ERROR says who is allowed to issue one.
- `ARCH-SECURE` covers the failure path of a single-shot parse of untrusted
  input. `ARCH-ERROR` covers the disposition of a failure that has already
  crossed a component boundary.
- `ARCH-CONSTRAINTS` owns overload behaviour as an *envelope*; `ARCH-ERROR` owns
  who acts when the envelope is exceeded.

### Deliberately deferred

**In-process spend history** — letting a call site read the timing of prior
calls in its own request, so a supervisor decides with evidence rather than
blind. Today that data flows outward to a human on a weeks-long loop
(telemetry → dashboard → operator → code change); inward it would close in
microseconds. The uses are real (actionable error text, deadline-aware
degradation targeted at the *sick* dependency rather than at optional work
generically, per-request circuit breaking, and assertable call-graph
invariants), and the reason frameworks lack it looks like a historical accident:
the trace was designed as a *shipping* format assembled out-of-process, so the
in-process representation is write-only.

Not in this entry. It is a mechanism proposal needing a design and probably a
prototype before it earns registry space; a principle that asks for something
nobody has built is aspiration (`#216`).

## Done when

- `ARCH-ERROR` is in the registry, satisfying the per-entry contract check
  (`architectureEntry` plus the `principle:`/`at-plan:`/`at-review:` bullets).
- `TestArchitectureMarkers`' hand-written `want` list gains `ARCH-ERROR` in
  registry order — this entry **does** add a marker, unlike `#216`, so that list
  must change and its deliberate non-derived tripwire comment stays intact.
- `sdlc arch-principles` renders **8** entries with the header count matching.
- Goldens re-captured; `{{ARCH_STAR}}` carries the marker into the planning,
  plan-quality and code-review prompts with no other hardcoded list edited.
- `TestDeferredPrinciplesReachNoGate` stays green.
- `atlas/workflow/architecture-principles.md` documents the entry, its trigger,
  and the boundaries against ORDER / SECURE / CONSTRAINTS, so the seams are not
  re-litigated.

## Plan

- [ ] Draft the entry from the clauses above; keep the trigger narrow and the
      `N/A` answer one line.
- [ ] Add `ARCH-ERROR` to the hand-written `want` list in
      `TestArchitectureMarkers` (do not derive it — see its comment).
- [ ] Re-capture goldens; verify 8 entries and marker propagation.
- [ ] Update `atlas/workflow/architecture-principles.md` with the entry and the
      three boundaries.
- [ ] Verify: `sdlc arch-principles` shows 8; `go test ./cmd/sdlc/...` green.

## Log

### 2026-09-07

From an operator brainstorm on error-handling architecture, opened by Erlang's
"let it crash" and the observation that the *placement* of error handling — not
its mechanics — is the architectural decision.

Three things the discussion changed, recorded because the entry's shape is
downstream of them:

1. **"Fail fast" was wrong and is withdrawn.** The operator's objection: if
   deadlines exist only where someone cares, a lower layer imposing its own
   timeout invents a budget it does not own. Corrected to *promptly honest, not
   fast* — block indefinitely, do not retry, do not swallow — which yields the
   cleaner symmetry now in clause 4.
2. **The actor abstraction's contribution was named.** Not concurrency: making
   the fault domain a *named, first-class object* so policy can be attached to
   it. That reframed clause 5 from "use supervisors" to "the domain already
   exists and is anonymous; name it".
3. **The stability clause came from the operator's challenge** that per-request
   adaptation may destabilise a system where traditional practice adapts on
   global statistics. Correct for load-redirecting decisions; the decomposition
   into adaptation / budgeting / attribution is what preserves the useful part.

The couch example is not illustrative decoration — it is the reason to prefer
this framing over quoting Erlang, because it is the case where the famous answer
is the wrong one and the codebase paid four rounds to discover it.
