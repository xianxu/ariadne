# Boundary Review — ariadne#280 (milestone M2)

| field | value |
|-------|-------|
| issue | 280 — Publish tested operation recovery contracts |
| repo | ariadne |
| issue file | workshop/issues/000280-operation-recovery-contracts.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 9933a3509db7ec9ff31563ba94c0a5b4701fd9bb..378eb85eb474d19de1d3964a1f14da46dde759db |
| command | sdlc milestone-close --issue 280 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-02T11:21:18-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**M2 delivers what it promised.** `sdlc help recovery` is a cobra help topic with no Run. Its scope line, class list, verb table and example are all rendered from the `internal/recovery` registry. The scheduling example is executed by `TestSchedulingExampleRuns` from the same `recovery.Example` data that renders it. A convergent-retry step is delivered twice, and the coordinator checks every step from a separate worktree. README, the atlas page and the process manual are all updated; the process manual picks the new helptext file up automatically.

I built the binary and rendered `sdlc help recovery`, `sdlc recovery`, `sdlc issue recovery [reconcile] --help` and `sdlc process-manual`. I ran `go test ./cmd/sdlc/internal/recovery/` and `go test ./cmd/sdlc -run 'TestSchedulingExampleRuns|TestRecoveryContractsAreProven|TestEveryFlagAppearsInItsHelp|TestNoCommandLongHasSurvivingPlaceholder|TestProcessManual'`. Everything passes; `TestSchedulingExampleRuns` takes 9.0s.

One Important finding remains: the plan's Core-concepts table still names files and symbols that M2 did not build, and no M2 Revisions entry records it. This is a repeat of the `plan-table-drift` family. Fixing it is a documentation edit, so it does not block the code.

**1. Strengths**
- **One source of data.** `cmd/sdlc/internal/recovery/page.go:61-83` is a single data table. It is rendered by `ExampleText` and run by `cmd/sdlc/recovery_example_test.go:67-108`. A coordinator step that is not `issue show` fails the test (`:89-91`), so the example cannot quietly grow an untested step kind.
- **Duplicate delivery is driven by the registry.** Convergent-retry steps run twice because the registry says they are convergent-retry (`recovery.For(args[0]).Class`, `recovery_example_test.go:79-82`), not because of a hand-kept list.
- **The "tracker unreachable" case is genuinely exercised.** The test renames the origin and checks that `tracker.state = stale`; it does not just describe the case.
- **`Lookup` is a small pure function with its own test.** It handles found and missing keyed elements, non-map intermediate values and number formatting (`contracts_test.go:53-66`).
- **The help topic needs no new command file.** It is added in `main.go:166` and filled from `renderLong` placeholders. `TestNoCommandLongHasSurvivingPlaceholder` still guards it, and the process manual includes it automatically.

**2. Critical findings:** none.

**3. Important findings**
- **The plan's Core-concepts table no longer matches the tree** (`workshop/plans/000280-operation-recovery-contracts-plan.md:36-63`). **This is the 2nd finding in family `plan-table-drift`.** The rows that are still wrong:
  - `recovery.Page()` in `render.go`: neither exists. The topic is `helptext/recovery.md` plus `ScopeText`, `ClassesText`, `Table` and `ExampleText` in `page.go`.
  - `recovery.Example` in `example.go`: it lives in `page.go`.
  - `sdlc help recovery` in `cmd/sdlc/recoverycmd.go`: that file does not exist; the topic is registered inline in `buildRoot` (`main.go:166`).
  - The new `Lookup`, `Step`, `Expect` and `Actor` are not listed at all.

  BR-2 even promised "`Page()` and `Example` are M2 deliverables", and M2 shipped a different shape. Earlier rounds patched individual rows. The rule that covers every instance: *each boundary sweeps every Core-concepts row against the tree (the path exists, the symbol exists) and records every divergence in one Revisions entry at that boundary, or rewrites the table to the current state.* Apply that sweep now (see section 7).

**4. Minor findings**
- **Inaccurate claim about the harness flags** (`recovery_example_test.go:19-20` and `atlas/workflow/recovery-contracts.md:52-53`). Both say the harness flags "stand in for model judges and estimates" and are "the only additions". But `--worktree=no` and `--no-atlas` bypass a gate or choose a layout; they are neither judges nor estimates. The text should describe the flags as they are. Family `stale-doc-comment`; this is its 2nd instance, and the rule is that a comment describing a set must match the set it describes.
- **The duplicate-delivery check misses two-word verbs** (`recovery_example_test.go:80`). `recovery.For(args[0])` looks up only the first word. A future recipient step such as `sdlc issue sync` would resolve to `issue`, find no contract, and quietly be delivered only once. Match the longest verb prefix instead. Family `caller-guard-gaps`; this is its 2nd instance, and the rule is that a guard derived from the registry must resolve the same verb identity the registry uses (multi-word `Verbs`).
- **The unreachable tracker is only restored at test cleanup** (`recovery_example_test.go:92-97`). Any step added after step 10 would silently run with no tracker. Restore the origin right after that step's observation.
- **Cross-link is one-way.** The plan says `sdlc help recovery` and `sdlc issue recovery` cross-link. `issue recovery reconcile --help` does point back to the topic, but the parent `sdlc issue recovery --help` does not.
- **The 30-second heuristic is not general guidance.** The plan asks the guidance to say that "roughly 30 s is a revisit heuristic, not proof of loss". It appears only inside the "otherwise" text of example step 3, not in the AGENT GUIDANCE list.
- **The example's actor routing has no default.** A step whose Actor is neither coordinator nor recipient falls through the `switch` silently. Add a `default: t.Fatalf`.

**5. Test coverage notes**
- The example test is a real end-to-end run against real git with a stubbed judge, and it checks every expectation from the data.
- It covers only one ordering: each recipient step finishes before the coordinator looks, apart from the single "before claim" probe. Lost-acknowledgement and race cases stay with M1's dedicated tests. That is acceptable for a documentation example, but the help's "so it cannot drift" applies to the commands and expectations, not to the "otherwise" branches.
- `TestPageRenderingCoversEveryContractAndStep` only checks that substrings are present. It is a reasonable smoke test.

**6. Architectural notes for upcoming work**
- **ARCH-DRY: pass.** `page.go` reuses `wrap` and the catalog. The repeated `strings.Repeat(" ", 19)` could become a constant.
- **ARCH-PURE: pass.** `page.go` and `Lookup` have no IO and are tested directly.
- **ARCH-PURPOSE: pass.** Every part of the topic derives from the registry, and the example is executed rather than restated.
- **ARCH-MOCK: pass.** The test runs against the real-git fixture, and the judge sits behind the existing `stubJudge` seam.
- **ARCH-CONSTRAINTS: pass.** This is help text plus a 9-second test.
- **ARCH-SECURE: pass.** `Lookup` returns false for any malformed path or shape and never makes up a value.
- **ARCH-ORDER: pass.** The help topic holds no state between events because it is static rendered text. The single-ordering limit of the example test is noted in section 5.
- **ARCH-FUNERAL: pass.** M2 creates nothing durable because the topic is embedded text and the test uses `t.TempDir` fixtures.
- **For pair#362:** `Lookup` keys lists by `boundary` only, so a second keyed list in the observation schema would need a key-field parameter. This is fine as long as `reviews` is the only keyed list.

**7. Plan revision recommendations**
- Add `## Revisions` → "2026-10-02 — M2 implementation" with these points:
  - The topic is `helptext/recovery.md` plus the `{{RECOVERY_SCOPE|CLASSES|TABLE|EXAMPLE}}` placeholders, rendered by `ScopeText`, `ClassesText`, `Table` and `ExampleText` in `recovery/page.go`. There is no `Page()` and no `render.go`.
  - `Example`, `Step`, `Expect`, `Actor` and `Lookup` live in `page.go`, not `example.go`.
  - The topic is registered inline in `buildRoot`; there is no `recoverycmd.go`.
  - The coordinator is a detached worktree, and the recipient is the fixture's main checkout.

```findings
findings:
  - id: new
    severity: Important
    family: plan-table-drift
    title: |
      Core-concepts table still names Page()/render.go, example.go and recoverycmd.go; M2 built page.go + helptext placeholders + inline buildRoot topic
    detail: |
      2nd finding in plan-table-drift. Rule covering all instances: each boundary sweeps every Core-concepts row against the tree (path and symbol exist) and records all divergences in one Revisions entry, or rewrites the table to current reality. Rows wrong now: Page() and render.go (absent), Example in example.go (page.go), recoverycmd.go (absent; main.go:166); Lookup/Step/Expect/Actor unlisted.
  - id: new
    severity: Minor
    family: stale-doc-comment
    title: |
      Harness-flag comment and atlas claim the flags only stand in for judges and estimates, but --worktree=no and --no-atlas are neither
    detail: |
      recovery_example_test.go:19-20 and atlas/workflow/recovery-contracts.md:52-53. Rule: a comment describing a set must match the set it describes.
  - id: new
    severity: Minor
    family: caller-guard-gaps
    title: |
      Duplicate-delivery class lookup uses recovery.For(args[0]), so a multi-word verb step (issue sync) would silently be delivered once
    detail: |
      recovery_example_test.go:80. Resolve the longest verb prefix against Contract.Verbs, the same identity the registry uses.
  - id: new
    severity: Minor
    family: example-step-ordering
    title: |
      TrackerUnreachable restores the origin only at t.Cleanup; any step appended after it runs with no tracker
    detail: |
      recovery_example_test.go:92-97; restore immediately after that step's observation.
  - id: new
    severity: Minor
    family: plan-claim-partially-delivered
    title: |
      Plan's two-way cross-link and the general 30 s revisit heuristic are only partly delivered
    detail: |
      sdlc issue recovery --help (parent) does not point to sdlc help recovery (reconcile does). The 30 s heuristic appears only in example step 3, not in AGENT GUIDANCE.
  - id: new
    severity: Minor
    family: caller-guard-gaps
    title: |
      Example actor switch has no default; a step with an unknown Actor is silently skipped
    detail: |
      recovery_example_test.go:69; add a default t.Fatalf.
```

---

## Re-review — 2026-10-02T11:24:03-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 280 — Publish tested operation recovery contracts |
| repo | ariadne |
| issue file | workshop/issues/000280-operation-recovery-contracts.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 9933a3509db7ec9ff31563ba94c0a5b4701fd9bb..eb280a5191d809a96b6b3866d5377ed315623cb6 |
| command | sdlc milestone-close --issue 280 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-02T11:24:03-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

M2 delivers what the plan asked for. `sdlc help recovery` renders the scope, the classes, the agent guidance (now including the 30 s revisit heuristic), one entry per contract and the scheduling example, all from the `recovery` registry. `TestSchedulingExampleRuns` runs the same `recovery.Example` data on a real-git fixture. The coordinator works in a separate worktree, steps of the convergent-retry class are delivered twice, and one step runs with the tracker remote unreachable. I ran `go test ./cmd/sdlc/internal/recovery/` and `go test ./cmd/sdlc/ -run 'TestSchedulingExampleRuns|TestEveryFlagAppearsInItsHelp|TestRecovery'`; both pass. I also rendered `sdlc help recovery` and checked that `sdlc process-manual` picks up `helptext/recovery.md`, which covers the plan's process-manual item. All six earlier findings are fixed in the tree. One new Minor remains: the plan-text drift family is fixed in the tables but not in the prose around them. Nothing blocks SHIP.

**1. Strengths**
- One source for documentation and test. `recovery.Example` (`cmd/sdlc/internal/recovery/page.go:62`) is both rendered by `ExampleText` and executed by the test. The help says "Executed step by step by TestSchedulingExampleRuns", and that claim is backed by the test.
- The example tests ordering problems the caller cannot block (ARCH-ORDER). It checks the card before the claim (a lost or unread message), sends each convergent-retry command twice (a duplicate message), and observes with the tracker unreachable (expects `stale`). `contractFor` (`recovery_example_test.go:139`) finds a command's class through the registry's own verb identity, not a hand-written list.
- `Lookup` (`page.go:106`) is pure. It returns false whenever a path doesn't resolve, so a missing or wrong field fails the expectation and is never read as a match. It is tested without IO, including the negative paths (`contracts_test.go:53`).
- The harness flags are kept in one documented map (`recovery_example_test.go:23`). The atlas lists the same flags (`atlas/workflow/recovery-contracts.md:52`).

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- Prose drift around the Core-concepts tables (plan-table-drift, 3rd instance; details in the findings block):
  - `workshop/plans/000280-operation-recovery-contracts-plan.md:44` lists a Contract field `Verb`, but the struct field is `Verbs []string` (`contracts.go:64`).
  - `:53` says the source feeds "the JSON", but PQ-2 dropped JSON output.
  - The issue's ticked M1 row still says "rendered via `{{RECOVERY}}`".
  - Rule: the boundary sweep from the new lesson should cover every symbol named anywhere in the Core-concepts section and the Plan rows, not just the table cells.
- README: the line "Older binaries refuse claimed" now breaks early (`README.md:38`). Cosmetic.

**5. Test coverage**
- The plan's four example phases are all exercised, plus the "otherwise" case before the claim.
- `TestPageRenderingCoversEveryContractAndStep` checks that every contract, step and class appears in the rendered page.
- An actor value the test doesn't know now fails (`default: t.Fatalf`).
- The tracker origin is restored right after its step, and `t.Cleanup` remains as a fallback.

**6. Architecture**
- ARCH-DRY: pass. Example text and test share one data source, and classes are looked up through the registry.
- ARCH-PURE: pass. `page.go` is pure; IO lives only in the test harness.
- ARCH-PURPOSE: pass. The Done-when scheduling example covers claim, branch/activity and gate progress, and the guidance's scope is stated.
- ARCH-MOCK: pass. The test runs real git fixtures with a stubbed judge behind the existing seam.
- ARCH-CONSTRAINTS: N/A. These are static help strings plus a test-only fixture.
- ARCH-SECURE: pass. `Lookup` refuses to resolve on malformed input instead of guessing a value.
- ARCH-ORDER: pass. Duplicates and an unreachable tracker are injected deterministically from the step data.
- ARCH-FUNERAL: pass. Nothing durable is created; the origin rename is undone at once, and `t.Cleanup` is a fallback.
- Upcoming work: pair#362 will consume this page. If consumers start parsing the example, `Lookup`'s `\w+` index key will need widening for boundary names containing hyphens.

**7. Plan revisions**
- Add one Revisions line: "Core-concepts prose: `Verb` → `Verbs`; 'the JSON' removed (PQ-2); the issue's M1 row now says the sections are appended by `attachRecoveryContracts`."

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      Both Core-concepts tables now match the tree: contracts.go holds Section/Scope/Validate/For/wrap, page.go holds Example/Step/Expect/Actor/Lookup and the renderers, attachRecoveryContracts is at main.go:197, cardpublish.go exists; a Revisions entry records the sweep and a lesson states the rule.
  - id: BR-9
    disposition: addressed
    note: |
      recovery_example_test.go:19-22 and atlas lines 52-53 now list judges, estimate, actual, worktree and atlas, which matches the map's set.
  - id: BR-10
    disposition: addressed
    note: |
      contractFor (recovery_example_test.go:139) resolves the longest verb prefix through recovery.For; reachable from the recipient branch at :80.
  - id: BR-11
    disposition: addressed
    note: |
      Origin is renamed back immediately after observeJSON (recovery_example_test.go:101-105); Cleanup remains only as a fallback.
  - id: BR-12
    disposition: addressed
    note: |
      issuerecovery.go Long now points to `sdlc help recovery`; AGENT GUIDANCE in helptext/recovery.md carries the 30 s revisit heuristic (verified in rendered output).
  - id: BR-13
    disposition: addressed
    note: |
      default: t.Fatalf("unknown actor") added at recovery_example_test.go:113.
findings:
  - id: new
    severity: Minor
    family: plan-table-drift
    title: |
      Core-concepts prose and the issue's M1 row still name superseded symbols (Contract field Verb, "the JSON", the RECOVERY placeholder)
    detail: |
      This is the 3rd finding in family plan-table-drift. The tables were fixed, but plan line 44 says the field is `Verb` (code: `Verbs`), line 53 says the source feeds "the JSON" (dropped in PQ-2), and the issue's M1 row says "rendered via {{RECOVERY}}" (now attachRecoveryContracts). Rule: the per-boundary sweep in the new lesson covers every symbol named in the Core-concepts section and the Plan rows, not only the table cells. Grep each backticked identifier against the tree and fix all misses in one Revisions entry.
```
