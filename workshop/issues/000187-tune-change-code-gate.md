---
id: 000187
status: working
deps: []
github_issue:
created: 2026-07-28
updated: 2026-07-29
estimate_hours:
started: 2026-07-29T10:26:05-07:00
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

## Spec

- **A1. Persist a plan-review sidecar.** Mirror `reviewsidecar.go`:
  `workshop/plans/NNNNNN-slug-plan-review.md`, appended one section per round,
  each finding carrying a stable id + severity.
- **A2. Feed the prior sidecar back.** `judge.PromptInput` gains the previous
  rounds. The judge must **dispose of every prior finding first**
  (`addressed` / `not-addressed` / `withdrawn`) before raising anything new, and
  must not re-raise a disposed finding at a lower severity.
- **A3. Converge deliberately.** Only unaddressed Critical/Important block. New
  findings at Info do not cost a round-trip — they land in the sidecar for the
  close review to pick up. Consider a round cap after which only Critical blocks.
- **B1. Reorder:** structural → plan-quality → (plan accepted) → estimate →
  estimate-recon → estimate-quality. Plan-quality becomes the gate that runs when
  no other trip is needed.
- **B2. Move estimate authoring downstream of plan approval.** Update the
  `change-code` helptext, `sdlc start-plan`'s output, `helptext/estimate.md`, and
  the base-layer `AGENTS.md` line that currently reads "The estimate is set later,
  at `start-plan` (required by `change-code`)" — it should say the estimate is
  derived **after** the plan clears plan-quality.
- **C1. Change what the plan gate demands of tests.** Ask for the **functions**
  that will be unit-tested by name, plus one line of strategy per risky function
  (the adversarial input class and the mechanical guard, e.g. "fuzz it"). Ask the
  judge to *reject* enumerated case lists, line-numbered call-site inventories,
  and procedural restatement of the diff — they pre-image code that is cheaper to
  write than to describe.
- **C2. Keep demanding what paid off**: decisions expensive to reverse (seams,
  layer boundaries, ownership, lock contracts); claims about existing behavior
  with the `file:line` backing them (both of pair#127's best catches were factual
  errors by the implementing agent about existing code); and explicit scope
  boundaries — what is deliberately NOT built, and why.
- **D1. Churn metric at close**, computed in the binary, four buckets:
  `code-prod` / `code-test` / `atlas` / `workshop`. **No comment-vs-non-comment
  split in v1** (deliberately deferred — this house style is comment-dense by
  design, so the ratio is descriptive at best and a Goodhart target at worst).
- **D2. Rework churn**: sum of insertions across the window's commits ÷ final
  insertions. Final-diff alone would have scored pair#127 as merely
  "process-heavy" and missed the actual waste, which was rewriting one file five
  times — much of it between invocations, never reaching a commit at all.
- **D3. Round-trips + disposition**, both free once A1 lands: rounds = sections in
  the sidecar; accepted-vs-forced = the disposition field. Emit alongside
  est/actual in the calibration ledger.

## Done when

- A second `change-code` on a plan whose Critical/Important findings were
  addressed passes without new blocking findings at a lower severity.
- The plan-review sidecar exists, accumulates rounds, and is read back in.
- `change-code` does not mention the estimate until plan-quality has passed; the
  helptext, `start-plan`, `helptext/estimate.md` and base-layer `AGENTS.md` all say
  so consistently.
- A plan naming test *functions* + strategy passes the gate; a plan enumerating 15
  prose test cases draws a finding telling it to compress.
- `sdlc close` prints the four-bucket churn split, rework ratio, round-trip count,
  and finding disposition; the calibration ledger records them.
- Replaying pair#127's first plan through the tuned gate costs materially fewer
  round-trips **while still producing the seam change and the absorb-layer
  removal** — those are the regression test for "did we weaken review".

## Plan

- [ ]

## Log

### 2026-07-28

- Filed from the pair#127 postmortem. Related: **#183** (FIX-THEN-SHIP needs
  better process) proposes gate-owned state keyed on content hashes for the
  *close* boundary; A1 here is the same shape — a gate that remembers what it
  asked for — applied to the *plan* boundary. Worth designing the two together so
  there is one notion of "gate state" rather than two (`ARCH-DRY`).
- Evidence base (pair#127): 6 `change-code` invocations / 5 rejections; +554 code
  vs +778 workshop; estimate re-derived 5×, of which 4 were forced by plan
  changes; ~15 prose test cases all rewritten as code, and the class of bug they
  missed (malformed input → panic) was caught only at close review.
- Counter-evidence to respect: the same gate produced the seam relocation, the
  absorb-layer removal, and two corrected factual claims about existing code. The
  goal is fewer trips at equal or better finding quality, not a lower bar.
