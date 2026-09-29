# Boundary Review — ariadne#267 (whole-issue close)

| field | value |
|-------|-------|
| issue | 267 — merge/pr: a PR behind the local branch reads as 'found 0 PRs'; pr re-run pushes then fails |
| repo | ariadne |
| issue file | workshop/issues/000267-stale-pr-head-messages.md |
| boundary | whole-issue close |
| milestone | — |
| window | 73a07a222e4f7c49b90e1b864f98995df0b947b2..7c0c6a42b036b07681fc863a9fb939bc208f9f6a |
| command | sdlc close --issue 267 |
| reviewer | claude |
| timestamp | 2026-09-28T22:35:50-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The change delivers all three Spec items. `selectLandingPR` is now split into an identity pass (`liveLandingPRs`) and a head filter. When no PR matches the local head but exactly one OPEN PR sits at an older head, `merge` refuses with a message naming `sdlc pr`, instead of the old "found 0". `runDurablePR` checks for existing PRs before it pushes. With one open PR it pushes, prints `updated PR #N to <sha>` and exits 0. With more than one it refuses before pushing. After the push it checks `branch.<b>.merge` and warns if it is empty. The Log explains the upstream root cause: the sandbox blocks writes to `.git/config`, and `git push -u` still exits 0. All three new tests go through the real entry points (`runMerge`/`runPR`) on the real-git fixture. `go test ./cmd/sdlc -run 'Landing|PR|FleetPlan'` passes (156.8s). The atlas and the `pr`/`merge` helptext are updated. Nothing blocks shipping; only Minor findings remain.

**1. Strengths**
- `cmd/sdlc/landing.go:189` — `liveLandingPRs` is one shared identity check, used by both `merge` (`selectLandingPR`) and `pr` (`runDurablePR`). The two paths cannot drift apart (ARCH-DRY pass).
- `cmd/sdlc/landing.go:215` — the stale-PR branch fires only when nothing matches the head and exactly one open PR is at another head. The later exact-head checks (`landing.go:377/432`, including the post-merge re-observe) keep their meaning.
- `runDurablePR` refuses when there is more than one open PR *before* it pushes, so the remote is never left changed on that refusal. The test covers this at `landing_test.go:254`.
- The Log records the upstream root cause, which it reproduced with a held `config.lock`. That meets the "or the Log records why not" clause of Done-when.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- The upstream warning guesses the cause instead of reporting it (`landing.go:562`). The Plan says "warn with the cause instead of dropping git's stderr". But `runGitCmd` (`runner.go:50`) still throws away stderr when the command succeeds, so the warning offers a hypothesis ("is .git/config write-protected…?"). Also, `upstream, _ :=` discards a real `git config` error (anything other than exit 1), which then shows up as the same sandbox guess.
- `sdlc pr --dry-run` (`landing.go:525`) still prints `Would: gh pr create` even when an open PR exists. The dry run never queries PRs, so its output no longer matches what a real run does.
- The warning is only tested in its "fires" state. Neither `TestLandingPRUpdatesOpenPR` nor any other happy-path test checks that stderr has no warning when the upstream *was* recorded. A warning that fires every time would go unnoticed.

**5. Test coverage notes**
Each Done-when clause has a test:
- a stale head makes `merge` refuse with the push instruction;
- `pr` with an open PR pushes, exits 0 and names the PR, and more than one open PR refuses;
- an unrecorded upstream warns.

The Log says all three fail against the pre-change `landing.go`. That is plausible: the old `selectLandingPR` returned "found 0", and the old `runDurablePR` always called `PRCreate`. The `fleet_plan_test.go` path fix is an unrelated side-quest and is committed separately.

**6. Architectural notes**
- ARCH-DRY: pass. The identity check is shared; the "keep only OPEN PRs" loop in `runDurablePR` is short and specific to that call.
- ARCH-PURE: pass. `liveLandingPRs` and `selectLandingPR` are pure, and the input/output stays in the `runDurablePR` shell.
- ARCH-PURPOSE: pass. All three Spec items are delivered, and the upstream item is resolved by diagnosing the cause, recording it in the Log, and adding a warning. That is an accepted outcome under Done-when.

**7. Plan revision recommendations:** optionally add a `## Revisions` note that the upstream warning infers its cause, because git's stderr is still dropped on success.

```findings
findings:
  - id: new
    severity: Minor
    family: success-path-stderr-dropped
    title: |
      Upstream warning guesses the cause; git's stderr and non-exit-1 config errors are discarded
    detail: |
      The Plan says "warn with the cause instead of dropping git's stderr", but runGitCmd still drops stderr on success, and `upstream, _ :=` at landing.go:562 discards a real config error, which then shows up as the sandbox guess. Only instance in this window.
  - id: new
    severity: Minor
    family: dry-run-mirrors-real-path
    title: |
      pr --dry-run still announces gh pr create when an open PR would be updated
    detail: |
      landing.go:525 returns before the PR query, so the dry run no longer matches the real path's update branch. Only instance in this window.
  - id: new
    severity: Minor
    family: two-state-clause-coverage
    title: |
      Upstream warning is tested only in its fires state
    detail: |
      No test asserts that stderr has no warning when branch.<b>.merge was recorded (TestLandingPRUpdatesOpenPR discards stderr), so a warning that fires every time would go unnoticed.
```

---

## Re-review — 2026-09-28T22:44:48-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 267 — merge/pr: a PR behind the local branch reads as 'found 0 PRs'; pr re-run pushes then fails |
| repo | ariadne |
| issue file | workshop/issues/000267-stale-pr-head-messages.md |
| boundary | whole-issue close |
| milestone | — |
| window | 73a07a222e4f7c49b90e1b864f98995df0b947b2..df7fedde652d10ea4ca73c0efdb370f411dbe72e |
| command | sdlc close --issue 267 |
| reviewer | claude |
| timestamp | 2026-09-28T22:44:48-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The window does what the Spec asks. `merge` now reports a stale open PR by number and says to push, instead of "found 0". `sdlc pr` pushes into an existing open PR and exits 0 with `updated PR #N to <sha>`, and refuses before pushing if there is more than one open PR. After the push it checks `branch.<b>.merge` and warns if git did not record it. The Log explains the upstream root cause, which satisfies that Done-when clause. The round-2 commit (`df7fedde`) handles all three prior Minors in code. BR-2 and BR-3 are fully resolved. BR-1's new error branch has no test, so it stays open as a Minor. Nothing blocks SHIP.

**1. Strengths**
- `cmd/sdlc/landing.go:189` — `liveLandingPRs` is the one identity check, and both `merge` and `pr` call it (ARCH-DRY pass).
- `landing.go:~215` — the stale-PR branch fires only when nothing matches the local head and exactly one OPEN PR sits at another head. MERGED PRs, ambiguous cases and `head == ""` keep the old exact-match meaning.
- `runDurablePR` refuses more than one open PR before it pushes, so a refusal never changes the remote. The test covers this at `landing_test.go:~254`.
- The Done-when clause about the upstream warning is now tested both ways. `TestLandingPRWarnsUnrecordedUpstream` covers the case where it fires. `TestLandingPRUpdatesOpenPR` covers the quiet case: it asserts no "upstream" on stderr after a real `push -u` records the upstream.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- BR-1 is still open, for test evidence only. The `cerr != nil` branch (`landing.go:~566`) now reports a failed config read instead of hiding it behind the sandbox guess, but no test reaches that branch. A `landingHookRunner.after` that returns an error for `args[0] == "config"` would make it cheap to add. This is still the 1st instance of the `success-path-stderr-dropped` family, not a new one. The push's own stderr is still dropped on success. The warning now presents the sandbox as an example ("as in an agent sandbox") rather than as the known cause, which is acceptable.

**5. Test coverage notes**
- Every Done-when clause has a test that goes through `runMerge` or `runPR` on the real-git fixture.
- The new dry-run wording has no assertion. That is acceptable: it is informational text, and a test would only check that the wording is present.

**6. Architectural notes**
- ARCH-DRY: pass. The small loop in `runDurablePR` that keeps only OPEN PRs is specific to that call.
- ARCH-PURE: pass. `liveLandingPRs` and `selectLandingPR` are pure functions, and IO stays in `runDurablePR`.
- ARCH-PURPOSE: pass. All three Spec items are delivered, and the upstream item is resolved as Done-when allows (a verified warning plus the Log's explanation).

**7. Plan revision recommendations:** none needed. You could optionally note that the upstream warning infers its cause, because git's push stderr is still discarded on success.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      Config-read errors now surface via the cerr branch, but no test reaches it; a landingHookRunner.after failing args[0]=="config" would pin it. Push stderr is still dropped on success.
  - id: BR-2
    disposition: addressed
    note: |
      landing.go dry-run line now names both outcomes (update the open PR, else gh pr create), which matches the real path while staying offline.
  - id: BR-3
    disposition: addressed
    note: |
      TestLandingPRUpdatesOpenPR now captures stderr and fails on any "upstream" text after a real push -u records the upstream, so a warning that always fires goes red.
```
