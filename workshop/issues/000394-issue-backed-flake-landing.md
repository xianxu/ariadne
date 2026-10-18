---
id: 000394
status: working
deps: [392]
github_issue:
created: 2026-10-15
updated: 2026-10-18
estimate_hours:
card_mirror: '6a53e9abee241df666ecadd2a0f95825c603cf07' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T15:44:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

Durable plan: `workshop/plans/000394-issue-backed-flake-landing-plan.md`, with the state table. It awaits the hub's decisions D1–D4 before code.

- [ ] M1 — Registry and classifier (pure) plus `sdlc flake register`/`list`.
- [ ] M2 — The gate: a targeted known-flake rerun in `runLocalCheck` (local-check run and merge), both attempts linked, the policy knob.
- [ ] M3 — No blind rerun (the success-post guard, republish after registration), the rules in tl-worker/help/atlas, a pair pilot.

## Log

### 2026-10-15

Filed from #390's operator-approved flake/landing decision. The weekly snapshot #392 supplies the baseline and durable measurement view.

### 2026-10-18

- Dispatched by the hub (ops:0): the registry of known flakes is tied to their issues, and a rerun of just the registered tests replaces the whole suite. #449 builds on this. Plan drafted; it surfaces D1 (the targeted rerun amends #390's approved full-rerun rule), D2 (the accept policy), D3 (the daily view's scope: CI to a follow-up) and D4 (the registry lives in issue frontmatter).
