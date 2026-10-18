# Boundary Review — ariadne#426 (whole-issue close)

| field | value |
|-------|-------|
| issue | 426 — tl-lead status: write the discussions index through couch's frontmatter read/write (pair#643), dropping the printed-diff workaround |
| repo | ariadne |
| issue file | workshop/issues/000426-tl-lead-status-write-the-discussions-index-through-couch-s-frontmatter-read-write-pair-643-dropping-the-printed-diff-workaround.md |
| boundary | whole-issue close |
| milestone | — |
| window | 810f125b368813c581bc80b8963fd29f69ed5550..1278a02061304bf1509f3a016e0be5922e872539 |
| command | sdlc close --issue 426 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T13:08:52-07:00 |
| verdict | SHIP |

## Review

Still sweeping the tree for stale mentions of the old whole-file write before I write up the verdict.

```verdict
verdict: SHIP
confidence: high
```

The change does what the issue asks. `syncIndex` now reads only the frontmatter (`ReadFrontmatter`), gets the body's links from `Peek` (which moves no baseline), and writes only the frontmatter through the hash check (CAS). The printed-diff workaround is gone: `readAllShowingUnread`, `windowChanges`, `syncIndexAfter` and submit's `printed` plumbing are all deleted. The fake now resets only the frontmatter's baseline (`shadow[""]`), and the live conformance test checks that a section edit is still in the diff after both a frontmatter read and a frontmatter write. `go test ./cmd/tl-lead ./pkg/sidedoc` passes in this checkout. I did not run the live conformance test; the issue log says it passed against real couch. What remains is minor: test-code leftovers, one stale comment, and an unused result from `FrontBlock` in `syncIndex`. None of it blocks.

**Strengths**
1. `cmd/tl-lead/index.go:33-60`: the retry loop is simple. It reads the frontmatter, peeks the body, joins `block+body`, reconciles, writes `FrontBlock(next)` with the read's hash, and retries on stale. A lock refusal still goes through `refusal`, so the user gets exit 3.
2. `pkg/sidedoc/index.go:66-83`: `FrontBlock` is pure. Its test covers the round trip (`block+rest == doc`), CRLF, a block with no trailing newline, an unclosed block, no block, and an empty string.
3. `cmd/tl-lead/index_test.go:150-153`: `TestIndexKeepsUnreadEdits` checks the exact diff after the write (`"## Needs you changed\n"`). That shows both that the operator's edit is still there and that the TL's own frontmatter write is not.
4. `conformance_test.go:77-122`: the fake and the live test changed together (ARCH-MOCK). The live test covers a wrong hash (exit 6), body bytes kept, the diff after the read and after the write, and the lock (exit 3), all on a scratch file under `t.TempDir()` as the hub required.
5. `newFake` now starts with `type: session-sync`, like the files couch serves. Some tests had used frontmatter-less files that real couch would refuse; this fixes that drift.

**Critical findings:** none.

**Important findings:** none.

**Minor findings**
- **Unused test helper.** `unreadChanges` (now at `conformance_test.go:164`) has no caller except its own test, `TestUnreadChanges` (`index_test.go:155-166`). The log says it moved to "its one user", but the conformance test no longer calls it. Delete the helper and its test.
- **Stale comment in the fake.** `sessionsync_fake_test.go:162-163` still says "WriteAll's parse returns each body byte…" and refers to "the whole-file write", but `WriteAll` is gone. The other `whole-file` mentions in the fake (lines 29 and 37) are still accurate: they describe the view and `MarkAllSeen`.
- **`FrontBlock(next)` result only partly checked.** At `index.go:51`, `front, _, _ := sidedoc.FrontBlock(next)` throws away `ok` and `rest`. `FrontBlock` accepts a closing line that is `---` after trimming spaces, while `frontmatter()` only accepts an exact `---`. So for a frontmatter containing an indented `  ---` line, a reconciled change that falls after that line would land in `rest` and never be written, yet would still be reported as dropped or added. A cheap guard closes it: refuse unless `ok && rest == body`.

**Test coverage notes:** The Done-when clauses are covered on both sides:
- `index` doing nothing: `TestIndexKeepsUnreadEdits`, first half.
- `link`, which writes: `TestIndexKeepsUnreadEdits`, second half.
- `submit`: `TestSubmitPrintsUnreadOnce`. Its count of exactly 1 would fail if the index printed the diff again.
- A racing frontmatter edit: `TestIndexCommand`. It goes stale, the write retries, the racing `note:` survives, and `frontWriteCalls == 3`.
- The lock: no write happens.

**Architecture**
- **ARCH-DRY:** pass. `ReadAll`/`WriteAll` were removed rather than kept next to the new verbs. One leftover: `FrontBlock` and `frontmatter()` in the same file detect the delimiter differently (trimmed vs exact `---`). That is the root of the minor guard finding above.
- **ARCH-PURE:** pass. `FrontBlock` and `ReconcileIndex` are pure, and `syncIndex` is thin glue over the injected `SessionSync`.
- **ARCH-PURPOSE:** pass. All three callers (`link`, `submit`, `index`) go through the one `syncIndex`. The other `read --all`/`write --all` references (`universal-review.md`, `file-edit.md`) describe the review skill's whole-document flow, which is out of scope here.

**Docs:** the atlas (`tl-skills.md`), `help.md` and the `status index` long help all describe the frontmatter write. No README change is needed because there are no new commands or flags. One small mismatch: the plan names atlas "collaboration-modes 'The index'", but the paragraph that changed is in `tl-skills.md`.

**Plan revision recommendations:** none needed. Optionally, correct the log line "moved into conformance_test.go, its one user": the helper's only user is now its own test.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: dead-test-helper-after-caller-removal
    title: |
      unreadChanges and TestUnreadChanges are left with no production or conformance caller
    detail: |
      unreadChanges (conformance_test.go:164) is called only by its own test, TestUnreadChanges (index_test.go:155-166); the conformance test no longer uses it, though the log says it moved to its one user. Delete both. No other instance of this in the window.
  - id: new
    severity: Minor
    kind: defect
    family: stale-comment-after-api-removal
    title: |
      The fake's text() comment still refers to the removed WriteAll and its whole-file write
    detail: |
      sessionsync_fake_test.go:162-163 still says WriteAll's parse returns each body byte. The other whole-file mentions in the fake (lines 29, 37) remain accurate (the view, MarkAllSeen).
  - id: new
    severity: Minor
    kind: defect
    family: frontmatter-delimiter-rule-divergence
    title: |
      syncIndex discards ok and rest from FrontBlock(next), and FrontBlock's trimmed delimiter differs from frontmatter()'s exact one
    detail: |
      index.go:51 uses front, _, _ := FrontBlock(next). If the frontmatter contains an indented '  ---' line, a reconciled entry after it lands in rest and is never written, yet is reported as dropped or added. Guard with: refuse unless ok and rest == body.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"df736ab6-ca82-4315-977e-34b329d0f70a","repository":"github.com/xianxu/ariadne","issue_id":"000426","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/dced4f9afdca824c","reviewed_head":"1278a02061304bf1509f3a016e0be5922e872539","slot":"ariadne:2","requested_at":"2026-10-18T20:07:35.370427Z"},"phase":"finished","pid":30792,"started_at":"2026-10-18T20:07:35.462545Z","finished_at":"2026-10-18T20:08:52.138448Z","session_id":"cee08e22-3c6e-4862-9f10-119db1b8b584","outcome":"returned"}] -->
