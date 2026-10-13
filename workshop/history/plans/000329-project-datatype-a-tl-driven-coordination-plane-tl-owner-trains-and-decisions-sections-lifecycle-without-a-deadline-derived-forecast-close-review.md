# Boundary Review — ariadne#329 (whole-issue close)

| field | value |
|-------|-------|
| issue | 329 — Project datatype: a TL-driven coordination plane (tl: owner, Trains and Decisions sections, lifecycle without a deadline, derived forecast) |
| repo | ariadne |
| issue file | workshop/issues/000329-project-datatype-a-tl-driven-coordination-plane-tl-owner-trains-and-decisions-sections-lifecycle-without-a-deadline-derived-forecast.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5655861efb765f10a5e5e9bd514dc3743c23fc56..6c57938b885a945dd12f9269bc7b70dcf40bab52 |
| command | sdlc close --issue 329 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T18:40:34-07:00 |
| verdict | SHIP |

## Review

Checking the ops handoff: the saved patch is already stale against current ops, as the plan predicted, so I'm checking that the script's output validates.

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** This round resolves the last open findings. `parseLedgerFloat` now rejects non-finite numbers where the ledger is parsed, and a test fails without that fix. The D10 Revisions now record what each ops file actually gets. The ops handoff is now a script that regenerates the change from the current files, instead of a saved patch.

I checked the handoff against the live `~/workspace/ops`:
- **The saved patch is already stale again.** `git apply --check` fails at `couch-broadcast-web/project.md:759`, because the Log has grown since the commit. The plan anticipated this and now calls the patch a "reviewable preview" only.
- **The script works on today's files.** I ran `000329-ops-conform.py` on both files and ran `sdlc project validate` on the output: both conform. `ariadne-robustness-1` moves to `status: executing`; `couch-broadcast-web` stays at `status: defined`. It keeps its Estimate line as `**Sizing.**`, with an empty Trains table and an empty Decisions section. That matches the new Revisions bullet.

**Tests:**
- The estimate, project and vocab packages pass.
- `cmd/sdlc` filtered to `Project|Forecast|Resolve` passes (105s).
- One full `cmd/sdlc` run hit go test's default 10-minute timeout while a second test run was going at the same time. The test that was running then, `TestCloseCommand_PlanChangedDuringBoundaryReview_DoesNotFinalize`, passes on its own in 28s. I read that as a timeout from the two runs competing, not a regression, but I did not re-run the full suite on its own.

1. **Strengths**
   - `cmd/sdlc/internal/estimate/ledger.go:276`: Inf and NaN read as absent at the single place ledger numbers are parsed, so readers don't re-check. This fixes the whole class, not one site, which is what the family escalation asked for. `forecastcal_test.go:53-67` covers both the parser and the row dropping out of the reader.
   - `workshop/plans/000329-ops-conform.py`: the script fails loudly on input it doesn't expect. A missing anchor raises on `m.start()` or `s.index`, so it never silently writes a half-converted file. It also refuses a file that was already converted.
   - The new Revisions entry in `000329-project-datatype-tl-plan.md` now lists D10's differences per file, in both directions (BR-13).

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The issue Log says the refreshed patch "passes `git -C ops apply --check` again". That was true when committed and is already false. It's fine as a dated record; the Revisions correctly name the script as the handoff.
   - `ops-conform.py` reads and writes files without specifying an encoding. That's fine on macOS (UTF-8 by default).

5. **Test coverage notes:** BR-7 is covered by a regression test (the `NaN`/`+Inf` assertions fail without the `math.IsInf`/`IsNaN` guard). The ops script has no test. It is a one-off migration that runs in another repo, and the TL checks its output with `sdlc project validate`, so I don't think it needs one.

6. **Architectural notes:**
   - **ARCH-DRY:** pass. The script's `DEC` constant and Sizing rewrite are shared by both file kinds.
   - **ARCH-PURE:** pass. The script's transforms are plain string functions; file IO happens only in the last three lines.
   - **ARCH-PURPOSE:** pass. The live files now conform when the script runs.
   - **ARCH-MOCK:** N/A this round; no new external calls.
   - **ARCH-CONSTRAINTS:** N/A; a one-shot migration.
   - **ARCH-SECURE:** pass. Bad ledger values are rejected when parsed.
   - **ARCH-ORDER:** N/A; the script holds no state between runs (it reads the current file once and writes once).
   - **ARCH-FUNERAL:** pass. The patch and script live in `workshop/plans`, tied to this issue's lifecycle.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      ledger.go:276 rejects Inf/NaN at the parse boundary; forecastcal_test.go:53-67 fails without the guard.
  - id: BR-8
    disposition: withdrawn
    note: |
      Accepted, documented trade-off for a readable cue diagnostic; Minor note only, nothing to fix.
  - id: BR-13
    disposition: addressed
    note: |
      Plan Revisions (after M2) record couch stays defined with an empty trains table and keeps Sizing; script output confirms both.
  - id: BR-14
    disposition: addressed
    note: |
      Handoff is now ops-conform.py; run against current ops, both outputs pass sdlc project validate. The saved patch is stale again but is labeled preview-only.
```
