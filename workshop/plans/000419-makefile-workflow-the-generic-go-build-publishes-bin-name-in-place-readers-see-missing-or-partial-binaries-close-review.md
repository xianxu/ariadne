# Boundary Review — ariadne#419 (whole-issue close)

| field | value |
|-------|-------|
| issue | 419 — Makefile.workflow: the generic Go build publishes bin/<name> in place; readers see missing or partial binaries |
| repo | ariadne |
| issue file | workshop/issues/000419-makefile-workflow-the-generic-go-build-publishes-bin-name-in-place-readers-see-missing-or-partial-binaries.md |
| boundary | whole-issue close |
| milestone | — |
| window | a5acc4c4ba680f1b5bbbd3c9f59b74dbf9c84252..2420f13a7f7530398a412a04c188c46c8f265260 |
| command | sdlc close --issue 419 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-17T00:19:03-07:00 |
| verdict | SHIP |

## Review

Diff read; now running the new and modified tests in the checkout.

Tests pass. Now confirming the new tests fail when the old in-place build is restored.

Revert check passed: all three new tests fail on the in-place build. Last, reading the issue's Spec for any requirement not yet covered.

```verdict
verdict: SHIP
confidence: high
```

The change does what #419 asks. `WF_GO_PUBLISH` (`Makefile.workflow:676-679`) builds each binary into its own `mktemp -d` staging directory next to the target. It then replaces the target with a single `mv -f`. A trap removes the staging directory when the build fails or is interrupted. Both callers use it: the generic `make build` loop and ariadne's own `tools`/`weave-build` targets. I ran the three new tests, the updated trimpath test and `portable-makefile.test.sh` at head, and all passed. I then put the old in-place `go build -o "bin/$$name"` line back and reran the new tests. All three failed with `bin/x = "partial"` in the mid-build, failed-build and interrupted-build cases. So the tests do catch the bug this issue fixes. All three "Done when" items are delivered. Only Minor notes remain.

**1. Strengths**
- **One publisher, two callers** (`Makefile.workflow:676`, `Makefile:36,43`). The pattern isn't copied into a second place (ARCH-DRY).
- **Safe staging.** The publisher runs in a subshell, so its staging variable `d` doesn't overwrite the build loop's own `$$d`. The staging directory sits next to the target, so the rename stays on one filesystem and is atomic. `"$${d:?}"` guards the `rm -rf`.
- **Strong tests** (`cmd/sdlc/makebuild_publish_test.go`). The fake `go` writes a partial file first, like the real compiler does. The tests run the real `Makefile.workflow` (symlinked into a temp repo) and pause the build at a controlled point, so they check what a reader sees mid-build, not just the end state. There is also a SIGTERM test on the process group.
- **Honest fixture fixes.** The fake `go` in `portable-makefile.test.sh` now does `chmod +x`, which matches real `go build` output. The Log explains why the rename made that necessary.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- **ARCH-FUNERAL:** if the build is killed with SIGKILL or the machine loses power, a `bin/.go-build.*` directory is left behind and nothing ever removes it. `bin/` is gitignored and each one is a single binary, so the cost is small. A future `make build` could sweep stale `.go-build.*` directories older than a few minutes.
- **Docs:** `atlas/workflow/setup-and-replication.md:215` describes `make build` but doesn't say it now publishes by rename. A one-line mention would help. There is no new command, flag or other user-facing surface, so README needs no change.

**5. Test coverage notes**
- The generic loop is covered by behavioural tests: paused, failing and interrupted builds.
- The owner `Makefile` targets are covered only by the trimpath test's match on the staging path. That's adequate because they use the same macro.
- Concurrent builds of the same target aren't tested. Each build gets its own staging directory and the last rename wins, which is acceptable.

**6. Architecture**
- **ARCH-DRY: pass.** The `pair` repo's `build-go-atomic.sh` is a cross-repo duplicate. That's outside this repository.
- **ARCH-PURE: pass.** This is a thin build shell with no business logic.
- **ARCH-PURPOSE: pass.** I checked the other `go build -o` sites. `Makefile.workflow:138` (issue-sync) and `scripts/merge-checks.d/40-duplicate-issue-id.sh:87` both build to private temporary paths that nothing else reads, so neither has this defect.
- **ARCH-MOCK: pass.** A fake `go` on `PATH` sits at the same seam the real compiler uses.
- **ARCH-CONSTRAINTS: pass.** It adds one rename per binary and nothing else.
- **ARCH-SECURE: pass.** There's no untrusted input or secrets, and the `${d:?}` guard protects the `rm -rf`.
- **ARCH-ORDER: pass.** The only state is the filesystem. Concurrent builds use separate staging directories, the single rename is the atomic step, and the interrupt ordering is tested.
- **ARCH-FUNERAL: note.** The SIGKILL residue above.

**7. Plan revision recommendations:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: artifact-without-removal-path
    title: |
      SIGKILL or power loss mid-build leaves bin/.go-build.* staging dirs with no sweep (ARCH-FUNERAL)
    detail: |
      The EXIT trap covers failure, INT and TERM but cannot run on SIGKILL, and nothing removes a stale staging dir later. Each one is a single binary in gitignored bin/. Optional fix: sweep .go-build.* dirs older than N minutes at the start of WF_GO_PUBLISH or make build.
  - id: new
    severity: Minor
    kind: defect
    family: atlas-surface-undocumented
    title: |
      atlas/workflow/setup-and-replication.md describes make build but not its publish-by-rename guarantee
    detail: |
      There is no new user-facing surface. A one-line atlas note would record that readers of bin/NAME never see a missing or partial binary during a rebuild.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"bb4f756e-2747-4ce1-89f6-e1b592bff97b","repository":"github.com/xianxu/ariadne","issue_id":"000419","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/d56aac107d4bea50","reviewed_head":"2420f13a7f7530398a412a04c188c46c8f265260","slot":"ariadne:4","requested_at":"2026-10-17T07:18:01.054879Z"},"phase":"finished","pid":51547,"started_at":"2026-10-17T07:18:01.151282Z","finished_at":"2026-10-17T07:19:03.292084Z","session_id":"0b097383-0f45-4c77-916d-ae464cffe30f","outcome":"returned"}] -->
