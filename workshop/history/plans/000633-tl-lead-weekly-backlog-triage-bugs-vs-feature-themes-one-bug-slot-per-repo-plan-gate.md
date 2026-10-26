---
reviewer_runs:
    - version: 1
      intent:
        id: 5f07970a-7bbc-43c3-8e60-ee1d1cbba186
        repository: github.com/xianxu/ariadne
        issue_id: "000633"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        reviewed_head: 7fd62025049b182cbab147fb0c6a1b5b8a99fc0e
        slot: ariadne:2
        requested_at: "2026-10-25T19:32:03.981995Z"
      phase: finished
      pid: 58591
      started_at: "2026-10-25T19:32:04.088501Z"
      finished_at: "2026-10-25T19:32:32.393467Z"
      session_id: 46950beb-a8af-4466-8e06-61ffe0e7d0cf
      outcome: returned
gate: plan-quality
issue: 633
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 4909442c-74b0-49d6-af35-5cfa85109a8b
            repository: github.com/xianxu/ariadne
            issue_id: "000633"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: bcff521eb8ecb106c9744252a6e874537720ff58
            slot: ariadne:2
            requested_at: "2026-10-25T19:29:21.801955Z"
          phase: finished
          pid: 5019
          started_at: "2026-10-25T19:29:22.048643Z"
          finished_at: "2026-10-25T19:30:28.335233Z"
          session_id: bafa41bc-2d32-4fb7-8d45-01e11c243db0
          outcome: returned
      "n": 1
      timestamp: "2026-10-25T12:30:28-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: PQ-1
          severity: Minor
          title: Back-fill relies on set-kind never bumping `updated`; state it and test it
          detail: updated is written only by close/status paths today (close.go:811, trackercompletion.go:151, completeop.go:118, externalmerge.go:417), not by newCardSetterCmd. The TL back-fill runs set-kind on closed cards, so a future setter that bumps updated would move every back-filled close into the back-fill week. Add a set-kind test asserting updated is unchanged.
          family: unstated-dependency-invariant
          round: 1
        - kind: defect
          id: PQ-2
          severity: Minor
          title: issue list already reads through listIssueStates, which drops terminal cards; legacy repos have no close history
          detail: runIssueList calls listIssueStates (issue.go ~683), the same path as state, which skips terminal card-only records (state.go:371-374), so --json must walk the record set directly. The table labels issue_list.go both new and modified. Legacy repos keep closed issues in workshop/history, so stats there would show no closes; say so.
          family: unbacked-existing-behavior-claim
          round: 1
        - kind: defect
          id: PQ-3
          severity: Minor
          title: M2 test bullets list cases; compress to one strategy line plus a stated open-at-week-end definition
          detail: Replace the case list with a generated-input property test (DST and ISO-year boundaries; opened minus closed reconciles with the change in open count) and define open-at-week-end as created <= end and (non-terminal or updated > end), noting that reopened cards make it approximate.
          family: test-prose-enumeration
          round: 1
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 67c04540-9f88-4f75-afb0-e1af52999372
            repository: github.com/xianxu/ariadne
            issue_id: "000633"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 7fd62025049b182cbab147fb0c6a1b5b8a99fc0e
            slot: ariadne:2
            requested_at: "2026-10-25T19:31:46.865062Z"
          phase: finished
          pid: 50616
          started_at: "2026-10-25T19:31:46.988645Z"
          finished_at: "2026-10-25T19:32:01.49535Z"
          session_id: f4ea7e42-dbb7-4d7d-8b21-a743762c93a8
          outcome: returned
      "n": 2
      timestamp: "2026-10-25T12:32:01-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: 'Revisions round 1: set-kind tests assert updated unchanged on a done card; the close/status-only writers are cited.'
          round: 2
        - id: PQ-2
          disposition: addressed
          note: 'Revisions round 1: --json walks the record set directly with a done card-only test; issue_list.go is new; legacy repos report "no tracker: open counts only".'
          round: 2
        - id: PQ-3
          disposition: addressed
          note: 'Revisions round 1: one generated-input property strategy (DST, ISO year, opened-closed reconciles with the change in open count) and an explicit open-at-week-end definition, approximate under reopens.'
          round: 2
      blocked: false
content_hash: 2e4bde14f6a062afe970710aa79e87d6aa46d556a87403e867649e9561d79827
---

# Gate ledger — ariadne#633 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-25T12:30:28-07:00 (claude) — passed

### Raised

- **PQ-1** [Minor] `unstated-dependency-invariant` (kind: defect) Back-fill relies on set-kind never bumping `updated`; state it and test it
  updated is written only by close/status paths today (close.go:811, trackercompletion.go:151, completeop.go:118, externalmerge.go:417), not by newCardSetterCmd. The TL back-fill runs set-kind on closed cards, so a future setter that bumps updated would move every back-filled close into the back-fill week. Add a set-kind test asserting updated is unchanged.
- **PQ-2** [Minor] `unbacked-existing-behavior-claim` (kind: defect) issue list already reads through listIssueStates, which drops terminal cards; legacy repos have no close history
  runIssueList calls listIssueStates (issue.go ~683), the same path as state, which skips terminal card-only records (state.go:371-374), so --json must walk the record set directly. The table labels issue_list.go both new and modified. Legacy repos keep closed issues in workshop/history, so stats there would show no closes; say so.
- **PQ-3** [Minor] `test-prose-enumeration` (kind: defect) M2 test bullets list cases; compress to one strategy line plus a stated open-at-week-end definition
  Replace the case list with a generated-input property test (DST and ISO-year boundaries; opened minus closed reconciles with the change in open count) and define open-at-week-end as created <= end and (non-terminal or updated > end), noting that reopened cards make it approximate.

## Round 2 — 2026-10-25T12:32:01-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Revisions round 1: set-kind tests assert updated unchanged on a done card; the close/status-only writers are cited.
- PQ-2 — addressed — Revisions round 1: --json walks the record set directly with a done card-only test; issue_list.go is new; legacy repos report "no tracker: open counts only".
- PQ-3 — addressed — Revisions round 1: one generated-input property strategy (DST, ISO year, opened-closed reconciles with the change in open count) and an explicit open-at-week-end definition, approximate under reopens.

## Open findings

(none — every finding has been disposed)
