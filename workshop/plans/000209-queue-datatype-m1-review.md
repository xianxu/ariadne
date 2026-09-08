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

---

## Re-review — 2026-09-07T17:36:16-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 209 — queue datatype for advisory work ordering |
| repo | ariadne |
| issue file | workshop/issues/000209-queue-datatype.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 00d7c75c5371bbb3e2f01f5e28780565c274f3a9..850eabd1c50b4dbbc07cd2c10ef64c04f464df93 |
| command | sdlc milestone-close --issue 209 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-09-07T17:36:16-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Round 1's five Important findings were genuinely worked, and I confirmed four of them by reverting the fix in a scratch worktree and watching the test go red (BR-1, BR-2, BR-4, and BR-5's new create-path test). The headline redesign — "observe git's state, don't parse its English" — is the right rule, and `Update`'s re-resolve-the-ref retry is strictly better evidence than the prose match it replaced. What blocks the boundary is that the rule was applied to two of the three classifiers and the third got a *different* conflating signal: `exists()` now decides "is this path on the trunk?" from `cat-file -e`'s exit code, which returns non-zero both for "absent from the tree" and for "the object is not available". I reproduced the consequence end-to-end against real git — `Read` returns `("", nil)` and `Update` publishes a commit that replaces `"important\ncontent\n"` with `"- mine\n"` on the trunk and reports success. That is silent destruction of published data from a primitive `ariadne#207` will point at issue files, and the function's own doc comment claims to prevent it. The same instance-not-class pattern shows up three more times in this round (two stale `combined output` comments in code, three in the plan), which is what the ARCH-PURPOSE lens is for. Package tests are green (`go test ./cmd/sdlc/internal/gitx/` ok, 7.7s); the only failure in `./cmd/sdlc/...` is `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`, which I confirmed fails identically at the base SHA.

## 1. Strengths

- **`trunkfile.go:269-286` — retryability is now an observation, not a reading of git's English.** Re-resolving the tracking ref after a failed push and retrying only if it actually moved is the correct shape, and `TestTrunkFile_NonRetryablePushFailsFast` (`trunkfile_test.go:354-383`) pins it: I restored the old `strings.Contains(out, "rejected")` classifier and the test went red with `transform called 3 times, want 1`. Confirmed-good ground.
- **`trunkfile_test.go:253-290` — BR-1's vacuous guard is genuinely repaired.** The failure is now injected at `write-tree`, *inside* `commitAndPush`, so `GIT_INDEX_FILE` is really set, and `t.Fatal` on `seen == ""` makes the guard unable to vacate again. Verified by moving `cleanupIdx()` off the defer onto the success path: `temp index …/index survived a mid-sequence git failure`.
- **`trunkfile.go:360-363` + `trunkfile_test.go:321-350` — `--type=bool` with a `yes`-configured repo.** Reverting to `--get` alone turns the test red. This is the right pin: it exercises git's normalization rather than restating the implementation.
- **`trunkfile.go:207-232` — the temp index moved into a 0700 `MkdirTemp` directory, and `TestTempIndexPath_NoUnclaimedNameWindow` asserts the mode, the absoluteness, that the file does *not* pre-exist, and that cleanup removes the whole directory.** That structurally guarantees the "never `$GIT_DIR/index`" clause rather than testing around it (ARCH-SECURE).
- **`atlas/workflow/sdlc-binary.md:646-687`** explains why `push <commit>:refs/heads/main` *is* the concurrency primitive and names the three load-bearing decisions; it matches the code I read.

## 2. Critical findings

**C-1 — `trunkfile.go:195-205`: `exists()` treats "object unavailable" as "path absent", so `Read` fabricates an empty file and `Update` wipes the trunk while reporting success.**

> **This is the 3rd finding in family `git-output-string-matching`.** Earlier rounds fixed instances (`isNonFastForward`'s prose match, `isMissingPath`'s prose match). Do not fix this instance alone — the rule below is what needs fixing.

`cat-file -e <ref>:<path>` exits non-zero for *both* "not in the tree" and "in the tree but the object cannot be produced". Reproduced with real git 2.50.1, two ways:

```
# normal clone, blob removed from .git/objects
git cat-file -e main:queue.md   -> exit 1     (path IS in the tree)
git ls-tree --name-only main -- queue.md -> exit 0, prints "queue.md"

# partial clone (--filter=blob:none), promisor unreachable
git cat-file -e origin/main:queue.md -> exit 128
git ls-tree --name-only origin/main -- queue.md -> exit 0, prints "queue.md"
```

End-to-end through the real type (scratch test, origin reachable, one loose blob removed):

```
Read -> "", err=<nil>
transform saw old=""
TRUNK NOW = "- mine\n"        # was "important\ncontent\n"
```

`Update` returned `nil`. The doc comment at `trunkfile.go:198-201` states this is exactly what the change was meant to prevent — *"silently reclassifies a real failure as 'absent' and hands the caller an empty file"* — and it is still reachable, because the replacement signal conflates the same two outcomes the prose did.

**The rule the family needs, stated once:** *a classifier in this file may only assert an outcome its signal actually distinguishes; where git offers no such signal, the ambiguous case must surface as an error, never as a value.* The enumeration that rule implies, swept in this round:

| site | decision | signal today | verdict |
|---|---|---|---|
| `exists` (`:202`) | absent vs unavailable | `cat-file -e` exit code | **conflates** → use `ls-tree --name-only <ref> -- <path>` (non-zero = error, zero+empty = absent), or distinguish exit 1 from 128 |
| `refExists` (`:143`) | ref present | `rev-parse --verify --quiet` | ok (called only with the tracking ref in a valid repo) |
| `Update` post-push (`:279-285`) | retryable vs not | re-resolve the ref | ok in principle, scoped too widely — see M-1 |
| `offlineError` (`:150`) | why the fetch failed | none | asserts a cause it never observed — BR-11, still open |
| `signs` (`:361`) | config truthiness | `--type=bool` | ok |

Fix sketch: make `readRef`/`readLocal` return `(nil, nil)` only on a *proven* absence and propagate an error otherwise, then pin it with a test that removes the loose blob and asserts `Read` errors rather than returning empty. A belt-and-braces guard in `Update` (refuse to publish content derived from an empty read when the path is present in the tree) is not a substitute for the classifier fix.

## 3. Important findings

None beyond C-1 and the prior-round dispositions below.

## 4. Minor findings

- `trunkfile.go:37` and `trunkfile.go:302` still say the shim returns "combined output", contradicting `trunkfile.go:39` four lines later and `commitAndPush`'s actual `errOut` return (family `stale-comment`, 2nd — see the rule in the findings block).
- `workshop/plans/000209-queue-datatype-plan.md:86`, `:136`, `:207` still specify `CombinedOutput` after the Revisions entry corrected `:64` (family `plan-traceability`, 2nd).
- `trunkfile_test.go:321-350` and `:430-466` are near-identical 25-line bodies differing only in the config value; one table over `{"true","yes","1","0","false"}` would also close the negative case, which nothing tests today (family `duplicated-helper`, 2nd).
- The contended path now costs 6 fetches + 3 pushes, up from 3+3: the post-push `t.fetch()` at `:279` is discarded and immediately repeated by the next iteration's `Read` (`:255`).
- `signs()` re-shells `git config` on every retry attempt; hoist it out of `commitAndPush`.

## 5. Test coverage notes

Sixteen tests, all against a real bare origin; the two `runGitIn` stubs are narrow (strip `-S`, force one push failure) and sit on top of the real fixture, so ARCH-MOCK holds. `TestFirstLine` is the right response to BR-10 — and eliminating the two prose classifiers outright was a better answer than table-testing them. Remaining gaps, in the order I'd close them: (1) the C-1 case — no test distinguishes "absent" from "unavailable", and the suite stays fully green with `refExists` hardwired to `true`, which is how I confirmed BR-3's byte assertion is still missing; (2) `TestTrunkFile_NoRemoteIsNamed` (`:509`) discards the returned bytes — `testfix.InitialCommit()` writes `README` = `"x\n"`, so `got` is assertable in one line; (3) the blob temp file (`writeTemp`) cleanup is correct but unpinned; (4) `Read`'s strict offline branch is exercised only transitively through `Update`'s refusal.

## 6. Architectural notes for upcoming work

**ARCH-DRY** — pass on the load-bearing call (one CAS/retry loop, `run` left alone, `firstLine` consolidated into `gitx.FirstLine` with a delegating alias so ~12 call sites read unchanged); flagged on the duplicated signing tests. **ARCH-PURE** — pass; `FirstLine` now has a direct table test and the impure classifiers correctly live as methods at the boundary. **ARCH-PURPOSE** — **flagged, and it is this round's theme.** Three of the round-1 fixes resolved the site the finding named while enumerable siblings survived: the `combined output` claim (grep finds 5, 2 were fixed), the plan's `CombinedOutput` restatement (4 sites, 1 fixed), and the classifier sweep (BR-3 asked for all three classifiers pinned; two were rewritten, one of them onto a signal that conflates the same two outcomes, and `offlineError` was untouched). The enumeration for each is a single `grep` — that is what "write the enumeration the class implies" means here. **ARCH-MOCK** — pass, exemplary. **ARCH-CONSTRAINTS** — flagged: still no context/timeout on `fetch`/`push`, and the round-1 fix doubled the fetch count on the contended path; M2 turns this into an interactive keystroke path, so decide there whether `sdlc queue` is where the binary's first `context.WithTimeout` lands. **ARCH-SECURE** — flagged via C-1 (a fabricated value that downstream code publishes as truth); the temp-index race is properly closed, and `path`/`msg` reach git as argv elements with no injection surface. **ARCH-ORDER** — pass, and still the strongest part of the diff: the transform is the seam through which a test chooses the interleaving, and retryability is now read off state rather than prose.

For M2: `Read` returning `(nil, nil)` for an absent path is what C-1 exploits. If the fix distinguishes proven-absence from error, `Doc.Parse` will still receive `nil` on the genuine seed — keep `nil` in the fuzz corpus alongside `[]byte{}`.

## 7. Plan revision recommendations

1. **Purge the surviving `CombinedOutput` restatements** at `:86` (the TrunkFile `Seam:` bullet), `:136` (Task 1 Step 4) and `:207` (Task 4 Step 5). The 2026-09-07 Revisions entry corrected the audit's Resolution paragraph only; as written the plan still instructs a future reader to build the shim that shipped the CRLF-in-blob-hash bug.
2. **Record the retryability redesign's actual scope.** The Revisions entry says `Update` "re-resolves the tracking ref after a failed push"; the code applies that classification to *any* `commitAndPush` failure — `read-tree`, `hash-object`, `update-index`, `write-tree`, `commit-tree` and the temp-file steps included. State it, or narrow the code to the push step.
3. **Add the third classifier decision** once C-1 is fixed: Task 1 Step 5 still describes "a missing path is not an error: return `(nil, nil)`" without saying how absence is *established*, which is the whole of C-1.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Verified by reverting: moving cleanupIdx() off the defer turns the test red. Blob temp-file cleanup remains unpinned (correct code, no test) — noted in coverage, not re-raised.
  - id: BR-2
    disposition: addressed
    note: |
      Verified by restoring the prose match: TestTrunkFile_NonRetryablePushFailsFast goes red with "transform called 3 times, want 1".
  - id: BR-3
    disposition: not-addressed
    note: |
      Prose matching is gone, but the byte assertion the finding asked for is still absent: I hardwired refExists to true and the entire suite stayed green, so the fabricated-empty-value path remains unpinned. TestTrunkFile_NoRemoteIsNamed:509 still discards the bytes; testfix.InitialCommit writes README="x\n", so this is a one-line assert. The replacement signal's own conflation is raised separately as C-1.
  - id: BR-4
    disposition: addressed
    note: |
      Verified by reverting to `config --get`: TestTrunkFile_SigningHonorsNonCanonicalBool goes red.
  - id: BR-5
    disposition: addressed
    note: |
      TestTrunkFile_UpdateCreatesAbsentPath covers the nested absent path and also asserts the sibling file survives read-tree.
  - id: BR-6
    disposition: addressed
    note: |
      tempIndexPath() now takes no parameter.
  - id: BR-7
    disposition: addressed
    note: |
      MkdirTemp 0700 + filepath.Join, pinned by TestTempIndexPath_NoUnclaimedNameWindow including the mode check.
  - id: BR-8
    disposition: addressed
    note: |
      gitx.FirstLine exported with a table test; issueids.go's firstLine is a delegating alias.
  - id: BR-9
    disposition: addressed
    note: |
      trunkfile_test.go:203-205 corrected. Two other sites carrying the same stale claim are raised as a family repeat, not as this finding.
  - id: BR-10
    disposition: addressed
    note: |
      Better than asked: isNonFastForward and isMissingPath were eliminated rather than table-tested, and FirstLine got TestFirstLine.
  - id: BR-11
    disposition: not-addressed
    note: |
      trunkfile.go:150 unchanged. Extending the finding to its class: Update:288 also reports "the trunk moved under 3 attempts" for any commitAndPush failure (read-tree/hash-object/write-tree/commit-tree/temp-file) that coincides with a peer push. Rule: an error message may only name a cause the code observed.
  - id: BR-12
    disposition: not-addressed
    note: |
      trunkfile.go:320-322 unchanged; no sentence saying .gitattributes is resolved from the checked-out tree rather than from the trunk being written.
  - id: BR-13
    disposition: addressed
    note: |
      Revisions section added, Tasks 1-6b ticked, ReadDegraded added to the TrunkFile entry, seam Resolution corrected, commit-granularity divergence recorded. Three unnamed CombinedOutput restatements survive and are raised as a family repeat.
findings:
  - id: new
    severity: Critical
    family: git-output-string-matching
    title: |
      exists() reads "object unavailable" as "path absent", so Update silently wipes the trunk file and reports success
    detail: |
      3rd finding in this family — fix the RULE, not the instance. trunkfile.go:202
      uses `cat-file -e <ref>:<path>`, whose non-zero exit means both "not in the
      tree" and "object cannot be produced". Reproduced on git 2.50.1: with the
      loose blob removed, `cat-file -e main:queue.md` exits 1 while `ls-tree
      --name-only` exits 0 and prints the path; on a --filter=blob:none clone with
      an unreachable promisor it exits 128 with the path still in the tree. Driven
      end-to-end through the real type with origin reachable, Read returned
      ("", nil) and Update published a commit replacing "important\ncontent\n"
      with "- mine\n", returning nil. The doc comment at trunkfile.go:198-201
      claims this exact failure is prevented. Rule for the whole family
      a classifier may only assert an outcome its signal distinguishes; the
      ambiguous case must surface as an error, never as a value. Enumeration
      exists (conflates, fix with ls-tree or an exit-code split), refExists (ok),
      Update's post-push check (ok but scoped to all of commitAndPush),
      offlineError (asserts an unobserved cause, BR-11), signs (ok).
  - id: new
    severity: Minor
    family: stale-comment
    title: |
      Two more "combined output" claims survive in trunkfile.go, one contradicting itself four lines later
    detail: |
      2nd finding in this family — state the rule rather than patching the site.
      Rule when a decision is reversed, sweep every restatement of it; the
      enumeration is a grep for the old term. `grep -n "combined\|Combined"` over
      trunkfile.go returns :37 ("returning combined output", contradicted by :39)
      and :302 ("Returns git's combined output", where commitAndPush actually
      returns stderr only). Round 1 fixed the test-comment site; these two are the
      same claim in the file a reader consults for the rationale.
  - id: new
    severity: Minor
    family: plan-traceability
    title: |
      The plan still instructs CombinedOutput at three sites after the Revisions entry corrected one
    detail: |
      2nd finding in this family — same rule as above, same grep. Plan lines 86
      (TrunkFile's `Seam:` bullet), 136 (Task 1 Step 4) and 207 (Task 4 Step 5)
      still specify CombinedOutput; only line 64's Resolution paragraph was
      corrected. As written the plan tells a future implementor to build the shim
      whose CRLF warning landed inside a parsed blob hash.
  - id: new
    severity: Minor
    family: duplicated-helper
    title: |
      TestTrunkFile_SigningHonorsNonCanonicalBool and TestTrunkFile_SigningRepoGetsSignedCommit are near-identical 25-line bodies
    detail: |
      2nd finding in this family — the rule is one definition per behavior,
      including in tests. trunkfile_test.go:321-350 and :430-466 differ only in the
      config value; both stub runGitIn, strip -S, and assert sawDashS. One table
      over {"true","yes","1","0","false"} collapses them and would also cover the
      negative case, which nothing tests today.
  - id: new
    severity: Minor
    family: repeated-external-call
    title: |
      The contended Update path now costs 6 fetches instead of 3
    detail: |
      trunkfile.go:279 fetches to decide retryability, discards the result, and the
      next iteration's Read (:255) immediately fetches again. Reuse the post-push
      fetch for the retry's base read (ARCH-CONSTRAINTS repeated expensive work).
      Also hoist signs() out of commitAndPush so it is not re-shelled per attempt.
```
