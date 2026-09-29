---
id: 000270
status: open
deps: []
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '1d02414ef5aa552d1815e46c018f3f3aa1630762' # card fields mirrored from issue-cards; edit via sdlc
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
- pair#247's measurement returns to its pre-rebase value (≈1.3–1.5 h) without
  undoing its rebase.

## Plan

- [ ]

## Log

### 2026-09-28

- Filed from pair#247 at the operator's request. pair#247's M2 close is on
  hold for this fix, so the calibration row gets a true measured value
  instead of a hand-typed one.
