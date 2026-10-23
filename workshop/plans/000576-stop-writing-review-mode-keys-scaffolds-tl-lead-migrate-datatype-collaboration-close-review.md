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
