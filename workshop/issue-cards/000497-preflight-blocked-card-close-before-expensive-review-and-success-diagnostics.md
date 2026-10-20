---
id: 000497
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
---

# Preflight blocked-card close before expensive review and success diagnostics

## Problem

A tracker-era whole-issue close may run its measured actual and expensive boundary review while the remote card is blocked, then reject completion only after SHIP. It misleadingly prints codecomplete publication success before the later blocked-card error. The recovery command repeats the same refusal until the owner changes status to working.

Observed in pair#722: sdlc close --issue 722 --no-plan-check --verified ... ran round5 to SHIP, made the close evidence commit, printed that codecomplete was published, then errored: card #000722 is blocked; only working (or a re-close of codecomplete) work closes. sdlc issue recovery reconcile --issue 722 refused completion.codecomplete with the same error. sdlc issue show 722 independently confirmed blocked/no card completion despite the recorded SHIP close. Supported repair was sdlc issue set-status working --issue 722, then recovery reconcile, which completed and pushed. No mirror edits or bypass were used.
