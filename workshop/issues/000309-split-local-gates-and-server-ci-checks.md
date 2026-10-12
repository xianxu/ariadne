---
id: 000309
status: working
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-12
estimate_hours:
card_mirror: '798c4e9910cbddf6fc6a5dc5a3e8086ea4be5df0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T11:13:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# Split local gates and server CI checks

## Problem

CI has been red in several repos for long stretches, and `sdlc merge` landed PRs over red or pending checks. #263 and #241 merged with failing checks (evidence `a2-0928:5142`, `:5047`); the id-lint CI failure became #266.

#052, the generic CI merge-check mechanism, shipped its M1 in May but was never ticked, and its M2 is stale.

The broader question was never settled: which checks run locally and which on the server. Ariadne strongly controls the local process, so local gates can do the heavy work. The server's job is mainly to make sure nobody quietly weakened the local gates.

## Spec

A design issue, to be refined at start-plan. Starting point (evidence file `workshop/pensive/2026-10-09-01-pensive-sdlc-robustness-evidence.md`, Part 1, under #052):
- **Local:** sdlc gates do the heavy checks (review, plan quality, tests).
- **Server CI** does three cheap things:
  1. deterministic checks (build, lint, id-lint);
  2. asserts the local-checks manifest is unchanged (gate config, `merge-checks.d/`, hooks, weave-generated gate surface), unless the PR carries an explicit operator-verified marker (cf. #297 `verify: operator`);
  3. `sdlc merge` refuses red or pending CI, and names the failing check.
- **Changing local checks legitimately:** a dedicated issue whose verification is operator-held.
- Fix the currently red CI in each ariadne-style repo as part of the rollout.

Replaces #052 (close it as superseded).

## Done when

- `sdlc merge` refuses a PR with red or pending required checks.
- Server CI passes on main in ariadne, pair, parley.nvim and tools.
- A PR that edits the local-checks manifest without the operator marker fails CI with a clear message.

## Plan

Durable plan: `workshop/plans/000309-split-local-gates-and-server-ci-checks-plan.md`.

- [ ] M1 — `sdlc merge` refuses red or pending CI on the PR head, both landing paths; `--no-ci` bypass
- [ ] M2 — `local-gates-guard.yml` on `pull_request_target` with the `local-gates-verified` label; merge-check on push to main; labels in four repos; atlas + AGENTS
- [ ] After landing: live check with a DO-NOT-MERGE PR; report seeded files to the hub

## Log

### 2026-10-09
Filed from the robustness evidence review (ariadne-robustness-1). Supersedes #052.

### 2026-10-12
Claimed at ariadne:2. Measured: merge-check is green on recent PRs in all four repos (no red CI to fix); only pair runs CI on push to main; `sdlc merge` reads no checks on either path. TL answers: label marker default (operator confirming), honest limit goes into the atlas verbatim; I create labels in all four repos, the TL runs the peer weave refreshes; one DO-NOT-MERGE live-check PR after landing is approved; push-to-main trigger approved.
