---
id: 000345
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '1da5f3fb7905297fce4aeb2674bd80eb15b68908' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T12:49:57-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "76100a65", done: "a03b5fee"}
---

# sdlc merge: a recorded local run substitutes for a CI job that never got a runner (not for one that failed)

## Problem

`sdlc merge` (#309) refuses while a required CI check is red or pending. On 2026-10-13 pair#523 sat for over an hour because GitHub's macOS arm64 runners were short: pair's `couch-zellij-conformance` job was 'not acquired by a hosted runner' and cancelled after 15 minutes with zero steps run. That job is the only routine run of pair's live Couch/Zellij tests (`make -f Makefile.local test-couch-zellij-live`, gated by PAIR_LIVE_COUCH=1, and `test-native-terminal-ci`): local `make test` skips them. Every slot runs on the same Mac, so the same tests can run locally.

## Spec

The operator (2026-10-13): a CI job that **could not get a runner** (no steps executed; cancelled/queued past a timeout) may be substituted by a **recorded local run** of the repo-declared equivalent target; a job that **ran and failed** still blocks.
1. A repo declares, per CI job, its local equivalent (e.g. pair: `couch-zellij-conformance` → `make -f Makefile.local test-couch-zellij-live test-native-terminal-ci`).
2. `sdlc merge` distinguishes 'no runner / zero steps' from 'failed' using the check run's data; for the former it offers the substitute: run the local target, record its exit and log as evidence on the PR (a comment or check), then proceed.
3. Pending-but-running jobs still wait; failed jobs still refuse; `--no-validate` stays forbidden.
4. pair's docs (AGENTS/README) say which local target substitutes for the macOS job, and that it is also worth running before close when a change touches Couch's session/Zellij code.

## Done when

- A merge whose required macOS job never got a runner proceeds after a recorded passing local run of the declared target; a failed job still blocks; tests cover both; pair declares its mapping.

## Plan

Durable plan: `workshop/plans/000345-local-check-substitutes-plan.md` (single pass, one close).

- [ ] Declaration parser (`.sdlc/local-checks.json`, strict)
- [ ] Runnerless classifier (structural: cancelled + no runner + zero steps)
- [ ] Gate decision with local evidence (`ciLocalNeeded`, `required`)
- [ ] gh seam (JobRun, PostStatus, PRComment) + stateful `ciFakeGH` + live conformance
- [ ] Local run shell (detached worktree at the head, test-lock slot, log, status)
- [ ] Wire into `runCIGate` for both merge paths (inline run, re-read, head pin)
- [ ] `sdlc local-check run` verb
- [ ] Atlas + help
- [ ] Verify, file the pair declaration issue unassigned, close, merge

## Log

### 2026-10-13

- Measured the runnerless signal: pair run 38453641991 attempt 1, job 115490631095 — check-run `conclusion=cancelled`; jobs API `runner_name=""`, `steps=0`; annotation "not acquired by Runner of type hosted". `gh pr checks` buckets it `cancel`, which `decideCIGate` (cmd/sdlc/cigate.go) reads as red today. Attempt 2 got a runner (14 steps, success).
- pair `main` has no branch protection, so `sdlc merge`'s CI gate is the only gate; a substitute there is enough.
- `local-gates-guard` already protects `.sdlc/*` → a declaration there is CI-tamper-guarded (the Revision's "CI is the tamper guard").
- Draft design (declaration in `.sdlc/local-checks.json`, `sdlc local-check run` posting an `sdlc-local/<name>` commit status on the head, gate accepts it only for runnerless cancels) sent to the hub for review before building.

- Hub approved the design (2026-10-13). (a) The pair declaration is a separate pair issue, filed unassigned at close. (b) merge may run a local check inline, but through the machine-wide suite lock, printing the command and log path first.

## Revisions

- 2026-10-13 — the operator's stance broadens this: for a small company, **local runs are the primary verification**
  (faster, more controlled iteration on the operator's Mac); **CI mainly guards that what agents aren't supposed to change
  hasn't changed** (#309's local-gates-guard, merge-check's integrity checks). So design the recorded local run as first-class
  evidence for heavy runtime suites, not only as a fallback, and keep CI as the tamper guard. Which jobs stay required on CI
  is the operator's call per repo.
