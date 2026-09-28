---
id: 000189
status: open
created: 2026-07-29
updated: 2026-07-29
estimate_hours:
github_issue:
---

# sdlc judge plan-quality renders an empty issue — no --issue path populates IssueContent/PlanContent

## Problem

`sdlc judge plan-quality --issue N` reviews an EMPTY issue. `cmd/sdlc/judge.go:109-113`
builds `judge.PromptInput` with `Diff`, `ChangedIssues`, `Base` and `Head` only —
`IssueContent` and `PlanContent` are never populated **for any category**. So
`{{ISSUE_CONTENT}}` renders empty and `{{PLAN_CONTENT}}` renders
`(no separate plan file)`, and the judge is asked to assess the quality of a plan it
cannot see. It will answer anyway.

The verb is advertised in the root helptext, so this is a defect in a documented surface,
not an internal edge. It silently produces a confident review of nothing — the worst
failure shape for a judge.

Found while designing ariadne#187 Task 14, which needed exactly this path and had to drive
`runPlanQualityJudge` directly instead. Deliberately left unfiled during #187 to keep it
out of that issue's scope (recorded in its plan's Task 14 rationale).
