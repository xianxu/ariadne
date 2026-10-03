# Boundary Review — ariadne#289 (whole-issue close)

| field | value |
|-------|-------|
| issue | 289 — Slot readiness in sdlc fleet inventory |
| repo | ariadne |
| issue file | workshop/issues/000289-slot-readiness-in-sdlc-fleet-inventory.md |
| boundary | whole-issue close |
| milestone | — |
| window | 86ab9fe92f9832cd8c63bfed3e22b212fef67e9a..457b63372007b9f878131aef2b8865733374a1e5 |
| command | sdlc close --issue 289 |
| reviewer | claude |
| timestamp | 2026-10-02T17:31:23-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Two of the four open findings are fixed. BR-7 (stale names) and BR-13 (the shared reader had no test of its own) are done. BR-14 has the right code fix: the pending-diagnostic flush now runs after dependency-clone collection (`inventory.go:127-135`). But its new regression test does not reach the code it is meant to protect. I put the old flush order back in a scratch worktree and `TestBrokenDependencyCloneIsReported` still passed. The empty `.git` it uses fails at `rev-parse --git-common-dir`, which records its diagnostic straight away and never reaches the pending path. BR-12 is unchanged: three `construct/deps` reads still bypass `layergraph.ReadDeclaration`. Both are Minor, so neither blocks the close. Nothing new was found in `slots.go` or `membership.go`.

1. **Strengths**
   - `pkg/layergraph/read.go`: one guarded reader. It checks the file before and after opening, opens with `O_NOFOLLOW|O_NONBLOCK`, and reads a bounded amount with an overflow probe. `read_test.go` now tests the symlink, FIFO, oversize and missing cases directly. `startplan.go:411` uses it too.
   - `pkg/workspace/operation.go`: one `OperationMarkers` list. `ActiveOperation` returns an error on any lstat failure other than "not found", so an unreadable directory can never read as "no operation".
   - `slots.go` `JudgeCheckout`/`AssembleSlots`/`withProbe` are pure, with a clear precedence. A failed probe can never leave a checkout `ready`, and `needs-recovery` wins over unread facts.
   - `DeclaredMembers` is pure: read, stat and canonicalization are passed in, and it applies weave's placement rule to the declared and canonical paths together.

2. **Critical**: none.

3. **Important**: none.

4. **Minor**
   - **BR-14, still open.** The fix is in place but has no regression test that fails without it. Suggested fix: in `fleetslots_test.go`, wrap `execGitReader` with an injected `GitReader` that fails `worktree list` only for the clone path. Then assert a `Stage: "worktrees"` diagnostic for that path. With the old flush order, that test should fail.
   - **BR-12, still open (3rd in family).** The rule: no non-test code opens `construct/deps` except through `layergraph.ReadDeclaration`. Sites that still break it:
     - `pkg/layergraph/fs.go:43`: `OSFS.ReadFile`, used by `walk.go:96`.
     - `cmd/weave/link.go:69`: `fs.ReadFile`.
     - `cmd/weave/internal/acquire/acquire.go:410`: falls back to `os.ReadFile` when `MaxDeclarationBytes<=0`.

     Enforce the rule with a guard test, not by fixing one site at a time.

5. **Test coverage notes:** BR-14's test passes whether or not the fix is present (checked by reverting the fix in a scratch worktree). `read_test.go` exercises the real filesystem cases. The slot verdict tests are pure.

6. **Architectural notes**
   - ARCH-DRY: pass, except the BR-12 reader sites.
   - ARCH-PURE: pass.
   - ARCH-PURPOSE: pass. Clones become rows and each slot gets one verdict.
   - ARCH-MOCK: pass. Git is injected through `GitReader`, and the e2e tests use real git in temporary directories.
   - ARCH-CONSTRAINTS: pass. Reads are bounded and no new git processes are added per clone beyond the row walk.
   - ARCH-SECURE: flagged (BR-12). Some readers still follow symlinks or FIFOs.
   - ARCH-ORDER: flagged (BR-14). The collect-then-flush order now looks right, but no test pins it.
   - ARCH-FUNERAL: pass. Nothing durable is created except review artifacts, which are archived with the issue.

7. **Plan revisions:** none needed. The plan matches the code.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      Issue line 104 now names the pkg/workspace detector; issuemovedetail.go:70 comment drops root; plan line 156 says REBASE_HEAD comes first.
  - id: BR-12
    disposition: not-addressed
    note: |
      Still bypassing ReadDeclaration: pkg/layergraph/fs.go:43 (OSFS.ReadFile used by walk.go:96), cmd/weave/link.go:69 fs.ReadFile, acquire.go:410 os.ReadFile fallback; no guard test.
  - id: BR-13
    disposition: addressed
    note: |
      pkg/layergraph/read_test.go covers ordinary, symlink, FIFO, oversize and missing.
  - id: BR-14
    disposition: not-addressed
    note: |
      Flush move is correct, but TestBrokenDependencyCloneIsReported passes with the old order (verified in a scratch worktree): an empty .git fails at rev-parse and is reported at once, never pending. Needs a test whose clone fails worktree list (injected GitReader).
```
