---
id: '000187'
status: done
started: 2026-07-29T10:26:05-07:00
created: 2026-07-28
updated: 2026-07-29
estimate_hours: 8.45
actual_hours: 2.32
---

# tune the change-code gate: stateful plan review, estimate after plan, churn metric

## Problem

Filed from a live cost postmortem on pair#127 (a two-defect terminal bugfix).
`sdlc change-code` was invoked **six times, five rejections**, ~3 min of judge
latency each, to gate a change whose final shape was 126 non-comment lines of new
logic. Final artifact churn: **+554 code / +778 workshop**. The operator's
question — "is the SDLC slowing you down?" — is fair, and the answer is that three
specific mechanics are, in ways fixable without weakening review.

The gates earned their keep on substance: round 1 moved a filter's seam and
deleted an entire category of specced work (tail-carry state); round 2 killed a
"defense in depth" layer that would have swallowed solicited terminal replies and
silently broken capability negotiation; the close review caught a live `panic`
(overlapping prefix/suffix checks inverting a slice bound) in code written twenty
minutes earlier. **This issue must not reduce that.** It targets the mechanics
that produced round-trips without producing findings.

### A. The plan-quality judge is stateless and memoryless — the root cause

`runPlanQualityJudge` (`cmd/sdlc/changecode.go:334`) builds a prompt from
`IssueContent` + `PlanContent`, dispatches, prints. It writes nothing and reads
nothing from any prior round. Contrast `reviewsidecar.go`, where **close and
milestone** reviews persist to `workshop/plans/NNNNNN-slug-{close,mx}-review.md`.

So every re-run is a brand-new reviewer that does not know it already reviewed.
It cannot converge, because it cannot see its earlier findings were addressed — it
re-derives an absolute bar each time and surfaces the next-deepest layer of a plan
that keeps improving. Observed on pair#127 as a clean descent: Critical →
Important → Info across rounds. A reviewer with memory says "you addressed my
three findings, ship" on round 2.

### B. Both estimate gates run BEFORE the plan judge

Order today (`changecode.go:120-190`): structural → estimate → estimate-recon →
plan-quality → estimate-quality. A reconciling itemized `## Estimate` block is
demanded before the plan has been looked at once. pair#127's first invocation
failed on "no `## Estimate` block" having received zero plan feedback.

The estimate is a **function of the plan**. The plan changed four times, so the
estimate was re-derived five times (1.26 → 1.69 → 1.38 → 1.05 → 1.47 → 1.40).
Only the last was an actual estimation error; the other four were forced
recomputation of a design that was not settled. Costing an unapproved plan is
waste by construction.

### C. The plan judge asks for test *enumeration*, which is pre-imaging code

pair#127's accepted plan listed ~15 test cases in prose. Every one was then
written as code — the prose was a lossy pre-image of an executable artifact.
Worse, the enumeration **missed the bug**: 30 hand-written cases all fed
syntactically valid sequences, and the close review found a panic on malformed
input. The one line that would have caught it is strategic, not enumerative:

> byte scanner over arbitrary device output → fuzz it, seeded with malformed forms

That single sentence subsumes the fifteen bullets and finds what they missed.

### D. No accepted-vs-forced record, and no cost measurement

`--force <reason>` prints to stderr and is not durable. Nothing records whether a
finding was acted on or overridden. So there is no way to answer "which gates earn
their cost" — and any cost metric built without that signal pushes toward
`--no-judge`, which is exactly how you lose the panic-catcher.
