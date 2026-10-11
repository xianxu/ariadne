---
id: 000189
status: codecomplete
created: 2026-07-29
updated: 2026-10-10
estimate_hours:
github_issue:
started: 2026-10-10T16:53:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.77
tracker:
    version: 1
    completion:
        token: close-37bf4cac8d74
        repository: github.com/xianxu/ariadne
        reviewed_head: f4fd6f8e55fd687d6d6dced68fcd2e5b11d38e9f
        evidence_commit: 96d7258f45ade26378f95d17647f296e0062dca4
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
