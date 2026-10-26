---
id: 000641
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: 'c1322710375963f49fd67738f5af10e63bc45467' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T17:26:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# Seeded merge-check and local-gates-guard workflows run on a self-hosted runner (CI_RUNS_ON), split from #637 for echo#51

## Problem

GitHub Actions stopped starting hosted jobs on the private repositories (10-25, a billing block), so echo's merges stop at the CI gate. echo#51 moves echo's CI to the operator's self-hosted runner. The seeded `merge-check` and `local-gates-guard` workflows hardcode `runs-on: ubuntu-latest` and install tools on the host (setup-homebrew, setup-go, brew bundle), which must never run on the operator's machine. Split out of ariadne#637 so it lands first (TL, 10-25); echo#51 bumps `.sdlc/weave-compiler-ref` to the commit where this lands.

## Spec

Both seeded workflows can run on a self-hosted runner, chosen per repository, with installed tools only; GitHub-hosted behaviour is unchanged when the variable is unset. From echo:2's spec (S1–S7), approved by the TL:

- **S1 Runner choice, both jobs:** `runs-on: ${{ vars.CI_RUNS_ON && fromJSON(vars.CI_RUNS_ON) || 'ubuntu-latest' }}`. Unset → ubuntu-latest; a malformed value fails visibly, never falls back to hosted. echo sets `["self-hosted","local-ci"]`.
- **S2 merge-check host-mutating steps** carry `runner.environment == 'github-hosted'` (combined with existing conditions): setup-homebrew and the three setup-go steps. The ariadne source-CI brew bundle is not skipped: the gateway step refuses source CI on a runner that is not GitHub-hosted, naming why (skipping the bundle would build the candidate without its tools). ariadne's own CI stays hosted.
- **S3** after checkout, self-hosted only: `command -v` go, git, python3 (fail naming each missing tool); print `go version`; write `GOTOOLCHAIN=local` and `UV_PYTHON_DOWNLOADS=never` to `GITHUB_ENV`. A host Go older than a pin's go.mod fails in `go build` with go's message naming the minimum, never downloads.
- **S4** after the pinned compiler build and before bootstrap.sh, self-hosted only: the same test bootstrap.sh line 6 makes (`command -v weave` and `weave dependencies --help`), failing otherwise, so bootstrap's brew-install branch is unreachable. bootstrap.sh unchanged.
- **S5 local-gates-guard** changes only runs-on (S1): pull_request_target, base checkout, PR head fetched as a ref, only `git diff --name-only` read.
- **S6 tests:** `scripts/test/portable-ci.test.sh` evaluates step conditions over a runner model and runs the actual blocks hosted and self-hosted; structural asserts that (a) runs-on in both seeds is the S1 expression, (b) every setup-homebrew/setup-go step carries the github-hosted guard and source CI refuses off hosted, (c) the S3 and S4 steps exist with the not-hosted guard, (d) no guard step checks out the PR head or runs a path from the tree. `scripts/test/local-gates-guard.test.sh` asserts its runs-on.
- **S7 out of scope:** permissions stay `contents: read`; no suite status; echo's own hook decides quick vs full.

## Done when

- Both seeds take runs-on from `vars.CI_RUNS_ON` with the ubuntu-latest default; host-mutating steps run only GitHub-hosted; the self-hosted tool checks exist; source CI refuses off hosted.
- portable-ci.test.sh runs the actual blocks under both runner models and passes; local-gates-guard.test.sh passes.
- Landed with the operator's local-gates-verified label; echo:2 has the merge commit.

## Plan

- [ ] Harness first: runner model in portable-ci.test.sh (conditions evaluated, hosted and self-hosted runs), red on the new asserts
- [ ] Workflow edits S1–S5
- [ ] Harnesses green; atlas/workflow/ci-merge-check.md notes CI_RUNS_ON

## Log


### 2026-10-25

- Filed by ariadne:2, split from #637 for echo#51; design approved by the TL as written in ## Spec (including refusing ariadne source CI off GitHub-hosted runners, and the harness evaluating step conditions over a hosted and a self-hosted runner model). Echo's full spec: echo branch 000051-ci-on-the-operator-s-self-hosted-runner-local-ci-quick-merge-checks-no-host, workshop/plans/000051-self-hosted-ci-plan.md.
- Paused by the fleet wind-down (operator, 10-25) before start-plan; no code written. Next step: `sdlc claim --issue 641`, `sdlc start-plan --issue 641`, then the Plan's first box: in scripts/test/portable-ci.test.sh replace the exact-string condition matches (`steps.gateway.outputs.pin2 != ''`, `source_ci == 'true'`) with an evaluator over a runner model (`runner.environment` github-hosted vs self-hosted), run the real blocks under both, and add the S6 structural asserts red before editing the two workflows. Landing needs the operator's local-gates-verified label (TL brings it); send echo:2 the merge commit.
