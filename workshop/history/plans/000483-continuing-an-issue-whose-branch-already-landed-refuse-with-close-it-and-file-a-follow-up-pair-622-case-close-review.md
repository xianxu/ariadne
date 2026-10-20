# Boundary Review — ariadne#483 (whole-issue close)

| field | value |
|-------|-------|
| issue | 483 — Continuing an issue whose branch already landed: refuse with 'close it and file a follow-up' (pair#622 case) |
| repo | ariadne |
| issue file | workshop/issues/000483-continuing-an-issue-whose-branch-already-landed-refuse-with-close-it-and-file-a-follow-up-pair-622-case.md |
| boundary | whole-issue close |
| milestone | — |
| window | 48c1d7a3c8f766bcd3e2662a6354b6d1cfb287ad..3c683e9cd174f608896220678ea12328eedd20f2 |
| command | sdlc close --issue 483 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-20T06:19:11-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still reviewing: the broad test run timed out, so I'm rerunning just the #483 tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The change does what the Spec's option (1) asks for. `start-plan` and `change-code` now refuse an issue whose branch has landed, and the refusal names the recovery, including `sdlc issue new '<title>' --deps N`. `close` now accepts a follow-up branch named with the issue's id prefix, but only after `refuseUnlandedBase` confirms it carries no other issue's unlanded work. The pair#622 shape, where the canonical branch was squash-merged and deleted and the evidence sits on `000009-live-acceptance`, is reproduced in a test. All eight targeted tests pass at head. The broader `-run 'Landed|Close|StartPlan|…'` run hit Go's 10-minute default timeout, so it says nothing either way. One cheap correctness gap stops a plain SHIP: the "landed" check only looks at the local copy of the canonical branch, never the remote one. A slot where that branch is absent locally therefore reads the issue as landed even when the remote branch has unlanded commits. The fix is a one-line reuse of `issueTip`.

1. **Strengths**
   - `landedissue.go:17` `landedPR` is a small pure decision, and an OPEN PR overrides any MERGED one. It is unit-tested in a table.
   - The close follow-up admission reuses `refuseUnlandedBase` (`trackerbranch.go:43`) instead of a new check (ARCH-DRY). `TestCloseRefusesFollowUpCarryingAnotherIssue` pins the guard against another issue's work riding along (ARCH-SECURE).
   - Using `LandingPRs` with `HeadOID` correctly handles squash merges, where a kept local branch's commits never appear on main by SHA. `TestLocalCanonicalAheadOfMergedHeadIsNotLanded` covers both sides of that.
   - Failures fail safe. Any gh, git or URL error reads as not landed, so the older refusals still apply (`TestCloseKeepsTodaysRefusalWhenNotLanded`, "gh error" case).
   - `start-plan` only checks issues that have already started, so a first `start-plan` makes no `gh` call (ARCH-CONSTRAINTS).

2. **Critical findings**: none.

3. **Important findings**
   - **The landed check ignores the remote canonical branch** (`landedissue.go:61-66`). `issueBranchLanded` only reads `refs/heads/<name>`, and when that is empty it returns landed. Failure scenario:
     - PR 406 squash-merges partway through the issue.
     - The worker keeps going in slot A and pushes new commits to `origin/000009-…`.
     - In slot B, which has no local branch, `sdlc start-plan --issue 9` refuses with "already landed", even though the remote holds unlanded work. Before this change it would have restored the branch.
     - The same false positive lets `close` accept a prefixed follow-up branch in slot B.

     The plan's own rule ("unlanded local commits → old path") is only half enforced. Fix: use `issueTip(env, name)`, which falls back to the remote ref, in place of the raw `for-each-ref` (this also removes duplicated logic, ARCH-DRY). Add a test where the local branch is deleted and the remote branch is ahead of `HeadOID`.

4. **Minor findings**
   - `landedPR` keeps the first MERGED PR, so when a branch has several merged PRs it depends on `gh pr list` returning newest first. A kept local tip at a later PR's head could then read as not landed. That direction is safe, but the assumption is undocumented.
   - `closeBranchRecoveryNote` (`trackerbranch.go:75`) says "already landed" whenever the tip differs from HEAD. It relies on `requireCloseIssueBranch` having admitted the branch first, and there is no comment stating that coupling.
   - Plan step 2 names "gh error → old path" tests for `start-plan`/`change-code`. Only `close` covers the gh-error case; the predicate is shared, so this is a coverage nit.

5. **Test coverage notes**
   - The tests use a fake behind the production `landingGH` seam, plus a GitHub-shaped remote URL with an offline `insteadOf` rewrite. Production and test therefore share the same boundary (ARCH-MOCK passes).
   - Missing: the case where the remote branch is ahead and there is no local branch (see Important finding).

6. **Architecture, marker by marker**

   | Marker | Result | Note |
   |---|---|---|
   | ARCH-DRY | pass, with one note | `issueTip` should also be reused by `issueBranchLanded` |
   | ARCH-PURE | pass | Pure `landedPR`, thin IO wrapper |
   | ARCH-PURPOSE | pass | Both verbs refuse, and the close path that motivated the issue (pair#622) is unblocked |
   | ARCH-MOCK | pass | |
   | ARCH-CONSTRAINTS | pass | One `gh pr list` per check, on infrequent verbs |
   | ARCH-SECURE | pass | The id prefix alone is not trusted; the unlanded-base check backs it up |
   | ARCH-ORDER | pass | No state is held; each verb re-derives "landed" at call time |
   | ARCH-FUNERAL | pass | Nothing durable is created |
   | ARCH-EVOLVE | pass | No persisted format changes |

   Docs: the atlas lifecycle entry #10 was added. README needs no change because there is no new flag or subcommand.

7. **Plan revision recommendations**
   - Step 1 names the pure decision `landedIssue(mergedPR bool, localTip, onMain)`. The code has `landedPR(prs []landingPR)` plus the IO wrapper `issueBranchLanded`. Add a `## Revisions` entry recording the rename and the switch to `LandingPRs`/`HeadOID` (already explained in the Log).
   - After the fix, reword the landed definition to "the canonical branch, local or remote, is absent or at/behind the merged head or main".

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: landed-evidence-ignores-remote-branch
    title: |
      issueBranchLanded reads only the local canonical ref; an absent local branch reads as landed even when the remote branch holds unlanded commits
    detail: |
      landedissue.go:61-66 uses for-each-ref refs/heads/<name> and returns landed when empty. In a slot without the local branch, a remote canonical branch pushed past the merged PR head (mid-issue squash landing, then continued work) is ignored: start-plan refuses instead of restoring it, and close admits a prefixed follow-up branch. Fix by using issueTip (local, else remote) and add a test where the local branch is deleted and the remote is ahead of HeadOID.
  - id: new
    severity: Minor
    kind: defect
    family: plan-entity-drift
    title: |
      Plan step names landedIssue(mergedPR, localTip, onMain); code delivers landedPR(prs) plus the issueBranchLanded wrapper
    detail: |
      Add a Revisions entry so the plan matches the delivered entities and the LandingPRs/HeadOID evidence source.
  - id: new
    severity: Minor
    kind: defect
    family: implicit-precondition-coupling
    title: |
      closeBranchRecoveryNote claims already landed on any tip mismatch, relying on requireCloseIssueBranch having admitted the branch first
    detail: |
      trackerbranch.go:75; document the precondition or gate the message on issueBranchLanded.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1077074f-152d-4a25-afec-931671823f12","repository":"github.com/xianxu/ariadne","issue_id":"000483","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/e829a2102610eb6e","reviewed_head":"3c683e9cd174f608896220678ea12328eedd20f2","slot":"ariadne:5","requested_at":"2026-10-20T13:07:23.179361Z"},"phase":"finished","pid":66692,"started_at":"2026-10-20T13:07:23.386153Z","finished_at":"2026-10-20T13:19:10.676064Z","session_id":"3184c72e-9656-425e-9fb0-05b6fe39d7cd","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T07:00:40-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 483 — Continuing an issue whose branch already landed: refuse with 'close it and file a follow-up' (pair#622 case) |
| repo | ariadne |
| issue file | workshop/issues/000483-continuing-an-issue-whose-branch-already-landed-refuse-with-close-it-and-file-a-follow-up-pair-622-case.md |
| boundary | whole-issue close |
| milestone | — |
| window | 48c1d7a3c8f766bcd3e2662a6354b6d1cfb287ad..69d2987ceff329b44ff50f7c8478a823c80698e2 |
| command | sdlc close --issue 483 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-20T07:00:40-07:00 |
| verdict | SHIP |

## Review

Targeted tests pass. Next I'm reverting the BR-1 fix to confirm its regression test goes red.

```verdict
verdict: SHIP
confidence: high
```

All three round-1 findings are fixed, and the regression test for the fix that changes behavior fails without it. BR-1 is fixed: `issueBranchLanded` now reads the issue branch through `issueTip` (the local branch, else the remote one). `TestRemoteCanonicalAheadOfMergedHeadIsNotLanded` covers it. I put the old local-only `for-each-ref` read back in this checkout and that test failed (`landedissue_test.go:177: a remote canonical branch ahead of the merged head read as landed`). After restoring the code, the targeted tests for this issue pass, including the `TestIssueHintsRenderDecimalIDs` guard. BR-2 is fixed by a `## Revisions` entry that matches the code, and BR-3 by a documented precondition on `closeBranchRecoveryNote`. Nothing blocks SHIP. The full `cmd/sdlc` package did not finish within `go test`'s default 10-minute timeout here. That is a timeout, not a test failure: the only FAIL lines are the package-level ones, and no individual test failed.

Done-when: current — `start-plan` and `change-code` both refuse a landed issue and name `sdlc issue new … --deps N` as the path forward. `TestStartPlanAndChangeCodeRefuseLandedIssue` pins this. The Revisions entry changes only how "landed" is detected; what the refusal does is unchanged.

1. **Strengths**
   - `landedissue.go:17` `landedPRs` is a small pure decision function, tested by a table with no IO. The IO wrapper is thin and treats every failure to read the evidence as "not landed", which keeps the older refusals (the safe direction).
   - `trackerbranch.go:61` `issueTip` pulls the "local, else remote" read into one helper used in three places (close admission, the landed check, the recovery note). That is the rule-level answer to the BR-1 family (ARCH-DRY).
   - The close admission reuses `refuseUnlandedBase`, so a follow-up branch with only the issue's prefix can't bring in another issue's work (ARCH-SECURE). `TestCloseRefusesFollowUpCarryingAnotherIssue` pins this.
   - `landedS09` rebuilds pair#622's exact shape (squash merge, branch deleted, follow-up branch holding only evidence) behind the real `landingGH` seam, with an offline `insteadOf` rewrite for the remote URL (ARCH-MOCK).
   - `landedRecovery` is one shared message for all three verbs.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `trackerbranch.go:42-47`: when the branch has the issue prefix but the issue isn't landed, `issueBranchLanded` runs twice (once directly, then again inside `withLandedRecovery`), so that refusal costs two `gh pr list` calls. This only happens on the refusal path, so I'm not filing it.

5. **Test coverage notes:** The tests cover:
   - landed, open-PR and gh-error outcomes;
   - a local branch ahead of the merged head, and a remote branch ahead of it;
   - a later merged PR's head counting as landed;
   - admission of a branch with the issue prefix, refusal when it carries another issue's work, and refusal of an unprefixed branch with the recovery text.

   There is no test for `issueTip` returning an error, but that path falls back to "not landed", which is benign.

6. **Architecture**
   - ARCH-DRY: pass (`issueTip`, `refuseUnlandedBase` and `landedRecovery` are each shared).
   - ARCH-PURE: pass. The decision is pure; the predicate that does IO is a thin wrapper.
   - ARCH-PURPOSE: pass. Both verbs refuse, and close is admitted from a follow-up branch, which unparks pair#622. The BR-1 class (every read of the issue branch's state for "landed") now goes through `issueTip`.
   - ARCH-MOCK: pass. `landedFakeGH` sits behind the production `landingGH` seam, and git runs against real local repos.
   - ARCH-CONSTRAINTS: pass. There is one gh call per check: `start-plan` checks only started cards, while `change-code` checks on every run, which the plan accepts for an infrequent verb.
   - ARCH-SECURE: pass. A branch prefix alone isn't trusted, and gh output comes through the typed `LandingPRs`.
   - ARCH-ORDER: pass. No state is held between events; "landed" is worked out again at each call.
   - ARCH-FUNERAL: pass. Nothing durable is created.
   - ARCH-EVOLVE: pass. No persisted format changes.

7. **Plan revisions:** none. The existing Revisions entry matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      landedissue.go:58 now uses issueTip (local else remote); TestRemoteCanonicalAheadOfMergedHeadIsNotLanded goes red when the local-only for-each-ref read is restored (verified by revert in this checkout).
  - id: BR-2
    disposition: addressed
    note: |
      Issue Revisions 2026-10-20 names landedPRs(prs) plus issueBranchLanded and LandingPRs/HeadOID as the evidence source, matching landedissue.go:17-79.
  - id: BR-3
    disposition: addressed
    note: |
      trackerbranch.go:70-72 doc comment states the precondition (describes only branches requireCloseIssueBranch admitted); requireCloseIssueBranch admits a tip mismatch only after issueBranchLanded.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"392fba76-64cd-4feb-9092-f723d031abf3","repository":"github.com/xianxu/ariadne","issue_id":"000483","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/b501ea97e89cb8ef","reviewed_head":"69d2987ceff329b44ff50f7c8478a823c80698e2","slot":"ariadne:5","requested_at":"2026-10-20T13:48:13.779615Z"},"phase":"finished","pid":87293,"started_at":"2026-10-20T13:48:13.867911Z","finished_at":"2026-10-20T14:00:39.239899Z","session_id":"3692e931-8925-4998-a303-40a49d74bbd2","outcome":"returned"}] -->
