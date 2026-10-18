---
id: 000394
status: working
deps: [392]
github_issue:
created: 2026-10-15
updated: 2026-10-18
estimate_hours: 1.6
card_mirror: '9b1d7c9630ed36af3583de09704352229383c349' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T15:44:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
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

## Revisions

- 2026-10-18 (hub D3): Done-when 3's CI failure classes move to #451. Here the daily view covers local attempts, and CI is shown as `not covered` until #451 lands, never as zero.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.2 impl=0.22
item: smaller-go-module      design=0.05 impl=0.14
item: smaller-go-module      design=0.05 impl=0.2
item: smaller-go-module      design=0.05 impl=0.14
item: atlas-docs             design=0.0 impl=0.08
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
design-buffer: 0.15
total: 1.6
```

Derivation (*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*). The design is settled by the hub's D1–D4 and two plan-quality rounds, so a ×0.2 spec discount applies to the greenfield design (0.5–2h becomes about 0.2). impl= is 40% of the v2 midpoints:
- `internal/flakes` (registry, classify, evidence, rerun) is greenfield: 0.55 → 0.22.
- The card field, vocabulary and `flake register`/`list`: 0.35 → 0.14.
- The runLocalCheck gate, the `runUnderSlot` extraction and the v2 attempts, the riskiest piece, take the top of the range: 0.5 → 0.2.
- The success-post guard, republish and the `Unattributed` parse: 0.35 → 0.14.
- tl-worker help with its phrase guards, the help text and the atlas: 0.08.
- Three milestone reviews.

The calibration doc is marked stale (#127), so these numbers are provisional.

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
- The hub decided D1–D4 as proposed: the targeted rerun counts (the operator's 10-18 ask); `accept`/`off` are built with the guarded config left at its default; CI in the daily view goes to a follow-up; the registry is issue frontmatter. Side doc 02 got a Revision note.
- Plan-quality round 1 answered in the plan's Revisions. PQ-4 changes the registry's storage from details frontmatter to a card field, with `sdlc flake register` as its setter, with no owner check and read via Fresh tracker records; the hub is told of the delta. PQ-1: the rerun takes its own slot, lease and signals through the shared `runUnderSlot`. PQ-2: classification happens in runLocalCheck; the main baseline stays merge's unrecorded diagnosis. PQ-3: the decoder reads v1 and v2, and the success guard fails closed.
- The hub approved the card-field registry (`flake_tests`, setter `sdlc flake register`, compare-and-swap with no owner check, read with the Fresh status read). Rollout: one rebuild of :0 after landing, before anything can write the field.
