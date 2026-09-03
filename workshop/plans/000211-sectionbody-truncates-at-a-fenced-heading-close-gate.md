---
gate: boundary-review
issue: 211
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-02T18:59:55-07:00"
      agent: sdlc
      findings:
        - id: BR-1
          severity: Minor
          title: Plan row 3 restates a stale consumer count ("four") that the Spec table has already superseded with six
          detail: |-
            2nd in family. Do not fix the instance: the rule is that the consumer set is
            enumerated in exactly one place, the grep-produced Spec table, and every other
            mention refers to it without restating a count. Measured prevalence 1 of 3
            references non-compliant (Plan row 3; Spec table and Estimate row 2 both comply).
            (carried from plan-quality PQ-5, deferred to the boundary review)
          family: consumer-enumeration-incomplete
          round: 1
        - id: BR-2
          severity: Minor
          title: Corpus-seeded property test does not separate stable invariants from an oracle that goes stale as issues are authored
          detail: |-
            "Concatenating visited + skipped reproduces the input" is content-independent;
            "no file loses a real section" over 406 live workshop files needs an oracle, and
            a golden captured today reddens on the next issue that quotes a fence. State that
            the corpus supplies inputs while assertions stay invariants, and name how the test
            in cmd/sdlc/internal/issue resolves the repo-root path.
            (carried from plan-quality PQ-6, deferred to the boundary review)
          family: mutable-corpus-as-test-oracle
          round: 1
      boundary: '*'
      no_cap: true
      blocked: false
    - "n": 2
      timestamp: "2026-09-02T18:59:55-07:00"
      agent: claude
      boundary: M1
      blocked: false
      protocol_error: no valid findings block
    - "n": 3
      timestamp: "2026-09-02T20:35:10-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Plan row 3 now reads "six sites"; verified the Spec table enumerates exactly six and the Estimate row 2 and Log both agree — 3 of 3 references consistent with the single grep-produced source.
          round: 3
        - id: BR-2
          disposition: addressed
          note: TestSectionBody_CorpusLosesNoRealSection states corpus-supplies-inputs / assertion-is-invariant explicitly and carries no golden; workshopMarkdown names its repo-root resolution and skips on an absent corpus.
          round: 3
      findings:
        - id: BR-3
          severity: Critical
          title: insertLogLine locates the `## Log` heading fence-aware but scans for the day header to EOF, still filing the close line inside a quoted fence
          detail: 'close.go:312 sets section = body[logStart:] ("the real Log section + anything after it"), so dayRE matches a `### <today>` inside a fenced example in a later section. Reproduced: with the real Log lacking today''s day header and an `## Appendix` quoting `### 2026-09-02` in a ```markdown block, the close line is written inside the fence and logHasEntryToday cannot see it. TestInsertLogLine_TargetsTheRealLogSection passes only because its fixture quotes 2020-01-01 while the inserted line is dated 2026-09-02 — a test written from the same mental model as the fix (#194). SectionByteBounds, added in this same diff, already returns the end; bound both the dayRE and insertRE searches to it. ARCH-SECURE: the write path degrades invisibly on hand-authored input.'
          family: unbounded-section-scan-window
          round: 3
        - id: BR-4
          severity: Important
          title: The plan-item fence filter reached 1 of 5 read sites and 0 of 1 write sites, so `sdlc state` and `sdlc close` now disagree about the same Plan
          detail: 2nd finding in family consumer-enumeration-incomplete — do not fix the instance. RULE - the fence filter belongs at the extraction boundary (PlanSectionBody), not at individual call sites, and any code reading or writing `- [s] ...` rows must take its window from SectionLineBounds/SectionByteBounds rather than the whole body. Measured - plan.go:30 CountPlanItems filtered; close.go:568 plan-unchecked, close.go:1727 findMilestonesMissingVerdict, structural.go:160 checkPlan, sizing.go:63-65 all unfiltered; close.go:558 milestone tick writes over the whole body, not scoped to Plan at all. On one fixture CountPlanItems reports 1/1 (state says 100%) while close refuses citing a quoted `- [ ] M9`; checkPlan accepts a Plan whose only items are fenced examples; findMilestonesMissingVerdict would demand an unsatisfiable verdict commit for a fenced milestone. Verified the tick regex rewrites 2 of 2 rows including the fenced one — and workshop/issues/000211-*.md:32 was ticked by an unanchored replace this very commit, breaking the Problem-section demonstration (the table below still says truth = 2 open items). Fix once inside PlanSectionBody and scope the tick to the Plan section, then drop StripFenced from plan.go:30. ARCH-PURPOSE, ARCH-DRY.
          family: consumer-enumeration-incomplete
          round: 3
        - id: BR-5
          severity: Important
          title: The indent-policy parameter has zero call sites and its comment names a caller (SplitFences) that never invokes it; the plan row promising a test pinning the choice is unpinned
          detail: maxIndentAny (fence.go:74) is referenced nowhere outside its declaration; fenceSpans and fenceMarkerIndent are only ever called with commonMarkMaxIndent. SplitFences uses strings.Index and calls neither, so the doc at fence.go:68-71 describes a wiring that does not exist. Meanwhile the ticked Plan row "Decide SplitFences' line-anchoring change explicitly ... with a test pinning the choice" has no test for the indent axis — structural_test.go is untouched in this window and the only added policy test covers the unterminated axis. Delete the dead parameter and add the real pin - an indented ``` is a fence for SplitFences and prose for FenceSpans, asserted side by side. RULE - a policy parameter with no non-default call site is documentation, not code; the axis must be pinned by test, not by an unused constant. ARCH-DRY.
          family: unwired-policy-parameter
          round: 3
        - id: BR-6
          severity: Important
          title: planGateContent's fence-awareness — one of the three swept sites — is covered by no test
          detail: 'changecode.go:740-751 now skips `## ` lines inside fences when computing the plan-gate hash. Verified reachable and correct (an issue quoting `## Estimate` in a fence keeps its real Problem prose), but changecode_test.go:602 only exercises estimate-edit invariance and passes with the fence check reverted. Per #194 the fix is not complete until a test reds without it - assert that real prose following a fenced `## Estimate` survives planGateContent.'
          family: claimed-fix-unpinned-by-test
          round: 3
        - id: BR-7
          severity: Important
          title: insertLogLine's doc block still specifies the last-match anchor that was just deleted, and three docs name a symbol that does not exist
          detail: 'close.go:292-299 still reads "Anchor: the last `## Log` header, not the first ... All offsets below are taken relative to that last header", three lines above the new comment saying the opposite; close.go:331 repeats it; close.go:335 references logHeaderRE, deleted from setstatus.go in this same commit. Separately stripEstimateForHash exists nowhere in the tree (the function is planGateContent) yet is named in atlas/workflow/issue-lifecycle.md:165, the M2 Plan row and the M2 Log. The atlas also asserts "Everything that finds a heading is fence-aware" while issue.go:530''s structure peek deliberately is not. Same class as the M1 review''s I6, recurring in the commit that closed it - delete the superseded prose rather than layering a correction beneath it.'
          family: doc-drifts-from-code
          round: 3
        - id: BR-8
          severity: Minor
          title: TestSectionByteBounds_MatchesSectionBody trims trailing newlines from both sides, so it does not pin the byte-identity section.go:87 claims
          detail: Byte-identity does hold — I verified body[start:end] == SectionBody(...) across all 2886 non-fenced sections in workshop/**/*.md with no panics — so this is a coverage gap, not a defect. Drop the TrimSuffix and fold SectionByteBounds/SectionHeadingByteOffset into the existing corpus walk; they carry all the index arithmetic and are currently exercised by four hand-written bodies.
          family: test-asserts-weaker-than-contract
          round: 3
        - id: BR-9
          severity: Minor
          title: The M2 Log does not record re-running `sdlc issue validate --all` after the stripCodeFences rebase, a Done-when item
          detail: The rebase from `(?s)```.*?``` ` to line-based StripFenced is the one change in this window that could move a gate verdict. It is safe — comparing old against new over every `## Spec` in the corpus, 1 file moves (000147-*.md, 246 to 247 words) and 0 cross the >=50 threshold — but the measurement is absent from the Log the Done-when asks for.
          family: unrecorded-gate-measurement
          round: 3
      boundary: M2
      blocked: true
---

# Gate ledger — ariadne#211 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-02T18:59:55-07:00 (sdlc) — passed

### Raised

- **BR-1** [Minor] `consumer-enumeration-incomplete` Plan row 3 restates a stale consumer count ("four") that the Spec table has already superseded with six
  2nd in family. Do not fix the instance: the rule is that the consumer set is
  enumerated in exactly one place, the grep-produced Spec table, and every other
  mention refers to it without restating a count. Measured prevalence 1 of 3
  references non-compliant (Plan row 3; Spec table and Estimate row 2 both comply).
  (carried from plan-quality PQ-5, deferred to the boundary review)
- **BR-2** [Minor] `mutable-corpus-as-test-oracle` Corpus-seeded property test does not separate stable invariants from an oracle that goes stale as issues are authored
  "Concatenating visited + skipped reproduces the input" is content-independent;
  "no file loses a real section" over 406 live workshop files needs an oracle, and
  a golden captured today reddens on the next issue that quotes a fence. State that
  the corpus supplies inputs while assertions stay invariants, and name how the test
  in cmd/sdlc/internal/issue resolves the repo-root path.
  (carried from plan-quality PQ-6, deferred to the boundary review)

## Round 2 — 2026-09-02T18:59:55-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 3 — 2026-09-02T20:35:10-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Plan row 3 now reads "six sites"; verified the Spec table enumerates exactly six and the Estimate row 2 and Log both agree — 3 of 3 references consistent with the single grep-produced source.
- BR-2 — addressed — TestSectionBody_CorpusLosesNoRealSection states corpus-supplies-inputs / assertion-is-invariant explicitly and carries no golden; workshopMarkdown names its repo-root resolution and skips on an absent corpus.

### Raised

- **BR-3** [Critical] `unbounded-section-scan-window` insertLogLine locates the `## Log` heading fence-aware but scans for the day header to EOF, still filing the close line inside a quoted fence
  close.go:312 sets section = body[logStart:] ("the real Log section + anything after it"), so dayRE matches a `### <today>` inside a fenced example in a later section. Reproduced: with the real Log lacking today's day header and an `## Appendix` quoting `### 2026-09-02` in a ```markdown block, the close line is written inside the fence and logHasEntryToday cannot see it. TestInsertLogLine_TargetsTheRealLogSection passes only because its fixture quotes 2020-01-01 while the inserted line is dated 2026-09-02 — a test written from the same mental model as the fix (#194). SectionByteBounds, added in this same diff, already returns the end; bound both the dayRE and insertRE searches to it. ARCH-SECURE: the write path degrades invisibly on hand-authored input.
- **BR-4** [Important] `consumer-enumeration-incomplete` The plan-item fence filter reached 1 of 5 read sites and 0 of 1 write sites, so `sdlc state` and `sdlc close` now disagree about the same Plan
  2nd finding in family consumer-enumeration-incomplete — do not fix the instance. RULE - the fence filter belongs at the extraction boundary (PlanSectionBody), not at individual call sites, and any code reading or writing `- [s] ...` rows must take its window from SectionLineBounds/SectionByteBounds rather than the whole body. Measured - plan.go:30 CountPlanItems filtered; close.go:568 plan-unchecked, close.go:1727 findMilestonesMissingVerdict, structural.go:160 checkPlan, sizing.go:63-65 all unfiltered; close.go:558 milestone tick writes over the whole body, not scoped to Plan at all. On one fixture CountPlanItems reports 1/1 (state says 100%) while close refuses citing a quoted `- [ ] M9`; checkPlan accepts a Plan whose only items are fenced examples; findMilestonesMissingVerdict would demand an unsatisfiable verdict commit for a fenced milestone. Verified the tick regex rewrites 2 of 2 rows including the fenced one — and workshop/issues/000211-*.md:32 was ticked by an unanchored replace this very commit, breaking the Problem-section demonstration (the table below still says truth = 2 open items). Fix once inside PlanSectionBody and scope the tick to the Plan section, then drop StripFenced from plan.go:30. ARCH-PURPOSE, ARCH-DRY.
- **BR-5** [Important] `unwired-policy-parameter` The indent-policy parameter has zero call sites and its comment names a caller (SplitFences) that never invokes it; the plan row promising a test pinning the choice is unpinned
  maxIndentAny (fence.go:74) is referenced nowhere outside its declaration; fenceSpans and fenceMarkerIndent are only ever called with commonMarkMaxIndent. SplitFences uses strings.Index and calls neither, so the doc at fence.go:68-71 describes a wiring that does not exist. Meanwhile the ticked Plan row "Decide SplitFences' line-anchoring change explicitly ... with a test pinning the choice" has no test for the indent axis — structural_test.go is untouched in this window and the only added policy test covers the unterminated axis. Delete the dead parameter and add the real pin - an indented ``` is a fence for SplitFences and prose for FenceSpans, asserted side by side. RULE - a policy parameter with no non-default call site is documentation, not code; the axis must be pinned by test, not by an unused constant. ARCH-DRY.
- **BR-6** [Important] `claimed-fix-unpinned-by-test` planGateContent's fence-awareness — one of the three swept sites — is covered by no test
  changecode.go:740-751 now skips `## ` lines inside fences when computing the plan-gate hash. Verified reachable and correct (an issue quoting `## Estimate` in a fence keeps its real Problem prose), but changecode_test.go:602 only exercises estimate-edit invariance and passes with the fence check reverted. Per #194 the fix is not complete until a test reds without it - assert that real prose following a fenced `## Estimate` survives planGateContent.
- **BR-7** [Important] `doc-drifts-from-code` insertLogLine's doc block still specifies the last-match anchor that was just deleted, and three docs name a symbol that does not exist
  close.go:292-299 still reads "Anchor: the last `## Log` header, not the first ... All offsets below are taken relative to that last header", three lines above the new comment saying the opposite; close.go:331 repeats it; close.go:335 references logHeaderRE, deleted from setstatus.go in this same commit. Separately stripEstimateForHash exists nowhere in the tree (the function is planGateContent) yet is named in atlas/workflow/issue-lifecycle.md:165, the M2 Plan row and the M2 Log. The atlas also asserts "Everything that finds a heading is fence-aware" while issue.go:530's structure peek deliberately is not. Same class as the M1 review's I6, recurring in the commit that closed it - delete the superseded prose rather than layering a correction beneath it.
- **BR-8** [Minor] `test-asserts-weaker-than-contract` TestSectionByteBounds_MatchesSectionBody trims trailing newlines from both sides, so it does not pin the byte-identity section.go:87 claims
  Byte-identity does hold — I verified body[start:end] == SectionBody(...) across all 2886 non-fenced sections in workshop/**/*.md with no panics — so this is a coverage gap, not a defect. Drop the TrimSuffix and fold SectionByteBounds/SectionHeadingByteOffset into the existing corpus walk; they carry all the index arithmetic and are currently exercised by four hand-written bodies.
- **BR-9** [Minor] `unrecorded-gate-measurement` The M2 Log does not record re-running `sdlc issue validate --all` after the stripCodeFences rebase, a Done-when item
  The rebase from `(?s)```.*?``` ` to line-based StripFenced is the one change in this window that could move a gate verdict. It is safe — comparing old against new over every `## Spec` in the corpus, 1 file moves (000147-*.md, 246 to 247 words) and 0 cross the >=50 threshold — but the measurement is absent from the Log the Done-when asks for.

## Open findings

- **BR-3** [Critical] `unbounded-section-scan-window` insertLogLine locates the `## Log` heading fence-aware but scans for the day header to EOF, still filing the close line inside a quoted fence
- **BR-4** [Important] `consumer-enumeration-incomplete` The plan-item fence filter reached 1 of 5 read sites and 0 of 1 write sites, so `sdlc state` and `sdlc close` now disagree about the same Plan
- **BR-5** [Important] `unwired-policy-parameter` The indent-policy parameter has zero call sites and its comment names a caller (SplitFences) that never invokes it; the plan row promising a test pinning the choice is unpinned
- **BR-6** [Important] `claimed-fix-unpinned-by-test` planGateContent's fence-awareness — one of the three swept sites — is covered by no test
- **BR-7** [Important] `doc-drifts-from-code` insertLogLine's doc block still specifies the last-match anchor that was just deleted, and three docs name a symbol that does not exist
- **BR-8** [Minor] `test-asserts-weaker-than-contract` TestSectionByteBounds_MatchesSectionBody trims trailing newlines from both sides, so it does not pin the byte-identity section.go:87 claims
- **BR-9** [Minor] `unrecorded-gate-measurement` The M2 Log does not record re-running `sdlc issue validate --all` after the stripCodeFences rebase, a Done-when item
