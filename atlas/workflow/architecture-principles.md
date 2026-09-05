# Architecture Principles

The canonical `ARCH-*` registry lives in
`cmd/sdlc/internal/judge/architecture.md`. It is embedded into plan-quality and
boundary-review prompts, `sdlc start-plan` pushes it into planning sessions, and
`go run ./cmd/sdlc arch-principles` renders it for non-gate design work.

Key consumers:

- `cmd/sdlc/internal/judge/architecture.go` extracts markers and renders the
  registry block.
- `cmd/sdlc/startplan.go` and `cmd/sdlc/archprinciples.go` are the planning-time
  push and standalone pull paths; both reuse `ArchitectureBlock`.
- `cmd/sdlc/internal/judge/judge_test.go` pins marker extraction and prompt
  embedding; `architecturedeferred_test.go` pins that deferred principles reach
  no gate.
- `cmd/sdlc/internal/judge/testdata/golden/*.prompt` pins the generated prompt
  bodies that carry the registry. A deliberate registry edit re-captures these
  (`-update-golden`); the ⛔ in `golden_test.go` forbids re-capturing to paper
  over *drift*, which is a different case.
- `cmd/sdlc/gatefindings.go` routes every fixer-facing findings refusal to
  `ARCH-PURPOSE` (#203) — the judges get the registry inlined because a marker
  alone would dangle in a fresh context, but these lines are read by the main
  thread, which already holds the block from `sdlc start-plan`. Guarded by
  `TestFixTheClassLine_RoutesToArchPrinciples`, which asserts the ROUTING and
  never the wording: asserting wording is what would let the line become a
  second copy of the principle.

`ARCH-PURPOSE` covers two axes of the same rule: deliver the issue's purpose
rather than its easy subset, and answer a review finding with the *class* it
names rather than the one site — a finding's named site being the easy subset
again. That is why #203 extended it instead of coining a fifth marker.

`ARCH-MOCK` codifies the external dependency rule: every relied-on external
binary/service should sit behind a seam with a stateful fake for integration and
end-to-end tests, plus live conformance checks where practical to keep the fake
honest against the real dependency. For owned libraries, services, and binaries,
the storage/backend layer should also boot from portable file folders and/or
database configuration without depending on production configuration or
production databases.

`ARCH-CONSTRAINTS` makes runtime behavior an explicit design input. Planning
classifies the workload/interaction path and records each material parameter's
budget or range, basis, and bounded behavior when exceeded; review checks the
implementation and representative measurements against that operating envelope.
Its domain prompts cover latency, workload/input scale and growth, CPU/memory/IO, concurrency,
environment/co-tenancy, and overload without imposing universal numeric defaults.

`ARCH-SECURE` (#208) is the trust lens: what a component consumes that it did
not produce, and what it holds that must not leak. Its clauses are drawn from
defects this fleet actually shipped rather than from a generic checklist —
a test run that overwrote the operator's live proxy config with a test key, a
picker that minted a `0600` key as a side effect of reading host/port, a
persisted cache that crashed on a hand-edited row. The registry had zero
security-shaped words before it, so those landed under test-hygiene, ARCH-PURPOSE
and crash-bug markers respectively. Trust is treated as a property of
*provenance*: an artifact this same program wrote is untrusted once it has
crossed a process, session, or version boundary.

`ARCH-ORDER` (#215) is the temporal lens: when a component holds state across
events arriving from outside it, the legal states and the transitions between
them are design, not an emergent property of the code. It sits beside two
neighbours it is easy to confuse with:

- **ARCH-PURE** is a *sibling, not a bullet under it.* `ARCH-PURE` names `clock`
  in its IO list, but its lens is business logic vs. IO — "don't bury logic in
  handlers." IO breaks purity by *doing* something; ordering breaks it by making
  the order of what already happened unreadable from the text. Different
  purity-breaker, different tell, so folding this in as an extra ARCH-PURE bullet
  was rejected: that entry's IO list enumerates *members*, and this is a
  difference in *kind*.
- **ARCH-SECURE** also says "make invalid state unrepresentable," and the split
  is **provenance vs. temporal**: ARCH-SECURE does it at a single-shot parse of
  input the component did not produce; ARCH-ORDER does it in the state a
  component carries *between* events. Illegal-state modeling that is neither —
  a flag constellation in synchronous state with no external events — has no
  registry home on purpose (YAGNI); coin an entry if it recurs.

Its `at-plan` lens deliberately *targets rather than sweeps*: it tells the
planner not to re-litigate what conventional UX already settles (block input,
show a spinner — advisory and UI-local) and to spend the enumeration on the
events the caller cannot block, naming which apply and why the rest do not. `N/A`
must be written as a falsifiable claim ("holds no state between events because
X"), because this principle's own premise is that the failure is invisible to the
author — a bare `N/A` is exactly what the author who cannot see the ordering will
write. Its `at-review` lens leads with the **oracle** clause (tests that can
observe only one interleaving) rather than a code-shape clause: a green run of
such a test is a sample of size one that reports no coverage, so it confirms
whichever ordering the author happened to get.

Like `ARCH-SECURE`, its clauses are grounded in defects this fleet shipped rather
than a generic checklist — `pair#182`/`#185`, where couch's park-then-resume was
designed correctly *because* it was treated as an ordering problem (four named
outcomes, each with its own recovery) while four separate defects were ordering
defects found late, and `go test -race` was recorded as close evidence from a
single run of a test that fails 3 in 10.

## Deferred principles (documented, not gated)

`cmd/sdlc/internal/judge/architecture-deferred.md` holds principles written down
but deliberately kept out of the gate prompts — currently `ARCH-AUTHORITY`
(confused-deputy risk and blast radius), deferred because parley.nvim#129 already
owns that decision and a registry entry would restate it.

**Activation is a MOVE and nothing else:** cut the section into
`architecture.md`. The count in the block header, `{{ARCH_STAR}}`, every prompt,
and `sdlc arch-principles` all follow, because they derive from
`ArchitectureMarkers()`. Re-capture the goldens and update the one hand-written
marker list (below).

`TestDeferredPrinciplesReachNoGate` makes "documented but not gated" a checked
property. Both of its halves are derived rather than listed:

- the forbidden marker set, by scanning `architecture-deferred.md` with
  `markersIn` — the *same* extraction `ArchitectureMarkers` uses, extracted for
  this in #208 — so activation *empties* the set and the guard stays green with
  no edit, which is what makes the "just move it" claim true. Two copies of that
  scan is the one duplication a disjointness guard cannot afford;
- the gate-facing text, by walking `promptFS` and rendering each prompt through
  `BuildPrompt`, plus both `ArchitectureBlock` lenses, `CodeReviewBody` and the
  raw registry — so a prompt added later is covered without anyone remembering.

An empty forbidden set would make the guard vacuous, so `deferredVerdict`
classifies why it is empty. That type is the single source of the mapping; it is
documented there and nowhere else, so this page describes it rather than
restating the rule:

| verdict | shape of the file | action |
| --- | --- | --- |
| `guard` | entries present | enforce disjointness |
| `broken` | sections, but no marker parses | FAIL — the guard is disarmed |
| `retired` | no sections, no markers | SKIP — every entry was activated |

`retired` **skips and never fails**. Failing it would mean activating the last
deferred entry reds the suite, which breaks the "activation is a pure move"
contract this design exists to keep. The classification is pure
(`parseDeferred`) and table-tested over synthetic content, because the committed
tree can only ever exhibit one of the three.

## Adding an entry

The registry is a real single source at runtime, but **one** test writes the
marker set by hand: `TestArchitectureMarkers` (`judge_test.go`). That is the
tripwire that makes an accidental registry edit visible, so it is deliberate —
do not "fix" it by deriving it, or the suite would assert nothing about the
registry's contents. Every other site derives from `ArchitectureMarkers()`.

So a new entry touches: `architecture.md`, that one list, and the goldens
(`-update-golden`). Before #208 it silently touched five places, three of which
would have kept passing while covering nothing.
