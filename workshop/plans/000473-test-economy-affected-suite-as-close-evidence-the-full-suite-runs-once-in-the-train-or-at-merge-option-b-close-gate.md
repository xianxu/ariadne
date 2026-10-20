---
gate: boundary-review
issue: 473
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 6565c97c-7663-4161-a479-04a5d599c1d3
            repository: github.com/xianxu/ariadne
            issue_id: "000473"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/2c46c994bc92bfc1
            reviewed_head: 36f16ff1e7a6c5b775a86aaaea601918089d89a9
            slot: ariadne:2
            requested_at: "2026-10-20T08:49:20.730538Z"
          phase: finished
          pid: 82816
          started_at: "2026-10-20T08:49:20.876995Z"
          finished_at: "2026-10-20T08:57:31.108352Z"
          session_id: 6f9f2218-be5e-4d86-9158-14cdee8133d6
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 36f16ff1e7a6c5b775a86aaaea601918089d89a9
        tracker_head: 97ccd9410a99b8cadb917699d8779fb664a986de
        branch_point: 0956e5e1c71bd6734f195929fc151e0c752dbc14
        since: "2026-10-19T20:26:49-07:00"
        until: "2026-10-20T01:48:42-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T20:26:49-07:00"
              end: "2026-10-20T01:48:42-07:00"
              began: "2026-10-19T20:26:49-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5a5aa7ef6697f8ce115263d087981e41bd7603b84db90304dbc47ea946847d1b
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 3148632
              windows:
                - start: "2026-10-19T20:26:49-07:00"
                  end: "2026-10-20T01:48:42-07:00"
        queue_waits:
            - attempt_id: bcd9b7407517e24367c744af82d5db34
              slot: ariadne:2
              record_digest: 38cc6a62061e1a988f1d7f0f855462befe0fbd9fd08ff6e8d5e7da15ab214da4
              start: "2026-10-20T03:26:49Z"
              end: "2026-10-20T03:54:30.635003Z"
              queue_seconds: 1661.635003
              raw_seconds: 1669.213411
              exact: true
            - attempt_id: 1e519eee13691a7f9ae238f3a34af108
              slot: ariadne:2
              record_digest: ec17427e86246fbb78fe51b1c02c6c1a3c11100644e7ee265a3bbb81ecc632fc
              start: "2026-10-20T04:15:40.4019Z"
              end: "2026-10-20T04:37:58.530161Z"
              queue_seconds: 1338.128261
              raw_seconds: 1338.128261
              exact: true
            - attempt_id: ec45d2b999a4f613e1087a285f7a4e97
              slot: ariadne:2
              record_digest: 44c532f2cb7fa09e8ca756cfd526f624744a71e1024a397fc4dcd2c47bf96ea1
              start: "2026-10-20T08:21:56.389306Z"
              end: "2026-10-20T08:36:54.647565Z"
              queue_seconds: 898.258259
              raw_seconds: 898.258259
              exact: true
        protected:
            - issue: "479"
              slot: ariadne:2
              start: "2026-10-19T21:12:26-07:00"
              end: "2026-10-20T04:15:25.377Z"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 27.769466666666666
      "n": 1
      timestamp: "2026-10-20T01:57:31-07:00"
      dispatched: "2026-10-20T01:49:20-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: A root Go package claims every changed path, so unclassified files never force the full suite
          detail: 'packageIn (graph.go:30) walks up to Dir "." and Plan (plan.go:44) marks the path claimed, as pinned by plan_test.go:125 where b/x.go maps to the root package. In a module whose root is a package, scripts/ and tests/lib/ changes select the root package and its dependents instead of Full, against the Spec''s unclassified-to-full-suite valve. Fix: limit ancestor claims (own dir plus testdata/, never the root for subdirectories) and add a root-package plus non-Go path table row expecting Full.'
          family: unclassified-path-fails-closed
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: An empty change set (rev at or behind origin/main) posts a "docs-only" affected success without running anything
          detail: In runAffectedCheck (localcheck_affected.go:33,43), publishGateHasCodeSurface of an empty path list is false, so a branch tip equal to main records sdlc-local/affected success "nothing affected (docs-only)" on a possibly-code commit. That is fabricated evidence the M2 stand-in and ancestor walk will trust. Refuse (or post nothing) when paths is empty, and add a test with sha equal to main.
          family: no-evidence-recorded-as-pass
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: README update appears missing for local-check run --name affected and .sdlc/affected.json
          detail: README.md:557-622 documents local-check run and the close suite but not the new name or config file. Plan Task 8 lists help text and atlas only, so README would be missed in M2 as well. Add README.md to Task 8 (plan revision) or update it now.
          family: docs-gate-new-surface
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: --all combined with --name affected silently drops --all
          detail: localcheck_cmd.go:148 skips selectLocalChecks when only affected is named, so the existing all-versus-names conflict error no longer fires.
          family: cli-flag-conflict
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: affected.ParseConfig accepts duplicate JSON keys (last one wins)
          detail: ARCH-EVOLVE asks that duplicate keys stay strict; encoding/json silently takes the last value.
          family: strict-persisted-format
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: The affected check carries KnownFlakes, but the flake step and success guard still key on "suite" until M2
          detail: This is inert in M1 (localcheck_run.go:136,184 and localcheck_republish.go:198). The docs-only record also saves no attempt, so it is not republishable. Note both in the Log so D4 is not assumed live.
          family: deferred-scope-visibility
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      blocked: true
---

# Gate ledger — ariadne#473 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T01:57:31-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `unclassified-path-fails-closed` (kind: defect) A root Go package claims every changed path, so unclassified files never force the full suite
  packageIn (graph.go:30) walks up to Dir "." and Plan (plan.go:44) marks the path claimed, as pinned by plan_test.go:125 where b/x.go maps to the root package. In a module whose root is a package, scripts/ and tests/lib/ changes select the root package and its dependents instead of Full, against the Spec's unclassified-to-full-suite valve. Fix: limit ancestor claims (own dir plus testdata/, never the root for subdirectories) and add a root-package plus non-Go path table row expecting Full.
- **BR-2** [Important] `no-evidence-recorded-as-pass` (kind: defect) An empty change set (rev at or behind origin/main) posts a "docs-only" affected success without running anything
  In runAffectedCheck (localcheck_affected.go:33,43), publishGateHasCodeSurface of an empty path list is false, so a branch tip equal to main records sdlc-local/affected success "nothing affected (docs-only)" on a possibly-code commit. That is fabricated evidence the M2 stand-in and ancestor walk will trust. Refuse (or post nothing) when paths is empty, and add a test with sha equal to main.
- **BR-3** [Important] `docs-gate-new-surface` (kind: defect) README update appears missing for local-check run --name affected and .sdlc/affected.json
  README.md:557-622 documents local-check run and the close suite but not the new name or config file. Plan Task 8 lists help text and atlas only, so README would be missed in M2 as well. Add README.md to Task 8 (plan revision) or update it now.
- **BR-4** [Minor] `cli-flag-conflict` (kind: defect) --all combined with --name affected silently drops --all
  localcheck_cmd.go:148 skips selectLocalChecks when only affected is named, so the existing all-versus-names conflict error no longer fires.
- **BR-5** [Minor] `strict-persisted-format` (kind: defect) affected.ParseConfig accepts duplicate JSON keys (last one wins)
  ARCH-EVOLVE asks that duplicate keys stay strict; encoding/json silently takes the last value.
- **BR-6** [Minor] `deferred-scope-visibility` (kind: defect) The affected check carries KnownFlakes, but the flake step and success guard still key on "suite" until M2
  This is inert in M1 (localcheck_run.go:136,184 and localcheck_republish.go:198). The docs-only record also saves no attempt, so it is not republishable. Note both in the Log so D4 is not assumed live.

## Open findings

- **BR-1** [Important] `unclassified-path-fails-closed` (kind: defect) A root Go package claims every changed path, so unclassified files never force the full suite
- **BR-2** [Important] `no-evidence-recorded-as-pass` (kind: defect) An empty change set (rev at or behind origin/main) posts a "docs-only" affected success without running anything
- **BR-3** [Important] `docs-gate-new-surface` (kind: defect) README update appears missing for local-check run --name affected and .sdlc/affected.json
- **BR-4** [Minor] `cli-flag-conflict` (kind: defect) --all combined with --name affected silently drops --all
- **BR-5** [Minor] `strict-persisted-format` (kind: defect) affected.ParseConfig accepts duplicate JSON keys (last one wins)
- **BR-6** [Minor] `deferred-scope-visibility` (kind: defect) The affected check carries KnownFlakes, but the flake step and success guard still key on "suite" until M2
