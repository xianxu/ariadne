---
gate: boundary-review
issue: 565
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 081ccaa4-559f-458b-b02b-92bbb8ae6fe7
            repository: github.com/xianxu/ariadne
            issue_id: "000565"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/05cafbac3a313b39
            reviewed_head: f0cf2e42a200e4047a9d3c69150cfdb901fa0c8a
            slot: ariadne:5
            requested_at: "2026-10-23T18:46:25.899641Z"
          phase: finished
          pid: 22195
          started_at: "2026-10-23T18:46:26.07681Z"
          finished_at: "2026-10-23T18:49:45.028308Z"
          session_id: 9ba7866f-9262-44b0-8440-8bbff34449c0
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: f0cf2e42a200e4047a9d3c69150cfdb901fa0c8a
        tracker_head: 2d218ac151ee6f837a97b511bf3dc32b742a3667
        branch_point: 668e7637556b4748178be10e69ee969922c01e29
        since: "2026-10-23T10:46:54-07:00"
        until: "2026-10-23T11:09:30-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T10:46:54-07:00"
              end: "2026-10-23T11:09:30-07:00"
              began: "2026-10-23T10:46:54-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: abc7ad2ee30a8aae4555a09e8530eb2e0f32d69694108c29c57f4afb07b98360
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 8776755
              windows:
                - start: "2026-10-23T10:46:54-07:00"
                  end: "2026-10-23T11:09:30-07:00"
      "n": 1
      timestamp: "2026-10-23T11:49:46-07:00"
      dispatched: "2026-10-23T11:46:25-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Status/sweep readers take the owner's exclusive try-lock, so a concurrent re-entry refuses as "already running" naming the dead pid
          detail: 'Intents.observe (intent.go:257-271) try-locks LOCK_EX on every intent, and Sweep runs a git rev-parse subprocess while holding it (mergeintent.go:68-70); every sdlc merge start and landing status sweeps all intents. A Hold during that window gets IntentHeldError naming the dead runner''s pid ("wait for it or stop it"). A throwaway test against the pinned head reproduced the refusal. A fleet re-entering after a full shutdown is exactly when these collide. Fix: Hold retries for a bounded time (about 2 s) before refusing; Sweep works out branch existence before locking; add a regression test that holds an observe lock while Hold runs.'
          family: reader-lock-masquerades-as-owner
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: formatMergeIntent prints Root and pause owner/reason unquoted, against the plan's every-field-%q promise
          detail: Valid only requires an absolute, clean root, so a hand-edited intent root (or pause.json reason) can put newlines or control characters into landing status / sdlc state output (mergeintent.go:124, mergeintent.go:106-110).
          family: printed-untrusted-fields-unquoted
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: Hold silently overwrites an unreadable same-branch intent, dropping a newer version's fields
          detail: When readIntent fails (corrupt or IntentVersion 2), Hold rewrites the file without merging extras (intent.go:189-200), although the atlas says unreadable intents are never acted on.
          family: tolerant-writer-drops-unknown-fields
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: The local branch-head rev-parse lookup is repeated three times in mergeintent.go
          detail: holdMergeIntent, sweepMergeIntents and observeMergeIntents each build the same rev-parse --verify --quiet refs/heads/B call; one helper would cover all three.
          family: duplicated-git-lookup
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a9ee95a4-0d3a-447f-9f61-7f0b5998e917
            repository: github.com/xianxu/ariadne
            issue_id: "000565"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/ce0e68e871df4120
            reviewed_head: 1a8a6b06a228bf8ae0bfb55ca8467de46e32b010
            slot: ariadne:5
            requested_at: "2026-10-23T18:53:33.223165Z"
          phase: finished
          pid: 4312
          started_at: "2026-10-23T18:53:33.405341Z"
          finished_at: "2026-10-23T18:56:03.148402Z"
          session_id: 014f220f-5b96-4c2e-8786-93617c6ced6d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 1a8a6b06a228bf8ae0bfb55ca8467de46e32b010
        tracker_head: ae6583f1981f2f5a900dea171fda70d62e0b2bdd
        branch_point: 668e7637556b4748178be10e69ee969922c01e29
        since: "2026-10-23T10:46:54-07:00"
        until: "2026-10-23T11:52:59-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T10:46:54-07:00"
              end: "2026-10-23T11:52:59-07:00"
              began: "2026-10-23T10:46:54-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: fad2e9f0d165fac91f955b8b32ff76d968209a9c009482a4927bbccb84ce0726
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 9001680
              windows:
                - start: "2026-10-23T10:46:54-07:00"
                  end: "2026-10-23T11:52:59-07:00"
        cuts:
            - dispatched: "2026-10-23T11:46:25-07:00"
              finalized: "2026-10-23T11:49:46-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T11:56:03-07:00"
      dispatched: "2026-10-23T11:53:33-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: observeIntent takes flock.Shared, Sweep decides with no lock and removes via removeUnchanged, Hold waits IntentHoldPatience; TestIntentReadersAreNotOwners fails on revert of either half (verified in checkout).
          round: 2
        - id: BR-2
          disposition: addressed
          note: Root and pause owner/reason are now %q (mergeintent.go:117,135), pinned by the forged-root and quoted pause tests; the remaining Path sibling is raised separately.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Hold refuses a non-empty unreadable or other-branch record (intent.go:215-223); TestIntentHoldNeverOverwritesAnUnreadableRecord asserts the bytes are unchanged.
          round: 2
        - id: BR-4
          disposition: addressed
          note: localBranchHead (mergeintent.go:66) is the single lookup used by hold, sweep and observe.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Minor
          title: formatMergeIntent prints the unreadable record's Path unquoted; Path embeds a file name from the shared intents directory
          detail: '2nd in family. Rule: every string in formatMergeIntent that comes from the shared directory (contents or entry names) is printed with %q; only sdlc-built values (Reentry, abbreviated head, ints, times) are raw. Path at mergeintent.go:129 is the only remaining violation.'
          family: printed-untrusted-fields-unquoted
          round: 2
        - kind: defect
          id: BR-6
          severity: Minor
          title: Plan Core concepts table names MergeIntent, IntentStore, intentReentry and runDurableMergeAttempt, but the code has Intent, Intents, reportMergeIntent and runDurableMerge
          detail: Same behaviour under renamed entities, and List returns []ObservedIntent rather than (observed, unreadable). Add a Revisions entry recording the names as built.
          family: plan-table-drifts-from-code
          round: 2
        - kind: defect
          id: BR-7
          severity: Minor
          title: A reader in the window between Hold's O_CREATE and its lock reports a live merge's empty file as unreadable and tells the operator to remove it
          detail: 'intent.go:190-194 creates the file before taking the exclusive lock, so observeIntent can see an empty, unlocked record ("intent: empty record"). Lock-state rule: no reader classification should advise an action on a record whose writer may hold it next; treat an empty, unlocked record as transient (skip it or report it live) rather than unreadable.'
          family: reader-lock-masquerades-as-owner
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: 1a8a6b06a228bf8ae0bfb55ca8467de46e32b010
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ac8be880-49e3-47a0-9c5c-3f95c127e975
            repository: github.com/xianxu/ariadne
            issue_id: "000565"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/4d0f435b9867b717
            reviewed_head: 5f60907996bc7a32ea2453e72e625043a4780a04
            slot: ariadne:5
            requested_at: "2026-10-23T18:58:59.470679Z"
          phase: finished
          pid: 12288
          started_at: "2026-10-23T18:58:59.630195Z"
          finished_at: "2026-10-23T19:00:56.167535Z"
          session_id: 66aae585-db32-4bee-9f27-2973edd3c148
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 5f60907996bc7a32ea2453e72e625043a4780a04
        tracker_head: 601017fa13bb2e11832fb8e58aa5cec5d20acce9
        branch_point: 668e7637556b4748178be10e69ee969922c01e29
        since: "2026-10-23T10:46:54-07:00"
        until: "2026-10-23T11:58:01-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T10:46:54-07:00"
              end: "2026-10-23T11:58:01-07:00"
              began: "2026-10-23T10:46:54-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: fe692fd8394eec593a175ae454c1b9a8ef6f560a4dd7efa3192ed6d11a7aab2a
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 9163745
              windows:
                - start: "2026-10-23T10:46:54-07:00"
                  end: "2026-10-23T11:58:01-07:00"
        cuts:
            - dispatched: "2026-10-23T11:46:25-07:00"
              finalized: "2026-10-23T11:49:46-07:00"
              finalizing: false
            - dispatched: "2026-10-23T11:53:33-07:00"
              finalized: "2026-10-23T11:56:03-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-23T12:00:57-07:00"
      dispatched: "2026-10-23T11:58:59-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-5
          disposition: addressed
          note: formatMergeIntent and Hold's refusal now print the path with %q; tests assert the quoted form (mergeintent_test.go:147, intent_test.go:291).
          round: 3
        - id: BR-6
          disposition: addressed
          note: The 2026-10-23 round-2 Revisions entry records Intent, Intents, List returning []ObservedIntent, reportMergeIntent/formatMergeIntent and runDurableMerge, all checked against landing.go:346 and mergeintent.go.
          round: 3
        - id: BR-7
          disposition: not-addressed
          note: The temp-then-link fix is present, but reverting it (restoring O_CREATE and the treat-empty-as-new branch) leaves every intent test passing, so no regression test pins the window.
          round: 3
      findings:
        - kind: defect
          id: BR-8
          severity: Minor
          title: sweepTemps can remove a running merge's temp file before create has locked it, so a concurrent merge fails to start
          detail: '3rd in this family. Rule: never judge a file''s owner dead from its lock state while the file can still be created but not yet locked. os.CreateTemp (intent.go:259) makes .tmp-* visible before flock.Exclusive (intent.go:263); another merge''s Sweep (mergeintent.go:42, intent.go:379-393) can lock and unlink it, and create then fails with EWOULDBLOCK or ENOENT on Link. Fix: have sweepTemps skip temp files younger than IntentHoldPatience, or retry create on a lock or link failure; add a test using a temp file that is fresh and unlocked.'
          family: reader-lock-masquerades-as-owner
          round: 3
      recipe: fix-review
      isolation: fenced
      reviewed: 5f60907996bc7a32ea2453e72e625043a4780a04
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 20c43f45-5874-41d4-812d-6132ad38a515
            repository: github.com/xianxu/ariadne
            issue_id: "000565"
            recipe: fix-review
            round: 4
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/004a8b631cbd5e9d
            reviewed_head: 0455eba5f1c5e3f4e8558a2e0449043185db0757
            slot: ariadne:5
            requested_at: "2026-10-23T21:49:28.065511Z"
          phase: finished
          pid: 19031
          started_at: "2026-10-23T21:49:28.171204Z"
          finished_at: "2026-10-23T22:01:43.771279Z"
          session_id: f39254ad-1c0c-400e-8f94-b94e73359f1e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 0455eba5f1c5e3f4e8558a2e0449043185db0757
        tracker_head: 989a9a71af55dadc67f3abd40b4fa7fb17dc3be4
        branch_point: 1e4dd8d59b367cda6d96684d7b6d6a05e00693bd
        since: "2026-10-23T10:46:54-07:00"
        until: "2026-10-23T13:42:52-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T10:46:54-07:00"
              end: "2026-10-23T13:42:52-07:00"
              began: "2026-10-23T10:46:54-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 3358605b867a18a82995bcbbd515929e9c6c76b759e69fb924bd5e98d02c9b17
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 10252146
              windows:
                - start: "2026-10-23T10:46:54-07:00"
                  end: "2026-10-23T13:42:52-07:00"
        queue_waits:
            - attempt_id: a5cdb53c3fbaa11f3ff4482ad3441cce
              slot: ariadne:5
              record_digest: fb5e1580853d0e2d4edeed1e72e12c4b38efcceda16f77fe770b9ef9ae7c2bd3
              start: "2026-10-23T19:02:34.915126Z"
              end: "2026-10-23T20:19:19.899172Z"
              queue_seconds: 4604.984046
              raw_seconds: 4604.984046
              exact: true
        cuts:
            - dispatched: "2026-10-23T11:46:25-07:00"
              finalized: "2026-10-23T11:49:46-07:00"
              finalizing: false
            - dispatched: "2026-10-23T11:53:33-07:00"
              finalized: "2026-10-23T11:56:03-07:00"
              finalizing: true
            - dispatched: "2026-10-23T11:58:59-07:00"
              finalized: "2026-10-23T12:00:57-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-23T15:01:44-07:00"
      dispatched: "2026-10-23T14:49:28-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-7
          disposition: addressed
          note: Temp-then-link creation is now pinned by TestIntentFirstRecordIsInvisibleUntilComplete; reverting to direct O_CREATE at path makes it fail (reader saw a record before link).
          round: 4
        - id: BR-8
          disposition: addressed
          note: sweepTemps skips temps younger than IntentHoldPatience (intent.go:399-403) and create retries errCreateRaced; reverting the age check fails TestIntentCreationLeavesNoWindowOrResidue on the fresh unlocked temp.
          round: 4
      recipe: fix-review
      isolation: fenced
      reviewed: 0455eba5f1c5e3f4e8558a2e0449043185db0757
      blocked: false
---

# Gate ledger — ariadne#565 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T11:49:46-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `reader-lock-masquerades-as-owner` (kind: defect) Status/sweep readers take the owner's exclusive try-lock, so a concurrent re-entry refuses as "already running" naming the dead pid
  Intents.observe (intent.go:257-271) try-locks LOCK_EX on every intent, and Sweep runs a git rev-parse subprocess while holding it (mergeintent.go:68-70); every sdlc merge start and landing status sweeps all intents. A Hold during that window gets IntentHeldError naming the dead runner's pid ("wait for it or stop it"). A throwaway test against the pinned head reproduced the refusal. A fleet re-entering after a full shutdown is exactly when these collide. Fix: Hold retries for a bounded time (about 2 s) before refusing; Sweep works out branch existence before locking; add a regression test that holds an observe lock while Hold runs.
- **BR-2** [Minor] `printed-untrusted-fields-unquoted` (kind: defect) formatMergeIntent prints Root and pause owner/reason unquoted, against the plan's every-field-%q promise
  Valid only requires an absolute, clean root, so a hand-edited intent root (or pause.json reason) can put newlines or control characters into landing status / sdlc state output (mergeintent.go:124, mergeintent.go:106-110).
- **BR-3** [Minor] `tolerant-writer-drops-unknown-fields` (kind: defect) Hold silently overwrites an unreadable same-branch intent, dropping a newer version's fields
  When readIntent fails (corrupt or IntentVersion 2), Hold rewrites the file without merging extras (intent.go:189-200), although the atlas says unreadable intents are never acted on.
- **BR-4** [Minor] `duplicated-git-lookup` (kind: defect) The local branch-head rev-parse lookup is repeated three times in mergeintent.go
  holdMergeIntent, sweepMergeIntents and observeMergeIntents each build the same rev-parse --verify --quiet refs/heads/B call; one helper would cover all three.

## Round 2 — 2026-10-23T11:56:03-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — observeIntent takes flock.Shared, Sweep decides with no lock and removes via removeUnchanged, Hold waits IntentHoldPatience; TestIntentReadersAreNotOwners fails on revert of either half (verified in checkout).
- BR-2 — addressed — Root and pause owner/reason are now %q (mergeintent.go:117,135), pinned by the forged-root and quoted pause tests; the remaining Path sibling is raised separately.
- BR-3 — addressed — Hold refuses a non-empty unreadable or other-branch record (intent.go:215-223); TestIntentHoldNeverOverwritesAnUnreadableRecord asserts the bytes are unchanged.
- BR-4 — addressed — localBranchHead (mergeintent.go:66) is the single lookup used by hold, sweep and observe.

### Raised

- **BR-5** [Minor] `printed-untrusted-fields-unquoted` (kind: defect) formatMergeIntent prints the unreadable record's Path unquoted; Path embeds a file name from the shared intents directory
  2nd in family. Rule: every string in formatMergeIntent that comes from the shared directory (contents or entry names) is printed with %q; only sdlc-built values (Reentry, abbreviated head, ints, times) are raw. Path at mergeintent.go:129 is the only remaining violation.
- **BR-6** [Minor] `plan-table-drifts-from-code` (kind: defect) Plan Core concepts table names MergeIntent, IntentStore, intentReentry and runDurableMergeAttempt, but the code has Intent, Intents, reportMergeIntent and runDurableMerge
  Same behaviour under renamed entities, and List returns []ObservedIntent rather than (observed, unreadable). Add a Revisions entry recording the names as built.
- **BR-7** [Minor] `reader-lock-masquerades-as-owner` (kind: defect) A reader in the window between Hold's O_CREATE and its lock reports a live merge's empty file as unreadable and tells the operator to remove it
  intent.go:190-194 creates the file before taking the exclusive lock, so observeIntent can see an empty, unlocked record ("intent: empty record"). Lock-state rule: no reader classification should advise an action on a record whose writer may hold it next; treat an empty, unlocked record as transient (skip it or report it live) rather than unreadable.

## Round 3 — 2026-10-23T12:00:57-07:00 (claude) — passed

### Disposed

- BR-5 — addressed — formatMergeIntent and Hold's refusal now print the path with %q; tests assert the quoted form (mergeintent_test.go:147, intent_test.go:291).
- BR-6 — addressed — The 2026-10-23 round-2 Revisions entry records Intent, Intents, List returning []ObservedIntent, reportMergeIntent/formatMergeIntent and runDurableMerge, all checked against landing.go:346 and mergeintent.go.
- BR-7 — not-addressed — The temp-then-link fix is present, but reverting it (restoring O_CREATE and the treat-empty-as-new branch) leaves every intent test passing, so no regression test pins the window.

### Raised

- **BR-8** [Minor] `reader-lock-masquerades-as-owner` (kind: defect) sweepTemps can remove a running merge's temp file before create has locked it, so a concurrent merge fails to start
  3rd in this family. Rule: never judge a file's owner dead from its lock state while the file can still be created but not yet locked. os.CreateTemp (intent.go:259) makes .tmp-* visible before flock.Exclusive (intent.go:263); another merge's Sweep (mergeintent.go:42, intent.go:379-393) can lock and unlink it, and create then fails with EWOULDBLOCK or ENOENT on Link. Fix: have sweepTemps skip temp files younger than IntentHoldPatience, or retry create on a lock or link failure; add a test using a temp file that is fresh and unlocked.

## Round 4 — 2026-10-23T15:01:44-07:00 (claude) — passed

### Disposed

- BR-7 — addressed — Temp-then-link creation is now pinned by TestIntentFirstRecordIsInvisibleUntilComplete; reverting to direct O_CREATE at path makes it fail (reader saw a record before link).
- BR-8 — addressed — sweepTemps skips temps younger than IntentHoldPatience (intent.go:399-403) and create retries errCreateRaced; reverting the age check fails TestIntentCreationLeavesNoWindowOrResidue on the fresh unlocked temp.

## Open findings

(none — every finding has been disposed)
