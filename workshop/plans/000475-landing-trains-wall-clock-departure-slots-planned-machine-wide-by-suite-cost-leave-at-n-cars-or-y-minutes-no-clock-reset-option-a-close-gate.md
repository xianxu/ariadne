---
gate: boundary-review
issue: 475
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: e23a1d03-611b-4c3a-8af2-99eb9b7564b9
            repository: github.com/xianxu/ariadne
            issue_id: "000475"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/5b6b7dbec9edb75f
            reviewed_head: b97bc6dd7d96fcd482b2fc57c6231b394cb0bf64
            slot: ariadne:4
            requested_at: "2026-10-20T11:40:01.533863Z"
          phase: finished
          pid: 39821
          started_at: "2026-10-20T11:40:01.627835Z"
          finished_at: "2026-10-20T11:51:26.839134Z"
          session_id: ce54466e-eca4-404e-8a65-78527918ab82
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: d75999f8bd8cab6da9b65500899c89f5364ae986
        tracker_head: d54ab3581632e28192fe40464fdd682d4fadf7e2
        branch_point: 7918d3d63c4c12e9a4537b51680e2143b29fd286
        since: "2026-10-20T03:03:17-07:00"
        until: "2026-10-20T04:38:21-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T03:03:17-07:00"
              end: "2026-10-20T04:38:21-07:00"
              began: "2026-10-20T03:03:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 3d08a2cde284a5ec5c5d1007a1ed6ba5714c4ff43cca95d0874bb8ea8ecb69df
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 96520513
              windows:
                - start: "2026-10-20T03:03:17-07:00"
                  end: "2026-10-20T04:38:21-07:00"
            - harness: claude
              id: 3fc56c78-9cad-4e12-a1a7-aed05bd0c63d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: b676d6d103897ad92941db345e4e23c27533a35891f17b7b537450d5976ee8ad
              started: "2026-10-20T11:21:49.422Z"
              spawned: false
              bytes: 384376
              windows:
                - start: "2026-10-20T11:21:49.422Z"
                  end: "2026-10-20T04:38:21-07:00"
            - harness: claude
              id: 95466ede-cf18-4115-b13f-1faa8f33645d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1b281c55e873b03c51fd775f439ef68f3e2a49b0a3f6e6097df1fe069c81db79
              started: "2026-10-20T11:21:30.104Z"
              spawned: false
              bytes: 514596
              windows:
                - start: "2026-10-20T11:21:30.104Z"
                  end: "2026-10-20T04:38:21-07:00"
            - harness: claude
              id: ac351657-0a41-4804-b564-3e710ea2dd55
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 762c45ca71122a0c54e4c27a0ff8e23460c28233ba39f58b9420ecf8d981cda4
              started: "2026-10-20T11:19:02.044Z"
              spawned: false
              bytes: 587879
              windows:
                - start: "2026-10-20T11:19:02.044Z"
                  end: "2026-10-20T04:38:21-07:00"
            - harness: claude
              id: f5886db6-bd63-4d0e-aec8-021fea4f0187
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ff3b2eef3104ceeab9229868a244b067cce02222561142ea5df5f3dc7ec0d977
              started: "2026-10-20T11:20:37.435Z"
              spawned: false
              bytes: 530611
              windows:
                - start: "2026-10-20T11:20:37.435Z"
                  end: "2026-10-20T04:38:21-07:00"
        queue_waits:
            - attempt_id: 3611f175d5a4d5f65847189d1c3190ac
              slot: ariadne:4
              record_digest: 0bfdc61fc34a7d4333b805a3bf32ca5a4d0d8dabb13eea6559cf546adf6609a2
              start: "2026-10-20T10:14:59.477681Z"
              end: "2026-10-20T10:30:07.515298Z"
              queue_seconds: 908.037617
              raw_seconds: 908.037617
              exact: true
            - attempt_id: b648e3ff6d96d6e5b355164d3dbee74c
              slot: ariadne:4
              record_digest: 11df85f04d90f5060d15b092a898573c63d1a6d63d3891bf06bff55a7853326d
              start: "2026-10-20T11:03:21.374039Z"
              end: "2026-10-20T11:04:29.093112Z"
              queue_seconds: 67.719073
              raw_seconds: 67.719073
              exact: true
        protected:
            - issue: "490"
              slot: ariadne:4
              start: "2026-10-20T10:05:34.25Z"
              end: "2026-10-20T10:11:07.378Z"
      "n": 1
      timestamp: "2026-10-20T04:51:27-07:00"
      dispatched: "2026-10-20T04:40:01-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: README departure rule is stale and omits train_max_wait_minutes
          detail: README.md:521-531 lists only train_interval_minutes/train_max_cars and describes departure as first train / one interval after the last / at once after left-ready. It omits the full-train and max-wait early departures, the new key, and that lone and granted landings no longer reset the clock. helptext/landing.md is equally stale.
          family: user-docs-track-new-surface
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: T5 is ticked, but landing status lacks the promised reasons and golden test
          detail: The plan promised train due (reason) / next train at HH:MM (slot or max-wait), including cars-left-ready, plus a golden test. The code prints a bare train due for slot and left-ready, and an RFC3339 time with no source. The only test change is a 1-ready substring (landingcmd_test.go:196). Add the reasons and a status test, or narrow T5 in Revisions.
          family: plan-claim-matches-delivery
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: Event.Departed has no reader; the slot check uses the stored NextDue, not the plan's lastDeparture
          detail: PQ-2 says lastDeparture reads Departed. The code compares a NextDue fixed at write time, so Departed is written but never read, and a later offset change only applies after the next train.
          family: plan-claim-matches-delivery
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: NextDeparture does not ignore old-format alone records by Result as the Ordering section claims
          detail: train.go:84 filters on non-zero NextDue only, so an alone record written before the upgrade (departed + interval) still delays the next train once.
          family: plan-claim-matches-delivery
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: The max-wait predicate is duplicated in DepartureDue
          detail: train.go:228 and :234 repeat the same three-part condition; one wait-deadline helper would do.
          family: shared-helper-not-repeated-predicate
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: An older sdlc re-vouching carries a stale since through Extra across a head change
          detail: mergeExtra keeps since as an unknown member for an old binary, so after a head change the new binary trusts the stale Since and may depart early on max-wait.
          family: unknown-field-carry-stale-semantics
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: c159b511-dfc9-4d80-91e3-f34150a94d55
            repository: github.com/xianxu/ariadne
            issue_id: "000475"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/30ef4d045dbeb182
            reviewed_head: fcc0be143bb399a3b0176477be2ed52ca16da8e0
            slot: ariadne:4
            requested_at: "2026-10-20T11:55:52.686773Z"
          phase: finished
          pid: 58606
          started_at: "2026-10-20T11:55:52.902227Z"
          finished_at: "2026-10-20T12:08:04.201777Z"
          session_id: dfa6f6eb-bdcf-4fe1-87d5-aa1bbe132208
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: fcc0be143bb399a3b0176477be2ed52ca16da8e0
        tracker_head: d54ab3581632e28192fe40464fdd682d4fadf7e2
        branch_point: 7918d3d63c4c12e9a4537b51680e2143b29fd286
        since: "2026-10-20T03:03:17-07:00"
        until: "2026-10-20T04:55:08-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T03:03:17-07:00"
              end: "2026-10-20T04:55:08-07:00"
              began: "2026-10-20T03:03:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 11d1cc2b288eedfc58db0d538b7bf9200b92bd8d2c485cff50fd5feed17411f1
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 96775005
              windows:
                - start: "2026-10-20T03:03:17-07:00"
                  end: "2026-10-20T04:55:08-07:00"
            - harness: claude
              id: 3fc56c78-9cad-4e12-a1a7-aed05bd0c63d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: b676d6d103897ad92941db345e4e23c27533a35891f17b7b537450d5976ee8ad
              started: "2026-10-20T11:21:49.422Z"
              spawned: false
              bytes: 384376
              windows:
                - start: "2026-10-20T11:21:49.422Z"
                  end: "2026-10-20T04:55:08-07:00"
            - harness: claude
              id: 95466ede-cf18-4115-b13f-1faa8f33645d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1b281c55e873b03c51fd775f439ef68f3e2a49b0a3f6e6097df1fe069c81db79
              started: "2026-10-20T11:21:30.104Z"
              spawned: false
              bytes: 514596
              windows:
                - start: "2026-10-20T11:21:30.104Z"
                  end: "2026-10-20T04:55:08-07:00"
            - harness: claude
              id: ac351657-0a41-4804-b564-3e710ea2dd55
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 762c45ca71122a0c54e4c27a0ff8e23460c28233ba39f58b9420ecf8d981cda4
              started: "2026-10-20T11:19:02.044Z"
              spawned: false
              bytes: 587879
              windows:
                - start: "2026-10-20T11:19:02.044Z"
                  end: "2026-10-20T04:55:08-07:00"
            - harness: claude
              id: f5886db6-bd63-4d0e-aec8-021fea4f0187
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ff3b2eef3104ceeab9229868a244b067cce02222561142ea5df5f3dc7ec0d977
              started: "2026-10-20T11:20:37.435Z"
              spawned: false
              bytes: 530611
              windows:
                - start: "2026-10-20T11:20:37.435Z"
                  end: "2026-10-20T04:55:08-07:00"
        queue_waits:
            - attempt_id: 3611f175d5a4d5f65847189d1c3190ac
              slot: ariadne:4
              record_digest: 0bfdc61fc34a7d4333b805a3bf32ca5a4d0d8dabb13eea6559cf546adf6609a2
              start: "2026-10-20T10:14:59.477681Z"
              end: "2026-10-20T10:30:07.515298Z"
              queue_seconds: 908.037617
              raw_seconds: 908.037617
              exact: true
            - attempt_id: b648e3ff6d96d6e5b355164d3dbee74c
              slot: ariadne:4
              record_digest: 11df85f04d90f5060d15b092a898573c63d1a6d63d3891bf06bff55a7853326d
              start: "2026-10-20T11:03:21.374039Z"
              end: "2026-10-20T11:04:29.093112Z"
              queue_seconds: 67.719073
              raw_seconds: 67.719073
              exact: true
        protected:
            - issue: "490"
              slot: ariadne:4
              start: "2026-10-20T10:05:34.25Z"
              end: "2026-10-20T10:11:07.378Z"
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T04:40:01-07:00"
              finalized: "2026-10-20T04:51:27-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-20T05:08:05-07:00"
      dispatched: "2026-10-20T04:55:52-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:521-538 and helptext/landing.md now list train_max_wait_minutes, full and max-wait early departures, and that lone/granted landings never push the train back.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Status calls DepartureDue and prints the reason (slot, interval, max wait, full, cars left ready); landingcmd_test.go:852-862 pins "next train at 2026-10-20 10:00 (interval)".
          round: 2
        - id: BR-3
          disposition: addressed
          note: nextDeparture works out the slot from departedOf(e) when it decides (train.go:373); TestNextDepartureReadsTheDeparture covers it.
          round: 2
        - id: BR-4
          disposition: addressed
          note: The alone case is skipped by Result (train.go:370); reverting it makes TestNextDepartureReadsTheDeparture fail (checked in the checkout).
          round: 2
        - id: BR-5
          disposition: addressed
          note: waitDeadline helper (train.go:330) used for both the due check and the reported next time.
          round: 2
        - id: BR-6
          disposition: addressed
          note: SinceHead guard in firstReady (train.go:449); reverting it makes TestStaleSinceFromAnOlderSdlc fail (checked). A sibling case is raised separately.
          round: 2
      findings:
        - kind: defect
          id: BR-7
          severity: Minor
          title: An older sdlc's not-ready then ready re-vouch at the same head still carries a stale since
          detail: 'This is the 2nd finding in this family. The rule: a derived field that an older writer can carry through Extra must be stamped against every condition that resets it, not just one. SinceHead covers a head change but not a not-ready vouch, so mergeExtra keeps since across old-binary not-ready then ready at the same head, and firstReady trusts it (train.go:448-465). Fix: stamp the reset conditions too, for example a since_at equal to the At of the vouch that set Since, read as valid only while every vouch since then was ready. Prevalence: 2 of 2 carried fields in this family (since; head-reset fixed, readiness-reset open). The window needs mixed binaries on one line, so this is Minor.'
          family: unknown-field-carry-stale-semantics
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: fcc0be143bb399a3b0176477be2ed52ca16da8e0
      blocked: false
---

# Gate ledger — ariadne#475 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T04:51:27-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `user-docs-track-new-surface` (kind: defect) README departure rule is stale and omits train_max_wait_minutes
  README.md:521-531 lists only train_interval_minutes/train_max_cars and describes departure as first train / one interval after the last / at once after left-ready. It omits the full-train and max-wait early departures, the new key, and that lone and granted landings no longer reset the clock. helptext/landing.md is equally stale.
- **BR-2** [Important] `plan-claim-matches-delivery` (kind: defect) T5 is ticked, but landing status lacks the promised reasons and golden test
  The plan promised train due (reason) / next train at HH:MM (slot or max-wait), including cars-left-ready, plus a golden test. The code prints a bare train due for slot and left-ready, and an RFC3339 time with no source. The only test change is a 1-ready substring (landingcmd_test.go:196). Add the reasons and a status test, or narrow T5 in Revisions.
- **BR-3** [Minor] `plan-claim-matches-delivery` (kind: defect) Event.Departed has no reader; the slot check uses the stored NextDue, not the plan's lastDeparture
  PQ-2 says lastDeparture reads Departed. The code compares a NextDue fixed at write time, so Departed is written but never read, and a later offset change only applies after the next train.
- **BR-4** [Minor] `plan-claim-matches-delivery` (kind: defect) NextDeparture does not ignore old-format alone records by Result as the Ordering section claims
  train.go:84 filters on non-zero NextDue only, so an alone record written before the upgrade (departed + interval) still delays the next train once.
- **BR-5** [Minor] `shared-helper-not-repeated-predicate` (kind: defect) The max-wait predicate is duplicated in DepartureDue
  train.go:228 and :234 repeat the same three-part condition; one wait-deadline helper would do.
- **BR-6** [Minor] `unknown-field-carry-stale-semantics` (kind: defect) An older sdlc re-vouching carries a stale since through Extra across a head change
  mergeExtra keeps since as an unknown member for an old binary, so after a head change the new binary trusts the stale Since and may depart early on max-wait.

## Round 2 — 2026-10-20T05:08:05-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — README.md:521-538 and helptext/landing.md now list train_max_wait_minutes, full and max-wait early departures, and that lone/granted landings never push the train back.
- BR-2 — addressed — Status calls DepartureDue and prints the reason (slot, interval, max wait, full, cars left ready); landingcmd_test.go:852-862 pins "next train at 2026-10-20 10:00 (interval)".
- BR-3 — addressed — nextDeparture works out the slot from departedOf(e) when it decides (train.go:373); TestNextDepartureReadsTheDeparture covers it.
- BR-4 — addressed — The alone case is skipped by Result (train.go:370); reverting it makes TestNextDepartureReadsTheDeparture fail (checked in the checkout).
- BR-5 — addressed — waitDeadline helper (train.go:330) used for both the due check and the reported next time.
- BR-6 — addressed — SinceHead guard in firstReady (train.go:449); reverting it makes TestStaleSinceFromAnOlderSdlc fail (checked). A sibling case is raised separately.

### Raised

- **BR-7** [Minor] `unknown-field-carry-stale-semantics` (kind: defect) An older sdlc's not-ready then ready re-vouch at the same head still carries a stale since
  This is the 2nd finding in this family. The rule: a derived field that an older writer can carry through Extra must be stamped against every condition that resets it, not just one. SinceHead covers a head change but not a not-ready vouch, so mergeExtra keeps since across old-binary not-ready then ready at the same head, and firstReady trusts it (train.go:448-465). Fix: stamp the reset conditions too, for example a since_at equal to the At of the vouch that set Since, read as valid only while every vouch since then was ready. Prevalence: 2 of 2 carried fields in this family (since; head-reset fixed, readiness-reset open). The window needs mixed binaries on one line, so this is Minor.

## Open findings

- **BR-7** [Minor] `unknown-field-carry-stale-semantics` (kind: defect) An older sdlc's not-ready then ready re-vouch at the same head still carries a stale since
