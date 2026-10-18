# Boundary Review — ariadne#442 (whole-issue close)

| field | value |
|-------|-------|
| issue | 442 — tl-worker: a review may overlap a queued recorded suite without hub approval |
| repo | ariadne |
| issue file | workshop/issues/000442-tl-worker-a-review-may-overlap-a-queued-recorded-suite-without-hub-approval.md |
| boundary | whole-issue close |
| milestone | — |
| window | 48b3b97b7cc45ab813214baddf6150549037225b..04363208a954f28e818a9b640b3c5f790aadcf42 |
| command | sdlc close --issue 442 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T08:32:11-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change adds the overlap rule beside #432's draft-PR rule in PERMISSIONS of `cmd/tl-worker/help.md` (lines 34–36). It states both conditions from the Spec and says no hub approval is needed. It adds two `own`-scope rows to `rules.tsv`, and `go test ./cmd/tl-worker/ ./cmd/tl-lead/` passes at head. Only one condition is phrase-guarded, though. The Spec's condition (1), "a recorded suite pass on the final head still precedes landing", has no `rules.tsv` row. Deleting that clause, which is the safety half of the rule, would leave every test green. The fix is one row, so it shouldn't block.

**1. Strengths**
- The rule sits exactly where the Plan says, right after the draft-PR rule (`help.md:32-33`). Its wording matches that neighbour ("needs no hub approval" and "without hub approval").
- Both guarded phrases fit on one line, so `strings.Contains` matches them without any whitespace handling (`help.md:34`, `:36`).
- The rows use `own` scope, so the tl-lead side test also checks that tl-lead doesn't repeat the rule. That test passes.
- The change stays in scope: two files and 5 lines, with no other edits.

**2. Critical findings**
None.

**3. Important findings**
- `cmd/tl-worker/testdata/rules.tsv:9-10`: Spec condition (1) is not guarded. The rows guard the overlap permission (`review-suite-overlap`) and condition (2) (`review-suite-requeue`). Nothing guards "a recorded suite pass on the final head still precedes landing" (`help.md:35`). I checked whether any other row covers it: `early-draft-gates` matches "gates still precede landing" on line 33, not line 35. A test run without this clause would stay green.
  - **Fix:** add a row such as `review-suite-pass-first	own	recorded suite pass on the final head still precedes landing`.
  - **Every instance of this gap in the window:** the permission is guarded, condition (2) is guarded, and condition (1) is the only one left. No other new clause lacks a guard.

**4. Minor findings**
- `review-suite-overlap` stops at "without hub" because "approval" wraps onto the next line. If the "no hub approval needed" wording were rewrapped or dropped, this guard might not catch it. It works today.

**5. Test coverage notes**
- `TestWorkerHelpCarriesEveryRule` runs both `--help` and the no-argument form, so the guarded phrases are checked in both modes.
- The tl-lead tests for `own`-scope rows also pass.
- Only condition (1) has no regression check, as described above.

**6. Architectural notes**
- **ARCH-DRY: pass.** The rule appears only in the tl-worker help, and the `own` scope keeps tl-lead from repeating it.
- **ARCH-PURE: pass.** The change is text plus a table row; no logic or IO was added.
- **ARCH-PURPOSE: flag.** This is the Important finding above. Done-when asks for the rule to be phrase-guarded, and the purpose is to remove ambiguity safely. Leaving the safety condition unguarded delivers only part of that.
- **Docs gate:** the change adds no new command, flag or architectural surface. The help text is the documentation this issue asked for, so no atlas or README update is needed.

**7. Plan revision recommendations**
- The Plan line says "two `rules.tsv` rows". It should say three, adding a row for condition (1). Alternatively, add a `## Revisions` entry noting the extra row.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: rule-clause-unguarded
    title: |
      Spec condition (1), the suite pass on the final head before landing, has no rules.tsv phrase guard
    detail: |
      help.md:35 "a recorded suite pass on the final head still precedes landing" matches no rules.tsv row (early-draft-gates matches line 33 only), so deleting the safety clause keeps TestWorkerHelpCarriesEveryRule green. Enumerated family in window: overlap permission (guarded), condition 2 requeue (guarded), condition 1 (unguarded, the only instance). Fix: add an own-scope row, e.g. review-suite-pass-first with phrase "recorded suite pass on the final head still precedes landing", and update the Plan line from two rows to three.
  - id: new
    severity: Minor
    kind: defect
    family: rule-clause-unguarded
    title: |
      review-suite-overlap phrase stops at "without hub" and omits "approval"
    detail: |
      The line wrap at help.md:34-35 means the guard does not pin "approval"; the no-approval wording could drift without the test noticing.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"51e80b50-9f09-4592-ae58-151c936f1fdb","repository":"github.com/xianxu/ariadne","issue_id":"000442","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/7f32db816d9c4151","reviewed_head":"04363208a954f28e818a9b640b3c5f790aadcf42","slot":"ariadne:2","requested_at":"2026-10-18T15:31:40.086073Z"},"phase":"finished","pid":69400,"started_at":"2026-10-18T15:31:40.173447Z","finished_at":"2026-10-18T15:32:11.57313Z","session_id":"c3331ca0-35d4-4994-a8a2-2df246f59de3","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T08:33:27-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 442 — tl-worker: a review may overlap a queued recorded suite without hub approval |
| repo | ariadne |
| issue file | workshop/issues/000442-tl-worker-a-review-may-overlap-a-queued-recorded-suite-without-hub-approval.md |
| boundary | whole-issue close |
| milestone | — |
| window | 48b3b97b7cc45ab813214baddf6150549037225b..f4dee0dcde01fb7d590b79c65ce9c2c641176f2c |
| command | sdlc close --issue 442 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T08:33:27-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Both prior findings are fixed, and I confirmed each fix by reverting it. The help text at `cmd/tl-worker/help.md:34-37` has been reflowed so each clause sits whole on one line: the overlap permission, the no-approval wording, condition (1) and condition (2). `cmd/tl-worker/testdata/rules.tsv` now has four own-scope rows, one per clause. `go test ./cmd/tl-worker/ ./cmd/tl-lead/` passes at head. Removing the condition (1) line makes `TestWorkerHelpCarriesEveryRule` fail in both the `--help` and no-argument modes. Changing "approval" to "sign-off" also makes it fail. The Plan line in the issue now says four rows, and the Log records round 1.

**1. Strengths**
- The reflow at `help.md:34-37` fixes the whole class of problem, not just the one case: every clause is a whole line, so every guard phrase is matched in full.
- There is exactly one `rules.tsv` row per clause (overlap, approval, pass-first, requeue), and each is own-scope, so tl-lead must not repeat the rule.
- The change stays in scope: two code files and 8 lines in total.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings:** none.

**5. Test coverage notes**
- `TestWorkerHelpCarriesEveryRule` checks every clause. I ran two reverts to confirm this: deleting the condition (1) line and rewording "approval" each make the test fail. The tl-lead own-scope test passes.

**6. Architectural notes**
- **ARCH-DRY: pass.** The rule appears only in the tl-worker help.
- **ARCH-PURE: pass.** The change is text and table rows; no logic was added.
- **ARCH-PURPOSE: pass.** Every clause of the rule, including the safety conditions, is now guarded by a phrase.
- **Docs gate:** no new command, flag or architectural surface, so no atlas or README update is needed.

**7. Plan revision recommendations:** none. The Plan already says four rows.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      rules.tsv adds review-suite-pass-first; deleting help.md condition (1) line turns TestWorkerHelpCarriesEveryRule red (verified by revert).
  - id: BR-2
    disposition: addressed
    note: |
      Reflowed so "without hub approval, under two conditions" is one line, guarded by review-suite-approval; rewording "approval" turns the test red (verified).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7240fa34-2277-4039-9e09-288e1a36218a","repository":"github.com/xianxu/ariadne","issue_id":"000442","recipe":"small-diff-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/a8a1a9fa7d5e6ca6","reviewed_head":"f4dee0dcde01fb7d590b79c65ce9c2c641176f2c","slot":"ariadne:2","requested_at":"2026-10-18T15:33:06.615053Z"},"phase":"finished","pid":75690,"started_at":"2026-10-18T15:33:06.706667Z","finished_at":"2026-10-18T15:33:27.239772Z","session_id":"d9ef123e-c3c6-4399-a70e-99fd907882ca","outcome":"returned"}] -->
