# Boundary Review — ariadne#593 (whole-issue close)

| field | value |
|-------|-------|
| issue | 593 — Measure close windows whose branch-attribution snapshot exceeds 64 KiB instead of refusing |
| repo | ariadne |
| issue file | workshop/issues/000593-measure-close-windows-whose-branch-attribution-snapshot-exceeds-64-kib-instead.md |
| boundary | whole-issue close |
| milestone | — |
| window | 31b03d889ad8a455cfc91ba3451ce2e5c8a8a29d..ca9d0d20966837e3ad342f30490ade836cf3cd05 |
| command | sdlc close --issue 593 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T07:14:36-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the codec now; next I'll run the tests and a revert check on the compaction.

Tests are slow; checking the partial output.

The gatestate package passes. Now I'll test whether the probe misreads key casing, using a scratch test in this throwaway checkout.

The probe concern doesn't hold: `jsonstrict` already refuses keys that collide case-insensitively. Checking the full test run's output.

The `cmd/sdlc` run hit Go's 10-minute default timeout. I'll rerun just the tests this diff touched.

The broad filter still matches too many tests. I'll stop it and run only the two edited tests by exact name.

The three edited tests pass. Next, a revert check: I'll turn off compaction and confirm the new tests fail.

The revert check confirms the claim: with compaction off, three of the new tests go red. Next I'll check whether branch intervals have any count bound, which decides the worst case.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The fix works. Measurement version 4 stores branch evidence in a compact form that decodes back to exactly what was written. The byte bound now depends on the version: 64 KiB through v3 and 128 KiB for v4. Version 3 snapshots replay byte-for-byte, and readers older than #593 refuse version 4. A shape like #586's (75 sources, 150 intervals) is pinned by a test, and so is one at 128 sources. I disabled compaction in this checkout and three tests went red (`TestCompactBranchesRoundTrip`, `TestMeasurementV4Issue586Shape`, `TestMeasurementV4RefusedByOlderReader`), which matches the implementer's own check. One gap is left against the second Done-when item. Branch intervals and windows per source have no count limit, so a snapshot within 128 sources can still pass the new 512 KiB parse-cost cap. That case still fails late inside `MarshalJSON`, and no test covers it. The fix is cheap.

1. **Strengths**
   - `compactBranches` / `expandBranches` (`measurement_branch_v4.go:245-300`) are pure, map one-to-one, and treat a branch that is present but `""` as literal. `expandBranches` refuses a negative index, an out-of-range index, an interval with both `source` and `source_index`, and one with neither (`TestExpandBranchesRefusesBadIndex`).
   - The codec reuses the shared helpers instead of copying `MeasurementInputs`' fields. Only the `branches` member is swapped (`compactWire`, `measurement_codec.go:445`).
   - Unknown members survive a rewrite at all three levels, byte-for-byte (`TestMeasurementV4PreservesUnknownMembers`). That satisfies ARCH-EVOLVE.
   - I checked whether a key differing only in case (`"Version":4` next to `"version":3`) could confuse the new probe in `UnmarshalJSON`. It can't: `jsonstrict` already refuses keys that collide case-insensitively.
   - No other code path checks the measurement version, so moving the writer to v4 at `actualinputs.go:254-256` misses no consumer.

2. **Critical findings:** none.

3. **Important findings**
   - **Very large snapshots still fail inside `MarshalJSON`** (`measurement.go:315-321`). `Validate` calls `json.Marshal(m)`, and the encoder's 512 KiB check fires before the counted refusal is reached. The message now includes the counts, but it is still the "does not serialize … error calling MarshalJSON" failure that the Done-when rules out. The 128-source test uses 2 intervals per source; the real worst case has no fixed size.
     - Fix: have `Validate` turn the structural-cap error into the same counted refusal, or measure size before calling `checkMeasurementJSON`.
     - Test: a fixture with many intervals that pushes past 512 KiB and asserts the error contains no "does not serialize". (ARCH-PURPOSE, ARCH-CONSTRAINTS.)

4. **Minor findings**
   - `helptext/actual.md:66`: the edited paragraph has an over-long line ("…crossing execution span. Existing adopted v1/v2"), so the wrapping breaks.
   - The plan's own `## Plan` checklist is still all `- [ ]`, while the issue shows every item done.
   - The v4 JSON writer re-serializes through `map[string]json.RawMessage`, so top-level keys come out sorted. It is deterministic and round-trip tests cover it; just note that v4 key order differs from v3.

5. **Test coverage notes**
   - `go test ./cmd/sdlc/internal/gatestate/` passes.
   - The three edited `cmd/sdlc` tests pass when run by exact name (23s).
   - The plan's broad `-run 'Actual|Measurement|Branch|Close'` hit Go's 10-minute default timeout twice in this environment. The output was cut off, so I couldn't see which test was still running. I'm not attributing that to this diff, but the implementer should confirm the broad run with `-timeout` raised.
   - `TestMeasurementV4At128Sources` stays green with compaction disabled. At 2 intervals per source, the larger bound alone is enough there, so the test pins the bound, not compaction. The #586-shape test covers compaction.

6. **Architectural notes**
   - ARCH-DRY: pass. One limit function, shared codec helpers.
   - ARCH-PURE: pass. The new functions are pure, and the codec is the thin seam.
   - ARCH-PURPOSE: flag, the Important finding above.
   - ARCH-MOCK: pass, N/A. No external dependency is involved.
   - ARCH-CONSTRAINTS: flag, the same finding. The stated envelope assumes about 2 intervals per source.
   - ARCH-SECURE: pass. Indices are range-checked, a strict scan runs before the probe, and expanded evidence still goes through `validateBranchEvidence`.
   - ARCH-ORDER: pass, N/A. The snapshot is an immutable value written once.
   - ARCH-FUNERAL: pass, with a note. The most a single close can add to a ledger rises from 64 to 128 KiB, but typical snapshots shrink (#586: about 67 KB down to about 55 KB).
   - ARCH-EVOLVE: pass. The version bump is justified: an old reader would see `source_index` as unknown and find no source. Note the consequence: as with v3, an older binary's `ParseSidecar` refuses the whole ledger once a v4 `actual_inputs` lands. Fleet binaries need upgrading before they read closed ledgers.

7. **Plan revision recommendations**
   - Add a "## Revisions" entry saying whether snapshots over 512 KiB, which unbounded intervals per source make possible, get the counted refusal. Tie the ARCH-CONSTRAINTS envelope to intervals per source, not only sources.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: snapshot-bound-refusal-not-total
    title: |
      A snapshot over the 512 KiB structural cap still fails late inside MarshalJSON instead of the counted refusal
    detail: |
      Validate (measurement.go:315) calls json.Marshal first, and encodeMeasurementObject's checkMeasurementJSON refuses at 512 KiB with the generic "does not serialize ... error calling MarshalJSON" text, before the counted check is reached. Branch intervals and windows per source have no count bound, so a window within SnapshotLimitSources can reach this path, which is exactly what Done-when 2 rules out. No test covers it (the 128-source fixture uses 2 intervals per source). Fix: map the structural-cap error to the counted refusal (size or bound, version, counts) and add a many-interval fixture over 512 KiB that asserts the error contains no "does not serialize".
  - id: new
    severity: Minor
    kind: defect
    family: helptext-wrap-width
    title: |
      The edited paragraph in helptext/actual.md has an over-long line and broken wrapping
    detail: |
      Line 66: "new method or version rather than silently drop a crossing execution span. Existing adopted v1/v2" runs past the paragraph's column width.
  - id: new
    severity: Minor
    kind: defect
    family: plan-checklist-drift
    title: |
      The plan file's Plan checklist is still unchecked though the issue marks every item done
    detail: |
      workshop/plans/000593-...-plan.md lists every step as "- [ ]" while the issue's Plan section shows "- [x]".
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"21405f87-d955-4351-af0a-6b110574a6a3","repository":"github.com/xianxu/ariadne","issue_id":"000593","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/a338cd41baef65e5","reviewed_head":"ca9d0d20966837e3ad342f30490ade836cf3cd05","slot":"ariadne:2","requested_at":"2026-10-24T13:50:55.796062Z"},"phase":"finished","pid":54168,"started_at":"2026-10-24T13:50:55.968315Z","finished_at":"2026-10-24T14:14:35.518827Z","session_id":"09c8a668-7ea3-43c1-af26-36bfcf51b435","outcome":"returned"}] -->
