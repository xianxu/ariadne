---
id: 000473
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '9f0b23231092a05c445723b39b10170fc89bdf66' # card fields mirrored from issue-cards; edit via sdlc
---

# test-economy: affected-suite as close evidence; the full suite runs once, in the train or at merge (option B')

## Problem

A change pays the full suite twice today: once on its branch (close evidence), and again at landing when main has moved
with code (the integrated suite), or in a train (the train suite). Measured 10-19 (`sdlc test-stats`, 7 days): pair suite
12.5 min median, slot wait 11 min median / 35 min p90, ready → landed 54 min median / 2.3 h p90; ariadne's suite takes both
slots. The machine is suite-bound. Discussion and the operator's decision: ops side doc
`teams/devinfra/tl-status-xian-xu.side/2026-10-19-11-landing-schedule.md` (option B′, approved 10-19).

## Spec

- **Close evidence becomes the affected suite:** the packages the change touches plus every package that depends on them
  (exact from `go list -deps` over the repo; non-Go parts by declared path rules in `.sdlc/local-checks.json`, e.g. Lua tests
  when `*.lua` changed; anything unclassifiable → full suite). A new local-check name (e.g. `affected`) records
  `sdlc-local/affected` on the head; `close` accepts it where it accepts `suite` today. The full suite still works as
  evidence (superset).
- **The full suite runs once, before merging:** in the train (as built, #449), or at merge on the integrated head for repos
  without trains (as today). A car no longer needs a branch full-suite record to board; it needs the affected record +
  review + CI.
- **Safety valves:** changes to build/test infrastructure (Makefiles, go.mod, test helpers declared in config, sdlc config)
  require the full suite on the branch; a per-repo setting can turn affected-mode off.
- Measure: suite slot-minutes per landed change before and after (test-stats).

## Done when

- Fixtures: a change in one leaf package records `affected` covering that package and its dependents only; a go.mod change
  forces the full suite; close accepts `affected`; a train boards a car with `affected` only.
- One week after rollout, slot-minutes per landed change in pair drop (test-stats snapshot in the Log).

## Plan

- [ ]

## Log

### 2026-10-19
