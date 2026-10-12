---
id: 000309
status: working
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-12
estimate_hours: 2.76
card_mirror: '8f47c0e4c6466e949775b711a6336aafa9e5d0d4' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T11:13:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
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

- [x] M1 — `sdlc merge` refuses red or pending CI on the PR head, both landing paths; `--no-ci` bypass
- [x] M2 — `local-gates-guard.yml` on `pull_request_target` with the `local-gates-verified` label; merge-check on push to main; labels in four repos; atlas + AGENTS
- [ ] After landing: live check with a DO-NOT-MERGE PR; report seeded files to the hub

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec                 design=1.0  impl=0.08
item: smaller-go-module          design=0.06 impl=0.2
item: skill-or-dispatcher        design=0.2  impl=0.2
item: real-api-discovery         design=0.0  impl=0.2
item: cross-repo-refactor-small  design=0.04 impl=0.08
item: atlas-docs                 design=0.04 impl=0.06
item: milestone-review           design=0.0  impl=0.2
item: milestone-review           design=0.0  impl=0.2
design-buffer: 0.15
total: 2.76
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* Items: issue-spec is this session's design and two review rounds; smaller-go-module is the CI gate, the `gh` adapter and the head pin on both merge paths (design ×0.2, the plan resolves it); skill-or-dispatcher is the guard workflow and its executed-run-block test; real-api-discovery covers `gh pr checks` exit codes and `pull_request_target` check reporting; cross-repo-refactor-small is the push trigger, the seed row and the labels in four repos.

## Log

### 2026-10-09
Filed from the robustness evidence review (ariadne-robustness-1). Supersedes #052.

### 2026-10-12
- 2026-10-12: closed M2 — Round 3. make test: exit 0 (1113 cmd/sdlc tests + 46 packages). BR-11: guard posts pending first, EXIT trap posts failure on any undecided exit (local-gates-guard.test.sh: pending-then-exactly-one-final; unfetchable PR ref ends in failure; deleting the trap fails the test); merge requires local-gates-guard/verdict when the target main carries the guard (TestRunCIGate missing/present/pending/unreadable-main-tree). Runner merge-checks.d/20-ci-workflow-tests.sh globs scripts/test/*.test.sh with two named exclusions; both included tests PASS. Classifier case added for the missing-verdict refusal. Earlier rounds: -z/quotePath, live labels, substring and matcher mutations all caught.; review verdict: SHIP
- 2026-10-12: closed M1 — Round 2. go test -run CIGate|TestRunMerge|TestLanding|DecideCIGate|ParsePRChecks: ok (399s); merge/tracker/landing/external-merge e2e set: ok (412s); processmanual ok. Mutation: removing either merge path runCIGate call alone fails its per-site test. BR-1 covered by TestRunCIGate lag/move/bound cases and TestRunMerge_LaggingPRHeadRefuses. make test (round 1): all shards pass except TestPlanningReviewCompetingResults, which fails identically on clean origin/main c1afb049 (pre-existing, reported to hub).; review verdict: SHIP
Claimed at ariadne:2. Measured: merge-check is green on recent PRs in all four repos (no red CI to fix); only pair runs CI on push to main; `sdlc merge` reads no checks on either path. TL answers: label marker default (operator confirming), honest limit goes into the atlas verbatim; I create labels in all four repos, the TL runs the peer weave refreshes; one DO-NOT-MERGE live-check PR after landing is approved; push-to-main trigger approved.
M1 implemented: `cmd/sdlc/cigate.go` (pure `decideCIGate`, `parsePRChecks`, `runCIGate`), `PRChecks` on the `ghCaller` seam, `PRMerge` pinned with `--match-head-commit`, `--no-ci` in the gate catalog. Mutation-checked: removing either site's `runCIGate` call alone fails `TestRunMerge_RedCIRefusesPreMerge` / `TestLandingRedCIRefusesBeforeMerge`. Revision: dropped "dry-run reports what the gate would say" — neither path's dry-run contacts GitHub for PR state today, and adding that is out of this issue's contour.
M1 review round 1 (FIX-THEN-SHIP): BR-1 fixed in the class — `runCIGate` brackets the checks read with two PR-head reads and refuses on any mismatch (both paths); tests `TestRunCIGate` (lag, move-during-read, bound green) and `TestRunMerge_LaggingPRHeadRefuses`. Minors fixed: plan Revisions entry, catalog refusal pattern now the `CI gate: ` prefix shared by every refusal, `exitErr` uses strconv.
M2 implemented: `.github/workflows/local-gates-guard.yml` (pull_request_target, read-only contents, statuses write; PR values via env:, fetches refs/pull/N/head, never runs PR code; posts a `local-gates-guard/verdict` status on the head), `scripts/test/local-gates-guard.test.sh` (executes the real run block: 16 patterns each with a generated sample and near miss, label × touched, deletion, hostile PR number; mutation-checked on the label and the matcher), merge-check on push to main over before..after (portable-ci.test.sh covers push and zero-before), seed row in base.manifest, atlas + AGENTS.base.md line. Labels `local-gates-verified` created in ariadne, pair, parley.nvim, tools. Duplicate-id check verified clean on two push ranges of main (c1afb049..bf837e55, 89213b48~1..89213b48). M1 advisory BR-3 folded in: classifier cases for every `CI gate: ` refusal, the ack, and the success lines; reverting the pattern fails 3.
Lesson for me: the full suite's hermeticity guard reads the real working tree, so editing during `make test` fails the run; park edits until it finishes.
M2 review round 1 (FIX-THEN-SHIP, 4 Important): BR-5 paths read `-z` with quotePath off (non-ASCII, quote and backslash cases; reverting fails the test); BR-6 labels read live with fail-closed exit 2, whole-name match (substring mutation fails the test), per-PR concurrency group; BR-7 README; BR-8 Revisions sweep of all M2 deviations. Minors: tests wired via `merge-checks.d/20-ci-workflow-tests.sh`, head-moved case tested; main-tip attachment goes to the live check. The reviewer, with mktemp failing, wrote ~/.git and ~/scripts; my rm and rename were refused by the classifier, so it went to the hub for the operator.
M2 review round 2 (FIX-THEN-SHIP, BR-11): guard posts pending first, EXIT trap posts failure on any undecided exit (tests: pending-then-one-final, unfetchable PR ref ends in failure); merge requires `local-gates-guard/verdict` whenever main carries the guard (`requiredChecks`, TestRunCIGate missing/present/pending/unreadable). Minor: the runner globs every scripts/test/*.test.sh with two named exclusions; plan wording no longer claims pre-push. Operator confirmed the design as the default merge behavior (TL, 10-12).
M2 advisory (stale origin/main on the legacy path): the gate now fetches `origin main` into `refs/remotes/origin/main` before deciding the verdict requirement. `TestRunMerge_GuardOnFreshMainRequiresVerdict` proves the requirement follows main as fetched; removing the new fetch alone does not fail it, because the in-place fixture's step 9b fetch already refreshes the ref. The linked-worktree topology, the case the fetch exists for, has no e2e fixture.
