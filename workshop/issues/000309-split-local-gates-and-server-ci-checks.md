---
id: 000309
status: open
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: 'd1833645db104bd132afaa12323bc41d6a8620d5' # card fields mirrored from issue-cards; edit via sdlc
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

## Log

### 2026-10-09
Filed from the robustness evidence review (ariadne-robustness-1). Supersedes #052.
