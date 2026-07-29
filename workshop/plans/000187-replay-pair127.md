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

### Rounds to acceptance: **2** (one rejection). Baseline: **6 invocations / 5 rejections.**

| round | wall | findings | outcome |
|---|---|---|---|
| 1 | 244s | 3 Important (PQ-1…3) + 2 Minor (PQ-4, PQ-5) | **BLOCKED** |
| 2 | 102s | 5 disposed `addressed`, **0 new**, 0 open | **PASSED** |

Round 2's message: *"no open blocking findings after 2 round(s)"*.

**The zero in "0 new" is the mechanism, not luck.** The stateless gate's pathology was
that each round re-derived an absolute bar and surfaced the next-deepest layer of a plan
that kept improving — five rejections for a 126-line change. Here round 2 was shown its own
prior findings, verified the five fixes, and stopped. It did not go looking for a sixth.

### Load-bearing check 1 — did the filter-seam relocation still surface? **YES.**

PQ-1 (Important), verbatim:

> Spec M2 says filter `tab.buffer` "as it is appended"; the Plan row says "stop replaying
> queries on redraw". Append-time runs under `m.mu` but needs carry state across `readPTY`'s
> 4096-byte chunks (run.go:707-721) and destroys the only record of app output. Replay-time
> stays pure but changes `redrawTab`'s lock contract, since `appendBuffer` re-slices
> `tab.buffer` from the output goroutine (run.go:748-751) — callers at run.go:699, 879, 926
> must either snapshot under `m.mu` or hand `redrawTab` the lock, with deadlock as the
> violation mode. Pick one and say which.

It surfaced **deeper** than the original gate's version: the original produced the seam
change; this one derives the concurrency contract each choice implies and names deadlock as
the failure mode of the alternative.

### Load-bearing check 2 — did the absorb-layer removal still surface? **YES.**

PQ-2 (Important), verbatim:

> `writeActive` targets the currently active tab (run.go:790) and the pump has no tab
> identity for an in-flight query, so a reply filter there also eats replies the active tab's
> own app solicited — silently breaking capability negotiation while the root-cause fix stays
> green. Done-when has no criterion for this layer. Either drop it with a Revisions note
> stating the input path is deliberately unfiltered, or name the tab-identity mechanism. Also
> state whether the query-in-flight-across-a-switch residual is accepted.

This is exactly the defect #187's Problem section credits to the old gate — "the 'defense in
depth' absorb layer that would have swallowed solicited terminal replies" — including the
"stays green" observation that makes it dangerous rather than merely wrong.

### Load-bearing check 3 (C1, negative) — **YES, and on the real artifact.**

The synthetic control was built because pair#127's ~15-prose-case plan is unrecoverable
(see the deviation above). It turned out not to be needed for this half: the gate fired C1
on the pre-gate revision's **four** enumerated cases. PQ-3 (Important), verbatim:

> The filter is a byte scanner over arbitrary child output, and Done-when lists only
> hand-written cases, which are blind to the malformed class (truncated CSI, overlapping OSC
> prefix/suffix on a short body, bare ESC at end of buffer) — a slice-bound panic there fires
> on the tab-switch path and kills `pair term` with every shell in the pane. **Replace the
> enumeration with one strategy line:** fuzz the filter seeded with truncated and malformed
> query forms; properties are never-panic, never-grow, never-drop-a-non-query-byte.

**This is the strongest single result in the replay, and it is stronger than the Done-when
asked for.** #187's evidence base records that in the original run, "~15 prose test cases
all rewritten as code, and *the class of bug they missed (malformed input → panic) was
caught only at close review*." The tuned gate caught that class **at plan time**, named the
three specific malformed forms, and named the blast radius. The gate did not merely survive
the tuning on this axis — it moved a bug class from close review to plan review.

### Load-bearing check 4 (C1, positive) — **YES.**

Round 2's between-rounds edit was deliberately written in strategy-line form (naming
`stripQueries` / `redrawTab` plus one adversarial-strategy line each, and two invariant
pins). It **passed with no test-surface finding**. So C1's two halves are both exercised on
the real artifact: the enumerated shape drew a finding, the strategy shape did not.

This is what makes check 3 mean something. A prompt that rejected *every* shape of test
description would satisfy check 3 alone and ship green.

### The synthetic C1 pair — run anyway, and it isolates shape from substance

`TestC1TestSurfaceShape`: one sound issue, two plans differing *only* in how the test work
is described. Both subtests passed (361s total).

**Enumerated variant → PQ-1 (Important), verbatim the Done-when's words:**

> **Test work is a 15-case prose enumeration; compress to functions-under-test plus one
> strategy line each.** The `- [ ] Tests:` block lists fifteen hand-written cases in prose
> ("a lone press is forwarded", "a wheel-up tick still routes to the scroll branch", …).
> Replace with the functions under test plus one adversarial-strategy line per risky
> function […] Name the test seam too (the existing `fakeMux` / `splitReader` doubles in
> `run_test.go`, no PTY) — that is the `ARCH-PURE` / `ARCH-MOCK` boundary the plan leans on
> silently.

**Strategy-line variant → no test-surface finding.** And this is the part that makes the
control informative rather than merely green: it still drew **three** findings, all about
substance —

- PQ-1 (Important): the plan names two SGR-terminator sites; there are three, and
  `parseSGRMousePress` (run.go:575) is the only place the release flag can be set, so the
  change as written leaves a third gate rejecting the release. Plus `rename_input.go:175-180`
  already recognizes the `M`/`m` pair, so the plan creates a second source of truth
  (`ARCH-DRY`).
- PQ-2 (Minor): no stated non-goals.
- PQ-3 (Minor): the held-buffer invariant is narrower than the bug class it closes — any
  `\x1b[<`-prefixed run with no terminator still accumulates unbounded.

So the strategy-line shape bought **no leniency on substance**. The gate reviewed it just as
hard and found a real third-site defect; it simply did not object to how the tests were
described. That is precisely the discrimination C1 asked for, and it could not be
demonstrated by either variant alone.

## Verdict

**The tuning did not weaken review.** Two rounds against six invocations, one rejection
against five — at strictly better finding quality on every axis this replay can measure:

- both findings the Done-when names as must-survive appeared in **round 1**, not eventually;
- the seam finding gained a concurrency contract the original did not produce;
- a bug class that originally escaped to close review was caught at plan time;
- and the gate stopped when its findings were addressed instead of descending further.

Round 1 also served as the live conformance check for the ` ```findings ` fence (Risk 5):
both rounds parsed with `protocol_error=""`, and round 2's five dispositions round-tripped
through the sidecar with their notes intact.
