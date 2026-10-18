---
reviewer_runs:
    - version: 1
      intent:
        id: f357cc04-ceac-42ca-bf95-711c9727c5b2
        repository: github.com/xianxu/ariadne
        issue_id: "000409"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        reviewed_head: a473a96cd0c9381ed0e45d3bf47c8879b49e7a0a
        slot: ariadne:2
        requested_at: "2026-10-18T12:25:41.24688Z"
      phase: finished
      pid: 20056
      started_at: "2026-10-18T12:25:41.375837Z"
      finished_at: "2026-10-18T12:26:10.365991Z"
      session_id: 946b723a-6614-4e42-8189-fff378170870
      outcome: returned
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
    - reviewer_runs:
        - version: 1
          intent:
            id: b349ca88-6c16-40d2-a92c-0080d007b2ae
            repository: github.com/xianxu/ariadne
            issue_id: "000409"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 62a097b594a8dc8dfd7cb7c076f107684e6515b4
            slot: ariadne:2
            requested_at: "2026-10-18T12:23:48.66276Z"
          phase: finished
          pid: 4267
          started_at: "2026-10-18T12:23:48.796479Z"
          finished_at: "2026-10-18T12:24:09.459244Z"
          session_id: 6be380fe-6c30-4408-b5b9-026237f91e30
          outcome: returned
      "n": 2
      timestamp: "2026-10-18T05:24:09-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Unavailable(run, stderr, launchErr) requires no assistant or tool event plus a pattern match; a stream that can't be parsed counts as started; RunFailure unchanged and its relation stated.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Kind is design-heavy or bounded, taken from the plan/flow decision; the default order differs by kind and can be overridden per kind.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: No cross mode; other is generalized and backward compatible; the default stays same behind one constant; recorded as a Revision.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: ReviewPlan takes an available(agent) predicate; codex under ContainCheckout on darwin and binaries missing from PATH are skipped before launch, with the reason recorded.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: 'Non-goals list added: quota accounting, doc-review (#441), semantic kind detection, a vendor quietly degrading its model.'
          round: 2
        - id: PQ-6
          disposition: addressed
          note: Author families are the union of the trailers and the closer's agentid.Detect, so a forged trailer can't remove the closer's own family.
          round: 2
        - id: PQ-7
          disposition: addressed
          note: Compressed to one strategy line per risky function in the Revisions.
          round: 2
        - id: PQ-8
          disposition: addressed
          note: Fixed trailer format, a test that parses it via git log trailers, and a tally helper.
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 02678ba6-1a49-46b4-8b04-feff6ccc2f2e
            repository: github.com/xianxu/ariadne
            issue_id: "000409"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: a473a96cd0c9381ed0e45d3bf47c8879b49e7a0a
            slot: ariadne:2
            requested_at: "2026-10-18T12:25:26.663011Z"
          phase: finished
          pid: 17399
          started_at: "2026-10-18T12:25:26.76627Z"
          finished_at: "2026-10-18T12:25:39.955954Z"
          session_id: 144a7ee5-cc37-41c5-ac73-b88bbf5223ba
          outcome: returned
      "n": 3
      timestamp: "2026-10-18T05:25:40-07:00"
      agent: claude
      blocked: false
      protocol_error: no valid findings block
content_hash: f6132a9a0331bdfee3d5b308bbedfd8e17cf739058bea7b9bd6fcd3ff33714bb
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

## Round 2 — 2026-10-18T05:24:09-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Unavailable(run, stderr, launchErr) requires no assistant or tool event plus a pattern match; a stream that can't be parsed counts as started; RunFailure unchanged and its relation stated.
- PQ-2 — addressed — Kind is design-heavy or bounded, taken from the plan/flow decision; the default order differs by kind and can be overridden per kind.
- PQ-3 — addressed — No cross mode; other is generalized and backward compatible; the default stays same behind one constant; recorded as a Revision.
- PQ-4 — addressed — ReviewPlan takes an available(agent) predicate; codex under ContainCheckout on darwin and binaries missing from PATH are skipped before launch, with the reason recorded.
- PQ-5 — addressed — Non-goals list added: quota accounting, doc-review (#441), semantic kind detection, a vendor quietly degrading its model.
- PQ-6 — addressed — Author families are the union of the trailers and the closer's agentid.Detect, so a forged trailer can't remove the closer's own family.
- PQ-7 — addressed — Compressed to one strategy line per risky function in the Revisions.
- PQ-8 — addressed — Fixed trailer format, a test that parses it via git log trailers, and a tally helper.

## Round 3 — 2026-10-18T05:25:40-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
