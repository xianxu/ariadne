# Boundary Review — ariadne#264 (whole-issue close)

| field | value |
|-------|-------|
| issue | 264 — weave compile: a refused compile must not rewrite .gitignore |
| repo | ariadne |
| issue file | workshop/issues/000264-weave-refused-compile-gitignore.md |
| boundary | whole-issue close |
| milestone | — |
| window | e0373799b0d4b1792b61f6ba548aa8c8a24ad340..1f9cb74806d4ac491f4ab03425cd2877c471e70d |
| command | sdlc close --issue 264 |
| reviewer | claude |
| timestamp | 2026-09-28T19:21:00-07:00 |
| verdict | unknown |

## Review

API Error: Couldn't connect through your proxy (ERR_PROXY_TUNNEL) — the proxy refused the tunnel: check its credentials and that it allows this host

---

## Re-review — 2026-09-28T19:22:59-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 264 — weave compile: a refused compile must not rewrite .gitignore |
| repo | ariadne |
| issue file | workshop/issues/000264-weave-refused-compile-gitignore.md |
| boundary | whole-issue close |
| milestone | — |
| window | e0373799b0d4b1792b61f6ba548aa8c8a24ad340..1f9cb74806d4ac491f4ab03425cd2877c471e70d |
| command | sdlc close --issue 264 |
| reviewer | claude |
| timestamp | 2026-09-28T19:22:59-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The fix is small and aimed at the real cause. `ApplyManaged` now retires the legacy fixed list only in the artifacts pass (`scope == ScopeArtifacts`, `ownership.go:180`), and it returns before any write when a pass has no actions and no prior identities in its scope (`ownership.go:104-108`). A refused artifacts pass fails in its preflight, which is still before `managedIgnore`, `saveInventory` and `writeManagedIgnore`. So after this change the data pass is the only step that can touch `.gitignore` before a refusal, and it now keeps the legacy lines. I checked every Done-when clause against the code and tests. `go test ./cmd/weave/...` passes on HEAD. Nothing blocks SHIP.

1. **Strengths**
   - The fix lives in the ownership rule (`splitIgnore`/`managedIgnore` take `migrate`), not in the order of the two passes in `compilePrepared`. Any other caller of a data-scope pass is protected too, as the Spec intends.
   - The early return also checks for prior identities in this scope, not just for zero actions. Data retirement with zero actions still works: `TestManagedScopesPreserveOtherInventoryAndIgnores` (`ownership_test.go:64`) still retires the data mounts.
   - Plain `Apply`/`EnsureGitignore` keeps migrating (`gitignore.go:33`). Compile's `EnsureGitignore` action goes through the managed path (`ownership.go:275`), so the flag is set correctly for every caller.
   - The two new regressions match the two Done-when clauses, including the "data pass then artifacts pass" order (`ownership_test.go:394-406`).
   - #263's `TestManagedPreservesCommittedBlockWithoutInventory` still passes. It now exercises the early-return path, and its assertions still hold.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **Artifacts-pass assertion is incomplete** (`ownership_test.go:403`): the migration check only proves `/.claude/skills/` is gone. `/construct/generated/` could survive outside the block and the test would still pass. `/AGENTS.md` is re-added inside the block, so its line proves nothing about migration. Fix: assert the text after `# END weave generated\n` equals `.goto\n` exactly.
   - **Atlas not updated** (`atlas/workflow/weave.md:46-53`): the `ApplyManaged` paragraph doesn't say that only the artifacts pass retires the legacy fixed list, or that an empty pass writes nothing. Adding one sentence would help the next reader, even though the current text isn't wrong.

5. **Test coverage notes**
   - Both Done-when regressions are pinned. The implementor's log says both went red on main.
   - The "artifacts pass refuses after a data pass" case is covered in parts rather than end to end: the refusal happens in preflight before any write, the data pass keeps the legacy lines, and the Log records a live parley.nvim check. An end-to-end test would chain a data pass with mounts and a refused artifacts pass, then assert the legacy lines survive. Nice to have, not needed.
   - The only `migrate=false` path is `managedIgnore` through a data scope, which both new tests reach. The malformed-block test still passes `true`, which is fine because malformed-block detection doesn't depend on the flag.

6. **Architectural notes**
   - **ARCH-DRY: pass.** One flag is threaded through the existing helpers, with no parallel copies.
   - **ARCH-PURE: pass.** `splitIgnore`/`managedIgnoreText` stay pure. The scope decision is made once, in the `ApplyManaged` shell.
   - **ARCH-PURPOSE: pass.** I checked every consumer of `splitIgnore`/`managedIgnoreText`: `ensureGitignoreText` and `managedIgnore`, where `managedIgnore` has two call sites. Each gets the flag from the caller's scope. Nothing still strips unconditionally. `GeneratedGitignoreEntries` for dry-run only plans and never writes.

7. **Plan revision recommendations:** none. The code matches the Spec, the Plan and the existing Revisions entry.

```findings
findings:
  - id: new
    severity: Minor
    family: assertion-underconstrains-contract
    title: |
      Artifacts-pass migration test does not prove every legacy line is retired
    detail: |
      ownership_test.go:403 checks only that /.claude/skills/ is gone; /construct/generated/ could survive outside the block and the test still passes (/AGENTS.md is re-added inside the block, so it proves nothing either). Assert the text after the END marker is exactly ".goto\n".
  - id: new
    severity: Minor
    family: atlas-surface-lag
    title: |
      weave atlas does not record the artifacts-only legacy migration or the empty-pass no-op
    detail: |
      atlas/workflow/weave.md:46-53 describes ApplyManaged's ignore ownership; one sentence noting that only the artifacts pass retires the legacy fixed list and that an empty pass writes nothing would keep the map current.
```

---

## Re-review — 2026-09-28T20:39:06-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 264 — weave compile: a refused compile must not rewrite .gitignore |
| repo | ariadne |
| issue file | workshop/issues/000264-weave-refused-compile-gitignore.md |
| boundary | whole-issue close |
| milestone | — |
| window | e0373799b0d4b1792b61f6ba548aa8c8a24ad340..1494c0b8df4f8af4641e5719d173ee5a018559fc |
| command | sdlc close --issue 264 |
| reviewer | claude |
| timestamp | 2026-09-28T20:39:06-07:00 |
| verdict | SHIP |

## Review

Checks are done (tests pass, both prior findings verified). Here is the verdict.

```verdict
verdict: SHIP
confidence: high
```

This change fixes the bug at its source, in the ownership rule, as ARCH requires. `splitIgnore` and `managedIgnoreText` now take a `migrate` flag. `ApplyManaged` sets it only when `scope == ScopeArtifacts`, and the plain `EnsureGitignore` path still migrates. A new early return makes a pass that has no actions and no prior identities in its scope write nothing. Both Done-when regressions are in place and `go test ./cmd/weave/...` passes at HEAD. Both prior Minor findings are fixed. Nothing blocks SHIP.

1. **Strengths**
   - `ownership.go:180` sets the migrate flag from the scope that owns the replacement. That is the root cause, and the flag is not keyed to caller order, so any other data-pass caller is covered too.
   - The early return at `ownership.go:104-108` runs after `readInventory` and before any write. It only skips work when the scope has nothing to retire, so retirement of existing entries is untouched.
   - `TestManagedDataPassKeepsLegacyListForArtifacts` covers both states of the migrate clause: a data pass keeps the list, then an artifacts pass retires it. The `HasSuffix(got, ignoreEnd+"\n.goto\n")` check pins the whole retirement.
   - I checked every caller of `splitIgnore`, `managedIgnoreText`, `managedIgnore` and `ApplyManaged`. All pass an explicit flag and none were missed (`gitignore.go:33`, `main.go:523`, `main.go:572`).

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** none new. The empty-pass test only runs `ScopeData`, though the rule applies to any scope. The artifacts case uses the same line of code and Done-when only names the data pass, so I'm not raising it.

5. **Test coverage:** Both regressions exist and fit the log's claim that they failed on main. Existing tests still cover the malformed-block cases (now called with `migrate=true`) and #263's committed-block preservation.

6. **Architecture:**
   - **ARCH-DRY: pass.** One flag is threaded through the existing helpers; no logic is copied.
   - **ARCH-PURE: pass.** `splitIgnore` and `managedIgnoreText` stay pure. The migrate decision is made once, in the IO shell (`ApplyManaged`).
   - **ARCH-PURPOSE: pass.** Both parts of the invariant are delivered: only the artifacts pass migrates, and an empty pass writes nothing. The code change and its tests are in the same commit, and nothing essential is left as a follow-up.

7. **Plan revisions:** none needed.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      ownership_test.go now asserts HasSuffix(got, ignoreEnd+"\n.goto\n"), so any legacy line (/construct/generated/, /.claude/skills/) surviving after the END marker fails the test.
  - id: BR-2
    disposition: addressed
    note: |
      atlas/workflow/weave.md:50-53 now states only the artifacts pass retires the pre-inventory fixed list and that a pass with nothing to apply or retire writes nothing; matches ownership.go:104-108,180.
```
