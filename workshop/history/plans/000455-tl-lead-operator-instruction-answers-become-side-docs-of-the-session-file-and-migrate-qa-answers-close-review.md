# Boundary Review — ariadne#455 (whole-issue close)

| field | value |
|-------|-------|
| issue | 455 — tl-lead: operator-instruction answers become side docs of the session file (and migrate qa/ answers) |
| repo | ariadne |
| issue file | workshop/issues/000455-tl-lead-operator-instruction-answers-become-side-docs-of-the-session-file-and-migrate-qa-answers.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5e0e91bc73e248068275f5887d7cf1d5b0102801..023d85a9241a99aac8751d5482937cd4afa7a08c |
| command | sdlc close --issue 455 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T22:50:01-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I've read the diff and am now checking one resume path (registration rename, then crash) and running the tests.

Tests pass. Now I'll confirm the suspected gap where a rerun skips the baseline read.

Confirmed: when the run stops after Rename, the rerun never reads the moved doc's baseline. Next I'll check the help, atlas, and instructions tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The issue is delivered. `link` now enforces side-doc answers (`cmd/tl-lead/answers.go`). The migration has a pure planner over a snapshot and a thin executor. The `SessionSync` seam gains `Apply`, `Registrations` and `Rename`, each with stateful fake behaviour and live conformance. Help and atlas are updated. Both packages' tests pass (`go test ./cmd/tl-lead/ ./pkg/sidedoc/`).

Two Important gaps remain.

1. **A rerun misses part of the Couch handover.** I confirmed this with a throwaway test in the disposable checkout. If the run stops after `Rename` but before `MarkAllSeen`, the rerun reads the file's "registered" state from its old path. That path is no longer registered, so the rerun never takes the new file's baseline. Couch then shows the TL the migration's own frontmatter, backlink and link edits as unseen operator edits. The rerun also never checks that file for an operator lock. The crash-resume test misses this because its oracle compares only the file tree and the registration list.
2. **The gitignore check is missing.** The Spec's "Gitignore follows wherever qa/ was ignored" became a plan promise that the command "checks it, and warns". Nothing implements it, and no Revision drops it.

Neither gap loses data, so the verdict is FIX-THEN-SHIP, not REWORK.

### 1. Strengths
- `pkg/sidedoc/rebase.go` is a well-bounded pure core. The fuzz test asserts two properties: the text is unchanged when nothing moved, and every rewritten target resolves. It already caught one panic, and that input is kept as a regression seed (`testdata/fuzz/FuzzRebaseLinks/8240217d9b6d7fe1`).
- `lineRecords` (`cmd/tl-lead/records.go:61`) proves its own records: it re-applies them and refuses if the result differs from the rewrite. This rules out a whole class of occurrence-count bugs before anything reaches Couch.
- On resume, the marker is pinned from the destination's own frontmatter (`migrate.go:161-164`). So a bullet deleted between runs can't turn the half-finished copy into a false conflict. `TestMigrateAnswersResumeKeepsMarker` covers this.
- The crash test stops at every step boundary and checks the rerun against the golden tree (`migrate_test.go:188`), not against another run. Conflict refusals assert that nothing at all was written.
- The decision to drop `ReadAll` in favour of `fileHash` of a Peek holds up. It keeps the operator's unread session-file edits in the TL's diff, and live conformance pins the hash formula.

### 2. Critical findings
None.

### 3. Important findings
- **`cmd/tl-lead/migrate.go:150` and `migrate_apply.go:185-195`: a resumed move skips `MarkAllSeen` and the lock check (ARCH-ORDER).**
  - **Cause:** `Registered` is `s.Registered[old]` only. After a stop between `Rename` and `MarkAllSeen`, Couch registers the new path, so the rerun has `Registered=false`. It then skips the `Rename` (which would succeed again harmlessly) and the `MarkAllSeen`. `s.Locked` is also never computed for the destination path.
  - **Evidence:** the throwaway probe failed at stop 4 (`2026-10-17-01-s4.md`) and at stop 9 (`…ci.side/01-why.md`). In both cases the doc was never marked seen.
  - **Fix:** set `Registered = s.Registered[old] || s.Registered[new]` when `atDest`, and include the destination paths in `readMigration`'s lock check (`migrate_apply.go:140`).
  - **Regression test:** have the fake's `MarkAllSeen` record the path, and assert in `TestMigrateAnswersResumesFromEveryStep` that every registered moved file was marked seen.
- **The gitignore check from the Spec and plan is not implemented (ARCH-PURPOSE).**
  - The plan says: "The command checks it, and warns if the side folder isn't ignored wherever qa/ was." No code mentions gitignore.
  - **Fix:** add the check (for example `git check-ignore` on a probe path under each folder) and print a warning in `renderPlan`. Or add a plan Revision explaining why ops' existing `.gitignore` makes it unnecessary.

### 4. Minor findings
- **Wrong base for outside files' index entries:** `rebaseIndex` maps `discussions: - doc:` entries as if they were relative to the root, but for outside files (`migrate.go:204`) they are relative to the file. Entries that name `../../qa/…` are therefore never rebased, and a subdirectory file's `qa/op/x` would be mis-mapped. Limit `rebaseIndex` to the session file, or resolve each entry against its own file first.
- **Fake `MarkAllSeen` ignores its path (ARCH-MOCK):** any call marks the session file's sections seen. Tests therefore can't tell whether the migration consumed the operator's unread session-file edits.
- **Repeated lines aren't pinned against real Couch (ARCH-MOCK):** live conformance only checks `occurrence: 1`. `lineRecords` emits `occurrence > 1` for repeated lines, and the counting rule the fake uses (each count starts after the previous match) is unverified against real Couch.
- **`--from` can point outside the team directory (ARCH-SECURE):** a value like `../x` is not refused, although the plan says every path must stay under the team directory.
- **Leftover temp files (ARCH-FUNERAL):** if the process dies between `CreateTemp` and the deferred remove, a `.tl-lead-migrate-*` file is left in the side folder, and no later run removes it.
- **`collaboration-mode` is overwritten:** an answer's existing value (32 ops docs have one) is replaced with `discussion` and the old value isn't kept. That matches the plan, but the dry run doesn't report it.
- **Duplicated temp-write code (ARCH-DRY):** `createExclusive` and `replaceIfUnchanged` repeat the same temp-write-close sequence; it could be one helper.
- **Dry-run evidence not recorded:** the plan's Verification asked for the ops dry-run counts (148 moves, 33 session edits…) to go in the Log, but they aren't there. The live checks are deferred to after landing, as the plan allows.

### 5. Test coverage notes
- The golden test, every-step crash test, conflict table, lock-mid-run test and marker-resume test are all real logic over a temp file tree plus the stateful fake.
- The crash test's oracle checks only the file tree and registrations, not Couch's seen baseline. That is why the Important ARCH-ORDER bug got through.
- `TestInstructionsLinkOnlySideDocAnswers` covers every refusal the plan lists, plus the quoted `parent:` and kept `process-instruction` cases.

### 6. Architectural notes
Each marker, pass or flag:

| Marker | Result | Notes |
|---|---|---|
| ARCH-DRY | pass, one minor | `forUnfenced` and `MoveUnder` are extracted and shared; `AnswerName` exists once. |
| ARCH-PURE | pass | The planner and rebasing are pure. The IO is in `readMigration`, `runMigration` and the seam. |
| ARCH-PURPOSE | flag | The gitignore check is missing; validation is enforced, not just documented. |
| ARCH-MOCK | pass, minors | The fake is stateful (registry, locks, all-or-none apply); see the two fake gaps above. |
| ARCH-CONSTRAINTS | pass | A one-time batch; one walk of the team directory. |
| ARCH-SECURE | pass, one minor | Malformed frontmatter and symlinks become conflicts; `marker:` is YAML-escaped. `--from` isn't confined. |
| ARCH-ORDER | flag | The resume path collapses "registered at the new path" into "not registered". |
| ARCH-FUNERAL | pass, one minor | The emptied qa folders are removed and dot folders are kept; only orphaned temp files after a crash have no removal path. |
| ARCH-EVOLVE | pass | Unknown frontmatter keys are kept in order (`migrate.go:311`); `parent:` keeps its quoting. |

### 7. Plan revision recommendations
- Add a Revision saying `ReadAll` was dropped. The Integration table still lists `SessionSync.ReadAll`, but the code uses `fileHash` plus the existing `MarkAllSeen`.
- Add a Revision saying there is no separate pre-flight pass: the plan is built from a fresh snapshot right before the apply.
- Add a Revision that either schedules the gitignore check or records why it was dropped.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: resume-reads-state-at-one-path
    title: |
      A rerun after a stop between Rename and MarkAllSeen never marks the moved doc seen or checks its lock
    detail: |
      fileMove.Registered is s.Registered[old] only (migrate.go:150), and couch already registers the new path after Rename, so the rerun skips Rename and MarkAllSeen (migrate_apply.go:185-195). Couch then shows the migration's edits to the TL as unseen operator edits. s.Locked is also never computed for destination paths (migrate_apply.go:140). A throwaway probe failed at stops 4 and 9. Fix: Registered = s.Registered[old] || s.Registered[new] when atDest, check locks on destinations too, and add a seenFiles assertion to TestMigrateAnswersResumesFromEveryStep.
  - id: new
    severity: Important
    kind: defect
    family: spec-clause-undelivered
    title: |
      The gitignore check promised by the Spec and plan is not implemented
    detail: |
      The Spec says "Gitignore follows wherever qa/ was ignored", and the plan says the command checks it and warns if the side folder isn't ignored where qa/ was. Nothing in cmd/tl-lead checks gitignore, and no plan Revision drops it. Either add the check with a warning in renderPlan, or add a Revision saying why it is unneeded.
  - id: new
    severity: Minor
    kind: defect
    family: path-base-mismatch
    title: |
      rebaseIndex maps outside files' discussions entries as root-relative, though they are file-relative
    detail: |
      For outside files (migrate.go:204), entries such as ../../qa/... are never rebased, and a subdirectory file's qa/op/x would be mis-mapped. Limit rebaseIndex to the session file, or resolve each entry against its own file.
  - id: new
    severity: Minor
    kind: defect
    family: fake-fidelity
    title: |
      The fake's MarkAllSeen ignores its path, and conformance never pins occurrence > 1
    detail: |
      Any MarkAllSeen call marks the session file seen in the fake. lineRecords can emit occurrence > 1, but live conformance only checks occurrence 1.
  - id: new
    severity: Minor
    kind: defect
    family: path-confinement
    title: |
      --from can name a folder outside the team directory
    detail: |
      The plan says every path stays under the team directory, but readMigration accepts --from ../x.
  - id: new
    severity: Minor
    kind: defect
    family: crash-residue
    title: |
      Temp files from createExclusive and replaceIfUnchanged survive a process death
    detail: |
      A .tl-lead-migrate-* file left in the side folder after a crash has no removal path.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7f48fecb-f530-4274-b5ae-97342aaf75ac","repository":"github.com/xianxu/ariadne","issue_id":"000455","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/7e10d7726a9e5cbb","reviewed_head":"023d85a9241a99aac8751d5482937cd4afa7a08c","slot":"ariadne:4","requested_at":"2026-10-19T05:46:13.496936Z"},"phase":"finished","pid":78929,"started_at":"2026-10-19T05:46:13.726555Z","finished_at":"2026-10-19T05:50:00.125643Z","session_id":"7a4254c0-279e-41c4-bdc1-a455a8ff7a0a","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T23:59:19-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 455 — tl-lead: operator-instruction answers become side docs of the session file (and migrate qa/ answers) |
| repo | ariadne |
| issue file | workshop/issues/000455-tl-lead-operator-instruction-answers-become-side-docs-of-the-session-file-and-migrate-qa-answers.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5e0e91bc73e248068275f5887d7cf1d5b0102801..7b3b597d0ce9de1f6cd252dfc2118aea4fb823fc |
| command | sdlc close --issue 455 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T23:59:19-07:00 |
| verdict | SHIP |

## Review

The round-1 fix commit is what I need to focus on first.

Tests pass. Now revert-verify the BR-1 fix: drop the `|| s.Registered[next]` and lock-at-dest changes and confirm the new assertions go red.

```verdict
verdict: SHIP
confidence: medium
```

All six round-1 findings are fixed in `7b3b597d`, and I checked each fix in the code rather than relying on the commit message. `go test ./cmd/tl-lead/ ./pkg/sidedoc/` passes at head. To confirm the BR-1 fix is pinned, I reverted it in the disposable checkout. `TestMigrateAnswersResumesFromEveryStep` then failed at stop 4 ("never marked seen"), and `TestMigrateAnswersLockAtDestination` failed as well. Reverting the temp-file sweep turns `TestMigrateAnswersSweepsItsTemps` red. The gitignore check is now implemented and tested against a real `git init` repo. The plan has a Revisions entry covering all six fixes. One new Minor test-coverage gap remains, and it does not block SHIP.

**Strengths**
- The resume path now reads state at both the old and the new path (`cmd/tl-lead/migrate.go:156-157`). The every-stop crash test asserts an invariant stated on its own terms (every moved registered doc gets marked seen), not a copy of the implementation.
- Live conformance now applies a record with occurrence 2 and checks the exact text that results (`conformance_test.go:209-226`). That pins the occurrence semantics the fake models.
- `rebaseIndex` now resolves each entry against its own file and reuses the exported `sidedoc.RelPath`, the same helper the links use, so it is one implementation rather than a copy (ARCH-DRY pass).
- `gitignoreGap` is quiet outside a git work tree and when git can't answer. It produces a visible warning, never a refusal (`migrate_apply.go:164-184`).

**Critical findings:** none.

**Important findings:** none.

**Minor findings**
- `migrate_apply.go:147`: the new clause that also checks locks inside the side folder is not pinned by any test. I deleted it and `TestMigrateAnswersLockAtDestination` still passed. The fixture's moved `s4` doc mentions `qa/op/log.jsonl` (testdata after, line 16), so the existing "outside file that names qa" clause fetches its lock anyway. Fix sketch: add a destination fixture doc that never mentions qa.
- `migrate_apply.go:166`: `gitignoreGap` runs `git` directly rather than through a seam. This matches the existing `freshness.go:165`, so it is consistent with the code around it, but `TestMigrateAnswersGitignoreFollows` fails outright on a host without git.

**Test coverage notes:** The resume, lock-at-destination, `--from` confinement, file-relative index and temp-sweep fixes all have tests that exercise them directly. Reverting the fix in the fake's `MarkAllSeen` changes no test result, so the fake is now more accurate but nothing asserts it. That is acceptable for a change to test code only.

**Architecture**
- **ARCH-DRY:** pass (the shared `RelPath` helper).
- **ARCH-PURE:** pass. Planning stays pure and the git probe lives in the IO-side reader.
- **ARCH-PURPOSE:** pass. The Spec's gitignore clause is now delivered.
- **ARCH-MOCK:** minor note above. Rename's idempotency on a rerun is pinned by conformance (`conformance_test.go:256`).
- **ARCH-CONSTRAINTS:** pass. `sweepTemps` walks the team folder once per `--apply` and skips dot-folders.
- **ARCH-SECURE:** pass. `--from` is confined to the team folder.
- **ARCH-ORDER:** pass. The resume table is covered at every stop. The lock-clause gap is noted above.
- **ARCH-FUNERAL:** pass. Temp files now have a removal path.
- **ARCH-EVOLVE:** N/A; no persisted format semantics changed this round.

**Plan revision recommendations:** none. The 2026-10-19 Revisions entry matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Registered and Locked are read at both paths (migrate.go:156-157). Reverting it makes ResumesFromEveryStep fail at stop 4 and LockAtDestination fail.
  - id: BR-2
    disposition: addressed
    note: |
      gitignoreGap (migrate_apply.go:164) warns in renderPlan. TestMigrateAnswersGitignoreFollows covers it with a real git repo.
  - id: BR-3
    disposition: addressed
    note: |
      rebaseIndex resolves each entry against path.Dir(file). TestRebaseIndexIsFileRelative covers the ../../qa case.
  - id: BR-4
    disposition: addressed
    note: |
      The fake's MarkAllSeen is now scoped to its own file. Live conformance applies an occurrence-2 record and checks the exact text. The fake-only change has no failing-without test, which is acceptable for test code.
  - id: BR-5
    disposition: addressed
    note: |
      readMigration refuses absolute paths, ".", ".." and "../" prefixes. TestMigrateAnswersFromStaysInside covers all four.
  - id: BR-6
    disposition: addressed
    note: |
      sweepTemps runs at the start of --apply. Reverting it makes TestMigrateAnswersSweepsItsTemps fail.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: resume-reads-state-at-one-path
    title: |
      No test pins the lock check on side-folder destinations; the fixture reaches it through the outside-file clause
    detail: |
      This is the 2nd finding in family resume-reads-state-at-one-path. The rule is that every piece of per-file state the resume reads (registration, lock, seen baseline) is read at both old and new path, and each read is pinned by a fixture that reaches it only through that clause. Removing the SideDir prefix clause at migrate_apply.go:147 leaves TestMigrateAnswersLockAtDestination green, because the moved s4 doc mentions qa/op/log.jsonl. Add a destination doc that never mentions qa.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a4f4d284-f2e7-4b9e-8c26-69900c1f7ce0","repository":"github.com/xianxu/ariadne","issue_id":"000455","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/a07222c75c04f0de","reviewed_head":"7b3b597d0ce9de1f6cd252dfc2118aea4fb823fc","slot":"ariadne:4","requested_at":"2026-10-19T06:57:30.829121Z"},"phase":"finished","pid":2157,"started_at":"2026-10-19T06:57:31.030313Z","finished_at":"2026-10-19T06:59:19.011734Z","session_id":"29515d65-af61-49f0-b62c-eeb4e6362927","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T00:01:14-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 455 — tl-lead: operator-instruction answers become side docs of the session file (and migrate qa/ answers) |
| repo | ariadne |
| issue file | workshop/issues/000455-tl-lead-operator-instruction-answers-become-side-docs-of-the-session-file-and-migrate-qa-answers.md |
| boundary | whole-issue close |
| milestone | — |
| window | 35c5a1a651e93dfafe6ccbca21d530e317ef1842..b4c241020db8006859a4c857d616962c6122b1a5 |
| command | sdlc close --issue 455 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T00:01:14-07:00 |
| verdict | SHIP |

## Review

Let me revert the side-folder clause and confirm the test goes red.

The test goes red when the clause is removed. Next, a quick check that the issue file's Spec, Revisions and Done-when sections didn't change in this window.

```verdict
verdict: SHIP
confidence: high
```

This window has one code change: `cmd/tl-lead/migrate_test.go`, which answers BR-7. The rest is bookkeeping: two new close-gate and close-review records under `workshop/plans/`, plus the issue file's status, card mirror, hours and close Log line. Nothing in `## Spec`, `## Revisions` or `## Done when` changed, so the scope hasn't moved. The rewritten `TestMigrateAnswersLockAtDestination` now uses the nested side doc `tl-status-op.side/2026-10-17-02-ci.side/01-why.md` as the destination. It stops the run right after that doc's rename, keyed on `f.registered[moved]` rather than on a fragile step count. It also asserts that the fixture doc never contains `qa/op`, so only the side-folder clause can reach it. I checked this directly. I removed `|| strings.HasPrefix(rel, sidedoc.SideDir(s.Session)+"/")` from `cmd/tl-lead/migrate_apply.go:147`, and the test went red: no conflict was reported and the summary showed `0 conflicts`. With the clause restored, all `TestMigrateAnswers*` tests pass.

1. **Strengths**
   - `migrate_test.go:379-383`: the stop condition depends on observable state (the doc's registration has moved), not on a call count. That makes it robust when the order of steps changes.
   - `migrate_test.go:388-390`: the test guards its own isolation. If someone later adds a `qa/op` mention to the fixture doc, the test fails loudly instead of quietly checking the outside-file clause again. This is the class-level protection the family `resume-reads-state-at-one-path` asks for.
2. **Critical findings:** none.
3. **Important findings:** none.
4. **Minor findings:** none.
5. **Test coverage notes:** I confirmed BR-7's regression test fails without the fix (clause reverted, test red, then restored).
6. **Architectural notes**
   - ARCH-DRY: pass. This is a test-only change with no duplicated logic.
   - ARCH-PURE: pass. The test drives the injected fake (`deps{ss: f}`) and runs no new IO.
   - ARCH-PURPOSE: pass. The fixture now isolates the clause BR-7 named.
   - ARCH-EVOLVE: pass, not applicable. No persisted format changed.
7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      The test now uses the nested side doc 2026-10-17-02-ci.side/01-why.md, which never names qa/, and asserts that. With the SideDir clause removed from migrate_apply.go:147 the test fails (0 conflicts); with it restored the test passes.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"fbfc1b94-fefe-4780-871f-d0099844328c","repository":"github.com/xianxu/ariadne","issue_id":"000455","recipe":"fix-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/fdbf839d7c005807","reviewed_head":"b4c241020db8006859a4c857d616962c6122b1a5","slot":"ariadne:4","requested_at":"2026-10-19T07:00:34.739689Z"},"phase":"finished","pid":27311,"started_at":"2026-10-19T07:00:34.841233Z","finished_at":"2026-10-19T07:01:14.398336Z","session_id":"436ffe81-9d44-4c1b-8f54-d8814db158c5","outcome":"returned"}] -->
