# Boundary Review — ariadne#145 (whole-issue close)

| field | value |
|-------|-------|
| issue | 145 — sdlc issue new: derive on-disk template from the issue.cue model |
| repo | ariadne |
| issue file | workshop/issues/000145-sdlc-issue-new-derive-on-disk-template-from-the-issue-cue-model.md |
| boundary | whole-issue close |
| milestone | — |
| window | f0e5a6deb8e1daf9e8c8aab6f64d451451a66952..HEAD |
| command | sdlc close --issue 145 |
| reviewer | claude |
| timestamp | 2026-07-05T16:01:31-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have everything I need. Let me record one memory-worthy nothing and finalize the review.

## Verdict

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The diff cleanly delivers the issue's purpose: `issue.Render` now derives its section list, order, seed placeholders, and initial `status:` from the `scaffold.sections` block in `issue.cue` (via `pkg/vocab`), not from Go literals. I verified this is real, not just claimed — the full suite passes, the embedded `pkg/vocab/issue.json` is byte-identical to a fresh `cmd/vocabulary export` (so the derivation seam is genuinely wired), and the byte-stable goldens + model-driven test prove the refactor preserves output while actually looping the model. The invariant chain (`structural.go` gated ⊆ `scaffold.sections` ⊆ `helptext/issue.md` documented) is enforced by three real tests, all PURE with no mocks. The one thing worth fixing before crossing is a documentation leftover: `helptext/issue.md` still claims to be "the single source of truth for the template" — the exact stale-claim symptom the issue's Problem statement set out to eliminate. Non-blocking, one-line fix.

### 1. Strengths
- **Derivation is genuinely wired, not just documented.** `pkg/vocab/issue.json` == fresh `cmd/vocabulary export --noun issue` (verified), so the `scaffold` block reaches runtime through the real embed seam. `Render` (`scaffold.go:117-136`) loops `m.Sections()`.
- **Byte-stability guarded both paths.** `TestRender_ByteStable` + `TestRender_ByteStable_FromGitHub` (`scaffold_test.go`) pin the exact pre-refactor output for blank and `--from-github`; both pass. The `TestRender_DrivenByModel` test proves order-derivation (not coincidence).
- **The shadow-sweep is real (ARCH-PURPOSE).** The plan enumerates every section consumer and the two non-derivers (`structural.go` subset, `helptext` superset) are held to the model by `TestGatedSectionsSubsetOfModel` and `TestIssueHelpDocumentsEveryScaffoldSection` — containment enforced, not just prose.
- **Name-coupling made safe.** The `Problem`/`Log` dynamic special-cases are keyed by name and pinned by `TestScaffold_SpecialSectionsPresent` (`scaffold_test.go`), with a matching comment on both cue and Go sides.
- **Atlas thoroughly updated** (`atlas/workflow/vocabulary.md:54-72`) — new accessors, the derivation, and the invariant chain all mapped.
- **`gatedSections` consts** (`structural.go:14-25`) correctly single-source the gated names *within* the file without over-modeling the bespoke gate logic into cue — the right DRY line.

### 2. Critical findings
None.

### 3. Important findings
None.

### 4. Minor findings
- **`cmd/sdlc/helptext/issue.md:16` — stale "single source of truth" claim.** The line reads: `` `sdlc issue new` writes — and this is the single source of truth for — the template below.`` After #145 the cue model is the single source of truth and this help doc is a drift-tested *superset consumer* (per the invariant chain and the atlas). The issue's own Problem statement (`workshop/issues/000145…md:16-21`) called out this exact "single source of truth" phrasing as the symptom; the `scaffold.go` doc comment was corrected but this one was missed. Fix sketch: reword to something like "`sdlc issue new` writes the template below; its section shape is derived from `construct/vocabulary/issue.cue` (`scaffold.sections`) — this doc is the human reference, not the source." (Behavior unaffected; the `TestIssueHelpDocumentsEveryScaffoldSection` test still passes since every modeled section stays documented.)

### 5. Test coverage notes
- Tests pin real logic, not mocks: goldens compare literal `Render()` bytes; `TestRender_DrivenByModel` would fail if `Render` hardcoded a divergent set. All PURE (no IO), consistent with ARCH-PURE.
- **Coverage gap worth noting (not blocking):** `TestRender_ByteStable` fixes the *input*, so it catches a byte change to the current model but would **not** catch a regression from *adding a trailing seedless section* to the model. `Render`'s last-section single-newline logic (`scaffold.go:131-135`) keys the `\n\n`-vs-`\n` separator on list index; if a future section is appended after `Log` with no seed and no special-case, that new last section `continue`s before the newline decision and the file would end with a header's `\n\n` (a trailing blank line) instead of a single newline. Harmless today (Log is always last and always has content), but the byte-stable golden won't guard it. A cheap follow-up would assert `strings.HasSuffix(out, "\n") && !strings.HasSuffix(out, "\n\n")` in `TestRender_DrivenByModel`.

### 6. Architectural notes for upcoming work
- **ARCH-DRY — pass.** Section vocabulary now lives once in cue; `status: open` derived via `InitialStatus()`; gated names consolidated to consts. No duplicated logic in the diff. (The byte-golden and `TestSections` restate the section list, but that's legitimate test-expectation data, not a source-of-truth duplication.)
- **ARCH-PURE — pass.** `Sections()`, `InitialStatus()`, `Render` are pure over the once-embedded model; every new test runs without IO/mocks. The build-time `//go:embed` is the correct thin seam.
- **ARCH-PURPOSE — pass (shadow-sweep completed).** The one true creation consumer (`Render`) derives; `structural.go` and `helptext` are correctly identified as non-derivers held by invariant tests; `parley.nvim#116` derives transitively by delegating to `sdlc issue new`. The only residual hand-maintained restatement is the help doc's intentional superset — but its *self-description* (Minor finding above) still contradicts the delivered model and should be corrected so no reader is told the doc is the source.

### 7. Plan revision recommendations
None — the plan already carries the accurate Task 7 revision (generated face is gitignored), and the code matches the Core-concepts table (every listed entity exists at its stated path: `scaffold.sections` in `issue.cue`, `Section`/`Scaffold`/`Sections()`/`InitialStatus()` in `vocab.go`, `gatedSections` in `structural.go`, both drift tests present). No contradiction between table and code.
