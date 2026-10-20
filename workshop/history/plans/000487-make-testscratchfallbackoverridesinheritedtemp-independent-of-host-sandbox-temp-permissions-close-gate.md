---
gate: boundary-review
issue: 487
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: d8d20531-3b3c-49a1-8dd1-8e05ee0ab40f
            repository: github.com/xianxu/ariadne
            issue_id: "000487"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/00ffd7a3ce3c2e07
            reviewed_head: 58adf0d713a29f8ac199c3cbdc5d376b1b0138bd
            slot: ariadne:3
            requested_at: "2026-10-20T07:35:52.55733Z"
          phase: finished
          pid: 14405
          started_at: "2026-10-20T07:35:52.640596Z"
          finished_at: "2026-10-20T07:37:54.564818Z"
          session_id: 01a1583c-f71f-7812-babb-589f170f9b06
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 58adf0d713a29f8ac199c3cbdc5d376b1b0138bd
        tracker_head: 701e3739bbd32f2c2056f9bb5770bf86271e9661
        branch_point: 50a6ca318d1804af77fad9453cfb283818fade18
        since: "2026-10-20T00:23:30-07:00"
        until: "2026-10-20T00:34:48-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T00:23:30-07:00"
              end: "2026-10-20T00:34:48-07:00"
              began: "2026-10-20T00:23:30-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 5f92d90085441a7f54ec8396e46a14725cd3d266d86c2170038a654b65cad0e6
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 65891636
              windows:
                - start: "2026-10-20T00:23:30-07:00"
                  end: "2026-10-20T00:34:48-07:00"
      "n": 1
      timestamp: "2026-10-20T00:37:55-07:00"
      dispatched: "2026-10-20T00:35:52-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: unverifiable
          id: BR-1
          severity: Important
          title: Denied-host-temp acceptance could not be verified
          detail: 'cmd/sdlc/testdata/487-deny-host-temp.sb:5: with TMPDIR and GOTMPDIR set to an owned /Volumes/TestScratch fixture, running sandbox-exec -f cmd/sdlc/testdata/487-deny-host-temp.sb go test ./cmd/sdlc ./cmd/sdlc/internal/scratchdisk -run ''Test(ScratchFallback|NormalTemp|MkdirTemp|UserTempDir)'' -count=1 failed before tests started: "sandbox-exec: sandbox_apply: Operation not permitted" (exit 71). Ordinary write probes succeeded in both the Darwin user-temp root and /tmp, so passing ordinary tests cannot establish this acceptance clause. The complete affected family is both fallback tests at suiteenvironment_test.go:87 and :140 and the candidate controls selected by that command. Supply measured host evidence showing the profile active, host-temp writes denied, fixture-root writes allowed, and the targeted tests passing.'
          family: sandbox-acceptance-evidence
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a58377e1-843c-4c70-942a-7872ab153749
            repository: github.com/xianxu/ariadne
            issue_id: "000487"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/a8e08ec6aaf9f04f
            reviewed_head: 8513558178988a1fe9af9ce1bc32f484fb7da941
            slot: ariadne:3
            requested_at: "2026-10-20T07:42:06.980584Z"
          phase: finished
          pid: 27683
          started_at: "2026-10-20T07:42:07.07241Z"
          finished_at: "2026-10-20T07:43:13.00626Z"
          session_id: 01a15842-add0-7d40-a3bd-2296e0447c9e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 8513558178988a1fe9af9ce1bc32f484fb7da941
        tracker_head: 701e3739bbd32f2c2056f9bb5770bf86271e9661
        branch_point: 50a6ca318d1804af77fad9453cfb283818fade18
        since: "2026-10-20T00:23:30-07:00"
        until: "2026-10-20T00:41:39-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T00:23:30-07:00"
              end: "2026-10-20T00:41:39-07:00"
              began: "2026-10-20T00:23:30-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: a0acd6d7d46c2d302baa7ed30b44914ee6a5bce3f5b23e8a39c3842fb1921fc4
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 66192462
              windows:
                - start: "2026-10-20T00:23:30-07:00"
                  end: "2026-10-20T00:41:39-07:00"
        cuts:
            - dispatched: "2026-10-20T00:35:52-07:00"
              finalized: "2026-10-20T00:37:55-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-20T00:43:14-07:00"
      dispatched: "2026-10-20T00:42:06-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: withdrawn
          note: 'The committed sandbox-host-evidence.txt answers the missing measurement: fixture create/remove succeeded, Darwin user-temp and /tmp writes were denied with errno 1, and both fallback tests plus candidate controls passed under the same profile with exit 0. The profile SHA256 matches this checkout; comparison against measured source HEAD 58adf0d713a29f8ac199c3cbdc5d376b1b0138bd confirms no subsequent production or test-code changes.'
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: 8513558178988a1fe9af9ce1bc32f484fb7da941
      blocked: false
---

# Gate ledger — ariadne#487 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T00:37:55-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `sandbox-acceptance-evidence` (kind: unverifiable) Denied-host-temp acceptance could not be verified
  cmd/sdlc/testdata/487-deny-host-temp.sb:5: with TMPDIR and GOTMPDIR set to an owned /Volumes/TestScratch fixture, running sandbox-exec -f cmd/sdlc/testdata/487-deny-host-temp.sb go test ./cmd/sdlc ./cmd/sdlc/internal/scratchdisk -run 'Test(ScratchFallback|NormalTemp|MkdirTemp|UserTempDir)' -count=1 failed before tests started: "sandbox-exec: sandbox_apply: Operation not permitted" (exit 71). Ordinary write probes succeeded in both the Darwin user-temp root and /tmp, so passing ordinary tests cannot establish this acceptance clause. The complete affected family is both fallback tests at suiteenvironment_test.go:87 and :140 and the candidate controls selected by that command. Supply measured host evidence showing the profile active, host-temp writes denied, fixture-root writes allowed, and the targeted tests passing.

## Round 2 — 2026-10-20T00:43:14-07:00 (codex) — passed

### Disposed

- BR-1 — withdrawn — The committed sandbox-host-evidence.txt answers the missing measurement: fixture create/remove succeeded, Darwin user-temp and /tmp writes were denied with errno 1, and both fallback tests plus candidate controls passed under the same profile with exit 0. The profile SHA256 matches this checkout; comparison against measured source HEAD 58adf0d713a29f8ac199c3cbdc5d376b1b0138bd confirms no subsequent production or test-code changes.

## Open findings

(none — every finding has been disposed)
