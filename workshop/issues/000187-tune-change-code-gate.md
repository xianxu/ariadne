---
id: 000187
status: working
deps: []
github_issue:
created: 2026-07-28
updated: 2026-07-29
estimate_hours: 8.45
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

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.* Design-buffer 0.15 (a thorough plan doc exists at
`workshop/plans/000187-tune-change-code-gate-plan.md`); `impl=` values are v2-table
implementation hours × 0.40 per v3.1, and — after the round-5 estimate-quality review —
**every item now sits at or under its primitive's v3.1-scaled ceiling.** Where work
exceeded a primitive, it is split into two items rather than written above the ceiling.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module      design=0.15 impl=0.20
item: greenfield-go-module   design=0.60 impl=0.32
item: smaller-go-module      design=0.10 impl=0.20
item: smaller-go-module      design=0.10 impl=0.20
item: smaller-go-module      design=0.05 impl=0.10
item: smaller-go-module      design=0.10 impl=0.20
item: skill-or-dispatcher    design=0.40 impl=0.20
item: cross-cutting-refactor design=0.30 impl=0.20
item: smaller-go-module      design=0.20 impl=0.20
item: cross-cutting-refactor design=0.30 impl=0.20
item: smaller-go-module      design=0.10 impl=0.20
item: atlas-docs             design=0.05 impl=0.08
item: milestone-review       design=0.00 impl=0.20
item: greenfield-go-module   design=0.50 impl=0.20
item: smaller-go-module      design=0.10 impl=0.16
item: smaller-go-module      design=0.10 impl=0.20
item: smaller-go-module      design=0.10 impl=0.16
item: atlas-docs             design=0.05 impl=0.08
item: milestone-review       design=0.00 impl=0.20
item: smaller-go-module      design=0.15 impl=0.20
item: smaller-go-module      design=0.10 impl=0.20
item: pensive                design=0.30 impl=0.12
design-buffer: 0.15
total: 8.45
```

Item → work mapping. **M1 (13):** `finding.cue` + vocab binding · `gatestate` ledger+Decide ·
`gatestate` parse + `FuzzParseFindingsBlock` · `gatestate` render/prompt +
`FuzzRenderParseRoundTrip` · first-fuzz-in-repo novelty (nothing to mirror — see below) ·
`planreview.go` IO shell · plan-quality prompt rewrite + goldens · `changecode.go`
gates-as-data refactor · stateful wiring + pass-through short-circuit · eleven doc surfaces
(incl. base-layer-symlinked atlas) · semantic-sweep consistency guard + `repoFilesExcluding` ·
`gate-state.md` + index · M1 review.
**M2 (6):** `churn` package · `churnForWindow` git seam · ten ledger columns · close wiring
(`closeFlags.PlansDir`, unconditional report, `DispositionCounts`) · atlas · M2 review.
**Replay (3):** harness · real-agent rounds + the strategy-line between-rounds edit ·
evidence record.

**Two model steps applied explicitly**, since the round-5 review flagged both as invisible:
- *Step 5 familiarity.* `familiarity` is block-level, so per-item novelty cannot be
  multiplied there. The fuzz work has nothing to mirror — verified, there is **no `func Fuzz`
  anywhere in this repo** — so `smaller-go-module` ("well-specced; mirror or extend") does not
  fit it alone; the novelty is carried as its own item (#5) rather than by inflating another.
- *Step 3 / Step 6 coupling.* The +0.15 buffer (not +0.30) is claimed on the thorough-plan
  ground, but the ×0.2 spec-quality discount is deliberately **not** applied to the design
  column. Reason: `started:` anchors the actual window at the claim commit (#116), so the
  plan authoring and all five gate rounds already spent land **inside** what `sdlc actual`
  will measure. Discounting design as "already credited" would systematically under-price
  exactly the hours this issue's own measurement will capture. The resulting ~52% design
  share matches recent ariadne rows (#172 61%, #180 41%, #186 53%).

**Calibration context:** the nearest structural analog is **#180 (est 8.10 → actual 11.93)**;
recent ariadne rows drift low (#171 0.41, #173 0.19, #186 0.56). 8.45 is therefore at the
optimistic end for work of this shape, and is recorded as such rather than padded.

## Plan

**M1 — the gate converges (A + B + C).** Plan: `workshop/plans/000187-tune-change-code-gate-plan.md` Tasks 1–10.

- [x] M1 — model the `finding` vocabulary in CUE + `pkg/vocab` binding, with the
      `code-review.md` severity drift guard (Task 1)
- [x] M1 — `gatestate` ledger: stable binary-assigned IDs, dispositions, open-set (Task 2)
- [x] M1 — `gatestate.Decide`: block only on undisposed Critical/Important; round cap
      demotes Important but never Critical (Task 3)
- [x] M1 — parse the fenced ` ```findings ` handoff, validated against the model,
      fail-closed (Task 4)
- [x] M1 — sidecar render/parse round-trip + the prior-findings prompt block (Task 5)
- [x] M1 — IO shell: persist `NNNNNN-slug-plan-gate.md`; corrupt sidecar errors rather
      than forgetting; archiving coverage pinned (Task 6)
- [x] M1 — rewrite `plan-quality.md`: dispose-first, strategy-not-enumeration, keep the
      hard-to-reverse-decision + unbacked-claim asks (Task 7)
- [x] M1 — wire `change-code`: stateful judge, `--force` recorded durably, estimate gates
      relocated below plan-quality (Task 8)
- [x] M1 — one story about estimate timing across helptext ×3, `startplan.go`,
      `AGENTS.base.md`, + `atlas/workflow/gate-state.md`, with a consistency guard (Task 9)
- [x] M1 — `sdlc milestone-close --issue 187 --milestone M1` (Task 10)

**M2 — the gate's cost becomes measurable (D).** Plan Tasks 11–15.

- [ ] M2 — `churn` package: four-bucket classification + rework ratio (Task 11)
- [ ] M2 — `churnForWindow` over the shared `boundaryWindowBase` window (Task 12)
- [ ] M2 — ten appended ledger columns + the close-time churn/round-trip lines, all
      degrading to zero rather than breaking a close (Task 13)
- [ ] M2 — replay pair#127's first plan: fewer rounds, but the seam relocation and the
      absorb-layer removal must still surface (Task 14)
- [ ] M2 — `sdlc close --issue 187` (Task 15)

## Revisions

### 2026-07-29 — plan authored; three deliberate deviations from the Spec

**Reason:** design decisions surfaced while writing
`workshop/plans/000187-tune-change-code-gate-plan.md` that change what the Spec asked for.

- **Sidecar is `NNNNNN-slug-plan-gate.md`, not `-plan-review.md`** (Spec A1).
  `construct/vocabulary/verdict.cue` declares `discovery: {home: "workshop/plans", glob:
  "*-review.md"}` — that glob asserts "this document carries a boundary verdict". A plan
  gate carries findings and no verdict, so filing it there would hand a future verdict
  consumer a document it cannot validate. Renaming costs nothing.
- **Severity vocabulary is `Critical | Important | Minor`, not `…| Info`** (Spec A3).
  `code-review.md` already owns exactly this taxonomy; single-sourcing the existing three
  names in `finding.cue` beats introducing a fourth (`ARCH-DRY`). Spec's "Info" ≡ "Minor".
  The judge *verdict token* `INFO` is a different noun and is untouched.
- **`--force` becomes durable** (Spec D says it "is not durable" — stated as a problem).
  A forced `change-code` now appends a round to the sidecar carrying the rationale. This
  is what makes D3's accepted-vs-forced count computable at all, so it is promoted from
  observation to deliverable.

**Delta:** no scope removed. Two milestones: M1 = A+B+C, M2 = D.

### 2026-07-29 — plan-quality round 1: FAILURE, seven findings, all confirmed

**Reason:** `sdlc change-code --issue 187` round 1. Every finding was verified against the
code before acting on it; all seven held up. Estimate 5.79 → 6.07 (added replay harness).

**Delta:**
- **[Critical]** Task 14's replay named `sdlc judge plan-quality`, which cannot do the job:
  `cmd/sdlc/judge.go:109-113` never populates `IssueContent`/`PlanContent` for any
  category, so the judge would review pair#127's plan without seeing it — and `sdlc judge`
  holds no ledger, so "rounds to acceptance" is uncountable through it. Rewritten to drive
  a scratch repo through the real `change-code` verb (manual-tagged harness). The
  `sdlc judge` defect is real but out of scope here → file as its own issue.
- **[Important]** Both protocol-miss paths returned before persisting, so `len(Rounds)`
  would stay 0 forever for a CLI that never emits the fence: the round cap could never
  fire, and M2's `gate_rounds` would report 0 for the most expensive sessions — inverting
  the D-series signal. Now a findings-less round with `protocol_error` is persisted on
  every path. Risk 2 gained a stated fail-closed trigger (<5% → drop the fallback;
  >20% → simplify the schema).
- **[Important]** Task 9 missed a sixth surface, `cmd/sdlc/changecode.go:147`, which prints
  "(set it at start-plan)" — the line an agent actually reads when the gate fires, and
  wrong at that moment after B1. Added to the file list and to the consistency guard
  (which now reads the two Go sources, not just prose). Its `rg` sweep was also
  unfalsifiable: scoped to `!workshop/history/**` it always matched this plan's own
  quotations. Rescoped to `!workshop/**`.
- **[Minor ×4]** `RenderBlockInstruction` moved to a method on `FindingModel` in
  `pkg/vocab`, mirroring `VerdictModel`'s placement rather than splitting the pattern
  across packages (`ARCH-DRY`); `--force` recorded only when the gate actually blocked
  (it is a *global* bypass, so unconditional stamping would over-report overrides in the
  one number meant to answer "which gates earn their cost"); Task 8's "real agent" step
  demoted to a prompt-shape smoke test (`--dry-run` returns before `judge.Dispatch`);
  `classifyFallback` declared as an explicit extraction of the existing
  `judge.Classify` switch.
- **Nits:** duplicate step numbering, `cap` shadowing the builtin, one self-contradictory
  test comment.

### 2026-07-29 — M2 pickup note (context handoff)

**M1 measured 1.55h against an 8.45h whole-issue estimate**, and M1 is the larger half
(13 of 22 estimate items). If M2 lands proportionally the ratio is roughly 4× OVER — in a
repo whose recent rows drift *under* (#171 0.41, #173 0.19, #186 0.56, and the named analog
#180 at 0.68). That is the opposite direction from the calibration context recorded when
the estimate was derived, and it is the single most useful signal this issue has produced
for velocity work so far.

Do **not** silently re-derive the estimate to make the ratio look better — the whole point
of the close-time ledger is to record it honestly. `sdlc close` will measure and adopt the
actual (#178); let it. Flag the variance in the close `--verified` evidence so the
calibration ledger carries the story, not just the number.

Worth noting *why* it may be inflated: the estimate was derived under gate pressure and
re-derived three times across five plan-quality rounds, each round adding items for work
the judge surfaced. Rounds that add scope to the plan also add items to the estimate, so a
stateless gate inflates estimates the same way it inflates round-trips. That is a #187-shaped
observation and belongs in the postmortem, not in a silent correction.

**Where M2 picks up:** `workshop/plans/000187-tune-change-code-gate-plan.md` Tasks 11–15,
all five unticked rows in `## Plan` above. Tasks 11–13 are the churn/rework/round-trip
metrics at close; Task 14 is the pair#127 replay (the regression test for "did we weaken
review", and per Risk 5 also the de-facto live conformance check for the ` ```findings `
fence); Task 15 is `sdlc close`. The plan's Task 12 and Task 14 already carry the
corrections from plan-gate rounds 1–2 (`gitx.RunGit` not `Capture`; drive
`runPlanQualityJudge`, not `runChangeCode`, which `os.Exit`s on gate failure).

**Loose end deliberately left unfiled:** `sdlc judge plan-quality` never populates
`IssueContent`/`PlanContent` (`cmd/sdlc/judge.go:109-113`), so any manual invocation
reviews an empty issue. Real defect in a verb the helptext advertises; recorded in Task 14's
rationale, out of scope for #187. File with `sdlc issue new` when convenient.

### 2026-07-29 — M1 boundary review: FIX-THEN-SHIP, all findings fixed before commit

**Reason:** `sdlc milestone-close --issue 187 --milestone M1`. Verdict FIX-THEN-SHIP;
per #174 the fixes land in the SAME commit as the close, and the review is not re-run.
Sidecar: `workshop/plans/000187-tune-change-code-gate-m1-review.md`.

**Fixed (all verified against the code first):**
- **C1 (Critical) — the `--force` consolidation silently killed two friction ACK patterns.**
  Task 8 collapsed five hand-written bypass messages into one template, changing the
  emitted strings (`structural gateS bypassed` → `structural gate bypassed`,
  `estimate-reconciliation` → `estimate-recon`), while `processmanual/gatesig.go:115,121`
  still matched the old wording. Those rows are `SilentAlone`, so the ACK line is the
  **only** observable evidence a gate was force-bypassed — a forced structural or
  estimate-recon bypass had become invisible to `sdlc process-manual` and every retro built
  on it. The consolidation was right; the consumer sweep was skipped. Patterns fixed, plus
  `TestForceAckMatchesGateCatalog`, which DERIVES the emitted string from
  `changeCodeGateOrder()` and asserts the catalog matches it — mutation-verified.
- **I1 (Important) — the pass-through did not prevent the re-dispatch it was built for.**
  `ContentHash` hashed the whole issue file, but the retry it exists to make cheap is by
  construction the one that edits both frontmatter (`estimate_hours:`) and body
  (`## Estimate`) — exactly what B2 instructs after a plan passes and the estimate gate
  refuses. So the hash differed and the judge re-dispatched: the very cost round 4 flagged.
  The live #187 exercise could not have caught it, because #187 already carried an estimate
  from the pre-M1 rounds. Fixed with `planGateContent`, which strips the estimate before
  hashing — correct on the merits too, since B1 removed the estimate from this gate's remit
  entirely. Two tests added, including the end-to-end "adding the estimate must not
  re-dispatch".
- **I2/I3 (Important) — two helptext surfaces Task 9 listed but never edited.**
  `change-code.md` still documented the pre-B1 gate order, never mentioned the stateful
  gate, omitted `--no-estimate-recon` from FLAGS, and left `WF_PLAN_ROUND_CAP` — a new
  operator-settable env var — undocumented against this binary's own convention.
  `start-plan.md` still said start-plan is where the estimate is set, contradicting
  `startplan.go`'s retimed runtime output on the one policy B2 exists to unify.
  **My own guard was blind to `start-plan.md`**: the sweep needs "estimate" and
  "start-plan" within 80 chars on one line, and there the identifier is the *filename*, not
  prose. Added it to the positive-assertion list with that reason recorded — a guard that
  passes while the thing it guards is broken is the failure mode worth naming.
- **I4 (Important) — `DispositionCounts` branched on disposition string literals**, the
  exact posture `finding.cue` argues against and `Closes()` exists to prevent. Latent
  today (M2 consumer), so it would have shipped as a silent under-count the moment a
  closing disposition was added. Now model-derived, with a coverage assertion.
- **Minors:** PQ-3's last stale "seven columns" (now ten); `TestRoundCapFromEnv` reads the
  ambient env before setting it (now hermetic); `structural()` told the operator to use
  `--force` even when `--force` was already supplied.

**Deferred with reasons, not silently:** promoting `exitWithCode` to a swappable var (the
review's suggestion for testing the gate loop end-to-end) — the C1 guard covers the message
contract, which was the actual risk, and changing a process-exit seam at a milestone
boundary is not a fix I want bundled into a FIX-THEN-SHIP commit. `finding.cue`'s
`discovery` block is declared but unconsumed — matching the `verdict.cue` precedent, so
it's a fleet-wide decision for its own issue, not a #187 change.

### 2026-07-29 — M1 landed; **the gate converged in 2 rounds, live, on its own plan**

**Reason:** Tasks 1–9 implemented and committed. The stateful gate was then exercised
end-to-end against a real judge on #187 itself — the first live proof of the whole loop.

**The headline Done-when, demonstrated:**

| round | findings | outcome |
|---|---|---|
| 1 | 1 Critical + 1 Important + 2 Minor | **BLOCKED** — ledger written, ids PQ-1…PQ-4 assigned |
| 2 | PQ-1 `addressed`, PQ-2 `addressed`, PQ-4 `addressed`, PQ-3 `not-addressed`, 1 new Minor | **PASSED** |

Round 2's message: *"no open blocking findings after 2 round(s); 1 advisory finding(s)
recorded for the close review."* Two rounds, against pair#127's six invocations / five
rejections — and the convergence is structural, not luck: PQ-3 was disposed
`not-addressed` (I had fixed only one of three stale mentions) and the gate correctly did
**not** block on it, because Minor never blocks. The new Minor raised in round 2 likewise
cost no round-trip. Both would have cost one under the old gate.

**Review quality was NOT weakened — round 1 caught two real defects:**
- **PQ-1 (Critical):** Task 14's replay harness read a return value `runChangeCode` never
  produces. Task 8's own gates-as-data refactor routes gate failures through
  `exitWithCode(1)` → `os.Exit`, so round 1 of the replay would have killed the test
  process — the round whose blocking *is* the point. Respecified onto
  `runPlanQualityJudge`, which returns an error and owns the full ledger path.
- **PQ-2 (Important):** Task 12 specified `churnForWindow` on `gitx.Capture`, which
  flattens every error to `""` (`window.go:50-56`, and its own doc warns against exactly
  this use). Task 13's "a git failure warns and leaves the values at zero" would have been
  unimplementable — a bad base SHA would print an all-zero churn line indistinguishable
  from an empty window. Switched to `gitx.RunGit`.

Both are seam-level findings about hard-to-reverse decisions — the class #187's Problem
section says the gate must keep catching.

**A fidelity bug live use found that the fuzz could not.** Round 1's PQ-1 detail came back
**truncated** mid-sentence. The judge wrote `## Estimate` inside a YAML *plain* scalar,
where ` #` begins a comment — so the finding silently lost its second half, and that
truncated text is what round 2 was shown as its prior finding. The gate's memory was
quietly lossy. `FuzzRenderParseRoundTrip` could not catch it: it fuzzes OUR render/parse,
not the judge's authoring. Fixed by rendering title/detail/note as `|` block scalars in the
handoff instruction (immune to `#` and `:`), pinned by a test that asserts both halves —
block form survives, plain form loses text — so nobody "simplifies" it back.

**Also from M1 implementation:** `FuzzRenderParseRoundTrip` failed within one second of its
first run on input `"\n0"` — `go.yaml.in/yaml/v3` mis-emits leading-newline strings,
writing a block-scalar indent indicator (`|4-`) contradicting its own 8-space indentation,
so its own parser rejects the result. A finding whose detail began with a newline would
have written a ledger that could never be read back, destroying that issue's gate memory
permanently. Fixed by canonicalizing agent prose at the schema boundary plus a defensive
canonicalization at the write boundary.

**Mutation-verified guards** (a guard that cannot fail is worse than none): swapping the
gate-order literal makes `TestGateOrderPlanBeforeEstimate` fail; reverting
`helptext/issue.md` to the old estimate story makes `TestEstimateTimingConsistency` fail.

**Estimate-quality on round 2: INFO** (non-blocking), five methodology notes. One is a
factual slip worth correcting here: the design-share comparison mixed conventions — the
peer figures (#172 61%, #180 41%, #186 53%) are pre-buffer `est_design/estimate`, so the
like-for-like number for this issue is **45.6%** (3.85/8.45), not the post-buffer 52%.
The conclusion is unchanged (45.6% sits inside 41–61%). The remaining four notes —
near-uniform ceiling selection on impl, the Step 3/Step 6 coupling, the un-itemized sunk
cost, and item 5 being a multiplier wearing an item's clothes — are recorded as advisory;
the estimate has been re-derived three times already and further tuning would cost more
than it informs.

### 2026-07-29 — gate PASSED on round 5; estimate re-derived on the estimate-quality review

**Reason:** `sdlc change-code --issue 187` round 5. **plan-quality cleared**; branch
`000187-tune-change-code-gate` created in place. estimate-quality returned **INFO**
(non-blocking) but with findings worth acting on, since the estimate feeds velocity
calibration and is measured against at close. Estimate 7.44 → 8.45.

**Delta (estimate only — no plan changes):**
- **Impl values were pinned at-or-above ceiling.** Five of sixteen items exceeded their
  primitive's v3.1-scaled ceiling and eight more sat exactly at it, so the primitive wasn't
  estimating — judgment was, layered on top, while the block's prose claimed "× 0.40 per
  v3.1". Fixed by **splitting** oversized work into honest second items rather than writing
  above the ceiling: 16 items → 22, and every impl now sits at or under its ceiling.
- **`atlas-docs` was the wrong slug for Task 9.** That task touches eleven doc surfaces
  (two base-layer-symlinked) *and* writes a new repo-wide semantic-sweep guard with a helper
  — priced as "docs maintenance" at 3.5× its ceiling. Now `cross-cutting-refactor` +
  `smaller-go-module` + a small `atlas-docs`.
- **Task 14 was priced as routine despite being the plan's own named risk** (round 3 logged
  it "advisory, not acted on"). Now three items covering the harness, the real-agent rounds
  + strategy-line edit, and the evidence record.
- **Arithmetic drift.** 7.44 was 0.12 above what its items derived (7.32) — it reconciled
  only because the 5% tolerance absorbed it. For a number the helptext calls "DERIVED, not
  guessed", riding the tolerance is the wrong habit. The new block is verified by
  recomputation: Σdesign 3.85 × 1.15 + Σimpl 4.02 = 8.4475 vs stated 8.45 (δ 0.0025).
- **Two model steps made explicit** rather than silently skipped: per-item novelty for the
  fuzz work (block-level `familiarity` can't express it, so it is its own item), and why the
  ×0.2 spec-quality discount is deliberately NOT applied (with `started:` anchoring at the
  claim commit, plan authoring + all five gate rounds land *inside* the measured window).
- **Calibration context recorded:** nearest analog #180 est 8.10 → **actual 11.93**; recent
  ariadne rows drift low. 8.45 is at the optimistic end for this shape, stated not padded.

**Gate cost, for this issue's own dataset:** 5 plan-quality rounds, ~4 min each, 0 forced.
Findings across rounds: R1 1C+2I+4m, R2 4I+4m, R3 1C+2I+3m, R4 1C+2I+3m, R5 clean. **Every
round found real defects** — all verified against the code before acting, none re-raised —
including one (R4) that showed B1 as specced would have made the gate *more* expensive.
That is the counter-evidence this issue's Problem section says to respect: the gate is
expensive AND it is earning it. What it never did across five rounds is *remember* — which
is exactly what A1/A2/A3 fix.

### 2026-07-29 — plan-quality round 4: FAILURE, 1 design flaw + 2 Important + 3 Minor

**Reason:** `sdlc change-code --issue 187` round 4. No re-raises again. This round found a
real flaw in **B1 itself**, not in the plan's rendering of it. Estimate 6.89 → 7.44.

**Delta:**
- **[Critical, design flaw in B1]** Moving the estimate gates below plan-quality inverts a
  cost that was free. Today an estimate failure exits at `changecode.go:143`/`:158` in
  milliseconds. After B1: plan-quality dispatches (~3 min) → passes → estimate gate fails
  (no `## Estimate` block yet — *exactly what B2's new prose instructs*) → the retry
  dispatches plan-quality **again**. So B2's own instruction guarantees a wasted 3-minute
  re-dispatch on every issue, and pair#127's one genuine estimation error would now cost a
  full judge round. **B1 as specced would have made the gate more expensive** — the
  opposite of this issue's purpose.
  Fix is the mechanism this issue's own Log already named: gate state keyed on a **content
  hash** (#183's shape, "worth designing the two together", `ARCH-DRY`). One `Ledger.ContentHash`
  field; when the last round passed and `sha256(issue+plan)` is unchanged, skip dispatch
  entirely, persist no round, fall through to the estimate gates. Three tests added
  (pass-through, re-dispatch on edit, never cache a blocking round).
  Two second-order bugs it also fixes: `CapReached` counted *invocations* (estimate-driven
  and `protocol_error` re-runs would burn the cap of 3 and silently demote a real Important
  at round 4), and M2's `gate_rounds` would have counted exactly the noise the reorder
  creates — reporting the tuning as *more* expensive for reasons unrelated to review quality.
- **[Important, `ARCH-PURPOSE`]** Task 9's surface list was incomplete and its sweep could
  not find the misses. A **semantic** sweep (`rg -i 'estimate.{0,80}start-plan|start-plan.{0,80}estimate'`)
  surfaces five more live surfaces the literal sweep never could: `helptext/issue.md:35`,
  `helptext/set-status.md:15`, `atlas/workflow/issue-lifecycle.md:35`+`:93`,
  `atlas/workflow/sdlc-binary.md:35` (its gate-order table row is verbatim the order B1
  inverts), and `startplan_test.go:116`. **`construct/base.manifest:214` is
  `symlink atlas/workflow`** — verified — so two of those propagate to every downstream
  repo. The consistency guard now *is* the semantic sweep with a two-entry allowlist,
  rather than a list of literals that passes clean while five surfaces tell the old story.
  Also corrected: `helptext/estimate.md` carries no timing claim at all, so counting it
  toward "all five places" while `issue.md`/`set-status.md` went unlisted was the miscount.
- **[Important, `ARCH-DRY`]** `changeCodeGateOrder()` guarded a *restatement*: moving the
  estimate block back above plan-quality would leave `TestGateOrderPlanBeforeEstimate`
  passing — the sole guard for B1 could not fail on its own regression. `runChangeCode` now
  **iterates** a `[]gate{name, run}` declaration, so the list *is* the order. Bonus: this
  collapses the five near-identical `--force` handling blocks (`changecode.go:124-136`,
  `:144-152`, `:159-167`, `:172-177`, `:178-183`) into one.
- **[Minor ×3]** Ledger column arithmetic was off by one throughout Task 13 — ten columns
  are appended, not nine, so they occupy indices 10–19 (total 20); `len(c) >= 19` would
  have admitted a 19-column row and panicked reading `c[19]`, and `cols[16]` is
  `gate_forced`, not `gate_addressed` (17). Cited `buildMilestoneReviewDispatch`, which
  does not exist — the function is `boundaryReviewDispatchOptions` (`milestoneclose.go:582`);
  the substantive claim held. C1's **positive** control was unverified (only the negative
  half was) — Task 14's between-rounds edit is now specified in strategy-line form and
  recorded as a fourth check, so a prompt that rejected *every* test-description shape
  can't ship green.

### 2026-07-29 — plan-quality round 3: FAILURE, 1 Critical + 2 Important + 3 Minor

**Reason:** `sdlc change-code --issue 187` round 3. Again no re-raises. Estimate 6.81 → 6.89.

**Delta:**
- **[Critical]** Task 14's replay harness passed `DryRun: true` — which, by this plan's own
  Task 8 Step 6 note (`changecode.go:357-367` returns before `judge.Dispatch`), means no
  agent runs, no findings block is produced, no round is persisted, and `err` is always
  nil. The one deliverable protecting the gate's *value* was a structural no-op. Dropped
  `DryRun`; added `NoEstimate`/`NoEstimateRecon` so the replay isolates the plan gate now
  that B1 puts the estimate gates downstream of it.
- **[Important, `ARCH-PURPOSE`]** The demotion policy's entire safety argument — "Minor and
  post-cap Important findings land in the sidecar for the close review to pick up" (Spec
  A3) — was asserted in four places and **wired in none**. Verified:
  `buildMilestoneReviewDispatch` (`milestoneclose.go:596-600`) passes no plan-gate content,
  and `grep -i 'plan-gate\|plan gate\|plan-quality' code-review.md` returns nothing. The
  cheap half (stop blocking) would have shipped while the half that makes it safe stayed
  documentation that doesn't derive. Task 9 Step 5 now adds a `Plan-gate carry-forward`
  block to `code-review.md` (the reviewer already has `Read`) plus a guard test.
- **[Important]** Done-when *"a plan enumerating 15 prose test cases draws a finding telling
  it to compress"* had **no verification step anywhere** — Task 7's two guards test
  plumbing (model renders, PriorFindings arrives), not C1's semantics. pair#127's plan is
  the natural positive control (its ~15 prose bullets are cited at `000127-*.md:346`), so
  Task 14's evidence record gains a fourth recorded question at zero extra agent cost.
- **[Minor ×3]** `TestPlanQualityPromptRendersFindingModel` wouldn't compile —
  `Dispositions` became `map[string][]string` when the closing/open partition landed, so
  the range yields `[]string`; switched to `AllDispositions()`. "Disposition" named two
  different nouns (D3's accepted-vs-forced vs. the `addressed|not-addressed|withdrawn`
  type) — resolved by emitting **both** and adding `gate_addressed`/`gate_withdrawn`/
  `gate_open` columns (nine appended, `len(c) >= 19`). Estimate-risk concentration in
  Tasks 5 and 14 noted as advisory, not acted on.

### 2026-07-29 — plan-quality round 2: FAILURE, four Important + four Minor, all confirmed

**Reason:** `sdlc change-code --issue 187` round 2. The judge explicitly re-verified round
1's seven fixes and re-raised none — the descent is into new substance, not noise. Every
finding was checked against the code before acting. Estimate 6.07 → 6.81.

**Delta:**
- **[Important]** Task 14 rested on a plan file that **does not exist**: `pair/workshop/history/`
  holds only `issues/000127-term-pane-stream-corruption.md` and its `-close-review.md`
  sidecar. pair#127's Plan lived *in the issue file* — the close-review sidecar archived
  fine, so the absent plan doc is real, not an archiving miss. The replay now recovers the
  issue file's `## Plan` via `git log -p --follow`, drops `writePlanFile`, and records
  "round 1 sees `PlanContent` empty" as a controlled condition (which is what the original
  round 1 saw).
- **[Important]** The churn report was planned inside `appendCalibrationRow`, which most
  closes never reach — `shouldLogCalibration` (`close.go:763`) requires
  `Milestone == "" && Actual != ""`, and the function returns early at `:808`/`:816` with
  no brain dir. So no churn output on a milestone close, under `--no-actual`, or in any
  downstream repo without a sibling brain — against a Done-when that says `sdlc close`
  prints it. Split into two consumers: the **ledger row** stays gated by #117's
  calibration-integrity rule; the **operator line** is unconditional in `applyClose`.
  Wiring gap closed too: `closeFlags` has no `PlansDir` field, so one is added and
  `close.go:959`/`:992`'s inline `envOr` calls switch to it (`ARCH-DRY`).
- **[Important]** `finding.cue` modeled disposition *names* but left the
  closes-vs-leaves-open semantics in a Go switch — the exact posture Task 1 argues
  against. Concretely: adding `deferred` would pass `ApplyChecked` and hit no case in
  `OpenFindings`, wedging the finding open forever. Dispositions now partition
  `closing`/`open` like severities do, with `Closes()` derived from it; `Decide`'s
  post-cap rule gets a modeled `hardBlocking` source instead of a `!= "Critical"` literal.
  Conformance test extended to pin both partitions and the hardBlocking ⊆ blocking subset.
- **[Important]** No fuzz on `ParseFindingsBlock` or the `Render`/`ParseSidecar`
  round-trip — both consume unbounded LLM text, and the planned coverage was 10
  hand-written well-formed cases. That is verbatim the pathology this issue indicts, and
  C1's own headline example describes this function. Added `FuzzParseFindingsBlock` and
  `FuzzRenderParseRoundTrip`, seeded with the structural hazards (`\n---\n` inside a
  detail, nested fences, block scalars, non-`<prefix>-<int>` ids). **There is no `func Fuzz`
  anywhere in this repo today** — verified — so this is also the reference instance of the
  mechanical guard the new prompt will demand of every future plan.
- **[Minor ×4]** archiving attribution corrected (`archivePlanArtifacts` is called from
  `push.go:610`/`merge.go:652`, never from `close`; test is `TestArchivePlanArtifacts`);
  the `gatestate` test helpers got a Step 0 with an explicit ID-sequence contract, since
  three tasks share them; Risk 5 added for live `findings`-fence conformance (Task 14's
  real-agent replay *is* the de-facto check for `claude` — stated, because it makes Task 14
  load-bearing twice); and the plan now states that its reproduced code blocks are
  illustrative, with the entity tables/tests/invariants authoritative.

Judge's clean list, worth keeping: `ARCH-PURE` split (clock + filesystem isolated in
`planreview.go`, `changeCodeGateOrder()` extracted so B1 is testable without spawning the
command), `ARCH-MOCK` (no new external dependency; `churnForWindow` drives real `git` in a
disposable repo), and the column-append discipline in Task 13 against `ParseRows`'
positional indexing.

## Log


- 2026-07-29: closed M1 — Live 2-round convergence on #187 own plan: round 1 blocked (1 Critical + 1 Important + 2 Minor, ledger written, PQ-1..PQ-4 assigned); round 2 passed ("no open blocking findings after 2 round(s); 1 advisory recorded") with a not-addressed Minor correctly NOT blocking. Review quality held: round 1 caught two real seam defects in M2 plan (runChangeCode os.Exit vs harness return value; gitx.Capture cannot report the failure Task 13 warns on). go test ./... green; 2 fuzz targets (13.8M execs) — FuzzRenderParseRoundTrip caught a yaml/v3 emitter bug that would have made ledgers unreadable. Mutation-verified: swapping the gate-order literal fails TestGateOrderPlanBeforeEstimate; reverting helptext/issue.md fails TestEstimateTimingConsistency.; review verdict: FIX-THEN-SHIP
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
