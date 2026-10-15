---
gate: boundary-review
issue: 373
id_prefix: BR
rounds:
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 6a45043f308811eb127044b83ebcdfbb788713e1
        tracker_head: e729638a8173ef693fdbe3b8781bccd5497515c4
        branch_point: 72bf044e19b3be76bc22851008aaf9fbf2f30dc8
        since: "2026-10-15T01:00:50-07:00"
        until: "2026-10-15T03:46:10-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T01:00:50-07:00"
              end: "2026-10-15T03:46:10-07:00"
              began: "2026-10-15T01:00:50-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 24c8d79e-33f9-4ab1-a6b5-8cb02e00580a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 9c6ff9b4b733fce50c31bd9cd95d472767e8d3401fae027a43bbcdde98a76604
              started: "2026-10-12T14:18:17.059Z"
              spawned: false
              bytes: 11516040
              windows:
                - start: "2026-10-15T01:00:50-07:00"
                  end: "2026-10-15T03:46:10-07:00"
      "n": 1
      timestamp: "2026-10-15T03:55:51-07:00"
      dispatched: "2026-10-15T03:46:53-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: 'README project section is stale: close example now fails without --verified; project tick, --dry-run, landing_tick undocumented'
          detail: README.md:468 shows `sdlc project close --slug example`, which now refuses (--verified required, projectclose.go:117). The new `project tick` verb, close --dry-run/--verified, status mvp/done_when output, landing_tick, and --force waiving the MVP gate are absent from README.
          family: docs-surface-not-swept
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: cmd/tl-lead/help.md contradicts the new linked-close and landing-tick behavior
          detail: 'help.md:85-92 tells TLs to move the team directory to teams/<team>/history/projects/ and drop the symlink at project close, and that workers never write project.md / TLs tick rows by hand. The code now updates the real file in place and moves the link into workshop/history/projects (folder stays), and merge ticks rows (opt-in via landing_tick) or prints `sdlc project tick`. ARCH-PURPOSE: a consumer not swept.'
          family: docs-surface-not-swept
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: Landing tick does unbounded peer-origin network IO while holding the merge lock
          detail: tickOnLanding runs inside f.withLock (landing.go:625, merge.go:750) and uses peerTrunk with commandContext (no deadline), unlike the landing's boundedLandingGit. A hung peer fetch/push blocks other slots' merges for warn-only work. Move it after the locked closure and give it a timeout (ARCH-CONSTRAINTS).
          family: optional-work-under-critical-lock
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: commitLinkedProjectClose leaves both links if removing the live link fails; rerun then refuses
          detail: 'projectclose.go:308: the real file is already closed and the archive link exists; a retry hits "archived project already exists".'
          family: partial-failure-recovery
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: MVP-gate --force waiver has no processmanual gatesig row
          detail: 'The project close retro waiver is cataloged (gatesig.go:128) but the new "--force: closing with the MVP incomplete" acknowledgement is not, so transcript friction analysis can''t attribute it.'
          family: waiver-catalog-coverage
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: Durable plan's Core-concepts signatures and Task checkboxes don't match the delivered code
          detail: Plan names landingTickTransform(ref, pr) and tickLanded(ctx, ref, pr). Code has landingTickTransform(repoName, issueID, pr, auto), tickLandedProjects, tickOnLanding and publishProjectEdit. The plan's Tasks are all unchecked while the issue shows them done.
          family: plan-code-drift
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 145f753912eecde709fe32b0a2a25931d1a30dcb
        tracker_head: 137b12ef1a729000883c727750b7b2245d2cac1a
        branch_point: 72bf044e19b3be76bc22851008aaf9fbf2f30dc8
        since: "2026-10-15T01:00:50-07:00"
        until: "2026-10-15T04:17:16-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T01:00:50-07:00"
              end: "2026-10-15T04:17:16-07:00"
              began: "2026-10-15T01:00:50-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 24c8d79e-33f9-4ab1-a6b5-8cb02e00580a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 978bf57cd6ddea0de484fb0038c1f6da8c07099650cc1829a6c65688444ed05f
              started: "2026-10-12T14:18:17.059Z"
              spawned: false
              bytes: 12069433
              windows:
                - start: "2026-10-15T01:00:50-07:00"
                  end: "2026-10-15T04:17:16-07:00"
        cuts:
            - dispatched: "2026-10-15T03:46:53-07:00"
              finalized: "2026-10-15T03:55:51-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-15T05:10:34-07:00"
      dispatched: "2026-10-15T04:59:51-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:468-499 now shows close --dry-run/--verified, project tick, --no-mvp/--force, status mvp/done_when, and landing_tick; matches projectclose.go and projectlandtick.go.
          round: 2
        - id: BR-2
          disposition: addressed
          note: cmd/tl-lead/help.md:85-96 now describes the printed project tick command, the landing_tick opt-in, and an in-place close that moves only the link; matches commitLinkedProjectClose.
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: Code is fixed (tick after withLock on both paths, 2m ctx deadline reaches peerTrunk), but no test fails without it. Add a Lock-wrapper test asserting the peer origin has no tick commit when run() returns, on both durable and legacy paths.
          round: 2
        - id: BR-4
          disposition: not-addressed
          note: 'Verified: with the fix reverted, TestProjectClose* still passes, because the old code already rolled back a rename failure. Pin the ordering with a rename stub that asserts the live link is gone. Also join the error from _ = os.Symlink restore (projectclose.go:318) into the returned error.'
          round: 2
        - id: BR-5
          disposition: addressed
          note: gatesig.go:131-134 adds the no-mvp row; gatesig_test GateFlagNames/GateFlagsFor pin it. Pattern-vs-emitted-text pinning is raised separately below.
          round: 2
        - id: BR-6
          disposition: addressed
          note: Plan Tasks are checked, and the 2026-10-15 Revisions entry records landingTickTransform(repoName, issueID, pr, auto), tickLandedProjects, tickOnLanding, and publishProjectEdit.
          round: 2
      findings:
        - kind: defect
          id: BR-7
          severity: Minor
          title: The no-mvp waiver's AckPat/RefusalPat are not pinned against the text close actually emits
          detail: '2nd in family. Rule: every GateCatalog row''s AckPat and RefusalPat must be matched by a test against the real waiver and refusal output. No test runs --no-mvp or matches "--no-mvp (or --force): closing with the MVP incomplete"; no-retro is pinned at projectclose_test.go:42. Sweep every catalog row, not only this one.'
          family: waiver-catalog-coverage
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f1f1e8c7-fcc1-4370-97a8-00b3dcf00c1d
            repository: github.com/xianxu/ariadne
            issue_id: "000373"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/23dee52d3448d231
            reviewed_head: 0aa64f88819a74197e75d1b0f6f99a032006ab1b
            slot: ariadne:4
            requested_at: "2026-10-15T12:51:29.117694Z"
          phase: finished
          pid: 36147
          started_at: "2026-10-15T12:51:29.373872Z"
          finished_at: "2026-10-15T12:53:46.777512Z"
          session_id: 19080286-ab7d-4d91-b47c-d6dfa784910b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 0aa64f88819a74197e75d1b0f6f99a032006ab1b
        tracker_head: ce1222b518f3d9fdc3b92f9cb566951954abac1e
        branch_point: 7bfdff6ee3dfc9141dc97525f865c5316b156595
        since: "2026-10-15T01:00:50-07:00"
        until: "2026-10-15T05:14:56-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T01:00:50-07:00"
              end: "2026-10-15T05:14:56-07:00"
              began: "2026-10-15T01:00:50-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 24c8d79e-33f9-4ab1-a6b5-8cb02e00580a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 91ac944565fa777093986d6e4081d4dde06435b5756eb142068b70a995aeb057
              started: "2026-10-12T14:18:17.059Z"
              spawned: false
              bytes: 12335296
              windows:
                - start: "2026-10-15T01:00:50-07:00"
                  end: "2026-10-15T05:14:56-07:00"
        cuts:
            - dispatched: "2026-10-15T03:46:53-07:00"
              finalized: "2026-10-15T03:55:51-07:00"
              finalizing: false
            - dispatched: "2026-10-15T04:59:51-07:00"
              finalized: "2026-10-15T05:10:34-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-15T05:53:48-07:00"
      dispatched: "2026-10-15T05:51:29-07:00"
      agent: claude
      dispose:
        - id: BR-3
          disposition: addressed
          note: tickOnLanding now runs after f.withLock in both landing.go:609-617 and merge.go:778-783, with a 2m context.WithTimeout that reaches TrunkFile git calls; the tickOutsideLock test asserts origin is untouched at lock release on both merge paths.
          round: 3
        - id: BR-4
          disposition: addressed
          note: commitLinkedProjectClose removes the live link before the rename and restores it on failure; TestProjectCloseLinkedRecordRollsBack forces the rename to fail and checks the link is restored, no archive link is left, and a rerun succeeds; TestProjectCloseLinkedRecord asserts the live link is gone at the commit point.
          round: 3
        - id: BR-7
          disposition: addressed
          note: TestProjectCloseGatesMatchTheCatalog loops over every project close catalog row (no-retro, no-mvp) and checks RefusalPat against the real refusal error and AckPat, via the new assertGatesigAcknowledges, against the real waiver stderr. Rows owned by other commands predate this issue.
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: 0aa64f88819a74197e75d1b0f6f99a032006ab1b
      blocked: false
---

# Gate ledger — ariadne#373 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-15T03:55:51-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `docs-surface-not-swept` (kind: defect) README project section is stale: close example now fails without --verified; project tick, --dry-run, landing_tick undocumented
  README.md:468 shows `sdlc project close --slug example`, which now refuses (--verified required, projectclose.go:117). The new `project tick` verb, close --dry-run/--verified, status mvp/done_when output, landing_tick, and --force waiving the MVP gate are absent from README.
- **BR-2** [Important] `docs-surface-not-swept` (kind: defect) cmd/tl-lead/help.md contradicts the new linked-close and landing-tick behavior
  help.md:85-92 tells TLs to move the team directory to teams/<team>/history/projects/ and drop the symlink at project close, and that workers never write project.md / TLs tick rows by hand. The code now updates the real file in place and moves the link into workshop/history/projects (folder stays), and merge ticks rows (opt-in via landing_tick) or prints `sdlc project tick`. ARCH-PURPOSE: a consumer not swept.
- **BR-3** [Important] `optional-work-under-critical-lock` (kind: defect) Landing tick does unbounded peer-origin network IO while holding the merge lock
  tickOnLanding runs inside f.withLock (landing.go:625, merge.go:750) and uses peerTrunk with commandContext (no deadline), unlike the landing's boundedLandingGit. A hung peer fetch/push blocks other slots' merges for warn-only work. Move it after the locked closure and give it a timeout (ARCH-CONSTRAINTS).
- **BR-4** [Minor] `partial-failure-recovery` (kind: defect) commitLinkedProjectClose leaves both links if removing the live link fails; rerun then refuses
  projectclose.go:308: the real file is already closed and the archive link exists; a retry hits "archived project already exists".
- **BR-5** [Minor] `waiver-catalog-coverage` (kind: defect) MVP-gate --force waiver has no processmanual gatesig row
  The project close retro waiver is cataloged (gatesig.go:128) but the new "--force: closing with the MVP incomplete" acknowledgement is not, so transcript friction analysis can't attribute it.
- **BR-6** [Minor] `plan-code-drift` (kind: defect) Durable plan's Core-concepts signatures and Task checkboxes don't match the delivered code
  Plan names landingTickTransform(ref, pr) and tickLanded(ctx, ref, pr). Code has landingTickTransform(repoName, issueID, pr, auto), tickLandedProjects, tickOnLanding and publishProjectEdit. The plan's Tasks are all unchecked while the issue shows them done.

## Round 2 — 2026-10-15T05:10:34-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — README.md:468-499 now shows close --dry-run/--verified, project tick, --no-mvp/--force, status mvp/done_when, and landing_tick; matches projectclose.go and projectlandtick.go.
- BR-2 — addressed — cmd/tl-lead/help.md:85-96 now describes the printed project tick command, the landing_tick opt-in, and an in-place close that moves only the link; matches commitLinkedProjectClose.
- BR-3 — not-addressed — Code is fixed (tick after withLock on both paths, 2m ctx deadline reaches peerTrunk), but no test fails without it. Add a Lock-wrapper test asserting the peer origin has no tick commit when run() returns, on both durable and legacy paths.
- BR-4 — not-addressed — Verified: with the fix reverted, TestProjectClose* still passes, because the old code already rolled back a rename failure. Pin the ordering with a rename stub that asserts the live link is gone. Also join the error from _ = os.Symlink restore (projectclose.go:318) into the returned error.
- BR-5 — addressed — gatesig.go:131-134 adds the no-mvp row; gatesig_test GateFlagNames/GateFlagsFor pin it. Pattern-vs-emitted-text pinning is raised separately below.
- BR-6 — addressed — Plan Tasks are checked, and the 2026-10-15 Revisions entry records landingTickTransform(repoName, issueID, pr, auto), tickLandedProjects, tickOnLanding, and publishProjectEdit.

### Raised

- **BR-7** [Minor] `waiver-catalog-coverage` (kind: defect) The no-mvp waiver's AckPat/RefusalPat are not pinned against the text close actually emits
  2nd in family. Rule: every GateCatalog row's AckPat and RefusalPat must be matched by a test against the real waiver and refusal output. No test runs --no-mvp or matches "--no-mvp (or --force): closing with the MVP incomplete"; no-retro is pinned at projectclose_test.go:42. Sweep every catalog row, not only this one.

## Round 3 — 2026-10-15T05:53:48-07:00 (claude) — passed

### Disposed

- BR-3 — addressed — tickOnLanding now runs after f.withLock in both landing.go:609-617 and merge.go:778-783, with a 2m context.WithTimeout that reaches TrunkFile git calls; the tickOutsideLock test asserts origin is untouched at lock release on both merge paths.
- BR-4 — addressed — commitLinkedProjectClose removes the live link before the rename and restores it on failure; TestProjectCloseLinkedRecordRollsBack forces the rename to fail and checks the link is restored, no archive link is left, and a rerun succeeds; TestProjectCloseLinkedRecord asserts the live link is gone at the commit point.
- BR-7 — addressed — TestProjectCloseGatesMatchTheCatalog loops over every project close catalog row (no-retro, no-mvp) and checks RefusalPat against the real refusal error and AckPat, via the new assertGatesigAcknowledges, against the real waiver stderr. Rows owned by other commands predate this issue.

## Open findings

(none — every finding has been disposed)
