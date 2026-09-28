---
id: '000143'
status: done
started: 2026-06-29T17:31:30-07:00
created: 2026-06-29
updated: 2026-06-29
estimate_hours: 0.59
actual_hours: 0.65
---

# Archive all issue artifacts to history on completion

## Problem

On issue completion, the archive sweep in `sdlc merge` (`archiveDoneIssuesInDir`,
`cmd/sdlc/merge.go:548`) and `sdlc push` (`archiveDoneIssues`,
`cmd/sdlc/push.go:468`) moves only `workshop/issues/NNNNNN-*.md` to
`workshop/history/`. Artifacts that share the issue's 6-digit id prefix — the
durable plan (`workshop/plans/NNNNNN-slug-plan.md`) and boundary-review sidecars
(`workshop/plans/NNNNNN-slug-{close,m<x>}-review.md`, #136) — are left behind in
`workshop/plans/` and accumulate indefinitely.

Observed at the #136 close: the plan + the close-review sidecar stayed in
`workshop/plans/` while the issue moved to history. The atlas already *claims*
plans are "Archived with issue" (`atlas/workflow/artifact-hierarchy.md`,
`atlas/workflow/sdlc-binary.md`) — so today's behavior contradicts the documented
lifecycle.
