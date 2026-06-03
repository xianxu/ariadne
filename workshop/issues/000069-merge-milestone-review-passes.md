---
id: 000069
status: open
deps: [000075]
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

## Refined design (2026-06-03)

Decided with operator (co-design session):

- **(A) Binary owns the review.** A review the agent can *skip* isn't a gate;
  binary-owned means it always runs, and the binary can also do the **cheap
  deterministic structural checks an agent forgets** (unticked-but-done boxes,
  missing `## Log` close entry, status not flipped) *before* spending tokens on
  the LLM pass. Division of labour: binary = cheap structural gate; one LLM
  review = judgment. So the agent runs `sdlc milestone-close`/`close`; it does the
  review; the agent does **not** separately dispatch `superpowers-requesting-code-
  review`.
- **One reviewer prompt** = reconcile the adapted superpowers `code-reviewer.md`
  (the general quality/architecture/testing/readiness checklist) with ariadne's
  milestone-review tweaks (issue-ref, the #70 `VERDICT:` contract line,
  `Review-Window` trailer, the **Atlas-update gate** and **Core-concepts
  cross-check**) **+ #75's `at-review` architectural lens**. Embed it as one
  source (à la #70/#75), used by the binary's review dispatch.
- **Close is a review boundary too.** `sdlc close` auto-dispatches the same
  review (whole-issue window): for a no-milestone issue that's the one review;
  for a multi-milestone issue it's an **end-of-issue review** (per-milestone
  reviews each see a slice — integration bugs + "do the milestones add up to the
  spec?" only show at the whole-issue diff). Plus the binary's structural checks.
- **Soft dep on #75** — the review consumes #75's `at-review` lens, so #75 lands
  first.

## Done when

- **One** fresh-context review per boundary (not two): the agent stops running a
  separate `superpowers-requesting-code-review` subagent; the binary's
  milestone-close/close review is the single pass.
- `sdlc close` runs the review (whole-issue window) + the cheap structural checks
  at the issue boundary; `sdlc milestone-close` does so at each milestone.
- The one reviewer prompt folds the superpowers checklist + ariadne tweaks + #75
  `at-review` lens; verdict feeds the `Review-Verdict:` trailer (#70 contract).
- AGENTS.md §3 + verb help reflect the single binary-owned pass (don't run both).

## Plan

*(draft — to refine at change-code; depends on #75)*

- [ ] M1 — reconcile the ONE reviewer prompt (superpowers `code-reviewer.md` +
  ariadne tweaks + #75 `at-review` lens), embedded as one source; `milestone-
  close` uses it; AGENTS.md §3 says the agent doesn't run a separate superpowers
  review. The adapted `construct/adapted/superpowers-requesting-code-review/`
  becomes the human-authored source the binary mirrors (or a thin pointer).
- [ ] M2 — `sdlc close` as a review boundary: cheap structural pre-checks
  (boxes/log/status) + auto-dispatch the one review on the whole-issue window
  (no-milestone → the review; multi-milestone → end-of-issue review); verb help +
  AGENTS.md updated.

## Log

### 2026-06-02

Filed from the sdlc tooling retro
(`workshop/pensive/2026-06-02-01-pensive-sdlc-tooling-retro.md`, finding F4). Operator:
"milestone-close = form; essence = adapted superpowers-requesting-code-review; tweak it
with the diff from milestone-close's judge."
