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
flow: {kind: quick, provenance: inferred, spec: "1e0dcaea", done: "6ff1911f"}
---

# flake list: drop a same-head flake once a closed issue's fix merged after that head

## Problem

`sdlc flake list` keeps showing a same-head fail→pass flake under "not registered (file or register today)" after an issue has fixed it. The 2026-10-19 pilot in pair hit this: five of the eight listed tests had been fixed and closed already (pair#585, #589, #609, #659), and each failing head predates its fix merge. The daily view asks for a filing that would duplicate a closed issue, and the real work gets lost among them. The same happened in ariadne with TestConstructScripts on a2b56335, which predates the #448 fix (6d623285).

## Spec

- A same-head flake counts as fixed when a closed (terminal) issue covers its test and that issue's landing merge is not contained in the failing head. To cover a test, the issue lists it in `flake_tests`, or names it as an exact test in its details. Containment is `git merge-base --is-ancestor`.
- Default view: fixed flakes are dropped. A verbose view (`flake list -v`) shows each one as `fixed by #N (merge <sha>)`.
- A failure on a head that already contains the fix merge is a recurrence. It stays listed, with `recurred after #N`, because that needs a new issue.
- An unreadable tracker, a missing merge commit, or ambiguous coverage (two closed issues) fails toward showing the flake, never toward hiding it.

## Done when

- Tests with fake cards and a real git history: a pre-fix flake is hidden by default and shown as fixed by #N with -v. A post-fix recurrence stays listed, marked as recurred. The ambiguous and unreadable cases stay listed.
- Run against pair's store, `flake list` no longer lists the five pilot tests fixed by pair#585, #589, #609 and #659.

## Plan

- [x] Build a read-only snapshot of terminal issue cards and their published active/archived details; require one unambiguous issue, a valid landing merge, and exact test ownership. Keep the open flake registry unchanged.
- [x] Classify each confirmed same-head failure against that snapshot and the real Git ancestry: pre-merge fixed, post-merge recurrent, or unknown. Unknown remains listed; `-v` explains fixed entries.
- [ ] Test the decision matrix with real Git history and fake cards, including unreadable/ambiguous/missing data, then verify pair's five pilot rows using the branch binary. Run `make -k` and the recorded suite, close on SHIP, and land.

Detailed design: [000458 plan](../plans/000458-flake-list-drop-a-same-head-flake-once-a-closed-issue-s-fix-merged-after-that-head-plan.md).

## Log

### 2026-10-19

Filed from ariadne:2 after the #394 pilot, on the hub's approval of the proposal.

Plan evidence: pair's five current unregistered rows are the broadcast #585, couchcmd #589, launcher #609, and two diagnosticlog #659 tests. #585's exact `- Test:` line is in its Log; #589's Log also mentions #609's failed test as an unrelated later suite result, so a whole-details search would create false ownership. ARCH-PURPOSE/ARCH-PURE: classify explicit ownership and keep uncertainty visible.

TDD: the new ownership/decision test first failed at compile because the classifier did not exist, then passed after implementation. Targeted `go test ./cmd/sdlc -run '^Test(FlakeListReadsArchivedDetails|FlakeListCachesAncestry|FixedFlake|FixedIssueCoverage)' -count=1 -v` passed, including a real two-parent Git merge and an archived-detail read from a local bare origin. The read-only pair pilot with the branch binary showed all five historical rows absent by default; `-v` showed #585 at 27a12fec, #589 at 5d40e6bd, #609 at 734dd7d1, and two #659 rows at 261951a8. All unrelated rows remained. One #171 legacy card has no landing merge or published details; the decision treats it as non-evidence unless its explicit card list claims the test. See the plan Revision.

The bare-name package-context regression went red on a wrong-package issue and on a contextless name, then green. The pair pilot remained five of five after the correction. The first recorded suite queued third and was canceled safely before running; it recorded no result.

The #171 legacy exception is now pinned by two tests: an unarchived/no-landing card with no explicit test list cannot block a different issue's proven fix; the same card explicitly listing that test does block it. The second suite attempt was also canceled in the queue before it ran so the final test-bearing head receives the recorded suite.

The recorded suite passed on 1b56f757d2b7852f06c162fc2b39a97c0078eb14: 1,687 cmd/sdlc tests in 12 shards and 71 other packages, zero failures; suite log SHA-256 begins a2638162d9d4. In the hub-granted live check, ariadne:0 staged that exact head and every bin/* revision matched. The read-only pair list hid all five pilot rows by default and verbose showed #585, #589, #609 and #659 twice with their landed merges. Stage clear restored all seven binaries to main a11504c417e9c8789327362895bc88c9fad042df, all unmodified.

Close review round 1 returned REWORK: the plan mislabeled the Git-backed classifier as PURE (BR-1); README omitted the changed list view (BR-2); command tests lacked observable default/recurrence proof (BR-3); and per-row uncertainty lost its reason (BR-4). The revision corrects the table and paths, adds the README text, and drives rendering through a real merge plus explicit coverage reasons. The new test failed against the prior return shape before the implementation changed.

Round 2 disposed BR-1 through BR-4 and found BR-5: matching a package's basename could misattribute a same-named test in another directory, then hide it. Red tests reproduced this in a Problem line, a dedicated `- Test:` Log line and a nested subpackage path. The correction matches exact import or repository-relative paths; a one-word label needs a cached Git-tree proof of unique package identity at the issue's landing merge. A real Git fixture checks uniqueness and ambiguity. The read-only pair pilot still shows five fixed rows in verbose output and none in the default same-head section; the two diagnosticlog test names also occur in an older, unrelated unexplained suite failure, which remains visible by design.
