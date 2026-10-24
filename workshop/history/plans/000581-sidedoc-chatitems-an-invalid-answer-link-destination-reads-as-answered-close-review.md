# Boundary Review — ariadne#581 (whole-issue close)

| field | value |
|-------|-------|
| issue | 581 — sidedoc ChatItems: an invalid answer-link destination reads as answered |
| repo | ariadne |
| issue file | workshop/issues/000581-sidedoc-chatitems-an-invalid-answer-link-destination-reads-as-answered.md |
| boundary | whole-issue close |
| milestone | — |
| window | 84b68e2fc5c0773ff38f13771507d93aaa08efd4..faa2aa966807907656a7ffa050c389b76eabedda |
| command | sdlc close --issue 581 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T19:37:50-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I'm checking whether the existing tests cover two of the "Done when" clauses (a link inside a fence, and headings inside an unread header). After that I'll run the tests.

Both fixes go red when reverted. Now I'm checking for other copies of the answer-target rule in the codebase.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Both rules are delivered and tested, and both tests do their job: I removed each fix in the disposable checkout and its test failed. Removing the `ValidAnswerTarget` check from `answered()` turns `TestChatItemsAnsweredNeedsAValidTarget` red. Removing the body-start move in `chatSections` turns both new `TestChatSections` rows red. `linkBullet` and `ChatItems` now use one target rule (ARCH-DRY), and `TestLinkBulletAgreesWithTheReader` checks that they agree. The export manifest only gained entries (ARCH-EVOLVE pass). The atlas, collaboration-modes.md §3 and the examples README were all updated. Two cheap things stand between this and SHIP:
- **Checker too broad:** the examples-checker change in `examples_test.go` hides more errors than its comment says.
- **Clause untested:** one Done-when clause ("no heading") is never checked by a test.

Neither is a correctness bug in the shipped reader.

**Test runs:** `pkg/sidedoc` and `pkg/reviewexport` pass. Every test touched by this change passes in `cmd/tl-lead`, including `TestLinkBullet`, `TestInstructions*` and `TestLinkBulletAgreesWithTheReader`. The rest of `cmd/tl-lead` fails here because the sandbox blocks `git init` ("outside the reviewer's directories"), which is unrelated to this change.

1. **Strengths**
   - `pkg/sidedoc/chat.go:381`: `ValidAnswerTarget` is the single rule. `linkBullet` (`cmd/tl-lead/instructions.go:113`) and `answered()` both call it, and no copy of the old inline condition remains in the diff.
   - `answered()` checks only the link that ends the line. So `→ [answer](broken path) → [answer](a.md)` correctly counts as answered, and there is a test row for it.
   - In `chatSections`, the body starts after the unread header, so headings, tags and the code mask are all skipped there in one move. The notice is kept, and headings after the header keep their real line numbers (the `{8, 9}` row).
   - The test rows include no-break spaces and tabs, which pins the `unicode.IsSpace` behaviour beyond a plain ASCII space.
   - The Done-when clause "a link inside a fence still never answers" was already covered by the `edge` case in `cmd/tl-lead/instructions_test.go:89`.

2. **Critical:** none.

3. **Important**
   - **Checker hides more errors than its comment says.** `pkg/reviewexport/examples_test.go:168`: the new guard `err != nil && len(sidedoc.SideLinks(text)) > 0` ignores *every* `Entries` error in a document with no side links. That includes a readable header whose `discussions:` index is broken (from `indexBlock` or `agreesWithYAML`), not just the unread header the comment talks about. A fixture `after/` file with a hand-broken index and no side link would now pass silently.
     - Fix: ignore only the unread-header case, for example `if err != nil && (len(SideLinks(text)) > 0 || !unreadHeader(text))`, where `unreadHeader` means `FrontBlock` found a closed block that `frontmatter()` refused. Or match on the `errFrontmatter` error.
     - This is the only instance of the problem in this change.
   - **"No heading" is never tested.** In `pkg/sidedoc/chat_test.go:31-43`, the Done-when clause says the unread-header rows yield "no section, no tag, no heading", but the table checks only sections, misplaced tags and notices. A tag would show up as a section, so that is covered indirectly. A `## Notes` heading leaking into `scan.Headings` would not be caught.
     - Fix: add a `headings []int` column, or a direct check of `scan.Headings` and `scan.Tags`, for both rows (spaced closing fence, and CR line endings).

4. **Minor**
   - `chat.go:147`: `TrimPrefix(TrimPrefix(rest, "\r"), "\n")` removes only one CR. A closing fence line like `---\r\r\n` (`FrontBlock` strips all trailing `\r\n`) leaves the body starting with an extra line, so line numbers and code-mask offsets are off by one. Using `_, body, _ = strings.Cut(rest, "\n")` is robust.
   - `cmd/tl-lead/migrate.go:111`: `answerLinkRE` (`[^)\s]*`) is an older, looser restatement of the target shape (it accepts an empty target and an opening paren). It only strips text, so it does no harm, but it is a fourth place that states the rule (ARCH-DRY, ARCH-PURPOSE). Worth a comment or deriving it from the shared rule.
   - The plan's last item is ticked, but "pair:3 and parley.nvim:3 are told the landed revision" can only happen after landing, so it can't be done at this review.

5. **Test coverage:** both fixes have tests that fail without them. Nothing in Go checks what the new fixtures mean (`reviewexport` passes with the header fix removed); they are prose fixtures for pair and parley, which is acceptable.

6. **Architecture:** ARCH-DRY passes for the writer/reader pair; the `migrate.go` regex is noted above. ARCH-PURE passes: the logic is pure functions and the tests use no IO. ARCH-PURPOSE passes: both scope items are delivered, apart from the Important test gap. ARCH-EVOLVE passes: the manifest only gains entries and there is no version bump.

7. **Plan revisions:** add a `## Revisions` note that the examples checker now skips index errors in documents with no side links. Narrow that to unread headers per the Important finding, or record the broader scope on purpose.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: guard-broader-than-stated-case
    title: |
      examples checker drops every Entries error in side-link-free documents, not only unread headers
    detail: |
      pkg/reviewexport/examples_test.go:168 gates on len(SideLinks(text)) > 0, so a readable header with a
      malformed discussions: index (indexBlock/agreesWithYAML error) and no side link now passes silently,
      but the comment claims only the unread-header case is excused. Fix: excuse only the case where FrontBlock
      found a closed block and frontmatter() refused it. This is the only instance in the window.
  - id: new
    severity: Important
    kind: defect
    family: done-when-clause-unasserted
    title: |
      TestChatSections unread-header rows never assert "no heading" (or Tags directly)
    detail: |
      pkg/sidedoc/chat_test.go:31-43: the Done-when clause requires no section, no tag and no heading from inside
      the header, but the table checks only sections, misplaced tags and notices, so a heading leaking into
      scan.Headings goes undetected. Add a headings/tags assertion to both rows (spaced fence, CR endings).
  - id: new
    severity: Minor
    kind: defect
    family: body-offset-derivation
    title: |
      unread-header body trim removes only one CR before the newline
    detail: |
      chat.go:147: a closing fence line like a triple dash followed by two CRs and a newline leaves the body
      starting with a stray line, so line numbers and code-mask offsets are off by one. Use
      strings.Cut(rest, "\n") to take the remainder.
  - id: new
    severity: Minor
    kind: defect
    family: answer-target-single-source
    title: |
      migrate.go answerLinkRE restates a looser answer-target shape
    detail: |
      cmd/tl-lead/migrate.go:111 uses a regex that accepts empty targets and an opening paren. It only strips
      text, so it is harmless, but it is another restatement of the rule ValidAnswerTarget now owns.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ad396c90-c1fb-4eef-b9cf-7188e34bd8e9","repository":"github.com/xianxu/ariadne","issue_id":"000581","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/1166d916b6ecc6ca","reviewed_head":"faa2aa966807907656a7ffa050c389b76eabedda","slot":"ariadne:2","requested_at":"2026-10-24T02:35:21.869455Z"},"phase":"finished","pid":21027,"started_at":"2026-10-24T02:35:22.015753Z","finished_at":"2026-10-24T02:37:48.870495Z","session_id":"35e96f63-19cd-4efb-90ad-fc8f06742871","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T19:44:49-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 581 — sidedoc ChatItems: an invalid answer-link destination reads as answered |
| repo | ariadne |
| issue file | workshop/issues/000581-sidedoc-chatitems-an-invalid-answer-link-destination-reads-as-answered.md |
| boundary | whole-issue close |
| milestone | — |
| window | 84b68e2fc5c0773ff38f13771507d93aaa08efd4..b2c0639bc2f2d058a8f0387790e23f6a908bd9d1 |
| command | sdlc close --issue 581 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T19:44:49-07:00 |
| verdict | SHIP |

## Review

Running the targeted tests now. The tl-lead failures come from the sandbox blocking `git init`, not from this change, so I'll run only the tests this window touches.

Now I'm reverting each round-1 fix one at a time to check that its regression test fails without it.

All four round-1 fixes are confirmed: each regression test fails when its fix is reverted. Next I'm sweeping the tree for any other copies of the answer-target rule and checking the fixtures.

Every round-1 disposition checks out, and the Done-when fence clause is covered by an existing tl-lead test. Last check: whether a test-table entry that looks duplicated is really a no-break space.

```verdict
verdict: SHIP
confidence: high
```

All four round-1 findings are fixed, and I confirmed each one by reverting its fix in this checkout and watching its regression test fail. `ValidAnswerTarget` is now the only answer-target rule in the tree. The reader (`answered()`), the writer (`linkBullet`) and the migrate path (`stripAnswerLinks`) all call it, and a search of non-test Go code found no other copy. The unread-header handling in `chatSections` is correct at its edges:
- `bodyStart` is the closing fence's line + 1.
- `strings.Cut(rest, "\n")` handles a fence on the document's last line and any number of CRs before the newline.
- The tag scan's offsets line up with `CodeLiteralMask(body)`.

`go test ./pkg/sidedoc/ ./pkg/reviewexport/` passes, and so do the new tl-lead tests. Other tests in `cmd/tl-lead` fail only because the sandbox refuses `git init` (sdlc #339), which has nothing to do with this change. The one new item is a Minor mismatch between the plan's wording and the fixture that shipped, so nothing blocks SHIP.

**Strengths**
- `pkg/sidedoc/chat.go:381` — one predicate. `answered()` captures the link target with a regex and lets the predicate decide. The tl-lead test `TestLinkBulletAgreesWithTheReader` checks both the cases the writer accepts and the ones it refuses.
- `pkg/sidedoc/chat.go:142-148` — the unread-header branch is placed after the `hasFM` body setup. The "YAML that does not parse" case already took the `hasFM` path, so only refused blocks are affected.
- `pkg/reviewexport/examples_test.go:169-172` — the narrowed guard excuses exactly one case: a closed block that `frontmatter()` refuses, in a document with no side link. `TestCheckCaseExcusesOnlyAnUnreadHeader` covers all three cases, including a readable header with a malformed index.
- The Done-when clause "a link inside a fence never answers" is already pinned by `cmd/tl-lead/instructions_test.go:90`, which runs ChatItems through `instructionBullets`.

**Critical:** none.

**Important:** none.

**Minor**
- The plan's third checklist item says `frontmatter-unread-tag`'s "body marker opens a side doc". The shipped fixture instead edits the paragraph (scope s1) as a command, because an unread header can't hold an index. The Log mentions this, but the ticked item still says the old thing.
- `pkg/sidedoc/chat_test.go:344`: the `a b.md` row tests the same thing as the `broken path` row (an ASCII space). It's harmless. The no-break-space case is the line above.

**Test coverage notes:** Revert results:
- **BR-1:** the old guard fails the "readable header, malformed index" case.
- **BR-2/BR-3:** removing the header skip fails 5 subtests. Going back to a single-CR trim fails "two CRs before the newline".
- **BR-4:** making `stripAnswerLinks` remove every link fails `TestStripAnswerLinks`. Reverting only the call site in `bulletMarkers` is not caught, which is acceptable for a Minor.

**Architecture**
- **ARCH-DRY: pass.** There is one predicate. The two link-capture regexes in `migrate.go` and `chat.go` now only find the link and leave the decision to the predicate.
- **ARCH-PURE: pass.** All the changed logic is pure parsing, tested without IO.
- **ARCH-PURPOSE: pass.** All consumers in this repo derive from the predicate. pair and parley pick it up through the two exported fixtures.
- **ARCH-EVOLVE: pass.** `export.json` only gains members, with no meaning change and no version bump, as the Done-when requires.

**Plan revision recommendations**
- Add a `## Revisions` entry: "`frontmatter-unread-tag`'s body marker is a command (an edit, s1), not a side doc: an unread header can hold no index, so the round can only edit (see Log 2026-10-23)."

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Guard excuses only closed && !readable && no side link; TestCheckCaseExcusesOnlyAnUnreadHeader goes red (readable malformed index) when reverted to the SideLinks-only gate.
  - id: BR-2
    disposition: addressed
    note: |
      TestChatSectionsIgnoreAnUnreadHeader asserts Headings, Tags, Misplaced, Sections and Notices for spaced fence, CR and two-CR rows; removing the header skip fails all of them.
  - id: BR-3
    disposition: addressed
    note: |
      chat.go:147 uses strings.Cut(rest, "\n"); reverting to a single-CR trim fails the two-CRs subtest (heading at line 8, tag misplaced).
  - id: BR-4
    disposition: addressed
    note: |
      stripAnswerLinks defers to sidedoc.ValidAnswerTarget; TestStripAnswerLinks goes red when it strips every link. No other restatement of the rule remains in non-test Go code.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: plan-claim-matches-delivery
    title: |
      Plan item 3 says frontmatter-unread-tag's body marker opens a side doc; the fixture edits instead
    detail: |
      The ticked plan item says "the body marker opens a side doc", but examples/frontmatter-unread-tag/after/note.md edits the paragraph (scope s1) as a command, since an unread header can hold no index. Only the Log records the change. Add a Revisions entry so the plan describes the shipped fixture. This is the only instance in the window.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1b2b5ec8-7bba-4ad9-94f4-893585c86765","repository":"github.com/xianxu/ariadne","issue_id":"000581","recipe":"small-diff-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/76e9c98973e0aeee","reviewed_head":"b2c0639bc2f2d058a8f0387790e23f6a908bd9d1","slot":"ariadne:2","requested_at":"2026-10-24T02:41:36.004749Z"},"phase":"finished","pid":91088,"started_at":"2026-10-24T02:41:36.1109Z","finished_at":"2026-10-24T02:44:48.992613Z","session_id":"5de31a41-c26a-4249-bf19-9a862f0a89c4","outcome":"returned"}] -->
