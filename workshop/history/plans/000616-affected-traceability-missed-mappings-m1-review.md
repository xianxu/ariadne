# Boundary Review — ariadne#616 (milestone M1)

| field | value |
|-------|-------|
| issue | 616 — Affected planning: traceability input and missed-mapping ledger |
| repo | ariadne |
| issue file | workshop/issues/000616-affected-traceability-missed-mappings.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | ce04fa47f783aa2c8e42e0cebc4f476ff7f9ed26..11267aeb4113d3d4cfa4cc991ac55badc0c62130 |
| command | sdlc milestone-close --issue 616 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T00:01:25-07:00 |
| verdict | SHIP |

## Review

Now the code diff (excluding the golden YAML and plan docs).

Tests and fuzzers are green; now checking the drop-cost script and a few parser edge cases.

```verdict
verdict: SHIP
confidence: high
```

M1 does what it says. A traceability file named in affected.json is read only from main. It is parsed by a strict `yaml.Node` walk that makes every refusal itself, and each entry becomes an ordinary `ownerRule`. So `Plan`, `Owned` and `splitImpact` all pick up traceability through `loadAffectedMapping` with no new planner code. I checked this against the code, not the commit messages:
- `go build ./...` and `go vet` on the touched packages are clean.
- These test suites pass: `affected`, `jsonstrict`, `gitx`, `landingline`, plus the `cmd/sdlc` affected, split and impact tests (203s).
- The three fuzz targets ran clean for 20s, 10s and 10s.
- `scripts/test-landing-drop-cost.py --timeout-seconds 400` passes with the new `-coverpkg`: growth 2988 then 5976, linear bound 5976.
- `BenchmarkPlanDeclared` with the 65-entry traceability case runs at 24 ms/op, against a 100 ms budget.
- A throwaway edge-case test, deleted afterwards, behaved as specified. Explicit and custom tags, an empty entry, a trailing `---`, a quoted `"<<"` key, a malformed glob and an entry missing `code` are each refused or skipped as the spec says.

Nothing blocks SHIP. The findings below are Minor: places where the plan's wording has drifted from the code.

1. **Strengths**
   - `traceability.go:105` `mappingPairs`: one helper owns the duplicate-key, alias/anchor, merge-key and non-string-key checks at every level. That answers PQ-1 for the whole class, not just the instance it named. The refusal table at `traceability_test.go:36` covers each level.
   - `mapping.go:76-92`: `Compile` loops over `cfg.Traceability`, the declared list, not over the files it was handed. So a declared file that main lacks refuses ("is missing"), and an undeclared file is ignored. Both are tested in `TestCompileTraceabilityRefuses`.
   - `snapshot.go:56-85`: `TreeFilesSelected` and `TreeFilesAt` share one `treeFiles` body. The `ls-tree` listing, the `cat-file` batch and the entry and blob limits live in one place, and gitlinks or non-blob entries still refuse.
   - `localcheck_affected_declared_test.go:146-203`: the real-git tests prove that entries only the head declares don't count. They also prove a missing or malformed file on main makes the run fall back to the full suite in ordinary mode, and is refused in boarding mode.
   - The codec move to `jsonstrict` changes no behavior. The diff is only renames, and landingline's tests and the drop-cost guard stay green.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The plan gives `ParseTraceability(path, data)`; the code has `ParseTraceability(data)` (`traceability.go:34`), and `Compile` adds the path to the error message.
   - The Core concepts table leaves out `TraceabilityFile`, `TraceEntry` and `gitx.TreeFilesAt`. Only the Revisions mention `TreeFilesAt`.
   - The plan says the `Owned` property extends to traceability. In fact `TestOwnedProperty` wasn't changed; a fixed four-path check was added inside `TestPlanTraceability` instead (`traceability_test.go:162-177`). It is adequate, but it is not a property test.
   - A file ending in a bare `---` (an empty second document) is refused with "more than one document". That is safe but strict; `traceability.go:43`.
   - The atlas says parley's `make test-spec` reads the same file unchanged. That can't be checked from this repo; it rests on the golden copy alone.

5. **Test coverage:** good for M1.
   - The golden test pins the entry count (64) and the whole first entry.
   - The fuzz target checks that anything accepted has clean paths and unique keys.
   - `TestPlanTraceability` covers the claim contract: a Go target, a group target, a target that claims nothing, a target in `full`, and an entry with no `tests`.
   - The loader is tested in both modes. No real-git test sends a non-Go spec through the loader, but the pure `Plan` test covers that path.

6. **Architecture notes**
   - **ARCH-DRY: pass.** The claim step, the tree reader and the JSON codec each have one copy. `cleanRelPath`/`validGlob` are reused. The `cleanRelPath && !hasGlobMeta` check appears twice, which is fine.
   - **ARCH-PURE: pass.** Parsing and compiling are pure; the loader is a thin IO step over `TreeFilesAt`.
   - **ARCH-PURPOSE: pass.** All three consumers of the mapping (run and resume through `localcheck_affected.go:233`, and `splitImpact` at `landingattribution_impact.go:85`) go through `loadAffectedMapping`.
   - **ARCH-MOCK: pass.** Tests use real git in temp repos, the git fake already established here.
   - **ARCH-CONSTRAINTS: pass.** The benchmark was measured, and going over a snapshot limit makes the selection full.
   - **ARCH-SECURE: pass.** Files are read from main only, aliases are refused (so no expansion attack), paths must be clean, and the 1 MiB per-blob limit applies.
   - **ARCH-ORDER: pass.** Nothing holds state between events, because parse, compile and plan are each a function of one call's inputs.
   - **ARCH-FUNERAL: pass.** M1 creates nothing durable; the mapping and sources live in memory for one run.
   - **ARCH-EVOLVE: pass.** The new affected.json member is additive, and an older binary refuses it and falls back to the full suite, which is safe. The traceability format is strict v1, and the moved codec behaves the same.
   - **For M2:** the `jsonstrict` surface (`DecodeObject`/`EncodeWithExtra`) is now exported and is the right base for the ledger. Keep the ledger's known-member check exact by name, as `object_test.go:28` does.

7. **Plan revisions:** add a "## Revisions" entry saying that:
   - `ParseTraceability` takes only `data`, and the path is added by `Compile`;
   - the Core concepts table gains `TraceabilityFile`, `TraceEntry` and `gitx.TreeFilesAt` (new);
   - the `Owned` check is a fixed-path assertion in `TestPlanTraceability`, not an extension of `TestOwnedProperty`.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: plan-surface-matches-code
    title: |
      Plan's ParseTraceability(path, data) signature and Core concepts table lag the built surface
    detail: |
      traceability.go:34 is ParseTraceability(data) and Compile adds the path to the error; TraceabilityFile, TraceEntry and gitx.TreeFilesAt are new entities that are missing from the Core concepts table. Add a Revisions entry.
  - id: new
    severity: Minor
    kind: defect
    family: plan-strategy-delivered-as-stated
    title: |
      The Owned property was not extended; a fixed four-path check in TestPlanTraceability stands in for it
    detail: |
      The plan says the existing Owned-includes-graph-owners property extends to traceability rules, but TestOwnedProperty is unchanged. traceability_test.go:162-177 checks four fixed paths instead. That is adequate coverage, but the plan wording should match it.
  - id: new
    severity: Minor
    kind: defect
    family: strict-parse-refusal-scope
    title: |
      A traceability file ending in a bare --- (empty second document) is refused as more than one document
    detail: |
      traceability.go:43 refuses any second Decode result that is not EOF, including an empty trailing document. This is safe (the selection goes full), but it may surprise a hand editor. Note it, or tolerate a null-only trailing document.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"722afff2-10ff-419f-9ba6-80d5285f3d13","repository":"github.com/xianxu/ariadne","issue_id":"000616","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/81364da035d612b2","reviewed_head":"11267aeb4113d3d4cfa4cc991ac55badc0c62130","slot":"ariadne:5","requested_at":"2026-10-25T06:53:06.016996Z"},"phase":"finished","pid":50213,"started_at":"2026-10-25T06:53:06.179343Z","finished_at":"2026-10-25T07:01:24.867474Z","session_id":"093e93bd-b7c5-4181-9134-5d1a4d94c2bc","outcome":"returned"}] -->
