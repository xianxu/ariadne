---
gate: boundary-review
issue: 425
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 8adc1145-9547-4852-903b-1cee43ff154f
            repository: github.com/xianxu/ariadne
            issue_id: "000425"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/626dcca66105ab49
            reviewed_head: 7d30d9a34392db47163f8225152dc4d1a4a31b25
            slot: ariadne:5
            requested_at: "2026-10-17T14:08:30.996756Z"
          phase: finished
          pid: 39858
          started_at: "2026-10-17T14:08:31.181331Z"
          finished_at: "2026-10-17T14:10:58.80651Z"
          session_id: 01a14a31-5dd6-7133-84d6-57bd37e5723a
          outcome: returned
      "n": 1
      timestamp: "2026-10-17T07:10:59-07:00"
      dispatched: "2026-10-17T07:08:30-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Important
          title: README update appears missing for the changed pre-close and landing sequence
          detail: 'cmd/sdlc/helptext/local-check.md:63–68 changes the user-facing procedure to close the tested branch without chasing newer main, then let sdlc merge integrate under admission. README.md:466–474 remains unchanged and omits that distinction. Family enumeration: the affected README passages are landing-line usage at lines 444–452 and pre-close suite usage at lines 466–474. Update these passages with the approved sequence, preserving freshness and branch-fix verification requirements.'
          family: user-facing-workflow-documentation
          round: 1
        - id: BR-2
          severity: Important
          title: The changed embedded workflow instructions have no regression checks
          detail: 'The production change is executable agent guidance, but all added tests exercise unchanged merge behavior. Existing help checks do not assert the new phase ownership or reject the removed instruction to merge main manually before landing. Family enumeration: cmd/tl-worker/help.md:34–40,83–85; cmd/tl-lead/help.md:327–333; cmd/sdlc/helptext/local-check.md:63–68. Extend rendered-help contract checks to protect the branch-suite → close → merge sequence, prohibition on manual chasing/holding, and freshness/code-fix exceptions; demonstrate that restoring the previous guidance fails those checks.'
          family: executable-guidance-regression-coverage
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 921afc49-2378-48e2-b1c1-3c556afe97d3
            repository: github.com/xianxu/ariadne
            issue_id: "000425"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/a74746d19fbbafd6
            reviewed_head: 40142a581f1c75bc04c7af76240f45ce433d5db6
            slot: ariadne:5
            requested_at: "2026-10-17T14:55:11.582659Z"
          phase: finished
          pid: 30153
          started_at: "2026-10-17T14:55:11.683823Z"
          finished_at: "2026-10-17T14:56:32.691828Z"
          session_id: 01a14a5c-1823-7da1-a011-c39f6de459e6
          outcome: returned
      "n": 2
      timestamp: "2026-10-17T07:56:33-07:00"
      dispatched: "2026-10-17T07:55:11-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:444–449 distinguishes branch verification from admitted integration; lines 468–479 preserve freshness, branch-fix verification, and conservative reuse. These match landing.go admission/gating and cigate.go ancestor-reuse behavior.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Rendered-help checks cover worker, lead, and local-check guidance. All pass at pinned head and all fail when the three help files are restored to base versions in a temporary copy, removed after verification.
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: 40142a581f1c75bc04c7af76240f45ce433d5db6
      blocked: false
---

# Gate ledger — ariadne#425 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T07:10:59-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `user-facing-workflow-documentation` README update appears missing for the changed pre-close and landing sequence
  cmd/sdlc/helptext/local-check.md:63–68 changes the user-facing procedure to close the tested branch without chasing newer main, then let sdlc merge integrate under admission. README.md:466–474 remains unchanged and omits that distinction. Family enumeration: the affected README passages are landing-line usage at lines 444–452 and pre-close suite usage at lines 466–474. Update these passages with the approved sequence, preserving freshness and branch-fix verification requirements.
- **BR-2** [Important] `executable-guidance-regression-coverage` The changed embedded workflow instructions have no regression checks
  The production change is executable agent guidance, but all added tests exercise unchanged merge behavior. Existing help checks do not assert the new phase ownership or reject the removed instruction to merge main manually before landing. Family enumeration: cmd/tl-worker/help.md:34–40,83–85; cmd/tl-lead/help.md:327–333; cmd/sdlc/helptext/local-check.md:63–68. Extend rendered-help contract checks to protect the branch-suite → close → merge sequence, prohibition on manual chasing/holding, and freshness/code-fix exceptions; demonstrate that restoring the previous guidance fails those checks.

## Round 2 — 2026-10-17T07:56:33-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — README.md:444–449 distinguishes branch verification from admitted integration; lines 468–479 preserve freshness, branch-fix verification, and conservative reuse. These match landing.go admission/gating and cigate.go ancestor-reuse behavior.
- BR-2 — addressed — Rendered-help checks cover worker, lead, and local-check guidance. All pass at pinned head and all fail when the three help files are restored to base versions in a temporary copy, removed after verification.

## Open findings

(none — every finding has been disposed)
