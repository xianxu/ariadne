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
