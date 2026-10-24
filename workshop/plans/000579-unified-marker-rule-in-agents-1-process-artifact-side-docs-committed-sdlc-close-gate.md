---
gate: boundary-review
issue: 579
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 0d6831e5-f20d-4f9f-af64-5e67053c39d2
            repository: github.com/xianxu/ariadne
            issue_id: "000579"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/d4b3cd1f8547cf54
            reviewed_head: 493d00a8e64923ccfa9443ad6e65e8a0f4352fa4
            slot: ariadne:5
            requested_at: "2026-10-24T01:04:16.612831Z"
          phase: finished
          pid: 58495
          started_at: "2026-10-24T01:04:16.709856Z"
          finished_at: "2026-10-24T01:26:52.535559Z"
          session_id: a4a68048-f2df-422c-9cfb-9bdc94ad333d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 493d00a8e64923ccfa9443ad6e65e8a0f4352fa4
        tracker_head: 1175ec40c5c3a33cd33a2a0f47f9196e2b77bab7
        branch_point: 56172a1009f553fa227d1085f8873c3d636da3ac
        since: "2026-10-23T16:38:06-07:00"
        until: "2026-10-23T18:03:49-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T16:38:06-07:00"
              end: "2026-10-23T18:03:49-07:00"
              began: "2026-10-23T16:38:06-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 2b4b362f441eb7688698abe5031b383ea9478e24e0b232376ec3ef90e052bb4c
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 12191995
              windows:
                - start: "2026-10-23T16:38:06-07:00"
                  end: "2026-10-23T18:03:49-07:00"
      "n": 1
      timestamp: "2026-10-23T18:26:53-07:00"
      dispatched: "2026-10-23T18:04:16-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: close side-link census runs under --no-side-links/--force and dies unbypassably on project discovery or peer git errors
          detail: 'close.go:744 computes issueSideLinkRefusals before consulting f.skip("side-links"), and any error (projectWorkspaceRoots, DiscoverInRoots walk, ls-tree HEAD in a peer project repo) hits die() with no bypass, even with --no-project. Close''s own project code (close.go:841-850) only warns on the same discovery error. Fix: skip the census when the gate is skipped, and degrade project discovery errors to a warning as close already does.'
          family: gate-bypass-must-skip-the-check
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: no tests for the project-file census branch or for runChangeCode's side-links flag and --force bypass
          detail: 'issueSideLinkRefusals'' cross-repo project path (linkguard.go:525-547) and the runChangeCode wiring (changecode.go:69-78: NoSideLinks, the force warning, exit code) are never exercised. The plan''s TestChangeCodeRefusesUncommittedSideLink and TestCloseRefusesProjectLinkToUncommittedFile are absent, and the Revisions do not record dropping them.'
          family: verb-wiring-untested
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: ls-files receives the link path as a glob pathspec; use :(literal)
          detail: 'linkguard.go:470: a target containing * or [ is pattern-matched, so its ignored-vs-uncommitted state can come out wrong.'
          family: git-pathspec-literal
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: gitLinkState's tree parameter is unused
          detail: linkguard.go:456; whether a path is in the tree is decided only by the injected inTree function.
          family: dead-parameter
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: republished issues run publishSideDocs twice (pre-effect loop and republishOwnedOnce)
          detail: issuepublish.go:410 and republish.go:189; the pre-effect check could keep its result, or only run for first publications.
          family: duplicate-census
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 65e16fe6-f9e6-48f7-821b-8bcf2ba78ae1
            repository: github.com/xianxu/ariadne
            issue_id: "000579"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/e2e4bdbb2822ab97
            reviewed_head: 712b54c08a7c4ae7178a4727b3ecc1a44bf492f6
            slot: ariadne:5
            requested_at: "2026-10-24T01:47:29.067738Z"
          phase: finished
          pid: 27781
          started_at: "2026-10-24T01:47:29.177282Z"
          finished_at: "2026-10-24T01:59:57.028931Z"
          session_id: 5f07c44f-4b4d-4fe1-b74f-b212a5fe5555
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 712b54c08a7c4ae7178a4727b3ecc1a44bf492f6
        tracker_head: aa2471597ac42861e8dab5a95d79b51f4df0766a
        branch_point: 56172a1009f553fa227d1085f8873c3d636da3ac
        since: "2026-10-23T16:38:06-07:00"
        until: "2026-10-23T18:46:22-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T16:38:06-07:00"
              end: "2026-10-23T18:46:22-07:00"
              began: "2026-10-23T16:38:06-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: e7d88e2176423c1e6f0a8a6690755d31da30c79ddb76ec652dadb7b5bc458387
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 12540560
              windows:
                - start: "2026-10-23T16:38:06-07:00"
                  end: "2026-10-23T18:46:22-07:00"
        cuts:
            - dispatched: "2026-10-23T18:04:16-07:00"
              finalized: "2026-10-23T18:26:53-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T18:59:57-07:00"
      dispatched: "2026-10-23T18:47:29-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: close.go:744 checks f.skip("side-links") before calling issueSideLinkRefusals; project discovery and peer errors are now warnings (linkguard.go:105-126); TestIssueSideLinkRefusalsJudgeProjects (unborn peer warns) and TestCloseRefusesIgnoredSideLink (skip acknowledgement) would both fail on the old code.
          round: 2
        - id: BR-2
          disposition: addressed
          note: TestChangeCodeRefusesUncommittedSideLink drives runChangeCode (exit 1, --force ack, --no-side-links, pass once committed); TestIssueSideLinkRefusalsJudgeProjects covers the peer-project branch; the folded test is recorded in the plan's Revisions. Both pass.
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: The :(literal) fix in ls-files/ls-tree is correct (checked empirically; ls-tree was already literal), but no test uses a target with a glob character, so the fix has no test that fails without it. Minor; non-blocking.
          round: 2
        - id: BR-4
          disposition: addressed
          note: The unused tree parameter is gone from gitLinkState (linkguard.go:35); a refactor that changes no behavior, and every caller compiles and its tests pass.
          round: 2
        - id: BR-5
          disposition: addressed
          note: judgePublishLinks (issuepublish.go:313) skips republications when no first publication shares the run; republishOwnedOnce's own pre-commit check remains the first. A refactor that changes no behavior; the publish tests pass.
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: 712b54c08a7c4ae7178a4727b3ecc1a44bf492f6
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 02561fd7-2540-4cae-93e4-fb870fc71527
            repository: github.com/xianxu/ariadne
            issue_id: "000579"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/b9878356f3581393
            reviewed_head: 5ead6cb66681b6be51fb18fc5d5c67dfa2f5e886
            slot: ariadne:5
            requested_at: "2026-10-24T02:01:35.549139Z"
          phase: finished
          pid: 16870
          started_at: "2026-10-24T02:01:35.699457Z"
          finished_at: "2026-10-24T02:02:35.811277Z"
          session_id: 14900503-65ea-48c9-8059-64c8c8ab590b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 5ead6cb66681b6be51fb18fc5d5c67dfa2f5e886
        tracker_head: ee100cde45192aeee027b9b225e43b28adf5b620
        branch_point: 56172a1009f553fa227d1085f8873c3d636da3ac
        since: "2026-10-23T16:38:06-07:00"
        until: "2026-10-23T19:01:03-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T16:38:06-07:00"
              end: "2026-10-23T19:01:03-07:00"
              began: "2026-10-23T16:38:06-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 584684ae6cf8ecbc72cb8ed0109240043d7de9ee049b4f5249f1c60599446614
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 12606543
              windows:
                - start: "2026-10-23T16:38:06-07:00"
                  end: "2026-10-23T19:01:03-07:00"
        cuts:
            - dispatched: "2026-10-23T18:04:16-07:00"
              finalized: "2026-10-23T18:26:53-07:00"
              finalizing: false
            - dispatched: "2026-10-23T18:47:29-07:00"
              finalized: "2026-10-23T18:59:57-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-23T19:02:36-07:00"
      dispatched: "2026-10-23T19:01:35-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-3
          disposition: addressed
          note: linkguard.go:49 and :63 pass ":(literal)"+p; TestGitLinkStateLiteralPath goes red (x*.md judged Ignored) with the literal prefix reverted, green with it.
          round: 3
      recipe: fix-review
      isolation: fenced
      reviewed: 5ead6cb66681b6be51fb18fc5d5c67dfa2f5e886
      blocked: false
---

# Gate ledger — ariadne#579 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T18:26:53-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `gate-bypass-must-skip-the-check` (kind: defect) close side-link census runs under --no-side-links/--force and dies unbypassably on project discovery or peer git errors
  close.go:744 computes issueSideLinkRefusals before consulting f.skip("side-links"), and any error (projectWorkspaceRoots, DiscoverInRoots walk, ls-tree HEAD in a peer project repo) hits die() with no bypass, even with --no-project. Close's own project code (close.go:841-850) only warns on the same discovery error. Fix: skip the census when the gate is skipped, and degrade project discovery errors to a warning as close already does.
- **BR-2** [Important] `verb-wiring-untested` (kind: defect) no tests for the project-file census branch or for runChangeCode's side-links flag and --force bypass
  issueSideLinkRefusals' cross-repo project path (linkguard.go:525-547) and the runChangeCode wiring (changecode.go:69-78: NoSideLinks, the force warning, exit code) are never exercised. The plan's TestChangeCodeRefusesUncommittedSideLink and TestCloseRefusesProjectLinkToUncommittedFile are absent, and the Revisions do not record dropping them.
- **BR-3** [Minor] `git-pathspec-literal` (kind: defect) ls-files receives the link path as a glob pathspec; use :(literal)
  linkguard.go:470: a target containing * or [ is pattern-matched, so its ignored-vs-uncommitted state can come out wrong.
- **BR-4** [Minor] `dead-parameter` (kind: defect) gitLinkState's tree parameter is unused
  linkguard.go:456; whether a path is in the tree is decided only by the injected inTree function.
- **BR-5** [Minor] `duplicate-census` (kind: defect) republished issues run publishSideDocs twice (pre-effect loop and republishOwnedOnce)
  issuepublish.go:410 and republish.go:189; the pre-effect check could keep its result, or only run for first publications.

## Round 2 — 2026-10-23T18:59:57-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — close.go:744 checks f.skip("side-links") before calling issueSideLinkRefusals; project discovery and peer errors are now warnings (linkguard.go:105-126); TestIssueSideLinkRefusalsJudgeProjects (unborn peer warns) and TestCloseRefusesIgnoredSideLink (skip acknowledgement) would both fail on the old code.
- BR-2 — addressed — TestChangeCodeRefusesUncommittedSideLink drives runChangeCode (exit 1, --force ack, --no-side-links, pass once committed); TestIssueSideLinkRefusalsJudgeProjects covers the peer-project branch; the folded test is recorded in the plan's Revisions. Both pass.
- BR-3 — not-addressed — The :(literal) fix in ls-files/ls-tree is correct (checked empirically; ls-tree was already literal), but no test uses a target with a glob character, so the fix has no test that fails without it. Minor; non-blocking.
- BR-4 — addressed — The unused tree parameter is gone from gitLinkState (linkguard.go:35); a refactor that changes no behavior, and every caller compiles and its tests pass.
- BR-5 — addressed — judgePublishLinks (issuepublish.go:313) skips republications when no first publication shares the run; republishOwnedOnce's own pre-commit check remains the first. A refactor that changes no behavior; the publish tests pass.

## Round 3 — 2026-10-23T19:02:36-07:00 (claude) — passed

### Disposed

- BR-3 — addressed — linkguard.go:49 and :63 pass ":(literal)"+p; TestGitLinkStateLiteralPath goes red (x*.md judged Ignored) with the literal prefix reverted, green with it.

## Open findings

(none — every finding has been disposed)
