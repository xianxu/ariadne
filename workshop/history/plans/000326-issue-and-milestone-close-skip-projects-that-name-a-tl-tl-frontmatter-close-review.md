# Boundary Review — ariadne#326 (whole-issue close)

| field | value |
|-------|-------|
| issue | 326 — Issue and milestone close skip projects that name a TL (tl: frontmatter) |
| repo | ariadne |
| issue file | workshop/issues/000326-issue-and-milestone-close-skip-projects-that-name-a-tl-tl-frontmatter.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3e80818d6213ac3e9f7c55306a9ccce111cf7cd4..d6e4c4e51a3ace586d33c76d319a4b263f40be14 |
| command | sdlc close --issue 326 |
| reviewer | claude |
| timestamp | 2026-10-11T00:09:04-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Single filter point confirmed; checking the error-sentinel path, then running the tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Verdict: FIX-THEN-SHIP.** The design is correct. The `tl:` filter sits at the one fleet-discovery loop in `computeClose` (`cmd/sdlc/close.go:799-811`). Everything downstream reads `projectEdits`, and I grepped to confirm it. That covers the tick, the detail-block upsert and its `--no-project` refusal, the peer-write (`close.go:920/950/978`, `closetracker.go:345`), and the review snapshot (`close.go:1617`). So one `continue` covers close and milestone-close in every downstream use.

The project, frontmatter and close tests pass, including all four `TestCloseCommands_TLOwnedProjectUntouched` subtests (442s). One regression blocks a clean SHIP. `TLOwners` decodes the whole typed `Metadata` struct. If any *unrelated* field has the wrong type, a project that has no `tl:` and used to be ticked is now skipped with only a warning. That breaks the Spec clause "Projects without `tl:` behave exactly as today."

1. **Strengths**
   - **One decision point (ARCH-DRY).** The filter runs before `newPT := pt`, so none of the downstream steps can disagree with it (`close.go:799-811`).
   - **Error sentinel instead of string matching.** `frontmatter.ErrNoFrontmatter` (`pkg/frontmatter/frontmatter.go:24`) passes through unwrapped via `issue.Parse` and `frontmatter.Split`, so the `errors.Is` check in `TLOwners` works.
   - **The integration test covers every combination.** It runs close and milestone-close, each with and without a TL edit during the review. It checks:
     - the `tl:` project stays byte-identical;
     - the skip line prints;
     - SHIP finalizes;
     - a sibling project without `tl:` is still ticked;
     - no `--no-project` waiver is needed, so it proves the detail-block gate doesn't fire for the TL project.
   - **`TestTLOwners` checks owned, not-owned and malformed cases separately**, with a specific error string for each.
   - **Docs are complete:** the vocabulary file (`tl?: string | [...string] | null`), the `project.md` table, both helptexts, and the atlas paragraph.

2. **Critical:** none.

3. **Important**
   - **Unrelated bad frontmatter now skips projects that have no `tl:`** (`cmd/sdlc/internal/project/metadata.go:78-89`). `TLOwners` calls `d.Metadata()`, which unmarshals into a struct with typed fields: `Deps []string`, `MVPScope []string`, `Name string`, and others.
     - **Example:** a project with `deps: ariadne#5` (a single string) or `mvp_scope: ariadne#1` makes yaml v3 return a type error. `TLOwners` then errors, and the close prints "cannot read tl: … skipping project update" and leaves untouched a project it ticked before this change.
     - **Fix:** decode only the `tl` key into a minimal struct, e.g. `struct{ TL any \`yaml:"tl"\` }`. Then only a malformed `tl:` (or YAML that won't parse at all) fails closed.
     - **Test:** add a unit row `"---\ndeps: ariadne#5\n---\n"` that is expected to be not-owned with no error.
     - **Family (ARCH-PURPOSE):** this is the only site in the window. No other caller uses `TLOwners`.

4. **Minor**
   - Files with CRLF line endings don't match the frontmatter regex, so a CRLF `tl:` project counts as not owned and gets ticked. This limitation was already there and affects every frontmatter reader.
   - The `Metadata.TL` field has only one consumer, `TLOwners`. It's fine as is, but it becomes unnecessary if the Important fix uses a minimal struct.

5. **Test coverage**
   - Each `## Done when` clause is tested in both modes (close and milestone-close) and in both states (with and without a concurrent edit).
   - The Log says a mutation run was done: with the filter disabled, the test fails.
   - Missing: a test that a non-TL project with an unrelated wrong-type field is still ticked. That is the regression above.

6. **Architecture**
   - **ARCH-DRY: pass.** There is one filter site, and `TLOwners` reuses `ParseDoc`.
   - **ARCH-PURE: pass.** `TLOwners` is a pure function of the text and is unit-tested without IO. The close loop is a thin caller.
   - **ARCH-PURPOSE: flag (the Important finding).** The purpose includes leaving projects without `tl:` unchanged. Beyond that, the purpose is met: every downstream consumer derives from the single filter, and the read paths (`projectfind.go:67`) are untouched.
   - Coordination with #308 is noted correctly. This issue decides *which* projects a close touches; #308 keeps *how* the rest are written.

7. **Plan revisions:** none needed beyond a Log line for the fix. The plan's "Unreadable `tl:` → warn and skip" should stay limited to the `tl:` key, not the whole metadata.

```findings
findings:
  - id: new
    severity: Important
    family: fail-closed-scope-too-broad
    title: |
      TLOwners decodes full typed Metadata, so an unrelated malformed field skips a project without tl:
    detail: |
      metadata.go:78 calls d.Metadata(); yaml type errors on Deps/MVPScope/Name (e.g. deps as a scalar string)
      make TLOwners error, and close.go:803 then skips a project with no tl: that it ticked before, which breaks
      the Spec clause "Projects without tl: behave exactly as today". Decode only the tl key (a minimal
      struct{TL any}) and add a not-owned unit row with deps as a scalar. This is the only instance in the window.
```

---

## Re-review — 2026-10-11T00:14:40-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 326 — Issue and milestone close skip projects that name a TL (tl: frontmatter) |
| repo | ariadne |
| issue file | workshop/issues/000326-issue-and-milestone-close-skip-projects-that-name-a-tl-tl-frontmatter.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3e80818d6213ac3e9f7c55306a9ccce111cf7cd4..f221a2aeb46970dd8c97b0295d10bcce39f20a8a |
| command | sdlc close --issue 326 |
| reviewer | claude |
| timestamp | 2026-10-11T00:14:40-07:00 |
| verdict | SHIP |

## Review

I've checked the fix and the tests pass, so I'm writing the verdict.

```verdict
verdict: SHIP
confidence: high
```

BR-1 is fixed. `TLOwners` (`cmd/sdlc/internal/project/metadata.go:79-118`) now decodes only the `tl` key into `struct{ TL any }`. A wrong type in another field, such as `deps: ariadne#5` or `mvp_scope: {a: b}`, no longer makes a project without `tl:` look unreadable. The new not-owned row `unrelated field malformed` covers this. It would fail under the old decode: `Metadata.Deps` is a `[]string` (`metadata.go:23`), and yaml v3 refuses to unmarshal a plain string into a slice. `TestTLOwners` and the frontmatter tests pass. The rest of the design is unchanged and correct: one filter at the close discovery loop (`close.go:799-811`), whose output feeds the tick, the detail-block gate, the peer-write and the review snapshot. Two small documentation problems remain, and neither blocks shipping.

1. **Strengths**
   - **Narrow decode.** The fail-closed check now only fails on the input it actually reads (`metadata.go:87-91`). Only a malformed `tl:`, or frontmatter that won't parse as YAML at all, skips a project.
   - **`ErrNoFrontmatter` sentinel** (`pkg/frontmatter/frontmatter.go:22-24`). `TLOwners` recognises "no frontmatter" with `errors.Is` instead of matching the error message.
   - **Unused field removed.** `Metadata.TL` is gone, so no field is left without a reader.
   - **Integration test.** It runs close and milestone-close, each with and without a TL edit during the review. It also checks that a sibling project without `tl:` is still ticked, and it passes no `--no-project` waiver.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **`workshop/lessons.md:44-47`: the new lesson splits an existing one.** It was inserted between the first line of "A guard must fail closed…" and that lesson's second line. "A catalogued flag may waive only the refusal named by that catalog entry." now reads as the end of the #326 lesson. Move the new bullet below that line. This is the only misplaced edit in the window.
   - **The issue's `## Plan` still says "`Metadata` gains `TL any`".** The BR-1 fix removed that field, as the Log records. Make the plan row match the code.

5. **Test coverage**
   - Every `## Done when` clause is tested in both modes (close and milestone-close) and both states (with and without a concurrent edit).
   - The BR-1 regression row has a real effect: it would fail against the old decode.

6. **Architecture**
   - **ARCH-DRY: pass.** There is one filter site, and `TLOwners` reuses `ParseDoc`.
   - **ARCH-PURE: pass.** `TLOwners` is a pure function of the file text and is tested without IO.
   - **ARCH-PURPOSE: pass.** Every downstream step works from the one filtered list, read paths are untouched, and projects without `tl:` are handled as before.

7. **Plan revisions:** add a `## Revisions` entry saying the `Metadata.TL any` step was replaced by a decode of only the `tl` key (BR-1).

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      TLOwners decodes only struct{TL any} (metadata.go:87-91); the not-owned row "unrelated field malformed" (deps scalar) would fail under the old Metadata decode because Deps is []string.
findings:
  - id: new
    severity: Minor
    family: edit-splits-neighbour
    title: |
      New lessons.md bullet was inserted inside an existing bullet, splitting it
    detail: |
      workshop/lessons.md:44-47. The line "A catalogued flag may waive only the refusal named by that catalog entry." now continues the 326 bullet instead of the fail-closed bullet above it. Move the new bullet below that line. This is the only instance in the window.
  - id: new
    severity: Minor
    family: plan-row-drifts-from-code
    title: |
      Plan row still says Metadata gains TL any, but the BR-1 fix removed that field
    detail: |
      The issue's Plan, first ticked row. Reword it to match the decode of only the tl key, or add a Revisions entry.
```
