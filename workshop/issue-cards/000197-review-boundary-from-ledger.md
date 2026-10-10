---
id: 000197
status: open
created: 2026-08-20
updated: 2026-10-09
estimate_hours:
github_issue:
started: 2026-10-09T18:40:42-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
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
