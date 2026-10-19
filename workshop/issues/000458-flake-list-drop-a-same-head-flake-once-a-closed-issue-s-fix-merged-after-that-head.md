---
id: 000458
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '6e6ff95efcab153b5bc7f2a543044513d252c3a8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T04:32:38-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# flake list: drop a same-head flake once a closed issue's fix merged after that head

## Problem

`sdlc flake list` keeps showing a same-head fail→pass flake under "not registered (file or register today)" after an issue has fixed it. The 2026-10-19 pilot in pair hit this: five of the eight listed tests had been fixed and closed already (pair#585, #589, #609, #659), and each failing head predates its fix merge. The daily view asks for a filing that would duplicate a closed issue, and the real work gets lost among them. The same happened in ariadne with TestConstructScripts on a2b56335, which predates the #448 fix (6d623285).

## Spec

- A same-head flake counts as fixed when a closed (terminal) issue's card lists the exact fully qualified test name in `flake_tests` and that issue's landing merge is not contained in the failing head. No issue Markdown establishes ownership. Containment is `git merge-base --is-ancestor`.
- Default view: fixed flakes are dropped. A verbose view (`flake list -v`) shows each one as `fixed by #N (merge <sha>)`.
- A failure on a head that already contains the fix merge is a recurrence. It stays listed, with `recurred after #N`, because that needs a new issue.
- An unreadable tracker, a missing merge commit, or ambiguous coverage (two closed issues) fails toward showing the flake, never toward hiding it.

## Done when

- Tests with fake cards and a real git history: a pre-fix flake is hidden by default and shown as fixed by #N with -v. A post-fix recurrence stays listed, marked as recurred. The ambiguous and unreadable cases stay listed.
- Run against pair's store before backfill: the five legacy pilot rows remain visible because those cards lack structured fully qualified ownership. A synthetic closed card with structured ownership hides its pre-fix row, while prose-only cards cannot hide it.

## Plan

- [x] Build a read-only snapshot of terminal issue cards; require one unambiguous card owner, a valid landing merge, and exact `flake_tests` ownership. Keep the open flake registry unchanged.
- [x] Classify each confirmed same-head failure against that snapshot and the real Git ancestry: pre-merge fixed, post-merge recurrent, or unknown. Unknown remains listed; `-v` explains fixed entries.
- [ ] Test the decision matrix with real Git history and fake cards, including unreadable/ambiguous/missing data, then verify pair's five pilot rows using the branch binary. Run `make -k` and the recorded suite, close on SHIP, and land.

Detailed design: [000458 plan](../plans/000458-flake-list-drop-a-same-head-flake-once-a-closed-issue-s-fix-merged-after-that-head-plan.md).

## Log

### 2026-10-19
- 2026-10-19: closed — Focused Go tests for card-only ownership, real Git recurrence and archived Log samples passed; make -k build and make sdlc-build passed. Read-only pair flake list showed all five legacy rows visible. Exact-head recorded suite is queued and must pass before landing; final plan row includes merge.; review verdict: SHIP
- 2026-10-19: flow upgraded quick → full — 224 added lines in code files (limit 100); an earlier round of this close already ran the full review

Filed from ariadne:2 after the #394 pilot, on the hub's approval of the proposal.

Plan evidence: pair's five current unregistered rows are the broadcast #585, couchcmd #589, launcher #609, and two diagnosticlog #659 tests. #585's exact `- Test:` line is in its Log; #589's Log also mentions #609's failed test as an unrelated later suite result, so a whole-details search would create false ownership. ARCH-PURPOSE/ARCH-PURE: classify explicit ownership and keep uncertainty visible.

TDD: the new ownership/decision test first failed at compile because the classifier did not exist, then passed after implementation. Targeted `go test ./cmd/sdlc -run '^Test(FlakeListReadsArchivedDetails|FlakeListCachesAncestry|FixedFlake|FixedIssueCoverage)' -count=1 -v` passed, including a real two-parent Git merge and an archived-detail read from a local bare origin. The read-only pair pilot with the branch binary showed all five historical rows absent by default; `-v` showed #585 at 27a12fec, #589 at 5d40e6bd, #609 at 734dd7d1, and two #659 rows at 261951a8. All unrelated rows remained. One #171 legacy card has no landing merge or published details; the decision treats it as non-evidence unless its explicit card list claims the test. See the plan Revision.

The bare-name package-context regression went red on a wrong-package issue and on a contextless name, then green. The pair pilot remained five of five after the correction. The first recorded suite queued third and was canceled safely before running; it recorded no result.

The #171 legacy exception is now pinned by two tests: an unarchived/no-landing card with no explicit test list cannot block a different issue's proven fix; the same card explicitly listing that test does block it. The second suite attempt was also canceled in the queue before it ran so the final test-bearing head receives the recorded suite.

The recorded suite passed on 1b56f757d2b7852f06c162fc2b39a97c0078eb14: 1,687 cmd/sdlc tests in 12 shards and 71 other packages, zero failures; suite log SHA-256 begins a2638162d9d4. In the hub-granted live check, ariadne:0 staged that exact head and every bin/* revision matched. The read-only pair list hid all five pilot rows by default and verbose showed #585, #589, #609 and #659 twice with their landed merges. Stage clear restored all seven binaries to main a11504c417e9c8789327362895bc88c9fad042df, all unmodified.

Close review round 1 returned REWORK: the plan mislabeled the Git-backed classifier as PURE (BR-1); README omitted the changed list view (BR-2); command tests lacked observable default/recurrence proof (BR-3); and per-row uncertainty lost its reason (BR-4). The revision corrects the table and paths, adds the README text, and drives rendering through a real merge plus explicit coverage reasons. The new test failed against the prior return shape before the implementation changed.

Round 2 disposed BR-1 through BR-4 and found BR-5: matching a package's basename could misattribute a same-named test in another directory, then hide it. Red tests reproduced this in a Problem line, a dedicated `- Test:` Log line and a nested subpackage path. The correction matches exact import or repository-relative paths; a one-word label needs a cached Git-tree proof of unique package identity at the issue's landing merge. A real Git fixture checks uniqueness and ambiguity. The read-only pair pilot still shows five fixed rows in verbose output and none in the default same-head section; the two diagnosticlog test names also occur in an older, unrelated unexplained suite failure, which remains visible by design.

The recorded suite on f2cbfe70e455169334d78e71e6c820e288089c4a passed 1,691 cmd/sdlc tests in 12 shards and 71 other packages, zero failures (log SHA-256 prefix f07a90acb6f5). Close review round 3 kept BR-5 open: a path suffix check could cross whitespace into later `.go` prose and again hide a nested-package failure. The plan now stops this incremental patching and defines a bounded, immediate path-token grammar with tests across every ownership section before more code changes.

The hub's 2026-10-19 revision supersedes that parser design: only fully qualified `flake_tests` or a dedicated fully qualified `- Test:` Log line can own a test. None of pair#585/#589/#609/#659 has either form, so the five pilot rows remain visible. This deliberately favors extra visible rows over an incorrect hide. Closed-card backfill was dropped: `flake register` refuses closed cards, and retained suite attempts age out after 30 days when later local checks prune the store. The Done when above reflects the revised live check.

Red-first correction: the prior matcher falsely owned a test from a Problem paragraph and #585's bare `- Test:` line; a real Git fixture showed a prose-only issue hiding the row. The structured matcher now rejects prose in the title, Problem, Spec, Plan, Done when, ordinary Log text and fenced text; it accepts an exact `flake_tests` value or a whole dedicated fully qualified Log line. The real-Git command test shows the prose-only row visible and the same row hidden after adding structured ownership. Targeted `TestFixed`/`TestFlakeList` tests passed, and `make -k build` plus `make sdlc-build` passed. Read-only pair pilot on the branch binary built from 658230a8440d9fea7438080f81268d0ca904bb45 with modified source showed all five legacy rows visible: broadcast #585, couchcmd #589, launcher #609, and two diagnosticlog #659 rows. The store has a 30-day retention (`suiteAttemptKeep`) pruned after future local checks, so those rows eventually leave the view without a card backfill.

The recorded suite on b64abf3422274c046cc283b613ca4a38836d2185 passed 1,688 cmd/sdlc tests in 12 shards and 71 other packages, zero failures (log SHA-256 prefix 56a36c950b88). Close review round 4 returned REWORK on BR-6: a sample `- Test:` in an unclosed fence, indented code or an HTML comment could still hide a flake. Red fixtures reproduced each form and a real-Git command suppression; the matcher now requires a top-level Log entry outside either kind of fence and HTML comments. The revised plan records the boundary.

The unchanged full-suite rerun on da8c62d2e111ef7fe86cc0c351e00d6512a5a87a passed 1,688 cmd/sdlc tests and 71 packages, zero failures (log SHA-256 prefix 7aa46b5ac1a4). The previous run on the same head timed out only in `TestConstructScripts/plain/dev-aliases.test.sh`; its exact-head targeted rerun passed 82/82 in 33.779s. The independent test-cost regression is filed as #465; #466 tracks exact subtest registration and narrowing. Close review round 5 kept BR-6 open on raw HTML samples and unmatched backticks. The hub approved a final simplification: only card `flake_tests` may establish ownership, so no Markdown parser remains in this path.

The final correction removes the detail-text loader and matcher entirely. A published-tracker fixture verifies that plain Log text, raw HTML and an unmatched backtick never claim a test; the card-only and real-Git command fixtures verify that an exact structured owner still hides a pre-fix row and labels recurrence. Focused tests passed. The five pair legacy rows remain visible and age out under the 30-day attempt retention; no closed-card backfill is planned.
