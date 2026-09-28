# Boundary Review — ariadne#259 (whole-issue close)

| field | value |
|-------|-------|
| issue | 259 — close evidence and other status readers drop the first modified path |
| repo | ariadne |
| issue file | workshop/issues/000259-status-paths-first-entry.md |
| boundary | whole-issue close |
| milestone | — |
| window | 968861f13e4d6a20fc281d79b4e080d405b6108d..30d9c17aaea0f60e033f0ac1a58b7cc78b1efe48 |
| command | sdlc close --issue 259 |
| reviewer | claude |
| timestamp | 2026-09-28T09:53:57-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** The fix goes after the root cause. There is now one byte-exact parser, `gitx.ParseStatusZ`, for `--porcelain=v1 -z` output, and `trackerEnv.statusEntries` gets its input without trimming, so the first entry's leading status space survives. Both callers named in the Spec (`closeEvidence` and migrate's `dirtyIssuePaths`) use it, and fleet's count now uses the same parser instead of its own copy. I checked the regression claim myself: in a scratch worktree at head, I put back the base versions of `closetracker.go` and `issuemigrate.go`, and both `TestTrackerReCloseCommitsTheModifiedGateLedger` and `TestIssueMigrateNamesADirtyPathWithASpace` failed. At head, the gitx and fleet packages and the three targeted tests pass. All three `## Done when` clauses are covered. Nothing blocks shipping; the findings below are all Minor.

1. **Strengths**
   - `cmd/sdlc/internal/gitx/status.go:21` fails closed. A trimmed first entry (`"M x\x00"`) is rejected rather than guessed at, and the test pins that case (`status_test.go:29`).
   - Moving fleet's validating counter into gitx gives one parser (ARCH-DRY). Fleet's fake git uses the same `ValidStatusCode`, so the fake and the parser can't drift apart.
   - Renames are handled correctly: in `-z` output `Path` is the destination and `Orig` the source. Migrate's old `dest`-preference logic therefore reduces to `e.Path` without changing behaviour.
   - The re-close test reproduces the real failure: the first close creates the ledger as new (`??`), and the second close has it as modified (` M`). It also checks that `workshop/plans` is left clean.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **`statusEntries` copies `gitEnv`'s exec plumbing** (`trackerenv.go:88`): the command setup, stderr capture and error formatting (ARCH-DRY). A small `gitRaw(dir, args) ([]byte, error)` that `gitEnv` also calls would remove the copy. The copy is cheap to leave.
   - **Some status readers still don't handle paths with spaces.** Merge's `porcelainPaths` (`merge.go:202`) splits on whitespace from trimmed output. Push's `parsePorcelainStatus` (`push.go:516`) slices `line[3:]` from untrimmed `CombinedOutput`, so the first-entry bug doesn't hit it, but it can't handle quoted paths with spaces. Both are outside the Spec's literal claim ("no reader slices fixed columns from trimmed text" holds), and the Log explains why merge was left alone (it fails closed). But the new lesson says to read status through `ParseStatusZ`, and these two readers still don't.
   - **Import grouping in `internal/fleet/facts.go:7`:** the gitx import sits in the standard-library group. goimports would move it to its own group.

5. **Test coverage**
   - The unit test covers modified-first, untracked, rename, spaces, empty input, and four malformed cases.
   - Both end-to-end tests fail without the fix; I checked this by reverting it.
   - Not covered: a copy entry (`C `) and a `" R"`-in-worktree code. They take the same branch as rename, so the risk is low.

6. **Architecture**
   - **ARCH-DRY: pass.** Two parsers became one. The exec copy is noted under Minor.
   - **ARCH-PURE: pass.** `ParseStatusZ` is pure and tested without IO. `statusEntries` is a thin IO layer around it.
   - **ARCH-PURPOSE: pass.** Every consumer named in the Spec now uses the single parser. The readers left out are recorded in the Log and don't drop paths.

   Going forward, new status readers should call `statusEntries` or `ParseStatusZ` and never `git()`.

7. **Plan revisions:** none needed. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: status-read-single-parser
    title: |
      Remaining porcelain readers (merge porcelainPaths, push parsePorcelainStatus) still bypass ParseStatusZ
    detail: |
      merge.go:202 splits on whitespace over trimmed output; push.go:516 slices line[3:] over untrimmed non -z output. Neither drops the first entry, but both mis-handle quoted paths with spaces; the new lesson prescribes ParseStatusZ for status reads.
  - id: new
    severity: Minor
    family: exec-plumbing-duplication
    title: |
      trackerEnv.statusEntries duplicates gitEnv's exec/stderr/error plumbing
    detail: |
      trackerenv.go:88 vs gitEnv; a shared untrimmed gitRaw(dir, args) helper would keep one exec path (ARCH-DRY).
  - id: new
    severity: Minor
    family: import-grouping
    title: |
      fleet/facts.go puts the gitx import in the stdlib import group
```
