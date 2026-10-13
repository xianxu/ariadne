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

## Open findings

- **BR-1** [Important] `load-bearing-branch-untested` The ref-containment half of moveSide.Leavable (for-each-ref --contains) has no test; mutating it to false keeps every stage and move test green
- **BR-2** [Minor] `render-format-consistency` The fleet render adds a bare tab-separated "staged" word to key=value output (render.go:58-60)
- **BR-3** [Minor] `load-bearing-branch-untested` The repeat stage at the same HEAD ("already staged", stage.go:84-85) has no test
- **BR-4** [Minor] `silent-error-swallowing` stageBuild treats every readBuildInfo error as "not a Go binary" and silently skips it (stagebuild.go:36-38)
- **BR-5** [Minor] `checklist-claims-match-evidence` The issue checks "xx-tl change sent to the hub", but the Log says the hub is told at landing; the durable plan's task boxes are all unchecked
