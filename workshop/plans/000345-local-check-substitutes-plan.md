# Recorded Local Checks: First-Class Merge Evidence Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A repository declares local checks (the local equivalents of heavy CI jobs, or local-only suites). `sdlc local-check run` runs one on the exact PR head and records the outcome as a GitHub commit status. `sdlc merge` accepts that recorded passing run in place of a CI job that never got a runner, and requires it for checks declared `required`. A CI job that ran and failed still blocks.

**Architecture:** A pure core in `localchecks.go` holds three pieces: the typed declaration parsed from `.sdlc/local-checks.json`, the job-runnerless classifier, and the extended gate decision. A thin IO shell in `localcheck_run.go` creates a detached worktree at the head, takes the machine-wide test-lock slot, runs the argv, keeps the log, and posts the status (and a comment on failure). `runCIGate` reads the declaration from the base tree, which is fetched `origin/main` and never the PR head. It asks GitHub about runner acquisition only for cancelled checks that have a declared substitute, and runs a missing local check inline through the same shell. Both merge paths (`merge.go`, `landing.go`) get this through `runCIGate`, so neither is edited beyond passing one dependency.

**Tech Stack:** Go (`cmd/sdlc`, cobra), `gh` CLI (REST: `actions/jobs/{id}`, `statuses/{sha}`, `issues/{n}/comments`), `git worktree`, `scripts/test-lock.py` (#319/#341).

**Approved design:** the hub (ops:0) approved it on 2026-10-13 (issue `## Log`). The pair declaration ships as a **separate pair issue** filed unassigned at close. When merge runs a check inline, it goes through the machine-wide suite gate and prints the exact command and log path first.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `localCheckDecl` / `localCheck` | `cmd/sdlc/localchecks.go` | new |
| `parseLocalChecks` | `cmd/sdlc/localchecks.go` | new |
| `jobRun` / `jobRun.runnerless` | `cmd/sdlc/localchecks.go` | new |
| `jobIDFromLink` | `cmd/sdlc/localchecks.go` | new |
| `localStatusContext` | `cmd/sdlc/localchecks.go` | new |
| `decideCIGate` (+ `ciInputs`, `ciVerdict.LocalNeeded`) | `cmd/sdlc/cigate.go` | modified |
| `localRunRecord` / `.statusDescription` / `.failureComment` | `cmd/sdlc/localchecks.go` | new |

- **localCheckDecl / localCheck**: the parsed declaration. `{Version int; Checks []localCheck}`, where `localCheck = {Name string; Command []string; Substitutes []string; Required bool; TimeoutMinutes int}`. Tests in `localchecks_test.go` need no IO.
  - **Relationships:** 1 decl per repo (read from the base tree). 1 check : N substituted CI check labels. A CI label may be substituted by at most one check (the parser refuses duplicates).
  - **DRY rationale:** this is the single source for "which local run stands in for which CI job". The verb, the gate and the docs all read it, and none restates the mapping.
  - **Future extensions:** a per-check `paths:` filter (run only when the diff touches them), mirroring the CI workflow's `paths:`. Not now (YAGNI).
- **parseLocalChecks(data []byte) (localCheckDecl, error)**: strict parse (ARCH-SECURE). It uses `json.Decoder.DisallowUnknownFields` and requires `version == 1`. Names must match `^[a-z0-9][a-z0-9-]{0,40}$`, because they become a status context and a path segment. `Command` must be non-empty argv with no empty elements. Labels must be non-empty and are unique across checks. `TimeoutMinutes` falls in 1..120 and defaults to 30. A missing file is an empty decl. A present but invalid file is an error, and the gate refuses on it rather than ignoring it.
- **jobRun{Status, Conclusion, RunnerName string; Steps int}** with **runnerless() bool**: true iff `Status=="completed" && Conclusion=="cancelled" && RunnerName=="" && Steps==0`. This is the structural signal measured on pair job 115490631095. Annotation text is never read.
- **jobIDFromLink(link string) (int64, bool)**: parses `https://github.com/<o>/<r>/actions/runs/<run>/job/<id>`. A status or third-party check link gives `false`, and such a check is never runnerless.
- **localStatusContext(name) string** = `"sdlc-local/" + name`. It is the one place the context string is built. The verb writes it and the gate reads it.
- **decideCIGate(in ciInputs) ciVerdict** (modified signature). `ciInputs{Checks []prCheck; HeadHasWorkflows bool; Local localCheckDecl; Runnerless map[string]bool /* by checkLabel */}`. The rules:
  1. A `cancel` check whose label is in `Runnerless` and has a declared substitute is *satisfied* if the checks contain `localStatusContext(sub)` with bucket `pass`. Otherwise the substitute goes into `LocalNeeded`. It does not add to `Offender`.
  2. Every `Required` local check without a `pass` status goes into `LocalNeeded`.
  3. Everything else follows today's ranking unchanged. A recorded local *failure* is a `fail` check, so it is red.
  4. New reason `ciLocalNeeded` ranks below red/unknown/pending. A red or pending CI check still wins, so nothing runs locally while CI is red or still running.
  - `checks == []` with `Required` local checks gives `ciLocalNeeded` rather than `ciNoWorkflows`, because a local-only repo still needs its evidence.
- **localRunRecord{Name, SHA string; Exit int; Duration time.Duration; LogSHA256, LogPath, Host string; Tail []string}**: the outcome of one run. `statusDescription()` stays within GitHub's 140-char limit ("exit 0 in 4m12s · log sha256:abcd1234… · host"). `failureComment()` renders a bounded body: at most 60 tail lines and 8 KiB.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `ghCaller.JobRun` | `cmd/sdlc/cigate.go` (realGH), `ghclient.go` (interface) | new | `gh api repos/{r}/actions/jobs/{id}` |
| `ghCaller.PostStatus` | same | new | `gh api -X POST repos/{r}/statuses/{sha}` |
| `ghCaller.PRComment` | same | new | `gh pr comment` |
| `localRunner` (`runLocalCheck`) | `cmd/sdlc/localcheck_run.go` | new | `git worktree`, `scripts/test-lock.py`, `exec` of the declared argv |
| `readLocalChecks(git, rev)` | `cmd/sdlc/localchecks.go` (thin) | new | `git show <rev>:.sdlc/local-checks.json` |
| `ciFakeGH` | `cmd/sdlc/cigate_fake_test.go` | new | stateful fake of the above + PRChecks/PRHeadOID |
| `newLocalCheckCmd` | `cmd/sdlc/localcheck_cmd.go` | new | cobra verb `sdlc local-check run` |

- **ghCaller.JobRun / PostStatus / PRComment**: added to the existing seam. `stubGH` (which every existing fake embeds) gets "not implemented" defaults, so the four existing fakes compile unchanged (ARCH-DRY: no fifth copy of PRChecks behaviour).
  - **Injected into:** `runCIGate`, `runLocalCheck`.
- **ciFakeGH** (ARCH-MOCK): the stateful fake. Its state is `heads map[int]string` (PR → head), `checks map[string][]prCheck` (sha → CI checks), `statuses map[string]map[string]string` (sha → context → state), `jobs map[int64]jobRun` and `comments []string`. `PRChecks` returns the CI checks plus the statuses for the PR's current head, with later statuses for the same context replacing earlier ones, which is GitHub's rollup. So a `PostStatus` made by the gate's inline run is visible to the next read, the interaction bug a function mock would miss. A `PostStatus` for a SHA the fake doesn't know returns the 422 GitHub gives for an unpushed commit.
- **Live conformance (ARCH-MOCK):** `TestLiveJobRunConformance` is skipped unless `SDLC_LIVE_GH=1`. It reads pair job 115490631095 through `realGH.JobRun` and asserts `runnerless()` is true. It also reads job 115497743342 (same run, attempt 2) and asserts false. This catches drift in the jobs API shape the classifier depends on. It's run by hand before close and noted in `--verified`.
- **localRunner / runLocalCheck(ctx, deps, check, sha, repoRoot) (localRunRecord, error)**: the IO shell.
  1. `git worktree add --detach <tmp> <sha>` in a fresh `os.MkdirTemp`, removed with `git worktree remove --force` in a `defer` and followed by `git worktree prune`. This guarantees the run is of exactly the pinned SHA on a clean tree, whatever the slot's checkout is (the landing path runs from the resting branch).
  2. Print `local check <name>: <argv> in <worktree> → log <path>`, then take the suite slot. `python3 <helper> acquire --watch <os.Getpid()>`, with the helper resolved exactly as `Makefile.workflow` does: `dir(realpath(<root>/Makefile.workflow))/scripts/test-lock.py` (ARCH-DRY: the same resolution and the same lock files, not a Go re-implementation of flock). Accept `acquired|reentrant|skipped`. Anything else, or a missing helper or `Makefile.workflow`, is an error ("refusing to run an unbounded suite"). After `acquired`, set `WF_TEST_LOCK_HELD_BY=<pid>` in the child env so a nested `make test` is reentrant.
  3. `exec.CommandContext` with the argv (no shell), cwd = worktree, a timeout of `TimeoutMinutes`, and stdout+stderr teed to the log file and to stderr.
  4. Hash the log and build the `localRunRecord`.
  5. `PostStatus(sha, localStatusContext(name), success|failure, rec.statusDescription())`. On failure, also `PRComment(rec.failureComment())` when a PR number is known.
  - **Injected into:** `runCIGate` (through a `localRun func(...)` field on a new `ciGateDeps` struct, so the gate tests inject a fake runner with no exec) and `newLocalCheckCmd`.
- **readLocalChecks**: `git show <base>:.sdlc/local-checks.json`. A "does not exist" error gives an empty decl, and any other git error is returned. It reads base, not head, so a PR can't add its own substitute to waive a CI job (ARCH-SECURE). The local-gates guard also refuses an unlabeled edit of `.sdlc/*`.
- **newLocalCheckCmd**: `sdlc local-check run (--name X | --all) [--rev HEAD]`. It resolves `--rev` to a SHA and refuses unless that SHA is the tip of the current branch on origin (statuses need a pushed commit). It reads the declaration from `origin/main`, finds the PR via `PRListForBranch` (optional, used only for the failure comment), and runs. It exits non-zero on a failed check. It is registered in `main.go` and marked mutating like `merge`.

### Operating envelope (ARCH-CONSTRAINTS)

| Constraint | Budget / behavior | Basis |
|---|---|---|
| Workload | Batch, operator-initiated (`local-check run`) or inside `sdlc merge`; not interactive. | design |
| Runtime of one check | Bounded by `timeout_minutes` (default 30, max 120); killed on expiry and recorded as a failure (exit −1, "timed out"). | pair's CI job is `timeout-minutes: 20` |
| CPU/concurrency | At most `WF_TEST_SUITES` (default 2) heavy suites machine-wide, shared with `make test`. The lock wait is bounded by `WF_TEST_LOCK_TIMEOUT`, and a `timeout` answer refuses. | #319/#341 measured |
| Lock holding | Held from acquire until the `sdlc` process exits. For inline merge that adds the post-run merge seconds, which is acceptable and documented in help. | design |
| gh calls | One `JobRun` per *cancelled* check with a declared substitute (≤ number of declared labels, single digits), plus 1 status and ≤1 comment per run. | design |
| Disk | One log per (sha, check) plus one transient worktree; see ARCH-FUNERAL. | design |

### ARCH-SECURE

- **Boundaries:** (1) `.sdlc/local-checks.json` is untrusted input. It's read from the base tree, parsed into a typed value, and refused when invalid. (2) GitHub API bodies (`jobs`, `checks`) are decoded into typed structs, and a job response missing `runner_name` or `steps` decodes as non-runnerless (zero steps but a runner name present ⇒ false; an absent `steps` key ⇒ unknown ⇒ false) so that unread never means substitutable. (3) Commit statuses: anyone holding a repo token can post `sdlc-local/*`. That's the accepted trust model under the operator's stance (local runs on the operator's machine are trusted; CI guards the gate configuration). The description carries host and log hash for audit.
- **Credentials:** no new credential. `gh` uses the existing auth. The log tail can contain test output. It's posted only on failure, bounded, and noted in help; the full log stays local.
- **Argv:** the command is an argv array executed without a shell (structural separation). The worktree path comes from `os.MkdirTemp`.

### ARCH-ORDER

`runCIGate` holds no state between events beyond one call. The sequence is: read head → read checks → read head (existing BR-1 head pin) → classify → [inline runs] → **re-read head and checks** → decide again. The inline run is the long, interruptible step:
- *The head moves during the run* (a push to the branch): the post-run re-read sees a different head → `ciHeadLag` refusal. The status already posted is on the old SHA and is harmless evidence about that SHA.
- *The process is killed mid-run*: the deferred worktree removal may not run. `git worktree prune` on the next run and the ARCH-FUNERAL sweep collect it. No status is posted (the post comes after the run), so nothing claims a pass.
- *PostStatus fails after a passing run* (network): the outcome is unconfirmed on GitHub, so the gate refuses ("recorded run not visible; re-run `sdlc local-check run --name X` or retry merge"). It never proceeds on an unrecorded local pass. A retry re-runs the check; duplicate statuses are idempotent (the latest state for the context wins).
- *A second actor posts a failure for the same context*: the post-run re-read shows `fail` → red.
The event most likely to be mishandled is the head moving during a long run, and the post-run re-read exists for exactly that.

### ARCH-FUNERAL

- **Logs** `.git/sdlc/local-checks/<sha>/<name>.log`: created by `runLocalCheck` and last needed by whoever audits a failed merge. Each run prunes `<sha>` directories beyond the newest 20 (by mtime), so the bound is 20 × checks × log size. Under `.git/`, so it's never committed and dies with the clone.
- **Worktrees**: created per run and removed in a `defer`. A crash leaves one, which is collected by `git worktree prune` plus deletion of stale `sdlc-local-*` temp dirs older than a day on the next run.
- **Commit statuses / PR comments** live on GitHub, bound to the commit or PR, and are bounded by GitHub. Nothing to remove.

### ARCH-PURPOSE

The issue's purpose is that a runnerless required macOS job no longer blocks pair. That needs the pair declaration, which ships as a separate pair issue by the hub's decision (one issue, one repo, and it needs the operator's `local-gates-verified` label). It's filed at close and isn't a deferral of this repo's purpose. In this repo the purpose is fully delivered: the declaration format, the run verb, gate acceptance for both merge paths, `required` local-only checks (the Revision), docs, and the atlas.

---

## Tasks

Single pass (one `sdlc close`), plain checkboxes.

### Task 1: Declaration parser

**Files:** Create `cmd/sdlc/localchecks.go`, `cmd/sdlc/localchecks_test.go`

- [ ] Step 1: Write `TestParseLocalChecks` as a table test covering:
  - a valid entry with defaults (`timeout_minutes` 30)
  - empty input or a missing file → empty decl
  - an unknown field (`"substitute"` typo) → error naming the field
  - `version: 2` → error
  - a bad name (`"A B"`, `"../x"`) → error
  - an empty command / empty argv element → error
  - the same label in two checks → error
  - timeout 0 or 121 → error
- [ ] Step 2: `go test ./cmd/sdlc -run TestParseLocalChecks` → FAIL (undefined).
- [ ] Step 3: Implement `localCheckDecl`, `localCheck`, `parseLocalChecks` (JSON tags: `version`, `checks`, `name`, `command`, `substitutes`, `required`, `timeout_minutes`), `localStatusContext`, `localChecksPath = ".sdlc/local-checks.json"`.
- [ ] Step 4: Re-run → PASS.
- [ ] Step 5: Commit `#345: local-check declaration parser`.

### Task 2: Runnerless classifier

**Files:** `cmd/sdlc/localchecks.go`, `localchecks_test.go`

- [ ] Step 1: Write `TestJobRunnerless` with these cases:
  - completed/cancelled/""/0 → true
  - cancelled with runner "GitHub Actions 1" → false
  - cancelled with steps 3 → false
  - failure/""/0 → false
  - queued → false

  Write `TestJobIDFromLink` with these cases:
  - `https://github.com/xianxu/pair/actions/runs/38453641991/job/115490631095` → 115490631095
  - a status link, an empty string, and `…/job/abc` → false
- [ ] Step 2: Run → FAIL.
- [ ] Step 3: Implement `jobRun` (JSON: `status`, `conclusion`, `runner_name`, `steps` as `[]json.RawMessage` counted, with a `StepsKnown` bool when the key is absent), `runnerless()`, `jobIDFromLink`.
- [ ] Step 4: PASS. Commit `#345: classify a job that never got a runner`.

### Task 3: Gate decision with local evidence

**Files:** Modify `cmd/sdlc/cigate.go` (`decideCIGate`, `ciReason`, `refusal`), `cmd/sdlc/cigate_test.go`

- [ ] Step 1: Port the existing `TestDecideCIGate` cases to the `ciInputs` form (no behaviour change). Add these cases:
  1. runnerless `cancel` + substitute + `sdlc-local/zellij` pass → green
  2. runnerless `cancel` + substitute, no status → `ciLocalNeeded` with `LocalNeeded == [zellij]` and no Offender
  3. a `cancel` not in `Runnerless` (it ran) + substitute + pass status → red
  4. a `fail` + substitute + pass status → red
  5. runnerless `cancel` with no declared substitute → red
  6. `pending` CI + a runnerless one → pending (local not run while CI pending)
  7. a required check with no CI checks at all → `ciLocalNeeded`
  8. required + `sdlc-local/x` fail → red
  9. required + pass → green
- [ ] Step 2: Run `go test ./cmd/sdlc -run 'TestDecideCIGate'` → FAIL.
- [ ] Step 3: Implement it. Add `ciLocalNeeded` to the rank map below `ciPending`. In `refusal`, render `ciLocalNeeded` as "PR #N needs recorded local checks: <names>; run `sdlc local-check run --name <n>` (merge runs them itself when it can)".
- [ ] Step 4: PASS. Commit `#345: CI gate accepts a recorded local run for a runnerless job`.

### Task 4: gh seam + stateful fake

**Files:** Modify `cmd/sdlc/ghclient.go` (interface), `cmd/sdlc/cigate.go` (realGH impls), `cmd/sdlc/fetch_test.go` (stubGH defaults). Create `cmd/sdlc/cigate_fake_test.go`.

- [ ] Step 1: Write `TestCIFakeGHRollup`. `PostStatus` on the head makes `PRChecks` return `sdlc-local/x` pass. A second post of `failure` replaces it. A post on an unknown SHA → error.
- [ ] Step 2: Add the interface methods:
  - `JobRun(ctx, repo string, id int64) (jobRun, error)`
  - `PostStatus(ctx, repo, sha, context, state, description string) error`
  - `PRComment(ctx, repo string, number int, body string) error`

  Implement the realGH methods via `runLandingGH`:
  - `api repos/R/actions/jobs/ID`
  - `api -X POST repos/R/statuses/SHA -f state= -f context= -f description=`
  - `pr comment N --repo R --body-file -` (stdin, not argv, for the body)

  Validate the repo, SHA (`landingOIDValid`) and context (`^sdlc-local/[a-z0-9-]+$`) before calling. Add stubGH defaults returning `errors.New("stubGH: not implemented")`. Implement `ciFakeGH`.
- [ ] Step 3: `go build ./cmd/sdlc && go test ./cmd/sdlc -run 'TestCIFakeGH|TestRunCIGate|TestRunMerge_|TestLanding'` → PASS (the existing fakes are unaffected).
- [ ] Step 4: Add `TestLiveJobRunConformance` (skipped unless `SDLC_LIVE_GH=1`). Run it once live: `SDLC_LIVE_GH=1 go test ./cmd/sdlc -run TestLiveJobRunConformance -v` → PASS.
- [ ] Step 5: Commit `#345: gh seam for jobs, statuses and comments, with a stateful fake`.

### Task 5: Local run shell

**Files:** Create `cmd/sdlc/localcheck_run.go`, `cmd/sdlc/localcheck_run_test.go`

- [ ] Step 1: Write `TestRunLocalCheck` against a real temp git repo plus a pushed bare origin (existing test helpers in `merge_e2e_test.go` set these up; reuse them). Use a fake `test-lock.py`: a script printing `acquired`, wired via a temp `Makefile.workflow` symlink. Use `ciFakeGH`. Cases:
  - argv `sh -c 'echo ok'` → status success on the SHA, log exists with "ok", worktree removed
  - `false` → status failure plus one comment whose body is ≤ 8 KiB
  - the helper prints `timeout` → error, no status, the argv never ran (a marker file is absent)
  - `Makefile.workflow` missing → error "refusing to run an unbounded suite"
  - a timeout of 1s on `sleep 5` (test-only override) → failure "timed out"
  - log pruning keeps the newest 20 SHA dirs
- [ ] Step 2: Run → FAIL.
- [ ] Step 3: Implement `runLocalCheck`, `resolveTestLockHelper(root)`, `pruneLocalCheckLogs(dir, keep)` and the stale-worktree sweep.
- [ ] Step 4: PASS. Commit `#345: run a declared local check on the exact head and record it`.

### Task 6: Wire into runCIGate (both merge paths)

**Files:** Modify `cmd/sdlc/cigate.go` (`runCIGate` → takes `ciGateDeps{GH, Git, Local func(...)…, Root string}`), `cmd/sdlc/merge.go:568`, `cmd/sdlc/landing.go:420`, `cmd/sdlc/cigate_test.go`, `cmd/sdlc/merge_e2e_test.go`

- [ ] Step 1: Tests:
  - `TestRunCIGate_RunnerlessRecorded`: `ciFakeGH` with a runnerless job plus a pass status → nil, and the output names the substitution.
  - `TestRunCIGate_RunnerlessRunsInline`: a fake local runner posts a pass → nil; the runner is called once with the pinned SHA; the printed command and log path precede the run.
  - `TestRunCIGate_InlineFails`: the runner posts a failure → refusal.
  - `TestRunCIGate_HeadMovesDuringRun`: the fake moves the head inside the runner → `another commit` refusal.
  - `TestRunCIGate_PostNotVisible`: the runner returns ok but posts nothing → refusal.
  - `TestRunCIGate_RanAndCancelled`: a job with steps 4 → red, and the runner is never called.
  - `TestRunCIGate_BadDeclaration`: invalid JSON on base → refusal naming the file.
  - `TestRunCIGate_DeclarationFromBaseOnly`: a substitute on head only → red.
  - Extend `merge_e2e_test.go` with `TestRunMerge_RunnerlessJobMergesAfterLocalRun` and `TestRunMerge_FailedJobStillBlocks` (the Done-when pair), plus one landing-path test `TestLandingRunnerlessJobMergesAfterLocalRun`.
- [ ] Step 2: Run → FAIL.
- [ ] Step 3: Implement. Classify, and when `ciLocalNeeded` print each `local check <name>: <argv> → log <path>` and run them sequentially. Then re-read head and checks (the existing before/after pin) and decide again; a second `ciLocalNeeded` refuses rather than looping. `--no-ci` still waives everything, unchanged.
- [ ] Step 4: `go test ./cmd/sdlc -run 'TestRunCIGate|TestDecideCIGate|TestRunMerge_|TestLanding'` → PASS.
- [ ] Step 5: Commit `#345: merge runs or accepts recorded local checks`.

### Task 7: Verb

**Files:** Create `cmd/sdlc/localcheck_cmd.go`, `cmd/sdlc/localcheck_cmd_test.go`; modify `cmd/sdlc/main.go` (register).

- [ ] Step 1: Tests:
  - `--name` unknown → error listing the declared names
  - `--rev` not on origin → refuses
  - `--all` runs each check in declaration order and exits non-zero if any fails
  - the help text names `.sdlc/local-checks.json`, the test lock, and that the failure comment posts a log tail
- [ ] Step 2: FAIL → implement → PASS.
- [ ] Step 3: Commit `#345: sdlc local-check run`.

### Task 8: Docs, atlas, help

**Files:** `atlas/workflow/ci-merge-check.md` (new section "Recorded local checks"), `atlas/index.md` (if a new file), the `sdlc merge` help (`merge.go` Long text) and `AGENTS.md`-shaped docs only if base-layer text mentions CI-only evidence (grep `no-ci` in `AGENTS.base.md`/`atlas/`).

- [ ] Step 1: Document the declaration schema with a pair-shaped example, the trust model (base-tree read, `.sdlc/*` guarded), the runnerless rule, `required`, the lock, and the log location and bound.
- [ ] Step 2: Commit `#345: atlas and help for recorded local checks`.

### Task 9: Verification and close

- [ ] Run targeted tests (above), the live conformance test, and `go vet ./cmd/sdlc`.
- [ ] Run the full `make test` in the background before close.
- [ ] Evidence for the real-GitHub path: the declaration is read from `origin/main`, so the full real path (status post → merge acceptance) is first exercised by the pair issue. Here the evidence is the fake-backed e2e plus the live jobs-API conformance. In addition, run `realGH.PostStatus` once live against this branch's own PR head with context `sdlc-local/selftest` and confirm `gh pr checks` lists it as a pass (the rollup assumption `ciFakeGH` models). Record all of this in `--verified`.
- [ ] File the pair issue unassigned: `sdlc issue new` in pair from `:0` semantics (unclaimed), titled "declare couch-zellij-conformance's local substitute in .sdlc/local-checks.json". It covers the docs lines (AGENTS.local/README: which target substitutes, and run it before close when touching Couch session/Zellij code), deps `ariadne#345`, and notes that it needs the operator's `local-gates-verified` label.
- [ ] `sdlc close --issue 345 --verified '…'`, then `sdlc merge --yes`, then report to the hub.
