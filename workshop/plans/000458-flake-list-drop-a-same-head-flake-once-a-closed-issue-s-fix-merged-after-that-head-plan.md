---
collaboration-mode: discussion
---

# Flake list after landed fixes: implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution. Use TDD and the SDLC close boundary for fresh review.

**Goal:** Stop asking the TL to refile a confirmed same-head flake already fixed by a uniquely identified, landed issue, while making recurrence and uncertainty visible.

**Architecture:** The open `flake_tests` registry remains the authority for landing policy. Only `flake list` gets a read-only catalogue of terminal cards plus published details, then an attribution decision per confirmed failure. One pinned main tree yields issue prose; the tracker yields card status, `flake_tests` and `completion.landed_commit`; real Git ancestry distinguishes a pre-fix head from a recurrence. Every missing or ambiguous input preserves the row.

**Tech stack:** Go, existing `tracker.Records`, `gitx.TrunkView.FilesSelected`, `flakes` core, Git ancestry, `cmd/sdlc` test fixtures.

## Core concepts

| Entity | Kind | Lives in | Status |
|---|---|---|---|
| `fixedFlakeIssue` evidence value | PURE | `cmd/sdlc/flake_fixed.go` | new |
| `fixedIssueCovers` ownership matcher | PURE | `cmd/sdlc/flake_fixed.go` | new |
| `classifyFixedFlake` attribution orchestration | INTEGRATION | `cmd/sdlc/flake_fixed.go` | new |

`fixedFlakeIssue` carries ID, landing merge and exact covered tests from a terminal issue. `fixedIssueCovers` matches explicit card or issue-detail ownership without I/O. `classifyFixedFlake` takes the candidates for one test, invokes injected Git ancestry, and returns fixed, recurrence or unknown with a coverage reason; it is integration logic, not a pure entity. Only exactly one eligible issue can justify hiding. It cannot guess when cards/details/commits are unreadable or multiple closed issues claim the test. No persisted data is created (ARCH-FUNERAL).

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| Published fixed-issue catalogue | `cmd/sdlc/flake_fixed.go` | new | one fresh tracker read and one pinned-main active/archived details snapshot |
| Git containment | `cmd/sdlc/flake_fixed.go` | new | `git merge-base --is-ancestor` for each unique (merge, failing head), cached per list run |
| CLI list rendering | `cmd/sdlc/flake_cmd.go` | modified | `-v`, existing test-stats and open registry seams |

The catalogue is a read-only addition to `flakeDeps`, leaving registration and landing policy alone (ARCH-DRY/PURPOSE). The main snapshot uses `FilesSelected` to batch issue detail reads, selecting only issue Markdown at the configured active and archived issue paths. No working-tree archive is read: a slot branch can lag main or be dirty. Join details by canonical numeric issue ID; duplicate or malformed detail evidence stays unknown. A card with terminal status but no trustworthy landing merge does not prove a fix.

**Ownership rule:** A card's exact `flake_tests` name covers a test. For older cards without that field, match the fully qualified `package.TestName`, or an exact top-level `TestName` token with the repository-relative package path on the same line, in title, Problem, Spec, Done when or Plan; a dedicated `- Test: TestName` Log line with that context also counts (the #585 format). A one-word package label counts only if the issue's landing tree contains that expected Go package and no other Go package with the same final directory name (the #659 format). A path to a different same-basename package or a nested subpackage is never a one-word label. Ordinary Log references to other suites do not claim tests (the #589 mention of #609 is the counterexample). Whole tokens avoid `TestA` matching `TestAB`; subtests resolve to their top-level name, consistent with `flakes.SplitTest`. A bare name shared by two packages in the confirmed set is ambiguous even with textual context. If two terminal issues cover the test, keep the row visible even when one appears fixed. `flake_tests` is not privileged over explicit details when checking ambiguity.

**Decision:** For a confirmed failure at H and exactly one covering terminal issue with merge M, verify M is a present commit reachable from current main. If `git merge-base --is-ancestor M H` is true, show `recurred after #N`. If false, default view hides that one row and `-v` prints `fixed by #N (merge <sha>)`. Any Git error other than the documented non-ancestor exit, an absent M, stale/unreadable tracker, unreadable/ambiguous details, or ambiguous coverage leaves the row under not registered with a coverage reason. The `unexplained` section is unchanged. A terminal issue without a merge, including an abandoned issue, never hides. No mutation occurs (ARCH-EVOLVE).

**Cost:** Drivers are A retained suite attempts (30-day store), D terminal issue details, U distinct confirmed test names, C unique `(merge, failing head)` pairs, P unique `(merge, package)` pairs needing one-word-label proof, and F tracked files in a queried merge tree. Existing attempt scan is O(A); detail coverage is O(D×U), Git ancestry is O(C), and historical package proof is O(P×F), never one Git call per card × attempt. `FilesSelected` batches each of the two published detail trees; ancestry and package proof are cached per run. Counted N/2N fixtures assert C ancestry calls and P package-tree reads, independent of repeated attempts on the same head.

## Revisions

### 2026-10-19 — dependency direction and legacy card

The pure ownership/decision code lives in `cmd/sdlc/flake_fixed.go` because `internal/issue` already imports `internal/flakes`; placing the document matcher in `flakes` would create a package cycle. The measured live pair read found one legacy terminal card (#171) with neither archived details nor a landing merge. Such a card cannot prove a fix and does not block an unrelated evidenced fix unless its `flake_tests` explicitly claims the test. A terminal card with a landing merge but unreadable details still makes coverage unknown. The cost statement now exposes U rather than hiding the D×U scan behind O(D).

### 2026-10-19 — package context for legacy bare names

A bare Go test function is not globally unique. The first implementation matched `TestName` in an issue whose text named a different package, which could hide an unrelated failure. A red test pinned that case; a legacy bare name now needs its package on the same ownership line. All five pair pilot rows still qualify. The first recorded suite was queued (third in line) and canceled through `sdlc local-check cancel` before it acquired a slot; the corrected pushed head gets the suite.

### 2026-10-19 — close review: IO boundary, command proof and uncertainty

The earlier PURE declaration for the Git-backed classifier was inaccurate. The table now names the pure ownership matcher and integration classifier as implemented. A command test drives a real two-parent merge through default and verbose rendering, including recurrence, and checks that ambiguous ownership stays visible with its reason. Per-row unreadable details, ambiguous ownership, missing merge and Git uncertainty carry a coverage reason into the list. README now describes the changed daily view. The implementation paths below are corrected from the abandoned `internal/flakes` location.

### 2026-10-19 — close review: historical package identity

Round 2 found that `path.Base(pkg)` can make `pkg/launcher.TestX` appear to cover `cmd/internal/launcher.TestX`. Exact import and repository-relative package paths now establish textual ownership; an issue's one-word package label requires uniqueness among Go package directories at its landing merge, queried read-only and cached by merge/package. This preserves #659's legacy table only when its `diagnosticlog` label is actually unique, and keeps ambiguous or unreadable evidence visible. Red fixtures cover same-basename paths in substantive sections and dedicated Log lines, plus a nested subpackage; a real Git tree proves the unique/ambiguous outcomes. The extra file-tree cost is named above.

## Tasks

### 1. Red coverage and decision tests

**Files:** `cmd/sdlc/flake_fixed.go`, `cmd/sdlc/flake_fixed_test.go`, `cmd/sdlc/flake_cmd_test.go`.

- [x] Add matcher and decision table tests for exact token versus substring, substantive section versus incidental Log mention, dedicated `- Test:` legacy evidence, two claimants, missing merge, and bare-name package collision. Run targeted test; confirm red.
- [x] Add command tests using a small real Git history: failing H before M hides by default and appears with `-v`; a later H containing M is marked recurrence. Missing merge object, unreadable tracker/details and ambiguous cards stay visible. Run targeted test; confirm red.

### 2. Catalogue and classifier

**Files:** `cmd/sdlc/flake_fixed.go`, `cmd/sdlc/flake_cmd.go`, relevant tests.

- [x] Implement pure exact-test ownership matching and the three-way integration decision; rerun targeted tests.
- [x] Load terminal card evidence plus one pinned published details snapshot; preserve unknown evidence explicitly. Add real tracker/main fixture tests for active/archive, missing and malformed paths; rerun targeted tests.
- [x] Add the cached, exact Git ancestry seam; render default and verbose list. Add counted scaling test and rerun command tests.

### 3. Live evidence and close

**Files:** `cmd/sdlc/helptext/flake*.md` if present, `atlas/workflow/*.md` as needed, issue Log.

- [x] Build the branch binary and run `flake list` and `flake list -v` read-only from pair's checkout. Verify the five pilot rows disappear by default and each is attributable in verbose output; verify other rows still show. Record exact commands, branch SHA and output evidence in the issue Log.
- [x] Run `make -k`, push the exact head, then `sdlc local-check run --name suite`; record the suite result and any scope/atlas changes.
- [ ] Stage for a short live check only with hub grant, close on SHIP and merge through the landing line. Report PR and real merge SHA to hub.
