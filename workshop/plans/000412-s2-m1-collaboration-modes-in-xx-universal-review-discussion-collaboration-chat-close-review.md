# Boundary Review — ariadne#412 (whole-issue close)

| field | value |
|-------|-------|
| issue | 412 — S2 M1: collaboration modes in xx-universal-review (discussion, collaboration, chat) |
| repo | ariadne |
| issue file | workshop/issues/000412-s2-m1-collaboration-modes-in-xx-universal-review-discussion-collaboration-chat.md |
| boundary | whole-issue close |
| milestone | — |
| window | 35ba24e6cad39344c347af981de7ac8fcc1d1c31..031ea5c641fd81964c0e613178f1bd93dced516c |
| command | sdlc close --issue 412 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-16T16:57:44-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking the chat-mode example fixtures and the tl-lead bullet parser against the new prose spec.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The work matches the issue's Spec and the plan's nine tasks. `collaboration-modes.md` holds the protocol in one place. `round.md` §4 sends each ask to it, it is in `export.json` (and the export test requires exactly one member for it), and all ten example cases pass a checker that was itself mutation-tested. `go test ./pkg/reviewexport/ ./cmd/tl-lead/` passes at the pinned head.

Two Important findings, neither blocking:
- **Reused side docs can link to the wrong discussion.** The side-doc reuse rule (§4 step 1) looks up a doc by `(parent, marker)` alone. When two different questions use the same short text, one gets linked to the other's discussion without any error.
- **The tl-lead parser no longer matches the rule it cites.** `tl-lead --help` now names the skill's chat mode as the rule for Operator instructions. But tl-lead's own bullet parser still defines an item differently from that rule, so the two can disagree about what an item is.

**Strengths**
- `pkg/reviewexport/examples_test.go:450-597`: `checkTree` and `checkCase` are pure functions over an `fs.FS`. `TestCheckCaseCatches` (`:672`) breaks a known-good case one way at a time, so the checker is shown to catch real errors rather than being trusted.
- `TestSideDocNames` pins the number widening (`99`→`100`) and checks that dated and numbered docs count separately. That edge is easy to get wrong.
- The marker is treated as untrusted input: it must be written as a double-quoted YAML string with only `\\` and `\"` escapes (`collaboration-modes.md:167-170`), and `quotedMarker` enforces this on every fixture (ARCH-SECURE).
- The write order in §4 (side docs, then the parent, then the rest) is spelled out, including a refused side-doc write after the fold has already landed ("don't fold again"). The adapter's new paragraph (`adapters/file-edit.md:71-82`) matches it.
- `help.md` now points to the skill instead of repeating the chat rules, and the old G2/G3 topic tests were moved to the new wording.

**Critical findings**
None.

**Important findings**
1. **ARCH-ORDER: a reused side doc can be the wrong one** (`construct/local/universal-review/collaboration-modes.md:188-191`). Step 1 reuses any side doc whose `parent:` and `marker:` match the ask, whether or not something already links to it and whatever section it came from.
   - Failure: last week 🤖[why?] in section A opened `01-why.md` and was replaced by its link. Today 🤖[why?] in section B finds `01-why.md` by `(parent, marker)` and links to it. The new question gets the old, unrelated discussion and no new doc is written.
   - Within a single round, two identical markers are not covered at all: the second either reuses the first doc or breaks the uniqueness rule the checker enforces.
   - Fix: reuse only a doc that no link in the parent points to. For discussion, also require the section named in the doc's backlink to match. State what happens with two identical markers in one round (for example, the second gets a new doc whose slug says which section it came from). Add an example or prose check for it.
2. **ARCH-DRY / ARCH-PURPOSE: tl-lead's bullet parser defines an item differently from chat mode** (`cmd/tl-lead/instructions.go:20,36-55` vs `collaboration-modes.md:112`). `help.md` now says chat mode is "the rule" for Operator instructions, but the parser's definition differs in three ways:
   - It only starts an item at an unindented `N. ` line. Chat mode also accepts `-` and `*` items at the section's shallowest indent.
   - It treats a `→ done:` line as answered. Chat mode has no such form.
   - It attaches any prose before the next bullet to the previous bullet.

   Failure: the operator writes `- check CI` after bullet 3. Chat mode sees a new ask, but `instructions list` doesn't show it, and `link 3` appends bullet 3's link to the `- check CI` line.

   Fix: make `instructionBullets` follow the chat-mode definition (items at the shallowest list indent, with `-` and `*` included), or have the help say the section takes numbered items only. Add a table test pinning whichever is chosen.

**Minor findings**
- Done-when 2 (the live session file runs on `collaboration-mode:`) is deferred to the TL's write in ops. The plan says so, but the close should list it as owed in the landing report rather than as done.
- The checker doesn't check that a fold marks the parent's link `(folded <date>)` (rows 7 and 10). Only the live check covers it.
- `sideDocName` ignores the error from `strconv.Atoi`. A number of 19 or more digits would wrap silently. This only affects test code.
- The new §4 resolve step in `round.md:69` is one over-long line compared with the wrapping around it.

**Test coverage notes**
- Ran `go test ./pkg/reviewexport/ ./cmd/tl-lead/`: both packages pass.
- Every state-table row is covered by an example case or a prose-phrase check (`TestCollaborationModesProse` checks rows 1–18 exist, plus seven rules a static case can't show).
- The new `instructions link` test with a `.side/` target only pins behaviour that already existed (no code changed).
- Nothing tests tl-lead's item parsing against chat mode's definition of an item (Important 2).

**Architecture**
- **ARCH-DRY:** flag (Important 2).
- **ARCH-PURE:** pass. The checker is pure over `fs.FS`, and `Build` is unchanged in shape.
- **ARCH-PURPOSE:** flag. I checked every place that consumes the protocol. `SKILL.md`, `round.md` §4, `export.json` and the help text all take it from `collaboration-modes.md`. The tl-lead Go parser is the one copy of the item definition that doesn't (same finding as Important 2).
- **ARCH-MOCK:** pass. No new external dependency, and Couch is still reached only through the adapter.
- **ARCH-CONSTRAINTS:** pass. Each question adds one file write in a single-turn round.
- **ARCH-SECURE:** pass. Markers are quoted and checked, slugs are limited to `[a-z0-9-]` and 40 characters, and parent paths with `#` are rejected.
- **ARCH-ORDER:** flag (Important 1). Otherwise the retry, refused-write and fold-landed-but-note-missing cases are handled well.
- **ARCH-FUNERAL:** pass. "Never pruned; the operator deletes; S3 adds closing" is stated, and the answer docs are gitignored like the QA files they replace.

**Plan revision recommendations**
- Add a Revisions entry recording that relinking is restricted to unlinked docs in the same section (once fixed), and add a test-list row for "same marker text, different section".
- Add a Revisions entry for task 6: tl-lead's item parser is lined up with chat mode's definition of an item, or the help says the section takes numbered items only.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: reuse-identity-too-coarse
    title: |
      Side-doc reuse by (parent, marker) alone can link a new ask to an unrelated, already-linked discussion
    detail: |
      collaboration-modes.md §4 step 1 reuses any side doc whose parent and marker match, even one already linked from another section, so a second identical marker (e.g. 🤖[why?] elsewhere) silently gets the old discussion; same-round identical markers are unspecified. Restrict reuse to unlinked docs (and same section for discussion) and add a case or check.
  - id: new
    severity: Important
    kind: defect
    family: cited-rule-restated-divergently
    title: |
      tl-lead's instructionBullets defines an item differently from the chat mode tl-lead --help now cites as the rule
    detail: |
      cmd/tl-lead/instructions.go only starts items at unindented `N. ` lines, treats `→ done:` as answered, and absorbs `-`/`*` lines into the previous bullet, while collaboration-modes.md §3 counts -, * and N. items at the shallowest indent; a `- foo` bullet is invisible to list and gets the previous bullet's link. Align the parser (with a table test) or narrow the help.
  - id: new
    severity: Minor
    kind: defect
    family: done-when-deferred
    title: |
      Done-when 2 (session file runs on collaboration-mode) depends on the TL's ops write; record it as owed in the landing report
  - id: new
    severity: Minor
    kind: defect
    family: checker-coverage-gap
    title: |
      Examples checker does not assert the (folded <date>) link mark for fold cases (rows 7 and 10)
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"18bcd2c0-3703-4b84-8353-59b03ee47297","repository":"github.com/xianxu/ariadne","issue_id":"000412","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/a271c0b6d57e2f38","reviewed_head":"031ea5c641fd81964c0e613178f1bd93dced516c","slot":"ariadne:2","requested_at":"2026-10-16T23:55:50.847012Z"},"phase":"finished","pid":21500,"started_at":"2026-10-16T23:55:51.782743Z","finished_at":"2026-10-16T23:57:40.343763Z","session_id":"ca937b3c-788e-48c2-a03c-5828a970022d","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T17:02:48-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 412 — S2 M1: collaboration modes in xx-universal-review (discussion, collaboration, chat) |
| repo | ariadne |
| issue file | workshop/issues/000412-s2-m1-collaboration-modes-in-xx-universal-review-discussion-collaboration-chat.md |
| boundary | whole-issue close |
| milestone | — |
| window | 35ba24e6cad39344c347af981de7ac8fcc1d1c31..255314b4ec7bd08fa70b031805fab03b12f29a06 |
| command | sdlc close --issue 412 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-16T17:02:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Focus first on the last commit, which addresses prior findings.

Checking how the last open prior finding (BR-3) was handled; then I'll write the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All four prior findings are dealt with. One gap remains, and it is cheap to close. BR-1's rule is now in the spec: reuse needs the same parent, marker and section, and the side doc must be unlinked (`collaboration-modes.md` §4 step 1, state-table row 19). There is a reference case for it (`examples/discussion-same-marker`). BR-2 is fixed: tl-lead's parser now follows chat mode's item rule, and a test fails without the fix. BR-4's fold-mark check has its own test that breaks the fixture on purpose and expects the checker to catch it. BR-3 is recorded as owed in the plan. The gap is that the example checker does not enforce row 19. I edited the row-19 case so both asks linked to the old `01-why-per-file.md` and deleted `02-…`, which is exactly the BR-1 defect. `go test ./pkg/reviewexport/ -run TestExamples` still **passed**. The new case shows the right answer, but nothing stops a regression to the defect. Both `go test ./cmd/tl-lead/ ./pkg/reviewexport/` packages are green at head.

1. **Strengths**
   - `cmd/tl-lead/instructions.go:40-79`: the parser now matches the §3 item rule (shallowest indent; sub-items and continuation lines included; prose after a blank line belongs to no item). `TestInstructionBulletsAreChatItems` checks mixed `-`/`*`/`N.` lists, including that a link lands on the right line. The old parser would fail it.
   - `examples_test.go`: identifying reuse by folder, parent, marker and section, plus whether the doc is linked, matches the spec exactly. The "unlinked doc not relinked" check is narrow and correct.
   - Two new cases in `TestCheckCaseCatches` (no backlink; a fold without its mark) each break a valid fixture and expect the checker to catch it, so they test the checker itself.
   - The spec now settles two identical asks in one section in the same round ("each candidate is reused at most once per round").

2. **Critical:** none.

3. **Important**
   - **This is the 2nd finding in family `checker-coverage-gap`.** At `pkg/reviewexport/examples_test.go` (around lines 176-184 and 262-270), the checker catches a missed relink but not the opposite: a doc that is already linked being reused. **Rule to fix the whole class:** the checker should derive every §4 step-1 invariant, not only the ones examples have shown so far. One rule covers row 19 and the same-round case: *in a parent, each side doc is the target of at most one `→ [discussion|answer]` link.* Count the links per target while checking the parent and fail on more than one. Then add a test that links two asks to one doc and expects the checker to fail.

4. **Minor**
   - `instructionBullets` works out `shallow` and the item lines without skipping fenced code blocks, so a `- x` inside a fence becomes a bullet. Indentation also counts spaces only, so a tab-indented item matches nothing.
   - `Answered` matches `→ done:` or `→ [answer](` anywhere in a bullet, including its sub-items. §3 talks about a *trailing* link on the item.
   - **ARCH-FUNERAL:** side docs are "never pruned". That is a deliberate choice and costs one small file per ask, but the spec does not say what bounds them or removes them.

5. **Test coverage**
   - tl-lead is well covered: the new table test plus the existing link/list tests.
   - The example checker has the row-19 gap above. Every other row's invariant I traced is covered, either by a case or by a test that breaks a fixture and expects a failure.

6. **Architecture**
   - **ARCH-DRY:** pass, with a note. The chat-item rule now exists as English in §3 and as Go in tl-lead. A future Go consumer should import the parser rather than copy it again.
   - **ARCH-PURE:** pass. Parsing, linking and checking are pure functions over strings or `fs.FS`; the IO lives in the `SessionSync` seam.
   - **ARCH-PURPOSE:** pass. tl-lead now follows the rule its help cites. Done-when 2 (the live session file) is explicitly owed to the TL.
   - **ARCH-MOCK:** N/A. The change adds no external dependency; tests use `fstest` and the existing fakes.
   - **ARCH-CONSTRAINTS:** N/A. These are small documents parsed in linear time.
   - **ARCH-SECURE:** pass. `linkBullet` refuses targets with whitespace or parentheses, and the checker rejects absolute or `#` parent paths.
   - **ARCH-ORDER:** pass for the spec (the refused-write retry is row 18, write order is in §4) and for the code (`link` is pinned to the hash `list` returned and never retried). The finding above is about the checker, not the protocol.
   - **ARCH-FUNERAL:** see the Minor note.

7. **Plan revisions:** add a "## Revisions" entry saying the checker enforces "at most one link per side doc", which covers row 19 and same-round identical asks.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Spec §4 step 1 and row 19 now require the same section and an unlinked doc; case discussion-same-marker added. Checker enforcement gap raised separately as checker-coverage-gap.
  - id: BR-2
    disposition: addressed
    note: |
      instructionBullets now follows the §3 item rule; TestInstructionBulletsAreChatItems fails on the old N.-only parser.
  - id: BR-3
    disposition: addressed
    note: |
      Plan revisions (plan line ~264) record Done-when 2 as owed by the TL in the landing report.
  - id: BR-4
    disposition: addressed
    note: |
      foldedRE check added, with a test that breaks a fixture's fold mark and expects the checker to fail.
findings:
  - id: new
    severity: Important
    kind: defect
    family: checker-coverage-gap
    title: |
      Example checker does not enforce row 19: a side doc that is already linked can be reused with no failure
    detail: |
      I edited discussion-same-marker/after so both asks link to 01-why-per-file.md and deleted 02, which is the BR-1 defect, and TestExamples still passed. This is the 2nd finding in the family. The rule covering the class: in a parent, each side doc is the target of at most one link. Enforce it in checkTree and add a test that links two asks to one doc and expects a failure.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"13685de5-96b5-4f53-b099-6306401ad4b3","repository":"github.com/xianxu/ariadne","issue_id":"000412","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/4da3b878f24d1b3b","reviewed_head":"255314b4ec7bd08fa70b031805fab03b12f29a06","slot":"ariadne:2","requested_at":"2026-10-17T00:01:22.89563Z"},"phase":"finished","pid":91621,"started_at":"2026-10-17T00:01:23.124489Z","finished_at":"2026-10-17T00:02:47.849521Z","session_id":"8f9eacdf-f967-44c3-9986-b3b01587aee4","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T17:06:57-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 412 — S2 M1: collaboration modes in xx-universal-review (discussion, collaboration, chat) |
| repo | ariadne |
| issue file | workshop/issues/000412-s2-m1-collaboration-modes-in-xx-universal-review-discussion-collaboration-chat.md |
| boundary | whole-issue close |
| milestone | — |
| window | 35ba24e6cad39344c347af981de7ac8fcc1d1c31..e123c3ca7429ac6e50381f13a646c024a08ea631 |
| command | sdlc close --issue 412 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-16T17:06:57-07:00 |
| verdict | REWORK |

## Review

Focus first on BR-5: the last commit claims the checker enforces every reuse invariant.

Now revert the fix to confirm the regression test goes red.

BR-5 is addressed. Now the issue, plan, and core spec.

Item 3 in the chat example has its link on the first line and a follow-up sub-item after it. Testing how the new "link ending the last line" rule handles that.

```verdict
verdict: REWORK
confidence: high
```

BR-5 is fixed. I re-ran the reviewer's regression: in `discussion-same-marker/after`, I pointed both asks at `01-why-per-file.md` and deleted `02`. `TestExamples` now fails with "two links point to design.side/01-why-per-file.md". I then disabled the at-most-one-link check (`targets[t] == 2` changed to `== 99`), and the mutation case "a linked doc reused" in `TestCheckCaseCatches` went red. Both changes were restored afterwards. The fix covers the whole class, not one instance: no side doc may be the target of two links (`examples_test.go:147`), and every link must come from the document its `parent:` names (`examples_test.go:150`).

The same commit (e123c3ca) also brought in a new correctness regression. The "parser edges" Minor changed what makes a chat item linked: both the skill and tl-lead now count only an answer link that ends the item's **last** line. When an operator adds a follow-up sub-item under a linked item (state-table row 14), the link is no longer on the last line, so the item reads as unlinked. That contradicts the committed `chat` example and row 14. In tl-lead it means `list` shows the item as open again, and `link` adds a second answer link to the operator's sub-item. This blocks SHIP. The fix is cheap.

**Strengths**
- The BR-5 fix states the invariant rather than patching the one case, and it has both a mutation test and an end-to-end check against the fixtures (`pkg/reviewexport/examples_test.go:136-153`, `:403-409`).
- The checker is a pure function over `fs.FS`, and `TestCheckCaseCatches` breaks a known-good case one way at a time. That is good evidence that each rule is actually enforced (ARCH-PURE).
- The export gains exactly one new member, with a rank and doc comment, and `examples/` is explicitly left out (`pkg/reviewexport/export.go:55,151`).
- `tl-lead --help` now points to the skill's chat mode instead of restating its rules, which keeps one source for them (ARCH-DRY, `cmd/tl-lead/help.md:171-199`).

**Critical**
1. **A linked chat item that gets a follow-up sub-item now reads as unlinked.**
   - **Where:** the rule is defined in `construct/local/universal-review/collaboration-modes.md:115-116` and implemented in `cmd/tl-lead/instructions.go:21,96`.
   - **Repro:** run `instructionBullets` on `3. how big is the backlog? → [answer](status.side/2026-10-15-01-backlog.md)\n   - and by repo?`. This is the exact shape of item 3 in `examples/chat`. It returns `answered=false`, and `linkBullet(sec,1,…)` writes `   - and by repo? → [answer](tl-status-op.side/…)`, a second link on the operator's line.
   - **Why it's wrong:** by the skill's own wording, the committed example's item 3 is a "new ask", yet the example and row 14 call it a follow-up that keeps its link. Before e123c3ca, `answeredRE` matched anywhere in the item, so this worked.
   - **Fix:** define "linked" so a link stays valid when lines are added after it. For example: "a ` → [answer](…)` ending a line of the item at the item's own indent, outside fences", or simply the item's first answer link. Update §3 and `answeredRE` together.
   - **Regression test:** add a tl-lead test built from the `chat` example's Operator-instructions section (read from `examples/chat/after/status.md`). That ties tl-lead's parser to the skill's reference fixture, so the two can't drift apart again.

**Important**
- None beyond the Critical above.

**Minor**
- `instructions_test.go:85` pins `2. r\n   - see → [answer](x.md)\n   more text` as unanswered. Re-check that expectation once "linked" is redefined.

**Test coverage notes**
- `go test ./cmd/tl-lead/ ./pkg/reviewexport/` passes at head. No test exercises row 14's item shape (a link followed by an added sub-item) against the item parser, and that is exactly why the regression went unnoticed.

**Architecture**
- **ARCH-DRY: pass.** The help cites the skill, and one item rule is shared by the skill and the parser. They do agree, but they agree on a definition that contradicts the skill's own example.
- **ARCH-PURE: pass.** The checker and `instructionBullets` are both pure.
- **ARCH-PURPOSE: flag (the Critical above).** Done-when 2 says the session file runs on chat mode "with no change in what the operator sees". Re-listing an answered item and double-linking it is such a change.
- **ARCH-MOCK: pass.** No new external dependency; Couch is reached only through the adapter's existing seam.
- **ARCH-CONSTRAINTS: pass.** Each new side doc costs one file write.
- **ARCH-SECURE: pass.** The marker must be a double-quoted YAML string, the link target is validated, and a parent path may not carry an anchor.
- **ARCH-ORDER: pass.** Retry and relink states are covered by `discussion-retry`, `chat-retry` and `discussion-same-marker`.
- **ARCH-FUNERAL: pass.** The skill now states the side docs' lifetime: no round prunes one, the operator deletes it, and a later slice adds closing.

**Docs**
- The atlas is updated (`atlas/workflow/universal-review.md`, `tl-skills.md`, the indexes). `README.md` never documented this skill's frontmatter, so no README gap.

**Plan revisions**
- Add to the round-2 Revisions entry: "an item is linked by its answer link wherever later lines put it, not only on its last line (row 14)", along with the regression test built from the chat example.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      Reviewer's regression (both asks -> 01, 02 deleted) now fails TestExamples; disabling the targets check turns "a linked doc reused" red.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: spec-contradicts-own-example
    title: |
      "Linked = answer link ending the item's last line" makes a row-14 follow-up item read as unlinked; tl-lead double-links it
    detail: |
      e123c3ca changed collaboration-modes.md:115-116 and instructions.go answeredRE (line 21, used at 96) so only a link on the last line counts. examples/chat item 3 ("3. ... → [answer](...)" then "   - and by repo?") is then a new ask by the spec's own rule, which contradicts row 14 and the example. Probe: instructionBullets gives answered=false, and linkBullet appends a second → [answer] to the operator's sub-item. Fix: redefine linked so lines added after the link don't unlink it (in both the skill and the regex), and add a tl-lead test that parses the chat example's Operator-instructions section.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2aaa082b-6e39-4a43-967a-2e206e39407b","repository":"github.com/xianxu/ariadne","issue_id":"000412","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/1676c00243d2f2c8","reviewed_head":"e123c3ca7429ac6e50381f13a646c024a08ea631","slot":"ariadne:2","requested_at":"2026-10-17T00:04:39.693039Z"},"phase":"finished","pid":26230,"started_at":"2026-10-17T00:04:39.893494Z","finished_at":"2026-10-17T00:06:56.755001Z","session_id":"e99fcf6d-5336-48db-94dd-b43f505f70d6","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T17:11:58-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 412 — S2 M1: collaboration modes in xx-universal-review (discussion, collaboration, chat) |
| repo | ariadne |
| issue file | workshop/issues/000412-s2-m1-collaboration-modes-in-xx-universal-review-discussion-collaboration-chat.md |
| boundary | whole-issue close |
| milestone | — |
| window | 35ba24e6cad39344c347af981de7ac8fcc1d1c31..6fec11e686a4fa6b3ac37980884247ea46258ec8 |
| command | sdlc close --issue 412 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-16T17:11:58-07:00 |
| verdict | SHIP |

## Review

I've confirmed the BR-6 fix: undoing it makes two tests fail. Next I'm checking for stale comments and leftover old "closing note" uses in the examples.

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** The BR-6 fix is real, it is applied in both places that define the rule, and a test catches it if it is removed. The skill (`collaboration-modes.md:124-126`) and the parser (`cmd/tl-lead/instructions.go:94-98`) now agree: a `→ [answer](…)` at the end of any of an item's lines outside fenced code makes the item linked. A follow-up added later (row 14) leaves it linked. I undid the parser change in the review checkout and two tests failed. `TestInstructionBulletsReadTheChatExample` reported items 1–3 open instead of 1–2, and item 3 open in `after/`, which is the double-link BR-6 described. With the fix back in place, `go test ./cmd/tl-lead/ ./pkg/reviewexport/` passes. The same commit also adds the "Decided (operator)" one-line record for decisions. That change is in scope: the plan's `## Revisions` lists it (plan:283-290), and it comes with row 20, the `discussion-decision` example, and checker updates. Nothing blocks. The only finding is one stale test comment.

1. **Strengths**
   - `cmd/tl-lead/instructions_test.go:222-245` makes tl-lead parse the skill's own chat example (both `before/` and `after/`). This links the parser to the reference example, so the two can't drift apart. That is the class-level fix for the `spec-contradicts-own-example` and `cited-rule-restated-divergently` families, not just a patch for this one case.
   - The edge-case test now covers a link inside a fence that must not count as an answer (`instructions_test.go:88-90`, `bs[0].Answered == false`). The fence exclusion is pinned, not just implemented.
   - A sweep for wording left over from the old rule ("last line" as the linked test, "Folded into the parent", "closing note") finds no stale copies in the skill, atlas, checker or examples. The remaining "last line" mentions only say where a link is *appended*, which is correct.
   - The atlas (`atlas/workflow/universal-review.md`) was updated in the same commit for the Decided line and for the example count going to twelve.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `cmd/tl-lead/instructions_test.go:84-85`: the old comment "only a link ending the bullet's last line answers it" is still there, directly above the new comment that says the opposite. Delete the two old lines.

5. **Test coverage notes:** BR-6 now has a regression test in two forms: a synthetic edge case and the real chat example. Reverting the fix turns both red. The fold-mark check in `examples_test.go` now looks for the new Decided-line wording, and `TestCheckCaseCatches` still includes the case that should trigger it.

6. **Architecture**
   - **ARCH-DRY:** pass. There is one linked rule, and the example test holds the parser to it.
   - **ARCH-PURE:** pass. `instructionBullets` is a pure function, tested without IO apart from reading the example file.
   - **ARCH-PURPOSE:** pass. The fix covers both the skill and the parser, and the class is enforced through the example.
   - **ARCH-MOCK:** pass. No new external calls; writes still go through the existing adapter and the faked lock.
   - **ARCH-CONSTRAINTS:** pass. Parsing is linear in the section's line count, and sections are small.
   - **ARCH-SECURE:** pass. Operator-edited markdown is parsed into typed bullets, and an already-linked bullet is refused (`errAlreadyLinked`), not double-written.
   - **ARCH-ORDER:** pass. The retry path (row 18, plus step 4's "folded today, approval still present") is written as explicit state checks.
   - **ARCH-FUNERAL:** pass. Side docs are never pruned by design, and orphans are named in the reply.

7. **Plan revision recommendations:** none. The Revisions section already records the Decided-line change.

```findings
dispose:
  - id: BR-6
    disposition: addressed
    note: |
      instructions.go:94-98 + collaboration-modes.md:124-126 now link on any non-fenced line; reverting the parser makes TestInstructionBulletsAreChatItems and TestInstructionBulletsReadTheChatExample fail (open 1-3 / after item 3 open); restored, suite green.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: cited-rule-restated-divergently
    title: |
      Stale test comment still states the old last-line rule
    detail: |
      cmd/tl-lead/instructions_test.go:84-85 still says only a link ending the bullet's last line answers it, directly above the new comment saying any line does. Delete the stale two lines.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b5b9ed1f-8e13-41cd-a135-19b40c710dfd","repository":"github.com/xianxu/ariadne","issue_id":"000412","recipe":"milestone-review","round":4,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/ef3323cd3bb5078b","reviewed_head":"6fec11e686a4fa6b3ac37980884247ea46258ec8","slot":"ariadne:2","requested_at":"2026-10-17T00:10:19.464592Z"},"phase":"finished","pid":4003,"started_at":"2026-10-17T00:10:19.761703Z","finished_at":"2026-10-17T00:11:30.2375Z","session_id":"6da505a9-d2e3-43ce-a7aa-58fc98f299b4","outcome":"returned"}] -->
