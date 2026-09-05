---
gate: plan-quality
issue: 215
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-04T21:53:12-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Eight accepted review deltas have no Plan row, no Done-when criterion, and the Spec's draft entry text is still the pre-delta version
          detail: |-
            Both `## Revisions` rounds say "Confirmed unchanged: the Plan's five steps",
            but the deltas are wording changes to the entry and Plan step 1 points at the
            Spec code block captioned "final wording is the implementer's" — which contains
            none of them. An implementer can tick every checkbox and ship the un-revised
            draft. Fold the eight deltas into the Spec's code block so there is one current
            version of the entry (ARCH-DRY), rather than a stale block plus a changelog the
            reader must replay; or add an explicit Plan row and Done-when criterion enumerating them.
          family: review-deltas-unlanded
          round: 1
        - id: PQ-2
          severity: Important
          title: Done-when and Plan step 3 misstate how the marker reaches its consumers — ARCH_STAR is not the mechanism, and the consumer list is still the one delta 5 flagged
          detail: |-
            `{{ARCH_STAR}}` exists only in `code-review.md:112` (substituted at `review.go:44`);
            plan-quality uses `{{ARCH_BLOCK}}` -> `ArchitectureBlock` (`prompts.go:60`) and
            start-plan calls `judge.ArchitectureBlock` directly (`startplan.go:75`).
            `judge_test.go:328-333` pins four BuildPrompt categories (PlanQuality,
            MilestoneReview, DRY, PURE); the goldens that change are dry/milestone-review/
            plan-quality/pure. Delta 5 named the incomplete list and it was never applied —
            sweep the class (every existing-behavior claim in Spec/Done-when/Plan), including
            that `architectureEntry` (`judge_test.go:195`) is ARCH-CONSTRAINTS-specific
            machinery, not the per-entry contract every marker satisfies
            (`TestArchitectureRegistry_Content`, `judge_test.go:119`). ARCH-PURPOSE.
          family: stale-consumer-claim
          round: 1
        - id: PQ-3
          severity: Minor
          title: Step 3 says "re-capture goldens" without naming which four or requiring the diff be inspected
          detail: |-
            `golden_test.go:35-38` carries an explicit prohibition on using `-update-golden`
            to paper over drift. Name the four registry-bearing goldens
            (dry/milestone-review/plan-quality/pure) and require the recaptured diff to
            contain only the ARCH-ORDER block plus the header count 6 -> 7, so an unrelated
            prompt edit in the same window cannot ride along invisibly.
          family: golden-recapture-blind-sweep
          round: 1
        - id: PQ-4
          severity: Minor
          title: An ARCH-<NAME> token in the new entry's prose without its own heading inflates the derived marker count
          detail: |-
            `markersIn` (`architecture.go:11,29`) scans the whole registry, so a bare marker
            mention becomes a marker. Deltas 1 and 4 deliberately add ARCH-SECURE and
            ARCH-CONSTRAINTS mentions (safe — both have headings), but the Spec's own
            `ARCH-FSM` naming rationale must not migrate into the entry text or
            `TestArchitectureRegistry_Content`'s "found it in prose only" branch fires.
          family: derived-marker-scan-trap
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-04T21:57:25-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Spec block is now the post-revision single version; all eight deltas verified present; Plan step 1 says "verbatim, no replay needed".
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Three-path mechanism enumerated correctly; prompts.go:60, startplan.go:75, review.go:44, judge_test.go:328 and the architectureEntry-vs-TestArchitectureRegistry_Content correction all verified.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Plan step 4 names dry/milestone-review/plan-quality/pure (exactly the four on disk) and requires the diff contain only the ARCH-ORDER block plus 6 -> 7.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: 'Done-when bullet plus Plan step 2 both pin it; the three cited markers all carry ## headings and ARCH-FSM stays out of the entry.'
          round: 2
      blocked: false
content_hash: 1c385faa66586e135c0b4f1e06afe517c83270174c1b7a29e0ded4d1fe528fcf
---

# Gate ledger — ariadne#215 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-04T21:53:12-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `review-deltas-unlanded` Eight accepted review deltas have no Plan row, no Done-when criterion, and the Spec's draft entry text is still the pre-delta version
  Both `## Revisions` rounds say "Confirmed unchanged: the Plan's five steps",
  but the deltas are wording changes to the entry and Plan step 1 points at the
  Spec code block captioned "final wording is the implementer's" — which contains
  none of them. An implementer can tick every checkbox and ship the un-revised
  draft. Fold the eight deltas into the Spec's code block so there is one current
  version of the entry (ARCH-DRY), rather than a stale block plus a changelog the
  reader must replay; or add an explicit Plan row and Done-when criterion enumerating them.
- **PQ-2** [Important] `stale-consumer-claim` Done-when and Plan step 3 misstate how the marker reaches its consumers — ARCH_STAR is not the mechanism, and the consumer list is still the one delta 5 flagged
  `{{ARCH_STAR}}` exists only in `code-review.md:112` (substituted at `review.go:44`);
  plan-quality uses `{{ARCH_BLOCK}}` -> `ArchitectureBlock` (`prompts.go:60`) and
  start-plan calls `judge.ArchitectureBlock` directly (`startplan.go:75`).
  `judge_test.go:328-333` pins four BuildPrompt categories (PlanQuality,
  MilestoneReview, DRY, PURE); the goldens that change are dry/milestone-review/
  plan-quality/pure. Delta 5 named the incomplete list and it was never applied —
  sweep the class (every existing-behavior claim in Spec/Done-when/Plan), including
  that `architectureEntry` (`judge_test.go:195`) is ARCH-CONSTRAINTS-specific
  machinery, not the per-entry contract every marker satisfies
  (`TestArchitectureRegistry_Content`, `judge_test.go:119`). ARCH-PURPOSE.
- **PQ-3** [Minor] `golden-recapture-blind-sweep` Step 3 says "re-capture goldens" without naming which four or requiring the diff be inspected
  `golden_test.go:35-38` carries an explicit prohibition on using `-update-golden`
  to paper over drift. Name the four registry-bearing goldens
  (dry/milestone-review/plan-quality/pure) and require the recaptured diff to
  contain only the ARCH-ORDER block plus the header count 6 -> 7, so an unrelated
  prompt edit in the same window cannot ride along invisibly.
- **PQ-4** [Minor] `derived-marker-scan-trap` An ARCH-<NAME> token in the new entry's prose without its own heading inflates the derived marker count
  `markersIn` (`architecture.go:11,29`) scans the whole registry, so a bare marker
  mention becomes a marker. Deltas 1 and 4 deliberately add ARCH-SECURE and
  ARCH-CONSTRAINTS mentions (safe — both have headings), but the Spec's own
  `ARCH-FSM` naming rationale must not migrate into the entry text or
  `TestArchitectureRegistry_Content`'s "found it in prose only" branch fires.

## Round 2 — 2026-09-04T21:57:25-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Spec block is now the post-revision single version; all eight deltas verified present; Plan step 1 says "verbatim, no replay needed".
- PQ-2 — addressed — Three-path mechanism enumerated correctly; prompts.go:60, startplan.go:75, review.go:44, judge_test.go:328 and the architectureEntry-vs-TestArchitectureRegistry_Content correction all verified.
- PQ-3 — addressed — Plan step 4 names dry/milestone-review/plan-quality/pure (exactly the four on disk) and requires the diff contain only the ARCH-ORDER block plus 6 -> 7.
- PQ-4 — addressed — Done-when bullet plus Plan step 2 both pin it; the three cited markers all carry ## headings and ARCH-FSM stays out of the entry.

## Open findings

(none — every finding has been disposed)
