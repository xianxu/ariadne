---
id: 000511
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '6ef9b693761093f19eb0e99166359eb428ecfa33' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T03:30:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "96c486c8", done: "bc4ffd7d"}
---

# landing trains: don't restart CI when main advances only by issue-detail commits (filings collided with train 197 twice on 10-20)

## Problem

Issue filings/publishes advanced main twice under train 197 on 10-20. Trains already retain suite results across docs-only reintegration, but publish a new train head and restart CI for each move.


`sdlc issue new` and `sdlc issue publish` commit issue details straight to main. On 10-20, two such commits (pair#745
and #746 filings) landed while pair train 197 waited on CI; each time the conductor re-integrated main and restarted
CI, delaying a train that had already passed its full suite. Filing and landing collide.

## Spec

Keep the exact train head whose suite, selected landing checks and CI passed when main advances only through issue-record Markdown. Reuse the single-car fast-path proof: per-commit movement classification (including merge parents and rename sources), same-PR/head/base CLEAN readiness, and landingCheckOwner.allowsMainMove. Recheck the train's instance/duplicate gates against the final fetched destination under the merge transaction, then send the existing expected-head merge. The server includes current main; never copy CI statuses to another commit.

Admit workshop/issue-cards/ Markdown alongside existing issue/plan/history Markdown. workshop/issue-tracker.json remains configuration and cannot qualify. A separate issue-tracker ref advance alone does not move main. Missing/stale readiness, malformed or unavailable movement, code/configuration, watched paths or invalid landing proof retain normal reintegration. Identity mismatch refuses, as on the single-car path. No train state/schema, retry budget or persisted artifact is added.


Options (choose in the plan):
1. **Trains ignore detail-only commits:** when main advances only by commits that touch `workshop/issues/` (tracker
   mirrors, issue details, archive moves), the conductor fast-forwards past them without re-running CI or the
   suite, since they can't change behavior. The merge stays a true merge with main.
2. Batch detail publishes so they don't land while a train's CI is in flight.

Option 1 is preferred: a mechanical, verifiable rule ("the diff since the verified base touches only workshop/
issues/ paths").

## Done when

- A real two-car train keeps its tested/published head, suite and CI evidence across issue-detail and card Markdown main advances; final main retains both new records and both cars.
- Configuration/code, watched landing paths, stale/unavailable/non-CLEAN readiness and code-add/revert cannot take the fast path; late duplicate IDs refuse before merge.
- Existing suite reuse and normal reintegration behavior remain covered; targeted controls, recorded full suite and boundary review pass before landing.


- A train whose main advanced only by issue-detail commits lands without restarting CI or the suite; any other
  commit still forces re-integration. Tests pin both.

## Plan

- [ ] Add failing real-Git/stateful-GitHub train regressions at the CI/readiness boundary, asserting unchanged head and one publication/CI cycle/suite; cover refused or fallback observations and late collisions.
- [ ] Reuse strict movement/readiness/landing proof in the train merge transaction; include card Markdown in the shared record classifier, and preserve normal reintegration on unavailable proof.
- [ ] Verify targeted train, classifier and landing-check controls; update README/atlas and record isolated proof and suite/close evidence through the workflow gates.

Actions: continue the current train; retry a refused landing with the named correction.

| Internal observation | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Exact checked head + record-only main + fresh CLEAN + valid landing proof | Train lands unchanged; new main records remain | Continue train | None |
| Missing/stale/non-CLEAN readiness, failed proof, code/config/watched movement | Existing reintegration and required checks | Continue train | New CI wait; suite rerun only under existing rules |
| Changed PR identity or duplicate ID | Refusal before merge | Retry after correcting identity/collision | None; cars retain branches |
| Merge acknowledgement uncertain | Existing probe/recovery path | Continue existing recovery | Outcome remains unconfirmed until observed |


- [ ]

## Log

### 2026-10-21

- Claimed on dispatch and started from fresh main after #510. Investigation: conductor.reintegrate already retains a docs-only suite, but the reducer republishes a new head and waits for fresh CI. The regression must assert head/publication/CI identity, not only suite count.
- ARCH-DRY: reuse mainMovementIssueOnly, readinessAllowsUnintegrated and landingCheckOwner.allowsMainMove; generic workshop docs classification alone is too broad (it permits tracker JSON). Alternative status copying would sever exact-head CI authority; broader docs exemptions would exceed the request. Keep the checked head instead.
- ARCH-ORDER/SECURE: observe readiness outside the transaction, compare its exact base with the final fetched main under the existing lock, and retain expected-head merge/probe. The lock does not freeze external publishers, so stale observations fall back; the server's CLEAN state and merge enforce current compatibility. No assertion of absence depends on a process scan.
- ARCH-PURE/MOCK: reuse pure path/readiness decisions with real Git histories and the existing stateful GitHub fake. Growth is one movement-history scan O(P) in changed path entries plus existing policy/car reads per observed main move; no per-record subprocess. Existing N=8/16 counted movement regression pins one history call. No durable record/format or background lifetime is added (ARCH-FUNERAL/EVOLVE). Existing bounded train retry/capacity remains (ARCH-CONSTRAINTS).
- Hub approved the shape and clarified card Markdown only, excluding workshop/issue-tracker.json (authenticated delivery 84b54387-9799-412f-a02e-133b1cbdfb2f). Expected implementation fits quick shell; if it exceeds 100 added executable/help lines, stop and plan full flow. Read-only delegated exploration confirmed the same reuse boundaries; no peer edits/tests.

### 2026-10-20
