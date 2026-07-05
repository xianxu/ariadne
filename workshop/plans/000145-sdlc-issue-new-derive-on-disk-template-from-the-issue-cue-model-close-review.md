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

---

## Re-review — 2026-07-05T16:16:23-07:00 (unknown)

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
| timestamp | 2026-07-05T16:16:23-07:00 |
| verdict | unknown |

## Review

Not logged in · Please run /login

---

## Re-review — 2026-07-05T16:25:14-07:00 (SHIP)

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
| timestamp | 2026-07-05T16:25:14-07:00 |
| verdict | SHIP |

## Review

All green: `go build ./...` and `go vet` pass, changed-package tests pass, the embed seam is verified (fresh cue export byte-identical to the committed embed), the provenance sweep is clean, and the data sweep confirms the only creation-template consumer (`Render`) derives while the other section-touching code (`close.go`/`setstatus.go`/`plan.go`) manipulates the Log/Plan of *existing* files — a separate concern, correctly out of scope.

I verified the two findings from the prior FIX-THEN-SHIP sidecar review are both resolved at HEAD: (1) `helptext/issue.md:16-19` no longer claims to be the source of truth — it now names the cue model as the source and itself as the human reference; (2) `TestRender_DrivenByModel` now carries the trailing-newline guard that closes the appended-seedless-section coverage gap.

```verdict
verdict: SHIP
confidence: high
```

This diff genuinely delivers the issue's purpose — `issue.Render` now derives its section list, order, seed placeholders, and initial `status:` from `scaffold.sections`/`categories.open` in `issue.cue` via `pkg/vocab`, not from Go literals — and I verified it's real rather than claimed: the embedded `pkg/vocab/issue.json` is byte-identical to a fresh `cmd/vocabulary export` (the derivation seam actually carries `scaffold` to runtime), the byte-stable goldens prove the refactor preserved output, and `TestRender_DrivenByModel` proves the output is model-*driven* not coincidental. The invariant chain (structural gated ⊆ `scaffold.sections` ⊆ helptext documented) is enforced by three PURE tests with no mocks. Nothing blocks the boundary; the one Minor and one coverage note from the prior close-review have already been applied.

### 1. Strengths
- **Derivation is genuinely wired, not documented-only.** `pkg/vocab/issue.json == cmd/vocabulary export --noun issue` (verified live), so the `scaffold` block reaches the runtime through the real `//go:embed` seam; `Render` loops `m.Sections()` (`scaffold.go:117-136`).
- **Byte-stability guarded on both paths.** `TestRender_ByteStable` + `TestRender_ByteStable_FromGitHub` (`scaffold_test.go:82-193`) pin exact pre-refactor output for blank and `--from-github`; both green.
- **Coverage gap from the prior review is closed.** `TestRender_DrivenByModel` now asserts `HasSuffix("\n") && !HasSuffix("\n\n")` (`scaffold_test.go:239-244`), catching a future appended seedless-last-section regression the fixed-input goldens couldn't.
- **Provenance self-claims corrected (the #122-class trap).** `helptext/issue.md:16-19` now reads "that model is the single source of truth, and this doc is the human reference"; a full `grep -i "source of truth|canonical"` sweep of every touched surface shows no file still falsely claiming to be the source.
- **ARCH-DRY line drawn correctly.** `gatedSections` consts (`structural.go:14-25`) single-source the gated names *within* the file without over-modeling the bespoke gate logic (word counts/regex/fallback) into cue.
- **Name-coupling made safe.** The `Problem`/`Log` dynamic special-cases are keyed by name and pinned by `TestScaffold_SpecialSectionsPresent`, with matching comments on both cue and Go sides.

### 2. Critical findings
None.

### 3. Important findings
None.

### 4. Minor findings
- `pkg/vocab/vocab.go:104` — `InitialStatus()` returns `categories.open[0]`, which assumes the open category's first member is the creation status. True today (sole member) and tested; note only if a second open-category status is ever added.

### 5. Test coverage notes
- Tests pin real logic, not mocks: goldens compare literal `Render()` bytes; `TestRender_DrivenByModel` would fail if `Render` hardcoded a divergent set/order; `TestSections`/`TestInitialStatus` pin the model decode. All PURE (no IO) — consistent with ARCH-PURE.
- The `helptext` drift test's regex (`docSectionRE`) was validated live — it extracts exactly `{Problem, Spec, Done when, Estimate, Plan, Log, Side quests}` and correctly treats the doc as a superset; a modeled section omitted from the doc fails the test (right direction).
- The three invariant-chain tests (`TestGatedSectionsSubsetOfModel`, `TestIssueHelpDocumentsEveryScaffoldSection`, `TestScaffold_SpecialSectionsPresent`) cover the exact bug classes this refactor could ship (a gate requiring an unwritten section; a doc silently dropping a section; a rename silently breaking dynamic injection).

### 6. Architectural notes for upcoming work
- **ARCH-DRY — pass.** Section vocabulary lives once in cue; `status: open` derived; gated names consolidated to consts. The goldens/`TestSections` restate the section list, but that's legitimate test-expectation data, not a source-of-truth duplication.
- **ARCH-PURE — pass.** `Sections()`, `InitialStatus()`, `Render` are pure over the once-embedded model; every new test runs without IO/mocks; the build-time `//go:embed` is the correct thin seam.
- **ARCH-PURPOSE — pass (shadow-sweep completed both axes).** *Data:* the one true creation consumer (`Render`) derives, proven by the propagation e2e recorded in the Log; `structural.go` (subset) and `helptext` (superset) are held to the model by tests, not prose; `parley.nvim#116` derives transitively by delegating to `sdlc issue new`; `close.go`/`setstatus.go`/`plan.go` touch the Log/Plan of *existing* files (a distinct concern, correctly not derived). *Provenance:* every self-claim now points at the cue model — the exact leak the just-added `lessons.md` rule targets, and it's clean here.

### 7. Plan revision recommendations
None — the plan already carries the accurate Task 7 revision (generated face is gitignored), and the Core-concepts table matches the code: `scaffold.sections` in `issue.cue`, `Section`/`Scaffold`/`Sections()`/`InitialStatus()` in `vocab.go`, `gatedSections` in `structural.go`, both drift tests present at their stated paths. No table/code contradiction.
