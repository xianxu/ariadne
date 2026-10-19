---
id: 000470
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# landing: a fleet-blocking grant holds the line even before the granted issue queues

## Problem

2026-10-19 ~12:20: ariadne main was red (#469 fixes it) and #469 held a fleet-blocking landing grant, but #469 had not yet
entered `sdlc merge`. The landing line admitted #465's waiting merge anyway: it integrated main and pushed, about to spend a
two-slot recorded suite on a known-broken main. A fleet-blocking grant only reorders merges already waiting; while the granted
issue is still being worked, everyone else keeps landing (and failing).
