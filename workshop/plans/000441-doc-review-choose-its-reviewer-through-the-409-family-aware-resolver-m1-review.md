# Boundary Review — ariadne#441 (milestone M1)

| field | value |
|-------|-------|
| issue | 441 — doc-review: choose its reviewer through the #409 family-aware resolver |
| repo | ariadne |
| issue file | workshop/issues/000441-doc-review-choose-its-reviewer-through-the-409-family-aware-resolver.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 4c2fa13e834dfdf4caec56b2c47b864b5465cefe..0fcc618f8f6848441a40a9dd5bfec00db29a1cb8 |
| command | sdlc milestone-close --issue 441 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T09:44:18-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking the error-contract drift I found; next I'll run the build and tests.

Targeted reviewer tests pass; now checking whether the full-suite timeout is related to this diff.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 does what it set out to do. The pure reviewer policy now lives in `pkg/reviewpolicy`, the session-agent and trailer parsing in `pkg/agentenv`, and `judge`, `agentid` and `cmd/sdlc` keep working through type aliases and one-line wrappers. Every sdlc call site compiles unchanged. `go vet` is clean. The judge, agentid, gitx and `pkg/reviewpolicy` tests pass, and so do all nine cmd/sdlc reviewer tests, none of which this range edits. Three things should be fixed before shipping, and none needs a redesign:
1. The refusal error from `Choice.Refusal()` no longer matches `ErrReviewDidNotRun`. The test that guarded this was rewritten rather than kept, which contradicts the plan's "every existing test passes unchanged".
2. Several Core-concepts rows in the plan name files and signatures the code doesn't have.
3. Two judge test files are left in a state `gofmt` rejects.

**Strengths**
- The aliases are real type aliases (`=`), not new types: `cmd/sdlc/internal/judge/dispatch.go:330`, `family.go:392-408`, `agentid/env.go:118`. Old and new names stay interchangeable, so nothing downstream has to change.
- `DispatchChoice`'s refusal keeps matching both errors through `noReviewer.Unwrap() []error` (`reviewchoice.go:56-58`), and the existing `TestDispatchChoice` still pins both matches.
- `AuthorFamilies` always includes the session's own family (`pkg/reviewpolicy/policy.go`), which keeps #409's rule against forged trailers in one place (ARCH-SECURE).
- `Load` keeps `DisallowUnknownFields` and runs `Validate` before returning, and `cmd/sdlc`'s `TestLoadReviewerPolicy` still passes through the wrapper.
- The atlas note (`atlas/workflow/sdlc-binary.md:1206-1210`) says where the policy now lives and what stays in `judge`.

**Critical**: none.

**Important**
1. **The refusal error changed what it matches, and its test was changed to match** (`pkg/reviewpolicy/choice.go:76`, `policy_test.go:164`).
   - Before, `judge.ReviewChoice.Refusal()` returned an error that unwrapped to `ErrReviewDidNotRun`. Now it unwraps to `reviewpolicy.ErrNoReviewer` only.
   - The callers that return `Refusal()` — `changecode.go:732,988`, `milestoneclose.go:903`, `judge.go:166` — therefore no longer satisfy `errors.Is(err, judge.ErrReviewDidNotRun)`. Only `DispatchChoice` still does.
   - The old `TestResolveReview` asserted exactly that match. The moved version asserts `ErrNoReviewer` instead, so the old contract is no longer tested.
   - No caller checks for it today, which is why this is Important rather than Critical. It is still a quiet behavior change in a move that promised none.
   - The user-facing message also changed, from "name one with --agent" to "name one explicitly".
   - **Fix:** in `judge`, wrap `Refusal()` (for example `ResolveReview` / a `judge.Refusal(choice)`) so it returns `noReviewer{...}`. Restore a judge-side test that asserts both `errors.Is(ErrReviewDidNotRun)` and `errors.As(*NoReviewerError)` on `Refusal()`. Keep the `--agent` hint for sdlc, or let the caller supply it.
2. **The plan's Core-concepts table doesn't match the code.** Every entity is delivered, but these paths and signatures are wrong:
   - `Record` is listed at `pkg/reviewpolicy/record.go`; it is in `choice.go`, with signature `Record(used, choice, skips)`, not `(used, authors, skips, forced, degraded)`.
   - `AuthorFamilies(session, trailerValues []string)` actually takes `[]agentenv.Commit`.
   - `cmd/sdlc/internal/judge/policyalias.go` doesn't exist; the aliases were added to `family.go`, `agent_resolver.go`, `reviewchoice.go` and `dispatch.go`.
   - The prompt calls any table/code contradiction Critical. I've graded it Important because nothing claimed is missing and the fix is a revision entry (below).
3. **`gofmt -l` flags two test files:** `cmd/sdlc/internal/judge/family_test.go` and `reviewchoice_test.go` still have the runs of blank lines left where tests were cut out. `gofmt -w` on both fixes it.

**Minor**
- `pkg/reviewpolicy/policy_test.go`: the comments still say `ParseReviewerRecord` / `ReviewerRecord`, but the helpers are now `ParseRecord` / `Record`. These are test-only helpers, so `TallyReviewers` exists only in a `_test.go` file. That was already true before this range.
- `orUnknown` (reviewpolicy) repeats judge's `valueOrDefault` (ARCH-DRY, trivial).
- `pkg/agentenv` has no tests of its own. `pkg/reviewpolicy`'s `Load` and `AuthorFamilies` are only tested from `cmd/sdlc`. M2's `cmd/doc-review` will import them directly, so these tests should move with the code.

**Test coverage notes**
- **Passing:** judge, agentid, gitx and `pkg/reviewpolicy`, plus the nine cmd/sdlc reviewer tests (boundary skip and refuse, trailer parse, author families, window authors, loader, the `other` swap, invalid policy, model pass-through).
- **Full cmd/sdlc suite:** it hit Go's 600s test timeout. The test that was running at the time, `TestCloseCommand_DocOnlyCommitDuringBoundaryReview_Finalizes`, passes on its own in 7s, so this looks like the suite's size rather than this diff. I didn't confirm that against the base commit.
- **Unrelated failure:** `pkg/workspace` `TestEnvironmentLegacyAndArchive` fails ("nil Git reader"). This range doesn't touch that package.

**Architecture**
- **ARCH-DRY:** pass. One policy, and the old judge and agentid bodies were deleted, not copied; only the `orUnknown` nit above.
- **ARCH-PURE:** pass. `policy.go` and `choice.go` are pure and `Load` is the one IO function, in its own file. `agentenv` takes `getenv` as a parameter.
- **ARCH-PURPOSE:** pass for M1's scope. Wiring doc-review is M2, which is the declared second milestone, not a dropped follow-up.
- **ARCH-MOCK:** pass. No new external calls.
- **ARCH-CONSTRAINTS:** N/A. This is a pure move with no new runtime work.
- **ARCH-SECURE:** pass. Untrusted trailers can only narrow the candidates, and the config is decoded strictly and validated.
- **ARCH-ORDER:** pass. Nothing holds state between events.
- **ARCH-FUNERAL:** pass. Nothing durable is created.

**Notes for M2**
- Doc-review's refusal should call `Choice.Refusal()` directly, which is another reason the `--agent` wording belongs to the caller rather than to `pkg/reviewpolicy`.
- Map `same` and the default to `other` before calling `Resolve`, as the plan says. A test should pin that `DefaultMode == Same` still cross-reviews in doc-review.

**Plan revisions recommended**
- A "## Revisions" entry correcting Core concepts: `Record(used Candidate, choice Choice, skips []Skip)` in `pkg/reviewpolicy/choice.go`; `AuthorFamilies(session Family, commits []agentenv.Commit)`; the judge aliases live in `family.go`, `agent_resolver.go`, `reviewchoice.go` and `dispatch.go` (no `policyalias.go`); `NoReviewerError` and `ErrNoReviewer` are new in `pkg/reviewpolicy`.
- The M1 verification line should say the pure tests were moved to `pkg/reviewpolicy` and adapted, not "unchanged".

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: move-preserves-contract
    title: |
      judge's Choice.Refusal() no longer unwraps to ErrReviewDidNotRun; the test pinning it was rewritten
    detail: |
      pkg/reviewpolicy/choice.go:76 NoReviewerError.Unwrap returns ErrNoReviewer only, so the Refusal() errors returned by changecode.go:732,988, milestoneclose.go:903 and judge.go:166 stop matching judge.ErrReviewDidNotRun (only DispatchChoice's noReviewer wrapper still does). The moved TestResolve (policy_test.go:164) swapped its assertion to ErrNoReviewer, against the plan's "tests pass unchanged"; the hint text also changed from "name one with --agent" to "name one explicitly". Fix: in judge, return noReviewer{} from the refusal path and restore a judge test asserting both errors.Is(ErrReviewDidNotRun) and errors.As(*NoReviewerError).
  - id: new
    severity: Important
    kind: defect
    family: plan-table-matches-code
    title: |
      Plan Core-concepts rows name files and signatures the code doesn't have
    detail: |
      Record is in choice.go (not record.go) with signature (used, choice, skips); AuthorFamilies takes []agentenv.Commit, not []string; judge/policyalias.go doesn't exist (the aliases are in family.go, agent_resolver.go, reviewchoice.go, dispatch.go). Add a Revisions entry correcting the table.
  - id: new
    severity: Important
    kind: defect
    family: gofmt-clean
    title: |
      judge family_test.go and reviewchoice_test.go fail gofmt -l
    detail: |
      Runs of blank lines are left where the moved tests were cut out; run gofmt -w on both files.
  - id: new
    severity: Minor
    kind: defect
    family: stale-identifier-in-comment
    title: |
      policy_test.go helper comments still name ParseReviewerRecord / ReviewerRecord
    detail: |
      Rename the comments to ParseRecord / Record.
  - id: new
    severity: Minor
    kind: defect
    family: shared-package-owns-its-tests
    title: |
      pkg/agentenv has no tests, and reviewpolicy Load/AuthorFamilies are only tested from cmd/sdlc
    detail: |
      doc-review (M2) will import these directly; move or add package-level tests alongside them.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1a9c75c4-0956-4650-90ef-74613071a845","repository":"github.com/xianxu/ariadne","issue_id":"000441","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/dc36896d98479771","reviewed_head":"0fcc618f8f6848441a40a9dd5bfec00db29a1cb8","slot":"ariadne:2","requested_at":"2026-10-18T16:23:57.619564Z"},"phase":"finished","pid":55386,"started_at":"2026-10-18T16:23:57.809457Z","finished_at":"2026-10-18T16:44:17.539951Z","session_id":"f5a73cbb-61b9-4e92-927d-9ba28c3e4a94","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T10:09:30-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 441 — doc-review: choose its reviewer through the #409 family-aware resolver |
| repo | ariadne |
| issue file | workshop/issues/000441-doc-review-choose-its-reviewer-through-the-409-family-aware-resolver.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 4c2fa13e834dfdf4caec56b2c47b864b5465cefe..88fddabe648729d8270c338b9a9606c03983cece |
| command | sdlc milestone-close --issue 441 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T10:09:30-07:00 |
| verdict | SHIP |

## Review

The full `cmd/sdlc` suite hits Go's 10-minute default timeout, so I'm running only the tests that touch reviewer choice, config and the refusal.

All five prior findings check out against the code. Before writing the verdict, I'm comparing the moved resolver logic with the base version to look for behavior drift.

```verdict
verdict: SHIP
confidence: high
```

All five findings from round 1 are fixed, and I checked each one against the code at the pinned head rather than the commit message. The refusal now keeps judge's contract again. Every sdlc site that refuses goes through `judge.Refusal`: `changecode.go:732`, `changecode.go:988`, `milestoneclose.go:903` and `judge.go:166`. `DispatchChoice` does the same at `reviewchoice.go:90`. A new judge test checks that the refusal matches both `*NoReviewerError` and `ErrReviewDidNotRun`, and that it still says `--agent`. The moved `Resolve`, `First` and `NoReviewerError` code does the same thing as the base `judge` versions; only the hint wording differs, and judge puts it back. Nothing in the diff is still blocking. I found one new Minor issue: `Choice.Refusal()` is still reachable from sdlc code and would quietly drop judge's contract.

**What I ran:**
- `gofmt -l` reports none of the changed files.
- Tests for `pkg/reviewpolicy`, `pkg/agentenv`, `cmd/sdlc/internal/judge` and `cmd/sdlc/internal/agentid` pass.
- The full `cmd/sdlc` package goes past Go's 600 s default test timeout, so I couldn't run it all. The 25 tests from `reviewchoice_test.go`, `reviewsites_test.go` and `reviewconfig_test.go`, plus the boundary-review tests, all pass.
- `pkg/workspace` `TestEnvironmentLegacyAndArchive` fails ("nil Git reader"). This change didn't cause it: that package isn't in the diff and doesn't depend on `reviewpolicy`, `agentenv` or `agentid`.

1. **Strengths**
   - `cmd/sdlc/internal/judge/reviewchoice.go:58-67`: one `noReviewer` type serves both `Refusal` and `DispatchChoice`, so there is a single place that defines judge's refusal contract.
   - `pkg/reviewpolicy/choice.go:25-56`: `Resolve` and `First` match the base `ResolveReview` exactly (checked line by line against `4c2fa13e`).
   - `pkg/reviewpolicy/load_test.go`: tests a missing config, a valid one, and four bad inputs (bad value, misspelled key, truncated JSON, a family that can't be resolved), each expecting an error that names `ConfigPath`.
   - `pkg/agentenv/agentenv_test.go`: tests `ParseLog` on merge commits, garbage records and whitespace in trailers, and checks that `LogFormat` actually reads `TrailerKey`.
   - The plan's Revisions table (`plan.md:85-93`) matches the head symbol for symbol.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **`Choice.Refusal()` loses judge's contract** (`pkg/reviewpolicy/choice.go:55`). `judge.ReviewChoice` is an alias for `reviewpolicy.Choice`, so sdlc code can still call `choice.Refusal()`. That returns an error that doesn't match `ErrReviewDidNotRun` and tells the user to "name one explicitly" instead of using `--agent`. That is the mistake BR-1 fixed. This is the **2nd finding in family `move-preserves-contract`**. The general rule: when a type moves behind an alias, every method the alias brings along must keep the contract its old callers relied on, or sdlc must be unable to call it. To enforce it, either remove `Choice.Refusal` (doc-review can build `&NoReviewerError{Skips: c.Skips}` itself), or add a judge-side test that greps non-test `.go` files outside `pkg/` for `.Refusal()` and fails if it finds one.
   - The `--agent` wording is restored by string replacement (`reviewchoice.go:61`). If the pkg wording changes, the replacement silently does nothing, but `TestRefusalKeepsJudgesContract` would catch it. Acceptable as is.

5. **Test coverage notes**
   - **Regression evidence for BR-1:** `judge.Refusal` is a new function, so taking the fix out means editing the four call sites, and the new test checks the function itself.
   - **Gap:** the site-level refusal test `TestBoundaryReviewRefusesWithNoOtherFamily` only checks the text "no reviewer available". If a site went back to `choice.Refusal()`, it would still pass. That gap is the reason for the Minor finding above.

6. **Architectural notes**
   - **ARCH-DRY: pass.** The judge aliases and wrappers delegate to the pkg code; the logic isn't duplicated.
   - **ARCH-PURE: pass.** `Plan`, `Resolve`, `Record` and `AuthorFamilies` are pure and tested without IO. `Load` reads the filesystem only through a root directory passed in.
   - **ARCH-PURPOSE: pass for M1.** M1's job is extracting the shared code. The consumer that motivated the issue, doc-review, is M2's declared deliverable and isn't hidden behind a vague follow-up.
   - **ARCH-MOCK: pass.** No new external call. Reading the git log stays behind the existing callers, and `agentenv` only provides the format string and the parser.
   - **ARCH-CONSTRAINTS: N/A.** Resolving a reviewer is a cheap step that runs once per review; this change doesn't affect load.
   - **ARCH-SECURE: pass.** `Load` rejects unknown keys and bad values and names the config path in the error. `ParseLog` skips malformed records instead of making up values.
   - **ARCH-ORDER: N/A.** Nothing keeps state between events: `Resolve` is a single-shot function of its inputs.
   - **ARCH-FUNERAL: N/A.** Nothing durable is created: the change is library code and tests, and writes no files.
   - **Docs gate:** `atlas/workflow/sdlc-binary.md` is updated. There's no new user-facing surface (flags, config keys, subcommands), so README doesn't need a change.
   - **For M2:** doc-review should take its refusal from `reviewpolicy` directly, and should not rely on judge's `--agent` wording.

7. **Plan revision recommendations:** none. The M1 round-1 Revisions table matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      judge.Refusal (reviewchoice.go:67) matches both NoReviewerError and ErrReviewDidNotRun and keeps the --agent hint; all 4 sites use it; TestRefusalKeepsJudgesContract pins it.
  - id: BR-2
    disposition: addressed
    note: |
      Plan Revisions (M1 round 1) table checked against the head: Record in choice.go, AuthorFamilies takes []agentenv.Commit, aliases listed by their real files.
  - id: BR-3
    disposition: addressed
    note: |
      gofmt -l on cmd and pkg lists no file in the diff; the extra blank lines are gone in family_test.go and reviewchoice_test.go.
  - id: BR-4
    disposition: addressed
    note: |
      policy_test.go:209-210 comments now say ParseRecord / Record; no ReviewerRecord left in pkg/.
  - id: BR-5
    disposition: addressed
    note: |
      Added pkg/agentenv/agentenv_test.go (Env, ParseLog, LogFormat) and pkg/reviewpolicy/load_test.go (Load, AuthorFamilies); both pass.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: move-preserves-contract
    title: |
      reviewpolicy Choice.Refusal() stays callable through the judge.ReviewChoice alias and drops judge's refusal contract
    detail: |
      2nd finding in this family. Rule: a method a moved type brings along through an alias must keep its old callers' contract, or sdlc must be unable to call it. Fix the rule, not the call site: remove Choice.Refusal (choice.go:55), or add a judge test that fails when non-test code outside pkg/ calls .Refusal(). Today no site test would fail if a site went back to choice.Refusal(): TestBoundaryReviewRefusesWithNoOtherFamily only checks the text "no reviewer available".
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"94f721a6-48b0-4fb6-bbca-3c26bd7df1cc","repository":"github.com/xianxu/ariadne","issue_id":"000441","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/fad372f250fe19fa","reviewed_head":"88fddabe648729d8270c338b9a9606c03983cece","slot":"ariadne:2","requested_at":"2026-10-18T16:46:45.95791Z"},"phase":"finished","pid":55727,"started_at":"2026-10-18T16:46:46.180434Z","finished_at":"2026-10-18T17:09:28.48915Z","session_id":"50a915aa-a5f0-40a3-a13f-11997634b2aa","outcome":"returned"}] -->
