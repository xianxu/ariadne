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
