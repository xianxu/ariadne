---
id: 000394
status: open
deps: [392]
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '4c3a48236fea2ec0b678eacea1120bcd181aa338' # card fields mirrored from issue-cards; edit via sdlc
---

# Issue-backed flake landing and daily triage

## Problem

An intermittent suite failure currently either blocks a landing or invites an unexplained retry. Both hide the actual test-health cost. #390's approved rule requires one passing exact-head suite while letting an already identified, separately owned flake be diagnosed and rerun on the same head.

## Spec

Keep the exact-head full-suite landing gate. When a test fails, retain the attempt and its log. If that test already has an open named flake issue, attach the failed attempt to the issue; a passing full-suite rerun on the same head may count for landing. An unknown failure blocks until classified: a reproducible branch bug is fixed on that branch; a newly established intermittent fault gets its own issue that day and only then becomes eligible for the known-flake rule. A blind passing retry never erases an unexplained failure. The TL owns a daily standing list across local and CI failures, with first-seen time, head, run/log, issue, owner and root-cause investigation or an explicit diagnostic blocker within one day. Flakes are fixed as separate root-cause issues, not by raising timeouts. Report CI failures separately from same-head local flakes; neither detector may treat absent data as success. Follow #390 §3 and approved side doc 02; make the policy an ariadne default for every SDLC repo after the ariadne/pair pilot.

## Done when

- [ ] A full suite passes on the exact landing head in every permitted landing; a known issue-backed flake's failed attempt and passing same-head rerun are both retained and linked.
- [ ] An unknown failure blocks even if a retry passes; classification tests distinguish branch bug, newly filed flake, known flake, and persistent main failure.
- [ ] The TL's daily view lists unowned or older-than-one-day intermittent failures from local attempts and CI with source coverage and issue links; CI failure classes remain distinct from confirmed same-head flakes.
- [ ] Tests use recorded attempts and stateful gate behavior to prove failure, issue filing, retry and head drift; no retry or timeout change silently turns red evidence green.

## Plan

- [ ] Design the issue-backed failure classifier and daily triage data sources with explicit unknowns.
- [ ] Test the gate transitions and reporting, implement the rule, and run a pilot in ariadne and pair.

## Log

### 2026-10-15

Filed from #390's operator-approved flake/landing decision. The weekly snapshot #392 supplies the baseline and durable measurement view.
