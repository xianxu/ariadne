---
id: 000197
status: open
deps: [ariadne#194]
github_issue:
created: 2026-08-20
updated: 2026-08-20
estimate_hours:
---

# derive the boundary-review window from the gate ledger, not a hand-pasted Review-Verdict trailer

## Problem

A boundary review's window base comes from `previousReviewBoundary`
(`cmd/sdlc/milestoneclose.go:313`): the most recent commit touching the issue file
whose message carries a `Review-Verdict:` trailer. That trailer is **not written by
the binary**. `sdlc milestone-close` prints it to stdout under "paste into commit
message" and trusts the agent to paste it.

When the paste is missed, `previousReviewBoundary` finds nothing and
`boundaryWindowBase` falls back to the branch start — so the next milestone
re-reviews the entire branch, and so does every milestone after it. The failure is
**silent**: nothing warns that the boundary did not advance, and the review still
produces a correct verdict, just over a window several times larger than it needed.

`previousReviewBoundary`'s own doc comment names the behavior and calls it safe:

> *If a prior close's trailer was never pasted into its commit, this finds nothing
> and the caller falls back to the branch start — over-covering (re-reviewing prior
> work) rather than under-covering, the safe direction.*

Safe for correctness. Not cheap: a boundary review is ~20 minutes of wall clock
during which the tree is effectively frozen, so a missed paste costs a full re-read
of the branch at every subsequent boundary.

Measured on ariadne#194: M1's close review covered 19 files / 1899 insertions;
M2's covered 40 files / 3302 insertions, because the missed paste widened its base
back to the issue-creation commit. Roughly double the diff, for a window whose first
half had already been reviewed once.

(An earlier draft of this issue also claimed the wider window pushed the dispatch
past a wall-clock threshold where it failed outright. That was **wrong** and has been
removed: the two failed M2 attempts were caused by the host's `pmset -c sleep` being
set to 1 minute, which made every unattended run a coin flip regardless of length.
The claim rested on both attempts stopping near 8 minutes, which was coincidence.
Recorded here because a plausible-but-unfounded causal story in an issue is worse
than no story.)

### Observed, on ariadne#194 itself

M1 closed with `Review-Verdict: FIX-THEN-SHIP`. The trailer was printed and not
pasted. `git log --grep="Review-Verdict" main..HEAD` returns **nothing** for the
whole branch. M2's review window therefore resolved to `9cb22e7^..HEAD` — the
issue-creation commit — so it re-read all of M1's diff plus every planning commit,
having already been reviewed once.

This is a concrete instance of the cost ariadne#194 was filed about. #194's
`## Revisions` attributes the repeated full-window reads to REWORK never advancing
the boundary (a non-finalizing verdict writes nothing). That is real, but this is a
second and probably larger cause: a missed paste means the boundary never advances
**for any reason**, for the remainder of the issue.

### It gates the close too, not just the window

The trailer is read by **two** consumers, and the second one refuses rather than degrades:

- `previousReviewBoundary` derives the review window from it (above) — a missed paste
  silently widens the window.
- `milestoneHasVerdictCommit` (`cmd/sdlc/close.go:1717`) is the **milestone-verdict close
  gate**: it requires, per milestone, a commit that touches the issue file, has subject
  `^#<issue> <Mx>[: ]`, AND carries `Review-Verdict:`. Miss the paste and
  `sdlc close` refuses outright — *"milestones M1 lack Review-Verdict trailer in close
  commits"* — with `--no-verdict` the only way past.

Observed on ariadne#194: the paste was missed three times in one issue. Twice it widened a
window; the third time it **blocked the issue close**, and the fix was a commit whose only
purpose was to carry a trailer the binary had already computed and printed. A gate whose
evidence the binary produces, prints, and then requires a human to copy back into git is a
gate that fails on clerical error rather than on substance.

That makes the ledger the better source for both consumers: it is written by the binary,
under the repo lock, at the moment the verdict is known.

## Spec

The anchor should come from state the binary owns, not from a manual step.

Since #194 M2 the boundary gate keeps a durable ledger
(`workshop/plans/NNNNNN-slug-close-gate.md`) with one round per review, each stamped
with its `Boundary` (`gatestate.Round.Boundary`) and written by the binary at
finalize time under the repo transaction lock. That is a better anchor than a
commit-message grep in every respect: it cannot be forgotten, it distinguishes
milestones without parsing subjects, and it already exists.

Sketch, to be designed properly:

- Record the reviewed SHA on the round (`Round.ReviewedSHA`) — #194 M1 already
  resolves it; the ledger just doesn't persist it yet.
- `boundaryWindowBase` consults the ledger for the newest FINALIZED round at a prior
  boundary and uses its reviewed SHA. Fall back to the trailer grep, then to the
  branch start, so existing issues and pre-#194 history keep working.
- Only a **finalized** round may advance the boundary. A REWORK round must not — the
  fixes it asked for have to land inside the next window. This is the property the
  trailer accidentally had (a REWORK writes no commit, so no trailer) and it must be
  preserved deliberately rather than by accident.
- When the ledger and the trailer disagree, say so rather than silently preferring
  one.

### Not just a cost problem

`sdlc close --help` and the atlas describe the window as "the previous review
boundary". With a missed paste that sentence is false, and nothing surfaces the
discrepancy. Either the window is derived from state the binary controls, or the
docs have to describe a window that depends on whether an agent remembered to paste
a line.

## Done when

- [ ] The boundary window is derived from the gate ledger when one exists, with the
      trailer grep and branch start as ordered fallbacks.
- [ ] A REWORK (non-finalizing) round does not advance the boundary; a test pins it.
- [ ] A missed trailer no longer changes the window — a test closes two milestones
      without ever committing a trailer and asserts M2's window starts at M1's
      reviewed SHA.
- [ ] When the ledger-derived base and the trailer-derived base disagree, the
      divergence is reported rather than silently resolved.
- [ ] `sdlc close --help` / `milestone-close --help` / atlas describe the real rule.
- [ ] The milestone-verdict close gate is satisfiable from binary-owned state — a missed
      paste must not be able to block a close. A test closes two milestones without ever
      committing a trailer and asserts the whole-issue close still passes its verdict gate.

## Plan

- [ ] Design via `sdlc start-plan` before implementing.

## Log

### 2026-08-20

Found while closing ariadne#194 M2, by reading why that review's window base was the
issue-creation commit. The trailer had been printed at M1's close and never pasted;
nothing anywhere reported that the boundary had failed to advance.

Depends on #194 M2 having landed — the ledger this proposes to read is what M2
builds. Filed rather than folded into #194 because #194's scope (anchor, ledger,
families, convergence) is already three milestones and this is a distinct change to
window derivation with its own correctness question: which rounds may advance a
boundary.
