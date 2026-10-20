---
gate: boundary-review
issue: 473
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 6565c97c-7663-4161-a479-04a5d599c1d3
            repository: github.com/xianxu/ariadne
            issue_id: "000473"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/2c46c994bc92bfc1
            reviewed_head: 36f16ff1e7a6c5b775a86aaaea601918089d89a9
            slot: ariadne:2
            requested_at: "2026-10-20T08:49:20.730538Z"
          phase: finished
          pid: 82816
          started_at: "2026-10-20T08:49:20.876995Z"
          finished_at: "2026-10-20T08:57:31.108352Z"
          session_id: 6f9f2218-be5e-4d86-9158-14cdee8133d6
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 36f16ff1e7a6c5b775a86aaaea601918089d89a9
        tracker_head: 97ccd9410a99b8cadb917699d8779fb664a986de
        branch_point: 0956e5e1c71bd6734f195929fc151e0c752dbc14
        since: "2026-10-19T20:26:49-07:00"
        until: "2026-10-20T01:48:42-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T20:26:49-07:00"
              end: "2026-10-20T01:48:42-07:00"
              began: "2026-10-19T20:26:49-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5a5aa7ef6697f8ce115263d087981e41bd7603b84db90304dbc47ea946847d1b
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 3148632
              windows:
                - start: "2026-10-19T20:26:49-07:00"
                  end: "2026-10-20T01:48:42-07:00"
        queue_waits:
            - attempt_id: bcd9b7407517e24367c744af82d5db34
              slot: ariadne:2
              record_digest: 38cc6a62061e1a988f1d7f0f855462befe0fbd9fd08ff6e8d5e7da15ab214da4
              start: "2026-10-20T03:26:49Z"
              end: "2026-10-20T03:54:30.635003Z"
              queue_seconds: 1661.635003
              raw_seconds: 1669.213411
              exact: true
            - attempt_id: 1e519eee13691a7f9ae238f3a34af108
              slot: ariadne:2
              record_digest: ec17427e86246fbb78fe51b1c02c6c1a3c11100644e7ee265a3bbb81ecc632fc
              start: "2026-10-20T04:15:40.4019Z"
              end: "2026-10-20T04:37:58.530161Z"
              queue_seconds: 1338.128261
              raw_seconds: 1338.128261
              exact: true
            - attempt_id: ec45d2b999a4f613e1087a285f7a4e97
              slot: ariadne:2
              record_digest: 44c532f2cb7fa09e8ca756cfd526f624744a71e1024a397fc4dcd2c47bf96ea1
              start: "2026-10-20T08:21:56.389306Z"
              end: "2026-10-20T08:36:54.647565Z"
              queue_seconds: 898.258259
              raw_seconds: 898.258259
              exact: true
        protected:
            - issue: "479"
              slot: ariadne:2
              start: "2026-10-19T21:12:26-07:00"
              end: "2026-10-20T04:15:25.377Z"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 27.769466666666666
      "n": 1
      timestamp: "2026-10-20T01:57:31-07:00"
      dispatched: "2026-10-20T01:49:20-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: A root Go package claims every changed path, so unclassified files never force the full suite
          detail: 'packageIn (graph.go:30) walks up to Dir "." and Plan (plan.go:44) marks the path claimed, as pinned by plan_test.go:125 where b/x.go maps to the root package. In a module whose root is a package, scripts/ and tests/lib/ changes select the root package and its dependents instead of Full, against the Spec''s unclassified-to-full-suite valve. Fix: limit ancestor claims (own dir plus testdata/, never the root for subdirectories) and add a root-package plus non-Go path table row expecting Full.'
          family: unclassified-path-fails-closed
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: An empty change set (rev at or behind origin/main) posts a "docs-only" affected success without running anything
          detail: In runAffectedCheck (localcheck_affected.go:33,43), publishGateHasCodeSurface of an empty path list is false, so a branch tip equal to main records sdlc-local/affected success "nothing affected (docs-only)" on a possibly-code commit. That is fabricated evidence the M2 stand-in and ancestor walk will trust. Refuse (or post nothing) when paths is empty, and add a test with sha equal to main.
          family: no-evidence-recorded-as-pass
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: README update appears missing for local-check run --name affected and .sdlc/affected.json
          detail: README.md:557-622 documents local-check run and the close suite but not the new name or config file. Plan Task 8 lists help text and atlas only, so README would be missed in M2 as well. Add README.md to Task 8 (plan revision) or update it now.
          family: docs-gate-new-surface
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: --all combined with --name affected silently drops --all
          detail: localcheck_cmd.go:148 skips selectLocalChecks when only affected is named, so the existing all-versus-names conflict error no longer fires.
          family: cli-flag-conflict
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: affected.ParseConfig accepts duplicate JSON keys (last one wins)
          detail: ARCH-EVOLVE asks that duplicate keys stay strict; encoding/json silently takes the last value.
          family: strict-persisted-format
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: The affected check carries KnownFlakes, but the flake step and success guard still key on "suite" until M2
          detail: This is inert in M1 (localcheck_run.go:136,184 and localcheck_republish.go:198). The docs-only record also saves no attempt, so it is not republishable. Note both in the Log so D4 is not assumed live.
          family: deferred-scope-visibility
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: cf024732-aa00-4e83-ad77-7c77f75d9d74
            repository: github.com/xianxu/ariadne
            issue_id: "000473"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/d546c3d9773d17bf
            reviewed_head: 387922e63d8c9d31022c1481dc6b0aa83a1fa270
            slot: ariadne:2
            requested_at: "2026-10-20T09:03:23.984159Z"
          phase: finished
          pid: 58208
          started_at: "2026-10-20T09:03:24.133797Z"
          finished_at: "2026-10-20T09:05:58.238916Z"
          session_id: 819a2b26-5aab-4d8f-a094-0d8e187c33ff
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 387922e63d8c9d31022c1481dc6b0aa83a1fa270
        tracker_head: 2a41686cbdf2a2a3a81fb907bb66985c222768af
        branch_point: 0956e5e1c71bd6734f195929fc151e0c752dbc14
        since: "2026-10-19T20:26:49-07:00"
        until: "2026-10-20T02:02:42-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T20:26:49-07:00"
              end: "2026-10-20T02:02:42-07:00"
              began: "2026-10-19T20:26:49-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6838ed8a96087db019b6c791859639aeb38f3eb8787105e13ffb1ff4a28ba1a6
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 3631411
              windows:
                - start: "2026-10-19T20:26:49-07:00"
                  end: "2026-10-20T02:02:42-07:00"
        queue_waits:
            - attempt_id: bcd9b7407517e24367c744af82d5db34
              slot: ariadne:2
              record_digest: 38cc6a62061e1a988f1d7f0f855462befe0fbd9fd08ff6e8d5e7da15ab214da4
              start: "2026-10-20T03:26:49Z"
              end: "2026-10-20T03:54:30.635003Z"
              queue_seconds: 1661.635003
              raw_seconds: 1669.213411
              exact: true
            - attempt_id: 1e519eee13691a7f9ae238f3a34af108
              slot: ariadne:2
              record_digest: ec17427e86246fbb78fe51b1c02c6c1a3c11100644e7ee265a3bbb81ecc632fc
              start: "2026-10-20T04:15:40.4019Z"
              end: "2026-10-20T04:37:58.530161Z"
              queue_seconds: 1338.128261
              raw_seconds: 1338.128261
              exact: true
            - attempt_id: ec45d2b999a4f613e1087a285f7a4e97
              slot: ariadne:2
              record_digest: 44c532f2cb7fa09e8ca756cfd526f624744a71e1024a397fc4dcd2c47bf96ea1
              start: "2026-10-20T08:21:56.389306Z"
              end: "2026-10-20T08:36:54.647565Z"
              queue_seconds: 898.258259
              raw_seconds: 898.258259
              exact: true
        protected:
            - issue: "479"
              slot: ariadne:2
              start: "2026-10-19T21:12:26-07:00"
              end: "2026-10-20T04:15:25.377Z"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 27.769466666666666
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T01:49:20-07:00"
              finalized: "2026-10-20T01:57:31-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-20T02:05:59-07:00"
      dispatched: "2026-10-20T02:03:23-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: packageIn is now exact (own dir, testdata/, embeds); TestPlanRootPackageDoesNotClaimSubdirectories and TestPackageOfIsExact go red when the old ancestor walk is restored.
          round: 2
        - id: BR-2
          disposition: addressed
          note: runAffectedCheck refuses on empty paths before any post; TestAffectedNoChangeRefuses goes red with the guard disabled.
          round: 2
        - id: BR-3
          disposition: addressed
          note: README.md:569-582 documents --name affected, .sdlc/affected.json, both env vars and all four outcomes; it links the atlas anchor, which exists (ci-merge-check.md:307).
          round: 2
        - id: BR-4
          disposition: not-addressed
          note: localcheck_cmd.go:145 restores the refusal, but no test runs --all with --name affected, so nothing fails without the fix.
          round: 2
        - id: BR-5
          disposition: addressed
          note: ParseConfig calls jsonstrict.RejectDuplicateKeys; config_test.go covers top-level and nested group duplicates.
          round: 2
        - id: BR-6
          disposition: addressed
          note: The issue Log's M1 review round 1 entry states that D4 is inert until M2 and that a docs-only record cannot be republished.
          round: 2
      findings:
        - kind: defect
          id: BR-7
          severity: Important
          title: packageIn stops at the first owner, so a file embedded by one package but inside another package's dir skips the embedder
          detail: 'graph.go:31-46 returns on the own-dir or testdata match before looking at Embeds, and Embeds is a single-valued map (graph.go:13). Verified with go list: a root package embedding sub/data.txt, where sub is its own package, is legal, yet Plan selects only sub. Rule for the family: a path selects the union of every claimant (own-dir package, testdata owner, every embedder, every matching group), and is Full only when that union is empty. Fix: make Embeds map[string][]string, have packageIn return a set, and add table rows for both shapes.'
          family: unclassified-path-fails-closed
          round: 2
        - kind: defect
          id: BR-8
          severity: Minor
          title: The plan still describes nearest-ancestor package ownership; no Revisions entry for the exact-ownership change
          detail: Core concepts rule 3 says "selects its nearest package", Task 2 says PackageOf is the "deepest dir that is a package", and the Graph struct has no Embeds field. The code and atlas now use exact ownership (own dir, testdata/, embeds).
          family: plan-tracks-code
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 48a54735-a380-40bb-b9d2-57d92641e825
            repository: github.com/xianxu/ariadne
            issue_id: "000473"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/acd521bd8eac615f
            reviewed_head: f4a4a688e1499c1f34e910719f33929d962b619b
            slot: ariadne:2
            requested_at: "2026-10-20T09:08:18.759877Z"
          phase: finished
          pid: 68795
          started_at: "2026-10-20T09:08:18.845901Z"
          finished_at: "2026-10-20T09:13:59.937446Z"
          session_id: cdbb0124-abc4-46fd-b39f-cf762c1bbdf6
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: f4a4a688e1499c1f34e910719f33929d962b619b
        tracker_head: 2a41686cbdf2a2a3a81fb907bb66985c222768af
        branch_point: 0956e5e1c71bd6734f195929fc151e0c752dbc14
        since: "2026-10-19T20:26:49-07:00"
        until: "2026-10-20T02:07:22-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T20:26:49-07:00"
              end: "2026-10-20T02:07:22-07:00"
              began: "2026-10-19T20:26:49-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: a8aa7d5a3ef374c068d21e44b188b7f6dfb8c555a55675e2d9053acdf2347ccd
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 3771238
              windows:
                - start: "2026-10-19T20:26:49-07:00"
                  end: "2026-10-20T02:07:22-07:00"
        queue_waits:
            - attempt_id: bcd9b7407517e24367c744af82d5db34
              slot: ariadne:2
              record_digest: 38cc6a62061e1a988f1d7f0f855462befe0fbd9fd08ff6e8d5e7da15ab214da4
              start: "2026-10-20T03:26:49Z"
              end: "2026-10-20T03:54:30.635003Z"
              queue_seconds: 1661.635003
              raw_seconds: 1669.213411
              exact: true
            - attempt_id: 1e519eee13691a7f9ae238f3a34af108
              slot: ariadne:2
              record_digest: ec17427e86246fbb78fe51b1c02c6c1a3c11100644e7ee265a3bbb81ecc632fc
              start: "2026-10-20T04:15:40.4019Z"
              end: "2026-10-20T04:37:58.530161Z"
              queue_seconds: 1338.128261
              raw_seconds: 1338.128261
              exact: true
            - attempt_id: ec45d2b999a4f613e1087a285f7a4e97
              slot: ariadne:2
              record_digest: 44c532f2cb7fa09e8ca756cfd526f624744a71e1024a397fc4dcd2c47bf96ea1
              start: "2026-10-20T08:21:56.389306Z"
              end: "2026-10-20T08:36:54.647565Z"
              queue_seconds: 898.258259
              raw_seconds: 898.258259
              exact: true
        protected:
            - issue: "479"
              slot: ariadne:2
              start: "2026-10-19T21:12:26-07:00"
              end: "2026-10-20T04:15:25.377Z"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 27.769466666666666
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T01:49:20-07:00"
              finalized: "2026-10-20T01:57:31-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T02:03:23-07:00"
              finalized: "2026-10-20T02:05:59-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-20T02:14:02-07:00"
      dispatched: "2026-10-20T02:08:18-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-4
          disposition: addressed
          note: splitAffectedName refuses --all with any --name (localcheck_cmd.go:98-104); disabling the guard makes TestSplitAffectedName fail.
          round: 3
        - id: BR-7
          disposition: addressed
          note: Embeds is map[string][]string and owners returns the union of own-dir, testdata ancestors and every embedder; reverting to first-match fails TestLoadGoGraphConformance (live go list), TestPlan BR-7 row and TestOwnersAreExactAndAUnion.
          round: 3
        - id: BR-8
          disposition: addressed
          note: Plan Revisions entry dated 2026-10-20 (plan.md:427-436) explicitly replaces "nearest package" (rule 3) and Task 2 PackageOf with Graph.Owners union semantics, matching graph.go:24-50.
          round: 3
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      reviewed: f4a4a688e1499c1f34e910719f33929d962b619b
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 62303091-c0a4-40ac-99d5-ad201c918fd0
            repository: github.com/xianxu/ariadne
            issue_id: "000473"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/89369333f40a517f
            reviewed_head: d6309691af55cb39765a55623ed01f5ab31212e2
            slot: ariadne:2
            requested_at: "2026-10-20T09:22:21.514451Z"
          phase: finished
          pid: 82302
          started_at: "2026-10-20T09:22:21.654275Z"
          finished_at: "2026-10-20T09:29:43.137115Z"
          session_id: a890961a-646b-4097-807b-a6708375d274
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: d6309691af55cb39765a55623ed01f5ab31212e2
        tracker_head: 2a41686cbdf2a2a3a81fb907bb66985c222768af
        branch_point: 0956e5e1c71bd6734f195929fc151e0c752dbc14
        since: "2026-10-19T20:26:49-07:00"
        until: "2026-10-20T02:21:33-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T20:26:49-07:00"
              end: "2026-10-20T02:21:33-07:00"
              began: "2026-10-19T20:26:49-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 8409afa1d5ecb412c9b5c215a85085e56fd7e44ed0a3d1f605e17a4186df7fd6
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 4492293
              windows:
                - start: "2026-10-19T20:26:49-07:00"
                  end: "2026-10-20T02:21:33-07:00"
        queue_waits:
            - attempt_id: bcd9b7407517e24367c744af82d5db34
              slot: ariadne:2
              record_digest: 38cc6a62061e1a988f1d7f0f855462befe0fbd9fd08ff6e8d5e7da15ab214da4
              start: "2026-10-20T03:26:49Z"
              end: "2026-10-20T03:54:30.635003Z"
              queue_seconds: 1661.635003
              raw_seconds: 1669.213411
              exact: true
            - attempt_id: 1e519eee13691a7f9ae238f3a34af108
              slot: ariadne:2
              record_digest: ec17427e86246fbb78fe51b1c02c6c1a3c11100644e7ee265a3bbb81ecc632fc
              start: "2026-10-20T04:15:40.4019Z"
              end: "2026-10-20T04:37:58.530161Z"
              queue_seconds: 1338.128261
              raw_seconds: 1338.128261
              exact: true
            - attempt_id: ec45d2b999a4f613e1087a285f7a4e97
              slot: ariadne:2
              record_digest: 44c532f2cb7fa09e8ca756cfd526f624744a71e1024a397fc4dcd2c47bf96ea1
              start: "2026-10-20T08:21:56.389306Z"
              end: "2026-10-20T08:36:54.647565Z"
              queue_seconds: 898.258259
              raw_seconds: 898.258259
              exact: true
        protected:
            - issue: "479"
              slot: ariadne:2
              start: "2026-10-19T21:12:26-07:00"
              end: "2026-10-20T04:15:25.377Z"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 27.769466666666666
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T01:49:20-07:00"
              finalized: "2026-10-20T01:57:31-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T02:03:23-07:00"
              finalized: "2026-10-20T02:05:59-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T02:08:18-07:00"
              finalized: "2026-10-20T02:14:02-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-20T02:29:44-07:00"
      dispatched: "2026-10-20T02:22:21-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-9
          severity: Important
          title: 'local-check republish --name affected always refuses for a saved known-flake pass: declaredCheck cannot resolve the reserved name'
          detail: 'runLocalCheckRepublish (localcheck_cmd.go:255-258) calls declaredCheck for KnownFlakePassed/Accept attempts, but declaredCheck (localcheck_cmd.go:290-307) only searches local-checks.json plus the default suite, where affected is reserved (localchecks.go:121). M2 newly lets affected runs reach KnownFlakePassed (localcheck_run.go:184), and a failed post then tells the operator to run this exact command (localcheck_run.go:230), so the only recovery is a full rerun. Rule (ARCH-PURPOSE sweep): every consumer that resolves a check name back to a localCheck must resolve affected from affected.json with the suite''s policy. Fix declaredCheck accordingly and add a republish regression test.'
          family: name-keyed-check-resolution
          round: 4
        - kind: defect
          id: BR-10
          severity: Important
          title: Plan D2/Task 8 still claim publishgate.go and merge.md wording changes that the code deliberately did not make
          detail: '2nd in family plan-tracks-code. The rule: any departure from a plan''s Decisions or Files list ships with a ## Revisions entry in the same commit. The publish gate''s --full-suite instruction stays on suite (the atlas explains why: a conflict resolution changes main''s code), and merge.md is untouched. D1-D6 and Tasks 6-8 swept against the diff show no other departures. Add the Revisions entry and amend D2.'
          family: plan-tracks-code
          round: 4
        - kind: defect
          id: BR-11
          severity: Minor
          title: A malformed affected.json on main makes every car's CI gate error instead of degrading to requiring suite
          detail: '3rd in family. Rule covering all three: an affected-suite input that cannot be trusted degrades to requiring the full suite, never to a hard stop or an empty selection. carCIGateFacts (landingcar.go:121) returns an error, which wedges the train. Dropping the stand-in with a warning matches the spec''s "affected unavailable; run the suite".'
          family: unclassified-path-fails-closed
          round: 4
      boundary: M2
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 150b7f6e-2da5-482a-bd39-20dfe5171031
            repository: github.com/xianxu/ariadne
            issue_id: "000473"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/a5c7ac163e197ce0
            reviewed_head: 22460760f3d5f188c542a29f95358a8ef7899a8a
            slot: ariadne:2
            requested_at: "2026-10-20T09:32:28.682985Z"
          phase: finished
          pid: 28315
          started_at: "2026-10-20T09:32:28.878294Z"
          finished_at: "2026-10-20T09:45:03.672704Z"
          session_id: 496da64a-c039-4d5d-b570-4cd323b4fdb3
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 22460760f3d5f188c542a29f95358a8ef7899a8a
        tracker_head: 2a41686cbdf2a2a3a81fb907bb66985c222768af
        branch_point: 0956e5e1c71bd6734f195929fc151e0c752dbc14
        since: "2026-10-19T20:26:49-07:00"
        until: "2026-10-20T02:31:31-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T20:26:49-07:00"
              end: "2026-10-20T02:31:31-07:00"
              began: "2026-10-19T20:26:49-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ff8f082f178c8d8c1eefd7235b9ff2cd2f752c69045bc83dfd26a21294f05368
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 4652541
              windows:
                - start: "2026-10-19T20:26:49-07:00"
                  end: "2026-10-20T02:31:31-07:00"
        queue_waits:
            - attempt_id: bcd9b7407517e24367c744af82d5db34
              slot: ariadne:2
              record_digest: 38cc6a62061e1a988f1d7f0f855462befe0fbd9fd08ff6e8d5e7da15ab214da4
              start: "2026-10-20T03:26:49Z"
              end: "2026-10-20T03:54:30.635003Z"
              queue_seconds: 1661.635003
              raw_seconds: 1669.213411
              exact: true
            - attempt_id: 1e519eee13691a7f9ae238f3a34af108
              slot: ariadne:2
              record_digest: ec17427e86246fbb78fe51b1c02c6c1a3c11100644e7ee265a3bbb81ecc632fc
              start: "2026-10-20T04:15:40.4019Z"
              end: "2026-10-20T04:37:58.530161Z"
              queue_seconds: 1338.128261
              raw_seconds: 1338.128261
              exact: true
            - attempt_id: ec45d2b999a4f613e1087a285f7a4e97
              slot: ariadne:2
              record_digest: 44c532f2cb7fa09e8ca756cfd526f624744a71e1024a397fc4dcd2c47bf96ea1
              start: "2026-10-20T08:21:56.389306Z"
              end: "2026-10-20T08:36:54.647565Z"
              queue_seconds: 898.258259
              raw_seconds: 898.258259
              exact: true
        protected:
            - issue: "479"
              slot: ariadne:2
              start: "2026-10-19T21:12:26-07:00"
              end: "2026-10-20T04:15:25.377Z"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 27.769466666666666
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T01:49:20-07:00"
              finalized: "2026-10-20T01:57:31-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T02:03:23-07:00"
              finalized: "2026-10-20T02:05:59-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T02:08:18-07:00"
              finalized: "2026-10-20T02:14:02-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-20T02:22:21-07:00"
              finalized: "2026-10-20T02:29:44-07:00"
              finalizing: false
      "n": 5
      timestamp: "2026-10-20T02:45:04-07:00"
      dispatched: "2026-10-20T02:32:28-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-9
          disposition: addressed
          note: declaredCheck resolves affected via affectedCheckDecl (localcheck_cmd.go:301-314); TestDeclaredCheckResolvesAffected goes red with the branch disabled (verified by revert).
          round: 5
        - id: BR-10
          disposition: addressed
          note: Plan Revisions 2026-10-20 (M2 Task 8, BR-9/BR-10) amends D2 and Task 8 Files; publishgate.go and merge.md are unchanged in the diff, matching it.
          round: 5
        - id: BR-11
          disposition: addressed
          note: carCIGateFacts drops the stand-in on readAffectedConfig error (landingcar.go:126); TestCarWithBrokenAffectedConfigNeedsSuite goes red when the error is returned (verified by revert).
          round: 5
      findings:
        - kind: defect
          id: BR-12
          severity: Minor
          title: test-stats emits a zero-run suite row for a repo whose window holds only affected attempts
          detail: aggregateCheckAttempts groups by repo over isFlakeEvidence, which now includes affected, before filtering rows by check (teststats.go:128,157), so the suite aggregate returns Runs 0 for such a repo and runTestStats appends it unconditionally (teststats_cmd.go:150-152). Filter suite rows with Runs > 0 as the affected row is, and add a runTestStats test with affected-only attempts.
          family: report-row-requires-runs
          round: 5
      boundary: M2
      recipe: milestone-review
      isolation: checkout
      reviewed: 22460760f3d5f188c542a29f95358a8ef7899a8a
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 05bc1e1c-46c6-4788-b8aa-2d9f40b49b0e
            repository: github.com/xianxu/ariadne
            issue_id: "000473"
            recipe: milestone-review
            round: 6
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/e99f14f8d5df9597
            reviewed_head: a3b1a8cb593559d6254cf0df09b0833e4fafcb46
            slot: ariadne:2
            requested_at: "2026-10-20T09:46:35.868741Z"
          phase: finished
          pid: 81339
          started_at: "2026-10-20T09:46:35.988473Z"
          finished_at: "2026-10-20T10:12:01.910331Z"
          session_id: 2ee48323-5fad-4b4e-a672-b1a99ccf4fce
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: a3b1a8cb593559d6254cf0df09b0833e4fafcb46
        tracker_head: 0c158439dfa9bdb30b733b79c8c0bfb6aa716bcd
        branch_point: 0956e5e1c71bd6734f195929fc151e0c752dbc14
        since: "2026-10-19T20:26:49-07:00"
        until: "2026-10-20T02:45:54-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T20:26:49-07:00"
              end: "2026-10-20T02:45:54-07:00"
              began: "2026-10-19T20:26:49-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f5282d472b8b6f5111567dab26b9b61964c510a755e10cc2d5b21acb54f2393e
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 4767502
              windows:
                - start: "2026-10-19T20:26:49-07:00"
                  end: "2026-10-20T02:45:54-07:00"
        queue_waits:
            - attempt_id: bcd9b7407517e24367c744af82d5db34
              slot: ariadne:2
              record_digest: 38cc6a62061e1a988f1d7f0f855462befe0fbd9fd08ff6e8d5e7da15ab214da4
              start: "2026-10-20T03:26:49Z"
              end: "2026-10-20T03:54:30.635003Z"
              queue_seconds: 1661.635003
              raw_seconds: 1669.213411
              exact: true
            - attempt_id: 1e519eee13691a7f9ae238f3a34af108
              slot: ariadne:2
              record_digest: ec17427e86246fbb78fe51b1c02c6c1a3c11100644e7ee265a3bbb81ecc632fc
              start: "2026-10-20T04:15:40.4019Z"
              end: "2026-10-20T04:37:58.530161Z"
              queue_seconds: 1338.128261
              raw_seconds: 1338.128261
              exact: true
            - attempt_id: ec45d2b999a4f613e1087a285f7a4e97
              slot: ariadne:2
              record_digest: 44c532f2cb7fa09e8ca756cfd526f624744a71e1024a397fc4dcd2c47bf96ea1
              start: "2026-10-20T08:21:56.389306Z"
              end: "2026-10-20T08:36:54.647565Z"
              queue_seconds: 898.258259
              raw_seconds: 898.258259
              exact: true
        protected:
            - issue: "479"
              slot: ariadne:2
              start: "2026-10-19T21:12:26-07:00"
              end: "2026-10-20T04:15:25.377Z"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 27.769466666666666
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T01:49:20-07:00"
              finalized: "2026-10-20T01:57:31-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T02:03:23-07:00"
              finalized: "2026-10-20T02:05:59-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T02:08:18-07:00"
              finalized: "2026-10-20T02:14:02-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-20T02:22:21-07:00"
              finalized: "2026-10-20T02:29:44-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-20T02:32:28-07:00"
              finalized: "2026-10-20T02:45:04-07:00"
              finalizing: true
      "n": 6
      timestamp: "2026-10-20T03:12:03-07:00"
      dispatched: "2026-10-20T02:46:35-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-12
          disposition: addressed
          note: testStatsRows filters both rows on Runs>0; reverting the suite filter turns TestTestStatsRowsNeedRuns red (verified).
          round: 6
      findings:
        - kind: defect
          id: BR-13
          severity: Minor
          title: Docs rule runs before ownership, so an embedded or testdata .md outside cmd/ records a docs-only affected pass
          detail: 'plan.go:30 skips isDoc paths before owners() and runAffectedCheck skips the graph when publishGateHasCodeSurface is false; churn.IsEmbedded only exempts cmd/. 4th in family: rule is "a path may be ignored only when no package owns it and no group matches it" - apply isDoc only to unowned paths and load the graph for any nonempty change set.'
          family: unclassified-path-fails-closed
          round: 6
        - kind: defect
          id: BR-14
          severity: Minor
          title: Plan ARCH-ORDER claims the affected record names its merge-base; the status carries only the scope
          detail: 'plan line 169 vs affectedScope/statusDescription (localcheck_affected.go:79, localchecks.go:315). 3rd in family: rule is "each plan claim about a code property is delivered or revised in Revisions" - sweep Core concepts and ARCH sections against code at close.'
          family: plan-tracks-code
          round: 6
        - kind: defect
          id: BR-15
          severity: Minor
          title: affected.json strict parse misses case-folded duplicate keys (command vs COMMAND), last one wins
          detail: 'jsonstrict.RejectDuplicateKeys compares keys case-sensitively but encoding/json matches field names case-insensitively, so DisallowUnknownFields accepts both. Rule: a strict decode rejects any key that is not an exact tag match.'
          family: strict-persisted-format
          round: 6
      recipe: milestone-review
      isolation: checkout
      reviewed: a3b1a8cb593559d6254cf0df09b0833e4fafcb46
      blocked: false
---

# Gate ledger — ariadne#473 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T01:57:31-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `unclassified-path-fails-closed` (kind: defect) A root Go package claims every changed path, so unclassified files never force the full suite
  packageIn (graph.go:30) walks up to Dir "." and Plan (plan.go:44) marks the path claimed, as pinned by plan_test.go:125 where b/x.go maps to the root package. In a module whose root is a package, scripts/ and tests/lib/ changes select the root package and its dependents instead of Full, against the Spec's unclassified-to-full-suite valve. Fix: limit ancestor claims (own dir plus testdata/, never the root for subdirectories) and add a root-package plus non-Go path table row expecting Full.
- **BR-2** [Important] `no-evidence-recorded-as-pass` (kind: defect) An empty change set (rev at or behind origin/main) posts a "docs-only" affected success without running anything
  In runAffectedCheck (localcheck_affected.go:33,43), publishGateHasCodeSurface of an empty path list is false, so a branch tip equal to main records sdlc-local/affected success "nothing affected (docs-only)" on a possibly-code commit. That is fabricated evidence the M2 stand-in and ancestor walk will trust. Refuse (or post nothing) when paths is empty, and add a test with sha equal to main.
- **BR-3** [Important] `docs-gate-new-surface` (kind: defect) README update appears missing for local-check run --name affected and .sdlc/affected.json
  README.md:557-622 documents local-check run and the close suite but not the new name or config file. Plan Task 8 lists help text and atlas only, so README would be missed in M2 as well. Add README.md to Task 8 (plan revision) or update it now.
- **BR-4** [Minor] `cli-flag-conflict` (kind: defect) --all combined with --name affected silently drops --all
  localcheck_cmd.go:148 skips selectLocalChecks when only affected is named, so the existing all-versus-names conflict error no longer fires.
- **BR-5** [Minor] `strict-persisted-format` (kind: defect) affected.ParseConfig accepts duplicate JSON keys (last one wins)
  ARCH-EVOLVE asks that duplicate keys stay strict; encoding/json silently takes the last value.
- **BR-6** [Minor] `deferred-scope-visibility` (kind: defect) The affected check carries KnownFlakes, but the flake step and success guard still key on "suite" until M2
  This is inert in M1 (localcheck_run.go:136,184 and localcheck_republish.go:198). The docs-only record also saves no attempt, so it is not republishable. Note both in the Log so D4 is not assumed live.

## Round 2 — 2026-10-20T02:05:59-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — packageIn is now exact (own dir, testdata/, embeds); TestPlanRootPackageDoesNotClaimSubdirectories and TestPackageOfIsExact go red when the old ancestor walk is restored.
- BR-2 — addressed — runAffectedCheck refuses on empty paths before any post; TestAffectedNoChangeRefuses goes red with the guard disabled.
- BR-3 — addressed — README.md:569-582 documents --name affected, .sdlc/affected.json, both env vars and all four outcomes; it links the atlas anchor, which exists (ci-merge-check.md:307).
- BR-4 — not-addressed — localcheck_cmd.go:145 restores the refusal, but no test runs --all with --name affected, so nothing fails without the fix.
- BR-5 — addressed — ParseConfig calls jsonstrict.RejectDuplicateKeys; config_test.go covers top-level and nested group duplicates.
- BR-6 — addressed — The issue Log's M1 review round 1 entry states that D4 is inert until M2 and that a docs-only record cannot be republished.

### Raised

- **BR-7** [Important] `unclassified-path-fails-closed` (kind: defect) packageIn stops at the first owner, so a file embedded by one package but inside another package's dir skips the embedder
  graph.go:31-46 returns on the own-dir or testdata match before looking at Embeds, and Embeds is a single-valued map (graph.go:13). Verified with go list: a root package embedding sub/data.txt, where sub is its own package, is legal, yet Plan selects only sub. Rule for the family: a path selects the union of every claimant (own-dir package, testdata owner, every embedder, every matching group), and is Full only when that union is empty. Fix: make Embeds map[string][]string, have packageIn return a set, and add table rows for both shapes.
- **BR-8** [Minor] `plan-tracks-code` (kind: defect) The plan still describes nearest-ancestor package ownership; no Revisions entry for the exact-ownership change
  Core concepts rule 3 says "selects its nearest package", Task 2 says PackageOf is the "deepest dir that is a package", and the Graph struct has no Embeds field. The code and atlas now use exact ownership (own dir, testdata/, embeds).

## Round 3 — 2026-10-20T02:14:02-07:00 (claude) — passed

### Disposed

- BR-4 — addressed — splitAffectedName refuses --all with any --name (localcheck_cmd.go:98-104); disabling the guard makes TestSplitAffectedName fail.
- BR-7 — addressed — Embeds is map[string][]string and owners returns the union of own-dir, testdata ancestors and every embedder; reverting to first-match fails TestLoadGoGraphConformance (live go list), TestPlan BR-7 row and TestOwnersAreExactAndAUnion.
- BR-8 — addressed — Plan Revisions entry dated 2026-10-20 (plan.md:427-436) explicitly replaces "nearest package" (rule 3) and Task 2 PackageOf with Graph.Owners union semantics, matching graph.go:24-50.

## Round 4 — 2026-10-20T02:29:44-07:00 (claude) — BLOCKED

### Raised

- **BR-9** [Important] `name-keyed-check-resolution` (kind: defect) local-check republish --name affected always refuses for a saved known-flake pass: declaredCheck cannot resolve the reserved name
  runLocalCheckRepublish (localcheck_cmd.go:255-258) calls declaredCheck for KnownFlakePassed/Accept attempts, but declaredCheck (localcheck_cmd.go:290-307) only searches local-checks.json plus the default suite, where affected is reserved (localchecks.go:121). M2 newly lets affected runs reach KnownFlakePassed (localcheck_run.go:184), and a failed post then tells the operator to run this exact command (localcheck_run.go:230), so the only recovery is a full rerun. Rule (ARCH-PURPOSE sweep): every consumer that resolves a check name back to a localCheck must resolve affected from affected.json with the suite's policy. Fix declaredCheck accordingly and add a republish regression test.
- **BR-10** [Important] `plan-tracks-code` (kind: defect) Plan D2/Task 8 still claim publishgate.go and merge.md wording changes that the code deliberately did not make
  2nd in family plan-tracks-code. The rule: any departure from a plan's Decisions or Files list ships with a ## Revisions entry in the same commit. The publish gate's --full-suite instruction stays on suite (the atlas explains why: a conflict resolution changes main's code), and merge.md is untouched. D1-D6 and Tasks 6-8 swept against the diff show no other departures. Add the Revisions entry and amend D2.
- **BR-11** [Minor] `unclassified-path-fails-closed` (kind: defect) A malformed affected.json on main makes every car's CI gate error instead of degrading to requiring suite
  3rd in family. Rule covering all three: an affected-suite input that cannot be trusted degrades to requiring the full suite, never to a hard stop or an empty selection. carCIGateFacts (landingcar.go:121) returns an error, which wedges the train. Dropping the stand-in with a warning matches the spec's "affected unavailable; run the suite".

## Round 5 — 2026-10-20T02:45:04-07:00 (claude) — passed

### Disposed

- BR-9 — addressed — declaredCheck resolves affected via affectedCheckDecl (localcheck_cmd.go:301-314); TestDeclaredCheckResolvesAffected goes red with the branch disabled (verified by revert).
- BR-10 — addressed — Plan Revisions 2026-10-20 (M2 Task 8, BR-9/BR-10) amends D2 and Task 8 Files; publishgate.go and merge.md are unchanged in the diff, matching it.
- BR-11 — addressed — carCIGateFacts drops the stand-in on readAffectedConfig error (landingcar.go:126); TestCarWithBrokenAffectedConfigNeedsSuite goes red when the error is returned (verified by revert).

### Raised

- **BR-12** [Minor] `report-row-requires-runs` (kind: defect) test-stats emits a zero-run suite row for a repo whose window holds only affected attempts
  aggregateCheckAttempts groups by repo over isFlakeEvidence, which now includes affected, before filtering rows by check (teststats.go:128,157), so the suite aggregate returns Runs 0 for such a repo and runTestStats appends it unconditionally (teststats_cmd.go:150-152). Filter suite rows with Runs > 0 as the affected row is, and add a runTestStats test with affected-only attempts.

## Round 6 — 2026-10-20T03:12:03-07:00 (claude) — passed

### Disposed

- BR-12 — addressed — testStatsRows filters both rows on Runs>0; reverting the suite filter turns TestTestStatsRowsNeedRuns red (verified).

### Raised

- **BR-13** [Minor] `unclassified-path-fails-closed` (kind: defect) Docs rule runs before ownership, so an embedded or testdata .md outside cmd/ records a docs-only affected pass
  plan.go:30 skips isDoc paths before owners() and runAffectedCheck skips the graph when publishGateHasCodeSurface is false; churn.IsEmbedded only exempts cmd/. 4th in family: rule is "a path may be ignored only when no package owns it and no group matches it" - apply isDoc only to unowned paths and load the graph for any nonempty change set.
- **BR-14** [Minor] `plan-tracks-code` (kind: defect) Plan ARCH-ORDER claims the affected record names its merge-base; the status carries only the scope
  plan line 169 vs affectedScope/statusDescription (localcheck_affected.go:79, localchecks.go:315). 3rd in family: rule is "each plan claim about a code property is delivered or revised in Revisions" - sweep Core concepts and ARCH sections against code at close.
- **BR-15** [Minor] `strict-persisted-format` (kind: defect) affected.json strict parse misses case-folded duplicate keys (command vs COMMAND), last one wins
  jsonstrict.RejectDuplicateKeys compares keys case-sensitively but encoding/json matches field names case-insensitively, so DisallowUnknownFields accepts both. Rule: a strict decode rejects any key that is not an exact tag match.

## Open findings

- **BR-13** [Minor] `unclassified-path-fails-closed` (kind: defect) Docs rule runs before ownership, so an embedded or testdata .md outside cmd/ records a docs-only affected pass
- **BR-14** [Minor] `plan-tracks-code` (kind: defect) Plan ARCH-ORDER claims the affected record names its merge-base; the status carries only the scope
- **BR-15** [Minor] `strict-persisted-format` (kind: defect) affected.json strict parse misses case-folded duplicate keys (command vs COMMAND), last one wins
