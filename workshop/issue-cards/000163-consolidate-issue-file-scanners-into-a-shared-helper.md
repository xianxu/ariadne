---
id: '000163'
status: done
started: 2026-07-12T23:38:52-07:00
created: 2026-07-03
updated: 2026-07-14
estimate_hours: 2.06
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
