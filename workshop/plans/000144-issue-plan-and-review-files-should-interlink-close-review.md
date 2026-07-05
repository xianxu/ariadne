# Boundary Review — ariadne#144 (whole-issue close)

| field | value |
|-------|-------|
| issue | 144 — sdlc resolve — read-only artifact-reference resolver |
| repo | ariadne |
| issue file | workshop/issues/000144-issue-plan-and-review-files-should-interlink.md |
| boundary | whole-issue close |
| milestone | — |
| window | 270662ff97907aa2ecba1bf1b6833857a893735f..HEAD |
| command | sdlc close --issue 144 |
| reviewer | claude |
| timestamp | 2026-07-05T12:57:40-07:00 |
| verdict | SHIP |

## Review

All four independent runs of the test command exited 0. I have everything I need. Here is my review.

```verdict
verdict: SHIP
confidence: high
```

**Summary.** This is a clean, well-architected boundary. `sdlc resolve`/`open` deliver exactly the #144 Spec: a read-only, lock-free (by construction — never tagged `markMutatingCommand`) artifact resolver whose ref grammar is single-sourced as the sole parser (`parseRef`) and whose locations derive entirely from the vocab `discovery:` model (zero hardcoded paths in `resolve.go`). The pure/IO split is textbook ARCH-PURE, and the test suite is thorough (parser edge cases, family ordering, exact-vs-prefix-vs-ambiguous repo matching, archive-union globbing, milestone narrow + distinct missing-error, github labeling, primary-target selection, and both structural + runtime lock-free proofs). Build passes; `go test ./cmd/sdlc/... ./pkg/vocab/...` is green across four independent runs. Nothing blocks SHIP; only minor polish notes below.

**1. Strengths**
- Genuinely pure core: `parseRef` / `classifyFamily` are string→struct and unit-tested with no IO (`resolve_test.go:325`, `:297`) — no mocks needed, so the PURE claim in the Core-concepts table holds.
- Lock-free proven two ways — structural `commandNeedsRepoLock(NewResolveCmd())==false` and a *runtime* test that holds a real `repolock.Acquire` lock then resolves (`resolve_test.go:177`, `:190`). This is the honest read-only path parley uses (calls `runResolve` directly, sidestepping the cwd-keyed Execute lock).
- Single-source discipline: no other grammar parser exists in-repo (grep for a Lua/Go re-encoding is empty), and `TestResolveDocExamplesParse` binds every documented example back to `parseRef` so the help can't drift (`resolve_test.go:22`).
- Model-derived locations: `familyFiles` globs `d.Home/d.Plans/d.Archive` from injected `vocab.Discovery` (`resolve.go:225`); the archive-union is what makes resolution correct post-`issues/→history/` — and `TestFamilyFiles` seeds a partially-archived family to prove it (`resolve_test.go:232`).
- `resolveArtifacts` is correctly shared between `runResolve` and `runOpen` (ARCH-DRY), so parse→glob→classify→narrow lives once.

**2. Critical findings** — none.

**3. Important findings** — none.

**4. Minor findings**
- `resolve.go:311` — `resolveResult.Files` lacks `omitempty` and isn't initialized, so a `gh#id` `--json` emits `"files":null` while the documented schema (`resolve.md:54`) shows `files:[...]`. A JSON consumer iterating `.files` on null must special-case it. Consider `omitempty` or `Files: []resolveFile{}`.
- `resolve.go:345-355` vs `:416-423` — the github-labeling block (`who` fallback to `filepath.Base(root)` + `github:%s#%d` print) is duplicated across `runResolve`/`runOpen`. Marginal, since the two differ slightly (JSON branch, suffix text), but a 3-line `githubLabel(ref, root)` helper would consolidate it.
- `atlas/index.md:13` says "12 verbs"; the visible verb list is now 18 (incl. the new `resolve`/`open`). This count was already stale pre-#144, but the diff adds two more. The load-bearing atlas file (`sdlc-binary.md`) *is* updated, so this is optional polish, not a gate miss.
- `parseRef` is slightly more lenient than the documented grammar: it accepts `repo #id` (space before `#`) even though the doc says "repo attaches directly to '#'", and it rejects uppercase milestone letters (`M4B`) which the doc ("M + digits + optional letter") doesn't flag as lowercase-only. Both harmless; a one-line doc note would close the gap.

**5. Test coverage notes**
- Excellent coverage of the bug classes this diff could ship. One small gap: the end-to-end cross-repo path where `repoDir != root` isn't exercised in a unit test — `TestResolveRun_JSON` uses `ariadne#144` with root already at `<parent>/ariadne`, so `resolveRepoDir` returns the root itself. The dir-resolution (`TestResolveRepoDir`) and globbing (`TestFamilyFiles`) are covered in isolation, and the issue `## Log` records real-repo `parley#160` verification, so this is a note, not a gap worth blocking on.

**6. Architectural notes for upcoming work**
- Cross-repo resolution applies the **current** repo's `vocab.Issue().Discovery()` to sibling repos (`resolve.go:283` inside `resolveArtifacts`). This silently assumes every peer shares ariadne's `workshop/{issues,plans,history}` layout — true for ariadne-styled peers today. The day a peer customizes its `discovery:`, `resolve <peer>#id` would return not-found rather than reading the peer's own model. Worth a comment or a future "load the sibling's own discovery" step if peer layouts ever diverge.
- The #163 follow-up (migrate `push`/`merge`/`state`/`close`/`reviewsidecar` off their `workshop/plans`/`workshop/history` hardcoders onto `Discovery()`) is correctly scoped out here — those predate #144 and aren't consumers of the ref grammar. ARCH-PURPOSE shadow-sweep passes: the resolver itself fully derives from the model; the deferred migration is a separable DRY consolidation, not the deferred point of the issue. Keep the #163 cross-link live so the accessor doesn't grow orphan consumers.

**7. Plan revision recommendations** — none. The plan matches the shipped code (Core-concepts table entries all verified at their stated paths/kinds), and the existing `## Revisions` entry already documents the M1/M2→single-boundary collapse that this close reflects.
