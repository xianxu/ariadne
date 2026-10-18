# Boundary Review — ariadne#409 (milestone M2)

| field | value |
|-------|-------|
| issue | 409 — Dynamic review policy: choose the reviewer model by author family and change kind (cross-family by default), extending #310 |
| repo | ariadne |
| issue file | workshop/issues/000409-dynamic-review-policy-choose-the-reviewer-model-by-author-family-and-change-kind-cross-family-by-default-extending-310.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | bc15307cf4a96f84853f113cd3c2c6351e9a1d54..b616aee0de69c94aa54dcbb4832c7e498e6a0ea3 |
| command | sdlc milestone-close --issue 409 --milestone M2 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-18T06:32:31-07:00 |
| verdict | REWORK |

## Review

Diff read; now checking remaining dispatch sites, the internal retry path, and change-kind wiring.

Confirmed one ordering bug via a disposable probe test (removed). Now checking the empty-candidate dry-run path and the sidecar consumers.

Accounting is agent-agnostic; fine. Running the changed packages' tests now.

The broad `cmd/sdlc` run hit Go's 10-minute test timeout (suite size, not a specific failure). Running the new tests directly.

```verdict
verdict: REWORK
confidence: high
```

The M2 wiring is mostly sound, but one rule from the plan is broken, and I confirmed it with a failing test. Every review site now picks its reviewer through `resolveReviewChoice` and `judge.DispatchChoice`. A reviewer that is missing before launch, fails to launch, or hits a quota error is skipped, and the skip is printed. When every candidate is skipped, the review refuses and names each one. The `Review-Reviewer:` trailer round-trips through `git interpret-trailers`. The new tests pass, and `go vet` is clean. The broken rule is "a reviewer that began never hands over". `Dispatch`'s internal no-verdict retry can return an `UnavailableError` on its second attempt, and `DispatchChoice` then switches to another family's reviewer. That happens after the first reviewer has already run a review. I reproduced it with a probe test in the disposable checkout and then deleted the test: the calls were `codex, codex, gemini` and gemini was used. This is the case the plan's state table forbids, and it blocks SHIP.

**1. Strengths**
- `cmd/sdlc/internal/judge/dispatch.go:647-254`: the order inside `classifyRunResult` is right. Cancellation is checked first, then launch failure, then `Unavailable`, then `RunFailure`. "Could not review at all" is one typed error (`UnavailableError`), and `DispatchChoice` switches on that error alone.
- `cmd/sdlc/internal/judge/reviewchoice.go:42-59`: `ResolveReview` is pure. `--agent` and `AGENT_CMD` keep their precedence, and a forced reviewer bypasses the availability check, so the operator can always override.
- `cmd/sdlc/reviewchoice.go:22-40`: the closing session's family is always among the authors. A forged `Agent:` trailer can add a family but can't remove the closer's own, as the plan's ARCH-SECURE revision requires.
- `cmd/sdlc/reviewchoice_test.go`: the boundary tests run the real `dispatchBoundaryReview` through the `judge.Run` seam. They cover the visible skip, the refusal with nothing launched, and the trailer format. `TestReviewReviewerTrailerParsesAsAGitTrailer` checks the Done-when tally against real git trailer parsing.
- `testmain_test.go:166-170`: `reviewAvailability` is made hermetic, so no test result depends on which CLIs the machine has installed.

**2. Critical findings**
- **ARCH-ORDER. A reviewer that began can hand over after its retry** (`internal/judge/dispatch.go:386-395`, `reviewchoice.go:98-104`).
  - What happens:
    1. Attempt 1 runs and ends with no verdict.
    2. Attempt 2 fails with a quota or unreachable error and no stream events.
    3. `dispatchAttempts` returns that `UnavailableError`, either directly (`ValidateResponse != nil`) or through `retryFailure`, which passes `err` through unwrapped.
    4. `DispatchChoice` calls `errors.As`, records a "skip", throws away attempt 1's output, and runs the next family.
  - Reproduced: fake codex returns `stdout "partial review"` with exit 0 on call 1, and the 429 `ERROR:` line with a non-zero exit on call 2. Result: `called=[codex codex gemini] used=gemini err=<nil>`.
  - Fix sketch: on attempt 2, demote any `UnavailableError` to an ordinary did-not-run error (for example `fmt.Errorf("retry after a started review: %s", e.Reason)` without `%w` to `ErrReviewerUnavailable`). Only a first-attempt unavailability should allow a switch. Add that two-call fake as a regression test in `TestDispatchChoice`.
  - ARCH-MOCK: `fakeAgents` returns one fixed answer per agent name, so it can't model a stateful sequence of calls. That is how this case was missed.
- **Plan's Core concepts disagree with the code** (`plan-table-matches-code`, the 2nd finding in this family; see the findings block).
  - The integration table names a `dispatchReview` loop in `dispatch.go`. The code has `judge.DispatchChoice` and `judge.ResolveReview` in `internal/judge/reviewchoice.go`, plus the shell `resolveReviewChoice` in `cmd/sdlc/reviewchoice.go`.
  - The plan says "the ledger round gains `family`, `model`, `author_families` and `degraded`". The code adds one string, `Round.Reviewer` (`gatestate/ledger.go:201-204`).
  - The fix is a Revisions entry (section 7). The rule is stated in the findings block.

**3. Important findings**
- **ARCH-DRY / `no-silent-same-family-review`, 2nd finding: an empty choice turns into claude.**
  - `ReviewChoice.First()` returns a zero `Candidate` when every candidate was skipped before launch.
  - `boundaryReviewOptionsFor` (`milestoneclose.go:897`) sets `opts.Agent = ""`. `agentArgs` maps `""` to claude (`dispatch.go:224`).
  - So `close --dry-run` under `other` in a claude-authored repo with codex and gemini unavailable prints a **claude** command, while the real run refuses.
  - The plan-quality, estimate-quality and `sdlc judge` sites print `invoking  for …` with a blank agent.
- **ARCH-PURPOSE: untested review-site wiring.** No test checks that plan-quality, estimate-quality or `sdlc judge` route through the policy. Only `reviewconfig_test.go` and `reviewchoice_test.go` write a `review.json`, and both cover the boundary review. If those three sites went back to `ResolveAgentCLI`, everything would stay green. Add one test per site under `{"boundary_reviewer":"other"}` and a claude session, asserting codex is dispatched.
- **`docs-track-behavior-change`, 3rd finding: stale atlas text.** `atlas/workflow/sdlc-binary.md:1010-1012` still says every judge dispatch "resolves through `internal/judge.ResolveAgentCLI`". Lines 1195-1197 still say the ledger's `agent` records who reviewed and that `ResolveAgentCLI` is the `same` case. Neither is true after this window.

**4. Minor findings**
- Plan drift: the plan says "Fix reviews inherit their close's kind", but `boundaryChangeKind` (`cmd/sdlc/reviewchoice.go:97`) makes every `FixReview` bounded. The default orders hide this for claude or codex authors; it shows up with gemini or mixed authors, or a custom `reviewer_order`.
- `dead-code`, 2nd finding: `ResolveAgentCLI` and `ResolveReviewerCLI` are called only from tests. `TallyReviewers` and `ParseReviewerRecord` are also test-only (help only documents the raw `git log` one-liner).
- ARCH-SECURE: a launch-failure reason (`dispatch.go:242-245`) is cut to 100 characters (`oneLine`) for the trailer and the ledger. That keeps the owner-bin and `PATH` prefix, so local paths go into commit history, and it drops the actual cause. Put the cause first, or record only `runErr`.
- ARCH-DRY: the "record if a reviewer was used" step is written three times (`milestoneclose.go:770`, `judge.go:208`, `planningreview.go:108`). `boundaryChangeKind` and `judgeChangeKind` are two category→kind maps that overlap.
- `branchWindowBase` hardcodes `origin/main`, ignoring the workspace's resting branch. It degrades to session-only authors, which is safe.
- The sidecar's reviewer row is a markdown table cell. A `|` in a degraded reason breaks the table.

**5. Test coverage notes**
- `go test ./cmd/sdlc/internal/judge ./cmd/sdlc/internal/gatestate` passes, and so do the four new or changed `cmd/sdlc` tests.
- A broad `go test ./cmd/sdlc -run 'Review|Boundary|…'` hit Go's 10-minute default timeout because of suite size, so I can't call the whole package green.
- Missing tests: unavailability on the retry attempt (the Critical above); routing at the non-boundary sites; dry run with zero candidates.

**6. Architecture**
| Principle | Result |
|---|---|
| ARCH-DRY | Minor flags (record closure written 3 times, two kind maps) |
| ARCH-PURE | Pass: `ResolveReview`, `ReviewerRecord` and `Tally` are pure; IO sits in `cmd/sdlc/reviewchoice.go` |
| ARCH-PURPOSE | Flag: untested wiring at three sites |
| ARCH-MOCK | Flag: per-name stateless fake missed the retry sequence |
| ARCH-CONSTRAINTS | Pass: the fallback is bounded by 3 candidates, each under its own `Dispatch` deadline |
| ARCH-SECURE | Pass on author families; Minor flag on paths in the trailer |
| ARCH-ORDER | **Flag (Critical)**: the began-then-unavailable sequence |
| ARCH-FUNERAL | Pass: one optional ledger field per round and one trailer line per close, nothing new on disk |

For upcoming work: carry a "began" bit out of `dispatchAttempts`, so the switch decision rests on the whole attempt history rather than on the last error's type.

**7. Plan revision recommendations**
- "## Revisions — M2 boundary review round 1":
  - Integration points become `judge.ResolveReview` and `judge.DispatchChoice` (`internal/judge/reviewchoice.go`) plus `resolveReviewChoice` and `reviewAvailability` (`cmd/sdlc/reviewchoice.go`).
  - The ledger record is one `reviewer` string in `ReviewerRecord` form, not four fields.
  - Fix-review kind is either changed to inherit the close's kind, or recorded as always bounded, with the reason.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: switch-only-before-review-began
    title: |
      A reviewer whose first attempt ran but whose no-verdict retry is unavailable hands the review to another family
    detail: |
      dispatchAttempts (internal/judge/dispatch.go:386-395) returns attempt 2's UnavailableError unchanged (directly, or via retryFailure). DispatchChoice (reviewchoice.go:98-104) then treats it as a skip, discards attempt 1's output and runs the next candidate. Reproduced with a probe test: codex call 1 gives a no-verdict review with exit 0, call 2 gives a 429 ERROR line with a non-zero exit; result called=[codex codex gemini], used=gemini, err=nil. This violates the plan's state-table row "started ... no switch to another family" (ARCH-ORDER). Fix: on attempt 2, demote UnavailableError to a plain did-not-run error, and add this two-call fake as a regression test.
  - id: new
    severity: Critical
    kind: defect
    family: plan-table-matches-code
    title: |
      The plan's integration table and recording section describe entities and ledger fields the code does not have
    detail: |
      This is the 2nd finding in family plan-table-matches-code. Rule: at every boundary close, each Core concepts row and each named persisted field is checked by grep against the head, and any mismatch goes into a Revisions entry in the same commit. Instances now: a dispatchReview loop in dispatch.go (actually judge.DispatchChoice and ResolveReview in internal/judge/reviewchoice.go, plus resolveReviewChoice in cmd/sdlc/reviewchoice.go); ledger fields family/model/author_families/degraded (actually one Round.Reviewer string).
  - id: new
    severity: Important
    kind: defect
    family: no-silent-same-family-review
    title: |
      An empty ReviewChoice becomes agent "" and so claude: a dry run shows a same-family command the real run would refuse
    detail: |
      This is the 2nd finding in family no-silent-same-family-review. Rule: a ReviewChoice with no candidates must never become a dispatchable zero value; First() returns (Candidate, bool), and every caller (boundaryReviewOptionsFor at milestoneclose.go:897, the changecode.go plan/estimate sites, judge.go) reports the NoReviewerError refusal instead. Today agentArgs maps "" to claude (dispatch.go:224), so close --dry-run prints a claude argv for a claude-authored window.
  - id: new
    severity: Important
    kind: defect
    family: each-wired-site-has-a-test
    title: |
      No test pins that plan-quality, estimate-quality or sdlc judge choose their reviewer through the policy
    detail: |
      Only the boundary review has a policy-driven test (reviewchoice_test.go, reviewconfig_test.go). If changecode.go:726/978 or judge.go:685 went back to ResolveAgentCLI, every test would still pass. Add one test per site under boundary_reviewer=other with a claude session, asserting codex is dispatched.
  - id: new
    severity: Important
    kind: defect
    family: docs-track-behavior-change
    title: |
      The atlas still says every judge dispatch resolves through ResolveAgentCLI, and that the ledger's agent field records the reviewer
    detail: |
      This is the 3rd finding in family docs-track-behavior-change. Rule: a change that retires or replaces a symbol greps atlas/ and helptext/ for that symbol in the same commit (ideally a test that every backticked judge.* identifier in atlas exists and has a non-test caller). Instances: atlas/workflow/sdlc-binary.md:1010-1012 and 1195-1197. Measured prevalence this round: 2 stale passages.
  - id: new
    severity: Minor
    kind: defect
    family: plan-decision-matches-code
    title: |
      Fix reviews are always bounded, but the plan says they inherit their close's kind
    detail: |
      boundaryChangeKind (cmd/sdlc/reviewchoice.go:97) maps FixReview to KindBounded even for a full-flow close. Either inherit the kind or record the deviation in Revisions.
  - id: new
    severity: Minor
    kind: defect
    family: dead-code
    title: |
      ResolveAgentCLI, ResolveReviewerCLI, TallyReviewers and ParseReviewerRecord have no production caller
    detail: |
      This is the 2nd finding in family dead-code. Rule: a migration that moves every caller off an exported function deletes it, or names its remaining consumer, in the same change. Tally helpers that only tests use should either back a command or be documented as test-only.
  - id: new
    severity: Minor
    kind: defect
    family: parse-config-to-typed-value
    title: |
      A launch-failure reason truncated to 100 characters puts local PATH text into the trailer and ledger and drops the cause
    detail: |
      dispatch.go:242-245 builds the reason as "could not launch: dispatch X (owner bin ... PATH=...)", and oneLine cuts it to 100 characters. Put the underlying runErr first.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a459a5a4-a640-42a5-a9f4-5960476d53db","repository":"github.com/xianxu/ariadne","issue_id":"000409","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/583459365d7fc14b","reviewed_head":"b616aee0de69c94aa54dcbb4832c7e498e6a0ea3","slot":"ariadne:2","requested_at":"2026-10-18T13:18:55.68487Z"},"phase":"finished","pid":96620,"started_at":"2026-10-18T13:18:55.856823Z","finished_at":"2026-10-18T13:32:28.88951Z","session_id":"7b029d83-e311-4e05-a0cf-bf8ad53c264b","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T07:02:17-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 409 — Dynamic review policy: choose the reviewer model by author family and change kind (cross-family by default), extending #310 |
| repo | ariadne |
| issue file | workshop/issues/000409-dynamic-review-policy-choose-the-reviewer-model-by-author-family-and-change-kind-cross-family-by-default-extending-310.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | bc15307cf4a96f84853f113cd3c2c6351e9a1d54..fdd8bffbce40de361f3b14cb5491ce644efca54f |
| command | sdlc milestone-close --issue 409 --milestone M2 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-18T07:02:17-07:00 |
| verdict | SHIP |

## Review

Checking each prior finding against the head. Next I'll revert the BR-9 fix in this throwaway checkout to confirm its regression test then fails.

Now the BR-12 check: make the planning sites skip the policy and confirm their tests fail.

Prior findings BR-9, BR-11 and BR-12 check out at the head. Now reading the plan's Revisions and the atlas for BR-10, BR-13 and BR-14.

```verdict
verdict: SHIP
confidence: medium
```

# Code review: ariadne#409 M2 close (round 2)

**Summary.** All four blocking findings from round 1 are fixed, and I checked each one against the code, not the commit message. For BR-9 (Critical) and BR-12 (Important), I reverted the fix in this checkout and the new regression tests failed. With the BR-9 fix reverted, `TestUnavailableRetryNeverHandsOver` reproduces the original bug: the call sequence is `codex,codex,gemini`. With the policy bypassed, the three per-site tests fail with "reviewed by claude". The checkout was restored afterwards and the tree is clean.

BR-10 and BR-13 are resolved by a grep-checked as-built table in Revisions and by corrected atlas text. Every row in that table resolves at the head. Only Minor items remain, so nothing blocks the boundary.

Confidence is medium because the full `cmd/sdlc` suite never finished here. Both runs hit Go's 10-minute default timeout. They reported no `--- FAIL` lines before the timeout. The `internal/judge` and `internal/gatestate` packages and the targeted site tests pass.

## 1. Strengths
- **BR-9 is fixed at the right layer.** At `cmd/sdlc/internal/judge/dispatch.go:388-394`, an `UnavailableError` on attempt 2 is turned into `ErrReviewDidNotRun`. `DispatchChoice` therefore can never treat a review that already began as a skip. Moving `Reason` and `Detail` into `UnavailableError` keeps PATH text out of the trailer and ledger.
- **An empty choice can no longer be dispatched.** `ReviewChoice.First() (Candidate, bool)` plus `Refusal()` (`reviewchoice.go:70-80`) means a caller must decide what to do with no reviewer. All four callers handle it: `changecode.go:730`, `changecode.go:986`, `judge.go:164` and `milestoneclose.go:901`.
- **Each site's policy test catches a real regression.** In `cmd/sdlc/reviewsites_test.go`, a repo set to `other` with a claude session must dispatch codex. Reverting to a no-policy resolve makes all three tests fail.
- **Dead code was deleted, not left behind.** `ResolveAgentCLI` and `ResolveReviewerCLI` are gone. Their precedence tests now run through `ResolveReview` (`judge_test.go:1367-1372`).

## 2. Critical findings
None.

## 3. Important findings
None.

## 4. Minor findings
- **BR-16 has no regression test (kept open as `not-addressed`).** The fix itself is right: at `dispatch.go:667`, `Reason` is now `"could not launch: " + runErr`, and `Detail` holds the owner bin and PATH. But no test pins it: grep finds no `could not launch` in any `_test.go`. Reverting would pass every test.
- `cmd/sdlc/internal/judge/reviewchoice_test.go` fails `gofmt -l`: there is no blank line before `// ReviewerEntry`, and a trailing blank line. This is new in the window.
- Now that `TallyReviewers` lives in a test file, it only tests itself, because the `close --help` sed one-liner is the real tally. `ParseReviewerRecord` still earns its place as the round-trip check. This is noted under BR-15, not raised again.

## 5. Test coverage notes
- **Confirmed by revert:**
  - BR-9 → `TestUnavailableRetryNeverHandsOver`.
  - BR-12 → `TestPlanQualityChoosesThroughThePolicy`, `TestEstimateQualityChoosesThroughThePolicy`, `TestSdlcJudgeChoosesThroughThePolicy`.
- **BR-11:** at the judge level, `TestResolveReview` checks that an empty choice gives `First()` false and that `Refusal` names the skips. At the boundary, `TestBoundaryReviewRefusesWithNoOtherFamily` now goes through `boundaryReviewOptionsFor`'s refusal path. There is still no dry-run test showing that `close --dry-run` prints no argv for an empty choice. That would be cheap to add.
- **Full suite:** `go test ./cmd/sdlc` takes more than 10 minutes on this host. Run it with `-timeout 30m` before the merge.

## 6. Architecture notes
- **ARCH-DRY: pass.** There is one resolver, `resolveReviewChoice` / `resolveReviewChoiceWith`, for every site.
- **ARCH-PURE: pass.** `ResolveReview`, `ReviewPlan` and `ReviewerRecord` are pure. IO sits behind `reviewAvailability` and `judge.Run`.
- **ARCH-PURPOSE: pass.**
  - Every site listed in the M2 plan goes through `ResolveReview` and `DispatchChoice`. That covers close, milestone, fix, plan-quality, estimate-quality and `sdlc judge`.
  - The hard-coded codex in doc-review is explicitly deferred to #441. That is a separate extension, not the point of this issue.
- **ARCH-MOCK: pass.** The fake `judge.Run` and the availability seam are the same seams production uses.
- **ARCH-CONSTRAINTS: pass.** Candidates are tried one after another, at most three, with the existing per-dispatch timeout.
- **ARCH-SECURE: pass.**
  - Agent trailers are treated as untrusted. They can only add families, never remove the closer's own (`reviewchoice.go:18-23`).
  - PATH is now kept out of persisted records.
- **ARCH-ORDER: pass.** The "began" state is enforced at the attempt boundary, and a test covers the ordering where the retry is unavailable.
- **ARCH-FUNERAL: pass.** Nothing new is stored durably except one `reviewer` string per ledger round, which lives as long as the existing ledger.

## 7. Plan revision recommendations
None. The 2026-10-18 Revisions entry matches the code: I grepped every row of its as-built table at the head, and it records the deviation that fix reviews are always bounded.

```findings
dispose:
  - id: BR-9
    disposition: addressed
    note: |
      dispatch.go:388-394 turns an unavailable attempt-2 failure into ErrReviewDidNotRun. With the fix reverted, TestUnavailableRetryNeverHandsOver fails with called [codex codex gemini].
  - id: BR-10
    disposition: addressed
    note: |
      The 2026-10-18 Revisions as-built table supersedes the old tables. Every row checked at the head (family.go, reviewchoice.go, stream.go Began, ledger.go Round.Reviewer, milestoneclose.go reviewTrailers).
  - id: BR-11
    disposition: addressed
    note: |
      First() returns (Candidate, bool) and all four production callers refuse with Refusal(). TestResolveReview pins the empty-choice result.
  - id: BR-12
    disposition: addressed
    note: |
      The three tests in reviewsites_test.go fail when planningReviewChoice and runJudge resolve with an empty policy (checked by revert).
  - id: BR-13
    disposition: addressed
    note: |
      The atlas at sdlc-binary.md:1012 and 1195 now names ResolveReview and the ledger's reviewer field. A grep of atlas/ and helptext/ finds no live mention of a retired symbol.
  - id: BR-14
    disposition: addressed
    note: |
      The deviation (fix reviews are always bounded) is recorded in Revisions with its reason.
  - id: BR-15
    disposition: addressed
    note: |
      ResolveAgentCLI and ResolveReviewerCLI are deleted. The parse and tally helpers moved into a test file and are documented as test-only.
  - id: BR-16
    disposition: not-addressed
    note: |
      The code fix is right (Reason holds runErr, Detail holds PATH), but no test pins it: no _test.go mentions could not launch. Minor, non-blocking.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: gofmt-clean-diff
    title: |
      reviewchoice_test.go in internal/judge is not gofmt-clean
    detail: |
      gofmt -l lists it. There is no blank line between TestReviewerRecordRoundTrip and the ReviewerEntry comment, and a trailing blank line at end of file. Run gofmt -w.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f70a3154-0e64-444f-a601-ff0a54cea014","repository":"github.com/xianxu/ariadne","issue_id":"000409","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/7001651fbb040aee","reviewed_head":"fdd8bffbce40de361f3b14cb5491ce644efca54f","slot":"ariadne:2","requested_at":"2026-10-18T13:39:11.416407Z"},"phase":"finished","pid":14858,"started_at":"2026-10-18T13:39:11.615862Z","finished_at":"2026-10-18T14:02:14.836676Z","session_id":"292b62a9-9001-4c51-b895-95fd8f20e2b3","outcome":"returned"}] -->
