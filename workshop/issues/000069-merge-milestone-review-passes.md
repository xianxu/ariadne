---
id: 000069
status: open
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
estimate_hours:
---

# merge the two per-milestone reviews into one boundary pass

## Problem

Each milestone boundary currently triggers **two independent fresh-context code
reviews** of the same diff:

1. The agent runs `superpowers-requesting-code-review` (a subagent) per AGENTS.md §3.
2. `sdlc milestone-close` **auto-dispatches** its own `sdlc judge milestone-review`
   (another agent) on the same window.

Observed in the 2026-06-02 `nous#41` session (4 milestones): the two passes were
**redundant** — the milestone-review judge mostly *confirmed* the superpowers review
rather than adding new findings — and **slow**: each `sdlc judge milestone-review` took
3–10 min (M3's worst), serializing the workflow with background waits the agent then had
to coordinate.

## Spec (operator framing)

> The two reviews should be merged. `superpowers-requesting-code-review` is borrowed from
> external (superpowers); `milestone-close` is home-grown. `sdlc milestone-close` intends
> to provide the **form**; the **essence** — in ariadne's design — should be the
> **adapted `superpowers-requesting-code-review`**. So fold milestone-close's judge into
> the adapted superpowers review: diff the two prompts and add milestone-close's tweaks
> (issue-ref awareness, the SHIP|FIX-THEN-SHIP|REWORK verdict line, the Review-Window
> trailer it emits) onto the adapted superpowers reviewer, then have milestone-close
> *invoke that one pass* instead of running a separate `judge milestone-review`.

There's already an adapted superpowers in the stack:
`construct/adapted/superpowers-writing-plans/` (sibling `superpowers-executing-plans/`) —
the requesting-code-review one should live alongside, and be what milestone-close calls.

## Done when

- One fresh-context review per milestone boundary (not two).
- `sdlc milestone-close` invokes the adapted `superpowers-requesting-code-review` (with
  the milestone-close tweaks folded in) and consumes its verdict for the
  `Review-Verdict:` trailer — no separate `sdlc judge milestone-review` dispatch.
- AGENTS.md §3 + the `sdlc judge`/`milestone-close` help reflect the single pass so the
  agent doesn't run both.

## Plan

- [ ] Diff `sdlc judge milestone-review`'s prompt vs the adapted
      `superpowers-requesting-code-review`; list what milestone-review adds (issue ref,
      verdict line, window trailer, atlas/lessons hooks).
- [ ] Fold those tweaks into the adapted superpowers reviewer (under `construct/adapted/`).
- [ ] Rewire `milestone-close` to invoke it once + parse the verdict; drop the separate
      judge dispatch.
- [ ] Update AGENTS.md §3 + verb help so only one review runs.

## Log

### 2026-06-02

Filed from the sdlc tooling retro
(`workshop/pensive/2026-06-02-01-pensive-sdlc-tooling-retro.md`, finding F4). Operator:
"milestone-close = form; essence = adapted superpowers-requesting-code-review; tweak it
with the diff from milestone-close's judge."
