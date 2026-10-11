---
id: 000269
status: working
deps: []
github_issue:
created: 2026-09-28
updated: 2026-10-10
estimate_hours:
card_mirror: '3cc07113d66c1c71a3916d9fab5e464949e32112' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-09T20:51:35-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "997c3fb5", done: "c0761a16"}
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

- **Interaction with #304:** milestone review windows are the interdiff since the last finalized review, replayed onto today's main. Integrating main (merge or rebase) therefore no longer widens a milestone window, and a conflict shows only its resolution. The freshness rule here can integrate main daily without resetting review cost. Project direction (ariadne-robustness-1): the rule is "the branch contains every origin/main commit older than the most recent noon in the repo timezone", applied at review boundaries, with #303 applying the same rule at claim/start-plan.

## Done when

- `sdlc close` refuses when the configured main has moved in code since the
  merge base, naming the rebase/merge next action by push state, and passes
  when main hasn't moved or moved only in docs. Tests use a real git fixture
  with a remote.
- A fetch failure is reported as inconclusive, never as up to date.
- `--no-integrate` bypasses it and is listed in `sdlc close --help`'s gate catalog.

## Plan

Durable plan: `workshop/plans/000269-integrate-gate-plan.md` (previous-noon rule, tri-state probe, shared `internal/freshness` for #303, gate at close + milestone-close).

- [ ] `internal/freshness` (Cutoff, LoadLocation, Check) with unit and real-git tests
- [ ] `integrategate.go` wired into `computeClose`; `--no-integrate` on close and milestone-close; fixture tests incl. remedy-from-refused-state
- [ ] gatesig catalog, help text, recovery catalog, atlas, `.sdlc/freshness.json`
- [ ] targeted tests, then full `make test`

## Log

### 2026-09-28

- Filed from pair#247 at the operator's request, after a discussion of
  whether to rebase before close. Checked the current behavior: milestone
  verdicts survive a rebase (message-matched), and a post-close rebase trips
  the publish gate.

### 2026-10-09
- 2026-10-09: unclaimed: TL edit published; not started

Spec line added by the TL (ariadne:1) at #304's request (its Done-when), with the previous-noon freshness direction from the ariadne-robustness-1 review.

### 2026-10-10
- Claimed at ariadne:4 (TL dispatch). Design: the previous-noon rule from project req. 6 subsumes the spec's "main moved since merge base". Only main commits landed before the most recent noon must be in the branch, and docs-only staleness passes. Applied at both boundaries, since #304 made milestone integration free. Fetch failure is inconclusive: warn, never the up-to-date line.
