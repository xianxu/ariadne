# Boundary Review — ariadne#287 (whole-issue close)

| field | value |
|-------|-------|
| issue | 287 — Reconcile merges done outside sdlc |
| repo | ariadne |
| issue file | workshop/issues/000287-reconcile-external-merges.md |
| boundary | whole-issue close |
| milestone | — |
| window | d042709c2bf937d5e44e510776ca09801910eb90..2e9298e16b790aa44411c01841cd9da4543307c3 |
| command | sdlc close --issue 287 |
| reviewer | claude |
| timestamp | 2026-10-08T18:30:38-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: medium
```

The feature works and does what the Spec asks. `sdlc state` reports issue branches merged outside sdlc, read-only, with the exact next action. `reconcile` and the next `sdlc merge` mark the card done, archive the details and plans in a narrow main commit, and delete the merged remote branch with a lease. Without the close, reconcile only prints the next action. All five #287 tests and the abandon suite pass at HEAD (`go test ./cmd/sdlc/ -run '...ExternalMerge...|TestAbandon'`, ok in 107s). Two things keep this from SHIP:

1. **The plan no longer matches the code.** The plan's Core concepts table and D5 describe a per-issue finisher, `finishExternalMerge(env, stderr, id)`, that reconcile runs "for that issue". That function doesn't exist. The code ships `finishLandedLeftovers(env, stderr, issuesDir)`, which works across every done card in the repo. The review rubric makes any table-vs-code contradiction Critical. The fix is only a `## Revisions` entry, so this rework is cheap.
2. **D1's network claim is wrong.** D1 says detection uses "no network beyond the fetch the reader already does", but `sdlc state` now fetches main and the tracker again on every run.

The rest are Minor.

**1. Strengths**
- `cmd/sdlc/archivemain.go:31`: `archiveIssueOnMain` is a clean extraction. Abandon and the new finisher share one archive path (ARCH-DRY), the abandon tests still pass unchanged, and the `Allow` hook keeps the ownership check with each caller.
- `externalmerge.go:58-62`: the first-parent rule ("a branch is merged if its tip is on main but not on main's first-parent line") is a sound way to tell a fresh branch from a merged one, and `TestStateIgnoresAFreshBranch` pins it.
- `externalMergeFindings` degrades to an `info` finding when it can't check (offline, no tracker), so `state` never fails because of it.
- The finisher is safe: it only deletes a remote branch the clone has fetched, that is contained in main, and that is still at its expected tip; failures warn and never fail the landing.
- The tests check the important side effect: `TestStateReportsExternalMerges` asserts the remote main and `issue-tracker` refs didn't move.

**2. Critical**
- Plan file, Core concepts table, D4 and D5 (against `externalmerge.go:184` and `issuerecovery.go:113`): the plan names the finisher `finishExternalMerge(env, stderr, id)` and says reconcile runs it "for that issue". The code runs `finishLandedLeftovers(env, stderr, issuesDir)` over every done card whose details are still on main, and it can delete other issues' remote branches. The issue's Log and the recovery catalog describe the actual behavior; the plan does not. **Fix:** append a `## Revisions` entry to the plan.

**3. Important**
- **ARCH-CONSTRAINTS** (`externalmerge.go:150-163`, `state.go:220`): `sdlc state` used to read the tracker with `PreferFresh`. Now it also calls `openTrackerAt`, `env.repo.Snapshot()` and `env.main.Snapshot()`, and each snapshot fetches. That adds about two fetches to every `state` run, which contradicts D1. Offline it degrades correctly to an info finding, but the latency is never stated or bounded. **Fix (cheap):** amend D1 to say what `state` now fetches and what happens offline, or reuse the tracker read `state` already does and fetch only main.

**4. Minor**
- **ARCH-FUNERAL** (`externalmerge.go:226-238`): if the remote branch delete fails after the archive, nothing ever retries it. A rerun skips the issue because its details are no longer live on main, so the branch stays on the remote until someone deletes it by hand.
- `branchTip` (`externalmerge.go:110`) prefers the remote-tracking ref without checking it is current. A stale tracking ref, for example for an issue reopened from done in another clone, could be reported as "merged outside sdlc".
- **ARCH-DRY**: the "evidence on main" check in `externalMerges` (`externalmerge.go:75-80`) re-implements the one in `ownedCompletions` (codecomplete, same repository, evidence is an ancestor of main). `state` and `reconcile` could drift apart; derive the settle set from `ownedCompletions` instead.
- `landing.go:497` passes `f.IssuesDir`, but `archiveIssueOnMain` takes its plans and history paths from the environment (`plansDir()`/`historyDir()`) and ignores `f.PlansDir`/`f.HistoryDir`.
- `finishLandedLeftovers` returns the IDs it archived, but both callers discard them (`_`). Drop the return value or use it.
- The settle message says "merged outside sdlc" even when an `sdlc merge` was interrupted or the close went straight to main via `sdlc push`. For an interrupted merge, the documented recovery is `sdlc merge --branch B --yes`, not reconcile.
- `state.go:227` finds the commit-subject finding to suppress by matching the message prefix `"looks done"`. That breaks silently if the message changes; key it on a structured field instead.

**5. Test coverage**
- `TestExternalMergeVerdict` restates the switch it tests, which is acceptable for a function this small.
- There is no test for a `blocked` card or for the offline path of `externalMergeFindings`.
- The race with an in-flight `sdlc merge` in another slot is untested; reading the code, it converges.

**6. Architectural notes**
- ARCH-PURE: pass. The verdict is pure; detection and finishing are thin IO glue.
- ARCH-MOCK: pass. Real git with a local remote, and the existing fake GitHub client for merge.
- ARCH-SECURE: pass. Only git refs and cards are read, and refs pass through argv.
- ARCH-ORDER: mostly pass. The finisher re-checks the card is done before publishing, and the delete is leased.
- ARCH-PURPOSE: pass. Both Done-when bullets are delivered.
- ARCH-FUNERAL: the remote branch can leak (see Minor).
- Squash/rebase detection will need the GitHub API, as the plan says.

**7. Plan revision recommendations**
- A Revisions entry: the finisher is `finishLandedLeftovers(env, stderr, issuesDir)`, applied to every done card still live on main (not per issue), with `reportUnclosedMerge` handling the named issue. Update the Core concepts row and D5 to match.
- A Revisions entry: `sdlc state` now fetches main and the tracker; give the offline behavior and a latency note (D1).

```findings
findings:
  - id: new
    severity: Critical
    family: plan-table-matches-code
    title: |
      Plan core-concepts table and D5 name a per-issue finishExternalMerge; code ships repo-wide finishLandedLeftovers
    detail: |
      The plan lists finishExternalMerge(env, stderr, id) run by reconcile for that issue only. Code has finishLandedLeftovers(env, stderr, issuesDir), which archives every done card still live on main and deletes their remote branches. Append a Revisions entry to the plan and correct the table row and D5.
  - id: new
    severity: Important
    family: declared-envelope-enforced
    title: |
      sdlc state now fetches main and the tracker every run, contradicting D1's no-extra-network claim
    detail: |
      externalMergeFindings calls openTrackerAt plus env.repo.Snapshot and env.main.Snapshot, each of which fetches, on every state run. Offline it degrades to an info finding, but the added latency is unstated. Amend D1 or reuse state's existing tracker read and fetch only main.
  - id: new
    severity: Minor
    family: cleanup-is-convergent
    title: |
      A failed remote-branch delete after archiving is never retried
    detail: |
      finishLandedLeftovers only revisits done cards whose details are still live on main, so after the archive succeeds and the delete warns, the merged branch stays on the remote with no automatic removal path.
  - id: new
    severity: Minor
    family: observation-freshness
    title: |
      branchTip prefers a possibly stale remote-tracking ref over the local branch
    detail: |
      A stale refs/remotes tracking ref, for example for an issue reopened from done in another clone, can be reported as merged outside sdlc.
  - id: new
    severity: Minor
    family: single-source-predicate
    title: |
      externalMerges re-implements ownedCompletions' evidence-on-main check
    detail: |
      state's settle verdict and reconcile's settle use two copies of the codecomplete, same-repository, evidence-is-ancestor check that can drift apart; derive the settle set from ownedCompletions.
  - id: new
    severity: Minor
    family: flags-thread-through
    title: |
      merge's finisher ignores f.PlansDir and f.HistoryDir
    detail: |
      archiveIssueOnMain uses the environment-derived plansDir() and historyDir(), while the merge hook passes f.IssuesDir from flags.
  - id: new
    severity: Minor
    family: unused-api-surface
    title: |
      finishLandedLeftovers returns archived IDs that both callers discard
  - id: new
    severity: Minor
    family: accurate-next-action
    title: |
      The settle finding says merged outside sdlc for interrupted sdlc merges and sdlc push closes
    detail: |
      For an interrupted sdlc merge the documented recovery is sdlc merge --branch B --yes, not reconcile.
  - id: new
    severity: Minor
    family: structured-not-string-match
    title: |
      state suppresses the close-off finding by matching the message prefix looks done
```
