---
id: 000194
status: working
deps: []
github_issue:
created: 2026-08-20
updated: 2026-08-20
estimate_hours: 4.40
started: 2026-08-20T15:52:21-07:00
---

# boundary reviews: anchor to the reviewed commit, and remember across rounds

## Problem

> **Diagnosis superseded 2026-08-20 — see `## Revisions`.** The freeze described
> below is real but is a *symptom*. The measured cost was the same commits being
> re-reviewed from the branch point every round. Kept verbatim as the filing record.

A boundary review (`sdlc close`, `sdlc milestone-close`) computes its window as
`BASE_SHA..HEAD`, dispatches a fresh-context reviewer, and on return refuses to
finalize if `HEAD` moved meanwhile — `closeReviewSnapshot.validate()`
(`cmd/sdlc/close.go:1198-1206`): *"HEAD changed from X to Y"*.

The review takes ~20 minutes of wall clock. For that whole window the working
tree is frozen: the agent cannot commit anything, on any branch, without
invalidating a run that is already most of the way through.

**This is a stop-the-world barrier, and it is the single largest cost in the
process.** Measured over one session (tools#1 + tools#2, 2026-08-20):

- 8 boundary-review runs, ~20 min each
- 4 of them were dead time — no other work was possible
- 1 run was killed outright, mid-flight, because the operator sent a change
  request and stalling them was the worse option
- 1 run completed and was then discarded, because a follow-up change had to land
  before the fixes could be applied

The failure mode is not theoretical: it converts an *asynchronous quality check*
into a *synchronous lock on the repository*, and the natural response — killing
the review to stay responsive — is exactly the behaviour the gate exists to
prevent.

## Spec

Three defects, one mechanism. A boundary review today is **anchored to nothing and
remembers nothing**: it does not record which commit it read, and it does not read
what the previous round said. Everything below follows from those two gaps.

### A. Anchor the review to the commit it read

`resolveReviewWindow` (`cmd/sdlc/milestoneclose.go:243`) returns `head` as the
**literal string `"HEAD"`**. It stays literal through `collectDiff`, the judge
prompt, the `Review-Verdict:`/`Review-Window:` trailer, and the #136 sidecar — every
one of the 66 archived sidecar files (of 70) records `<base>..HEAD`, across 84 of 86
window rows — a re-run appends a `## Re-review` section, so rows outnumber files. So
the durable record of a review does not say what was reviewed.

Worse, `reviewThenFinalizeLocked` releases the repo lock *before*
`dispatchBoundaryReview`, and `boundaryReviewDispatchOptions` then resolves `"HEAD"`
independently — so the snapshot's `rev-parse` and the diff the reviewer actually read
can already name different commits.

Resolve HEAD once, under the lock, and thread that SHA through the diff, the prompt,
the trailer, the sidecar, and the finalize check. **This is the primitive B and C both
need** — neither round-scoping nor "what changed since the last review" is expressible
without it.

### B. Give the boundary review the ledger the plan gate already has

Reuse `gatestate.Ledger` rather than writing a second one (ARCH-DRY). It is already
gate-generic — `Gate string` and `IDPrefix string` are fields, and its own
`ContentHash` comment names the close boundary as an intended user (#183's
`--fixed-to-ship`). A boundary re-run then reads the prior rounds' findings, must
dispose of each (`addressed` / `not-addressed` / `withdrawn`) before raising new ones,
and cannot silently renumber. `judge.PromptInput.PriorFindings` already exists and is
documented "Empty for every category but plan-quality" — extend it to
`MilestoneReview`.

The #136 sidecar stays as-is: it persists the reviewer's *prose*, which is the right
artifact for a human or a resuming agent. The ledger is the *addressable* record the
next round consumes. Two artifacts, two consumers — not a replacement.

### C. Tag a `family:` on each finding, and escalate on repeat

A short slug naming the *underlying rule*, not the symptom — `block-opener-rule`,
`oracle-blind-direction`, `doc-claim-exceeds-measurement`. Assigned by the reviewer,
carried in the ledger. When a finding's family already has ≥1 disposed finding in a
prior round, the reviewer is instructed to change its recommendation:

> **This is the 3rd finding in family `block-opener-rule`.** Rounds 1 and 2 fixed
> instances. Do not fix this instance — state the rule that covers all three, and fix
> that. If the rule cannot be stated, say why, and record the family in `Limits` with
> its measured prevalence.

That instruction is the substance. The ledger exists to make it mechanical rather than
dependent on a stateless reviewer noticing across a document it was never shown.

Evidence (`tools#1`, four M1 rounds + two close rounds): one missing rule — *"a
part-of-speech word only opens a block when structurally placed"* — surfaced in four
shapes across four rounds. Each round fixed the instance; only the third wrote the
rule. A second family — *the measurement cannot fail in the direction that matters* —
spanned three more rounds. **Family escalation would have collapsed at least two of
the four M1 rounds.**

### D. Convergence line in the verdict

> round 4 — 2 new findings, 0 repeat families, 6 disposed. **Converging.**
> round 2 — 3 new findings, 2 repeat families. **Not converging: fix rules.**

Capping on finding *count* is arbitrary; capping when families stop repeating is not.

### E. Scope a re-review to what changed since the last round — REJECTED

> **Rejected 2026-08-20 by the operator: the reviewer keeps seeing the whole
> branch.** Kept as the record of what was considered and why it lost. B's
> recorded-coverage argument is not a substitute for a reviewer that has actually
> read the integrated result. See `## Revisions`.

With A and B in place this becomes expressible for the first time. Today
`boundaryWindowBase` gives a whole-issue close `merge-base(main, HEAD)` —
unconditionally, with no re-close narrowing anywhere in the tree — and a milestone
close the previous *finalized* boundary, which a REWORK never advances (it writes
nothing). So every round re-reads the full window and pays the full ~20 minutes.

A re-run whose prior round has an anchor and a ledger should review
`lastReviewedSHA..HEAD` with the prior findings carried in, rather than re-deriving
the whole branch.

**This trades integration coverage for wall clock, and the trade was refused.** The
whole-issue window is merge-base-scoped *by design* (#77) so the final review sees the
branch as it will ship. B makes coverage *recorded* (every prior finding disposed)
rather than *re-derived* — but recorded coverage is a weaker claim than a reviewer
having read the integrated result, and the integration read is the one this gate
exists for. A–D stand alone; the wall-clock win comes from B and C collapsing the
NUMBER of rounds instead of shrinking each one.

### F. A mid-review commit is a delta, not an invalidation

`closeReviewSnapshot.validate()` (`cmd/sdlc/close.go:1198`) refuses on any HEAD change.
`publishgate.go:175-193` already classifies the same question better one stage later
(`publishGateHasCodeSurface`, #174). Reuse it: a doc-only delta finalizes; a code delta
refuses **naming the commits** rather than reporting a bare "HEAD changed".

A code delta must still refuse. `runPublishGate` anchors on
`codecompleteAnchorCommit` — the *close commit* — so finalizing above an unreviewed
code delta would put the close commit on top of it, merge would compute
`closeCommit..HEAD` = 0, print `reviewed-HEAD-unchanged ✓`, and ship unreviewed code.
Making the other branch safe means re-anchoring the publish gate, which is out of scope.

Ranked last deliberately: it removes the freeze on close-time *bookkeeping*
(lessons.md, atlas, plan ticks), which is real but smaller than B–E.

## Done when

- [ ] A boundary review records the concrete SHA it read — in the trailer, the
      sidecar, and the finalize check — and the diff it dispatches is pinned to that
      same SHA (closing the lock-release drift).
- [ ] Boundary-review findings have stable ids and survive across rounds, using
      `gatestate.Ledger` rather than a second implementation.
- [ ] A re-run must dispose of every prior finding before raising new ones.
- [ ] Each finding carries a `family:` slug.
- [ ] A repeat family changes the reviewer's recommendation from "fix this instance"
      to "state the rule", and the verdict names the family and the repeat count.
- [ ] The verdict carries a one-line convergence summary.
- [ ] Every boundary review still reads the whole branch — the window source is
      unchanged, and a test pins that `boundaryWindowBase` still returns
      `merge-base(main, HEAD)` for a whole-issue close after this work.
- [ ] A doc-only commit landing mid-review no longer discards the review; a code
      commit still refuses, naming the commits it did not cover.
- [ ] The publish-time invariant is unchanged: no code ships unreviewed.
- [ ] Verified against a real multi-round history — `tools#1`'s
      `000001-define-m1-review.md` has four rounds and two clear families, so it is a
      ready-made fixture: a correct implementation flags `block-opener-rule` at round
      2, not round 3.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.30
item: cross-cutting-refactor design=0.1 impl=0.15
item: greenfield-go-module design=0.2 impl=0.25
item: smaller-go-module design=0.1 impl=0.15
item: greenfield-go-module design=0.15 impl=0.3
item: typed-data-prototype design=0.2 impl=0.15
item: smaller-go-module design=0.15 impl=0.2
item: smaller-go-module design=0.05 impl=0.15
item: atlas-docs design=0.05 impl=0.1
item: milestone-review design=0.0 impl=0.33
item: milestone-review design=0.0 impl=0.33
item: milestone-review design=0.0 impl=0.33
item: milestone-review design=0.0 impl=0.33
item: milestone-review design=0.0 impl=0.33
total: 4.40
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

Derivation, in plan order:

| item | covers |
|---|---|
| `cross-cutting-refactor` | M1.1 — thread the resolved SHA through `resolveReviewWindow` → `collectDiff`/prompt/trailer/sidecar/snapshot; mechanical but multi-file, and it moves test fixtures. |
| `greenfield-go-module` | M1.2 — `reviewanchor.go`: pure classifier + IO shell + two formatters, plus the goroutine/channel interleaving integration tests. |
| `smaller-go-module` | M2 — `boundaryledger.go`, a mirror of `planreview.go`. Cheap precisely because it is a mirror. |
| `greenfield-go-module` | M2 — `gatestate` (`Round.Boundary`, `FilterBoundary`, `BoundaryAll`), the D2 seeding, the `PriorFindings` wiring, and regenerating the byte-pinned `golden_test.go` fixtures. |
| `typed-data-prototype` | M3 — `family` on `#Finding`: the CUE model first, then `gatestate.Finding`, the parser, and `RenderBlockInstruction`. Typed-data because the closed schema is the source and three consumers derive from it. |
| `smaller-go-module` ×2 | M3 — `FamilyCounts`/`normalizeFamily`/`ConvergenceLine` + escalation rendering; then the `tools#1` fixture, the D3 near-miss and synonym tests, and the D5 + window regression tests. |
| `atlas-docs` | `ledger-landscape.md` (a second gate ledger joins the table) + close/milestone-close helptext. |
| `milestone-review` ×5 | Four boundaries (M1, M2, M3, whole-issue close) plus one expected re-round. |

### Two deliberate departures from the v3.1 defaults

**`milestone-review` is priced UNSCALED, at 0.33h.** v3.1's ×0.40 impl scaling exists
because the v2 table's implementation hours represent roughly sequential build effort
while `sdlc actual` measures compressed wall-clock. A boundary review is *already*
wall-clock and is *not* compressible — this issue's `## Problem` measures it directly
("8 boundary-review runs, ~20 min each") and its whole thesis is that the tree is frozen
throughout ("4 of them were dead time — no other work was possible"). Under #118 a
blocking subagent span counts as elapsed, so `sdlc actual` will bill every minute.
Applying a compression factor to an incompressible chunk is a model misapplication; the
unscaled v2 band (0.2–0.5) brackets the measured ~0.33h correctly. Raised by the
estimate-quality judge and adopted.

**`design-buffer` is 0.30, not 0.15.** The two tiers must agree: +15% presupposes v2.1's
×0.2 spec-quality discount, but three of the `smaller-go-module` design values (0.1,
0.15, 0.15) sit above the ×0.2 ceiling of 0.06 — they are priced at the ×0.5 tier, which
v2.1 pairs with +30%. The design values are the honest ones (D1–D5 are settled, but M3's
family anchoring still carries an open judgement call), so the buffer moves to match them
rather than the reverse.

### The fifth review chunk

Budgeting exactly one review per boundary would be optimistic on the very issue that
exists because rounds repeat: `tools#1` ran four rounds on one milestone, and this issue
burned three plan-quality rounds before implementation began. The sharper risk is
structural — M2 and M3 each close *through gate code the previous milestone just
rewrote*, so a regression there blocks the boundary rather than merely failing a test.
One extra chunk is a thin hedge against that, not a generous one.

## Plan

Durable design: `workshop/plans/000194-review-anchor-commit-plan.md`.
Three review boundaries, closed separately (AGENTS.md §3).

- [x] Design via `sdlc start-plan` before implementing.
- [x] Establish that #136 persists prose while #187's ledger persists addressable
      findings, and that only the latter is read back — the asymmetry this folds.
- [x] M1 — Anchor: resolve the review head to a concrete SHA once under the lock and
      thread it through diff/prompt/trailer/sidecar/finalize (closing the
      lock-release drift); classify a mid-review delta instead of refusing on HEAD
      identity, reusing `publishGateHasCodeSurface`. Standalone value: fixes a live
      defect regardless of M2–M3.
- [ ] M2 — Ledger: declare the boundary gate's `Gate`/`IDPrefix` pair that
      `planreview.go:26-30` already anticipates; wire `PriorFindings` into
      `MilestoneReview`; require every prior finding disposed before new ones.
- [ ] M3 — Families: model `family` on `#Finding` (closed schema — model first, then
      Go, then `RenderBlockInstruction`); escalate a repeat family from "fix this
      instance" to "state the rule"; emit the convergence line. Verified against
      `tools#1`'s four-round history as a copied fixture.
- [ ] M3 also pins the window: a regression test asserting the whole-issue review
      window stays `merge-base(main, HEAD)` (M4 was considered and rejected).

## Log

### 2026-08-20 (M2)

- **D4's protocol-miss clause revised: warn and persist, do not halt.** The plan said a
  boundary review emitting no `findings` fence should route to `closeHalt`. Two facts
  from the code overturned it: the fallback's failure mode (next round blind) is exactly
  the pre-#194 status quo rather than a regression, and the only ways past a halt —
  `--no-judge`, `--force` — skip the review *entirely*, so strictness would turn an
  occasional LLM formatting miss into a routine reason to run no review at all. The
  round is still persisted and marked `Blocked`, so the miss is visible rather than
  absorbed. Full reasoning in the plan's D4.
- **D2 landed as a seed, and the old instruction was deleted in the same commit.**
  `code-review.md`'s "Plan-gate carry-forward" section told the reviewer to read
  `-plan-gate.md` itself and re-raise open `PQ-*` findings — which, once a `BR-*` ledger
  existed, would have put two id namespaces into one output fence with no rule for
  disposing an id this ledger never issued. Now the still-open plan-gate findings are
  seeded into the boundary ledger under `BR-*` ids at `BoundaryAll`.
- `TestCodeReviewCarriesPlanGateForward` failed on that deletion, correctly — it guards
  the invariant that the plan gate's round-cap demotion is safe *only because* the
  boundary review picks those findings up. The mechanism moved, so the guard moved with
  it (`TestBoundaryReviewIsAskedToDisposePriorFindings` for the prompt half,
  `TestBoundaryReview_SeedsDeferredPlanGateFindings` for the mechanism half) rather than
  being deleted. `Decide`'s doc comment pointed at the removed section and was updated.
- ARCH-DRY: `readPlanGateLedger`/`writePlanGateLedger` and the boundary pair differed by
  nothing but a (gate, prefix, suffix) triple, so the bodies were **extracted** into
  `readGateLedger`/`writeGateLedger` rather than mirrored — a second copy of "a corrupt
  sidecar is an error, never a silent reset" is exactly the drift that rule guards.
- The ledger write is deferred to finalize time, matching the sidecar: `dispatch` runs
  with the repo transaction lock released, and the ledger is a repo write.

### 2026-08-20 (M1)
- 2026-08-20: closed M1 — go build ./... && go test ./... && go vet all green; 10 new pure unit tests over classifyReviewAnchor + both formatters (no repo), 2 window/abbrevSHA tests, and both interleaving directions in close_finalize_test.go — doc-only commit mid-review finalizes emitting the pass line, code commit refuses naming "concurrent #69 side change" and asserts absence of "HEAD changed from"; review verdict: FIX-THEN-SHIP

- M1 landed in three commits: `b19ac6b` (pin the SHA), `b2294ac` (snapshot takes
  it as a parameter), `7fd89bd` (classify the delta).
- The issue's Spec claimed `BASE_SHA..HEAD` was "already resolved to a concrete
  SHA at dispatch". Only the base was. `head` was the literal string `"HEAD"`
  from `resolveReviewWindow` all the way to the trailer and sidecar — which is
  why all 67 archived sidecars that carry a window row say `<base>..HEAD`.
- The live defect underneath: `reviewThenFinalizeLocked` releases the repo lock
  **before** `dispatchBoundaryReview`, and `boundaryReviewDispatchOptions` then
  re-resolved `"HEAD"` itself. A commit landing in that gap meant the diff handed
  to the reviewer named a different commit than the snapshot pinned. Resolving
  once under the lock closes it; that identity is now structural, not incidental.
- ARCH-DRY twice. `classifyReviewAnchor` calls `publishGateHasCodeSurface` rather
  than restating the docs-vs-code rule — the reuse the Spec asked for. And a
  planned `abbrevSHA` turned out to already exist in `state.go`; the duplicate
  was deleted and the original's doc extended.
- Found while wiring display: `shortSHA` runs `git rev-parse --short`, so it
  RESOLVES its argument — `shortSHA("HEAD")` returns the *ambient* repo's HEAD.
  Rendering the fallback head through it would print a commit the review never
  read. The trailer uses the pure `abbrevSHA`, so a degraded window stays
  visibly `..HEAD` rather than silently naming the wrong commit. Test pins it.
- ARCH-PURE: `classifyReviewAnchor` and both formatters are pure and unit-tested
  with no repo; `gatherReviewAnchorDelta` is the only thing that touches git.
- Four outcomes, not three. `reviewed` may not be an ancestor of HEAD at all
  after a rebase or `reset --hard`, and `git diff A B` between unrelated commits
  returns paths happily — without the ancestry check a rebase-away would
  classify as doc-only and finalize.

### 2026-08-20

Filed after a session that ran 8 boundary reviews across two issues in the
`tools` repo. The guard never produced a wrong verdict — every one of those
reviews found real, live-verified defects. The cost was entirely in
serialization, and the sharpest evidence is that the pragmatic response to an
operator request mid-review was to *kill the review*.

`publishgate.go` already contains the doc-only/code-delta classifier this needs;
the guard at `close.go:1198` predates it and asks a blunter question.

## Revisions

### 2026-08-20 — folded #195 in; corrected the diagnosis

**Reason.** The operator identified that the 8 boundary-review runs behind this issue
were driven by *iteration with a human* (repeated close attempts as feedback arrived),
not by the tree being frozen — and judged the agent's close-on-spec behavior correct,
since a human can always say "don't close, I want to smoke test first."

Checking that against the code disproved the hopeful reading that a re-close reviews
only the new commit. `boundaryWindowBase` (`cmd/sdlc/milestoneclose.go:271-283`) gives
a whole-issue close `gitx.MergeBaseWithMain()` unconditionally — there is no re-close
narrowing anywhere — and a milestone close the previous *finalized* boundary, which a
REWORK never advances because `finalizeBoundaryReview`'s `closeRework` branch writes
nothing. So each round re-read the whole window: 70 archived sidecar files carrying 86
window rows across 69 distinct window strings.

**Delta.**
- The original diagnosis ("stop-the-world barrier … the single largest cost") is
  demoted to a symptom; the cost is re-reading unchanged commits every round. The
  `## Problem` text is kept verbatim as the filing record, flagged at its head.
- `#195` (carry findings across rounds, tag families, escalate on repeat) is folded in
  as Spec B–D and punted with a pointer here. The two issues shared one root cause: a
  review that neither records what it read nor reads what it said. Neither half is
  useful alone — the ledger without an anchor cannot scope a round, and the anchor
  without a ledger has nothing to carry.
- New scope E (round-scoped re-review) — the lever that actually addresses the
  operator's loop, and only expressible once A and B exist.
- Original scope (anchor + delta classification) is retained as A and F, with F ranked
  last since it addresses bookkeeping friction rather than the measured cost.

### 2026-08-20 (later) — M4 rejected; the window stays whole-branch

**Reason.** Operator decision: *"I'd rather have review see whole branch still."*

Round-scoping a re-review (`lastReviewedSHA..HEAD`) was the only scope item that
would have shortened an individual round, but it means no single reviewer ever reads
the integrated branch — and the whole-issue boundary is merge-base-scoped precisely
so the last review sees what ships (#77). The ledger's recorded-coverage argument
(every prior finding disposed) is real but weaker: it certifies that findings were
handled, not that anyone read the result of handling them.

**Delta.** Spec E marked REJECTED in place (kept as the record of the road not
taken); M4 dropped from the plan; the corresponding Done-when item inverted into a
**regression test** that pins the window at `merge-base(main, HEAD)`, so a future
change cannot quietly narrow it. The wall-clock win now comes entirely from B and C
reducing the NUMBER of rounds rather than the size of each — which is the mechanism
the `tools#1` evidence actually supports (family escalation would have collapsed at
least two of four M1 rounds).

### 2026-08-20 (later still) — estimate revised 3.20 → 4.40 after estimate-quality

The gate passed the estimate as `info`, so this was optional; adopting it anyway,
because a knowingly-low estimate is exactly the pollution `--actual`'s
measured-not-typed rule exists to prevent, one stage earlier.

Three of the judge's findings were right and are applied in `## Estimate`:
`milestone-review` priced unscaled (v3.1's ×0.40 compresses sequential build effort;
a boundary review is already wall-clock and this issue's own Problem section
stopwatches it at ~20 min), the design-buffer tier raised to +30% to match design
values priced at v2.1's ×0.5 tier, and one out-of-band `smaller-go-module` re-slugged
to `greenfield-go-module` where 0.3 impl actually fits. A fifth review chunk was added
against the M2/M3 self-hosting risk — each closes through gate code the previous
milestone just rewrote.

The judge's fifth point — that recent ariadne v3.1 rows over-estimate — was **not**
applied. Those rows have actuals in the 0.22–2.32h range that `baseline-v3.1.md`
itself flags as possible measurement artifacts, and the two rows from the session that
produced this issue point the other way. The revision rests on this issue's own
stopwatch, not on the ledger.

