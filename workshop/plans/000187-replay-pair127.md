---
type: prose
issue: ariadne#187
milestone: M2
task: 14
created: 2026-07-29
---

# Replaying pair#127's first plan through the tuned gate

The regression test for the one thing #187 could plausibly have broken: **did making the
gate converge faster make it review worse?** Fewer round-trips is only a win if the
findings that mattered still surface.

Everything below is the output of `cmd/sdlc/planreplay_test.go` (build tag `manual`), run
against a real judge. The harness records; the judgements here are read off its logs.

## Provenance of the round-1 input

`cmd/sdlc/testdata/pair127-round1-issue.md` is pair#127's issue file at
**`pair@70f6ac0`** (`issue-sync: update issues`, 2026-07-28) — recovered with
`git log --follow` over `workshop/history/issues/000127-term-pane-stream-corruption.md`,
since the file has since been archived.

Two independent signals confirm this is the state at the **first `change-code` invocation
for M2**, before any gate feedback landed:

- **`estimate_hours:` is empty.** pair#127 ran under the pre-#187 gate order, where the
  estimate gate came *first*. An empty estimate means no `change-code` had yet passed.
- **`## Plan` shows `- [x] M1` and `- [ ] M2`**, and M1's code lands in the *next* commit
  (`ac76a4e`). So this is mid-work: M1 finished, M2 planned and not yet gated — and the two
  findings #187 credits to the old gate (the filter-seam relocation and the absorb-layer
  removal) are both **M2** Spec bullets. This is the plan those rounds were arguing about.

**There is no plan document.** `pair/workshop/history/` holds exactly two #127 artifacts —
the issue file and `plans/000127-…-close-review.md`. The close-review sidecar archived
fine, so the absent plan doc is real rather than an archiving miss: #127's Plan lived *in
the issue file*. The replay therefore passes `planContent: ""`, and **round 1 sees
`PlanContent` empty — a controlled condition, not a gap.** It is what the original round 1
saw.

## Deviation: C1's controls are synthetic, and why that is sharper

Task 14 Step 4 planned to use pair#127's own ~15 prose test bullets as C1's negative
control, citing `000127-*.md:346`. **That text is not recoverable.** Git holds two states:
the pre-gate revision above (whose test work is *four* cases in one prose bullet) and the
final accepted plan (whose test steps are already in strategy form, with rationale). The
~15-prose-case version #187's evidence base cites lived in an **intermediate plan state
that was never committed** — the issue file was committed at `70f6ac0` and then not again
until the code landed.

So C1 is tested with a **controlled pair** instead: one issue, two plans differing *only*
in how the test work is described (`TestC1TestSurfaceShape`). That is strictly better
evidence for a claim about shape — the historical artifact would have varied substance and
shape at once, leaving any finding ambiguous between "too enumerated" and "underspecified".
The synthetic issue shell is deliberately sound (real defect, real root cause, real
done-when) so that a finding can only be about the test description.

Both halves are run, because only the pair is informative:

- **negative** — the enumerated plan must DRAW a finding telling it to compress;
- **positive** — the strategy-line plan must PASS without a test-surface finding.

A prompt that rejected *every* shape of test description would satisfy the negative half
alone and ship green. That is the failure mode this pair exists to catch.

## Results

<!-- filled from the harness run; see ## Verdict -->

## Verdict

<!-- filled from the harness run -->
