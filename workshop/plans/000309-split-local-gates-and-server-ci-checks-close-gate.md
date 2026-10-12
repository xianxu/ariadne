---
gate: boundary-review
issue: 309
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T12:25:10-07:00"
      dispatched: "2026-10-12T11:59:20-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Legacy merge path never verifies the checks gh reports belong to the pinned head
          detail: merge.go:552-566 pins head from local origin/<branch> but gh pr checks reads GitHub's current PR head; right after a push the PR object can lag, old-head green checks pass the gate, and --match-head-commit only protects if GitHub has not yet synced. Compare PR headRefOid to head (treat mismatch as pending) and add a fake test.
          family: checks-bound-to-merged-head
          round: 1
        - id: BR-2
          severity: Minor
          title: Dropped M1 dry-run bullet is recorded only in the issue Log, not the plan Revisions
          detail: Add a Revisions entry so the plan stops claiming the dry-run behaviour.
          family: plan-revisions-record-scope-changes
          round: 1
        - id: BR-3
          severity: Minor
          title: no-ci RefusalPat misses the head-tree-read and cannot-pin refusals
          detail: 'gatesig.go:163 matches "PR N has|reports" and "could not read", but not "CI gate: read <head>''s tree" or "CI gate: cannot pin the merge".'
          family: gate-catalog-covers-all-refusals
          round: 1
        - id: BR-4
          severity: Minor
          title: exitErr test helper builds the exit code via rune arithmetic, valid only for 0-9
          detail: cigate_test.go:68; use strconv.Itoa.
          family: test-helper-correctness
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-12T12:56:15-07:00"
      dispatched: "2026-10-12T12:43:31-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: cigate.go:216-230 reads the head before and after the checks; reverting the comparison turns TestRunCIGate and TestRunMerge_LaggingPRHeadRefuses red.
          round: 2
        - id: BR-2
          disposition: addressed
          note: The plan's Revisions section now records the dropped dry-run bullet (2026-10-12, M1 implementation).
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: 'The prefix pattern covers the class, but no classifier test pins any no-ci refusal: reverting gatesig.go to the 80160793 pattern leaves processmanual tests green. Add TestClassifyOutputLine cases for every refusal string, plus negatives for the CI checks: success line and the ACK.'
          round: 2
        - id: BR-4
          disposition: addressed
          note: cigate_test.go:72 now uses strconv.Itoa.
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      reviewed: 1f915d6065bf307057fd2c7c4a88d6bdcfe67b13
      blocked: false
    - "n": 3
      timestamp: "2026-10-12T13:13:44-07:00"
      dispatched: "2026-10-12T13:08:33-07:00"
      agent: claude
      findings:
        - id: BR-5
          severity: Important
          title: local-gates-guard misses paths git prints in quoted form (non-ASCII, quotes, backslashes)
          detail: 'git diff --name-only without -z wraps such paths in quotes with escapes (reproduced: scripts/merge-checks.d/50-é.sh and .github/workflows/a"b.yml both pass as untouched). Use git -c core.quotePath=false diff -z and read -d ''''; add a non-ASCII protected-path test case that fails before the fix.'
          family: guard-parses-input-losslessly
          round: 3
        - id: BR-6
          severity: Important
          title: Guard decides on the event-payload label snapshot with no per-PR concurrency group
          detail: A re-run of an old labeled run, or a label removal racing a labeled run, can leave a success status after the label is gone. Read labels live via gh api at decision time (as the head is re-read) and add a per-PR concurrency group (ARCH-ORDER).
          family: guard-decides-on-live-state
          round: 3
        - id: BR-7
          severity: Important
          title: README does not mention the seeded local-gates-guard, the operator-only label, or the push-to-main run
          detail: README.md:186-192 documents the seeded CI workflow for consumers; the new seeded workflow and the local-gates-verified label are consumer- and operator-facing surface.
          family: docs-cover-new-surface
          round: 3
        - id: BR-8
          severity: Important
          title: M2 deviations from D4 (statuses write, verdict status context, head-moved no-op, branches filter) have no Revisions entry
          detail: '2nd finding in this family. Rule: at each milestone-close, compare every Decision and checklist bullet against what shipped; any added permission, output or context, dropped item, or new fail-open/no-op branch gets a Revisions line in the same commit. Sweep all of M2, not just this instance.'
          family: plan-revisions-record-scope-changes
          round: 3
        - id: BR-9
          severity: Minor
          title: local-gates-guard.test.sh and portable-ci.test.sh run in no suite or CI job
          detail: Manual-only, as before; the guard's regression test never runs automatically. The head-moved exit is also untested.
          family: tests-wired-into-suite
          round: 3
        - id: BR-10
          severity: Minor
          title: Confirm in the live check whether failed guard job runs mark main's tip red
          detail: pull_request_target check runs may attach to the base SHA, which would muddy D7's main status.
          family: live-check-covers-attachment
          round: 3
      boundary: M2
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 4
      timestamp: "2026-10-12T13:26:17-07:00"
      dispatched: "2026-10-12T13:20:09-07:00"
      agent: claude
      dispose:
        - id: BR-5
          disposition: addressed
          note: local-gates-guard.yml:84-92 uses core.quotePath=false diff -z with read -d ''; reverting both makes the test fail on scripts/merge-checks.d/50-é.sh (verified in this checkout).
          round: 4
        - id: BR-6
          disposition: addressed
          note: Labels are read live via gh api, failing closed with exit 2; a per-PR concurrency group is in place; swapping in a payload LABELS variable makes the test fail (verified).
          round: 4
        - id: BR-7
          disposition: addressed
          note: README.md:192-201 documents the push-to-main run, the seeded local-gates-guard.yml, the operator-only label and gh label create, matching the workflow and atlas.
          round: 4
        - id: BR-8
          disposition: addressed
          note: Plan Revisions entry "M2 review round 1" sweeps permissions, branches filter, live labels, concurrency, the head-moved no-op, -z, env for D7, wiring, README and live check; small omissions are noted as Minor.
          round: 4
        - id: BR-9
          disposition: addressed
          note: scripts/merge-checks.d/20-ci-workflow-tests.sh (mode 755) runs both tests through run-merge-checks.sh in merge-check CI; the head-moved case is tested (posts nothing, exits 0).
          round: 4
        - id: BR-10
          disposition: addressed
          note: The plan's after-landing live check now explicitly checks whether the guard's job run shows on main's tip; this is a plan item, not runtime behavior.
          round: 4
      findings:
        - id: BR-11
          severity: Important
          title: Guard exits before its decision (fetch/merge-base error, cancelled run, head-moved no-op) leave no verdict status, and absence reads as green
          detail: 'local-gates-guard.yml:76-82 can exit under set -e before verdict() is defined, so no local-gates-guard/verdict lands on the head, and decideCIGate only refuses checks that exist. If pull_request_target job runs attach off the PR head (the BR-10 question), merge sees no guard and passes; the atlas claim "whichever way GitHub attaches" fails here. Concrete case: an operator re-run of an old event cancels the current head''s run via the concurrency group, then no-ops on head-moved, leaving no verdict. Fix: define verdict() early and add an EXIT trap that posts failure on any non-zero exit after input validation, and consider posting pending on head-moved; add a test where the fetch fails and assert state=failure.'
          family: gate-fails-closed-on-absent-evidence
          round: 4
        - id: BR-12
          severity: Minor
          title: Script tests are wired by name, not by class; the Revisions "pre-push" claim is unsupported and slot-dependencies.test.sh runs nowhere
          detail: '2nd finding in this family. The rule: every scripts/test/*.test.sh is reached by an automated runner. 20-ci-workflow-tests.sh should glob scripts/test/*.test.sh, excluding explicitly any test that needs arguments (release-weave, which weave-release.yml runs). No ariadne pre-push hook calls run-merge-checks.sh, so correct the plan''s "CI and pre-push" wording.'
          family: tests-wired-into-suite
          round: 4
      boundary: M2
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 5
      timestamp: "2026-10-12T13:51:36-07:00"
      dispatched: "2026-10-12T13:41:09-07:00"
      agent: claude
      dispose:
        - id: BR-11
          disposition: addressed
          note: pending-first + EXIT trap (local-gates-guard.yml:76-89) and merge requires the verdict while main carries the guard (cigate.go:264-272); both tests go red on revert.
          round: 5
        - id: BR-12
          disposition: addressed
          note: 20-ci-workflow-tests.sh globs scripts/test/*.test.sh with two reasoned exclusions; plan wording no longer claims pre-push.
          round: 5
      findings:
        - id: BR-13
          severity: Minor
          title: Legacy merge path reads a possibly stale origin/main to decide whether the guard verdict is required
          detail: 'merge.go:563 passes origin/main to runCIGate, but it is fetched only when inPlace (merge.go:534 admits it is stale). A local origin/main that predates the guard skips the requirement, so a guard run that never reported passes again. Rule: every runCIGate caller passes a freshly fetched base. Fix: fetch origin main before the gate, as lines 507 and 539 already do.'
          family: guard-decides-on-live-state
          round: 5
      boundary: M2
      recipe: milestone-review
      isolation: checkout
      reviewed: 719cf451355b8661b6662d69d8d0fc20f32e8c20
      blocked: false
---

# Gate ledger — ariadne#309 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T12:25:10-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `checks-bound-to-merged-head` Legacy merge path never verifies the checks gh reports belong to the pinned head
  merge.go:552-566 pins head from local origin/<branch> but gh pr checks reads GitHub's current PR head; right after a push the PR object can lag, old-head green checks pass the gate, and --match-head-commit only protects if GitHub has not yet synced. Compare PR headRefOid to head (treat mismatch as pending) and add a fake test.
- **BR-2** [Minor] `plan-revisions-record-scope-changes` Dropped M1 dry-run bullet is recorded only in the issue Log, not the plan Revisions
  Add a Revisions entry so the plan stops claiming the dry-run behaviour.
- **BR-3** [Minor] `gate-catalog-covers-all-refusals` no-ci RefusalPat misses the head-tree-read and cannot-pin refusals
  gatesig.go:163 matches "PR N has|reports" and "could not read", but not "CI gate: read <head>'s tree" or "CI gate: cannot pin the merge".
- **BR-4** [Minor] `test-helper-correctness` exitErr test helper builds the exit code via rune arithmetic, valid only for 0-9
  cigate_test.go:68; use strconv.Itoa.

## Round 2 — 2026-10-12T12:56:15-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — cigate.go:216-230 reads the head before and after the checks; reverting the comparison turns TestRunCIGate and TestRunMerge_LaggingPRHeadRefuses red.
- BR-2 — addressed — The plan's Revisions section now records the dropped dry-run bullet (2026-10-12, M1 implementation).
- BR-3 — not-addressed — The prefix pattern covers the class, but no classifier test pins any no-ci refusal: reverting gatesig.go to the 80160793 pattern leaves processmanual tests green. Add TestClassifyOutputLine cases for every refusal string, plus negatives for the CI checks: success line and the ACK.
- BR-4 — addressed — cigate_test.go:72 now uses strconv.Itoa.

## Round 3 — 2026-10-12T13:13:44-07:00 (claude) — BLOCKED

### Raised

- **BR-5** [Important] `guard-parses-input-losslessly` local-gates-guard misses paths git prints in quoted form (non-ASCII, quotes, backslashes)
  git diff --name-only without -z wraps such paths in quotes with escapes (reproduced: scripts/merge-checks.d/50-é.sh and .github/workflows/a"b.yml both pass as untouched). Use git -c core.quotePath=false diff -z and read -d ''; add a non-ASCII protected-path test case that fails before the fix.
- **BR-6** [Important] `guard-decides-on-live-state` Guard decides on the event-payload label snapshot with no per-PR concurrency group
  A re-run of an old labeled run, or a label removal racing a labeled run, can leave a success status after the label is gone. Read labels live via gh api at decision time (as the head is re-read) and add a per-PR concurrency group (ARCH-ORDER).
- **BR-7** [Important] `docs-cover-new-surface` README does not mention the seeded local-gates-guard, the operator-only label, or the push-to-main run
  README.md:186-192 documents the seeded CI workflow for consumers; the new seeded workflow and the local-gates-verified label are consumer- and operator-facing surface.
- **BR-8** [Important] `plan-revisions-record-scope-changes` M2 deviations from D4 (statuses write, verdict status context, head-moved no-op, branches filter) have no Revisions entry
  2nd finding in this family. Rule: at each milestone-close, compare every Decision and checklist bullet against what shipped; any added permission, output or context, dropped item, or new fail-open/no-op branch gets a Revisions line in the same commit. Sweep all of M2, not just this instance.
- **BR-9** [Minor] `tests-wired-into-suite` local-gates-guard.test.sh and portable-ci.test.sh run in no suite or CI job
  Manual-only, as before; the guard's regression test never runs automatically. The head-moved exit is also untested.
- **BR-10** [Minor] `live-check-covers-attachment` Confirm in the live check whether failed guard job runs mark main's tip red
  pull_request_target check runs may attach to the base SHA, which would muddy D7's main status.

## Round 4 — 2026-10-12T13:26:17-07:00 (claude) — BLOCKED

### Disposed

- BR-5 — addressed — local-gates-guard.yml:84-92 uses core.quotePath=false diff -z with read -d ''; reverting both makes the test fail on scripts/merge-checks.d/50-é.sh (verified in this checkout).
- BR-6 — addressed — Labels are read live via gh api, failing closed with exit 2; a per-PR concurrency group is in place; swapping in a payload LABELS variable makes the test fail (verified).
- BR-7 — addressed — README.md:192-201 documents the push-to-main run, the seeded local-gates-guard.yml, the operator-only label and gh label create, matching the workflow and atlas.
- BR-8 — addressed — Plan Revisions entry "M2 review round 1" sweeps permissions, branches filter, live labels, concurrency, the head-moved no-op, -z, env for D7, wiring, README and live check; small omissions are noted as Minor.
- BR-9 — addressed — scripts/merge-checks.d/20-ci-workflow-tests.sh (mode 755) runs both tests through run-merge-checks.sh in merge-check CI; the head-moved case is tested (posts nothing, exits 0).
- BR-10 — addressed — The plan's after-landing live check now explicitly checks whether the guard's job run shows on main's tip; this is a plan item, not runtime behavior.

### Raised

- **BR-11** [Important] `gate-fails-closed-on-absent-evidence` Guard exits before its decision (fetch/merge-base error, cancelled run, head-moved no-op) leave no verdict status, and absence reads as green
  local-gates-guard.yml:76-82 can exit under set -e before verdict() is defined, so no local-gates-guard/verdict lands on the head, and decideCIGate only refuses checks that exist. If pull_request_target job runs attach off the PR head (the BR-10 question), merge sees no guard and passes; the atlas claim "whichever way GitHub attaches" fails here. Concrete case: an operator re-run of an old event cancels the current head's run via the concurrency group, then no-ops on head-moved, leaving no verdict. Fix: define verdict() early and add an EXIT trap that posts failure on any non-zero exit after input validation, and consider posting pending on head-moved; add a test where the fetch fails and assert state=failure.
- **BR-12** [Minor] `tests-wired-into-suite` Script tests are wired by name, not by class; the Revisions "pre-push" claim is unsupported and slot-dependencies.test.sh runs nowhere
  2nd finding in this family. The rule: every scripts/test/*.test.sh is reached by an automated runner. 20-ci-workflow-tests.sh should glob scripts/test/*.test.sh, excluding explicitly any test that needs arguments (release-weave, which weave-release.yml runs). No ariadne pre-push hook calls run-merge-checks.sh, so correct the plan's "CI and pre-push" wording.

## Round 5 — 2026-10-12T13:51:36-07:00 (claude) — passed

### Disposed

- BR-11 — addressed — pending-first + EXIT trap (local-gates-guard.yml:76-89) and merge requires the verdict while main carries the guard (cigate.go:264-272); both tests go red on revert.
- BR-12 — addressed — 20-ci-workflow-tests.sh globs scripts/test/*.test.sh with two reasoned exclusions; plan wording no longer claims pre-push.

### Raised

- **BR-13** [Minor] `guard-decides-on-live-state` Legacy merge path reads a possibly stale origin/main to decide whether the guard verdict is required
  merge.go:563 passes origin/main to runCIGate, but it is fetched only when inPlace (merge.go:534 admits it is stale). A local origin/main that predates the guard skips the requirement, so a guard run that never reported passes again. Rule: every runCIGate caller passes a freshly fetched base. Fix: fetch origin main before the gate, as lines 507 and 539 already do.

## Open findings

- **BR-3** [Minor] `gate-catalog-covers-all-refusals` no-ci RefusalPat misses the head-tree-read and cannot-pin refusals
- **BR-13** [Minor] `guard-decides-on-live-state` Legacy merge path reads a possibly stale origin/main to decide whether the guard verdict is required
