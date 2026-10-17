# Boundary Review — ariadne#415 (whole-issue close)

| field | value |
|-------|-------|
| issue | 415 — S2 M2: every managed document type declares a default collaboration-mode |
| repo | ariadne |
| issue file | workshop/issues/000415-s2-m2-every-managed-document-type-declares-a-default-collaboration-mode.md |
| boundary | whole-issue close |
| milestone | — |
| window | 63b1a010fc175881b8311e8588676a70991a8ad3..614b8f11f0eb8bd986d57eac0df1963831c3eaf9 |
| command | sdlc close --issue 415 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-16T19:45:26-07:00 |
| verdict | SHIP |

## Review

Inspecting done; now running the targeted test suites to check behavior.

Checking whether plan readers tolerate the new frontmatter block, then writing the verdict.

Last check: whether `datatype lint` is actually run anywhere (Makefile/CI).

```verdict
verdict: SHIP
confidence: high
```

The issue asked for one declared default collaboration mode per managed type, a writer that puts it into new documents, and a lint. The diff delivers all three, matching the plan table. Issues and projects take their default from `scaffold.collaboration_mode` in their CUE models. `issue.Render`, `DetailsFromCard` and `project.RenderScaffold` all write that value, and tests compare it to the model rather than to a hard-coded string. The ten datatype prototypes each declare theirs in a Frontmatter-shape row. The new `datatype lint` reads every type the registry resolves, and running it here prints "10 types, each declares a default collaboration mode". The plan header in the writing-plans skill now has frontmatter, and a regex test pins it. The TL session file (`cmd/tl-lead/help.md:136`) and the side and answer docs (`collaboration-modes.md` §4, around line 187) already declared their defaults before this change. The targeted suites pass: `./pkg/vocab`, `./cmd/vocabulary`, `./cmd/sdlc/internal/issue`, `./cmd/sdlc/internal/project` and `./cmd/datatype`. `TestFindRepoRoot_GitFallback` failed until I set `TMPDIR=/tmp`. That test is older than this change; it fails because the review checkout puts its temp directory inside the repo. Nothing blocks shipping; the findings below are all Minor.

1. **Strengths**
   - **One source per type.** Go reads the default from the CUE model (`pkg/vocab/vocab.go:105-114`, `project.go:55-57`). The project prototype row is checked against `project.cue` (`cmd/datatype/collabmode_test.go:190-199`), so the two can't drift apart.
   - **The lint catches new prototypes automatically.** It runs over whatever `resolveTypes()` returns, not a hand-kept list (`collabmode_test.go:179-188`). It is also a real subcommand, not just a test.
   - **The parser is pure.** `collaborationModeDefault` works on a string and is tested with no IO. It ignores a decoy row in the Body skeleton section (`collabmode_test.go:144`).
   - **The writer tests compare against the model**, e.g. `"collaboration-mode: " + vocab.Issue().CollaborationMode()` (`handoff_test.go:64`, `project/scaffold_test.go:19`).
   - **The `pkg/vocab/issue.json` setter change is not scope creep.** The embed was stale relative to `issue.cue:108`, and regenerating it fixed that.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **Lint not run in derived repos** (ARCH-PURPOSE): no make target, gate or CI step runs `datatype lint`. In ariadne the Go test covers it, but a derived repo's own prototypes are only checked if someone runs the command by hand.
   - **Three copies of the mode list** (ARCH-DRY): the values appear in `vocab.CollaborationModes`, inline in `pensive.cue:25` and in `collaboration-modes.md`. Also, `scaffold: collaboration_mode` has no type in CUE, so only the Go test rejects a typo.
   - **Parser vs. stated contract:** `collaborationModeDefault` accepts the first backticked value anywhere in the notes, though the docs say "notes start with". It also fails on CRLF prototypes with the misleading message "no Frontmatter shape section".
   - **Possible test panic:** `TestEveryTypeDeclaresACollaborationMode` indexes `protos[i]` without checking for `i == -1`, so it panics instead of failing if `project` ever stops resolving.
   - **Task 5 evidence not logged:** the plan asked to check that plan readers tolerate a frontmatter block, but the Log doesn't record the result. I looked: `issueTitleFromContent` and the flow line counter are fine with it; the line counter now counts four more lines per durable plan.
   - **Spec vs. issue default:** the Spec says "design docs → discussion", but in ariadne the design lives in the issue's `## Spec` section, and issues default to `collaboration`. A per-section override could reconcile this later.

5. **Test coverage notes:** the lint has both positive and negative tests. All three writers are pinned. The plan header and the drift between `project.md` and `project.cue` are both tested. `runLint`, the IO wrapper, has no test, which is acceptable for thin glue.

6. **Architectural notes**
   - **ARCH-DRY:** pass, apart from the Minor note on the copied mode list.
   - **ARCH-PURE:** pass. The parser is pure and the IO sits in `runLint`.
   - **ARCH-PURPOSE:** pass. Every row of the plan table derives from its declaration; the derived-repo lint is the separable extension noted above.
   - **ARCH-MOCK:** N/A. No external binary or service is involved.
   - **ARCH-CONSTRAINTS:** N/A. The lint reads about 10 files per run.
   - **ARCH-SECURE:** pass. The prototypes are repo-local, and a malformed one produces a visible error rather than a made-up default.
   - **ARCH-ORDER:** N/A. Nothing holds state between events; each writer is a single render.
   - **ARCH-FUNERAL:** N/A. The change adds no new durable artifact family; existing documents just gain one frontmatter line.

7. **Plan revision recommendations:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: lint-not-wired-to-gate
    title: |
      datatype lint is not invoked by any make target, gate, or CI in derived repositories
    detail: |
      Only TestEveryTypeDeclaresACollaborationMode exercises it in ariadne; a derived repo's local prototype without a collaboration-mode row passes unless the operator runs the command by hand.
  - id: new
    severity: Minor
    kind: defect
    family: enum-single-source
    title: |
      The collaboration-mode value set is restated in Go, pensive.cue and collaboration-modes.md; the CUE scaffold value is untyped
    detail: |
      Only the Go vocab test rejects a typo in scaffold.collaboration_mode; a shared CUE enum would make the CUE model enforce it too.
  - id: new
    severity: Minor
    kind: defect
    family: parser-matches-stated-contract
    title: |
      collaborationModeDefault accepts the first backticked token anywhere in the notes, not at the start; CRLF prototypes report a misleading error
    detail: |
      The docs say the notes start with the default. A CRLF file fails with "no Frontmatter shape section".
  - id: new
    severity: Minor
    kind: defect
    family: test-guards-lookup
    title: |
      TestEveryTypeDeclaresACollaborationMode indexes protos[i] without checking i != -1
    detail: |
      If project ever stops resolving, the test panics instead of reporting a failure.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"9f699c5f-e7f1-489f-a49c-126afeef5ca2","repository":"github.com/xianxu/ariadne","issue_id":"000415","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/6daf07d4d8089694","reviewed_head":"614b8f11f0eb8bd986d57eac0df1963831c3eaf9","slot":"ariadne:2","requested_at":"2026-10-17T02:43:21.74706Z"},"phase":"finished","pid":37113,"started_at":"2026-10-17T02:43:21.883985Z","finished_at":"2026-10-17T02:45:26.004376Z","session_id":"283abace-6516-4233-8215-cc8a4f1b8d71","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T19:47:52-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 415 — S2 M2: every managed document type declares a default collaboration-mode |
| repo | ariadne |
| issue file | workshop/issues/000415-s2-m2-every-managed-document-type-declares-a-default-collaboration-mode.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5a4dbf182040b89e2050c906b21502eb3006274e..21ba3b6c13b4123a639a58666228c4764436ee22 |
| command | sdlc close --issue 415 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-16T19:47:52-07:00 |
| verdict | SHIP |

## Review

I've finished checking the fixes and confirmed both parser regressions by reverting them. The verdict follows.

```verdict
verdict: SHIP
confidence: high
```

All four Minor findings from round 1 are fixed in this window, and the fixes add no Critical or Important problem. I reverted each of the two parser fixes separately in the disposable checkout and ran the tests, and each revert made `TestCollaborationModeDefault` fail. With the old code the CRLF test fails with "no Frontmatter shape section", and the "not at the start" test fails because no error is returned. With both fixes in place, `./cmd/datatype` and `./pkg/vocab` pass. Nothing in the issue file's Spec, Revisions or Done-when changed. The only edits there are Log lines, the status, `actual_hours` and the flow upgrade, so the scope did not change after close. No main conflict resolutions are listed for this window.

1. **Strengths**
   - **BR-1 is fixed without extra setup.** The lint now runs inside `renderToOutput` (`cmd/datatype/main.go:91-96`), which `construct/local/datatype/.dynamic-skill:6` calls during every repository's weave compile. So a derived repository's own prototypes get checked with no extra wiring. It warns rather than failing, so a bad prototype in a derived repo can't break its compile.
   - **The parser now does what the docs say.** It normalizes CRLF first (`collabmode.go:18`), then requires the notes to *start* with the backticked default (`collabmode.go:29`). Each of the two new test cases fails if its fix is removed.
   - **One test ties all three copies of the mode list to the Go list.** `TestCollaborationModesAgreeAcrossSources` (`pkg/vocab/vocab_test.go:280-299`) checks `pensive.cue` and `collaboration-modes.md` against `vocab.CollaborationModes`, so the three can't drift apart unnoticed.
   - **The lookup guard turns a possible panic into a clear failure** (`collabmode_test.go:71-73`).

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - No test covers the new warning in `renderToOutput`. It is thin IO glue around `lintTypes`, which is tested, so this is acceptable.

5. **Test coverage notes**
   - Both BR-3 regressions are confirmed red without their fixes.
   - BR-2 is guarded by a drift test, not by one shared enum. That catches drift, but the CUE scaffold value still has no type of its own (see the BR-2 note below).
   - BR-4 is a test-only guard and needs no further evidence.

6. **Architectural notes**
   - **ARCH-DRY:** pass. The mode list is still written out in three places, but a test now binds them to the Go list. A shared CUE enum, exported to Go, would remove the copies entirely if this area is reworked.
   - **ARCH-PURE:** pass. The parser stays a pure function on a string, and the new IO is one loop in the existing IO shell.
   - **ARCH-PURPOSE:** pass. BR-1 is fixed at the point every repository's compile passes through, not only for ariadne. BR-3 fixes both parts of its finding (start-of-notes and CRLF). BR-2 binds every copy of the list, not just the one the finding named.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      renderToOutput (run by every repo's weave compile via construct/local/datatype/.dynamic-skill) now calls lintTypes and warns on stderr (main.go:91-96).
  - id: BR-2
    disposition: addressed
    note: |
      TestCollaborationModesAgreeAcrossSources binds pensive.cue's enum and collaboration-modes.md's error text to vocab.CollaborationModes. The scaffold value still has no CUE type; TestScaffoldsDeclareACollaborationMode covers that in Go, which is acceptable at Minor.
  - id: BR-3
    disposition: addressed
    note: |
      CRLF is normalized and the notes must start with the default (CutPrefix on the trimmed cell). Reverting either fix makes TestCollaborationModeDefault fail (CRLF case, and the not-at-the-start case).
  - id: BR-4
    disposition: addressed
    note: |
      collabmode_test.go:71-73 calls t.Fatal when the project lookup returns i < 0.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: lint-not-wired-to-gate
    title: |
      The new compile-time warning loop in renderToOutput has no test
    detail: |
      lintTypes is tested, but no test checks that the render step prints a warning, so the BR-1 wiring could be dropped without any test failing. Acceptable as thin IO glue; a small test of renderToOutput against a fixture tree with a prototype lacking a collaboration-mode row would pin it.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"0a7260da-9e79-46e1-95cb-1991e40df217","repository":"github.com/xianxu/ariadne","issue_id":"000415","recipe":"fix-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/d3075f6a596001ae","reviewed_head":"21ba3b6c13b4123a639a58666228c4764436ee22","slot":"ariadne:2","requested_at":"2026-10-17T02:46:43.158057Z"},"phase":"finished","pid":25086,"started_at":"2026-10-17T02:46:43.273768Z","finished_at":"2026-10-17T02:47:43.271045Z","session_id":"b61a53e8-fe0f-45f7-b7df-29d05fedb13c","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T19:49:54-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 415 — S2 M2: every managed document type declares a default collaboration-mode |
| repo | ariadne |
| issue file | workshop/issues/000415-s2-m2-every-managed-document-type-declares-a-default-collaboration-mode.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1c03779ee7ef6f700657c3f61b556fb3bad263df..c812b6e5e3756aa302676ecfb1f576a791ac0287 |
| command | sdlc close --issue 415 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-16T19:49:54-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The only open finding, BR-5, is fixed. `renderToOutput` now takes an `io.Writer` for its warnings (`cmd/datatype/main.go:83,96`). `main` passes in `os.Stderr`, so what the binary does is unchanged. The new test `TestRenderWarnsOnATypeWithoutADefault` (`cmd/datatype/collabmode_test.go:100-124`) checks three things: the warning appears for a prototype with no collaboration-mode row, it doesn't appear for a prototype that has one, and `SKILL.md` is still written.

I checked that the test fails without the fix. In the disposable checkout I replaced the warning print with a no-op and ran the test. It failed with `warnings "", want one for bare only` (collabmode_test.go:119). After restoring the file, the tree is clean and the test passes. `./cmd/datatype` passes in full with `TMPDIR=/tmp`. With the default sandbox `TMPDIR`, which sits inside the checkout, `TestFindRepoRoot_GitFallback` fails. This window doesn't touch that test. The failure comes from the temp directory being nested inside the repository.

The scope did not change. The window only touches `main.go`, `collabmode_test.go` and the two generated gate/review ledgers. The Spec, Revisions and Done-when are untouched, and no conflict resolutions are listed.

1. **Strengths**
   - BR-5 is fixed as the finding suggested: a fixture tree with one good and one bare prototype runs through the real render path. The test checks for missing warnings and for extra ones.
   - Passing the writer in (`main.go:83`) makes the warning testable without capturing `os.Stderr`. The IO shell stays thin, and the caller decides where warnings go.
   - The test also checks that a warning doesn't block the render, which pins the "warn, don't fail the compile" design.

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none new.
5. **Test coverage notes:** The new test fails with the warning print replaced by a no-op and passes with it restored. `t.Chdir` limits the cwd change to this test.
6. **Architectural notes:**
   - **ARCH-DRY: pass.** The test reuses the existing `proto` and `shapeRow` helpers.
   - **ARCH-PURE: pass.** The writer passed into the IO shell keeps the logic in `lintTypes`, which is pure.
   - **ARCH-PURPOSE: pass.** The finding's class was "the lint call in the compile has no test", and that call is the only one in this window, so the whole class is covered.
7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      TestRenderWarnsOnATypeWithoutADefault (collabmode_test.go:100-124) runs renderToOutput, now taking an io.Writer, on a fixture tree; making the warning print a no-op turns it red ("warnings \"\", want one for bare only").
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f94fe7c2-3bb0-4aee-b7bf-aa25a0e5c363","repository":"github.com/xianxu/ariadne","issue_id":"000415","recipe":"fix-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c38d9cbf89291776","reviewed_head":"c812b6e5e3756aa302676ecfb1f576a791ac0287","slot":"ariadne:2","requested_at":"2026-10-17T02:49:05.417384Z"},"phase":"finished","pid":48653,"started_at":"2026-10-17T02:49:05.594496Z","finished_at":"2026-10-17T02:49:46.988917Z","session_id":"c393b803-80f3-44ab-8211-d37f6f0a3a67","outcome":"returned"}] -->
