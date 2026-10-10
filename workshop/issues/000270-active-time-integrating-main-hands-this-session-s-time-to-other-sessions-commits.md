---
id: 000270
status: working
deps: []
github_issue:
created: 2026-09-28
updated: 2026-10-09
estimate_hours:
card_mirror: '89dbeaf6d7230579d668516b59a9bd68f576f7aa' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-09T20:01:25-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# active-time: integrating main hands this session's time to other sessions' commits

## Problem

`sdlc actual` (active-time-v3) cuts a session's activity into segments at issue
commit boundaries, and each segment goes to the issue its bounding commit
references. The boundaries come from `git log HEAD` over the time window
(`internal/activetime/commit.go`, `loadWindowCommits`). That includes every
commit that became an ancestor of the branch, including other sessions' work
that arrived when main was integrated. Those foreign commits then claim this
session's time.

Observed in pair#247 on 2026-09-28, after rebasing the issue branch onto
`origin/main` (the practice ariadne#269 proposes):
- Before the rebase, `sdlc actual --issue 247` measured 1.24 h at M1 close.
  After it, 0.34 h.
- `sdlc active-time` over this session's transcripts showed the design and M1
  segment (21:24–22:04, 39.4 min of #247 work) bounded by `57feb10 #341: …`, a
  commit another session made on main at 22:04. All 39.4 min went to #341.
- Rebasing again with `--committer-date-is-author-date` only moved it to
  0.40 h. That residue is a second bug: `git log --since/--until` filter by
  COMMITTER date while segments use author date (`%aI`), so a plain rebase
  (which restamps committer dates) also moves commits in or out of the window.

Merging main in has the same effect, because merged commits are ancestors
too. So any integration before close (#269) corrupts the hours `close` adopts
into velocity calibration, unless this is fixed first.

## Spec

- **Boundaries are the issue branch's own commits.** A commit reachable from
  the configured main at the branch's merge base, and not referencing the
  measured issue, is not a boundary for this measurement: it belongs to
  whatever produced main's history. Commits on the branch itself, issue-tracker
  refs (#252) and main commits that reference this issue (e.g. its creation)
  still count. The segment's time is then attributed by the next real boundary
  or the existing fallback rules.
- **Window filtering by author date.** Select the window by the author date
  that segmentation already uses, not `git log --since/--until` (committer date).
  Otherwise a rebase changes the answer.
- A test fixture reproduces the incident: a session's transcript, a branch with
  its own issue commits, and interleaved foreign-issue commits on main pulled in
  by (a) rebase and (b) merge. The measured hours are identical before and after
  each integration.

Related: ariadne#269 (integrate main before close) depends on this; land this
first, or #269's gate makes the undercount routine.

## Done when

- The reproduction measures the same `sdlc actual` before and after rebasing or
  merging main in, both with and without restamped committer dates.
- A foreign-issue commit reachable only through main never bounds a segment
  that this session's issue commits enclose.
- pair#247's measurement, replayed at its branch tip without undoing its
  rebase, credits #341 nothing and gives #247 the slot's whole in-window time.
- A claimed issue's window starts at its claim, even when it was filed
  earlier, so scoped boundaries can't sweep the slot's other work since the
  filing into the issue. `windowStart`/`resolveWindowStart` tests pin it.

## Plan

Design (ARCH-DRY: one boundary loader feeds both the segment boundaries and
the peer set):

- `activetime.Options` gains `Scope{BranchPoint, Issue}`. `computeActual`
  fills it from `gitx.MergeBaseWithMain()` ("" on main / no divergence →
  legacy, unscoped behavior, so the direct-on-main flow is unchanged).
- `loadWindowCommits` reads `git log HEAD <extraRefs>` with no
  `--since/--until` and filters by author date in Go (ariadne's full history
  walk is ~50 ms). With a scope, a commit is a boundary iff it is the
  branch's own (`git rev-list HEAD ^<branch point>`) or its subject
  references `Scope.Issue`. That drops merged/rebased-in main commits **and**
  other issues' tracker card commits (another slot's claim at 14:00 is the
  same foreign-boundary bug); this issue's tracker claim/close commits stay.
- Peers come from the same scoped commit set (`activetime.WindowIssues`),
  replacing `gitx.DiscoverWindowIssues` (committer-date, HEAD-only); its
  #190 foreign-ref test moves with it.

Steps:
- [x] Failing real-git fixture test: branch with own #N commits, interleaved
      foreign #M commits on main; integrate by rebase, rebase
      `--committer-date-is-author-date`, and merge; boundary set (and peers)
      identical across all four states, and no #M boundary.
- [x] Author-date window filter + branch scope in `loadWindowCommits`.
- [x] `WindowIssues` replaces `DiscoverWindowIssues`; wire `computeActual`.
- [x] End-to-end check: `Compute` minutes equal before/after on the fixture.
- [x] Re-measure pair#247 (branch tip in a temp worktree) and log the value.
- [x] Atlas note on boundary scoping.
- [x] Claim-anchored window start (from #254; see Revisions).

## Log

### 2026-09-28

- Filed from pair#247 at the operator's request. pair#247's M2 close is on
  hold for this fix, so the calibration row gets a true measured value
  instead of a hand-typed one.

### 2026-10-09
- 2026-10-09: closed — make test green except load/sandbox flakes (processgroup /bin/ps blocked by sandbox; TestPlanningReviewConcurrencySchedules and TestCLISignalCancelsOwnedReviewer pass in isolation, review-dispatch timing). TestBoundariesSurviveIntegratingMain: not integrated / rebase / rebase --committer-date-is-author-date / merge-end / merge-mid give identical boundaries, peers, Compute minutes; mutation-checked. TestActualScope pins scope wiring (mutation-checked). Claim anchor pinned by TestWindowStart/TestResolveWindowStart; #270 measures from its 20:01 claim instead of the 09-28 filing (23.26h -> minutes). pair#247 replay: #247 0.60h -> 2.33h, #341 0. Review round 2 fixes: BR-1..4, BR-6; BR-5 declined (see Log).; review verdict: FIX-THEN-SHIP
- 2026-10-09: flow upgraded quick → full — 177 added lines in code files (limit 100); an earlier round of this close already ran the full review

- Dispatched by the ariadne-robustness-1 TL (ariadne:1); evidence finding A1.
  Claimed in ariadne:3.
- Fixture `TestBoundariesSurviveIntegratingMain` (real git): not integrated,
  plain rebase, rebase `--committer-date-is-author-date`, merge at the end,
  merge mid-work. All five give boundaries `[#9 a, #9 b]`, peers `[9]`, and
  `Compute` #9 = 50 min, #5 = 0. Mutation-checked: dropping the scope filter
  fails the 4 integrated states; keeping merges fails merge-mid-work; putting
  back `--since/--until` fails the plain rebase; dropping the
  issue-named-trunk exception fails
  `TestScopedBoundaryKeepsTrunkCommitsNamingTheIssue`.
- Scope decision: with a branch point, other issues' **tracker** card commits
  also stop being boundaries. Another slot's claim is the same foreign-boundary
  bug as a main commit. The measured issue's own claim/close commits still
  count. Merge commits are excluded from the branch's own set, since they are
  integration, not work.
- `gitx.DiscoverWindowIssues` (committer date, HEAD only) is replaced by
  `activetime.WindowIssues` over the same scoped commits (ARCH-DRY). Its #190
  foreign-ref test moved with it. Event and commit window parsing share
  `windowBounds`.
- pair#247 replay (local clone at its close commit 93994d3d, pair-slot1 and
  brain transcripts, branch point 25233cdf = main before its PR merge, window
  claim 21:31:56 → close 23:59:36):
  - old engine (installed `sdlc active-time`): #247 0.60 h, #341 1.50 h. The
    81.5 min design+M1 run (21:32–22:53, 14 #247 mentions) went to #341 via
    `d7699cb`.
  - new, unscoped (author-date only): #247 1.96 h.
  - new, scoped: #247 2.33 h = all of the slot's in-window activity, with no
    #341.
  - M1 window only (→ 22:20:04), scoped: 0.86 h. The pre-rebase 1.24 h opened
    the window at the branch's old parent `6f62a7ec` (20:54), i.e. 38 min
    before the claim; see Revisions.
- Residual rebase sensitivity, handed to #254: `CommitWindow`'s left edge is
  the parent of the first `#N` commit when that is earlier than the claim, and
  a rebase moves that parent. #254 anchors the window at the claim, which
  removes it. #254 is next in this slot.
- Close review rounds 1–2 (round 1 never ran: the sandbox proxy refused the
  reviewer's API call, so the verdict was "unknown"; round 2 rerun with the
  API host allowed): FIX-THEN-SHIP.
  - BR-1 (Important), stale atlas text: fixed. sdlc-binary.md now says "claim
    when present, else parent" and names `activetime.WindowIssues`.
  - BR-2: fixed with gofmt.
  - BR-3: fixed. `sdlc active-time --branch-point <rev>` applies the same
    scope, and the first `--issue` is the measured one. This is the replay
    tool pair#247 needed.
  - BR-4 + BR-6: fixed. `actualScope` also scopes the issue's own branch
    before its first commit (where the branch point is HEAD).
    `TestActualScope` pins the wiring and fails when that branch is removed.
  - BR-5 (prefilter with `--since`): declined. The full walk is about 50 ms,
    twice. git's `--since` stops the walk at the first commit with an
    old-enough committer date, so a `--committer-date-is-author-date` rebase
    can hide in-window commits behind older-dated ones. That reopens the
    rebase sensitivity this issue removes.
- Close review round 3: FIX-THEN-SHIP, no blocking findings. The one advisory
  was BR-3's missing test. Fixed in the same round: `activeTimeScope` is
  extracted, and `TestActiveTimeScope` covers the scope, the unscoped default,
  and the `--issue` refusal through the CLI.

## Revisions

- 2026-10-09 — Done-when 3's "≈1.3–1.5 h" was a pre-rebase figure whose window
  opened 38 min before the claim, at the old branch parent. With boundaries
  scoped and the window at the claim, pair#247 measures 0.86 h for M1 and
  2.33 h for the whole issue, against 1.81 h hand-composed at close (M2's
  0.57 h was itself measured after the rebase). Done-when 1 holds for the
  engine (boundaries, peers and window filter). The window's left edge becomes
  rebase-invariant with #254.
- 2026-10-09 — Folded #254's window-start half (Spec 2) into this issue. The
  first `sdlc close` measured 23.26 h. The window opened at the 09-28 filing
  commit, and with main's other-issue commits no longer boundaries, all of
  ariadne:3's work since then flowed to #270's commits. So scoped boundaries
  are only safe with a claim-anchored window. `windowStart` now returns the
  claim whenever one exists, and the commit parent only without one. #270
  measures 0.14 h at 20:17 (claimed 20:01). #254 keeps the PR-number
  extraction half (Spec 1) and its before/after re-measure.
