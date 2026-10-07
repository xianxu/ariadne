# Boundary Review — ariadne#301 (whole-issue close)

| field | value |
|-------|-------|
| issue | 301 — Pin the rebase-aware close rule: e2e test and atlas |
| repo | ariadne |
| issue file | workshop/issues/000301-close-survives-rebase-followup.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3b72bffdb5f5f15473154dc40832c8e0e1d47d9d..3451f75803d3e981212d0dc497c49c80f0f1b67e |
| command | sdlc close --issue 301 |
| reviewer | claude |
| timestamp | 2026-10-07T13:23:06-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** This is a small window that delivers the issue's purpose. I ran the new end-to-end test `TestTrackerCloseSurvivesARebase` in a scratch worktree: it passes, and all three variants are green. I then checked the main Done-when clause myself instead of relying on the issue Log. Swapping `env.closeAncestorOf` for `env.ancestorOf` at `cmd/sdlc/closetracker.go:418` fails exactly the "close, old review pruned" variant. Swapping it at `cmd/sdlc/issuerecovery.go:164` fails exactly the "reconcile, old review pruned" variant. Both callers are now pinned. The atlas states the rule, the recovery catalog cites the new test, and the project file records the scope change. I did not run the full `make test`; I ran only the targeted test, so that Done-when clause rests on the implementor's run.

**1. Strengths**
- The revert check holds and each variant guards its own caller (`closetracker_test.go:298-306`). The Log admits that the first draft missed the closetracker revert and explains why the pruned SHIP variant was added. The code confirms that explanation.
- The test checks its own fixture before trusting it (`closetracker_test.go:326`, `:332`):
  - the rebase must actually rewrite the reviewed commit off the branch;
  - the prune must actually remove that commit.
  
  Without these checks the test could pass for the wrong reason.
- The final assertion checks a new token, an evidence commit equal to `HEAD^` (the existing `evidenceRev` convention), and `status: codecomplete`. That is the whole Spec claim ("lands, bound to an evidence commit on the branch"), not just the absence of an error.
- The atlas addition (`atlas/workflow/issue-tracker.md:280-286`) is accurate against the wiring: it covers ancestry, the case where a rebase rewrote the close off the branch, `closeAncestorOf` treating an unknown commit as "not an ancestor", and the rule that a stale receipt still loses. It sits beside the binding description in § Readers and completion.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- The Spec asks for "one sentence" in the atlas; the addition is four sentences. It is correct and readable, so this is not worth changing.
- `gitSucceeds` (`closetracker_test.go:361`) is a new helper. Existing helpers (`gitOut`, `gitIn`, `trackerRepo.git`) all fail the test on a non-zero exit, so none of them can serve as a boolean probe. The new helper is justified and is not duplication.

**5. Test coverage**
- The test covers close and reconcile, each with the old reviewed commit pruned, plus a SHIP variant with it present.
- It does not cover reconcile with the old commit present. That combination is already reached by plain ancestry and by the existing unit tests, so it is not needed.
- The "stale pre-rebase receipt still loses" half stays unit-pinned by `TestNewestCloseAfterARebase`. That is appropriate.

**6. Architecture**
- **ARCH-DRY: pass.** The test reuses `closeReady`, `peerCommit`, `stubJudge`, `runRecoveryReconcile` and `evidenceRev`. The new helper has no existing equivalent.
- **ARCH-PURE: pass.** This is an integration test over real git, which is the right tier for a wiring check. The pure rule keeps its own unit test.
- **ARCH-PURPOSE: pass.** The class here is "callers of the close-generation rule". Grepping `NewCompletionOp(` finds exactly two call sites, and both are pinned. Nothing is deferred.

**7. Plan revisions:** none.

```findings
findings:
  - id: new
    severity: Minor
    family: spec-scope-wording
    title: |
      The atlas addition is four sentences where the Spec asked for one
    detail: |
      atlas/workflow/issue-tracker.md:280-286. The content is accurate; only the length differs from the Spec. No change needed.
```
