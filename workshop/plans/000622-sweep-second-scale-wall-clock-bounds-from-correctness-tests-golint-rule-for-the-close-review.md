# Boundary Review — ariadne#622 (whole-issue close)

| field | value |
|-------|-------|
| issue | 622 — Sweep second-scale wall-clock bounds from correctness tests; golint rule for the shape |
| repo | ariadne |
| issue file | workshop/issues/000622-sweep-second-scale-wall-clock-bounds-from-correctness-tests-golint-rule-for-the.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2bc11af146651898f1506721f33fcc41f250f3b4..5338705620e4698fbda6e487c90214ecc7394d7f |
| command | sdlc close --issue 622 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T10:19:45-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Disposing BR-11 (verified with tests); now reading the rest of the M2 test diffs.

Now running the lint, helper, and compile checks across the changed packages.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2 delivers what the issue is for. `wallclockPending` and its test are gone, so `test-tight-wallclock` now checks every file, and `pkg/golint` passes against the repo catalog (`go test ./pkg/golint ./pkg/testgate`: green). The group B kill/cancel sites now run their bounded call through the new `testgate.Returns`, and each hung fixture is made to outlive `HangCeiling`. If a kill or timeout breaks, the test fails at the hang guard rather than on a timing threshold. I ran a sample of the converted tests (boundary push, gh cancel, help probe, integrate gate, split impact, cleanup git, retention, suite queue, planning review, makebuild, gitx, fleet, testfix, acquire, golist ×3, couch deadline) and all pass. The one exception is `TestFleetResumeBoundedRecheck`, which is refused by this reviewer sandbox's `git init` confinement (sdlc #339); that is not a code defect. `go vet ./cmd/... ./pkg/...` is clean. BR-11 is fixed and I checked it by mutation: with the closed-channel guard removed, `TestAwaitConditionRefusesUnbufferedOrClosedWorkers` fails with `panic: send on closed channel`, and it passes again once restored. One cheap item remains before SHIP: the atlas does not describe the new `testgate.Returns` helper or its "outlive the ceiling" convention, which group B relies on at about 15 sites.

### 1. Strengths
- **Shared helpers:** `pkg/testgate/await.go:19-44`. The workers check at entry (`cap < 1` causes Fatalf) and the `err, open` receive close BR-11 for every caller. `Returns` (`await.go:50-66`) is built on `AwaitCondition`, so there is still only one wait helper.
- **Hung fixtures:** the group B rewrites make every hung fixture `sleep 600` (or 10 min), so a broken bound is the only way a test can reach the ceiling. This matches the plan's "mutation → hang-guard failure, not timing" proof.
- **No timer race in `golist_lifetime_test.go:28-63`:** cancellation now waits for the descendant's pid (no timer), and the descendant's death is checked with `AwaitExit`. Both the timer race and the old inference from the marker file are gone.
- **Worker results are not lost:** `planningreview_test.go:38-49` (`runPlanningCLI`) and `localcheck_lease`'s `exited` channel turn the existing completion channels into `chan error` workers. Because a consumed result is put back, the cleanup paths that read `done` still receive it.
- **Group C allows** (`judge/cancellation_unix_test.go:98`, `localcheck_lease_test.go:226`, `net_signal_unix_test.go:72`) each name the contract they bound and why no event can stand in.

### 2. Critical findings
None.

### 3. Important findings
- **`atlas/workflow/sdlc-binary.md:2627`: the atlas update for `testgate.Returns` appears to be missing.** M2 adds an exported helper with a convention that test authors must follow: make the work that the bound must stop outlast `HangCeiling`, and run the bounded call under `Returns`. Group B now depends on it at about 15 sites. The atlas paragraph covers `AwaitCondition` only, and it says nothing about the new up-front refusal of unbuffered or closed channels.
  - This is the 2nd finding in family `atlas-surface-coverage`.
  - The rule that covers both findings: every exported `pkg/testgate` or `testfix` helper that carries a usage convention gets its atlas line in the same commit.
  - Fix: one sentence after the `AwaitCondition` sentence, covering `Returns` and the refusal.

### 4. Minor findings
- **Hand-rolled poll loops remain** (3rd in `duplicated-poll-loop`). The rule: every test poll-until-condition goes through `testgate.AwaitCondition`. Measured prevalence of what's left:
  - `landingdrop_train_test.go:119` `awaitDropCondition` and `merge_stage_lock_test.go:61` `mergeStageAwait`: worker-less, budgeted by `landingCompletionBudget`, 8 call sites between them.
  - `testfix.AwaitPIDFile` and `testfix.AwaitExit`.

  The M2 enumeration covered only HangCeiling-bounded loops. Either widen it, or record in the plan why loops bounded by the test-deadline budget are excluded.
- **Import grouping:** about 15 files add `"github.com/xianxu/ariadne/pkg/testgate"` inside the stdlib import group (for example `boundarypush_test.go:6`, `queue_suite_status_test.go:9`). gofmt accepts it, but goimports grouping does not.
- **Plan checkboxes:** the plan file's checkboxes for M1 and M2 are still `[ ]`, while the issue file shows `[x]`.
- **Orphaned sleeps:** when a kill is broken, the `sleep 600` children outlive the failed test by up to about 10 minutes. This only happens on failure, so it is acceptable, but it would be worth a line in the atlas convention.

### 5. Test coverage notes
- `await_test.go` covers four cases: fail-fast with put-back, the condition holding, the unbuffered and closed refusals, and `Returns`. The unbuffered case would hang until `HangCeiling` if its guard were removed, so it still fails, just slowly.
- Group B relies on fixtures that outlive the ceiling instead of elapsed checks. The Log's `runBounded` mutation (kill removed → "not reached within 2m0s") shows this works. Nothing replaces the old checks that the probe's own five-second ceiling held and that a cancelled observation started no slow work. Now only "returned before the ceiling" is asserted, which matches what the issue asks for.

### 6. Architectural notes
- ARCH-DRY: pass, apart from the duplicated-poll-loop Minor above.
- ARCH-PURE: pass; the rule is pure AST code.
- ARCH-PURPOSE: pass. The pending list is deleted and the rule enforces repo-wide; this is the purpose, not deferred.
- ARCH-MOCK: N/A; the seams are unchanged.
- ARCH-CONSTRAINTS: pass. Every wait is bounded by `HangCeiling`.
- ARCH-SECURE: N/A.
- ARCH-ORDER: pass. The golist cancellation now waits for the descendant's pid.
- ARCH-FUNERAL: pass. `wallclockPending` was created in M1 and deleted in M2, as planned.
- ARCH-EVOLVE: N/A; no persisted formats change.
- Downstream (`base-layer-downstream-impact`): unchanged since M1. pair#885 still governs the pair pin bump, and the pending-list deletion only affects ariadne's own paths.

### 7. Plan revision recommendations
- Tick the M1 and M2 checkboxes in the plan.
- Add a "## Revisions" line that either lists `awaitDropCondition`, `mergeStageAwait` and `AwaitPIDFile` as converted, or names them as deliberately left (budgeted by the test deadline, or helpers parameterised by a ceiling).

```findings
dispose:
  - id: BR-11
    disposition: addressed
    note: |
      await.go:21-25 refuses cap<1 up front and 29-33 receives err,open; TestAwaitConditionRefusesUnbufferedOrClosedWorkers goes red (send on closed channel panic) with the guard reverted, green restored.
findings:
  - id: new
    severity: Important
    kind: defect
    family: atlas-surface-coverage
    title: |
      atlas update appears missing for testgate.Returns and its outlive-the-ceiling convention
    detail: |
      2nd in family; rule: every exported pkg/testgate or testfix helper that carries a usage convention gets its atlas line in the same commit. atlas/workflow/sdlc-binary.md:2627 covers AwaitCondition only; Returns (await.go:50), which about 15 group B sites use, plus the unbuffered/closed-worker refusal, are undocumented. Fix: one sentence beside the AwaitCondition sentence.
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-poll-loop
    title: |
      Worker-less hand-rolled poll loops remain outside AwaitCondition
    detail: |
      3rd in family; rule: every test poll-until-condition goes through testgate.AwaitCondition. Prevalence: awaitDropCondition (landingdrop_train_test.go:119) and mergeStageAwait (merge_stage_lock_test.go:61), budgeted by landingCompletionBudget, 8 call sites between them; testfix.AwaitPIDFile/AwaitExit. Either convert them or record in the plan why loops bounded by the test-deadline budget are excluded.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d8dd21a1-9bd3-4990-aeb0-382e10ebe75f","repository":"github.com/xianxu/ariadne","issue_id":"000622","recipe":"milestone-review","round":5,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/d184366a789fb611","reviewed_head":"5338705620e4698fbda6e487c90214ecc7394d7f","slot":"ariadne:2","requested_at":"2026-10-25T17:16:45.427713Z"},"phase":"finished","pid":86347,"started_at":"2026-10-25T17:16:45.523863Z","finished_at":"2026-10-25T17:19:43.944119Z","session_id":"be2d092f-f193-48c0-83c8-6d91d5ad6996","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T10:25:28-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 622 — Sweep second-scale wall-clock bounds from correctness tests; golint rule for the shape |
| repo | ariadne |
| issue file | workshop/issues/000622-sweep-second-scale-wall-clock-bounds-from-correctness-tests-golint-rule-for-the.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2bc11af146651898f1506721f33fcc41f250f3b4..37c8aece1620c0422195dc5de38fe45bfebbe14e |
| command | sdlc close --issue 622 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T10:25:28-07:00 |
| verdict | SHIP |

## Review

Focus on the last commit, which disposes BR-12/BR-13.

```verdict
verdict: SHIP
confidence: high
```

Both open findings are fixed at the pinned head 37c8aece, and no new blocking issues turned up. For BR-12, `atlas/workflow/sdlc-binary.md:2627` now documents `testgate.Returns`. It covers the outlive-the-ceiling convention (the work the bound must stop runs past `HangCeiling`, e.g. `sleep 600`), the refusal of unbuffered or closed worker channels, and the fact that a child outlives a broken bound. Each claim matches `pkg/testgate/await.go`. For BR-13, the finding allowed two fixes, and the commit uses the second: the plan's 2026-10-25 Log entry and the atlas both state the rule ("`AwaitCondition` owns every poll bounded by `HangCeiling`"). They also name the waits it excludes and why. `awaitDropCondition` and `mergeStageAwait` are bounded by `landingCompletionBudget`, which is derived from the test's own deadline (`landingwait_test.go:95`). `AwaitPIDFile` and `AwaitExit` return a value to a caller that sets the ceiling. I checked `awaitDropCondition` at `landingdrop_train_test.go:119`, and it does take its budget from the test deadline, so the stated scope is accurate. `go test ./pkg/testgate ./pkg/golint` passes, and `go vet` over cmd/sdlc, cmd/tl-lead and cmd/weave is clean, which confirms the moved imports compile.

1. **Strengths**
   - `pkg/testgate/await.go:19-44`: `AwaitCondition` checks buffering before it starts polling, puts each worker's result back on its channel, and is bounded only by `HangCeiling`. This one helper is shared by sdlc, weave and tl-lead.
   - `Returns` (`await.go:50`) proves a bound by structure, not by elapsed time: only a broken bound can reach the ceiling.
   - The plan's Log records which waits the rule covers and why each exclusion exists. It does not just leave them out, which stops this family from repeating.
   - `TestAwaitConditionRefusesUnbufferedOrClosedWorkers` tests both refusal paths through an injected `fatalTB`, without any IO.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The atlas says the helper refuses a closed worker channel "up front", but `await.go` only notices a closed channel on its first poll. If `cond()` is already true when the call starts, a closed channel is never reported. This is harmless, since the wait succeeded anyway; the fix is either "on the first poll" in the atlas or moving the check into the pre-loop pass. I'm not raising it as a finding.

5. **Test coverage:** the testgate unit tests cover the condition holding, a worker ending first, the refusals, and `Returns` waiting for its call. The golint rule tests passed. The other changes in this round are only prose and import reordering, so no new regression test is needed.

6. **Architecture**
   - ARCH-DRY: pass. There is one shared helper, and the waits left outside it have a stated reason.
   - ARCH-PURPOSE: pass. The lint rule enforces on every file now that `wallclockPending` is deleted.
   - ARCH-PURE: pass. The rule is pure AST code.
   - ARCH-MOCK: N/A. The change is test infrastructure and adds no external dependency.
   - ARCH-CONSTRAINTS: pass. Waits are bounded by `HangCeiling`.
   - ARCH-SECURE: N/A. Nothing reads untrusted input or touches secrets.
   - ARCH-ORDER: pass. Waits fail fast when a worker ends first, which takes the ordering assumption out of the oracle.
   - ARCH-FUNERAL: pass. The atlas notes that a long child outlives a failed test.
   - ARCH-EVOLVE: N/A. No persisted format changes.

7. **Plan revisions:** none. The M1/M2 boxes are ticked and the scope is recorded. The sub-item checkboxes are still unticked, which is cosmetic.

```findings
dispose:
  - id: BR-12
    disposition: addressed
    note: |
      atlas/workflow/sdlc-binary.md:2627 now documents testgate.Returns, its outlive-HangCeiling convention, worker-channel refusal and the orphan-child consequence; matches pkg/testgate/await.go:19-64.
  - id: BR-13
    disposition: addressed
    note: |
      Took the finding's second option: the plan Log (2026-10-25) and the atlas state the rule (AwaitCondition owns HangCeiling-bounded polls) and why awaitDropCondition and mergeStageAwait (test-deadline budget via landingCompletionBudget) and AwaitPIDFile/AwaitExit (return to the caller) are excluded; verified at landingdrop_train_test.go:119.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2bc00693-7063-4587-903b-396c5b0479a6","repository":"github.com/xianxu/ariadne","issue_id":"000622","recipe":"milestone-review","round":6,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/6a3be93da0d4c526","reviewed_head":"37c8aece1620c0422195dc5de38fe45bfebbe14e","slot":"ariadne:2","requested_at":"2026-10-25T17:24:38.739831Z"},"phase":"finished","pid":17160,"started_at":"2026-10-25T17:24:38.90007Z","finished_at":"2026-10-25T17:25:22.915845Z","session_id":"2c7ebd8f-61d8-4888-aefb-cecde22a7d9b","outcome":"returned"}] -->
