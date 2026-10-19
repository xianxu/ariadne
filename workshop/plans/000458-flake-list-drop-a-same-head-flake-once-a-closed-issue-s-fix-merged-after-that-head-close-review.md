# Boundary Review — ariadne#458 (whole-issue close)

| field | value |
|-------|-------|
| issue | 458 — flake list: drop a same-head flake once a closed issue's fix merged after that head |
| repo | ariadne |
| issue file | workshop/issues/000458-flake-list-drop-a-same-head-flake-once-a-closed-issue-s-fix-merged-after-that-head.md |
| boundary | whole-issue close |
| milestone | — |
| window | a11504c417e9c8789327362895bc88c9fad042df..1b56f757d2b7852f06c162fc2b39a97c0078eb14 |
| command | sdlc close --issue 458 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T05:27:07-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The ancestry decision and conservative fallback behavior passed the targeted tests. SHIP is blocked by a contradiction in the plan’s PURE classification, missing command-level regression coverage, and missing README documentation. Per-row uncertainty also loses the explanation promised by the plan.

```findings
findings:
  - id: new
    severity: Critical
    family: pure-core-contract
    title: |
      ARCH-PURE: the declared pure decision performs injected Git IO
    detail: |
      cmd/sdlc/flake_fixed.go:219–223 invokes contains, wired to real Git in production and in TestFixedFlakeAgainstRealGitMergeHistory. The plan’s Core concepts table at lines 19–22 instead declares a pure FixedFlakeDecision taking ancestry answers; neither that symbol nor FixedFlakeIssue exists under the declared names. Move ancestry execution into the integration shell and pass explicit results into a pure decision, or revise the table to name the actual entities and classify the orchestration as INTEGRATION.
  - id: new
    severity: Important
    family: user-surface-documentation
    title: |
      README update appears missing for flake list -v
    detail: |
      cmd/sdlc/flake_cmd.go:261 introduces --verbose/-v and changes default list behavior, but README.md:525 still describes only the previous daily view. Update that passage with default suppression, verbose attribution, and recurrence behavior; atlas coverage does not replace README usage.
  - id: new
    severity: Important
    family: observable-behavior-regression-coverage
    title: |
      Command tests do not pin default suppression or recurrence rendering
    detail: |
      cmd/sdlc/flake_fixed_test.go:83 tests the classifier against real Git, while lines 124–147 exercise only verbose fixed-row rendering with synthetic ancestry answers. No test asserts that the default command hides a proven fixed row or that a recurrence stays visible with its annotation. Add command-level tests covering both modes and conservative failure cases, using the real-history seam required by Done when.
  - id: new
    severity: Important
    family: uncertainty-remains-observable
    title: |
      Per-row evidence failures discard the promised coverage reason
    detail: |
      cmd/sdlc/flake_fixed.go:203–225 collapses unreadable details, ambiguous ownership, and ancestry errors into flakeUnknown without a reason. cmd/sdlc/flake_cmd.go:487 prints those rows unchanged; its coverage message handles only catalogue-wide FixedErr. Preserve a reason in the decision result and render it so these cases satisfy the plan’s line-34 promise that unknown rows remain visible with a coverage reason.
```

1. **Strengths**

   - Published details come from one pinned main snapshot, including archived issues (`flake_fixed.go:87–120`).
   - Missing merge objects and uncertain ancestry preserve visibility; real Git tests exercise these decisions.
   - Ancestry results are cached by merge/head pair, with repeated-attempt tests confirming two calls (`flake_cmd.go:438–459`).
   - Ordinary Log mentions do not claim ownership; dedicated legacy test lines receive separate handling.

2. **Critical findings**

   - `pure-core-contract`: reconcile the actual IO boundary with the Core concepts table as described above. This is a classification and separation issue, not a request to export internal names.

3. **Important findings**

   - Document `-v` and changed defaults in README.
   - Test observable default and recurrence output.
   - Preserve and display per-row uncertainty reasons.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Ran:

   `go test ./cmd/sdlc -run '^(TestFlake|TestFixed)' -count=1`

   Passed in 47.111 seconds. Ownership matching, conservative classification, real merge ancestry, archived-detail loading, and ancestry caching have coverage. The command-output gaps above remain. The pair pilot and full-suite claims were not independently verified. The checkout remained clean.

6. **Architectural notes**

   - **ARCH-DRY — pass:** reuses tracker, snapshot, registry, and Git helpers.
   - **ARCH-PURE — flag:** decision orchestration executes Git callbacks despite its PURE declaration.
   - **ARCH-PURPOSE — flag:** observable command behavior lacks the promised regression coverage and uncertainty explanations.
   - **ARCH-MOCK — pass for this scope:** immutable ancestry queries have an injected seam and real local-Git fixtures; no new external mutation sequence.
   - **ARCH-CONSTRAINTS — pass for Git fan-out:** batching and per-pair caching are exercised. The scaling test establishes Git-call counts, not total CPU cost.
   - **ARCH-SECURE — pass:** malformed or unavailable evidence conservatively retains rows; Git uses argument arrays.
   - **ARCH-ORDER — pass:** each invocation evaluates snapshots and holds no authoritative state between events.
   - **ARCH-FUNERAL — pass:** no new durable artifact family.
   - **ARCH-EVOLVE — pass:** no persisted-format rewrite or schema change.

7. **Plan revision recommendations**

   Add a `## Revisions` entry naming the implemented entities and their actual PURE/INTEGRATION boundaries. Update the obsolete task paths under `internal/flakes`, and reconcile task completion with the committed command tests. Retain the uncertainty-explanation requirement and implement it.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"8d765e44-9a65-4e75-91b9-905d30a95b5e","repository":"github.com/xianxu/ariadne","issue_id":"000458","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/28a4f2053ee07160","reviewed_head":"1b56f757d2b7852f06c162fc2b39a97c0078eb14","slot":"ariadne:1","requested_at":"2026-10-19T12:25:12.330439Z"},"phase":"finished","pid":47557,"started_at":"2026-10-19T12:25:12.427098Z","finished_at":"2026-10-19T12:27:07.102181Z","session_id":"01a1541f-7f0e-7d72-a761-660c9890a1fd","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T06:02:41-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 458 — flake list: drop a same-head flake once a closed issue's fix merged after that head |
| repo | ariadne |
| issue file | workshop/issues/000458-flake-list-drop-a-same-head-flake-once-a-closed-issue-s-fix-merged-after-that-head.md |
| boundary | whole-issue close |
| milestone | — |
| window | a11504c417e9c8789327362895bc88c9fad042df..e4e8590d273723c84d7cc000c381abc16d622cdd |
| command | sdlc close --issue 458 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T06:02:41-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

All four prior findings are addressed. The targeted tests pass, and mutation checks confirm that the new command test detects regressions in suppression, recurrence annotation, and uncertainty rendering. One correctness bug blocks shipping: package-basename matching can attribute an unrelated test to a closed issue and silently hide its failure.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Plan lines 19–23 now name the implemented entities and classify classifyFixedFlake as INTEGRATION, matching its injected Git calls. fixedIssueCovers remains independently testable without IO.
  - id: BR-2
    disposition: addressed
    note: |
      README.md:527–533 documents default suppression, flake list -v, recurrence visibility, and uncertain evidence; these descriptions match flake_cmd.go:474–486.
  - id: BR-3
    disposition: addressed
    note: |
      TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit exercises default and verbose output, recurrence, and ambiguous coverage using real merge history. Temporary source overlays disabling suppression or recurrence annotation each made it fail.
  - id: BR-4
    disposition: addressed
    note: |
      classifyFixedFlake now returns reasons for unreadable details, ambiguous ownership, missing merges, and ancestry failures. Decision tests assert these reasons; removing coverage-reason rendering made the real-history command test fail.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: exact-test-ownership
    title: |
      ARCH-PURPOSE: package-basename matching can hide an unrelated test
    detail: |
      cmd/sdlc/flake_fixed.go:171 reduces package identity to path.Base(pkg). For a failure named github.com/xianxu/ariadne/cmd/internal/launcher.TestX, details reading “TestX fails in pkg/launcher/run_test.go” incorrectly establish ownership. A temporary regression test confirmed that classifyFixedFlake returns flakeFixed when the issue’s merge is on main but absent from the failing head, causing default rendering to suppress the unrelated failure. bareCollision does not protect this case when only one package occurs in retained attempts. Require unambiguous package identity for textual ownership, retain uncertain rows, and cover same-basename packages across substantive sections and dedicated Log lines.
```

1. **Strengths**

   - Published active and archived details come from one pinned main snapshot.
   - Git failures preserve visibility and now explain uncertainty.
   - Ancestry caching has counted repeated-attempt coverage.
   - The revised concept table accurately distinguishes pure matching from integration orchestration.

2. **Critical findings**

   - `flake_fixed.go:171`: preserve package identity when matching legacy textual ownership. A shared final directory name cannot establish that two tests belong to the same package.

3. **Important findings**

   None remaining from the prior round.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Passed: `go test ./cmd/sdlc -run '^Test(Flake|Fixed)' -count=1` — 47.558 seconds.
   - Three temporary mutations independently made the command regression test fail.
   - The temporary same-basename ownership regression failed against the pinned implementation, confirming the new defect.
   - Temporary overlays were removed; tracked files were unchanged. The reported pair pilot and full suite were not independently rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** existing tracker, snapshot, registry, and Git helpers are reused.
   - **ARCH-PURE — pass:** ownership matching is pure; Git orchestration is correctly classified.
   - **ARCH-PURPOSE — flag:** incorrect ownership can suppress a failure without an applicable fix.
   - **ARCH-MOCK — pass for this scope:** ancestry uses an injected seam and portable real-Git fixtures.
   - **ARCH-CONSTRAINTS — pass for Git fan-out:** per-pair caching is tested; total CPU scaling was not measured.
   - **ARCH-SECURE — flag:** external issue text is promoted to exact ownership despite ambiguous package identity.
   - **ARCH-ORDER — pass:** invocation-local snapshots introduce no authoritative state across events.
   - **ARCH-FUNERAL — pass:** no new durable artifact family.
   - **ARCH-EVOLVE — pass:** no persisted-format changes or rewrites.

7. **Plan revision recommendations**

   Add a `## Revisions` entry clarifying that legacy package context must identify the package unambiguously. Enumerate and test all textual ownership paths: substantive sections and dedicated Log lines, including different packages sharing the same basename.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"09f5f5d6-fdf5-4217-9b31-b3337ecf0308","repository":"github.com/xianxu/ariadne","issue_id":"000458","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/feb76cf032713ca5","reviewed_head":"e4e8590d273723c84d7cc000c381abc16d622cdd","slot":"ariadne:1","requested_at":"2026-10-19T12:59:54.84211Z"},"phase":"finished","pid":84333,"started_at":"2026-10-19T12:59:54.936514Z","finished_at":"2026-10-19T13:02:41.515882Z","session_id":"01a1543f-45bc-7b93-b21d-fabbef7cb8e7","outcome":"returned"}] -->
