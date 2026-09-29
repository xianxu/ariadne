# Boundary Review — ariadne#266 (whole-issue close)

| field | value |
|-------|-------|
| issue | 266 — sdlc id lint cannot resolve publication target in CI checkouts |
| repo | ariadne |
| issue file | workshop/issues/000266-id-lint-ci-publication-target.md |
| boundary | whole-issue close |
| milestone | — |
| window | 70fd1077511c3f526817e9b910384aa44046cbec..65a903603c8562221b70283a58bd9f474620f6e9 |
| command | sdlc close --issue 266 |
| reviewer | claude |
| timestamp | 2026-09-28T21:52:07-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The fix targets the right place. The CI failure happened because `cardlessAdditions` → `tracker.RepositoryForCheckout` → `RepositoryFor` resolved the resting branch's publication target (`branch.main.remote`/`.merge`). A detached `actions/checkout` has no such config, so the lint could not tell which remote to read. `--remote NAME` now opens `tracker.NewRepository` directly. The script passes `--remote origin`, so the trunk and the cards come from the same remote. Without the flag, operators keep the existing path. The regression test runs the real `40-duplicate-issue-id.sh` in a fresh-init, fetched, detached checkout with no local `main`, and checks both outcomes: carded details exit 0, a hand-made file exits 1. I ran `TestDuplicateIDCheckRunsInADetachedCICheckout` and `TestLintIDs*` and they pass. Nothing blocks the ship. The one gap worth closing first is that the new branch's fail-closed claim ("a marked repository whose `origin` has no tracker still exits 2") has no test. I traced the code and the claim holds today, but no test protects it.

1. **Strengths**
   - `cmd/sdlc/issuelintids.go:244-250`: the change is small and branches at the only point that needed it. Resolution without the flag is unchanged, so operators see no difference. After either branch, `GuardCutoverAt(dirs.Top, headRef)` still applies the head-side cutover guard. The `RepositoryFor` path had used `GuardCutover(root)`, and it is overridden here as before, so the new path loses no guard.
   - The new path works in shallow CI checkouts because `TrunkFile.Snapshot` → `refreshTip` → `fetch` (`gitx/trunkfile.go:197`) fetches `refs/heads/issue-tracker` itself. The test's full `+refs/heads/*` fetch is therefore not hiding a dependency.
   - It fails closed. A missing tracker ref on the remote makes the fetch fail, which produces `offlineError`, then `degraded`, then exit 2. It never reports clean.
   - The script already exits 0 when there is no `origin` (`40-duplicate-issue-id.sh:45`), so `--remote origin` never names a remote that doesn't exist.
   - The test drives the real script with the owner-repo `dev-aliases.sh` indirection, so it covers the script and the flag together, as consumer CI runs them.

2. **Critical:** none.

3. **Important**
   - The `--remote` path's fail-closed branch (marked repository, `origin` with no `issue-tracker` branch → exit 2) is untested. This is the #213 invariant ("a check that did not look must not report clean"). The Spec, the atlas and the script comment all state it, but only by tracing through `gitx`. A later change that made the tracker read tolerant of a missing branch, for example through `ReadDegraded` or a `nil, nil` return for "no tracker", would turn CI green without anyone noticing. Fix: in `TestDuplicateIDCheckRunsInADetachedCICheckout`, or a sibling test, delete `issue-tracker` from the origin (`git push origin :issue-tracker`) or point the CI clone at a remote without it, then assert exit 2. Other claims of this kind in the window: the same sentence appears in `atlas/workflow/ci-merge-check.md:98` and in the Spec. Both depend on this one behavior, so one test covers all three.

4. **Minor**
   - The first `## Done when` clause ("parley.nvim merge-check passes on a PR run") and the third Plan box can only be checked after merge, as the Log says. When closing, say so in `--verified`, or add a `## Revisions` entry saying that clause is verified after merge. Otherwise the close claims a Done-when that nobody observed.
   - The Log has an empty duplicate `### 2026-09-28` heading.
   - `issuelintids.go:244-245`: the `var repo …; var err error` pair followed by an if/else is fine. A small `openTracker(ctx, top, remote)` helper would read more cleanly, but this is optional.

5. **Test coverage notes:** Done-when clause 2 ("a detached CI-style checkout with a fetched tracker") is covered, with both outcomes (0 and 1) asserted. The Log says the test exited 2 with the exact CI message before the fix, which matches the code: a detached checkout with no `branch.main.*` makes `ResolvePublicationTarget` fail while `FetchedTracker` is true, and that goes to the "publication target is unusable" error. I did not re-run the test with the fix reverted. The fail-closed case of the new branch is missing (see Important).

6. **Architectural notes**
   - **ARCH-DRY: pass.** The script still hands all logic to the verb. The remote is named once, as `origin`, in the script, which already uses `origin` for the trunk. The verb reuses `tracker.NewRepository`, the same constructor `RepositoryFor` uses.
   - **ARCH-PURE: pass.** The IO seam is unchanged. `addedDetails` and `detailsWithoutCards` stay pure, and the new branch only chooses which repository constructor to call.
   - **ARCH-PURPOSE: pass.** The purpose was to make the read-only lint independent of publication config in CI, and the only CI consumer, `40-duplicate-issue-id.sh`, now passes the flag in both of its invocations (the fallback-base and trunk-base forms). No other script calls `lint-ids`. The one other reference, `testdata/legacy-equivalence.sh`, runs the operator path on purpose.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry saying the parley.nvim PR-run evidence (Done-when clause 1, Plan box 3) is verified after merge, because parley's script symlinks into ariadne and its CI clones ariadne `main`.

```findings
findings:
  - id: new
    severity: Important
    family: fail-closed-path-untested
    title: |
      The --remote path's fail-closed case (marked repo, origin without an issue-tracker branch, exit 2) has no test
    detail: |
      Spec, atlas/workflow/ci-merge-check.md:98 and the script comment all claim a marked repository whose origin carries no tracker still exits 2. It holds today only because TrunkFile.fetch fails and becomes offlineError, then degraded. Add a case to TestDuplicateIDCheckRunsInADetachedCICheckout that deletes origin's issue-tracker branch and asserts exit 2, so a tolerant tracker read cannot silently turn CI green.
  - id: new
    severity: Minor
    family: done-when-unobserved-at-close
    title: |
      Done-when clause 1 (parley.nvim PR run green) can only be verified after merge
    detail: |
      Plan box 3 is unchecked and the Log explains why. Record this in a Revisions entry or in close --verified so the close does not claim an unobserved Done-when.
  - id: new
    severity: Minor
    family: log-hygiene
    title: |
      Duplicate empty "### 2026-09-28" heading in the issue Log
```

---

## Re-review — 2026-09-28T21:54:14-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 266 — sdlc id lint cannot resolve publication target in CI checkouts |
| repo | ariadne |
| issue file | workshop/issues/000266-id-lint-ci-publication-target.md |
| boundary | whole-issue close |
| milestone | — |
| window | 70fd1077511c3f526817e9b910384aa44046cbec..f864843970b724a3276d40a0f109fa3de71ccce9 |
| command | sdlc close --issue 266 |
| reviewer | claude |
| timestamp | 2026-09-28T21:54:14-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** The fix is small and addresses the root cause. When CI passes `--remote`, `lint-ids` opens the tracker on the named remote with `tracker.NewRepository`. It no longer resolves `branch.main.*` publication config just to find out where to read. The script passes `--remote origin`, so the trunk and the cards come from the same remote. The operator path, without the flag, is unchanged. All three prior findings are addressed. At head `f8648439` I ran `go test ./cmd/sdlc -run 'TestDuplicateIDCheckRunsInADetachedCICheckout|TestLintIDs'` and it passed. The only thing left is one Minor wording problem in the issue file.

1. **Strengths**
   - `cmd/sdlc/issuelintids.go:244-250`: the change is limited to how the repository is opened. The marker check, the nil-repo guard and `GuardCutoverAt(dirs.Top, headRef)` still run on both paths. The `--remote` path loses nothing, because `RepositoryFor`'s `GuardCutover` is replaced by `GuardCutoverAt` afterwards on either path.
   - `cmd/sdlc/trackedlegacy_test.go:152`: the test runs the real `40-duplicate-issue-id.sh` in a checkout shaped like CI: a fresh `init`, refs fetched from the remote, a detached HEAD and no local `main`. It covers the three required outcomes: carded details exit 0, hand-made details exit 1, and a missing tracker exits 2.
   - The fail-closed case deletes the tracker on origin while the CI clone still has the old fetched ref. So it proves the stale fetched ref cannot stand in for the tracker. That is the kind of silent false pass this check exists to prevent.
   - Help text, the script comment and the atlas all explain *why* the flag exists (read-only lookup versus the publication target a writer needs), and they agree with the code.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - In the `## Revisions` entry, the clause numbers point the wrong way. The last commit reordered `## Done when`: the regression test is now clause 1 and the parley.nvim PR run is clause 2. The Revisions entry still calls the parley run "clause 1" and says "Close attests clause 2", which is backwards. The parenthetical names what is meant, so nothing is ambiguous in practice.
     - This is the only instance of this family in the window.

5. **Test coverage**
   - Both flag modes are covered. `--remote` is covered by the new test through the script. The no-flag operator path is covered by the existing `TestLintIDs*` tests, which still pass.
   - I did not re-run the revert-the-fix mutation, because my tools were read-only. The Log records that the test was red before the fix (exit 2 with the exact CI message) and that a mutation check turned the fail-closed case red.

6. **Architecture**
   - **ARCH-DRY: pass.** `NewRepository` is reused rather than writing a second tracker reader. The trunk and the tracker share one remote.
   - **ARCH-PURE: pass.** The IO choice sits at the constructor seam; `detailsWithoutCards` and `addedDetails` stay pure.
   - **ARCH-PURPOSE: pass.** Every consumer runs the script, and the script now passes `--remote`. The parley.nvim run can only be observed after merge, and the issue is honest about that.

7. **Plan revisions:** correct the clause numbers in the existing Revisions entry (parley.nvim run = clause 2, close attests clause 1).

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      trackedlegacy_test.go deletes origin's issue-tracker and asserts exit 2 plus COULD NOT RUN with the stale fetched ref present; passes at head.
  - id: BR-2
    disposition: addressed
    note: |
      Done-when split into close-time and post-merge clauses, plus a Revisions entry explaining the post-merge evidence.
  - id: BR-3
    disposition: addressed
    note: |
      The issue Log now has a single 2026-09-28 heading.
findings:
  - id: new
    severity: Minor
    family: issue-prose-cross-reference-drift
    title: |
      Revisions entry cites Done-when clause numbers that are inverted after the reorder
    detail: |
      The Revisions entry calls the parley.nvim PR run clause 1 and says close attests clause 2. The final commit reordered Done-when so the regression test is clause 1 and the parley run is clause 2. Swap the numbers. This is the only instance in the window.
```
