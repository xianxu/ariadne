# Issue-backed known-flake landing Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** When a recorded suite (`sdlc local-check run`, or `sdlc merge`'s integration suite) fails only on tests that an open flake issue names, sdlc reruns just those tests on the same head and records the pass with the issue link. Any other failure still blocks, and a blind full rerun can no longer turn an unexplained failure green.

**Architecture:**
- **Registry:** the registry is the flake issues themselves. An open issue whose frontmatter lists `flake_tests:` registers those tests. The list is read from origin/main's tree, with the issue's status from its tracker card, so a branch cannot excuse its own failure, and closing the issue unregisters the tests.
- **Classifier:** a pure classifier maps an attempt's failures, the registry, a targeted rerun's result and the main baseline to one outcome.
- **Gate:** `runLocalCheck`, shared by `local-check run` and merge, executes that outcome's effects. The test-stats attempt store keeps every attempt and links each one to its issue.

**Tech Stack:** Go standard library; `go test -run` for the targeted rerun; the existing test-stats store, tracker records, GitHub status seam (`ghCaller`) and its stateful fake (`ciFakeGH`).

---

## Decisions for the hub/operator before code

1. **D1 — what counts as evidence (amends #390's approved rule 2).** #390 side doc 02 and the tl-worker help say a known flake needs *a passing full-suite rerun on the same head*. The dispatch asks to rerun *just the registered failing tests*. **Proposed:** the targeted rerun counts. The failed attempt already ran every other test on the exact head and they all passed, and the rerun passes the rest, so every test has passed once on that head. The combined evidence is attempt 1 (everything except the registered tests passed) plus attempt 2 (the registered tests passed). One passing exact-head suite remains the invariant, as the union of the two. This takes minutes, not a 10-minute suite plus a 20-50 minute queue. The tl-worker help's "one unchanged rerun of the full suite" changes to match.
2. **D2 — accepting without a rerun ("by policy").** **Proposed:** a field on the suite in `.sdlc/local-checks.json`, `"known_flakes": "rerun"`, with values `rerun` (the default), `accept` or `off`. That file is under local-gates, so only the operator changes it. `accept` records the pass with the issue link and no rerun, which is weaker evidence chosen on purpose.
3. **D3 — the daily TL view (Done-when 3).** #449 needs only the gate. **Proposed:** in this issue, a local view, `sdlc flake list`. It shows:
   - the registry (test, issue, owner, age);
   - every recent same-head fail→pass attempt that is not registered;
   - every unexplained failure older than a day.

   Each item comes with its attempt id and log. CI failure classes become a separate follow-up issue, "test economy: CI failures in the daily flake ledger". Until then, `flake list` prints CI as `not covered` and never as zero, so absent data never reads as success.
4. **D4 — where the registry lives.** **Proposed:** in issue frontmatter (`flake_tests: [<pkg>.<Test>, …]`), not a `.sdlc/` file. A `.sdlc/` list is local-gates guarded, so each daily change would need the operator's label. An issue file is published by the TL with `sdlc issue publish`, so the issue is the single source of truth. Registration goes through a verb (`sdlc flake register`) that requires same-head fail→pass evidence, per #390's "intermittency proof before flake eligibility".

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `FlakeEntry{Test, Issue, Status}`, `Registry` (`Lookup(test)`), `parseFlakeTests(frontmatter)` | `cmd/sdlc/internal/flakes/registry.go` | new |
| `Outcome` (tagged: `Passed`, `KnownFlakeRerun`, `KnownFlakeAccepted`, `KnownFlakeFailedTwice`, `Unknown`, `Unattributable`, `PersistentMain`, `UnexplainedPriorFailure`) and `Classify(in ClassifyInput) Outcome` | `cmd/sdlc/internal/flakes/classify.go` | new |
| `Evidence` check: `sameHeadFailPass(attempts, sha, test) (failedID, passedID, ok)` | `cmd/sdlc/internal/flakes/evidence.go` | new |
| `rerunCommands(tests []string) [][]string` (per-package `go test -count=1 -run '^(A\|B)$' <pkg>`) | `cmd/sdlc/internal/flakes/rerun.go` | new |
| `SuiteLogFacts.Unattributed` (a package-level FAIL, a shard exiting with no failing test, tests that never ran, a build error, a timeout) | `cmd/sdlc/teststats_parse.go` | modified |
| `SuiteAttempt` gains `known_flakes [{test, issue}]`, `rerun_of`, `outcome` (schema version 2) | `cmd/sdlc/teststats.go` | modified |
| `localCheck.KnownFlakes` (`rerun`\|`accept`\|`off`) | `cmd/sdlc/localchecks.go` | modified |

- **Registry**: built from each origin/main issue file's `flake_tests:`, joined with its card status. It is 1:N, one issue to several tests. A test that two open issues name refuses as ambiguous. Only an open or active status counts. A terminal status (`done`, `wontfix`, `punt`) unregisters the test.
- **Classify** is the gate's whole decision (ARCH-PURE, ARCH-ORDER). Its input is `{facts, registry, policy, rerun *Result, mainBaseline *Result, priorAttemptsOnSHA}` and its output is one `Outcome` with the issues it names. The tests are a table over every row of the state table below.
- **DRY:** it reuses `failingTests`/`parseSuiteLogFacts` (names), `aggregateSuiteAttempts` (the same-head fail→pass rule, extracted to the shared `sameHeadFailPass`), `diagnoseSuiteFailure` (the main baseline), `loadIssueRecords` (status), and `readLocalChecks`' origin/main read pattern.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `loadFlakeRegistry(ctx, git, records)` | `cmd/sdlc/flakeregistry.go` | new | `git show origin/main:workshop/issues/*`, tracker records |
| known-flake step in `runLocalCheck` (targeted rerun in the same worktree, under the held suite slot) | `cmd/sdlc/localcheck_run.go` | modified | `go test` |
| success-post guard (no `success` while an unexplained failure is on the SHA) | `cmd/sdlc/localcheck_republish.go` | modified | `ghCaller.PostStatus` |
| `sdlc flake register --issue N --test T [--attempt ID]`, `sdlc flake list` | `cmd/sdlc/flake_cmd.go` | new | the issue publish path, the test-stats store |

- ARCH-MOCK: the existing seams carry it. `ciFakeGH` (stateful statuses) is used for posting. A fake test-stats store directory covers attempts, and real git temp repos cover origin/main issue files, as the landing tests already do. The rerun goes through the existing `localRunDeps` command seam, with a fake runner that scripts each attempt's log and exit.

## State table (the gate, per suite run on one head)

Actions: run the suite (`local-check run`, `merge`); register a flake (`sdlc flake register`); close the flake issue; `local-check republish`; `local-check cancel`.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Every test passes | `exit 0`, status success | none | none |
| Fails; every failure attributable, every test registered to an open issue; policy `rerun` | `known flakes #N: rerunning T…` | automatic targeted rerun | none (both attempts kept) |
| Targeted rerun passes | success: `pass · known flake #N rerun (attempt A)` | none | none |
| Targeted rerun fails | failed: `known flake #N failed twice on this head` | investigate: a real fault or a hard flake; the issue's owner is named | none |
| Fails; policy `accept`, all failures registered | success: `pass · known flake #N accepted (policy)` | none | rerun evidence (operator's choice) |
| Any unregistered failure (registered ones listed) | failed: `unknown failure: T (known: #N)` | fix on the branch, or after a same-head fail→pass, register it | none |
| Unattributable (build error, package-level, shard with no test, never ran, timeout) | failed: `unattributable failure` | fix; not flake-eligible | none |
| Unknown failure that also fails on the main baseline | failed: `persistent main failure: T` | fix on main; never flake-eligible without fail→pass | none |
| Registry unreadable (origin/main or tracker unreachable, a malformed `flake_tests`) | failed: `known-flake rule unavailable: <why>` (treated as unknown) | retry once reachable, fix the issue file | none; never treated as an empty registry |
| Flake issue closed between runs | its tests are unregistered, so unknown | reopen, or the fix landed: investigate | none |
| A full rerun passes after an unexplained failure on the same SHA | pass recorded, **no success posted**: `unexplained failure A on this head` | register (this pair is the evidence), then `local-check republish`; or fix | none |
| The process dies or is interrupted during the targeted rerun | no record (as today); the failed attempt stays in test-stats | rerun the exact head (`local-check run`) | rerun time |
| The head moves (a new commit, merge integration) | earlier attempts belong to another head | the new head runs fresh | none |
| Two slots run the same SHA | each attempt is its own record; the guard reads every attempt in the common store | none | none |

**Covered:** every outcome the classifier names, the interrupt, a moved head and a concurrent run. **Not covered:**
- Failed attempts on another machine (the store is per machine). The guard sees only this machine's attempts, which is stated in the help.
- Non-Go suites: a name the rerun can't map is classed unattributable, so it blocks.
- CI failure classes (D3 follow-up).

The event most likely to be mishandled is a full rerun passing after an unknown failure: today it turns green, and that is the loophole the success-post guard closes.

## Plan

- [ ] M1 — **Registry and classifier, pure, plus `flake register`/`list`.**
  - `internal/flakes` (registry, classify, evidence, rerun commands) with table tests over every state-table row, including hostile and ambiguous `flake_tests`.
  - `parseSuiteLogFacts.Unattributed` with log fixtures for each unattributable shape.
  - `SuiteAttempt` v2 fields.
  - `sdlc flake register` writes `flake_tests` through the issue publish path after `sameHeadFailPass` finds evidence in the store. `--evidence TEXT` covers a run outside the store, such as CI, and is recorded verbatim.
  - `sdlc flake list` covers the local view (D3), with CI shown as `not covered`.
- [ ] M2 — **The gate.**
  - `runLocalCheck` classifies a failed attempt, runs the targeted rerun under the held slot in the same worktree, and records both attempts linked (`rerun_of`, `known_flakes`).
  - It posts the status the outcome names, with the description carrying the issue.
  - merge's `diagnoseSuiteFailure` supplies the main baseline for `PersistentMain`.
  - The policy knob is read from origin/main's `.sdlc/local-checks.json`.
  - e2e tests (`ciFakeGH`, fake runner) cover known flake → rerun → landed; unknown → blocked; failed twice → blocked; accept policy; registry unreadable → blocked.
- [ ] M3 — **No blind rerun, and the rules everywhere.**
  - The success-post guard in `postLocalCheckRecord`/`republishSavedPass` blocks success while an unexplained failure sits on the SHA. Registration then lets `republish` post the saved pass.
  - Tests cover full rerun after unknown → not green; register → republish → green; a closed issue → unknown.
  - The tl-worker help and its phrase guards, the local-check and merge help, and the atlas (sdlc-binary test economy) are updated.
  - Base-layer propagation is checked: the code ships in sdlc, which every repo runs.
  - Pilot: a pair run on a real registered flake, with the hub.

Design notes:
- **ARCH-SECURE:**
  - The registry is read from origin/main, so a PR's own issue edits can't excuse its failures.
  - `flake_tests` is parsed into exact `<pkg>.<Test>` names; a regex-shaped or empty name refuses.
  - The rerun builds `-run '^(A|B)$'` from validated identifiers only, using an argv array, never a shell.
  - Status descriptions stay within the 140-character cap and carry no log text.
- **ARCH-ORDER:**
  - The gate's state is per head and per attempt: Classify is pure, and the effects (rerun, post, record) run after it.
  - The interrupt rule is unchanged: an interrupted run records nothing, which is not success.
  - A rerun never repeats in a loop, so there is at most one targeted rerun per failed attempt.
  - The status check rereads the registry at post time, so an issue closed mid-run unregisters its tests.
- **ARCH-FUNERAL:** the new attempt fields ride the store's existing 30-day pruning. Registry entries end when their issue closes. `flake_tests` stays in the archived issue as history.
- **ARCH-CONSTRAINTS:**
  - A targeted rerun is bounded by the number of registered failures in one attempt. The cap is 5 tests; more refuses as "not a flake pattern, investigate" (tl-worker's "a second unrelated flake: stop").
  - It runs within the check's `timeout_minutes`.
  - Reading the registry costs one `git ls-tree` plus `git show` per issue that lists `flake_tests`, bounded by the number of issue files, which is in the hundreds. It is cached per command.
- **Schema v2 rollout:** test-stats readers use `DisallowUnknownFields`. Agent slots run :0's binaries, so the TL rebuilds :0 and a v2 store is read by v2 binaries. A v1 binary refuses a v2 attempt loudly rather than misreading it.
- **Size:** about 450 added code lines, full flow, with three review boundaries.

## Revisions

### 2026-10-18 — hub decisions D1–D4

- **D1, approved:** the targeted rerun counts, the lever the operator asked for on 2026-10-18 ("bypass it during landing"). By the union argument, every test has passed once on the exact head. Side doc 02's rule 2 and the tl-worker help are amended to match, citing that ask, with a Revision note appended to the side doc, which is not overwritten. The tl-worker help changes in M3.
- **D2, approved:** `known_flakes` with `rerun` (the default), `accept` and `off` is built and tested. The guarded `.sdlc/local-checks.json` stays at its default (the field is absent, meaning `rerun`), so this issue needs no `local-gates-verified` label.
- **D3, approved:** a local `sdlc flake list`, with CI shown as `not covered` until #451 (filed, unclaimed) lands.
- **D4, approved:** the registry is issue frontmatter read from origin/main, so closing the issue unregisters it, and registration requires same-head fail→pass evidence.
