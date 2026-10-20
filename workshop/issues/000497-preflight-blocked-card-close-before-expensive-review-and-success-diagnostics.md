---
id: 000497
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '871768d88cbdb9a5dfdd7e00d0ea32f26065f447' # card fields mirrored from issue-cards; edit via sdlc
---

# Preflight blocked-card close before expensive review and success diagnostics

## Problem

A tracker-era whole-issue close may run its measured actual and expensive boundary review while the remote card is blocked, then reject completion only after SHIP. It misleadingly prints codecomplete publication success before the later blocked-card error. The recovery command repeats the same refusal until the owner changes status to working.

Observed in pair#722: sdlc close --issue 722 --no-plan-check --verified ... ran round5 to SHIP, made the close evidence commit, printed that codecomplete was published, then errored: card #000722 is blocked; only working (or a re-close of codecomplete) work closes. sdlc issue recovery reconcile --issue 722 refused completion.codecomplete with the same error. sdlc issue show 722 independently confirmed blocked/no card completion despite the recorded SHIP close. Supported repair was sdlc issue set-status working --issue 722, then recovery reconcile, which completed and pushed. No mirror edits or bypass were used.

## Spec

Preflight the same tracker status precondition before actual measurement and judge dispatch. A blocked-card close should name the supported working/status-resume action without paying for a reviewer or writing a pending completion that cannot publish. Final success messages must follow confirmed card publication. Preserve existing ownership, re-close and interrupted-operation recovery guarantees.

## Done when

- A tracker-era blocked-card close refuses before judge invocation, measured-actual adoption and completion/evidence publication, with a supported next action.
- Working and valid codecomplete re-close remain accepted; recovery still completes a valid prepared close after the correct precondition is restored.
- Diagnostics do not report codecomplete publication success when the tracker transaction subsequently refuses it.
- Tests pin refused status and ordering via an injected reviewer counter and tracker publication failure.

## Plan

## Log

Filed from pair:5 after reproducing and recovering the pair#722 SHIP close. This follow-up is independent of drain behavior and is intentionally unclaimed.
