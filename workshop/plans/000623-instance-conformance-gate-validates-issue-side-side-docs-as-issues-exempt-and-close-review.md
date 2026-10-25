# Boundary Review — ariadne#623 (whole-issue close)

| field | value |
|-------|-------|
| issue | 623 — Instance-conformance gate validates issue .side/ side docs as issues; exempt and validate by type |
| repo | ariadne |
| issue file | workshop/issues/000623-instance-conformance-gate-validates-issue-side-side-docs-as-issues-exempt-and.md |
| boundary | whole-issue close |
| milestone | — |
| window | a141d5da4078701e806ca7716ca9591aff9fd0c9..d2e00528b4fcf5a5f1bc14ea77c441a70247cd11 |
| command | sdlc close --issue 623 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-25T12:58:05-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The shared ownership rule and archive coverage are well implemented, and the targeted tests pass. Shipping is blocked by a reproduced validation bypass: malformed typed side docs can silently become exempt attachments. README coverage and evidence for the explicitly required historical repair are also missing.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: typed-input-fail-closed
    title: |
      Malformed typed side documents bypass conformance validation
    detail: |
      cmd/sdlc/validategate.go:220-240 returns false when YAML parsing fails and the fallback cannot recognize the declaration. Closed frontmatter containing `"type": [invalid` or `type : [invalid` passes validateChangedInstances with zero validator calls; unterminated typed frontmatter also passes. A temporary overlay test reproduced all three against pinned production code. ARCH-SECURE/ARCH-PURPOSE: distinguish absent declarations from malformed typed input, preserve visible refusal, and add regression tests covering the declaration syntax class.
  - id: new
    severity: Important
    kind: defect
    family: user-surface-documentation
    title: |
      README update appears missing for side-doc validation and deployment
    detail: |
      construct/vocabulary/side-doc.cue:5 introduces the side-doc validation noun and its required type/parent contract, while atlas/workflow/sdlc-binary.md:393 documents paired vocabulary/sdlc deployment. README.md is unchanged in the pinned range, despite the plan's explicit README task. Add the validation invocation, typed-versus-untyped attachment behavior, and paired deployment guidance to the user entry point.
  - id: new
    severity: Important
    kind: defect
    family: acceptance-deliverable-completion
    title: |
      Required historical side-document repair lacks completion evidence
    detail: |
      The issue's Done when at workshop/issues/000623-instance-conformance-gate-validates-issue-side-side-docs-as-issues-exempt-and.md:42 explicitly requires archiving parley.nvim's leftover file. Its latest evidence at lines 101-106 still records a requested owner-run repair, and the repair checklist remains unchecked. ARCH-PURPOSE: before whole-issue close, retain the owner's repair commit, matching source/destination blob identities, active-path removal, and corrected archived linkage.
```

**1. Strengths**

- `pkg/sidedoc/tree.go:18` adds content-independent ownership while preserving the existing folder-root exclusion contract.
- Archive discovery covers absent plan documents across preparation, recovery, immutable landing, abandon, reopen, and replacement episodes.
- Real-Git regressions verify blob identity, unrelated-tree preservation, and idempotent archive retry.
- Native CUE tests validate required fields and additive metadata without imposing parent-target existence.

**2. Critical findings**

- `validategate.go:220-240`: malformed typed input silently bypasses validation. Replace the boolean ambiguity with classification that distinguishes untyped evidence from malformed declarations.

**3. Important findings**

- README documentation for the new validation surface and paired deployment is missing.
- The required historical repair has no completion proof in the pinned tracker.

**4. Minor findings**

None.

**5. Test coverage notes**

- Focused sidedoc, issue-ID, and vocabulary tests passed.
- The targeted SDLC side-related and ordinary-gate controls passed in **231.316s**.
- Supplementary overlay tests failed for quoted malformed declarations, spaced malformed declarations, and unterminated frontmatter: each produced **zero validator calls and a nil gate error**.
- The overlay was removed; the checkout remains clean. No full suite was run.

**6. Architectural notes**

- **ARCH-DRY — Pass:** ownership is shared across the three consumer families.
- **ARCH-PURE — Pass:** listed core entities exist; path transformations remain separate from IO.
- **ARCH-PURPOSE — Flag:** malformed-input enforcement and required repair evidence remain incomplete.
- **ARCH-MOCK — Pass:** changed integrations retain existing injected seams and portable private-Git fixtures.
- **ARCH-CONSTRAINTS — Pass:** side-member indexing avoids repeated owner scans; counted growth controls exercise it.
- **ARCH-SECURE — Flag:** parse failure is currently treated as exemption.
- **ARCH-ORDER — Pass:** existing archive transaction boundaries remain intact; recovery and retry tests exercise them.
- **ARCH-FUNERAL — Pass:** no new runtime residue family; moved artifacts retain the archive lifecycle.
- **ARCH-EVOLVE — Pass:** unknown metadata is accepted, and archive moves preserve source bytes.

**7. Plan revision recommendations**

Append a `## Revisions` entry enumerating malformed declaration cases and their visible refusal behavior, including quoted/spaced keys and incomplete frontmatter. Retain the historical repair requirement and record its completion evidence before close.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"81d71839-70dc-480a-9018-3c007e0b3b51","repository":"github.com/xianxu/ariadne","issue_id":"000623","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/6b7beb50ca5b00b7","reviewed_head":"d2e00528b4fcf5a5f1bc14ea77c441a70247cd11","slot":"ariadne:1","requested_at":"2026-10-25T19:52:14.527287Z"},"phase":"finished","pid":75197,"started_at":"2026-10-25T19:52:14.674722Z","finished_at":"2026-10-25T19:58:02.729615Z","session_id":"01a1749e-ee0b-75d1-950f-a391228b1d3c","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T13:14:40-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 623 — Instance-conformance gate validates issue .side/ side docs as issues; exempt and validate by type |
| repo | ariadne |
| issue file | workshop/issues/000623-instance-conformance-gate-validates-issue-side-side-docs-as-issues-exempt-and.md |
| boundary | whole-issue close |
| milestone | — |
| window | a141d5da4078701e806ca7716ca9591aff9fd0c9..b3cdd26015edacff6f1803334633fc3d797b80d5 |
| command | sdlc close --issue 623 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-25T13:14:40-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Shared side ownership, orphan-tree archival, and README documentation are validated. Two prior findings remain blocking: malformed typed YAML still bypasses conformance for additional declaration forms, and the required historical repair remains explicitly unlanded.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      cmd/sdlc/validategate_sidedoc.go:88-95,139-140 still exempts malformed explicit-flow and tagged-flow declarations. Temporary overlay regressions confirmed that closed headers containing {? type: [invalid} or !!map {type: [invalid} pass validateChangedInstances with zero validator calls, although valid counterparts are recognized as typed. The ten committed regressions pass and all fail with the previous discriminator restored through an overlay. Complete the declaration-class rule across mapping prefixes and key syntax, with malformed counterparts tested. ARCH-SECURE/ARCH-PURPOSE; existing family typed-input-fail-closed.
  - id: BR-2
    disposition: addressed
    note: |
      README.md:128-137 now documents typed versus untyped attachments, vocabulary validate-instance --type side-doc PATH, required type/parent fields, and paired deployment. These match construct/vocabulary/side-doc.cue:5-10, gate dispatch, and atlas/workflow/sdlc-binary.md:393-401. This is a prose correction; native CUE and gate tests validate the documented behavior subject to the remaining BR-1 defect.
  - id: BR-3
    disposition: not-addressed
    note: |
      The issue's Done when still requires the historical relocation. Its pinned Log at workshop/issues/000623-instance-conformance-gate-validates-issue-side-side-docs-as-issues-exempt-and.md:113 explicitly calls the repair UNLANDED and says landing confirmation remains outstanding. Branch blob proof is useful progress but does not establish completion. Retain the landed commit and verify matching blobs, active-path removal, and archived linkage on the destination main tree before whole-issue close. ARCH-PURPOSE; existing family acceptance-deliverable-completion.
```

1. **Strengths**

   - `pkg/sidedoc/tree.go:18` supplies shared ownership without requiring an owner document; existing `Folder` semantics remain intact.
   - Orphan archive tests verify exact Git blob preservation, unrelated-tree isolation, absence of invented plans, and idempotent retry.
   - The new CUE schema validates required fields while accepting additive metadata.
   - README and atlas explain compatibility and paired deployment.

2. **Critical findings**

   - **BR-1 remains open:** the fallback parser loses explicit-flow key syntax and tagged mapping prefixes. Recognize the complete admitted declaration class independently of malformed values, preserving nested/literal attachment exemptions.

3. **Important findings**

   - **BR-3 remains open:** record and verify the historical repair’s landed result before closing the issue.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Focused gate, ownership, ID, native CUE, filesystem archive, recovery, landing, and abandon/reopen tests passed. Overlay mutation confirmed all ten committed malformed-header regressions fail without the correction. Two additional overlay regressions exposed the remaining BR-1 bypass. No full suite ran; tracked files remain unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** membership consumers share `Owner`; archive paths share inventory projection.
   - **ARCH-PURE — pass:** path classification and inventory logic remain separate from IO.
   - **ARCH-PURPOSE — flag:** BR-1 and BR-3 leave stated acceptance requirements incomplete.
   - **ARCH-MOCK — pass:** existing seams and portable private Git fixtures exercise production archive paths; native CUE checks validate schema behavior.
   - **ARCH-CONSTRAINTS — pass:** indexed ownership and counted growth tests support the declared batch envelope.
   - **ARCH-SECURE — flag:** BR-1 converts malformed typed input into an exemption.
   - **ARCH-ORDER — pass:** existing publication/recovery authority remains; interrupted recovery and retry tests pass.
   - **ARCH-FUNERAL — pass:** no new runtime residue family; moved files retain the archive lifecycle.
   - **ARCH-EVOLVE — pass:** validation performs no rewrite; additive metadata and exact archived bytes are preserved.

7. **Plan revision recommendations**

   Append a `## Revisions` entry refining BR-1’s declaration rule to include explicit-flow keys and tagged mapping prefixes, with valid/malformed paired regressions. Keep historical repair completion pending until landed evidence exists.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"5f4a8b13-1a4b-4c85-834f-1d0e6e2e6ed5","repository":"github.com/xianxu/ariadne","issue_id":"000623","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/abaa590146e4d8ec","reviewed_head":"b3cdd26015edacff6f1803334633fc3d797b80d5","slot":"ariadne:1","requested_at":"2026-10-25T20:11:59.576941Z"},"phase":"finished","pid":16887,"started_at":"2026-10-25T20:11:59.670589Z","finished_at":"2026-10-25T20:14:39.227715Z","session_id":"01a174b1-0269-7ee2-82db-3e7e17e93b32","outcome":"returned"}] -->
