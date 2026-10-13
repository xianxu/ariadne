---
gate: boundary-review
issue: 343
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T15:41:43-07:00"
      dispatched: "2026-10-13T15:31:11-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: New e2e test copies the killed-owner setup of TestCompileRecoversAfterProcessDeathDuringGeneration
          detail: ARCH-DRY. dynamic_test.go:460-495 repeats about 25 lines from dynamic_test.go:405-440 (fixture, marker swap, helper-child kill, marker restore, run, stage-dir assertion); only the marker tail differs. Extract a helper such as killOwnerMidGeneration(t, tail). These two tests are the only instances in the window.
          family: test-setup-duplication
          round: 1
        - id: BR-2
          severity: Minor
          title: Reclaim now spends up to Settle (1s) per dead-owner stage still held by a live orphan
          detail: The stages are waited on one after another inside the Reclaim loop (stage.go:88), and Reclaim is called from publication.go:34, plan/staged.go:40, release/main.go:65 and acquire/stage.go:9. This is acceptable because the case is rare, but worth a sentence in the atlas or code comment.
          family: undocumented-latency-bound
          round: 1
      recipe: small-diff-review
      isolation: checkout
      reviewed: c2585f059c31350ef226d24e7dc7828e9245e005
      blocked: false
    - "n": 2
      timestamp: "2026-10-13T15:44:07-07:00"
      dispatched: "2026-10-13T15:43:01-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Shared killOwnerMidGeneration(t, tail) + compileLeavesNoOwnedStage(t, leaf) at dynamic_test.go:407/436 now cover both tests; the reverted fix still turns the e2e test red (3/3).
          round: 2
        - id: BR-2
          disposition: addressed
          note: Cost written down in the stage.go:88-91 comment and atlas/workflow/weave.md:354-355; matches Settle=1s (lease.go:28) and the serial Reclaim loop.
          round: 2
      recipe: fix-review
      isolation: checkout
      reviewed: 06c9afd08bbc9ca7b05e54601cb285e558076a13
      blocked: false
---

# Gate ledger — ariadne#343 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T15:41:43-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `test-setup-duplication` New e2e test copies the killed-owner setup of TestCompileRecoversAfterProcessDeathDuringGeneration
  ARCH-DRY. dynamic_test.go:460-495 repeats about 25 lines from dynamic_test.go:405-440 (fixture, marker swap, helper-child kill, marker restore, run, stage-dir assertion); only the marker tail differs. Extract a helper such as killOwnerMidGeneration(t, tail). These two tests are the only instances in the window.
- **BR-2** [Minor] `undocumented-latency-bound` Reclaim now spends up to Settle (1s) per dead-owner stage still held by a live orphan
  The stages are waited on one after another inside the Reclaim loop (stage.go:88), and Reclaim is called from publication.go:34, plan/staged.go:40, release/main.go:65 and acquire/stage.go:9. This is acceptable because the case is rare, but worth a sentence in the atlas or code comment.

## Round 2 — 2026-10-13T15:44:07-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Shared killOwnerMidGeneration(t, tail) + compileLeavesNoOwnedStage(t, leaf) at dynamic_test.go:407/436 now cover both tests; the reverted fix still turns the e2e test red (3/3).
- BR-2 — addressed — Cost written down in the stage.go:88-91 comment and atlas/workflow/weave.md:354-355; matches Settle=1s (lease.go:28) and the serial Reclaim loop.

## Open findings

(none — every finding has been disposed)
