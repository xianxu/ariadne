---
gate: boundary-review
issue: 326
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-11T00:09:05-07:00"
      dispatched: "2026-10-11T00:00:37-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: 'TLOwners decodes full typed Metadata, so an unrelated malformed field skips a project without tl:'
          detail: |-
            metadata.go:78 calls d.Metadata(); yaml type errors on Deps/MVPScope/Name (e.g. deps as a scalar string)
            make TLOwners error, and close.go:803 then skips a project with no tl: that it ticked before, which breaks
            the Spec clause "Projects without tl: behave exactly as today". Decode only the tl key (a minimal
            struct{TL any}) and add a not-owned unit row with deps as a scalar. This is the only instance in the window.
          family: fail-closed-scope-too-broad
          round: 1
      recipe: small-diff-review
      blocked: true
    - "n": 2
      timestamp: "2026-10-11T00:14:40-07:00"
      dispatched: "2026-10-11T00:13:58-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: TLOwners decodes only struct{TL any} (metadata.go:87-91); the not-owned row "unrelated field malformed" (deps scalar) would fail under the old Metadata decode because Deps is []string.
          round: 2
      findings:
        - id: BR-2
          severity: Minor
          title: New lessons.md bullet was inserted inside an existing bullet, splitting it
          detail: workshop/lessons.md:44-47. The line "A catalogued flag may waive only the refusal named by that catalog entry." now continues the 326 bullet instead of the fail-closed bullet above it. Move the new bullet below that line. This is the only instance in the window.
          family: edit-splits-neighbour
          round: 2
        - id: BR-3
          severity: Minor
          title: Plan row still says Metadata gains TL any, but the BR-1 fix removed that field
          detail: The issue's Plan, first ticked row. Reword it to match the decode of only the tl key, or add a Revisions entry.
          family: plan-row-drifts-from-code
          round: 2
      recipe: small-diff-review
      reviewed: f221a2aeb46970dd8c97b0295d10bcce39f20a8a
      blocked: false
---

# Gate ledger — ariadne#326 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-11T00:09:05-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `fail-closed-scope-too-broad` TLOwners decodes full typed Metadata, so an unrelated malformed field skips a project without tl:
  metadata.go:78 calls d.Metadata(); yaml type errors on Deps/MVPScope/Name (e.g. deps as a scalar string)
  make TLOwners error, and close.go:803 then skips a project with no tl: that it ticked before, which breaks
  the Spec clause "Projects without tl: behave exactly as today". Decode only the tl key (a minimal
  struct{TL any}) and add a not-owned unit row with deps as a scalar. This is the only instance in the window.

## Round 2 — 2026-10-11T00:14:40-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — TLOwners decodes only struct{TL any} (metadata.go:87-91); the not-owned row "unrelated field malformed" (deps scalar) would fail under the old Metadata decode because Deps is []string.

### Raised

- **BR-2** [Minor] `edit-splits-neighbour` New lessons.md bullet was inserted inside an existing bullet, splitting it
  workshop/lessons.md:44-47. The line "A catalogued flag may waive only the refusal named by that catalog entry." now continues the 326 bullet instead of the fail-closed bullet above it. Move the new bullet below that line. This is the only instance in the window.
- **BR-3** [Minor] `plan-row-drifts-from-code` Plan row still says Metadata gains TL any, but the BR-1 fix removed that field
  The issue's Plan, first ticked row. Reword it to match the decode of only the tl key, or add a Revisions entry.

## Open findings

- **BR-2** [Minor] `edit-splits-neighbour` New lessons.md bullet was inserted inside an existing bullet, splitting it
- **BR-3** [Minor] `plan-row-drifts-from-code` Plan row still says Metadata gains TL any, but the BR-1 fix removed that field
