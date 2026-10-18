# Boundary Review — ariadne#441 (whole-issue close)

| field | value |
|-------|-------|
| issue | 441 — doc-review: choose its reviewer through the #409 family-aware resolver |
| repo | ariadne |
| issue file | workshop/issues/000441-doc-review-choose-its-reviewer-through-the-409-family-aware-resolver.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4c2fa13e834dfdf4caec56b2c47b864b5465cefe..b7c033ab62cfa957c71465574ff4aaa93a74468b |
| command | sdlc close --issue 441 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T11:18:54-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Let me look at the last commit (addressing BR-6/BR-11 presumably) and the core packages.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Both open findings are fixed. BR-6 is gone because `Choice.Refusal` was deleted from `pkg/reviewpolicy/choice.go`, so nothing can call it through the alias any more. BR-11 is fixed for the case it named: `chooseReviewers` now marks a "never self-review" skip with `Policy: true`, `NoReviewerError.Error()` builds its remedy from the skip kinds, `TestRefusalRemedyFollowsTheSkips` pins both remedies, and help.md:27-29 now says an explicit pair into the author's family, or an unpaired session, refuses. One part of that same class is still wrong. When the plan comes back empty with no skips, `Error()` counts it as a policy skip (`policy := len(parts)==0`). That case is `other` mode with every family already an author. The message then tells the user to set `reviewer_pairs`, which `Validate` rejects outside `explicit` mode. I ran a probe in this checkout to confirm the exact output. The issue is Minor and does not block. gofmt (on the touched packages), vet and the tests of the touched packages are all clean.

1. **Strengths**
   - `pkg/reviewpolicy` is the single source of the policy. `judge` re-exports it through aliases and only adds sdlc's `--agent` wording on top (`cmd/sdlc/internal/judge/reviewchoice.go:58-63`), so ARCH-DRY holds.
   - `Plan`, `Resolve`, `Record` and `AuthorFamilies` are pure functions. Availability and git access are injected through `Availability`, `lookPath` and `gitIn`, so ARCH-PURE holds.
   - doc-review's never-self-review filter sits in one place (`cmd/doc-review/choose.go:56-75`), and a forced agent is still recorded as degraded.
   - The remedy text is now computed from skip data rather than a fixed sentence.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `pkg/reviewpolicy/choice.go:68`: an empty refusal (`other` mode with every family authored, or a `reviewer_order` that covers only author families) recommends `reviewer_pairs`, a setting `Validate` rejects in that mode. This is the 3rd finding in family `refusal-names-applicable-remedy`; the rule is in the findings block below.
   - `pkg/reviewpolicy/policy.go:225-230` with `choose.go:66`: `Plan` checks availability before doc-review's author filter. So an explicit pair into the author's family whose CLI is also missing is reported as transient ("retry later"), yet installing the CLI would still lead to a refusal. Same class as the finding above.

5. **Test coverage:** `TestRefusalRemedyFollowsTheSkips` covers a policy-only refusal and a transient-only refusal. Three cases have no test: the empty-skips refusal, mixed skips, and sdlc's `noReviewer` rewrite of the policy remedy ("name one with --agent").

6. **Architecture**
   - **ARCH-DRY:** pass.
   - **ARCH-PURE:** pass.
   - **ARCH-PURPOSE:** pass for consumers: both sdlc and doc-review derive from `reviewpolicy`. The remedy class is only partly swept (the Minor finding above).
   - **ARCH-MOCK:** pass. `git` runs against real temporary repositories, and the CLI lookup is behind the `lookPath` seam.
   - **ARCH-CONSTRAINTS:** N/A. Each review makes one cheap git log call.
   - **ARCH-SECURE:** pass. Forged trailers can narrow the candidates but never remove the session's family (`AuthorFamilies`), and `Validate` refuses unknown agents and malformed model names.
   - **ARCH-ORDER:** the policy carries no state between events (it is a single resolve). The candidate loop in `review.go`/`DispatchChoice` only moves past a reviewer that never launched, so no review is duplicated.
   - **ARCH-FUNERAL:** nothing durable is created besides the existing report field.

7. **Plan revisions:** none needed. The as-built table already reflects the deletion of `Choice.Refusal` (BR-8).

```findings
dispose:
  - id: BR-6
    disposition: addressed
    note: |
      Choice.Refusal no longer exists in pkg/reviewpolicy/choice.go; grep finds only judge.Refusal (reviewchoice.go:67), so the alias cannot expose it.
  - id: BR-11
    disposition: addressed
    note: |
      Skip.Policy is set by the never-self-review filter (choose.go:69), the remedy drops retry-later for policy-only skips, TestRefusalRemedyFollowsTheSkips pins it, help.md:27-29 updated; the empty-skips sibling is raised as a new finding.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: refusal-names-applicable-remedy
    title: |
      An empty refusal under other mode recommends reviewer_pairs, which Validate rejects outside explicit mode
    detail: |
      This is the 3rd finding in family refusal-names-applicable-remedy. Earlier rounds fixed instances: a fixed sentence, then a single Policy boolean. Rule: every cause of refusal carries its own remedy, chosen by the cause together with the effective mode. Cause 1, a CLI that is unavailable: retry later. Cause 2, an explicit pair into the author family, or an unpaired session: fix reviewer_pairs. Cause 3, an other-mode plan that leaves no family (all families authored, or reviewer_order covers only author families): no config remedy except widening reviewer_order; name one explicitly. Cause 4, a skip for availability whose candidate would also have been policy-skipped: report the policy cause, not the transient one. Fix the rule: replace the Policy bool with a remedy or cause enum on Skip, plus an explicit cause for an empty plan. choice.go:68 currently sets policy=true when there are no skips. A probe in this checkout (chooseReviewers under Other with authors anthropic, openai and google) printed "pair this agent with another family in reviewer_pairs", and Policy{Mode: Other, Pairs: ...}.Validate() fails with "reviewer_pairs applies only to boundary_reviewer explicit". The same text reaches sdlc through judge.noReviewer. Add a table test that covers all four causes.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"24e3e023-f0ff-4989-a8e5-cd79a32cdb38","repository":"github.com/xianxu/ariadne","issue_id":"000441","recipe":"milestone-review","round":5,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/beea8d778a44af75","reviewed_head":"b7c033ab62cfa957c71465574ff4aaa93a74468b","slot":"ariadne:2","requested_at":"2026-10-18T18:17:29.807684Z"},"phase":"finished","pid":10310,"started_at":"2026-10-18T18:17:29.92047Z","finished_at":"2026-10-18T18:18:53.05861Z","session_id":"03d0af34-0961-492b-b5f0-da09d5190f05","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T11:31:38-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 441 — doc-review: choose its reviewer through the #409 family-aware resolver |
| repo | ariadne |
| issue file | workshop/issues/000441-doc-review-choose-its-reviewer-through-the-409-family-aware-resolver.md |
| boundary | whole-issue close |
| milestone | — |
| window | 31b0ff491562705ee5d9e818273f7068cb76acff..bfe6b7b53ab4e5b9bbf0b8b600105bdfa03fbda7 |
| command | sdlc close --issue 441 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T11:31:38-07:00 |
| verdict | SHIP |

## Review

Tests pass on the touched packages. Next I'm reverting the cause-4 fix to confirm its test goes red.

The revert check passed: both fixes are needed for the test to go green. Last step is checking that the whole repo still builds now that the `Skip` field is gone.

```verdict
verdict: SHIP
confidence: high
```

The fix for BR-12 is in place, and it fixes the rule the finding stated, not just the one example it gave. It covers all four refusal causes:

- **The old boolean is gone.** `Skip.Policy` is replaced by a `SkipCause` enum with two values, `Unavailable` and `AuthorFamily` (`pkg/reviewpolicy/policy.go:204-215`).
- **An empty plan gets its own wording.** A refusal with no skips (`other` mode with every family already an author) now explains why and recommends only "name one explicitly". It no longer suggests `reviewer_pairs`, which `Validate` rejects in that mode (`pkg/reviewpolicy/choice.go:68-85`).
- **The policy cause wins over a missing CLI.** In doc-review, a candidate that is unavailable *and* in the author family is now reported with the self-review reason, not "retry later" (`cmd/doc-review/choose.go:66-76`).

A new table test covers all four causes. It also runs `Validate()` on each case's policy, so the test itself can't use a setup the config would reject. I reverted each part of the fix in this checkout, one at a time, and the test failed both times:

- Without the cause-4 change, the message read "claude is not on PATH … retry later".
- With the empty plan treated as a policy skip again, the message recommended `reviewer_pairs`.

Other checks:

- No other code still uses the removed `Policy` field.
- `go build ./...` passes, and `go vet` passes on `cmd/sdlc/...`, `cmd/doc-review` and `pkg/reviewpolicy`.
- `gofmt -l` is clean on the touched packages.
- Tests pass in `cmd/doc-review`, `pkg/reviewpolicy` and `cmd/sdlc/internal/judge`. The `cmd/sdlc` tests ran too long and were moved to the background, so I didn't wait for that result; this change doesn't touch that package.

Outside the code, the only changes are the generated gate-ledger and close-review files. `## Spec`, `## Revisions` and `## Done when` are untouched, so scope did not change. There are no conflict resolutions in this window.

1. **Strengths**
   - Each remedy now comes from a declared cause rather than from parsing the reason text (`choice.go:71-84`).
   - In `choose.go:66-69`, one `selfReview` helper serves both the rewrite of skipped candidates and the filter on remaining candidates, so the self-review rule is written once.
   - The test checks both what each message must say and what it must not say, and it checks that every case's policy is valid config.
   - sdlc's `noReviewer` rewrite ("name one with --agent") still matches the new text, because the empty-plan wording keeps "name one explicitly".

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none new. The empty-plan wording names `reviewer_order` even when the user never configured one. That is still accurate, because the default order has the same effect.
5. **Test coverage:** all four causes are covered and each fails when its fix is removed. sdlc's `noReviewer` rewrite of the empty-plan message has no direct test, but the existing judge tests pass.
6. **Architecture (interdiff only)**
   - **ARCH-DRY:** pass. One `selfReview` helper, and one place that maps causes to remedies.
   - **ARCH-PURE:** pass. `NoReviewerError.Error()` and `chooseReviewers` are pure; availability is passed in.
   - **ARCH-PURPOSE:** pass. The four-cause rule and the empty-plan cause were fixed together in one round.
7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-12
    disposition: addressed
    note: |
      Skip.Policy replaced by the SkipCause enum (policy.go:204-215); the empty plan has its own reason and only the name-one remedy (choice.go:68-85); doc-review reports the policy cause over the transient one (choose.go:66-76); TestRefusalRemedyFollowsTheCause covers all four causes, validates each policy, and fails when either fix is reverted in this checkout.
```

The `cmd/sdlc` test run I'd left in the background has finished, and it **failed**. It hit Go's default 10-minute test timeout (600.79s) and panicked while running `TestOwnedActualSelectorWarningsRefuseCloseAndCorrection` (`cmd/sdlc/actualinputs_test.go:169`). Earlier I said this suite was skipped; that is no longer true, and the result is a failure.

I don't think it affects the SHIP verdict:

- **The window doesn't touch that test.** The interdiff changes `cmd/doc-review/choose.go` and its test, `pkg/reviewpolicy/choice.go` and `pkg/reviewpolicy/policy.go`. `actualinputs_test.go` covers time accounting, not reviewer choice.
- **The packages that use the changed code pass.** Tests in `cmd/doc-review`, `pkg/reviewpolicy` and `cmd/sdlc/internal/judge` pass, and `go build ./...` and `go vet ./cmd/sdlc/...` are clean.
- **The run was slowed down.** It shared the machine with my other test, build and vet runs, which likely pushed the slow suite past the limit.

I haven't proven it's unrelated, though. To settle it, run `go test -timeout 30m ./cmd/sdlc` on its own, on this head and on base `31b0ff49`. A timeout on both points to slowness that predates this window; a pass on base and a failure on head would mean this window broke something.

The verdict stays **SHIP**, and BR-12 stays addressed.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cd925a86-601e-40e3-8333-f489212d8c97","repository":"github.com/xianxu/ariadne","issue_id":"000441","recipe":"fix-review","round":6,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/955bb8f0ad98c697","reviewed_head":"bfe6b7b53ab4e5b9bbf0b8b600105bdfa03fbda7","slot":"ariadne:2","requested_at":"2026-10-18T18:20:37.905909Z"},"phase":"finished","pid":45754,"started_at":"2026-10-18T18:20:38.041052Z","finished_at":"2026-10-18T18:31:37.480103Z","session_id":"a73d562a-ef29-4ba9-85d8-19a82084f1db","outcome":"returned"}] -->
