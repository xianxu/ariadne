---
gate: boundary-review
issue: 371
id_prefix: BR
rounds:
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 4ad181aafb45810d42611e4881e3245a0801d1af
        tracker_head: 78c07eb085600bc8f021d817e3603367d051a3e9
        branch_point: 0820d860ceb54656f8a7c6cab6f209e2b6a65729
        since: "2026-10-15T00:11:03-07:00"
        until: "2026-10-15T00:18:39-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T00:11:03-07:00"
              end: "2026-10-15T00:18:39-07:00"
              began: "2026-10-15T00:11:03-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: fce81c2c48c8e6a0033fd4f2617950e643107a19f94419d52f0cb71835ab548a
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 8011437
              windows:
                - start: "2026-10-15T00:11:03-07:00"
                  end: "2026-10-15T00:18:39-07:00"
      "n": 1
      timestamp: "2026-10-15T01:32:27-07:00"
      dispatched: "2026-10-15T01:29:22-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: 'helptext/issue.md: the start-plan parenthetical now attaches to the --no-claim sentence'
          detail: The inserted --details/--no-claim sentences separated "(`start-plan` branches from fresh main)" from "next main merge", which it qualifies. Move the parenthetical back after "merge".
          family: doc-edit-displaces-context
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: No test covers --details on the legacy (untracked) issue new path
          detail: legacymode.go passes Sections to Render, but no test checks the body there; a legacy-mode fixture asserting the supplied Spec appears would pin it.
          family: test-coverage-parallel-path
          round: 1
      recipe: milestone-review
      isolation: fenced
      reviewed: 4ad181aafb45810d42611e4881e3245a0801d1af
      blocked: false
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 4ad181aafb45810d42611e4881e3245a0801d1af
        tracker_head: 78c07eb085600bc8f021d817e3603367d051a3e9
        branch_point: 0820d860ceb54656f8a7c6cab6f209e2b6a65729
        since: "2026-10-15T00:11:03-07:00"
        until: "2026-10-15T00:18:39-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T00:11:03-07:00"
              end: "2026-10-15T00:18:39-07:00"
              began: "2026-10-15T00:11:03-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: fce81c2c48c8e6a0033fd4f2617950e643107a19f94419d52f0cb71835ab548a
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 8011437
              windows:
                - start: "2026-10-15T00:11:03-07:00"
                  end: "2026-10-15T00:18:39-07:00"
      "n": 2
      timestamp: "2026-10-15T01:36:36-07:00"
      dispatched: "2026-10-15T01:34:48-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: helptext/issue.md now reads "next main merge (`start-plan` branches from fresh main)" with the --details/--no-claim sentences after it.
          round: 2
        - id: BR-2
          disposition: addressed
          note: TestLegacyRepositoryIssueNewDetails asserts the supplied Spec and Done-when, and it goes red when legacymode.go:157 Sections pass-through is reverted.
          round: 2
      recipe: fix-review
      isolation: fenced
      reviewed: ff2f38fb068d73f3a209d221832de93271b13f70
      blocked: false
---

# Gate ledger — ariadne#371 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-15T01:32:27-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `doc-edit-displaces-context` (kind: defect) helptext/issue.md: the start-plan parenthetical now attaches to the --no-claim sentence
  The inserted --details/--no-claim sentences separated "(`start-plan` branches from fresh main)" from "next main merge", which it qualifies. Move the parenthetical back after "merge".
- **BR-2** [Minor] `test-coverage-parallel-path` (kind: defect) No test covers --details on the legacy (untracked) issue new path
  legacymode.go passes Sections to Render, but no test checks the body there; a legacy-mode fixture asserting the supplied Spec appears would pin it.

## Round 2 — 2026-10-15T01:36:36-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — helptext/issue.md now reads "next main merge (`start-plan` branches from fresh main)" with the --details/--no-claim sentences after it.
- BR-2 — addressed — TestLegacyRepositoryIssueNewDetails asserts the supplied Spec and Done-when, and it goes red when legacymode.go:157 Sections pass-through is reverted.

## Open findings

(none — every finding has been disposed)
