---
gate: boundary-review
issue: 573
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 08e0496f-5021-4ac4-b3e5-6bd1f8fb0930
            repository: github.com/xianxu/ariadne
            issue_id: "000573"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/5ac2508f0f921db2
            reviewed_head: d8789eaa01cfde72a5863da477af87aeee73dec7
            slot: ariadne:2
            requested_at: "2026-10-23T16:08:23.27629Z"
          phase: finished
          pid: 71041
          started_at: "2026-10-23T16:08:23.371697Z"
          finished_at: "2026-10-23T16:10:06.387925Z"
          session_id: b7949763-de2d-418b-9257-022d7495d1f9
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: d8789eaa01cfde72a5863da477af87aeee73dec7
        tracker_head: b7403b40b9c8f3c46b10a84f85d1ecd7bf5fc868
        branch_point: 3639d6162dd1ac3ea3b88ce031cecd7ba1e6136e
        since: "2026-10-23T08:01:51-07:00"
        until: "2026-10-23T08:14:22-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T08:01:51-07:00"
              end: "2026-10-23T08:14:22-07:00"
              began: "2026-10-23T08:01:51-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c34cb061b3a5dab8783c68b64ac2202adecfe40f97d2af1af31892f42f9ac96f
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 6395259
              windows:
                - start: "2026-10-23T08:01:51-07:00"
                  end: "2026-10-23T08:14:22-07:00"
      "n": 1
      timestamp: "2026-10-23T09:10:07-07:00"
      dispatched: "2026-10-23T09:08:23-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: dev-aliases-shadow.test.sh expects /bin/ls, but merged-usr Linux (ubuntu-latest CI) resolves it to /usr/bin/ls
          detail: 'The test''s PATH lists /usr/bin before /bin. On Ubuntu 24.04, `type -P ls` returns /usr/bin/ls, and lsalias (a link to /bin/ls) resolves through `cd -P /bin` to /usr/bin/ls. The greps at lines 64 and 65 for "the system command /bin/ls;" would then fail. Every instance: lines 64 and 65 fail; line 24 (pure unit table) and line 45 (the link target) are the same assumption but harmless. Fix: compute the canonical system ls path at test start and use it in both greps. Not run on Linux: the Docker daemon was not running, so the conclusion comes from the PATH order and how cd -P resolves /bin.'
          family: test-hardcodes-host-filesystem-layout
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: 'ARCH-DRY: the symlink loop in system_command nearly repeats the script-path dereference at dev-aliases.sh:151-160'
          detail: Both use readlink plus path_parts to follow links by hand, which bash 3.2 needs. Both could share one resolve_link helper.
          family: duplicated-symlink-resolution
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 9fc1f045-0878-4023-8d75-38beb0af90ae
            repository: github.com/xianxu/ariadne
            issue_id: "000573"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/695dc5e30e4c26ff
            reviewed_head: 5903d2db7eb2916c3f5a3ae130cd99e31d54da05
            slot: ariadne:2
            requested_at: "2026-10-23T16:11:31.897541Z"
          phase: finished
          pid: 1849
          started_at: "2026-10-23T16:11:32.000518Z"
          finished_at: "2026-10-23T16:13:22.543303Z"
          session_id: 44a4cfe1-6bf1-426b-8f0b-02cd2460bfee
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 5903d2db7eb2916c3f5a3ae130cd99e31d54da05
        tracker_head: b7403b40b9c8f3c46b10a84f85d1ecd7bf5fc868
        branch_point: 3639d6162dd1ac3ea3b88ce031cecd7ba1e6136e
        since: "2026-10-23T08:01:51-07:00"
        until: "2026-10-23T09:11:04-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T08:01:51-07:00"
              end: "2026-10-23T09:11:04-07:00"
              began: "2026-10-23T08:01:51-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f2e6083102483aa35eb3ca2c93c72b5e01f7fc7cfc345d2e1d508def3f4ac2ad
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 6510622
              windows:
                - start: "2026-10-23T08:01:51-07:00"
                  end: "2026-10-23T09:11:04-07:00"
        queue_waits:
            - attempt_id: 276b751079eec255663510761f321f91
              slot: ariadne:2
              record_digest: fda5634cbb50f671ec8cdeac36e14011ff4d614c9c28ba6d0d03a227b481b42b
              start: "2026-10-23T15:14:55.719901Z"
              end: "2026-10-23T15:53:13.207442Z"
              queue_seconds: 2297.487541
              raw_seconds: 2297.487541
              exact: true
        cuts:
            - dispatched: "2026-10-23T09:08:23-07:00"
              finalized: "2026-10-23T09:10:07-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T09:13:23-07:00"
      dispatched: "2026-10-23T09:11:31-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: The test now derives ls_want and lsalias_want with cd -P, the same as system_command does; traced for merged-usr. Not run on Linux (no Docker).
          round: 2
        - id: BR-2
          disposition: addressed
          note: follow_links (dev-aliases.sh:138) is now shared by the script-path dereference (:165) and system_command (:473); the path, slot and base tests pass under bash 3.2.
          round: 2
      findings:
        - kind: defect
          id: BR-3
          severity: Minor
          title: Readlink count of exactly 2 and lsalias_want both assume no host system binary reached by the fixture is a symlink
          detail: '2nd finding in this family. Rule: every test expectation that depends on the host filesystem must be derived from the host with the generator''s own resolution (follow_links then cd -P), or the fixture must not reach host binaries at all. system_command runs even for names already refused as builtins (dev-aliases.sh:472). So on Debian with zsh and mailcap (/usr/bin/print is a symlink to run-mailcap) the count is 3, and on busybox (where /bin/ls is a symlink to busybox) both the count and lsalias_want are wrong. Fix: call system_command only when the bash and zsh kinds gave no reason, and derive the expected hop count and lsalias_want with that same resolution.'
          family: test-hardcodes-host-filesystem-layout
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: 5903d2db7eb2916c3f5a3ae130cd99e31d54da05
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: bbf6c2e9-452a-45c6-8610-3319a00f731a
            repository: github.com/xianxu/ariadne
            issue_id: "000573"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/1fcc860100633bf9
            reviewed_head: c0cb4a5090b5b149295e422a4b558d88717a2d51
            slot: ariadne:2
            requested_at: "2026-10-23T16:15:30.703983Z"
          phase: finished
          pid: 50972
          started_at: "2026-10-23T16:15:30.803547Z"
          finished_at: "2026-10-23T16:16:30.467416Z"
          session_id: 8ba5a2c1-718d-4bee-a808-93fde86ef083
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: c0cb4a5090b5b149295e422a4b558d88717a2d51
        tracker_head: 2dd00d5bfd2c76f4d98d2421601c905d9b631ec3
        branch_point: 3639d6162dd1ac3ea3b88ce031cecd7ba1e6136e
        since: "2026-10-23T08:01:51-07:00"
        until: "2026-10-23T09:15:03-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T08:01:51-07:00"
              end: "2026-10-23T09:15:03-07:00"
              began: "2026-10-23T08:01:51-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: a3209830796dd580e93b1dba3e64cc49f6e1964f741e14f40f4e0dd730a9a88d
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 6802469
              windows:
                - start: "2026-10-23T08:01:51-07:00"
                  end: "2026-10-23T09:15:03-07:00"
        queue_waits:
            - attempt_id: 276b751079eec255663510761f321f91
              slot: ariadne:2
              record_digest: fda5634cbb50f671ec8cdeac36e14011ff4d614c9c28ba6d0d03a227b481b42b
              start: "2026-10-23T15:14:55.719901Z"
              end: "2026-10-23T15:53:13.207442Z"
              queue_seconds: 2297.487541
              raw_seconds: 2297.487541
              exact: true
        cuts:
            - dispatched: "2026-10-23T09:08:23-07:00"
              finalized: "2026-10-23T09:10:07-07:00"
              finalizing: false
            - dispatched: "2026-10-23T09:11:31-07:00"
              finalized: "2026-10-23T09:13:23-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-23T09:16:31-07:00"
      dispatched: "2026-10-23T09:15:30-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-3
          disposition: addressed
          note: 'system_command now runs only when neither shell refuses the name, and the test derives ls_want, lsalias_want and want_hops via the generator''s own follow_links plus cd -P. Reverting the skip turns the test red with "8 names: 3 readlink calls, want 2".'
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: c0cb4a5090b5b149295e422a4b558d88717a2d51
      blocked: false
---

# Gate ledger — ariadne#573 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T09:10:07-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `test-hardcodes-host-filesystem-layout` (kind: defect) dev-aliases-shadow.test.sh expects /bin/ls, but merged-usr Linux (ubuntu-latest CI) resolves it to /usr/bin/ls
  The test's PATH lists /usr/bin before /bin. On Ubuntu 24.04, `type -P ls` returns /usr/bin/ls, and lsalias (a link to /bin/ls) resolves through `cd -P /bin` to /usr/bin/ls. The greps at lines 64 and 65 for "the system command /bin/ls;" would then fail. Every instance: lines 64 and 65 fail; line 24 (pure unit table) and line 45 (the link target) are the same assumption but harmless. Fix: compute the canonical system ls path at test start and use it in both greps. Not run on Linux: the Docker daemon was not running, so the conclusion comes from the PATH order and how cd -P resolves /bin.
- **BR-2** [Minor] `duplicated-symlink-resolution` (kind: defect) ARCH-DRY: the symlink loop in system_command nearly repeats the script-path dereference at dev-aliases.sh:151-160
  Both use readlink plus path_parts to follow links by hand, which bash 3.2 needs. Both could share one resolve_link helper.

## Round 2 — 2026-10-23T09:13:23-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — The test now derives ls_want and lsalias_want with cd -P, the same as system_command does; traced for merged-usr. Not run on Linux (no Docker).
- BR-2 — addressed — follow_links (dev-aliases.sh:138) is now shared by the script-path dereference (:165) and system_command (:473); the path, slot and base tests pass under bash 3.2.

### Raised

- **BR-3** [Minor] `test-hardcodes-host-filesystem-layout` (kind: defect) Readlink count of exactly 2 and lsalias_want both assume no host system binary reached by the fixture is a symlink
  2nd finding in this family. Rule: every test expectation that depends on the host filesystem must be derived from the host with the generator's own resolution (follow_links then cd -P), or the fixture must not reach host binaries at all. system_command runs even for names already refused as builtins (dev-aliases.sh:472). So on Debian with zsh and mailcap (/usr/bin/print is a symlink to run-mailcap) the count is 3, and on busybox (where /bin/ls is a symlink to busybox) both the count and lsalias_want are wrong. Fix: call system_command only when the bash and zsh kinds gave no reason, and derive the expected hop count and lsalias_want with that same resolution.

## Round 3 — 2026-10-23T09:16:31-07:00 (claude) — passed

### Disposed

- BR-3 — addressed — system_command now runs only when neither shell refuses the name, and the test derives ls_want, lsalias_want and want_hops via the generator's own follow_links plus cd -P. Reverting the skip turns the test red with "8 names: 3 readlink calls, want 2".

## Open findings

(none — every finding has been disposed)
