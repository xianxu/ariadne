---
gate: boundary-review
issue: 350
id_prefix: BR
rounds:
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: b696b7d263d37bc9e4287a0c73ebd13f920afdf5
        tracker_head: 11346d7d8670215cdb8a925e96e50b21ab84730d
        branch_point: 1c39182532fa3cc78ea4935abff4ab1b87a6f01d
        since: "2026-10-13T17:19:46-07:00"
        until: "2026-10-13T18:32:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-13T17:19:46-07:00"
              end: "2026-10-13T18:32:41-07:00"
              began: "2026-10-13T17:19:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: bd759883-0d13-49e6-8d0c-73ff96809176
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5c34a25ce33e7b3b7cd7e70959429d55d73e25fcd681beea2202618f73af9d4a
              started: "2026-10-14T00:28:59.212Z"
              spawned: false
              bytes: 421205
              windows:
                - start: "2026-10-14T00:28:59.212Z"
                  end: "2026-10-13T18:32:41-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 8782a827ae1f577e22e936106a153674d6e285b9a73da68ab1ac78a37b528986
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 3258388
              windows:
                - start: "2026-10-13T17:19:46-07:00"
                  end: "2026-10-13T18:32:41-07:00"
      "n": 1
      timestamp: "2026-10-14T08:47:13-07:00"
      dispatched: "2026-10-14T08:44:50-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: README still lists only sdlc/datatype/vocabulary/doc-review as ariadne's make tools binaries
          detail: README.md:195 is stale now that the Makefile tools target builds tl-worker and tl-lead. Add both binaries and a one-line pointer to their --help.
          family: readme-tool-surface-sync
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: TestWorkerHelpDoesNotRestateSection15 omits the planned "whole Bash call" refusal; help.md:11 contains it
          detail: Either reword the citation in cmd/tl-worker/help.md or add a plan Revisions entry saying the one-line citation is allowed.
          family: plan-test-contract-drift
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: Live session-sync conformance never exercises fold, so the fake's fold refusals are unchecked against couch
          detail: cmd/tl-lead/conformance_test.go never folds. The argv form matches couch help, but the fold-not-owed refusal the fake models has no live check. Revisit at the M3 close.
          family: conformance-covers-modeled-behavior
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: Plan names ErrDeliveryGap and sdlc_fake_test.go; the code has neither (fake is fake_test.go, no gap error)
          detail: Add Revisions entries so the Core concepts table and the SessionSync description match the code.
          family: plan-core-concepts-drift
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: A fleet-inventory failure aborts the whole trains render, including the independent Trains block
          detail: cmd/tl-lead/trains.go:68 returns fleetErr before rendering the Trains block. It could render the off-train block as ? instead.
          family: degrade-visibly-not-abort
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 392c1079c8f81086a51ede0b695c4cc6e8bebdcc
        tracker_head: 11346d7d8670215cdb8a925e96e50b21ab84730d
        branch_point: 1c39182532fa3cc78ea4935abff4ab1b87a6f01d
        since: "2026-10-13T17:19:46-07:00"
        until: "2026-10-14T08:49:10-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-13T17:19:46-07:00"
              end: "2026-10-14T08:49:10-07:00"
              began: "2026-10-13T17:19:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: bd759883-0d13-49e6-8d0c-73ff96809176
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5c34a25ce33e7b3b7cd7e70959429d55d73e25fcd681beea2202618f73af9d4a
              started: "2026-10-14T00:28:59.212Z"
              spawned: false
              bytes: 421205
              windows:
                - start: "2026-10-14T00:28:59.212Z"
                  end: "2026-10-14T08:49:10-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5f209512d67c50e4e8f85ef6895bdc36f617afbb66fb503454f3f37c953c4313
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 3433594
              windows:
                - start: "2026-10-13T17:19:46-07:00"
                  end: "2026-10-14T08:49:10-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.052116666666666665
        cuts:
            - boundary: M1
              dispatched: "2026-10-14T08:44:50-07:00"
              finalized: "2026-10-14T08:47:13-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-14T08:51:31-07:00"
      dispatched: "2026-10-14T08:49:27-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:195 and atlas base-layer.md:75, setup-and-replication.md:54, sandbox.md:13 all list tl-worker and tl-lead; a grep sweep finds no stale list.
          round: 2
        - id: BR-2
          disposition: addressed
          note: help.md:11 reworded; the test forbids "whole Bash call"; restoring the old help.md fails TestWorkerHelpDoesNotRestateSection15.
          round: 2
        - id: BR-3
          disposition: addressed
          note: conformance_test.go:53-68 folds an ID not owed against real couch; it passed live on this host with TL_LEAD_LIVE=1.
          round: 2
        - id: BR-4
          disposition: addressed
          note: ErrDeliveryGap exists and is mapped in decodeReply; reverting the mapping fails TestDecodeReplyDeliveryGap; a plan Revisions entry fixes the fake file names.
          round: 2
        - id: BR-5
          disposition: addressed
          note: trains.go:63-70 renders Trains and marks Slots unread; reverting fails TestGatherTrainsFleetFailureDegrades.
          round: 2
      findings:
        - kind: defect
          id: BR-6
          severity: Minor
          title: TestSessionSyncConformance doc comment still says the test never folds
          detail: cmd/tl-lead/conformance_test.go:16-17 says "the test never folds", but lines 53-68 now fold a not-owed ID; reword it to say the fold names an ID that isn't owed.
          family: comment-tracks-code
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: 392c1079c8f81086a51ede0b695c4cc6e8bebdcc
      blocked: false
---

# Gate ledger — ariadne#350 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T08:47:13-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `readme-tool-surface-sync` (kind: defect) README still lists only sdlc/datatype/vocabulary/doc-review as ariadne's make tools binaries
  README.md:195 is stale now that the Makefile tools target builds tl-worker and tl-lead. Add both binaries and a one-line pointer to their --help.
- **BR-2** [Minor] `plan-test-contract-drift` (kind: defect) TestWorkerHelpDoesNotRestateSection15 omits the planned "whole Bash call" refusal; help.md:11 contains it
  Either reword the citation in cmd/tl-worker/help.md or add a plan Revisions entry saying the one-line citation is allowed.
- **BR-3** [Minor] `conformance-covers-modeled-behavior` (kind: defect) Live session-sync conformance never exercises fold, so the fake's fold refusals are unchecked against couch
  cmd/tl-lead/conformance_test.go never folds. The argv form matches couch help, but the fold-not-owed refusal the fake models has no live check. Revisit at the M3 close.
- **BR-4** [Minor] `plan-core-concepts-drift` (kind: defect) Plan names ErrDeliveryGap and sdlc_fake_test.go; the code has neither (fake is fake_test.go, no gap error)
  Add Revisions entries so the Core concepts table and the SessionSync description match the code.
- **BR-5** [Minor] `degrade-visibly-not-abort` (kind: defect) A fleet-inventory failure aborts the whole trains render, including the independent Trains block
  cmd/tl-lead/trains.go:68 returns fleetErr before rendering the Trains block. It could render the off-train block as ? instead.

## Round 2 — 2026-10-14T08:51:31-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — README.md:195 and atlas base-layer.md:75, setup-and-replication.md:54, sandbox.md:13 all list tl-worker and tl-lead; a grep sweep finds no stale list.
- BR-2 — addressed — help.md:11 reworded; the test forbids "whole Bash call"; restoring the old help.md fails TestWorkerHelpDoesNotRestateSection15.
- BR-3 — addressed — conformance_test.go:53-68 folds an ID not owed against real couch; it passed live on this host with TL_LEAD_LIVE=1.
- BR-4 — addressed — ErrDeliveryGap exists and is mapped in decodeReply; reverting the mapping fails TestDecodeReplyDeliveryGap; a plan Revisions entry fixes the fake file names.
- BR-5 — addressed — trains.go:63-70 renders Trains and marks Slots unread; reverting fails TestGatherTrainsFleetFailureDegrades.

### Raised

- **BR-6** [Minor] `comment-tracks-code` (kind: defect) TestSessionSyncConformance doc comment still says the test never folds
  cmd/tl-lead/conformance_test.go:16-17 says "the test never folds", but lines 53-68 now fold a not-owed ID; reword it to say the fold names an ID that isn't owed.

## Open findings

- **BR-6** [Minor] `comment-tracks-code` (kind: defect) TestSessionSyncConformance doc comment still says the test never folds
