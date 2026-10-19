# CI merge-check — the pluggable publish gate

A generic, server-side gate every ariadne derivative inherits: on a PR to `main`,
run the repo's own checks over the PR's changes and report (and, if made required,
block). Each derivative plugs in its own logic; a repo with no checks is a no-op.
Introduced in #52; pairs with the in-place-branch workflow (#51) and the LLM
judges (`pre-merge-checks.md`).

## Three layers (and why)

Two CI-isolation facts force the split: GitHub Actions does **not** discover
symlinked *workflow* files, and an isolated CI checkout (the consumer repo only)
does **not** contain the base-layer sibling that a *script* symlink points at:

| Layer | File | Delivery | Why |
|---|---|---|---|
| Shim | `.github/workflows/merge-check.yml` | `seed` | Real file per repo (Actions ignores symlinked workflows); thin + stable so it rarely changes |
| Runner | `scripts/run-merge-checks.sh` | `symlink` → `../<upstream>` | Generic logic, single-source via the symlink; but the target is the base-layer *sibling*, absent in an isolated CI checkout — so the shim clones peers first (below) |
| Checks | `scripts/merge-checks.d/*` | `scaffold` | Repo-owned; each derivative drops its own (or none) |

The shim first runs `BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh` to clone the upstream
peer chain as siblings (so the runner symlink resolves; `CLONE_ONLY` skips the
heavy `make bootstrap` handoff; guarded so the root repo is a no-op). Then it
computes the PR range and calls the runner; the runner discovers and runs every
executable in `merge-checks.d/`, aggregating exit codes. **Empty dir = pass.**
(The clone step was added after the first live PR exercise — you-decide#4 / #53
Phase E — showed the symlinked runner `exit 127`ing in CI.)

## The check contract

```
scripts/merge-checks.d/10-foo.sh  <BASE_SHA>  <HEAD_SHA>     # exit 0 pass, non-0 fail

**Checks in ariadne today:** `30-weave-drift.sh` (generated skills match a fresh
`weave compile`) and `40-duplicate-issue-id.sh` (#213 — refuse a PR reusing an
issue id). The latter is the enforcement half of #213: the gate inside `sdlc
merge` is operator feedback, bypassable by a GitHub-UI merge, a bare `gh pr
merge`, `--no-validate`, or an actor who has not pulled the fix.

Its logic lives in `sdlc issue lint-ids`, not in the script (ARCH-DRY) — filename
parsing, the three id-bearing directories, and the introduced-vs-pre-existing
split are decided once in Go with tests. The script builds sdlc from the checkout
under test rather than trusting a PATH binary, and evaluates **every skip
condition before any side effect** (no `./cmd/sdlc`, no `go`, no writable temp
dir), so a derivative that cannot run the check exits cleanly instead of dying in
setup.
```

Run in filename order; `README*`/`*.md` ignored; findings to stderr. The shim passes
`merge-base(base, head)` as `<BASE_SHA>` so a two-dot `base..head` diff is exactly the
PR's changes.

## One runner, two call sites

`run-merge-checks.sh` is invoked by **CI** (the shim) *and* by a repo's local
**`pre-push` hook** — so local fast-feedback and the server gate run the identical
check-set and can't drift.

## Strictness is opt-in

Without branch protection the check is **advisory** (runs + reports; doesn't block the
merge button; direct pushes to `main` stay open). Making `merge-check` a *required*
status check — and creating the remote repo — is the `make remote-init` operator verb
(#52 M2; needs `gh` auth). Mechanism works advisory-only without it.

## Local gates and server CI: the split (#309)

Local sdlc gates do the heavy checks (review, plan quality, tests), because
ariadne controls the local process. Server CI does cheap deterministic checks,
and `sdlc merge` will not land over it:

- **Merge reads CI** (`cmd/sdlc/cigate.go`). Both landing paths read the checks
  GitHub reports on the PR head (`gh pr checks --json`) and refuse unless every
  one passed or was skipped, naming each failing or pending check. An unknown
  state or an unreadable list refuses (fail closed). No checks while the head
  carries `.github/workflows/` is pending; no checks and no workflows passes.
  The merge is pinned to the read head with `--match-head-commit`. `--no-ci`
  waives only this gate. gh exits 1 (failing) or 8 (pending) while still
  printing the JSON, so the adapter reads stdout before the exit status.
- **Durable CI wait** (`cmd/sdlc/cigate_wait.go`). The durable path polls only
  pending or unregistered checks every 30 seconds, for up to 20 minutes of CI
  waiting (local suite runtime excluded). Typed refusals keep red, missing,
  unknown, head-lag and read errors terminal. Immutable main/head facts, including
  the make probe, are loaded once per wait. The legacy path still refuses pending.
- **The local-gates guard** (`.github/workflows/local-gates-guard.yml`, seeded).
  It runs on `pull_request_target`, so main's copy of the guard decides and a
  PR cannot weaken the guard within itself; it never checks out or runs PR
  code, only the changed-path list of `refs/pull/<N>/head`, and takes PR values
  through `env:`. A PR touching a protected path fails unless it carries the
  `local-gates-verified` label. It reads the changed paths NUL-delimited and
  unquoted (`-z`, `core.quotePath=false`), so a non-ASCII or quoted path is
  still matched; it reads the labels live from the API at decision time, not
  from the event payload, and fails closed when it cannot; one run per PR at a
  time (`concurrency`, newer cancels older); a run whose event head is no
  longer the PR head decides nothing. It also posts a `local-gates-guard/verdict`
  commit status on the PR head, so `gh pr checks` (and merge's CI gate) see the
  latest verdict for that head whichever way GitHub attaches the job's check.
  Tested by executing the real run block: `scripts/test/local-gates-guard.test.sh`,
  run in CI by ariadne's `merge-checks.d/20-ci-workflow-tests.sh`, which runs
  every `scripts/test/*.test.sh` except two named exclusions. Among the scripts it
  runs, `scripts/test/construct-scripts.test.sh` (#362) runs every
  `construct/scripts/test/*.test.sh` and names its own exclusions (none today);
  `make test` runs the same runner through the Go test beside it
  (`scripts/test/scripts_test.go`), so CI and the local suite share one list.
  The Go entry point repeats that list under plain, no-color/CI, and forced-color
  terminal environments (#363). Color assertions in `vm-log.test.sh` clear inherited
  color flags before setting each case's policy.
- **Absence is not green.** The guard posts `pending` first and turns any exit
  before a decision into `failure`; merge's CI gate requires the
  `local-gates-guard/verdict` status whenever the target main carries the
  guard, so a run that never reported refuses.
- **The label is the operator's marker.** Only the operator applies
  `local-gates-verified`; no sdlc verb does and agents are told never to. A
  local-gates change is its own issue with operator-held verification. Honest
  limit: agents share the operator's gh identity, so the label is an explicit,
  audited acknowledgment, not a cryptographic one.
- **How the protected list is derived** (re-sweep from these, don't hand-pick):
  `construct/base.manifest` rows whose file runs or configures a merge-blocking
  check, or the agent's sandbox and permissions; repo-root config that sdlc and
  weave read (`.sdlc/`, `workshop/review.json`, `construct/deps`, the manifest
  itself); CI-only hooks (`scripts/ci-setup.sh`, `.claude/hooks/`). Gate code
  (`cmd/sdlc`, vocabulary) is out: changing it is ordinary ariadne work and the
  boundary review owns it. The list lives once, in the guard workflow.
- **merge-check runs on push to main too**, over `before..after` (a new ref
  checks nothing), so main has a CI status of its own.
- **Not covered:** `sdlc push` lands on main with no PR, so no PR CI gates it;
  the push run reports after the fact. Branch protection (required checks) is
  not configured: merge's refusal is the gate for agents, the guard the
  server's detector.

## Recorded local checks (#345)

The operator's stance: for a small team, local runs are the primary verification
(faster and more controlled on the operator's machine), and CI mainly guards that
what agents are not supposed to change has not changed. So a local run is
first-class evidence, recorded on the commit it ran on:

- **Declaration** (`.sdlc/local-checks.json`, parsed strictly in
  `cmd/sdlc/localchecks.go`). Each check has a `name`, an argv `command`,
  optional `substitutes` (CI check labels, `workflow / job`), optional
  `required`, and `timeout_minutes`. The gate reads it from **main's** tree,
  never the PR head, and `.sdlc/` is under the local-gates guard, so an agent
  cannot declare its way past a CI job.
- **Recording** (`sdlc local-check run`, `cmd/sdlc/localcheck_run.go`). It runs
  in a fresh detached worktree at the exact commit, under one of the
  machine-wide full-suite slots (the same `scripts/test-lock.py` that
  `make test` takes, found beside `Makefile.workflow`'s real path; it refuses
  rather than run unbounded). The slot covers the suite's execution only: a
  per-run holder child is the watched pid and is released when the suite ends,
  so merge's GitHub polling and landing hold no slot (#383). The log is kept under
  `<git common dir>/sdlc/local-checks/<sha>/` (newest 20 commits), and the run
  posts the commit status `sdlc-local/<name>`. A failure also posts a bounded
  log tail as a PR comment.
- **Merge** (`decideCIGate`, `runCIGate`). A `cancel` check counts as
  *runnerless* only when the Actions jobs API reports no runner and zero steps
  (structural, measured on pair job 115490631095; a live conformance test reads
  it). A runnerless job with a declared substitute passes on that substitute's
  recorded pass. A `required` check needs its recorded pass whatever CI says.
  A job that ran and failed or was cancelled stays red. When a needed run is
  missing, merge first posts a saved pass of that check on the exact head (the
  newest attempt there, its log hash verified, #431); without one it runs the
  check itself through the same path, then re-reads the head
  and checks (a push during the run, or a pass that never reached GitHub,
  refuses). Red, unknown and pending CI outrank a needed local run, so nothing
  runs locally while CI is red or still running.
- **Known flakes (#394).** When the `suite` check fails, `knownFlakeStep`
  (`cmd/sdlc/localcheck_flakes.go`) classifies the attempt
  (`internal/flakes.Classify`) against the registry: open issues whose card
  lists `flake_tests`, read fresh from the tracker. If every failure is a
  named, registered test, those tests alone are rerun once on the same head
  (`go test -v -run`, weight 1, through the shared `runUnderSlot`, so it has its
  own lease and #383 slot release, bounded by the check's timeout). It lands
  only when each named test shows `--- PASS`. The registry and the policy are
  rechecked just before posting (`flakes.Revalidate`), and again when a saved
  known-flake pass is reused or republished. The status reads "pass · known
  flake #N rerun". The rerun is saved as a v2 attempt of check `suite-rerun`
  (reserved; the decoder ties it to `rerun_of`), with its log kept per ID, so
  no reader of suite attempts can take it for a suite run. A saved
  known-flake pass is reused from its parent attempt's outcome. These block:
    - an unregistered failure;
    - an unattributable one (a build error, a timeout, a crashed shard, an
      unparseable log);
    - a known flake that fails the rerun again or proves nothing in it;
    - more than five at once;
    - an unreadable registry.
  An interrupted rerun posts nothing. **No blind rerun lands:**
  `guardSuitePass` (`postLocalCheckRecord`, so every run, reuse and republish)
  refuses a suite success on a commit while a failed suite attempt there is
  unexplained (`flakes.Unexplained`: not a known-flake landing, and not every
  failure a test registered now; an unattributable or rejected failure never).
  It fails closed on a test-stats store it cannot read whole (a record pruned
  mid-read is skipped, not unreadable). The reader's caps (10,000 records,
  64 MiB) now gate landings; the writer is bounded by the 30-day prune, and
  ariadne's store measured 155 records / 620K on 2026-10-18, about 60x under. Store readers
  use `isSuiteRun` (counts, reuse) or `isFlakeEvidence` (a suite run or a
  passing `suite-rerun`). `known_flakes` (`rerun` \| `accept` \|
  `off`) on the declared suite is the operator's policy, under the same
  local-gates guard. Test-stats counts a rerun as flake evidence, never as a
  suite run.
- **Attempt compatibility (#460).** The test-stats reader accepts additive
  unknown JSON members at the attempt root and inside duration units and known
  flake facts. It still rejects duplicate keys, aliases of known field names,
  malformed known values, and unsupported record versions. Attempt records are
  read-only evidence; each run publishes a new record instead of rewriting one.
- **Trust limit:** anyone with the repository's token can post an
  `sdlc-local/*` status. That is accepted: the local machine is trusted, and
  what CI guards is the declaration and the gate code.
- **Fake:** `ciFakeGH` (`cmd/sdlc/cigate_fake_test.go`) models heads, checks,
  the status rollup, jobs and comments, behind the same `ghCaller` seam.
- **Live conformance:** run
  `SDLC_LIVE_GH=1 SDLC_LIVE_COMMIT_STATUSES_GH="$(command -v gh)" go test ./cmd/sdlc -run '^TestLiveCommitStatusesConformance$' -count=1`
  when changing the status adapter, parser or fake, and before landing behavior
  that depends on them. It only reads GitHub (requires authenticated access to
  `xianxu/ariadne`): PR #229's head has a known guard history with pending,
  failure and success states. The check verifies that `realGH.CommitStatuses`
  selects the newest state. The explicit installed CLI path avoids the test
  harness's fake `gh`; changed fixture history requires independent inspection
  before updating expectations. Ordinary tests skip this network check.
- **The suite on every landing (#332).** The check named `suite` is required
  for every landing: either the declared one, or the repository's `make test`.
  The default's existence and its weight (`WF_TEST_WEIGHT`) come from probing
  **main's** Makefile in a detached worktree, under a goal that is not `test`,
  so the probe takes no lock. A branch therefore cannot drop or lighten its own
  suite. A declared `suite` overrides execution settings but is always required,
  even when `required` is omitted or false. Agents run it before close with
  `sdlc local-check run --name suite`.
  - A record on a first-parent ancestor counts when the difference since is
    docs-only (`publishGateHasCodeSurface`, the publish gate's line). Close's
    own commits don't force a rerun.
  - A declared `weight` reaches `test-lock.py --slots`, so `make test` under it
    holds the repository's weight. Before #332 it held one slot and went
    reentrant.
  - Before local-check worktree maintenance or queueing, SDLC probes the resolved
    helper's side-effect-free `acquire --help` for the required `--watch` and
    `--slots` declarations. Missing or unverifiable capability refuses without a
    check run or posted status. Help is bounded to five seconds and 64 KiB, with
    process-group cancellation and a one-second pipe drain.
  - A stale helper refusal names its actual supplying checkout and HEAD, then
    gives `weave refresh --substrate` in the consumer checkout. This repairs
    private substrates from a clean issue branch without rebasing that branch.
    If the helper is tracked in the consumer itself, the refusal instead says
    to merge current main into the issue branch; substrate refresh cannot adopt
    newer tracked source. Neither repair runs automatically.

The durable landing CI wait has a 20-minute budget covering inflight GitHub
and Git reads. Every read receives the remaining deadline, and late green
results refuse. Local suite runtime, including machine-lock queueing, is
excluded. Cleanup receives fresh bounded IO after cancellation.
The Makefile facts probe also inherits the remaining CI budget; its normal
two-minute ceiling still applies outside a wait.
If the integrated head's suite fails, landing runs that same suite once on
main as an unrecorded diagnostic baseline. The refusal separates named tests
introduced by the branch from those also failing on main. An unparseable
baseline failure is reported as such rather than assigning its failures to
the branch; neither failed tree can land. Both runs retain their local logs,
but only the tested branch head posts a commit status or PR comment.
Recorded checks use a detached worktree of the pushed commit. The suite and
its assertions therefore must work without ignored, woven projections from
an agent's long-lived slot; tests that need such files supply their own
fixtures or condition their expectations on the projection being present.

## Not the same as the publish gate or the close review

Three distinct tiers, none an LLM at merge-time (#160):

- **This CI merge-check** — **deterministic, server-side, repo-specific** (a substrate
  gate, a lint, a test), gating the PR.
- **The publish gate** (`sdlc merge`/`push`, `cmd/sdlc/publishgate.go`) — **deterministic,
  local**: the reviewed-HEAD-unchanged invariant + the `codecomplete → done` flip. No LLM.
- **The close boundary review** (`sdlc close`) — the **LLM, local** acceptance review;
  all LLM review is now close-time, NOT merge-time (the old merge-time `plan`/`specs`
  judges were removed, #142 folded into #160). See `pre-merge-checks.md`.

## Example consumer

you-decide drops `merge-checks.d/10-review-gate.sh` (wraps its `review: passed`
substrate gate) and points its `pre-push` hook at the runner. See you-decide#4 M3.

### Two kinds of not-running, decided apart (#213 BR-28)

`40-duplicate-issue-id.sh` distinguishes them, because collapsing them is how a
check suppresses a real refusal:

| | condition | exit |
| --- | --- | --- |
| **nothing to check** | no `workshop/issues/`, or no `origin` remote | 0 — the check is inapplicable and clean is honest |
| **could not check** | no `go`, no `cmd/sdlc` to build, no writable temp dir, build failed, `origin/main` unreadable | 2 — there IS something to check, so clean would be a false pass |

In an issue tracker repository (#252) new IDs are allocated on cards, so a
collision needs a hand-made details file; the check still guards details a
branch adds. It reads the cards with `lint-ids --remote origin`, the same remote
as the trunk, not through the resting branch's publication target: a CI checkout
is detached with no local `main` configured, so that lookup made every consumer
PR exit 2 (#266). A marked repository whose `origin` has no tracker still exits 2.

Both sets are evaluated before any side effect. The `sdlc issue lint-ids` it
shells to uses the same three codes (0 clean, 1 introduced, 2 could-not-run), so
the distinction survives to CI rather than being flattened at the boundary.

## Standalone consumers (#225)

The upstream-seeded workflow clones peers first, then selects the declared Go
version from the consumer's go.mod or the bootstrapped sibling ariadne/go.mod.
An executable repo-owned `scripts/ci-setup.sh` runs before checks; failure stops
the job. Consumer tool provisioning belongs in this hook, which weave does not
overwrite. The check runner resolves from `scripts/run-merge-checks.sh` or
`../ariadne/scripts/run-merge-checks.sh`; neither present is a visible failure.
It always runs the consuming repository's checks from that repository's cwd.
`bash scripts/test/portable-ci.test.sh` executes the workflow's actual run
blocks against a scratch Git repository and real runner, including hook failure
and local/fallback resolution. No credentials or hosted runner are required.

The SDLC boundary reviewer distinguishes executable fixes from prose-only repairs
when disposing prior findings. Executable changes retain fails-without-fix
regressions; README/plan wording uses inspected pinned diffs and the referenced
source. Markdown prompts themselves remain executable contracts. The rendered
review prompt and its golden pin this distinction (`cmd/sdlc/internal/judge`).
