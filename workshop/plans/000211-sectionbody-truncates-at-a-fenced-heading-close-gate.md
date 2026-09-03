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
    - "n": 4
      timestamp: "2026-09-02T21:03:11-07:00"
      agent: claude
      dispose:
        - id: BR-3
          disposition: addressed
          note: 'Revert-verified: restoring body[logStart:] reds TestInsertLogLine_IgnoresAQuotedDayHeaderLater.'
          round: 4
        - id: BR-4
          disposition: not-addressed
          note: Readers routed to PlanItemsBody, but the tick writer is unscoped, no test pins the routing, and issue line 32 is still ticked.
          round: 4
        - id: BR-5
          disposition: not-addressed
          note: Dead parameter and false comment removed; the indent-axis pin the finding asked for is still absent while the plan row claiming it stays ticked.
          round: 4
        - id: BR-6
          disposition: addressed
          note: 'Revert-verified: dropping the !inside[i] guard reds TestPlanGateContent_IgnoresAQuotedEstimateHeading.'
          round: 4
        - id: BR-7
          disposition: not-addressed
          note: close.go:292 and :331 fixed; close.go:340 still names logHeaderRE (dead), and stripEstimateForHash survives at atlas:168 plus issue lines 289 and 333.
          round: 4
        - id: BR-8
          disposition: not-addressed
          note: Test still TrimSuffixes both sides; byte-identity independently re-verified on 9 edge shapes including CRLF.
          round: 4
        - id: BR-9
          disposition: not-addressed
          note: M2 Log still records no validate run; I measured it — PlanItemsBody routing leaves `sdlc issue validate --all` byte-identical.
          round: 4
      findings:
        - id: BR-10
          severity: Important
          title: BR-4's consumer routing is pinned by no test — reverting all four call sites leaves the suite green
          detail: |-
            2nd in family. RULE - a test standing in for a production path must call the same
            extractor that path calls, and its fixture must contain the shape that extractor
            exists to filter. Measured - close_test.go:240,257,440 and planfence_test.go:44,63
            all call PlanSectionBody while close.go:572, close.go:1726, plan.go:26,
            structural.go:160 and sizing.go:63 now call PlanItemsBody; close_test.go:238's
            comment still claims it routes "through the production extractor". Reverting every
            production routing to PlanSectionBody changes no test result. A correctly-written
            pin (fixture with a fenced `- [ ] M9` row, read through PlanItemsBody) reds on
            revert with CountPlanItems = (5,3) want (3,3) — I verified both directions in
            scratch. ARCH-MOCK - production and test flow no longer share the boundary.
          family: claimed-fix-unpinned-by-test
          round: 4
        - id: BR-11
          severity: Important
          title: Bounding a search to a fence-aware section does not make the search fence-aware — four within-section matchers remain unfiltered
          detail: |-
            3rd in family — do NOT fix these sites individually. RULE - any structural match
            run INSIDE an extracted section must go over StripFenced(section) or a
            FenceSpans-filtered line walk, not the raw section text. Measured members -
            close.go:328 dayRE (a fenced `### <date>` inside the real Log section still
            captures the close-line insert, which is BR-3's defect one level further down),
            project/guards.go:17 retroHeadingRE (retro-recorded is a presence-means-pass guard,
            so a quoted retro heading satisfies it), project/retro.go:8 retroDateRE, and
            close.go:547 (see the sibling finding). Prevalence across all 409 workshop/**/*.md -
            0 fenced day-headers in a Log section, 0 fenced retro headings; latent today,
            exactly as the `## Plan` truncation was latent when this issue was filed. Write the
            helper once and sweep the enumeration.
          family: consumer-enumeration-incomplete
          round: 4
        - id: BR-12
          severity: Important
          title: The milestone tick rewrites every matching row in the whole body, and its "rewrites a line it already matched" rationale is false
          detail: |-
            close.go:547-552 does pat.ReplaceAllString(newBody, ...) with no Plan scoping and no
            fence filter, so it rewrites ALL matches anywhere in the document — not one already
            matched line. Verified - a body quoting `- [ ] M2` in a fenced Problem-section
            example plus one real row ticks both (matches=2); a body whose ONLY `- [ ] M3` is
            fenced still yields n>0, so close prints `ticked M3 in ... ## Plan` while the real
            Plan is untouched. PlanItemsBody sees only the real row, so writer and readers now
            disagree about the same Plan — the disagreement BR-4 was raised to end. Scope the
            replace to SectionByteBounds(body, "Plan", ...) and skip FenceSpans-inside lines.
          family: consumer-enumeration-incomplete
          round: 4
        - id: BR-13
          severity: Important
          title: Done-when cites a `## Revisions` section that does not exist, and a ticked plan row still claims SplitFences was rebuilt
          detail: |-
            2nd in family. RULE - when a decision reverses what an artifact says, delete or
            rewrite the superseded text in the same commit; never layer a correction beneath it,
            and never leave a cross-reference pointing at a section you did not write. Measured -
            the issue file's headings are Problem, Spec, Done when, Estimate, Plan, Log, with no
            `## Revisions`, yet Done-when line 190 says "a deliberate, recorded exception ... (see
            Revisions)" and AGENTS.md section 1 requires appending one on mid-stream revision (the
            Done-when was rewritten in place). Plan line 296 is ticked claiming "Rebuild
            stripCodeFences + SplitFences on the scanner" while the Log explains SplitFences was
            deliberately not rebuilt. Same rule covers BR-7's residue - close.go:340 names the
            deleted logHeaderRE, and stripEstimateForHash (a symbol that exists nowhere) survives
            at atlas/workflow/issue-lifecycle.md:168 and issue lines 289 and 333.
          family: doc-drifts-from-code
          round: 4
        - id: BR-14
          severity: Minor
          title: insertLogLine computes the Log section twice and discards the second ok, leaving a latent slice panic
          detail: |-
            close.go:310 calls SectionHeadingByteOffset (which internally calls
            SectionByteBounds), then close.go:317 calls SectionByteBounds again with `_` for ok.
            The two agree today because the arguments are identical, so the panic is
            unreachable — but body[logStart:0] is what a false ok would produce. Take
            start/end/ok from one SectionByteBounds call and derive the heading offset from it.
            ARCH-DRY, ARCH-SECURE.
          family: redundant-recompute-drops-error
          round: 4
        - id: BR-15
          severity: Minor
          title: Stale line reference and a missing entry in the atlas's fence-aware inventory
          detail: |-
            workshop/issues/000211-...:37 cites close.go:563 for the plan-unchecked guard, now
            close.go:572. atlas/workflow/issue-lifecycle.md:166-168 lists the fence-aware
            heading finders but omits PlanItemsBody, the single extraction point M2 introduced.
          family: doc-drifts-from-code
          round: 4
      boundary: M2
      blocked: true
    - "n": 5
      timestamp: "2026-09-02T22:25:55-07:00"
      agent: claude
      boundary: M2
      blocked: true
      protocol_error: no valid findings block
    - "n": 6
      timestamp: "2026-09-02T22:46:09-07:00"
      agent: claude
      dispose:
        - id: BR-4
          disposition: addressed
          note: All five readers route through PlanItemsBody and the tick is Plan-scoped; residual corrupted demo moved under BR-15.
          round: 6
        - id: BR-5
          disposition: not-addressed
          note: 'maxIndentAny is gone (grep: zero references), but the promised indent-axis pin still does not exist — TestSplitFences (structural_test.go:188) has no indented case and no side-by-side FenceSpans counterpart.'
          round: 6
        - id: BR-7
          disposition: not-addressed
          note: close.go:292/:331 fixed; close.go:342 still names logHeaderRE (absent from the tree), atlas:184 still names stripEstimateForHash (function is planGateContent, changecode.go), and atlas:181's "Everything that finds a heading is fence-aware" is still unqualified while issue.go:530 is deliberately not.
          round: 6
        - id: BR-8
          disposition: not-addressed
          note: fence_test.go:228 still TrimSuffix-es both sides; SectionByteBounds/SectionHeadingByteOffset are still exercised only by the four hand-written bodies, not the corpus walk.
          round: 6
        - id: BR-9
          disposition: not-addressed
          note: No M2 Log entry records re-running `sdlc issue validate --all`; the only record (issue :434, :448) is M1's.
          round: 6
        - id: BR-10
          disposition: not-addressed
          note: Measured by revert - reverting close.go:596, close.go:1750, sizing.go:63 and structural.go:160 to PlanSectionBody leaves the entire suite green. Only CountPlanItems (plan.go:26) reds, at (4,2) want (2,1). TestPlanItemReadersAgree extracts planBody itself and applies the regexes inline, so it pins 1 of 5 readers against a comment claiming 4.
          round: 6
        - id: BR-11
          disposition: not-addressed
          note: 2 of the 4 enumerated members swept (close.go:328 dayRE, close.go:547 tick) plus logHasEntryToday; project/guards.go:17 retroHeadingRE and project/retro.go:8 retroDateRE still match against raw d.SectionBody("Log").
          round: 6
        - id: BR-12
          disposition: not-addressed
          note: The close.go:564-588 fix is correct and corpus-verified non-regressive, but unpinned - reverting it to the whole-body ReplaceAll leaves TestMilestoneTick_OnlyTicksTheRealPlan PASSING, because the test re-implements the loop rather than calling computeClose.
          round: 6
        - id: BR-13
          disposition: not-addressed
          note: Done-when's Revisions cross-reference and the SplitFences plan row are both corrected; close.go:342 (logHeaderRE), atlas:184 + issue :289/:369 (stripEstimateForHash) survive, and section.go:72-74 still says the tick writer "keeps PlanSectionBody ... rewrites a specific line it already matched" — the rationale close.go:568 declares false in the same commit, about code that no longer calls PlanSectionBody at all.
          round: 6
        - id: BR-14
          disposition: not-addressed
          note: close.go:308 and close.go:315 still compute the Log bounds twice, second call discarding ok.
          round: 6
        - id: BR-15
          disposition: not-addressed
          note: Atlas now lists PlanItemsBody (addressed), but issue :37 and :143 still cite close.go:563 (guard is at close.go:596) and :32 still carries the tick-corrupted `- [x] M2 — NOT done` under a table asserting 2 open items.
          round: 6
      findings:
        - id: BR-16
          severity: Important
          title: The M2 Log records a revert-verification for BR-10 that I measured to be false
          detail: |-
            This is the 2nd finding in family `unrecorded-gate-measurement`. Do not fix
            the sentence - fix the rule. RULE: a Log line asserting evidence must name
            the command run and the observed result, and a "revert-verified" claim is
            only true when the named test went RED with the production change undone.
            Measured: issue :332-341 says BR-10 and BR-11 were "Both revert-verified
            with an explicit build check first". BR-11 is true (I reverted close.go's
            dayRE and setstatus.go's StripFenced; planfence_test.go:310 reds on both
            assertions). BR-10 is false (reverting close.go:596, close.go:1750,
            sizing.go:63, structural.go:160 leaves every test green). Prevalence in
            this issue: 2 of 3 revert claims in the M2 Log are unverifiable as written
            - the third, "Revert-verified on the live instance" at :368, names no test.
            A build check is not a revert check.
          family: unrecorded-gate-measurement
          round: 6
        - id: BR-17
          severity: Important
          title: FindLineOutsideFences documents a line range and returns a match range, on a splice-offset API
          detail: |-
            This is the 4th finding in family `doc-drifts-from-code`. Do not fix this
            instance alone. RULE: every symbol name and every behavioral claim written
            into a comment, the atlas or an issue must be greppable-true against the
            tree at the commit that writes it; the enumeration is one grep per named
            symbol plus one check per "everything/all X" claim. Prevalence at HEAD, all
            four members verified by grep: fence.go:156 says "byte range of the first
            LINE matching re" while :171 returns off+loc[0], off+loc[1] (the match);
            close.go:342 names logHeaderRE, absent from the tree; atlas:184 and issue
            :289/:369 name stripEstimateForHash, absent (it is planGateContent);
            section.go:72-74 states a tick-writer rationale that close.go:568 calls
            false in the same commit. The FindLineOutsideFences case is the only one
            with a byte-corruption path: the sole caller anchors ^...$ so match == line
            today, but close.go:333 splices at the returned offset, so an unanchored
            regex from a future caller inserts mid-line into an issue file. Either
            return the line bounds or say "match range" and rename the returns.
          family: doc-drifts-from-code
          round: 6
      boundary: M2
      blocked: false
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

## Round 4 — 2026-09-02T21:03:11-07:00 (claude) — BLOCKED

### Disposed

- BR-3 — addressed — Revert-verified: restoring body[logStart:] reds TestInsertLogLine_IgnoresAQuotedDayHeaderLater.
- BR-4 — not-addressed — Readers routed to PlanItemsBody, but the tick writer is unscoped, no test pins the routing, and issue line 32 is still ticked.
- BR-5 — not-addressed — Dead parameter and false comment removed; the indent-axis pin the finding asked for is still absent while the plan row claiming it stays ticked.
- BR-6 — addressed — Revert-verified: dropping the !inside[i] guard reds TestPlanGateContent_IgnoresAQuotedEstimateHeading.
- BR-7 — not-addressed — close.go:292 and :331 fixed; close.go:340 still names logHeaderRE (dead), and stripEstimateForHash survives at atlas:168 plus issue lines 289 and 333.
- BR-8 — not-addressed — Test still TrimSuffixes both sides; byte-identity independently re-verified on 9 edge shapes including CRLF.
- BR-9 — not-addressed — M2 Log still records no validate run; I measured it — PlanItemsBody routing leaves `sdlc issue validate --all` byte-identical.

### Raised

- **BR-10** [Important] `claimed-fix-unpinned-by-test` BR-4's consumer routing is pinned by no test — reverting all four call sites leaves the suite green
  2nd in family. RULE - a test standing in for a production path must call the same
  extractor that path calls, and its fixture must contain the shape that extractor
  exists to filter. Measured - close_test.go:240,257,440 and planfence_test.go:44,63
  all call PlanSectionBody while close.go:572, close.go:1726, plan.go:26,
  structural.go:160 and sizing.go:63 now call PlanItemsBody; close_test.go:238's
  comment still claims it routes "through the production extractor". Reverting every
  production routing to PlanSectionBody changes no test result. A correctly-written
  pin (fixture with a fenced `- [ ] M9` row, read through PlanItemsBody) reds on
  revert with CountPlanItems = (5,3) want (3,3) — I verified both directions in
  scratch. ARCH-MOCK - production and test flow no longer share the boundary.
- **BR-11** [Important] `consumer-enumeration-incomplete` Bounding a search to a fence-aware section does not make the search fence-aware — four within-section matchers remain unfiltered
  3rd in family — do NOT fix these sites individually. RULE - any structural match
  run INSIDE an extracted section must go over StripFenced(section) or a
  FenceSpans-filtered line walk, not the raw section text. Measured members -
  close.go:328 dayRE (a fenced `### <date>` inside the real Log section still
  captures the close-line insert, which is BR-3's defect one level further down),
  project/guards.go:17 retroHeadingRE (retro-recorded is a presence-means-pass guard,
  so a quoted retro heading satisfies it), project/retro.go:8 retroDateRE, and
  close.go:547 (see the sibling finding). Prevalence across all 409 workshop/**/*.md -
  0 fenced day-headers in a Log section, 0 fenced retro headings; latent today,
  exactly as the `## Plan` truncation was latent when this issue was filed. Write the
  helper once and sweep the enumeration.
- **BR-12** [Important] `consumer-enumeration-incomplete` The milestone tick rewrites every matching row in the whole body, and its "rewrites a line it already matched" rationale is false
  close.go:547-552 does pat.ReplaceAllString(newBody, ...) with no Plan scoping and no
  fence filter, so it rewrites ALL matches anywhere in the document — not one already
  matched line. Verified - a body quoting `- [ ] M2` in a fenced Problem-section
  example plus one real row ticks both (matches=2); a body whose ONLY `- [ ] M3` is
  fenced still yields n>0, so close prints `ticked M3 in ... ## Plan` while the real
  Plan is untouched. PlanItemsBody sees only the real row, so writer and readers now
  disagree about the same Plan — the disagreement BR-4 was raised to end. Scope the
  replace to SectionByteBounds(body, "Plan", ...) and skip FenceSpans-inside lines.
- **BR-13** [Important] `doc-drifts-from-code` Done-when cites a `## Revisions` section that does not exist, and a ticked plan row still claims SplitFences was rebuilt
  2nd in family. RULE - when a decision reverses what an artifact says, delete or
  rewrite the superseded text in the same commit; never layer a correction beneath it,
  and never leave a cross-reference pointing at a section you did not write. Measured -
  the issue file's headings are Problem, Spec, Done when, Estimate, Plan, Log, with no
  `## Revisions`, yet Done-when line 190 says "a deliberate, recorded exception ... (see
  Revisions)" and AGENTS.md section 1 requires appending one on mid-stream revision (the
  Done-when was rewritten in place). Plan line 296 is ticked claiming "Rebuild
  stripCodeFences + SplitFences on the scanner" while the Log explains SplitFences was
  deliberately not rebuilt. Same rule covers BR-7's residue - close.go:340 names the
  deleted logHeaderRE, and stripEstimateForHash (a symbol that exists nowhere) survives
  at atlas/workflow/issue-lifecycle.md:168 and issue lines 289 and 333.
- **BR-14** [Minor] `redundant-recompute-drops-error` insertLogLine computes the Log section twice and discards the second ok, leaving a latent slice panic
  close.go:310 calls SectionHeadingByteOffset (which internally calls
  SectionByteBounds), then close.go:317 calls SectionByteBounds again with `_` for ok.
  The two agree today because the arguments are identical, so the panic is
  unreachable — but body[logStart:0] is what a false ok would produce. Take
  start/end/ok from one SectionByteBounds call and derive the heading offset from it.
  ARCH-DRY, ARCH-SECURE.
- **BR-15** [Minor] `doc-drifts-from-code` Stale line reference and a missing entry in the atlas's fence-aware inventory
  workshop/issues/000211-...:37 cites close.go:563 for the plan-unchecked guard, now
  close.go:572. atlas/workflow/issue-lifecycle.md:166-168 lists the fence-aware
  heading finders but omits PlanItemsBody, the single extraction point M2 introduced.

## Round 5 — 2026-09-02T22:25:55-07:00 (claude) — BLOCKED

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 6 — 2026-09-02T22:46:09-07:00 (claude) — passed

### Disposed

- BR-4 — addressed — All five readers route through PlanItemsBody and the tick is Plan-scoped; residual corrupted demo moved under BR-15.
- BR-5 — not-addressed — maxIndentAny is gone (grep: zero references), but the promised indent-axis pin still does not exist — TestSplitFences (structural_test.go:188) has no indented case and no side-by-side FenceSpans counterpart.
- BR-7 — not-addressed — close.go:292/:331 fixed; close.go:342 still names logHeaderRE (absent from the tree), atlas:184 still names stripEstimateForHash (function is planGateContent, changecode.go), and atlas:181's "Everything that finds a heading is fence-aware" is still unqualified while issue.go:530 is deliberately not.
- BR-8 — not-addressed — fence_test.go:228 still TrimSuffix-es both sides; SectionByteBounds/SectionHeadingByteOffset are still exercised only by the four hand-written bodies, not the corpus walk.
- BR-9 — not-addressed — No M2 Log entry records re-running `sdlc issue validate --all`; the only record (issue :434, :448) is M1's.
- BR-10 — not-addressed — Measured by revert - reverting close.go:596, close.go:1750, sizing.go:63 and structural.go:160 to PlanSectionBody leaves the entire suite green. Only CountPlanItems (plan.go:26) reds, at (4,2) want (2,1). TestPlanItemReadersAgree extracts planBody itself and applies the regexes inline, so it pins 1 of 5 readers against a comment claiming 4.
- BR-11 — not-addressed — 2 of the 4 enumerated members swept (close.go:328 dayRE, close.go:547 tick) plus logHasEntryToday; project/guards.go:17 retroHeadingRE and project/retro.go:8 retroDateRE still match against raw d.SectionBody("Log").
- BR-12 — not-addressed — The close.go:564-588 fix is correct and corpus-verified non-regressive, but unpinned - reverting it to the whole-body ReplaceAll leaves TestMilestoneTick_OnlyTicksTheRealPlan PASSING, because the test re-implements the loop rather than calling computeClose.
- BR-13 — not-addressed — Done-when's Revisions cross-reference and the SplitFences plan row are both corrected; close.go:342 (logHeaderRE), atlas:184 + issue :289/:369 (stripEstimateForHash) survive, and section.go:72-74 still says the tick writer "keeps PlanSectionBody ... rewrites a specific line it already matched" — the rationale close.go:568 declares false in the same commit, about code that no longer calls PlanSectionBody at all.
- BR-14 — not-addressed — close.go:308 and close.go:315 still compute the Log bounds twice, second call discarding ok.
- BR-15 — not-addressed — Atlas now lists PlanItemsBody (addressed), but issue :37 and :143 still cite close.go:563 (guard is at close.go:596) and :32 still carries the tick-corrupted `- [x] M2 — NOT done` under a table asserting 2 open items.

### Raised

- **BR-16** [Important] `unrecorded-gate-measurement` The M2 Log records a revert-verification for BR-10 that I measured to be false
  This is the 2nd finding in family `unrecorded-gate-measurement`. Do not fix
  the sentence - fix the rule. RULE: a Log line asserting evidence must name
  the command run and the observed result, and a "revert-verified" claim is
  only true when the named test went RED with the production change undone.
  Measured: issue :332-341 says BR-10 and BR-11 were "Both revert-verified
  with an explicit build check first". BR-11 is true (I reverted close.go's
  dayRE and setstatus.go's StripFenced; planfence_test.go:310 reds on both
  assertions). BR-10 is false (reverting close.go:596, close.go:1750,
  sizing.go:63, structural.go:160 leaves every test green). Prevalence in
  this issue: 2 of 3 revert claims in the M2 Log are unverifiable as written
  - the third, "Revert-verified on the live instance" at :368, names no test.
  A build check is not a revert check.
- **BR-17** [Important] `doc-drifts-from-code` FindLineOutsideFences documents a line range and returns a match range, on a splice-offset API
  This is the 4th finding in family `doc-drifts-from-code`. Do not fix this
  instance alone. RULE: every symbol name and every behavioral claim written
  into a comment, the atlas or an issue must be greppable-true against the
  tree at the commit that writes it; the enumeration is one grep per named
  symbol plus one check per "everything/all X" claim. Prevalence at HEAD, all
  four members verified by grep: fence.go:156 says "byte range of the first
  LINE matching re" while :171 returns off+loc[0], off+loc[1] (the match);
  close.go:342 names logHeaderRE, absent from the tree; atlas:184 and issue
  :289/:369 name stripEstimateForHash, absent (it is planGateContent);
  section.go:72-74 states a tick-writer rationale that close.go:568 calls
  false in the same commit. The FindLineOutsideFences case is the only one
  with a byte-corruption path: the sole caller anchors ^...$ so match == line
  today, but close.go:333 splices at the returned offset, so an unanchored
  regex from a future caller inserts mid-line into an issue file. Either
  return the line bounds or say "match range" and rename the returns.

## Open findings

- **BR-5** [Important] `unwired-policy-parameter` The indent-policy parameter has zero call sites and its comment names a caller (SplitFences) that never invokes it; the plan row promising a test pinning the choice is unpinned
- **BR-7** [Important] `doc-drifts-from-code` insertLogLine's doc block still specifies the last-match anchor that was just deleted, and three docs name a symbol that does not exist
- **BR-8** [Minor] `test-asserts-weaker-than-contract` TestSectionByteBounds_MatchesSectionBody trims trailing newlines from both sides, so it does not pin the byte-identity section.go:87 claims
- **BR-9** [Minor] `unrecorded-gate-measurement` The M2 Log does not record re-running `sdlc issue validate --all` after the stripCodeFences rebase, a Done-when item
- **BR-10** [Important] `claimed-fix-unpinned-by-test` BR-4's consumer routing is pinned by no test — reverting all four call sites leaves the suite green
- **BR-11** [Important] `consumer-enumeration-incomplete` Bounding a search to a fence-aware section does not make the search fence-aware — four within-section matchers remain unfiltered
- **BR-12** [Important] `consumer-enumeration-incomplete` The milestone tick rewrites every matching row in the whole body, and its "rewrites a line it already matched" rationale is false
- **BR-13** [Important] `doc-drifts-from-code` Done-when cites a `## Revisions` section that does not exist, and a ticked plan row still claims SplitFences was rebuilt
- **BR-14** [Minor] `redundant-recompute-drops-error` insertLogLine computes the Log section twice and discards the second ok, leaving a latent slice panic
- **BR-15** [Minor] `doc-drifts-from-code` Stale line reference and a missing entry in the atlas's fence-aware inventory
- **BR-16** [Important] `unrecorded-gate-measurement` The M2 Log records a revert-verification for BR-10 that I measured to be false
- **BR-17** [Important] `doc-drifts-from-code` FindLineOutsideFences documents a line range and returns a match range, on a splice-offset API
