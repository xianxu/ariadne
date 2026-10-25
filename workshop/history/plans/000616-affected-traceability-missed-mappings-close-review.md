# Boundary Review — ariadne#616 (whole-issue close)

| field | value |
|-------|-------|
| issue | 616 — Affected planning: traceability input and missed-mapping ledger |
| repo | ariadne |
| issue file | workshop/issues/000616-affected-traceability-missed-mappings.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8f3eec5e1b92a6d6f3e183267b9680e0385df2de..1202e2c2f70aa952a140c9bc03574e5791b8ce69 |
| command | sdlc close --issue 616 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T03:06:53-07:00 |
| verdict | SHIP |

## Review

Prior-round fixes check out. Now reviewing the M2 CLI, the merge and train hooks, and their tests.

Tests pass. Now checking the hook tests, and whether those tests fail when the hooks are removed.

```verdict
verdict: SHIP
confidence: high
```

All three prior-round Minors are fixed, and M2 delivers what the Spec asks for. The pieces are the pure `Missed` and `MissedLedger`, a locked ledger file with crash-safe writes, the merge-diagnosis hook, the train-split hook and `sdlc affected missed list|dismiss`. Tests, README, atlas and recovery-catalog entries come with them. I built the code at the pinned head and ran `go test` on `internal/affected`, `jsonstrict`, `gitx`, `landingline` and `recovery`, plus `./cmd/sdlc -run 'Missed|Affected|SplitImpact|Attribution|SuiteFailure|Diagnose'`. Everything passed (the `cmd/sdlc` run took 286s). Nothing blocks SHIP. Three new Minors are below. The most notable is that the ledger's `hits` counter also counts multiple failing tests in the same package and change, not just repeat sightings.

### 1. Strengths
- **One shared planner.** `affectedSelectionAt`, `changedPathsBetween` and `retiringCleanupGit` (`cmd/sdlc/landingattribution_impact.go:117-175`) were pulled out of `splitImpact`. The merge recorder, `list`'s coverage check and the split attribution all plan through the same code, so there is no second planner (ARCH-DRY).
- **Recording can't change a gate's outcome, and tests prove it.** On the merge side, `recordMissedMappings` (`suitebaseline.go:78-87`) turns any recorder error into a label, and `TestDiagnoseSuiteFailureRecordsMissedMappings` checks `errors.Is(err, gateErr)` in every case. On the train side, the split verdict is computed before `recordSplitMisses` runs (`landingtrain_attribution.go:116-117`).
- **The train hook adds no new rule.** It reuses `ImpactUntouched`, which `splitImpactPaths` only assigns when neither end of the split is uncertain (`landingattribution_impact.go:93-97`). It only runs when the earlier prefix passed and the later one failed (`assessSplit`, line 67). That matches `Missed`'s rule of "narrowed, group-free selections only".
- **The ledger format can evolve safely (ARCH-EVOLVE).** It decodes strictly through `jsonstrict` and keeps unknown members at both the top level and inside entries. A nested `triage` object survives a round trip, and duplicate keys, wrong types and missing `paths` are refused (`missed_test.go:137-170`). The fuzzer once found an entry without `paths`; that crash is kept as a seed. Refusing other versions is justified in the code: an entry's members decide dedupe and pruning, so a newer file must not be rewritten.
- **Writes are locked and crash-safe.** Every change is read-modify-write under a flock, written to a temp file and renamed into place. Stale temp files are swept under the lock. A `list` that prunes nothing doesn't rewrite the file (`errNoLedgerChange`). There is a test for concurrent writers.

### 2. Critical findings
None.

### 3. Important findings
None.

### 4. Minor findings
- **`hits` counts tests, not sightings** (`internal/affected/missed.go:108-125`). `Missed` dedupes by (Test, Pkg), but the ledger keys entries by (Pkg, path set). So two failing tests in one package from one merge produce two `Record` iterations against the same ID. The result is `hits` = 2, `recorded` = 2, and `Test` keeps only the last name. Train subtest leaves (`pkg.TestA/x`, `pkg.TestA/y`) inflate it the same way. `TestMissedLedgerSequence` only ever records one miss per call, so it doesn't catch this. Fix sketch: group misses by ID inside `Record` (one hit per observation, keeping all test names or a count), and add a test that records two tests from one package in one call.
- **`list --json` hides the "nothing pruned" reason** (`affected_missed_cmd.go:157-168`). If `covErr` is set in JSON mode, the "nothing pruned" reason is never printed. Write it to stderr instead.
- **`list` checks only the first 20 stored paths** (`affected_missed_cmd.go:126`). If the path that actually needs the declaration was cut off, the entry never drops out of `list` on its own and has to be dismissed by hand. The atlas could say so.

### 5. Test coverage notes
- **Merge hook:** covered with a stubbed recorder, plus real-git tests of the production recorder (`TestMissedMergeRecorderRealGit`) and of main's coverage check (`TestMainMissedCoverageRealGit`).
- **Train hook:** covered end to end in the split-attribution test, both registered and unregistered, plus `TestRecordSplitMisses` for filtering and store failure.
- **Store:** concurrency, refusal of a bad ledger, temp sweeping and honouring the context on the lock wait are each tested.
- **Gap:** no test records several misses that share an ID in one `Record` call (see the first Minor).

### 6. Architectural notes
- **ARCH-DRY: pass.** See Strengths. The train hook's default `record` closure is trivial wiring.
- **ARCH-PURE: pass.** `Missed`, `MissedLedger.Record/Dismiss/Prune` and the codec are pure and tested without IO. `missedLedgerStore` is the thin shell, and the train hook's sink is an injected seam (`conductor.missed`).
- **ARCH-PURPOSE: pass.** Both producers (merge and train) are built, and both `list` and `dismiss` exist. `list` prunes against main's current mapping, so adding a declaration on main closes the loop. Nothing that is the point of the issue was deferred.
- **ARCH-MOCK: pass.** git is exercised against real temporary repositories, as elsewhere in the repo. No new external service.
- **ARCH-CONSTRAINTS: pass.** The merge recorder has its own 2-minute budget. The train recorder runs inside the split's existing 2-minute budget. `list` builds one main graph per call, and an operator runs it on demand.
- **ARCH-SECURE: pass.** The ledger file comes from a shared directory and is treated as untrusted. It is decoded strictly, and a malformed one is refused with its path named, never overwritten. No credentials are touched.
- **ARCH-ORDER: pass.** All state changes are serialised by the flock. A read-only `list` aborts without writing. The coverage predicate is computed before the lock is taken, and a stale view can only under-prune, which is safe.
- **ARCH-FUNERAL: pass.** The ledger is capped at 200 entries (oldest last-seen drops first), crashed writers' temp files older than an hour are swept, the lock file stays a constant size, and private checkouts are removed on a separate 5-second budget.
- **ARCH-EVOLVE: pass.** See Strengths.

### 7. Plan revision recommendations
None required. The Revisions entries "M2 as built" and "M1 as built" match the code. Optionally, also correct the line in the Core concepts prose that still reads `ParseTraceability(path, data)`; the Revisions entry already records the change.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Plan Revisions (M1 as built) records ParseTraceability(data) and Compile adding the path; Core concepts table now lists TraceabilityFile/TraceEntry and TreeFilesAt.
  - id: BR-2
    disposition: addressed
    note: |
      TestOwnedProperty (declared_test.go) now compiles a traceability source and checks the ui/* paths; the fixed four-path loop was removed and Revisions records it.
  - id: BR-3
    disposition: addressed
    note: |
      traceability.go tolerates empty trailing documents via emptyDocument; TestParseTraceabilityRefuses accepts ok+"---" and ok+"---\n---" and still refuses "---\n~" and a content document.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: dedupe-key-granularity
    title: |
      MissedLedger.Record counts each failing test in a package as a hit, not each sighting
    detail: |
      Missed dedupes by (Test, Pkg) but entries key by (Pkg, path set), so N failing tests (or train subtest leaves) in one package from one observation give hits=N, recorded=N, and keep only the last Test. Group misses by ID inside Record and add a two-tests-one-package case to TestMissedLedgerSequence.
  - id: new
    severity: Minor
    kind: defect
    family: diagnostic-visible-in-every-output-mode
    title: |
      affected missed list --json silently omits why nothing was pruned
    detail: |
      runMissedList prints the coverage error only in text mode (affected_missed_cmd.go:157-168); in JSON mode it is dropped. Write it to stderr.
  - id: new
    severity: Minor
    kind: defect
    family: stored-evidence-sufficient-for-reader
    title: |
      list coverage plans only the first 20 stored paths, so a truncated culprit path never auto-prunes
    detail: |
      mainMissedCoverage plans e.Paths (capped at missedPathCap) not the full change; when the declaration-relevant path was truncated the entry stays until dismissed. Document it in the atlas, or note it in list output when PathsTotal exceeds the stored paths.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"dcb222ab-5f5b-44a3-8ebf-b4e2043aa0f9","repository":"github.com/xianxu/ariadne","issue_id":"000616","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/9bb86c558d094a3c","reviewed_head":"1202e2c2f70aa952a140c9bc03574e5791b8ce69","slot":"ariadne:5","requested_at":"2026-10-25T09:58:30.612755Z"},"phase":"finished","pid":90912,"started_at":"2026-10-25T09:58:30.705471Z","finished_at":"2026-10-25T10:06:18.238233Z","session_id":"811a77e0-b3f1-42aa-8bb0-ad6102adfc9e","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T04:09:58-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 616 — Affected planning: traceability input and missed-mapping ledger |
| repo | ariadne |
| issue file | workshop/issues/000616-affected-traceability-missed-mappings.md |
| boundary | whole-issue close |
| milestone | — |
| window | e5756748b4e982b3933b90e1c5a0b917bb5ef536..bfc0039e11803740b669bc253d0a1b086828905b |
| command | sdlc close --issue 616 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T04:09:58-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

All three open Minors (BR-4, BR-5, BR-6) are fixed in this window, and the window adds no new Critical or Important defect. BR-4 is fixed for the whole class: `Record` now dedupes by entry ID within one observation. Any number of misses that resolve to the same (package, path set) entry count as one hit, so this covers several failing tests in one package as well as a train's subtest leaves. I confirmed the regression test catches the bug: I removed the `seen[id]` guard and `TestMissedLedgerOneSightingPerObservation` failed; with the guard restored it passes. BR-5 sends every note to stderr under `--json`, both the coverage error and the "covered now, removed" lines, and the test now checks stderr in JSON mode. BR-6 is documented in the atlas and marked in the `list` output. The issue-file diff only changes status, timestamps and log lines; `## Spec`, `## Revisions` and `## Done when` are unchanged, so scope did not change. `go test ./internal/affected/ . -run Missed` passes at head.

1. **Strengths**
   - `cmd/sdlc/internal/affected/missed.go:112-118`: the dedupe sits at the entry-ID level inside the pure `Record`, so it fixes the whole family and needs no IO in its test.
   - `cmd/sdlc/affected_missed_cmd.go:156-166`: one `note` writer chosen per output mode replaces the old text-only branch. There is now one place that emits notes, so a later note cannot be added to only one mode.
   - `atlas/workflow/ci-merge-check.md:431-433` states the truncation limit in the same section that defines `affected.Missed`.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - `missed.go:102`: the edited doc comment runs past the file's usual wrap width. Style only; not raised.

5. **Test coverage notes**
   - BR-4: the new test fails without the fix and covers one package with two tests plus a second package.
   - BR-5: the JSON-mode stderr check covers the pruned-note path. The coverage-error note in JSON mode goes through the same `note` writer but has no dedicated assertion. That is acceptable because the routing is shared.
   - BR-6: the change is to wording and output text only. The atlas passage matches the code: `kept` is capped at `missedPathCap` and the ID is keyed by the full `paths`.

6. **Architectural notes (interdiff only)**
   - **ARCH-DRY: pass.** The duplicated per-mode note printing was merged into a single path.
   - **ARCH-PURE: pass.** The dedupe logic is in the pure ledger, and the command layer only picks a writer.
   - **ARCH-PURPOSE: pass.** BR-4 was fixed for every way misses can share an ID, not only the case the finding named. BR-5 covers both note kinds.
   - **ARCH-EVOLVE: pass.** The persisted format did not change, and `Extra` (unknown members) is still carried over on update.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      Record dedupes by entry ID per observation (missed.go:112-118); TestMissedLedgerOneSightingPerObservation fails with the guard removed (verified by revert) and passes at head.
  - id: BR-5
    disposition: addressed
    note: |
      runMissedList routes both the coverage-error and pruned notes to stderr under --json (affected_missed_cmd.go:156-166); the test asserts the stderr note in JSON mode, which the prior code would fail.
  - id: BR-6
    disposition: addressed
    note: |
      Atlas ci-merge-check.md:431-433 documents that pruning plans only the stored 20 paths; list output now marks truncated entries. Matches missed.go kept/missedPathCap and MissID keyed by full paths.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"079c6056-a1ef-41d4-95fb-f2fe8a19be5f","repository":"github.com/xianxu/ariadne","issue_id":"000616","recipe":"fix-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/0680b5b85f07c9ff","reviewed_head":"bfc0039e11803740b669bc253d0a1b086828905b","slot":"ariadne:5","requested_at":"2026-10-25T11:08:52.688912Z"},"phase":"finished","pid":91636,"started_at":"2026-10-25T11:08:52.784617Z","finished_at":"2026-10-25T11:09:56.968513Z","session_id":"93c160db-d340-45fe-8cb2-6f98abe7315c","outcome":"returned"}] -->
