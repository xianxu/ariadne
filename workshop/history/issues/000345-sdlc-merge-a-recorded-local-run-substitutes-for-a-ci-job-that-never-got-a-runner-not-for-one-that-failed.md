---
id: 000345
status: done
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '832144b8f903ed12a909cd033471cc3213eb91ae' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T12:49:57-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.61
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

Durable plan: `workshop/plans/000345-sdlc-merge-a-recorded-local-run-substitutes-for-a-ci-job-that-never-got-a-runner-not-for-one-that-failed-plan.md` (single pass, one close).

- [x] Declaration parser (`.sdlc/local-checks.json`, strict)
- [x] Runnerless classifier (structural: cancelled + no runner + zero steps)
- [x] Gate decision with local evidence (`ciLocalNeeded`, `required`)
- [x] gh seam (JobRun, PostStatus, PRComment) + stateful `ciFakeGH` + live conformance
- [x] Local run shell (detached worktree at the head, test-lock slot, log, status)
- [x] Wire into `runCIGate` for both merge paths (inline run, re-read, head pin)
- [x] `sdlc local-check run` verb
- [x] Atlas + help
- [x] Verify, file the pair declaration issue unassigned, close, merge

## Log

### 2026-10-13
- 2026-10-13: closed — Full make test on 34e4cff2 (review fixes + origin/main with #346): exit 0, all 12 cmd/sdlc shards (1120 tests) and all 46 other packages green. Targeted #345 suites green: parser, runnerless classifier, decideCIGate table incl. provenance (an Actions job named sdlc-local/x is not a recorded run), runCIGate against the stateful ciFakeGH (inline run on the pinned head; head moved, unrecorded pass, head-only declaration, invalid declaration, unreadable job, and a missing required verdict before any run all refuse), runLocalCheck against real git, merge e2e for both Done-when paths (TestRunMerge_RunnerlessJobMergesAfterLocalRun, TestRunMerge_FailedJobStillBlocks) plus the landing path, the verb (--rev must be an origin branch tip). Live: TestLiveJobRunConformance passed against pair jobs 115490631095 (runnerless) and 115497743342 (ran). Round-1 review findings (1 Important, 5 Minor) all fixed. Pair declaration filed as pair#534.; review verdict: FIX-THEN-SHIP
- 2026-10-13: flow upgraded quick → full — 927 added lines in code files (limit 100); an earlier round of this close already ran the full review

- Measured the runnerless signal: pair run 38453641991 attempt 1, job 115490631095 — check-run `conclusion=cancelled`; jobs API `runner_name=""`, `steps=0`; annotation "not acquired by Runner of type hosted". `gh pr checks` buckets it `cancel`, which `decideCIGate` (cmd/sdlc/cigate.go) reads as red today. Attempt 2 got a runner (14 steps, success).
- pair `main` has no branch protection, so `sdlc merge`'s CI gate is the only gate; a substitute there is enough.
- `local-gates-guard` already protects `.sdlc/*` → a declaration there is CI-tamper-guarded (the Revision's "CI is the tamper guard").
- Draft design (declaration in `.sdlc/local-checks.json`, `sdlc local-check run` posting an `sdlc-local/<name>` commit status on the head, gate accepts it only for runnerless cancels) sent to the hub for review before building.

- Hub approved the design (2026-10-13). (a) The pair declaration is a separate pair issue, filed unassigned at close. (b) merge may run a local check inline, but through the machine-wide suite lock, printing the command and log path first.
- Built (2026-10-13): `localchecks.go` (pure: strict parse, runnerless, records), `localcheck_run.go` (shell: detached worktree, suite lock via `Makefile.workflow`'s helper, log, status), `runCIGate` reads main's declaration and runs a missing check inline, then re-reads (ARCH-ORDER: a push during the run or an unrecorded pass refuses), `sdlc local-check run`. ARCH-DRY: reused `abbrevSHA` and `stubGH`; the status-context check derives from the name regex. ARCH-MOCK: `ciFakeGH` models GitHub's status rollup; `TestLiveJobRunConformance` passed live against pair jobs 115490631095 (runnerless) and 115497743342 (ran).
- The plan's live `PostStatus` self-test is dropped as redundant: that `gh pr checks` lists commit statuses by context is already load-bearing and live for `local-gates-guard/verdict` (#309), so posting a junk status on this PR would prove nothing new.
- A background merge/landing run reported all PASS but tripped the hermeticity guard, because I committed in this tree while it ran (HEAD 8477263c → 46bcbfb2, my own commit). Not a leaking test.
- Full `make test` on 40debd85: all 12 cmd/sdlc shards green (1116 tests); one failure, `internal/fleet TestHangingRemoteDegradesOnlyItsRepository` (3.6s git deadline under load ~150), passed 5/5 in isolation, and the branch does not touch that package.
- Filed pair#534 unclaimed (from pair:0, deps ariadne#345): declare `couch-zellij-conformance / conformance` → `make -f Makefile.local test-couch-zellij-live test-native-terminal-ci` in `.sdlc/local-checks.json`, plus the AGENTS.local/README lines; it needs the operator's `local-gates-verified` label.
- Boundary review round 1: FIX-THEN-SHIP, 1 Important + 5 Minor, all fixed in-round. BR-1: README gains the local-check paragraph. Recorded passes count only from commit statuses (an Actions job named `sdlc-local/x` no longer satisfies; ARCH-SECURE). `--rev` must be an origin branch tip. A missing required verdict refuses before any inline run. Fixed the atlas table cell. The fakes share `ciSeamGH` (nil-safe) instead of per-fake delegation (ARCH-DRY). Targeted suites green (656s).

## Revisions

- 2026-10-13 — the operator's stance broadens this: for a small company, **local runs are the primary verification**
  (faster, more controlled iteration on the operator's Mac); **CI mainly guards that what agents aren't supposed to change
  hasn't changed** (#309's local-gates-guard, merge-check's integrity checks). So design the recorded local run as first-class
  evidence for heavy runtime suites, not only as a fallback, and keep CI as the tamper guard. Which jobs stay required on CI
  is the operator's call per repo.
