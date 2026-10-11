# Boundary Review — ariadne#202 (whole-issue close)

| field | value |
|-------|-------|
| issue | 202 — A finding has no implementer-side disposition: only the judge can retract |
| repo | ariadne |
| issue file | workshop/issues/000202-a-finding-has-no-implementer-side-disposition-only-the-judge-can-retract.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3e80818d6213ac3e9f7c55306a9ccce111cf7cd4..c79efc0373fffd5fd340d6929c7dc6116b46112d |
| command | sdlc close --issue 202 |
| reviewer | claude |
| timestamp | 2026-10-11T00:42:41-07:00 |
| verdict | SHIP |

## Review

Checking how scoped ledgers carry disputes next.

```verdict
verdict: SHIP
confidence: high
```

The change does what the Spec asks. `finding.cue` now has a closing `disputed` disposition, an `authority` partition between judge and implementer, and `appealOnly: hardBlocking`. All consumers derive from the model: `ApplyChecked`, `closedSet`, `DispositionCounts`, the prompt, the fence instruction and the churn columns. Judge rounds and implementer disputes merge into one ordered timeline in a single pure place (`gatestate.timeline`), so there is one ordering rule instead of several restated ones. The verb is a thin IO shell over a pure admission check (`AdmitDispute`), and it refuses before writing anything. I ran the targeted suites (`pkg/vocab`, `gatestate`, `estimate`, `recovery`, and `cmd/sdlc -run 'Dispute|Finding|CloseMetrics|…'`) and all pass. Nothing blocks SHIP. The findings below are Minor.

1. **Strengths**
   - `cmd/sdlc/internal/gatestate/dispute.go:49`: `timeline()` is the only place the two streams are ordered. `closedSet`, `DispositionCounts` and the prompt all read it through `latestStates`, which replaces three copies of the "last disposition wins" loop (ARCH-DRY).
   - `settles()` at `dispute.go:80` reads appeal-only and authority from the model instead of switching on literals. `appealOnly: hardBlocking` is a reference to that set, not a copy, so the two cannot drift apart.
   - `pkg/vocab/finding.go:121`: `RenderBlockInstruction` offers only `JudgeDispositions()`, and `ApplyChecked` drops a judge `disputed` as a protocol error. This closes the hole from both sides. The fence in this very prompt shows only `addressed | withdrawn | not-addressed`, which confirms it.
   - The loudness is built in. `stampAndPersist` prints `DisputeLine()` on every round, whether it passes or not. Both gates are covered, and `TestGatePersist_AnnouncesDisputesOnAPass` pins it.
   - The TSV append is backward compatible: the `disputeCols = 30` guard means rows written before this change still parse.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **Implementer text reaches the judge prompt unfenced.** `renderLiveDisputes` (`prompt.go:125`) inlines the rationale and evidence directly into the instructions. An implementer could word the evidence as an instruction to the judge. The impact is low, because the judge is told to re-run the measurement. Rendering it as quoted data would still be cheap. (ARCH-SECURE)
   - **Help text overstates what a dispute unblocks at the boundary.** `finding-dispute.md` says an Important or Minor "stops blocking". At the boundary, a REWORK verdict still blocks regardless of the ledger, so a judge that ignores a dispute but answers REWORK holds the close. One sentence in the help would fix this.
   - **The plan gate can lose a dispute in theory.** `runPlanQualityJudge` reads its ledger before dispatch and persists afterwards, so a dispute written during the dispatch could be dropped. The boundary path re-reads at persist time. The mutating-command lock very likely serializes these runs, so this is a note rather than a defect.

5. **Test coverage notes:** The tests check real logic:
   - the `<` vs `<=` ordering in the timeline;
   - the appeal-only settle rule;
   - the judge-only fence;
   - every refusal path writing nothing;
   - the render→parse round trip;
   - the end-to-end refusal routing (`TestLedgerGateRefusalsRouteToDispute`).

   No test covers a second `not-addressed` → re-dispute → `withdrawn` sequence across two rounds. `AdmitDispute` covers re-admission, but not the timeline across three events.

6. **Architectural notes**
   - ARCH-DRY: pass.
   - ARCH-PURE: pass. The policy lives in `gatestate` and `vocab`; `finding.go` only does IO.
   - ARCH-PURPOSE: pass. Every consumer named in the Spec derives from the model.
   - ARCH-MOCK: pass. Tests run against temporary repos, and no new external dependency was added.
   - ARCH-CONSTRAINTS: N/A. These are small ledgers on a non-hot path.
   - ARCH-SECURE: Minor (the prompt finding above).
   - ARCH-ORDER: pass. The ordering is explicit in `timeline`, and the judge-vs-implementer interleaving is tested.
   - ARCH-FUNERAL: pass. Disputes are bounded by findings × rounds and archived with the ledger. The TSV growth note in `ledger.go` is updated.

7. **Plan revision recommendations:** none. The Log already records the two design changes (`DisputeLine` on the Decision, and `AdmitDispute` replacing `DecideDispute`).

```findings
findings:
  - id: new
    severity: Minor
    family: untrusted-text-in-judge-prompt
    title: |
      Dispute rationale/evidence is inlined unfenced into the next judge prompt
    detail: |
      renderLiveDisputes (gatestate/prompt.go) emits implementer-authored free text as prompt prose; render it as quoted data so it cannot read as instructions to the judge (ARCH-SECURE).
  - id: new
    severity: Minor
    family: help-overstates-gate-effect
    title: |
      finding-dispute help says Important/Minor "stops blocking" but a REWORK verdict still blocks the boundary
    detail: |
      The dispute settles the ledger side only; close.go's verdict-AND-ledger rule means a judge answering REWORK still holds the close. One sentence in helptext/finding-dispute.md fixes it.
  - id: new
    severity: Minor
    family: ledger-read-modify-write-window
    title: |
      Plan gate reads its ledger before dispatch and persists after, so a dispute written mid-dispatch could be dropped
    detail: |
      The boundary path re-reads at persist time; the plan gate does not. Likely serialized by the mutating-command lock, so this is noted rather than a defect.
```
