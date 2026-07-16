# Boundary Review — ariadne#180 (milestone M4)

| field | value |
|-------|-------|
| issue | 180 — project vocabulary model: schematize project like issue |
| repo | ariadne |
| boundary | milestone M4 |
| window | 0a39068935d18b70cfe5e39ad14fb5571acdd3fc..5bb8b99 |
| reviewer | codex |
| final verdict | FIX-THEN-SHIP |

## Review sequence

The full M4 window received repeated fresh-context reviews because each REWORK
pass found a distinct correctness issue below the previous layer. The durable
record keeps the findings and resolutions without retaining the generated raw
prompts, diffs, and duplicated transcripts.

1. **REWORK — close contract and transaction.** Close selected known guards
   rather than executing the complete modeled list, could calibrate from partial
   actuals, and changed the project before the sibling ledger write. Phase-A was
   misclassified as PURE; README and real-instance evidence were incomplete.
   Remediation added a fail-closed guard registry, complete-only calibration,
   staged/compensated writes, runnable docs, and live read-only dogfood.
2. **REWORK — malformed estimates.** Invalid `estimate_hours` silently became a
   plausible zero. Typed parsing now surfaces an explicit board warning.
3. **REWORK — YAML and Phase-A semantics.** Raw scalar readers broke valid
   quoted/block YAML; absent and malformed Phase-A values collapsed together;
   rollback coverage stopped too early. One typed metadata decoder, tri-state
   Phase-A parser, post-ledger rollback test, and last-retro age resolved these.
4. **REWORK — finite numbers and residual consumers.** NaN/Inf could enter
   calibration, board/stale-retro consumers still read raw values, dates could
   cross midnight, and lifecycle literals remained. Finite-positive validation,
   typed consumers, one captured date, and model-event derivation resolved them.
5. **SHIP.** The preceding remediation passed with no findings. A subsequent
   accidental re-run reopened the whole window and exposed further valid issues;
   those later findings were treated as authoritative rather than ignored.
6. **REWORK — duplicate MVP identity.** Exact and alias-equivalent scope refs
   could double-count calibration. Shared repository-plus-ID identity now rejects
   duplicates before lookup or mutation.
7. **REWORK — unavailable-peer compatibility.** Filesystem-required identity
   resolution defeated the documented incomplete-actuals bypass. Best-effort
   identity preserves `--no-ledger` degraded close while still deduplicating
   resolvable aliases.
8. **REWORK — alias-sensitive board graph.** Raw ref spellings split one logical
   dependency graph into false parallel threads and duplicate tasks inflated
   remaining effort. Lookup now carries canonical identity into the pure board.
9. **REWORK — ledger structure.** EOF without a trailing newline could panic,
   and substring heading search could target prose/fenced examples. A pure,
   fence-aware, section-scoped line transformation now handles both EOF forms.
10. **FIX-THEN-SHIP — shadow atlas.** No code or architecture blocker remained.
    The only Important finding was the stale brain-era project row in
    `atlas/workflow/ledger-landscape.md`; it was corrected in the sanctioned
    boundary commit without re-running the review.

## Final verification

- `go test ./... -count=1`
- focused race suite across `cmd/sdlc`, project/process-manual internals, and
  `pkg/vocab`
- `bash construct/vocabulary/vet_test.sh`
- `git diff --check`
- live `project status --slug project-management-primitive`

ARCH-DRY, ARCH-PURE, and ARCH-PURPOSE passed in the final review. The boundary
commit carries `Review-Verdict: FIX-THEN-SHIP` and the reviewed window trailer.
