---
type: continuation
slug: sdlc-painpoint-audit-m2
agent: claude
created: 2026-07-14T12:32:49
branch: 000172-sdlc-painpoint-audit
worktree: /Users/xianxu/workspace/ariadne
issues: [000172, 000170]
---

# Continuation: sdlc-painpoint-audit — M2

## NEXT ACTION

Build **#172 M2** on branch `000172-sdlc-painpoint-audit` (M1 closed, FIX-THEN-SHIP,
verdict fixes applied). M2 = the two remaining T1 detectors, over the
`SdlcInvocation` stream M1 already produces. In `cmd/sdlc/internal/processmanual/`,
TDD, **real captured fixtures**:

1. **`detectRefusalRetries`** — pair a `Refusal` `GateEvent` with the next same-verb
   +same-issue invocation in the same transcript (time-ordered). `Resolved` = the
   retry had no refusal. Per-gate refusal signatures are already in `GateCatalog`
   (`RefusalPat`). Carry `Observability` (flag-omitted for merge/push publish-gate,
   which don't name the flag → best-effort attribution).
   - **Fix the M1-review Minor here:** the `no-validate` refusal emits TWO matching
     lines (`validategate.go:82` cwarn + the die-wrapped err) → **dedupe gate events
     per invocation** before pairing, else refusal→retry resolution rates skew.
2. **`detectFiringOrder`** — per-issue, iteration-aware. Order oracle = AGENTS.md's
   workflow DAG `claim(0) ≺ start-plan(1) ≺ change-code(2) ≺ milestone-close(3) ≺
   close(4) ≺ merge(5)`. Legal loops (do NOT flag): `milestone-close→change-code`,
   `start-plan` re-runs, `close→change-code`/`start-plan` after REWORK/reopen
   (`codecomplete→working`, issue.cue:129/132). `issueID` is on `SdlcInvocation`;
   **merge/push carry no `--issue`** → attribute stage-5 from segment context (nearest
   preceding `--issue` invocation) or mark `unattributed`. **Edit/Write/MultiEdit →
   `KindFileEdit` capture (deferred from M1 Task 3) lands here** for the `skill-late`
   arm (`Skill(writing-plans)` after a file edit in the same segment+issue).
3. Fold both into `FrictionReport`/`renderFrictionReport` (+ `--json`); add the
   render/JSON-shape test the M1 review flagged missing. Then
   `sdlc milestone-close --issue 172 --milestone M2`.

After M2 → **M3** (net-new Go codex parser from the atlas spec — fork-skip-40-not-119;
share `transcripts/codex.go`'s first-`session_meta` decode loop shape) → **M4** (T2
triage + T3 coverage-gap read). After #172 → **#170** (audit umbrella; #172 is its
last open dep).

## State of play

Run `sdlc state`. Branch `000172-sdlc-painpoint-audit`, 8 commits ahead of main.
- **#172** — status `working`, M1 **closed** (verdict FIX-THEN-SHIP). Estimate 8.07h;
  M1 actual 5.04h. `go build`/`go vet` clean; processmanual suite green. The
  instrument runs: `go run ./cmd/sdlc process-manual --friction-report`.
- **#170** — open, the audit umbrella (`deps:[#169,#172]`); #172 is its last dep.
- **#173** — done+merged (codex ingest; the atlas codex-format spec M3's Go parser derives from).
- **#169** — done (introspect diminishing returns).

## The instrument (M1, working)

`sdlc process-manual --friction-report` — whole-corpus, cross-repo, command-anchored
per-gate **bypass** measure. Signal is in the tool_result **content-block** (anchored
to `Bash(sdlc <verb>)` calls, `tool_use_id`-linked). `classifyOutputLine` (friction.go)
matches each output line against `GateCatalog` (gatesig.go — 12 gates / 16 sigs / 3 ACK
grammars), reset-gated ACKs + grammar/digit refusals, rejecting warmup/source/cat-n.
Cross-command drift guard (`gates_test.go`). Observability keyed per (command, flag),
intrinsic (`gateObs`).

**M1 headline (clean anchored, real corpus):** no-judge dominant (close 7 + mclose 10 =
17, **full** obs), no-actual 8, no-atlas 5; **no-verified = 0** (design works); bypasses
concentrate in **peers — brain 19, pair 15 — vs ariadne 3**. Raw grep over-counted ~4×
(unlinked echoes anchoring excludes).

## Thread arc & user model

This session: resumed the #173 codex-ingest continuation → finished #173 (codex
dogfood = "no, codex doesn't reopen the taste well"; 3 adapter fixes; merged) →
user steered *"the next step is #172"* → designed #172 through **three plan-review
rounds** (v1 premise inverted → v3.1 build-ready) → *"continue on #172"* ×2 →
change-code + built **all of M1** + closed it. User then: *"close M1, and make a
continuation."*

User model (per AGENTS.md *Model User Intention*): systems-thinker, terse decisive
steers, engages deeply with structured tradeoff tables (answered every
AskUserQuestion crisply). Prizes **agent-neutrality**, **precision over recall**,
**honest calibration** (accepted #173's 5.2× estimate overrun as data), **not
bypassing gates** (chose all-spine-gates scope; the atlas gate honored at M1 rather
than --no-atlas'd), and **review rigor** — the fresh-eyes reviews repeatedly caught
real, empirically-grounded errors and he wants them fixed at root. Drives depth-first
to completion; keeps saying "continue."

## Open questions

None blocking. All #172 scope forks decided (all 12 spine gates; AGENTS.md workflow-DAG
firing-order; whole cross-repo corpus). Run order: #172 → #170.

## Artifact map

- **`workshop/plans/000172-sdlc-painpoint-audit-plan.md`** — the build plan. The
  **signal-catalog master table** is the classifier's ground truth. Read the
  `## Revisions` v1→v3.1 + the **M1 execution-deviations** entry for why each shape is
  what it is (encodes 3 review rounds + the M1 boundary review — don't repeat them).
- **`workshop/plans/000172-*-m1-review.md`** — the M1 boundary-review sidecar (2
  Important fixed; Minors listed for M2).
- **`workshop/issues/000172-sdlc-painpoint-audit.md`** — Spec (T1-T3 Done-when +
  contaminated-grep baseline) + Log (design phase, M1).
- **Code (branch):** `cmd/sdlc/internal/processmanual/` — `gatesig.go` (catalog),
  `friction.go` (`classifyOutputLine`/`SdlcInvocation`/`sdlcInvocations`/`aggregate`/
  `RunFrictionReport`), `gates_test.go` (drift guard, in `package main`), `session.go`
  (`rec` gained `Result`/content-block; `classifyToolUse`), `processmanual.go`
  (`--friction-report` dispatch). Tests: `gatesig_test.go`, `friction_test.go`.
- **`atlas/workflow/sdlc-binary.md`** → "Friction audit" section.
- **`atlas/workflow/introspect.md`** → "Codex transcript format" — M3's Go-parser source.

## Decisions & dead ends

- **Signal is in the tool_result content-block, anchored to `Bash(sdlc <verb>)`** —
  not stderr (v1 dead end), not raw grep (contaminated ~4×). The content-block is more
  complete than `toolUseResult.stdout` (no-judge ACK 105× vs 49×).
- **Drift guard enforced the OTHER way** (test introspects registered flags vs the
  catalog) rather than `AllGates()` sourcing registrations — production untouched, same
  invariant. Recorded as an M1 deviation.
- **Observability is a per-(command,flag) intrinsic property** (`gateObs`,
  HasRefusal-guarded), not per-flag last-write-wins (the M1-review defect).
- **IsHelp is NOT a skip** — a compound `sdlc close --no-judge && x --help` carries a
  real bypass; the classifier's specific patterns reject help-text lines anyway.
- **Deferred to M2:** dedupe gate events per invocation (no-validate double-line);
  Edit/Write capture; go-run-invocation footnote; error-on-zero-transcripts; render/JSON test.

## Lessons learned

- **A working instrument still needs its counts sanity-checked against a raw baseline**
  — M1 first reported 14 no-judge; a diagnostic caught the IsHelp over-skip (compound
  commands), and content-block-vs-stdout fixed the linkage; final 17 vs ~21-linked is
  the honest anchored count (the rest were echoes).
- **The boundary review earns its keep even on a green milestone** — it caught the
  observability mislabel (the design's own honesty pillar) that all local tests missed
  because no test covered the aggregation/render layer. Add render-layer tests in M2.
- **Estimate design mid-weighting (not ×0.2) is holding** — M1 actual 5.04h of the
  8.07h estimate is tracking, unlike #173's 5.2× under-run.
