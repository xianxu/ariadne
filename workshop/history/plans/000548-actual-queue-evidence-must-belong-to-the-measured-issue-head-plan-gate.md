---
reviewer_runs:
    - version: 1
      intent:
        id: 881e3734-e3e0-4db7-9c89-f028b1269e00
        repository: github.com/xianxu/ariadne
        issue_id: "000548"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: codex
        root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        reviewed_head: 549a055ab91a51ec76f2e43cf34424c91f5088fe
        slot: ariadne:1
        requested_at: "2026-10-22T10:49:42.980106Z"
      phase: finished
      pid: 38444
      started_at: "2026-10-22T10:49:43.090151Z"
      finished_at: "2026-10-22T10:50:12.101803Z"
      session_id: 01a1633b-26e2-7b72-a1b9-a13cb5613d82
      outcome: returned
gate: plan-quality
issue: 548
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 10578df4-f9c2-4e4b-9b30-fb6b8e797904
            repository: github.com/xianxu/ariadne
            issue_id: "000548"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: 509ec9e25248eb478c5d45abea40bc9b71bccd1a
            slot: ariadne:1
            requested_at: "2026-10-22T10:43:26.147005Z"
          phase: finished
          pid: 97227
          started_at: "2026-10-22T10:43:26.278346Z"
          finished_at: "2026-10-22T10:44:29.504022Z"
          session_id: 01a16335-672a-7933-b4ea-2925ff2e5217
          outcome: returned
      "n": 1
      timestamp: "2026-10-22T03:44:29-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Disable Git replacement objects when reading immutable attribution evidence
          detail: 'ARCH-SECURE: The plan requires exact immutable OID authority, but its proposed cat-file reader reuses runGitBoundedInputContext, which forwards the requested Git invocation without disabling replacement objects (cmd/sdlc/internal/gitx/boundedoutput.go:42–53). Git replacement refs can substitute a commit''s message and parents while the batch header still identifies the requested OID, defeating the planned identity/frame checks. Require --no-replace-objects or an equivalent enforced setting at this reader boundary, and verify with a real replacement-ref fixture that membership remains bound to the original object.'
          family: immutable-object-read-authority
          round: 1
        - id: PQ-2
          severity: Important
          title: Specify aggregate traversal and retained-memory limits
          detail: 'ARCH-CONSTRAINTS: “One bounded history budget per unique lineage” and a “complete-history bound/sentinel” leave the actual envelope undefined. The existing store admits up to 10,000 attempts (cmd/sdlc/teststats_cmd.go:19), while command-local memoization retains facts across batches; a per-batch output cap does not bound total retained bytes or total traversal work. State concrete command-wide node/byte limits and a per-lineage depth limit, their initial basis, and the named refusal on exhaustion; exercise those limits through readQueueHeadMemberships.'
          family: explicit-command-resource-envelope
          round: 1
        - id: PQ-3
          severity: Minor
          title: Remove the duplicated prose test-case inventory
          detail: The “Verify direct heads, neutral metadata/integration heads…” checkbox in the implementation section enumerates cases already covered by the named function strategy table. Replace that inventory with references to the function contracts; keep concrete cases in executable tests.
          family: concise-function-test-strategies
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 4aa52137-70dd-406a-bf57-8acfdc067c14
            repository: github.com/xianxu/ariadne
            issue_id: "000548"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: 549a055ab91a51ec76f2e43cf34424c91f5088fe
            slot: ariadne:1
            requested_at: "2026-10-22T10:47:56.44846Z"
          phase: finished
          pid: 50697
          started_at: "2026-10-22T10:47:56.617499Z"
          finished_at: "2026-10-22T10:48:23.429776Z"
          session_id: 01a16339-875a-7873-833a-823b635425ec
          outcome: returned
      "n": 2
      timestamp: "2026-10-22T03:48:23-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: The narrow reader disables replacement objects for every attribution read and requires a real replacement-ref regression.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Concrete aggregate node/byte and lineage-depth ceilings include accounting rules, named refusals, and production-walker boundary tests.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: The verification checklist references named function contracts instead of repeating the prose case inventory.
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: ee8b5081-82f3-447a-a0f4-5221428726ce
            repository: github.com/xianxu/ariadne
            issue_id: "000548"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: 549a055ab91a51ec76f2e43cf34424c91f5088fe
            slot: ariadne:1
            requested_at: "2026-10-22T10:49:05.808867Z"
          phase: finished
          pid: 7205
          started_at: "2026-10-22T10:49:05.926325Z"
          finished_at: "2026-10-22T10:49:41.185608Z"
          session_id: 01a1633a-95d7-7431-bc8d-a693237878ce
          outcome: returned
      "n": 3
      timestamp: "2026-10-22T03:49:41-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      blocked: false
      protocol_error: no valid findings block
content_hash: b64a92e0735947bc3e541c13a08f26867c113a5758907f67a192479376fd3632
---

# Gate ledger — ariadne#548 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T03:44:29-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `immutable-object-read-authority` Disable Git replacement objects when reading immutable attribution evidence
  ARCH-SECURE: The plan requires exact immutable OID authority, but its proposed cat-file reader reuses runGitBoundedInputContext, which forwards the requested Git invocation without disabling replacement objects (cmd/sdlc/internal/gitx/boundedoutput.go:42–53). Git replacement refs can substitute a commit's message and parents while the batch header still identifies the requested OID, defeating the planned identity/frame checks. Require --no-replace-objects or an equivalent enforced setting at this reader boundary, and verify with a real replacement-ref fixture that membership remains bound to the original object.
- **PQ-2** [Important] `explicit-command-resource-envelope` Specify aggregate traversal and retained-memory limits
  ARCH-CONSTRAINTS: “One bounded history budget per unique lineage” and a “complete-history bound/sentinel” leave the actual envelope undefined. The existing store admits up to 10,000 attempts (cmd/sdlc/teststats_cmd.go:19), while command-local memoization retains facts across batches; a per-batch output cap does not bound total retained bytes or total traversal work. State concrete command-wide node/byte limits and a per-lineage depth limit, their initial basis, and the named refusal on exhaustion; exercise those limits through readQueueHeadMemberships.
- **PQ-3** [Minor] `concise-function-test-strategies` Remove the duplicated prose test-case inventory
  The “Verify direct heads, neutral metadata/integration heads…” checkbox in the implementation section enumerates cases already covered by the named function strategy table. Replace that inventory with references to the function contracts; keep concrete cases in executable tests.

## Round 2 — 2026-10-22T03:48:23-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — The narrow reader disables replacement objects for every attribution read and requires a real replacement-ref regression.
- PQ-2 — addressed — Concrete aggregate node/byte and lineage-depth ceilings include accounting rules, named refusals, and production-walker boundary tests.
- PQ-3 — addressed — The verification checklist references named function contracts instead of repeating the prose case inventory.

## Round 3 — 2026-10-22T03:49:41-07:00 (codex) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
