---
id: 000572
status: open
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
github_issue:
---

# local-check: refuse to record a suite attempt invalidated by sandbox denials

## Problem

A failed local-check attempt has only two dispositions: register its tests as flakes (fail then pass is the evidence) or fix the code and push a new head (#394). Neither fits an attempt that its environment invalidated. In #561, `sdlc local-check run --name suite` was piped through `tail`. That broke the whole-call sandbox exemption, so the suite ran sandboxed and every failure was `operation not permitted` (resume-environment tests, TestQueueRealHelperPause, a golint socket test). The unpiped rerun on the same head passed, but sdlc refused to record it because the earlier failure was unexplained. Registering those tests as flakes would have falsified the flake record. The only honest way out was a new head and a second full suite: about 15 minutes of suite time plus queue wait. `local-check reclassify` covers startup-only failures, so it didn't apply.
