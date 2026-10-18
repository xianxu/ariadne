---
gate: boundary-review
issue: 409
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: c4bc4de4-13e2-4ce2-b0f5-a06f4142d8e3
            repository: github.com/xianxu/ariadne
            issue_id: "000409"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/76308b093647e048
            reviewed_head: 7fe6c39b917990a37c04efa94aa6b96e954646c9
            slot: ariadne:2
            requested_at: "2026-10-18T12:32:24.753047Z"
          phase: finished
          pid: 18367
          started_at: "2026-10-18T12:32:24.914333Z"
          finished_at: "2026-10-18T12:35:51.678033Z"
          session_id: 0b2d6a50-30f4-45a0-a498-d2d2c3bafad4
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7fe6c39b917990a37c04efa94aa6b96e954646c9
        tracker_head: 5578fe23834f9c2160c593a05397eaad33b10d1b
        branch_point: d3ebb55260c6647fbc70b423efed20cd8c734c5b
        since: "2026-10-18T05:16:36-07:00"
        until: "2026-10-18T05:31:57-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:16:36-07:00"
              end: "2026-10-18T05:31:57-07:00"
              began: "2026-10-18T05:16:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 144a7ee5-cc37-41c5-ac73-b88bbf5223ba
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5f3da10d43be251869bdafd6a98084e32ec627ca1a5a4b930a44fbe4f280c081
              started: "2026-10-18T12:25:28.29Z"
              spawned: false
              bytes: 482080
              windows:
                - start: "2026-10-18T12:25:28.29Z"
                  end: "2026-10-18T05:31:57-07:00"
            - harness: claude
              id: 5ca5160d-491c-4aec-b8f5-f5e40a69efb5
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 42052d978d0f33bca789480a4b74830893f6dd041d8d8d097390922c5c9823d6
              started: "2026-10-18T12:21:14.242Z"
              spawned: false
              bytes: 563079
              windows:
                - start: "2026-10-18T12:21:14.242Z"
                  end: "2026-10-18T05:31:57-07:00"
            - harness: claude
              id: 6be380fe-6c30-4408-b5b9-026237f91e30
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 70f6e935edb05d9fad671da7d73d9b88df89415544fb98457e596558f0fb3332
              started: "2026-10-18T12:23:49.894Z"
              spawned: false
              bytes: 497183
              windows:
                - start: "2026-10-18T12:23:49.894Z"
                  end: "2026-10-18T05:31:57-07:00"
            - harness: claude
              id: 946b723a-6614-4e42-8189-fff378170870
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 18bbe8b8ed1e1d030e367159b09a2e30f9b27a75f8dcc5f5048ff5d6d9bb34f1
              started: "2026-10-18T12:25:42.657Z"
              spawned: false
              bytes: 366511
              windows:
                - start: "2026-10-18T12:25:42.657Z"
                  end: "2026-10-18T05:31:57-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 46da947c66a4904a05bce842137321a76f9d16471d09636a5b6b06196fb57eb8
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 150009716
              windows:
                - start: "2026-10-18T05:16:36-07:00"
                  end: "2026-10-18T05:31:57-07:00"
      "n": 1
      timestamp: "2026-10-18T05:35:52-07:00"
      dispatched: "2026-10-18T05:32:24-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: reviewer_order lacking another family makes Reviewer() silently return the author's own agent
          detail: Mode other with reviewer_order {"bounded":["claude"]} (or an empty list) and a claude main agent makes ReviewPlan return an empty plan, and Reviewer() at agent_resolver.go:85-91 falls through to `return main`, so claude reviews claude with no notice. Validate (agent_resolver.go:120-129) accepts the config. Reject empty, duplicate or single-family orders in Validate, and never return main under other. ARCH-PURPOSE.
          family: no-silent-same-family-review
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Unavailable can never classify a codex or gemini quota or auth failure; the fallback is inert for the default cross-family reviewer
          detail: dispatch.go:629-633 runs ReadStream only for claude, so codex and gemini runs have Stream=false and Unavailable (family.go:131-134) returns false except on a launch error. M1 promised tests for the quota and auth texts each CLI prints; TestUnavailable has only claude stream-json fixtures. Parse codex exec --json (and gemini's output) into Began, or state and test the limit in a plan Revision. ARCH-PURPOSE, ARCH-ORDER.
          family: unavailable-evidence-per-cli
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: close help still says gemini is reviewed by claude; M1 changed it to codex, and the new review.json keys are missing
          detail: cmd/sdlc/helptext/close.md:417 describes the old `other` mapping. The changed judge_test.go:1336 now expects codex. The REVIEWER section omits reviewer_order and reviewer_models, and atlas sdlc-binary.md:1191 still describes other as claude and codex only. A cheap fix, and it should not wait for M2.
          family: docs-track-behavior-change
          round: 1
        - kind: defect
          id: BR-4
          severity: Important
          title: reviewer_models is accepted and validated but has no consumer, and no milestone owns the agentArgs wiring
          detail: reviewconfig.go loads ReviewerModels into Policy.Models and ReviewPlan copies it into Candidate.Model, but agentArgs (dispatch.go:219-258) never passes a model, so a configured model is silently ignored. The plan's Integration section says it is passed as --model/-m, but neither M1 nor M2 lists that work. Add it to M2, or warn while it is unwired.
          family: config-key-without-consumer
          round: 1
        - kind: defect
          id: BR-5
          severity: Important
          title: 'Plan Core concepts table contradicts the code: file paths, FamilyOf and ReviewPlan signatures, missing entities'
          detail: The table puts ReviewPlan in reviewplan.go and Unavailable in apifailure.go (both are in family.go), lists FamilyOf(agent, model) with qoder resolved through its model (the code is FamilyOf(agent) with no qoder handling), lists ReviewPlan(authors, kind, policy) (the code adds main and available), and omits LocalAvailability and AgentRun.Began. Add a Revisions entry.
          family: plan-table-matches-code
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: trimReason in family.go has no caller
          family: dead-code
          round: 1
        - kind: defect
          id: BR-7
          severity: Minor
          title: reviewer_models allows a leading '-', and reviewer_order allows duplicate agents
          detail: Reject a flag-like model value before M2 passes it to the CLI (ARCH-SECURE), and reject an agent listed twice in one order.
          family: parse-config-to-typed-value
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#409 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T05:35:52-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `no-silent-same-family-review` (kind: defect) reviewer_order lacking another family makes Reviewer() silently return the author's own agent
  Mode other with reviewer_order {"bounded":["claude"]} (or an empty list) and a claude main agent makes ReviewPlan return an empty plan, and Reviewer() at agent_resolver.go:85-91 falls through to `return main`, so claude reviews claude with no notice. Validate (agent_resolver.go:120-129) accepts the config. Reject empty, duplicate or single-family orders in Validate, and never return main under other. ARCH-PURPOSE.
- **BR-2** [Important] `unavailable-evidence-per-cli` (kind: defect) Unavailable can never classify a codex or gemini quota or auth failure; the fallback is inert for the default cross-family reviewer
  dispatch.go:629-633 runs ReadStream only for claude, so codex and gemini runs have Stream=false and Unavailable (family.go:131-134) returns false except on a launch error. M1 promised tests for the quota and auth texts each CLI prints; TestUnavailable has only claude stream-json fixtures. Parse codex exec --json (and gemini's output) into Began, or state and test the limit in a plan Revision. ARCH-PURPOSE, ARCH-ORDER.
- **BR-3** [Important] `docs-track-behavior-change` (kind: defect) close help still says gemini is reviewed by claude; M1 changed it to codex, and the new review.json keys are missing
  cmd/sdlc/helptext/close.md:417 describes the old `other` mapping. The changed judge_test.go:1336 now expects codex. The REVIEWER section omits reviewer_order and reviewer_models, and atlas sdlc-binary.md:1191 still describes other as claude and codex only. A cheap fix, and it should not wait for M2.
- **BR-4** [Important] `config-key-without-consumer` (kind: defect) reviewer_models is accepted and validated but has no consumer, and no milestone owns the agentArgs wiring
  reviewconfig.go loads ReviewerModels into Policy.Models and ReviewPlan copies it into Candidate.Model, but agentArgs (dispatch.go:219-258) never passes a model, so a configured model is silently ignored. The plan's Integration section says it is passed as --model/-m, but neither M1 nor M2 lists that work. Add it to M2, or warn while it is unwired.
- **BR-5** [Important] `plan-table-matches-code` (kind: defect) Plan Core concepts table contradicts the code: file paths, FamilyOf and ReviewPlan signatures, missing entities
  The table puts ReviewPlan in reviewplan.go and Unavailable in apifailure.go (both are in family.go), lists FamilyOf(agent, model) with qoder resolved through its model (the code is FamilyOf(agent) with no qoder handling), lists ReviewPlan(authors, kind, policy) (the code adds main and available), and omits LocalAvailability and AgentRun.Began. Add a Revisions entry.
- **BR-6** [Minor] `dead-code` (kind: defect) trimReason in family.go has no caller
- **BR-7** [Minor] `parse-config-to-typed-value` (kind: defect) reviewer_models allows a leading '-', and reviewer_order allows duplicate agents
  Reject a flag-like model value before M2 passes it to the CLI (ARCH-SECURE), and reject an agent listed twice in one order.

## Open findings

- **BR-1** [Important] `no-silent-same-family-review` (kind: defect) reviewer_order lacking another family makes Reviewer() silently return the author's own agent
- **BR-2** [Important] `unavailable-evidence-per-cli` (kind: defect) Unavailable can never classify a codex or gemini quota or auth failure; the fallback is inert for the default cross-family reviewer
- **BR-3** [Important] `docs-track-behavior-change` (kind: defect) close help still says gemini is reviewed by claude; M1 changed it to codex, and the new review.json keys are missing
- **BR-4** [Important] `config-key-without-consumer` (kind: defect) reviewer_models is accepted and validated but has no consumer, and no milestone owns the agentArgs wiring
- **BR-5** [Important] `plan-table-matches-code` (kind: defect) Plan Core concepts table contradicts the code: file paths, FamilyOf and ReviewPlan signatures, missing entities
- **BR-6** [Minor] `dead-code` (kind: defect) trimReason in family.go has no caller
- **BR-7** [Minor] `parse-config-to-typed-value` (kind: defect) reviewer_models allows a leading '-', and reviewer_order allows duplicate agents
