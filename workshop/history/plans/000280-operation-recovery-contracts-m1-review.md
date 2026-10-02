# Boundary Review — ariadne#280 (milestone M1)

| field | value |
|-------|-------|
| issue | 280 — Publish tested operation recovery contracts |
| repo | ariadne |
| issue file | workshop/issues/000280-operation-recovery-contracts.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 3b7315b7fde5083f036e377d41e8053193e52ef3..3667fb9b7bff19f757258d57c7a424e7e7684195 |
| command | sdlc milestone-close --issue 280 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-02T11:08:23-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**FIX-THEN-SHIP.** M1 delivers what the plan says it would, and the new tests pass. One registry, `internal/recovery`, now drives both the help text and the contract test. Every single-card compare-and-swap verb (claim, adopt, relocation, the card setters, reclaim) publishes through one function, `cardPublish`. A grep shows no other `UpdateCard` call sites remain; the only other card write is `ChangeCard` in landing completion, which is a separate path. Every verb that gets an "uncertain" error now tells the agent to rerun, worded the same way everywhere. The tests that lose an acknowledgement do real work: the publish lands, the error is injected afterwards, and the rerun returns the settled answer. One thing must be fixed before the boundary: the catalog describes only tracker repositories, but its text appears in `--help` for every repository, and some statements are false in legacy repositories. The rest are cheap cleanups.

**What I ran:**
- `go test ./cmd/sdlc/internal/recovery/` passes.
- In `./cmd/sdlc`: `TestRecoveryContractsAreProven`, the two rerun-after-lost-response tests, `TestCloseRerunAfterShipStartsANewGeneration`, `TestLandingLeavesAReopenedCardAlone`, `TestCardPublishCallers`, `TestReclaimStaleAndLostResponse`, `TestNoCommandLongHasSurvivingPlaceholder`, `TestEveryFlagAppearsInItsHelp`, and the help and process-manual tests all pass.
- `sdlc claim --help` and `sdlc issue set-title --help` both show exactly one recovery section.
- I could only run the git commands with the sandbox disabled.

## Strengths
- **The required verb set is derived, not hand-listed.** `recovery_contract_test.go:57-71` walks the command tree for the repo-lock annotation. That covers `markManualLockCommand`, so change-code, close and milestone-close are required automatically. It then adds start-plan and issue show. It also flags exemptions that no longer match a command, and contracts that name something that isn't a command.
- **Help is attached in one pass.** `attachRecoveryContracts` (`main.go:187-201`) adds the section to every page, including the issue subcommands whose help is written inline. Dropping the `{{RECOVERY}}` placeholder is recorded in a plan revision.
- **One wrapper for uncertain outcomes.** `uncertainCardWrite` (`cardpublish.go:26`) keeps the error chain intact with `%w`. Callers' `errors.Is(ErrCardChanged / ErrNoChange)` checks still work, and an uncertain outcome is never reported as success or failure (ARCH-ORDER).
- **Lost acknowledgements are tested properly.** `loseResponses` (`recovery_proofs_test.go:21`) lets the real publish run against a real repository, then injects the lost response. That models the uncertainty itself rather than mocking it away (ARCH-MOCK).
- **Missing proofs show as UNPROVEN.** A claim with no test renders as "UNPROVEN (no test; treat as unknown)" (`contracts.go:146`), which meets the Done-when requirement that unproven guarantees are explicitly unknown.

## Critical
None.

## Important
1. **The catalog is wrong in legacy repositories (no issue tracker).** `catalog.go`, ARCH-PURPOSE.
   - The help section appears in every repository that uses the base layer, but it only describes tracker repositories.
   - The claim contract says "one compare-and-swap on the card … Never pushes main". In a legacy repository, `runLegacyClaim` (`legacymode.go:224+`) publishes to main through `newTrunkPublisher` and `UpdateMany`.
   - The change-code contract says "Never pushes" and "local only; rerun". In a legacy repository, `syncLegacyIssue` (`changecode.go:204-238`) commits and publishes.
   - set-status has a legacy path too (`setstatus.go:84`).
   - An agent in a downstream legacy repository would be told the wrong evidence and the wrong lost-response action.
   - **Fix:** say "in an issue-tracker repository (#252)" in each section header (`Section` in `contracts.go`), plus one line that legacy repositories are not covered and their recovery is unknown. Alternatively, add legacy wording to the claim, change-code and set-status contracts.
2. **The plan's Core concepts table contradicts the code.**
   - It lists `recovery.Section` / `Page()` / `JSON()` in `internal/recovery/render.go`. That file doesn't exist: `Section` and `wrap` are in `contracts.go`.
   - `JSON()` was dropped in the PQ-2 revision.
   - `Page()` is M2 work and the table doesn't say so.
   - **Fix:** a Revisions entry (see the end of this review).

## Minor
- The package doc in `contracts.go` (lines 1-6) still says the catalog renders through `({{RECOVERY}})`. That placeholder was dropped.
- Every section ends with "see `sdlc help recovery`", but that topic doesn't exist until M2. The pointer dangles at the M1 boundary; it's fine as long as M2 lands before merge.
- `TestCardPublishCallers` (`recovery_proofs_test.go:132`) has two gaps:
  - `reclaimEffect` is not in `allowed`, so the test won't notice if reclaim stops using the seam.
  - It only looks at variable declarations named `reclaimEffect`. Any other `var x = func(){ cardPublish(...) }` would get past the guard.
- The plan lists `TestUpdateMany_LostAcknowledgmentIsUncertainWithoutReplay` as lost-acknowledgement proof, but no catalog entry cites it.
- `TestLandingLeavesAReopenedCardAlone` passes only because the landing code filters on `status == codecomplete`; it doesn't exercise any generation guard. It's an accurate proof of "a reopened card is left alone". The stronger reopen-then-reclose case is already covered by `TestPublishFlipLeavesAReclosedGenerationAlone`.
- The command-tree walk is written twice: once in `attachRecoveryContracts`, once in the test. A small `walkCommands(root, fn)` helper would remove the duplicate (ARCH-DRY, trivial).

## Test coverage notes
- The contract test fails on:
  - a required verb with neither a contract nor an exemption;
  - an exemption that no longer matches a command;
  - a contract for something that isn't a command;
  - a missing help section;
  - a proof naming a test that doesn't exist (found by an AST scan).
- The implementer's Log says a mutation check was done: a removed contract and a contract for a non-command each turn the test red. I did not re-run that check.
- `Validate` has tests that feed it a deliberately broken contract for each rule.
- There are no tests for how legacy repositories behave under the contracts. That is consistent with Important #1: either declare legacy out of scope, or prove it.

## Architecture
- **ARCH-DRY: pass.** There is one seam, one uncertain-outcome message and one registry. The duplicated walk is the minor note above.
- **ARCH-PURE: pass.** `recovery` is pure data plus rendering, and its tests do no IO. The IO stays in `cardPublish`.
- **ARCH-PURPOSE: flag.** See Important #1 (contracts that are wrong outside tracker repositories). The search for other card-publication sites found only landing's `ChangeCard`, which is a separate path.
- **ARCH-MOCK: pass.** The lost-acknowledgement double wraps the real publish against a real git fixture.
- **ARCH-CONSTRAINTS: pass.** The attach pass runs on every `buildRoot` and renders 16 short sections, which costs microseconds and is negligible.
- **ARCH-SECURE: pass (N/A).** The change only parses local test sources and handles no secrets.
- **ARCH-ORDER: pass.** Uncertain outcomes are preserved, not converted to success or failure. The reruns are decided by the compare-and-swap on the card. The generation tests cover a close re-run after SHIP and a reopen.
- **ARCH-FUNERAL: pass.** The change creates nothing durable: the registry is static, and the tests use temporary fixtures.

## Notes for M2
- The `recovery.Example` steps should run in the scheduling-example test and render on the page from the same data, so the page and the test can't drift. If the legacy scope stays out, the page should say so.

## Plan revision recommendations
- Add a "## Revisions" entry: Core concepts — `Section`/`wrap` are in `contracts.go` (there is no `render.go`); `JSON()` was removed (PQ-2); `Page()` and `example.go` are M2. Also note whether the contracts cover tracker repositories only.

```findings
findings:
  - id: new
    severity: Important
    family: contract-scope-unstated
    title: |
      Recovery contracts describe tracker repositories only but render into every repository's help; claim, change-code and set-status are wrong in legacy repositories
    detail: |
      claim says "Never pushes main" and change-code says "Never pushes", but runLegacyClaim (legacymode.go) and syncLegacyIssue (changecode.go) publish to main. Either scope each Section to issue-tracker repositories (#252) and mark legacy as unknown, or add legacy wording to those contracts.
  - id: new
    severity: Important
    family: plan-table-drift
    title: |
      The plan's Core concepts table names render.go, Page() and JSON(), which do not exist (Section is in contracts.go; JSON was dropped in PQ-2)
    detail: |
      Add a Revisions entry so the table matches the code, and mark Page() and Example as M2.
  - id: new
    severity: Minor
    family: stale-doc-comment
    title: |
      The recovery package doc still describes the dropped {{RECOVERY}} placeholder
  - id: new
    severity: Minor
    family: dangling-help-pointer
    title: |
      Every M1 help section points at sdlc help recovery, which does not exist until M2
  - id: new
    severity: Minor
    family: caller-guard-gaps
    title: |
      TestCardPublishCallers does not require reclaimEffect and skips other var-declared function literals
  - id: new
    severity: Minor
    family: plan-proof-omitted
    title: |
      The plan's lost-acknowledgement proof TestUpdateMany_LostAcknowledgmentIsUncertainWithoutReplay is cited by no catalog entry
```

---

## Re-review — 2026-10-02T11:10:43-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 280 — Publish tested operation recovery contracts |
| repo | ariadne |
| issue file | workshop/issues/000280-operation-recovery-contracts.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 3b7315b7fde5083f036e377d41e8053193e52ef3..135d7d4b60c8a24171b811930f9ea40fe5fcc3ff |
| command | sdlc milestone-close --issue 280 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-02T11:10:43-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

All six round-1 findings are fixed, and I checked each one against the code and tests rather than the commit message. The one new finding is Minor and doesn't block.

**What was fixed:**
- **Legacy-repository scope (BR-1):** a single `recovery.Scope` constant is now rendered into every help section. `Section` writes the `Scope:` line at `contracts.go:135`, and `TestSectionMarksUnprovenClaims` checks that `Scope:`, `legacy` and `unknown` are present. That test fails if the scope line is removed.
- **Caller guard (BR-5):** `reclaimEffect` is now in the allowed set, so the "no longer publishes" loop requires it. The guard now looks at every package-level initializer instead of one hard-coded name.
- **Uncited lost-acknowledgement test (BR-6):** the catalog now cites the gitx test, and that test exists.
  - `cardPublish` → `UpdateCardWithTrailers` → `UpdateManyPrepared` runs through the same `updateMany` core that the gitx test drives, so the citation is accurate.

**What I ran:**
- `go test ./cmd/sdlc/internal/recovery/` passes.
- `go test ./cmd/sdlc -run 'TestRecoveryContractsAreProven|TestCardPublishCallers|TestNoCommandLongHasSurvivingPlaceholder|TestEveryFlagAppearsInItsHelp'` passes.
- The gitx lost-acknowledgement test passes.
- `sdlc claim --help` shows the scope line.
- The git commands only worked with the sandbox disabled.

1. **Strengths**
   - The scope text is defined once (`contracts.go:53`) and shown through the one `Section` renderer. It isn't copied into each catalog entry, which would follow ARCH-DRY.
   - The caller guard now encodes a general rule ("any initializer that calls `cardPublish` must be allowed") instead of naming one exception (`recovery_proofs_test.go:136-170`). It also works in both directions: the allowed callers must still use the seam.
   - The lost-acknowledgement proof chain goes all the way down to the push. `TestUpdateMany_LostAcknowledgmentIsUncertainWithoutReplay` checks that a lost push response is reported as uncertain, that the change is not prepared a second time (`calls == 1`), and that the push really reached origin.
   - The plan's Revisions entry clearly replaces the outdated table rows and records that M2 owns the `sdlc help recovery` page.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The atlas page `atlas/workflow/recovery-contracts.md` never says that the contracts only cover issue-tracker repositories. It describes the catalog as the rulebook with no limit on where it applies. **This is the 2nd finding in family `contract-scope-unstated`.**
     - The rule that covers both: every surface that describes the contracts states their scope, and generated surfaces take it from `recovery.Scope`. Today those surfaces are the help sections (done), the M2 `sdlc help recovery` page (should render `Scope`, not retype it) and the atlas (hand-written, so add one line there).
     - Fix the class in M2: render `Scope` in `Page()`, and add the scope line to the atlas in the same round.

5. **Test coverage notes**
   - The scope assertion checks for words, not the whole line; that's fine for M1.
   - The caller guard still credits any `cardPublish` use inside a function, closures included, to that function. That's acceptable.
   - Legacy repositories have no recovery tests. That's consistent with their being declared out of scope.

6. **Architecture**
   - **ARCH-DRY:** pass. Scope is single-sourced. The command-tree walk is still written twice (once in `attachRecoveryContracts`, once in the test); that was already noted in round 1 and is trivial.
   - **ARCH-PURE:** pass. The recovery package is pure data and rendering.
   - **ARCH-PURPOSE:** pass. The legacy gap is now explicitly marked unknown, which fits the Done-when line "unproven guarantees are explicitly unknown". The scheduling example and the help page remain M2 work, as planned.
   - **ARCH-MOCK:** pass. The lost-acknowledgement tests run the real git push against a real fixture.
   - **ARCH-CONSTRAINTS:** pass. Negligible cost.
   - **ARCH-SECURE:** N/A. The change only parses local Go sources and handles no secrets.
   - **ARCH-ORDER:** pass. An uncertain publish stays uncertain and is never retried automatically.
   - **ARCH-FUNERAL:** pass. Nothing durable is created; the catalog is static.

7. **Plan revisions:** none beyond what was added. In M2, record that `Page()` renders `recovery.Scope`.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      recovery.Scope rendered by Section into every help section (contracts.go:135); TestSectionMarksUnprovenClaims asserts Scope/legacy/unknown; verified in sdlc claim --help.
  - id: BR-2
    disposition: addressed
    note: |
      Plan Revisions entry supersedes the table rows: no render.go, JSON dropped (PQ-2), Page/Example are M2.
  - id: BR-3
    disposition: addressed
    note: |
      Package doc now names attachRecoveryContracts; no placeholder reference remains.
  - id: BR-4
    disposition: addressed
    note: |
      Revisions records that sdlc help recovery lands on this branch in M2 before anything ships; acceptable at an intra-branch milestone.
  - id: BR-5
    disposition: addressed
    note: |
      reclaimEffect added to allowed (so required by the not-seen loop); guard inspects every ValueSpec initializer.
  - id: BR-6
    disposition: addressed
    note: |
      catalog.go:25 cites the test; it exists in internal/gitx and passes; it shares the updateMany core with UpdateCardWithTrailers.
findings:
  - id: new
    severity: Minor
    family: contract-scope-unstated
    title: |
      atlas/workflow/recovery-contracts.md presents the contracts as the rulebook without their issue-tracker-only scope
    detail: |
      Repeat of the family. Rule: every surface describing the contracts states the scope; generated surfaces (help sections, M2 Page) render recovery.Scope, and the hand-written atlas page carries one scope line. Fix the class in M2 by rendering Scope in Page() and adding the line to the atlas.
```
