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
