# Boundary Review — ariadne#296 (whole-issue close)

| field | value |
|-------|-------|
| issue | 296 — weave: recover a sourceless substrate's source from the primary checkout's sibling, and migrate the row |
| repo | ariadne |
| issue file | workshop/issues/000296-weave-recover-a-sourceless-substrate-s-source-from-the-primary-checkout-s-sibling-and-migrate-the-row.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0fb92aa4f4492026d003ef8fbdf9aa30f07fe49d..6e622048033fce61ca040be301a997647ea65bc5 |
| command | sdlc close --issue 296 |
| reviewer | claude |
| timestamp | 2026-10-06T09:18:03-07:00 |
| verdict | unknown |

## Review

API Error: Couldn't connect through your proxy (ERR_PROXY_TUNNEL) — the proxy refused the tunnel: check its credentials and that it allows this host

---

## Re-review — 2026-10-06T09:24:02-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 296 — weave: recover a sourceless substrate's source from the primary checkout's sibling, and migrate the row |
| repo | ariadne |
| issue file | workshop/issues/000296-weave-recover-a-sourceless-substrate-s-source-from-the-primary-checkout-s-sibling-and-migrate-the-row.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0fb92aa4f4492026d003ef8fbdf9aa30f07fe49d..6e622048033fce61ca040be301a997647ea65bc5 |
| command | sdlc close --issue 296 |
| reviewer | claude |
| timestamp | 2026-10-06T09:24:02-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The issue's purpose is delivered, and the change matches the Spec and the operator's choice of migration (a). In a slot, a sourceless substrate row now recovers its source from the primary-side sibling's remote `origin`. Every recovery prints a notice that names the row, URL and sibling. All five refusal cases keep the missing-substrate error and add the reason. A real compile or `weave dependencies` in a primary checkout writes the origin into the row through the row writer that was extracted from `weave link` (`declareSubstrate`). Slots, linked worktrees, dry runs and failed restores never write. I inspected the window with the stat, name-status and full-diff commands. `go test ./cmd/weave/...` passes at HEAD 6e622048. Nothing blocks the close. Two Important findings are cheap to fix first: the README doesn't mention that compile now edits `construct/deps`, and the "own checkout with a remote origin" check is written twice with different reporting.

## 1. Strengths
- **Recovery reuses the existing acquisition path.** `cmd/weave/internal/acquire/acquire.go:347-355` swaps the recovered `Source` in before the identity-conflict map, `Ensure` and the dry-run `Missing` handling. Recovery therefore gets the same post-clone `origin/main` and manifest checks as a declared source. It doesn't build a parallel clone path.
- **The scope is right.** Recovery is keyed on `Policy.PrimaryRoot`, which comes from Git-proved discovery (`environment.go:43`), so it only happens inside a numbered environment.
- **Untrusted origins can't leak or misbehave.** The sibling's origin goes through `sourceAt`/`NormalizeSource`, which rejects credentials, whitespace, `#`, a leading dash and file sources. The error messages never echo the raw URL (ARCH-SECURE).
- **`declareSubstrate` is a clean extraction** (`link.go:69`). The migration and `weave link` now share one writer, and it keeps comments.
- **The tests use real git and stateful fixtures.** They include a warm-reuse check after the sibling is deleted, a positive control in `TestMigrateSourcelessNeverWritesOutsidePrimary`, and an end-to-end `weave dependencies` run in a numbered environment that uses `url.insteadOf`.

## 2. Critical findings
None.

## 3. Important findings
- **The README doesn't mention the new user-visible behaviour.** `README.md:80-85` still says "record a source to restore a missing one". Two things a reader would notice are missing: `weave compile`/`weave dependencies` in a primary checkout now edit the tracked file `construct/deps`, and slots recover sourceless rows. The atlas was updated; the README wasn't. Fix: add a sentence or two near line 84.
- **ARCH-DRY / ARCH-MOCK: the same check is written twice.** `recoverSource` (`acquire.go:398-427`) and `remoteOrigin` (`migrate.go:80-91`) both test "own toplevel, has an origin, origin parses, not a `file:` source". They differ in three ways:
  - One uses `c.git` + `canonical`; the other uses `setupGitReader` + `samePath`.
  - `remoteOrigin` reduces the result to a bool, so a local, credentialed or invalid origin all produce the warning "has no remote origin", which is wrong for those cases.
  - `migrateSourceless` bypasses the injected `client` that `finishRestore` already receives; it calls the package-level `acquire.Origin`.

  Fix: export one `acquire.Client.RemoteOrigin(ctx, dir) (Source, error)` that returns the reason. Use it in both places, and print that reason in the migration warning.

## 4. Minor findings
- The plan says that outside a numbered environment the error stays "unchanged". The code adds `(no source declared; no numbered environment to recover it from)`. The added text is harmless, but the plan and the code disagree.
- `migrateSourceless` runs two `git rev-parse` calls on every primary compile, even when there are no sourceless rows. Parsing the rows first would skip them.
- `finishRestore` takes eight parameters and passes `err` straight through. A small result struct, or calling it only on success plus a separate notice printer, would read better.

## 5. Test coverage notes
All Done-when items are covered: the slot clone and its notice, five refusals, the outside-environment case, the primary rewrite (idempotent, comment kept), and no writes in a linked worktree, a slot, a dry run or after a failed restore. The live `tools:1` reconcile is recorded in the Log; I didn't re-run it. Gaps:
- No test for a sourceless row that is missing in a slot and declared by a transitively cloned substrate. The owner in that case is a private clone, and the notice tells the operator to edit that clone's deps.
- No test for a primary-side sibling that is a symlink or a linked worktree.

## 6. Architectural notes
- **ARCH-DRY:** flagged above.
- **ARCH-PURE:** pass, with a note. The checks are git-probe glue, which is acceptable. The pure part, row selection, sits inside IO loops but is small.
- **ARCH-PURPOSE:** pass. All Done-when items are delivered, and #293 is revised to drop the manual fleet edit.
- **ARCH-MOCK:** minor flag. It's folded into the DRY finding: migration bypasses the `Client.Git` seam. The tests use real git, which follows this repo's convention.
- **ARCH-CONSTRAINTS:** pass. A few git calls per compile.
- **ARCH-SECURE:** pass. The origin is parsed into a typed `Source` and credentials are rejected.
- **ARCH-ORDER:** pass. Nothing is held between events: `Restore` reads the filesystem as its source of truth, and slot writes happen under the existing setup lease. The primary-side deps edit has no lease, but two concurrent compiles in one primary checkout are an operator-level race, not a new one.
- **ARCH-FUNERAL:** pass. The recovered clone belongs to the environment's existing clone family. The only other durable artifact is the one-line deps edit, which the operator commits.

## 7. Plan revision recommendations
- Add a `## Revisions` note that the error outside an environment now carries a "no numbered environment" reason, or make the code match the plan's "unchanged".

```findings
findings:
  - id: new
    severity: Important
    family: readme-user-surface-sync
    title: |
      README omits that compile/dependencies in a primary checkout rewrite construct/deps and that slots recover sourceless rows
    detail: |
      README.md:80-85 still says to record a source by hand to restore a missing one. Compile now edits a tracked file in a primary checkout, and slots recover from the primary-side sibling. Both are user-visible and should be described there; only atlas/ was updated.
  - id: new
    severity: Important
    family: shared-helper-not-extracted
    title: |
      The "own checkout with a remote origin" check is written twice (recoverSource and remoteOrigin) with different reporting
    detail: |
      acquire.go:398-427 and migrate.go:80-91 both check own toplevel, origin, parse, and not-file, using different git readers and path comparisons. remoteOrigin reduces this to a bool, so the migration warning wrongly says "has no remote origin" for local, credentialed or invalid origins, and it bypasses the injected Client.Git seam. Export one acquire.Client.RemoteOrigin(ctx, dir) (Source, error) that returns the reason, and use it in both places.
  - id: new
    severity: Minor
    family: plan-code-drift
    title: |
      The error outside a numbered environment is extended, but the plan says it stays unchanged
    detail: |
      The code adds "(no source declared; no numbered environment to recover it from)". Either note this in the plan's Revisions or keep the old message.
  - id: new
    severity: Minor
    family: avoidable-repeated-work
    title: |
      migrateSourceless runs the primaryCheckout git probes on every compile, even with no sourceless rows
  - id: new
    severity: Minor
    family: function-signature-shape
    title: |
      finishRestore takes eight parameters and passes Restore's err straight through
```
