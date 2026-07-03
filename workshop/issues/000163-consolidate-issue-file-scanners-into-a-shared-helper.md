---
id: 000163
status: open
deps: []
github_issue:
created: 2026-07-03
updated: 2026-07-03
estimate_hours:
---

# consolidate issue-file scanners into a shared helper

## Problem

Four `cmd/sdlc` helpers converged on the same shape — **enumerate issue files
(glob or `git diff` window) → `issue.Parse` → read `status` → filter/act** — after
#160 added the third and fourth. The #160 M3 and whole-issue boundary reviews both
flagged the duplication (ARCH-DRY), noting the comments even say *"mirrors … (ARCH-DRY)"*
but mirror rather than reuse:

- `mergedCodecompleteIssues(issuesDir)` — `cmd/sdlc/publishgate.go`: `git diff
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
// changedIssueFiles returns the issue files (repo-relative path + parsed status)
// in a window (baseRef..HEAD) OR — when baseRef == "" — the whole issuesDir glob.
type issueFileRef struct { Path, Status string }
func changedIssueFiles(baseRef, issuesDir string, r gitRunner) ([]issueFileRef, error)
```

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
  unit-testable without git.
- The `gitRunner` seam matters for `touchedIssuesNotDone`'s existing tests; the
  publishgate helpers currently use `gitx.RunGit` directly (cwd). Reconcile — either
  thread `gitRunner` through, or standardize on `gitx` — without regressing the
  merge/push test seams (`runPublishGateFn`, the e2e stubs).
- This is base-layer `cmd/sdlc` code — no behavior change, pure refactor.

## Done when

- [ ] A single `changedIssueFiles`-style helper backs all four scanners; no caller
      re-implements the glob/diff + parse + status-read boilerplate.
- [ ] Behavior is unchanged (the `codecomplete` carve-out, terminal filters, and
      window vs dir-wide scoping all preserved) — existing tests pass untouched where
      they assert behavior.
- [ ] The pure status-filters are unit-tested; the git/IO seam is exercised against a
      real temp repo (the `publishgate_test.go` / `milestonewindow_test.go` pattern).

## Plan

- [ ] Inspect the four scanners; identify the shared parse core vs the per-caller filter.
- [ ] Extract `changedIssueFiles` (window + dir-wide) + `issueFileRef`; reconcile the
      `gitRunner` vs `gitx` seam.
- [ ] Rewrite the four callers as filters over it; keep their signatures/behavior.
- [ ] Tests: pure filters + temp-repo seam; confirm the existing merge/push/publishgate
      suites stay green.

## Log

### 2026-07-03

- Created as a follow-up from #160 (the codecomplete two-gate model), which added the
  third + fourth scanner. Flagged by #160's M3 and whole-issue boundary reviews as an
  ARCH-DRY consolidation to do "before a fifth appears." Pure refactor, no behavior
  change.
