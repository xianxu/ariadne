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
    - "n": 2
      timestamp: "2026-09-04T22:28:48-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'Verified by measurement, not by the commit message: 12-word-window scan of the atlas paragraph against the registry entry finds 0 shared spans post-fix vs 26 overlapping windows pre-fix; no sibling atlas paragraph restates its entry either.'
          round: 2
        - id: BR-2
          disposition: addressed
          note: Recorded in Revisions with the measurement and the reason not to trim; the finding itself scoped this as a note for the next entry. No follow-up issue was filed for the stated "worth its own issue".
          round: 2
        - id: BR-3
          disposition: addressed
          note: Done-when bullet corrected as the finding asked. A sibling site in the same family remains at Plan step 5 and is raised below rather than re-raised here.
          round: 2
      findings:
        - id: BR-4
          severity: Minor
          title: Plan step 5 still lists "the count" as an atlas deliverable and is ticked, contradicting the corrected Done-when
          detail: 'This is the 2nd finding in family `derived-fact-restated`. BR-3 named the Done-when bullet and that bullet was fixed; the enumerable sibling at workshop/issues/000215-arch-order-architecture-principle.md:243 was not — "Update atlas...: the entry, the count, the ARCH-PURE boundary..." is still there and marked [x], asserting delivery of the hand-maintained restatement Done-when now forbids. Do not fix only this site: state the rule — a fact derived at runtime from the registry (entry count, marker list, ARCH_STAR) may appear as a verification criterion about a derived output, never as a hand-maintained requirement in any artifact — and sweep the enumeration. Measured prevalence for this issue: grep over the issue file and atlas returns exactly one demand site (:243); all other "seven"/"7" hits are verification assertions the rule permits. ARCH-PURPOSE.'
          family: derived-fact-restated
          round: 2
        - id: BR-5
          severity: Minor
          title: The map-don't-restate rule BR-1 established has no durable home; the atlas "Adding an entry" checklist still omits the atlas paragraph entirely
          detail: 'This is the 2nd finding in family `atlas-restates-single-source`. BR-1''s instances are genuinely gone (0 shared 12-word spans, and no sibling paragraph restates), so the sweep is clean — what is missing is the rule. atlas/workflow/architecture-principles.md:143 says "a new entry touches: architecture.md, that one list, and the goldens", omitting the map-level paragraph in that same file that every entry since ARCH-MOCK has in fact required. The rule is currently recorded only in this issue''s ## Revisions, which archives to workshop/history/ that AGENTS.md marks "don''t read", so the next entry-adder reads a three-item checklist and repeats BR-1. Fix the class: add the fourth touch point to :143 stating the shape — this file gets a MAP-level paragraph (boundaries against confusable neighbours, deliberate shaping choices, in-fleet origin) and never a restatement of the entry''s clauses, the same route-don''t-restate discipline gatefindings.go already follows at :24-30. Measured prevalence: 1 restatement in 1 entry-add under the unwritten rule; 0 remaining after this round''s sweep. ARCH-DRY, ARCH-PURPOSE.'
          family: atlas-restates-single-source
          round: 2
      blocked: false
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

## Round 2 — 2026-09-04T22:28:48-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Verified by measurement, not by the commit message: 12-word-window scan of the atlas paragraph against the registry entry finds 0 shared spans post-fix vs 26 overlapping windows pre-fix; no sibling atlas paragraph restates its entry either.
- BR-2 — addressed — Recorded in Revisions with the measurement and the reason not to trim; the finding itself scoped this as a note for the next entry. No follow-up issue was filed for the stated "worth its own issue".
- BR-3 — addressed — Done-when bullet corrected as the finding asked. A sibling site in the same family remains at Plan step 5 and is raised below rather than re-raised here.

### Raised

- **BR-4** [Minor] `derived-fact-restated` Plan step 5 still lists "the count" as an atlas deliverable and is ticked, contradicting the corrected Done-when
  This is the 2nd finding in family `derived-fact-restated`. BR-3 named the Done-when bullet and that bullet was fixed; the enumerable sibling at workshop/issues/000215-arch-order-architecture-principle.md:243 was not — "Update atlas...: the entry, the count, the ARCH-PURE boundary..." is still there and marked [x], asserting delivery of the hand-maintained restatement Done-when now forbids. Do not fix only this site: state the rule — a fact derived at runtime from the registry (entry count, marker list, ARCH_STAR) may appear as a verification criterion about a derived output, never as a hand-maintained requirement in any artifact — and sweep the enumeration. Measured prevalence for this issue: grep over the issue file and atlas returns exactly one demand site (:243); all other "seven"/"7" hits are verification assertions the rule permits. ARCH-PURPOSE.
- **BR-5** [Minor] `atlas-restates-single-source` The map-don't-restate rule BR-1 established has no durable home; the atlas "Adding an entry" checklist still omits the atlas paragraph entirely
  This is the 2nd finding in family `atlas-restates-single-source`. BR-1's instances are genuinely gone (0 shared 12-word spans, and no sibling paragraph restates), so the sweep is clean — what is missing is the rule. atlas/workflow/architecture-principles.md:143 says "a new entry touches: architecture.md, that one list, and the goldens", omitting the map-level paragraph in that same file that every entry since ARCH-MOCK has in fact required. The rule is currently recorded only in this issue's ## Revisions, which archives to workshop/history/ that AGENTS.md marks "don't read", so the next entry-adder reads a three-item checklist and repeats BR-1. Fix the class: add the fourth touch point to :143 stating the shape — this file gets a MAP-level paragraph (boundaries against confusable neighbours, deliberate shaping choices, in-fleet origin) and never a restatement of the entry's clauses, the same route-don't-restate discipline gatefindings.go already follows at :24-30. Measured prevalence: 1 restatement in 1 entry-add under the unwritten rule; 0 remaining after this round's sweep. ARCH-DRY, ARCH-PURPOSE.

## Open findings

- **BR-4** [Minor] `derived-fact-restated` Plan step 5 still lists "the count" as an atlas deliverable and is ticked, contradicting the corrected Done-when
- **BR-5** [Minor] `atlas-restates-single-source` The map-don't-restate rule BR-1 established has no durable home; the atlas "Adding an entry" checklist still omits the atlas paragraph entirely
