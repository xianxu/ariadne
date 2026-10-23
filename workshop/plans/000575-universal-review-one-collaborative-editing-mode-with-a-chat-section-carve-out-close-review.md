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

---

## Re-review — 2026-10-23T12:18:31-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 575 — universal-review: one collaborative editing mode with a @@chat@@ section carve-out; mode keys become read-only aliases |
| repo | ariadne |
| issue file | workshop/issues/000575-universal-review-one-collaborative-editing-mode-with-a-chat-section-carve-out.md |
| boundary | whole-issue close |
| milestone | — |
| window | 588a0b0af09a8965d472f7acc216581db3afa815..f1e4752bd9b7920606ada1739383c73ffc1575f6 |
| command | sdlc close --issue 575 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T12:18:31-07:00 |
| verdict | SHIP |

## Review

Focus on the last fix commit, then chat.go.

I'm checking the fixes by reverting each one. The BR-3 revert went red as expected. The BR-2 revert didn't compile, so I'm redoing it.

```verdict
verdict: SHIP
confidence: high
```

All six prior findings are fixed, and I checked each one in the code, not from the commit message. For the two behavior fixes (BR-2, BR-3) I reverted the fix in this checkout and the regression tests failed; with the fixes restored, `pkg/sidedoc` and `pkg/reviewexport` pass. `TestUniversalReviewSkillDiscoveredInDerivedRepo` in `cmd/weave` passes. The other `cmd/weave` failures come from this review sandbox refusing `git init` ("outside the reviewer's directories (sdlc #339)") and are unrelated to this diff. I found one new problem, Minor and so not blocking: legacy frontmatter can still go unread with no notice, by routes other than CRLF frontmatter. It is the second finding in family `legacy-key-notice-coverage`, so it is written up as a general rule. Nothing blocks SHIP.

1. **Strengths**
   - `pkg/sidedoc/chat.go:109-127`: tags are found once in `tagLine`, checked against the code mask, and the same set is used both for misplaced tags and for tagged sections. Before, two different checks did this, and that is how BR-2 happened. Reverting the mask check fails both the fenced-code and the indented-code test rows.
   - `pkg/reviewexport/examples_test.go:747-754`: `checkChat` now compares every case's reply against `ChatSections(before).Notices`. The fixtures are now checked against the oracle itself, so they can't drift from it again, rather than one fixture being fixed by hand. Restoring the old `legacy-chat-comment/case.md` makes it fail with the exact missing notices.
   - `collaboration-modes.md` §7 now says line numbers count lines of the submitted document, frontmatter included. That matches the code: `readComment` gets `h.line+1`, and line 10 in `before/design.md` is the comment.
   - The plan has `## Revisions` entries covering the new location of `ChatSections` and the guard's real scope (plan lines 189-191).

2. **Critical**: none.

3. **Important**: none.

4. **Minor**
   - `pkg/sidedoc/chat.go:66-74`, `index.go:42`: a legacy `collaboration-mode: chat` scalar is dropped with no notice in three cases I probed:
     - LF frontmatter where some body line has a CR (`frontmatter()` refuses any doc containing `\r`);
     - unparseable YAML (`readFrontmatter` returns silently);
     - a `...` marker inside the frontmatter.

     Each probe returned `{Sections:[] Notices:[]}`. The CRLF fix only covered docs that start with `---\r\n`. This belongs to family `legacy-key-notice-coverage`; see the findings block for the rule.
   - `checkChat` looks for notices anywhere in `case.md`, not only under `## Reply`. That's low risk because the notice strings are long and specific.
   - Plan Task 1 (lines 114, 126) still names `pkg/reviewexport/chat.go`. The Revisions entry covers it.

5. **Test coverage**
   - Setext-tag-is-misplaced, indented-code tag, map `default:` notice and CRLF notice each have a row in `chat_test.go`.
   - Every case's reply is now checked against the oracle's notices.
   - Nothing tests frontmatter that is refused for reasons other than CRLF.

6. **Architecture**
   - ARCH-DRY pass: reuses `frontmatter`, `CodeLiteralMask` and shared `RetiredKeys`.
   - ARCH-PURE pass: `ChatSections` is a pure string-to-value function, tested with no IO.
   - ARCH-PURPOSE pass: every fixture is checked against the oracle, and the stale-mode prose sweep of `construct/local/universal-review`, the atlas and README found nothing left.
   - ARCH-MOCK N/A: no external dependency.
   - ARCH-CONSTRAINTS pass: work grows with the number of headings over document-sized input, which is fine.
   - ARCH-SECURE flag (Minor): unreadable frontmatter in an untrusted doc is silently treated as having no keys (the item above).
   - ARCH-ORDER N/A: it keeps no state between events, because it is a one-shot scan.
   - ARCH-FUNERAL pass: it creates nothing durable, and `modes/` and the retired cases are deleted.
   - ARCH-EVOLVE pass: legacy keys are read around and never rewritten, and `export.json` is regenerated (`TestExportCurrent` is green).

7. **Plan revisions**: none required. Optionally, add a line to the round-1 Revisions entry once the frontmatter-notice rule lands.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Spec line 107 and the chat.go:53 doc now say ATX only; a chat_test.go row pins a setext-adjacent tag as misplaced.
  - id: BR-2
    disposition: addressed
    note: |
      tagLine applies the CodeLiteralMask; reverting the mask check fails the indented-code and fenced test rows.
  - id: BR-3
    disposition: addressed
    note: |
      checkChat asserts the oracle notices verbatim; reverting legacy-chat-comment/case.md fails TestExamples; spec section 7 states the line-number basis.
  - id: BR-4
    disposition: addressed
    note: |
      The Core concepts row now says pkg/sidedoc/chat.go; Revisions lines 189-190 record the move and the guard's actual scope.
  - id: BR-5
    disposition: addressed
    note: |
      Map default notice and CRLF-prefix notice added, with test rows; the remaining unread-frontmatter routes are raised below as the family rule.
  - id: BR-6
    disposition: addressed
    note: |
      file-edit.md:140 reworded; a grep sweep of construct/local/universal-review, the atlas and README finds no remaining mode-era phrasing.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: legacy-key-notice-coverage
    title: |
      Frontmatter that frontmatter() refuses or YAML cannot parse drops legacy keys with no notice
    detail: |
      This is the 2nd finding in family legacy-key-notice-coverage, so fix the rule, not this instance. Rule: when the doc opens with a --- fence and its keys were not read for any reason (frontmatter() refused, any CR in the doc, a ... or % line, or yaml.Unmarshal error), ChatSections emits one notice. Probes in which a legacy chat scalar vanished silently: LF frontmatter plus a CR in the body, a malformed flow sequence, a ... marker. Replace the CRLF-prefix check at chat.go:67 and the silent return at chat.go:173 with that one rule, and add rows for each route.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7ff57d0e-7ef2-4b9e-86b6-5124415082f0","repository":"github.com/xianxu/ariadne","issue_id":"000575","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/e8a01914f50112e4","reviewed_head":"f1e4752bd9b7920606ada1739383c73ffc1575f6","slot":"ariadne:2","requested_at":"2026-10-23T19:16:09.565549Z"},"phase":"finished","pid":67873,"started_at":"2026-10-23T19:16:09.774686Z","finished_at":"2026-10-23T19:18:30.916027Z","session_id":"2135aa92-3250-486d-8fb1-ce9cbfadad1d","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T12:22:01-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 575 — universal-review: one collaborative editing mode with a @@chat@@ section carve-out; mode keys become read-only aliases |
| repo | ariadne |
| issue file | workshop/issues/000575-universal-review-one-collaborative-editing-mode-with-a-chat-section-carve-out.md |
| boundary | whole-issue close |
| milestone | — |
| window | 588a0b0af09a8965d472f7acc216581db3afa815..213f0fa2d58f44975260da50b823ea9f57760f41 |
| command | sdlc close --issue 575 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T12:22:01-07:00 |
| verdict | SHIP |

## Review

The open finding is BR-7 in chat.go. Let me read it.

Status: BR-7's fix is confirmed. Reverting it turns its new tests red, and the tree is restored. Next I'm probing edge cases of the new rule and checking the rest of the diff.

Wrapping up: BR-7 holds up, and I'm writing the verdict now.

```verdict
verdict: SHIP
confidence: high
```

The one open finding, BR-7, is fixed the way it asked: one rule now covers every route, not just the next instance. `pkg/sidedoc/chat.go:73` gives the "frontmatter could not be read" notice to any document that opens with a `---` fence when `frontmatter()` refuses it. That covers a CR anywhere in the document, a `...`, `---` or `%` line inside the frontmatter, and an unclosed fence. `chat.go:178` gives the same notice when the YAML does not parse. The two can't fire on the same document: when `frontmatter()` refuses, `fm` is empty and `readFrontmatter` returns early. Each route has its own test row (`chat_test.go:67-86`). I restored the pre-fix `chat.go` from `213f0fa2d^` and four rows went red: CRLF, LF frontmatter with a CR in the body, a document marker, and YAML that does not parse. I then put the file back, and the tree is clean. The spec text matches the code (`collaboration-modes.md:344`). `go test ./pkg/sidedoc/ ./pkg/reviewexport/` passes, and so does `go test ./cmd/weave/ -run 'Universal|Review'`. Some other `cmd/weave` tests fail because the sandbox refuses `git init` ("outside the reviewer's directories (sdlc #339)"). That is the environment, not this diff. Nothing new blocks shipping.

1. **Strengths**
   - `chat.go:50`: one `unreadFrontmatter` constant feeds both routes, so the wording can't drift between them (ARCH-DRY).
   - `ChatSections` is pure (only string and goldmark parsing). The test table runs with no IO, and the fixtures are checked against it through `examples_test.go:699-750` (`checkChat`). The fixtures and the code that decides chat sections stay in step.
   - `ParseBrief` (`export.go`) rejects retired mode flags and unknown keys with named errors. A v3 mode file can't pass as a brief by accident.
   - `exportedCases` is the single list read by both `Build` and `TestExportCurrent`.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings** (not raised as findings)
   - A document whose first line is a thematic break (`---\n\ntext`) now gets the notice even though it has no frontmatter, as my probe confirmed. The wording ("any legacy mode keys … were not read") makes this harmless.
   - A document with a BOM before its first `---` gets no notice. This is consistent with `frontmatter()`, which also doesn't treat it as frontmatter.

5. **Test coverage notes:** each route that leaves frontmatter unread has a row, and reverting the fix turns four rows red. The empty-frontmatter case (`---\n---\n`) correctly produces no notice. I probed it but there is no table row for it, which is optional.

6. **Architecture**
   - ARCH-DRY: pass.
   - ARCH-PURE: pass. `ChatSections` is pure and IO stays in the export shell.
   - ARCH-PURPOSE: pass. BR-7 was fixed at the rule level.
   - ARCH-MOCK: pass. No new external dependencies.
   - ARCH-CONSTRAINTS: pass. The scan is linear in document size.
   - ARCH-SECURE: pass. Malformed persisted frontmatter now shows up as a notice instead of being ignored silently.
   - ARCH-ORDER: pass. `ChatSections` keeps no state between calls; it is a single-shot scan.
   - ARCH-FUNERAL: pass. Nothing durable is added beyond the fixtures, which are versioned.
   - ARCH-EVOLVE: pass. Going to Version 4 removes the `mode` role and its flags. That changes what the manifest means, and the version comment documents it, so refusal by v3 readers is justified.

7. **Plan revision recommendations:** add a Revisions entry for round 2 at `plan.md:191`. Round 1 says "unread CRLF frontmatter are now noticed"; the rule is now broader, covering every unread fence (CR, document marker or directive, YAML that does not parse).

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      chat.go:73 and chat.go:178 apply one rule for every unread fence; chat_test.go:67-86 rows go red when the pre-fix chat.go is restored (checked in this checkout).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"aee73bfb-0a0a-4f5b-b204-58fe9f882387","repository":"github.com/xianxu/ariadne","issue_id":"000575","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/5d4601b41e81e3d7","reviewed_head":"213f0fa2d58f44975260da50b823ea9f57760f41","slot":"ariadne:2","requested_at":"2026-10-23T19:20:14.177636Z"},"phase":"finished","pid":90290,"started_at":"2026-10-23T19:20:14.366087Z","finished_at":"2026-10-23T19:22:01.134438Z","session_id":"72d0912c-a9a2-4b2f-8b31-f1be1dd2713a","outcome":"returned"}] -->
