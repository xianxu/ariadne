# Boundary Review — ariadne#575 (whole-issue close)

| field | value |
|-------|-------|
| issue | 575 — universal-review: one collaborative editing mode with a @@chat@@ section carve-out; mode keys become read-only aliases |
| repo | ariadne |
| issue file | workshop/issues/000575-universal-review-one-collaborative-editing-mode-with-a-chat-section-carve-out.md |
| boundary | whole-issue close |
| milestone | — |
| window | 588a0b0af09a8965d472f7acc216581db3afa815..10bd1d202c13536678cf5e194b733b002c5fcae5 |
| command | sdlc close --issue 575 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T12:11:33-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewed the plan, the issue, and the new `ChatSections` scanner. Next I'm running the test suites and checking a few edge cases.

The tests pass, apart from tl-lead and weave tests that fail only because this sandbox refuses `git init`. Now I'm checking the spec prose and examples against the scanner.

Spec prose matches the model. Last checks: duplicated helpers (ARCH-DRY) and export.json.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The change delivers what the issue asks for. The spec now describes one collaborative mode, the `@@chat@@` carve-out, scope rules s1–s3 and the legacy-key table. Export v4 has five flagless briefs and 15 exported fixture cases. `TestNoSpecWritesModeKeys` enforces the Done-when rule that no spec text asks anyone to write mode keys. `go test ./pkg/reviewexport ./pkg/sidedoc` passes. The review-related subsets of `cmd/weave` and `cmd/tl-lead` also pass. Their full suites fail here only because this sandbox refuses `git init` outside the reviewer's directories; those failures are unrelated to the diff. I mutation-tested `checkChat` by deleting one chat answer link in `chat-tag/after/session.md`, and the test went red as expected. Nothing here is a correctness break. The main gap is that the new reference scanner (`ChatSections`, which pair and parley are told to copy) disagrees with the spec and the fixtures in a few places that no test checks. Each fix is small.

**1. Strengths**
- `pkg/sidedoc/chat.go:59` `ChatSections` is pure and gets headings from goldmark, so a heading inside a fence is treated as code. It sits in the shared side-doc library rather than a third heading scanner, and the issue Log records that choice (ARCH-DRY).
- `exportedCases` (`pkg/reviewexport/export.go:56`) is the only list of exported cases. Both `Build` and the checker read it.
- The checker rule for new side docs (`examples_test.go:354-365`) shares `sidedoc.RetiredKeys` with the scanner, so the checker and the scanner name the same retired keys. Pre-existing legacy side docs keep their bytes (`TestMonthSideDocNoModeKeys`).
- The legacy fixtures (`legacy-keys-ignored`, `legacy-chat-comment`, `legacy-collaboration`) keep every key line byte for byte and add only the index.
- `ParseBrief` refuses retired flags and unknown keys, so v3 flags can't slip back into v4.

**2. Critical**
None.

**3. Important**
- **The spec says "ATX or setext" headings, but the scanner can't support setext.** See `collaboration-modes.md:107-108` and the `chat.go:53` doc comment. In CommonMark, `@@chat@@\nOps\n---` is a single two-line setext heading, so `ChatSections` reports the tag as misplaced. The plan said ATX only. Fix: say "ATX heading" in the spec and the doc comment, and add a test row in `chat_test.go` that pins the setext case as misplaced.
- **An indented-code `@@chat@@` still makes a section chat.** Section detection (`chat.go:120`) compares the previous line's trimmed text and ignores the code mask that the misplaced loop (`chat.go:108`) uses. I ran a probe: `para\n\n    @@chat@@\n## Ops` returns `Source: tag`. Fix: check `literal` before setting `source = "tag"`, and add a test row.
- **No test checks the reply notices against the scanner, and the fixtures have already drifted.** Spec §7 (`collaboration-modes.md:340`) says the reply carries "one notice line per document … as `ChatSections` words them". In practice:
  - `ChatSections` emits several notices: one for the frontmatter key, one for retired keys, one per heading comment, and one per unmatched map name.
  - The `legacy-chat-comment` reply says "the heading comment at line 11". The scanner reports the comment at line 12 of the after document; line 11 is the heading.
  - `checkChat` never reads `scan.Notices`.

  Fix: have `checkChat` assert that every entry in `scan.Notices` appears in the case reply, fix that fixture, and reword §7 and `round.md` §6. Also say whether line numbers refer to the document before or after the write; the write can add a `discussions:` index, which shifts lines.
- **The plan's Core concepts table and Task 5 no longer match the code.**
  - The table puts `chatSections` in `pkg/reviewexport/chat.go`. The code has `ChatSections` in `pkg/sidedoc/chat.go`.
  - Task 5 says the retired-phrase guard covers `adapters/file-edit.md` and the README. `TestNoSpecWritesModeKeys` excludes both on purpose, because of the Couch legacy text.

  The Log records both deviations, but the plan has no `## Revisions` entry for either. The review rubric calls a table/code contradiction Critical; I rated it Important because the Log already documents the deviation.

**4. Minor**
- A legacy map `default: discussion` or `default: debate` gets only the generic "map is a legacy key" notice. The retired value itself is never named, unlike the scalar case.
- In a CRLF document, `frontmatter()` refuses the block, so a legacy `collaboration-mode: chat` is silently not read and no notice says so.
- `adapters/file-edit.md:140` still says "A round in discussion or chat mode".
- A tag directly above the next heading is counted in the previous section's `End`. It's harmless, but untested.

**5. Test coverage notes**
- The `ChatSections` table covers every row the plan lists.
- Still missing: rows for setext headings, indented code, CRLF, and a map default that names a retired value.
- `checkChat` checks unanswered tagged items and misplaced-tag naming, and the link-drop mutation confirms it. It does not check notices or items in legacy chat sections.

**6. Architectural notes**
- ARCH-DRY: pass. There is one case list and one retired-key list. tl-lead's adoption of `@@chat@@` is filed under #576.
- ARCH-PURE: pass. The scanner and `ParseBrief` are pure, and the tests need no mocks.
- ARCH-PURPOSE: pass. The Done-when is about spec text, which is delivered. The remaining Go writers are a separate concern, filed as #576. The gaps in checking the scanner against the spec and fixtures are listed above.
- ARCH-MOCK: N/A. There is no external dependency.
- ARCH-CONSTRAINTS: pass. The document is parsed twice per call, which is fine for files of this size.
- ARCH-SECURE: pass. Malformed YAML degrades instead of crashing, but it does so with no notice (see the CRLF item).
- ARCH-ORDER: N/A. This is single-shot pure scanning with no state between events.
- ARCH-FUNERAL: pass. The retired modes, cases and test were deleted.
- ARCH-EVOLVE: pass. The v4 bump with refusal is justified, because consumers would misread the removed mode flags. The legacy keys in documents are preserved byte for byte.

**7. Plan revision recommendations**
- `## Revisions`: "`chatSections` is `ChatSections` in `pkg/sidedoc/chat.go`, not `pkg/reviewexport/chat.go`, so tl-lead and sdlc reuse it (ARCH-DRY)."
- `## Revisions`: "`TestNoSpecWritesModeKeys` covers `round.md`, `SKILL.md`, `briefs/*` and `collaboration-modes.md` before §7. It excludes `adapters/file-edit.md` and the README, which describe Couch's legacy propose behavior until pair#838."

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: chat-grammar-oracle-parity
    title: |
      Spec and ChatSections doc claim setext headings, but a tag above a setext heading is always misplaced
    detail: |
      collaboration-modes.md:107 and chat.go:53 say "ATX or setext"; CommonMark folds the tag line into the setext heading, so ChatSections reports it misplaced. Plan said ATX only. Fix the prose to ATX and pin setext as misplaced in chat_test.go.
  - id: new
    severity: Important
    kind: defect
    family: chat-grammar-oracle-parity
    title: |
      Indented-code @@chat@@ directly above a heading makes the section chat
    detail: |
      chat.go:120 checks the previous line's trimmed text without the CodeLiteralMask used at chat.go:108; a probe of "para, blank, 4-space-indented tag, then a heading" returns Source tag. Check the mask there and add a test row.
  - id: new
    severity: Important
    kind: defect
    family: fixture-oracle-crosscheck
    title: |
      Reply notices are not cross-checked with ChatSections and the exported fixtures already diverge
    detail: |
      Spec section 7 says replies word notices as ChatSections does, but checkChat ignores scan.Notices. The legacy-chat-comment reply cites line 11 while the oracle reports 12, and the scanner emits several notices where the spec says one line. Assert notices in checkChat, fix the fixture, and state which document the line numbers refer to.
  - id: new
    severity: Important
    kind: defect
    family: plan-table-drift
    title: |
      Plan Core concepts and Task 5 no longer match the code: ChatSections location and guard scope
    detail: |
      The table lists pkg/reviewexport/chat.go, but the code is pkg/sidedoc/chat.go. Task 5 says the guard covers the adapter and README, but TestNoSpecWritesModeKeys excludes both. The Log records both; add Revisions entries.
  - id: new
    severity: Minor
    kind: defect
    family: legacy-key-notice-coverage
    title: |
      Map default of a retired or unknown value is not named, and CRLF frontmatter legacy chat is silently unread
    detail: |
      readFrontmatter emits only the generic map notice for default discussion or debate. frontmatter() refuses CRLF, so a legacy chat scalar is dropped with no notice.
  - id: new
    severity: Minor
    kind: defect
    family: stale-mode-prose
    title: |
      adapters/file-edit.md:140 still says "A round in discussion or chat mode"
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"985e0dd7-a2f5-48ad-a9a9-c0669fd0e544","repository":"github.com/xianxu/ariadne","issue_id":"000575","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/c4e5cd8ddb547d5b","reviewed_head":"10bd1d202c13536678cf5e194b733b002c5fcae5","slot":"ariadne:2","requested_at":"2026-10-23T19:07:55.921707Z"},"phase":"finished","pid":48460,"started_at":"2026-10-23T19:07:56.080757Z","finished_at":"2026-10-23T19:11:32.750156Z","session_id":"e1cb1cda-78ca-4e26-8f83-555fd09b50a1","outcome":"returned"}] -->
