---
gate: plan-quality
issue: 409
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 99146027-4a87-4045-911c-3d4ea32cfac1
            repository: github.com/xianxu/ariadne
            issue_id: "000409"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: a0fe884074803f72d22922b2ebad7dfc20f169ce
            slot: ariadne:2
            requested_at: "2026-10-18T12:21:12.989659Z"
          phase: finished
          pid: 14177
          started_at: "2026-10-18T12:21:13.140282Z"
          finished_at: "2026-10-18T12:22:33.554162Z"
          session_id: 5ca5160d-491c-4aec-b8f5-f5e40a69efb5
          outcome: returned
      "n": 1
      timestamp: "2026-10-18T05:22:33-07:00"
      agent: claude
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: The fallback's "could not run, never started" test doesn't match how dispatch classifies failures
          detail: RunFailure (apifailure.go:37) labels any run with no verdict plus is_error or a non-zero exit as ErrAPIUnreachable when the output matches the broad `API Error` regex, including a run that fails mid-review. Dispatch also swallows non-zero exits and returns the output (dispatch.go doc), so quota and auth text never reaches an `Unavailable(err)`. Define the "never started" evidence (for example, no assistant or tool events before the failure), give Unavailable the AgentRun and stderr as input, and state how it relates to RunFailure (ARCH-ORDER).
          family: unconfirmed-outcome-treated-as-confirmed
          round: 1
        - kind: defect
          id: PQ-2
          severity: Important
          title: ChangeKind is the review stage, not the spec's design-heavy vs bounded kind, and no rule uses it
          detail: Spec and Done-when require the policy to pick by change kind (new mechanism, state or invariant vs bounded). The plan lists stages, gives no default rule where kind changes the reviewer, and doesn't say how each site derives kind. "Per-kind overrides" is left empty (ARCH-PURPOSE).
          family: deferred-purpose
          round: 1
        - kind: defect
          id: PQ-3
          severity: Important
          title: 'Plan reverses the spec''s "static #310 policy is the floor" and adds `cross` next to the existing `other` mode without deciding `other`''s fate'
          detail: agent_resolver.go:47 already has `other` (claude and codex cross). State whether `other` is kept, made an alias of `cross`, or deprecated, and how Validate handles it, and record the spec change as a Revision. review.json is a cross-repo config contract (ARCH-DRY).
          family: unstated-contract-change
          round: 1
        - kind: defect
          id: PQ-4
          severity: Important
          title: Default cross sends claude-authored closes to codex, which milestoneclose.go:699 already says fails on darwin ContainCheckout, and the fallback won't skip it
          detail: That failure is a launched-but-failing run, not Unavailable, so the close halts instead of trying gemini. The condition is known before launch. Feed it to ReviewPlan or the dispatch loop as a recorded pre-launch skip.
          family: known-unavailability-not-prefiltered
          round: 1
        - kind: defect
          id: PQ-5
          severity: Important
          title: No stated non-goals; the spec's "over budget" fallback is silently dropped
          detail: 'Add an explicit non-goals list: budget or quota accounting (only reactive quota-error detection is built), doc-review''s hard-coded codex, kind auto-detection if deferred, each with a reason.'
          family: missing-non-goals
          round: 1
        - kind: defect
          id: PQ-6
          severity: Minor
          title: The ARCH-SECURE claim that forged Agent trailers can only steer the reviewer toward another family is false
          detail: Stamping a claude-authored window as codex makes cross pick claude, giving same-family review. Restate the trust assumption, or cross-check trailers against agentid.Detect.
          family: untrusted-input-claim
          round: 1
        - kind: defect
          id: PQ-7
          severity: Minor
          title: M2 lists fake-runner test cases one by one in prose; compress to one strategy line per risky function
          family: test-prose-enumeration
          round: 1
        - kind: defect
          id: PQ-8
          severity: Minor
          title: Done-when's "one week of trailers can be tallied" has no tally test or command; at least test that Review-Reviewer parses via git interpret-trailers
          family: done-when-unbacked
          round: 1
      blocked: true
---

# Gate ledger — ariadne#409 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T05:22:33-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `unconfirmed-outcome-treated-as-confirmed` (kind: defect) The fallback's "could not run, never started" test doesn't match how dispatch classifies failures
  RunFailure (apifailure.go:37) labels any run with no verdict plus is_error or a non-zero exit as ErrAPIUnreachable when the output matches the broad `API Error` regex, including a run that fails mid-review. Dispatch also swallows non-zero exits and returns the output (dispatch.go doc), so quota and auth text never reaches an `Unavailable(err)`. Define the "never started" evidence (for example, no assistant or tool events before the failure), give Unavailable the AgentRun and stderr as input, and state how it relates to RunFailure (ARCH-ORDER).
- **PQ-2** [Important] `deferred-purpose` (kind: defect) ChangeKind is the review stage, not the spec's design-heavy vs bounded kind, and no rule uses it
  Spec and Done-when require the policy to pick by change kind (new mechanism, state or invariant vs bounded). The plan lists stages, gives no default rule where kind changes the reviewer, and doesn't say how each site derives kind. "Per-kind overrides" is left empty (ARCH-PURPOSE).
- **PQ-3** [Important] `unstated-contract-change` (kind: defect) Plan reverses the spec's "static #310 policy is the floor" and adds `cross` next to the existing `other` mode without deciding `other`'s fate
  agent_resolver.go:47 already has `other` (claude and codex cross). State whether `other` is kept, made an alias of `cross`, or deprecated, and how Validate handles it, and record the spec change as a Revision. review.json is a cross-repo config contract (ARCH-DRY).
- **PQ-4** [Important] `known-unavailability-not-prefiltered` (kind: defect) Default cross sends claude-authored closes to codex, which milestoneclose.go:699 already says fails on darwin ContainCheckout, and the fallback won't skip it
  That failure is a launched-but-failing run, not Unavailable, so the close halts instead of trying gemini. The condition is known before launch. Feed it to ReviewPlan or the dispatch loop as a recorded pre-launch skip.
- **PQ-5** [Important] `missing-non-goals` (kind: defect) No stated non-goals; the spec's "over budget" fallback is silently dropped
  Add an explicit non-goals list: budget or quota accounting (only reactive quota-error detection is built), doc-review's hard-coded codex, kind auto-detection if deferred, each with a reason.
- **PQ-6** [Minor] `untrusted-input-claim` (kind: defect) The ARCH-SECURE claim that forged Agent trailers can only steer the reviewer toward another family is false
  Stamping a claude-authored window as codex makes cross pick claude, giving same-family review. Restate the trust assumption, or cross-check trailers against agentid.Detect.
- **PQ-7** [Minor] `test-prose-enumeration` (kind: defect) M2 lists fake-runner test cases one by one in prose; compress to one strategy line per risky function
- **PQ-8** [Minor] `done-when-unbacked` (kind: defect) Done-when's "one week of trailers can be tallied" has no tally test or command; at least test that Review-Reviewer parses via git interpret-trailers

## Open findings

- **PQ-1** [Important] `unconfirmed-outcome-treated-as-confirmed` (kind: defect) The fallback's "could not run, never started" test doesn't match how dispatch classifies failures
- **PQ-2** [Important] `deferred-purpose` (kind: defect) ChangeKind is the review stage, not the spec's design-heavy vs bounded kind, and no rule uses it
- **PQ-3** [Important] `unstated-contract-change` (kind: defect) Plan reverses the spec's "static #310 policy is the floor" and adds `cross` next to the existing `other` mode without deciding `other`'s fate
- **PQ-4** [Important] `known-unavailability-not-prefiltered` (kind: defect) Default cross sends claude-authored closes to codex, which milestoneclose.go:699 already says fails on darwin ContainCheckout, and the fallback won't skip it
- **PQ-5** [Important] `missing-non-goals` (kind: defect) No stated non-goals; the spec's "over budget" fallback is silently dropped
- **PQ-6** [Minor] `untrusted-input-claim` (kind: defect) The ARCH-SECURE claim that forged Agent trailers can only steer the reviewer toward another family is false
- **PQ-7** [Minor] `test-prose-enumeration` (kind: defect) M2 lists fake-runner test cases one by one in prose; compress to one strategy line per risky function
- **PQ-8** [Minor] `done-when-unbacked` (kind: defect) Done-when's "one week of trailers can be tallied" has no tally test or command; at least test that Review-Reviewer parses via git interpret-trailers
