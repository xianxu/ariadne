---
id: 000194
status: working
deps: []
github_issue:
created: 2026-08-20
updated: 2026-08-20
estimate_hours:
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
one of the 86 archived sidecars records `<base>..HEAD`. So the durable record of a
review does not say what was reviewed.

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

### E. Scope a re-review to what changed since the last round

With A and B in place this becomes expressible for the first time. Today
`boundaryWindowBase` gives a whole-issue close `merge-base(main, HEAD)` —
unconditionally, with no re-close narrowing anywhere in the tree — and a milestone
close the previous *finalized* boundary, which a REWORK never advances (it writes
nothing). So every round re-reads the full window and pays the full ~20 minutes.

A re-run whose prior round has an anchor and a ledger should review
`lastReviewedSHA..HEAD` with the prior findings carried in, rather than re-deriving
the whole branch.

**This trades integration coverage for wall clock, and that trade must be made
explicitly.** The whole-issue window is merge-base-scoped *by design* (#77) so the
final review sees the branch as it will ship. The mitigation is B: coverage becomes
*recorded* (every prior finding disposed) instead of *re-derived*. If that argument
does not survive the plan-quality gate, E is the part to drop — A–D stand alone.

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
- [ ] A re-run reviews only the commits since its prior round's anchor, carrying the
      prior findings — or, if that trade is rejected at plan-quality, the rejection is
      recorded in `## Log` with the reason and E is dropped.
- [ ] A doc-only commit landing mid-review no longer discards the review; a code
      commit still refuses, naming the commits it did not cover.
- [ ] The publish-time invariant is unchanged: no code ships unreviewed.
- [ ] Verified against a real multi-round history — `tools#1`'s
      `000001-define-m1-review.md` has four rounds and two clear families, so it is a
      ready-made fixture: a correct implementation flags `block-opener-rule` at round
      2, not round 3.

## Plan

Durable design: `workshop/plans/000194-review-anchor-commit-plan.md` (being rewritten
for the folded scope).

- [x] Design via `sdlc start-plan` before implementing.
- [x] Establish that #136 persists prose while #187's ledger persists addressable
      findings, and that only the latter is read back — the asymmetry this folds.
- [ ] Re-plan for the folded scope; re-run the plan-quality gate.

## Log

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
nothing. So each round re-read the whole window: 86 archived sidecars, 69 distinct
window strings.

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
