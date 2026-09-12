# Boundary Review — ariadne#188 (whole-issue close)

| field | value |
|-------|-------|
| issue | 188 — allocate issue IDs from origin/main with retry-on-reject |
| repo | ariadne |
| issue file | workshop/issues/000188-issue-id-allocation-from-main.md |
| boundary | whole-issue close |
| milestone | — |
| window | ff1b52b6311ce92bd014c626f1b089fc01aa7860..b7f08ecdf8bec790fbeaa8711b13e860772df1ac |
| command | sdlc close --issue 188 |
| reviewer | claude |
| timestamp | 2026-09-09T21:17:06-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The structural move is right and well-argued: `syncViaMainWorktree` and its three shared-checkout guards are genuinely gone, `gitx.TrunkFile` is consumed rather than reimplemented, the collision decision is a pure three-outcome function that sits *inside* `prepare` (proven by a real-git CAS test), and the load-bearing mid-retry test seeds the peer inside the window and asserts `prepare` ran twice — the exact assertion an earlier draft got backwards. What blocks SHIP is that the two riskiest contracts the Spec wrote out by hand are not delivered by the code: (a) `sdlc claim` and `sdlc issue new` on a feature branch now push a trunk commit with an **empty subject**, because `syncViaTrunk` passes `msg` straight through and dropped the deleted arm's `syncMessage()` default; (b) the re-allocate ordering — "write the new local path before the push, remove the old one only after the push succeeds" — is inverted by `rc.finish()` running unconditionally, and the Done-when test for it was never written (the `fakePublisher.err` seam that exists to inject it is at zero call sites). A third: the free-id search consults only the trunk, so the collision resolver can mint a fresh duplicate. All four findings below were reproduced with scratch tests against a copy of the tree (repo left untouched).

## 1. Strengths

- **`TrunkWrite.Delete` as a separate field** (`cmd/sdlc/internal/gitx/updatemany.go:22`) — the "absent key cannot mean removal" argument is correct and `TestUpdateMany_WholeSetEarlyReturn` pins the hard half (a *pending* delete with matching writes must still commit). The test comment even records that an earlier version passed on the defect. This is the right shape.
- **`TestUpdateMany_PathsAreDerivedPerAttempt`** (`updatemany_test.go:120`) is real: bare origin, a genuine peer push between attempts, a genuine non-fast-forward rejection, and it asserts the peer's file was not clobbered. This is the test the whole design rests on and it is not a mock reasserting the implementation.
- **`decideCollision`** (`cmd/sdlc/issuecollision.go:47`) is pure, table-tested with no IO, and separates `publish` / `reallocate` / `refuse` — including the `mineOnTrunk && foreign` cell that a two-outcome design would have collapsed. `TestSyncViaTrunk_RefusesWhenTrunkHasForeignSlugAtOurID` closes the wiring gap the author found by mutation.
- **`rewriteIdentity`'s anchored regex** (`cmd/sdlc/reallocate.go:37`) plus the refusal on a file with no `id:` line — and the test comment noting the *first* version of that test used a non-numeric prose value and so pinned nothing. That is the mutation discipline the round asks for, applied honestly.
- **`refIDSpace` and `issueIDFromPath` reused rather than re-derived** — the fifth trunk-id-space reader PQ-3 warned about was not written.

## 2. Critical findings

**C1 — `sdlc claim` / `issue new` push trunk commits with an empty subject** (`cmd/sdlc/synctrunk.go:96` → `claim.go:159`)

`syncViaTrunk` hands `msg` straight to `pub.UpdateMany(msg, …)` → `commit-tree -m ""`. The deleted arm defaulted it: `syncMessage(msg, fmt.Sprintf("issue-sync: update issues from branch '%s'", branch))`. Both `runClaim` (`claim.go:114`) and `runIssueNew` (`issue.go:339`) pass `""` — and `issue.go:337`'s comment still claims `""` "keeps issue new's historical subject", which is now false on this arm. Reproduced: publishing from a feature branch lands a commit on `origin/main` with `%s` and `%B` both empty. This breaks AGENTS.md §12 and the `git log --grep "^#N"` traceability agents depend on, on the tracker's own trunk.

Fix sketch: in `syncViaTrunk`, `msg = syncMessage(msg, fmt.Sprintf("issue-sync: update issues from branch '%s'", branch))` — thread `branch` in (it is already computed at `claim.go:141`, and doing so also retires the `_ = branch` line). Add an assertion on the pushed subject to `TestSyncIssuesToMain_PublishesFromBranchWithNoMainWorktree` for the `""` case, not just the caller-supplied case.

**C2 — a failed publish deletes the ORIGINAL local issue file** (`cmd/sdlc/synctrunk.go:49-56`, `cmd/sdlc/reallocate.go:73`)

```go
rc, err := syncViaTrunkWithRealloc(...)
if rc != nil { rc.finish() }   // runs on error too
return err
```
`finish()` unconditionally `os.Remove(rc.OldPath)`. The Spec is explicit — *"write the new local path BEFORE the push, remove the old one only after the push succeeds"* — and the Done-when asks for a test injecting a failure between the two. Reproduced: with the push rejected, `000700-mine.md` is gone and only the unpublished `000701-mine.md` remains, while the trunk holds nothing. On retry exhaustion the same happens with two orphans removed and the issue silently renumbered to an id that was never published.

Fix sketch: gate cleanup on success — `if err == nil && rc != nil { rc.finish() }`; on failure, remove only the *candidate* paths in `rc.written` and leave `OldPath` alone. Pin it with the test C2 describes, using the existing-but-unused `fakePublisher.err` field.

**C3 — the free-id search is against the trunk only, so re-allocation can mint the duplicate it is preventing** (`cmd/sdlc/issuecollision.go:80` `nextFreeID`, called at `synctrunk.go:129`)

The rule #213 established is that an id must be free across the **union** of trunk and local (`allocateIssueID`, `issueids.go:33`: *"Union, not replacement: an unpushed issue on this branch is real"*). `nextFreeID(id+1, space)` sees only `space`, the trunk map. Two instances of the same class, both reproduced:

- *Same prepare pass.* Two local files at one id both re-allocate; `space` is not updated between them, so both pick `000701` and **both land in ONE commit** — `Write` contains `000701-graft.md` and `000701-mine.md`. The collision resolver publishes a fresh duplicate id, which is the exact failure #188/#207 exist to prevent. `rc` is also mispaired afterwards (`OldPath` from one file, `NewPath` from the other), so `finish()` then removes `000701-graft.md` as an "orphan" *after publishing it*.
- *Local unpushed ids.* Trunk holds `000700-theirs`, local holds an unpushed `000701-local-unpushed.md`; re-allocation picks `000701` and collides locally. Reachable in this fleet's normal multi-session pattern (the issue Log records nine issues filed in one session).

Fix sketch: give `nextFreeID` the full space — seed a working copy of `space` from `refIDSpace(v.Ref())` ∪ `issue.LocalPathsByID(dirs.Abs)` at the top of `prepare`, and record each id it hands out back into that map before the next file is considered. Tests: two same-id local files must land under two distinct ids; a local unpushed id must be skipped.

## 3. Important findings

**I1 — stale `rc` across attempts deletes a just-published file** (`cmd/sdlc/synctrunk.go:120-146`) — ARCH-ORDER. `rc` is set on attempt N and never invalidated if attempt N+1 decides `verdictPublish` for the original path (reachable: a peer withdraws its colliding file, exactly how this repo's own `000207` duplicate was resolved). Reproduced: the final set publishes `000700-mine.md`, then `finish()` removes it locally and keeps an unpublished `000701-mine.md`. Fix: reset `rc = nil` at the top of each `prepare` invocation, and make the reallocation set a value the *last* attempt produces rather than an accumulator.

**I2 — no success confirmation at all on the publish arm** (`cmd/sdlc/synctrunk.go`). The deleted arm printed `cok(stderr, "Issues synced to main and pushed to origin.")` and `fmt.Fprintln(stdout, "synced")`. `syncViaTrunk` prints `"Publishing issue changes to the trunk:"` plus the file list and then nothing — success and a no-op look identical, and the `synced` marker `issue.go:335` still calls a contract is gone. Fix: emit both on the success return (and on the whole-set no-op, distinguished).

**I3 — `sdlc claim --help` documents the deleted route** (`cmd/sdlc/helptext/claim.md:32-53,68`). Still lists "find the main worktree", "pull --rebase on the main worktree", "detect conflicts … print resolution steps", an entire `CONFLICT BEHAVIOR` section for machinery that no longer exists, exit-code-1 causes "missing main worktree, dirty main, conflicts", and the `"issue-sync: update issues from branch '<branch>'"` default that C1 dropped. This is user-facing surface a reader types — the README-gate class.

**I4 — atlas not updated in this range** (Docs gate, AGENTS.md §8). At `b7f08ec`, `atlas/workflow/sdlc-binary.md:89,117,149,660` still documents `syncViaMainWorktree`; the `syncViaTrunk` rewrite exists only as an **uncommitted** working-tree edit, outside the window. `atlas/workflow/issue-sync.md:22-42` still describes the worktree route, the conflict detector and "on a feature branch the file routes to the main worktree" — and is not touched even in the working tree.

**I5 — Done-when bullets not delivered** (test coverage): (a) *"publish … while another worktree on main is dirty or mid-rebase"* has no test — `TestSyncIssuesToMain_PublishesFromBranchWithNoMainWorktree` covers only the absent-worktree half; (b) the re-allocate-ordering test (C2) is absent, and `fakePublisher.err` is a failure seam at **zero call sites**, kept compilable by a bare `var _ = errors.Is` (`synctrunk_test.go:157`); (c) *"the published blob round-trips byte-identical, with attributes applied"* is asserted only as `strings.Contains`.

## 4. Minor findings

- `cmd/sdlc/claim.go:16-22` — bullet 2 renamed to `syncViaTrunk` but its six sub-bullets still describe the deleted worktree route step by step; actively misleading now.
- `cmd/sdlc/claim.go:159` — `_ = branch` is dead; `branch` is already used at line 142.
- `cmd/sdlc/internal/gitx/updatemany.go:42,45,48` — `ViewOf` justified by consumer tests (fair), but `TrunkView.Read` and `TrunkFile.TrackingRef` are exported at zero call sites anywhere.
- `cmd/sdlc/internal/gitx/updatemany_test.go:178` — `commitCount` computes `len(strings.Fields(...))*0 + atoi(...)`, a dead term that costs an extra `git rev-list` per call.
- `cmd/sdlc/synctrunk_test.go:113` — `TestSyncViaTrunk_CheckRunsOnEveryAttempt` asserts `pub.runs == 2`, a counter the fake increments unconditionally; it cannot fail unless `syncViaTrunk` errors. `TestUpdateMany_PathsAreDerivedPerAttempt` is the real one.
- `go test ./cmd/sdlc/` is **red** on `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (`workshop/plans/000200-…-plan.md` archived at `dfeba9c`, which precedes the base) — pre-existing, not introduced here, but the package cannot be cited as green close evidence without saying so.

## 5. Test coverage notes

Real-git coverage is genuinely good where it exists (bare origin, actual CAS rejection, commit-count assertions). The gap is uniform and diagnostic: **every failure path is untested.** No test drives `UpdateMany` to a non-nil error through `syncViaTrunk`; no test drives retry exhaustion; no test injects a failure between the local write and the push. That is precisely why C2 and I1 shipped — the `finish()` call sits on a path nothing exercises. Adding the C2 test (using the already-present `err` field) is cheap and would go red against all three of C2, I1, and the exhaustion case.

## 6. Architectural notes

- **ARCH-DRY — flag (C3).** `refIDSpace` and `issueIDFromPath` reuse is exemplary. But `nextFreeID` is a *second allocator* whose rule diverges from `allocateIssueID`'s union. One reader, two allocation policies.
- **ARCH-PURE — pass, one note.** `decideCollision`, `nextFreeID`, `rewriteIdentity` are pure and IO-free in test. The note: the `prepare` closure (`synctrunk.go:118-148`) folds the decision, an `os.WriteFile`, and mutation of `rc` into one body — every one of C2/I1/C3 lives inside it. Extracting a pure `plan(changed, space, firstPub) → (writes, deletes, realloc, err)` would make them unit-testable without git.
- **ARCH-PURPOSE — flag (C3).** Shadow-sweep of "an id must be unique across the published space": `allocateIssueID` derives from the union; `nextFreeID` restates a trunk-only version. The class is "free-id search over the complete space", and it was swept in one of its two members.
- **ARCH-MOCK — pass, one note.** gitx tests run the real stack against a real bare origin; `fakePublisher` is stateful enough to model the retry. Note: its `err` seam is at zero call sites, so the fake cannot currently model the one interaction that matters most.
- **ARCH-CONSTRAINTS — pass.** Bounded at 3 attempts, one fetch per attempt, `ls-tree` reads names only. `nextFreeID`'s unbounded `for` is bounded in practice by the map.
- **ARCH-SECURE — pass.** `--end-of-options` on `refIDSpace`, structurally separated argv, anchored frontmatter regex, no credentials in scope. Paths come from `git diff --name-only` (repo-relative, no traversal), and `rewriteIdentity` validates the basename shape.
- **ARCH-ORDER — flag (I1, C2).** `rc` is nullable state carried between retry attempts with no written legal combinations; `(attempt, verdict) → (rc, effects)` is not enumerable from the code, and the two cells that matter — "re-allocated then published at the old path" and "re-allocated then push failed" — are both wrong. The interrupting events PQ-4 asked to be named (crash between rename and push, a second publisher taking the new id) are named in the Spec and not enforced in the code.

## 7. Plan revision recommendations

`workshop/issues/000207-sync-without-worktree.md` needs a `## Revisions` entry stating:

1. **The re-allocate ordering (Spec step 3) is not what the code does.** Either the code changes to match ("old path removed only after the push succeeds") or the Spec is amended to say cleanup runs unconditionally — but the current text asserts a guarantee the tree does not hold.
2. **Step 1's "pick the next free id from `refIDSpace`" is under-specified.** It must say: over the union of trunk, local unpushed files, and ids already handed out earlier in the same `prepare` pass. As written it licenses exactly the duplicate C3 reproduces.
3. **Commit-subject default.** The Spec's "Scope: replace `syncViaMainWorktree` only" implies message parity; record that the branch-arm default (`issue-sync: update issues from branch '<b>'`) must survive the replacement.
4. **Done-when status.** Mark "publish while another worktree on main is dirty or mid-rebase", "the re-allocate ordering is asserted", and "the published blob round-trips byte-identical" as not yet delivered, rather than leaving them as claims.
5. `## Plan` steps 1–6 are all still `- [ ]` while the code is written — tick what landed and note what did not, or the close gate's `plan-unchecked` guard is reporting truthfully.

`workshop/issues/000188-issue-id-allocation-from-main.md`'s supersession table attributes the last bullet to #207. That attribution holds structurally (the decision *is* inside the loop), but C3 means the bullet's own guarantee — "never land two files at one id" — is not yet true on the re-allocation path. Worth a line in the table rather than closing on it as-is.

```findings
findings:
  - id: new
    severity: Critical
    family: sync-arm-contract-drift
    title: |
      claim / issue new push trunk commits with an EMPTY subject on the feature-branch arm
    detail: |
      syncViaTrunk passes msg straight to UpdateMany -> `commit-tree -m ""`; the
      deleted arm defaulted it via syncMessage(msg, "issue-sync: update issues
      from branch '<b>'"). runClaim (claim.go:114) and runIssueNew (issue.go:339)
      both pass "". Reproduced against a real bare origin: %s and %B are both
      empty on the pushed commit. Breaks AGENTS.md 12 and `git log --grep "^#N"`
      on the tracker's own trunk; issue.go:337's comment now states the opposite
      of what happens.
  - id: new
    severity: Critical
    family: realloc-ordering-contract
    title: |
      a failed publish removes the ORIGINAL local issue file, inverting the Spec's ordering
    detail: |
      syncViaTrunk (synctrunk.go:49-56) calls rc.finish() whenever rc != nil,
      regardless of err, and finish() os.Remove's rc.OldPath. The Spec says
      "remove the old one only after the push succeeds" and the Done-when asks
      for a test injecting a failure between the local write and the push.
      Reproduced: with the push rejected, 000700-mine.md is deleted and only the
      unpublished 000701-mine.md survives while the trunk holds nothing. The
      fakePublisher.err seam that exists to inject this is at zero call sites,
      propped up by a bare `var _ = errors.Is`.
  - id: new
    severity: Critical
    family: free-id-space-incomplete
    title: |
      nextFreeID searches the trunk only, so re-allocation can mint the duplicate it prevents
    detail: |
      allocateIssueID (issueids.go:33, #213) allocates from the UNION of trunk and
      local. nextFreeID (issuecollision.go:80) sees only the trunk map. Two
      reproduced instances of the class: (a) two local files at one id both
      re-allocate in one prepare pass and both pick 000701, landing two files at
      one id in ONE commit, with rc mispaired so finish() then deletes one of
      them AFTER publishing it; (b) a local unpushed 000701 is ignored, so
      re-allocation collides with it. Feed nextFreeID trunk union local union
      ids-issued-this-pass.
  - id: new
    severity: Important
    family: realloc-ordering-contract
    title: |
      rc is stale state carried across retry attempts and is never invalidated
    detail: |
      A re-allocation recorded on attempt N survives into attempt N+1 even when
      that attempt decides verdictPublish for the original path (reachable when a
      peer withdraws its colliding file, exactly how this repo's own 000207
      duplicate was resolved). Reproduced: the final set publishes
      000700-mine.md, then finish() removes it locally and keeps an unpublished
      000701-mine.md. ARCH-ORDER: reset rc per prepare invocation.
  - id: new
    severity: Important
    family: publish-success-signal
    title: |
      syncViaTrunk emits no success confirmation and drops the `synced` stdout marker
    detail: |
      The deleted arm printed cok(stderr, "Issues synced to main and pushed to
      origin.") and Fprintln(stdout, "synced"). syncViaTrunk prints the file list
      and then nothing, so a successful publish and a whole-set no-op are
      indistinguishable to the operator. issue.go:335 still describes the
      "synced" marker as a contract.
  - id: new
    severity: Important
    family: stale-user-facing-docs
    title: |
      sdlc claim --help still documents the deleted main-worktree route
    detail: |
      cmd/sdlc/helptext/claim.md:32-53,68 lists the worktree hunt, pull --rebase,
      merge-base conflict detection, an entire CONFLICT BEHAVIOR section, exit-1
      causes "missing main worktree, dirty main, conflicts", and the
      "issue-sync: update issues from branch '<branch>'" default the code
      dropped. This is surface a reader types.
  - id: new
    severity: Important
    family: stale-user-facing-docs
    title: |
      atlas is not updated inside the review window
    detail: |
      At b7f08ec, atlas/workflow/sdlc-binary.md:89,117,149,660 still documents
      syncViaMainWorktree; the syncViaTrunk rewrite exists only as an
      uncommitted working-tree edit, outside this range.
      atlas/workflow/issue-sync.md:22-42 still describes the worktree route and
      conflict detector and is untouched even in the working tree.
  - id: new
    severity: Important
    family: donewhen-not-delivered
    title: |
      three Done-when bullets have no test
    detail: |
      (a) "publish while another worktree on main is dirty or mid-rebase" — only
      the absent-worktree half is covered; (b) "a failure injected between the
      local write and the push" — no test, and the seam is unused; (c) "the
      published blob round-trips byte-identical, with attributes applied" —
      asserted only via strings.Contains.
  - id: new
    severity: Minor
    family: stale-user-facing-docs
    title: |
      claim.go's header renames the bullet to syncViaTrunk but keeps the deleted route's six steps
  - id: new
    severity: Minor
    family: dead-code
    title: |
      `_ = branch` at claim.go:159 is dead; branch is already used at line 142
  - id: new
    severity: Minor
    family: dead-code
    title: |
      TrunkView.Read and TrunkFile.TrackingRef are exported at zero call sites
  - id: new
    severity: Minor
    family: dead-code
    title: |
      commitCount's `len(strings.Fields(...))*0 +` term is dead and costs an extra git call
  - id: new
    severity: Minor
    family: tautological-test
    title: |
      TestSyncViaTrunk_CheckRunsOnEveryAttempt asserts a counter the fake increments unconditionally
    detail: |
      pub.runs == 2 cannot fail unless syncViaTrunk errors;
      TestUpdateMany_PathsAreDerivedPerAttempt is the load-bearing one.
  - id: new
    severity: Minor
    family: preexisting-red-suite
    title: |
      go test ./cmd/sdlc/ is red on TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory
    detail: |
      workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md was archived at
      dfeba9c, which precedes the review base — pre-existing, not introduced
      here, but the package cannot be cited as green close evidence without
      naming it.
```
