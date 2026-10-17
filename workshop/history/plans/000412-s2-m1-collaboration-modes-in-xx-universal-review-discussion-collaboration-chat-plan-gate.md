---
reviewer_runs:
    - version: 1
      intent:
        id: 10383b66-264f-4eac-bb96-31dfbe8944c4
        repository: github.com/xianxu/ariadne
        issue_id: "000412"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        reviewed_head: 9a2d9f731f47c7a591f13ac7fcbba256824c44b3
        slot: ariadne:2
        requested_at: "2026-10-16T22:56:11.635805Z"
      phase: finished
      pid: 73093
      started_at: "2026-10-16T22:56:11.960795Z"
      finished_at: "2026-10-16T22:56:37.733605Z"
      session_id: e1a45820-de66-478e-bdce-95f833a64410
      outcome: returned
gate: plan-quality
issue: 412
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 8399402e-3ff5-489d-bbcc-b1f46175e9ed
            repository: github.com/xianxu/ariadne
            issue_id: "000412"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 3ffd721437171286e29f9cec7f733ebca757ee4e
            slot: ariadne:2
            requested_at: "2026-10-16T22:49:08.779575Z"
          phase: finished
          pid: 39807
          started_at: "2026-10-16T22:49:09.03784Z"
          finished_at: "2026-10-16T22:50:09.229936Z"
          session_id: 6681a7b6-a1e9-4bf6-8743-64ae382a6945
          outcome: returned
      "n": 1
      timestamp: "2026-10-16T15:50:10-07:00"
      agent: claude
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: The examples' side-doc layout contradicts itself (after.side/ vs before.side/ links) and can't represent before and after states of side docs
          detail: Core concepts says after.side/, but rows 3 and 12 link before.side/. Rows 6, 7, 14, 17b and chat/retry need side docs in both a before and an after state, which one folder can't hold. Define per-state trees (e.g. before/doc.md + doc.side/, after/doc.md + doc.side/) so the task 4 checker and the task 8 live diff have one layout to resolve links against.
          family: fixture-layout-unspecified
          round: 1
        - kind: defect
          id: PQ-2
          severity: Minor
          title: Moving answer docs to tl-status-<op>.side/ doesn't mention gitignoring that folder in ops
          detail: help.md:185-186 says QA files are gitignored like the status file; the new folder needs the same, or private answers get committed to ops. Add it to the TL's landing report.
          family: cross-repo-landing-steps
          round: 1
        - kind: defect
          id: PQ-3
          severity: Minor
          title: Chat follow-ups on bullets already linked to qa/ answer files aren't addressed
          detail: The chat follow-up rule should follow the existing link target rather than assume <name>.side/, or old bullets get duplicate answer docs.
          family: legacy-document-compat
          round: 1
        - kind: defect
          id: PQ-4
          severity: Minor
          title: Non-goals are scattered across the plan instead of stated in one place
          family: missing-non-goals
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: fd505325-6bd9-43ba-a86d-21535d607aa9
            repository: github.com/xianxu/ariadne
            issue_id: "000412"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 9a2d9f731f47c7a591f13ac7fcbba256824c44b3
            slot: ariadne:2
            requested_at: "2026-10-16T22:53:42.81272Z"
          phase: finished
          pid: 32479
          started_at: "2026-10-16T22:53:43.159054Z"
          finished_at: "2026-10-16T22:54:01.226396Z"
          session_id: c16f9f99-f7dc-493c-80e6-9950883fe1cf
          outcome: returned
      "n": 2
      timestamp: "2026-10-16T15:54:02-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Per-case before/ and after/ trees, each with its own .side/ folder, plus case.md; swept across the test list and tasks 3, 4 and 8.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Task 6 states the gitignore; the landing report carries the ops .gitignore entry for tl-status-*.side/.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: 'Task 6 and a chat contract edge: a follow-up extends the file the existing qa/ link names.'
          round: 2
        - id: PQ-4
          disposition: addressed
          note: A single Non-goals section now exists.
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: b8877626-22b2-49e8-83f8-2ab5ceda5c1d
            repository: github.com/xianxu/ariadne
            issue_id: "000412"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 9a2d9f731f47c7a591f13ac7fcbba256824c44b3
            slot: ariadne:2
            requested_at: "2026-10-16T22:55:48.996281Z"
          phase: finished
          pid: 65144
          started_at: "2026-10-16T22:55:49.29019Z"
          finished_at: "2026-10-16T22:56:08.47968Z"
          session_id: a3c26c68-b277-41fb-8571-456fa2883740
          outcome: returned
      "n": 3
      timestamp: "2026-10-16T15:56:09-07:00"
      agent: claude
      blocked: false
      protocol_error: no valid findings block
content_hash: ee3e02bc48ca3bdad170ab4b5194f5fc4f6d52527c842dc91a296b8e7f4014ab
---

# Gate ledger — ariadne#412 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-16T15:50:10-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `fixture-layout-unspecified` (kind: defect) The examples' side-doc layout contradicts itself (after.side/ vs before.side/ links) and can't represent before and after states of side docs
  Core concepts says after.side/, but rows 3 and 12 link before.side/. Rows 6, 7, 14, 17b and chat/retry need side docs in both a before and an after state, which one folder can't hold. Define per-state trees (e.g. before/doc.md + doc.side/, after/doc.md + doc.side/) so the task 4 checker and the task 8 live diff have one layout to resolve links against.
- **PQ-2** [Minor] `cross-repo-landing-steps` (kind: defect) Moving answer docs to tl-status-<op>.side/ doesn't mention gitignoring that folder in ops
  help.md:185-186 says QA files are gitignored like the status file; the new folder needs the same, or private answers get committed to ops. Add it to the TL's landing report.
- **PQ-3** [Minor] `legacy-document-compat` (kind: defect) Chat follow-ups on bullets already linked to qa/ answer files aren't addressed
  The chat follow-up rule should follow the existing link target rather than assume <name>.side/, or old bullets get duplicate answer docs.
- **PQ-4** [Minor] `missing-non-goals` (kind: defect) Non-goals are scattered across the plan instead of stated in one place

## Round 2 — 2026-10-16T15:54:02-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Per-case before/ and after/ trees, each with its own .side/ folder, plus case.md; swept across the test list and tasks 3, 4 and 8.
- PQ-2 — addressed — Task 6 states the gitignore; the landing report carries the ops .gitignore entry for tl-status-*.side/.
- PQ-3 — addressed — Task 6 and a chat contract edge: a follow-up extends the file the existing qa/ link names.
- PQ-4 — addressed — A single Non-goals section now exists.

## Round 3 — 2026-10-16T15:56:09-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
