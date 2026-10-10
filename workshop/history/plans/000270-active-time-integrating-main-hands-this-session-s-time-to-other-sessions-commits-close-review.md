# Boundary Review — ariadne#270 (whole-issue close)

| field | value |
|-------|-------|
| issue | 270 — active-time: integrating main hands this session's time to other sessions' commits |
| repo | ariadne |
| issue file | workshop/issues/000270-active-time-integrating-main-hands-this-session-s-time-to-other-sessions-commits.md |
| boundary | whole-issue close |
| milestone | — |
| window | b77e971865f485b0b0fca59a03130d7febb04154..981512b3ff71122baf43388bd696b78b27319839 |
| command | sdlc close --issue 270 |
| reviewer | claude |
| timestamp | 2026-10-09T20:30:41-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

## Review summary: ariadne#270

**Verdict: FIX-THEN-SHIP.** The engine change is correct and well tested. One Important finding should be fixed before close: `atlas/workflow/sdlc-binary.md` still describes the old window-start rule and names a function this diff deletes.

**What the change does.**
- On an issue branch, only two kinds of commit can bound a segment: the branch's own non-merge commits (`rev-list --no-merges HEAD ^<branch point>`), and commits whose subject names the measured issue.
- The window is filtered by author date in Go, not by `git log --since/--until`, which uses committer dates.
- The peer set comes from the same commit loader.
- The window now starts at the claim whenever a claim exists.

**How I checked it.**
- I ran the stat and name-status recipes, then read the full diff.
- `go test ./cmd/sdlc/internal/activetime/ ./cmd/sdlc/internal/gitx/` and `go test ./cmd/sdlc/ -run 'WindowStart|Actual'` both pass.
- `go vet` is clean on the touched packages. `gofmt -l` flags one file (Minor 1).
- Every Plan item and every Done-when item maps to code or a test, except the pair#247 re-measure. That is evidence recorded in the Log, so I didn't re-run it.

### 1. Strengths
- **The fixture reproduces the real incident** (`cmd/sdlc/internal/activetime/branchscope_test.go:72`). It uses real git and covers five states: not integrated, plain rebase, rebase with `--committer-date-is-author-date`, merge at the end, and merge mid-work. Each state checks the boundaries, the peers and the minutes `Compute` assigns. Each of the three fixes is caught by at least one state, which matches the mutation results in the Log.
- **Merge commits are left out of the branch's own commits** (`commit.go:115`). This is the detail that makes "merged mid-work" give the same answer as "not integrated".
- **The design follows ARCH-DRY.** `WindowIssues` reads the same commits `Compute` segments on, so the peer set and the boundaries cannot disagree. Before this change they came from different git queries (committer date vs. author date, HEAD only vs. tracker refs).
- **Duplicate code is shared.** `windowBounds` now handles window parsing for both the commit loader and the event loader. The ARCH-DRY intent of the #190 foreign-ref test is kept, since it moved to `TestWindowIssuesExcludesForeignRefs`.
- **`windowStart` is simpler and no longer fragile.** It no longer compares ISO strings as text, so the old caveat about UTC offsets and DST is gone along with the code that needed it.
- **Peer discovery now reports errors.** The old `DiscoverWindowIssues` hid a failed git call by returning just the primary issue. Now `computeActual` returns `actualError` instead.

### 2. Critical
None.

### 3. Important
- **The atlas contradicts itself** (`atlas/workflow/sdlc-binary.md`).
  - Around line 1005 it still says "The window-**start** is the *earlier* of `CommitWindow`'s parent-of-first-`#N`-commit and the engagement anchor". The paragraph added later in the same section says the claim always wins.
  - Lines 715 and 980 still name `DiscoverWindowIssues` as a gitx export and as the source of peers. This diff deletes it.
  - **Fix:** rewrite the "earlier of" sentence to "the claim when present, else the parent of the first `#N` commit". Replace `DiscoverWindowIssues` with `activetime.WindowIssues (branch-scoped)` at both places.

### 4. Minor
1. **`cmd/sdlc/internal/gitx/window.go:17`** isn't gofmt-clean: removing the imports left a blank line before `)` in the import block. Run `gofmt -w` on the file.
2. **ARCH-PURPOSE:** the standalone `sdlc active-time` (`cmd/sdlc/activetime.go:212`) builds `Options` with no `Scope` and has no flag to set one. The Log used this tool for its "old engine" numbers, and its output can still differ from what `sdlc actual` adopts. A `--branch-point`/scope flag, or a note in its help text, would align them.
3. **The scope turns off when the branch has no commits of its own yet.** `MergeBaseWithMain()` returns `""` when HEAD equals the merge base. At that point other slots' tracker claim/close commits become boundaries again, unlike the scoped rule. It doesn't matter at close, where the branch has commits, but it affects an early `sdlc actual` preview.
4. **ARCH-CONSTRAINTS:** `loadWindowCommits` now walks the whole HEAD and tracker history with no bound. It runs twice per measurement (once in `WindowIssues`, once in `Compute`), plus two `rev-list` calls. That's fine at ariadne's size (~50 ms), but the cost grows with history. Keeping `--since=<since>` as a committer-date prefilter would bound the walk without bringing back the rebase bug, because a rebase only moves committer dates later. Only the `--until` side was the problem.
5. **ARCH-PURE:** the window and scope filter logic sits inside the git-reading loop in `loadWindowCommits`. A pure `keepBoundary(ts, sha, issues, since, until, own, scope)` helper could be unit-tested without git or the `gitRun` fake.
6. **Untested wiring in `computeActual`:** no test covers how it builds the scope (`actual.go:114`, `Scope: scope` at line 133). The fixture calls `loadWindowCommits` and `Compute` directly, so dropping `Scope: scope` from the `Options` would not fail any test.

### 5. Test coverage
- **Well covered:** boundaries, peers and minutes are identical across all five integration states. The unscoped behaviour on main is unchanged. A commit on main that names the measured issue is still a boundary. `windowStart` and `resolveWindowStart` cover the claim-wins cases, including the one whose expected value flipped, "started later than parent".
- **Not covered:**
  - a measured-issue claim/close commit on the tracker ref while scoped (only the trunk-commit case is tested);
  - a branch with zero own commits (Minor 3);
  - the glue in `computeActual` (Minor 6).

### 6. Architecture principles
| Marker | Result |
|---|---|
| ARCH-DRY | Pass: one loader feeds both the boundaries and the peers, and `windowBounds` is shared. |
| ARCH-PURE | Pass, with Minor 5. |
| ARCH-PURPOSE | Pass. The main consumer, `computeActual`, uses the scope. The standalone `active-time` doesn't (Minor 2). |
| ARCH-MOCK | Pass: tests use real git through `gitRun`, and the existing fakes still work. |
| ARCH-CONSTRAINTS | Pass, with Minor 4. |
| ARCH-SECURE | Pass: git output is parsed with a tab split, and malformed lines are skipped. |
| ARCH-ORDER | Pass: the code keeps no state between events, since every measurement is a fresh git read. |
| ARCH-FUNERAL | Pass: the code writes nothing durable; it only reads git and returns values. |

Note for #254: the PR-number extraction half still to come should go through the same branch-scoped loader, not a new git query.

### 7. Plan revisions
None. The existing Revisions entries already record the folding of #254's window half and the change to Done-when 3's figure.

```findings
findings:
  - id: new
    severity: Important
    family: atlas-stale-after-surface-change
    title: |
      sdlc-binary.md still says window-start is the earlier of parent and claim, and names deleted DiscoverWindowIssues
    detail: |
      atlas/workflow/sdlc-binary.md around line 1005 says "the window-start is the earlier of" the two anchors, which the new paragraph contradicts, and lines 715 and 980 still name DiscoverWindowIssues, which this diff deletes. Rewrite the sentence to "claim when present, else parent" and name activetime.WindowIssues.
  - id: new
    severity: Minor
    family: gofmt-clean
    title: |
      gitx/window.go import block keeps a blank line before the closing paren
  - id: new
    severity: Minor
    family: single-source-consumer-sweep
    title: |
      standalone sdlc active-time builds Options without a Scope, so it can disagree with sdlc actual
  - id: new
    severity: Minor
    family: boundary-scope-consistency
    title: |
      scope is off when HEAD equals the merge base, so other issues' tracker commits bound segments before the branch's first commit
  - id: new
    severity: Minor
    family: unbounded-history-walk
    title: |
      loadWindowCommits walks all history twice per measurement; keeping --since as a committer-date prefilter would bound it
  - id: new
    severity: Minor
    family: glue-wiring-untested
    title: |
      no test covers the scope wiring in computeActual (MergeBaseWithMain into Options.Scope)
```

---

## Re-review — 2026-10-09T20:48:38-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 270 — active-time: integrating main hands this session's time to other sessions' commits |
| repo | ariadne |
| issue file | workshop/issues/000270-active-time-integrating-main-hands-this-session-s-time-to-other-sessions-commits.md |
| boundary | whole-issue close |
| milestone | — |
| window | b77e971865f485b0b0fca59a03130d7febb04154..c897abcb45184c0843dccd6c2c6a68c8bab9c980 |
| command | sdlc close --issue 270 |
| reviewer | claude |
| timestamp | 2026-10-09T20:48:38-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Verdict: ready to ship after one small test.** Five of the six prior findings are resolved. BR-1, BR-2, BR-4 and BR-6 are fixed, and the reason for declining BR-5 holds up. BR-3 is only half done. The `sdlc active-time --branch-point` flag now applies the same scope as `sdlc actual`, but no test exercises it, so under the claimed-fix rule (a behaviour change needs a test that fails without it) it stays `not-addressed`. It is Minor and does not block the gate.

**Checks I ran:**
- `gofmt -l cmd/sdlc/` reports nothing.
- `go vet` is clean on `internal/activetime` and `internal/gitx`.
- `go test` passes for `internal/activetime`, `internal/gitx`, and the `cmd/sdlc` tests that match `TestActualScope`, `TestActual`, `TestComputeActual` and `ActiveTime`.

1. **Strengths**
   - **One commit list feeds both peers and boundaries.** `activetime.WindowIssues` (`internal/activetime/commit.go`) reads the same `loadWindowCommits` output that `Compute` segments on. The tracked issues and the segment boundaries therefore can't disagree (ARCH-DRY), which removes the old split where `gitx.DiscoverWindowIssues` used committer dates.
   - **Merges don't count as the branch's own commits.** `branchOwn` uses `rev-list --no-merges HEAD ^bp`, so merging main in mid-issue adds no boundaries. This is correct and stated in the comment.
   - **Errors are no longer swallowed.** A failed peer lookup now returns `actualError`; before, it quietly fell back to just the primary issue.
   - **The early-preview gap is covered by a test that would catch a regression.** `actualScope` handles an issue branch that has no commits yet (`actual.go:155-171`). `TestActualScope` asserts the branch point is HEAD in that state, and it fails if that branch of the code is removed.
   - **BR-5's rationale is sound.** git's `--since` can stop the history walk early (it compares committer dates), so after a rebase it would miss in-window commits. Adding it as a prefilter would bring back exactly the bug this issue fixes, and the full walk costs about 50 ms.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - BR-3 remains open (see the dispose block): `--branch-point` has no test.

5. **Test coverage**
   - The scope engine has three focused tests: boundaries survive integrating main (rebase and merge cases), the unscoped path on main, and trunk commits that name the issue are kept.
   - `TestActualScope` covers how `computeActual` builds its scope.
   - Nothing tests the standalone CLI flag (BR-3).

6. **Architecture**
   - **ARCH-DRY: pass.** The only remaining mention of `DiscoverWindowIssues` is a provenance comment in a test.
   - **ARCH-PURE: pass.** git access goes through the `gitRun` shim, and the filtering is pure apart from it.
   - **ARCH-PURPOSE: pass.** Both `sdlc actual` and `sdlc active-time` take the scope from the same `Scope` type. The other `MergeBaseWithMain` caller, `milestoneclose.go`, uses it for a review window, not for time attribution.
   - **ARCH-MOCK: pass.** The tests run real git in temporary repos.
   - **ARCH-CONSTRAINTS: pass.** The full-history walk was measured at about 50 ms, and the BR-5 decision is recorded.
   - **ARCH-SECURE: N/A.** The diff only adds reads of the local repo's git history (no credentials, no external input).
   - **ARCH-ORDER: pass.** The change keeps no state between events, and the integration order is covered by the rebase and merge subtests.
   - **ARCH-FUNERAL: pass.** The change creates nothing durable; everything it builds lives in memory for one call.

7. **Plan revisions:** none needed. The Log already records the round-2 dispositions.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      sdlc-binary.md:715/980/1005 now say claim-else-parent and name activetime.WindowIssues; grep finds no stale DiscoverWindowIssues/"earlier of" in atlas.
  - id: BR-2
    disposition: addressed
    note: |
      gitx/window.go import block fixed; gofmt -l cmd/sdlc/ is clean.
  - id: BR-3
    disposition: not-addressed
    note: |
      --branch-point wiring (activetime.go:222-227, incl. the needs---issue error) is a behavior change with no test exercising the flag; add a small cobra-level test.
  - id: BR-4
    disposition: addressed
    note: |
      actualScope falls back to BranchPoint() on the issue's own undiverged branch; TestActualScope asserts it and goes red without the branch.
  - id: BR-5
    disposition: withdrawn
    note: |
      Decline is correct: git --since walk-termination uses committer dates and would reintroduce the rebase sensitivity #270 removes; ~50ms cost is acceptable.
  - id: BR-6
    disposition: addressed
    note: |
      computeActual now derives scope via actualScope(issueNum), pinned by cmd/sdlc/actualscope_test.go.
```
