# Boundary Review — ariadne#424 (whole-issue close)

| field | value |
|-------|-------|
| issue | 424 — side docs S3: the discussions index, link removal, fold-in from the parent, and moves that carry the .side folder (skill, tl-lead, sdlc archiving) |
| repo | ariadne |
| issue file | workshop/issues/000424-side-docs-s3-the-discussions-index-link-removal-fold-in-from-the-parent-and-moves-that-carry-the-side-folder-skill-tl-lead-sdlc-archiving.md |
| boundary | whole-issue close |
| milestone | — |
| window | 146796d884ccf9a9d589a8766d7627762b4571e6..5990bf04a5d7cebb58dcf279b19e142ae3f2fe5f |
| command | sdlc close --issue 424 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-17T07:29:12-07:00 |
| verdict | SHIP |

## Review

Checking the four open findings against the code and running the targeted tests now.

Instead, I'll run the two specific BR-17/BR-18 tests directly in the foreground.

The broader run hit Go's 10-minute default timeout (the regex matched many git-heavy archive tests). I'll rerun the side-folder tests on their own.

All side-folder tests pass. Now I'll revert the BR-18 byte comparison to confirm its regression test goes red without it.

BR-18's test fails without the fix, as it should. Last, a quick docs check on the new `tl-lead status index` surface.

```verdict
verdict: SHIP
confidence: high
```

All four open findings are fixed, and I checked each one against the code at the pinned head. Commit messages alone weren't taken as evidence. BR-11 and BR-12 are fixed in `cmd/tl-lead/index.go` and `status.go`, and a new test pins the BR-12 fix. BR-17 is fixed in `cmd/sdlc/archiveside.go`. BR-18 has a regression test that fails when the fix is reverted. The targeted packages pass: `cmd/tl-lead`, `pkg/sidedoc`, `pkg/reviewexport`, and the 14 side-folder, archive, reopen and successor tests matched by `-run Side` in `cmd/sdlc`. Nothing new rises above Minor, so nothing blocks the close.

1. **Strengths**
   - `cmd/sdlc/archiveside.go:56-77`: `prepareSideArchive` refuses every bad destination before anything moves: a file, a symlink, an unreadable folder, or a folder that already holds files. That keeps the archive from stopping halfway.
   - `cmd/sdlc/abandon.go` (reopen restore loop): side docs and assets get the same byte comparison and refusal, so there is one rule for both. Reverting the side-doc half (`|| sidedoc.InSideDir(m[1])`) makes `TestReopenDoneSideFolder` fail. I restored the file afterwards.
   - `cmd/tl-lead/index.go`: `windowChanges` now just keeps the changed lines of the existing `lineDiff`, so the package has one diff algorithm.
   - `recoverSideArchiveMoves` returns a read error as an IO error. It no longer reports an unreadable pair as a "duplicate".

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `index.go` `readAllShowingUnread`: submit skips the second print only when the new diff text matches what it just printed exactly. If an edit lands between the two reads, both diffs are printed in full, though only the new change was unread. That is harmless, since nothing is lost.

5. **Test coverage**
   - Running `go test ./cmd/sdlc -run 'Side|Reopen|Archive|Landing'` in one go passed tl-lead, sidedoc and reviewexport, then hit Go's 10-minute default timeout in `cmd/sdlc`. The pattern matches a large number of git-heavy tests, so I don't think this is a hang. `-run Side` on its own passes in 56 s.
   - I did not run the full suite.

6. **Architecture**
   - **Pass:** ARCH-DRY, ARCH-PURE, ARCH-PURPOSE, ARCH-SECURE and ARCH-ORDER.
     - ARCH-DRY passes now that BR-11 and BR-16 are fixed.
     - ARCH-PURPOSE: the side folder moves with its document in every archive path (merge, push, push's recovery, trunk, landing, reopen, abandon's successor).
     - ARCH-SECURE: symlinks are refused, and frontmatter that can't be parsed is refused rather than guessed at.
     - ARCH-ORDER: the index decides with `Peek`, which leaves couch's seen-baseline alone, and the fake models that baseline.
   - **ARCH-MOCK:** pass. The couch seam has a stateful fake and a conformance test.
   - **ARCH-CONSTRAINTS:** pass. The work is per-archive and bounded by the number of side docs.
   - **ARCH-FUNERAL:** pass. The source folders that empty out are removed, and the archived side docs live and die with their archived document.

7. **Plan revisions:** none. The plan's Log already records the scope beyond (a)–(f): the landing archive, reopen, and abandon's kept-branch side folders.

```findings
dispose:
  - id: BR-11
    disposition: addressed
    note: |
      index.go windowChanges now filters lineDiff output; lineChanges removed (f83186dd).
  - id: BR-12
    disposition: addressed
    note: |
      submit passes its printed diff to syncIndexAfter, which skips an identical reprint; TestSubmitPrintsUnreadOnce pins it.
  - id: BR-17
    disposition: addressed
    note: |
      archiveside.go refuses a non-folder destination and a readDirNames error at prepare; recovery returns ReadFile errors; TestPrepareSideArchiveRefusesFileDestination passes.
  - id: BR-18
    disposition: addressed
    note: |
      reopen byte-compares side docs like assets; reverting the InSideDir clause makes TestReopenDoneSideFolder fail (verified).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a346d263-89c0-4b8a-997b-33364c3098e1","repository":"github.com/xianxu/ariadne","issue_id":"000424","recipe":"milestone-review","round":9,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/92c7a101ea4e6a3a","reviewed_head":"5990bf04a5d7cebb58dcf279b19e142ae3f2fe5f","slot":"ariadne:2","requested_at":"2026-10-17T14:14:56.438235Z"},"phase":"finished","pid":38154,"started_at":"2026-10-17T14:14:56.694041Z","finished_at":"2026-10-17T14:29:11.581166Z","session_id":"febcb445-8b80-4f9a-af60-2a3e616ef19d","outcome":"returned"}] -->
