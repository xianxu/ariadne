# Boundary Review — ariadne#329 (milestone M2)

| field | value |
|-------|-------|
| issue | 329 — Project datatype: a TL-driven coordination plane (tl: owner, Trains and Decisions sections, lifecycle without a deadline, derived forecast) |
| repo | ariadne |
| issue file | workshop/issues/000329-project-datatype-a-tl-driven-coordination-plane-tl-owner-trains-and-decisions-sections-lifecycle-without-a-deadline-derived-forecast.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 4e488a8fc6a021557e11c0f9f7fb96cfa271bb2d..6eec26cf6a73175f5b45e8c2ebb394a706807c14 |
| command | sdlc milestone-close --issue 329 --milestone M2 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T18:09:50-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checked the diff. Next I'm reading the M2 section of the issue's plan.

Plan and issue read. Now checking D5 against the code and running the tests.

Targeted tests pass. Next: running the validator on this repo, checking `issue validate` for section logic `project validate` could share, and reverting the close fix to confirm its test goes red.

Found one gap: the D10 ops patch isn't recorded anywhere in the repo. Now reverting the close-tick skip to confirm its test fails without it.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The M2 code is correct and well tested, and nothing blocks the boundary. `find` and `resolve` now name the train and car. A record that carries the issue only as a train car is now a match, marked `RowRef=false`. The close tick skips such a record with an info line instead of demanding a detail block for it. `project validate` now checks that the scaffold sections are present. The non-finite ledger fix is done at the parse boundary, so it covers every reader. The targeted tests pass: `internal/project`, `internal/estimate`, and the `cmd/sdlc` tests for find, resolve, validate, close and the TL-owned skip. I also removed the `!m.RowRef` skip in `close.go` and ran the test again: `TestCloseCommands_TrainOnlyProjectUntouched` fails without it, then I restored the file. Running `sdlc project validate --all` with this branch's binary reports the migrated `claimant-ownership.md` as conforming. Two things should be fixed first, both cheap. The plan was not updated for what M2 actually delivered, and the D10 patch file isn't recorded anywhere. Two prose descriptions of the matching rule are also now out of date.

**1. Strengths**
- `cmd/sdlc/internal/project/discover.go:210-220`: the train lookup reuses the M1 `ParseDoc`/`Trains`/`Locate` path instead of parsing the table a second time. A row-ref match and a train match are separate facts (`RowRef`, `Train`), so each consumer can decide what to do.
- `cmd/sdlc/close.go:815-818`: the train-only skip comes before the legacy, TL and edit logic. The test shows the file stays byte-for-byte unchanged under both `close` and `milestone-close`, and it fails without the fix.
- `cmd/sdlc/internal/estimate/ledger.go:273-282`: `parseLedgerFloat` turns Inf/NaN into "absent" at the single parse boundary. That handles the whole class of bad ledger values, not just the one site the earlier finding named.
- `TestProjectFind_NamesTrainAndCar` covers both match shapes (train-only and row-only), uses real separators and stripped parenthesised text, and checks the 2/3 car position.

**2. Critical findings**
None.

**3. Important findings**
- **I-1 (`plan-code-drift`, ARCH-PURPOSE). This is the 2nd finding in this family.** The rule: at every milestone close, each milestone checklist item and each D-decision it touches must be matched to a delivered artifact. Any difference must get an entry in `## Revisions`, in both directions:
  - Promised but not delivered: D10 says to "write a patch file for the TL". No patch file is in the repo, and the Log doesn't give its path. The claim that "the patched copies conform" can't be checked, and Done-when #4 has nothing to point at.
  - Delivered but not planned: M2 also changes the matching rule (a train car alone now counts as a match; D5 only said the fields are "filled by parsing the matched file's Trains"), adds the close-tick skip, adds the `project validate` section check, and migrates `claimant-ownership.md`. None of these has a Revisions entry.
  - Fix: record where the patch lives, and add an M2 Revisions entry covering all of the above.
- **I-2 (`stale-match-rule-restatement`, ARCH-PURPOSE shadow sweep).** Two prose descriptions of the matching rule still say "body contains the `[repo#id]` marker" only. Train-only matches and the new `train`/`car` JSON fields are missing from both:
  - the doc comment on `DiscoverByIssueRef` at `cmd/sdlc/internal/project/discover.go:186-192`;
  - `cmd/sdlc/helptext/resolve.md:54-60`.
  - Fix: update both, or have both point at `DiscoverInRoots` as the single statement of the rule.

**4. Minor findings**
- `TestResolveRun_KindProjectJSONCarriesTrain` (`projectfind_test.go`) never checks that `t.md` is in `res.Files`, so it would pass if `t.md` were missing.
- `missingProjectSections` (`project.go:182`) treats scaffold sections as required sections. If someone adds an optional scaffold section later, every existing project will fail validation. Consider a `required` flag in the model.
- `strings.Split(buf.String(), "p.md")[1][:1]` in the find test is fragile; parse the output line by line instead.
- `Locate` returns the first train only. If an issue appears in two trains, the second is silently ignored. Document this or warn.
- `Trains(...).Warnings` are dropped in discovery. That's fine for best-effort navigation, but `find` could show them.

**5. Test coverage notes**
- The close train-only test fails without the fix (checked by reverting it).
- The non-finite ledger test would fail without the fix (NaN previously parsed with `ok=true`).
- The section-check test checks both the missing and the present cases.
- No test covers a match that is both a row ref and a train car. It would print the train suffix and still be ticked; cheap to add.

**6. Architectural notes**
- ARCH-DRY: pass. The project section loop is specific to projects; `issue.CheckSectionsPresence` has different, status-dependent rules.
- ARCH-PURE: pass. The new logic is the pure `ParseDoc`/`Locate`, and file reads stay in the existing walker.
- ARCH-PURPOSE: flagged in I-1 and I-2.
- ARCH-MOCK: pass. No new external dependency.
- ARCH-CONSTRAINTS: pass. One extra `ParseDoc` per project file per close or find is negligible.
- ARCH-SECURE: pass. Peer project files go through the fuzz-tested trains parser, and ledger numbers are checked at the parse boundary.
- ARCH-ORDER: pass. Discovery and validation hold no state between events; each call is a single read.
- ARCH-FUNERAL: pass. Nothing durable is created; the D10 patch is an artifact handed to the TL outside this repo.

**7. Plan revision recommendations**
Add to `## Revisions`, dated 2026-10-12, M2:
- (a) Matching now counts a train car as a match (`RowRef=false`), and the close tick skips such a record with an info line.
- (b) `project validate` checks the scaffold sections; the changed-instance gate stays frontmatter-only.
- (c) `claimant-ownership.md` is migrated.
- (d) The D10 patch was delivered at `<path>`, along with the read-only conformance results.

```findings
findings:
  - id: new
    severity: Important
    family: plan-code-drift
    title: |
      M2 has no Revisions entry for unplanned scope, and the D10 patch artifact is not recorded anywhere
    detail: |
      2nd finding in this family. Rule: at every milestone close, match each checklist item and D-decision to a delivered artifact and record any difference in Revisions, in both directions. Promised but missing: D10 says to write a patch file for the TL, but no patch is in the range and the Log gives no path, so Done-when 4 cannot be checked. Delivered but unplanned: train-only matching (D5 only said fields are filled for matched files), the close-tick RowRef skip, the project validate section check, and the claimant-ownership.md migration, none recorded in Revisions.
  - id: new
    severity: Important
    family: stale-match-rule-restatement
    title: |
      Prose descriptions of the project match rule still say marker-only after train-only matching was added
    detail: |
      The DiscoverByIssueRef doc comment (discover.go:186-192) and helptext/resolve.md:54-60 both say a match means the body contains the [repo#id] marker. They omit train-car matches and the new train/car JSON fields. Update both, or point both at DiscoverInRoots as the single source of the rule.
  - id: new
    severity: Minor
    family: vacuous-test-assertion
    title: |
      TestResolveRun_KindProjectJSONCarriesTrain does not check that t.md is in the results
    detail: |
      The loop only checks matching entries, so the test passes if the train-only file is missing from res.Files.
  - id: new
    severity: Minor
    family: scaffold-as-requirement
    title: |
      project validate treats every scaffold section as required
    detail: |
      missingProjectSections iterates over the scaffold Sections(), so a future optional scaffold section would make every existing project fail. Consider a required flag in the model.
```

---

## Re-review — 2026-10-12T18:16:09-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 329 — Project datatype: a TL-driven coordination plane (tl: owner, Trains and Decisions sections, lifecycle without a deadline, derived forecast) |
| repo | ariadne |
| issue file | workshop/issues/000329-project-datatype-a-tl-driven-coordination-plane-tl-owner-trains-and-decisions-sections-lifecycle-without-a-deadline-derived-forecast.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 4e488a8fc6a021557e11c0f9f7fb96cfa271bb2d..2e5ef5a3c43b6528d82b14a8fbadd8a78dcaa9a2 |
| command | sdlc milestone-close --issue 329 --milestone M2 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T18:16:09-07:00 |
| verdict | SHIP |

## Review

Checking the issue's M2 plan items next, then verifying the fixes for the prior findings.

Sweeping the remaining match-rule wording, then checking the trains-matching code for edge cases.

Nearly done — checking the ops repo (if present) to verify the patch, then writing up the verdict.

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** All four findings from the earlier round are fixed, and the fixes are backed by evidence. The targeted tests pass at the pinned head: `pkg/vocab`, `internal/project`, `internal/estimate`, and `cmd/sdlc` filtered to ProjectFind, KindProject, TrainOnly, TLOwned, ProjectValidate, MissingSections, ProjectDefaults and ParseRows. The matching rule now has one owner, `DiscoverInRoots`, and every written description of the rule I found now matches it: the code comments, `resolve.md`, `project.md` and the atlas. I'm raising two new Minor findings and nothing that blocks:
- The ops patch no longer applies cleanly to ops as it is today.
- The D10 Revisions entry doesn't mention two things the patch does differently from the plan for `couch-broadcast-web`.

1. **Strengths**
   - `DiscoverInRoots` (`cmd/sdlc/internal/project/discover.go:203-226`) is the only place the matching rule lives. `RowRef` and `Train` are set independently, so close, find and resolve all read the same match result.
   - The close step that ticks project rows skips any record that carries the issue only as a train car (`close.go:815-818`). It runs before the TL and detail-block logic, so a train-only record is never written to and never asks for a detail block. `TestCloseCommands_TrainOnlyProjectUntouched` checks this for both close and milestone-close, byte for byte.
   - `missingSections` (`project.go:189`) is a pure function and is tested without any IO. The file-reading wrapper around it is thin.
   - `parseLedgerFloat` now rejects non-finite numbers in one place, at the parse step, so no reader has to re-check (`ledger.go:273-282`).

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **Patch is already stale.** `workshop/plans/000329-ops-conformance.patch` fails `git apply --check` against the current `~/workspace/ops`. Both files' last hunk (the appended `## Decisions`) fails because the ops Log sections have grown since the patch was cut. It does apply with `-C1`. The TL should apply it with `--3way` or reduced context, or the patch should be regenerated just before handoff. A patch whose last hunk depends on the end of a file that is appended to daily will go stale quickly.
   - **D10 Revisions entry is incomplete.** The patch leaves `couch-broadcast-web` at `status: defined`, while D10 says "sets `status: executing`". It also deletes that file's Estimate text ("Not yet. Sized after…") instead of keeping it as Sizing, which is what it does for `ariadne-robustness-1`. The Revisions entry records neither.

5. **Test coverage notes**
   - BR-11: the `sawTrain` assertion would now fail if `t.md` were missing.
   - BR-12: `TestMissingSectionsSkipsOptional` would fail without the `!sec.Optional` guard, because section C would be reported.
   - Nothing tests that `optional` flows from CUE through `vocabulary export` into the embedded JSON. No section sets it today, so `project.json` is unchanged and that is acceptable for now.

6. **Architecture (each principle)**
   - **ARCH-DRY: pass.** The match rule is written once.
   - **ARCH-PURE: pass.**
   - **ARCH-PURPOSE: pass.** I swept every consumer of `DiscoverByIssueRef`/`DiscoverInRoots` and every written description of the rule; none still says marker-only.
   - **ARCH-MOCK: N/A.** No new external dependency.
   - **ARCH-CONSTRAINTS: pass.** Discovery now parses every project file it walks, which is bounded by fleet size and happens off the keystroke path.
   - **ARCH-SECURE: pass.** If a hand-edited project file fails to parse, matching falls back to the marker alone, and malformed train rows are skipped.
   - **ARCH-ORDER: N/A.** These are stateless read-and-match functions; no state is held between events.
   - **ARCH-FUNERAL: pass.** The patch is part of the issue's plans and is archived with them.

7. **Plan revision recommendations:** add one Revisions bullet saying that the D10 patch leaves `couch-broadcast-web` at `defined`, because its trains table is empty and `trains-present` would refuse `executing`. If dropping its Estimate text was intended, say so in the same bullet.

```findings
dispose:
  - id: BR-9
    disposition: addressed
    note: |
      tl-plan.md Revisions 2026-10-12 M2 entry matches D5/D10 to deliveries and lists the four unplanned items; patch committed at workshop/plans/000329-ops-conformance.patch.
  - id: BR-10
    disposition: addressed
    note: |
      discover.go:187-206 points at DiscoverInRoots as the single rule; helptext/resolve.md:54-58 names row marker OR train car plus train/car JSON; sweep found no remaining marker-only restatement.
  - id: BR-11
    disposition: addressed
    note: |
      projectfind_test.go now asserts sawTrain, so a missing t.md fails the test.
  - id: BR-12
    disposition: addressed
    note: |
      Section.Optional + #ScaffoldSection.optional; TestMissingSectionsSkipsOptional fails without the !sec.Optional guard in missingSections.
findings:
  - id: new
    severity: Minor
    family: plan-code-drift
    title: |
      D10 Revisions omit that the ops patch leaves couch-broadcast-web at status defined and drops its Estimate text
    detail: |
      This is the 3rd finding in family plan-code-drift. The covering rule is the one BR-9 stated: at milestone close, every D-decision clause is matched to the artifact in both directions. D10 says the patch "sets status: executing", but only ariadne-robustness-1 is changed; couch-broadcast-web stays defined (its trains are empty) and its "Not yet. Sized after..." line is deleted rather than kept as Sizing. Record both in the D10 Revisions bullet.
  - id: new
    severity: Minor
    family: stale-handoff-artifact
    title: |
      The ops conformance patch already fails git apply --check against current ops; the appended Decisions hunks no longer match the grown Log tails
    detail: |
      Run against ~/workspace/ops, git apply --check fails at ariadne-robustness-1:687 and couch-broadcast-web:754, and succeeds with -C1. Tell the TL to apply with --3way or reduced context, or regenerate the patch at handoff.
```
