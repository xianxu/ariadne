# Boundary Review — ariadne#588 (whole-issue close)

| field | value |
|-------|-------|
| issue | 588 — tl-lead instructions link: trailing space trips the chat guard, and a refusal leaves a half-done write |
| repo | ariadne |
| issue file | workshop/issues/000588-tl-lead-instructions-link-trailing-space-trips-the-chat-guard-and-a-refusal.md |
| boundary | whole-issue close |
| milestone | — |
| window | 500f48f1ea8add6e2b05fc39c9a7fd72edc9157a..54d9672559571d87593307d1a9386a28948031ad |
| command | sdlc close --issue 588 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T00:16:30-07:00 |
| verdict | SHIP |

## Review

Diff and issue read; now running the targeted tests and a revert check.

```verdict
verdict: SHIP
confidence: high
```

Both parts of the issue are fixed, and each fix has a test that fails without it. I checked this by reverting each fix in this checkout. First, `linkBullet` (`cmd/tl-lead/instructions.go:128-129`) now adds the link after the ask without changing any of its bytes, and keeps a trailing `\r` at the end of the line. With the old trimming put back, `TestLinkBullet` and `TestIndexLinkUnderChatGuard` both fail. Second, when the link's write is refused, the new `unlist` puts the `discussions:` index back the way it was. With the `unlist` call removed, `TestIndexLinkRefusedLeavesFile` fails ("a stale link: exit 6").

The design changed during the work, and the change holds up. The plan first wrote everything in one `Apply`. It was replaced with two scoped writes and an undo, because a whole-file apply in couch would mark all of the operator's unread edits as seen. That reason is recorded in the plan's Revisions section, and the fake now models it (`TestFakeApplyAdvancesBaseline`).

The targeted tests pass (`-run 'TestIndex|TestLink|TestInstructions|TestFake'`). Running the whole package has one failure, `TestTeamNewSetupEndsWithExplicitTrackerInit`. It fails because the review sandbox refuses `git init` outside the reviewer's directories, which has nothing to do with this change. The atlas entry was updated. Nothing blocks shipping; only Minor findings remain.

1. **Strengths**
   - `instructions.go:128-129`: the fix is small and correct. The ask is kept as an exact prefix of the new line and the `\r` stays last. The table test at `instructions_test.go:158-166` covers a trailing space, a tab, a CRLF line and a space followed by CRLF.
   - The undo in `unlist` (`instructions.go:213-236`) is careful about when it restores. It restores the old bytes only if the frontmatter is exactly what link's own write produced. If anything else changed it, it falls back to a reconcile. A stale restore also falls back to a reconcile, not to a blind overwrite.
   - The fake now enforces the chat guard (`chatAppendOnly`), and `TestIndexLinkUnderChatGuard` checks that this model rejects the old trimmed output. So the test would have caught the original bug.
   - All three refusal outcomes are tested: the undo restores the old bytes, a frontmatter edit in between leads to a reconcile, and a lock between the two writes refuses the undo and reports the entry left behind.
   - The CRLF limit is written down honestly. `ReconcileIndex` refuses CRLF files, so the CRLF case is tested at the `linkBullet` level only.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **The comment on `linkInstruction` describes the wrong recovery path** (`instructions.go:154-158`). It says "a reconcile drops the entry it orphaned, so the refusal leaves the file as it was". In the code, the main path restores the old bytes; the reconcile is only the fallback. "Leaves the file as it was" is not true on the fallback path (a reconcile may also drop other orphaned entries) or when the undo itself is refused. This is the only claim of this kind in the window: the `unlist` doc comment and the atlas text both describe it correctly.
   - **The reported failure is not run end to end.** The issue's case is a chat-guard refusal after the index write. The tests reach the undo through a stale write and through the lock, and they hit the guard by calling `WriteSection` directly. No test sends a guard refusal through `link` and then the undo. It runs the same code as the lock case, so the gap is small.
   - **`help.md` doesn't explain the new message.** Around line 443 it doesn't say what "stays listed without its link" means or what the TL should do about it (the next reconcile drops the entry, or rerun `link`). The `Short` text was updated.

5. **Test coverage notes:** both "Done when" clauses are covered. The space and tab clause is exercised in both forms, and each form checks that the ask's bytes are intact. The "refused link leaves the document unchanged" clause is checked with `f.text()` compared to the original, apart from the operator's racing edit. Both fixes fail their tests when reverted.

6. **Architectural notes**
   - **ARCH-DRY: pass.** `unlist` reuses `syncIndex`, `sidedoc.ReconcileIndex` and `writeBody`. No answer-link appender exists anywhere else in the tree; I searched for `[answer](`.
   - **ARCH-PURE: pass.** `linkBullet` is still pure and tested without IO. `unlist` is a thin IO step around pure reconcile helpers, and it is tested through the injected fake.
   - **ARCH-PURPOSE: pass.** Both parts of the fix are delivered. The single-`Apply` design was dropped for a documented safety reason, and its follow-up, pair#862 (an apply should advance the baseline only where it edited), is a separate extension, not the point of this issue.
   - **ARCH-EVOLVE: pass.** No persisted format changes.

7. **Plan revision recommendations:** none. The Revisions entry for 2026-10-24 already matches the code.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: doc-claim-from-memory
    title: |
      linkInstruction's comment says a reconcile undoes the index entry and "leaves the file as it was"
    detail: |
      cmd/tl-lead/instructions.go:154-158. The main recovery path in unlist restores the frontmatter byte for byte; the reconcile is only the fallback. On the fallback path, or when the undo itself is refused, the file is not exactly as it was. This is the only instance in the window: unlist's own comment and atlas/workflow/tl-skills.md describe it correctly.
  - id: new
    severity: Minor
    kind: defect
    family: done-when-path-not-end-to-end
    title: |
      No test drives a chat-guard refusal through link and the undo
    detail: |
      TestIndexLinkUnderChatGuard reaches the guard by calling WriteSection directly. Through link, the undo is exercised only by a stale write (TestIndexLinkRefusedLeavesFile) and by the lock, where the undo is refused too. A guard refusal followed by a successful undo, which is the issue's reported case, runs the same branch as the lock case but is not exercised.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"69fdfbeb-9054-4498-bd4f-4853c8333095","repository":"github.com/xianxu/ariadne","issue_id":"000588","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/3fd8cc2f6171ed64","reviewed_head":"54d9672559571d87593307d1a9386a28948031ad","slot":"ariadne:5","requested_at":"2026-10-24T07:14:19.358258Z"},"phase":"finished","pid":81981,"started_at":"2026-10-24T07:14:19.482178Z","finished_at":"2026-10-24T07:16:28.722395Z","session_id":"49bf248d-8148-4d55-89e4-221663659803","outcome":"returned"}] -->

---

## Re-review — 2026-10-24T00:23:59-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 588 — tl-lead instructions link: trailing space trips the chat guard, and a refusal leaves a half-done write |
| repo | ariadne |
| issue file | workshop/issues/000588-tl-lead-instructions-link-trailing-space-trips-the-chat-guard-and-a-refusal.md |
| boundary | whole-issue close |
| milestone | — |
| window | 087a25d271336cad71d42ca78893b96a2e248837..584fcb609daa49b807fbc6f8e88a4ac62a89039f |
| command | sdlc close --issue 588 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T00:23:59-07:00 |
| verdict | SHIP |

## Review

Checking the new guard-refusal test: I'm reverting the undo to confirm only that test block goes red. The many unrelated failures come from the sandbox refusing `git init`.

```verdict
verdict: SHIP
confidence: high
```

Both prior Minors are fixed in this window, and the fixes add no new defects. BR-1: the comment on `linkInstruction` now matches `unlist`. The undo restores the frontmatter's old bytes, and it falls back to a reconcile only when the frontmatter changed in the meantime. BR-2: a chat-guard refusal now runs all the way through `link` and the undo. I removed the call to `unlist` in this checkout, and the new block failed (`index_test.go:108: a guard refusal: …editorial-chat…`). With the code restored it passes, and so do the rest of the targeted tests (`TestIndex|TestLinkBullet|TestFake` → ok). The full `go test ./cmd/tl-lead/` run fails many unrelated tests (operating, session move, team new, fleet and so on). Those tests create git repos, and this sandbox refuses `git init` outside the reviewer's directories ("git init refused … outside the reviewer's directories (sdlc #339)"). This window doesn't touch any of them. The window doesn't change the Spec or Done-when; the issue file only gains the status change, the card mirror, `actual_hours` and Log entries. The help.md addition documents behavior that already existed and was flagged in the prior review. No conflict resolutions are listed.

1. **Strengths**
   - `cmd/tl-lead/sessionsync_fake_test.go:96-98`: `refuseAt` is a general, per-call refusal knob on the fake. It runs before the lock check, so a guard refusal and the lock stay separate cases.
   - `cmd/tl-lead/index_test.go:102-109`: the new case checks that the error passes through (`editorial-chat`), that the file is byte-identical, and that the output says `index: unlisted`. That last check proves the byte-restore path ran, not the reconcile fallback.
   - `cmd/tl-lead/help.md:444-446`: the quoted string "stays listed without its link" matches what `instructions.go:202` actually prints, and the recovery advice (next reconcile, or rerun `link`) is correct.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `instructions.go:157` and `index_test.go:80-82`: after the edit, the comment lines wrap unevenly (one runs long, one stops after "back to its"). Cosmetic only.

5. **Test coverage:** all three ways `link` can be refused now have tests that go through `link` and the undo: a stale write, the lock (where the undo is refused too), and the chat guard. Removing the undo turns the new guard case red.

6. **Architecture (interdiff only)**
   - ARCH-DRY: pass. `refuseAt` doesn't repeat `lockAt`: the lock changes who holds the file, while `refuseAt` makes one specific call fail.
   - ARCH-PURE: pass. The changes are test-only plus comments and help text.
   - ARCH-PURPOSE: pass. BR-2's class is "a refused link, end to end through the undo". All three refusal types are now covered, not just the one the finding named.
   - ARCH-EVOLVE: pass. No persisted format changes.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      instructions.go:156-157 now says unlist restores the frontmatter's old bytes, or reconciles when the frontmatter moved; matches unlist at instructions.go:214-235.
  - id: BR-2
    disposition: addressed
    note: |
      index_test.go:102-109 drives a refuseAt guard refusal through link; with the unlist call removed it fails at index_test.go:108, and it passes on head.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"37ea8ae2-aa0f-4488-8690-dfe2d588ee48","repository":"github.com/xianxu/ariadne","issue_id":"000588","recipe":"fix-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/cd1ef6f6986a94a7","reviewed_head":"584fcb609daa49b807fbc6f8e88a4ac62a89039f","slot":"ariadne:5","requested_at":"2026-10-24T07:21:24.24614Z"},"phase":"finished","pid":59955,"started_at":"2026-10-24T07:21:24.541322Z","finished_at":"2026-10-24T07:23:46.181866Z","session_id":"5f3f0c40-cc86-462f-9366-f0b00b70db24","outcome":"returned"}] -->

---

## Re-review — 2026-10-24T00:25:57-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 588 — tl-lead instructions link: trailing space trips the chat guard, and a refusal leaves a half-done write |
| repo | ariadne |
| issue file | workshop/issues/000588-tl-lead-instructions-link-trailing-space-trips-the-chat-guard-and-a-refusal.md |
| boundary | whole-issue close |
| milestone | — |
| window | e396210aa706c11044fee34fd3155cef952fa6b2..12e1207287eedf8e2b1c0c19a3489e3d8d2d2f83 |
| command | sdlc close --issue 588 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T00:25:57-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This window changes no behavior. It does two things:
- It rewraps two comments in `cmd/tl-lead/instructions.go` and `cmd/tl-lead/index_test.go`, which is the earlier review's cosmetic Minor about uneven wrapping. The words are the same as before; only the line breaks moved.
- It adds generated tracker records: the round-2 gate ledger and the re-review record, plus the card mirror, `actual_hours` and the close Log entry in the issue file.

`## Spec`, `## Revisions` and `## Done when` are unchanged, so scope did not change. No conflict resolutions are listed. `gofmt -l` is clean on both files, and `go test ./cmd/tl-lead/ -run 'TestIndex|TestLinkBullet'` passes. BR-1 and BR-2 were already disposed as addressed in round 2. Nothing in this window reopens them: the BR-1 comment keeps its corrected meaning after the rewrap, and the BR-2 test body is untouched.

1. **Strengths**
   - `cmd/tl-lead/instructions.go:154-159`: the rewrapped doc comment still describes `unlist` correctly: it restores the old bytes, and falls back to a reconcile only when the frontmatter changed in the meantime.
   - `cmd/tl-lead/index_test.go:80-86, 130-132`: the comments now wrap evenly at the file's usual width, which resolves the round-2 cosmetic note.

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none.
5. **Test coverage:** comments only, so no new tests are needed. The targeted tests pass at head.
6. **Architecture (interdiff only):**
   - ARCH-DRY: pass. Nothing is duplicated.
   - ARCH-PURE: pass. No logic changed.
   - ARCH-PURPOSE: pass. No fix was deferred; the earlier cosmetic note is fixed in both files it named.
   - ARCH-EVOLVE: pass. No persisted format changed; the ledger entries only add records.
7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Disposed in round 2; this window only rewraps instructions.go:154-159 and keeps the wording, so the unlist byte-restore and reconcile-fallback description still holds.
  - id: BR-2
    disposition: addressed
    note: |
      Disposed in round 2; this window only rewraps comments in index_test.go and leaves the guard-refusal test body unchanged; targeted TestIndex tests pass at head.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ed69c776-29e6-4016-a61a-5864d6b60eba","repository":"github.com/xianxu/ariadne","issue_id":"000588","recipe":"fix-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/45c09ee112bf05a3","reviewed_head":"12e1207287eedf8e2b1c0c19a3489e3d8d2d2f83","slot":"ariadne:5","requested_at":"2026-10-24T07:25:26.431099Z"},"phase":"finished","pid":69362,"started_at":"2026-10-24T07:25:26.62631Z","finished_at":"2026-10-24T07:25:55.237139Z","session_id":"dfb4a920-fb43-4450-a1a3-18cdb3ee4f04","outcome":"returned"}] -->
