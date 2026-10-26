---
id: 000641
status: codecomplete
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: 'ea1aeaad0a6cfdbcb35fd1a9d8f3597631e99aed' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T17:26:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "ea365fd5", done: "7fb9bde4"}
actual_hours: 0.57
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

- [x] Harness first: runner model in portable-ci.test.sh (conditions evaluated, hosted and self-hosted runs), red on the new asserts
- [x] Workflow edits S1–S5
- [x] Harnesses green; atlas/workflow/ci-merge-check.md notes CI_RUNS_ON

## Revisions

- 2026-10-25 (close review round 1, BR-2): S6's runs-on assertion for local-gates-guard.yml lives in scripts/test/portable-ci.test.sh, which checks both seeds against one RUNS_ON constant; local-gates-guard.test.sh is unchanged (a second copy would restate the expression).

## Log


### 2026-10-25
- 2026-10-25: closed — Round 4 Minor fixed in fd6e1970: the missing-go PATH dedupe uses os.path.lexists so a dangling system symlink cannot raise. portable-ci.test.sh green locally; hosted CI reruns on the push. (fixed-then-ship) (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-25: closed — Hosted merge-check on PR #448 failed in the missing-go case: the Ubuntu image has a go in its system dirs, so command -v found it. Fixed in f085a20c: the case runs on the fixture commands plus /usr/bin and /bin symlinks without go. portable-ci.test.sh green locally; hosted CI reruns on the push. 756fbd92 removes a stray file an unexpanded glob created in f085a20c. (fixed-then-ship) (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-25: closed — Round-2 Minor fixed after close in b67a69bf9: README host requirements (no pinned weave on the host; Go new enough for the pin), merge-check.yml comment and atlas say invalid JSON fails while a falsy CI_RUNS_ON reads as unset, atlas wording. Comment and docs only; portable-ci.test.sh rerun green. (fixed-then-ship) (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-25: closed — portable-ci.test.sh runs the actual merge-check blocks under a GitHub-hosted and a self-hosted runner model with conditions evaluated: self-hosted gives the same event rows, takes only actions/checkout, writes GOTOOLCHAIN=local and UV_PYTHON_DOWNLOADS=never; missing go refuses naming it; a weave without dependencies refuses with the S4 message before bootstrap; ariadne source CI refuses off GitHub-hosted. Structural asserts on both seeds (runs-on expression, guards on setup-homebrew/setup-go, self-hosted steps run no tree path, no checkout ref). Red before the workflow edits; mutations turn it red. local-gates-guard.test.sh green. Pre-close suite green on f8e342c2; 1994b7b7 adds README/atlas docs and tightens one assert (harness rerun green). Round 1 BR-1..3 fixed. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

- Filed by ariadne:2, split from #637 for echo#51; design approved by the TL as written in ## Spec (including refusing ariadne source CI off GitHub-hosted runners, and the harness evaluating step conditions over a hosted and a self-hosted runner model). Echo's full spec: echo branch 000051-ci-on-the-operator-s-self-hosted-runner-local-ci-quick-merge-checks-no-host, workshop/plans/000051-self-hosted-ci-plan.md.
- Paused by the fleet wind-down (operator, 10-25) before start-plan; no code written. Next step: `sdlc claim --issue 641`, `sdlc start-plan --issue 641`, then the Plan's first box: in scripts/test/portable-ci.test.sh replace the exact-string condition matches (`steps.gateway.outputs.pin2 != ''`, `source_ci == 'true'`) with an evaluator over a runner model (`runner.environment` github-hosted vs self-hosted), run the real blocks under both, and add the S6 structural asserts red before editing the two workflows. Landing needs the operator's local-gates-verified label (TL brings it); send echo:2 the merge commit.
- Resumed under the move plan's phase A1 (TL, operator-approved). Harness first: portable-ci.test.sh now parses every step (uses ones too) and evaluates conditions over a runner model (== / != conjunctions over runner.environment and gateway outputs; anything else fails the harness), runs the actual blocks GitHub-hosted and self-hosted, and asserts S6 (a)-(d) structurally for both seeds. Red on runs-on before the workflow edits; green after. Self-hosted cases: same event rows as hosted, only actions/checkout taken, GITHUB_ENV carries GOTOOLCHAIN=local and UV_PYTHON_DOWNLOADS=never; missing go refuses naming it; a weave without dependencies refuses before bootstrap; source CI refuses naming GitHub-hosted. Mutations: dropping the source-CI refusal and dropping setup-homebrew's guard each turn it red. local-gates-guard.test.sh green.
- Close review round 1 FIX-THEN-SHIP: BR-1 README now documents CI_RUNS_ON (gh variable set, host tools, private repositories only); the reviewer's fork-PR note added to README and atlas; BR-2 Revisions entry; BR-3 the NO_DEPENDENCIES case matches the S4 message.
- Close review round 2 SHIP with one Minor (operator-doc claims): README no longer says the host needs the pinned weave (the job builds it; the host needs a Go new enough for the pin), and the workflow comment and atlas say invalid JSON fails while a falsy CI_RUNS_ON reads as unset; atlas wording nit fixed. Fixed after the close, docs and a comment only.
- Merge refused: hosted merge-check failed in portable-ci.test.sh's missing-go case. The GitHub-hosted Ubuntu image has a go in its system directories, so command -v go found it and real go ran on the fixture (macOS has none, so it passed locally). The missing-go PATH is now the fixture commands plus symlinks of /usr/bin and /bin without go, and nothing else.
- Round 4 SHIP with one Minor: the missing-go PATH dedupe used Path.exists(), which follows symlinks, so a dangling system link would make symlink_to raise on a merged-/usr host; now os.path.lexists.
