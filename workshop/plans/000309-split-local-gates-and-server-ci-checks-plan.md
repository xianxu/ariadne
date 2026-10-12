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
  The PR head is already pinned (local = remote = PR head, then
  `--match-head-commit`), so the checks read belong to the head that merges.
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
  protected by `.github/workflows/**`):
  `.github/workflows/**`, `scripts/run-merge-checks.sh`,
  `scripts/merge-checks.d/**`, `scripts/ci-setup.sh`, `scripts/pre-merge-checks.sh`,
  `bootstrap.sh`, `.claude/settings*.json`, `.claude/hooks/**`,
  `.gemini/settings.json`, `.sdlc/**`, `workshop/review.json`,
  `construct/base.manifest`, `construct/deps`.
  Gate CODE (`cmd/sdlc/**`) is deliberately absent: changing it is ariadne's
  normal work, and the boundary review owns it. The list covers configuration
  that switches or loosens gates.
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
- [ ] Tests first (`cigate_test.go`): table over buckets — all pass; one fail
  named; one pending named; cancel red; unknown bucket refuses; zero checks
  with workflows pending; zero checks without workflows pass; query error
  refuses. Each refusal row asserts its own reason (lessons: rejection table).
  Parser test over real `gh` JSON, including the no-checks message.
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
- [ ] `scripts/test/local-gates-guard.test.sh`: executes the workflow's real
  run block against a scratch repo (portable-ci.test.sh pattern): untouched →
  pass; protected path touched → fail naming the path; touched + label →
  pass; label on an untouched PR → pass; each protected pattern matches a
  sample path (enumerate from the list, not from examples); a near-miss path
  (`scripts/merge-checks.d.md`, `cmd/sdlc/x.go`) does not match.
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
  Confirm the guard fails, `sdlc merge` refuses naming it; then close the PR
  and delete its branch. Report seeded files to the hub for the TL's refresh.

## Verification

- Targeted: `go test ./cmd/sdlc/ -run 'CIGate|Merge|Landing'`,
  `bash scripts/test/local-gates-guard.test.sh`, `bash scripts/test/portable-ci.test.sh`.
- Full `make test` before each close (one suite at a time; the lock waits).
- Live: the after-landing check above.
