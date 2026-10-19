---
id: 000463
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ff06260404a0574c04926c6000ee32bc16792225' # card fields mirrored from issue-cards; edit via sdlc
---

# local-check: resume recorded known-flake failures with targeted tests

## Problem

ariadne#461's recorded suite attempt fa09ccf7eb6fd0862fadf1d94e865944 failed
only TestSuiteQueuePriorityRealHelper when its private provider launch returned
Operation not permitted. The exact same commit passed that targeted test without
changes, and the test was registered on ariadne#462. Existing local-check run
can invoke the known-flake step only while processing a fresh full suite;
republish reposts an already-saved failure unchanged. There is no supported
way to finish this failed attempt with just the now-registered targeted tests,
so the hub approved another full suite for #461. This is a #394 follow-up.

## Spec

Add an SDLC verb to resume an already-failed recorded suite attempt whose named
failures are all currently registered known flakes. Run only those tests on
the original exact head through the existing known-flake verification path,
then post the suite success with its flake issue links and both attempt records.
Reuse the existing registry, classifier, rerun checks and evidence model rather
than introducing a second acceptance policy. Do not treat manual target output
or a republished failure as a recorded passing suite.

## Done when

- A saved failed suite whose only failures are registered known flakes can be
  resumed with named-test reruns only, without executing the full command again.
- The rerun uses the original exact commit and verifies the saved log/digest,
  repository identity, attempt completion and current registry/policy first.
- Success links the original failure, targeted rerun and open flake issues;
  every required test has an actual PASS. Other tests retain their original
  full-suite evidence instead of being silently marked as newly run.
- Unknown/unnamed failures, timeouts, missing or changed evidence, disabled
  policy, and closed/ambiguous flake registrations cannot become success.
  Revalidate registration and policy before posting.
- Cancellation and concurrent/repeated resume are handled with existing run
  ownership/evidence rules; no duplicate or unconfirmed run becomes a pass.
- Stateful regression tests prove targeted-only execution, exact-head binding,
  refusal paths and linked commit status/reporting.

## Plan

- [ ]

## Log

Filed unclaimed at the hub's explicit request, discovered during ariadne#461.
The current workflow required one extra full recorded suite despite established
same-exact-head fail/pass evidence. Original full log is retained under
ariadne/.git/sdlc/test-stats/failed/fa09ccf7eb6fd0862fadf1d94e865944.log.
The classified intermittent provider failure itself belongs to ariadne#462;
this issue owns only recovery of already-recorded failed attempts.
