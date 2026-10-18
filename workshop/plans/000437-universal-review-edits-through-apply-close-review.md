# Boundary Review — ariadne#437 (whole-issue close)

| field | value |
|-------|-------|
| issue | 437 — xx-universal-review edits through couch --session-sync apply; retire deletions: propose-strike |
| repo | ariadne |
| issue file | workshop/issues/000437-universal-review-edits-through-apply.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2e783ac22e1a79d3a6cddce34016f5353679415c..930ab339c87d9cedeae94897f2e226cf3e8a6272 |
| command | sdlc close --issue 437 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T12:15:40-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: medium
```

The exporter migration, shared proposal-link parser, and marker-based examples pass their focused checks. Shipping is blocked by incomplete proposal handling outside the marker-based path and recovery instructions that can mistake an earlier round for a successful apply. README coverage is also missing.

```findings
findings:
  - id: new
    severity: Critical
    family: presentation-all-write-paths
    title: |
      Untracked edits bypass the document's proposal presentation
    detail: |
      construct/local/universal-review/round.md:50-56 now requires clean edits, forbids agent-written proposal markers, and promises proposals under presentation: propose. However, adapters/file-edit.md:124-127 still directs untracked documents straight through harness file tools, with no converter. An unregistered document outside Couch therefore receives clean replacements instead of proposals; round.md:115 also requires round --json output that this path cannot produce. The family includes the submitted untracked document and any untracked parent routed through adapters/file-edit.md:115-118. Define a supported presentation and reporting path for both, and exercise propose on tracked and untracked documents. ARCH-PURPOSE: the document-level contract must reach every supported writer.
  - id: new
    severity: Critical
    family: structural-links-without-markers
    title: |
      The proposal fallback lacks a supported path when no ready marker exists
    detail: |
      adapters/file-edit.md:95-99 unconditionally routes large or unencodable rewrites to an applied proposal link, but its direct-link exception at lines 62-69 requires an operator's ready marker. A marker-free developmental pass is explicitly supported by round.md:45 and modes/developmental.md:7-10, yet cannot use that exception. The same gap occurs in fix/SKILL.md:229's hosted fallback and collaboration-modes.md:135-137's marker-free chat answer links, routed through the shared link contract at lines 204-215. Provide a supported marker-free structural-link path, preserving link/index consistency, and add conformance cases without ready markers. The new large-proposal fixture covers only the marker-present case. ARCH-PURPOSE: sweep these consumers together.
  - id: new
    severity: Critical
    family: apply-recovery-operation-correlation
    title: |
      A current stored round does not prove the attempted apply landed
    detail: |
      construct/local/fix/SKILL.md:234-238 instructs the agent to commit stored Body whenever Current is true after a lost response. If a previous round is still current and the new apply never reaches Pair, these instructions commit the previous Body while leaving the new requested edits unapplied. Branch/file checks cannot distinguish those operations. The family includes lost-response recovery here and interrupted-commit recovery at lines 258-261, which also needs the successful operation's identity. Correlate the stored round with the attempted apply using a request or round identity and its base/result hashes; test both an old current round with an unreceived request and a newly landed request with a lost response.
  - id: new
    severity: Important
    family: user-surface-readme-coverage
    title: |
      README update appears missing for the new review surface
    detail: |
      README.md is unchanged in the pinned range and contains no presentation or review-apply guidance. This window introduces the document presentation key and the pair review apply / pair review round workflow, documented in round.md:50-62 and fix/SKILL.md:211-243. These are all instances of the new user-facing surface requiring README coverage under this boundary's docs gate. Add concise usage and migration guidance, linking to the updated atlas.
```

1. **Strengths**

   - V3 explicitly rejects retired mode-level `deletions` and omits it from exported records.
   - Proposal links reuse `pkg/sidedoc` rather than introducing a parallel parser.
   - The tracked adapter clearly distinguishes invalid, stale, locked, and operator-owned refusals.
   - Atlas coverage is updated; its count of 21 reference examples is correct.

2. **Critical findings**

   The three findings above require supported paths and regression evidence before crossing the boundary.

3. **Important findings**

   Add README coverage for the enumerated commands and configuration key.

4. **Minor findings**

   None.

5. **Test coverage notes**

   `go test ./pkg/reviewexport ./pkg/sidedoc -count=1` and `git diff --check` passed. The contract test checks phrases; the examples checker validates static trees, so neither establishes the missing execution paths above.

   The supplemental weave suite was stopped without results. Its focused skill-discovery test failed because its temporary derived directory was not recognized as a Git repository. That test is unchanged in this range; I have not attributed its failure to this patch. The checkout remains unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared export and side-doc machinery remains consolidated.
   - **ARCH-PURE — pass:** changed Go logic remains directly testable over text or injected filesystems; no new business logic is buried in IO.
   - **ARCH-PURPOSE — flag:** presentation and structural-link handling do not cover every supported consumer. The plan’s listed entities and seams otherwise exist at their stated locations.

7. **Plan revision recommendations**

   Add `## Revisions` entries defining untracked presentation/reporting, marker-free structural-link handling across both adapters and chat, and operation-correlated apply recovery. Reconcile Task 4’s unchecked items with measured completion evidence.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d2cda983-e7ec-413e-bb89-5e6b9abf4157","repository":"github.com/xianxu/ariadne","issue_id":"000437","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/73816bf80a42a6d2","reviewed_head":"930ab339c87d9cedeae94897f2e226cf3e8a6272","slot":"ariadne:1","requested_at":"2026-10-18T19:10:24.93917Z"},"phase":"finished","pid":93984,"started_at":"2026-10-18T19:10:25.141143Z","finished_at":"2026-10-18T19:15:40.044587Z","session_id":"01a1506c-1efa-7f80-886d-d33831248123","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T13:30:32-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 437 — xx-universal-review edits through couch --session-sync apply; retire deletions: propose-strike |
| repo | ariadne |
| issue file | workshop/issues/000437-universal-review-edits-through-apply.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2e783ac22e1a79d3a6cddce34016f5353679415c..0db63032d318ffa5a1da0b8b10f36153f8b15e15 |
| command | sdlc close --issue 437 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T13:30:32-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The focused suites pass, the v3 export is current, and README coverage is repaired. Shipping remains blocked by missing behavioral regression evidence for BR-1 and BR-3, plus a reproduced parser bug that drops an index entry while its proposed link remains pending.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      The refusal and reporting instructions now cover submitted documents and untracked parents at adapters/file-edit.md:119-138. Behavioral regression evidence is still missing: TestReviewSkillCloseRecoveryContract only checks phrases, and TestExamples only validates the stored unchanged fixture. In a temporary copy, replacing the refusal with unconditional direct editing while retaining its keywords in a comment left both tests green. Exercise tracked/untracked proposal routing and untracked parent routing through a meaningful skill scenario.
  - id: BR-2
    disposition: addressed
    note: |
      Both adapters and collaboration-modes.md:216-222 now support pending marker-free links, atomic index entries, and rejection cleanup. The new SideLinks and ReconcileIndex tests exercise pending-link retention and rejection; restoring the base parser through a temporary Go overlay makes those tests fail. A separate multiplicity defect is reported below.
  - id: BR-3
    disposition: not-addressed
    note: |
      fix/SKILL.md:234-245 and 269-275 now require attempted-base correlation and readback for both lost-response and interrupted-commit recovery. However, no test exercises the required old-current/unreceived-request and newly-landed/lost-response scenarios. TestReviewSkillCloseRecoveryContract merely searches for keywords. Restoring unsafe Current-only recovery in a temporary copy, with those keywords retained in a comment, still passes. Add scenario evidence that rejects the old round, admits the matching landed round, and handles changed bytes without a recorded round.
  - id: BR-4
    disposition: addressed
    note: |
      README.md:15-20 now introduces presentation: propose, Couch apply, hosted Pair apply, and untracked migration, linking to the atlas. atlas/workflow/universal-review.md:41-79 explains the apply/round workflow and recovery checks, consistent with the revised adapters.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: structural-links-without-markers
    title: |
      A pending proposal containing multiple links loses later index entries
    detail: |
      pkg/sidedoc/links.go:21,51-52 consumes the entire pending marker while capturing only its first escaped link. Reproduced with 🤖{ → \[proposal\](d.side/01-a.md) and → \[answer\](d.side/02-b.md)}: SideLinks returns only 01-a.md, and ReconcileIndex silently drops 02-b.md despite its pending link. This is the 2nd finding in family structural-links-without-markers. ARCH-PURPOSE: state and enforce the rule that every pending link retains its entry; extract every link within each recognized marker. The affected family includes discussion, answer, notes, and proposal labels through this shared parser, consumed by both adapters and collaboration guidance. Add multi-link retention and rejection tests rather than fixing one label or consumer.
```

1. **Strengths**

   - V3 rejects retired mode-level `deletions` and omits it from exported records.
   - Refusal guidance distinguishes invalid edits, stale snapshots, operator locks, and operator-owned fields.
   - Pending-link parsing reuses `pkg/sidedoc`; its new single-link tests demonstrably fail with the base parser.

2. **Critical findings**

   BR-1 and BR-3 remain open for meaningful regression evidence. The new finding requires parsing every link inside a pending marker.

3. **Important findings**

   None beyond the retained dispositions.

4. **Minor findings**

   None.

5. **Test coverage notes**

   `go test ./pkg/reviewexport ./pkg/sidedoc -count=1` and the pinned-range `git diff --check` passed. Temporary mutation checks exposed the two phrase-test gaps; a direct Go reproduction confirmed the multi-link defect. Temporary artifacts were removed and the checkout remains unchanged. External live-conformance claims were not independently rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared export and side-document machinery remain consolidated.
   - **ARCH-PURE — pass:** changed parsing logic remains directly testable over text; the plan’s entity locations match the implementation.
   - **ARCH-PURPOSE — flag:** pending-link preservation covers only the first link per marker, and claimed recovery coverage does not exercise recovery decisions.

7. **Plan revision recommendations**

   Add `## Revisions` entries requiring all pending links to survive reconciliation and meaningful scenario tests for untracked routing and operation-correlated recovery. Correct Task 5’s completed regression claim until those scenarios are demonstrated.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e93b6f0c-c23f-4d71-82ca-e7137aea2f37","repository":"github.com/xianxu/ariadne","issue_id":"000437","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/e44f26f54729c2ce","reviewed_head":"0db63032d318ffa5a1da0b8b10f36153f8b15e15","slot":"ariadne:1","requested_at":"2026-10-18T20:27:54.411325Z"},"phase":"finished","pid":62758,"started_at":"2026-10-18T20:27:54.597469Z","finished_at":"2026-10-18T20:30:32.327017Z","session_id":"01a150b3-10e0-7c21-9222-bbb4a8f0d812","outcome":"returned"}] -->
