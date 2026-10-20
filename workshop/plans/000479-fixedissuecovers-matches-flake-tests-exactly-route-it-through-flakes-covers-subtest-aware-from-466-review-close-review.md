# Boundary Review — ariadne#479 (whole-issue close)

| field | value |
|-------|-------|
| issue | 479 — fixedIssueCovers matches flake_tests exactly; route it through flakes.Covers (subtest-aware, from #466 review) |
| repo | ariadne |
| issue file | workshop/issues/000479-fixedissuecovers-matches-flake-tests-exactly-route-it-through-flakes-covers-subtest-aware-from-466-review.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4ce53fe9728bee39a0beb2d57cd51b678a13146a..04bf0c5e603d97658e0e3827ef8364aa49664498 |
| command | sdlc close --issue 479 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T21:12:25-07:00 |
| verdict | SHIP |

## Review

The tests pass. Next I'll do a quick mutation check to confirm the new tests actually catch a regression.

```verdict
verdict: SHIP
confidence: high
```

The change does what #479 asked for. `flake list`'s fixed-issue path now uses the same coverage rule as the gate: `flakes.Uncovered`, which reuses the gate's own `explain`, gives the failing leaves, and `flakes.Covers` picks the covering name, nearest first. Both clauses of Done-when are delivered, and `fixedIssueCovers` is deleted. The relevant tests pass in this checkout (`go test ./internal/flakes/` and `go test -run 'Flake|Fixed|Parallel' .`), and `go vet` is clean. I tried one mutation: making `fixedCandidates` match names exactly turns `TestFixedIssueCoverageFollowsTheGateRule` red. Only Minor notes remain, and none blocks the boundary.

1. **Strengths**
   - `cmd/sdlc/internal/flakes/evidence.go:88-98`: `Explained` is now defined by `Uncovered`. Both wrap the gate's `explain`, so leaf splitting has a single source and `flake list` can no longer drift from it.
   - `cmd/sdlc/flake_fixed.go:107-124`: when one name is the nearest cover, its issue wins. Two issues naming that same name stay ambiguous, which keeps the "multiple terminal issues" unknown path. One issue naming both `x` and `x/sub` is counted once. Any two names that cover the same leaf are its ancestors, so they always differ in length. That makes the length comparison a sound way to rank them.
   - `cmd/sdlc/flake_fixed_test.go:21-55`: the table covers every case the Plan lists: exact match, a sibling, a parent, an ancestor, nearest-wins in both orders, a tie, and the "name prefix is not a test" boundary (`TestXExtra`).
   - `cmd/sdlc/flake_cmd.go:658`: each row now names the uncovered leaf, which is the name someone would register. A failure the open registry already covers is never checked against fixed issues, so the existing precedence holds.
   - `atlas/workflow/sdlc-binary.md` is updated with `flakes.Uncovered` and `fixedCandidates`.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - `flake_fixed_test.go:93`: `parser.ParseDir` is deprecated in Go 1.26 (it ignores build tags). That makes no difference to this name check, but a future staticcheck run would flag it (SA1019).
   - `TestNoParallelFlakeCoverageMatcher` only catches functions whose names end in `Covers`. An inline `==` coverage match would get past it. The Log's audit of the remaining `==` sites (identity checks, not coverage) is what actually settles Done-when 2. That's acceptable, but the lint is narrower than its comment suggests.
   - ARCH-DRY: `fixedCandidates` repeats the nearest-name loop in `Registry.Cover` (`internal/flakes/registry.go:183`). Keeping it separate is justified, because `Cover` keeps one owner per name and ties are impossible there, while fixed issues need ties to come out as ambiguous. A shared `flakes.Nearest(names, leaf)` helper could still feed both.
   - Under the exact-match mutation, `TestFlakeListJudgesFixedSubtestsFromTheLog` stays green, because the logged leaf is exactly the fixed name. No end-to-end `flake list` test covers a fixed top-level `TestX` hiding a logged `TestX/sub`; only the unit table does.

5. **Test coverage notes:** the unit table pins the matching rule, and the mutation confirms it. The `flake list` test covers both states of the logged-leaf clause (the fixed subtest is hidden, a sibling stays listed by name). The Log says a mutation restoring the old top-level loop fails it; I did not re-run that one.

6. **Architectural notes**
   - ARCH-DRY: pass, with the Minor note above. The coverage predicate has one source (`flakes.Covers` and `explain`).
   - ARCH-PURE: pass. `fixedCandidates` is pure and tested directly; `classifyFixedFlake` keeps git ancestry injected.
   - ARCH-PURPOSE: pass. I checked every consumer of fixed-issue matching: `classifyFixedFlake` is the only one, and its single caller is `runFlakeListMode`. It now goes through the shared rule.
   - ARCH-EVOLVE: pass, not applicable. No persisted format changes; `Subtests` stays unpersisted.

7. **Plan revision recommendations:** none. The Plan says `Explained` becomes `len(Unknown)==0`, and the Log already records the rename to `Uncovered`.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: deprecated-stdlib-api
    title: |
      TestNoParallelFlakeCoverageMatcher uses go/parser.ParseDir, deprecated in Go 1.26
    detail: |
      flake_fixed_test.go:93. It works for this name-only scan, but staticcheck SA1019 would flag it. Use go/packages, or os.ReadDir plus parser.ParseFile. This is the only instance in the window.
  - id: new
    severity: Minor
    kind: defect
    family: lint-proxy-narrower-than-claim
    title: |
      The no-parallel-matcher lint only catches functions whose names end in Covers, not inline exact-match coverage
    detail: |
      Done-when 2 actually rests on the Log's manual audit of the remaining == sites (removeFlakeTests, runFlakeRegister, Registry.Lookup). The comment on the AST test could say it guards only that one regression shape.
  - id: new
    severity: Minor
    kind: defect
    family: nearest-cover-duplicated
    title: |
      ARCH-DRY: fixedCandidates repeats Registry.Cover's nearest-covering-name loop
    detail: |
      flake_fixed.go:107 and internal/flakes/registry.go:183 both pick the longest name that Covers the leaf. Fixed issues need ties to be ambiguous, so the loops can't simply be merged. A shared flakes helper that returns the nearest covering name could feed both. Only these two instances exist.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"0017bec4-9149-4a4a-a710-8cb74c288941","repository":"github.com/xianxu/ariadne","issue_id":"000479","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/007572c22853e62d","reviewed_head":"04bf0c5e603d97658e0e3827ef8364aa49664498","slot":"ariadne:2","requested_at":"2026-10-20T04:08:55.844434Z"},"phase":"finished","pid":76336,"started_at":"2026-10-20T04:08:56.022127Z","finished_at":"2026-10-20T04:12:25.454834Z","session_id":"1bae3aab-a324-47d8-8169-344b0b0a34ad","outcome":"returned"}] -->
