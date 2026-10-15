---
id: 000395
status: open
deps: [328]
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '8350df28aa55e78ad9b6ba5a2848c39869395b81' # card fields mirrored from issue-cards; edit via sdlc
---

# Feature-scoped mutation-guided test minimization

## Problem

Tests accumulate without a measured removal process. Coverage overlap alone cannot prove that a test adds no fault detection. #390's approved algorithm requires feature-level traceability before identifying candidates and mutation evidence before a reviewer decides what to keep.

## Spec

After #328 supplies a feature → code → tests map, run one mapped feature at a time. Measure each test's line/branch coverage and runtime alone. Nominate a test only when its covered set is contained in the union of the other mapped tests' sets; rank by possible time saved. Mutate a bounded slice of that feature's code, record which tests kill each mutant, and keep any uniquely killing test. For a candidate with no unique measured kill, a reviewer chooses remove, merge or keep based on assertions, requirements and distinct integration paths; mutation evidence does not automate deletion. Changes ship as ordinary reviewed issues and must preserve feature coverage and mutation score while reporting before/after time. Tool failures or incomplete mappings are unknown, never redundancy. Produce a feature report suitable for #327's monthly GC cadence and store measurement evidence with weekly snapshots under brain's charter. Follow #390 §5 and approved side doc 03; the method becomes an ariadne default for every SDLC repo after the ariadne/pair pilot.

## Done when

- [ ] A first mapped feature (couchcore if #328 maps it) has reproducible per-test coverage/time, subset nominations, a bounded mutant set and per-test kill matrix; unmeasured cases are explicit.
- [ ] A reviewer-facing report records each remove/merge/keep decision and reason, including tests retained for unique mutants or contracts absent from the mutant set.
- [ ] A test-only removal or merge, if justified, preserves the feature's coverage and mutation score and records measured time saved; otherwise the report explicitly recommends no deletion.
- [ ] #327's monthly GC procedure can consume the report feature by feature; tests prevent coverage-only automatic removal and catch mutation-score regression.

## Plan

- [ ] Design the per-feature report and mutation seam after #328's map lands.
- [ ] Pilot one feature with tests, review evidence and a normal close/landing path for any actual removal.

## Log

### 2026-10-15

Filed from #390's approved minimization algorithm. #328 is a blocking dependency; #327 owns the recurring GC schedule, this issue owns the feature-level procedure and pilot.
