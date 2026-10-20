---
id: 000502
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c438b6f4cccd5ca9396fcd6787918ef0fae35eda' # card fields mirrored from issue-cards; edit via sdlc
---

# Reopened issue can silently remove a waiting car publish obligation

## Problem

A landing-train car can remain ready after its issue owner reopens it, even with fresh tracker reads. selectTrackedLandingIssues in cmd/sdlc/landingarchive.go delegates to ownedCompletions, which skips working cards. validatePublishAnchors in cmd/sdlc/publishgate.go accepts an empty selection, so the reopened card silently removes the review obligation.

Observed while fixing #499. A temporary extension to TestCarReadinessObservesExternalReclose used a real close-ready repo, SHIP close, correction, corrected SHIP close and productionCarGates with NoCI true. It then externally changed the card from codecomplete to working, preserving the existing completion binding, and called carReadiness again with the fresh per-pass records scope. Expected not-ready; got Ready:true PR:499. Command: go test ./cmd/sdlc -run '^TestCarReadinessObservesExternalReclose$' -count=1. Exit 1 in 13.650s; assertion “reopened issue authorized landing”. Host evidence: /tmp/ariadne499-reopen-red.log and /tmp/ariadne499-reopen-probe.go in ariadne:5; probe was removed from the passing cache fix after capture.

## Spec

Policy proposal for TL triage: an externally reopened card refuses its car with a named reason, since landing work its owner reopened is surprising. Decide how to establish branch ownership after the reopen and how truly issue-free branches behave; do not turn all empty publish selections into refusals accidentally.

## Done when

- A deterministic production-gate regression starts with a ready closed car, externally reopens its card, and observes not-ready with a named reason on the next readiness pass.
- Controls preserve legitimate issue-free branches and current completion generation rules.

## Plan

- [ ]

## Log

2026-10-20: TL ops:0 requested this separate unclaimed filing; #499 remains scoped to cache freshness/status visibility. Policy not implemented or presumed approved.
