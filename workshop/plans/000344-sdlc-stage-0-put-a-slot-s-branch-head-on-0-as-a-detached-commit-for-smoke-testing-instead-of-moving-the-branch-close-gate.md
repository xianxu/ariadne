---
gate: boundary-review
issue: 344
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T14:56:48-07:00"
      dispatched: "2026-10-13T14:50:33-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: The ref-containment half of moveSide.Leavable (for-each-ref --contains) has no test; mutating it to false keeps every stage and move test green
          detail: cmd/sdlc/move.go:296-302. This path is what the stage recovery contract's LostResponse and the plan's effect-order argument depend on (a crash before the record is written). Add a real-git row with :0 detached at a branch-contained commit and no record, and assert that stage, --clear and move all accept it.
          family: load-bearing-branch-untested
          round: 1
        - id: BR-2
          severity: Minor
          title: The fleet render adds a bare tab-separated "staged" word to key=value output (render.go:58-60)
          detail: Use staged=true, or a key=value form, to keep the line format consistent.
          family: render-format-consistency
          round: 1
        - id: BR-3
          severity: Minor
          title: The repeat stage at the same HEAD ("already staged", stage.go:84-85) has no test
          family: load-bearing-branch-untested
          round: 1
        - id: BR-4
          severity: Minor
          title: stageBuild treats every readBuildInfo error as "not a Go binary" and silently skips it (stagebuild.go:36-38)
          family: silent-error-swallowing
          round: 1
        - id: BR-5
          severity: Minor
          title: The issue checks "xx-tl change sent to the hub", but the Log says the hub is told at landing; the durable plan's task boxes are all unchecked
          family: checklist-claims-match-evidence
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-13T15:15:56-07:00"
      dispatched: "2026-10-13T15:10:55-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: TestStageAcceptsADetachARefKeeps (stage_test.go:192) goes red in all three subtests when move.go:301's ref check is forced false; verified by mutation in the review checkout.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: render.go:59 now emits staged=true, but no render_test row covers a staged row, so nothing fails without the fix; add one row to render_test.go.
          round: 2
        - id: BR-3
          disposition: addressed
          note: stage_test.go:66-69 asserts "already staged on" and the unchanged stage on a repeat at the same HEAD.
          round: 2
        - id: BR-4
          disposition: not-addressed
          note: stagebuild.go:37-41 now warns on non-"not Go" errors, but no test fakes readBuildInfo returning another error; the fix is unpinned (and notGoBinary matches on error text).
          round: 2
        - id: BR-5
          disposition: addressed
          note: Plan task boxes are ticked, and the issue's Docs item now says the xx-tl change lives in ops and goes to the hub in the landing report, matching the Log.
          round: 2
      findings:
        - id: BR-6
          severity: Minor
          title: The stage recovery contract's LostResponse claim does not cite TestStageAcceptsADetachARefKeeps as its proof
          detail: 'This is the 2nd finding in family checklist-claims-match-evidence. The rule: every claim a durable contract makes (plan box, catalog field) names the test that proves it. Sweep the stage Contract in recovery/catalog.go:199-213 so that each of Effects, Repeat, LostResponse and Ends maps to a listed Proof.'
          family: checklist-claims-match-evidence
          round: 2
      recipe: milestone-review
      isolation: checkout
      reviewed: 911b8afa4437f1ab911c32d2c819d3fb329c9274
      blocked: false
---

# Gate ledger — ariadne#344 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T14:56:48-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `load-bearing-branch-untested` The ref-containment half of moveSide.Leavable (for-each-ref --contains) has no test; mutating it to false keeps every stage and move test green
  cmd/sdlc/move.go:296-302. This path is what the stage recovery contract's LostResponse and the plan's effect-order argument depend on (a crash before the record is written). Add a real-git row with :0 detached at a branch-contained commit and no record, and assert that stage, --clear and move all accept it.
- **BR-2** [Minor] `render-format-consistency` The fleet render adds a bare tab-separated "staged" word to key=value output (render.go:58-60)
  Use staged=true, or a key=value form, to keep the line format consistent.
- **BR-3** [Minor] `load-bearing-branch-untested` The repeat stage at the same HEAD ("already staged", stage.go:84-85) has no test
- **BR-4** [Minor] `silent-error-swallowing` stageBuild treats every readBuildInfo error as "not a Go binary" and silently skips it (stagebuild.go:36-38)
- **BR-5** [Minor] `checklist-claims-match-evidence` The issue checks "xx-tl change sent to the hub", but the Log says the hub is told at landing; the durable plan's task boxes are all unchecked

## Round 2 — 2026-10-13T15:15:56-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — TestStageAcceptsADetachARefKeeps (stage_test.go:192) goes red in all three subtests when move.go:301's ref check is forced false; verified by mutation in the review checkout.
- BR-2 — not-addressed — render.go:59 now emits staged=true, but no render_test row covers a staged row, so nothing fails without the fix; add one row to render_test.go.
- BR-3 — addressed — stage_test.go:66-69 asserts "already staged on" and the unchanged stage on a repeat at the same HEAD.
- BR-4 — not-addressed — stagebuild.go:37-41 now warns on non-"not Go" errors, but no test fakes readBuildInfo returning another error; the fix is unpinned (and notGoBinary matches on error text).
- BR-5 — addressed — Plan task boxes are ticked, and the issue's Docs item now says the xx-tl change lives in ops and goes to the hub in the landing report, matching the Log.

### Raised

- **BR-6** [Minor] `checklist-claims-match-evidence` The stage recovery contract's LostResponse claim does not cite TestStageAcceptsADetachARefKeeps as its proof
  This is the 2nd finding in family checklist-claims-match-evidence. The rule: every claim a durable contract makes (plan box, catalog field) names the test that proves it. Sweep the stage Contract in recovery/catalog.go:199-213 so that each of Effects, Repeat, LostResponse and Ends maps to a listed Proof.

## Open findings

- **BR-2** [Minor] `render-format-consistency` The fleet render adds a bare tab-separated "staged" word to key=value output (render.go:58-60)
- **BR-4** [Minor] `silent-error-swallowing` stageBuild treats every readBuildInfo error as "not a Go binary" and silently skips it (stagebuild.go:36-38)
- **BR-6** [Minor] `checklist-claims-match-evidence` The stage recovery contract's LostResponse claim does not cite TestStageAcceptsADetachARefKeeps as its proof
