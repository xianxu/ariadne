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
