# Boundary Review — ariadne#424 (milestone M1)

| field | value |
|-------|-------|
| issue | 424 — side docs S3: the discussions index, link removal, fold-in from the parent, and moves that carry the .side folder (skill, tl-lead, sdlc archiving) |
| repo | ariadne |
| issue file | workshop/issues/000424-side-docs-s3-the-discussions-index-link-removal-fold-in-from-the-parent-and-moves-that-carry-the-side-folder-skill-tl-lead-sdlc-archiving.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 409e789d83910de366d11b8a536c2c6d8dc9900e..8517f816bd85d33877493b1c257dfe24218f2df4 |
| command | sdlc milestone-close --issue 424 --milestone M1 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-17T03:31:29-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Running the test suites and checking export consumers next.

Reviewing the contract prose and examples now.

Nearly done — checking the weave test result, then writing up.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers what Tasks 1 to 5 claim. `pkg/sidedoc` gives one home to the side-link regexp and the `discussions:` index. `collaboration-modes.md` now carries 07's sections 3 to 7 (no closing, the index, fold-in from the parent, moves) and rows 21 to 29. The six new examples and five updated ones pass the extended checker. The move cases are exported as `fixture` members in a regenerated v2 `export.json`. `go test ./pkg/sidedoc/ ./pkg/reviewexport/` passes. `cmd/weave`'s universal-review test passes once TMPDIR is outside the review checkout; inside it, it fails only because of the review sandbox's `GIT_CEILING_DIRECTORIES`. Three cheap fixes remain, none of them a blocker for the boundary:
- `ReconcileIndex` can silently misread or corrupt a hand-edited document in two ways I reproduced. Its own plan promises it refuses such input.
- The checker does not enforce the contract's "a fold adds no index entry". I confirmed this with a mutation that passed.

M2's tl-lead work will consume `ReconcileIndex` directly, so the parser fixes are worth doing now.

**1. Strengths**
- **One home for the side-doc facts (ARCH-DRY).** `pkg/sidedoc/links.go:13` replaces the checker's private `sideLinkRE`. `ParseCase` (`pkg/reviewexport/export.go:246`) serves both the export and the checker, replacing the removed `caseFields`.
- **Move cases are checked exactly.** `checkCase` compares `after/` file by file with `projectMove(before/)` (`examples_test.go:272-293`). The same before/after trees ship to parley, so the agent's prose, the Go reference and parley's mover share one fixture set.
- **Refusing leaves the input untouched.** When `ReconcileIndex` refuses, it returns the input byte for byte. A fuzz test pins this, and its committed corpus entry (a leading-space path) came from a real round-trip bug the fuzzer found.
- **The mutation suite grew with the invariants.** `TestCheckCaseCatches` gained one mutation per new invariant: entry resolution, an entry without its link, an entry dropped while its link remains, a link removed by the round, an unjustified add, and a wrong move projection.
- **Atlas is updated in the same range.** `atlas/workflow/universal-review.md` covers S3, the eighteen examples, and export v2.

**2. Critical findings**
None.

**3. Important findings**
- **A blank line inside the `discussions:` list ends the parse (`pkg/sidedoc/index.go:75-79`).** The check `l == t && !HasPrefix(t, "- ")` breaks on an empty line before the `t == ""` continue is reached. Entries after the blank line are never pruned, and new adds are spliced into the middle of what YAML still reads as one list. I reproduced it:
  - Input: `discussions:\n  - doc: d.side/01-a.md\n\n  - doc: d.side/09-gone.md` (no link to 09).
  - Result: 09 is not dropped, and 02-b is inserted between 01-a and 09-gone.
  - Fix: treat blank lines as continuation (check `t == ""` first), or refuse.
  - Add a test case. The fuzz oracle can't catch this, because it reads the output back with the same parser.
  - ARCH-SECURE.
- **A frontmatter `ReconcileIndex` can't split is treated as absent, and a second one is prepended (`index.go:33-48`, `:164`).** I reproduced two cases:
  - A CRLF document (`---\r\n…`) comes back as `---\ndiscussions:…\n---\n---\r\ndiscussions:\r\n…`, so the original frontmatter becomes body text.
  - An unclosed `---\nunclosed: x\nbody` gets the same treatment.
  - Fix: if the document starts with `---` but no frontmatter can be split, return an error and the input unchanged, as the plan's Trust section promises.
  - ARCH-SECURE.
- **The checker doesn't enforce "a fold adds no index entry" (`examples_test.go:384-388`).** The add check only fires when the side doc is unchanged and was already linked. A fold always changes the side doc by adding the closing note, so a fold that adds an entry passes. I confirmed this by adding `discussions: - doc: design.side/01-naming.md` to `fold-from-parent/after/design.md`: `TestExamples` stayed green. `foldNoteRE` already identifies a fold's closing note, so the fix is to flag a newly added entry whose side doc newly gained a fold note. Add one mutation for it. ARCH-PURPOSE.

**4. Minor findings**
- `move-rename/case.md` says "notes.md's link into the folder is not rewritten", but the fixture has no `notes.md` (the log says it was dropped). The text is stale and ships to parley as a fixture.
- The fuzz property (`index_test.go:110`) checks the output against `Entries`, which uses the same `indexBlock` parser. Parse bugs that are self-consistent pass, as the blank-line case shows. An independent oracle, such as a YAML decode of the frontmatter, would catch them.
- Plan drift: Task 1 lists `pkg/sidedoc/links_test.go`, but `TestSideLinks` lives in `index_test.go`. Task 2 and Task 3 name `.agents/skills/xx-universal-review/…`, but the source is `construct/local/universal-review/`.
- `checkIndexTransitions`' check for an entry dropped while its link remains (`examples_test.go:376`) compares raw entry strings, so a reformatted `./`-prefixed entry would be a false positive. Compare `path.Join`-ed paths instead.
- `TestUniversalReviewSkillDiscoveredInDerivedRepo` fails under the review confinement (`GIT_CEILING_DIRECTORIES` covers the in-checkout TMPDIR) and passes with TMPDIR outside it. This is environmental, not a defect in the diff.

**5. Test coverage notes**
- `ReconcileIndex` has good table coverage of the one-way rule and of refusals. It is missing cases for blank lines, CRLF line endings and unclosed frontmatter (see the Important findings).
- `TestBuildFixtures` pins Version 2, membership limited to move cases, and fixtures sorting last. `TestExportCurrent` checks that `export.json` is current (16 fixture members, matching 16 move-case files).
- Nothing tests a fold-chain or fold-from-parent mutation (adding an entry, or a marked link without a closing note).

**6. Architecture**
- **ARCH-DRY: pass.** Link regexp and case parsing are unified. `checkIndexTransitions` builds its own `path.Join` link sets alongside `sidedoc.linkedSet`; that's acceptable because the trees need joins relative to each file.
- **ARCH-PURE: pass.** `sidedoc` is pure over strings, and its tests use no IO or fakes. The checker runs over `fs.FS` / `fstest.MapFS`.
- **ARCH-PURPOSE: flag (Important #3).** Every checker invariant the issue names is delivered, but one rule the contract states (a fold adds no index entry) isn't enforced.
- **ARCH-MOCK: pass.** M1 calls no external binary or service. The Couch `write --all` seam belongs to M2.
- **ARCH-CONSTRAINTS: pass.** Tens of entries and files; the export grows by one member per move-case file.
- **ARCH-SECURE: flag (Important #1, #2).** Hand-edited frontmatter is the untrusted input, and two of its shapes are silently misread or corrupted instead of refused.
- **ARCH-ORDER: pass.** M1 holds no state between events: `ReconcileIndex` is a pure function of a single read. The prose covers the refused-write recovery ("that round adds both the link and the entry"). M2's two-write order is planned for later.
- **ARCH-FUNERAL: pass.** The one-way rule bounds index entries by live links. `.side/` folders end with their document's archive, which the prose states and M3 implements. Fixture members are bounded by the move cases.
- **For M2:** the tl-lead `syncIndex` will feed real session files into `ReconcileIndex`, so fix #1 and #2 first. Consider having the renderer keep the operator's original indentation and quoting for entries it keeps, so a no-drop add doesn't reformat the list.

**7. Plan revision recommendations**
- Add a `## Revisions` entry: "M1: the new state rows are 21 to 29, not 21 to 31. 07's rows 23 and 24 update the skill's rows 10 and 7. The move cases are rows 27 to 29." Also correct the Done-when ("rows 21 to 31") and the row numbers in Task 3.
- Correct the skill paths in Tasks 2, 3 and 5 to `construct/local/universal-review/` (the authored source), and note that `TestSideLinks` lives in `index_test.go`.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: hand-edited-index-parse-refuses
    title: |
      ReconcileIndex stops parsing discussions: at a blank line, so later entries are never pruned and adds splice mid-list
    detail: |
      pkg/sidedoc/index.go:75-79 breaks on an empty line before the t=="" continue. Reproduced with "discussions:\n  - doc: d.side/01-a.md\n\n  - doc: d.side/09-gone.md" (no link to 09): 09 is kept, and the add lands between 01-a and 09-gone, which YAML still reads as one list. Fix: treat blank lines as continuation or refuse; add a table case.
  - id: new
    severity: Important
    kind: defect
    family: hand-edited-index-parse-refuses
    title: |
      A frontmatter ReconcileIndex can't split (CRLF, unclosed) is treated as absent and a second one is prepended
    detail: |
      index.go:33-48 returns ok=false for "---\r\n..." and for an unclosed "---\n..."; ReconcileIndex (:164) then prepends a fresh frontmatter, and the original becomes body text. Reproduced for both. Fix: a document starting with --- whose frontmatter can't be split returns an error and the input unchanged, as the plan's Trust section promises.
  - id: new
    severity: Important
    kind: defect
    family: contract-rule-unenforced-by-checker
    title: |
      The examples checker does not enforce "a fold adds no index entry"
    detail: |
      checkIndexTransitions (examples_test.go:384-388) only flags an add when the side doc is unchanged and was already linked; a fold always adds the closing note, so it passes. Confirmed: adding the 01-naming entry to fold-from-parent/after/design.md leaves TestExamples green. Fix: flag a new entry whose side doc newly matches foldNoteRE, and add a mutation for it.
  - id: new
    severity: Minor
    kind: defect
    family: fixture-prose-matches-tree
    title: |
      move-rename/case.md describes a notes.md link that the fixture does not contain
    detail: |
      The case is exported to parley as a fixture; drop the sentence or add the file.
  - id: new
    severity: Minor
    kind: defect
    family: independent-test-oracle
    title: |
      FuzzReconcileIndex reads its output back with the same parser it is testing
    detail: |
      Self-consistent misparses (the blank-line case) pass the property; use an independent YAML decode of the frontmatter as the oracle.
  - id: new
    severity: Minor
    kind: defect
    family: plan-matches-code
    title: |
      Plan still says rows 21 to 31, a links_test.go, and .agents/skills/xx-universal-review paths
    detail: |
      The delivered rows are 21 to 29 (explained in the log, but no Revisions entry), TestSideLinks is in index_test.go, and the authored source is construct/local/universal-review/. Add a Revisions entry.
```

That background job was my first test run, the one that timed out and moved to the background. It covered `pkg/sidedoc`, `pkg/reviewexport` and `cmd/weave`. Its only failure was the `cmd/weave` test, and that failure comes from the review sandbox's `GIT_CEILING_DIRECTORIES` (the temp directory sits inside the review checkout). With TMPDIR moved outside the checkout, the same test passes.

The verdict stays FIX-THEN-SHIP with the same three Important findings:
- **Blank line in the index:** `ReconcileIndex` stops reading the `discussions:` list at a blank line, so entries after it are never pruned.
- **Unreadable frontmatter:** when the frontmatter can't be split (CRLF line endings, or no closing `---`), it adds a second frontmatter on top instead of refusing.
- **Fold rule not checked:** the examples checker doesn't enforce the rule that a fold adds no index entry.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d0bbd2a3-1565-4700-8724-42785c89b21d","repository":"github.com/xianxu/ariadne","issue_id":"000424","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/819ca98d954e956b","reviewed_head":"8517f816bd85d33877493b1c257dfe24218f2df4","slot":"ariadne:2","requested_at":"2026-10-17T10:25:46.221408Z"},"phase":"finished","pid":70769,"started_at":"2026-10-17T10:25:46.438189Z","finished_at":"2026-10-17T10:31:28.62108Z","session_id":"1a1f82ce-8ae9-405f-b233-8ade14ec25bd","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T03:52:02-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 424 — side docs S3: the discussions index, link removal, fold-in from the parent, and moves that carry the .side folder (skill, tl-lead, sdlc archiving) |
| repo | ariadne |
| issue file | workshop/issues/000424-side-docs-s3-the-discussions-index-link-removal-fold-in-from-the-parent-and-moves-that-carry-the-side-folder-skill-tl-lead-sdlc-archiving.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 409e789d83910de366d11b8a536c2c6d8dc9900e..a085177edd84ce83ea91aaccf9203e1b2159f9ff |
| command | sdlc milestone-close --issue 424 --milestone M1 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-17T03:52:02-07:00 |
| verdict | REWORK |

## Review

Probing done: a column-0 comment inside the index list still makes the parser stop early — same class as BR-1. Now checking that the claimed fixes go red when reverted.

```verdict
verdict: REWORK
confidence: medium
```

I checked all six prior findings against the code. Each is fixed: I reverted each behavior fix in the checkout and its regression test failed. BR-3's mutation fails when the fold check is removed. When `index.go` is restored to its pre-fix version, the blank-line table case, both unsplittable-frontmatter cases and four fuzz corpus entries all fail. The two prose fixes (BR-4, BR-6) are correct. `go test ./pkg/sidedoc/ ./pkg/reviewexport/` passes at head.

One thing blocks SHIP. This is the 3rd finding in family `hand-edited-index-parse-refuses`, and the class is still open. A column-0 YAML comment inside the `discussions:` list ends the block early, just as the blank line did in BR-1. Given `discussions:\n  - doc: d.side/01-a.md\n# c\n  - doc: d.side/09-gone.md\n` and an add, `ReconcileIndex` returns no error. It reports Entries `[01-a, 02-b]` and writes `01-a, 02-b, # c, 09-gone`. YAML reads that as three entries, so the unlinked `09-gone` is never pruned and the add is spliced into the middle of the list. Tab indentation (`\t- doc: …`), which YAML rejects, is also accepted and silently rewritten. Both contradict the new Revisions entry, which says it "reads only frontmatter that YAML reads the same way… Anything else is refused." The fuzz oracle (now a real YAML decode) would catch the comment case, but the corpus never reaches it.

1. **Strengths**
   - The fuzz oracle at `pkg/sidedoc/index_test.go:131-154` is now an independent `yaml.v3` decode. It found four real misparses, all now kept in the corpus.
   - Refusals return the document unchanged, and the tests assert this (`index_test.go:105-112`, the fuzz refusal branch).
   - `plainPathRE` (`index.go:26`) is an allowlist rather than a denylist, which is the right shape for untrusted input (ARCH-SECURE).
   - The BR-3 checker rule (`examples_test.go:389-393`) is backed by a mutation case that fails without it.
   - The fixture prose matches the trees: I swept every file named in the new `case.md` files and none is missing.

2. **Critical:** none.

3. **Important**
   - `pkg/sidedoc/index.go:87-110`: the block scan ends at any column-0 line that isn't `-`/`- …`, including a `# comment`. It also accepts tab-indented items.
     - **The rule to fix:** the index `ReconcileIndex` reads must be the index YAML reads, or it refuses.
     - **Structural fix:** before writing, decode the frontmatter lines with `yaml.v3` (already a test dependency). Refuse if it fails to decode, or if the decoded `discussions[].doc` list differs from `entries`. This turns the fuzz oracle into a production guard and covers comments, tabs and any later instance. A narrower option is to end the block only at a line matching a top-level key (`^[A-Za-z_][\w-]*:(\s|$)`) and refuse any other unindented line; that still needs a table case for each.
     - Then correct the Revisions entry so it states the guard.

4. **Minor:** none new. `discussions: # c` (valid YAML) is refused; that is the safe direction and acceptable.

5. **Test coverage notes**
   - The fuzz seed list has no comment line. Add `"discussions:\n  - doc: d.side/01-a.md\n# c\n  - doc: d.side/09-gone.md\n"` and a tab-indent case as table cases in `TestReconcileIndex` / `TestReconcileIndexRefuses`.
   - The full `go test ./...` hit the 120s foreground timeout and was moved to the background; I did not see its result. The two packages in scope passed in the foreground.

6. **Architectural notes**
   - **ARCH-DRY: pass.** Side links and the index live in one place, `pkg/sidedoc`, which the checker uses.
   - **ARCH-PURE: pass.** Everything is pure over strings, and the tests do no IO.
   - **ARCH-PURPOSE: flag.** The class from the finding above is only partly swept.
   - **ARCH-MOCK: pass.** No external dependency; the package is pure text.
   - **ARCH-CONSTRAINTS: pass.** It is a small per-document text transform with no hot path.
   - **ARCH-SECURE: flag.** A hand-edited file is trusted where the parser's reading differs from YAML's. This is the same finding.
   - **ARCH-ORDER: pass.** The package holds no state between calls; each reconcile is a single function of (doc, add).
   - **ARCH-FUNERAL: pass.** It creates only fuzz corpus files and fixtures, both versioned in the repo.

7. **Plan revision recommendations**
   - The 2026-10-17 "M1 as built" entry overstates what is refused. Reword it once the YAML cross-check lands, for example: "the reconcile decodes the frontmatter with YAML and refuses when its reading differs."

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      index.go:87-96 continues through blank lines; table case fails on revert to 8517f816.
  - id: BR-2
    disposition: addressed
    note: |
      errFrontmatter refuses CRLF and unclosed blocks unchanged; TestReconcileIndexRefuses fails on revert.
  - id: BR-3
    disposition: addressed
    note: |
      examples_test.go:389-393 flags fold-added entries; mutation "an entry added by a fold" fails with the check disabled.
  - id: BR-4
    disposition: addressed
    note: |
      notes.md sentence removed from move-rename/case.md; export.json hash updated; swept all new case.md file references, none missing.
  - id: BR-5
    disposition: addressed
    note: |
      Fuzz oracle is now a yaml.v3 decode of the reconciled frontmatter (index_test.go:131-154).
  - id: BR-6
    disposition: addressed
    note: |
      Revisions entry records rows 21-29, the construct/local source path, and TestSideLinks location.
findings:
  - id: new
    severity: Important
    family: hand-edited-index-parse-refuses
    title: |
      A column-0 comment inside discussions: ends the block, so later entries are never pruned and adds splice mid-list (3rd in family)
    detail: |
      index.go:93 breaks at any unindented non-item line. For "discussions:\n  - doc: d.side/01-a.md\n# c\n  - doc: d.side/09-gone.md\n" with add 02-b, ReconcileIndex returns nil error, reports Entries [01-a 02-b], and writes 01-a, 02-b, then the comment, then 09-gone, which YAML reads as three entries. Tab-indented items (invalid YAML) are also accepted and rewritten. This is the 3rd instance in this family. Rather than patching this site, enforce the rule: the reconcile decodes the frontmatter with yaml.v3 and refuses when the decoded discussions[].doc differs from what it parsed. Add table cases for the comment and tab inputs, and correct the plan Revisions claim "Anything else is refused".
```

The full test suite finished. Two failures came up, both outside the reviewed range, and I don't think either changes the verdict.

- **`cmd/datatype` `TestFindRepoRoot_GitFallback`:** it expects the git root of a temp repo, but got the review checkout's root instead. The temp directory sits inside the checkout, at `.sdlc-review/tmp/...`, so the lookup most likely walks up and finds the outer repo first. That points to how this review checkout is set up, not to #424.
- **A package timed out after 10 minutes in `TestArchiveDoneIssuesInDir_UntrackedSidecar_RealRepo`:** the log doesn't name the package. #424's M1 doesn't touch archiving (that's M3), so this also looks like a slow or hanging test in this environment, not something the diff caused.

I didn't rerun either at the base commit, so I can't say for sure whether they already failed there. If the main agent wants certainty, run `go test ./cmd/datatype/` and the archive tests at `409e789d` in the same checkout and compare.

The two packages in scope, `pkg/sidedoc` and `pkg/reviewexport`, pass. The verdict stays REWORK because of the one Important finding: a column-0 `# comment` inside the `discussions:` list makes `ReconcileIndex` misread the index and write the wrong entries.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ce460e10-dce7-44d7-a171-b96272dff9c5","repository":"github.com/xianxu/ariadne","issue_id":"000424","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/95c83b13d48626aa","reviewed_head":"a085177edd84ce83ea91aaccf9203e1b2159f9ff","slot":"ariadne:2","requested_at":"2026-10-17T10:40:40.154656Z"},"phase":"finished","pid":95532,"started_at":"2026-10-17T10:40:40.43711Z","finished_at":"2026-10-17T10:52:01.648846Z","session_id":"64c810b9-15c0-4ef2-b73b-4207b2ff712a","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T03:57:31-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 424 — side docs S3: the discussions index, link removal, fold-in from the parent, and moves that carry the .side folder (skill, tl-lead, sdlc archiving) |
| repo | ariadne |
| issue file | workshop/issues/000424-side-docs-s3-the-discussions-index-link-removal-fold-in-from-the-parent-and-moves-that-carry-the-side-folder-skill-tl-lead-sdlc-archiving.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 409e789d83910de366d11b8a536c2c6d8dc9900e..b4b62c42ffb718dc720f79d7908aa8ae56bbd901 |
| command | sdlc milestone-close --issue 424 --milestone M1 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T03:57:31-07:00 |
| verdict | SHIP |

## Review

Focus on BR-7 first: the last commit.

```verdict
verdict: SHIP
confidence: high
```

BR-7, the one open finding, is fixed, and I found no new blocking issues. The fix treats YAML as the authority, as the finding asked, rather than patching one more input shape. `ReconcileIndex` and `Entries` now decode the frontmatter with yaml.v3 and refuse when YAML reads the index differently from what the line parser found. `ReconcileIndex` also checks its own rewritten output before returning it (`pkg/sidedoc/index.go:198-201` and `:228-231`). I tested the fix by making `agreesWithYAML` return nil in the disposable checkout. `TestReconcileIndexRefuses` then failed on both BR-7 inputs: the column-0 comment and the tab-indented item. With the code restored, `go test ./pkg/sidedoc/ ./pkg/reviewexport/` passes, and a 20-second fuzz run of `FuzzReconcileIndex` (about 990k executions) found no failures. The plan's Revisions entry now describes the YAML-authority rule and no longer says "Anything else is refused".

1. **Strengths**
   - `pkg/sidedoc/index.go:247-266` (`agreesWithYAML`) fixes the whole `hand-edited-index-parse-refuses` family, not just the latest case. Anything YAML reads differently is refused: comments, tabs, merge keys or anchors that inject `discussions`, duplicate keys, nulls.
   - Checking both before and after the rewrite means that even if the line parser finds the wrong span, the document can't be silently corrupted. That check is what turns the line parser's job into "find the block", not "decide what the index says".
   - When the reconcile refuses, the document comes back byte-for-byte unchanged, and both the table tests and the fuzz test check this (`index_test.go:95-97`, `FuzzReconcileIndex`).
   - `pkg/sidedoc/links.go` is a small pure package of text functions with no IO, and the examples checker, export and later consumers can all reuse it.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `SideLinks` closes a fence on any line starting with the opening 3-character prefix. A fence opened with four backticks therefore closes on a three-backtick line, which differs from CommonMark. This is an edge case and doesn't block.
   - Since BR-7, the fuzz oracle and the implementation use the same decoder (yaml.v3). That is intended, because YAML is now the authority, but the oracle can no longer catch a yaml.v3 quirk.

5. **Test coverage:** The BR-7 refusal cases go red when the fix is removed, so they test the actual change. The `extra: x` case is already rejected by the line parser, so it doesn't exercise the YAML check, but it does no harm. The no-frontmatter, CRLF, unclosed-frontmatter and blank-line-inside-the-list cases are all still covered.

6. **Architecture**
   - **ARCH-DRY: pass.** One YAML-agreement helper is shared by `Entries` and `ReconcileIndex`. `linkedSet` and `SideLinks` are each written once.
   - **ARCH-PURE: pass.** `sidedoc` works only on strings. The `reviewexport` changes read through an `fs.FS`, so tests can use in-memory file systems.
   - **ARCH-PURPOSE: pass.** This round fixed the family's rule rather than another instance.
   - **ARCH-MOCK: N/A.** No external binaries or services are involved.
   - **ARCH-CONSTRAINTS: pass.** The work runs on small, single-document frontmatter, and the extra YAML decode is negligible.
   - **ARCH-SECURE: pass.** A hand-edited index is untrusted input, and it is now parsed into typed values. A bad index is refused with a visible error and the document unchanged; nothing is made up.
   - **ARCH-ORDER: N/A.** The reconcile keeps no state between calls; each call is a pure function of the document and the list of adds.
   - **ARCH-FUNERAL: pass.** Index entries are removed by the one-way rule (a deleted link drops its entry). The new export fixtures are versioned files, and Version is bumped to 2.

7. **Plan revisions:** none needed. The M1-as-built Revisions entry now matches the code.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      agreesWithYAML decodes the frontmatter with yaml.v3 before and after the rewrite (index.go:198,228). With it reverted, TestReconcileIndexRefuses fails on the comment and tab cases; restored, the tests and 20s of fuzzing pass. The plan Revisions entry is corrected.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7ebaa550-16d6-4725-9f61-f9fdd9ed83c2","repository":"github.com/xianxu/ariadne","issue_id":"000424","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/956b70ec7253af13","reviewed_head":"b4b62c42ffb718dc720f79d7908aa8ae56bbd901","slot":"ariadne:2","requested_at":"2026-10-17T10:56:13.187162Z"},"phase":"finished","pid":14792,"started_at":"2026-10-17T10:56:13.279868Z","finished_at":"2026-10-17T10:57:30.522148Z","session_id":"90a8d77e-c5a7-4ab7-aaf4-2a9b2e91154b","outcome":"returned"}] -->
