# #309 — Split local gates and server CI checks: plan

## Goal

Local sdlc gates keep doing the heavy work. Server CI does three cheap things:
the deterministic checks it already runs, a guard that a PR has not quietly
changed the local gates, and a status that `sdlc merge` refuses to land over
when it is red or pending.

## Current state (measured 10-12)

- `.github/workflows/merge-check.yml` (seeded) runs on `pull_request` only. It
  is green on recent PRs in ariadne, pair, parley.nvim and tools. Only pair runs
  anything on push to main (`couch-zellij-conformance`). No red CI to fix today.
- `sdlc merge` never looks at checks. Both paths merge over red or pending CI:
  the durable landing (`landing.go:runDurableMerge`, numbered slots and :0) and
  the legacy path (`merge.go:runMerge`, ordinary worktrees and clones).
- Nothing on the server notices a PR that edits merge-checks, workflows, hooks,
  sandbox settings or review config. The PR's own workflow runs on
  `pull_request`, so a PR can weaken the check that would have caught it.

## Decisions

- **D1. Merge reads checks on the PR head.** `gh pr checks <N> --json
  name,bucket,workflow,link`. `pass` and `skipping` are green; `fail` and
  `cancel` are red; `pending` is pending; any other bucket refuses
  (fail closed). The refusal names every non-green check with its link.
  Zero checks while the head tree carries `.github/workflows/` counts as
  pending: GitHub has not registered the suites yet, and clean would be a
  false pass. A repository with no workflows at all passes with a note.
  The checks read must belong to the head that merges. The durable path
  already pins it (local = remote = PR head, then `--match-head-commit`). The
  legacy path does not: `realGH.PRMerge` (`ghclient.go`) merges whatever the
  PR head is. M1 changes the seam to `PRMerge(repo, branch, expectedHead)`,
  passing `--match-head-commit`, with the head read from `origin/<branch>`
  after step 4 (which already proves local = remote).
  **`gh pr checks` exit codes** (gh 2.93): measured 10-12, a PR with no
  checks prints `no checks reported on the '<branch>' branch` on stderr and
  exits 1; a green PR prints a JSON array and exits 0. gh documents exit 8 for
  pending and 1 for failing checks. So the adapter parses stdout first: a JSON
  array is the answer whatever the exit code; the no-checks message is an
  empty list; anything else with a non-zero exit is a query failure. Fixtures:
  the captured pass and no-checks outputs now; the live check captures a real
  failing and pending output, which then replace the constructed fixtures.
  A repository whose workflows never trigger on PRs stays pending forever
  under this rule; the help text says so and names `--no-ci` for that case.
- **D2. One gate, both paths.** A pure `decideCIGate(checks, hasWorkflows)`
  returns pass, or a refusal carrying its reason class (red, pending,
  unknown, query failed). Both merge paths call one `runCIGate`. The `gh`
  call sits behind a `ciChecksGH` interface so tests use a fake.
- **D3. `--no-ci` bypass**, announced loudly like `--no-validate`. It waives
  only this gate. Help and the merge help text say it is for a broken CI
  system, with the reason in the log.
- **D4. Guard runs on `pull_request_target`.** New seeded workflow
  `.github/workflows/local-gates-guard.yml`. `pull_request_target` runs the
  BASE branch's copy of the workflow, so a PR cannot weaken the guard within
  itself. It never checks out or executes PR code: it fetches the PR head
  ref and computes `git diff --name-only merge-base..head`. Permissions:
  `contents: read`. Triggers: opened, synchronize, reopened, labeled,
  unlabeled. Only same-repository PRs exist here; the guard does not need
  secrets, so a fork PR gets no extra exposure.
- **D5. The protected list lives in the guard workflow** (one source, itself
  protected by `.github/workflows/**`). It is derived, not hand-picked, from
  three sources, swept 10-12:
  1. `construct/base.manifest` rows whose delivered file runs or configures a
     merge-blocking check, or the agent's sandbox and permissions:
     `bootstrap.sh`, `.codex/config.toml`, `.claude/settings*.json` (the
     symlinked source and the merged target), `scripts/pre-merge-checks.sh`,
     `scripts/parallel-checks.sh`, `.github/workflows/**`,
     `scripts/run-merge-checks.sh`, `scripts/merge-checks.d/**`,
     `.openshell/policy.yaml`;
  2. repo-root config that sdlc and weave read (grep of `cmd/`, `pkg/` for
     repo-relative config paths): `.sdlc/**` (fleet, freshness),
     `workshop/review.json`, `construct/deps`; plus `construct/base.manifest`
     itself, which chooses what is woven;
  3. CI-only hooks: `scripts/ci-setup.sh`, `.gemini/settings.json`,
     `.claude/hooks/**`.
  Some entries exist only in peer repos or not yet anywhere
  (`scripts/ci-setup.sh`, `workshop/review.json`, `construct/deps` in
  ariadne, `.claude/hooks/`); they are listed so the seeded guard covers every
  consumer. Deliberately absent: gate CODE (`cmd/sdlc/**`, vocabulary `.cue`)
  — changing it is ariadne's normal work and the boundary review owns it — and
  build tooling that blocks nothing (`Makefile.workflow`). The atlas records
  this rule so the next sweep starts from it.
- **D6. The marker is the PR label `local-gates-verified`**, applied only by
  the operator. No sdlc verb applies it; AGENTS tells agents never to. Honest
  limit (TL, 10-12): agents share the operator's gh identity, so the label is
  an explicit, audited acknowledgment, not a cryptographic one. The guard's
  failure message lists the touched protected paths and says the operator
  applies the label.
- **D7. merge-check also runs on push to main**, so main has a CI status. The
  push run uses `before..after`; a zero `before` (new ref) checks nothing.
  Each existing check must give the right answer on a push range (duplicate-id
  compares against the trunk tip, which on push equals head: verify it reports
  clean, not a self-collision).
- **D8. Rollout.** The guard and the shim reach pair, parley.nvim and tools
  through the TL's weave refresh per slot (seed rows). This issue creates the
  label in all four repos with `gh label create`, and reports the seeded files.
- **D9. Out of scope.** `sdlc push` lands on main without a PR, so no PR CI
  gates it; the push-to-main run reports after the fact. Branch protection /
  required checks (`make remote-init`, #052 M2) stays out: sdlc merge's refusal
  is the gate for agents, and the guard is the server's detector.

ARCH notes: ARCH-DRY (one gate decision for both merge paths; one protected
list). Fail closed on unknown buckets and query failure (lessons, "guards and
proof"). ARCH-FUNERAL: creates nothing durable beyond a GitHub label per repo
and a workflow file; check runs age out under GitHub's retention.

## Milestones

### M1 — `sdlc merge` refuses red or pending CI

- [ ] `cmd/sdlc/cigate.go`: `prCheck` type, `ciChecksGH` interface,
  `realGH.PRChecks` (parses `gh pr checks --json`; "no checks reported" is an
  empty list, other failures are errors), pure `decideCIGate`, `runCIGate`.
- [ ] Tests first (`cigate_test.go`). Classifier: a table over every bucket
  plus an unknown one, zero checks with and without workflows, and query
  failure, each refusal row asserting its own reason. Adapter: fixtures of
  real `gh` output (pass, no-checks; fail and pending after the live check)
  across exit codes.
- [ ] `PRMerge` takes the expected head and passes `--match-head-commit`;
  argv test; legacy merge reads the head from `origin/<branch>`.
- [ ] Wire into `runDurableMerge` after the publish gate, before
  `revalidateLanding`; and into `runMerge` before `PRMerge` (open PR only;
  resume of an already-merged PR skips it). Per-site tests: each path refuses
  a red check with its fake, and each fails when its call alone is removed.
- [ ] `--no-ci` flag on merge, loud warning, both paths; help text in
  `helptext/merge.md`; dry-run reports what the gate would say.
- [ ] Existing merge tests keep passing (their fakes report green, or the
  test seam defaults to a green fake).

### M2 — server guard, push-to-main, docs

- [ ] `.github/workflows/local-gates-guard.yml` per D4–D6, run block in Bash
  with the protected list as a pattern list.
- [ ] The run block takes PR values only through `env:` (number, base SHA,
  labels as JSON) and fetches `refs/pull/<N>/head` by number; it never
  interpolates `${{ github.event.pull_request.head.ref }}` or titles into the
  script (ARCH-SECURE).
- [ ] `scripts/test/local-gates-guard.test.sh` executes the workflow's real run
  block against a scratch repo (portable-ci.test.sh pattern). Matcher: cases
  generated from the pattern list in the workflow, plus near-miss mutations of
  each; decision: touched/untouched × labeled/unlabeled.
- [ ] `merge-check.yml`: add `push: branches: [main]`; range from
  `before`/`after` on push. Extend portable-ci.test.sh for the push event;
  verify duplicate-id is clean on a push range.
- [ ] `construct/base.manifest`: `seed .github/workflows/local-gates-guard.yml`.
- [ ] Label `local-gates-verified` created in ariadne, pair, parley.nvim, tools.
- [ ] Atlas: `atlas/workflow/ci-merge-check.md` gains the local/server split,
  the merge CI gate, the guard and the label with its honest limit;
  `atlas/index.md` link if a new file. AGENTS.base.md: one line that agents
  never apply `local-gates-verified`, and a local-gates change is its own
  issue with operator-held verification.
- [ ] Close #052 as superseded (if still open).

### After landing (not a boundary)

- [ ] Live check (TL-approved): a throwaway ariadne PR titled
  "DO NOT MERGE: local-gates-guard live check" touching a protected path.
  Confirm the guard fails, `sdlc merge` refuses naming it; capture the real
  failing and pending `gh pr checks` output as fixtures; add the label and
  confirm `gh pr checks` reports the re-run (latest run per name), not the
  earlier failure; check whether the guard's job run shows on main's tip
  (pull_request_target attaches to the base); then remove the label, close
  the PR and delete its branch. Report seeded files to the hub for the TL's refresh.

## Verification

- Targeted: `go test ./cmd/sdlc/ -run 'CIGate|Merge|Landing'`,
  `bash scripts/test/local-gates-guard.test.sh`, `bash scripts/test/portable-ci.test.sh`.
- Full `make test` before each close (one suite at a time; the lock waits).
- Live: the after-landing check above.

## Revisions

- 2026-10-12, plan-quality round 1: pin the head on the legacy path
  (PRMerge `--match-head-commit`); parse `gh pr checks` stdout before its exit
  code, with measured fixtures; D5 derived from three stated sources
  (adds `.codex/config.toml`, `.openshell/policy.yaml`,
  `scripts/parallel-checks.sh`); guard run block takes PR values via `env:`;
  test bullets compressed to strategies; live check confirms re-run
  supersession.
- 2026-10-12, M1 implementation: dropped "dry-run reports what the gate
  would say" — neither merge path's dry-run contacts GitHub for PR state, and
  adding it is outside this issue.
- 2026-10-12, M1 review BR-1: `gh pr checks` reads the checks of GitHub's
  current PR head, which can lag a push. `runCIGate` now reads the PR head
  (`gh pr view --json headRefOid`) before and after the checks and refuses
  unless both equal the pinned head, on both paths. Every CI-gate refusal, and
  only refusals, starts `CI gate: ` (success lines say `CI checks:`), so the
  gate catalog matches the prefix.
- 2026-10-12, M2 review round 1 — sweep of every M2 decision and bullet
  against what shipped:
  - D4 permissions grew: `statuses: write` (the guard posts a
    `local-gates-guard/verdict` commit status on the PR head, so merge's CI
    gate sees the verdict whichever SHA GitHub attaches the job's check to),
    and `issues: read` / `pull-requests: read` for the live label read.
  - D4 trigger narrowed to PRs targeting `main` (`branches: [main]`).
  - D4/D6 labels are read live (`gh api .../issues/N/labels`) at decision
    time, not from the event payload, failing closed (exit 2, failure status)
    when unreadable; a per-PR `concurrency` group cancels an older run.
  - New no-op branch: a run whose event head is no longer the PR head exits 0
    and posts nothing; the run for the newer push decides. Not fail-open: the
    newer head's own run posts its verdict, and merge's gate reads the head.
  - Changed paths are read with `-z` and `core.quotePath=false` (BR-5).
  - D7: merge-check run blocks take event values through `env:` too.
  - Tests are wired: `scripts/merge-checks.d/20-ci-workflow-tests.sh` runs the
    guard and portable-ci tests in CI and pre-push (ariadne-local).
  - README documents the guard, the label and the push run (BR-7).
  - The live check also confirms whether a failed guard job marks main's tip.
