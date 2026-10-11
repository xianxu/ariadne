---
id: 000321
status: codecomplete
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: '8b622085cf6ffe7b2c361a2a9b6aa22cfcbf2797' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T17:24:16-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "75f7d8ad", done: "2408e024"}
actual_hours: 0.03
---

# active-time: off the issue branch, every slot's tracker card commits bound and claim runs, so sdlc actual after landing hands an issue's run to an unrelated card write

## Problem

After ariadne#317 landed, `sdlc actual --issue 317` from slot 3's resting
branch says "found no measurable activity", although the session's transcript
is in the selected dir. The cause is not the lead-ref rule (every #317
subject leads with #317) and not transcript selection: the slot-3 dir is
selected and gives 127 events in the window.

The cause is that **the measurement is unscoped off the issue branch, but it
still reads the tracker ref.** `actualScope` returns no branch point when HEAD
has not diverged from main and is not the issue's branch, which covers a
resting branch or main after landing. `loadWindowCommits` then keeps every
commit in the window, and `actualTrackerInputs` adds `origin/issue-tracker`,
whose history interleaves **every slot's** card writes (`#189: tracker: update
card`, `#183: …`, `#320: …`). Each one leads with its issue, so each is a
claimant boundary. `selectClaimant` then gives a run to the nearest
*outside* boundary whenever a next boundary exists. Commits inside the run
are ignored in that case. So a card write by another slot two seconds before
the run starts takes the whole run.

Evidence (Compute with `computeActual`'s inputs: window 16:53:10 → 17:17:10,
dirs brain + slot 3, ExtraRefs `origin/issue-tracker`, peers [183 189 317 320]):

- unscoped (what `sdlc actual` does from main-slot3): one run 16:53:15 → 17:16,
  23.4 min, claimed by `e932356 #189: tracker: update card` (16:53:13);
  per-issue = {189: 23.4}, #317 = 0 → "no measurable activity".
- same inputs, but boundaries limited to commits whose lead names #317 (Scope
  with a branch point): the run is claimed by `9c03e2e #317: tracker: update
  card`; #317 = 23.4 min (0.39h), which matches the ~16:52–17:18 session.

Not caused by #317. The pre-#317 binary gives the same answer, because the
card subjects lead with their issue under both rules.

**Related observation, not this bug:** the close-time numbers follow from the
window rule. The first close measured 16:53:10 → 17:03:37, the last #317
commit then, which gives 0.17h. The re-close extended the window to 17:14:51,
giving 0.36h, which the card holds. The window ends at the last `#N` commit,
so verification between that commit and the close is not counted: the full
`make test` and the close review, about 11 min here. Decide separately
whether that is intended.

## Spec

- Tracker-ref commits are this issue's activity only when their lead names
  the measured issue. Every other slot's card write is unrelated noise in
  every scope, not only on an issue branch. Filter the tracker history to
  the measured issue's lead whether or not the measurement is scoped
  (ARCH-DRY with #270's `Scope.Issue` rule).
- Decide whether an unscoped measurement after landing should scope to the
  landed branch (its PR's branch point), so it reproduces the close-time
  number instead of treating main's other work as boundaries.
- Consider whether `selectClaimant` should prefer the run's own inside
  claimants over a nearer outside boundary. A boundary 2 s before a 23-min run
  that contains five commits of another issue is a weak claim. That is a
  #92 model change, so weigh it separately.

## Done when

- `sdlc actual --issue 317` from a resting branch at or after the landing
  measures about 0.39h for #317, not 0. Replay window 16:53:10 → 17:17:10,
  slot-3 + brain dirs.
- A test pins it: an unscoped window whose tracker ref carries another
  issue's card commit just before a run does not give that run to the other
  issue.

## Plan

- [x] Failing test: an unscoped window whose tracker carries another issue's
      card write 2 s before a run (`TestTrackerCommitsOfOtherIssuesDoNotClaim`)
- [x] `loadWindowCommits`: commits reachable only from extra refs
      (`besideHead`, `rev-list <refs> ^HEAD`) bound only when their lead
      names `Scope.Issue`, in every scope
- [x] Verify `sdlc actual --issue 317` unscoped after landing
- [x] atlas

## Log

### 2026-10-10
- 2026-10-10: closed — make test: all cmd/sdlc shards pass; the only failure is processgroup TestCancellationKillsDescendants, which fails only under the agent sandbox (fork/exec /bin/ps not permitted) and passed unsandboxed in #317. New TestTrackerCommitsOfOtherIssuesDoNotClaim fails before the fix (#5 card write 2 s before the run took all 40 min) and passes after it. Done-when: `sdlc actual --issue 317`, unscoped from HEAD = main with #317 landed, measures 0.39h (it said "no measurable activity" before).; review verdict: SHIP

- Filed by ariadne:3, a TL-requested diagnosis of the low #317 measurement
  after #317 landed. Diagnosis only (TL: file, don't fix).
- 2026-10-10 TL: fix it now. In scope: the tracker filter in every scope,
  plus Done-when. Landed-branch scoping only if it falls out naturally. The
  selectClaimant change (#92) and the verification-tail question are out.
- Implemented the filter. The zero `Scope.Issue` keeps every commit, which
  matches the zero Scope's contract and the standalone `active-time` CLI,
  which reads no extra refs. Landed-branch scoping was not needed for
  Done-when, so it is left out. `sdlc actual --issue 317` from this branch
  (HEAD = main with #317 landed, not diverged, so unscoped): **0.39h**, was
  "no measurable activity".

