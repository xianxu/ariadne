# Boundary Review — ariadne#441 (milestone M2)

| field | value |
|-------|-------|
| issue | 441 — doc-review: choose its reviewer through the #409 family-aware resolver |
| repo | ariadne |
| issue file | workshop/issues/000441-doc-review-choose-its-reviewer-through-the-409-family-aware-resolver.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 60461c842629b1fcb6889bba8989cf474a39c940..a57f01df3f2eb012cc6a88a2e394da92b9383909 |
| command | sdlc milestone-close --issue 441 --milestone M2 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T10:17:06-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

M2 does most of what it promised. doc-review drops its own agent list in favour of `reviewpolicy.Agent`. It picks a reviewer through `reviewpolicy.Resolve` using the design-heavy order and PATH availability. It treats `same` and the default as `other`, and the agent named on the command line wins. It moves on only when the CLI is missing or fails to launch; a codex crash keeps the hint and does not switch. When no reviewer is left it refuses and names each candidate, and it writes the record to the report's `reviewer_record:` field. Help, atlas and the fix skill are updated.

The build, `go vet` and gofmt are clean for the changed files. The tests for `cmd/doc-review`, `pkg/reviewpolicy` and `judge` pass. `pkg/workspace`'s `TestEnvironmentLegacyAndArchive` fails and gofmt flags `pkg/workspace/identity.go`, but the window does not touch `pkg/workspace`.

One thing blocks SHIP. Under `boundary_reviewer: "explicit"`, a session agent that has no pair is planned as its own reviewer. doc-review then reviews the document with its author's own family, and the record does not mark it as degraded. That breaks the "never self-reviews" promise in the help and the M2 commit. I confirmed it with a throwaway probe test in this checkout. The test that should catch it skips explicit mode in its family check.

1. **Strengths**
   - `cmd/doc-review/choose.go:70-76`: the never-same rule is one small adapter in front of the shared `Resolve`. It sets `Same` to `Other` and clears `AgentCmd`, without copying any of the policy (ARCH-DRY).
   - `cmd/doc-review/review.go:601-644`: `reviewOnce` separates `launchErr` from `err`, so the fallback rule from the plan revision ("a launch failure moves on; a crash does not") is visible in the function's signature. Each attempt gets its own codex `-o` temp file.
   - `cmd/doc-review/choose_test.go:224-297`: the `runReview` tests use a real temporary git repo with `Agent:` trailers and a stateful `runAgent`/`lookPath` fake. They cover the missing CLI, failed launch, crash, refusal and `same` cases end to end.
   - The report path and the record both come from the candidate that actually ran (`outPath(f, abs, c.Agent)`), so a fallback names the right sidecar.

2. **Critical**
   - `cmd/doc-review/choose.go:70-76` (and `pkg/reviewpolicy/policy.go:226-227` / `Reviewer`): explicit mode can self-review without saying so.
     - **Trigger:** `{"boundary_reviewer":"explicit","reviewer_pairs":{"codex":"claude"}}`, a claude session and a claude-written document.
     - **Result:** `Plan` goes to `Reviewer(claude)`, finds no pair, and returns `claude`. The candidates are `[claude]` and the record is `"claude (anthropic; author anthropic)"` with no degraded marker. A pair whose reviewer shares a family with the document's authors behaves the same way.
     - **Fix:** in `chooseReviewers`, when the choice is not forced, move any candidate whose family is among `authors` into `Skips` with the reason "same family as the author". If nothing is left, it refuses.
     - **Regression test:** add the unpaired-explicit case to `TestChooseReviewers` and stop exempting `Explicit` from the family check at `choose_test.go:140`.

3. **Important**
   - The plan's as-built table (`…-plan.md:90`) still lists `Choice.Refusal` in `pkg/reviewpolicy/choice.go`, but this window deleted that method. **This is the 2nd finding in family `plan-table-matches-code`.** The rule is: a commit that adds, renames or removes an identifier listed in the as-built table updates the table in the same commit. Check it by grepping every identifier in the table against the head at each boundary close. Removing a method from a shared package is also an API change that M1 had just published; it should get a Revisions note.

4. **Minor**
   - `cmd/doc-review/main.go:2-3,15` still says "codex by default" and "defaults to codex". **This is the 2nd finding in family `stale-identifier-in-comment`.** The rule is: when a behaviour or identifier changes, grep the package's comments for the old claim in the same commit. Here that means `rg -n 'codex by default|defaults to codex|DefaultAgent' cmd/doc-review`.
   - If an explicitly named agent fails to launch, `NoReviewerError` tells the user to "name one explicitly", which they already did.
   - `docAuthors` runs `git log` without `--follow`, so authors from before a rename are lost and the choice can narrow incorrectly.

5. **Test coverage**
   - All six planned scenarios are covered and pass.
   - Gap: the "no candidate shares an author family" check is turned off for explicit mode, which is exactly where the bug is.
   - Nothing tests the `docAuthors` "authors unread" degraded path when the document is outside any repo.
   - Nothing tests that a dry run's `record:` line matches the record a real run writes.

6. **Architecture**
   - **ARCH-DRY:** pass. The duplicate agent list is gone.
   - **ARCH-PURE:** pass. `chooseReviewers` is pure; git and PATH are injected package vars.
   - **ARCH-PURPOSE:** flag. "Never self-review" is the point of the issue and holds for every mode except explicit (see Critical).
   - **ARCH-MOCK:** pass. git runs against a temporary repo, and the `runAgent` and `lookPath` fakes sit behind the production seam.
   - **ARCH-CONSTRAINTS:** pass. One `git log` per run.
   - **ARCH-SECURE:** pass. The session's family is always among the authors, so a forged trailer can only narrow the candidates. `review.json` is checked by `Validate` via `Load`.
   - **ARCH-ORDER:** pass. Each run is a single ordered fallback loop, with no state kept between runs.
   - **ARCH-FUNERAL:** pass. Per-attempt temp files are removed with `defer`, and the report is the only durable output.

7. **Plan revisions**
   - Add a "## Revisions" entry for the M2 boundary. It should say that `Choice.Refusal` was removed and that refusals are now built as `&NoReviewerError{Skips}`, and update the as-built table at line 90 to match.
   - It should also state the explicit-mode rule for doc-review: a non-forced candidate of the author's family is skipped, never run without a mark.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: never-self-review-holds-in-every-mode
    title: |
      doc-review self-reviews under explicit mode when the session agent has no pair, and the record does not mark it degraded
    detail: |
      chooseReviewers (cmd/doc-review/choose.go:70-76) passes explicit policies straight to Resolve. Plan then asks Policy.Reviewer(main), which returns main when the session has no pair (pkg/reviewpolicy/policy.go:120-124). Probe: Pairs{codex:claude}, a claude session and a claude-authored doc give candidates=[claude] and record "claude (anthropic; author anthropic)" with no degraded marker. A pair whose reviewer shares the authors' family also slips through. This contradicts the help ("never reviews a document with its author's own family") and the M2 commit. Fix: in chooseReviewers, for a non-forced choice, move candidates whose family is among the authors into Skips (refusing if none remain). Add the case to TestChooseReviewers and remove the Explicit exemption at choose_test.go:140.
  - id: new
    severity: Important
    kind: defect
    family: plan-table-matches-code
    title: |
      The plan's as-built table still lists Choice.Refusal, which this window deleted from pkg/reviewpolicy/choice.go
    detail: |
      This is the 2nd finding in family plan-table-matches-code. Rule: any commit that adds, renames or removes an identifier listed in the as-built table updates the table in the same commit, checked by grepping every table identifier against the head at boundary close. Here: remove Refusal from plan.md:90's Choice row and add a Revisions note that the shared package's Choice.Refusal API was dropped (refusals are now &NoReviewerError{Skips}).
  - id: new
    severity: Minor
    kind: defect
    family: stale-identifier-in-comment
    title: |
      cmd/doc-review/main.go:2-3 and :15 still say the reviewer defaults to codex
    detail: |
      This is the 2nd finding in family stale-identifier-in-comment. Rule: when a behaviour or identifier changes, grep the package's comments for the old claim in the same commit (rg 'codex by default|defaults to codex|DefaultAgent' cmd/doc-review).
  - id: new
    severity: Minor
    kind: defect
    family: refusal-names-applicable-remedy
    title: |
      When an explicitly named agent fails to launch, the refusal tells the user to name one explicitly
    detail: |
      A forced candidate that fails to launch ends in NoReviewerError, whose text says "name one explicitly", though the user already did. Return the launch error directly when choice.Forced.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"9ed6e410-73b7-4476-9541-8520fc9569ac","repository":"github.com/xianxu/ariadne","issue_id":"000441","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/5b0857c0d46fe800","reviewed_head":"a57f01df3f2eb012cc6a88a2e394da92b9383909","slot":"ariadne:2","requested_at":"2026-10-18T17:14:35.019626Z"},"phase":"finished","pid":55612,"started_at":"2026-10-18T17:14:35.193984Z","finished_at":"2026-10-18T17:17:06.542053Z","session_id":"9effd4c5-06be-4db4-9c56-c178602c9f7b","outcome":"returned"}] -->
