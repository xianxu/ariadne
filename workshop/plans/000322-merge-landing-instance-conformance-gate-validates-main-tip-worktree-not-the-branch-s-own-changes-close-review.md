# Boundary Review — ariadne#322 (whole-issue close)

| field | value |
|-------|-------|
| issue | 322 — merge: landing instance-conformance gate validates main-tip..worktree, not the branch's own changes |
| repo | ariadne |
| issue file | workshop/issues/000322-merge-landing-instance-conformance-gate-validates-main-tip-worktree-not-the-branch-s-own-changes.md |
| boundary | whole-issue close |
| milestone | — |
| window | 368314170c805e8be5280e6c8e4ba323e0a9e792..dfa7d125af64faf13feb320e0cd95e1328337862 |
| command | sdlc close --issue 322 |
| reviewer | claude |
| timestamp | 2026-10-10T18:15:38-07:00 |
| verdict | SHIP |

## Review

Diff read; now checking that the landing checkout is clean and sits at the branch head, since the frontmatter and section reads come from disk.

Code looks consistent. Running the four new tests to confirm they pass at head.

```verdict
verdict: SHIP
confidence: high
```

The fix does what the issue set out to do. `runDurableMerge` used to call the instance-conformance gate with `pr.BaseOID` (main's current tip) and the working tree. It now calls `runLandingInstanceGate(targetMain, head, …)`, which validates `merge-base(targetMain, head)..head`. So when main archives a nonconforming issue after the branch point, that file no longer reaches the gate. The merge-base comes from one helper, `landingMergeBase`, which the duplicate-id gate now uses too (ARCH-DRY). If the merge-base can't be found, the gate fails closed. I checked the neighbourhood: frontmatter and section reads still come from disk, which is safe here. `nextLandingAction` refuses to start a landing from rest or from a missing branch, and `landingCheckoutReady` requires a clean checkout on that branch, so the working tree equals `head` when the gate runs. `head` can't be empty on this path. The targeted tests pass at head (`TestLandingInstanceGate*`, `TestLandingPassesBranchWindowToInstanceGate`, `TestLandingDuplicate*`: ok, 8.7s). The legacy `merge.go` and `push.go` callers use `gitx.DiffBase()` and are correctly out of scope.

1. **Strengths**
   - `cmd/sdlc/landinggate.go:48`: `landingMergeBase` is a small fail-closed helper with one consumer per gate, and its comment explains why no gate may fall back to main's tip.
   - `cmd/sdlc/landinggate_instance_test.go`: the fixture uses real git and recreates the exact #189/#317 shape (main archives a nonconforming issue after the branch is cut).
   - The "1 changed instance file(s) conform" count assertion catches a mutation inside the gate (diffing from `mainOID`) even with frontmatter stubbed to pass, because the archived file would raise the count to 2.
   - `TestLandingPassesBranchWindowToInstanceGate` pins the production call site. The old code passed `pr.BaseOID` and `""`, which fails both its base and head assertions, so the call-site mutation claim checks out.
   - Both cases in the "keeps its teeth" clause are covered: the branch's own nonconforming issue refuses, and main's archived issue is ignored. The fail-closed path has its own test.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `landinggate.go:30` vs `landing.go:406`: the duplicate gate measures from `"HEAD"` and the instance gate from the pinned `head`. They are the same commit after the clean-checkout check, but passing `head` to both would make the two gates measure the same pinned window by construction.

5. **Test coverage notes:** the Done-when clauses are covered in both modes (main's change ignored, branch's own change refused), and fail-closed is covered. I didn't revert code in a scratch copy, because review tools are read-only. I checked the mutation claims by reading the assertions against the old arguments instead.

6. **Architectural notes**
   - ARCH-DRY: pass. The two copies of the merge-base probe are now one helper.
   - ARCH-PURE: pass. The gate stays a thin git-glue layer over the injectable `validateChangedInstancesFn` and `diffNameStatusFn`.
   - ARCH-PURPOSE: pass. The production call site is fixed and pinned by a test. The other `validateChangedInstancesFn` callers (`merge.go:328`, `push.go:129`) already diff from the branch point (`gitx.DiffBase()`), so none of the family remains.

7. **Plan revision recommendations:** none.

```findings
findings:
  - id: new
    severity: Minor
    family: landing-gates-pinned-window
    title: |
      Duplicate-id gate measures merge-base from "HEAD" while the instance gate uses the pinned head
    detail: |
      landinggate.go:30 calls landingMergeBase(r, mainOID, "HEAD"); landing.go:406 now passes the pinned head to the instance gate. They are equal after landingClean, but threading head into runLandingDuplicateGate would make both gates measure the same pinned window by construction. Only these two landing gates are in this family.
```
