---
gate: plan-quality
issue: 441
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: dbeab6d0-b03a-4793-8f4b-c375306adb8b
            repository: github.com/xianxu/ariadne
            issue_id: "000441"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 7800e68d27d767dc49b7dc9decae4e850bd72036
            slot: ariadne:2
            requested_at: "2026-10-18T16:13:06.049263Z"
          phase: finished
          pid: 41968
          started_at: "2026-10-18T16:13:06.183784Z"
          finished_at: "2026-10-18T16:13:55.551121Z"
          session_id: 4945cf1c-b2de-4b88-8487-960125262396
          outcome: returned
      "n": 1
      timestamp: "2026-10-18T09:13:55-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: Session-agent detection and Agent-trailer parsing live in cmd/sdlc/internal/agentid; the plan never says how doc-review gets them
          detail: FamilyOf's constants (dispatch.go:29 agentid.Claude), MainAgent (agent_resolver.go:19) and authorFamiliesFromLog (cmd/sdlc/reviewchoice.go:39, agentid.LogFormat/ParseLog) all depend on agentid, which doc-review cannot import. M2 hand-rolls a trailer format (duplication) and names no session-agent source, yet never-self-review rests on the session family being among the authors. State the move (agentid env/trailer parsing to pkg/) or the substitute, and the behavior when the session agent is unknown.
          family: internal-boundary-strands-dependency
          round: 1
        - kind: defect
          id: PQ-2
          severity: Minor
          title: ResolveReview's forced path (explicit arg, AGENT_CMD, Forced for the record) is not extracted; chooseReviewers will re-implement it
          family: shared-policy-incomplete-extraction
          round: 1
        - kind: defect
          id: PQ-3
          severity: Minor
          title: Spec makes the codex-crash hint an automatic fallback; the plan falls back only on a missing CLI or failed launch
          detail: A codex crash is a non-zero run (review.go:97-105), not a launch failure. Pick one rule and say how launch failure is told apart under CombinedOutput.
          family: spec-plan-contradiction
          round: 1
        - kind: defect
          id: PQ-4
          severity: Minor
          title: reportPath (review.go:56) and codex -o capture are keyed on the agent; with fallback they must derive from the candidate actually run
          family: fallback-derived-state
          round: 1
        - kind: defect
          id: PQ-5
          severity: Minor
          title: M2 enumerates six test cases in prose; compress to the function plus a property-style strategy line
          family: test-prose-enumeration
          round: 1
      blocked: true
---

# Gate ledger — ariadne#441 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T09:13:55-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `internal-boundary-strands-dependency` (kind: defect) Session-agent detection and Agent-trailer parsing live in cmd/sdlc/internal/agentid; the plan never says how doc-review gets them
  FamilyOf's constants (dispatch.go:29 agentid.Claude), MainAgent (agent_resolver.go:19) and authorFamiliesFromLog (cmd/sdlc/reviewchoice.go:39, agentid.LogFormat/ParseLog) all depend on agentid, which doc-review cannot import. M2 hand-rolls a trailer format (duplication) and names no session-agent source, yet never-self-review rests on the session family being among the authors. State the move (agentid env/trailer parsing to pkg/) or the substitute, and the behavior when the session agent is unknown.
- **PQ-2** [Minor] `shared-policy-incomplete-extraction` (kind: defect) ResolveReview's forced path (explicit arg, AGENT_CMD, Forced for the record) is not extracted; chooseReviewers will re-implement it
- **PQ-3** [Minor] `spec-plan-contradiction` (kind: defect) Spec makes the codex-crash hint an automatic fallback; the plan falls back only on a missing CLI or failed launch
  A codex crash is a non-zero run (review.go:97-105), not a launch failure. Pick one rule and say how launch failure is told apart under CombinedOutput.
- **PQ-4** [Minor] `fallback-derived-state` (kind: defect) reportPath (review.go:56) and codex -o capture are keyed on the agent; with fallback they must derive from the candidate actually run
- **PQ-5** [Minor] `test-prose-enumeration` (kind: defect) M2 enumerates six test cases in prose; compress to the function plus a property-style strategy line

## Open findings

- **PQ-1** [Important] `internal-boundary-strands-dependency` (kind: defect) Session-agent detection and Agent-trailer parsing live in cmd/sdlc/internal/agentid; the plan never says how doc-review gets them
- **PQ-2** [Minor] `shared-policy-incomplete-extraction` (kind: defect) ResolveReview's forced path (explicit arg, AGENT_CMD, Forced for the record) is not extracted; chooseReviewers will re-implement it
- **PQ-3** [Minor] `spec-plan-contradiction` (kind: defect) Spec makes the codex-crash hint an automatic fallback; the plan falls back only on a missing CLI or failed launch
- **PQ-4** [Minor] `fallback-derived-state` (kind: defect) reportPath (review.go:56) and codex -o capture are keyed on the agent; with fallback they must derive from the candidate actually run
- **PQ-5** [Minor] `test-prose-enumeration` (kind: defect) M2 enumerates six test cases in prose; compress to the function plus a property-style strategy line
