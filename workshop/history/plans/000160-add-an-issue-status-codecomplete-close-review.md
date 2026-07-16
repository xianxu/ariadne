# Boundary Review — ariadne#160 (whole-issue close)

| field | value |
|-------|-------|
| issue | 160 — add an issue status: codecomplete |
| repo | ariadne |
| issue file | workshop/issues/000160-add-an-issue-status-codecomplete.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6ed0d8a6336dda64eeba1835f669d61ffd6f6042..HEAD |
| command | sdlc close --issue 160 |
| reviewer | claude |
| timestamp | 2026-07-02T18:04:38-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have a complete picture. Let me write up the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** Issue #160 delivers the two-gate publish model cleanly and completely. `codecomplete` is added to the single CUE source and I verified `pkg/vocab/issue.json` is **byte-identical** to what `issue.cue` regenerates (single-source / ARCH-PURPOSE holds); the lifecycle edges, the extended `actual_hours!` guard, set-status's dual `→ codecomplete`/`→ done` refusals, `sdlc close`'s flip to `codecomplete`, and the deterministic `runPublishGate` invariant + `codecomplete → done` flip on merge/push are all present and correct. The build is clean, the targeted logic/integration tests pass (publishgate, transition guards, close-finalize, merge/push flip+archive), and the two Important findings from the M3 milestone review — the `touchedIssuesNotDone` carve-out test and the push-side flip+archive integration test — were both remediated (verified in the diff). Docs (atlas issue-lifecycle / vocabulary / pre-merge-checks / ci-merge-check) and the curated help are comprehensive; README (a 9-line intro) has no status/flow surface to update. Nothing blocks SHIP. What keeps it from a clean SHIP is purely cosmetic: the cobra-registered flag/`Short` usage strings still say "judges" (they render in `--help` and now contradict the curated body that says "publish gate, no LLM").

### 1. Strengths
- **True single source, verified.** `go run ./cmd/vocabulary export --noun issue` is byte-identical to committed `pkg/vocab/issue.json` — the enum is not hand-restated anywhere; consumers derive (`IsTerminal`/`IsActive`, `RenderLifecycleHelp`). ARCH-PURPOSE/ARCH-DRY pass.
- **The anchor invariant rests on an enforced property, not a claim.** `codecompleteAnchorCommit` (publishgate.go:40) is trustworthy *because* `set-status` Guard 1b (setstatus.go:261) refuses `→ codecomplete`, making `close` the sole writer. The multi-issue "latest anchor" reasoning (min-ahead = newest close = whole-branch review) is sound (publishgate.go:104-121), avoiding false per-issue drift refusals.
- **Fail-closed on git error.** `revCount` returns `ok=false` on a git error so the gate refuses rather than treating an error as "no drift" (publishgate.go:113-116, 173-180) — the right default for a safety gate.
- **Dry-run doesn't mutate.** Both merge (early-return at merge.go:389-402, before step 10.5) and push (early-return at push.go:168-171, before step 6.5) return before the flip — the flip is never reached on `--dry-run`. Verified by tracing both flows.
- **Shadow-sweep is complete.** `state.go` `detectDrift`/`closeOffFinding` handle `codecomplete` gracefully (not terminal → no "should be archived"; not open/working → not a close-off candidate) — verified at state.go:316-346. The `--no-judge` emergency path still flips (TestRunMerge_CodecompleteFlippedToDoneAndArchived uses `NoJudge:true`).

### 2. Critical findings
None.

### 3. Important findings
None. (The M3 boundary review's two Important findings — the untested `touchedIssuesNotDone` codecomplete carve-out and the missing push-side flip+archive integration test — are both delivered: `push_test.go` `TestTouchedIssuesNotDone` now seeds `000004-cc.md` and asserts it's absent from not-done; `TestPushPublishSequence_CodecompleteFlippedThenArchived` covers the push flip→archive. Verified in the diff.)

### 4. Minor findings
- **Stale "judges" in `--help`.** The cobra-registered flag usage strings still say "judges", and they render *in addition to* the curated helptext, producing a self-contradiction in `sdlc merge --help` / `sdlc push --help`:
  - `merge.go:109` / `push.go:68` — `--no-judge` usage: `"skip pre-merge judges (emergency-only)"` → the publish gate. (Renders at merge --help line 134 / push line 91, right below the correct curated line.)
  - `merge.go:108` — `--yes` usage: `"…merge fail-fasts before the judges when stdin is not a terminal"` → "before the publish gate".
  - `push.go:59` — `Short`: `"Ship from main: auto-commit, run pre-merge judges, push, archive done issues"` → drop "run pre-merge judges" (now the deterministic publish gate).
- **Dir-wide flip scope (already documented).** `publishCodecompleteIssues` (publishgate.go:143) globs *all* codecomplete issues in the dir, while `runPublishGate` verifies the invariant only for the window-scoped `mergedCodecompleteIssues` set — so a codecomplete issue outside the merge window would be flipped+archived without its invariant checked. Safe under the healthy-single-main assumption; acknowledged in the plan's `## Revisions`. Keep as-is or scope the flip to the verified set.
- **Issue `## Done when` boxes are all left `- [ ]`** despite every item being delivered (I checked each against the code). Cosmetic — `close`'s plan-check only scans `## Plan` (all three `Mx` ticked), so it won't refuse; tick them for an accurate record.

### 5. Test coverage notes
- Pure/invariant logic is well covered against **real temp-git repos** (publishgate_test.go: clean/drift/multi-issue/re-close/no-op), the correct choice over mocks per §5. Transition guards, close-finalize (SHIP→codecomplete; REWORK/error/unknown→no-flip), lessons-emit gating, and both flip+archive paths are pinned.
- Acceptable gaps: no single end-to-end `close→merge` test (the pieces cover it); the README docs-gate is pinned only as prompt text + golden (can't unit-test LLM judgment). Neither is a boundary blocker.
- Full `go test ./cmd/sdlc/...` couldn't run here — the parent `sdlc close` that dispatched this review holds `.git/sdlc.lock`, so lock-acquiring mutating tests block (a harness artifact, not a defect). I ran the non-locking logic + integration subset (green) and the build (clean).

### 6. Architectural notes for upcoming work
- **ARCH-DRY — pass, with a convergence to watch.** Four issue-file scanners now share a shape (`mergedCodecompleteIssues`, `touchedIssuesNotDone`, `publishCodecompleteIssues`, `archiveDoneIssues` — glob/diff → parse → status-filter). The comments even say "mirrors … (ARCH-DRY)" but mirror rather than reuse. Extract a shared `changedIssueFiles(baseRef, issuesDir) []{Path,Status}` before a fifth appears (noted by M3 too).
- **ARCH-PURE — pass.** Vocab is pure data; predicates/guards are pure string→error, tested without IO; publishgate helpers are thin git-glue tested against a real process-level git repo (not function mocks). `publishCodecompleteIssues` reads the clock directly (`time.Now()`), acceptable for a boundary write helper.
- **ARCH-PURPOSE — pass, one posture shift to record.** The purpose (split "I think I'm done" from "reviewed AND published", with the invariant *enforced*) is delivered and the invariant is enforced (runPublishGate + Guard 1b), not just documented. Note for the future: review-enforcement now depends on the agent actually running `sdlc close` — merge/push with no codecomplete issue in the window is a deterministic no-op that forces no review; the not-done warn (`--yes`-skippable), the #124 conformance gate, and server-side CI are the merge-side backstops. This is by design (Spec Q1–Q5), but it's a genuine change from "merge always ran plan/specs" worth keeping in mind.

### 7. Plan revision recommendations
None required. The M3 review already prompted the `## Revisions` entry in `000160-codecomplete-status-plan.md` reconciling the Core-concepts table with the code (the new `publishCodecompleteIssues` helper; the dir-wide vs window-scoped flip; `runPublishGate` does the check only) — it's present and accurate. The `#142` plan carries a correct PENDING-SEQUENCING banner + Revisions documenting its subsumption. Only non-plan follow-up: tick the issue's `## Done when` boxes (Minor above).
