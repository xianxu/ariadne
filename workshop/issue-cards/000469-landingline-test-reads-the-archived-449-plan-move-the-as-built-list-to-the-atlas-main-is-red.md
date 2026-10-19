---
id: 000469
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# landingline: test reads the archived #449 plan; move the as-built list to the atlas (main is red)

## Problem

ariadne main is red since #449 landed: `cmd/sdlc/internal/landingline/trainstate_test.go` `TestPlanNamesEveryTrainEventAndEffect`
reads `../../../../workshop/plans/000449-landing-train-plan.md`, which the close archived to `workshop/history/`. Every
recorded suite now fails there (seen on #458's integrated suite, e07ab835), so every ariadne landing is blocked.
