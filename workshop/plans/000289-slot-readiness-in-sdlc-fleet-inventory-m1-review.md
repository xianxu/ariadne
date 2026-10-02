# Boundary Review — ariadne#289 (milestone M1)

| field | value |
|-------|-------|
| issue | 289 — Slot readiness in sdlc fleet inventory |
| repo | ariadne |
| issue file | workshop/issues/000289-slot-readiness-in-sdlc-fleet-inventory.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 86ab9fe92f9832cd8c63bfed3e22b212fef67e9a..47ca2a056b6c9864635aa088b6f14cd6dceb7443 |
| command | sdlc milestone-close --issue 289 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-02T16:40:22-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers what the plan claims. There is one shared operation detector (`gitx.OperationMarkers` / `ActiveOperation` / `WorktreeGitDir`) that reads files only, and landing, `move` and `move-detail` now use it. `MeasuredFacts` records the operation without starting another git process, and the call-count assertion proves it. The pure `JudgeCheckout` follows the revised precedence: needs-recovery, then unknown, then holds-work, then ready. The build is clean, `go vet` is clean, and the `gitx`, `fleet` and `cmd/sdlc` tests (Landing, Move, MoveDetail, Operation) pass.

One verdict-correctness problem should be fixed before M2 builds on the verdict. `JudgeCheckout` decides that the issue lookup failed by looking for an empty association on an issue-prefixed branch. That treats "the lookup worked and found no matching issue" the same as "the lookup failed".

**1. Strengths**
- `gitx/operation.go:24-35`: `ActiveOperation` treats any `lstat` error other than "does not exist" as an error, never as "no operation". `facts.go:73-80` keeps that rule: a failed read clears `Operation` and records `OperationError`.
- `operation_test.go:83-113`: an end-to-end test with real git runs a conflicted rebase in a linked worktree. It checks that the detector follows the `gitdir:` pointer and that the primary checkout is unaffected. This is a real conformance check, not a mock.
- `facts_test.go:378-399`: the `countingReader` proves no git process was added, which is the operating-envelope requirement (ARCH-CONSTRAINTS).
- `slots.go:96-108`: the precedence is easy to follow. The one exception (a checkout already holding work while its claims are unread) is named and tested (`slots_test.go:52-55`).
- Every marker in the list belongs to a single worktree in git's layout (none is shared through the common directory), so moving landing and `move-detail` from one `--git-path` call per marker to a single `--absolute-git-dir` call does not change what they detect.

**2. Critical:** none.

**3. Important**
- **`cmd/sdlc/internal/fleet/slots.go:80-85`: the issue verdict is inferred from an empty association.**
  - **What happens:** `AssociateBranchIssue` (`issues.go:165-187`) returns an empty list with no error in three cases:
    - the lookup is nil;
    - the branch contains a backslash;
    - the lookup worked but found no single matching issue (`len(matches) != 1`).
  - **Consequence:** a branch like `000123-foo` whose issue does not exist in this repository (for example a peer's issue number, or an issue that was deleted) is judged `unknown`. Its error text says "could not be read", which is false because the read succeeded. Such a checkout can never become ready.
  - **The real error is lost:** when the lookup does fail, its actual message goes only to `Diagnostics` (`inventory.go:286-289`), and the verdict carries the made-up text instead.
  - **Repeated parsing (ARCH-DRY):** `issueBranchStem` copies the branch-prefix parsing in `AssociateBranchIssue` (`issues.go:169-176`).
  - **Fix:** record the association outcome on the row, the same way claims are recorded (`IssuesError`, or an `IssuesState` like `ClaimsState`), set it at `inventory.go:286`, and have `JudgeCheckout` read it. Then delete `issueBranchStem`.
  - **Tests to add:** "issue-prefixed branch, lookup found nothing → ready / unlanded-commits" and "lookup error → unknown carrying the real error".

**4. Minor**
- **A third marker list remains (ARCH-DRY / ARCH-PURPOSE).** `cmd/weave/internal/refresh/git.go:104` still has its own list, which lacks REBASE_HEAD and BISECT_LOG. The atlas wording, "single-sourced inside `sdlc`", is accurate, but the repository still has two lists. `gitx` lives under `cmd/sdlc/internal`, so weave cannot import it. The fix is to move `OperationMarkers` to `pkg/` (`pkg/` already holds several shared packages) and have weave use it.
- **Unused parameter.** In `issuemovedetail.go:71`, `gitOperationInProgress` no longer uses its `root` parameter. Drop it, and update the caller in `move.go:249`.
- **Global test seams (`facts.go:64-67`).** `readGitPointer` and `lstatMarker` are package-level variables that tests swap out, so those tests can never run in parallel. Acceptable, but injecting them would be cleaner.
- **`json:"-"` on `Operation` / `OperationError`.** A row that goes through JSON and back loses its operation, so a consumer that judges a deserialized row will miss it. M2 should judge before serializing, or document this.
- **Stat → Lstat.** Landing used to call `os.Stat` and now uses `Lstat`. If a marker were a symlink, the result would now differ. The impact is negligible.

**5. Test coverage notes**
- `TestJudgeCheckout` has 18 cases covering the precedence, and it checks that every probe reason carries an error.
- Missing: the case where the issue lookup worked but found nothing (see the Important finding).
- The new markers that landing gains (REBASE_HEAD, BISECT_LOG) and that `move-detail` gains (sequencer, BISECT_START) are tested only through the shared `gitx` table, not end to end. That is acceptable because both callers now go through the same function.

**6. Architectural notes for upcoming work (M2)**

| Principle | Result | Note |
|---|---|---|
| ARCH-PURE | pass | `JudgeCheckout` and `ActiveOperation` are pure over injected functions. |
| ARCH-MOCK | pass | A real-git conformance test exists. |
| ARCH-CONSTRAINTS | pass | Zero extra processes, asserted. |
| ARCH-SECURE | pass | The `.git` pointer is parsed strictly: malformed or multi-line content is an error, and relative paths are resolved. |
| ARCH-ORDER | pass | Holds no state between events; it is a single observation, and `Verdict`'s comment says it must be re-checked at action time. |
| ARCH-FUNERAL | pass | Creates nothing durable. |
| ARCH-DRY / ARCH-PURPOSE | flagged | See above. |

When `AssembleSlots` arrives, it should take each member's verdict from row fields only, never from what is missing, which is the lesson of the Important finding.

**7. Plan revision recommendations**
- Add a `## Revisions` entry. The decision "unknown when the issue named by an issue-prefixed branch [is unread]" needs a recorded probe outcome on the row, distinct from "no matching issue". Also note that weave's marker list is either swept in or explicitly deferred.

```findings
findings:
  - id: new
    severity: Important
    family: verdict-infers-probe-failure-from-absence
    title: |
      JudgeCheckout reads an empty issue association on an issue-prefixed branch as a failed probe, conflating "lookup ok, no matching issue" with a lookup error
    detail: |
      AssociateBranchIssue returns empty with nil error for len(matches)!=1, nil lookup, or a backslash branch (issues.go:165-187); slots.go:80-85 then reports unknown with fabricated text "could not be read", and the real lookup error is only in Diagnostics. issueBranchStem also duplicates AssociateBranchIssue's prefix parsing (ARCH-DRY). Record the association outcome on TreeRow (like ClaimsState/ClaimsError), have JudgeCheckout read it, delete issueBranchStem, and add tests for lookup-ok-no-match (ready) and lookup-error (unknown with the real error).
  - id: new
    severity: Minor
    family: single-source-marker-list
    title: |
      cmd/weave/internal/refresh/git.go:104 keeps a third, divergent operation-marker list
    detail: |
      Lacks REBASE_HEAD and BISECT_LOG. gitx is under cmd/sdlc/internal so weave cannot import it; moving OperationMarkers to pkg/ would make it one source repository-wide.
  - id: new
    severity: Minor
    family: dead-parameter
    title: |
      gitOperationInProgress no longer uses its root parameter (issuemovedetail.go:71)
```

---

## Re-review — 2026-10-02T16:48:15-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 289 — Slot readiness in sdlc fleet inventory |
| repo | ariadne |
| issue file | workshop/issues/000289-slot-readiness-in-sdlc-fleet-inventory.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 86ab9fe92f9832cd8c63bfed3e22b212fef67e9a..df660572057fc52f341b3ef0be68670757c511ed |
| command | sdlc milestone-close --issue 289 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-02T16:48:15-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Verdict: FIX-THEN-SHIP.** All three prior findings are fixed and backed by evidence. `JudgeCheckout` now reads the recorded lookup outcome (`TreeRow.IssuesError`), and `issueBranchStem` is gone. The marker list now has one source, `pkg/workspace`, and weave's refresh uses it too. The unused `root` parameter has been removed. The targeted packages pass: `go test ./pkg/workspace/ ./cmd/sdlc/internal/fleet/ ./cmd/sdlc/internal/gitx/ ./cmd/weave/internal/refresh/`. One cheap fix remains before the boundary. The atlas, and the issue Log, still say the list is `gitx.OperationMarkers`, but the move to `pkg/` changed that and they weren't updated. Nothing blocks shipping.

### 1. Strengths
- `pkg/workspace/operation.go:24-35`: `ActiveOperation` reports any `lstat` failure other than "not found" as an error, never as "no operation". The tests check this at the unit level (`operation_test.go` "probe fails") and through `CollectFacts` (`facts_test.go`, the permission-denied case).
- Replacing one `rev-parse --git-path` call per marker with a single `--absolute-git-dir` is correct. The markers (MERGE_HEAD, sequencer, BISECT_*, rebase-*) are all stored per worktree in Git, not in the shared directory. The end-to-end test with a real conflicted rebase in a linked worktree (`gitx/operation_test.go`) confirms it.
- `facts_test.go`: counting git calls proves the "no new git process" limit the plan sets (ARCH-CONSTRAINTS).
- `slots.go:47-106`: the order in which reasons decide the verdict matches the plan's Decisions exactly. That includes the exception where unread claims don't matter if the checkout already holds work. The test table covers each branch of it.

### 2. Critical
None.

### 3. Important
- `atlas/workflow/workspace-branching.md:35` says the list is "single-sourced as `gitx.OperationMarkers`". The code has it at `workspace.OperationMarkers` in `pkg/workspace/operation.go`, and the atlas omits weave refresh as a consumer. The issue Log line ("M1 implemented: `gitx.OperationMarkers`…") is stale in the same way. **Fix:** change both to `workspace.OperationMarkers (pkg/workspace)` and add weave refresh to the atlas's list of consumers.

### 4. Minor
- `issues.go:183`: `len(matches) > 1` (an ambiguous issue ID) returns no association with no error, so `JudgeCheckout` reads it as ready. An ambiguous answer becomes "no issue". It's rare, but on an issue-prefixed branch it should probably count as `probe:issue`.
- `cmd/sdlc/internal/gitx/operation_test.go` only exercises `pkg/workspace` functions. It lives in `gitx` only because `testfix` is internal, so the location is misleading. A comment saying why would help.
- `facts_test.go`: the test swaps the package-level `lstatMarker`, so it can't use `t.Parallel()`. Acceptable as is.

### 5. Test coverage
- The BR-1 fix has regression tests: the "lookup ok, no match → ready" case would go red against the old code that inferred failure from an empty association.
- The inventory wiring (`inventory.go:286-303`) has no direct test. It's trivial, and M2's assembly fixtures will exercise it.
- Weave refresh's existing tests pass on the shared detector.

### 6. Architecture
- **ARCH-DRY: pass.** There's one marker list and one detector, and all four readers use it.
- **ARCH-PURE: pass.** `ActiveOperation`, `WorktreeGitDir` and `JudgeCheckout` are pure with injected reads, and `collectOperation` is a thin layer around them.
- **ARCH-PURPOSE: pass for M1.** The whole consumer list is covered, and the atlas wording is the only leftover.
- **ARCH-MOCK: pass.** The checks are file probes behind seams, and a real-git conformance test covers them.
- **ARCH-CONSTRAINTS: pass.** No extra git process, and it's asserted.
- **ARCH-SECURE: pass.** A malformed `.git` pointer file fails loudly ("is not a gitdir pointer") instead of being trusted. The only weak spot is the ambiguous-match Minor above.
- **ARCH-ORDER: pass.** Each read is a single shot and keeps no state between events.
- **ARCH-FUNERAL: pass.** The change creates nothing durable.

### 7. Plan revisions
None: the plan already names `pkg/workspace`. Only the atlas and the issue Log need the correction.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      TreeRow.IssuesError recorded in inventory.go:286-303 and read at slots.go:77; issueBranchStem deleted; slots_test.go cases "lookup failed" (unknown with real error) and "lookup ok, no match" (ready) go red on the prior inference.
  - id: BR-2
    disposition: addressed
    note: |
      weave refresh/git.go:105-114 now uses workspace.ActiveOperation over the one pkg/workspace list; refresh tests pass.
  - id: BR-3
    disposition: addressed
    note: |
      gitOperationInProgress(git) drops root (issuemovedetail.go:72); callers at :100 and move.go:249 updated.
findings:
  - id: new
    severity: Important
    family: docs-restate-moved-identifier
    title: |
      Atlas (and issue Log) name gitx.OperationMarkers; the list lives at workspace.OperationMarkers and weave is an unlisted consumer
    detail: |
      atlas/workflow/workspace-branching.md:35 was written before the move to pkg/workspace and not updated by df660572; replace gitx.OperationMarkers with workspace.OperationMarkers (pkg/workspace) everywhere it appears (atlas, issue Log) and add weave refresh to the consumer list.
  - id: new
    severity: Minor
    family: verdict-infers-probe-failure-from-absence
    title: |
      An ambiguous issue match (len(matches)>1) collapses to no association, so an issue-prefixed branch reads ready
    detail: |
      This is the 2nd finding in the family. The rule is that an answer which is not a clean match/no-match (an error or an ambiguous result) must be recorded on the row as an outcome rather than folded into empty. issues.go:183 should return an error or set IssuesError for ambiguity, so that JudgeCheckout reports probe:issue.
  - id: new
    severity: Minor
    family: test-placement
    title: |
      gitx/operation_test.go tests only pkg/workspace functions; it lives in gitx because testfix is internal
```

---

## Re-review — 2026-10-02T16:54:54-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 289 — Slot readiness in sdlc fleet inventory |
| repo | ariadne |
| issue file | workshop/issues/000289-slot-readiness-in-sdlc-fleet-inventory.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 86ab9fe92f9832cd8c63bfed3e22b212fef67e9a..80aa0751afa6bf1e5751adf4ca18cb571142548c |
| command | sdlc milestone-close --issue 289 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-02T16:54:54-07:00 |
| verdict | SHIP |

## Review

I've checked the three open findings and am finishing the architecture pass.

```verdict
verdict: SHIP
confidence: high
```

All three open findings are fixed, with evidence:

- **BR-4:** the atlas and the issue Log now name `workspace.OperationMarkers`, and weave refresh is in the consumer list.
- **BR-5:** an ambiguous match now returns `ErrAmbiguousIssue`. The inventory records it on the row as `IssuesError` and as a diagnostic, and `JudgeCheckout` turns that into `probe:issue`. Tests cover each step: the `wantErr` row, the cardinality property test, and the judge's "lookup failed" row.
- **BR-6:** the real-git rebase test now lives in `pkg/workspace/operation_test.go`, next to the code it tests, and runs its own git commands without the internal `testfix` helper.

`go build ./...` and `go vet` are clean, and `go test ./pkg/workspace ./cmd/sdlc/internal/fleet` passes. One stale identifier is left: the issue's Plan row still says `gitx` detector. It is Minor and doesn't block.

**1. Strengths**
- Every reader now uses one marker list: `pkg/workspace/operation.go:16`. landing, move, move-detail, weave refresh and fleet each ask git for the directory once with `--absolute-git-dir`, where they used to run one `rev-parse` per marker.
- `ActiveOperation` (`operation.go:24`) treats any lstat error other than "not found" as an error. An unreadable directory can never read as "no operation".
- `JudgeCheckout` (`slots.go:53`) is pure. The order it checks things in is written down (recovery, then probe, then holds, then ready), and a table test covers it, including "dirty and base unavailable".
- `collectOperation` (`facts.go:73`) reads files only and starts no extra git process, which keeps fleet's cost budget.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- `workshop/issues/000289-…md:104` still says "shared file-based `gitx` operation detector". `issuemovedetail.go:70` still says "in root's worktree", but that parameter was removed. Plan line ~151 says the rebase test "reports `rebase-merge`", but `REBASE_HEAD` comes first in the marker order, so that is what gets reported, and the test only checks for "rebase". This is the second finding in family `docs-restate-moved-identifier`. The underlying rule: when a commit moves or renames an identifier, or removes a parameter, grep the issue, the plan, the atlas and nearby comments for the old name in the same commit.
- landing switched from `os.Stat` to `Lstat`. It now also refuses on REBASE_HEAD and BISECT_LOG. That is stricter and intended (the atlas says so); I'm noting it only because it changes behavior.

**5. Test coverage**
- The cardinality property test pins ambiguous as an error and never an empty association.
- The judge table covers lookup failed versus lookup succeeded with no match.
- The real-git linked-worktree rebase test covers the `gitdir:` pointer path.
- Nothing tests the full path from an ambiguous match in the inventory to `probe:issue`. The pieces are each tested, and M2's real-git fixtures could add that link.

**6. Architecture (each principle checked)**
- **ARCH-DRY: pass.** Three divergent marker lists became one.
- **ARCH-PURE: pass.** The verdict is pure, and the file reads are passed in as functions.
- **ARCH-PURPOSE: pass.** The consumer sweep is complete, and grep finds no leftover marker lists.
- **ARCH-MOCK: pass.** Tests inject a fake `lstat` and a fake pointer reader, and one real-git test checks against actual git.
- **ARCH-CONSTRAINTS: pass.** Fleet starts no new process, and landing and move start fewer.
- **ARCH-SECURE: pass.** A malformed `gitdir:` file (multiple lines or empty) is rejected with a visible error.
- **ARCH-ORDER: pass.** The verdict is a single pure decision and holds no state between events.
- **ARCH-FUNERAL: pass.** Nothing durable is created; the operation and issue-error fields have `json:"-"` and are never saved.

**7. Plan revisions:** fix the issue's M1 row so it says `pkg/workspace` instead of `gitx`, and change the plan's expected rebase marker to "a rebase marker (REBASE_HEAD first in order)".

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      atlas/workflow/workspace-branching.md:35-38 now names workspace.OperationMarkers (pkg/workspace/operation.go) and lists weave refresh; issue Log uses the new name.
  - id: BR-5
    disposition: addressed
    note: |
      issues.go:190 returns ErrAmbiguousIssue; inventory.go:286 records it as IssuesError, so JudgeCheckout gives probe:issue; tests: the issues_test wantErr row and the cardinality property test.
  - id: BR-6
    disposition: addressed
    note: |
      The gitx test was deleted; the same scenario is now in pkg/workspace/operation_test.go with its own git helper, and it passes.
findings:
  - id: new
    severity: Minor
    family: docs-restate-moved-identifier
    title: |
      Issue Plan M1 row still says gitx detector; issuemovedetail comment names removed root param
    detail: |
      2nd in family. Rule: a commit that moves or renames an identifier, or drops a parameter, greps the issue, the plan, the atlas and nearby comments for the old name in the same commit. Remaining sites: issue line 104 ("`gitx` operation detector"), issuemovedetail.go:70 ("in root's worktree"), and plan Chunk 1 ("reports rebase-merge"; REBASE_HEAD is reported first).
```
