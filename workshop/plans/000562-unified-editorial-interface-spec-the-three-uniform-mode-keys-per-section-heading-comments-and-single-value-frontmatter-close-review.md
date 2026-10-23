# Boundary Review — ariadne#562 (whole-issue close)

| field | value |
|-------|-------|
| issue | 562 — unified-editorial-interface: spec the three uniform mode keys, per-section heading comments and single-value frontmatter |
| repo | ariadne |
| issue file | workshop/issues/000562-unified-editorial-interface-spec-the-three-uniform-mode-keys-per-section-heading-comments-and-single-value-frontmatter.md |
| boundary | whole-issue close |
| milestone | — |
| window | ada6c156d263d1ad3376e44147dfdacfe2c4d90e..fef3c6b67b0878d6d61e5d6eff545bc2cf2d94fa |
| command | sdlc close --issue 562 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T23:17:40-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The three-key contract, compatibility rules, worked examples, and export updates are largely delivered, and focused tests pass. Shipping is blocked by contradictory instructions in active consumer prompts and incomplete preservation coverage in the new fill fixtures.

Done-when: current — “all of the above” covers the revised compatibility rules, and preserving existing values covers legacy materialization; the acceptance criteria remain applicable.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: shared-contract-consumer-consistency
    title: |
      Active consumer instructions still prescribe the retired creation contract
    detail: |
      ARCH-DRY / ARCH-PURPOSE: collaboration-modes.md:241 requires every new side doc to use YYYY-MM/YYYY-MM-DD-NN-slug.md, but its routing rule at :146 and state-table rows at :401 and :410 still explicitly create flat paths. The remaining active prompt instances are construct/local/fix/SKILL.md:87 (flat path and same-flat-folder instruction), construct/local/fix/review-convention.md:55 (same flat folder), and cmd/tl-lead/help.md:436-442 (flat answer paths and incomplete written defaults); the TL session-creation instruction at :340 also still emits the retired collaboration-mode map. These are executable instructions, not the explicitly grandfathered examples. The stale grammar is exported alongside the new contract. Sweep these consumers into agreement, preferably by referencing the shared contract, regenerate the export, and exercise creation through the complete prompt composition. Coordinate the TL naming instructions with cmd/tl-lead/answers.go:24, which currently rejects month-directory targets.
  - id: new
    severity: Important
    kind: defect
    family: preservation-conformance-all-fields
    title: |
      Fill conformance examples do not exercise preservation of existing collaboration choices or instructions
    detail: |
      pkg/reviewexport/editorial_fill_test.go:17-19 covers empty, partially canonical, and legacy-only frontmatter, but :36 hardcodes collaboration-mode: discussion for every result. Across all three fixture families, no input has an existing collaboration-mode, and no input has an existing canonical processing-instruction. Consequently the preservation loop at :43 never tests those two fields, although the contract says existing values remain untouched. Add cases preserving a nondefault collaboration-mode and existing processing-instruction while missing mode keys are filled, and exercise them in the contract scenarios. These are the two uncovered field instances in the new preservation fixtures.
```

1. **Strengths**

   - Canonical-key precedence, alias preservation, and lossless multi-mode refusal are explicit in `collaboration-modes.md:42–59`.
   - The month-layout checker covers invalid dates, backlinks, cross-layout numbering, and retry identity with mutation cases.
   - README and atlas both describe the new surface and identify downstream runtime ownership.
   - Export hashes and membership pass `TestExportCurrent`.

2. **Critical findings**

   - **Shared contract consistency:** sweep the active prompt instances enumerated above. Existing compatibility fixtures can remain unchanged.

3. **Important findings**

   - **Preservation coverage:** extend the fill oracle and scenarios to cover existing collaboration settings and canonical instructions.

4. **Minor findings**

   - None.

5. **Test coverage notes**

   Passed:

   - `go test ./pkg/reviewexport ./pkg/vocab ./pkg/sidedoc`
   - `go test ./cmd/weave/internal/walk -run 'Skill' -count=1`
   - `go test ./cmd/tl-lead -run '^TestInstructionsLinkOnlySideDocAnswers$' -count=1`

   The fill tests validate committed before/after artifacts; they do not execute a filler. The issue log’s fresh-context results were not treated as independently verified behavioral evidence. No full-suite pass is claimed.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY: flag** — active prompts duplicate creation rules and now disagree.
   - **ARCH-PURE: pass** — no production IO shell was added; checker mutations use an in-memory filesystem.
   - **ARCH-PURPOSE: flag** — the promised contradictory-instruction sweep remains incomplete.
   - **ARCH-EVOLVE: pass within this window** — aliases and legacy layouts remain readable, destructive migration is avoided, and no persisted decoder/writer changed.

   TL runtime adoption needs explicit attention: its current answer validator accepts only flat paths. Pair/parley ownership does not cover that ariadne-local seam.

7. **Plan revision recommendations**

   Append `## Revisions` entries recording:

   - The expanded active-consumer sweep, including the exported grammar, hosted fix prompt, and TL instructions, with explicit coordination of TL runtime adoption.
   - Preservation scenarios for existing `collaboration-mode` and `processing-instruction`, alongside the already-covered presentation and processing choices.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"dd728520-2827-437b-8bdb-1948128a8376","repository":"github.com/xianxu/ariadne","issue_id":"000562","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/35c1e181a342a6ee","reviewed_head":"fef3c6b67b0878d6d61e5d6eff545bc2cf2d94fa","slot":"ariadne:5","requested_at":"2026-10-23T06:14:38.216534Z"},"phase":"finished","pid":54799,"started_at":"2026-10-23T06:14:38.313337Z","finished_at":"2026-10-23T06:17:39.619899Z","session_id":"01a16765-ac1e-7412-86bc-f3eef0948a93","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T23:26:55-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 562 — unified-editorial-interface: spec the three uniform mode keys, per-section heading comments and single-value frontmatter |
| repo | ariadne |
| issue file | workshop/issues/000562-unified-editorial-interface-spec-the-three-uniform-mode-keys-per-section-heading-comments-and-single-value-frontmatter.md |
| boundary | whole-issue close |
| milestone | — |
| window | ada6c156d263d1ad3376e44147dfdacfe2c4d90e..2a40b9e992ed64dddb675af0beff4026b012fa1e |
| command | sdlc close --issue 562 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T23:26:55-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned changes deliver the revised editorial contract and resolve both prior findings. Prompt instructions, exported contract, TL path validation, preservation fixtures, README, and atlas agree. No new blocking findings remain.

Done-when: current — documenting “all of the above” covers the clarified compatibility rules, and preservation tests cover existing choices and instructions.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Inspected every previously enumerated consumer: creation now uses monthly paths and explicit defaults or references the shared contract. Export freshness passes. Temporarily restoring pre-fix prompts and answers.go makes composition, rendered-help, and monthly-link regression tests fail; restored-head tests pass.
  - id: BR-2
    disposition: addressed
    note: |
      editorial_fill_test.go:20-21 and the canonical-instruction/canonical-mode fixtures exercise existing chat/collaboration choices and canonical instructions alongside independent legacy fallback. Mutating the preserved collaboration and instruction values makes both value and preservation assertions fail.
```

1. **Strengths**

   - Canonical precedence, alias retention, and lossless multi-mode refusal are explicit.
   - Monthly answer validation checks calendar dates, directory matching, and parent resolution while retaining dated flat paths.
   - Refusal tests verify that invalid targets publish neither links nor index entries.
   - README and atlas document the surface and downstream implementation ownership.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed:

   - Full `pkg/reviewexport`, `pkg/vocab`, and `pkg/sidedoc` tests.
   - Full `cmd/tl-lead` tests.
   - Weave skill discovery checks.
   - Export freshness and `git diff --check`.

   Revert and mutation checks confirmed the regressions go red. Temporary changes were restored. Fill fixtures validate conformance examples; they do not execute downstream fillers. No repository-wide suite pass is claimed.

6. **Architectural notes**

   - **ARCH-DRY: pass** — active creation consumers reference or agree with the shared contract.
   - **ARCH-PURE: pass** — validator IO remains injected; checker mutations use in-memory filesystems.
   - **ARCH-PURPOSE: pass** — the contract covers all settled axes, compatibility, completion ownership, and naming.
   - **ARCH-EVOLVE: pass** — legacy formats remain readable; the validator does not rewrite documents or discard unknown fields.

7. **Plan revision recommendations:** None; existing revisions cover the bounded runtime correction and preservation cases.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"40943610-ca13-4cba-8d33-d75f252c1fc3","repository":"github.com/xianxu/ariadne","issue_id":"000562","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/07b048c516b13232","reviewed_head":"2a40b9e992ed64dddb675af0beff4026b012fa1e","slot":"ariadne:5","requested_at":"2026-10-23T06:23:35.410507Z"},"phase":"finished","pid":7202,"started_at":"2026-10-23T06:23:35.525106Z","finished_at":"2026-10-23T06:26:54.916955Z","session_id":"01a1676d-de38-72c2-ab25-a4b81e1ced9b","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T00:33:04-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 562 — unified-editorial-interface: spec the three uniform mode keys, per-section heading comments and single-value frontmatter |
| repo | ariadne |
| issue file | workshop/issues/000562-unified-editorial-interface-spec-the-three-uniform-mode-keys-per-section-heading-comments-and-single-value-frontmatter.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6318fc7bc700173acdcea67f3993d326f8947845..904fc9c470c097e4624b540a788de9fa73d682d5 |
| command | sdlc close --issue 562 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T00:33:04-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range delivers the revised editorial contract, compatibility examples, export, and monthly TL answer-path support. Both prior findings are addressed; no new blocking findings remain.

Done-when: current — documenting “all of the above” covers the presentation fallback clarification, while preservation criteria still cover unchanged canonical values and aliases.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Inspected all previously enumerated creation instructions: monthly paths and explicit defaults now appear directly or through the shared contract. Regression assertions distinguish the retired base instructions from head; composition, rendered-help, monthly-link and export tests pass.
  - id: BR-2
    disposition: addressed
    note: |
      editorial_fill_test.go exercises canonical-instruction and canonical-mode fixtures with existing chat/collaboration choices, canonical instructions and independent legacy fallback. Value and source-line assertions enforce preservation; focused tests pass.
```

1. **Strengths**

   - Canonical presence, presentation fallback, and strict heading declarations are clearly distinguished in `construct/local/universal-review/collaboration-modes.md:50`.
   - Monthly answer validation retains flat-path compatibility and checks calendar dates, directory matching, and relative parents (`cmd/tl-lead/answers.go:22`).
   - Invalid-link tests verify that neither the operator section nor its index is written.
   - README and atlas document the new surface and downstream ownership; export freshness passes.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed:

   - `go test ./pkg/reviewexport ./pkg/vocab ./pkg/sidedoc ./cmd/tl-lead`
   - Weave skill discovery tests.
   - `git diff --check` for the pinned range.

   Fill tests validate conformance artifacts, not downstream runtime fillers. No repository-wide suite or downstream implementation pass is claimed.

6. **Architectural notes**

   - **ARCH-DRY: pass** — creation consumers agree with or reference the shared contract; completion sources are named.
   - **ARCH-PURE: pass** — the production validator retains injected reads; checker tests use in-memory filesystems.
   - **ARCH-PURPOSE: pass** — all three axes, precedence, compatibility, completion requirements, and side-doc layout are covered within the stated specification scope.
   - **ARCH-EVOLVE: pass** — legacy forms remain readable, fill preserves existing lines, and answer validation performs no document rewrite.

7. **Plan revision recommendations:** None. Existing revisions account for the presentation clarification and bounded production validator change.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"411d0776-84c8-4341-87ff-b81caf51595e","repository":"github.com/xianxu/ariadne","issue_id":"000562","recipe":"small-diff-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/56a25e8830a63376","reviewed_head":"904fc9c470c097e4624b540a788de9fa73d682d5","slot":"ariadne:5","requested_at":"2026-10-23T07:31:08.571914Z"},"phase":"finished","pid":7064,"started_at":"2026-10-23T07:31:08.662835Z","finished_at":"2026-10-23T07:33:04.256676Z","session_id":"01a167ab-b60b-7ab3-9868-ab929687030d","outcome":"returned"}] -->
