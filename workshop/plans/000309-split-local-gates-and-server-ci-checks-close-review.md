# Boundary Review — ariadne#309 (whole-issue close)

| field | value |
|-------|-------|
| issue | 309 — Split local gates and server CI checks |
| repo | ariadne |
| issue file | workshop/issues/000309-split-local-gates-and-server-ci-checks.md |
| boundary | whole-issue close |
| milestone | — |
| window | e1a463ec6875969047968d05b01e05031f0776af..1a7aa8de60c4245f06267e2b9e94241f221dc895 |
| command | sdlc close --issue 309 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T14:21:35-07:00 |
| verdict | SHIP |

## Review

Running a revert test of the BR-13 fix (Go build is slow); reviewing the gatesig tests for BR-3 in the meantime.

```verdict
verdict: SHIP
confidence: high
```

Both `runCIGate` callers now decide from a freshly fetched main. The durable path at `landing.go:388` uses `t.fetchMain`. The legacy path at `merge.go:563-567` now fetches `+refs/heads/main:refs/remotes/origin/main` before the gate. So the rule BR-13 named covers every caller, not just the one it pointed at. BR-3 is fixed: the no-ci `RefusalPat` is now the `CI gate: ` prefix, which every refusal carries. `friction_test.go` has a case for each refusal, including the head-tree read and cannot-pin, plus negative cases for the ack and success lines. One evidence gap remains on BR-13. I reverted the new fetch in a scratch copy and ran `TestRunMerge_GuardOnFreshMainRequiresVerdict` (`go test -count=1`), and it still passed. That is because step 9b's in-place fetch already refreshes `origin/main`; the issue Log says this openly (line 92). Under the ariadne#194 rule, that disposition is therefore `not-addressed` until a regression test exists. It is Minor and the code is correct, so it does not block.

1. **Strengths**
   - `gatesig.go:164`: a single-prefix refusal pattern with a stated invariant ("every refusal, and only refusals, carry this prefix"). It also covers the new `CI gate: git fetch origin main` message without any change. This fixes the whole class, not one site.
   - `friction_test.go:130-144`: one classifier case for every refusal string in `cigate.go`/`merge.go`, plus negative cases for the green and none success lines.
   - `merge.go:563-566`: an explicit refspec fetch, so a linked worktree's `refs/remotes/origin/main` really moves. A failed fetch fails closed with the gate prefix.
   - Issue Log line 92 states honestly what the new test does and does not prove.

2. **Critical**: none.

3. **Important**: none.

4. **Minor**
   - BR-13 (re-disposed `not-addressed`): no test fails without the fix. Fix sketch: add a linked-worktree e2e fixture (`git worktree add` from `tempRepo`, push the guard to main from a second clone, run `runMerge` from the worktree, expect the `local-gates-guard/verdict` refusal). That is the topology where step 9b does not fetch.

5. **Test coverage**
   - Reverting the BR-13 fetch leaves the full `TestRunMerge_GuardOnFreshMainRequiresVerdict` run green (both runs `ok`). The scratch tree was restored and `git status` is clean.
   - BR-3 is covered by the per-refusal classifier rows.

6. **Architecture**
   - ARCH-DRY: pass. `runMerge` has three `fetch origin main` sites, each with its own reason and error text. That is acceptable, though a later `fetchMainFresh(why)` helper would make them one.
   - ARCH-PURE: pass. The gate decision is in `cigate.go`, with git and gh injected.
   - ARCH-PURPOSE: pass. Both gate callers read fresh main (class swept).
   - ARCH-MOCK: pass. The e2e uses a real local git origin and a fake gh behind the client seam.
   - ARCH-CONSTRAINTS: pass. It adds one fetch per merge, on the operator path.
   - ARCH-SECURE: pass. No new untrusted input.
   - ARCH-ORDER: pass. The fetch is ordered before the gate read. The only interleaving gap is the untested worktree topology noted under BR-13.
   - ARCH-FUNERAL: pass. Nothing durable is created.

7. **Plan revisions**: none needed. The issue Log already records the BR-13 approach and its test limitation.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      RefusalPat is now the `CI gate: ` prefix (gatesig.go:164); friction_test.go adds rows for the head-tree read and cannot-pin plus every other refusal and negative success lines.
  - id: BR-13
    disposition: not-addressed
    note: |
      The fix at merge.go:563-566 is correct, but TestRunMerge_GuardOnFreshMainRequiresVerdict still passes with it reverted (checked; the step 9b in-place fetch refreshes the ref). Add a linked-worktree fixture that goes red without the fetch.
```
