---
gate: boundary-review
issue: 215
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-04T22:17:18-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: atlas paragraph restates ARCH-ORDER clause wording near-verbatim instead of mapping to it
          detail: |-
            atlas/workflow/architecture-principles.md:79-89 copies three registry sentences
            (":88" is character-identical to architecture.md:176; the N/A clause and the
            principle opener are near-verbatim). Nothing pins the copy, so a later rewording
            of the registry strands it. Fleet convention is route-don't-restate
            (construct/intents/superpowers.md:184,196,204) and the same atlas file says so at
            :24-27; sibling paragraphs (ARCH-SECURE :50-60, ARCH-MOCK, ARCH-CONSTRAINTS) all
            paraphrase. Keep the two boundary paragraphs and the pair#182/#185 origin — those
            are atlas-only content — and compress :79-89 to one map-level sentence. ARCH-DRY.
          family: atlas-restates-single-source
          round: 1
        - id: BR-2
          severity: Minor
          title: ARCH-ORDER is 1.4x the next-largest entry and grew the injected registry 35 percent
          detail: |-
            531 words vs ARCH-PURPOSE's 371 and ARCH-DRY's 92; architecture.md went 9,510 to
            12,868 bytes, and every registry-bearing prompt gained ~3.3 KB (dry/pure +32%).
            Six prompt paths carry it and no length norm exists for the artifact. Note for the
            next entry, not a defect in this one. ARCH-CONSTRAINTS, on the prompt budget.
          family: registry-entry-length-unbudgeted
          round: 1
        - id: BR-3
          severity: Minor
          title: Done-when asks the atlas to document the entry count, which is derived and was never carried there
          detail: |-
            The count comes from len(ArchitectureMarkers()) at architecture.go:57 and no
            "six"/"seven" ever appeared in atlas/workflow/architecture-principles.md. Writing
            one in would be the hand-maintained restatement the file's own "Adding an entry"
            section rules out. The criterion is wrong, not the implementation — drop "the
            count" from the bullet via a Revisions entry.
          family: derived-fact-restated
          round: 1
      blocked: true
---

# Gate ledger — ariadne#215 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-04T22:17:18-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `atlas-restates-single-source` atlas paragraph restates ARCH-ORDER clause wording near-verbatim instead of mapping to it
  atlas/workflow/architecture-principles.md:79-89 copies three registry sentences
  (":88" is character-identical to architecture.md:176; the N/A clause and the
  principle opener are near-verbatim). Nothing pins the copy, so a later rewording
  of the registry strands it. Fleet convention is route-don't-restate
  (construct/intents/superpowers.md:184,196,204) and the same atlas file says so at
  :24-27; sibling paragraphs (ARCH-SECURE :50-60, ARCH-MOCK, ARCH-CONSTRAINTS) all
  paraphrase. Keep the two boundary paragraphs and the pair#182/#185 origin — those
  are atlas-only content — and compress :79-89 to one map-level sentence. ARCH-DRY.
- **BR-2** [Minor] `registry-entry-length-unbudgeted` ARCH-ORDER is 1.4x the next-largest entry and grew the injected registry 35 percent
  531 words vs ARCH-PURPOSE's 371 and ARCH-DRY's 92; architecture.md went 9,510 to
  12,868 bytes, and every registry-bearing prompt gained ~3.3 KB (dry/pure +32%).
  Six prompt paths carry it and no length norm exists for the artifact. Note for the
  next entry, not a defect in this one. ARCH-CONSTRAINTS, on the prompt budget.
- **BR-3** [Minor] `derived-fact-restated` Done-when asks the atlas to document the entry count, which is derived and was never carried there
  The count comes from len(ArchitectureMarkers()) at architecture.go:57 and no
  "six"/"seven" ever appeared in atlas/workflow/architecture-principles.md. Writing
  one in would be the hand-maintained restatement the file's own "Adding an entry"
  section rules out. The criterion is wrong, not the implementation — drop "the
  count" from the bullet via a Revisions entry.

## Open findings

- **BR-1** [Important] `atlas-restates-single-source` atlas paragraph restates ARCH-ORDER clause wording near-verbatim instead of mapping to it
- **BR-2** [Minor] `registry-entry-length-unbudgeted` ARCH-ORDER is 1.4x the next-largest entry and grew the injected registry 35 percent
- **BR-3** [Minor] `derived-fact-restated` Done-when asks the atlas to document the entry count, which is derived and was never carried there
