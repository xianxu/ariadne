# Boundary Review — ariadne#279 (milestone M1)

| field | value |
|-------|-------|
| issue | 279 — Expose authoritative workflow observations for agents |
| repo | ariadne |
| issue file | workshop/issues/000279-workflow-observables.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 11ce13fafae075a7878bc68aadbe2db135975813..39092042fdf95ba327d8849640acc5f70da68dac |
| command | sdlc milestone-close --issue 279 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-02T00:11:32-07:00 |
| verdict | REWORK |

## Review

I've confirmed two real bugs. Writing up the verdict now.

```verdict
verdict: REWORK
confidence: high
```

M1's pure core is solid. `internal/observe` keeps read quality (`state`) separate from value fields, the strict JSON refuses unknown and duplicate keys and validates in both directions, and the real-git tests cover the parked-slot, stale-tracker and no-mutation cases from the plan. I ran `go vet`, the observe package tests, and `TestObserve*`, `TestIssueShow*` and `TestEveryFlag*`; all pass. One defect still blocks the boundary. `issueShowRepo` (`cmd/sdlc/issue.go:626`) finds the repository from the current directory (`.`), but the old text path finds it from `--issues-dir`. I reproduced two failures from this with a built binary:
- **Outside any git repository**, `sdlc issue show 5` now dies with `. is not inside a repository`. The base version read the details-only directory through `LoadRecords(nil, …)`.
- **With `--issues-dir` in a non-git temp dir, run from inside ariadne**, the text view printed the temp issue (`status: open`). Under it, the "observations" block described **ariadne's own #5**: `status: done`, `landing: landed, archived at workshop/history/issues/000005-issue-sync-makefile-target.md`. That is a confident answer about the wrong issue.

The same root cause makes the existing test `TestRunIssueShow_HeadersNotBodies` run a real `PreferFresh` tracker fetch against the developer's ariadne checkout.

1. **Strengths**
   - `assemble.go:62-69`: every section derives from the tracker read through `tracked`/`valued`. Stale stays stale with its reason, and unknown is never collapsed into absent. `FuzzAssemble` checks this for arbitrary card bytes.
   - `json.go:44-56`: the "relation/outcome is set exactly when the read yielded a value" invariant enforces the PQ-3 rule in code, not just in prose.
   - `json.go:104`: `rejectDuplicateKeys` closes the `encoding/json` last-key-wins gap. The test checks that every bad fixture actually changes the document.
   - `records.go`: `FetchErr` is a minimal, well-placed change, so a stale answer carries its reason.
   - `observe_test.go:44`: the parked-slot test snapshots `status` and `for-each-ref` on both checkouts, which really pins "queries do not mutate".

2. **Critical findings**
   - `cmd/sdlc/issue.go:626-640` (`issueShowRepo`): without `--repo`, find the root from the issues directory, as `loadIssueRecords` does. With no repository, keep the old details-only behaviour: skip the observation, or emit it with the tracker `absent`. Never die.
   - Add regression tests for both cases:
     - `issue show` in a non-git dir succeeds;
     - `--issues-dir` pointing at another repository observes that repository, not the current directory's.

3. **Important findings**
   - **Golden fixture missing.** The M1 Plan's first item promises "a golden fixture" for the strict JSON, and none is in the diff. The round-trip test compares the document with itself, so it can't catch a wire-format change to `schema_version: 1`. Add a golden file under `internal/observe/testdata`.
   - **A unit test reaches real repository state (ARCH-SECURE).** Because of the Critical, `TestRunIssueShow_HeadersNotBodies` now fetches the real tracker in whatever checkout runs `go test`, and updates its remote-tracking ref. It goes away once the root follows the issues directory. Make sure no other test with an absolute temp `IssuesDir` keeps the cwd fallback.

4. **Minor findings**
   - `observe.go:47`: a `rev-parse` failure on `rs.Ref` is silently dropped, leaving the tracker `present` with an empty `ref`. Put it into `TrackerErr`, or record it.
   - `Validate` checks `landing.outcome` against its enum but not `relation` or `claimant_worktree`.
   - `text.go:48`: the `map[bool]string{…}[…]` idiom is obscure; a plain `if` reads better.
   - `issue.go:16`: the `internal/observe` import sits inside the stdlib group (goimports style).
   - `assemble.go:53`: the M2 sections are emitted as `unknown` with "not observed by this build". This is fine for one milestone, but it overloads `unknown`, which means "the read failed". Make sure M2 removes it.
   - `assemble.go:156`: `in.Me == nil` gives `FateUnknown` even when the recorded machine could be compared. That is acceptable.

5. **Test coverage notes**
   - The Assemble table covers relation × claimant worktree × freshness well.
   - Missing coverage:
     - no golden file;
     - no test for `issue show` outside a repository, or with a foreign `--issues-dir` (the bug shipped here);
     - no subprocess-count budget assertion. That is reasonable to defer to M2, where the fan-out grows.
   - The `--repo` test passes a path inside `workshop/`, which is good: it exercises "containing this path".

6. **Architectural notes**
   - **ARCH-DRY: pass.** Judgments call `CardClaimant`, `MatchClaimant`, `CardCompletion` and `ParseCard`; nothing is restated.
   - **ARCH-PURE: pass.** `Assemble` has no IO, and the collector is thin.
   - **ARCH-PURPOSE: pass for M1 scope**, apart from the missing golden.
   - **ARCH-MOCK: pass.** Real-git fixtures use a local bare origin, and the tests use the existing harness.
   - **ARCH-CONSTRAINTS: pass.** The CLI's records scope (`main.go:91`) folds the text path's two loads into one fetch when the roots match. After the root fix, both loads must still hit the same scope key.
   - **ARCH-SECURE: flag.** Wrong-repo provenance gives a fabricated observation that downstream code would read as evidence, and a unit test reaches real repository state (see above).
   - **ARCH-ORDER: pass.** This is a single-shot read that holds no state between events, and `observed_at` plus the tracker ref anchor the snapshot.
   - **ARCH-FUNERAL: pass.** It creates nothing durable. The only side effect is the documented remote-tracking ref update.
   - For M2: derive the branch name from the details stem rather than the card path's basename (`assemble.go:163`), so untracked repositories work the same way.

7. **Plan revision recommendations**
   - If the golden is deliberately deferred to M2, add a Revisions entry saying so. Otherwise none.

```findings
findings:
  - id: new
    severity: Critical
    family: repo-root-from-wrong-anchor
    title: |
      issue show resolves the observed repository from cwd, not the issues dir: dies outside a repo and observes the wrong repo with --issues-dir
    detail: |
      issueShowRepo (cmd/sdlc/issue.go:626) uses repoRootOf("."). Reproduced: `sdlc issue show 5` in a non-git dir now dies with ". is not inside a repository" (the base version worked via LoadRecords(nil)). Run from ariadne with --issues-dir pointing at a temp dir, it printed the temp issue (open) and then ariadne's own #5 observation (done, landed). Derive the root from the issues dir unless --repo is given, degrade without a repository, and add regression tests for both.
  - id: new
    severity: Important
    family: plan-deliverable-dropped
    title: |
      M1 promises a golden fixture for the schema_version 1 JSON; none is in the diff
    detail: |
      The round-trip test compares the document with itself, so a wire-format drift of the v1 contract goes undetected. Add internal/observe/testdata/*.golden.
  - id: new
    severity: Important
    family: test-reaches-real-state
    title: |
      TestRunIssueShow_HeadersNotBodies now does a real tracker fetch against the developer's checkout
    detail: |
      With an absolute temp IssuesDir and cwd inside ariadne, the deferred observation runs loadIssueRecordsAt(PreferFresh) on the real repository and updates its remote-tracking ref during go test (ARCH-SECURE). Fixed by the root-anchor fix; check that no other test keeps the cwd fallback.
  - id: new
    severity: Minor
    family: silent-error-swallow
    title: |
      rev-parse failure on the tracker ref is dropped, leaving tracker present with an empty ref
  - id: new
    severity: Minor
    family: contract-enum-validation
    title: |
      Validate checks landing.outcome against its enum but not assignment.relation or claimant_worktree
  - id: new
    severity: Minor
    family: state-semantics-overload
    title: |
      M2 sections are emitted as unknown ("not observed by this build"), overloading the read-failed meaning; remove in M2
```

---

## Re-review — 2026-10-02T00:17:40-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 279 — Expose authoritative workflow observations for agents |
| repo | ariadne |
| issue file | workshop/issues/000279-workflow-observables.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 11ce13fafae075a7878bc68aadbe2db135975813..6140631c21a02a01fd5199f0953ce4dfe1a432d9 |
| command | sdlc milestone-close --issue 279 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-02T00:17:40-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The rework fixes the three blocking findings from round 1, and I checked each one in the code rather than relying on the commit message. `issueShowRepo` (`cmd/sdlc/issue.go:629`) now takes the repository from the issues directory it was given, or from `--repo`. It no longer uses the process cwd. An issues directory outside any repository falls back to the details-only read (root `""`). `TestObserveAnchorsOnTheIssuesDir` tests both halves: running outside a repository, and pointing at a loose issues directory from inside a tracked repository, with the tracker tip and remote refs checked to be unchanged. Both halves fail against the old `repoRootOf(".")` code. A v1 golden fixture now pins the wire format. `TestRunIssueShow_HeadersNotBodies` now runs on a temp directory that isn't a git repository, so it no longer fetches against the developer's checkout. `go test ./cmd/sdlc/internal/observe/` and `go test ./cmd/sdlc/ -run 'TestObserve|TestRunIssueShow'` both pass. What's left is Minor:
- Two of the Minor fixes have no test that would fail without them.
- The enum validation closed only the fields that were named, not the rest of that class.
- The M2 placeholder is intentionally still there.

1. **Strengths**
   - The issues-directory anchor is a simple rule that's easy to explain. The new `workshop/lessons.md` entry states it as a general rule ("resolve from the path given, never the cwd"), not as a one-off fix.
   - `TestObservationGolden` compares against a file that was written separately. It also decodes the golden through the strict decoder, so the fixture is itself checked against the contract.
   - `Tracker.RefError` keeps a failed rev-parse separate from both the read state and the ref (`assemble.go:79-83`), so a failure no longer leaves an empty ref with no reason.
   - `Validate` now requires `claimant_worktree` to be set exactly when a claimant is (`json.go:66`), which is a real cross-field rule.
   - The stale-tracker test in `TestObserveStaleTrackerSaysSo` checks the error text, the stale state and the value, so it follows the "read quality separate from value" contract all the way through.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - BR-4: the `TrackerRefErr` path is wired up, but no test reaches it, at either the assembler level or the collector level.
   - BR-5 (family `contract-enum-validation`): `relation` and `claimant_worktree` are now validated, but `Authority` on every section is not. Neither is the requirement that each section carries its own fixed authority (card, assignment, completion and landing are `tracker`; branch and checkpoints are `committed`; workspaces are `worktree`). Neither of these is covered by a test either.
   - BR-6: the "not observed by this build" placeholders are now also baked into the golden fixture. M2 has to update both together.

5. **Test coverage**
   - The real-git tests cover the M1 plan items: the parked slot, the stale tracker, unchanged local state, landed, and `--repo`.
   - The fuzz test covers malformed card bytes only. The ledger and details inputs it is meant to cover arrive in M2.
   - The plan's subprocess-count budget test isn't written yet. It isn't an M1 checklist item, so carry it into M2.

6. **Architecture**
   - **ARCH-DRY: pass.** Judgments reuse the existing `issue.*` parsers.
   - **ARCH-PURE: pass.** `internal/observe` does no IO, and the collector is a thin layer.
   - **ARCH-PURPOSE: pass for M1.**
   - **ARCH-MOCK: pass.** Tests run on the real-git fixture harness.
   - **ARCH-CONSTRAINTS: pass.** In text mode, the records scope reuses the same root+dir key, so there is one fetch.
   - **ARCH-SECURE: pass.** BR-3 is closed, and strict decoding rejects unknown and duplicate keys.
   - **ARCH-ORDER: pass.** This is a single-shot read that holds no state between events.
   - **ARCH-FUNERAL: pass.** It creates nothing durable. The only side effect is the existing remote-tracking ref update.
   - For M2:
     - Make each enum type validate itself, so new sections can't skip it.
     - With `--repo` plus an absolute `--issues-dir` in a different repository, the tracker of one repository is paired with the details of another. Make that combination an error.

7. **Plan revisions:** none needed. The plan still matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      issueShowRepo anchors on the abs issues dir or --repo; TestObserveAnchorsOnTheIssuesDir fails under the old cwd anchor (both the non-repo and the wrong-repo halves).
  - id: BR-2
    disposition: addressed
    note: |
      testdata/observation-v1.golden.json, compared byte-for-byte by TestObservationGolden and strictly decoded.
  - id: BR-3
    disposition: addressed
    note: |
      newTestDirs is a non-git temp dir, so root is empty and nothing is fetched; no other test calls runIssueShow with a cwd-relative dir inside the real checkout.
  - id: BR-4
    disposition: not-addressed
    note: |
      TrackerRefErr/RefError is wired up, but no test reaches the rev-parse failure path.
  - id: BR-5
    disposition: not-addressed
    note: |
      relation and claimant_worktree are now checked, but there is no rejection test, and Authority on every section, plus the fixed authority per section, is still unvalidated. Rule: every enum-typed contract field validates itself, with one table test covering all of them.
  - id: BR-6
    disposition: not-addressed
    note: |
      Still present by design until M2, and now also baked into the v1 golden fixture; M2 must update both.
```
