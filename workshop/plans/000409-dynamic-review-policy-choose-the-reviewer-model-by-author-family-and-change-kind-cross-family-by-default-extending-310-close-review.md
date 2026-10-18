# Boundary Review — ariadne#409 (whole-issue close)

| field | value |
|-------|-------|
| issue | 409 — Dynamic review policy: choose the reviewer model by author family and change kind (cross-family by default), extending #310 |
| repo | ariadne |
| issue file | workshop/issues/000409-dynamic-review-policy-choose-the-reviewer-model-by-author-family-and-change-kind-cross-family-by-default-extending-310.md |
| boundary | whole-issue close |
| milestone | — |
| window | d3ebb55260c6647fbc70b423efed20cd8c734c5b..e5dca284acc7c1d10d3b354130fac751304ceca2 |
| command | sdlc close --issue 409 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-18T07:51:45-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking the earlier findings: BR-17's gofmt issue is fixed. I'm now checking whether BR-16's fix has a regression test.

BR-16 remains not-addressed (code is right, still no test). Now running the build and tests, then reading the core files.

Tests for the diff pass (judge, gatestate, and the main-package review tests). Now checking the default mode against the spec and the trailer-tally one-liner.

Wiring looks consistent across sites. I'm now probing how authors are attributed (`reviewAuthorFamilies`) and its tests.

The function that reads author families from the window's `Agent:` trailers has no test. I'll check it against a scratch repo with a temporary test (in this disposable checkout).

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

#409 delivers what it set out to do. Every production review dispatch now goes through one resolver and one dispatch loop: `ResolveReview` → `DispatchChoice` (`cmd/sdlc/internal/judge/reviewchoice.go:50,104`). A grep finds no other production `Dispatch(` caller. Skipped reviewers are announced, an empty choice refuses at every site, and the `Review-Reviewer:` trailer, ledger round and sidecar all carry the same `ReviewerRecord` text. `go vet ./...` is clean. `go test ./internal/judge ./internal/gatestate` passes. The `#409`-related tests in the main package pass (194s). The full main-package run timed out after 600s in `TestAssetDestinationPolicyRejectsNonregularAndOversizedFiles`, which this diff doesn't touch. I ran the help's weekly tally one-liner against all three record shapes and it gives the right `author -> reviewer` pairs.

What stands between this and SHIP:
- **Important:** the IO step that turns the window's `Agent:` trailers into author families has no test.
- **Minor:** BR-16 still has no regression test.
- **Minor:** a new instance of the silent same-family family.

**BR-17 is fixed:** `gofmt -l cmd/` no longer lists `reviewchoice_test.go`. The two files it does list (`activetime/split_test.go`, `fleet/types.go`) are outside this diff.

### 1. Strengths
- **One choice serves every review site:** boundary, fix, plan-quality, estimate-quality and `sdlc judge` all take the reviewer from `ReviewChoice`. `First()` returns `(Candidate, bool)`, so the zero candidate (which would have run claude) can't be dispatched (`reviewchoice.go:72`).
- **Treating forged trailers as untrusted is the right model:** `cmd/sdlc/reviewchoice.go:18-24` always includes the session's own family. A forged trailer can only add a family, so it can never steer the review back to the closer's own family.
- **The unavailability evidence is careful, per CLI:**
  - codex: causes are read only from its `ERROR:` lines, and any activity block (`codex`/`exec`/`thinking`) means the review began (`family.go:157-161,179`).
  - claude: a stream counts as unavailable only if it ended before any assistant or tool event (`AgentRun.Began`).
- **The end-to-end boundary tests check what the operator sees:** `TestBoundaryReviewSkipsAnUnavailableReviewerVisibly` checks the skip line, the record and the trailer. `TestBoundaryReviewRefusesWithNoOtherFamily` checks that nothing launched (`cmd/sdlc/reviewchoice_test.go:53,79`).
- **Model values are checked before they reach any argv:** a `reviewer_models` value that is blank, has whitespace or a newline, or starts with `-` is refused (`agent_resolver.go:127`).

### 2. Critical findings
None.

### 3. Important findings
- **`reviewAuthorFamilies` is untested** (`cmd/sdlc/reviewchoice.go:23-41`). This function is the Spec's first input: the author family read from the commits. ARCH-PURPOSE: the Done-when says "the policy picks per author family". Today only the pure `ReviewPlan`, given hand-made author sets, is tested. Every site test runs a window with no `Agent:` trailers, so the whole chain goes unchecked: `git log --first-parent --no-merges` → `agentid.ParseLog` → cutting at `/` → `FamilyOf`.
  - I confirmed by hand that it works today: a scratch repo with `Agent: codex/gpt-5/high` and a claude session gives `[anthropic openai]`.
  - Nothing would catch a regression in the trailer format or the parsing. Such a regression would silently send codex-authored work to codex.
  - Fix: add a table test over a temp repo covering a codex trailer, a model suffix, a merge commit that should be ignored, and an unknown agent. Better, extract the trailer→families step as a pure function that takes the log text, and test it without IO (ARCH-PURE).

### 4. Minor findings
- **BR-16 is still not-addressed.** The code is right (`dispatch.go:667`: `Reason` holds `runErr`, `Detail` holds PATH), but no `_test.go` asserts that a launch-failure `UnavailableError.Reason` excludes `PATH=` and leads with the cause. The existing `TestDispatch_LaunchError_Surfaces` only checks `err.Error()`, which still contains PATH through `Detail`.
- **Author-read failure is silent** (`reviewchoice.go:29-31`). If `git log` fails, the trailer authors are dropped and only the session's family is excluded. I confirmed this with a bad base, which returned `[anthropic]`. The same happens when `branchWindowBase` returns "" (no `origin/main`). So a codex-authored change can be reviewed by codex, and the trailer then misrecords `author anthropic`. **This is the 3rd finding in family `no-silent-same-family-review`.** The rule that covers all three: *any input the cross-family exclusion depends on (policy, order, authors) that cannot be read must either refuse or be recorded as degraded in the `ReviewerRecord`; it must never narrow the excluded families silently.* Apply that rule to `ReviewPlan`'s inputs as a whole, not just this call.
- **The help's kind description doesn't mention fix reviews or `sdlc judge`.** `close.md` says "a bounded (quick-flow) change", but fix reviews are also bounded (`boundaryChangeKind`), and `sdlc judge` maps kind by recipe.
- **The not-run record names a reviewer that never ran.** When every candidate fails at runtime (a `NoReviewerError` from `DispatchChoice`), `r.Agent` still holds the first candidate's name (`milestoneclose.go:741`). The reason text names the skips, so this is cosmetic.

### 5. Test coverage notes
- Covered: the demotion in `TestUnavailableRetryNeverHandsOver` (BR-9), and every-site policy tests for plan-quality, estimate-quality and `sdlc judge` (BR-12).
- Gaps: author-family extraction (Important above) and the launch-failure `Reason` contents (BR-16).

### 6. Architectural notes
- **ARCH-DRY: pass.** One resolver, one loop, one record function. `oneLine`/`firstLine` are reused.
- **ARCH-PURE: pass, with a note.** `ReviewPlan`, `Unavailable`, `ReviewerRecord` and `LocalAvailability` (lookPath injected) are pure. The trailer→families step is mixed into the git call; pulling it out would make the Important finding cheap to test.
- **ARCH-PURPOSE: pass, with a note.** The Done-when tally is a documented git one-liner rather than a subcommand, which the plan records. The `same` default contradicts the Spec's "default rule: cross-family", but the plan records it as an operator decision kept in one constant (`DefaultReviewerMode`). The author-input test gap is above.
- **ARCH-MOCK: pass.** The agent CLIs are reached through the `judge.Run` and `reviewAvailability` seams. Site tests run `fakeReviewer` behind the same seams production uses.
- **ARCH-CONSTRAINTS: pass.** Candidates are tried one after another, at most three, with no fan-out.
- **ARCH-SECURE:**
  - Trailers are treated as untrusted and model values are validated: pass.
  - Silent author-read failure: flagged (Minor above).
- **ARCH-ORDER: pass.** Switching reviewers only before a review began is enforced: an unavailable result on attempt 2 is demoted. `DispatchChoice` is a bounded, ordered loop that is tested through injected outcomes.
- **ARCH-FUNERAL: pass.** No new artifact family. Each ledger round and trailer gains one bounded string.

### 7. Plan revision recommendations
- Add a Revisions note on how author families are read: from the first-parent `Agent:` trailers plus the session. Say what happens when they can't be read (after the fix above: degraded or refused).

```findings
dispose:
  - id: BR-16
    disposition: not-addressed
    note: |
      Code is correct at dispatch.go:667 (Reason holds runErr, Detail holds PATH) but still no test asserts that Reason excludes PATH and leads with the cause.
  - id: BR-17
    disposition: addressed
    note: |
      gofmt -l cmd/ at e5dca284 no longer lists internal/judge/reviewchoice_test.go.
findings:
  - id: new
    severity: Important
    family: io-input-path-tested
    title: |
      reviewAuthorFamilies, the Spec's author-family input from Agent: trailers, has no test
    detail: |
      cmd/sdlc/reviewchoice.go:23-41 turns the window's first-parent Agent: trailers into families, but every site test runs a trailer-less window and only the pure ReviewPlan is tested with hand-made author sets. A regression in agentid parsing or the cut at the slash would silently send codex-authored work to codex. Add a temp-repo table test (codex trailer with a model suffix, merge commit ignored, unknown agent), or extract a pure log-text to families function and test that.
  - id: new
    severity: Minor
    family: no-silent-same-family-review
    title: |
      A failed git log or a missing origin/main silently drops the trailer authors from the exclusion
    detail: |
      This is the 3rd finding in family no-silent-same-family-review. reviewchoice.go:29-31 returns only the session family when git log fails (confirmed with a bad base, which gave anthropic only), and branchWindowBase returns empty without origin/main, so a codex-authored change can be reviewed by codex and the trailer records author anthropic. Rule for the class: any input the cross-family exclusion depends on (policy, order, authors) that cannot be read must refuse or be recorded as degraded in ReviewerRecord, never silently narrow the excluded families.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cc4effe3-dcb7-4d0b-bec4-cca56ce00f0a","repository":"github.com/xianxu/ariadne","issue_id":"000409","recipe":"milestone-review","round":6,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c05b825593fbc47e","reviewed_head":"e5dca284acc7c1d10d3b354130fac751304ceca2","slot":"ariadne:2","requested_at":"2026-10-18T14:34:49.347125Z"},"phase":"finished","pid":97088,"started_at":"2026-10-18T14:34:49.435409Z","finished_at":"2026-10-18T14:51:44.314908Z","session_id":"48f48ad3-5673-4dd3-aab1-722acf4a3624","outcome":"returned"}] -->
