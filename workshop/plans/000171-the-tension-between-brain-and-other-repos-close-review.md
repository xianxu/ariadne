# Boundary Review — ariadne#171 (whole-issue close)

| field | value |
|-------|-------|
| issue | 171 — the tension between brain and other repos |
| repo | ariadne |
| issue file | workshop/issues/000171-the-tension-between-brain-and-other-repos.md |
| boundary | whole-issue close |
| milestone | — |
| window | b45c686f49e62262fd67dfe9e7d194da0440de1a..HEAD |
| command | sdlc close --issue 171 |
| reviewer | claude |
| timestamp | 2026-07-17T23:05:38-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

All verification is done. Here is the boundary review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This whole-issue close delivers the full lift the Done-when promised, and I verified the end-state directly rather than trusting the Log: `brain/data/project/` holds only the active `metis-v2-experiment-algebra.md`; the four terminal records sit at their operator-confirmed destinations (`nous`/`kbench`/`metis` `workshop/history/projects/`); the close gate discovers fleet-wide with the pure `planPeerWrites` commit decision and snapshot-before-write ordering; parley's `gP`/`--kind project` wiring exists in the peer tree with unit specs pinning the argv; and the docs sweep (README, five helptext files, atlas, datatype, AGENTS.base + woven faces) is genuinely complete in this repo. One Important item keeps this off SHIP: the Done-when clause "downstream repos re-woven" is still unfinished — `42shots` and `pair` faces verifiably carry the pre-#171 brain line, and the only reminder lives in a Log line about to be archived. Caveat on verification: Bash is broken at the harness level in this session (EPERM creating `session-env`, reproduced unsandboxed and in a subagent), so I could not execute the suite; however, every milestone fix bundle recorded green `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/ -count=1` runs, and the windows after the last recorded green (M5/M6) contain only docs and data-op changes — the Go code at HEAD is the state last verified green. The main agent should still run the suite once before finalizing, per the established pattern.

### 1. Strengths

- **The pure/IO split is real, not claimed.** `planPeerWrites` (`cmd/sdlc/peerwrite.go:66`) is a genuinely pure decision core whose test is a zero-IO decision table covering all seven rows (commit, off-main, staged, brain, unknown-state, dirty-target, undeterminable-branch), and the state snapshot in `applyClose` (`cmd/sdlc/close.go` peer-write section) is correctly placed *before* the file writes with a comment explaining why — the "absorb another session's work" hazard is closed by construction, not by luck.
- **`containsIssueMarker` (`cmd/sdlc/internal/project/discover.go:143`) is the right fix at the right layer** — the ID-boundary bug was in the shared M2 match, so fixing it there protects both navigation and the close gate's tick, and `TestDiscoverByIssueRef_IDBoundary` + `TestContainsIssueMarker` pin both directions including the EOF-boundary edge.
- **The #139 TOCTOU snapshot generalized cleanly to N edits** (`closeReviewSnapshot.validate` loops `projects []projectEdit` carrying `oldText`) — the plan-then-apply invariant survived the single→many refactor intact, with `TestCloseCommand_ProjectChangedDuringBoundaryReview_DoesNotFinalize` updated to the fleet home.
- **Fail-safe direction is consistent everywhere:** unknown git state, unborn HEAD, brain, dirty targets → report-only with the exact finish command; `applyPeerWrites` failures warn-and-continue (pinned by the failing-runner stub); a peer-write outcome never fails the close. This matches the Done-when's "committed (scoped) or loudly reported" verbatim.
- **Honest bookkeeping under review pressure.** Six FIX-THEN-SHIP cycles all resolved in-bundle with plan `## Revisions` entries that match the shipped code (I spot-checked M3's delta 6 and M6's stale-binary correction against the code and fleet state — both accurate), plus a lessons.md entry generalizing the stale-binary evidence failure.

### 2. Critical findings

None.

### 3. Important findings

1. **Done-when "downstream repos re-woven" is incomplete: `42shots` and `pair` still carry the pre-#171 constitution** (`/Users/xianxu/workspace/42shots/AGENTS.md:20` and `pair/{AGENTS,CLAUDE,GEMINI}.md:20` all read "brain = special peer holding cross-cutting state (`project`, `roadmap`)" — the exact text this issue retires). M5's propagate-base skipped them as dirty (the designed refusal) and the Log carries "re-run at issue close" — this close is that moment. Fix: re-run `sdlc propagate-base` for the two dependents (out-of-sandbox, per M5's documented constraint) before or with this close; if their trees are still dirty with another session's work, record a durable deferral (a follow-up issue or a note somewhere that survives the issue's archival to `workshop/history/`), not just the soon-to-be-archived Log line. ARCH-PURPOSE: two consumers of the residency model still restate the old model.

### 4. Minor findings

- `cmd/sdlc/close.go:667-679` — the `closeResult` literal has mixed field alignment (`fm:` group vs the old wider `body:`/`repoName:` alignment); likely gofmt-unclean. Run `gofmt -w cmd/sdlc/close.go` (I could not execute gofmt to confirm).
- `ActiveOnly` drops terminal-status matches only for *legacy* (brain) records; a `done` record transiently sitting in a peer's active `workshop/projects/` would still be re-ticked. Window is tiny (`project close` archives immediately) — note only.
- brain's `git rm` of the four is verified gone from the working tree, but whether nous's auto-commit swept the staged deletion into a commit couldn't be checked without git — the M6 Log's carried item ("confirm nous swept") is still open; confirm at close.
- `isFleetSibling` excludes the worktree container by exact name `worktree` only; a `worktrees/` or similarly-named container wouldn't be skipped. Cold-path nit.
- `dropTerminalLegacy` re-reads files `scan` already read — already noted in the plan as the `ProjectMatch.Status` future extension; fine to leave.
- Test-fixture duplication: `gitIn`/`initFleetRepo` (another copy of the git-fixture idiom, flagged at M3 as ~the 12th) and near-identical `writeProject` helpers in `discover_test.go` vs `projectfind_test.go`. Pre-existing debt class; a shared internal test-fixture package remains a good side-quest candidate.

### 5. Test coverage notes

Coverage is strong and matches the risk profile: discovery (all-match, scope split, terminal-legacy drop, brain-predicate-not-basename, symlink dedup, unreadable skip, ID boundary, stale-sibling skip), the full peer-write decision table, real multi-repo git fixtures for commit/report-only/dirty-target/current-repo-omitted, close-level multi-project tick, TOCTOU, and the default-resolve regression pin. Known accepted gap: the milestone-close-mode peer commit is untested (same `applyClose` path; deferral documented in the M3 Log). The one structural gap in this session is execution: **the suite could not be run here** (harness EPERM, reproduced in a subagent) — the main agent must run `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/ -count=1` and record green in the close bundle, per the pattern every prior boundary followed.

### 6. Architectural notes (explicit ARCH sweep)

- **ARCH-DRY: pass.** One fleet walker (`SiblingRepoDirs`) shared by `resolveRepoDir` and discovery; one brain predicate (`gitx.IsBrainRepo`) adopted at all four sites; one discovery function feeding three consumers differing only in scope; one shared seam (`discoverProjectsForRef`) behind both navigation surfaces; shared `printProjectMatches`. The helptext restatements were swept at M5 and now agree with the flag registrations.
- **ARCH-PURE: pass.** `planPeerWrites` is pure with a mock-free decision table; discovery is an fs-seam tested against temp dirs (correctly reclassified from "PURE" in the plan's M2 Revisions — the Core Concepts table contradiction was resolved there, so no table/code mismatch remains); all git IO lives in `readRepoGitState`/`applyPeerWrites` behind the existing `gitRunner` seam.
- **ARCH-PURPOSE: pass, with the one flag above.** Shadow-sweep of the Done-when: residency ✓ (verified on disk), brain capture-only with the metis-v2 amendment ✓, fleet-wide close gate with safe peer-write ✓, parley navigation ✓ (verified in the peer tree), residency docs ✓ in this repo — but the two stale dependents (finding 3.1) are hand-maintained-stale consumers of the residency model until rewoven. Nothing that was the point of the issue got deferred as follow-up; #185 (roadmap lift) is a genuine separable extension.
- For upcoming work: the `resolve.go` sibling-discovery-model caveat (this repo's model applied to siblings) now spans issue *and* project discovery — worth remembering when any peer first customizes `discovery:`.

### 7. Plan revision recommendations

None required — the plan's six Revisions entries accurately reconcile every delta I checked against the code and the fleet. Optionally, when finding 3.1 is resolved, append one line to the M5 entry recording that the 42shots/pair re-weave completed at issue close (closing the loop the entry's delta 2 opened).
