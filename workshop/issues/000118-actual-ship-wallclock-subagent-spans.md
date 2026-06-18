---
id: 000118
status: done
deps: []
github_issue:
created: 2026-06-18
updated: 2026-06-18
estimate_hours: 0.96
started: 2026-06-18T10:30:48-07:00
actual_hours: 0.72
---

# Measure ship wall-clock: fill subagent spans in active-time so actual matches the estimate's unit

## Problem

`sdlc actual` (active-time) measures the operator's **interaction time** — it sums
gaps between events in the operator's main transcript, capping each at 15 min to
delete idle. But the estimate (`estimate_logic-v2` build-effort) targets the thing
we actually want: **wall-clock for one engineer + AI to ship**. These are different
units, and the gap is the bulk of the apparent "v2 overshoots ~3.5×" finding —
*a measurement artifact, not an estimation error.*

The unit mismatch has a precise cause: **subagent execution spans are wrongly
truncated as idle.** When the main agent dispatches a Task subagent, the subagent
runs in its own transcript (outside the dirs active-time reads), so the operator's
main session shows `dispatch → … → return` as one big gap — capped at 15 min. A
2-hour subagent build registers as ~15 min. But that span is **active project work**
(the AI is shipping), and it is **bounded by observable events** (the Task tool-use
and its tool-result are both in the operator's transcript, timestamped). So we can
distinguish "subagent grinding" from "operator at lunch" — the blanket 15-min cap
conflates them.

Reframe (operator, 2026-06-18): the estimate exists for **launch predictability**
(when can we ship), not for tracking operator-attention. So `actual` should measure
**ship wall-clock**, and operator-attention is *not* the right unit — it belongs one
level up, as the parallelism/throughput limit (one human ≈ 2 concurrent sessions),
not as the per-issue measure. This supersedes #112's operator-attention-model
direction (parked → effectively moot).

## Spec

Make active-time measure **active ship wall-clock**: idle removed, but
**subagent-execution spans kept**.

- **Detect Task spans.** In the main transcript, a synchronous Task call is an
  assistant `tool_use` block (subagent dispatch) followed by its `tool_result`
  block (return), both timestamped. The gap between them is active subagent work.
  The event parser (`internal/activetime`, `walkSessionEvents`) currently keys off
  text/mentions; extend it to recognize Task dispatch/return so the span is
  identifiable.
- **Fill, don't cap, those spans.** `activeMinutes` caps every gap at 15 min;
  instead, a gap **bounded by a Task dispatch→return counts in full** (it's work,
  not idle). Non-Task gaps (operator stepped away, nothing running) keep the 15-min
  truncation.
- **Ship-time, not summed effort.** Filling naturally yields wall-clock: parallel
  subagents produce overlapping spans whose union is the wall-clock, not the summed
  effort. That's the right quantity (when can we ship), and it sidesteps the
  parallelism-discount question (see non-goals).
- **Window** is already claim→close (#116's `started:` anchor); unchanged.
- **Reconcile the framing.** The "operator-attention" language #117 put in the
  contract + docs is now wrong. Update `helptext/{issue,change-code,estimate,close}.md`,
  `brain/.../velocity/calibration-findings.md`, and the calibration-ledger header to
  say estimate + actual are both **ship wall-clock** (the loop measures
  estimate-vs-actual ship-time).

## Done when

- active-time fills Task-bounded gaps (full duration) and still truncates non-Task
  idle at 15 min. Unit-tested on crafted event streams (a Task span survives; a bare
  idle gap is capped).
- `sdlc actual` on a delegated issue (e.g. one implemented via subagents) measures a
  ship wall-clock close to the real elapsed working time, not ~15 min.
- Docs/framing reconciled from "operator-attention" to "ship wall-clock"
  (helptext + calibration-findings + ledger header).
- Decide + document what to do with the 4 pre-fix ledger rows (re-measure #116/#117
  under the new engine where transcripts survive, or mark them legacy-method).
- Atlas note on the active-time semantics change.

## Scope / non-goals

- **Within-session parallelism discount is OUT.** When the main agent fans out
  parallel subagents, ship wall-clock compresses below v2's sequential sum. Real and
  already happening — but agentic parallelism behavior changes too fast to model
  durably right now (operator, 2026-06-18). Filling spans gives the *wall-clock*
  (overlap union), which is correct; any *estimate-side* parallelism discount is
  deferred until the behavior stabilizes.
- **Background/async subagents** (main session stays busy, no gap) need no special
  handling — the truncation problem only arises with synchronous blocking spans.
- Not a new estimate model. v2 stays; this fixes the *ruler* so v2 can be judged
  fairly. Plan after this lands: observe estimate↔actual for 1–2 weeks, then
  recalibrate v2 → v2.1 from the corrected ledger.

## Estimate

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: smaller-go-module   design=0.1 impl=0.5
item: atlas-docs          design=0.1 impl=0.2
design-buffer: 0.30
total: 0.96
```

- `smaller-go-module` — extend `internal/activetime` (TaskSpan + union gap-math +
  Agent-span parsing + threading through Compute/buildSegments). Design low: the
  durable plan pre-resolves the decisions; impl upper-range (3 files, ~10 new
  tests, parity-sensitive).
- `atlas-docs` — helptext UNIT-NOTE reconcile + calibration-findings banner +
  ledger header + atlas note, across ariadne + brain.

## Plan

Durable plan: `workshop/plans/000118-actual-ship-wallclock-subagent-spans-plan.md`
(authored via superpowers-writing-plans). Single-pass; the mandatory fresh-eyes
review runs at `sdlc close` (branch-point→HEAD).

- [x] Task 1 — `TaskSpan` + union gap-math core (`activeMinutesUnion`/`unionMinutes`/`clampSpans`); `activeMinutes` → wrapper (parity-preserving)
- [x] Task 2 — parse `Agent` dispatch→return spans in `walkSessionEvents`/`loadEvents`
- [x] Task 3 — thread spans through `Compute` + `buildSegments` (fill, don't cap; clamp per segment; extend final boundary)
- [x] Task 4 — reconcile framing: helptext UNIT NOTE + calibration-findings banner + ledger header + atlas note; document 4-pre-fix-rows decision
- [x] Task 5 — verify: full suite + vet + build + real-data smoke

## Log

### 2026-06-18
- 2026-06-18: closed — go test ./... green; go vet clean; go build OK. New tests: union fill/parity/overlap, clampSpans, Agent-span parse, dangling-dispatch, window-clamp, fills-span, commit-inside-span (plan-review blocker), tail-past-last-event. 4 pre-fix ledger rows = DEMONSTRATED no-op: pre/post-#118 binaries identical over identical windows (#116 0.54==0.54, #117 1.31==1.31, #118 0.58==0.58). Docs reconciled: estimate.md UNIT NOTE + atlas/ledger-landscape + brain calibration-findings banner/Revisions + ledger header.; review verdict: FIX-THEN-SHIP
Created from the launch-predictability reframe: the estimate targets ship
wall-clock, but `actual` measured operator-interaction time — so the ~3.5× was
largely a wrong-ruler artifact. Supersedes #112's direction (operator-attention
model; #112 stays parked in history). Sequence: fix the ruler (this) → observe 1–2
weeks → recalibrate v2 → v2.1.

### 2026-06-18 — built + verified

**Premise correction (surfaced during exploration, operator-confirmed "build it,
correct the rationale"):**
- The subagent-dispatch tool is named **`Agent`**, not `Task`, in real transcripts
  (the `Task*` tools are the todo list). Detection keys off `tool_use.name=="Agent"`.
- **Magnitude claim was wrong.** Census of all 33 historical `Agent` spans → every
  one is **under** the 15-min cap. So span-fill changes no current row; the ~3.5×
  supervised overshoot is most consistent with **stale v2 numbers**, not a
  wrong-ruler artifact. #118 is unit-correctness + forward-looking (bites once a
  delegated run exceeds the cap). Calibration banner corrected accordingly.

**Implementation (ARCH-DRY + ARCH-PURE):**
- `activeMinutes` → thin wrapper over a single union-of-intervals core
  `activeMinutesUnion` (capped gaps ∪ full task spans). Parity exact: with no spans
  union == old sum-of-capped-gaps (`TestActiveMinutes` + `TestAttributionGolden`
  unchanged). `ARCH-DRY` — one gap-math implementation.
- Span detection in the IO seam (`walkSessionEvents`), structural (tool_use/
  tool_result), independent of `includeAssistant`; union/clamp math pure,
  unit-tested with real temp files (`ARCH-PURE`).
- `buildSegments`: clamp spans per segment, fill in full, extend final boundary for
  a trailing span, and **don't skip a span-bearing zero-event segment** — fixes the
  **commit-inside-span** blocker a fresh-eyes plan review caught (a subagent that
  commits mid-run leaves the post-commit tail in an event-less segment; the return
  is a dropped `tool_result`). Span boundaries deliberately NOT added → commit-
  anchored attribution preserved. `TestBuildSegmentsCommitInsideSpan` guards it.
- `loadEvents` clamps spans to `[since,until]` (all measured time ∈ window).

**Verification (evidence):**
- `go test ./...` green, `go vet ./cmd/sdlc/...` clean, `go build ./cmd/sdlc` OK.
- New tests: union fill/parity/overlap, clampSpans, Agent-span parse, dangling
  dispatch, window-clamp, fills-span, commit-inside-span, tail-past-last-event.
- **4-pre-fix-rows decision = KEEP AS-IS (demonstrated, not asserted):** built the
  pre-#118 binary from the base commit and compared old vs new over identical
  windows — #116 0.54h==0.54h, #117 1.31h==1.31h, #118 0.58h==0.58h. The engine
  change is a provable no-op on sub-cap windows. (Ledger's 0.41/0.93 differ only
  because closed-issue windows now extend to HEAD — a window artifact, not #118.)
- Docs reconciled: estimate.md UNIT NOTE, atlas/ledger-landscape, brain
  calibration-findings banner + `## Revisions`, ledger header.

### 2026-06-18 — close-review fixes (FIX-THEN-SHIP → resolved)
Boundary review (window 59366da..HEAD) = FIX-THEN-SHIP, no Critical, 2 Important — both fixed:
1. Framing reconciliation missed the contract-bearing estimate-quality **judge prompt**
   (`cmd/sdlc/internal/judge/prompts.go`) — reframed operator-attention → ship wall-clock
   (same unit as build-effort; residual heavy-fan-out gap = parallelism/overlap discount,
   #118 non-goal). (`propagatebase.go`/`vocab.go` mentions are out of scope.)
2. Delivered the committed end-to-end test `TestComputeFillsAgentSpanEndToEnd` (real git +
   30-min Agent span through Compute → TotalActive≈35, PerIssue[#8]≈35; un-filled ≈5).
Minor: per-file `pending` (dispatch/return across a compaction boundary unpaired) — caveat
added to atlas/ledger-landscape. Re-verified: `go test ./...` green, vet clean, build OK.
