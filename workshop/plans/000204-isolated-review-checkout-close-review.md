# Boundary Review — ariadne#204 (whole-issue close)

| field | value |
|-------|-------|
| issue | 204 — Run each boundary reviewer in a disposable isolated checkout |
| repo | ariadne |
| issue file | workshop/issues/000204-isolated-review-checkout.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6fdaee9d2ac90d5146c09612063aeecac28b1d5c..d02c457a45dc593b72d7a8918b0fd09ecf7a8a71 |
| command | sdlc close --issue 204 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T08:06:20-07:00 |
| verdict | REWORK |

## Review

Running the reviewbox tests worked. The live fence test skips here because this review is itself nested in a sandbox. Next I'm checking the cmd-level isolation tests and how cancellation is handled.

I've found a likely problem with the prompt wording. Now I'm checking whether one undeletable directory left behind by a crashed review stops every later review from running.

```verdict
verdict: REWORK
confidence: medium
```

The isolation design holds up: a shared clone with no remote, a lock taken before its name is visible, one argv wrap for every agent, and a degrade decision that is recorded on the round and sidecar. The reviewbox and boundary-wiring tests pass here: `go test ./cmd/sdlc/internal/reviewbox/` and `go test ./cmd/sdlc/ -run 'BoundaryReview|Sidecar'`. The live fence test skipped because this review is itself nested in a sandbox.

One defect blocks SHIP. The reviewer prompt `code-review.md` now tells every reviewer it is in a disposable checkout and should revert fixes there. That same template drives `sdlc judge milestone-review`, which this issue left running in the caller's real checkout. On that path the prompt now directs edits into the operator's live tree, which is the parley#237 failure this issue was filed to stop.

There are also three Important problems, each cheap to fix:
- **Shared retry checkout:** a no-verdict retry runs in the same checkout as the first attempt.
- **Poisoned sweep:** one checkout that cannot be deleted makes every later boundary review fail to run.
- **Fenced path never seen working:** the OS-fenced path has not been observed passing anywhere.

1. **Strengths**
   - `reviewbox/checkout.go:63-75`: the lock is held before it is renamed into view, and the sweep relies on the kernel's flock instead of guessing from pids. Close is idempotent.
   - `reviewbox/fence.go:51-66`: `Wrap` is one pure transform shared by all three agents. It is applied last in `BuildArgs` (`judge/dispatch.go:141-148`), so `--dry-run` shows the real command and the existing 61 `judge.Run` fakes keep their signature.
   - `reviewbox/checkout_test.go`: real-git tests cover source-untouched (refs, config, untracked files), a fresh checkout per round, sweeping dead versus live owners, and a create failure leaving nothing behind. The log records that the push test was strengthened after a mutation check showed it passed without the fix.
   - `boundaryOrientation` resolves the repo name before `RepoRoot` moves to the checkout (`milestoneclose.go:791-797`). The `isolation` field is `omitempty`, so existing ledgers stay byte-identical.

2. **Critical**
   - `cmd/sdlc/internal/judge/code-review.md:71-73` (ARCH-PURPOSE, ARCH-SECURE): the unconditional sentence "you run in a disposable checkout of the pinned head… so revert and test there" is also rendered by `cmd/sdlc/judge.go:109-126`.
     - That path is ad-hoc `sdlc judge milestone-review`, including working-tree mode, and it still sets `RepoRoot` to the source checkout.
     - The prompt now actively tells those reviewers to revert fixes in the operator's live tree.
     - Fix: render the sentence only when isolated (a template variable set by `boundaryReviewOptionsFor`), keep "in a scratch copy" otherwise, and pin both renderings with golden tests. Alternatively, isolate the committed-range `judge milestone-review` path as well.

3. **Important**
   - **Retry reuses the checkout** (`milestoneclose.go:669-681` with `judge/dispatch.go:226-236`, ARCH-ORDER): `Dispatch` retries a no-verdict review inside the same checkout.
     - Retries are triggered by #271/#300, which is exactly the case of a reviewer that started revert/mutation work in the background and then ended its turn.
     - Attempt 2 can therefore inspect a tree whose fix is still reverted, and wrongly report the fix missing.
     - Fix: add a per-attempt hook such as `DispatchOptions.BeforeAttempt` that resets the checkout (`checkout --force --detach head`, `clean -fdx`, re-apply the overlays) or recreates it. Add a test whose fake reviewer leaves residue on attempt 1.
   - **One bad checkout blocks all reviews** (`reviewbox/checkout.go:52-54, 104-109, 145-148`, ARCH-FUNERAL): this chain is confirmed.
     - A reviewer leaves a read-only directory in its checkout.
     - `Close` fails on `RemoveAll` but still deletes the lock, leaving a directory with no lock.
     - The next `Create` runs `Sweep`, which returns an error. I reproduced this: `unlinkat …: permission denied`.
     - `Create` aborts, so every later boundary review comes back VerdictNotRun until someone cleans up by hand.
     - Fix: make the sweep best-effort (warn on an orphan, never fail `Create` because of one), `chmod -R u+w` before removing, and keep the lock when `RemoveAll` fails.
   - **Fenced path unverified, protect paths not canonical** (`reviewbox/probe_test.go:21-26`, `reviewisolation.go:64-70`, ARCH-MOCK, ARCH-SECURE):
     - The only live conformance test skips wherever the probe fails, and the log says it has never been seen passing.
     - Protect paths are passed to Seatbelt without symlink resolution. The test's `t.TempDir()` lives under `/var/folders`, which is a symlink to `/private/var`. Seatbelt subpath rules are generally believed to match canonical paths, so the deny rules may not match. I could not confirm this here because nesting is refused.
     - Fix: apply `filepath.EvalSymlinks` to each Protect path and to `Dir`, then run `TestFenceDeniesSourceWrites` once unsandboxed and log the result. Until then the "fenced" claim is unproven.
   - **Deny list widens the agents' write scope** (`reviewisolation.go:64-70`, ARCH-SECURE): under the fence each agent's own sandbox is turned off. Writes are then allowed everywhere except the source worktree and its git common dir.
     - That includes sibling linked worktrees of the same repo (and the main working tree when the source is itself a linked worktree), plus the rest of `$HOME`.
     - Cheap mitigation: add every path from `git worktree list --porcelain` to Protect. Also note the remaining widening in the atlas.

4. **Minor**
   - `judge/dispatch.go:182`: gemini now gets `--sandbox=false`, but the plan says gemini is unchanged. The atlas does describe the flag, so only the plan needs a revision entry.
   - `TestArchitecture_NarrativeRoutesToArchPrinciples` fails inside a review checkout because `AGENTS.md` is untracked, so it is absent from a clean clone. Tests that depend on untracked generated files now give reviewers false failures.
   - `<id>.lock.tmp` files left by a create that was killed mid-way are never swept. Each is tiny, but nothing bounds how many accumulate.
   - `plannedReviewIsolation` and `openReviewIsolation` duplicate the protect-path and root resolution (`reviewisolation.go:22-60`). A small shared helper would remove it.
   - When the issue file path is absolute, the overlay is skipped and the reviewer reads the source path directly. That is read-only and harmless, but undocumented.

5. **Test coverage**
   - Good: the lifecycle, overlay, failure and cancellation cleanup, fresh checkout per round, fenced argv and stamp, and dry-run tests.
   - Missing:
     - retry-attempt residue (Important 1);
     - a sweep across an orphan that cannot be removed (Important 2);
     - a golden test showing the ad-hoc `judge milestone-review` prompt does not claim a disposable checkout (Critical);
     - a live fence run that has actually been observed passing.

6. **Architectural notes**
   - Seatbelt can't nest, so in the usual slot-agent setup the guarantee is checkout-only. Consider failing louder, or tracking how often rounds run as `checkout` in the ledger, so the TL can see whether the fence ever actually applies.
   - An allow-list profile (deny all writes, then allow the checkout, temp space and the agent's state directories) would give a strictly narrower scope than the agents' own sandboxes did.

7. **Plan revision recommendations**
   - "Gemini under Fenced passes `--sandbox=false`." This replaces the plan's "Gemini … unchanged".
   - "`code-review.md` is shared with `sdlc judge milestone-review`, which is not isolated; the disposable-checkout sentence is rendered only for isolated boundary dispatches."
   - "A no-verdict retry gets a reset (or fresh) checkout; 'one round, one checkout' means one per attempt."

**Architecture markers**

| Marker | Result | Note |
|---|---|---|
| ARCH-DRY | pass | Minor duplication only |
| ARCH-PURE | pass | |
| ARCH-PURPOSE | flag | Prompt contract spills into a path that isn't isolated |
| ARCH-MOCK | flag | Live conformance never observed passing |
| ARCH-CONSTRAINTS | pass | Clone measured at about 70 ms |
| ARCH-SECURE | flag | Protect paths not canonical; sibling worktrees unprotected |
| ARCH-ORDER | flag | Retry residue |
| ARCH-FUNERAL | flag | A sweep failure poisons every later review |

```findings
findings:
  - id: new
    severity: Critical
    family: isolation-claim-on-unisolated-path
    title: |
      code-review.md tells every reviewer it runs in a disposable checkout, but ad-hoc sdlc judge milestone-review still runs in the source checkout
    detail: |
      cmd/sdlc/judge.go:109-126 renders the same template with RepoRoot set to the live checkout, so the prompt now tells those reviewers to revert fixes in the operator's tree (the parley#237 failure). Render the sentence only for isolated dispatches; keep the scratch-copy wording otherwise; pin both renderings with golden tests.
  - id: new
    severity: Important
    family: retry-attempt-shares-mutable-state
    title: |
      No-verdict retry in judge.Dispatch reuses the same checkout, so attempt 2 can inspect attempt 1's reverted or mutated tree
    detail: |
      Retries fire exactly when a reviewer ended mid-work (#271/#300). Add a per-attempt reset or recreate hook (checkout --force head, clean -fdx, re-overlay) and a test with residue left by attempt 1.
  - id: new
    severity: Important
    family: sweep-failure-blocks-dispatch
    title: |
      One checkout that cannot be removed makes Sweep fail, which fails Create and turns every later boundary review into NotRun
    detail: |
      Reproduced: a read-only directory in an orphan makes Sweep return unlinkat permission denied. Close also deletes the lock when RemoveAll fails, creating exactly that lockless orphan. Make the sweep best-effort with a warning, chmod u+w before removing, and keep the lock when RemoveAll fails.
  - id: new
    severity: Important
    family: fence-unverified-live
    title: |
      The fenced path has never been seen passing live, and protect paths are not resolved through symlinks before reaching Seatbelt
    detail: |
      TestFenceDeniesSourceWrites skips wherever the probe fails and has not been seen green. Its t.TempDir paths under /var are symlinks, and Seatbelt subpath rules are generally believed to match canonical paths. EvalSymlinks the Protect paths and Dir, run the test once unsandboxed, and log the result.
  - id: new
    severity: Important
    family: fence-scope-wider-than-agent-sandbox
    title: |
      The deny-list fence turns off the agents' own sandboxes but protects only the source worktree and git common dir, leaving sibling worktrees and HOME writable
    detail: |
      Add every path from git worktree list --porcelain to Protect, or move to an allow-list profile; note the remaining widening in the atlas.
  - id: new
    severity: Minor
    family: plan-code-drift
    title: |
      Fenced gemini gets --sandbox=false while the plan says gemini is unchanged
  - id: new
    severity: Minor
    family: tests-depend-on-untracked-files
    title: |
      TestArchitecture_NarrativeRoutesToArchPrinciples fails in a review checkout because AGENTS.md is untracked
  - id: new
    severity: Minor
    family: artifact-without-removal-path
    title: |
      lock.tmp files left by a killed Create are never swept
```

---

## Re-review — 2026-10-11T08:40:53-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 204 — Run each boundary reviewer in a disposable isolated checkout |
| repo | ariadne |
| issue file | workshop/issues/000204-isolated-review-checkout.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6fdaee9d2ac90d5146c09612063aeecac28b1d5c..388cceb8bbf6cdce4b428171d7705d3801b785cf |
| command | sdlc close --issue 204 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T08:40:53-07:00 |
| verdict | SHIP |

## Review

Let me look at the fix commit primarily, plus the core files.

Checking whether this review is running under the sdlc fence or only in its checkout, since a write to the source git dir just succeeded.

All prior fixes go red when reverted. Now checking whether the retry reset can race with background processes left over from attempt 1.

Wrapping up: all eight prior findings verified, and two new Minor issues found.

```verdict
verdict: SHIP
confidence: medium
```

All eight prior findings are fixed. For each behaviour fix I reverted the change in this checkout and confirmed its regression test went red, then restored it. The prompt only says "revert in place" when the review really runs in an isolated checkout. A retry now starts from a fresh clone. A sweep problem is reported as a warning and no longer stops the review. `Close` keeps the lock file when it can't remove a checkout. The fence's protected paths are canonicalized and now cover every worktree. Two things remain. The fenced path has still never been seen working in a real run. The probe now checks that the fence actually denies a write before it reports one, so a round can't claim to be fenced when it isn't. Nothing here blocks SHIP. I found two new Minor issues. Confidence is medium because the full `cmd/sdlc` suite hit Go's 10-minute timeout here and never finished.

**1. Strengths**
- `reviewbox/probe.go:44-84`: the probe runs a write through the same `Isolation.Wrap` the review uses. It must succeed in an allowed scratch directory and be refused in a protected one before it reports a fence. So a `fenced` stamp can't be recorded unless a write was actually refused first, and a fence that denies nothing is rejected (`TestProbeRejectsAFenceThatDoesNotDeny`).
- `judge/review.go:57-65`: the revert instruction (`{{REVERT_CHECK}}`) depends on `PromptInput.Isolated`, so unisolated prompts are unchanged byte for byte. `TestRevertCheckFollowsIsolation` covers all three review recipes.
- `reviewisolation_test.go:264` is an end-to-end retry test. Attempt 1 changes the checkout and leaves extra files behind, and the test asserts attempt 2 sees neither. When I removed `opts.BeforeRetry = checkout.Reset`, it failed with "attempt 2 read attempt 1's mutation".
- `checkout.go`: `Close` keeps the lock file when removal fails, `removeTree` restores write permission first, and `.lock.tmp` files are swept. Each has a test that fails without it. `TestCloseFailureLeavesLockForSweep` uses `chflags uchg` to make a file that can't be removed, so the failure path is real.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- **Sweep can make a concurrent `Create` fail** (`reviewbox/checkout.go:66-75`, family `sweep-failure-blocks-dispatch`, ARCH-ORDER).
  - The comment says the lock is held "before its name is visible", but that is no longer true. `<id>.lock.tmp` is visible between `OpenFile` and `flock`, and since the BR-8 fix the sweep reclaims unheld `.lock.tmp` files.
  - If another review sweeps in that window, it deletes the tmp file. The first review's `Create` then fails at `rename` with ENOENT and that review ends as NotRun.
  - **This is the 2nd finding in family `sweep-failure-blocks-dispatch`.** The rule that covers both: the sweep may only act on a name whose owner is provably dead, so every name it acts on must be locked from the moment it exists. Fix that rule, not this site: either take the lock in the same call that creates the file (`O_EXLOCK` on darwin), or have the sweep skip a `.lock.tmp` younger than a short grace period. Then add a test that has a sweep reclaim the tmp file just before `Create` takes its lock (the window before `flock` is where this goes wrong).
- **Tests check a dispatch path production no longer uses** (`milestoneclose.go:747`). `boundaryReviewDispatchOptions` (no isolation) is now called only from tests (`closereview_test.go:718,755`, `smalldiffreplay_test.go:45`). Point those tests at `boundaryReviewOptionsFor` with an isolation, or delete the helper.

**5. Test coverage notes**
- Results:
  - `reviewbox` and `judge` pass.
  - `go test ./cmd/sdlc -run 'BoundaryReview|…'` passes.
  - The full `go test ./cmd/sdlc/...` run timed out on `cmd/sdlc` after 600s.
  - `processgroup` fails with `fork/exec /bin/ps: operation not permitted`, which is this sandbox, not the diff.
- So I can't say whether any other `cmd/sdlc` test depends on gitignored files, which is BR-7's class.
- `TestFenceDeniesSourceWrites` skipped here: `sandbox_apply` was refused.
- This round itself ran checkout-only: `SANDBOX_RUNTIME=1`, the reviewer's own sandbox couldn't nest, and the source `.git` was writable from inside the review. That matches the documented fallback ("slot agents run clone-only"). It also means that from slots, the fence protecting the source is effectively never on.

**6. Architectural notes**
- **ARCH-DRY: pass.** One `Wrap` is shared by all agents and dry-run, and `populate` is shared by `Create` and `Reset`.
- **ARCH-PURE: pass.** `Wrap`, `seatbeltProfile` and `revertCheck` are pure. Git and filesystem access stays in the thin `reviewisolation.go` and `checkout.go` layer.
- **ARCH-MOCK: pass, with a note.** Tests use real git in temporary repos. The live check (`TestFenceDeniesSourceWrites`) has no scheduled unsandboxed run, so the fenced path has never been seen working. Schedule one.
- **ARCH-CONSTRAINTS: pass.** The probe is one spawn per dispatch, and `Reset` is a shared clone.
- **ARCH-SECURE: pass.** Overlay paths are checked, Seatbelt strings are escaped, and protected paths are canonical. HOME stays writable under the fence, and the atlas says so.
- **ARCH-ORDER: flag.** See the Minor finding about the `.lock.tmp` window. Retry versus attempt-1 leftovers is handled: `runProcess` kills the process group before `BeforeRetry` runs.
- **ARCH-FUNERAL: pass.** Clones are removed in a defer, crashed owners' clones are swept on every dispatch, and the probe's scratch directory is removed. An entry the sweep can never remove is warned about on every review and never grows by itself.
- **ARCH-PURPOSE: pass.** I swept for "disposable checkout" claims, and they now appear only on isolated paths.

**7. Plan revision recommendations:** none needed. The Revisions entry dated 2026-10-11 matches the code, including correcting "gemini unchanged".

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      revertCheck(in.Isolated) at judge/review.go:48; TestRevertCheckFollowsIsolation fails when Isolated is forced true; goldens keep the scratch-copy text.
  - id: BR-2
    disposition: addressed
    note: |
      BeforeRetry=checkout.Reset; TestBoundaryReviewRetryStartsFromAPristineCheckout fails without it ("attempt 2 read attempt 1's mutation").
  - id: BR-3
    disposition: addressed
    note: |
      Sweep is best effort and runs outside Create, removeTree restores write permission, Close keeps the lock on failure; each has a test that fails when its change is reverted.
  - id: BR-4
    disposition: addressed
    note: |
      Paths are canonicalized; the probe now proves a denial through the production Wrap before reporting fenced, so a false stamp cannot be recorded. The live test is still unrun, noted under ARCH-MOCK.
  - id: BR-5
    disposition: addressed
    note: |
      Every git worktree list path is protected; TestBoundaryReviewFencedArgvAndStamp fails without it; the atlas states the HOME widening.
  - id: BR-6
    disposition: addressed
    note: |
      The plan Revisions entry corrects "gemini unchanged" to match --sandbox=false.
  - id: BR-7
    disposition: addressed
    note: |
      The test reads the tracked AGENTS.base.md and passes in this clean clone.
  - id: BR-8
    disposition: addressed
    note: |
      Sweep reclaims unheld .lock.tmp; TestSweepReclaimsDeadOwnerAndSparesLiveOne fails without it.
findings:
  - id: new
    severity: Minor
    family: sweep-failure-blocks-dispatch
    title: |
      A concurrent sweep can reclaim a just-created <id>.lock.tmp before Create locks it, so that review fails as NotRun
    detail: |
      2nd finding in this family. Rule: the sweep may only act on a name whose owner is provably dead, so every name it acts on must be locked from creation. Fix the rule: create the lock file and lock it in one call (O_EXLOCK on darwin), or skip .lock.tmp younger than a short grace period. Then test a sweep reclaiming the tmp file just before Create locks it. The comment at checkout.go:66 is now false.
  - id: new
    severity: Minor
    family: tests-pin-unused-path
    title: |
      boundaryReviewDispatchOptions (no isolation) is now called only by tests
    detail: |
      closereview_test.go and smalldiffreplay_test.go check prompts that production never sends; point them at boundaryReviewOptionsFor with an isolation, or delete the helper.
```
