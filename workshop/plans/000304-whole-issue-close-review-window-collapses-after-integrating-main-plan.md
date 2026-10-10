# Branch-patch review windows, publish anchor and close binding — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every main-relative review mechanism is defined on the branch patch
`diff(merge-base(main, HEAD), HEAD)`. Merging or rebasing main never widens a review
window, never forces a re-close and never orphans a close binding. A later review reads
only the interdiff since the last reviewed patch.

**Architecture:** One git primitive, *the reviewed patch rebased onto today's main*,
serves every consumer. A finalized boundary review records the head it read (`H_r`) in
the gate ledger. Its base is derivable as `merge-base(main, H_r)`, because main only
grows. `git merge-tree --write-tree --merge-base=B_r B_now H_r` builds the tree "today's
main plus the reviewed patch", `T'`. Then:
- The milestone window is `diff(T', HEAD)`.
- The publish gate passes when that diff has no code surface.
- A rebased close commit is recognised by its stable patch-id.

A per-issue ref pins `H_r`'s objects across a rebase. The ledger, which the binary
writes, replaces the hand-pasted trailer as the boundary source (#197). Milestone-close
commits its own evidence.

**Tech Stack:** Go (`cmd/sdlc`), git ≥ 2.40 (`merge-tree --write-tree --merge-base`,
`patch-id --stable`), the existing `gatestate` ledger, the tracker card's completion
binding.

**Issue:** ariadne#304 (umbrella for ariadne-robustness-1 requirement 5). It absorbs
#197 (archived wontfix into this issue), evidence A2, A3 and A4, and B3's close-binding
half (evidence pensive `workshop/pensive/2026-10-09-01-pensive-sdlc-robustness-evidence.md`,
Part 1 under #304/#197, Part 2 §A). #183 (`--fixed-then-ship`) builds on the interdiff
primitive delivered here.

---

## Background: what is wrong today (verified in source, 2026-10-09)

| # | Mechanism | Today | Failure after integrating main |
|---|-----------|-------|--------------------------------|
| 1 | Milestone window base | `boundaryWindowBase` → `previousReviewBoundary` (`milestoneclose.go:312,356`): newest commit touching the issue file whose message has `^Review-Verdict:` | After a **merge** of main, `git diff <prior boundary> HEAD` includes every file main changed (pair#410: 93 foreign commits, ~11.8k lines). A missed paste silently widens the window to the branch start (#197). |
| 2 | Whole-issue window | `merge-base(main, HEAD)` | Correct; the reviewer is handed a tree diff. But `close` prints `commit window: X → Y` (`close.go:560`), which reads as "only the post-merge commits" (pair#410, ~40 min lost reading sdlc source). |
| 3 | Milestone evidence | `milestone-close` prints trailers "paste into commit message" (`milestoneclose.go:449`) | A missed paste breaks (1) and makes `close` refuse "milestones Mx lack Review-Verdict trailer" (`close.go:1895`). The only way past is `--no-verdict`. |
| 4 | Publish gate (A4) | `validatePublishAnchors` (`publishgate.go:131`): `revCount(anchor..HEAD) > 0` with code paths → refuse | A merge of main after close puts main's commits after the anchor, so the gate refuses "N commit(s) landed after sdlc close" and forces a full re-close (≥ 8 occurrences, 10–60 min each). |
| 5 | Close binding (B3) | `ownedCompletions` (`trackercompletion.go:36`) owns a card iff `EvidenceCommit` is an ancestor of the head | A **rebase** after close rewrites the evidence commit, so the card is silently *not owned*. `runPublishGate` then sees "no codecomplete issues — nothing to verify" and lands with no check. The card stays `codecomplete` forever. This **fails open**. |
| 6 | `Review-Window:` trailer (A2) | `reviewTrailers` renders `base..head` SHAs | After a rebase they name pre-rebase commits. Nothing mechanical reads them except (1). |

## Design decisions

- **D1. The reviewed patch's identity is its head commit `H_r`.** The base is
  derived, not stored: `B_r = merge-base(main, H_r)`. That holds because main
  only grows, and the TL dispatch and the project file assume a non-rewritten
  main. One field (`Round.Reviewed`) is the whole identity. The card's existing
  `reviewed_head` already carries it, so the card schema doesn't change
  (ARCH-DRY: no second copy of a derivable value).
- **D2. The rebased reviewed tree** is `T' = merge-tree(base=B_r, B_now, H_r)`.
  It is the reviewed patch replayed onto today's main, and git does the replay,
  not us.
  - When `B_now == B_r`, `T' = tree(H_r)`, and every consumer reduces exactly to
    today's behaviour. That is the regression anchor.
  - A conflict (merge-tree exit 1) means "main rewrote lines the review read".
    The interdiff is then undefined, and each consumer fails safe in its own
    direction: windows over-cover to the full branch patch (loud warning), and
    the publish gate refuses (the resolution is unreviewed).
- **D3. The synthetic base commit.** Consumers want a commit SHA (the manifest
  validates `^{commit}`, `diff --name-only` takes commits), so `T'` is wrapped as
  `S = commit-tree T' -p B_now`.
  - Fixed author/committer identity and date make `S` deterministic for a given
    `(B_r, B_now, H_r)`. The snapshot re-check after the review then compares
    equal values.
  - ARCH-FUNERAL: `S` is unreferenced and collected by `git gc` after
    `gc.pruneExpire` (2 weeks default). There is one small commit object per
    boundary invocation, and its trees are shared with main.
- **D4. Pin ref `refs/sdlc/reviewed/<6-digit id>`** points at the newest commit a
  finalized boundary of the issue produced. That is the evidence commit when the
  binary makes one, else `H_r`. Either way it descends from `H_r`, so `H_r`
  survives a rebase.
  - Refs are shared across linked worktrees, so `sdlc move` between slots keeps
    it.
  - ARCH-FUNERAL: created and advanced at boundary finalize. Deleted when the
    card goes `done` (`completeOnCard`), when the issue is abandoned
    (`runAbandon`), and by the archive step in `merge`/`push` for legacy repos.
    It is bounded to one ref per in-flight issue.
  - If `H_r` cannot be resolved (another clone, or a ref deleted by hand), the
    consumer fails closed: windows over-cover with a warning, and the publish
    gate refuses with "re-run `sdlc close`".
- **D5. Only a finalizing round advances the boundary.** `Round.Reviewed` is
  stamped **iff** the verdict finalizes (SHIP / FIX-THEN-SHIP per
  `vocab.Verdict().IsFinalizing`) *and* the ledger decision does not block, or
  its block was waived by `--no-ledger`. REWORK, a halt, a protocol error or a
  `--no-judge` skip never stamps it, so the next window still covers
  everything unreviewed.
  - FIX-THEN-SHIP stamps the **pre-fix** head. The fixes are therefore in the
    next interdiff, which is exactly what #183 will review.
- **D6. The milestone base is the newest stamped round whose boundary is not the
  current one.** A whole-issue close keeps the full branch patch (base
  `B_now`); #183 will add its own interdiff. A pre-#304 ledger with no stamped
  round falls back to `previousReviewBoundary`, then to the branch point. That
  covers issues in flight at rollout.
  - ARCH-FUNERAL: the trailer fallback is removable once no open issue predates
    #304. This is noted in the code comment and the atlas, not filed as a
    separate issue.
- **D7. Milestone-close commits its own evidence (#197)** in tracker-era
  repositories, exactly as whole-issue close does (`closetracker.go`).
  - The commit holds the issue details, `<stem>-*` plans files (ledger,
    sidecar) and project edits, through the same `closeEvidence` set and a
    temporary index, so staged unrelated work is untouched (#206's class guard
    `TestGitCommitsCarryTheirPathspec` applies).
  - Its subject is `#N Mx: close`, with the verdict trailers in the body, so the
    existing `milestoneHasVerdictCommit` gate is satisfied by the binary's own
    commit. No second verdict source is needed.
  - Legacy repositories keep the paste protocol, as their whole-issue close
    does. Their window still comes from the ledger (D6), which they commit with
    their close.
- **D8. A rebased close is recognised by patch-id (B3).** `ownedCompletions`
  also owns a codecomplete card when the evidence commit is not an ancestor of
  the head, but a commit in `merge-base(main, head)..head` has the same
  `git patch-id --stable` as the evidence commit.
  - The evidence commit touches only details, plans files and project records,
    so its patch-id survives any rebase that doesn't conflict on those files.
  - A conflict there yields a different patch-id. The card is then not owned,
    and that case is made **loud**, not silent: a codecomplete card bound to
    this repository whose details file the branch patch changes, but which
    is owned by neither rule, refuses the publish ("re-run `sdlc close`").
    This closes the fail-open in Background row 5.
- **D9. Publish gate on the branch patch (A4).** For each owned completion,
  `R = H_r`: the card's `reviewed_head`, or, in legacy repos, the codecomplete
  anchor commit, which sits in HEAD's history. The gate builds `S` (D2/D3) and
  diffs `S..HEAD`, excluding nothing: the existing `publishGateHasCodeSurface`
  decides.
  - No code surface → pass, with the existing docs-only line.
  - Code surface → refuse, naming the code paths, because a commit count
    misleads once merges are involved.
  - A conflict → refuse: "main's changes conflict with the reviewed patch; the
    resolution is unreviewed".
  - A merge of main with no conflict therefore passes with no re-close.
    `quickGrewPastReview` already measures the branch patch and is unchanged.
- **D10. The printed window says what it is (A3).**
  - Whole-issue close: `branch patch vs main@<B8>: <n> issue commit(s), <f> file(s)`.
    `<n>` is `rev-list --count --first-parent --no-merges B..HEAD`, which counts
    the issue's own commits, never main's.
  - Milestone close: `interdiff since <Mprev> review (<H_r8>), rebased onto main@<B8>: <f> file(s)`.
    When `B_now == B_r` it reads `interdiff since <Mprev> review (<H_r8>): …`.
  - The `Review-Window:` trailer keeps `base..head`, but `base` is the real
    `B_now` for a whole-issue close, and `H_r`'s short SHA for a milestone. The
    synthetic `S` is never printed: it means nothing to a human and is
    collected later.
  - Help text states that the trailer records commit ids *as of review time*,
    and that a rebase leaves them historical; the ledger and the pin are the
    record (A2). Nothing mechanical parses the trailer's range.
- **Out of scope (stated so the reviewer doesn't hunt for it):**
  - #194's in-review anchor (`reviewanchor.go`): a rebase *during* a ~20-minute
    review still classifies as diverged and refuses. That is rare, it happens
    inside one invocation, and re-running is correct.
  - Active time (#270), and the freshness rule (#269/#303), which explicitly
    lands after this.
  - The `--fixed-then-ship` verb (#183).

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Round.Reviewed` | `cmd/sdlc/internal/gatestate/ledger.go` | new (field) |
| `LatestReviewed` | `cmd/sdlc/internal/gatestate/reviewed.go` | new |
| `reviewWindow` | `cmd/sdlc/reviewwindowplan.go` | new |
| `planReviewWindow` | `cmd/sdlc/reviewwindowplan.go` | new |
| `formatReviewWindow` | `cmd/sdlc/reviewwindowplan.go` | new |
| `classifyPublishDelta` | `cmd/sdlc/publishgate.go` | new |
| `validatePublishAnchors` | `cmd/sdlc/publishgate.go` | modified |
| `reviewTrailers` | `cmd/sdlc/milestoneclose.go` | modified |

- **`Round.Reviewed`** is a `string` holding the full 40/64-hex `H_r`, with
  YAML key `reviewed` and `omitempty`, so pre-#304 ledgers stay byte-identical.
  It is set only by the finalize path (D5).
  - **Relationships:** 0..1 per round; N rounds per boundary.
  - **Future extensions:** #183 reads the FIX-THEN-SHIP round's `Reviewed` as
    its interdiff base.
- **`LatestReviewed(l Ledger, exclude string) (sha, boundary string, ok bool)`**
  returns the newest round with `Reviewed != ""` whose `Boundary` is neither
  `exclude` nor `BoundaryAll`. It is pure, and unit-tested on in-memory ledgers.
  - **DRY rationale:** this is the single answer to "where did review last
    stop?" Today it is re-derived by trailer grep.
- **`reviewWindow`** is a struct `{Base, Head string; Kind windowKind;
  MainBase, Reviewed, PriorBoundary string; Files int; IssueCommits int;
  Fallback string}`. `Kind` is one of `branchPatch`, `interdiff` or
  `branchPatchFallback`.
  - **Relationships:** 1 per close/milestone-close invocation. It feeds the
    review manifest, the atlas gate, the trailer, the printed line and the
    sidecar meta (one source, which keeps today's #58 ARCH-DRY property that
    the atlas gate and the review cover the same diff).
- **`planReviewWindow(in windowFacts) reviewWindow`** is the pure decision
  over gathered facts `{MainBase, Head, Reviewed, PriorBoundary,
  ReviewedResolvable bool, Rebased string /*S or ""*/, Conflict bool,
  LegacyTrailerBase string, BranchStart string}`. It is total, and every
  `windowFacts` combination maps to exactly one `Kind` (table-tested).
- **`formatReviewWindow(w reviewWindow) string`** renders the D10 line. It is
  pure, and it is kept separate from the refusal vocabulary (the
  `gatesig`/#172 constraint).
- **`classifyPublishDelta(d publishDelta) (pass bool, msg string)`** is pure.
  Its input is `{Anchor, MainBase string; Conflict bool; Paths []string;
  Unresolvable bool}`, and it reuses `publishGateHasCodeSurface`.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `rebasedReviewedBase` | `cmd/sdlc/internal/gitx/rebasedpatch.go` | new | `git merge-base`, `merge-tree`, `commit-tree` |
| `stablePatchID` | `cmd/sdlc/internal/gitx/rebasedpatch.go` | new | `git diff-tree -p` piped to `git patch-id --stable` |
| `pinReviewed` / `unpinReviewed` | `cmd/sdlc/reviewpin.go` | new | `git update-ref` |
| `gatherWindowFacts` | `cmd/sdlc/reviewwindowplan.go` | new | ledger read + `rebasedReviewedBase` |
| `commitMilestoneEvidence` | `cmd/sdlc/closetracker.go` | new | temp index + `commit-tree` + branch CAS (reuses `closeEvidence`, `commitIndexOnto`) |
| `ownedCompletions` | `cmd/sdlc/trackercompletion.go` | modified | adds the D8 patch-id rule and the loud unowned refusal |
| `boundaryWindowBase` / `resolveReviewWindow` | `cmd/sdlc/milestoneclose.go` | modified | become thin wrappers over `gatherWindowFacts` + `planReviewWindow` |

- **`rebasedReviewedBase(run, mainRef, reviewed string) (s string, conflict bool, err error)`**
  computes `B_r = merge-base(main, reviewed)` and `B_now = merge-base(main, HEAD)`.
  - It runs `merge-tree --write-tree --merge-base=B_r B_now reviewed`. Exit 1
    means conflict and returns `conflict=true` with no error; any other non-zero
    exit is an error.
  - It then runs `commit-tree` with `GIT_{AUTHOR,COMMITTER}_{NAME,EMAIL,DATE}`
    fixed to `sdlc`/`sdlc@invalid`/`@0 +0000`, and refuses a git older than
    2.40 with a message naming the requirement.
  - **Injected into:** `gatherWindowFacts` and the publish gate, through the
    package `run` shim already used by `gitx` tests. Tests use **real git** in
    hermetic temp repos (`hermeticRepo`), not mocks: the behaviour under test
    *is* git's merge.
- **`pinReviewed(id, commit)` / `unpinReviewed(id)`** run
  `git update-ref refs/sdlc/reviewed/<id> <commit>` / `update-ref -d`. Unpinning
  a missing ref is a no-op, and any failure is a warning, never a refusal: the
  pin is retention, not authority. That follows the lessons rule that a
  best-effort projection must not add a refusal.

## Chunk 1: M1 — windows on the branch patch, ledger-sourced boundary, milestone evidence commit

### Task 1: `Round.Reviewed` + `LatestReviewed` (pure)

**Files:** Modify `cmd/sdlc/internal/gatestate/ledger.go` (the `Round` struct).
Create `cmd/sdlc/internal/gatestate/reviewed.go` and `reviewed_test.go`.

- [ ] **Step 1: Write failing tests** in `reviewed_test.go`:
  - an empty ledger gives `ok=false`;
  - rounds `M1(reviewed=a)`, `M1(REWORK, no reviewed)`, `M2(reviewed=b)` with
    exclude `M3` give `b, "M2"`;
  - exclude `M2` gives `a, "M1"`;
  - a `BoundaryAll` seed round with `reviewed` set (impossible in production,
    but asserting it is ignored) still gives the non-seed round;
  - a YAML round-trip of a round with no `Reviewed` is byte-identical to the
    pre-change fixture in `parse_test.go`.
- [ ] **Step 2:** `go test ./cmd/sdlc/internal/gatestate/ -run Reviewed` — expect FAIL (undefined).
- [ ] **Step 3: Implement**
  ```go
  // ledger.go, in Round, after Recipe:
  // Reviewed is the head commit a FINALIZED boundary review read (#304, #197): the
  // identity of the reviewed branch patch, whose base is merge-base(main, Reviewed).
  // Set only when the round finalized; REWORK/halt/protocol-error rounds leave it
  // empty so they never advance the boundary.
  Reviewed string `yaml:"reviewed,omitempty"`
  ```
  ```go
  // reviewed.go
  // LatestReviewed returns where review last stopped outside `exclude`: the newest
  // round carrying a Reviewed head whose boundary is neither exclude nor BoundaryAll.
  func LatestReviewed(l Ledger, exclude string) (sha, boundary string, ok bool) {
  	for i := len(l.Rounds) - 1; i >= 0; i-- {
  		r := l.Rounds[i]
  		if r.Reviewed == "" || r.Boundary == exclude || r.Boundary == BoundaryAll {
  			continue
  		}
  		return r.Reviewed, r.Boundary, true
  	}
  	return "", "", false
  }
  ```
- [ ] **Step 4:** Re-run and expect PASS. Run `go test ./cmd/sdlc/internal/gatestate/`
  and expect it to stay green.
- [ ] **Step 5:** `git commit -m '#304 M1: gatestate: Round.Reviewed + LatestReviewed' -- cmd/sdlc/internal/gatestate/`

### Task 2: stamp `Reviewed` only on a finalizing round (D5)

**Files:** Modify `cmd/sdlc/boundaryledger.go` (`persistBoundaryRound`) and
`cmd/sdlc/close.go` (`finalizeBoundaryReview`, ~1355). Test in
`cmd/sdlc/close_ledger_test.go`.

The round is persisted *before* the verdict switch, so the stamp decision moves
into the persist call. Pass `finalizes bool` as `closeVerdictOutcome(review.Verdict) == closeFinalize`
and `waived bool` as `f.skip("ledger")`. Inside `stampAndPersist`'s
`Write` closure, after the decision `d` is known, set
`out.Rounds[last].Reviewed = review.Head` iff `finalizes && (!d.Block || waived)`
and `isResolvedSHA(review.Head)`.

- [ ] **Step 1: Failing table test** `TestPersistBoundaryRound_StampsReviewedOnlyWhenFinalizing`
  drives `persistBoundaryRound` with a temp plans dir:

  | Case | Ledger state | Expected `reviewed` |
  |------|--------------|---------------------|
  | SHIP | no open findings | set to head |
  | FIX-THEN-SHIP | — | set |
  | REWORK | — | empty |
  | SHIP | open Important | empty |
  | SHIP | open Important, waived | set |
  | protocol error | — | empty |
  | head literal `"HEAD"` | — | empty |

  Each row asserts the YAML on disk.
- [ ] **Step 2:** Run and expect FAIL.
- [ ] **Step 3:** Implement as described. Keep the `--no-judge` paths
  (`runMilestoneClose` and close's skip) **not** calling persist, as today.
  Add a comment there that this is why they never advance the boundary.
- [ ] **Step 4:** PASS. Mutation check: delete the `!d.Block` term, and the
  "open Important" row must fail. Restore it.
- [ ] **Step 5:** Commit `#304 M1: boundary ledger records the reviewed head on finalize`.

### Task 3: `rebasedReviewedBase` + `stablePatchID` (git seam)

**Files:** Create `cmd/sdlc/internal/gitx/rebasedpatch.go` and
`rebasedpatch_test.go` (real git in `t.TempDir()`, using the same init helpers
as `window_test.go`).

Fixture builder `patchRepo(t)`:
1. main: `a.txt`, `shared.txt`.
2. branch: change `a.txt` (M1), commit `H1`.
3. main advances: add `foreign.txt` and edit the *end* of `shared.txt`
   (non-overlapping).
4. Return the SHAs.

- [ ] **Step 1: Failing tests:**
  - (a) no integration (`B_now == B_r`): `tree(S) == tree(H1)` and
    `parent(S) == B_r`;
  - (b) after `git merge main` on the branch: `git diff --name-only S HEAD` is
    empty;
  - (c) after `git rebase main`: also empty;
  - (d) a branch commit `H2` touching `b.txt` after the merge: the diff is
    exactly `[b.txt]`;
  - (e) main edits the same line of `a.txt` that M1 changed, then the branch is
    merged with a resolution: `conflict == true`, `err == nil`;
  - (f) a reviewed SHA not resolvable: error mentions "reviewed head";
  - (g) determinism: two calls return the same `S`;
  - (h) `stablePatchID(c) == stablePatchID(rebased copy of c)`, and the IDs
    differ for different changes.
- [ ] **Step 2:** Run `go test ./cmd/sdlc/internal/gitx/ -run RebasedReviewed` and expect FAIL.
- [ ] **Step 3: Implement**
  ```go
  // RebasedReviewedBase returns S: a deterministic, unreferenced commit whose tree is
  // today's merge-base(main, HEAD) plus the patch `reviewed` carried against its own
  // merge-base with main (#304). diff(S, HEAD) is the interdiff since that review.
  // conflict=true (no error) when main rewrote lines the reviewed patch touched.
  func RebasedReviewedBase(mainRef, reviewed string) (s string, conflict bool, err error)
  ```
  - `B_r := merge-base(mainRef, reviewed)` and `B_now := merge-base(mainRef, HEAD)`.
    An empty result is an error, never a degraded `""`, per the lesson about
    a silent empty answer.
  - Run `merge-tree --write-tree --merge-base=<B_r> <B_now> <reviewed>`. Exit 1
    means conflict. On exit 0, the first stdout line is the tree.
  - Run `commit-tree <tree> -p <B_now> -m "sdlc: reviewed patch rebased (#304)"`
    with the fixed identity env, via a new `runEnv` shim beside `run`.
  - `StablePatchID(commit)` runs `git diff-tree -p <commit>` piped to
    `git patch-id --stable`, taking the first field. An empty output (for an
    empty commit) returns `""` with no error. Callers treat `""` as "no match",
    never as equal.
- [ ] **Step 4:** PASS. Mutation: drop `--merge-base=` (it lets git pick the
  base), and case (b) must still pass while case (d) must not regress. If
  nothing fails, add case (i): main reverts an earlier main commit (a criss-cross
  base), which shows why the explicit base is needed.
- [ ] **Step 5:** Commit `#304 M1: gitx: rebased reviewed base + stable patch-id`.

### Task 4: the window planner (pure + thin shell), replacing `boundaryWindowBase`'s milestone branch

**Files:** Create `cmd/sdlc/reviewwindowplan.go` and `reviewwindowplan_test.go`.
Modify `cmd/sdlc/milestoneclose.go:278-331` (`resolveReviewWindow`,
`boundaryWindowBase`).

`planReviewWindow` rules, in order:
1. `milestone == ""`: `branchPatch`, with base `MainBase`, else `BranchStart`
   (on-main fallback, as today).
2. `Reviewed` found, resolvable and not conflicting: `interdiff`, base `S`.
3. `Reviewed` found but unresolvable or conflicting: `branchPatchFallback`,
   base `MainBase`. `Fallback` names why ("main rewrote lines M1's review read",
   or "reviewed head <x> is not in this repository").
4. No stamped round, but a `LegacyTrailerBase` (pre-#304 ledger): `interdiff`
   with base = trailer commit, and `Fallback = "pre-#304 boundary from the
   Review-Verdict trailer"`.
5. Otherwise: `branchPatch` (first milestone).

`gatherWindowFacts` reads the ledger with `readBoundaryGateLedger`. An
unreadable ledger is treated as a warning plus rule 5's over-cover, never a
refusal (the review is still worth running, matching `boundaryPriorFindings`).
It calls `gitx.RebasedReviewedBase(gitx.TrunkRef(), sha)` only when a stamped
round exists, and counts `IssueCommits` / `Files`.

`boundaryWindowBase(issueStr, milestone, issuePath)` keeps its signature and
callers (the close atlas gate at `close.go:551`, `quickGrewPastReview`) and
returns `planReviewWindow(...).Base`. `resolveReviewWindow` additionally returns
the `reviewWindow`, and its three call sites take it.

- [ ] **Step 1: Failing pure table test** `TestPlanReviewWindow` covers each
  rule. It includes the regression row (M2 with reviewed = H1 and no
  integration gives `interdiff` with base S, where the facts say
  `tree(S)==tree(H1)`) and asserts `Kind`, `Base` and `Fallback`.
- [ ] **Step 2: Failing fixture test, the issue's Done-when:**
  `TestMilestoneWindow_ExcludesIntegratedMain` in `milestonewindow_test.go`.
  - Setup: issue commits (M1 file `m1.go`); write a ledger round
    `{Boundary: M1, Reviewed: H1}`; `git merge` a main with `foreign.go`; then
    commit `m2.go`.
  - Assert the M2 window's `git diff --name-only base HEAD` is exactly
    `[m2.go]`.
  - Assert the whole-issue window gives `[m1.go m2.go]` plus the issue/plans
    files, and **no** `foreign.go`.
  - Repeat the same fixture with `git rebase` instead of merge (sub-test).
- [ ] **Step 3:** Run both and expect FAIL.
- [ ] **Step 4:** Implement. Update the doc comments on `boundaryWindowBase` and
  `previousReviewBoundary`: the latter becomes "legacy fallback (D6); remove
  once no open issue predates #304".
- [ ] **Step 5:** PASS. Run all existing `TestBoundaryWindowBase_*` and expect
  them to stay green: they have no stamped rounds, so rule 4 or 5 applies,
  which is identical to today. Mutation: make rule 2 return `MainBase`, and the
  fixture must fail.
- [ ] **Step 6:** Commit `#304 M1: review windows planned on the branch patch`.

### Task 5: printed window + trailer (A3, A2)

**Files:** `cmd/sdlc/close.go:556-563`, `cmd/sdlc/milestoneclose.go`
(`reviewTrailers`, `resolveReviewWindow` callers), `cmd/sdlc/helptext/close.md`,
`cmd/sdlc/helptext/milestone-close.md`. Tests in `reviewwindowplan_test.go`.

- [ ] **Step 1: Failing tests** `TestFormatReviewWindow` (one golden string per
  `Kind`, exactly the D10 shapes) and `TestReviewTrailers_InterdiffNamesReviewedHead`.
  The trailer for an interdiff window is `Review-Window: <H_r8>..<head8>`, never
  the synthetic S.
- [ ] **Step 2:** FAIL.
- [ ] **Step 3:** Implement.
  - `close.go` prints `formatReviewWindow(w)` in place of `commit window: …`.
    `branchPatchFallback` prints through `cwarn` and names the reason.
  - `reviewResult` gains `WindowBase string` (the human base: `B_now` or `H_r`),
    and `reviewTrailers` renders that, not `BaseLong`.
  - Help text: one paragraph saying the window is the branch patch (or the
    interdiff since the previous milestone's review), that integrating main
    never widens it, and that `Review-Window:` records commit ids as of the
    review, so a rebase leaves them historical (the ledger and
    `refs/sdlc/reviewed/<id>` are the record).
- [ ] **Step 4:** PASS. `go test ./cmd/sdlc/ -run 'Helptext|Trailer|Window'`.
- [ ] **Step 5:** Commit `#304 M1: close prints the branch patch it reviews`.

### Task 6: milestone-close commits its own evidence; pin the reviewed head (#197, D4, D7)

**Files:**
- Create `cmd/sdlc/reviewpin.go` and `reviewpin_test.go`.
- Modify `cmd/sdlc/closetracker.go`: add `commitMilestoneEvidence`.
- Modify `cmd/sdlc/close.go` (`finalizeBoundaryReview`, the `f.Milestone != ""`
  finalize branch).
- Modify `cmd/sdlc/helptext/milestone-close.md`.
- Tests in `milestoneclose_test.go` and `tracker_e2e_test.go`.

- [ ] **Step 1: Failing tests.**
  - (a) `TestPinReviewed`: pin, re-pin to a descendant, unpin, and unpin of a
    missing ref is a no-op. A failure under a read-only `.git/refs` returns a
    warning string, not an error.
  - (b) A tracker e2e, `TestMilestoneClose_CommitsItsOwnEvidence`:
    - Stage an unrelated file. Run milestone-close with a fake SHIP reviewer
      (the existing reviewer fake used by `closereview_test.go`).
    - Assert HEAD's subject is `#N M1: close`, and that its body carries
      `Review-Verdict: SHIP`.
    - Assert it changes only the details file and `<stem>-*` plans files, that
      the unrelated file is **still staged and not committed**, and that
      `refs/sdlc/reviewed/<id>` equals the new commit.
    - Assert a following `sdlc close` passes the milestone-verdict check
      **without** `--no-verdict`.
  - (c) FIX-THEN-SHIP at a milestone commits evidence immediately, and the
    ledger round's `reviewed` is the pre-commit head.
  - (d) Legacy mode (no tracker): no commit is made, trailers are printed as
    today, and the ledger is still stamped.
- [ ] **Step 2:** FAIL.
- [ ] **Step 3: Implement.**
  - `commitMilestoneEvidence(env, r, plansDir, review)`:
    - `entries := closeEvidence(env, r, plansDir)`;
    - build a temp index from HEAD plus the entries (the same body as
      `gitEvidence.Prepare`; extract the shared part into
      `evidenceIndexOnto(env, head, entries)` and keep `Prepare` calling it, per
      ARCH-DRY);
    - `commitIndexOnto(env, idx, head, "#N Mx: close\n\n"+trailers)`;
    - move the branch exactly as `gitEvidence.Apply` does: `update-ref <branch>
      <new> <head>`, a CAS that refuses if the branch moved, then
      `git reset -q -- <evidence paths>`. Only those paths' index entries follow
      the new HEAD, so other staged work stays staged. Share Apply's body as
      `applyEvidenceCommit(env, branch, commit, paths)` instead of copying it.
  - In `finalizeBoundaryReview`'s finalize branch, when `r.tracker && f.Milestone != ""`,
    call it, then call `pinReviewed(id, newCommit)`. For a whole-issue tracker
    close, call `pinReviewed(id, evidence)` after `tracker.Drive` succeeds, and
    `pinReviewed(id, review.Head)` on the FIX-THEN-SHIP deferral. In legacy
    mode, call `pinReviewed(id, review.Head)`.
  - Help text: milestone-close commits its evidence itself in tracker-era
    repositories, so there is no paste step; legacy keeps the paste.
- [ ] **Step 4:** PASS. Run `go test ./cmd/sdlc/ -run 'TestGitCommitsCarryTheirPathspec|MilestoneClose|PinReviewed'`.
  The class guard must stay green because the commit goes through a temp index
  and `commit-tree`, not `git commit`.
- [ ] **Step 5:** Mutation: drop the CAS old-value argument, and a test that
  moves the branch between the review and the commit must fail. Add that test
  if it is missing.
- [ ] **Step 6:** Commit `#304 M1: milestone-close commits its own evidence (#197)`.

### Task 7: M1 docs + boundary

- [ ] Update `atlas/workflow/gate-state.md`: the ledger's `reviewed` field is
  the boundary source. Update `atlas/workflow/sdlc-binary.md`: review windows
  are on the branch patch and the interdiff, `refs/sdlc/reviewed/<id>` and its
  lifecycle, and the legacy trailer fallback. Keep `atlas/index.md` linking.
- [ ] Add one line to #269's spec noting the milestone-window interaction (Done-when).
- [ ] Commit `#304 M1: atlas: branch-patch review windows`.
- [ ] Run `make test`, then `sdlc milestone-close --issue 304 --milestone M1`.

## Chunk 2: M2 — publish gate and close binding on the branch patch

### Task 8: publish delta classification (pure) (D9)

**Files:** `cmd/sdlc/publishgate.go` and `publishgate_test.go`.

- [ ] **Step 1: Failing table test** `TestClassifyPublishDelta`:
  - no paths: pass, with the "reviewed patch unchanged" line;
  - docs only: pass, with the docs-only line (reuse
    `formatPublishGateDocsOnly`'s vocabulary constraint);
  - code paths: refuse, naming each code path and the re-close command;
  - conflict: refuse, with "main's changes conflict with the reviewed patch —
    the resolution is unreviewed";
  - unresolvable: refuse, with "reviewed head <x> is not in this repository
    (rebased in another clone?)".

  Every refusal row asserts its own distinguishing token, per lesson BR-8, so
  that no row passes vacuously.
- [ ] **Step 2:** FAIL. **Step 3:** Implement `publishDelta` + `classifyPublishDelta`. **Step 4:** PASS.
- [ ] **Step 5:** Commit `#304 M2: publish gate: pure branch-patch delta classifier`.

### Task 9: `validatePublishAnchors` on the rebased reviewed patch (A4)

**Files:** `cmd/sdlc/publishgate.go:131-191`, `trackercompletion.go`
(`ownedPublishIssues` passes `Binding.ReviewedHEAD`), and `publishgate_test.go`.

`publishIssue` gains `Reviewed string`.
- Tracker mode: `Reviewed` is `Binding.ReviewedHEAD`, and the evidence commit
  stays the ownership token.
- Legacy mode: `Reviewed` is the codecomplete anchor commit.

For each entry, call `gitx.RebasedReviewedBase(TrunkRef, Reviewed)`, then
`DiffNames(S, "HEAD")`, then `classifyPublishDelta`. Any refusal refuses the
publish. The old `revCount(anchor..HEAD)` minimum-anchor logic is deleted: each
issue is checked against its own reviewed patch, which removes the "newest
anchor" approximation as well.

- [ ] **Step 1: Failing fixture tests**, extending `publishRepo`:
  - (a) close, then `git merge` main with foreign code: **pass**. This is the
    A4 regression, and it fails today with "N commit(s) landed after".
  - (b) close, then `git rebase` main: pass.
  - (c) close, merge main, then add a code commit: refuse, naming that file
    only, not main's files.
  - (d) close, then a docs commit: pass with the docs-only line (the existing
    `TestRunPublishGate_DocsOnly` keeps passing).
  - (e) close, then a conflicting merge with a resolution: refuse with the
    conflict token.
  - (f) the existing `TestRunPublishGate` and `_QuickGrewPastReview` stay green.
- [ ] **Step 2:** FAIL on (a)–(c) and (e). **Step 3:** Implement. **Step 4:** PASS.
  Mutation: diff against `MainBase` instead of `S`, and (a) must still pass
  while (c) must fail by including main's files. Diff against `Reviewed`
  instead of `S`, and (a) must fail.
- [ ] **Step 5:** Commit `#304 M2: publish gate compares the rebased reviewed patch (A4)`.

### Task 10: close binding survives a rebase; unowned codecomplete details refuse (B3, D8)

**Files:** `cmd/sdlc/trackercompletion.go` (`ownedCompletions`),
`trackercompletion_test.go`, `tracker_e2e_test.go`.

- [ ] **Step 1: Failing tests.**
  - (a) A tracker e2e: close, then `git rebase` main (the evidence commit is
    rewritten). `branchOwnedCompletions` returns the card, and the publish gate
    runs and passes.
  - (b) After the same rebase, plus a code commit: the publish gate refuses.
    It fails open today: "nothing to verify".
  - (c) A conflict on the details file during the rebase changes the patch-id.
    The card is not owned by either rule, the branch patch changes its details
    file, and the gate refuses with "#N is codecomplete but this branch carries
    no close for it — re-run `sdlc close --issue N`".
  - (d) A card bound to the repo, with an evidence commit on main already
    (settled), is unaffected.
  - (e) `completeLandingPR` with a squash landing still completes (existing
    test).
- [ ] **Step 2:** FAIL. **Step 3: Implement** inside `ownedCompletions`.
  - After the `inHead` check fails, compute `pid := StablePatchID(evidence)`.
    The evidence object must be present; the pin keeps it alive.
  - Scan `git rev-list merge-base(main, head)..head -- <binding details path>`,
    the commits touching the details file, which is a small set. If any has an
    equal non-empty patch-id, treat the card as owned, and log
    `cinfo("#N: close recognised after a rebase (evidence <old8> → <new8>)")`.
  - Separately, `runPublishGate` collects codecomplete cards bound to this repo
    whose details path appears in `DiffNames(merge-base, HEAD)` but which are
    not owned, and refuses. The owned set is computed once and shared.
- [ ] **Step 4:** PASS. Mutation: remove the patch-id rule, and (a) must fail.
  Remove the unowned refusal, and (c) must fail with "nothing to verify".
- [ ] **Step 5:** Commit `#304 M2: close binding survives a rebase (B3)`.

### Task 11: pin lifecycle end (ARCH-FUNERAL)

**Files:** `cmd/sdlc/trackercompletion.go` (`completeOnCard`),
`cmd/sdlc/abandon.go`, `cmd/sdlc/landingarchive.go` (legacy archive), and tests
beside each.

- [ ] **Step 1: Failing tests:**
  - the ref is gone after `completeOnCard`, after `abandon`, and after a legacy
    archive;
  - an unpin failure only warns.
- [ ] **Steps 2–4:** FAIL, implement `unpinReviewed(id)` at the three sites,
  PASS. Mutation: remove each site alone, and its own test must fail (lesson
  #286 BR-2: one test per call site).
- [ ] **Step 5:** Commit `#304 M2: reviewed-head pin is removed when the issue ends`.

### Task 12: M2 docs + close

- [ ] Atlas: `atlas/workflow/pre-merge-checks.md` and the publish-gate section
  of `sdlc-binary.md` (the reviewed-patch delta, the rebase-recognised binding,
  the unowned refusal).
- [ ] Lessons (only for what code doesn't enforce): "Define a main-relative
  check on the branch patch (`merge-base(main,HEAD)..HEAD`) or on the rebased
  reviewed patch, never on commits after an anchor. Merges and rebases of main
  change commits, not the patch."
- [ ] Post a project-file scope note under `#ariadne-304` (the close gate ticks it).
- [ ] Run `make test`, then `sdlc close --issue 304 --verified '<evidence>'`.

## Verification (manual, after both milestones)

Replay pair#410's shape in a scratch clone:
1. M1, then `milestone-close`.
2. `git merge origin/main` (≥ 50 foreign commits).
3. M2 commits, then `milestone-close --dry-run`.
4. Expect the window line to read `interdiff since M1 review …` with M2's file
   count only.
5. `close`, then merge main again, then `sdlc pr` / `sdlc merge --dry-run`.
6. Expect the publish gate to pass with no re-close.

## Risks / coordination

- **ariadne:2 (#300)** edits the judge dispatch/parse path and the round's
  blocked/passed stamping (D2/D3 of #300). The only overlap is `gatestate.Round`
  (an additive `reviewed` field) and `persistBoundaryRound`, where the stamp is
  computed. Whichever lands second rebases onto the other; the conflict is
  textual and local.
- **git ≥ 2.40** is required for `merge-tree --merge-base`. The fleet runs 2.54.
  The binary refuses older git with a named requirement, never with a silent
  fallback.
- **The patch-id rule (D8) could match a cherry-picked close of the same issue
  on an unrelated branch.** That requires the same issue's identical evidence
  commit on two branches, which "one issue, one branch" (#272) forbids. The
  details-path scan keeps the cost negligible.

## Revisions
