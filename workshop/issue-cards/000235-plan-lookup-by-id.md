---
id: 000235
status: open
created: 2026-09-18
updated: 2026-09-18
estimate_hours:
github_issue:
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
