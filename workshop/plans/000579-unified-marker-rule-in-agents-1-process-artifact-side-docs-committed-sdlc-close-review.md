# Boundary Review — ariadne#579 (whole-issue close)

| field | value |
|-------|-------|
| issue | 579 — Unified marker rule in AGENTS §1; process-artifact side docs committed; sdlc refuses links to uncommitted files |
| repo | ariadne |
| issue file | workshop/issues/000579-unified-marker-rule-in-agents-1-process-artifact-side-docs-committed-sdlc.md |
| boundary | whole-issue close |
| milestone | — |
| window | 56172a1009f553fa227d1085f8873c3d636da3ac..493d00a8e64923ccfa9443ad6e65e8a0f4352fa4 |
| command | sdlc close --issue 579 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T18:26:53-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the #579 diff: I've read the plan and the full patch. Now I'm checking how close handles errors, the TrunkView directory semantics, and running the tests.

The new #579 tests pass. The broader publish/change-code suites exceed the time limit in this sandbox, so I'm checking a few specific error paths by reading the code instead.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The issue's goal is delivered. AGENTS.base.md §1, review-convention §6 and collaboration-modes §4 all state the unified marker rule and the commit-with-artifact rule. A search found no copy of the old "resolve it in place" wording outside #579's own issue text. The link grammar now has one source: `sidedoc.LocalLinks`, which `RebaseLinks` also uses. The guard is a pure census with injected state and read functions. `change-code`, `close`/`milestone-close` and `issue publish` all call it before their first effect. D1 (publish carries the linked side docs) also re-checks main's copy inside the commit's prepare step. All six new or changed #579 tests pass here, plus the `sidedoc`, `linkguard` and `processmanual` packages. Two Important items keep this from SHIP. First, close's guard runs even when bypassed, and any project-discovery or peer-repo git error kills the close with no way around it, where close's own project code only warns on the same error. Second, the cross-repo project branch and change-code's real flag and `--force` handling have no tests, though the plan listed two tests for them.

**Strengths**
- `pkg/sidedoc/rebase.go:115-164`: pulling out `localPath` and `eachTarget` gives one link grammar for "what moves with the archive" and "what must be committed". The existing `RebaseLinks` tests pass unchanged (ARCH-DRY).
- `cmd/sdlc/internal/linkguard/guard.go:199-234`: `Census` is pure. It asks each target's state once, follows nested side docs with a visited set so cycles end, and its test checks every ask happens once (ARCH-PURE, ARCH-CONSTRAINTS).
- `cmd/sdlc/republish.go:296-315`: a carried side doc is checked again inside the commit's prepare step against main's copy at judgment time. A concurrent edit on main therefore refuses instead of being overwritten (ARCH-ORDER).
- `cmd/sdlc/linkguard.go:578-586`: publish only carries regular files inside the issue's own `.side/` folder, and the test proves a symlink is refused (ARCH-SECURE).
- The change-code check runs on every flow, outside `activeChangeCodeGates`. This is correctly recorded in the plan's Revisions, because the quick flow skips that gate list.

**Critical findings:** none.

**Important findings**
1. `cmd/sdlc/close.go:744-748`: the census runs even under `--no-side-links` or `--force`, and its error path `die`s with no bypass. The census includes `projectWorkspaceRoots`, `DiscoverInRoots`, and `git ls-tree HEAD` in each peer project's repository. Close's own project code a few lines later (`close.go:841-850`) only warns on a discovery error ("skipping project update"). So after this change, a broken or unreadable peer root, or a peer repo with an unborn HEAD, blocks a close that used to succeed, even with `--no-project`. Fix: skip the census when `f.skip("side-links")`, and treat project discovery errors the same way close does (warn and skip projects). change-code already sends errors through its `--force` path.
2. Test coverage. Nothing tests the project branch of `issueSideLinkRefusals` (`linkguard.go:525-547`: cross-repo `censusRepo`, `r.Doc = path.Join(m.Repo, …)`). Nothing runs `runChangeCode` itself either, so the `NoSideLinks` flag, the `--force` bypass warning and the exit code (`changecode.go:69-78`) are unchecked; `TestIssueSideLinkRefusals` only calls the helper. The plan's Task 3 listed `TestChangeCodeRefusesUncommittedSideLink` and `TestCloseRefusesProjectLinkToUncommittedFile`. Neither exists, and the Revisions don't say they were dropped.

**Minor findings**
- `linkguard.go:456`: the `tree` parameter of `gitLinkState` is never used.
- `linkguard.go:470`: the path reaches `ls-files` as a pathspec, so glob characters in a link target (`*`, `[`) get pattern-matched. `:(literal)` would make the match exact.
- `issuepublish.go:410` and `republish.go:189`: every republished issue runs the full census twice, once before any effect and once inside the publication.
- `issuepublish.go:403`: the error from `snap.Require` is dropped. It is safe today only because the earlier loop already succeeded on it.
- `changecode.go:71` passes raw `f.PlansDir`, while close passes `resolvePlansDir(...)`. Worth checking that both resolve to the same directory.
- `linkguard.go:581` then `:597`: there is a gap between the regular-file `Lstat` check and the `ReadFile` that reads the carried bytes. The local-only exposure is small.

**Test coverage notes**
Passed here: `pkg/sidedoc`, `internal/linkguard`, `internal/processmanual`, and `TestIssuePublishCarriesLinkedSideDocs`, `TestIssuePublishRefusesUncommittedLink` (3 subtests), `TestIssueSideLinkRefusals`, `TestCloseRefusesIgnoredSideLink`, `TestCloseFlags_Skip`, `TestGateCatalogMatchesRegisteredFlags`. Broader `-run 'Publish|Republish|ChangeCode'` runs did not finish within 9–10 minutes in this sandbox, so I didn't check for regressions in the rest of the publish and change-code suites. The plan says the recorded suite runs on the pushed head before merge.

**Architecture**
- ARCH-DRY: pass. One grammar via `LocalLinks`; the only repetition is the double census (Minor).
- ARCH-PURE: pass. `Census` and `FormatRefusals` are pure; the git and disk calls sit in `cmd/sdlc/linkguard.go`.
- ARCH-PURPOSE: pass. All three verbs and the prose are done, and the old wording appears nowhere outside #579's own issue file.
- ARCH-MOCK: pass. Tests use real git in temp repos, the existing fixture pattern.
- ARCH-CONSTRAINTS: pass. A few git calls per link over a handful of documents.
- ARCH-SECURE: pass. Paths go to git after `--` or as `tree:path`, and the carried set is closed (Minor TOCTOU noted).
- ARCH-ORDER: pass. Judgment is re-checked inside the CAS prepare step, and `sideChanged` covers the finish.
- ARCH-FUNERAL: pass. Nothing new is durable; carried side docs archive with the issue.
- ARCH-EVOLVE: pass. No persisted format changed.

**Architectural notes for upcoming work**
The plan names a dangling-link check as a third consumer of `LocalLinks`. It can reuse `Census` with `Absent` treated as a refusal.

**Plan revision recommendations**
- Add a Revisions entry recording that `TestChangeCodeRefusesUncommittedSideLink` and `TestCloseRefusesProjectLinkToUncommittedFile` were not built (or build them), and record the close guard's bypass and error policy once Important 1 is fixed.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: gate-bypass-must-skip-the-check
    title: |
      close side-link census runs under --no-side-links/--force and dies unbypassably on project discovery or peer git errors
    detail: |
      close.go:744 computes issueSideLinkRefusals before consulting f.skip("side-links"), and any error (projectWorkspaceRoots, DiscoverInRoots walk, ls-tree HEAD in a peer project repo) hits die() with no bypass, even with --no-project. Close's own project code (close.go:841-850) only warns on the same discovery error. Fix: skip the census when the gate is skipped, and degrade project discovery errors to a warning as close already does.
  - id: new
    severity: Important
    kind: defect
    family: verb-wiring-untested
    title: |
      no tests for the project-file census branch or for runChangeCode's side-links flag and --force bypass
    detail: |
      issueSideLinkRefusals' cross-repo project path (linkguard.go:525-547) and the runChangeCode wiring (changecode.go:69-78: NoSideLinks, the force warning, exit code) are never exercised. The plan's TestChangeCodeRefusesUncommittedSideLink and TestCloseRefusesProjectLinkToUncommittedFile are absent, and the Revisions do not record dropping them.
  - id: new
    severity: Minor
    kind: defect
    family: git-pathspec-literal
    title: |
      ls-files receives the link path as a glob pathspec; use :(literal)
    detail: |
      linkguard.go:470: a target containing * or [ is pattern-matched, so its ignored-vs-uncommitted state can come out wrong.
  - id: new
    severity: Minor
    kind: defect
    family: dead-parameter
    title: |
      gitLinkState's tree parameter is unused
    detail: |
      linkguard.go:456; whether a path is in the tree is decided only by the injected inTree function.
  - id: new
    severity: Minor
    kind: defect
    family: duplicate-census
    title: |
      republished issues run publishSideDocs twice (pre-effect loop and republishOwnedOnce)
    detail: |
      issuepublish.go:410 and republish.go:189; the pre-effect check could keep its result, or only run for first publications.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"0d6831e5-f20d-4f9f-af64-5e67053c39d2","repository":"github.com/xianxu/ariadne","issue_id":"000579","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/d4b3cd1f8547cf54","reviewed_head":"493d00a8e64923ccfa9443ad6e65e8a0f4352fa4","slot":"ariadne:5","requested_at":"2026-10-24T01:04:16.612831Z"},"phase":"finished","pid":58495,"started_at":"2026-10-24T01:04:16.709856Z","finished_at":"2026-10-24T01:26:52.535559Z","session_id":"a4a68048-f2df-422c-9cfb-9bdc94ad333d","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T18:59:57-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 579 — Unified marker rule in AGENTS §1; process-artifact side docs committed; sdlc refuses links to uncommitted files |
| repo | ariadne |
| issue file | workshop/issues/000579-unified-marker-rule-in-agents-1-process-artifact-side-docs-committed-sdlc.md |
| boundary | whole-issue close |
| milestone | — |
| window | 56172a1009f553fa227d1085f8873c3d636da3ac..712b54c08a7c4ae7178a4727b3ecc1a44bf492f6 |
| command | sdlc close --issue 579 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T18:59:57-07:00 |
| verdict | SHIP |

## Review

Still reviewing. The broad test run timed out on an unrelated lock test, so I'm rerunning only the side-link tests.

Checking the publish pre-effect loop against how first publications are classified.

```verdict
verdict: SHIP
confidence: high
```

All five findings from round 1 were fixed in `712b54c08`. I accept four of them. BR-3 stays open as a Minor: the fix is right, but no test fails without it. Under its bypass, close now skips the side-link check entirely. It used to compute the check first and only then look at `f.skip`. Project discovery problems and peer repositories that cannot be judged now come back as warnings, the same way close's own project update treats them. The two missing tests now exist, and they exercise the real code paths, not mocks. The side-link tests I ran pass: `TestIssueSideLinkRefusals`, `TestIssueSideLinkRefusalsJudgeProjects`, `TestChangeCodeRefusesUncommittedSideLink`, `TestCloseRefusesIgnoredSideLink`, `TestIssuePublishCarriesLinkedSideDocs`, `TestIssuePublishRefusesUncommittedLink`, `TestGateCatalogMatchesRegisteredFlags`, plus the `pkg/sidedoc`, `linkguard` and `processmanual` packages. A broader `-run 'Close|ChangeCode|…'` sweep did not finish: it hung for 10 minutes in `TestCloseCommandsReleaseLockDuringBoundaryReview`, a repo-lock test this change doesn't touch, so I narrowed the run. Nothing blocks SHIP.

1. **Strengths**
   - `cmd/sdlc/close.go:744-757`: the bypass decision now comes before the check. The acknowledgement text ("skipping the side-link check") is emitted only in that branch, and the catalog pattern (`gatesig.go:105`) and the test both pin it.
   - `cmd/sdlc/linkguard.go:96-128`: errors are split by where they happen. An error in the issue's own repository still refuses (and says which bypass to use). An error in a peer project becomes a warning, matching close.go's existing project-update policy.
   - `TestChangeCodeRefusesUncommittedSideLink` runs `runChangeCode` itself through `expectExit`. It covers exit code 1, the `--force` acknowledgement, `--no-side-links` suppressing the check, and passing once the side doc is committed.
   - `TestIssueSideLinkRefusalsJudgeProjects` uses a real peer repository. It checks that an unborn peer warns, and that a committed peer refuses with the refusal named `peerrepo/…`.
   - `readRegular` (`linkguard.go:259`) checks that the file is the same regular file before and after the read (Lstat, read, Lstat). This closes the window where a symlink could be swapped in between the check and the carry.

2. **Critical**: none.

3. **Important**: none.

4. **Minor**
   - BR-3 is still open. The `:(literal)` fix is correct: I confirmed empirically that `ls-files` pattern-matches a `*` without it, while `ls-tree` (`headHas`, and `entryAt` behind `main.Exists`) already treats paths literally. But no test has a target with a glob character, so removing `:(literal)` would leave every test green.
   - The atlas side-link guard entry (`atlas/workflow/sdlc-binary.md:465`) doesn't mention that a project which can't be found or judged only warns. Worth one clause.

5. **Test coverage notes**
   - The BR-1 skip is pinned through the acknowledgement message. No test has a project-discovery failure under `--force`, but the restructure means the check is never computed on that path, so a test like that would add little.
   - Publish's mixed batch (first publication plus republication) still checks republished issues twice: once before any effect, then again in `republishOwnedOnce`. This is intentional and documented in the plan's Revisions.

6. **Architecture**
   - ARCH-DRY: pass. One link grammar (`LocalLinks`), and one census shared by change-code and close.
   - ARCH-PURE: pass. `Census` is pure with an injected state function, and git sits behind `sideLinkGit`.
   - ARCH-PURPOSE: pass. The class behind BR-1 (a bypass must skip the check) was checked across all three verbs. Publish has no bypass by design. Change-code under `--force` still runs the check but tolerates its errors, which matches its other gates.
   - ARCH-MOCK: pass. Tests use real git in temp repositories, as the existing sdlc fixtures do.
   - ARCH-CONSTRAINTS: pass. Work is O(links) over the issue's own documents.
   - ARCH-SECURE: pass. Paths go to git as literal pathspecs after `--`, and only regular files are carried.
   - ARCH-ORDER: pass. The check runs once before the first effect, and carried docs are re-checked against main inside the publish commit.
   - ARCH-FUNERAL: pass. The guard creates nothing durable. Carried side docs archive with the issue.
   - ARCH-EVOLVE: pass. No persisted format changed.

7. **Plan revisions**: none needed. The round-1 Revisions entry matches the code, including folding `TestCloseRefusesProjectLinkToUncommittedFile` into the shared project-branch test.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      close.go:744 checks f.skip("side-links") before calling issueSideLinkRefusals; project discovery and peer errors are now warnings (linkguard.go:105-126); TestIssueSideLinkRefusalsJudgeProjects (unborn peer warns) and TestCloseRefusesIgnoredSideLink (skip acknowledgement) would both fail on the old code.
  - id: BR-2
    disposition: addressed
    note: |
      TestChangeCodeRefusesUncommittedSideLink drives runChangeCode (exit 1, --force ack, --no-side-links, pass once committed); TestIssueSideLinkRefusalsJudgeProjects covers the peer-project branch; the folded test is recorded in the plan's Revisions. Both pass.
  - id: BR-3
    disposition: not-addressed
    note: |
      The :(literal) fix in ls-files/ls-tree is correct (checked empirically; ls-tree was already literal), but no test uses a target with a glob character, so the fix has no test that fails without it. Minor; non-blocking.
  - id: BR-4
    disposition: addressed
    note: |
      The unused tree parameter is gone from gitLinkState (linkguard.go:35); a refactor that changes no behavior, and every caller compiles and its tests pass.
  - id: BR-5
    disposition: addressed
    note: |
      judgePublishLinks (issuepublish.go:313) skips republications when no first publication shares the run; republishOwnedOnce's own pre-commit check remains the first. A refactor that changes no behavior; the publish tests pass.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"65e16fe6-f9e6-48f7-821b-8bcf2ba78ae1","repository":"github.com/xianxu/ariadne","issue_id":"000579","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/e2e4bdbb2822ab97","reviewed_head":"712b54c08a7c4ae7178a4727b3ecc1a44bf492f6","slot":"ariadne:5","requested_at":"2026-10-24T01:47:29.067738Z"},"phase":"finished","pid":27781,"started_at":"2026-10-24T01:47:29.177282Z","finished_at":"2026-10-24T01:59:57.028931Z","session_id":"5f07c44f-4b4d-4fe1-b74f-b212a5fe5555","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T19:02:36-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 579 — Unified marker rule in AGENTS §1; process-artifact side docs committed; sdlc refuses links to uncommitted files |
| repo | ariadne |
| issue file | workshop/issues/000579-unified-marker-rule-in-agents-1-process-artifact-side-docs-committed-sdlc.md |
| boundary | whole-issue close |
| milestone | — |
| window | 05173098826052d7842caa9b0aea35532118ef07..5ead6cb66681b6be51fb18fc5d5c67dfa2f5e886 |
| command | sdlc close --issue 579 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T19:02:36-07:00 |
| verdict | SHIP |

## Review

Checking the issue-card change for scope drift, then writing the verdict.

```verdict
verdict: SHIP
confidence: high
```

This window covers only what changed after the close review. That is one new regression test in `cmd/sdlc/linkguard_test.go`, plus close bookkeeping: the gate ledger, the review record, and the issue card moving to `codecomplete` and switching to the full flow with a closing Log entry. `## Spec`, `## Revisions` and `## Done when` are unchanged, so scope did not move. The BR-3 code fix was already in the code before this window started: `cmd/sdlc/linkguard.go:49` (ls-files) and `:63` (ls-tree) both pass `":(literal)"+p` after `--`. The new test is what proves it. I took `:(literal)` out of both call sites in the disposable checkout, and `TestGitLinkStateLiteralPath` failed (`x*.md: 2 <nil>, want 1`: the file came out as Ignored when it should be Uncommitted). With the fix restored, it passes along with the ChangeCode/SideLink tests. No open blocking findings remain, and the window adds no defects.

1. **Strengths**
   - The test checks the real failure. An untracked `x*.md` sits next to an ignored `xy.md`, which is exactly where a glob pathspec gets it wrong. The test uses real git with no mocks, and it goes red when the fix is removed (`linkguard_test.go:251-264`).
   - One table covers all three outcomes in one place: Uncommitted, Ignored and Absent (`x?.md`).
   - The close bookkeeping touches only the status, the flow and the Log. No spec text changed.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** none new. The test only weakly covers the `headHas` ls-tree call: no file whose name contains glob characters is committed, so taking `:(literal)` out of that call alone would not turn the test red. This is not worth a finding, since BR-3 named the ls-files call.

5. **Test coverage notes:** BR-3 now has a regression test that goes red without the fix. The existing tests for BR-1, BR-2, BR-4 and BR-5 are outside this window and were not re-run beyond the ChangeCode/SideLink subset.

6. **Architectural notes (interdiff only):**
   - ARCH-DRY: pass. The test reuses the existing helpers `claimSetRepo`, `writeRepoFile` and `execSideLinkGit`.
   - ARCH-PURE: pass. Running real git is right for an integration seam whose whole job is git pathspec behaviour.
   - ARCH-PURPOSE: pass. Both pathspec calls in `linkguard.go` are literal, so the whole class in this file is fixed, not just the one site.
   - ARCH-EVOLVE: pass. No persisted format changed, apart from the additive `actual_hours` and the flow update on the card.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      linkguard.go:49 and :63 pass ":(literal)"+p; TestGitLinkStateLiteralPath goes red (x*.md judged Ignored) with the literal prefix reverted, green with it.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"02561fd7-2540-4cae-93e4-fb870fc71527","repository":"github.com/xianxu/ariadne","issue_id":"000579","recipe":"fix-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/b9878356f3581393","reviewed_head":"5ead6cb66681b6be51fb18fc5d5c67dfa2f5e886","slot":"ariadne:5","requested_at":"2026-10-24T02:01:35.549139Z"},"phase":"finished","pid":16870,"started_at":"2026-10-24T02:01:35.699457Z","finished_at":"2026-10-24T02:02:35.811277Z","session_id":"14900503-65ea-48c9-8059-64c8c8ab590b","outcome":"returned"}] -->
