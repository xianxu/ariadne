# Boundary Review — ariadne#444 (whole-issue close)

| field | value |
|-------|-------|
| issue | 444 — review convention: one turn per marker; a longer exchange moves to a side doc |
| repo | ariadne |
| issue file | workshop/issues/000444-review-convention-one-turn-per-marker-a-longer-exchange-moves-to-a-side-doc.md |
| boundary | whole-issue close |
| milestone | — |
| window | cf565ef7914b53b0485a587586f4b094643cf13f..89d3effe978c89d08f872c9b78425ad53f4d1d40 |
| command | sdlc close --issue 444 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T15:20:32-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The convention, skills, and revised examples consistently route continued exchanges into indexed side docs. Focused tests and Parley compatibility tests pass. Two inexpensive fixes remain: preserve legacy chains across fixture moves, and document the changed user-facing convention in README.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: legacy-preservation-across-moves
    title: |
      The new fixture guard rejects unchanged legacy chains after a file move
    detail: |
      pkg/reviewexport/examples_test.go:260–264 compares each after-file only with the identical before-path. A valid move from design.md to spec.md containing an unchanged 🤖[ask]{reply} therefore reports “agent wrote a second turn” because beforeText["spec.md"] is empty. The complete affected family is the renamed parent document, side docs moved with its folder, and destination orphan side docs displaced by the move contract. Validate move cases through their existing exact projectMove projection before applying the emission guard, or compare against a projected baseline. Add positive move regressions alongside the existing append-rejection tests.
  - id: new
    severity: Important
    kind: defect
    family: user-facing-convention-documentation
    title: |
      README update appears missing for the changed marker convention
    detail: |
      construct/local/fix/review-convention.md:45–55 changes the syntax users write and the clarification workflow, but README.md is unchanged in the pinned range. The affected surface comprises the one-commentary-block limit, continued exchanges opening linked and indexed side docs, and legacy-chain compatibility. Add a short explanation beside README.md:23–28’s existing universal-review usage paragraph, with a link to the canonical convention.
```

1. **Strengths**

   - Both ambiguous-ask examples now contain valid side-doc links, matching index entries, and backlinks.
   - The canonical convention distinguishes writer restrictions from legacy reader compatibility.
   - The export manifest remains consistent with the changed protocol files; `pkg/reviewexport` passes.
   - Parley’s existing tests confirm both human-ended and agent-ended legacy chains remain parseable.

2. **Critical findings**

   None.

3. **Important findings**

   The two findings above require fixes before close.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Ran `go test ./pkg/reviewexport -count=1`, `git diff --check`, and Parley’s `make test-spec SPEC=modes/review`; all passed. Append mutations cover a new second turn and extension of a legacy chain. Missing coverage is preservation of legacy chains across the three move categories enumerated above.

6. **Architectural notes**

   - **ARCH-DRY — pass:** the canonical grammar remains shared through the existing symlink and export mechanism.
   - **ARCH-PURE — pass:** no new business logic is mixed into an IO shell.
   - **ARCH-PURPOSE — pass:** the changed agent routes deliver one-turn emission and side-doc escalation; the compatibility guard needs the move correction above.

7. **Plan revision recommendations**

   Add a `## Revisions` entry expanding verification to cover unchanged legacy chains in renamed parents, moved side docs, and displaced orphan folders.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"eaeafdcc-5e85-4baa-9beb-f422f4cc07c6","repository":"github.com/xianxu/ariadne","issue_id":"000444","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/0152b5d1bc9a9d28","reviewed_head":"89d3effe978c89d08f872c9b78425ad53f4d1d40","slot":"ariadne:1","requested_at":"2026-10-18T22:18:56.338672Z"},"phase":"finished","pid":83495,"started_at":"2026-10-18T22:18:56.432008Z","finished_at":"2026-10-18T22:20:32.166272Z","session_id":"01a15118-b725-7131-9a8c-4c6a1525342b","outcome":"returned"}] -->
