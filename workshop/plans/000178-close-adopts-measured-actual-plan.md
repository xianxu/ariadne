# close adopts its measured actual — Implementation Plan

> **For agentic workers:** single-pass atomic change (one review boundary — no Mx tags). TDD per task; steps use checkbox syntax.

**Goal:** `sdlc close` / `milestone-close` without `--actual` ADOPTS the active-time-v3 measurement (loud info line with hours + window + peer attribution) instead of refusing with a "→ close with: --actual N" suggestion the agent copies verbatim. The refusal survives ONLY for the unmeasurable statuses. Kills the spine's second-largest refusal volume (#172: 48 no-actual refusals, ~45 verbatim copies) at zero calibration risk — the gate prevents *guessed* hours, and a value sdlc measured cannot be a guess.

**Architecture:** the change lives entirely in `computeClose`'s actual-gate block (`close.go:344-370`), shared by both verbs. Reorder: handle the omit-path FIRST — measure via a new `computeActualForCloseFn` seam (the file's existing `validateChangedIssuesFn`/`runPublishGateFn` pattern); on `actualMeasured` adopt into `f.Actual` and mark adopted; else `explainActual` + exit (unchanged next-action refusal — it already renders the telemetry-gap/no-window/error diagnostics). The pass-path deviation check (#87) skips when adopted — re-checking a value we just measured would re-run the engine to compare it with itself. `--no-actual` keeps its exact meaning (genuinely nothing to measure → `actual_hours: N/A`).

**Observability note (#172 instrument):** the no-actual refusal signature (`Pass --no-actual (or --force) only when measurement is not applicable`) still fires only on the true-refusal path, so the friction report's refusal counts collapse to the unmeasurable cases going forward — the intended measurement effect. The adopt info line must NOT match any `GateCatalog` ACK/refusal pattern (it names no flag; verify against `gatesig.go`).

## Tasks

### Task 1: adopt-path in computeClose (TDD)
**Files:** `cmd/sdlc/close.go`; tests in `cmd/sdlc/close_actualdev_test.go` (or sibling).
- [x] **Step 1: Failing tests** via the `computeActualForCloseFn` override seam:
  (a) omit + `actualMeasured{1.23}` → close proceeds, frontmatter lands `actual_hours: 1.23` (format pinned to `%.2f`, identical to the info line), stderr carries the adopt line (hours + window + peers);
  (b) omit + `actualTelemetryGap` → refusal (exit path) with the existing explainActual shape;
  (c) `--no-actual` → `N/A` sentinel unchanged;
  (d) passed `--actual` still deviation-checked (existing #87 tests keep passing);
  (e) adopted value skips the deviation check (engine invoked exactly once — assert via the seam's call count);
  (f) milestone-mode: adopt line renders under `mode=milestone` too — noting the cumulative-window semantics (computeActual's window is issue-scoped; at M2+ the adopted value is cumulative issue hours, same number today's suggest-path proposes — pre-existing semantics, now stated).
- [x] **Step 2–4:** implement (seam var + reorder + `adopted` flag + info line); `go test ./cmd/sdlc/...` PASS.
- [x] **Step 5: Commit** — `#178: close adopts measured actual (omit-path measures + proceeds)`.

### Task 2: wording sweep (shadow docs)
- [x] `close.go` warmup (`:217` "close suggests a number") + `explainActual` head/tail (now the *unmeasurable* explainer — say so) → "close measures and uses it".
- [x] `helptext/close.md` (3 spots), `AGENTS.md` + `AGENTS.base.md` §5 close bullet ("computes + suggests" → "measures + adopts"), atlas mentions if any describe the two-step.
- [x] Commit — `#178: wording sweep (measure-and-adopt, not compute-then-ask)`.

### Task 3: verify + close
- [x] Live verify: in a hermetic repo (or this one), `sdlc close`-shape omit-path run shows the adopt line; unmeasurable fixture still refuses. `go build/vet/test` green.
- [x] gatesig no-collision check: the adopt info line classifies to `none` against every `GateCatalog` pattern (explicit test — the verifiable close evidence, vs the aspirational corpus-collapse claim).
- [x] Close: `sdlc close --issue 178 --verified '<evidence>'` — which itself dogfoods the adopt path.

## ARCH notes
- **ARCH-PURE:** the adopt decision is a reorder of existing pure-ish gate logic; the engine stays behind one seam; tests need no real transcripts.
- **ARCH-DRY:** one measurement call feeds adopt + (skipped) deviation; `resolveActualRoots` reuse; no new verb.
- **Root cause:** deletes the loop instead of teaching agents to live with it (#172 T2 verdict).

## Revisions

### 2026-07-14 — executed (all tasks complete)

Deltas from the task text, all small: (1) `adoptOmittedActual` returns the
measurement so the refusal arm's `explainActual` REUSES it (signature gained a
`res actualResult` param — no double engine run; the old code measured once,
the naive port would have measured twice on refusals). (2) `explainActual`'s
head now states it is the unmeasurable-only explainer; its tail (the gatesig
refusal anchor) is byte-identical. (3) The sweep found a straggler the plan
missed: `helptext/milestone-close.md` — updated with the cumulative-window
note. (4) AGENTS.md is generated (gitignored); the durable edit is
AGENTS.base.md, with the working copy updated in place for session consistency.
(5) Live verify used the omit+no-verified trick: the actual gate adopts
(`using measured actual: 0.47h (window e4e06cbe → HEAD)`), then the verified
gate refuses before mutation — real-engine adoption observed with zero side
effects. Task 1 test (e)'s "call count" pins the single engine run inside the
adopt arm; the pass-path deviation skip is the adjacent `!adopted` guard.
