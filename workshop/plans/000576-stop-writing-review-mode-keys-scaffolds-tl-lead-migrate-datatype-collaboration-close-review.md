# Boundary Review — ariadne#576 (whole-issue close)

| field | value |
|-------|-------|
| issue | 576 — Stop writing review mode keys: scaffolds, tl-lead migrate, datatype collaboration-mode rows |
| repo | ariadne |
| issue file | workshop/issues/000576-stop-writing-review-mode-keys-scaffolds-tl-lead-migrate-datatype-collaboration.md |
| boundary | whole-issue close |
| milestone | — |
| window | c39d1899c04e1176d5bdce47f9f06848dd17ce2c..4096b2e43bf00d217e1c4afdfb8384eed34bd4b1 |
| command | sdlc close --issue 576 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T14:52:38-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still reviewing: I've read the tl-lead and sidedoc diffs. Next I'm checking the guard test and the vocab changes, then running the test suites.

Running the affected test suites now, then writing up the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The main goal of #576 is met: nothing in ariadne writes a review mode key any more. The three sdlc writers (`scaffold.go`, `handoff.go`, `project/scaffold.go`) are gone, along with the CUE `scaffold.collaboration_mode` field and its JSON, the vocab accessors, all 11 datatype rows, the `type.md` requirement, the writing-plans header key and the `answerFrontmatter` writer. A grep finds no remaining Go reader or writer of the retired keys outside legacy readers and test fixtures. `answerFrontmatter` is now lossless. The touched packages pass: `pkg/sidedoc`, `pkg/vocab`, `pkg/reviewexport`, `cmd/datatype`, `cmd/sdlc/internal/issue` and `cmd/sdlc/internal/project`. The tl-lead, weave, layergraph, workspace and testfix failures I hit are all `git init refused … outside the reviewer's directories (sdlc #339)`, which is this sandbox, not the code. One thing keeps this from SHIP: the plan promised `link` would reuse the item rules `list` used (an ARCH-ORDER guard), and that guard is missing. Because the section hash doesn't cover the `@@chat@@` tag line or the frontmatter, `link` can attach an answer to the wrong bullet. Two promised tests are also missing.

1. **Strengths**
   - `sidedoc.ChatItems` and `CodeLines` (`pkg/sidedoc/chat.go:779`, `pkg/sidedoc/literals.go:1180`) replace three separate item parsers (tl-lead's, the checker's `numberedItem`, and pair's copy) with one shared set of rules (ARCH-DRY). Code detection now uses CommonMark via goldmark instead of a fence toggle, and the three fence-mismatch cases are pinned in `TestChatItemsCodeFollowsCommonMark`.
   - The linear-time claims are tested, not just stated: tests compare work counts at N and 2N for `chatSections`, `chatItems` and `codeLines`.
   - `TestChatSectionsLaminar` checks `Innermost` and nesting against an independent reading of 500 random documents rather than restating the stack implementation.
   - `answerFrontmatter` (`cmd/tl-lead/migrate.go:445`) keeps every line except type/parent/marker verbatim, and the golden diff only removes the two lines it used to add (ARCH-EVOLVE pass).
   - The `TestNoSpecWritesModeKeys` guard was widened to cover every creator's instructions, and its regex now matches the key names in shape-table rows too, not just `key:`.

2. **Critical findings:** none.

3. **Important findings**
   - **`link` can attach an answer to the wrong bullet** (`cmd/tl-lead/instructions.go:127`, ARCH-ORDER).
     - The plan says `link` reuses the item rules `list` printed, but `linkInstruction` re-reads the file and recomputes `legacyGrammar(doc)`.
     - The hash check only covers the section body. The `@@chat@@` tag sits above the heading, and a legacy map sits in the frontmatter, so adding either between `list` and `link` changes how bullets are counted without changing the hash.
     - Example: a section `- note\n1. a\n2. b`. Under the legacy rules `list` shows position 2 = `1. a`. After the operator adds `@@chat@@` (the migration help.md now prescribes), `link 2` writes to `2. b` instead.
     - Fix sketch: have `list` print the rule set it used (e.g. a `grammar: legacy|numbered` line next to `hash:`) and make `link` take it as a flag, or refuse when it changed. Add a sequence test using the fake that runs list, then adds the tag, then links.
   - **Promised fake-backed list/link tests are missing** (`cmd/tl-lead/instructions_test.go:306`). Task 1 promised tests for a heading-comment file and a `@@chat@@` file that run through the session-sync fake. `TestLegacyGrammar` only tests the pure helper. Nothing exercises `sectionBullets`, the `Peek` call in `link`, or the `list` output on a tagged file.
   - **The card-publish round-trip test is missing** (`cmd/sdlc/internal/issue/nomodekeys_test.go:42`). Task 2 (marked done) and Done-when promise a legacy `collaboration-mode:` surviving `SetField` *and* a card publish (SplitCard/compose). Only `SetField` is tested.

4. **Minor findings**
   - `legacyGrammar` looks for an exact `## Operator instructions` line and only checks sections whose heading is that line. If the section is nested under a tagged parent, or the heading carries a trailing annotation, it falls back to the legacy rules instead of using `Innermost`.
   - The plan says `legacy` is "false when no legacy form names it", but the code treats an undeclared file as legacy. The code's choice is the sensible backward-compatible one; the plan text should be updated.
   - The plan says it reads frontmatter plus the section; the code uses `Peek` on the whole file. This deviation is logged but the plan wasn't revised.
   - The `TestNoSpecWritesModeKeys` filter skips any line containing the word "legacy". That is a broad hole: a line telling a creator to write a key would pass if it also says "legacy".
   - `codeLines`: an empty fenced block (no content lines) leaves its fence lines unmarked. It is harmless for finding items today.
   - `nomodekeys_test.go:31` throws away the `SetCardHandoff` error.
   - `project/scaffold_test.go` ends with a trailing blank line (gofmt).
   - The `instructions_test.go` import isn't grouped (the sidedoc import sits in the stdlib block).

5. **Test coverage notes**
   - The guard's mutation check (which caught missing shape-table rows) and the work-count tests are good evidence.
   - The gaps are the three Important items above. Most important: no test can observe the list→link ordering, so a green run only proves the single path the author happened to take.

6. **Architectural notes for upcoming work**
   - ARCH-DRY: pass. One set of item rules now serves the checker, tl-lead and pair.
   - ARCH-PURE: pass. `ChatItems`, `legacyGrammar` and `answerFrontmatter` are pure; IO stays in `Peek`/`ReadSection`.
   - ARCH-PURPOSE: pass. Every writer listed in the surface table is gone, the creators' instructions are guarded, and `vocab.CollaborationModes` is reduced to the legacy set.
   - ARCH-MOCK: pass, with a gap. The existing session-sync fake was the right seam but isn't used for the new flow.
   - ARCH-CONSTRAINTS: pass. The scans are linear and counted; there is one extra `Peek` per command.
   - ARCH-SECURE: pass. Unreadable frontmatter only produces a notice and the legacy rules, never a crash.
   - ARCH-ORDER: **flag**. This is the list/link finding above.
   - ARCH-FUNERAL: N/A. The change creates nothing durable; it deletes code and rows.
   - ARCH-EVOLVE: pass. The key removal is additive, legacy lines survive `SetField` and migrate, and `pensive.cue`'s optional enum is kept.

7. **Plan revision recommendations**
   - "## Revisions: `link` reuses the item rules `list` printed (a printed `grammar:` line passed back with `--hash`). This replaces recomputing them from a fresh `Peek`."
   - "## Revisions: tl-lead decides the item rules from `ChatSections(Peek(file))`, not frontmatter plus section. A file with no chat declaration keeps the legacy rules (`-`, `*`, `N.`)."
   - "## Revisions: workingdoc.md's own frontmatter key was removed too, not kept as the surface table states."

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: list-link-grammar-pinning
    title: |
      tl-lead link recomputes the item grammar from a fresh Peek instead of reusing the grammar list printed, so a tag or frontmatter-map change outside the hashed section re-numbers bullets and link writes to the wrong one
    detail: |
      cmd/tl-lead/instructions.go:127 calls legacyGrammar(ss.Peek(file)) at link time. The section hash covers only the section body; the @@chat@@ tag line (above the heading) and a frontmatter collaboration-mode map do not affect it. Example: section "- note\n1. a\n2. b". list (legacy) shows position 2 = "1. a". The operator adds @@chat@@ as help.md now prescribes, and link 2 under the numbered grammar appends to "2. b". The plan's ARCH-ORDER note promised that link reuses list's grammar. Fix: have list print the grammar and link take it as a flag (or refuse on mismatch), plus a fake-backed list -> add tag -> link sequence test.
  - id: new
    severity: Important
    kind: defect
    family: promised-test-not-delivered
    title: |
      Task 1's fake-backed list/link tests for a heading-comment session file and a @@chat@@ session file are missing
    detail: |
      Only the pure helper is tested (TestLegacyGrammar, instructions_test.go:306). No test drives sectionBullets, link's Peek path, or the list output on a tagged file through the session-sync fake, so the new IO wiring has no regression coverage.
  - id: new
    severity: Important
    kind: defect
    family: promised-test-not-delivered
    title: |
      The legacy round-trip test covers SetField only; the card publish named in Task 2 and Done-when is not tested
    detail: |
      nomodekeys_test.go TestLegacyModeKeySurvivesSetField checks SetField alone. Task 2 (checked) promises that a legacy collaboration-mode line survives "SetField and a card publish" byte for byte. Add a SplitCard/compose round trip on an issue that carries collaboration-mode: collaboration.
  - id: new
    severity: Minor
    kind: defect
    family: plan-code-drift
    title: |
      The plan still says legacy=false for an undeclared section, uses frontmatter+section, and keeps workingdoc.md's key; the code does otherwise
    detail: |
      legacyGrammar returns true for an undeclared file and reads Peek(file); workingdoc.md's own key was removed. Add ## Revisions entries.
  - id: new
    severity: Minor
    kind: defect
    family: guard-skip-too-broad
    title: |
      TestNoSpecWritesModeKeys skips any creator line containing the word "legacy", a broad hole in the guard
    detail: |
      A line instructing a creator to write a key would pass if it also mentions "legacy". Consider skipping by explicit allow-list lines or by a marker phrase instead.
  - id: new
    severity: Minor
    kind: defect
    family: heading-lookup-exact-match
    title: |
      legacyGrammar finds the section by an exact "## Operator instructions" line and ignores enclosing sections
    detail: |
      A tagged parent heading, or a trailing heading annotation, falls back to the legacy grammar; ChatScan.Innermost would resolve it.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"393a7478-9e0e-474a-bb19-c380e18fe64a","repository":"github.com/xianxu/ariadne","issue_id":"000576","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/e94d3ba9daee6b13","reviewed_head":"4096b2e43bf00d217e1c4afdfb8384eed34bd4b1","slot":"ariadne:2","requested_at":"2026-10-23T21:48:36.873152Z"},"phase":"finished","pid":11238,"started_at":"2026-10-23T21:48:36.983063Z","finished_at":"2026-10-23T21:52:37.732996Z","session_id":"96bfb0ed-07f1-4de9-b80c-2fb9ee8570c3","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T14:59:10-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 576 — Stop writing review mode keys: scaffolds, tl-lead migrate, datatype collaboration-mode rows |
| repo | ariadne |
| issue file | workshop/issues/000576-stop-writing-review-mode-keys-scaffolds-tl-lead-migrate-datatype-collaboration.md |
| boundary | whole-issue close |
| milestone | — |
| window | c39d1899c04e1176d5bdce47f9f06848dd17ce2c..72459336c6d684be44f6a4cb4de15e09e71886a8 |
| command | sdlc close --issue 576 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-23T14:59:10-07:00 |
| verdict | SHIP |

## Review

Focus on the last commit (fixes), plus plan.

Status: the relevant tests pass, and reverting each fix turns its regression test red. Here is the review.

```verdict
verdict: SHIP
confidence: high
```

All six findings from round 1 are dealt with, and the four that change behaviour have tests that fail without the fix. I checked this in the disposable checkout:

- **BR-1:** I disabled the grammar check in `linkInstruction`. `TestInstructionsGrammarThroughTheFake/a_tag_added_after_list` then fails with "exit 0, writes 1; want exit 6 and no write".
- **BR-6:** I removed the `Innermost` fallback. `TestLegacyGrammar` "an enclosing tagged section governs" then fails.

The source is restored; `git status` is clean. The affected packages pass: `sidedoc`, `vocab`, `reviewexport`, `sdlc/internal/issue`, `datatype`, and the instruction, legacy-grammar, migrate and help tests in `tl-lead`. The other `cmd/tl-lead` failures (team, fleet, operating, session-move and a few others) come from the review sandbox refusing `git init` ("outside the reviewer's directories (sdlc #339)"). They are not caused by this diff. Nothing blocks SHIP; one new Minor doc nit remains.

1. **Strengths**
   - `cmd/tl-lead/instructions.go:84-93`: the `hash:grammar` token carries the grammar from `list` to `link`, so `link` refuses instead of guessing. A bare hash from before #576 means legacy, which keeps older callers working. A tagged file then refuses them, and the test covers that.
   - `cmd/tl-lead/instructions_test.go:332-372`: the tests run list → change → link through the session-sync fake and assert both the exit code and that nothing was written.
   - `pkg/reviewexport/examples_test.go:816,851-856`: the guard now excuses only the "legacy `key`" span, not any line containing "legacy". The Log records that an instruction phrased as "on a legacy document, write `presentation-mode`" now fails the guard.
   - `cmd/sdlc/internal/issue/nomodekeys_test.go:50-71`: checks that a card publish keeps the legacy keys in the details byte for byte and leaves them off the card.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `cmd/tl-lead/help.md:433-436` still says `list` prints "the section's hash", but `list` now prints the token `HASH:GRAMMAR`. Callers who copy H verbatim still work. The help should say the token also pins the grammar, and that exit 6 can mean "the grammar changed".

5. **Test coverage notes:** each behaviour fix was checked by reverting it (see above). BR-3's test pins existing `SplitCard` behaviour, so it guards against regressions rather than testing a new fix; that is what Task 2 promised.

6. **Architecture**
   - **Pass:** ARCH-DRY (link reuses `ChatItems`, `ChatSections` and `Innermost`), ARCH-PURE (the grammar helpers are pure and IO stays in `sectionBullets` and `linkInstruction`), ARCH-PURPOSE (the guard covers every creator's instructions), ARCH-MOCK (tested through the fake), ARCH-CONSTRAINTS (linear scans), ARCH-SECURE (an unknown grammar value refuses instead of being trusted).
   - **ARCH-ORDER — pass:** the race between list, an outside edit and link now goes through the explicit token check, and a test covers it.
   - **ARCH-FUNERAL — pass:** nothing durable is added.
   - **ARCH-EVOLVE — pass:** legacy keys survive SetField and card publish byte for byte. The token format is additive (a bare hash is still accepted).

7. **Plan revisions:** none needed. The Revisions entries dated 2026-10-23 match the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The token is now hash:grammar and link refuses with exit 6 on a change. Disabling the check makes TestInstructionsGrammarThroughTheFake fail (exit 0, writes 1).
  - id: BR-2
    disposition: addressed
    note: |
      TestInstructionsGrammarThroughTheFake runs list and link through the fake for a tagged session file and a heading-comment one.
  - id: BR-3
    disposition: addressed
    note: |
      TestLegacyModeKeysSurviveCardPublish checks that the details frontmatter keeps a byte prefix and all three legacy keys, and that the card has none.
  - id: BR-4
    disposition: addressed
    note: |
      Plan Revisions entries record Peek(file), legacy=true for an undeclared section, and workingdoc.md's key removal; each matches instructions.go:39-57.
  - id: BR-5
    disposition: addressed
    note: |
      The guard now excuses only the "legacy `key`" span (examples_test.go:816,855); the rest of the line is still checked.
  - id: BR-6
    disposition: addressed
    note: |
      An enclosing chat section now governs through Innermost, and a revert check confirms the test goes red. A trailing annotation on the heading itself still uses the legacy grammar, a known low-value gap.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: plan-code-drift
    title: |
      help.md still says list prints "the section's hash"; it now prints a HASH:GRAMMAR token
    detail: |
      cmd/tl-lead/help.md:433-436 should say the token also pins the item grammar, and that exit 6 can mean a tag or legacy declaration changed. This is the 2nd finding in family plan-code-drift. The rule: when a fix changes a printed or accepted token, the same commit updates every prose surface that describes it (help.md, the plan, atlas). Copying H verbatim still works, so this is cosmetic.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"860ee805-a595-4e6b-8b5c-c18e274eed06","repository":"github.com/xianxu/ariadne","issue_id":"000576","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/b6b051cb513eeb35","reviewed_head":"72459336c6d684be44f6a4cb4de15e09e71886a8","slot":"ariadne:2","requested_at":"2026-10-23T21:57:44.642944Z"},"phase":"finished","pid":2349,"started_at":"2026-10-23T21:57:44.78014Z","finished_at":"2026-10-23T21:59:09.984941Z","session_id":"78202f8f-3a2f-4dfe-b161-5d283a2cbda5","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T15:03:14-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 576 — Stop writing review mode keys: scaffolds, tl-lead migrate, datatype collaboration-mode rows |
| repo | ariadne |
| issue file | workshop/issues/000576-stop-writing-review-mode-keys-scaffolds-tl-lead-migrate-datatype-collaboration.md |
| boundary | whole-issue close |
| milestone | — |
| window | c39d1899c04e1176d5bdce47f9f06848dd17ce2c..2b4928789db3a7e43f455453a4b811ff871b233c |
| command | sdlc close --issue 576 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T15:03:14-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Check BR-7 first: the last commit's changes.

Tests for the touched logic pass; the failures that remain are the sandbox refusing `git init`. Next I'm checking how `legacyGrammar` matches the heading, and the sidedoc chat grammar.

Nearly done — checking whether any remaining non-test file still writes a mode key.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The diff does what #576 asks. Nothing writes a review mode key any more: not the sdlc scaffolds, the vocab, the datatype prototypes, the plan header or tl-lead migrate. `datatype lint` is gone, along with its render warning and every doc reference to it. tl-lead now reads `## Operator instructions` with the shared `sidedoc.ChatItems` grammar, choosing numbered or legacy items from a `Peek` of the file. `link` checks the grammar it gets back from the `HASH:GRAMMAR` token that `list` printed. BR-7 is fixed: help.md, the `list` one-liner and the atlas now describe the token and the new reason for exit 6. Nothing blocks. Two new Minor findings remain, both repeats of known families, so each comes with the rule that covers it rather than a one-line fix.

**Test run:** the tests for the touched logic pass (instructions, grammar, migrate golden/resume, help, sidedoc, vocab, reviewexport, issue, project, datatype). The tl-lead failures I saw are all the review sandbox refusing `git init` ("outside the reviewer's directories (sdlc #339)"). None of them is a test this diff touches.

### 1. Strengths
- `cmd/tl-lead/instructions.go:83-88,153-163`: the `HASH:GRAMMAR` token closes the gap between `list` and `link` without a new state store. A bare token from before #576 still parses as legacy, so older output keeps working.
- `TestInstructionsGrammarThroughTheFake` (`cmd/tl-lead/instructions_test.go:332`) covers the key ordering case through the session-sync fake. When a tag is added outside the hashed section after `list`, `link` exits 6 and writes nothing (`f.writes == 0`).
- `cmd/sdlc/internal/issue/nomodekeys_test.go`: covers both creation paths (`Render` and `DetailsFromCard`) and checks byte-for-byte that legacy keys survive both `SetField` and the `SplitCard` card publish (ARCH-EVOLVE).
- `pkg/reviewexport/examples_test.go:825`: the guard now covers every creator's instructions. The exemption is narrowed to the "legacy \`key\`" span instead of the whole line, which fixes BR-5's hole.

### 2. Critical findings
None.

### 3. Important findings
None.

### 4. Minor findings
- **`cmd/tl-lead/instructions.go:216,226,123`** (plan-code-drift, 3rd finding in the family): three surfaces still describe the old bare hash.
  - The cobra `Use` still says `--hash H`.
  - The flag help says "the section hash `list` printed".
  - The comment on `renderOpen` says "then the section hash".

  The rule: every description of the token should come from one source, a single `const` used by `Use`, the flag help and `Short`. Then sweep with `grep -n -- "--hash H\|section's hash\|section hash" cmd/tl-lead`.
- **`cmd/tl-lead/instructions.go:41-43`** (heading-lookup-exact-match, 2nd finding in the family): `legacyGrammar` finds the heading by comparing raw lines. I checked this with a throwaway test. Each of these `@@chat@@`-tagged files returned `legacy=true` when it should be numbered:
  - an earlier copy of the heading inside a fenced block;
  - a closing `##` after the heading text;
  - a heading indented by one space.

  `list` and `link` make the same choice, so nothing gets linked to the wrong bullet; the only effect is the wrong grammar. The rule: find a heading through the markdown parse, never a raw line compare. `ChatSections` already builds the heading set from the AST, so expose a lookup there (for example `ChatScan.HeadingLine(name)`) and have `legacyGrammar` use it.

### 5. Test coverage notes
The behaviour this diff could break (a grammar change outside the hash, legacy keys dropped on rewrite, a creator emitting a key) is covered by tests that would fail. Not covered: the heading-match variants above.

### 6. Architectural notes
- **ARCH-DRY: pass.** `ChatItems` is the one grammar, shared by tl-lead and the checker. The token text in `Use` and the flag help is duplicated by hand; that is the Minor finding above.
- **ARCH-PURE: pass.** The parser, the linker, `legacyGrammar` and the token are pure; the IO goes through `deps.ss`.
- **ARCH-PURPOSE: pass.** I checked every writer of the keys: no non-test Go file still writes one, and the remaining mentions are comments or `collaboration-modes.md` paths.
- **ARCH-MOCK: pass.** The tests go through the session-sync fake.
- **ARCH-CONSTRAINTS: pass.** There is one extra `Peek` per `list`/`link`, a local file read.
- **ARCH-SECURE: pass.** A malformed grammar in the token refuses with exit 6.
- **ARCH-ORDER: pass.** The hash pins the section, the grammar check pins the rest of the file, and the write is a CAS. A grammar edit between the `Peek` and the write is accepted as advisory.
- **ARCH-FUNERAL: pass.** Nothing new is persisted; the token lives in stdout only.
- **ARCH-EVOLVE: pass.** Legacy keys are kept on `SetField` and card publish, and migrate keeps them.

### 7. Plan revision recommendations
None required. Plan line 55 still describes the original frontmatter-plus-section design, but the 10-23 Revisions entry supersedes it and BR-4 is already disposed.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      help.md:430-440 now describes the HASH:GRAMMAR token and exit 6 on a grammar change; atlas tl-skills.md:146 and the list Short match. The cobra Use and flag text are raised separately.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: plan-code-drift
    title: |
      link's cobra Use, --hash flag help and renderOpen comment still describe a bare section hash
    detail: |
      This is the 3rd plan-code-drift finding. Sites: instructions.go:216 Use says --hash H, :226 flag help says the section hash list printed, :123 renderOpen comment says the section hash. Rule: every description of a printed or accepted token should come from one source, a const used by Use, the flag help and Short. Then sweep with grep for --hash H and section hash in cmd/tl-lead.
  - id: new
    severity: Minor
    kind: defect
    family: heading-lookup-exact-match
    title: |
      legacyGrammar finds the Operator instructions heading by a raw line compare, not the markdown parse
    detail: |
      This is the 2nd heading-lookup-exact-match finding. A throwaway test shows three @@chat@@-tagged files returning legacy=true when they should be numbered: an earlier copy of the heading inside a fenced block, a closing ## after the heading text, and a heading indented by one space. list and link make the same choice, so no bullet is mislinked; the only effect is the wrong grammar. Rule: find a heading through the parsed AST that ChatSections already builds (for example a ChatScan heading lookup), never a raw line compare.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"38bc9677-ef8b-4c9d-a736-3e15bf628e00","repository":"github.com/xianxu/ariadne","issue_id":"000576","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/630ef39e4c25b48d","reviewed_head":"2b4928789db3a7e43f455453a4b811ff871b233c","slot":"ariadne:2","requested_at":"2026-10-23T22:01:10.740419Z"},"phase":"finished","pid":13711,"started_at":"2026-10-23T22:01:10.847412Z","finished_at":"2026-10-23T22:03:14.439061Z","session_id":"359c08fd-e64b-4456-8da6-6e8f96f06957","outcome":"returned"}] -->
