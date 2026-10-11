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

## Open findings

- **BR-1** [Important] `fail-closed-scope-too-broad` TLOwners decodes full typed Metadata, so an unrelated malformed field skips a project without tl:
