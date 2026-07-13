---
id: 000163
status: codecomplete
deps: []
github_issue:
created: 2026-07-03
updated: 2026-07-13
estimate_hours: 2.06
started: 2026-07-12T23:38:52-07:00
actual_hours: 4.98
---

# consolidate issue-file scanners into a shared helper

## Problem

Four `cmd/sdlc` helpers converged on the same shape — **enumerate issue files
(glob or `git diff` window) → `issue.Parse` → read `status` → filter/act** — after
#160 added the third and fourth. The #160 M3 and whole-issue boundary reviews both
flagged the duplication (ARCH-DRY), noting the comments even say *"mirrors … (ARCH-DRY)"*
but mirror rather than reuse:

- `mergedCodecompleteIssues(baseRef, issuesDir)` — `cmd/sdlc/publishgate.go`: `git diff
  --name-only baseRef..HEAD -- issuesDir/*.md` → parse → keep `status == codecomplete`.
- `touchedIssuesNotDone(baseRef, issuesDir, r)` — `cmd/sdlc/push.go`: same window diff
  → parse → keep non-terminal (and, post-#160, not `codecomplete`).
- `publishCodecompleteIssues(issuesDir)` — `cmd/sdlc/publishgate.go`: glob
  `NNNNNN-*.md` → parse → flip `codecomplete → done`.
- `archiveDoneIssues` / `archiveDoneIssuesInDir` — `cmd/sdlc/push.go` / `merge.go`:
  glob → parse → act on terminal.

Each re-derives the glob/diff + parse + `GetField("status")` boilerplate. A fifth
scanner is likely (this pattern recurs), and the divergence is a real hazard — e.g.
the `codecomplete` carve-out had to be added to `touchedIssuesNotDone` by hand (#160
review #2) and could drift from the others.

## Spec

Extract one shared helper that both the window-scoped and dir-wide callers use, e.g.:

```go
// scanIssueFiles returns parsed issue files
// in a window (baseRef..HEAD) OR — when baseRef == "" — the whole issuesDir glob.
type issueFileRef struct { Path, Status, Frontmatter, Body string }
func scanIssueFiles(baseRef, issuesDir string, runGit func(...string) ([]byte, error)) ([]issueFileRef, error)
```

The helper name must not collide with claim's existing, behaviorally different
`changedIssueFiles(*claimFlags, gitRunner)`, which enumerates dirty/staged/untracked
issue records for tracker synchronization. Retaining the parsed frontmatter and body
in `issueFileRef` is deliberate: publish needs them to compose the status update and
archive needs frontmatter for `github_issue`. Returning only path/status would make
those callers immediately re-read and re-parse the same file, leaving the duplication
half-consolidated (ARCH-DRY).

Then the four callers become status-filters over its result:
- `mergedCodecompleteIssues` → `filter(status == "codecomplete")` on the window.
- `touchedIssuesNotDone` → `filter(!IsTerminal && != "codecomplete")` on the window.
- `publishCodecompleteIssues` → `filter(status == "codecomplete")` on the dir glob, then flip.
- `archiveDoneIssues` → `filter(IsTerminal)` on the dir glob, then archive.

Design notes / constraints:
- Preserve the **window vs dir-wide** distinction (some callers scan `baseRef..HEAD`,
  others glob the whole dir) — the helper should support both (baseRef sentinel, or
  two entry points sharing a parse core).
- Keep it a thin git/IO seam feeding pure status-filters (ARCH-PURE); the filters are
  unit-testable without git. Keep the filter/action boundary explicit: the shared
  helper enumerates and parses; GitHub closes, writes, renames, plan sweeps, and
  logging remain in the callers.
- The `gitRunner` seam matters for `touchedIssuesNotDone`'s existing tests; the
  publishgate helpers currently use `gitx.RunGit` directly (cwd). Reconcile — either
  thread `gitRunner` through, or standardize on `gitx` — without regressing the
  merge/push test seams (`runPublishGateFn`, the e2e stubs).
- Preserve the two window callers' distinct diagnostics: `mergedCodecompleteIssues`
  wraps the underlying `gitx.RunGit` error with `%w`, while `touchedIssuesNotDone`
  includes `gitRunner.Git`'s combined output. The shared scanner accepts a narrow git
  function and returns a typed error carrying raw output plus the underlying error so
  each caller retains its current contract.
- Preserve current edge semantics: a failed window `git diff` returns an error;
  unreadable or malformed issue files are skipped; a missing status is still reported
  as `unset` by the not-done warning; dir-wide glob results stay sorted while window
  results retain git's order. Window enumeration preserves the existing
  `issuesDir/*.md` git pathspec; only dir-wide enumeration applies the six-digit
  `NNNNNN-*.md` filename restriction.
- Reuse the existing issue-filename grammar everywhere: one `issueFilenamePattern`
  feeds directory globbing (including `buildPushCommitMessage`) and membership; a
  small pure parts helper replaces `state.go`'s parallel capture regex while preserving
  its non-empty-slug rule, supplies `issueIDPrefix` for archive-plan lookup, and
  replaces branch creation's equivalent `issueIDRE` full-filename check. Do not
  introduce another six-digit literal, full-filename regex, or manual digit loop while
  removing scanner duplication (ARCH-DRY). The scaffold's prefix-only parser remains
  distinct because it does not validate a full filename.
- Preserve merge's path topology: a dir-wide scan under `mainPath` may return absolute
  filesystem paths, while `archiveDoneIssuesInDir` must continue recording
  `mainPath`-relative paths for `GitInDir` staging.
- This is base-layer `cmd/sdlc` code — no behavior change, pure refactor.

## Done when

- [x] The shared `scanIssueFiles` helper backs all four scanners; no caller
      re-implements the glob/diff + parse + status-read boilerplate.
- [x] The six-digit issue filename pattern has one definition shared by directory
      scanning, `buildPushCommitMessage`, `issueFilename`, `issueIDPrefix`, and state
      inventory/untracked-branch parsing.
- [x] Behavior is unchanged (the `codecomplete` carve-out, terminal filters, and
      window vs dir-wide scoping all preserved) — existing tests pass untouched where
      they assert behavior.
- [x] The pure status-filters are unit-tested across terminal, `codecomplete`, active,
      and missing statuses; the git/IO seam is exercised against a real temp repo,
      including malformed/unreadable/deleted records, the six-digit dir-wide glob,
      ordering, and a non-six-digit `.md` included by the window scan but excluded by
      the dir-wide scan.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.15 impl=0.10
item: smaller-go-module design=0.10 impl=0.20
item: smaller-go-module design=0.05 impl=0.20
item: cross-cutting-refactor design=0.20 impl=0.20
item: cross-cutting-refactor design=0.20 impl=0.20
item: atlas-docs design=0.05 impl=0.10
item: milestone-review design=0.00 impl=0.20
design-buffer: 0.15
total: 2.06
```

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only. The thorough reviewed spec earns the v2.1 design
discount and 15% design buffer; v3.1 implementation values use 40% of the v2 ranges.
The second module/refactor primitives cover the additional filename/state consumers
and their focused tests. The calibration source is currently marked stale, so this
estimate is provisional.

## Plan

Durable execution plan:
`workshop/plans/000163-consolidate-issue-file-scanners-into-a-shared-helper-plan.md`.

- [x] Inspect the four scanners; identify the shared parse core vs the per-caller filter.
- [x] Extract `scanIssueFiles` (window + dir-wide) + `issueFileRef`; reconcile the
      `gitRunner` vs `gitx` seam.
- [x] Rewrite the four callers as filters over it; keep their signatures/behavior.
- [x] Tests: pure filters + temp-repo seam; confirm the existing merge/push/publishgate
      suites stay green.

## Log

### 2026-07-03

- Created as a follow-up from #160 (the codecomplete two-gate model), which added the
  third + fourth scanner. Flagged by #160's M3 and whole-issue boundary reviews as an
  ARCH-DRY consolidation to do "before a fifth appears." Pure refactor, no behavior
  change.

### 2026-07-12

- Claimed and entered planning. Traced the push/merge publish and archive flows plus
  their real-repo and injected-runner test seams. Design approved: one window/dir scan
  helper returns a complete parsed record, with pure status filters and caller-owned
  side effects (ARCH-DRY, ARCH-PURE, ARCH-PURPOSE).

### 2026-07-13
- 2026-07-13: closed — Boundary remediation added AST source guard for the sole filename literal and all named consumers; focused guard/scanner tests passed; go test ./cmd/sdlc -count=1 and go test ./... -count=1 passed; committed window whitespace clean; no atlas surface change; unrelated pre-existing worktree edits excluded.; review verdict: SHIP
- 2026-07-13: closed — Focused scanner/caller regressions passed; go test ./cmd/sdlc -count=1 and go test ./... -count=1 passed; committed window git diff --check clean; ARCH-DRY sweeps show one filename pattern and no legacy full-filename regex/manual digit loop; remaining parsers are distinct historical-anchor/archive-recovery jobs; no atlas surface change; unrelated pre-existing worktree whitespace in process-manual.md and #170 excluded.; review verdict: FIX-THEN-SHIP
- The boundary review found no correctness defects and confirmed the consolidation,
  error contracts, side-effect topology, and ARCH-PURE separation. Its one Important
  finding was an enforcement gap: the plan promised an automated source guard, while
  execution had only run the equivalent manual sweep. Added an AST-based regression
  that requires one production pattern literal and verifies every named consumer
  references its canonical constant/helper (ARCH-DRY, ARCH-PURPOSE).

- Implemented the parsed scanner, typed git failure, pure status filters, and one
  shared six-digit filename grammar. Rewired both window callers and all three
  directory action paths while retaining caller-owned diagnostics and side effects.
- TDD evidence: pure-filter and scanner tests were observed RED before implementation,
  then GREEN; window and directory caller characterization passed before and after
  rewiring. Added real-git, injected-order, malformed/unreadable/deleted, missing-
  status, error-chain, mutation-body, GitHub-close, and relative-path coverage.
- Verification: focused scanner/caller regressions passed; `go test ./cmd/sdlc
  -count=1` passed; `go test ./... -count=1` passed; the committed branch window is
  whitespace-clean. ARCH-DRY sweeps found one production filename-pattern definition
  and no legacy full-filename regex/manual digit loop. The two remaining status
  parsers are behaviorally distinct: historical close-anchor inspection and
  interrupted-archive validation.
- Atlas assessment: no live atlas map names or describes these internal scanner
  implementations, so the refactor introduces no atlas surface change (ARCH-PURPOSE).
  Repo-wide `git diff --check` still reports only pre-existing user edits in
  `atlas/workflow/process-manual.md` and issue #170; #163's committed window is clean.

## Revisions

### 2026-07-12T23:50:00-07:00 — approved design after source-grounded context pass

- Replaced the illustrative helper name because `changedIssueFiles` already names the
  claim-sync scanner; selected `scanIssueFiles` for this distinct status scanner.
- Expanded `issueFileRef` to retain parsed frontmatter/body so publish and archive do
  not reparse.
- Pinned existing error, malformed-file, ordering, missing-status, and merge-relative-
  path behavior as explicit no-change constraints and test obligations.

### 2026-07-13T00:02:00-07:00 — fresh-context spec review

- Corrected the stale `changedIssueFiles` name in Done-when and Plan so every section
  consistently names `scanIssueFiles` and cannot be read as merging with claim sync.
- Made the enumeration grammar testable: window scope keeps `issuesDir/*.md`, while
  dir-wide scope alone requires the six-digit issue filename convention.

### 2026-07-13T00:15:00-07:00 — implementation plan and derived estimate

- Added the durable TDD plan and a reconciled estimate-logic-v3.1 breakdown totaling
  1.05 ship-wall-clock hours. Kept the refactor atomic with one close-time review
  boundary; no artificial milestone tags.

### 2026-07-13T00:27:00-07:00 — fresh-context plan review

- Corrected the Problem's stale `mergedCodecompleteIssues` signature.
- Narrowed scanner injection from the broad `gitRunner` interface to a git function,
  preserving `gitx.RunGit` for the publish gate and `r.Git` for warning callers.
- Made raw git output and error unwrapping part of the shared scan-error contract so
  consolidation cannot silently change caller diagnostics.

### 2026-07-13T00:47:00-07:00 — change-code plan-quality refusal

- The gate found that the planned directory glob would duplicate `issueFilename`'s
  existing six-digit grammar. Revised the design so one `issueFilenamePattern`
  constant feeds both glob enumeration and filename membership (ARCH-DRY).

### 2026-07-13T00:55:00-07:00 — second change-code plan-quality refusal

- Expanded the filename single source to `buildPushCommitMessage` and state inventory,
  replacing the latter's equivalent capture regex with a pure parts helper while
  retaining its non-empty-slug behavior.
- Added a fake-runner test whose deliberately non-lexicographic output proves window
  order is not sorted; a real git repo alone cannot expose that mutation.
- Re-derived the estimate as 2.06h for the expanded consumer/test surface; the prior
  1.05h no longer matched the executable plan.

### 2026-07-13T01:02:00-07:00 — durable-plan discovery correction

- `change-code` reviews `<issue-filename-stem>-plan.md` exactly. Renamed the shortened
  plan slug to match the issue stem so the gate receives the detailed executable plan
  instead of reviewing only the issue's abbreviated checklist.

### 2026-07-13T01:15:00-07:00 — filename shadow-consumer sweep

- The detailed gate review found `issueIDPrefix` manually revalidating the same six
  digits and hyphen for archive-plan lookup. Added it to `issueFilenameParts` and the
  equivalence tests/sweep so no manual digit loop survives the single source.

### 2026-07-13T01:22:00-07:00 — full-filename regex consumer sweep

- Added branch creation's equivalent `issueIDRE`/`listUntrackedIssues` consumer to the
  shared filename predicate and tests. Kept the internal scaffold's prefix-only parser
  separate because it intentionally does not validate a complete filename.
