---
id: 000164
status: open
deps: []
github_issue:
created: 2026-07-05
updated: 2026-07-05
estimate_hours:
---

# docflow ship: ask whether to squash review rounds

## Problem

Today `docflow ship` (`scripts/docflow.sh`, `cmd_ship`) does a **`--no-ff` merge**
of `review/<slug>` into base and deletes the branch — deliberately preserving
every human/agent round commit so `git log` shows step-by-step how the artifact
was constructed (the script header explicitly states "No squash"). That's the
right default for durable authoring history.

But sometimes the author wants the opposite: **land the finished doc as one clean
commit and keep the inner making-of private** — the round-by-round back-and-forth
(and the agent's rationale in each round body) is scaffolding they don't want in
the published history of a manuscript.

Split out of **pair#89** (which kept the concurrent-edit-reconciliation scope);
this is the "ask to squash on ship" piece, which is substantively `docflow` +
`xx-fix` skill work and deserves its own design.

## Spec

Sketch (to be brainstormed — "a bunch of designs" here):

- **`docflow ship --squash`** — squash-merge `review/<slug>` into base as a single
  commit (subject e.g. `review(<slug>): <N> round(s), squashed`), then delete the
  branch. The `--no-ff` default is unchanged; `--squash` is opt-in. Guard behavior
  (refuse while `🤖` markers remain) is identical. Decide what the squashed commit
  body should retain (nothing / a rounds summary / the final agent rationale).
- **The agent asks at ship time.** When the operator says "ship it", the agent
  offers the choice: keep the rounds (default, `--no-ff`) or squash (`--squash`) —
  "shows how it was built" vs "keep the making-of private". This is `xx-fix` SKILL
  guidance (the "Shipping" section) — pair's `:PairReviewShip` poke carries the
  ship request; where the choice is surfaced (pane menu vs. agent chat prompt) is
  a design question, coordinate with the pair review-workbench flow
  (`pair/workshop/targets/review-protocol.md`).

Open design questions for the brainstorm:
- Does squash lose the per-round agent rationale that some authors *do* want? Offer
  a middle path (squash the doc but stash the rationale in the squashed body / a
  note)?
- Interaction with `--force` (the abandon path) — orthogonal, keep separate.
- Should the default be configurable per-repo (some repos always want private
  making-of)?

## Done when

-

## Plan

- [ ]

## Log

### 2026-07-05
- Created, split from pair#89. Grounding: `scripts/docflow.sh` `cmd_ship`
  (`--no-ff` merge, header "No squash"), `.claude/skills/xx-fix/SKILL.md`
  "Shipping" section. Needs its own brainstorm before planning.
