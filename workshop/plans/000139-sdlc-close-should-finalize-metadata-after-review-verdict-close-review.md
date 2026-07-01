# Boundary Review — ariadne#139 (whole-issue close)

| field | value |
|-------|-------|
| issue | 139 — sdlc close should finalize metadata after review verdict |
| repo | ariadne |
| issue file | workshop/issues/000139-sdlc-close-should-finalize-metadata-after-review-verdict.md |
| boundary | whole-issue close |
| milestone | — |
| window | 539fff5cedaf6d84e99b59f79971449b7e635ed1..HEAD |
| command | sdlc close --issue 139 |
| reviewer | claude |
| timestamp | 2026-06-30T17:57:32-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have a complete picture. The core mechanism is sound and well-tested; I found two Important edge-path issues (a contradictory operator message and a `--force` contract regression) plus a test gap. Verdict below.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** #139 delivers its purpose cleanly: `runClose` is split at its real compute/write seam into a read-only `computeClose` (all gates, composes text, writes nothing) and `applyClose` (writes + ledger + deferred success messages), and both full-issue close and milestone-close reorder to **compute → review-the-un-mutated-tree → finalize-on-verdict** through the shared `reviewThenFinalize`. The finalize policy derives from `vocab.Verdict()` (no hardcoded enumeration — honors #147), REWORK/unknown correctly leave the issue `working` with no stale bookkeeping, and the rerun-without-`--no-reclose-guard` behavior is pinned by a real test. Tests pass (`ok cmd/sdlc 14.1s`, `pkg/vocab` ok) — the earlier "failures" I hit were an environmental stale `.git/sdlc.lock` from an interrupted real `sdlc close --issue 139` (dead PID 65499), which I cleared; not a code defect. What keeps this from SHIP: on the review-can't-run / dispatch-error path the code now prints "close succeeded" and then halts without finalizing (a direct contradiction), and `milestone-close --force` no longer guarantees finalization — both are non-blocking but should be fixed.

### 1. Strengths (confirmed-good ground)
- **The compute/apply seam is the right cut** (`close.go:330` `computeClose`, `close.go:625` `applyClose`). Deferring the `cok("flipped → done")` messages into `closeResult.appliedMsgs` (`close.go:322`, emitted at `close.go:636`) so they print *only* post-write is exactly right — a REWORK never claims a write that didn't happen. This directly closed a plan-quality finding.
- **`closeVerdictOutcome` derives from the model** (`close.go:845`) via `vocab.Verdict().IsFinalizing/IsBlocking` rather than a `SHIP/FIX-THEN-SHIP/REWORK` switch — ARCH-DRY / ARCH-PURPOSE pass; a new verdict token in `verdict.cue` flows here automatically instead of silently falling to `closeHalt`.
- **One finalize policy, shared** (`reviewThenFinalize`, `close.go:861`) across close and milestone-close, keyed on `f.Milestone` for the trailer/annotation — no duplicated dispatch logic (ARCH-DRY).
- **Genuinely behavioral tests**: `TestRunCloseWithReview_RerunAfterREWORK` (`close_finalize_test.go:96`) pins Done-when-2 with an exactly-one-`closed —`-line assertion; the REWORK/unknown/milestone cases drive the full flow against a real temp git repo (`closeRepo`) + injected `judge.Run` stub — INTEGRATION tested via an injected fake, not mock-reasserting-impl (ARCH-PURE).
- **Atlas updated in-window** (`atlas/workflow/sdlc-binary.md`) with an accurate two-phase description — atlas gate satisfied.

### 2. Critical findings
None. No state-corruption, crash, or silent error-swallowing where the source raised.

### 3. Important findings

**I1 — Contradictory "close succeeded" message on the not-run / dispatch-error path.** `dispatchBoundaryReview` still prints `"close succeeded; re-run judge manually if needed"` at `milestoneclose.go:484` and `:493` when the review is skipped (no window) or `judge.Dispatch` errors. Those lines were written under the pre-#139 premise "the close has already happened; the review is a follow-on." That premise is now false: `dispatchBoundaryReview`'s only non-test caller is `reviewThenFinalize`, which maps the returned `VerdictNotRun` → `closeHalt` and prints `"close NOT finalized; issue left at status: working"` (`close.go:884`) with a non-zero exit and **no write**. So the operator sees `close succeeded` immediately followed by `close NOT finalized` — telling them the close is done when it isn't. Not a state bug (the issue is correctly left `working`), but genuinely misleading.
  *Fix sketch:* drop/replace the two `"close succeeded…"` `cwarn`s (they no longer hold for either caller). Something like `"boundary review could not run — close NOT finalized; re-run once the judge is available, or pass --no-judge to skip"`.

**I2 — `milestone-close --force` no longer guarantees finalization.** The judge-skip branch keys on the raw flag: `case f.NoJudge:` (`milestoneclose.go:149`), not `f.Force || f.NoJudge`. The `--force` help says "bypass ALL close gates (≡ every --no-* flag)" (`milestoneclose.go:87`), and `--no-judge` is one of those flags — so `--force` is documented to imply it. Full-issue close honors this via `f.skip("judge")` (which includes `Force`, `close.go:794`); milestone-close doesn't. Pre-#139 this was latent (the write happened first, so the still-dispatched review couldn't block an already-`done` close). Post-#139 it has teeth: a `--force` milestone-close still dispatches the review and will **halt/rework** on a non-finalizing verdict, leaving the milestone unwritten — violating the `--force` "bypass ALL" contract exactly when an operator is relying on it (emergencies).
  *Fix sketch:* `case f.Force || f.NoJudge:` in `runMilestoneClose`'s switch, mirroring full-issue close's `f.skip("judge")`.

**I3 — The dispatch-error → halt path is untested.** Done-when 5 claims tests cover "judge failure," and it's marked `[x]`, but only the parse-*unknown* halt is exercised (`TestRunCloseWithReview_Unknown_Halts` stubs a body with no `VERDICT:` line). No test stubs `judge.Run` returning an error — every stub in the suite returns `(output, nil)`. The policy is pinned at the pure level (`TestCloseVerdictOutcome` maps `VerdictNotRun→closeHalt`), but the *wired* dispatch-error flow isn't — and an integration test there would have caught I1's contradictory message.
  *Fix sketch:* add a close + a milestone case with `judge.Run = func(...) ([]byte, error) { return nil, errors.New("boom") }`, asserting: non-nil error, issue still `working`, no `closed` line, and no `"close succeeded"` in stderr.

### 4. Minor findings
- **Dry-run lost its per-change preview.** Pre-#139 dry-run printed `flipped → done` / `ticked` / `appended` (inline during compute) before `DRY=1`; now those live in `applyClose.appliedMsgs`, which dry-run never calls, so `printCloseDryRun` (`close.go:613`) emits only `Would update: <path>`. Arguably cleaner (it no longer says "flipped" when nothing flipped), but it's a behavior change vs the plan's "behavior-preserving / dry-run tests stay green" claim. If the preview was valued, print `r.appliedMsgs` in `printCloseDryRun`.
- **`kind`/`verb` string derivation** is computed in `reviewThenFinalize` (`close.go:863`) but the milestone `--no-judge`/`--dry-run` branches hardcode `"milestone-close"` separately (`milestoneclose.go:153,161`). Trivial; leave it.

### 5. Test coverage notes
- Well-covered: SHIP finalize (full-flow + rerun), REWORK not-finalize (close + milestone), unknown-verdict halt, `--no-judge` finalize, `closeVerdictOutcome` table, and the #69 "milestone close doesn't double-dispatch" invariant all pass.
- Gaps: (a) I3 — no wired dispatch-error/skipped-window → halt test; (b) FIX-THEN-SHIP finalize is only covered at the pure-policy level, not a full flow (acceptable — it shares the `closeFinalize` branch with SHIP); (c) no test for `milestone-close --force` (would pin I2).

### 6. Architectural notes for upcoming work
- **Halt-on-dispatch-error in judge-less environments.** This is intended and documented (plan D2, atlas), but note the downstream consequence: a base-layer consumer without a configured agent CLI now gets `judge.Dispatch` error → not-run → **halt** on every `sdlc close`, and must pass `--no-judge` to finalize — a real friction change from the prior graceful degradation. The non-goals section already flags hardening `ParseVerdict`/the prompt to make `unknown` rare as a separate issue; that same follow-up should decide whether a genuine *dispatch* failure (agent absent/timeout) deserves the same halt as an *ambiguous verdict*, or a distinct "review-unavailable" degradation. Worth a tracked issue.

### 7. Plan revision recommendations
- No table/code contradictions — every Core-concepts entity exists at its stated path with the stated status (`closeResult` `close.go:314`, `computeClose` `:330`, `closeOutcome`/`closeVerdictOutcome` `:831`/`:845`, `applyClose` `:625`, `printCloseDryRun` `:613`, `reviewThenFinalize` `:861`, `runClose` wrapper `:652`, `runMilestoneClose` two-phase `milestoneclose.go:104`). No plan revision required for accuracy.
- Optional (if I2/I3 are accepted): add a `## Revisions` note that milestone-close's judge-skip must honor `--force` (align with full-issue `f.skip("judge")`), and that the Done-when "judge failure" coverage means a *dispatch-error* integration test, not only the parse-unknown case.
