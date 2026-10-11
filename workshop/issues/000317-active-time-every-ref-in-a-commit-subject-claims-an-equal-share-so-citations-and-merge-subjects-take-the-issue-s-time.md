---
id: 000317
status: codecomplete
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-10
estimate_hours:
card_mirror: 'a676275adc50be3d9392112f049655ffed377a3c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T16:53:10-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "5bdcac5d", done: "0655980a"}
actual_hours: 0.36
---

# active-time: every ref in a commit subject claims an equal share, so citations and merge subjects take the issue's time

## Problem

`sdlc actual` gives every local `#N` in a commit subject an equal share of the
activity run that commit claims (`activetime.attributeRun`). A subject that
*cites* another issue therefore hands that issue a share of the measured
issue's time. Two cases, both found reproducing ariadne:4's report that #304
measured 2.23 h before merging main and 1.01 h after (wall clock about 2.4 h,
claim 20:01 → close 22:26).

1. **Integration merge (regression from #270).** #270 drops merge commits
   from a branch's own boundary set, but keeps any commit whose subject names
   the measured issue. The merge `#304: merge origin/main (#300, #270
   landed)` names #304, so it stays a boundary. Its run is then split three
   ways. Replay at the post-merge head, scoped as `sdlc actual` does:
   #304 60.8 m, #300 44.6 m, #270 44.6 m, so **1.01 h**. Without that one
   boundary it is 1.38 h, and the pre-merge head gives 1.37 h. The number
   moves on integration, which #270 exists to prevent.
2. **Citations (pre-existing, v3 rule).** `#304: log: --no-validate rationale
   for the landing (F1, #308)` cites #308, which then takes **66 min** of
   #304's time at both heads. The window's total active time is 2.48 h, and
   all of it was #304 work in slot :4, but #304 measures 1.37 h.

Replay method: a local clone at each head; Compute with the slot-4 and brain
transcript dirs, tracker ref, claim window 20:01:24 → the head's last #304
commit, and branch points b77e9718 (pre) / 3fddafdf (post, main before PR 174).

## Spec

- Only a subject's **leading** issue ref(s) claim time. That is the `#N:` /
  `#N Mx:` lead (and the `ariadne#N` self-qualified form), and any refs the
  convention puts at the lead (`close #N`, `#N, #M:`). Refs later in the
  subject are citations. They stay in the mention scope but are not
  claimants. One extractor in `issueref`, shared by boundaries and peers
  (ARCH-DRY with #254's merge-PR mask).
- A branch's own integration merge is never a claimant beyond its lead.
  Leading-only covers it (`#304: merge origin/main (…)` claims for #304 only).
  Decide whether such merges should be boundaries at all; with leading-only
  they at least cannot leak.
- Replay test from real subject shapes: the #304 merge subject, a
  parenthesised citation, `#174-#176`, `chore: bump (refs #1, #2)`, and a
  subject with no lead (it stays a neutral boundary, or claims its refs; pick
  one and record why).

## Done when

- The #304 replay measures the same before and after its main merge, and #308
  and #270/#300 get nothing from #304's subjects.
- A table test pins lead vs citation extraction over the shapes above.
- Re-measure two or three recent closes before/after and record the deltas
  for calibration consumers.

### Decisions (2026-10-10, ariadne:3)

- **Lead grammar** (`issueref.Lead`): after any `area:` labels (`sdlc: `,
  `issues: `), optionally one lead verb from a closed list (`close`, `closes`,
  `file`, `issue`, `plan`), the subject must start with a ref; the lead is
  that ref plus refs chained to it by `-`, `–`, `,`, `/`, `+`, `&`
  (`#174-#176`, `#N, #M:`). Anything after is a citation. The verb list is
  closed on purpose: an open "any word" verb would read
  `side-quest: follow #N's plan` as a claim, which is the leak.
- **No-lead subject → neutral boundary.** By the commit convention (§12) the
  issue a commit works for leads; `chore: bump (refs #1, #2)` names no owner.
  Claiming its refs is the citation leak again. Neutral still ends a
  claimant's reach and leaves the run to mentions.
- **Citations stay in the mention scope:** `Commit.Refs` keeps every local
  ref (feeds `WindowIssues`); `Commit.Issues` becomes the lead claimants
  (feeds boundaries/attribution).
- **Scoped boundaries use the lead:** a main commit is a boundary only if its
  *lead* names the issue (a main commit that cites #N no longer bounds #N's
  sessions), and **merge commits are never boundaries in a scoped
  measurement**. Integration is not work (#270's rule for `branchOwn`); with
  it the boundary set before and after integrating main is identical by
  construction, not just leak-free.

## Plan

- [x] `issueref.Lead` / `LeadLocalNums` + table test over real subject shapes
- [x] activetime: `Commit.Refs` (mention scope) vs `Commit.Issues` (lead
      claimants); scoped filter uses the lead and drops merges
- [x] Tests: commit-load test for citation/merge subjects; #304-shaped
      regression (pre == post, no share to #308/#300/#270)
- [x] Replay #304 (pre/post) and 2–3 recent closes old vs new; record deltas
- [x] atlas: active-time attribution section

## Log

### 2026-10-09

- Filed by ariadne:3 at the ariadne:1 TL's request, investigating ariadne:4's
  #304 hours report. Not fixed yet (TL: file, don't fix).

### 2026-10-10
- 2026-10-10: closed — Re-close after the close-review Minor fix (doc/comment wording only: helptext, Scope comment, atlas). Prior evidence stands: make test 12 shards 0 failed; processgroup fails only under the sandbox (ps fork denied), passes unsandboxed; targeted activetime/issueref tests pass; #304 replay pre 2.48h / post 2.50h all #304.; review verdict: SHIP
- 2026-10-10: closed — make test: 12 cmd/sdlc shards 0 failed; other packages pass except processgroup TestCancellationKillsDescendants, which fails only under the agent sandbox (fork/exec /bin/ps not permitted) and passes unsandboxed. New tests: issueref TestLeadLocalNums (19 real subject shapes), activetime TestLoadWindowCommitsLeadClaimsCitationsDont, TestCitationsDoNotClaim, TestBoundariesSurviveIntegratingMain/"subject names the issue" (#304 merge shape). Replay of #304 through Compute as sdlc actual calls it: pre-merge 1.37h(+#308 1.10h) -> 2.48h all #304; post-merge 1.01h(+#300/#270 0.74h each) -> 2.50h all #304. Recent closes old->new: #319/#300/#270/#315/#296 unchanged, #283 6.28->7.35h.; review verdict: SHIP

- ariadne:3 claimed (TL dispatch, batch 2). Replay reproduced through
  `Compute` exactly as `sdlc actual` calls it (clones at 7885ed1e / 8470f8eb,
  slot-4 + brain transcript dirs, tracker ref, since 20:01:24): pre #304
  82.4m + #308 66.2m (2.48h total); post #304 60.8m, #300 44.6m, #270 44.6m.
- Implemented: `issueref.Lead`/`LeadLocalNums` (closed-verb lead grammar);
  `Commit.Issues` = lead claimants, `Commit.Refs` = all local refs (mention
  scope); `branchCommits` reads `rev-list --parents` once and the scoped filter
  drops merges and keys main commits on the lead. `TestAttributionGolden`'s
  `#10 second (#8)` encoded the old citation-claims rule; it is now the lead
  list `#10, #8: second`, which keeps the split it pins. Help text
  (`helptext/active-time.md`) and atlas (`sdlc-binary.md`) updated.
- **Replay after the fix** (same harness, scoped as `sdlc actual`):
  | #304 head | old | new |
  |---|---|---|
  | pre-merge 7885ed1e | #304 1.37h, #308 1.10h | #304 2.48h (all) |
  | post-merge 8470f8eb | #304 1.01h, #300 0.74h, #270 0.74h | #304 2.50h (all) |
  | close 85394ff5 | #304 2.68h | #304 2.68h |
  Pre vs post differ only by the window end (22:14 vs 22:15). At close the
  old engine already gave 2.68h: the merge sat inside a run whose next
  boundary was the close. The 1.01h reading depends on when you measure.
- **Calibration deltas (old → new), recent closes:** #319 0.26 → 0.26,
  #300 1.92 → 1.92, #270 0.79 → 0.79, #315 0.01 → 0.01, #296 0.33 → 0.33 (no
  cited subjects in a claiming position); **#283 6.28 → 7.35h** (`#283:
  project: … file #284-#287` had handed #284/#287 1.07h; #284 keeps 0.54h
  from its own filing commits on the branch). Recorded actuals on cards
  differ from both because they were measured at other heads/times.

