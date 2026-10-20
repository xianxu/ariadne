---
id: 000499
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1be4e2fb5d7144bdf40593ff8b0cde92c3522dbc' # card fields mirrored from issue-cards; edit via sdlc
---

# Waiting landing-train readiness retains a stale close anchor across external re-close

## Problem

A waiting landing-train car can remain permanently unready after its own branch is re-closed. During pair#711, merge pid 92126 started at 06:32 PDT on 2026-10-20, before a test-only correction and narrow re-close. It initially observed close evidence a64d76a2. The subsequent SHIP close bound c38a0d8a, and the branch advanced to 3a1cd887. The corrected recorded suite passed on its code ancestor 62319325 in 8m13s. Nevertheless landing status still reported the old a64d76a2 anchor and refused the two corrected test files. The runner remained alive, waiting, with no admitted landing.

Read-only source inspection: landingcar.go re-runs readiness every minute and reads the new local/remote PR head, but runLandingPublishGate composes tracker records through issuerecord.go. recordsScope caches root+issuesDir for the entire command; fresh records satisfy subsequent Fresh reads, and only writes by that command invalidate the cache. An external close therefore cannot update the cached completion anchor. The same class could hide an external reopen or changed completion, which should refuse rather than authorize stale evidence.

Recovery used: interrupt only the owning waiting merge runner (exit 130), then rerun central sdlc merge --yes. No gate waiver, shared queue mutation, status fabrication or source change.

## Spec

A waiting car readiness pass must evaluate one fresh, coherent tracker generation for that pass, including external close/reopen changes, rather than retain a command-lifetime generation indefinitely. Reuse within a single pass is fine; preserve exact-head, reviewed-anchor and fresh-card safety. Do not turn this into repeated unbounded fetches for each individual issue or weaken gates.

## Done when

- A deterministic waiting-car test observes an initial not-ready close, an external corrected SHIP close and head advance, then becomes ready without restarting the merge process.
- A negative control externally reopens or supersedes a previously ready close and prevents admission on old authority.
- Existing one-fetch/coherent-generation behavior within one readiness pass stays bounded and counted.

## Plan

- [ ]

## Log

Filed from pair:7 while landing pair#711, unclaimed. Native evidence: sdlc landing status reported no current landing, pid 92126 first in line and old-anchor publish refusal after the recorded suite succeeded. Source locations: cmd/sdlc/landingcar.go carReadiness/vouchWhileWaiting, landinggate.go runLandingPublishGate, issuerecord.go loadIssueRecordsAt/recordsScope.get. No Ariadne implementation changes.
