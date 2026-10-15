---
id: 000397
status: open
deps: [392]
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: 'cafc88ad65acb5deb8589d5df8864c7ace6f70ae' # card fields mirrored from issue-cards; edit via sdlc
---

# Measured soft package test budgets

## Problem

#390 identified package-runtime growth as a long-term suite cost, but the operator withdrew the proposed couchcore 600 s ceiling and 240 s target because one week of partial trend data does not justify values. The fleet needs several weekly snapshots before choosing budgets.

## Spec

After #392 has produced several comparable weekly summaries, present the observed per-package p90 distributions, sample coverage, host/cache differences and suite critical path for operator selection. Only after that decision, design versioned soft package budgets in `.sdlc/` for every SDLC repo by default. A budget overrun requests a measured explanation or optimization issue; it does not make a correct suite fail or raise a test timeout. Missing/insufficient measurements stay unknown. Pilot with pair couchcore, whose 481.3 s p90 over 21 samples is an observation, not an approved threshold. Follow #390's track-first revision; do not install the withdrawn 600/240 s values.

## Done when

- [ ] Several weekly snapshots with comparable package durations and coverage are reviewed, and the operator records pilot values and their measurement basis before any gate/declaration change.
- [ ] A versioned `.sdlc/` declaration and report compare complete package p90 samples to approved soft budgets; insufficient data stays unknown and one slow run cannot ratchet the baseline.
- [ ] Tests cover overrun, unknown coverage, budget revision and gate configuration; an overrun prompts explanation rather than masking a failed test or increasing a timeout.
- [ ] The pilot and default rollout are documented; local-gates changes follow operator verification.

## Plan

- [ ] Wait for the trend and operator-chosen values, then design the schema and soft response.
- [ ] Implement and verify without changing suite correctness gates.

## Log

### 2026-10-15

Filed as deferred follow-up from #390; no budget value is approved yet.
