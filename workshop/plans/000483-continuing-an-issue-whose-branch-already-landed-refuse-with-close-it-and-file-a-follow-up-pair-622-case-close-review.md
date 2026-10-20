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
