# Re-close Actuals Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3. Execute with the executing-plans skill; SDLC owns the fresh-context boundary review. Steps use checkbox syntax.

**Goal:** A post-close fix can be re-closed with measured actuals tied to its newly reviewed head, and ambiguous legacy calibration has a supported, auditable repair path.

**Architecture:** Separate candidate-close measurement from completed-generation replay. Reuse the attribution engine and snapshot schema. Add explicit selection of one eligible raw calibration row for otherwise ambiguous correction; existing correction receipts retain that exact target and completion binding.

**Tech Stack:** Go, Cobra, real scratch Git/tracker repositories, portable transcript fixtures, existing calibration correction state machine.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| Measurement purpose | cmd/sdlc/actualinputs.go | new |
| Exact calibration target selection | cmd/sdlc/actualcorrection.go | modified |

- Measurement purpose is typed caller intent: historical replay or close candidate. Candidate inputs are newly derived and retained with the new close; no retained pin is rewritten. No schema change is needed.
- Target selection requires an effective row matching issue, recorded actual and close date, or a corrected row's exact completion token. An optional full RawID selects within that eligible set. Default ambiguity never chooses by ordering. Caller selection is explicit evidence supplied with a bounded reason, not a claim that old rows carried generation metadata.

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| Owned measurement resolver | cmd/sdlc/actualinputs.go | modified | pinned Git/tracker history and transcript discovery |
| Close measurement caller | cmd/sdlc/close.go, actual.go | modified | candidate resolver plus existing active-time engine |
| Correction command | cmd/sdlc/actualcorrection.go | modified | target parser and existing receipt/card/main/calibration transaction |

Reuse real-Git tracker fixtures, portable transcripts and the existing judge dispatch seam. No live judge or user calibration is modified. No new external binary or fake is introduced.

## Alternatives and chosen behavior

1. **Chosen: re-measure at close.** Candidate measurement bypasses prior completion replay, including prior correction snapshots, and derives current ownership-proven inputs at candidate HEAD. Historical preview remains strict. Output names fresh provenance and code/tracker pins; the new review round retains those inputs.
2. Carrying the first actual forward would need an explicit measurement-generation relationship. Merely changing saved.CodeHead fabricates provenance; relaxing equality loses the invariant. Do neither.
3. Legacy correction cannot infer the correct ordinary row from date/rounded hours. Add `--target-row <full RawID>` for a TL/operator-selected eligible row. Ambiguity names the ledger path, completion token, eligible IDs with distinguishing metadata, and exact dry-run command. Never select newest automatically.

Historical reconstruction is not an implicit fallback: a fixed-then-ship Review-Window may name only the fix interdiff, which cannot prove the original branch point. If numeric correction cannot trust retained inputs, preserve them and name the exact same-target `--exclude` command. Exclusion already records pin-only provenance without invented hours. On an active branch a fresh re-close produces a valid new generation. No refusal advises hand-typed hours.

## Invariants and transitions

- Candidate close -> fresh inputs -> review -> new completion: measurement code head equals the head reviewed. Fix activity uses existing ownership rules and phase cuts. Prior snapshots remain immutable.
- Completed preview -> exact retained replay: later commits do not advance measurement. Wrong repository, malformed/missing sources, excluded corrections and mismatched bindings still refuse.
- Review REWORK/cancellation -> no completion replacement. Existing repo locking and review revalidation protect concurrent changes.
- Ambiguous correction -> unchanged refusal with target-selection action. Explicit eligible target -> existing measured correction or explicit exclusion. Unknown, superseded or foreign target -> refusal before writes.
- Pending recovery always resumes its pinned target; a conflicting selector refuses. Completed repeat likewise cannot silently ignore a conflicting selector. Existing correction CAS, lease and ledger lock remain authoritative.
- No new concurrency, journal or schema. The selector reuses existing target hashes. Calibration retains one existing-format row per correction, with original rows immutable. No credentials are involved.

## Plan

### Reproduce and fix candidate measurement

**Files:** cmd/sdlc/actualinputs_test.go, closefixed_test.go; cmd/sdlc/actualinputs.go, actual.go, close.go.

- [x] Drive measured whole close -> post-close fix -> measured fixed-then-ship re-close through the production command, with real owned tracker/transcript fixtures. Do not stub measurement. Assert latest ActualInputs.CodeHead equals completion.ReviewedHEAD, fix activity is admitted, historical preview replays the new generation, and publication passes.
- [x] Keep historical mismatch rejection and finalized milestone tests. Add prior numeric-correction/exclusion controls for candidate measurement, and REWORK preservation of the old completion.
- [x] Run RED; distinguish stale-head failure from fixture errors.
- [x] Introduce typed purpose through the shared resolver/engine. Close requests candidate mode, reading current tracker authority and local boundary cuts. Historical actual/correction callers retain replay mode. Finalized milestone previews keep their behavior.
- [x] Run focused GREEN. Restore the old purpose at the production close caller as a mutation; the sequence test must fail.
- [x] Commit and log evidence.

### Support exact legacy correction targets

**Files:** cmd/sdlc/actualcorrection.go, actualcorrection_test.go; internal/actualcorrection tests if recovery intent changes.

- [x] Exercise actualCorrectionTargetForGeneration against adversarial calibration histories whose values coincide but provenance differs. Selection must be within the independently derived eligible effective set; preserve default uniqueness. Mutation removing the eligibility predicate or ignoring the digest must fail a foreign-row or exact-selection assertion.
- [x] Drive runActualCorrection through the real command/portable ledger and tracker, perturbing selection intent before initial admission and after durable progress. Assert exact-target publication on valid intent and fingerprint all stores to prove unchanged refusal/dry-run; replay must preserve its original target. Mutation of selector-to-receipt wiring must fail the durable-target assertion.
- [x] Add optional --target-row. Reuse target eligibility and existing durable target digest. Selection only admits an initial target; recovery cannot retarget it.
- [x] Numeric measurement failure names its cause and exact supported same-target exclusion command. Preserve generation/source/repository checks; malformed evidence cannot become a numeric success.
- [x] Run RED/GREEN and mutate eligibility filtering and selector-to-receipt wiring to prove guards.
- [x] Commit and log evidence.

### Documentation and acceptance

**Files:** README.md; atlas/workflow/sdlc-binary.md; cmd/sdlc/helptext/actual.md, close.md, issue.md; actual.go/close.go diagnostic wording and tests.

- [x] Remove manual/judgment-hour fallback advice from these measured workflows. Explain candidate versus historical inputs, target selection, recovery and exclusion. Sweep README/help/atlas/comments for competing contracts.
- [x] Update issue Log and checkboxes. Run focused actual/adoption/fix-review/correction tests and internal estimate/actualcorrection packages; git diff --check.
- [ ] Push with sdlc pr; recorded full suite via sdlc local-check run --name suite, using the machine gate without overrides.
After verification, close through mandatory review and fix blocking findings. Integrate current main, pass exact-head suite and PR CI, sdlc merge --yes without waivers, verify card/archive/hashes, report the landed outcome to hub and stop.

## Test strategy by risky function

- **resolveOwnedActualInputs / computeActualForCloseFn / adoptOmittedActual:** real close/fix/re-close/preview sequence; distinct source/head pins; prior corrected/excluded controls. Mutation restores historical purpose and must fail the production sequence.
- **actualCorrectionTargetForGeneration:** adversarial histories deliberately share issue/date/value but differ in provenance; independently derive eligible effective IDs and assert selected membership or unchanged refusal. Mutations allowing foreign targets or ignoring the selector must go red.
- **runActualCorrection:** real tracker/main/ledger and production measurement where practical; existing bounded measurement seam only for calibration transaction cases. Assert durable exact target, unchanged other rows, no effects from dry-run, and no retargeting after interruption.
- **printActual/explainActual:** all unmeasurable statuses name supported actions without manual-hour fallback, including production command output.

## Architecture lenses

- ARCH-DRY: one attribution engine, one target predicate, existing correction model/locks.
- ARCH-PURE: purpose and selection deterministic; Git/discovery remain integration.
- ARCH-PURPOSE: first stale adoption and subsequent hard failure, ambiguous correction and unsupported fallback advice all covered.
- ARCH-MOCK: existing portable tracker/transcript/calibration fixtures and judge seam; real Git verifies histories.
- ARCH-CONSTRAINTS: synchronous CLI, existing finite tracker/source limits and cancellation; no extra scans, fan-out or service; machine-gated full suite.
- ARCH-SECURE: validate persisted snapshots/rows; selected digest cannot bypass eligibility or mint hours.
- ARCH-ORDER: explicit candidate/replay purpose; existing correction transitions and exact completion checks; conflicting retry tests.
- ARCH-FUNERAL: no new artifact family or schema; existing snapshot/archive/supersession and recovery cleanup owns residue. Original rows remain audit history.

## Revisions

- 2026-10-15 — PQ-1: replace correction case inventories with named-function adversarial strategies and mechanical mutation oracles. Name the resolver and close caller explicitly. Scope and operator-approved design unchanged.

- 2026-10-15T04:18:20-07:00 — Keep the future close/landing workflow outside verification checkboxes; no scope or acceptance change.
