# Boundary Review — ariadne#209 (milestone M1)

| field | value |
|-------|-------|
| issue | 209 — queue datatype for advisory work ordering |
| repo | ariadne |
| issue file | workshop/issues/000209-queue-datatype.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 00d7c75c5371bbb3e2f01f5e28780565c274f3a9..2c756f91a15f81bf27770b6f8e65e88f4caaa060 |
| command | sdlc milestone-close --issue 209 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-09-07T17:17:20-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers `gitx.TrunkFile` as specified: it reads and CAS-writes one path on `origin/main` with no working tree, the retry re-runs the transform on the moved base (the load-bearing design claim), gitattributes and signing round-trip, offline is asymmetric (read degrades, write refuses), and the temp index is absolute and deferred-cleaned. I verified the whole package green (`go test ./cmd/sdlc/internal/gitx/` — 12/12 pass, 9.9s); the only failure in `./cmd/sdlc/...` is `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`, which the plan's Verification section already documents as pre-existing (ariadne#210). Atlas is updated with a substantive, accurate section. Nothing here is Critical — no correctness bug on the happy path, and I confirmed the first-ever-write plumbing sequence works by replaying it by hand against a throwaway bare origin. What blocks a clean SHIP is a cluster of cheap gaps: one test whose entire assertion body is unreachable (so the Done-when clause it claims to pin is unpinned), three classifiers that branch on git's human-facing stderr with no test producing that text — one of which I reproduced misclassifying a hook-declined push as a CAS collision — and a `commit.gpgsign` check that only recognizes the literal `true`.

## 1. Strengths

- **`trunkfile.go:227-251` + `trunkfile_test.go:166-199` — the retry test is a genuine two-publisher race made deterministic by injecting the peer push *inside the transform*.** This is exactly what ARCH-ORDER asks for at review: a seam that lets a test choose the interleaving, rather than a green run that samples one arbitrary ordering. Asserting `calls == 2` pins re-running the transform, not just "both lines present" — a re-push of the original bytes would fail the count. Confirmed-good ground.
- **`trunkfile.go:34-52` — separating stdout from stderr instead of `CombinedOutput`, with the reason recorded.** I verified the rationale is real, not folklore: `git hash-object -w --path x.md <file>` under `*.md text eol=lf` emits `warning: ... CRLF will be replaced by LF` on stderr, which combined output would have concatenated into the parsed blob hash and, on reads, into file content. This deviates from what the plan's seam table specified, and the deviation is correct. It is pinned by `trunkfile_test.go:130-132` (stdout must stay empty) and by `TestTrunkFile_RoundTripsUnderGitattributes`.
- **`trunkfile.go:78-87` + `trunkfile_test.go:45-49` — the empty-dir guard written before the feature.** The dangerous failure for this type is silently fetching from and pushing to the *real* ariadne origin during `go test`; making it unrepresentable beats a convention. This is the correct disposition of gate finding PQ-7 and it is pinned by a test that fails without it.
- **`trunkfile_test.go:16-35` — real bare origin as the fake (ARCH-MOCK).** A function-call mock cannot produce a genuine non-fast-forward rejection; the fixture can, and `TestTrunkFile_RetryExhaustionSurfacesGitRejection` asserts on git's own text rather than a wrapper message.
- **`atlas/workflow/sdlc-binary.md:646-687` — the atlas section explains *why* the push is the concurrency primitive** and names the three load-bearing decisions, rather than restating the API. It matches the code I read, including `push <commit>:refs/heads/main`.

## 2. Critical findings

None.

## 3. Important findings

**I-1 — `trunkfile_test.go:258-265`: the temp-index assertions are unreachable, so the Done-when clause they claim to pin is unpinned.**
`Update` (`trunkfile.go:229-238`) calls the transform *before* `commitAndPush`, and `GIT_INDEX_FILE` is only set at `trunkfile.go:271` inside `commitAndPush`. A transform that errors immediately means the spy at line 241-248 never observes that env var, so `seen == ""` and the whole `if seen != ""` block — the `os.Stat` removal check *and* the `filepath.IsAbs` check — is skipped every run. The test's name and comment ("the temp index is gone. Cleanup on the ERROR path is the case a success-only defer misses") describe something it does not assert. The Done-when clause *"The temp index is removed on every exit path, including the error paths, and is never `$GIT_DIR/index`"* has no coverage at all, and neither does the blob temp file. Plan Task 6 Step 1 also asked for a `$GIT_DIR/index` mtime comparison, which is absent.
Fix sketch: assert on a path that actually reaches `commitAndPush` — e.g. capture `seen` from a successful `Update` and check the file is gone afterwards, plus a rejected-push case (reuse the exhaustion fixture) where `commitAndPush` runs three times; and add `t.Fatal` if `seen == ""` so the guard can never silently vacate again. Snapshot `filepath.Join(repo, ".git", "index")` mtime+size around `Update` for the `$GIT_DIR/index` half. The production code appears correct (`defer cleanupIdx()` at line 267 precedes every return) — this is a coverage defect, not a behavior defect.

**I-2 — `trunkfile.go:256-261`: `isNonFastForward` matches bare `"rejected"`, so non-retryable push refusals are retried three times and then reported as "the trunk moved".** Reproduced: a `pre-receive` hook that exits non-zero produces `! [remote rejected] main -> main (pre-receive hook declined)`. That substring-matches `"rejected"`, so `Update` treats a policy/permission/protected-branch refusal as a CAS collision, burns two more fetch+build+push cycles, and returns `publish %s: the trunk moved under 3 attempts` — a next-action spec pointing at the wrong action, which is precisely what `sdlc --help`'s error contract forbids. `"non-fast-forward"` and `"fetch first"` already cover the real CAS case (the exhaustion test asserts only those two).
Fix sketch: drop the `"rejected"` clause, or narrow it to `"[rejected]"` co-occurring with `"fetch first"`/`"non-fast-forward"`. Add a table test over `isNonFastForward` with the three real outputs (non-ff, hook-declined, permission-denied) as fixtures.

**I-3 — `trunkfile.go:181-186` and `104-136`: the missing-path/no-remote classification branches on git text that current git does not emit, and nothing pins it (same family as I-2).** `isMissingPath` looks for `"Not a valid object name"`; git 2.50.1 says `fatal: invalid object name 'refs/remotes/origin/main'.` (lowercase, different wording). `ReadDegraded`'s local-branch fallback at line 128-134 fires *only because* that phrase fails to match — i.e. the documented behavior "a repo with no `origin` reads the local ref" is load-bearing on a **negative** match against an unpinned string. If a git version emits the matched phrasing, `readRef` returns `(nil, nil)`, the fallback is skipped, and `ReadDegraded` returns **empty content** alongside a warning claiming it read the stale ref — a fabricated value presented as evidence (ARCH-SECURE's at-review lens: "degrade visibly rather than substituting a fabricated value"). `TestTrunkFile_NoRemoteIsNamed` (`trunkfile_test.go:397-410`) asserts only that the warning contains `"origin"`; it never asserts the returned bytes, so it passes either way.
Fix sketch: sweep the class rather than the instance — enumerate every site in this file that classifies git by its output (`isMissingPath`, `isNonFastForward`, and `offlineError`'s implicit "any fetch failure means offline"), and give each a fixture test that runs the real git command producing the real text. For `isMissingPath` specifically, prefer classifying by *ref existence* (`git rev-parse --verify --quiet <ref>`) rather than by prose, and make `TestTrunkFile_NoRemoteIsNamed` assert the README bytes came back.

**I-4 — `trunkfile.go:318-321`: `signs()` compares the raw config string to `"true"`, so a repo configured `commit.gpgsign = 1|yes|on` gets unsigned commits.** Verified: `git config --get commit.gpgsign` returns `1` / `yes` verbatim, while `git config --get --type=bool commit.gpgsign` returns `true` for all of them. Git treats all four as true. The function's own comment says "A signing repo that silently produced unsigned commits through this path would be a regression" — that regression is reachable today. `TestTrunkFile_SigningRepoGetsSignedCommit` sets the literal `"true"`, so it cannot catch this. This also compounds with I-2: on a trunk with signature-requiring branch protection, the unsigned push is refused with `[remote rejected]` and then misreported as "the trunk moved".
Fix sketch: `runGitIn(t.dir, nil, "config", "--get", "--type=bool", "commit.gpgsign")`, and add `"1"` to the test's table.

**I-5 — no test for `Update` on a path absent from the trunk, which is the exact path M2 Task 13's seed takes.** `workshop/queue.md` does not exist on `origin/main`, so the first real use of this primitive is a create, not an append. Every `Update` test seeds `queue.md` via `trunkFixture` first. I replayed the plumbing by hand (`read-tree` → `hash-object -w --path workshop/queue.md` → `update-index --add --cacheinfo` → `write-tree` → `commit-tree` → `push`) against a throwaway bare origin and it lands correctly, so this is a coverage gap rather than a bug — but it is the one path with no regression guard, and it is the next path the code will run in anger.
Fix sketch: one test — `trunkFixture` then `Update("workshop/queue.md", ...)` on a nested path that is not in the trunk tree; assert `git show main:workshop/queue.md` in the bare origin.

## 4. Minor findings

- `trunkfile.go:194` — `tempIndexPath(dir string)` never uses `dir`; the parameter is vestigial and its doc comment implies otherwise. Drop it or use it.
- `trunkfile.go:195-201`, `324-325` — `os.CreateTemp("")` then `os.Remove` leaves a create-then-race window on the index path; on a shared `/tmp` an attacker can plant a symlink git then writes the index through. `os.MkdirTemp` + `filepath.Join(dir, "index")` closes it (ARCH-SECURE, low exposure on a single-user dev box).
- `trunkfile.go:146` — `firstLineOf` duplicates `firstLine` (`cmd/sdlc/issueids.go:112`); `main` already imports `gitx`, so exporting one and deleting the other is free (ARCH-DRY).
- `trunkfile_test.go:202-203` — comment says the rejection text is "reachable only because runGitIn uses CombinedOutput"; `runGitIn` deliberately does *not* use CombinedOutput. Stale claim in the one place a reader checks the rationale.
- `trunkfile.go:256`, `181`, `146` — these are pure functions inside an IO shell but have no direct table tests; they are exercised only through 2.5s integration tests. Pulling them into a pure table test is what would have surfaced I-2 and I-3 (ARCH-PURE).
- `trunkfile.go:104-109` — `Read`'s `offlineError` labels *every* fetch failure "unreachable (offline?)", including "branch does not exist on origin". Misattributes a configuration error as a network one.
- `trunkfile.go:281-286` — `hash-object --path` resolves `.gitattributes` from the *working tree*, i.e. the currently checked-out branch, not from the trunk being written. Correct in practice (attributes rarely differ per branch) but worth one sentence in the doc comment, since "no working tree involved" is the type's headline claim.
- Plan Tasks 1–6b are all still `- [ ]` in `workshop/plans/000209-queue-datatype-plan.md` despite being delivered; and commit granularity diverges (Tasks 2+4 landed as `b93375c`, Tasks 3+5 as `26c672b`). Traceability only.

## 5. Test coverage notes

Twelve tests, all passing, all driving real git against a real bare origin — the right shape for this type. `TestTrunkFile_ReadsFromTrunkNotWorktree` deliberately skipping `testfix.Chdir()` is the correct instinct: a cwd-dependent implementation fails there instead of passing in CI. The race test and the exhaustion test both assert transform call counts, which is what makes them non-tautological.

Gaps, in the order I would close them: the vacuous temp-index block (I-1); the create-a-new-path case (I-5); pure table tests for the three git-text classifiers (I-2, I-3); a `commit.gpgsign = 1` row (I-4). One more not yet raised — `Read` (the strict variant) has no direct offline test; it is covered only transitively through `Update`'s refusal. Cheap to add while touching `TestTrunkFile_OfflineDegradesReadRefusesWrite`.

## 6. Architectural notes for upcoming work

Marker-by-marker: **ARCH-DRY** — pass on the big call (one CAS/retry loop for both #209 and #207, `run` correctly left alone rather than widened); flagged only on the duplicated `firstLine` helper and on the divergent degraded-read return shape (`ReadDegraded` returns `(bytes, warn string, err)` where the cited precedent `publishedIDSpace` returns `(space, staleErr error, err error)` — the *policy* matches, which is what mattered, but #207 and the M2 verb will each have to learn a second idiom; worth settling before a second consumer arrives). **ARCH-PURE** — pass; the pure classifiers should get direct tests. **ARCH-PURPOSE** — pass: M1's purpose was the primitive, the deferrals (`syncInPlace`, `syncViaMainWorktree`) are declared in the Spec as #207's job, and no competing trunk-write path was introduced. **ARCH-MOCK** — pass, and exemplary. **ARCH-CONSTRAINTS** — flagged lightly: no context/timeout on `fetch`/`push`, which the plan names as accepted residue matching the binary's existing convention. That was fine for a library; at M2 the verb makes it an interactive path, so a hung fetch blocks `sdlc queue` unboundedly. Worst case is now 3 fetches + 3 pushes per `Update`, each unbounded. Decide at M2 whether the queue verb is where the binary's first `context.WithTimeout` lands. **ARCH-SECURE** — flagged via I-3 (fabricated empty value on a classification miss) and the temp-file race; `msg` and `path` reach git as argv elements, so no injection surface. **ARCH-ORDER** — pass, and the strongest part of the diff.

For M2: `TrunkFile.Read` returns `(nil, nil)` for an absent path, so `Intent.Apply` will receive `nil` on the seed. Make sure `Doc.Parse(nil)` is in the fuzz seed corpus — an empty-vs-nil distinction is exactly what a round-trip invariant trips on. And `Update` surfaces the transform's error unwrapped (`trunkfile.go:236-238`), which `TestTrunkFile_TransformErrorLeavesNoIndexAndNoCommit` pins with `errors.Is` — good, and it is what lets Task 11 render `ErrAnchorMissing` alongside the trunk state.

## 7. Plan revision recommendations

The plan file has no `## Revisions` section at all, though it was edited mid-stream (Tasks 6b and 13b were appended in this window). Per AGENTS.md §1, append one rather than continuing to edit in place. Entries needed:

1. **`runGitIn`'s signature changed from the seam audit's specification.** The Core-concepts seam table still says *"add `gitx.runGitIn(dir string, env []string, args ...string) ([]byte, error)` — `CombinedOutput` with `cmd.Dir` and `cmd.Env` set"*, and the `stderr` row says the same. The code returns `(stdout, stderr []byte, err error)` deliberately, because combined output folds git's CRLF warning into the parsed blob hash. Record the delta and the reason; as written the plan documents a design that would have shipped a bug.
2. **`TrunkFile`'s surface grew a third method.** The Core-concepts entry lists `Read(path)` and `Update(path, msg, transform)`; the code also exports `ReadDegraded(path) ([]byte, string, error)`, which is the method M2's list path will actually call. Add it to the table with its return contract.
3. **Tasks 6b and 13b were added after plan approval** (from the estimate-quality round). They are in the file but with no revision note explaining when or why.
4. **Tick Tasks 1–6b's delivered steps**, and note that Tasks 2/4 and 3/5 landed as merged commits with different subjects than the plan's Step-5 commit messages.

```findings
findings:
  - id: new
    severity: Important
    family: vacuous-test-guard
    title: |
      Temp-index assertions in TestTrunkFile_TransformErrorLeavesNoIndexAndNoCommit are unreachable
    detail: |
      trunkfile_test.go:258-265 nests the os.Stat removal check and the
      filepath.IsAbs check under `if seen != ""`, but Update calls the transform
      before commitAndPush, and GIT_INDEX_FILE is only set at trunkfile.go:271
      inside commitAndPush — so the spy never sees it and the block never runs.
      The Done-when clause "the temp index is removed on every exit path,
      including the error paths, and is never $GIT_DIR/index" has no coverage;
      neither does the blob temp file. Assert on a path that reaches
      commitAndPush and t.Fatal when seen is empty so it cannot vacate silently.
  - id: new
    severity: Important
    family: git-output-string-matching
    title: |
      isNonFastForward matches bare "rejected", retrying non-retryable push refusals
    detail: |
      trunkfile.go:260. Reproduced: a pre-receive hook declining a push emits
      "! [remote rejected] main -> main (pre-receive hook declined)", which
      substring-matches "rejected". A policy, permission or protected-branch
      refusal is then retried 3 times and reported as "the trunk moved under 3
      attempts" — the wrong next action. "non-fast-forward" and "fetch first"
      already cover the real CAS case. Add a pure table test over the three real
      git outputs.
  - id: new
    severity: Important
    family: git-output-string-matching
    title: |
      isMissingPath carries a phrase git no longer emits, and the no-remote fallback silently depends on it
    detail: |
      trunkfile.go:185 matches "Not a valid object name"; git 2.50.1 says
      "invalid object name". ReadDegraded's local-branch fallback
      (trunkfile.go:128-134) fires only because that match FAILS, so the
      documented "a repo with no origin reads the local ref" behavior rests on a
      negative match against unpinned prose. If a git version emits the matched
      wording, ReadDegraded returns empty bytes with a warning claiming it read
      the stale ref — a fabricated value (ARCH-SECURE). TestTrunkFile_NoRemoteIsNamed
      asserts only the warning, never the bytes. Sweep the class: pin all three
      classifiers (isMissingPath, isNonFastForward, offlineError) against real git
      output, and prefer rev-parse --verify over prose for ref existence.
  - id: new
    severity: Important
    family: git-config-bool
    title: |
      signs() compares the raw config string to "true", so commit.gpgsign = 1|yes|on yields unsigned commits
    detail: |
      trunkfile.go:319-320. Verified: `git config --get commit.gpgsign` returns
      "1"/"yes" verbatim while `--type=bool` normalizes all of git's truthy forms
      to "true". The function's own comment calls silently-unsigned commits a
      regression, and it is reachable today; the test only exercises the literal
      "true". Use --type=bool and add a "1" row. Compounds with the isNonFastForward
      finding: on a signature-requiring trunk the refused push is misreported as a
      CAS collision.
  - id: new
    severity: Important
    family: untested-first-write
    title: |
      No test for Update on a path absent from the trunk — the exact path M2's seed takes
    detail: |
      Every Update test seeds queue.md via trunkFixture first, but
      workshop/queue.md does not exist on origin/main, so Task 13's seed is a
      create. I replayed the plumbing by hand against a throwaway bare origin and
      it works, so this is a coverage gap rather than a bug — but it is the only
      path with no regression guard, and it is the next one to run in anger. One
      test on a nested, absent path closes it.
  - id: new
    severity: Minor
    family: dead-parameter
    title: |
      tempIndexPath(dir string) never uses dir
    detail: |
      trunkfile.go:194 — the parameter is vestigial and the doc comment implies it
      matters. Drop it or use it.
  - id: new
    severity: Minor
    family: temp-file-race
    title: |
      os.CreateTemp then os.Remove leaves a symlink-plant window on the index path
    detail: |
      trunkfile.go:195-201 and 324-325. os.MkdirTemp plus filepath.Join(dir,
      "index") removes the window (ARCH-SECURE; low exposure on a single-user
      machine, but free to fix).
  - id: new
    severity: Minor
    family: duplicated-helper
    title: |
      firstLineOf duplicates firstLine in cmd/sdlc/issueids.go:112
    detail: |
      trunkfile.go:146. main already imports gitx, so exporting one and deleting
      the other is free (ARCH-DRY).
  - id: new
    severity: Minor
    family: stale-comment
    title: |
      Test comment claims runGitIn uses CombinedOutput, which it deliberately does not
    detail: |
      trunkfile_test.go:202-203. The separation is the correct design and is
      explained at trunkfile.go:34-52; the comment contradicts it in the one place
      a reader checks the rationale.
  - id: new
    severity: Minor
    family: pure-logic-untested
    title: |
      isNonFastForward, isMissingPath and firstLineOf are pure but only exercised through integration tests
    detail: |
      Direct table tests over these three would have surfaced both
      git-output-string-matching findings, and run in microseconds instead of
      seconds (ARCH-PURE).
  - id: new
    severity: Minor
    family: error-misattribution
    title: |
      offlineError labels every fetch failure "unreachable (offline?)"
    detail: |
      trunkfile.go:104-109 and 140-142. A branch that does not exist on origin, or
      an auth failure, reads to the operator as a network outage.
  - id: new
    severity: Minor
    family: gitattributes-source
    title: |
      hash-object --path resolves .gitattributes from the working tree, not from the trunk being written
    detail: |
      trunkfile.go:281-286. Correct in practice, but "no working tree involved" is
      the type's headline claim and this is the one place it is not strictly true.
      One sentence in the doc comment.
  - id: new
    severity: Minor
    family: plan-traceability
    title: |
      Plan Tasks 1-6b left unchecked and the plan has no Revisions section despite mid-stream edits
    detail: |
      workshop/plans/000209-queue-datatype-plan.md — Tasks 6b and 13b were appended
      in this window with no revision note, the seam table still specifies a
      CombinedOutput two-value runGitIn that the code deliberately does not
      implement, and TrunkFile's entry omits the exported ReadDegraded. Commit
      granularity also diverges (Tasks 2+4 merged, 3+5 merged).
```
