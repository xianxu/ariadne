# Boundary Review — ariadne#490 (whole-issue close)

| field | value |
|-------|-------|
| issue | 490 — Train instance gate reads issue files from the conductor checkout, not the train-prefix tree: a later car's new issue file is held as missing |
| repo | ariadne |
| issue file | workshop/issues/000490-train-instance-gate-reads-issue-files-from-the-conductor-checkout-not-the-train-prefix-tree-a-later-car-s-new-issue-file-is-held-as-missing.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3e01f169ac8a4b8e27277fdf1692149e6ca483f1..63346f514919ad5b093b8e3184c068222e28a49f |
| command | sdlc close --issue 490 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-20T03:15:03-07:00 |
| verdict | SHIP |

## Review

The targeted tests pass. Next I'll revert the fix in the disposable checkout to confirm the new test fails without it.

```verdict
verdict: SHIP
confidence: high
```

The fix is small and correct. `validateChangedInstances` and `runLandingInstanceGate` now take a `root`, which is the tree that holds the gated head. The train passes `c.wt`, its own worktree. After a car's `merge --no-ff`, that tree is checked out at exactly the head being gated (`landingtrain.go:270-275`). Single landing, merge and push pass `""`, so they behave as before. The file list still comes from `git diff --name-status` on OIDs, which works from any worktree because they share one object store. Only the file read and the validator's path change. I confirmed the regression test fails without the fix. I put the reviewed head back afterwards (`git checkout landingtrain.go`, and `git status` was clean). The only item left is the live Done-when clause, and the plan already marks it unchecked on purpose.

1. **Strengths**
   - `cmd/sdlc/validategate.go:89`: one `file := filepath.Join(root, ch.Path)` now feeds both the `vocabulary` call and `readIssueFileFn`. Problem messages still use the repo-relative `ch.Path`, so the held-car detail names the file the way users expect.
   - `cmd/sdlc/landingtrain.go:296`: reusing `c.wt` follows the Spec's "no second resolver". Nothing new to keep in sync.
   - The tests are real regression evidence. I changed `c.wt` back to `""`. `TestLandingTrainGatesEachPrefixInItsOwnTree` then failed with pair's exact symptom ("vocabulary: read that file: open workshop/issues/000002-second.md: no such file or directory"). The control test failed too, because the held detail named the read failure instead of the rule. With the fix, both pass.
   - The `gateInstancesByReading` stub actually reads the path it is handed, like the real validator. So the test checks how paths are resolved, not just that a mock was called.
   - The class sweep holds up. `runLandingDuplicateGate` only uses `refIDSpace` on OIDs (`landinggate.go`), so it doesn't depend on which tree it runs in.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - Done-when clause 3 is still open: "pair's next train with 2+ cars lands both; recorded in test-economy". It can only happen after landing, and the Plan marks it `[ ]`. Track it so the issue isn't archived without the live evidence.

5. **Test coverage notes:** The "car 2 adds a file the conductor lacks" clause is tested both ways. With a good file the car boards. With a malformed one it is held, and the file and rule are named (`(section)`). The `""` root path is still covered by the existing `validategate_test.go` and `landinggate_instance_test.go` cases, all updated to the new signature. These tests pass: `TestLandingTrain*`, `TestLandingInstanceGate*`, `TestValidateChanged*`, `TestLandingPasses*`.

6. **Architectural notes**
   - **ARCH-DRY: pass.** One join site, no duplicate resolver.
   - **ARCH-PURE: pass.** The IO seams (`validateFrontmatterFn`, `readIssueFileFn`, `diffNameStatusFn`) are still injected. `root` is just a plain parameter.
   - **ARCH-PURPOSE: pass.** All consumers of the instance gate were checked: train, durable merge (two sites), `runMerge`, `runPush`. Only the train gates a tree other than its own working directory, and the sibling per-prefix gate was swept.
   - **ARCH-EVOLVE: pass.** No persisted format changed.
   - **Docs:** no new user-facing surface, so no atlas or README update is needed.
   - **Pre-existing, outside this issue's scope:** the single durable landing gates `head` while passing `""` (the current directory). That assumes the checkout matches the head's changed files. A future issue might pass that landing's own integrated tree too.

7. **Plan revision recommendations:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: done-when-live-evidence-pending
    title: |
      Done-when clause 3 (pair's next 2+ car train lands both, recorded in test-economy) is not yet evidenced
    detail: |
      This can only happen after landing, and the Plan marks it unchecked. It is the only unmet Done-when clause. Record the live train result in test-economy before archiving the issue.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"9d384104-6876-48fb-b229-d83caeb9caca","repository":"github.com/xianxu/ariadne","issue_id":"000490","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/662e37be396acbc4","reviewed_head":"63346f514919ad5b093b8e3184c068222e28a49f","slot":"ariadne:4","requested_at":"2026-10-20T10:11:55.465551Z"},"phase":"finished","pid":44916,"started_at":"2026-10-20T10:11:55.661086Z","finished_at":"2026-10-20T10:15:00.406001Z","session_id":"48d55c11-cda1-4722-8692-31d4b55d6687","outcome":"returned"}] -->
