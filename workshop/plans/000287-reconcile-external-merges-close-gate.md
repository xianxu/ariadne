---
gate: boundary-review
issue: 287
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-08T18:30:38-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Critical
          title: Plan core-concepts table and D5 name a per-issue finishExternalMerge; code ships repo-wide finishLandedLeftovers
          detail: The plan lists finishExternalMerge(env, stderr, id) run by reconcile for that issue only. Code has finishLandedLeftovers(env, stderr, issuesDir), which archives every done card still live on main and deletes their remote branches. Append a Revisions entry to the plan and correct the table row and D5.
          family: plan-table-matches-code
          round: 1
        - id: BR-2
          severity: Important
          title: sdlc state now fetches main and the tracker every run, contradicting D1's no-extra-network claim
          detail: externalMergeFindings calls openTrackerAt plus env.repo.Snapshot and env.main.Snapshot, each of which fetches, on every state run. Offline it degrades to an info finding, but the added latency is unstated. Amend D1 or reuse state's existing tracker read and fetch only main.
          family: declared-envelope-enforced
          round: 1
        - id: BR-3
          severity: Minor
          title: A failed remote-branch delete after archiving is never retried
          detail: finishLandedLeftovers only revisits done cards whose details are still live on main, so after the archive succeeds and the delete warns, the merged branch stays on the remote with no automatic removal path.
          family: cleanup-is-convergent
          round: 1
        - id: BR-4
          severity: Minor
          title: branchTip prefers a possibly stale remote-tracking ref over the local branch
          detail: A stale refs/remotes tracking ref, for example for an issue reopened from done in another clone, can be reported as merged outside sdlc.
          family: observation-freshness
          round: 1
        - id: BR-5
          severity: Minor
          title: externalMerges re-implements ownedCompletions' evidence-on-main check
          detail: state's settle verdict and reconcile's settle use two copies of the codecomplete, same-repository, evidence-is-ancestor check that can drift apart; derive the settle set from ownedCompletions.
          family: single-source-predicate
          round: 1
        - id: BR-6
          severity: Minor
          title: merge's finisher ignores f.PlansDir and f.HistoryDir
          detail: archiveIssueOnMain uses the environment-derived plansDir() and historyDir(), while the merge hook passes f.IssuesDir from flags.
          family: flags-thread-through
          round: 1
        - id: BR-7
          severity: Minor
          title: finishLandedLeftovers returns archived IDs that both callers discard
          family: unused-api-surface
          round: 1
        - id: BR-8
          severity: Minor
          title: The settle finding says merged outside sdlc for interrupted sdlc merges and sdlc push closes
          detail: For an interrupted sdlc merge the documented recovery is sdlc merge --branch B --yes, not reconcile.
          family: accurate-next-action
          round: 1
        - id: BR-9
          severity: Minor
          title: state suppresses the close-off finding by matching the message prefix looks done
          family: structured-not-string-match
          round: 1
      recipe: milestone-review
      blocked: true
---

# Gate ledger — ariadne#287 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-08T18:30:38-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `plan-table-matches-code` Plan core-concepts table and D5 name a per-issue finishExternalMerge; code ships repo-wide finishLandedLeftovers
  The plan lists finishExternalMerge(env, stderr, id) run by reconcile for that issue only. Code has finishLandedLeftovers(env, stderr, issuesDir), which archives every done card still live on main and deletes their remote branches. Append a Revisions entry to the plan and correct the table row and D5.
- **BR-2** [Important] `declared-envelope-enforced` sdlc state now fetches main and the tracker every run, contradicting D1's no-extra-network claim
  externalMergeFindings calls openTrackerAt plus env.repo.Snapshot and env.main.Snapshot, each of which fetches, on every state run. Offline it degrades to an info finding, but the added latency is unstated. Amend D1 or reuse state's existing tracker read and fetch only main.
- **BR-3** [Minor] `cleanup-is-convergent` A failed remote-branch delete after archiving is never retried
  finishLandedLeftovers only revisits done cards whose details are still live on main, so after the archive succeeds and the delete warns, the merged branch stays on the remote with no automatic removal path.
- **BR-4** [Minor] `observation-freshness` branchTip prefers a possibly stale remote-tracking ref over the local branch
  A stale refs/remotes tracking ref, for example for an issue reopened from done in another clone, can be reported as merged outside sdlc.
- **BR-5** [Minor] `single-source-predicate` externalMerges re-implements ownedCompletions' evidence-on-main check
  state's settle verdict and reconcile's settle use two copies of the codecomplete, same-repository, evidence-is-ancestor check that can drift apart; derive the settle set from ownedCompletions.
- **BR-6** [Minor] `flags-thread-through` merge's finisher ignores f.PlansDir and f.HistoryDir
  archiveIssueOnMain uses the environment-derived plansDir() and historyDir(), while the merge hook passes f.IssuesDir from flags.
- **BR-7** [Minor] `unused-api-surface` finishLandedLeftovers returns archived IDs that both callers discard
- **BR-8** [Minor] `accurate-next-action` The settle finding says merged outside sdlc for interrupted sdlc merges and sdlc push closes
  For an interrupted sdlc merge the documented recovery is sdlc merge --branch B --yes, not reconcile.
- **BR-9** [Minor] `structured-not-string-match` state suppresses the close-off finding by matching the message prefix looks done

## Open findings

- **BR-1** [Critical] `plan-table-matches-code` Plan core-concepts table and D5 name a per-issue finishExternalMerge; code ships repo-wide finishLandedLeftovers
- **BR-2** [Important] `declared-envelope-enforced` sdlc state now fetches main and the tracker every run, contradicting D1's no-extra-network claim
- **BR-3** [Minor] `cleanup-is-convergent` A failed remote-branch delete after archiving is never retried
- **BR-4** [Minor] `observation-freshness` branchTip prefers a possibly stale remote-tracking ref over the local branch
- **BR-5** [Minor] `single-source-predicate` externalMerges re-implements ownedCompletions' evidence-on-main check
- **BR-6** [Minor] `flags-thread-through` merge's finisher ignores f.PlansDir and f.HistoryDir
- **BR-7** [Minor] `unused-api-surface` finishLandedLeftovers returns archived IDs that both callers discard
- **BR-8** [Minor] `accurate-next-action` The settle finding says merged outside sdlc for interrupted sdlc merges and sdlc push closes
- **BR-9** [Minor] `structured-not-string-match` state suppresses the close-off finding by matching the message prefix looks done
