# Boundary Review — ariadne#355 (whole-issue close)

| field | value |
|-------|-------|
| issue | 355 — dev-aliases: build to a temp file and rename into place, so a rebuilding alias never leaves the binary missing |
| repo | ariadne |
| issue file | workshop/issues/000355-dev-aliases-build-to-a-temp-file-and-rename-into-place-so-a-rebuilding-alias-never-leaves-the-binary-missing.md |
| boundary | whole-issue close |
| milestone | — |
| window | 57f9b5c9e1b2e4c1b021dce8f09959c9eacc58b4..f1d6eae5d416f581980b86e5bbcfe249191929df |
| command | sdlc close --issue 355 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T20:42:39-07:00 |
| verdict | SHIP |

## Review

Test suite passes (48/0 under /bin/bash 3.2). Now checking that the generated function also works under zsh, which is the operator's shell.

```verdict
verdict: SHIP
confidence: high
```

This change fixes the reported failure, where `bin/X` went missing while an alias rebuilt it. The emitter at `construct/dev-aliases.sh:272` now builds into a fresh `mktemp -d bin/.X.XXXXXX` directory, then uses `mv -f` to rename the result over `bin/X` and removes the build directory. The exit status is kept, so a failed build still fails the call and runs nothing.

I checked it four ways:
- **Diff inspection:** I ran the stat, name-status and full-patch commands on the pinned range.
- **Test suite:** `dev-aliases.test.sh` passes under `/bin/bash` 3.2 (48 passed, 0 failed).
- **zsh check:** the emitted function works under `zsh -f`, which is what the operator's shell runs. A successful build installs the binary and leaves only `bin/foo`. A failed build returns rc=1, keeps the previous binary and leaves no build directory.
- **Revert check:** I swapped the base version of the emitter back in. The tests went red: 7 failed. Four of those are shape assertions. The other three are behavioural: the poller found the binary missing during a rebuild, the final binary was not whole, and the previous binary was lost. So the new tests do catch the bug this issue is about.

Nothing blocks SHIP.

Done-when: current — its clauses all still hold: build and rename, a concurrency test with two rebuilds and an exec poller, and a passing run under bash 3.2. The only drift is that it says "temp file" where the code uses a temp directory. `## Revisions` records and explains that change: `go build -o` won't overwrite a file it didn't build, and `$$` is not unique inside the function's subshell.

1. **Strengths**
   - `construct/dev-aliases.sh:272`: the build directory sits inside `bin/`, so the rename stays on one filesystem and is atomic. `s=$?; rm -rf "${d:?}"; exit $s` cleans up on both success and failure without losing the exit status, and `${d:?}` guards `rm -rf` against an empty `d`.
   - The concurrency test (`dev-aliases.test.sh:224-248`) uses a slow stub that writes its output in two steps. The poller checks for a complete binary, not just that the file exists. This is what makes the test catch the old `rm -f` window.
   - The failed-build test (`:250-263`) covers both outcomes: success, and failure that keeps the old binary. The old form destroyed the previous binary on failure, and the revert run confirms the test catches that.
   - The plain-form test still compares the entire emitted line exactly, so any accidental change to its shape will fail.
   - The atlas (`atlas/workflow/base-layer.md`) and the header comment were updated in the same range. Their claims match the code: atomic rename, separate directories per call, last rename wins, a new inode, and a killed build leaves its directory behind.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The emitted form is written out by hand in three places: the format string, the header comment at `dev-aliases.sh:12-13` and the atlas paragraph. This duplication existed before this change, and the exact-line test only guards the format string (ARCH-DRY, minor). Those three places are the whole set.
   - The note that a killed build's directory is "gitignored with bin/" is confirmed only for ariadne (`.gitignore:77`). I couldn't check other owner repos from here.

5. **Test coverage:** this issue's behaviour is covered in both directions, success and failure, and concurrently. The poller runs for a fixed 3 seconds while the builds take about 0.8 seconds, which leaves enough margin. There is no zsh test; my manual zsh run passed.

6. **Architecture**
   - **ARCH-DRY:** pass, apart from the doc duplication noted under Minor.
   - **ARCH-PURE:** pass. The change is a shell emitter with no business logic mixed into IO.
   - **ARCH-PURPOSE:** pass. Every emitted function comes from the single printf at `:272`, so every owner and binary gets the fix. A search finds no other `rm -f bin/X` rebuild form in the tree.

7. **Plan revisions:** none needed. `## Revisions` already records the switch from a temp file to a temp directory.

```findings
findings:
  - id: new
    severity: Minor
    family: hand-restated-emitted-form
    title: |
      Emitted alias form is hand-restated in the header comment and atlas besides the printf
    detail: |
      There are three copies: the printf at construct/dev-aliases.sh:272, the header comment at construct/dev-aliases.sh:12-13 and the atlas/workflow/base-layer.md paragraph. Only the printf is guarded by the exact-line test. All three match today. This duplication predates this change.
```
