---
id: 000235
status: open
deps: []
github_issue:
created: 2026-09-18
updated: 2026-09-18
estimate_hours:
---

# Durable-plan lookup is exact-name: a plan under another slug is invisible

## Problem

sdlc finds an issue's durable plan by ONE exact name: the issue file's stem plus
`-plan.md` (`readOptionalPlanFile`, `cmd/sdlc/changecode.go`). A plan saved under
any other slug for the same issue id is invisible to it, silently.

Found in pair#283 (Log, 2026-09-18). The plan was saved as
`000283-default-cursor-style-plan.md`, while the issue file is
`000283-default-cursor-style-overrides-terminal.md`. Two consequences:

- **#231's flow inference said quick** although a durable plan existed. Renaming
  the file fixed it.
- **Plan-quality never saw the plan.** This predates #231: a misnamed plan is
  skipped by the plan gate too, and so the plan the agent wrote goes unreviewed.

The close-time review manifest resolves the plan the same way
(`reviewPlanPaths` / `optionalReviewPlanPath`, `cmd/sdlc/reviewwindow.go`), so
the boundary reviewer misses a misnamed plan as well.

## Spec

Find the plan by issue ID, in one place that every reader uses:

- A plan is `<plansDir>/NNNNNN-*-plan.md` for the issue's zero-padded id. The
  exact `-plan.md` suffix excludes the sidecars (`-plan-gate.md`,
  `-close-gate.md`, `-*-review.md`) by construction.
- Exactly one match is the plan, whatever its slug. A match that isn't at the
  canonical name also prints a one-line note naming the canonical name.
- Two or more matches are ambiguous: refuse, naming them, rather than pick one.
- No match means no plan, as today.

Route `readOptionalPlanFile` (change-code: flow inference and plan-quality) and
the review manifest's plan resolution through it. `familyFiles`
(`cmd/sdlc/resolve.go`) already globs an issue's family by id, so reuse it rather
than write a third lookup (ARCH-DRY).

## Done when

- A plan named `NNNNNN-<any-slug>-plan.md` is found by change-code (the flow
  inference and plan-quality) and by the close review manifest. The canonical
  name still works.
- Two candidate plans for one issue id refuse, naming both.
- A sidecar (`-plan-gate.md`, `-close-gate.md`, `-*-review.md`) is never taken
  for the plan.
- A test reproduces pair#283's names: an off-slug plan infers full and reaches
  plan-quality.

## Plan

- [ ] One id-based plan lookup; route change-code's plan read and the review
      manifest through it
- [ ] Tests: the canonical name, pair#283's off-slug name, two candidates
      refused, sidecars ignored

## Log

### 2026-09-18

Filed from pair#283 at the operator's request. It is meant as an example of the
small bug the quick flow (#231) should accelerate: a few lines in one or two
files, with a crisp Done-when and no durable plan.

One conflict to resolve before it can be one: in ariadne, #231 declared sdlc's
whole build closure (`cmd/sdlc/`, `pkg/`, go.mod/sum) a shared surface. sdlc
runs from the branch under review, so a quick branch could otherwise change the
gate it closes under. As a result every sdlc fix in ariadne, this one included,
is upgraded to full at close. Whether that declaration should be narrower is
the operator's call, and #231's trial is the place to decide it.
