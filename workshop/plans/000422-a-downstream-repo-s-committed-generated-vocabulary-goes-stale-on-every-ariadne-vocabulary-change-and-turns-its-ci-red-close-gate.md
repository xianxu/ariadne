---
gate: boundary-review
issue: 422
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ca8c0394-73c6-4091-8443-b054b3fd9eb8
            repository: github.com/xianxu/ariadne
            issue_id: "000422"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/b900e5bf7e217d3a
            reviewed_head: 62afe6f15e4672de6f38d1ba55a22c683bdcf24d
            slot: ariadne:4
            requested_at: "2026-10-17T09:27:34.140536Z"
          phase: finished
          pid: 61921
          started_at: "2026-10-17T09:27:34.234664Z"
          finished_at: "2026-10-17T09:33:05.626077Z"
          session_id: 3b982b31-71b9-4769-ac3e-de6df77ca595
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 62afe6f15e4672de6f38d1ba55a22c683bdcf24d
        tracker_head: 71054ab21437a764cac99f8d03432c483c831c90
        branch_point: 18a4476aa996c36a8df7d647c750e3c130ab2b57
        since: "2026-10-17T01:34:29-07:00"
        until: "2026-10-17T02:27:19-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T01:34:29-07:00"
              end: "2026-10-17T02:27:19-07:00"
              began: "2026-10-17T01:34:29-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: f560be0f4446c047c041b28773a1fb68f797f0036b59f210d1f45056e37b2019
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 29010375
              windows:
                - start: "2026-10-17T01:34:29-07:00"
                  end: "2026-10-17T02:27:19-07:00"
      "n": 1
      timestamp: "2026-10-17T02:33:08-07:00"
      dispatched: "2026-10-17T02:27:34-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: The non-regular sidecar/output branch of vouched (symlink or directory) has no test
          detail: vouched returns false for a non-regular Lstat result, but no TestShippedOutputOwnership row enters that branch. Add a row where the sidecar is a symlink to a matching digest, expecting a cold refusal.
          family: untrusted-input-branch-untested
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: Plan Core concepts gives ShippedPaths as walk.Layer returning []string; the code takes layer.Layer and returns ([]string, error)
          detail: Add a Revisions entry, and tick the plan file's task checkboxes (still unchecked although the issue checklist is complete).
          family: plan-signature-drift
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: The ship-row warning says "names no generated output" for a row naming a staged symlink output or an exported ancestor row
          detail: ShipStaged only matches staged WriteFile outputs, so a symlink output, or an export row reaching a consumer that doesn't generate it, gets a misleading warning. Consider naming the reason.
          family: misleading-diagnostic
          round: 1
      recipe: milestone-review
      isolation: checkout
      reviewed: 62afe6f15e4672de6f38d1ba55a22c683bdcf24d
      blocked: false
---

# Gate ledger — ariadne#422 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T02:33:08-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `untrusted-input-branch-untested` (kind: defect) The non-regular sidecar/output branch of vouched (symlink or directory) has no test
  vouched returns false for a non-regular Lstat result, but no TestShippedOutputOwnership row enters that branch. Add a row where the sidecar is a symlink to a matching digest, expecting a cold refusal.
- **BR-2** [Minor] `plan-signature-drift` (kind: defect) Plan Core concepts gives ShippedPaths as walk.Layer returning []string; the code takes layer.Layer and returns ([]string, error)
  Add a Revisions entry, and tick the plan file's task checkboxes (still unchecked although the issue checklist is complete).
- **BR-3** [Minor] `misleading-diagnostic` (kind: defect) The ship-row warning says "names no generated output" for a row naming a staged symlink output or an exported ancestor row
  ShipStaged only matches staged WriteFile outputs, so a symlink output, or an export row reaching a consumer that doesn't generate it, gets a misleading warning. Consider naming the reason.

## Open findings

- **BR-1** [Minor] `untrusted-input-branch-untested` (kind: defect) The non-regular sidecar/output branch of vouched (symlink or directory) has no test
- **BR-2** [Minor] `plan-signature-drift` (kind: defect) Plan Core concepts gives ShippedPaths as walk.Layer returning []string; the code takes layer.Layer and returns ([]string, error)
- **BR-3** [Minor] `misleading-diagnostic` (kind: defect) The ship-row warning says "names no generated output" for a row naming a staged symlink output or an exported ancestor row
