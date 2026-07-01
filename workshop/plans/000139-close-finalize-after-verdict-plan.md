# Close Finalizes After Verdict Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Both `sdlc close` and `sdlc milestone-close` finalize the issue (flip `status: done`, write `actual_hours`, append the "closed" log line, write project/ledger) **only after** the boundary review returns a finalizing verdict — so a REWORK or an *unexpected* verdict leaves the issue at `status: working` with no stale close bookkeeping, a rerun needs no `--no-reclose-guard`, and anything ambiguous halts for a human instead of being papered over.

**Architecture:** `runClose` already computes the entire new issue/project text **in memory** (close.go:311–574) before one consolidated write block (586–603). Split that seam: read-only `computeClose` returns a `closeResult` (no writes); `applyClose(closeResult)` performs the three writes (issue, project, ledger). A shared `reviewThenFinalize` runs the boundary review against the *un-mutated* working tree and dispatches on the verdict via `closeVerdictOutcome` (one of: finalize / rework / halt). **Both** full-issue close (`runCloseWithReview`) and milestone-close (`runMilestoneClose`) reorder to **computeClose → reviewThenFinalize**, so neither writes before the verdict and the two paths share one finalize policy (ARCH-DRY). `runClose` stays a thin wrapper (`computeClose` → dry-run-or-`applyClose`) for the non-review/dry-run callers. Because nothing is written pre-review, the reviewer reads the honest in-progress issue (`status: working`, real Plan/Done-when) — the spec's intent.

**Tech Stack:** Go; unit tests in `cmd/sdlc` via the existing `closeRepo` fixture + `judge.Run` stub (the verdict-outcome matrix + rerun-after-REWORK); close gate docs in the atlas.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `closeResult` | `cmd/sdlc/close.go` | new |
| `computeClose` | `cmd/sdlc/close.go` | new (extracted) |
| `closeOutcome` + `closeVerdictOutcome` | `cmd/sdlc/close.go` | new |

- **closeResult** — value bundling everything `applyClose` needs, computed without writes: `issuePath`, `issueText` (orig), `newIssueText`, `projectEditPath`, `projectEditText`, and the calibration-ledger inputs (`fm`, `body`, `repoName`, `issueStr`, `today`). The seam that lets compute and apply straddle the review.
  - **DRY rationale:** one struct carries the computed close; the eager wrapper, full-issue close, and milestone-close all consume it — none recompute.
- **computeClose** — `computeClose(stderr io.Writer, f *closeFlags) closeResult`: current `runClose` body lines 311–574 (all validation gates + in-memory compose). Keeps `die()` on gate failure (validation fails fast, *before* any review). Writes nothing.
- **closeOutcome + closeVerdictOutcome** — the finalize policy, named once and
  **derived from `vocab.Verdict()`** (the #147 single-source — NOT a hardcoded
  switch, which would reintroduce the verdict enumeration #147 killed):
  ```go
  type closeOutcome int
  const ( closeFinalize closeOutcome = iota; closeRework; closeHalt )
  func closeVerdictOutcome(v judge.Verdict) closeOutcome {
      switch t := string(v); {
      case vocab.Verdict().IsFinalizing(t): return closeFinalize // SHIP, FIX-THEN-SHIP
      case vocab.Verdict().IsBlocking(t):   return closeRework   // REWORK
      default:                              return closeHalt     // unknown, not-run(dispatch error)
      }
  }
  ```
  A pure `TestCloseVerdictOutcome` still enumerates the expected mapping, but the
  policy itself reads the model — so a new verdict token in `verdict.cue` flows here
  automatically (finalizing/blocking) rather than silently falling to `closeHalt`.
  - **`unknown` → `closeHalt`, never finalize.** An `unknown`/`not-run`-from-error verdict means the review ran but produced no clear SHIP/FIX-THEN-SHIP/REWORK — an *unexpected* state (usually a gate/prompt bug). The close must NOT paper over it by finalizing; it halts so a human investigates the root cause. (An explicit `--no-judge` skip is handled *before* dispatch and finalizes — that's an operator decision, not an ambiguous review.)

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `applyClose` | `cmd/sdlc/close.go` | new (extracted) | filesystem writes |
| `printCloseDryRun` | `cmd/sdlc/close.go` | new (extracted) | dry-run print |
| `reviewThenFinalize` | `cmd/sdlc/close.go` | new | review + outcome dispatch |
| `runClose` | `cmd/sdlc/close.go:310` | modified (wrapper) | compute+apply |
| `runCloseWithReview` | `cmd/sdlc/close.go:724` | modified (two-phase) | compute→review→finalize |
| `runMilestoneClose` | `cmd/sdlc/milestoneclose.go:102` | modified (two-phase) | compute→review→finalize |

- **applyClose** — `applyClose(stderr io.Writer, f *closeFlags, r closeResult)`: the current write block (close.go:586–603) — issue/project `os.WriteFile` + `appendCalibrationRow` + the `cok("done — …")`.
- **printCloseDryRun** — the dry-run print (577–584), extracted so all callers show identical "Would update…".
- **reviewThenFinalize** — `reviewThenFinalize(stdout, stderr io.Writer, f *closeFlags, r closeResult, p boundaryReviewParams) error`: dispatch the boundary review, then on `closeVerdictOutcome`:
  - **finalize** → `applyClose` → `emitTrailerBlock` + `annotateLogLineWithVerdict(f.IssuesDir, f.Issue, f.Milestone, verdict)`; return nil.
  - **rework** → `emitTrailerBlock`; warn "REWORK — not finalized; fix findings and re-run `sdlc <kind>`"; return a non-nil REWORK error. No writes.
  - **halt** → `emitTrailerBlock`; warn "verdict %q is UNEXPECTED — not finalized; the review produced no clear verdict (gate/prompt bug?). STOP, investigate the review output (sidecar), consult a human before re-running"; return a non-nil error. No writes.
  Shared verbatim by close (`f.Milestone==""`) and milestone-close (`f.Milestone==Mx`) — the annotation already keys on `f.Milestone`, so one helper serves both.
- **runClose** *(wrapper)* — `r := computeClose(stderr, f); if f.DryRun { printCloseDryRun(stderr, r); return nil }; applyClose(stderr, f, r); return nil`. Used by the non-review/dry-run callers; behavior identical to today.
- **runCloseWithReview** *(two-phase)* — `r := computeClose(...)`; resolve window; dry-run prints + returns; `--no-judge` → `applyClose` + not-run trailer (explicit skip finalizes); else `reviewThenFinalize(...)`.
- **runMilestoneClose** *(two-phase)* — replace its "Step 1 `runClose` (eager) → Step 2–5 review+trailer+annotate" with `computeClose` → (dry-run/no-judge as above) → `reviewThenFinalize`. Same finalize-after-verdict guarantee for milestones.

**Test surface.** `closeRepo` + `judge.Run` stub drive both `runCloseWithReview` and `runMilestoneClose`. Cover the outcome matrix: `SHIP`/`FIX-THEN-SHIP` finalize (issue `done`, exactly one close line, verdict annotated); `REWORK` → not finalized (still `working`, no close line, no `actual_hours`, non-nil error); **`unknown`** (stub a body with no `VERDICT:` line) → not finalized + halt error; judge dispatch-error → halt; `--no-judge` → finalize; **rerun after REWORK** with a `SHIP` stub → clean finalize, one log line, **no `--no-reclose-guard`**. Plus a pure `closeVerdictOutcome` table. Mirror the REWORK + finalize cases for milestone-close.

---

## Design decisions

- **D1 — split at the existing compute/write boundary** (close.go:586). No gate-logic change; low risk (ARCH-DRY).
- **D2 — three outcomes, and `unknown` HALTS.** finalize = {SHIP, FIX-THEN-SHIP, explicit `--no-judge`}; rework = {REWORK}; halt = {unknown, judge dispatch error}. `unknown` is treated as a *bug signal*, not a pass — the close stops and tells the operator to investigate + consult a human, rather than finalizing (the user's "fix the bug, don't dance around it"). Named in `closeVerdictOutcome`, documented, tested.
- **D3 — non-finalize verdicts exit non-zero, write nothing.** REWORK and halt both leave `status: working` (no flip, no `actual_hours`, no "closed" line) → reruns need no `--no-reclose-guard` and produce no stale lines (Done-when 1, 2, 3). They differ only in the message (rework: a clear next action; halt: stop + consult a human).
- **D4 — the reviewer reads the honest state.** `applyClose` deferred ⇒ the working-tree issue stays `working` during review; the reviewer reads the real Spec/Plan/Done-when (status isn't what the boundary review checks).
- **D5 — the final close commit is the operator's, post-gate.** `sdlc close` writes + prints "review with git diff, then commit"; it never commits, and the boundary review fires only inside the close verb — so the close commit can carry `Review-Verdict:` with no recursive review by construction (Done-when 4; Log note).
- **D6 — milestone-close folds in.** Same two-phase via the shared `reviewThenFinalize`; the existing `#69` invariant (milestone-close owns its per-milestone review; full-issue close doesn't double-review a milestone) is preserved — only the *write ordering* changes.

---

## Chunk 1: Extract compute/apply

### Task 1: `computeClose` + `applyClose` + `printCloseDryRun` + `runClose` wrapper

**Files:** Modify `cmd/sdlc/close.go`; Test `cmd/sdlc/close_test.go`

- [ ] **Step 1: Baseline green** — `go test ./cmd/sdlc/ -run 'Close|Frontmatter|insertLog|Calibration|Ledger'`.
- [ ] **Step 2: Extract** `closeResult`; move 311–574 → `computeClose` (return the struct); 577–584 → `printCloseDryRun`; 586–603 → `applyClose`. Rewrite `runClose` as the wrapper.
- [ ] **Step 3: Build + close suite green** — behavior-preserving; every existing close/milestone/dry-run/N-A/frontmatter test stays green unchanged.
- [ ] **Step 4: Commit** — `#139: extract computeClose/applyClose (behavior-preserving)`

## Chunk 2: Two-phase finalize (close + milestone), three outcomes

### Task 2: `closeVerdictOutcome` + `reviewThenFinalize`, rewire `runCloseWithReview`

**Files:** `cmd/sdlc/close.go`; `cmd/sdlc/closereview_test.go`

- [ ] **Step 1: Failing tests** — REWORK → non-nil error + issue still `working` + no `closed —` line + no `actual_hours`; `unknown` (stub output without a `VERDICT:` line) → non-nil halt error + issue still `working`; rerun after REWORK with SHIP → finalize, one log line, no `--no-reclose-guard`. SHIP/FIX-THEN-SHIP/`--no-judge` stay green.
- [ ] **Step 2: Run → fails.**
- [ ] **Step 3: Implement** `closeOutcome` + `closeVerdictOutcome` + `reviewThenFinalize`; rewrite `runCloseWithReview` (milestone short-circuit → `runClose`; else compute → window → dry-run/`--no-judge` branches → `reviewThenFinalize`).
- [ ] **Step 4: Run → PASS**; full `go test ./cmd/sdlc/...`.
- [ ] **Step 5: Commit** — `#139: full-issue close finalizes only after a finalizing verdict (REWORK/unknown halt)`

### Task 3: fold milestone-close in

**Files:** `cmd/sdlc/milestoneclose.go`; `cmd/sdlc/milestoneclose_test.go`

- [ ] **Step 1: Failing test** — a milestone-close with a REWORK stub leaves the milestone NOT closed (status/log unchanged), non-nil error; with SHIP it closes + annotates as today.
- [ ] **Step 2: Implement** — replace `runMilestoneClose`'s eager `runClose` + manual review/trailer/annotate with `computeClose` → dry-run/`--no-judge` → `reviewThenFinalize` (shared helper). Preserve the `#69` invariant + the milestone verdict gate.
- [ ] **Step 3: Run → PASS**; full suite.
- [ ] **Step 4: Commit** — `#139: milestone-close also finalizes only after its verdict`

### Task 4: outcome-matrix coverage

**Files:** `cmd/sdlc/closereview_test.go`, `cmd/sdlc/close_test.go`

- [ ] **Step 1:** Pure `TestCloseVerdictOutcome` (SHIP/FIX-THEN-SHIP→finalize, REWORK→rework, unknown/not-run→halt) + confirm the five close flows + the milestone REWORK/SHIP cases are all asserted.
- [ ] **Step 2: Commit** — `#139: tests — verdict-outcome matrix + rerun-after-REWORK`

## Chunk 3: Docs

### Task 5: atlas

**Files:** `atlas/workflow/sdlc-binary.md`

- [ ] **Step 1:** Document two-phase close/milestone-close: validate+compute → review → finalize-on-verdict; REWORK leaves `working` (rework+rerun, no `--no-reclose-guard`); **`unknown`/judge-error halt for a human** (don't finalize an ambiguous gate). One paragraph.
- [ ] **Step 2: Commit** — `#139: atlas — close finalizes after the boundary verdict`

---

## Done-when mapping

| Issue Done-when | Delivered by |
|---|---|
| REWORK does not leave the issue at `status: done` | Tasks 2–3 (D3) |
| rerun after fixing findings needs no `--no-reclose-guard` | Task 2 (issue stays `working`) |
| exactly one final "closed" log line | Tasks 2–3 (apply writes once, only on finalize) |
| final close commit carries `Review-Verdict:` w/o recursive review | D5 (structural — Log note) |
| tests cover SHIP / FIX-THEN-SHIP / REWORK / judge failure / `--no-judge` (+ unknown halt, + milestone) | Tasks 2–4 |

## Non-goals / follow-up

- **Verdict-parser robustness is NOT in scope here, but is the root cause of the frequent `unknown` this session** (reviewers wrote prose like "the verdict stands: FIX-THEN-SHIP" instead of a `VERDICT:` line). With D2's halt-on-unknown, those closes will halt until that's fixed — a **separate issue** to harden `ParseVerdict`/the prompt so `unknown` becomes genuinely rare. Flagged for the operator to file; #139 deliberately surfaces the bug rather than hiding it.
- No change to the gate set, the trailer format, or the sidecar.

## Revisions

- **2026-06-30 — resumed after #147 (dep satisfied).** `closeVerdictOutcome` now
  **derives from `vocab.Verdict().IsFinalizing/IsBlocking`** (the #147 single-source)
  instead of a hardcoded SHIP/FIX-THEN-SHIP/REWORK switch — so #139 doesn't
  reintroduce the verdict enumeration #147 collapsed (agent-binary-handoff-schema
  target). With #147's structured handoff, `unknown` is now rare/genuine, so
  halt-on-`unknown` (closeHalt) is sound. The rest of the design is unchanged.
