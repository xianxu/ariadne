---
reviewer_runs:
    - version: 1
      intent:
        id: c7fcc2f0-56d7-4633-8178-e54e64792945
        repository: github.com/xianxu/ariadne
        issue_id: "000456"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        reviewed_head: f028a38e7d01f6b3b57f3f82186225de578b7013
        slot: ariadne:4
        requested_at: "2026-10-19T20:51:04.779514Z"
      phase: finished
      pid: 70199
      started_at: "2026-10-19T20:51:04.880697Z"
      finished_at: "2026-10-19T20:51:40.59981Z"
      session_id: ec355cf4-beff-4405-8e7c-aaa0b02ac6c0
      outcome: returned
gate: plan-quality
issue: 456
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: eb55d9d2-f946-4495-b761-09f852ac6085
            repository: github.com/xianxu/ariadne
            issue_id: "000456"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: aa1587a4c14eae3d392e3c7dae31ec9faf124ccd
            slot: ariadne:4
            requested_at: "2026-10-19T20:45:53.02653Z"
          phase: finished
          pid: 5986
          started_at: "2026-10-19T20:45:53.188233Z"
          finished_at: "2026-10-19T20:47:37.784402Z"
          session_id: 0045277a-2192-47eb-ac53-898254e7ea0f
          outcome: returned
      "n": 1
      timestamp: "2026-10-19T13:47:38-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: Done-when (issue and plan) still names dropped verbs and the old team-new shape; R5 is absent
          detail: Plan Done-when requires team commit, --team resolution and the default status FILE, all withdrawn in Revisions. The issue's Done-when says `team new <name>` scaffolds a team folder (now PATH creating or adopting an ops repo). The live echo proof is deferred on the grounds of pair#694, which is now parked. fleet resume has no acceptance line. Rewrite both Done-when blocks for the revised scope.
          family: done-when-tracks-revision
          round: 1
        - kind: defect
          id: PQ-2
          severity: Important
          title: fleet resume targets "slots of repos this hub claims" instead of Couch's owner resolution, so it can resume another team's slots
          detail: In couchcore/hub_ownership.go:82-116 an exact repo:N claim by another hub outranks a repo claim, and unclaimed repos go to the first hub. Select resume targets with R3's resolver (only slots resolving to this hub; ties listed as ambiguous), not a separate repo-claim filter (ARCH-DRY, ARCH-PURPOSE).
          family: ownership-single-resolver
          round: 1
        - kind: defect
          id: PQ-3
          severity: Minor
          title: tl-worker filing rule uses `issue new` then `unclaim`; `sdlc issue new --no-claim` exists for this case
          family: help-matches-tool-contract
          round: 1
        - kind: defect
          id: PQ-4
          severity: Minor
          title: Original Tasks 1-7 remain unmarked beside R1-R7; mark the dropped tasks as superseded
          family: superseded-plan-content
          round: 1
        - kind: defect
          id: PQ-5
          severity: Minor
          title: R1 keeps a team.md parse that nothing in the revised scope reads
          family: yagni-unconsumed-parse
          round: 1
        - kind: defect
          id: PQ-6
          severity: Minor
          title: planResume must treat an owner without a local workspace, and a failed actors read, as unknown, never as a match or as down
          family: failed-read-not-absence
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 9f88b15d-8589-4191-8162-703a8a44f348
            repository: github.com/xianxu/ariadne
            issue_id: "000456"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: 8b1615b381054440cf8e026782108c1cfca8c5c7
            slot: ariadne:4
            requested_at: "2026-10-19T20:48:25.491197Z"
          phase: finished
          pid: 15263
          started_at: "2026-10-19T20:48:25.818774Z"
          finished_at: "2026-10-19T20:49:20.10583Z"
          session_id: ab89d487-8a7b-416a-96ba-d59ad7542385
          outcome: returned
      "n": 2
      timestamp: "2026-10-19T13:49:20-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Both Done-when blocks were rewritten for the revised scope; fleet resume has an acceptance line; the pair#694 dependency is dropped.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: fleet resume targets come from ownerOf, ties and unclaimed slots are ambiguous, and the inventory covers the whole fleet.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: The round-1 revision and the Done-when use --no-claim.
          round: 2
        - id: PQ-4
          disposition: addressed
          round: 2
        - id: PQ-5
          disposition: addressed
          round: 2
        - id: PQ-6
          disposition: addressed
          round: 2
      findings:
        - kind: defect
          id: PQ-7
          severity: Important
          title: ownerOf omits Couch's hub-slot rule and its runtime couch hub move overrides, so fleet resume can resume a slot moved to another hub
          detail: 'This is the 2nd finding in family ownership-single-resolver. Rule: tl-lead''s ownership copy covers only the layer Couch persists (.hub); any action that depends on it must either read Couch''s own answer or list the layers it can''t see and treat those slots as unknown. couchcore/hub_ownership.go:86-91 checks hubForSlot and the ephemeral overrides (couchcmd/hub_move.go:83) before the claim ranks, and OwnerHub reaches only Couch''s menu (couchtty/thread_presentation.go:281), not a JSON read. Fix: add an owner read to the Couch seam (fake plus live decode), or state that runtime moves are invisible and stop --apply from resuming a slot until Couch confirms its owner; also model the hub-slot rule in ownerOf.'
          family: ownership-single-resolver
          round: 2
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 3287b402-34f4-4db6-bc67-2f61bad5bce9
            repository: github.com/xianxu/ariadne
            issue_id: "000456"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: 9a2994632184c565279c0798bfac472350d29ee2
            slot: ariadne:4
            requested_at: "2026-10-19T20:50:01.627053Z"
          phase: finished
          pid: 59596
          started_at: "2026-10-19T20:50:01.722117Z"
          finished_at: "2026-10-19T20:50:23.263673Z"
          session_id: 8b6e6e4a-fd47-42e4-8519-5bf10cd63759
          outcome: returned
      "n": 3
      timestamp: "2026-10-19T13:50:23-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-7
          disposition: addressed
          note: Round-2 revision mirrors hubForSlot, states the family rule (copy covers only what Couch persists), names runtime overrides as invisible with a run-time caveat and pre-apply slot list, and bounds the consequence; verified against pair couchcore/hub_ownership.go:61,82-116.
          round: 3
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: c354702a-a316-46c3-b397-f0819bfb5d22
            repository: github.com/xianxu/ariadne
            issue_id: "000456"
            recipe: plan-quality
            round: 4
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: f028a38e7d01f6b3b57f3f82186225de578b7013
            slot: ariadne:4
            requested_at: "2026-10-19T20:50:54.741634Z"
          phase: finished
          pid: 69039
          started_at: "2026-10-19T20:50:54.839835Z"
          finished_at: "2026-10-19T20:51:03.379249Z"
          session_id: 68bad21f-151d-4bc6-ac7a-fe968a6e52e1
          outcome: returned
      "n": 4
      timestamp: "2026-10-19T13:51:03-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      blocked: false
      protocol_error: no valid findings block
content_hash: da9e3d20d0be8d045d36c6d3818a2f56543fc09324ab14b6ff7d433b7c867f5d
---

# Gate ledger — ariadne#456 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T13:47:38-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `done-when-tracks-revision` (kind: defect) Done-when (issue and plan) still names dropped verbs and the old team-new shape; R5 is absent
  Plan Done-when requires team commit, --team resolution and the default status FILE, all withdrawn in Revisions. The issue's Done-when says `team new <name>` scaffolds a team folder (now PATH creating or adopting an ops repo). The live echo proof is deferred on the grounds of pair#694, which is now parked. fleet resume has no acceptance line. Rewrite both Done-when blocks for the revised scope.
- **PQ-2** [Important] `ownership-single-resolver` (kind: defect) fleet resume targets "slots of repos this hub claims" instead of Couch's owner resolution, so it can resume another team's slots
  In couchcore/hub_ownership.go:82-116 an exact repo:N claim by another hub outranks a repo claim, and unclaimed repos go to the first hub. Select resume targets with R3's resolver (only slots resolving to this hub; ties listed as ambiguous), not a separate repo-claim filter (ARCH-DRY, ARCH-PURPOSE).
- **PQ-3** [Minor] `help-matches-tool-contract` (kind: defect) tl-worker filing rule uses `issue new` then `unclaim`; `sdlc issue new --no-claim` exists for this case
- **PQ-4** [Minor] `superseded-plan-content` (kind: defect) Original Tasks 1-7 remain unmarked beside R1-R7; mark the dropped tasks as superseded
- **PQ-5** [Minor] `yagni-unconsumed-parse` (kind: defect) R1 keeps a team.md parse that nothing in the revised scope reads
- **PQ-6** [Minor] `failed-read-not-absence` (kind: defect) planResume must treat an owner without a local workspace, and a failed actors read, as unknown, never as a match or as down

## Round 2 — 2026-10-19T13:49:20-07:00 (claude) — BLOCKED

### Disposed

- PQ-1 — addressed — Both Done-when blocks were rewritten for the revised scope; fleet resume has an acceptance line; the pair#694 dependency is dropped.
- PQ-2 — addressed — fleet resume targets come from ownerOf, ties and unclaimed slots are ambiguous, and the inventory covers the whole fleet.
- PQ-3 — addressed — The round-1 revision and the Done-when use --no-claim.
- PQ-4 — addressed
- PQ-5 — addressed
- PQ-6 — addressed

### Raised

- **PQ-7** [Important] `ownership-single-resolver` (kind: defect) ownerOf omits Couch's hub-slot rule and its runtime couch hub move overrides, so fleet resume can resume a slot moved to another hub
  This is the 2nd finding in family ownership-single-resolver. Rule: tl-lead's ownership copy covers only the layer Couch persists (.hub); any action that depends on it must either read Couch's own answer or list the layers it can't see and treat those slots as unknown. couchcore/hub_ownership.go:86-91 checks hubForSlot and the ephemeral overrides (couchcmd/hub_move.go:83) before the claim ranks, and OwnerHub reaches only Couch's menu (couchtty/thread_presentation.go:281), not a JSON read. Fix: add an owner read to the Couch seam (fake plus live decode), or state that runtime moves are invisible and stop --apply from resuming a slot until Couch confirms its owner; also model the hub-slot rule in ownerOf.

## Round 3 — 2026-10-19T13:50:23-07:00 (claude) — passed

### Disposed

- PQ-7 — addressed — Round-2 revision mirrors hubForSlot, states the family rule (copy covers only what Couch persists), names runtime overrides as invisible with a run-time caveat and pre-apply slot list, and bounds the consequence; verified against pair couchcore/hub_ownership.go:61,82-116.

## Round 4 — 2026-10-19T13:51:03-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
