---
reviewer_runs:
    - version: 1
      intent:
        id: 9f2e28f4-eead-4936-b663-93d2dd7c462a
        repository: github.com/xianxu/ariadne
        issue_id: "000422"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        reviewed_head: fca84162a33a703e6d2fde9459bfc156ef78c24b
        slot: ariadne:4
        requested_at: "2026-10-17T08:46:05.357848Z"
      phase: finished
      pid: 38669
      started_at: "2026-10-17T08:46:05.46862Z"
      finished_at: "2026-10-17T08:46:30.471491Z"
      session_id: 9a6cd143-f8a3-43cc-8aad-ca8e76d94949
      outcome: returned
gate: plan-quality
issue: 422
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 5fcc21c4-30d0-4b1d-b472-b343beaaccf7
            repository: github.com/xianxu/ariadne
            issue_id: "000422"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: 3e777703b0512e32b5c5015f436783761dbedb75
            slot: ariadne:4
            requested_at: "2026-10-17T08:42:38.683939Z"
          phase: finished
          pid: 6459
          started_at: "2026-10-17T08:42:38.890676Z"
          finished_at: "2026-10-17T08:43:52.821238Z"
          session_id: 25bc77d8-eb41-47d9-9d3c-ab8b15dad67e
          outcome: returned
      "n": 1
      timestamp: "2026-10-17T01:43:53-07:00"
      agent: claude
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: Warm pulled-copy case displaces the recorded sidecar, contradicting the planned Displaced-empty assertion
          detail: The sidecar is inventoried after the first apply; a git pull that changes it hits ownership.go:360 (recorded and not a dir, so it is displaced) because the plan makes it a plain non-strict output with no vouching rule. Make a sidecar count as weave's own when its paired declared output vouches, or change the test expectation.
          family: paired-artifact-ownership
          round: 1
        - kind: defect
          id: PQ-2
          severity: Minor
          title: 'Gitignore handling decided twice: Task 4 keep set vs the issue''s parley-side negation line'
          detail: managedIgnoreText already lets authored negations override the block (managed_ignore.go:60-61) and parley already uses one for issue.json; choose either the keep set or the negation line, not both.
          family: single-mechanism-per-decision
          round: 1
        - kind: defect
          id: PQ-3
          severity: Minor
          title: Sidecar retirement claim holds only for warm checkouts; label is ARCH-FUNERAL, not ARCH-CONSTRAINTS
          detail: On a cold checkout the sidecar is not inventoried, so dropping the ship row leaves a committed sidecar until someone runs git rm; state that.
          family: funeral-claim-scope
          round: 1
        - kind: defect
          id: PQ-4
          severity: Minor
          title: Tasks 2, 3 and 5 enumerate test cases in prose instead of one strategy line per risky function
          family: plan-test-prose-enumeration
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: d519022e-eee7-4b8f-a700-9df3785582dd
            repository: github.com/xianxu/ariadne
            issue_id: "000422"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: 8c0c2f152f6fc6775cbcda46c14ebde367254382
            slot: ariadne:4
            requested_at: "2026-10-17T08:44:35.452959Z"
          phase: finished
          pid: 12454
          started_at: "2026-10-17T08:44:35.640699Z"
          finished_at: "2026-10-17T08:45:07.406516Z"
          session_id: 7f0b55b7-2b4b-4bcc-a7d7-071cac906664
          outcome: returned
      "n": 2
      timestamp: "2026-10-17T01:45:07-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: The pair rule makes both the output and the sidecar own when the pair matches, so the recorded-and-changed displacement at ownership.go:360 no longer fires for a pulled refresh.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Keep set dropped; only the authored-negation mechanism remains.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Removal scoped to warm checkouts, git rm stated for cold ones, filed under ARCH-FUNERAL.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Tasks 2-4 now give one strategy line (matrix plus guard) each.
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: da35a172-665e-4793-82f5-05200ca3c710
            repository: github.com/xianxu/ariadne
            issue_id: "000422"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: fca84162a33a703e6d2fde9459bfc156ef78c24b
            slot: ariadne:4
            requested_at: "2026-10-17T08:45:49.871687Z"
          phase: finished
          pid: 33855
          started_at: "2026-10-17T08:45:49.977761Z"
          finished_at: "2026-10-17T08:46:04.097106Z"
          session_id: f45f86f1-ade1-4003-8a2b-d4bbc8c634c1
          outcome: returned
      "n": 3
      timestamp: "2026-10-17T01:46:04-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Pair rule makes weave own both the output and its sidecar; the pulled-copy row asserts Displaced is empty.
          round: 3
        - id: PQ-2
          disposition: addressed
          note: Keep set dropped; the existing authored-negation mechanism is the only one.
          round: 3
        - id: PQ-3
          disposition: addressed
          note: Removal scoped to warm checkouts, under ARCH-FUNERAL, with the cold git rm duty.
          round: 3
        - id: PQ-4
          disposition: addressed
          note: Tasks 2 to 4 use one strategy line each.
          round: 3
      blocked: false
content_hash: e5158e5540b7a2f35732e2f0b37642aaae740caa25ea51fb36c64d2d64166570
---

# Gate ledger — ariadne#422 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T01:43:53-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `paired-artifact-ownership` (kind: defect) Warm pulled-copy case displaces the recorded sidecar, contradicting the planned Displaced-empty assertion
  The sidecar is inventoried after the first apply; a git pull that changes it hits ownership.go:360 (recorded and not a dir, so it is displaced) because the plan makes it a plain non-strict output with no vouching rule. Make a sidecar count as weave's own when its paired declared output vouches, or change the test expectation.
- **PQ-2** [Minor] `single-mechanism-per-decision` (kind: defect) Gitignore handling decided twice: Task 4 keep set vs the issue's parley-side negation line
  managedIgnoreText already lets authored negations override the block (managed_ignore.go:60-61) and parley already uses one for issue.json; choose either the keep set or the negation line, not both.
- **PQ-3** [Minor] `funeral-claim-scope` (kind: defect) Sidecar retirement claim holds only for warm checkouts; label is ARCH-FUNERAL, not ARCH-CONSTRAINTS
  On a cold checkout the sidecar is not inventoried, so dropping the ship row leaves a committed sidecar until someone runs git rm; state that.
- **PQ-4** [Minor] `plan-test-prose-enumeration` (kind: defect) Tasks 2, 3 and 5 enumerate test cases in prose instead of one strategy line per risky function

## Round 2 — 2026-10-17T01:45:07-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — The pair rule makes both the output and the sidecar own when the pair matches, so the recorded-and-changed displacement at ownership.go:360 no longer fires for a pulled refresh.
- PQ-2 — addressed — Keep set dropped; only the authored-negation mechanism remains.
- PQ-3 — addressed — Removal scoped to warm checkouts, git rm stated for cold ones, filed under ARCH-FUNERAL.
- PQ-4 — addressed — Tasks 2-4 now give one strategy line (matrix plus guard) each.

## Round 3 — 2026-10-17T01:46:04-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Pair rule makes weave own both the output and its sidecar; the pulled-copy row asserts Displaced is empty.
- PQ-2 — addressed — Keep set dropped; the existing authored-negation mechanism is the only one.
- PQ-3 — addressed — Removal scoped to warm checkouts, under ARCH-FUNERAL, with the cold git rm duty.
- PQ-4 — addressed — Tasks 2 to 4 use one strategy line each.

## Open findings

(none — every finding has been disposed)
