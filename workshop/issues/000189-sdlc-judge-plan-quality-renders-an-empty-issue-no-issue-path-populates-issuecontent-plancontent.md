---
id: 000189
status: open
deps: []
github_issue:
created: 2026-07-29
updated: 2026-07-29
estimate_hours:
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

## Spec

- `sdlc judge <category> --issue N` populates `IssueContent` from the issue file and
  `PlanContent` from the durable plan, the same way `change-code` does — one resolution
  shared between them rather than a second copy (`ARCH-DRY`).
- Categories that take no issue context are unaffected.
- **Decide what `sdlc judge plan-quality` means without a ledger.** `change-code` owns the
  gate state; `judge` holds none, so it cannot count rounds or dispose findings. Either it
  reads the sidecar read-only (so a manual invocation sees prior findings but records
  nothing), or it refuses `plan-quality` and points at `change-code`. Refusing may be the
  honest answer — a stateless plan-quality run is the pre-#187 behavior this repo just spent
  an issue removing.

## Done when

- `sdlc judge plan-quality --issue N` renders a prompt containing the issue's `## Spec`
  text and, when a durable plan exists, its body — pinned by a golden prompt so an empty
  render cannot pass again.
- The issue/plan resolution is SHARED with `change-code` rather than a second copy
  (`ARCH-DRY`): one of them changing cannot leave the other reading a different file.
- The no-ledger question is SETTLED in the code, not left implicit — either the sidecar is
  read read-only (prior findings shown, nothing recorded) or the category refuses and points
  at `change-code`. Whichever is chosen, a test asserts it, so "stateless plan-quality" can
  never quietly return.
- Categories that take no issue context render byte-identically to today.

## Plan

- [ ] Failing test: `sdlc judge plan-quality --issue N` renders a prompt containing the
      issue's `## Spec` text — currently empty
- [ ] Populate `IssueContent`/`PlanContent` from the shared resolution
- [ ] Settle the no-ledger question above and implement the chosen answer
- [ ] Golden-prompt coverage so an empty render cannot pass again

## Log

### 2026-07-29
