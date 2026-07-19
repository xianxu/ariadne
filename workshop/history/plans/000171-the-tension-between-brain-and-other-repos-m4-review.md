# Boundary Review — ariadne#171 (milestone M4)

| field | value |
|-------|-------|
| issue | 171 — the tension between brain and other repos |
| repo | ariadne |
| issue file | workshop/issues/000171-the-tension-between-brain-and-other-repos.md |
| boundary | milestone M4 |
| milestone | M4 |
| window | 2d898dfa05e2db7ec4b97e00f3316a18a663fc75..HEAD |
| command | sdlc milestone-close --issue 171 --milestone M4 |
| reviewer | claude |
| timestamp | 2026-07-17T21:19:43-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M4 delivers exactly what the boundary claims: one shared discovery seam (`discoverProjectsForRef`) feeding both `sdlc project find` and `resolve --kind project` under `ActiveAndArchive` scope, the default issue-kind resolution pinned unchanged by a dedicated regression test, and the parley `gP` jump verified present in the peer tree (`parley.nvim/lua/parley/init.lua:4201`, `artifact_ref.lua:122`, unit specs pinning the `--kind` argv). I cross-checked the marker convention against the M2 close gate — both build the marker from `filepath.Base(repoDir)` + `strconv.Itoa(id)` (`close.go:420,351` vs `projectfind.go:55-56`), so prefix tokens like `met#18` normalize correctly before matching. Two Important findings keep this off SHIP: the README's `sdlc project` subcommand listing was not updated for `find` (the exact #142-class gap this gate exists to catch), and the shared marker match has an ID-boundary false positive that M4 turns into user-facing surface. Caveat: I could not execute the test suite — Bash is unavailable in this session (persistent harness EPERM creating its session-env dir, before any command runs) — so this review is static; the main agent should run `go test ./cmd/sdlc/... -count=1` when applying fixes.

### 1. Strengths

- **The shared seam is genuine ARCH-DRY, not lip service.** `discoverProjectsForRef` (`cmd/sdlc/projectfind.go:43`) owns parse → sibling-resolve → discovery for both surfaces; `runResolveProjects` (`cmd/sdlc/resolve.go:425`) is a thin formatter over it. The plan's sketch implied two parallel wirings; consolidating was the right delta.
- **`TestResolveRun_DefaultKindUnchangedByProjects`** (`projectfind_test.go:134`) is precisely the regression that matters — it plants a project record next to an issue family and proves default resolve doesn't leak it. This pins the "default resolution unchanged" contract with evidence, not assertion.
- **Repo-token normalization before marker construction** — resolving through `resolveRepoDir` and then taking `filepath.Base(repoDir)` means `find` and the close gate agree on the full-basename convention (`[parley.nvim#84`, not the typed short token). The sketch's verbatim `ref.Repo` would have been a silent-miss bug; the implementor caught it.
- **Docs updated in-window:** helptext for both commands, a new atlas section (`atlas/workflow/sdlc-binary.md`), and an honest plan `## Revisions` entry whose three claimed deltas all check out against the code (including the peer repo).
- **`--kind` as a resolution mode, not new ref grammar** — keeps `parseRef` the single-sourced grammar authority, with unknown kinds erroring loudly (`resolve.go:392`) and a test pinning that.

### 2. Critical findings

None.

### 3. Important findings

1. **README update missing for the new CLI surface** (`README.md:21-33`). The README enumerates every `sdlc project` subcommand (new/list/show/validate/set-status/status/retro/close) but the diff adds `find` without touching it, and `resolve --kind project` appears nowhere. This is exactly the class of gap the docs gate names (#142). Fix: add `sdlc project find --issue metis#18` to the command listing and a line for `sdlc resolve --kind project`.
2. **ID-boundary false positive in the shared marker match** — `discover.go:78,99` matches via `strings.Contains(data, "["+repo+"#"+id)`, so searching `#18` matches a record containing `[metis#180]`, and `#17` matches `[ariadne#171]`. The code is M2 (outside this window), but M4 promotes it to user-facing navigation, and the close gate shares it (a close of issue #18 would falsely tick #180's project). Cheap fix in `DiscoverByIssueRef`: accept only the two documented forms — `marker+"]"` or `marker+" "` (i.e., a non-digit boundary after the id) — plus a test seeding `[metis#180]` and asserting `find --issue metis#18` misses it. Since the fix lands in M2 code, also worth a one-line plan note.

### 4. Minor findings

- `sdlc resolve --kind project '#171 M4'` silently ignores the milestone token; either error ("milestones don't apply to project kind") or document the ignore in `helptext/resolve.md`.
- Text-mode `path + " (legacy)"` print loop is duplicated between `projectfind.go:75-81` and `resolve.go:437-443` (ARCH-DRY nit) — extract a shared `printProjectMatches(w, matches)`.
- JSON rows drop the `Legacy` bit (text mode flags it; a JSON consumer can't distinguish a brain-legacy record). Documented in atlas so evidently deliberate, but a `legacy` field is cheap and parley may eventually want it.
- `helptext/resolve.md` "(`[repo#id`)" reads as an unbalanced-bracket typo unless the reader already knows the open-bracket marker convention — spell it out ("records containing the `[repo#id …]` marker").

### 5. Test coverage notes

Strong for the new surface: the fleet fixture covers active + archived + brain-legacy in one sweep, legacy flagging both directions, prefix-token resolution, the distinct no-match error, JSON shape (kind on every row, ID), unknown-kind error, and the default-kind pin. Gaps: (a) the ID-boundary collision above — the test fixture uses only `#18` with no `#180`-style neighbor, which is why the bug is invisible; (b) no test for `--kind project` with a GitHub ref (the `projectfind.go:48` error arm is untested); (c) milestone-token-with-project-kind behavior is unpinned. I could not run the suite myself (Bash broken in this session); the tests read as correct and self-contained (`t.TempDir` fleets, injected roots, no exec/net — consistent with the repo's INTEGRATION-via-temp-repo idiom).

### 6. Architectural notes (explicit ARCH pass)

- **ARCH-DRY: pass** (with the minor print-loop nit). One walker (`SiblingRepoDirs`), one discovery (`DiscoverByIssueRef`), one seam for two command surfaces; helptext points `find` and `resolve --kind project` at each other as equivalents.
- **ARCH-PURE: pass.** No new business logic buried in IO — the new code is thin composition of the existing pure core (`parseRef`) and existing IO seams; tests run against real temp filesystems with injected roots, zero mocks.
- **ARCH-PURPOSE: pass.** Shadow-sweep of the M4 Done-when slice: `sdlc project find` ✓, `resolve` project kind archive-inclusive ✓, parley always-cross-repo project navigation ✓ (verified in the peer tree, not taken from the commit message). The "artifact class" → "jump binding" reframe is a legitimate adaptation to parley's actual structure and is documented in the plan's Revisions rather than silently narrowed; fleet-wide *search* (vs jump-from-ref) is served by the CLI. Nothing that is the point of M4 was deferred as follow-up.
- For M5: `resolveArtifacts`'s existing note (`resolve.go:312`) about applying this repo's discovery model to siblings now also applies to project discovery — if a peer ever customizes `discovery:`, both paths need the sibling's model. Worth one line in the M5 docs sweep.

### 7. Plan revision recommendations

None required — the plan's 2026-07-17 M4 Revisions entry already reconciles all three deltas and matches the shipped code. Two honest open boxes remain by design: Task 4.3 Step 3 (live in-editor `gP` jump, deliberately deferred to Manual Verification — the operator should actually perform it before M5) and Step 5 (this milestone-close). If the ID-boundary fix (Important 2) is taken now, add a one-line Revisions note since it amends M2's discovery contract.
