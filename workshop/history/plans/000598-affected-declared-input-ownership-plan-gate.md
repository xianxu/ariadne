---
reviewer_runs:
    - version: 1
      intent:
        id: 0763ac6e-2cee-4408-92c3-b7b5393628e4
        repository: github.com/xianxu/ariadne
        issue_id: "000598"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        reviewed_head: 5e2768ed9ec7098f2c6600d573fa3db09315d644
        slot: ariadne:5
        requested_at: "2026-10-24T22:15:43.136109Z"
      phase: finished
      pid: 83838
      started_at: "2026-10-24T22:15:43.383547Z"
      finished_at: "2026-10-24T22:16:12.916427Z"
      session_id: b67424a5-4450-47c8-97b8-2728a3e3e2e8
      outcome: returned
gate: plan-quality
issue: 598
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: a7fb4185-c7b8-4ae3-8101-fd0f824c7c55
            repository: github.com/xianxu/ariadne
            issue_id: "000598"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
            reviewed_head: 6c4d9de4b21f8f2322577d5db80cdb642fa64efa
            slot: ariadne:5
            requested_at: "2026-10-24T22:09:32.40095Z"
          phase: finished
          pid: 60856
          started_at: "2026-10-24T22:09:32.60134Z"
          finished_at: "2026-10-24T22:11:33.495292Z"
          session_id: 892024fe-47ce-4a43-b1a5-4663182c168f
          outcome: returned
      "n": 1
      timestamp: "2026-10-24T15:11:33-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: readAffectedConfig takes a git closure, but TrunkView.FilesSelected needs a TrunkFile with its own context; the plan doesn't say how the seam changes
          detail: 'cigate.go:488 takes a git func closure; splitImpact passes a runner bounded by the landing deadline (landingattribution_impact.go:61-63). TrunkView sits on a TrunkFile (dir, remote, branch, own ctx; trunkfile.go:186-205), so the plan must choose: change the signature for all six callers, or add a bounded batch read over the closure. It must also say whether precloseck.go:33 and landingcar.go:164, which only need the declared bit, now pay for and can fail on the sidecar reads.'
          family: unstated-seam-change
          round: 1
        - kind: defect
          id: PQ-2
          severity: Important
          title: Done-when has no acceptance criteria for M2 (traceability input) or M3 (missed-mappings ledger, merge and train hooks, CLI)
          detail: M2 and M3 were added from the operator's direction recorded in the Log, but the issue's Done-when still covers only the M1 scope. Add a Revisions entry plus Done-when bullets so the close gate can check them.
          family: done-when-covers-scope
          round: 1
        - kind: defect
          id: PQ-3
          severity: Minor
          title: M3's new package-of-test parser duplicates flakes.ParseName / flakes.SplitTest
          detail: internal/flakes/registry.go:87 and the use in landingtrain_attribution.go:84-89 already split a failure name into package and test; suite names can be bare (suitefailures.go:56-62).
          family: reuse-existing-helper
          round: 1
        - kind: defect
          id: PQ-4
          severity: Minor
          title: M1 test rows list cases instead of naming the functions under test with one strategy line each
          detail: ParseSidecar, ParseTraceability and the Strip derivation parse bytes from outside the process; fuzz them, seeded with traversal and malformed forms, instead of listing cases in prose.
          family: test-prose-over-strategy
          round: 1
        - kind: defect
          id: PQ-5
          severity: Minor
          title: The trust section says a branch-added certificate's Strip alias only adds coverage, but it narrows a path that was full to one package
          detail: This is intended and trusted because main's config declares it; the wording should say narrowing rather than adding coverage.
          family: trust-claim-accuracy
          round: 1
        - kind: defect
          id: PQ-6
          severity: Minor
          title: The plan doesn't say what a sidecar with no reads member means, or who enforces reads per disposition
          detail: State that absent reads means no declaration, and that pair#853's audit enforces required reads, since the spec says a missing declaration must never authorize skipping coverage.
          family: missing-declaration-semantics
          round: 1
        - kind: defect
          id: PQ-7
          severity: Minor
          title: No explicit non-goals section
          detail: Several non-goals are implied (dispositions opaque to ariadne, sidecars never rewritten by ariadne, group failures out of scope for M3); collect them in one list.
          family: stated-non-goals
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 654259d3-23fd-4b20-b20f-9bcfd6d92c81
            repository: github.com/xianxu/ariadne
            issue_id: "000598"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
            reviewed_head: 87c5c9009b85ac2ae6b48113a012c0deb3e889be
            slot: ariadne:5
            requested_at: "2026-10-24T22:13:39.814807Z"
          phase: finished
          pid: 4666
          started_at: "2026-10-24T22:13:39.967385Z"
          finished_at: "2026-10-24T22:14:13.815109Z"
          session_id: 9fc438f6-1e46-462b-8a7e-093f8f570d6a
          outcome: returned
      "n": 2
      timestamp: "2026-10-24T15:14:14-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: TreeFilesSelected is extracted with an explicit ctx and dir; readAffectedConfig is unchanged; preclose and the car stand-in never read sidecars.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: M2 and M3 Done-when bullets added, with a Revisions entry.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Missed reuses flakes.SplitTest and ParseName; a bare name is not a miss.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Rows name their functions with one strategy each; ParseSidecar and Strip are fuzzed with traversal seeds.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: The wording now says narrowing, trusted because main declares the convention.
          round: 2
        - id: PQ-6
          disposition: addressed
          note: Absent reads declares nothing; pair#853's audit owns the per-disposition requirement.
          round: 2
        - id: PQ-7
          disposition: addressed
          note: A Non-goals section has been added.
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 5bed175c-bc76-41af-9153-4a6f83dbc290
            repository: github.com/xianxu/ariadne
            issue_id: "000598"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
            reviewed_head: 5e2768ed9ec7098f2c6600d573fa3db09315d644
            slot: ariadne:5
            requested_at: "2026-10-24T22:15:21.038618Z"
          phase: finished
          pid: 71772
          started_at: "2026-10-24T22:15:21.284298Z"
          finished_at: "2026-10-24T22:15:40.032261Z"
          session_id: e4e8cf8d-96e4-41c5-8294-ef8fa7a128ca
          outcome: returned
      "n": 3
      timestamp: "2026-10-24T15:15:40-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      blocked: false
      protocol_error: no valid findings block
content_hash: 1368972541e0e929c8882f2ab750a4e623d17418e5c205121425c93ca2cf2574
---

# Gate ledger — ariadne#598 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-24T15:11:33-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `unstated-seam-change` (kind: defect) readAffectedConfig takes a git closure, but TrunkView.FilesSelected needs a TrunkFile with its own context; the plan doesn't say how the seam changes
  cigate.go:488 takes a git func closure; splitImpact passes a runner bounded by the landing deadline (landingattribution_impact.go:61-63). TrunkView sits on a TrunkFile (dir, remote, branch, own ctx; trunkfile.go:186-205), so the plan must choose: change the signature for all six callers, or add a bounded batch read over the closure. It must also say whether precloseck.go:33 and landingcar.go:164, which only need the declared bit, now pay for and can fail on the sidecar reads.
- **PQ-2** [Important] `done-when-covers-scope` (kind: defect) Done-when has no acceptance criteria for M2 (traceability input) or M3 (missed-mappings ledger, merge and train hooks, CLI)
  M2 and M3 were added from the operator's direction recorded in the Log, but the issue's Done-when still covers only the M1 scope. Add a Revisions entry plus Done-when bullets so the close gate can check them.
- **PQ-3** [Minor] `reuse-existing-helper` (kind: defect) M3's new package-of-test parser duplicates flakes.ParseName / flakes.SplitTest
  internal/flakes/registry.go:87 and the use in landingtrain_attribution.go:84-89 already split a failure name into package and test; suite names can be bare (suitefailures.go:56-62).
- **PQ-4** [Minor] `test-prose-over-strategy` (kind: defect) M1 test rows list cases instead of naming the functions under test with one strategy line each
  ParseSidecar, ParseTraceability and the Strip derivation parse bytes from outside the process; fuzz them, seeded with traversal and malformed forms, instead of listing cases in prose.
- **PQ-5** [Minor] `trust-claim-accuracy` (kind: defect) The trust section says a branch-added certificate's Strip alias only adds coverage, but it narrows a path that was full to one package
  This is intended and trusted because main's config declares it; the wording should say narrowing rather than adding coverage.
- **PQ-6** [Minor] `missing-declaration-semantics` (kind: defect) The plan doesn't say what a sidecar with no reads member means, or who enforces reads per disposition
  State that absent reads means no declaration, and that pair#853's audit enforces required reads, since the spec says a missing declaration must never authorize skipping coverage.
- **PQ-7** [Minor] `stated-non-goals` (kind: defect) No explicit non-goals section
  Several non-goals are implied (dispositions opaque to ariadne, sidecars never rewritten by ariadne, group failures out of scope for M3); collect them in one list.

## Round 2 — 2026-10-24T15:14:14-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — TreeFilesSelected is extracted with an explicit ctx and dir; readAffectedConfig is unchanged; preclose and the car stand-in never read sidecars.
- PQ-2 — addressed — M2 and M3 Done-when bullets added, with a Revisions entry.
- PQ-3 — addressed — Missed reuses flakes.SplitTest and ParseName; a bare name is not a miss.
- PQ-4 — addressed — Rows name their functions with one strategy each; ParseSidecar and Strip are fuzzed with traversal seeds.
- PQ-5 — addressed — The wording now says narrowing, trusted because main declares the convention.
- PQ-6 — addressed — Absent reads declares nothing; pair#853's audit owns the per-disposition requirement.
- PQ-7 — addressed — A Non-goals section has been added.

## Round 3 — 2026-10-24T15:15:40-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
