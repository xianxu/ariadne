# Gate state — how an SDLC gate remembers what it asked for

A **gate ledger** is the durable memory of one SDLC gate on one issue: the findings a
fresh-context judge raised, the stable ids the binary assigned them, how later rounds
disposed of them, and whether each round blocked. Introduced by ariadne#187 for the
`change-code` plan-quality gate; designed gate-agnostic so #183's close-boundary gate
consumes the same notion rather than inventing a second one.

## Why it exists

`sdlc change-code` used to dispatch a brand-new plan reviewer on every invocation, print
its output, and forget. A reviewer with no memory of its own prior findings cannot
converge: it re-derives an absolute bar each round and surfaces the next-deepest layer of
a plan that keeps improving. On pair#127 that showed up as a clean descent — Critical →
Important → Info across six invocations, five rejections, for a 126-line change.

The fix is not a lower bar. It is memory: a reviewer that can see its three findings were
addressed says "ship" on round 2.

## Where it lives

`workshop/plans/NNNNNN-slug-plan-gate.md` (plan-quality) and
`workshop/plans/NNNNNN-slug-close-gate.md` (the boundary review, #194) — beside the durable plan and the boundary-review
sidecars, archived with them by `sdlc push`/`sdlc merge` (the `<id>-*` glob).

Deliberately **not** `*-plan-review.md`: `construct/vocabulary/verdict.cue` claims the
`*-review.md` glob for the verdict noun, and that glob asserts "this document carries a
boundary verdict". A gate ledger carries findings and no verdict.

The file is two projections of one `Ledger`:

- **YAML frontmatter** — the machine view, and the only thing the parser reads.
- **Generated prose** — the human view, derived from the same struct at render time.

Neither is hand-maintained, so the document cannot disagree with itself.

## The vocabulary

`construct/vocabulary/finding.cue` (Go binding: `pkg/vocab/finding.go`) is the single
source for:

| concept | values | meaning |
|---|---|---|
| `categories.blocking` | `Critical`, `Important` | undisposed ⇒ the gate refuses |
| `categories.advisory` | `Minor` | recorded; never blocks |
| `hardBlocking` | `Critical` | still blocks past the round cap |
| `dispositions.closing` | `addressed`, `withdrawn` | settled; stops blocking |
| `dispositions.open` | `not-addressed` | still open; keeps blocking |

Severity names are `code-review.md`'s existing three — the boundary review and the plan
gate share one taxonomy, pinned by a drift test.

Dispositions are **partitioned**, not listed flat, and that is load-bearing: a flat list
plus a prose gloss would leave the closes-vs-leaves-open decision in a Go switch, so adding
`deferred` later would pass validation, match no case, and wedge a finding open forever.

## The handoff

The judge emits a fenced ` ```findings ` block — the schema'd stochastic→deterministic
boundary the [`agent-binary-handoff-schema`](../../workshop/targets/agent-binary-handoff-schema.md)
target requires. The binary validates it against the model and **fails closed**: a missing
or model-invalid block is a protocol error, never a prose guess.

The judge writes `id: new`; **the binary assigns the stable id**, so the judge only ever
has to REFER to identifiers it was handed. Ids never reuse — one that changed meaning
between rounds would let a later round dispose the wrong finding.

## The decision

`gatestate.Decide` is a pure function over the accumulated ledger, not a read of the
judge's verdict token. It blocks iff some finding is still **open** at a blocking severity.

That is the mechanic that converges: a fresh `Minor` cannot cost a round-trip, and disposed
blockers open the gate. Past `WF_PLAN_ROUND_CAP` rounds (default 3) only `Critical` blocks.

**The demotion is safe only because the boundary review picks the rest up.** Since #194
that pickup is a SEED, not an instruction: on the boundary ledger's first round,
`seedFromPlanGate` copies the plan gate's still-open findings into it under `BR-*` ids at
`BoundaryAll`, so they arrive through the same prior-findings block as everything else and
can be disposed by name. (Before #194, `code-review.md` told the reviewer to read
`-plan-gate.md` off disk itself — which, once the boundary gate had its own ledger, would
have put two id namespaces in one output fence.) A guard test pins the prompt half; another
pins the seeding. Without both, the demotion becomes a silent loss.

**At the boundary gate a demotion means something different**, and is announced. There is no
later gate to pick up what the cap demotes — the boundary review is the last read before
publish — so each demoted finding gets a warning naming it. Its cap knob is
`WF_BOUNDARY_ROUND_CAP`.

## The pass-through

`Ledger.ContentHash` records `sha256(issue+plan)` as of the last **passing** round. When
content is unchanged since then, `change-code` skips dispatch entirely and persists no
round.

This is what makes #187's gate reorder (plan-quality before the estimate gates) a net win
rather than a net cost. Without it, every estimate-gate failure would pay a fresh
multi-minute judge dispatch on the retry — and since the estimate is now derived *after*
the plan clears, that retry is guaranteed on every issue. It also keeps the round cap and
the close-time `gate_rounds` metric counting review rounds rather than retries.

Same mechanism #183 wants at the close boundary, which is why it lives on the shared
`Ledger`.

## Protocol misses still count

A round whose judge emitted no valid block is persisted with `protocol_error` set and no
findings. Dropping it would freeze `len(Rounds)` at 0 for an agent CLI that never emits
the fence: the prompt would announce "this is the FIRST round" on invocation six, the round
cap could never fire, and `gate_rounds` would report 0 for the most expensive sessions.

## What close reports

A gate that remembers can also be *costed*. `sdlc close` and `sdlc milestone-close` print
two lines derived from the window they just reviewed:

```
  [ok] churn: prod 554 / test 300 / atlas 20 / workshop 778 (final 1652, rework 2.4×)
  [ok] plan gate: 6 round(s), 1 forced; findings 4 addressed / 1 withdrawn / 2 still open
```

**Churn** splits the window's surviving insertions four ways (`atlas/` → atlas,
`workshop/` → workshop, `*_test.go`/`testdata/` → test, **everything else** → prod —
including embedded prompt markdown and build files, which are production artifacts here).
**Rework** is insertions-across-commits over insertions-in-the-final-diff: `1.0` means
every line written survived, `2.4×` means the window wrote 2.4 lines per line it kept.
That ratio is the point — pair#127's final diff read as merely process-heavy and said
nothing about the file it rewrote five times.

**Two senses of "disposition", both reported.** *Forced* counts rounds the operator
bypassed; *addressed / withdrawn / still open* counts what happened to the findings. The
first answers "was the gate worked around", the second "were its findings acted on". Both
are free from the ledger, so neither is inferred from the other.

Three properties worth not breaking:

- **The window is `boundaryWindowBase`'s**, not a second derivation — so the churn provably
  covers the commits the review saw. The close's own uncommitted edits fall outside it, and
  that is the correct trade: diverging from the review's window would make the numbers
  incomparable across issues.
- **The operator lines are unconditional; the ledger row is not.** The calibration row is
  gated by #117's integrity rule (a milestone close carries a partial actual and must not
  pollute the ledger) and skipped again with no `brain/`. The report inherits none of that —
  emitting it from inside `appendCalibrationRow` would silently produce nothing on a
  milestone close, under `--no-actual`, or in any downstream repo without a sibling brain.
- **Everything degrades to zero with a warning.** A bad base SHA, a corrupt sidecar: warn,
  zero, close. No diagnostic is worth failing a close over. (An *absent* sidecar is not a
  failure — it is every issue closed before #187, and reads as honest zeroes.)

The same values land in the calibration ledger as ten appended columns — see
[ledger-landscape.md](ledger-landscape.md).

## Code map

| file | role |
|---|---|
| `construct/vocabulary/finding.cue` | the noun: severities, dispositions, handoff shape |
| `pkg/vocab/finding.go` | Go binding; renders the prompt's block instruction from the model |
| `cmd/sdlc/internal/gatestate/` | PURE core — ledger, ids, decision, parse, render, prompt block |
| `cmd/sdlc/planreview.go` | the IO shell: read/write the ledger (filesystem + clock live here only) |
| `cmd/sdlc/changecode.go` | wires it into the plan-quality gate |
| `cmd/sdlc/internal/judge/prompts/plan-quality.md` | dispose-first prompt |
| `cmd/sdlc/boundaryledger.go` | the boundary gate's ledger pair + seeding + per-round decision (#194) |
| `cmd/sdlc/internal/churn/` | PURE churn: four-bucket classification, rework ratio, numstat parsing |
| `cmd/sdlc/churnreport.go` | the close-time cost report — git seam + the two operator lines |
