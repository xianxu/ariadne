---
id: 000572
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'f3f888238822657c6cdc3089b8efc2d355e9966c' # card fields mirrored from issue-cards; edit via sdlc
---

# local-check: refuse to record a suite attempt invalidated by sandbox denials

## Problem

A failed local-check attempt has only two dispositions: register its tests as flakes (fail then pass is the evidence) or fix the code and push a new head (#394). Neither fits an attempt that its environment invalidated. In #561, `sdlc local-check run --name suite` was piped through `tail`. That broke the whole-call sandbox exemption, so the suite ran sandboxed and every failure was `operation not permitted` (resume-environment tests, TestQueueRealHelperPause, a golint socket test). The unpiped rerun on the same head passed, but sdlc refused to record it because the earlier failure was unexplained. Registering those tests as flakes would have falsified the flake record. The only honest way out was a new head and a second full suite: about 15 minutes of suite time plus queue wait. `local-check reclassify` covers startup-only failures, so it didn't apply.

## Spec

Detect a run invalidated by its environment and refuse to record it as a failure at all. When the failing tests' output is dominated by sandbox denials (`operation not permitted` / EPERM on file, socket or process operations), the run records neither a failure nor a pass. The checker posts no failing status and saves no failed attempt that blocks the head. It prints an error naming the cause: the run was sandboxed, so rerun it as the whole Bash call. Open questions for design:
- the classification rule: every failure carries a denial, or only a majority does;
- whether to add an explicit operator disposition for an existing attempt, such as `local-check invalidate --attempt ID --reason environment`, so attempts recorded before the detector existed can be cleared without a new head;
- how to keep a real permission bug in the code under test from being waved away. One idea is to require that the same tests pass on a rerun of the same head.

## Done when

- A suite run whose failures are all sandbox denials leaves no failed attempt and no failing `sdlc-local/*` status on the head, and says why.
- A run with any failure that is not a denial records as a failure exactly as today.
- Tests cover both cases plus a mixed run, using captured failure output.

## Plan

- [ ]

## Log

### 2026-10-23
