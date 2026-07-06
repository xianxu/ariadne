# Boundary Review — ariadne#146 (whole-issue close)

| field | value |
|-------|-------|
| issue | 146 — sdlc milestone review bypass command too ambiguous |
| repo | ariadne |
| issue file | workshop/issues/000146-sdlc-milestone-review-bypass-command-too-ambiguous.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4d2dd20809fe6885855b66ae7740cf67d790ddb1..HEAD |
| command | sdlc close --issue 146 |
| reviewer | claude |
| timestamp | 2026-07-05T17:12:54-07:00 |
| verdict | unknown |

## Review



---

## Re-review — 2026-07-05T17:43:30-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 146 — sdlc milestone review bypass command too ambiguous |
| repo | ariadne |
| issue file | workshop/issues/000146-sdlc-milestone-review-bypass-command-too-ambiguous.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4d2dd20809fe6885855b66ae7740cf67d790ddb1..HEAD |
| command | sdlc close --issue 146 |
| reviewer | claude |
| timestamp | 2026-07-05T17:43:30-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Vet is clean (exit 0, no diagnostics). I now have everything I need. Let me compose the review.

## Summary of verification performed

- **Build**: `go build ./...` clean (exit 0). **Vet**: `go vet ./cmd/sdlc/...` clean (exit 0).
- **#146 tests**: `TestCloseVerb`, `TestRerunCmd`, `TestClose_MilestoneRefusesWithRedirect`, `TestRunCloseWithReview_MilestoneRefuses`, milestone-close registration tests — **all PASS in 0.4s**.
- **Full-suite "FAIL" diagnosed**: isolated `go test ./cmd/sdlc/` timed out at 4m — traced to `TestSetStatusAlias_BothPathsMutate` blocking in `repolock.Acquire` on `.git/sdlc.lock`, held by **PID 66727 = the live `sdlc close --issue 146` transaction that dispatched this very review**. 30-min lock WaitTimeout > 4-min test deadline. Every other package is `ok`. This is a review-time lock artifact, not a diff defect — #146 touches none of repolock/set-status/issue_test. Left the lock alone (can't be certain the close isn't still live).

---

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change cleanly does what #146 set out to do: `sdlc close` hides `--milestone` and refuses it with a teaching redirect to `sdlc milestone-close` (reviewed) / `--no-judge` (labeled skip), the two gate re-run hints are repointed via a single-sourced `closeVerb`/pure `rerunCmd` pair, the milestone-close in-process mechanics are untouched, and the derived entry files (`CLAUDE.md`/`AGENTS.md`/`GEMINI.md`) were regenerated flag-free. Build/vet clean; the diff's own tests pass; the sole suite "failure" is a mutating-command test blocking on the repo lock held by the close transaction running this review — not attributable to the diff. What keeps this from a straight SHIP is one shipped user-facing help-text line that still advertises the removed capability (the exact drift class the issue exists to eliminate) — cheap to fix, non-blocking at the gate.

### 1. Strengths
- **Refusal placement is correct and load-bearing** (`close.go:791`): the `f.Milestone != ""` guard sits at the *top* of `runCloseWithReview`, before `computeClose`, so it fires ahead of any issue-existence check (a nonexistent `--issue 999` still gets the redirect) and milestone-close — which never routes through `runCloseWithReview` — is provably unaffected.
- **Returnable error, not `die()`** (`close.go:791-797`): the right call so the command-level test can assert the refusal instead of `os.Exit` killing the test binary; the reasoning is captured in the comment.
- **ARCH-DRY consolidation is genuine**: `closeVerb` single-sources the mode→verb mapping across `reviewThenFinalize` + both hints, and `rerunCmd` DRYs the two hint tails (`close.go:870-893`); the previously-inline selection at `reviewThenFinalize` is now a call.
- **ARCH-PURE**: `closeVerb`/`rerunCmd` are pure string→string, unit-tested directly without IO (`close_test.go:18-45`); the refusal is a thin IO guard.
- **Obsolete guard test repurposed, not deleted** (`closereview_test.go:254`): `TestRunCloseWithReview_MilestoneRefuses` now asserts the refusal + zero dispatch, preserving coverage of the invariant.
- **ARCH-PURPOSE shadow-sweep of derived consumers done**: verified `CLAUDE.md`/`AGENTS.md`/`GEMINI.md:50` match the regenerated `AGENTS.base.md` and carry no `close --milestone`.

### 2. Critical findings
None.

### 3. Important findings

**`cmd/sdlc/helptext/milestone-close.md:81` — RELATED line still advertises the removed capability.**
```
RELATED
  sdlc close             same close logic without milestone-review auto-dispatch
```
Post-#146 this is false on both counts: `sdlc close` refuses `--milestone` (it can no longer close a milestone at all), and a whole-issue `close` *does* auto-dispatch the boundary review. This is user-facing help (`sdlc milestone-close --help`), it sits in the very file the diff edited (line 8 was changed), and it points the reader straight at the "close a milestone without the review" path the issue removed — i.e. the exact drift the Done-when says to eliminate. The plan's Task 4/shadow-sweep enumerated `helptext/milestone-close.md` but only named line ~8; the RELATED cross-ref at the bottom slipped through.
*Fix sketch:* reword to describe what `sdlc close` actually is now, e.g. `sdlc close   whole-issue close (auto-dispatches the end-of-issue boundary review)`, or drop the "without milestone-review auto-dispatch" clause.

### 4. Minor findings
- **`close.go:655-656` and `789-790` mis-describe the call graph.** They state milestone-close performs "its in-process mechanical step (runClose with the milestone set)" / "runClose(Milestone=…), which milestone-close calls in-process." Since #139, milestone-close calls `computeClose` directly (`milestoneclose.go:135`), *not* `runClose`; `milestoneclose.go:112`'s "delegate the mechanical close to runClose" comment is likewise stale. With the `close --milestone` short-circuit now removed, **`runClose` has zero production callers** — only `close_test.go:390` and `close_ledger_test.go:121,134`. The safety conclusion ("milestone-close unaffected") is right, but the stated mechanism is wrong; given #146's whole thesis is call-graph precision, worth correcting the comments to say `computeClose` (and noting `runClose` is now a test-only shared wrapper, or inlining it later).
- **`closereview_test.go:249`** header comment ("a milestone close routed through runClose (as milestone-close does)") repeats the same inaccuracy; the test itself is correct.

### 5. Test coverage notes
- Good: the pure builders are pinned directly (`TestCloseVerb`, `TestRerunCmd`) and the refusal is covered at both the function level (`TestRunCloseWithReview_MilestoneRefuses`) and the cobra command level (`TestClose_MilestoneRefusesWithRedirect`, which also asserts the flag is hidden + still parses).
- Gap (acceptable): no test asserts that `explainActual`/`explainVerified` actually thread `milestone` into `rerunCmd` — they shell out to `computeActual`, which is why the plan deliberately extracted the pure builder and tested that instead. The wiring is one line each; the ARCH-PURE split is the right trade-off, so this isn't a finding, just noted.

### 6. Architectural notes for upcoming work
- **ARCH-DRY: PASS** — the mode→verb mapping is now single-sourced; the only residual DRY smell is the now-test-only `runClose` duplicating the `computeClose→applyClose` sequence that both real close paths inline. A future cleanup could delete `runClose` and have the ledger/close tests drive `computeClose`+`applyClose` (or `runCloseWithReview`) directly.
- **ARCH-PURE: PASS** — pure core (`closeVerb`, `rerunCmd`) / thin IO shell (cobra refusal, explainers) split is clean and directly tested.
- **ARCH-PURPOSE: PASS with the §3 caveat** — the three purpose components (flag hidden+refused, hints repointed, derived entry files regenerated) are all delivered and verified. The one remaining hand-maintained restatement of the removed capability (`milestone-close.md:81`) is the deferred consumer the shadow-sweep should have caught — fixing it closes ARCH-PURPOSE fully.

### 7. Plan revision recommendations
Add a `## Revisions` entry to `workshop/plans/000146-remove-close-milestone-bypass-plan.md` noting:
- Task 4's doc shadow-sweep must include `helptext/milestone-close.md`'s **RELATED** section (line ~81), not just line ~8 — it still cross-references the removed `close --milestone` behavior.
- The verification grep asserted in Task 5 (`grep -c -- '--milestone Mx' … → 0`) is imprecise: the corrected `sdlc milestone-close --issue N --milestone Mx` text legitimately contains the substring `--milestone Mx`, so that count is expected to be non-zero. The real invariant is "no `close --milestone`" (and no `[--milestone Mx]` on the close example), which does hold — record the corrected check so the plan doesn't claim a grep result the shipped docs don't satisfy.

(Housekeeping, outside the diff: `.git/sdlc.lock` is currently held by the `sdlc close --issue 146` PID 66727, now shown as gone — if the close transaction has truly exited without releasing it, that lock is stale and should be removed before the next mutating `sdlc` command, per the lock's own "remove only if no transaction is running" guidance. I left it in place since I'm the review dispatched by that transaction and can't rule out it's still live.)

---

## Re-review — 2026-07-05T18:49:58-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 146 — sdlc milestone review bypass command too ambiguous |
| repo | ariadne |
| issue file | workshop/issues/000146-sdlc-milestone-review-bypass-command-too-ambiguous.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4d2dd20809fe6885855b66ae7740cf67d790ddb1..HEAD |
| command | sdlc close --issue 146 |
| reviewer | claude |
| timestamp | 2026-07-05T18:49:58-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have everything I need. `Makefile.workflow` is a **base-layer file** (`construct/base.manifest:112`, symlinked into downstream repos), so the `make close-issue MILESTONE=Mx` breakage propagates to every dependent repo. Let me write up the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The core change cleanly delivers #146's purpose on the primary interface: `sdlc close` hides `--milestone` and refuses it with a teaching redirect to `sdlc milestone-close` (reviewed) / `--no-judge` (labeled skip); the two gate re-run hints are repointed via a single-sourced `closeVerb` + pure `rerunCmd`; the milestone-close mechanics (`computeClose` → `reviewThenFinalize`, verified at `milestoneclose.go:136,169`) are untouched; and the derived entry files + helptext (including the RELATED line the prior review flagged) are swept clean. Build/vet/tests all green. What holds it back from a straight SHIP is one **base-layer consumer the shadow-sweep missed**: the `make close-issue MILESTONE=Mx` target still shells to the now-refused `bin/sdlc close --milestone`, so a documented, downstream-propagating interface now hard-fails — the exact "enumerate every consumer of the removed capability" step ARCH-PURPOSE calls for. It fails-safe (prints the redirect, no silent wrong close) and the primary `sdlc milestone-close` path is unaffected, so it's non-blocking at the gate but should be fixed before merge.

### 1. Strengths

- **Refusal placement is load-bearing and correct** (`close.go:794`): the `f.Milestone != ""` guard sits at the *top* of `runCloseWithReview`, before `computeClose`, so it fires ahead of any issue-existence check and — critically — milestone-close never reaches this function (it drives `computeClose`/`reviewThenFinalize` directly, `milestoneclose.go:136,169`), so the mechanical milestone close is provably unaffected.
- **Returnable error, not `die()`** (`close.go:795`): the right call so the command-level test can assert the refusal instead of `os.Exit` killing the test binary; the reasoning is captured in the comment.
- **ARCH-DRY consolidation is genuine**: `closeVerb` single-sources the mode→verb mapping across `reviewThenFinalize` (`close.go:909`) + both hints via `rerunCmd`; the previously-inline `verb := "sdlc close"; if … { verb = … }` at `reviewThenFinalize` is now a call.
- **ARCH-PURE**: `closeVerb`/`rerunCmd` are pure string→string, unit-tested directly without IO (`close_test.go:18-45`); the refusal is a thin cobra-path guard.
- **Obsolete guard test repurposed, not deleted** (`closereview_test.go:254`): `TestRunCloseWithReview_MilestoneRefuses` now asserts the refusal + zero dispatch, preserving the invariant's coverage.
- **Prior FIX-THEN-SHIP findings actually applied**: `helptext/milestone-close.md:81` RELATED line and the `close.go`/`milestoneclose.go` call-graph comments were corrected to say `computeClose` — I verified them at HEAD, not just in the commit message.

### 2. Critical findings

None.

### 3. Important findings

**`Makefile.workflow:104-106` — `make close-issue MILESTONE=Mx` still invokes the refused `close --milestone` path (base-layer; propagates downstream).**
```make
close-issue:
	@if [ -x bin/sdlc ]; then \
	    bin/sdlc close \
	      $${ISSUE:+--issue "$$ISSUE"} \
	      $${MILESTONE:+--milestone "$$MILESTONE"} \   # ← now refuses
```
When `MILESTONE` is set, this target shells to `bin/sdlc close --milestone Mx`, which #146 made refuse (`close.go:794`). So a documented interface (`Makefile.workflow:55` advertises `make close-issue ISSUE=N [MILESTONE=Mx] …`; `:79` shows a worked MILESTONE example) now hard-fails. `Makefile.workflow` is `symlink`ed in `construct/base.manifest:112`, so this breakage propagates to every downstream ariadne-styled repo. It's also self-contradicted by the diff's own atlas edit: `atlas/workflow/sdlc-binary.md:40` (changed in this window) claims `make close-issue MILESTONE=Mx` **is** "THE milestone-close path … auto-dispatched boundary review" — but the target routes to the no-longer-valid `close --milestone`, not `milestone-close`. This is the ARCH-PURPOSE shadow-sweep miss: the sweep found the doc consumers (entry files) but not the *executable* consumer.
*Mitigation (why Important, not Critical):* it fails-safe — the operator sees the redirect telling them to run `sdlc milestone-close` — and the constitution-directed primary path (`sdlc milestone-close`) works; whole-issue `make close-issue` (no MILESTONE) is unaffected.
*Fix sketch:* branch the target on `MILESTONE` — call `bin/sdlc milestone-close …` when it's set, else `bin/sdlc close …`. That both restores the interface and makes atlas row 40's claim true (the make path would then actually dispatch the review). (Separately note: the Python fallback `scripts/close-issue.py` does its own no-review milestone close, so on a repo without `bin/sdlc` built, `make close-issue MILESTONE=Mx` still silently skips the review — that's the deprecated pre-binary path slated for M8 removal, out of #146's scope, but worth a one-line acknowledgment in the Log.)

### 4. Minor findings

- **`atlas/workflow/sdlc-binary.md:290` and `:346`** still describe milestone-close as `forwards the same flags into its delegated runClose` / `inherits it (wraps runClose)`. Since #139 milestone-close calls `computeClose` directly, and #146 made `runClose` test-only (zero production callers, per `close.go:654-660`), both lines are stale — and they contradict the same file's own `:477-480` ("compute → review → finalize"). Pre-existing (outside the diff window), but this is the exact "sweep at SECTION granularity, not file granularity" lesson #146 itself just added (`lessons.md`, commit fc84936): the diff edited line 40 of this file but didn't grep the rest of it for `runClose`. Cheap to correct while here.

### 5. Test coverage notes

- Well covered: the pure builders are pinned directly (`TestCloseVerb`, `TestRerunCmd`), and the refusal is asserted at both the function level (`TestRunCloseWithReview_MilestoneRefuses`, incl. zero dispatch) and the cobra level (`TestClose_MilestoneRefusesWithRedirect`, which also asserts the flag is hidden yet still parses). All pass (verified: 0.42s).
- Gap that would have caught the Important finding: there is **no test exercising the `make close-issue MILESTONE=Mx` → binary wiring**. A wrapper-level breakage like this is inherently outside Go unit tests; the `close-review` e2e checklist (Task 5) checked the *flag* and *help text* but not that the `make` interface still functions. Not a blocker, but the class of miss argues for a smoke check (`make close-issue MILESTONE=… DRY=1` expecting the milestone-close path) if that interface is meant to survive.
- Acceptable gap: no test asserts `explainActual`/`explainVerified` thread `milestone` into `rerunCmd` (they shell to `computeActual`); the ARCH-PURE extraction of the pure builder + testing that directly is the right trade-off.

### 6. Architectural notes for upcoming work

- **ARCH-DRY: PASS** — mode→verb mapping is now single-sourced. Residual smell: the now-test-only `runClose` duplicates the `computeClose→applyClose` sequence both real paths inline; a future cleanup could delete it and have the ledger/close tests drive `computeClose`+`applyClose` directly.
- **ARCH-PURE: PASS** — pure core (`closeVerb`, `rerunCmd`) / thin IO shell (cobra refusal, explainers) split is clean and directly tested.
- **ARCH-PURPOSE: FLAG** — the shadow-sweep enumerated the doc consumers but not the two *executing* consumers of `close --milestone`: the base-layer `Makefile.workflow` target (Important above) and the legacy `scripts/close-issue.py` fallback. "Every consumer derives from the source" here means every surface that *invokes or advertises* the removed flag; the make target both advertises (`:55`) and invokes (`:106`) it. Fixing the Makefile closes ARCH-PURPOSE fully.

### 7. Plan revision recommendations

Add a `## Revisions` entry to `workshop/plans/000146-remove-close-milestone-bypass-plan.md`:
- **Scope miss — `make close-issue MILESTONE=Mx` consumer.** The "Scope / blast radius" section and Task 4's doc-sweep enumerated the entry files but omitted `Makefile.workflow:104-106`, which shells `MILESTONE` into `bin/sdlc close --milestone` — now a refused path. It's a base-layer file (`construct/base.manifest:112`), so the breakage propagates downstream, and it contradicts the plan's own atlas edit (row 40 = "the milestone-close path"). The shadow-sweep must include command-invoking wrappers (Makefiles/scripts), not only prose docs; the fix is to route the target to `bin/sdlc milestone-close` when `MILESTONE` is set.
- **Atlas call-graph residue.** Note that `atlas/workflow/sdlc-binary.md:290,346` still say milestone-close "delegates/wraps `runClose`" — the same #139 call-graph inaccuracy the plan's earlier Revisions corrected in code comments, left uncorrected in the atlas (a per-file, not per-section, sweep).
