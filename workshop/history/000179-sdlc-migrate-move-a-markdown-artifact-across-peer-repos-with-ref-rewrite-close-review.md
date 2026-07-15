# Boundary Review — ariadne#179 (whole-issue close)

| field | value |
|-------|-------|
| issue | 179 — sdlc migrate: move a markdown artifact across peer repos with ref rewrite |
| repo | ariadne |
| issue file | workshop/issues/000179-sdlc-migrate-move-a-markdown-artifact-across-peer-repos-with-ref-rewrite.md |
| boundary | whole-issue close |
| milestone | — |
| window | 60bb23096de7ce024860fdbd81541b354a0b0b85..HEAD |
| command | sdlc close --issue 179 |
| reviewer | claude |
| timestamp | 2026-07-15T12:27:56-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I've completed the review — verified the diff against the real seams (`parseRef`, `resolveArtifacts`, `die`/`expectDie`, helptext embed, lock wrapper, `WorkflowVerbs`), cross-checked the plan's Core concepts table, and worked the ARCH markers. One caveat up front: **the Bash tool is broken in this session at the harness level** (EPERM creating `~/.claude/session-env/...` — affects subagents too, the operator may want to check `/sandbox`), so I could not execute the test suite myself; everything below is verified by reading code, not by running it.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

This boundary delivers the issue's full Done-when: the pure `rewriteRefs` core implements exactly the three rewrite rules with fence/span awareness and the parseRef candidate filter, the IO shell verifies every rewrite from the destination's vantage before any write, commits are scoped per side, the inbound sweep reports `repo/file:line`, and the round-trip test pins canonicalization + idempotence. The plan's Core concepts table matches the code row-for-row. What keeps this from a clean SHIP is one guard gap (`--dest-path` is unvalidated and can write a stray file outside the destination repo before dying) plus its complete absence from the test suite — both cheap to fix — and the fact that I could not independently execute the tests (confidence: medium for that reason alone; the main agent should re-run `go test ./cmd/sdlc/... ./cmd/sdlc/internal/...` once Bash works).

**1. Strengths**

- The ARCH-DRY discipline around the grammar is real, not claimed: the scanner (`migrate.go:41`) only finds candidates, `rewriteToken` (`migrate.go:60`) routes every one through `parseRef`, and `TestRefScan_GrammarRoundTrip` pins the construction. I diffed the scanner regex against `parseRef` (`resolve.go:56-101`) and the subset relationship holds (attached-token-only is deliberate and documented as the `ariadne #15` caveat).
- Dest-vantage verification (`migrate.go:292-298`) genuinely exercises the post-move resolution path — `resolveArtifacts(target, destTop)` re-runs `resolveRepoDir` from the destination, which is what makes the non-sibling-dest refusal fall out for free, and the test for it exists.
- The `SplitFences` / `stripCodeFences` divergence (different unterminated-fence policies) is documented on **both** sides with cross-references (`structural.go:222-227, 240-244`) — exactly how a deliberate non-refactor should be recorded.
- Test design is strong: concatenation-identity property in `TestSplitFences`, commit-scope assertion via `git show --stat` pipe-count, `expectDie` refusal coverage for all seven guards, and the symlinked-`$PWD` regression test capturing a real dogfood bug.
- Guard ordering is right: brain-dest check (step 3) fires before the dirty check (step 4), so the brain refusal isn't masked; verification (step 5) strictly precedes the first write (step 6).

**2. Critical findings** — none.

**3. Important findings**

- `cmd/sdlc/migrate.go:265-270` — **`--dest-path` is unvalidated; a traversal value escapes the destination repo and violates the fail-closed claim.** `filepath.Join(destTop, filepath.FromSlash(destRel))` cleans the path, so `--dest-path ../evil.md` yields a `destFile` *outside* `destTop`. The existence check, `MkdirAll`, and `WriteFile` all operate on that outside path; only the subsequent `git add -- ../evil.md` fails (git refuses paths outside the repo), by which point a stray file has been written — potentially into a *third* repo's worktree. The source is untouched, but "nothing half-moved / fail-closed" is broken in letter. Fix sketch: after computing `destFile`, mirror the source-side guard — `rel, err := filepath.Rel(destTop, destFile); if err != nil || strings.HasPrefix(rel, "..") { die(...) }` (also normalizes the absolute-`--dest-path` oddity, where `Join(destTop, "/abs/x")` silently nests it under destTop).
- `cmd/sdlc/migrate_test.go` — **`--dest-path` has zero test coverage** (only the `""` default branch is exercised). This is a shipped user-facing flag; the kind of bug above is exactly what a happy-path `--dest-path other/dir/q.md` test plus a traversal-refusal test would catch. Add both alongside the guard fix.

**4. Minor findings**

- `migrate.go:141` — `migrateOpts.stdout` is assigned in `NewMigrateCmd` but never used; all output goes to stderr. Drop the field or use it for the happy-path summary.
- `migrate.go:299-301` — a file with **zero rewrites** skips verification entirely, so the non-sibling-dest backstop (a side effect of verification) is vacuous for ref-free files; such a file can migrate to a non-sibling repo and the parent-dir inbound sweep then won't cover the destination. Harmless in v1 (report-only sweep), worth a one-line comment or a cheap sibling check.
- `migrate.go:164-168` — `gitInDir` is a fine local helper, but `propagatebase.go:214-271` has five inline `exec.Command("git", "-C", ...)` calls it could consolidate (ARCH-DRY, future sweep — same class as the atlas note about migrating archive logic onto `Discovery()`).
- `migrate.go:307` — `os.WriteFile(..., 0o644)` discards the source file's mode; irrelevant for markdown, noting for completeness.
- `inlineSpanRE` handles single-backtick spans only; a double-backtick span (`` ``…`` ``) is scanned as prose. Rare in artifact bodies and rewrites there are usually *desired*; fine for v1.
- `atlas/index.md:13` — the sdlc-binary hook line enumerates `resolve`/`open` (#144) but not `migrate`; the linked file has the migrate section, so this is just the index one-liner lagging.

**5. Test coverage notes**

Pure core: excellent — 20-case table covering all three rules, fence/span/milestone/gh/edge-digit forms, the `\b` seven-digit pin, plus line-number and grammar-drift tests, all IO-free (PURE table rows check out). IO shell: all guards refusal-tested via `expectDie`, happy path asserts content + commit scope + summary, `--no-commit` asserts staged-only, round-trip asserts the exact canonicalization delta then byte-stability, gatesig collision covered. Gaps: `--dest-path` (above), and no direct test that the skipped-report suppresses verification for gh refs (implicitly covered by the fixture). **I could not run the suite** — the main agent must re-run build/vet/test before recording this verdict's close.

**6. Architectural notes**

- **ARCH-DRY: pass.** Grammar single-sourced in `parseRef` (filter + drift test enforce it); `gitStatusPorcelain` reused for the dest check; the SplitFences non-refactor is documented both sides. The `gitInDir`/propagatebase consolidation is a future minor.
- **ARCH-PURE: pass.** `rewriteRefs`/`SplitFences` are string→string with IO-free tests; `runMigrate` is long but is genuinely guards + orchestration — the rewrite semantics live entirely in the pure core.
- **ARCH-PURPOSE: pass.** Shadow-sweep of the Done-when: all four bullets delivered in this diff; the deferred items (issue renumbering, inbound *rewrite*, gh-ref rewrite) are declared v2 in the issue Spec itself, so they're scoped extensions, not the deferred point. The round-trip bullet is delivered as canonicalization+idempotence — a reasoned refinement the plan records explicitly.
- For v2 renumbering: `rewriteToken`'s switch is well-shaped to take the fourth (self-ref) rule, as the plan anticipates.

**7. Plan revision recommendations**

None required — the plan matches the shipped code, and the symlink-cwd fix is already logged in the issue's `## Log`. If the `--dest-path` guard from finding 3 is applied, append a one-line `## Revisions` entry to `workshop/plans/000179-sdlc-migrate-plan.md` noting the added within-dest-repo containment check (the plan's guard list currently ends at "dest path free").
