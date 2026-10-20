---
gate: boundary-review
issue: 486
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: aa0528ba-ebf8-4857-8a39-fc201ee6876e
            repository: github.com/xianxu/ariadne
            issue_id: "000486"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/425b922304953e34
            reviewed_head: 4a966a8883623a8eefd5582498139a7dbcaf13b8
            slot: ariadne:4
            requested_at: "2026-10-20T08:19:39.26669Z"
          phase: finished
          pid: 74671
          started_at: "2026-10-20T08:19:39.35695Z"
          finished_at: "2026-10-20T08:21:15.75727Z"
          session_id: 432bde0a-908a-4f2c-815a-5b5be639d0a1
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 4a966a8883623a8eefd5582498139a7dbcaf13b8
        tracker_head: 642a42774c32e249af6f54db96d762f5411fd69a
        branch_point: 1cc7620f848ce10c0834d213257b8ea5c47d214d
        since: "2026-10-20T01:02:55-07:00"
        until: "2026-10-20T01:19:10-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T01:02:55-07:00"
              end: "2026-10-20T01:19:10-07:00"
              began: "2026-10-20T01:02:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 7780672e21e7d4ed7911cc4d94d1e527f7949df8ef22c8044bc95eb1a1dfe25a
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 94259729
              windows:
                - start: "2026-10-20T01:02:55-07:00"
                  end: "2026-10-20T01:19:10-07:00"
      "n": 1
      timestamp: "2026-10-20T01:21:16-07:00"
      dispatched: "2026-10-20T01:19:39-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: 'workflowFile''s anchor fallback misses pair''s case: Makefile.workflow is gitignored and absent from the train tree'
          detail: 'helpercapability.go:72-76 falls back only through os.Readlink of the tree''s own link. In pair (and every consumer that slot-dependencies provisions), Makefile.workflow is untracked (.gitignore:78), so the train worktree has no such file, Readlink gets ENOENT, and the function returns pair''s exact error "lstat .../Makefile.workflow: no such file or directory". A probe test confirmed it: workflowFile(treeWithoutFile, slotWithSymlink) returns that error. Fix: when root''s copy fails to resolve and anchor != "", return filepath.EvalSymlinks(filepath.Join(anchor, "Makefile.workflow")). Every instance of the family in this window: symlinkedWorkflowTrain (landingtrain_test.go:820-824) commits the symlink to git, so TestLandingTrainResolvesASymlinkedWorkflow and TestLandingTrainStoppedByEnvironmentReleasesItsCars both run on a shape no layered repo has. Make the fixture gitignore Makefile.workflow and create the symlink only in each car checkout, then confirm the e2e test fails without the fix.'
          family: fixture-diverges-from-real-repo-shape
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: landingtrain.go:229-232 comment still says the lock helper resolves from the train's tree
          detail: With WorkflowAnchor it now falls back to the checkout the train was built from. Update the comment, and the plan's claim that the test "fails without the fix with pair's exact error".
          family: stale-comment-after-behavior-change
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f22bf1ca-71f6-4fb9-a088-720d69b78afe
            repository: github.com/xianxu/ariadne
            issue_id: "000486"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/e1624071ff38efaf
            reviewed_head: dc5a9f192f259a627a04e183e8197fc405a0edbe
            slot: ariadne:4
            requested_at: "2026-10-20T08:29:52.507787Z"
          phase: finished
          pid: 85159
          started_at: "2026-10-20T08:29:52.743624Z"
          finished_at: "2026-10-20T08:36:54.14004Z"
          session_id: 146fb686-5cc2-4493-9948-20463cf91b58
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: dc5a9f192f259a627a04e183e8197fc405a0edbe
        tracker_head: abf41e8bb5a49add8a18f31b938f525996445085
        branch_point: 1cc7620f848ce10c0834d213257b8ea5c47d214d
        since: "2026-10-20T01:02:55-07:00"
        until: "2026-10-20T01:22:51-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T01:02:55-07:00"
              end: "2026-10-20T01:22:51-07:00"
              began: "2026-10-20T01:02:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 4fa0d4306b291f45658e5231986132943ecc231d3bebf3dcb9f761f3cb7dc8fc
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 94455759
              windows:
                - start: "2026-10-20T01:02:55-07:00"
                  end: "2026-10-20T01:22:51-07:00"
        cuts:
            - dispatched: "2026-10-20T01:19:39-07:00"
              finalized: "2026-10-20T01:21:16-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-20T01:36:54-07:00"
      dispatched: "2026-10-20T01:29:52-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: helpercapability.go:80 now falls back to the anchor's own Makefile.workflow. The fixture gitignores it and links it untracked in each car. With the fallback or the WorkflowAnchor line reverted, TestLandingTrainResolvesASymlinkedWorkflow fails with pair's exact lstat error, and with the fallback reverted the unit test fails too.
          round: 2
        - id: BR-2
          disposition: addressed
          note: landingtrain.go:229-232 now describes the fallback to the checkout the train was built from. The plan's fails-without-fix claim is now true; confirmed by reverting.
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: dc5a9f192f259a627a04e183e8197fc405a0edbe
      blocked: false
---

# Gate ledger — ariadne#486 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T01:21:16-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `fixture-diverges-from-real-repo-shape` (kind: defect) workflowFile's anchor fallback misses pair's case: Makefile.workflow is gitignored and absent from the train tree
  helpercapability.go:72-76 falls back only through os.Readlink of the tree's own link. In pair (and every consumer that slot-dependencies provisions), Makefile.workflow is untracked (.gitignore:78), so the train worktree has no such file, Readlink gets ENOENT, and the function returns pair's exact error "lstat .../Makefile.workflow: no such file or directory". A probe test confirmed it: workflowFile(treeWithoutFile, slotWithSymlink) returns that error. Fix: when root's copy fails to resolve and anchor != "", return filepath.EvalSymlinks(filepath.Join(anchor, "Makefile.workflow")). Every instance of the family in this window: symlinkedWorkflowTrain (landingtrain_test.go:820-824) commits the symlink to git, so TestLandingTrainResolvesASymlinkedWorkflow and TestLandingTrainStoppedByEnvironmentReleasesItsCars both run on a shape no layered repo has. Make the fixture gitignore Makefile.workflow and create the symlink only in each car checkout, then confirm the e2e test fails without the fix.
- **BR-2** [Minor] `stale-comment-after-behavior-change` (kind: defect) landingtrain.go:229-232 comment still says the lock helper resolves from the train's tree
  With WorkflowAnchor it now falls back to the checkout the train was built from. Update the comment, and the plan's claim that the test "fails without the fix with pair's exact error".

## Round 2 — 2026-10-20T01:36:54-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — helpercapability.go:80 now falls back to the anchor's own Makefile.workflow. The fixture gitignores it and links it untracked in each car. With the fallback or the WorkflowAnchor line reverted, TestLandingTrainResolvesASymlinkedWorkflow fails with pair's exact lstat error, and with the fallback reverted the unit test fails too.
- BR-2 — addressed — landingtrain.go:229-232 now describes the fallback to the checkout the train was built from. The plan's fails-without-fix claim is now true; confirmed by reverting.

## Open findings

(none — every finding has been disposed)
