# Boundary Review — ariadne#259 (whole-issue close)

| field | value |
|-------|-------|
| issue | 259 — close evidence and other status readers drop the first modified path |
| repo | ariadne |
| issue file | workshop/issues/000259-status-paths-first-entry.md |
| boundary | whole-issue close |
| milestone | — |
| window | 968861f13e4d6a20fc281d79b4e080d405b6108d..30d9c17aaea0f60e033f0ac1a58b7cc78b1efe48 |
| command | sdlc close --issue 259 |
| reviewer | claude |
| timestamp | 2026-09-28T09:53:57-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** The fix goes after the root cause. There is now one byte-exact parser, `gitx.ParseStatusZ`, for `--porcelain=v1 -z` output, and `trackerEnv.statusEntries` gets its input without trimming, so the first entry's leading status space survives. Both callers named in the Spec (`closeEvidence` and migrate's `dirtyIssuePaths`) use it, and fleet's count now uses the same parser instead of its own copy. I checked the regression claim myself: in a scratch worktree at head, I put back the base versions of `closetracker.go` and `issuemigrate.go`, and both `TestTrackerReCloseCommitsTheModifiedGateLedger` and `TestIssueMigrateNamesADirtyPathWithASpace` failed. At head, the gitx and fleet packages and the three targeted tests pass. All three `## Done when` clauses are covered. Nothing blocks shipping; the findings below are all Minor.

1. **Strengths**
   - `cmd/sdlc/internal/gitx/status.go:21` fails closed. A trimmed first entry (`"M x\x00"`) is rejected rather than guessed at, and the test pins that case (`status_test.go:29`).
   - Moving fleet's validating counter into gitx gives one parser (ARCH-DRY). Fleet's fake git uses the same `ValidStatusCode`, so the fake and the parser can't drift apart.
   - Renames are handled correctly: in `-z` output `Path` is the destination and `Orig` the source. Migrate's old `dest`-preference logic therefore reduces to `e.Path` without changing behaviour.
   - The re-close test reproduces the real failure: the first close creates the ledger as new (`??`), and the second close has it as modified (` M`). It also checks that `workshop/plans` is left clean.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **`statusEntries` copies `gitEnv`'s exec plumbing** (`trackerenv.go:88`): the command setup, stderr capture and error formatting (ARCH-DRY). A small `gitRaw(dir, args) ([]byte, error)` that `gitEnv` also calls would remove the copy. The copy is cheap to leave.
   - **Some status readers still don't handle paths with spaces.** Merge's `porcelainPaths` (`merge.go:202`) splits on whitespace from trimmed output. Push's `parsePorcelainStatus` (`push.go:516`) slices `line[3:]` from untrimmed `CombinedOutput`, so the first-entry bug doesn't hit it, but it can't handle quoted paths with spaces. Both are outside the Spec's literal claim ("no reader slices fixed columns from trimmed text" holds), and the Log explains why merge was left alone (it fails closed). But the new lesson says to read status through `ParseStatusZ`, and these two readers still don't.
   - **Import grouping in `internal/fleet/facts.go:7`:** the gitx import sits in the standard-library group. goimports would move it to its own group.

5. **Test coverage**
   - The unit test covers modified-first, untracked, rename, spaces, empty input, and four malformed cases.
   - Both end-to-end tests fail without the fix; I checked this by reverting it.
   - Not covered: a copy entry (`C `) and a `" R"`-in-worktree code. They take the same branch as rename, so the risk is low.

6. **Architecture**
   - **ARCH-DRY: pass.** Two parsers became one. The exec copy is noted under Minor.
   - **ARCH-PURE: pass.** `ParseStatusZ` is pure and tested without IO. `statusEntries` is a thin IO layer around it.
   - **ARCH-PURPOSE: pass.** Every consumer named in the Spec now uses the single parser. The readers left out are recorded in the Log and don't drop paths.

   Going forward, new status readers should call `statusEntries` or `ParseStatusZ` and never `git()`.

7. **Plan revisions:** none needed. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: status-read-single-parser
    title: |
      Remaining porcelain readers (merge porcelainPaths, push parsePorcelainStatus) still bypass ParseStatusZ
    detail: |
      merge.go:202 splits on whitespace over trimmed output; push.go:516 slices line[3:] over untrimmed non -z output. Neither drops the first entry, but both mis-handle quoted paths with spaces; the new lesson prescribes ParseStatusZ for status reads.
  - id: new
    severity: Minor
    family: exec-plumbing-duplication
    title: |
      trackerEnv.statusEntries duplicates gitEnv's exec/stderr/error plumbing
    detail: |
      trackerenv.go:88 vs gitEnv; a shared untrimmed gitRaw(dir, args) helper would keep one exec path (ARCH-DRY).
  - id: new
    severity: Minor
    family: import-grouping
    title: |
      fleet/facts.go puts the gitx import in the stdlib import group
```

---

## Re-review — 2026-09-28T10:28:53-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 259 — close evidence and other status readers drop the first modified path |
| repo | ariadne |
| issue file | workshop/issues/000259-status-paths-first-entry.md |
| boundary | whole-issue close |
| milestone | — |
| window | 968861f13e4d6a20fc281d79b4e080d405b6108d..55d34aabdf58cb345e408bb3da352e44b1e99591 |
| command | sdlc close --issue 259 |
| reviewer | claude |
| timestamp | 2026-09-28T10:28:53-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The window does what the issue set out to do. `gitx.ParseStatusZ` (`internal/gitx/status.go:21`) is now the only parser of `git status` paths. Every status reader that uses paths goes through it: close evidence, migrate's dirty list, merge's `worktreeDirty`/`assessDirty`, the start-plan dirty count, push's interrupted-archive recovery, and fleet's count. The two old hand parsers (`porcelainPaths` and `parsePorcelainStatus`) are deleted. The `status --porcelain` calls left in the tree (`changecode.go:288`, `migrate.go:246`, `peerwrite.go:118`, `propagatebase.go:271`, `push.go:112`, `issuemigrate.go:526`, `planningbranch.go:39`) only test whether the output is empty, and trimming can't change that answer. All three prior Minors are fixed. What remains are two advisory Minors and nothing blocks SHIP. I ran the targeted regression tests (`AssessDirty`, `WorktreeDirty`, `PreparedArchive*`, `RecoverInterruptedArchive*`, `IssueMigrateNamesADirtyPathWithASpace`, `TrackerReCloseCommitsTheModifiedGateLedger`) and the gitx and fleet packages; all pass.

**Strengths**
- `ParseStatusZ` refuses a trimmed first entry (`status_test.go:29`) instead of guessing at it. That turns the #259 bug class into a visible error rather than a silently wrong path.
- In `-z` output a rename is written as target first, then source, and the code handles that order correctly. It maps to `Path`/`Orig`, and `push.go:458` swaps back to src→dest for archive-move detection.
- `trackerEnv.gitRaw` (`trackerenv.go:106`) is now the single exec path. `gitEnv` trims its output and `statusEntries` parses it (this fixes BR-2, ARCH-DRY).
- The tests exercise the real parser. The `statusZ`/`statusOf` helpers (`merge_test.go:43`) encode literal porcelain lines as real `-z` bytes, so the merge and push tests run through `ParseStatusZ` instead of a stub. There is also a new case for a path containing a space.
- The end-to-end regression `TestTrackerReCloseCommitsTheModifiedGateLedger` checks the exact failure: the modified ledger must be in the evidence commit and the plans tree must be left clean.

**Critical findings:** none.

**Important findings:** none.

**Minor findings**
- **Stderr can leak into the status stream.** The production `execGitRunner.Git`/`GitInDir` use `CombinedOutput` (`runner.go:37`), so for merge, push recovery and start-plan, git's stderr is mixed into the `-z` stream. Any `git status` warning, such as an unreadable directory, now makes `ParseStatusZ` fail with an opaque "field N is not XY+path status" error. In `startplan.go:455` that error is swallowed, so the dirty count silently reads 0. The parser is right to reject it; the input is the problem. This is the 2nd finding in the `status-read-single-parser` family, so the fix should be the rule, not this instance:
  - **Rule:** status is read from stdout only, through one runner-level helper that runs `status --porcelain=v1 -z` and returns parsed entries.
  - Today the "run + parse" step is written out at 4 sites; that helper would replace all of them.
- **Import grouping in `merge_test.go:6`.** The `gitx` import sits in the standard-library group. This is the 2nd finding in the `import-grouping` family. The rule is `goimports -local github.com/xianxu/ariadne`, enforced by a lint step. Neither `goimports` nor `gci` is installed here, which is why it keeps recurring. I found about 10 files with the same pattern (`branchcreate.go`, `setstatus.go`, `workspace.go`, `workspacepaths.go`, several `_test.go` files), so a one-file fix doesn't solve anything.

**Test coverage:** Good. The parser has unit tests for modified-first, untracked, rename, spaces, and four malformed shapes. Both reported symptoms (close and migrate) have end-to-end regressions. No test covers stderr interleaving; that is outside this diff's contract.

**Architecture**
- **ARCH-DRY:** pass. There is one parser and one exec path in `trackerEnv`; the remaining repetition is noted in the first Minor.
- **ARCH-PURE:** pass. `ParseStatusZ`, `assessDirty` and `preparedArchiveMoves` now take typed entries; only the git calls are shells.
- **ARCH-PURPOSE:** pass. The sweep covered every status reader that uses paths, and I checked each remaining `--porcelain` caller.
- **ARCH-MOCK:** pass. Fleet's fake git validates its codes through the shared `gitx.ValidStatusCode`, and the tests feed real `-z` bytes.
- **ARCH-CONSTRAINTS:** not applicable; the parse is linear over one status read.
- **ARCH-SECURE:** pass. Status output is parsed into a typed value at the boundary, and malformed input is a visible error. The stderr-mixing gap is in the Minor above.
- **ARCH-ORDER:** not applicable. Holds no state between events; it is a single-shot parse.
- **ARCH-FUNERAL:** not applicable. Creates nothing durable beyond the existing gate ledgers.

**Architectural notes for upcoming work:** Adding a stdout-only `StatusEntries` on the git runner would retire the last hand-written "run + parse" repetition and close the stderr gap.

**Plan revisions:** none; the plan matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      porcelainPaths and parsePorcelainStatus deleted; merge/push/startplan read -z via gitx.ParseStatusZ; spaced-path case in TestAssessDirty; remaining --porcelain callers only test emptiness.
  - id: BR-2
    disposition: addressed
    note: |
      trackerenv.go gitRaw is the single exec path; gitEnv trims it, statusEntries parses it.
  - id: BR-3
    disposition: addressed
    note: |
      fleet/facts.go now puts gitx in its own group after the stdlib imports.
findings:
  - id: new
    severity: Minor
    family: status-read-single-parser
    title: |
      Status reads via CombinedOutput runners mix stderr into the -z stream; start-plan swallows the parse error
    detail: |
      2nd in family. runner.go:37 uses CombinedOutput, so a git status warning fails ParseStatusZ opaquely in merge/push and zeroes DirtyCode in startplan.go:455. Rule: status is read stdout-only through one runner-level StatusEntries helper, replacing the 4 run+parse sites.
  - id: new
    severity: Minor
    family: import-grouping
    title: |
      merge_test.go puts the gitx import in the stdlib group
    detail: |
      2nd in family. Rule: goimports -local github.com/xianxu/ariadne enforced by lint; no goimports/gci installed. Measured prevalence about 10 files (branchcreate.go, setstatus.go, workspace.go, workspacepaths.go, several _test.go files), so fix via tooling, not per instance.
```

---

## Re-review — 2026-09-28T10:33:00-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 259 — close evidence and other status readers drop the first modified path |
| repo | ariadne |
| issue file | workshop/issues/000259-status-paths-first-entry.md |
| boundary | whole-issue close |
| milestone | — |
| window | 968861f13e4d6a20fc281d79b4e080d405b6108d..be261d03b7845ecde7f95d0571c9fdc3f45334c4 |
| command | sdlc close --issue 259 |
| reviewer | claude |
| timestamp | 2026-09-28T10:33:00-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This PR delivers what the Spec asks for. `gitx.ParseStatusZ` (`cmd/sdlc/internal/gitx/status.go:21`) is now the only thing that reads paths out of `git status`. Close evidence, migrate's dirty-issue list, merge's dirty check and its re-check, start-plan's dirty count, push's archive recovery and fleet's count all go through it. The hand-written parsers `porcelainPaths` and `parsePorcelainStatus` are deleted. `trackerEnv.gitRaw` is the one exec path. I checked every other `status --porcelain` caller in the tree: `issuemigrate.go:526`, `planningbranch.go:39`, `changecode.go:288`, `peerwrite.go:118`, `propagatebase.go:271` and `landing.go:147` only test whether the output is empty. None reads a path, so the Done-when line "no status reader parses status text by hand" holds. I ran the gitx and fleet packages and the targeted command tests (ReClose, spaced migrate path, PlanningContention, AssessDirty, WorktreeDirty, PreparedArchive, RecoverInterrupted); all pass. The one open item is BR-4's recommended runner-level fix: it wasn't applied. What remains is unclear error messages, not a safety problem, so it doesn't block SHIP.

1. **Strengths**
   - `ParseStatusZ` rejects a trimmed first entry instead of guessing what it was (`status_test.go:28`). That makes the #259 bug class fail loudly.
   - Rename handling is correct across all callers. In `-z` output the new path comes first and the source second; push (`push.go:458-461`) and merge (`merge.go:177`) map `Path` and `Orig` correctly.
   - The start-plan fix now reports a status it can't read as unavailable, never as clean (`startplan.go:455-462`). This matches how merge and push already fail when they can't parse status. `TestPlanningContentionReportsUnreadableStatus` pins it.
   - Test fakes now feed real `-z` bytes through the production parser (`statusZ`/`statusOf` in `merge_test.go`), so the tests exercise the real parsing instead of restating it.
   - The regression tests cover the exact failure seen in the field: a re-close whose gate ledger is modified, plus a dirty path containing a space.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - BR-4 is still open (disposed below).
   - When git itself fails, start-plan's error text leaves out git's own message: `errors.Join(err, perr)` reports "exit status 128" plus "missing NUL terminator" instead of the `fatal:` line (`startplan.go:460`). Same family as BR-4.
   - The comment on `baseContention.Unavailable` (`startplan.go:172`) still says "identity resolution failed", but the field now also carries status and issue-read failures.

5. **Test coverage:** good. Both main regression tests are claimed to fail without the fix. The fake-noise runner test covers start-plan. Merge and push have no test for a noisy or malformed status stream, but their error path is a plain return.

6. **Architecture**
   - **ARCH-DRY:** pass. There is one parser. Four call sites repeat a two-line "run, then parse" pattern; that's what BR-4 would fold into one helper.
   - **ARCH-PURE:** pass. The parser and `assessDirty`/`preparedArchiveMoves` are pure and take parsed entries.
   - **ARCH-PURPOSE:** pass. The sweep covers every reader in the class.
   - **ARCH-MOCK:** pass. The existing runner seam and fleet's fake git reuse `ValidStatusCode`.
   - **ARCH-CONSTRAINTS:** N/A. Status output is small and read once per verb.
   - **ARCH-SECURE:** pass. Git output is untrusted input and is now parsed into a typed value at the boundary; malformed input fails visibly.
   - **ARCH-ORDER:** N/A. Nothing keeps state between events; every read is a one-shot snapshot.
   - **ARCH-FUNERAL:** N/A. The diff creates nothing durable apart from the gate artifacts the SDLC already archives.

7. **Plan revisions:** none. The plan matches the code.

```findings
dispose:
  - id: BR-4
    disposition: not-addressed
    note: |
      Half fixed: start-plan no longer swallows the error; it reports the base as unavailable, and TestPlanningContentionReportsUnreadableStatus pins that. The rule was not applied: execGitRunner (runner.go:37,43) still uses CombinedOutput, so a stderr warning still fails merge and push with an opaque "field 1 is not XY+path status", and start-plan's message drops git's own text. The run-and-parse pattern is still repeated at four sites (merge.go:130, push.go:377, startplan.go:457, trackerenv.go:89). The fix is still one stdout-only StatusEntries helper on the runner. This affects error messages, not safety; everything fails closed.
  - id: BR-5
    disposition: addressed
    note: |
      merge_test.go now groups the gitx import with the project imports, and every import block this diff touches is grouped correctly (checked push.go, startplan.go, trackerenv.go, merge.go and fleet/fakegit_test.go). The tooling rule (goimports -local in lint) covers roughly 10 files that predate this diff and belongs outside this issue.
```
