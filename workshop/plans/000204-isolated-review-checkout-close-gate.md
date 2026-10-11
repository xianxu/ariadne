---
gate: boundary-review
issue: 204
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-11T08:06:20-07:00"
      dispatched: "2026-10-11T08:02:39-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Critical
          title: code-review.md tells every reviewer it runs in a disposable checkout, but ad-hoc sdlc judge milestone-review still runs in the source checkout
          detail: cmd/sdlc/judge.go:109-126 renders the same template with RepoRoot set to the live checkout, so the prompt now tells those reviewers to revert fixes in the operator's tree (the parley#237 failure). Render the sentence only for isolated dispatches; keep the scratch-copy wording otherwise; pin both renderings with golden tests.
          family: isolation-claim-on-unisolated-path
          round: 1
        - id: BR-2
          severity: Important
          title: No-verdict retry in judge.Dispatch reuses the same checkout, so attempt 2 can inspect attempt 1's reverted or mutated tree
          detail: Retries fire exactly when a reviewer ended mid-work (#271/#300). Add a per-attempt reset or recreate hook (checkout --force head, clean -fdx, re-overlay) and a test with residue left by attempt 1.
          family: retry-attempt-shares-mutable-state
          round: 1
        - id: BR-3
          severity: Important
          title: One checkout that cannot be removed makes Sweep fail, which fails Create and turns every later boundary review into NotRun
          detail: 'Reproduced: a read-only directory in an orphan makes Sweep return unlinkat permission denied. Close also deletes the lock when RemoveAll fails, creating exactly that lockless orphan. Make the sweep best-effort with a warning, chmod u+w before removing, and keep the lock when RemoveAll fails.'
          family: sweep-failure-blocks-dispatch
          round: 1
        - id: BR-4
          severity: Important
          title: The fenced path has never been seen passing live, and protect paths are not resolved through symlinks before reaching Seatbelt
          detail: TestFenceDeniesSourceWrites skips wherever the probe fails and has not been seen green. Its t.TempDir paths under /var are symlinks, and Seatbelt subpath rules are generally believed to match canonical paths. EvalSymlinks the Protect paths and Dir, run the test once unsandboxed, and log the result.
          family: fence-unverified-live
          round: 1
        - id: BR-5
          severity: Important
          title: The deny-list fence turns off the agents' own sandboxes but protects only the source worktree and git common dir, leaving sibling worktrees and HOME writable
          detail: Add every path from git worktree list --porcelain to Protect, or move to an allow-list profile; note the remaining widening in the atlas.
          family: fence-scope-wider-than-agent-sandbox
          round: 1
        - id: BR-6
          severity: Minor
          title: Fenced gemini gets --sandbox=false while the plan says gemini is unchanged
          family: plan-code-drift
          round: 1
        - id: BR-7
          severity: Minor
          title: TestArchitecture_NarrativeRoutesToArchPrinciples fails in a review checkout because AGENTS.md is untracked
          family: tests-depend-on-untracked-files
          round: 1
        - id: BR-8
          severity: Minor
          title: lock.tmp files left by a killed Create are never swept
          family: artifact-without-removal-path
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-11T08:40:53-07:00"
      dispatched: "2026-10-11T08:26:46-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: revertCheck(in.Isolated) at judge/review.go:48; TestRevertCheckFollowsIsolation fails when Isolated is forced true; goldens keep the scratch-copy text.
          round: 2
        - id: BR-2
          disposition: addressed
          note: BeforeRetry=checkout.Reset; TestBoundaryReviewRetryStartsFromAPristineCheckout fails without it ("attempt 2 read attempt 1's mutation").
          round: 2
        - id: BR-3
          disposition: addressed
          note: Sweep is best effort and runs outside Create, removeTree restores write permission, Close keeps the lock on failure; each has a test that fails when its change is reverted.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Paths are canonicalized; the probe now proves a denial through the production Wrap before reporting fenced, so a false stamp cannot be recorded. The live test is still unrun, noted under ARCH-MOCK.
          round: 2
        - id: BR-5
          disposition: addressed
          note: Every git worktree list path is protected; TestBoundaryReviewFencedArgvAndStamp fails without it; the atlas states the HOME widening.
          round: 2
        - id: BR-6
          disposition: addressed
          note: The plan Revisions entry corrects "gemini unchanged" to match --sandbox=false.
          round: 2
        - id: BR-7
          disposition: addressed
          note: The test reads the tracked AGENTS.base.md and passes in this clean clone.
          round: 2
        - id: BR-8
          disposition: addressed
          note: Sweep reclaims unheld .lock.tmp; TestSweepReclaimsDeadOwnerAndSparesLiveOne fails without it.
          round: 2
      findings:
        - id: BR-9
          severity: Minor
          title: A concurrent sweep can reclaim a just-created <id>.lock.tmp before Create locks it, so that review fails as NotRun
          detail: '2nd finding in this family. Rule: the sweep may only act on a name whose owner is provably dead, so every name it acts on must be locked from creation. Fix the rule: create the lock file and lock it in one call (O_EXLOCK on darwin), or skip .lock.tmp younger than a short grace period. Then test a sweep reclaiming the tmp file just before Create locks it. The comment at checkout.go:66 is now false.'
          family: sweep-failure-blocks-dispatch
          round: 2
        - id: BR-10
          severity: Minor
          title: boundaryReviewDispatchOptions (no isolation) is now called only by tests
          detail: closereview_test.go and smalldiffreplay_test.go check prompts that production never sends; point them at boundaryReviewOptionsFor with an isolation, or delete the helper.
          family: tests-pin-unused-path
          round: 2
      recipe: milestone-review
      isolation: checkout
      reviewed: 388cceb8bbf6cdce4b428171d7705d3801b785cf
      blocked: false
    - "n": 3
      timestamp: "2026-10-11T08:46:20-07:00"
      dispatched: "2026-10-11T08:44:31-07:00"
      agent: claude
      dispose:
        - id: BR-9
          disposition: addressed
          note: Sweep skips .lock.tmp younger than tmpLockGrace (checkout.go:175); the comment at Create was corrected; the test spares a fresh tmp and fails (4 removed vs 3) when the grace check is reverted.
          round: 3
        - id: BR-10
          disposition: addressed
          note: boundaryReviewDispatchOptions deleted; tests use plannedBoundaryReviewOptions (the production dry-run builder via boundaryReviewOptionsFor with isolation); the replay harness opens a real checkout.
          round: 3
      recipe: fix-review
      isolation: checkout
      reviewed: 09a5b7643d10dcb84a430d5ee30ad7a8fbd1b485
      blocked: false
---

# Gate ledger — ariadne#204 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-11T08:06:20-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `isolation-claim-on-unisolated-path` code-review.md tells every reviewer it runs in a disposable checkout, but ad-hoc sdlc judge milestone-review still runs in the source checkout
  cmd/sdlc/judge.go:109-126 renders the same template with RepoRoot set to the live checkout, so the prompt now tells those reviewers to revert fixes in the operator's tree (the parley#237 failure). Render the sentence only for isolated dispatches; keep the scratch-copy wording otherwise; pin both renderings with golden tests.
- **BR-2** [Important] `retry-attempt-shares-mutable-state` No-verdict retry in judge.Dispatch reuses the same checkout, so attempt 2 can inspect attempt 1's reverted or mutated tree
  Retries fire exactly when a reviewer ended mid-work (#271/#300). Add a per-attempt reset or recreate hook (checkout --force head, clean -fdx, re-overlay) and a test with residue left by attempt 1.
- **BR-3** [Important] `sweep-failure-blocks-dispatch` One checkout that cannot be removed makes Sweep fail, which fails Create and turns every later boundary review into NotRun
  Reproduced: a read-only directory in an orphan makes Sweep return unlinkat permission denied. Close also deletes the lock when RemoveAll fails, creating exactly that lockless orphan. Make the sweep best-effort with a warning, chmod u+w before removing, and keep the lock when RemoveAll fails.
- **BR-4** [Important] `fence-unverified-live` The fenced path has never been seen passing live, and protect paths are not resolved through symlinks before reaching Seatbelt
  TestFenceDeniesSourceWrites skips wherever the probe fails and has not been seen green. Its t.TempDir paths under /var are symlinks, and Seatbelt subpath rules are generally believed to match canonical paths. EvalSymlinks the Protect paths and Dir, run the test once unsandboxed, and log the result.
- **BR-5** [Important] `fence-scope-wider-than-agent-sandbox` The deny-list fence turns off the agents' own sandboxes but protects only the source worktree and git common dir, leaving sibling worktrees and HOME writable
  Add every path from git worktree list --porcelain to Protect, or move to an allow-list profile; note the remaining widening in the atlas.
- **BR-6** [Minor] `plan-code-drift` Fenced gemini gets --sandbox=false while the plan says gemini is unchanged
- **BR-7** [Minor] `tests-depend-on-untracked-files` TestArchitecture_NarrativeRoutesToArchPrinciples fails in a review checkout because AGENTS.md is untracked
- **BR-8** [Minor] `artifact-without-removal-path` lock.tmp files left by a killed Create are never swept

## Round 2 — 2026-10-11T08:40:53-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — revertCheck(in.Isolated) at judge/review.go:48; TestRevertCheckFollowsIsolation fails when Isolated is forced true; goldens keep the scratch-copy text.
- BR-2 — addressed — BeforeRetry=checkout.Reset; TestBoundaryReviewRetryStartsFromAPristineCheckout fails without it ("attempt 2 read attempt 1's mutation").
- BR-3 — addressed — Sweep is best effort and runs outside Create, removeTree restores write permission, Close keeps the lock on failure; each has a test that fails when its change is reverted.
- BR-4 — addressed — Paths are canonicalized; the probe now proves a denial through the production Wrap before reporting fenced, so a false stamp cannot be recorded. The live test is still unrun, noted under ARCH-MOCK.
- BR-5 — addressed — Every git worktree list path is protected; TestBoundaryReviewFencedArgvAndStamp fails without it; the atlas states the HOME widening.
- BR-6 — addressed — The plan Revisions entry corrects "gemini unchanged" to match --sandbox=false.
- BR-7 — addressed — The test reads the tracked AGENTS.base.md and passes in this clean clone.
- BR-8 — addressed — Sweep reclaims unheld .lock.tmp; TestSweepReclaimsDeadOwnerAndSparesLiveOne fails without it.

### Raised

- **BR-9** [Minor] `sweep-failure-blocks-dispatch` A concurrent sweep can reclaim a just-created <id>.lock.tmp before Create locks it, so that review fails as NotRun
  2nd finding in this family. Rule: the sweep may only act on a name whose owner is provably dead, so every name it acts on must be locked from creation. Fix the rule: create the lock file and lock it in one call (O_EXLOCK on darwin), or skip .lock.tmp younger than a short grace period. Then test a sweep reclaiming the tmp file just before Create locks it. The comment at checkout.go:66 is now false.
- **BR-10** [Minor] `tests-pin-unused-path` boundaryReviewDispatchOptions (no isolation) is now called only by tests
  closereview_test.go and smalldiffreplay_test.go check prompts that production never sends; point them at boundaryReviewOptionsFor with an isolation, or delete the helper.

## Round 3 — 2026-10-11T08:46:20-07:00 (claude) — passed

### Disposed

- BR-9 — addressed — Sweep skips .lock.tmp younger than tmpLockGrace (checkout.go:175); the comment at Create was corrected; the test spares a fresh tmp and fails (4 removed vs 3) when the grace check is reverted.
- BR-10 — addressed — boundaryReviewDispatchOptions deleted; tests use plannedBoundaryReviewOptions (the production dry-run builder via boundaryReviewOptionsFor with isolation); the replay harness opens a real checkout.

## Open findings

(none — every finding has been disposed)
