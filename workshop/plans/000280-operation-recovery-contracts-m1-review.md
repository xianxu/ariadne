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
