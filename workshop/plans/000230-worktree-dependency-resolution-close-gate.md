---
gate: boundary-review
issue: 230
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T13:25:50-07:00"
      dispatched: "2026-10-13T13:18:44-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Dangling or unstat-able local substrate target is classified Absent and silently bypassed for the primary fallback
          detail: pkg/layergraph/anchor.go:51-58 probes with FS.Stat (follows symlinks); a dangling fleet/worktree/leaf/base symlink yields Walk=[fleet/base, br] with no error (reproduced). Spec item 5 requires an existing invalid local target to be an error. Probe presence with Lstat (or treat non-IsNotExist Stat errors as NotLayer) and add a regression row.
          family: existing-target-not-bypassed
          round: 1
        - id: BR-2
          severity: Important
          title: No-anchor reasons are computed but dropped; compile missing-substrate errors never name why the fallback did not apply
          detail: acquire.go:320-333 discards every reason except FallbackAbsentError (bare primary, Git failure, unavailable primary), contradicting plan Decision 5, the ticked issue Task 4 'anchor reason in missing errors', and Done-when bullet 4; a sourced row then silently clones a duplicate beside the worktree. DeclaredSubstrate.Via/Unanchored have zero production consumers. Wrap the missing-substrate errors with the reason and test via Client.Anchors, or revise the plan.
          family: computed-diagnostic-not-surfaced
          round: 1
        - id: BR-3
          severity: Important
          title: README not updated for linked-worktree resolution; its 'record a source to restore a missing checkout' is now false in linked worktrees
          detail: README.md:103-110 covers primary and Couch-slot behavior only; the new fallback, its compile notice, and the absent-everywhere refusal are documented only in atlas.
          family: readme-tracks-user-surface
          round: 1
        - id: BR-4
          severity: Minor
          title: NewResolver defaults every FS (including in-memory fakes) to GitAnchors; no fake implements Anchors despite the plan's hermeticity claim
          family: fake-seam-unwired
          round: 1
        - id: BR-5
          severity: Minor
          title: workspace.ExecGit duplicates cmd/weave setupGitReader and lacks a context/timeout
          family: shared-git-reader
          round: 1
        - id: BR-6
          severity: Minor
          title: Atlas says the absent-everywhere refusal happens before compile writes anything; earlier rows may already have been acquired
          family: doc-overclaims-behavior
          round: 1
        - id: BR-7
          severity: Minor
          title: DeclaredSubstrate.Unanchored also carries FallbackAbsentError, where an anchor did exist
          family: field-name-matches-meaning
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-13T13:50:31-07:00"
      dispatched: "2026-10-13T13:45:01-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Resolver.probe uses Lstat (anchor.go:61); walk uses res.Probe; TestWalkLinkedWorktreeDanglingLocalNotBypassed goes red with the Lstat path disabled.
          round: 2
        - id: BR-2
          disposition: addressed
          note: 'NoAnchorError typed; Restore appends "no worktree fallback: <reason>"; TestRestoreMissingNamesNoFallbackReason goes red when the append is reverted; anchor failures fail closed (tested).'
          round: 2
        - id: BR-3
          disposition: addressed
          note: README.md:112-119 describes the linked-worktree fallback, compile notice and the absent-everywhere refusal; matches acquire.go:322-331 behavior.
          round: 2
        - id: BR-4
          disposition: addressed
          note: TestWalkUsesTheFSAnchors wires an FS implementing Anchors with no Git; in-memory fakes without it hit Lstat-not-exist so no Git runs.
          round: 2
        - id: BR-5
          disposition: not-addressed
          note: Duplicate removed and ExecGit gained Context, but GitAnchors (anchor.go:24) still builds ExecGit{} without one, so Restore's anchor Git calls ignore its ctx.
          round: 2
        - id: BR-6
          disposition: addressed
          note: setup-and-replication.md now says composition writes nothing and already-acquired dependencies stay, matching Restore's BFS order.
          round: 2
        - id: BR-7
          disposition: addressed
          note: Renamed to NoFallback with doc covering both NoAnchorError and FallbackAbsentError (walk.go:47-49).
          round: 2
      findings:
        - id: BR-8
          severity: Important
          title: weave refresh declaredOrigins re-resolves substrate rows beside the worktree, bypassing layergraph.Resolver; fails from a linked worktree
          detail: refresh.go:35-40 joins s.path+row.Path and EvalSymlinks it, while snapshots come from Restore's fallback-resolved Layers; a sourced ../base absent beside the worktree errors with ENOENT. Use Restore's resolved edges or the Resolver, and add a linked-worktree refresh test.
          family: single-resolver-all-consumers
          round: 2
        - id: BR-9
          severity: Important
          title: 'Plan still states the pre-fix contract: Git failure as a no-anchor value, "never fatal", Unanchored, Resolution without NoFallback'
          detail: '2nd finding in this family. Rule: a review fix that changes a contract updates every artifact restating it in the same commit; enumerate by grepping the old identifier and semantic phrase across workshop/plans, atlas, README and doc comments. Measured stale sites: plan:57-62, 80, 96-97, 106-108, 273, 341. Add a plan Revisions entry.'
          family: doc-overclaims-behavior
          round: 2
        - id: BR-10
          severity: Minor
          title: sdlc substrateWalk's fail-closed anchor-error branch is untested, has no Anchors seam, and re-derives the local path
          detail: startplan.go:464-475 builds NewResolver(OSFS) inline, so the branch is unreachable from tests; it also skips the seen-dedupe.
          family: behavior-branch-needs-regression
          round: 2
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 3
      timestamp: "2026-10-13T14:51:32-07:00"
      dispatched: "2026-10-13T14:45:03-07:00"
      agent: claude
      dispose:
        - id: BR-5
          disposition: addressed
          note: 'setupGitReader removed; workspace.ExecGit{Context} used in environment.go:21 and main.go:636. Residual: GitAnchors still uses ExecGit{} with no context (anchor.go:24), note only.'
          round: 3
        - id: BR-8
          disposition: addressed
          note: refresh.go:29-47 resolves via layergraph.Resolver with c.Anchors; TestRefreshLinkedWorktreeSubstrate fails with ENOENT on worktree/host/dep when the fix is reverted, passes at head.
          round: 3
        - id: BR-9
          disposition: addressed
          note: Plan sites 57-62, 80, 97, 106-114, 273, 341 are marked Superseded and point to Revisions 2026-10-13, which states the typed no-anchor, the fatal Git failure and the NoFallback contract.
          round: 3
        - id: BR-10
          disposition: addressed
          note: substrateWalkWith seam (startplan.go:446); TestSubstrateWalkAnchorFailureReportsAbsent covers the Git-failure branch and absent dedupe; the path comes from res.Path.
          round: 3
      findings:
        - id: BR-11
          severity: Minor
          title: Resolver doc comment and atlas weave.md list its consumers without refresh's declaredOrigins
          detail: 'This is the 3rd finding in family doc-overclaims-behavior. Rule: any prose that lists a component''s consumers or contract must be regenerated from a grep of the component''s call sites whenever a fix adds or removes one. Here, grep NewResolver/Resolver.Substrate callers and update both lists in pkg/layergraph/anchor.go:28-30 and atlas/workflow/weave.md:158-159, which still omit cmd/weave/internal/refresh after BR-8.'
          family: doc-overclaims-behavior
          round: 3
      recipe: milestone-review
      isolation: checkout
      reviewed: 69aa17f639a9ef681c2ddf8396cdbc3d9eadb6ea
      blocked: false
---

# Gate ledger — ariadne#230 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T13:25:50-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `existing-target-not-bypassed` Dangling or unstat-able local substrate target is classified Absent and silently bypassed for the primary fallback
  pkg/layergraph/anchor.go:51-58 probes with FS.Stat (follows symlinks); a dangling fleet/worktree/leaf/base symlink yields Walk=[fleet/base, br] with no error (reproduced). Spec item 5 requires an existing invalid local target to be an error. Probe presence with Lstat (or treat non-IsNotExist Stat errors as NotLayer) and add a regression row.
- **BR-2** [Important] `computed-diagnostic-not-surfaced` No-anchor reasons are computed but dropped; compile missing-substrate errors never name why the fallback did not apply
  acquire.go:320-333 discards every reason except FallbackAbsentError (bare primary, Git failure, unavailable primary), contradicting plan Decision 5, the ticked issue Task 4 'anchor reason in missing errors', and Done-when bullet 4; a sourced row then silently clones a duplicate beside the worktree. DeclaredSubstrate.Via/Unanchored have zero production consumers. Wrap the missing-substrate errors with the reason and test via Client.Anchors, or revise the plan.
- **BR-3** [Important] `readme-tracks-user-surface` README not updated for linked-worktree resolution; its 'record a source to restore a missing checkout' is now false in linked worktrees
  README.md:103-110 covers primary and Couch-slot behavior only; the new fallback, its compile notice, and the absent-everywhere refusal are documented only in atlas.
- **BR-4** [Minor] `fake-seam-unwired` NewResolver defaults every FS (including in-memory fakes) to GitAnchors; no fake implements Anchors despite the plan's hermeticity claim
- **BR-5** [Minor] `shared-git-reader` workspace.ExecGit duplicates cmd/weave setupGitReader and lacks a context/timeout
- **BR-6** [Minor] `doc-overclaims-behavior` Atlas says the absent-everywhere refusal happens before compile writes anything; earlier rows may already have been acquired
- **BR-7** [Minor] `field-name-matches-meaning` DeclaredSubstrate.Unanchored also carries FallbackAbsentError, where an anchor did exist

## Round 2 — 2026-10-13T13:50:31-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Resolver.probe uses Lstat (anchor.go:61); walk uses res.Probe; TestWalkLinkedWorktreeDanglingLocalNotBypassed goes red with the Lstat path disabled.
- BR-2 — addressed — NoAnchorError typed; Restore appends "no worktree fallback: <reason>"; TestRestoreMissingNamesNoFallbackReason goes red when the append is reverted; anchor failures fail closed (tested).
- BR-3 — addressed — README.md:112-119 describes the linked-worktree fallback, compile notice and the absent-everywhere refusal; matches acquire.go:322-331 behavior.
- BR-4 — addressed — TestWalkUsesTheFSAnchors wires an FS implementing Anchors with no Git; in-memory fakes without it hit Lstat-not-exist so no Git runs.
- BR-5 — not-addressed — Duplicate removed and ExecGit gained Context, but GitAnchors (anchor.go:24) still builds ExecGit{} without one, so Restore's anchor Git calls ignore its ctx.
- BR-6 — addressed — setup-and-replication.md now says composition writes nothing and already-acquired dependencies stay, matching Restore's BFS order.
- BR-7 — addressed — Renamed to NoFallback with doc covering both NoAnchorError and FallbackAbsentError (walk.go:47-49).

### Raised

- **BR-8** [Important] `single-resolver-all-consumers` weave refresh declaredOrigins re-resolves substrate rows beside the worktree, bypassing layergraph.Resolver; fails from a linked worktree
  refresh.go:35-40 joins s.path+row.Path and EvalSymlinks it, while snapshots come from Restore's fallback-resolved Layers; a sourced ../base absent beside the worktree errors with ENOENT. Use Restore's resolved edges or the Resolver, and add a linked-worktree refresh test.
- **BR-9** [Important] `doc-overclaims-behavior` Plan still states the pre-fix contract: Git failure as a no-anchor value, "never fatal", Unanchored, Resolution without NoFallback
  2nd finding in this family. Rule: a review fix that changes a contract updates every artifact restating it in the same commit; enumerate by grepping the old identifier and semantic phrase across workshop/plans, atlas, README and doc comments. Measured stale sites: plan:57-62, 80, 96-97, 106-108, 273, 341. Add a plan Revisions entry.
- **BR-10** [Minor] `behavior-branch-needs-regression` sdlc substrateWalk's fail-closed anchor-error branch is untested, has no Anchors seam, and re-derives the local path
  startplan.go:464-475 builds NewResolver(OSFS) inline, so the branch is unreachable from tests; it also skips the seen-dedupe.

## Round 3 — 2026-10-13T14:51:32-07:00 (claude) — passed

### Disposed

- BR-5 — addressed — setupGitReader removed; workspace.ExecGit{Context} used in environment.go:21 and main.go:636. Residual: GitAnchors still uses ExecGit{} with no context (anchor.go:24), note only.
- BR-8 — addressed — refresh.go:29-47 resolves via layergraph.Resolver with c.Anchors; TestRefreshLinkedWorktreeSubstrate fails with ENOENT on worktree/host/dep when the fix is reverted, passes at head.
- BR-9 — addressed — Plan sites 57-62, 80, 97, 106-114, 273, 341 are marked Superseded and point to Revisions 2026-10-13, which states the typed no-anchor, the fatal Git failure and the NoFallback contract.
- BR-10 — addressed — substrateWalkWith seam (startplan.go:446); TestSubstrateWalkAnchorFailureReportsAbsent covers the Git-failure branch and absent dedupe; the path comes from res.Path.

### Raised

- **BR-11** [Minor] `doc-overclaims-behavior` Resolver doc comment and atlas weave.md list its consumers without refresh's declaredOrigins
  This is the 3rd finding in family doc-overclaims-behavior. Rule: any prose that lists a component's consumers or contract must be regenerated from a grep of the component's call sites whenever a fix adds or removes one. Here, grep NewResolver/Resolver.Substrate callers and update both lists in pkg/layergraph/anchor.go:28-30 and atlas/workflow/weave.md:158-159, which still omit cmd/weave/internal/refresh after BR-8.

## Open findings

- **BR-11** [Minor] `doc-overclaims-behavior` Resolver doc comment and atlas weave.md list its consumers without refresh's declaredOrigins
