---
id: 000269
status: open
deps: []
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: 'b161c09e525b0b4cdc2c7675ce7cd58be28fc4e8' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc close: integrate a moved origin/main before the boundary review

## Problem

`sdlc close` reviews the branch, and the publish gate (`merge`/`push`) only
checks that nothing changed since that review (the #160 reviewed-HEAD-unchanged
invariant). Neither looks at what `origin/main` gained after the branch point.
So a close review and its full-suite evidence can describe a tree that never
ships: the server-side merge is textually clean, and still breaks the build.

Integrating after `close` is worse. It moves HEAD, the publish gate refuses, and
`close` has to run again, re-reviewing the change since the last review. The
cheap moment is right before `close`, and nothing prompts it.

Observed in pair#247 (2026-09-28):
- Main gained pair#340's `keyhelp/review.go` without registering it in the
  artifact inventory, which failed
  `TestProductionArtifactReferencesAreExactlyClassified` on main. The #247
  branch only found out because it happened to be based after that commit.
- The branch's full-suite run hit an out-of-date generated runtime bundle
  (ignored files) left from before #340's nvim changes, and reported key-help
  failures that aren't real. A fresh `make test` regenerates the bundle and
  passes.

## Spec

A close-time gate. Before dispatching the boundary review, `sdlc close` fetches
the configured main and compares it with the branch's merge base.

- **Nothing new on main** → pass silently.
- **Main moved only in process/doc paths** (`workshop/`, `atlas/`, markdown
  outside shipped trees; the same "code surface" classification the atlas gate
  already uses, #177) → pass with an info line.
- **Main moved in code** → refuse, with the next action:
  - an unpushed branch: `git rebase <remote>/main`, re-run the suite, re-run close;
  - a pushed branch: `git merge <remote>/main` instead (a rebase would need a
    force-push, which fights merge's local/remote/PR head check).
- Bypass: `--no-integrate` (the per-gate convention), reason in `--verified`.
- Offline or fetch failure → report "could not check", don't pass it as
  up to date (probes are false/true/inconclusive).

Also worth deciding at design time:
- Whether `milestone-close` should warn (not refuse) on the same condition.
  Milestone windows are found by commit message, so a rebase there costs only
  cosmetic `Review-Window` SHAs, which then name commits no longer on the branch.
- A rebase leaves earlier milestones' `Review-Window:` trailers naming
  pre-rebase SHAs. Gates don't read them (verdicts are matched by message,
  `close.go` verdict scan), but `sdlc state`/history do. Either document that,
  or record windows in a rebase-stable form.

Related: #249 (resting branch diverges after issue sync) is the resting-branch
side of the same staleness; this issue is the issue-branch side, at close.

## Done when

- `sdlc close` refuses when the configured main has moved in code since the
  merge base, naming the rebase/merge next action by push state, and passes
  when main hasn't moved or moved only in docs. Tests use a real git fixture
  with a remote.
- A fetch failure is reported as inconclusive, never as up to date.
- `--no-integrate` bypasses it and is listed in `sdlc close --help`'s gate catalog.

## Plan

- [ ]

## Log

### 2026-09-28

- Filed from pair#247 at the operator's request, after a discussion of
  whether to rebase before close. Checked the current behavior: milestone
  verdicts survive a rebase (message-matched), and a post-close rebase trips
  the publish gate.
