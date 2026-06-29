# Boundary Review — ariadne#136 (whole-issue close)

| field | value |
|-------|-------|
| issue | 136 — sdlc boundary review sidecar |
| repo | ariadne |
| issue file | workshop/issues/000136-review-sidecar.md |
| boundary | whole-issue close |
| milestone | — |
| window | fa432305957410402314e06c3a99d7b9de47171f..HEAD |
| command | sdlc close --issue 136 |
| reviewer | claude |
| timestamp | 2026-06-29T16:29:01-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Confirmed: `--agent` defaults to `""`, and the resolved reviewer (via `judge.ResolveAgentCLI` → `opts.Agent`/`agent` at `milestoneclose.go:494`) is **not** threaded into the sidecar — `writeReviewSidecar` records the raw `p.Agent`. So in the common default invocation the `| reviewer |` cell renders empty despite the value being known. I have everything I need.

VERDICT: FIX-THEN-SHIP (confidence: high)

The sidecar feature is cleanly built and well-tested: a genuinely pure render/path/meta core behind a thin atomic-write IO seam, wired into the single shared `dispatchBoundaryReview` so both close paths inherit persistence for free, and strictly additive (trailers/verdict/log/gates untouched). All 16 diff-relevant tests pass, including the no-judge/dry-run skip paths and graceful non-fatal degradation when params are absent. Nothing blocks the boundary. The findings below are non-blocking fidelity/coverage gaps worth closing before the next boundary — chiefly that the spec-required "reviewer agent" metadata renders empty in the default invocation even though the resolved agent is in hand at the call site.

*(Process note, not a code finding: the full `go test ./cmd/sdlc/` run hangs at 300–600s because a live `sdlc close --issue 136` process — PID 94761, started 16:04 — is still holding the real-repo lock `.git/sdlc.lock` and blocks `TestSetStatusAlias_BothPathsMutate`. This is unrelated to the diff; the diff-relevant tests use temp repos / call dispatch directly. The main agent should clear that stale/hung process+lock before relying on a full-suite run. I left it untouched — read-only review.)*

### 1. Strengths
- **Exemplary ARCH-PURE separation** (`reviewsidecar.go:18-93`): `sidecarMeta`/`sidecarPath`/`renderReviewEntry`/`boundaryKind`/`sidecarCommand` are all deterministic and unit-tested with zero IO; the clock is isolated in `nowRFC3339` and injected as a string. This is the design the plan promised.
- **Single write site (ARCH-DRY)** (`milestoneclose.go:519-527`): persistence lives only in the shared `dispatchBoundaryReview`; both close and milestone-close inherit it. Slug derived from the resolved issue-filename stem (`reviewsidecar.go:38`) reuses one source of truth instead of re-slugifying the title.
- **Non-fatal write is correct** (`milestoneclose.go:522-527`): a write failure is warned, not propagated — the review already ran. Confirmed robust by `TestDispatchBoundaryReview_AgentDefaultUsesPairAgent` passing even though it supplies no `IssueNum`/`PlansDir` (sidecar write errors → `cwarn`, dispatch still returns the verdict).
- **D4 honored**: `--no-judge`/`--dry-run`/not-run never reach `dispatchBoundaryReview` (`close.go:740-751`, `milestoneclose.go:148-153`), so no stub sidecars — verified by reading both call paths.
- **Re-run preserves evidence** (`reviewsidecar.go:152-155` + `TestWriteReviewSidecar_CreateThenAppend`): read-modify-rewrite-via-rename keeps exactly one H1 and both bodies; no temp-file leak asserted.

### 2. Critical findings
None.

### 3. Important findings
- **Reviewer metadata renders empty in the default case (ARCH-PURPOSE)** — `reviewsidecar.go:147` sets `Agent: p.Agent`, but `p.Agent` is the raw `--agent` flag which defaults to `""` (`close.go:148`, `milestoneclose.go:97`). The *resolved* reviewer is computed at dispatch and sits in scope as `agent := opts.Agent` (`milestoneclose.go:494`). So a normal `sdlc close` (no `--agent`) writes `| reviewer |  |` — empty — even though the spec explicitly requires "reviewer agent/model if known" and it *is* known. Fix sketch: thread the resolved agent into the write, e.g. in `dispatchBoundaryReview` set `p.Agent = agent` before calling `writeReviewSidecar`, or pass `opts.Agent` as an explicit arg used for `sidecarMeta.Agent`. Cheap, and it's the durable record's whole point.
- **D4 (skip-on-non-run) is documented but untested** — `TestRunCloseWithReview_NoJudge_Skips` and the dry-run test assert behavior but never assert the *absence* of a sidecar file. A future refactor could start emitting a stub for skipped boundaries with no test catching it. Fix sketch: add `if _, err := os.Stat(filepath.Join("workshop/plans", "000069-x-close-review.md")); !os.IsNotExist(err) { t.Error(...) }` to the no-judge/dry-run cases.

### 4. Minor findings
- `repoIdentity()` (`reviewsidecar.go:98-104`) is a third inline copy of `filepath.Base(gitx.RepoTopLevel())` — also at `close.go:367-371` and `branchcreate.go:145-149` (ARCH-DRY). Consider promoting a `gitx.RepoName()` and routing all three through it. Pre-existing duplication the new code joins rather than resolves; non-blocking.
- Window cell records `<longbase>..HEAD` with the literal `"HEAD"` (`milestoneclose.go` threads `Head: "HEAD"`; rendered at `reviewsidecar.go:83`). For a *durable* artifact, a symbolic ref is weaker than a resolved short SHA (it's only recoverable via the sidecar's own commit provenance). Matches the existing trailer convention, so low priority — but resolving head to a SHA before the write would make the record self-contained.
- Title fallback is doubly-defaulted (`reviewsidecar.go:133` `"(no title)"` and `issueTitleFromContent` also returns `"(no title)"`); harmless, just redundant.

### 5. Test coverage notes
- Pure path/render/write coverage is strong (close + 2 milestone path cases + absolute-basename; metadata completeness + revision heading + no-H1-on-revision; create→append with both bodies + single H1 + no temp leak; missing-issue error). Integration covers both close (`TestRunCloseWithReview_IssueClose_Dispatches`) and milestone (`TestDispatchBoundaryReview_WritesMilestoneSidecar`) through the real dispatch.
- Gaps: (a) no test asserts the *resolved* reviewer reaches the sidecar — which is exactly why the empty-reviewer bug shipped silently (the close integration test asserts `| verdict | SHIP |` but not the reviewer cell); (b) no test pins D4's no-write guarantee.

### 6. Architectural notes for upcoming work
- The sidecar is written to the working tree *after* the close commit (the review is a follow-on), so it lands **untracked**. The issue's own Non-goals flags that the merge-sweep to `workshop/history/` relies on the file being carried along — which only holds if the agent commits it. Worth an explicit "commit the sidecar" step in the close runbook/atlas so durable records aren't orphaned as untracked files. (This review *is* that close; its `000136-review-sidecar-close-review.md` will need committing.)
- `sidecarMeta` is a clean extension point (the plan notes `DiffStat`/`PromptSHA` could be added without touching the writer) — good. When the reviewer fix lands, prefer storing the resolved agent so historical records stay comparable.

### 7. Plan revision recommendations
None required for table-vs-code fidelity: every Core-concepts entity exists at its stated path with the stated kind (PURE entities tested without IO; INTEGRATION via temp dirs), and modified entities (`reviewResult`, `boundaryReviewParams`, `dispatchBoundaryReview`) show the expected changes. Optional: if the reviewer-agent fix is adopted, add a one-line `## Revisions` note to the plan recording that `sidecarMeta.Agent` is populated from the *resolved* dispatch agent (`opts.Agent`), not the raw `--agent` flag, so the metadata contract is unambiguous.
