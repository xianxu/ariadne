---
id: 000499
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '073b5afbb03190e30e5f4f73131c7f650e645172' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T08:58:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "32259c11", done: "6b1c0c23"}
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
- A negative control observes external status changes at the readiness gate seam; a later pass must not authorize using the previous pass's cached status. The publish gate's policy for reopened cards is tracked separately.
- Existing one-fetch/coherent-generation behavior within one readiness pass stays bounded and counted.

## Plan

- [x] Reproduce stale readiness with one command context, a real tracked close, code/head advance, and an external corrected SHIP close.
- [x] Give each carReadiness call a nested records scope (ARCH-DRY), preserving parent context and pass-local coherence.
- [x] Verify external status changes at the gate seam, one fetch per pass at N/2N reads, and isolation from the parent scope.
- [ ] Run focused tests and recorded suite, then close/review and land through SDLC.

## Log

Filed from pair:7 while landing pair#711, unclaimed. Native evidence: sdlc landing status reported no current landing, pid 92126 first in line and old-anchor publish refusal after the recorded suite succeeded. Source locations: cmd/sdlc/landingcar.go carReadiness/vouchWhileWaiting, landinggate.go runLandingPublishGate, issuerecord.go loadIssueRecordsAt/recordsScope.get. No Ariadne implementation changes.

2026-10-20: Claimed and started #499. Root cause confirmed: recordsScope outlives carReadiness polls. Quick-flow plan posted to hub, proceeding unless objected. Reuse withIssueRecordsScope rather than invalidating shared state; no new state or persisted format. Each pass performs one tracker fetch/composition and O(N) card composition, with repeated reads served by that snapshot; count fetches for 8/16 reads. ARCH-PURE: retain production publish gate and real local bare remote in regression.

2026-10-20: Red first: TestCarReadinessObservesExternalReclose failed on the old reviewed anchor after a corrected SHIP close; TestCarReadinessRecordsGeneration failed at both 8/16 reads because external status remained cached (13.424s). Added only a nested records scope to carReadiness. Green: both regressions passed (15.159s), then focused readiness/scope/wait-loop/context guard tests passed (28.896s). Real bare remotes and production close/publish gates cover the original incident; the gate seam checks status visibility, pass coherence, bounded fetch count and parent-scope isolation. No lexical lint added: cache lifetime is a semantic observation-boundary property, defended by the regression rather than function-name matching. No new durable state or background work. Reopen policy probe failed independently with Ready:true after external codecomplete→working; filed unclaimed #502 per TL. Temporary failing probe removed after capture.

## Revisions

### 2026-10-20T09:03-07:00 — separate cache visibility from reopen policy

TL approved limiting #499 to fresh status visibility. Original acceptance: “A negative control externally reopens or supersedes a previously ready close and prevents admission on old authority.” Current acceptance checks status visibility at the injected gate seam. Existing selectTrackedLandingIssues skips working cards and validatePublishAnchors accepts no selections, so a reopened issue can currently remain ready even with fresh reads; that policy is a separately filed issue, not a new refusal in this cache fix.
