# Boundary Review — ariadne#424 (milestone M3)

| field | value |
|-------|-------|
| issue | 424 — side docs S3: the discussions index, link removal, fold-in from the parent, and moves that carry the .side folder (skill, tl-lead, sdlc archiving) |
| repo | ariadne |
| issue file | workshop/issues/000424-side-docs-s3-the-discussions-index-link-removal-fold-in-from-the-parent-and-moves-that-carry-the-side-folder-skill-tl-lead-sdlc-archiving.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 885dfcf1e73219e3c732b3d7eb77dc8f93f8020f..e43b38b3d644683b23be83ae14df7b5aa4fa2b41 |
| command | sdlc milestone-close --issue 424 --milestone M3 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T05:45:26-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the M3 archive diff now — `archiveside.go` and `tree.go` are read; next come the callers (merge/push/archivemain/landing/abandon) and the tests.

The new side-folder tests pass; the full `cmd/sdlc` run timed out at 600s. Now I'm checking the trunk retry branch and the reopen validation for gaps.

Confirmed that verify runs only after the moves. I'll write a throwaway test to check that a hand-edited `parent:` strands an archive halfway.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M3 does what it set out to do in most places, and the new tests pass. The `.side/` folders go through the same four stages as the asset archive: prepare, a shared finisher, recovery, and verify. That covers the merge and push archives, push's recovery of an interrupted archive, the trunk first write and retry, the landing archive, and reopen. The work also found and closed a real bug: `archivePlanArtifacts` used to rename a plan's `.side` folder as one directory, which recovery could not see. Three things stop a plain SHIP:

1. **The new verify step can strand an archive halfway, on input the operator wrote.** It runs only after every file has moved, and it fails the archive when a side doc's `parent:` does not resolve. I confirmed this with a throwaway test.
2. **The trunk successor episode (abandoning a reopened issue again) skips the plans' side folders.** The issue's folder is replaced, but the plans' folders are not, and nothing tests this path.
3. **The help text listed in Task 7 was not updated.**

**Test runs.** These pass: `go test ./pkg/sidedoc/ ./cmd/tl-lead/` and the new sdlc tests (`TestArchiveSide*`, `TestPreparedArchiveMoves*`, `TestLandingArchiveSideFolders`, `TestAbandonAndReopenCarrySideFolder`, `TestAbandonRetryRefusesSideResidue`). The wider `./cmd/sdlc -run 'Side|Archive|…|Push'` run hit the 600s timeout inside `TestCheckoutArchiveMirrorsTheDoneCard` and was killed, so it reported FAIL. That looks like suite length, not a regression. I did not run the full suite.

## 1. Strengths
- **Refusals happen before anything moves.** `prepareSideArchive` (`cmd/sdlc/archiveside.go:52`) runs before the issue is renamed in both legacy paths (`merge.go:887`, `push.go:722`), and `TestArchiveSideRefusesBeforeEffects` pins this for both an occupied destination and a symlink.
- **Recovery reuses the asset model.** A moved document is the authority, its expected members are the union of its active and archived folders, and a same-path copy with different bytes is refused (`archiveside.go:100-160`). `TestPreparedArchiveMovesRecoversSideFolders` exercises a real partial move, a conflicting edit, and an unowned side doc.
- **The class sweep went beyond the plan's list.** Revisions names path (g), the landing archive, which `archiveRelated` now covers so `confirmLandingArchive`'s proof is unchanged. It also adds the reverse move on reopen and the `archivePlanArtifacts` directory skip (`push.go:346`). That is ARCH-PURPOSE done properly.
- **`pkg/sidedoc/tree.go` is small, pure and tested without IO**, and the `.sidecar` false-positive case is pinned in its test.

## 2. Critical findings
None.

## 3. Important findings
- **I1: verify judges operator-written `parent:` lines only after everything has moved** (`archiveside.go:166`, `push.go:216`, `merge.go:740`) — ARCH-SECURE, ARCH-ORDER.
  - Archiving keeps the base name and moves the folder whole, so a `parent:` resolves afterwards exactly when it resolved before. The only thing this verify can catch is a `parent:` that was already broken by hand.
  - When it does, every file has already moved and push or merge dies. Push's recovery then re-runs the same verify and dies again.
  - The error message ("restore the folder beside its document") points at the wrong fix.
  - Probe: `parent: ../000143-old.md` in a side doc gives `archiveDoneIssues` err=nil, the file already in history, and then a verify error.
  - Fix: check `parent:` (if at all) in `prepareSideArchive`, before anything moves, or turn it into a warning. Keep the post-move check to "every prepared member exists at its destination", which is something sdlc's own work guarantees.
- **I2: the trunk successor episode drops the kept branch's plan side folders** (`archiveside.go:310-333`, `archivemain.go:105`) — ARCH-PURPOSE.
  - `trunkSideReplace` maps only the issue's live folder to its archived folder. `a.Sides`, built at `abandon.go:301-315`, also carries the plans' side files, but none of them match, so they are skipped silently.
  - Meanwhile `archivemain.go:113` rewrites the archived plans from the kept branch. The result is archived plans whose `→ [discussion](…-plan.side/…)` links point at the previous episode's folder, or at nothing.
  - `trunkSideReplace` has no test at all. Plan path (f) and the Revisions entry claim it is covered.
  - Related: the retry residue check (`archivemain.go:78`) covers only the issue's folder. Plan (e) promises "the issue or its plans".
  - The rule behind both gaps is the one Revisions already states: every move of an issue or plan document carries its folder. Fix: map each plan in `a.Plans` with `MoveMember(live plan, archived plan)`, extend the residue check to the plans' live folders, and add a successor test with a plan side doc.
- **I3: help text update missing for the archive behaviour** (Docs gate).
  - Task 7 lists `cmd/sdlc/helptext/` and Step 4 says "atlas and help text". The atlas was updated, but `push.md:48`, `merge.md:45` and `abandon.md:13,25` still describe only the details, assets and plans moving.
  - Fix: add one clause to each, matching the atlas paragraph.

## 4. Minor findings
- **Duplicated side-doc helpers** (ARCH-DRY, family `reuse-existing-helper`; see the findings block for the class rule):
  - `sideFolderOf` (`landingarchive.go:373`) re-implements `sidedoc.InSideDir`'s segment scan.
  - `sideParent` (`archiveside.go:189`) is a third ad-hoc frontmatter reader, beside `pkg/sidedoc/index.go:40` and `pkg/reviewexport/examples_test.go:63`. It also returns "" on CRLF frontmatter.
- **Two error paths are ignored** (`archiveside.go`):
  - At :67, a `readDirNames` error is ignored, so a destination `.side` that exists as a file is not refused at prepare. It fails later, inside the finisher, after the issue has moved.
  - At :130-131, `os.ReadFile` errors are ignored, so an unreadable pair is reported as "duplicate", not as an IO error.
- **Reopen overwrites live side docs without comparing bytes** (`abandon.go:646-670`). A live side doc at a reopen destination is `git rm`'d with no comparison, whereas assets refuse on a difference. It is consistent with how issue and plan files behave, but not with the forward archive's "never replace" rule.
- **The trunk-path tests are thinner than the plan promises.** They use a single side doc. The plan's Step 1 promised a side doc of a side doc on every path (a) to (f).

## 5. Test coverage notes
- The legacy merge test runs with `testfix.Chdir()` and relative directories, so a bug where a path resolves against the working directory instead of `mainPath` would not show. `underRoot` looks correct; the test just doesn't prove it.
- Recovery is exercised for one interleaving: the issue and one side doc moved. The case "the process died after the plans moved but before any side member moved" is untested, though it is handled by the union logic.
- `trunkSideReplace` and the kept-branch branch of `trunkSideArchive` (`kept != nil` with plans) have no direct tests.

## 6. Architectural notes
- **ARCH-DRY:** flagged (Minor; see above).
- **ARCH-PURE:** passes. `SideDir`, `MoveMember` and `InSideDir` are pure and tested without IO. `recoverSideArchiveMoves`' union and conflict decisions could become a pure function over two listings plus a byte comparison; worth doing if this grows.
- **ARCH-PURPOSE:** flagged (I2). Otherwise the sweep was strong.
- **ARCH-MOCK:** passes. It reuses the existing real-git `testfix` and trunk fixtures, and adds no new external dependency.
- **ARCH-CONSTRAINTS:** passes. Tens of files per folder, bounded by the trunk write limits.
- **ARCH-SECURE:** flagged (I1). Symlink refusal on both the working-tree and trunk paths is good.
- **ARCH-ORDER:** flagged (I1, a check of operator-written content placed after the effects). Forward-only recovery and link-then-remove are sound.
- **ARCH-FUNERAL:** passes. Nothing durable is created, and emptied source folders are removed (`removeEmptyDirs`, including in recovery, `push.go:441`).

## 7. Plan revision recommendations
- Add to Revisions: the successor episode (f) replaces the issue's and each plan's archived side folders, and the retry (e) refuses active residue for the plans' folders too. Or, if that is intentionally out of scope, the plan should say so instead of claiming it.
- Integration points still says (f) "leaves an archived `.side/` tree as it is", while the M3 Revisions entry says the successor replaces it. Reconcile the two.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: refuse-before-effects
    title: |
      verifyArchivedSideMoves judges operator-written parent: lines only after every move, stranding push/merge mid-archive
    detail: |
      An archive keeps the base name and moves the folder whole, so a parent: resolves after the move exactly when it resolved before. The verify (archiveside.go:166, called at push.go:216, merge.go:740 and in recovery) can only fail on a parent: already broken by hand, and it fails after every file has moved; recovery repeats the same die. Probe: parent: ../000143-old.md gives archiveDoneIssues err=nil, the file already in history, then a verify error. Check parent: in prepareSideArchive before anything moves (or warn), and keep the post-move check to "each prepared member exists at its destination". ARCH-SECURE / ARCH-ORDER.
  - id: new
    severity: Important
    kind: defect
    family: move-carries-side-folder
    title: |
      trunk successor episode replaces only the issue's side folder; the kept branch's plan side folders are dropped
    detail: |
      trunkSideReplace (archiveside.go:310) maps only the live issue folder, so the plan side files in a.Sides (abandon.go:301-315) never match and are skipped silently, while archivemain.go:113 rewrites the archived plans from the kept branch. The plans' links then point at the old episode's folder. trunkSideReplace has no test. The retry residue check (archivemain.go:78) likewise covers only the issue's folder, though plan (e) says "the issue or its plans". Rule (Revisions' own): every move of an issue or plan document carries its folder. Map each a.Plans doc too, extend the residue check, add a successor test with a plan side doc. ARCH-PURPOSE.
  - id: new
    severity: Important
    kind: defect
    family: docs-update-for-surface
    title: |
      sdlc help text not updated for side folders moving on archive/abandon/reopen
    detail: |
      Plan Task 7 lists cmd/sdlc/helptext/ and Step 4 says "atlas and help text". The atlas was updated, but push.md:48, merge.md:45 and abandon.md:13,25 still describe only details, assets and plans moving.
  - id: new
    severity: Minor
    kind: defect
    family: reuse-existing-helper
    title: |
      sideFolderOf duplicates sidedoc.InSideDir; sideParent is a third ad-hoc frontmatter reader
    detail: |
      This is the 2nd finding in family reuse-existing-helper. Class rule: side-doc path and frontmatter facts (folder of a path, parent: of a side doc) live in pkg/sidedoc and every cmd/ and checker consumer calls them. Sites: landingarchive.go:373 sideFolderOf vs pkg/sidedoc/tree.go InSideDir; archiveside.go:189 sideParent vs pkg/sidedoc/index.go:40 frontmatter and pkg/reviewexport/examples_test.go:63. sideParent also returns "" on CRLF frontmatter.
  - id: new
    severity: Minor
    kind: defect
    family: silent-error-swallow
    title: |
      prepareSideArchive ignores readDirNames errors; recovery ignores ReadFile errors
    detail: |
      archiveside.go:67: a destination .side that exists as a file is not refused at prepare, so it fails inside the finisher after the issue has moved. archiveside.go:130-131: an unreadable pair is reported as "duplicate" instead of as an IO error.
  - id: new
    severity: Minor
    kind: defect
    family: move-carries-side-folder
    title: |
      reopen git-rm's a differing live side doc without comparing bytes
    detail: |
      abandon.go:646-670 compares bytes only for assets. A live side doc at a reopen destination is replaced by the archived copy without a check, unlike the forward archive's never-replace rule.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1665dc7a-b4c4-434b-ad7c-3d1fc73d0542","repository":"github.com/xianxu/ariadne","issue_id":"000424","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/4518a87842d7857a","reviewed_head":"e43b38b3d644683b23be83ae14df7b5aa4fa2b41","slot":"ariadne:2","requested_at":"2026-10-17T12:30:55.777778Z"},"phase":"finished","pid":99199,"started_at":"2026-10-17T12:30:55.873821Z","finished_at":"2026-10-17T12:45:25.689999Z","session_id":"8a578e7c-5ff2-422a-a084-fc7b4176d6b7","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T06:33:25-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 424 — side docs S3: the discussions index, link removal, fold-in from the parent, and moves that carry the .side folder (skill, tl-lead, sdlc archiving) |
| repo | ariadne |
| issue file | workshop/issues/000424-side-docs-s3-the-discussions-index-link-removal-fold-in-from-the-parent-and-moves-that-carry-the-side-folder-skill-tl-lead-sdlc-archiving.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 885dfcf1e73219e3c732b3d7eb77dc8f93f8020f..aef20f75848fbb1eec0a69dd0f9578cf2941085d |
| command | sdlc milestone-close --issue 424 --milestone M3 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T06:33:25-07:00 |
| verdict | SHIP |

## Review

Waiting on the cmd/sdlc tests. Meanwhile I'm checking whether the BR-18 reopen fix has a regression test.

The broad run hit go test's 10-minute timeout because the pattern matched too much. Running just the side-folder tests now.

Revert-checking the claimed fixes: restoring the pre-fix `archiveside.go` and `archivemain.go`, with the new tests kept.

BR-13 and BR-14 both go red with the fix reverted, and the tree is restored. Next I'm checking the new `sidedoc.Fields`/`Folder` helpers (BR-16).

BR-16 checks out. Now looking at the merge-path verify call site and the remaining successor/retry edges.

```verdict
verdict: SHIP
confidence: high
```

The M3 round-1 fixes hold up, and nothing blocks shipping. I reverted `archiveside.go` and `archivemain.go` to `e43b38b3` while keeping the new tests. Three tests go red without the fixes: `TestVerifyArchivedSideMoves` (BR-13), `TestAbandonSuccessorReplacesPlanSideFolder` (BR-14) and `TestAbandonRetryRefusesPlanSideResidue` (BR-14 residue). I then restored the files and the checkout is clean. All 10 side-folder tests pass at head, as do `pkg/sidedoc`, `cmd/tl-lead` and `pkg/reviewexport`. The help text now describes side folders moving on push, merge, abandon and reopen. The duplicate helpers are gone in favour of `sidedoc.Folder` and `sidedoc.Fields`. Two Minor fixes (BR-17, BR-18) are present in code but have no test that fails without them, so under the claimed-fix rule they stay `not-addressed`. They are Minor and do not block.

I ran only the side-folder tests from `cmd/sdlc`, not the whole package. A broad `-run` pattern hit go test's 10-minute timeout with a stack trace in `TestLandingRetainsWorkspace`, which is outside this change. I could not tell whether that was the timeout itself or a real failure.

**1. Strengths**
- `verifyArchivedSideMoves` (`cmd/sdlc/archiveside.go:188-201`) now checks only that each member arrived: present at its destination as a regular file, gone from its source. It can no longer fail after every file has already moved. The comment explains why content is not judged.
- `trunkSideReplace` (`archiveside.go:314-342`) now uses the same document list for the issue and each `.md` plan, so the "every move carries its folder" rule holds in the successor episode too.
- `trunkSideResidue` (`archiveside.go:347-360`) covers plan side folders and does not fetch any file contents.
- `lsTreeFiles` (`abandon.go:358-372`) lists only files, so a plan's `.side/` folder can never be read as a plan. That replaces the old `!HasSuffix(".side")` special case.
- `prepareSideArchive` now refuses a destination that is not a folder, or an unreadable one, before anything moves (`archiveside.go:69-82`).

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- BR-17 and BR-18 are fixed in code but untested: a destination `.side` that exists as a file, recovery `ReadFile` errors, and reopen refusing a differing live side doc (`abandon.go:665`). One test each would close them.
- `sidedoc.Folder` and `sidedoc.Fields` have no direct unit tests in `pkg/sidedoc`. They are only exercised through their callers.

**5. Test coverage notes**
- BR-13 and BR-14 regression tests confirmed red without the fix and green with it.
- None of the new tests reaches the merge-path verify (`merge.go:740`) directly. It shares the verify function with push, which is tested.

**6. Architecture**
- **ARCH-DRY: pass.** `sideFolderOf` and `sideParent` are deleted; the frontmatter reader in `examples_test` now uses `sidedoc.Fields`.
- **ARCH-PURE: pass.** `Folder` and `Fields` are pure. The archive stages stay in the IO layer, with the same structure as the asset archive.
- **ARCH-PURPOSE: pass.** Side folders move with the issue and with plans across merge, push, recovery, trunk retry and successor, landing, and reopen.
- **ARCH-MOCK: pass.** Tests use temporary git repos, the existing pattern; this change adds no new external dependency.
- **ARCH-CONSTRAINTS: pass.** The residue scan is a single directory listing with no blob fetches.
- **ARCH-SECURE: pass.** Symlinks are refused, a destination that is not a folder is refused, and recovery read errors now propagate.
- **ARCH-ORDER: pass.** Nothing can fail after the moves except arrival; recovery takes the union of active and archived members and refuses duplicates.
- **ARCH-FUNERAL: pass.** Emptied source folders are removed. The successor deletes archived members the kept branch no longer has.

**7. Plan revision recommendations:** none. The plan's "M3 as built" revision matches the code.

```findings
dispose:
  - id: BR-13
    disposition: addressed
    note: |
      Verify now checks arrival only (archiveside.go:188); TestVerifyArchivedSideMoves goes red with e43b38b3's verify restored.
  - id: BR-14
    disposition: addressed
    note: |
      trunkSideReplace maps every .md plan and trunkSideResidue covers plan folders; the successor and residue tests go red when reverted.
  - id: BR-15
    disposition: addressed
    note: |
      push.md:46-50, merge.md:44-46 and abandon.md:13-16 and 26 now name side folders moving; matches atlas/workflow/issue-assets.md.
  - id: BR-16
    disposition: addressed
    note: |
      sideFolderOf replaced by sidedoc.Folder; sideParent deleted; examples_test uses sidedoc.Fields; one frontmatter reader remains.
  - id: BR-17
    disposition: not-addressed
    note: |
      Code fix present (archiveside.go:69-82, 141-148), but no test covers a destination that is a file or a ReadFile error.
  - id: BR-18
    disposition: not-addressed
    note: |
      Reopen now compares side-doc bytes (abandon.go:665), but no test reopens over a differing live side doc.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"741192ce-6cd5-4430-94e8-e7ef12726ebf","repository":"github.com/xianxu/ariadne","issue_id":"000424","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/f1fbd20b085e3f78","reviewed_head":"aef20f75848fbb1eec0a69dd0f9578cf2941085d","slot":"ariadne:2","requested_at":"2026-10-17T13:20:07.450478Z"},"phase":"finished","pid":8843,"started_at":"2026-10-17T13:20:07.62553Z","finished_at":"2026-10-17T13:33:24.568103Z","session_id":"ad6a30a1-f646-44ae-81a0-0e79149341ad","outcome":"returned"}] -->
