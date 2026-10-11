---
id: 000189
status: working
deps: []
github_issue:
created: 2026-07-29
updated: 2026-10-10
estimate_hours:
card_mirror: '15d773045dcacbf9279177c1c80aacb64aafe1d4' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T16:53:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "6d3c6f44", done: "e5bf0e1c"}
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

### Decision (2026-10-10)

- **plan-quality: render read-only, refuse to dispatch.** `sdlc judge plan-quality --issue N
  --dry-run` renders exactly the prompt `change-code` would send — issue, durable plan, and
  the ledger's prior findings, read and never written. A live run refuses and points at
  `sdlc change-code`: the gate's decision is the ledger's, and a dispatch nobody records is
  the stateless pre-#187 review this repo removed.
- Refuses without `--issue N` (it judges one issue; an empty render is the bug).
- estimate-quality was considered and is out: it renders `{{ISSUE_CONTENT}}` too, but it is not
  a `sdlc judge` category (`judge.IsValid` rejects it), so it has no manual surface to fix.
- Shared resolution (`ARCH-DRY`): `resolveChangeCodeName` (issue file by id), one
  `planArtifactPath` (also behind `readOptionalPlanFile`), one `planningIssueRef`, and one
  `planQualityPromptInput` that both verbs build the prompt through. The other categories' path
  is untouched (the `internal/judge` golden and the `dry` dry-run tests still pin it). `judge` reads the issue's
  disk bytes; `change-code` reads them with the card mirror refreshed (card fields only).
- ARCH-FUNERAL: creates nothing durable — the judge path reads the ledger and writes nothing.

## Plan

- [x] Failing test: `sdlc judge plan-quality --issue N --dry-run` renders a prompt containing
      the issue's `## Spec` text and the plan body — currently empty
      (`TestJudgePlanQuality_DryRunRendersChangeCodePrompt`, red before the fix)
- [x] Extract the shared resolution + prompt input; `change-code` and `judge` both use it
- [x] plan-quality live run refuses → `change-code`, no dispatch
      (`TestJudgePlanQuality_LiveRunRefuses`); prior findings shown read-only on dry-run and
      the ledger bytes unchanged (asserted in the dry-run test)
- [x] Refuses without `--issue` (`TestJudgePlanQuality_RefusesWithoutIssue`)
- [x] Pin: judge's plan-quality prompt is byte-equal to `change-code --dry-run`'s for the same
      fixture (same dry-run test); no-issue categories unchanged (existing `internal/judge`
      golden + `TestJudgeAgentDefault_*` dry tests)
- [x] Helptext (`helptext/judge.md`) + atlas (`workflow/sdlc-binary.md`)
- [x] Full `make test` (see Log: environmental failures only), then `sdlc close`

## Log

### 2026-07-29

### 2026-10-10
- Claimed from ops:0 TL dispatch (ariadne-robustness-1, batch 2).
- estimate-quality fold dropped mid-implementation: the test showed `sdlc judge` rejects the
  category outright — no manual surface, no defect there.
- Smoke: `sdlc judge plan-quality --issue 189 --dry-run` renders this issue's Decision section;
  the live run refuses with the change-code pointer.
- Full `make test`: `TestCLISignalCancelsOwnedReviewer/judge/*` drove a LIVE
  `judge plan-quality` dispatch to exercise signal cancellation — the very stateless path now
  refused. Its judge leg moved to `judge dry` (any live-dispatching category serves).
  `TestPlanningReviewConcurrencySchedules` failed once under the sharded load and passes in
  isolation; untouched by this diff.
- Second full run: three environmental failures, none in this diff's path.
  - `TestPlanningReviewConcurrencySchedules`: its unrelated `set-status` completes, then is
    killed at the test's 2s deadline. Load average was about 50 from peer slots. 4/4 pass on
    origin/main and 4/4 pass on this branch when rerun.
  - `processgroup.TestCancellationKillsDescendants`: the sandbox blocks `/bin/ps`.
  - The hermeticity guard flagged slot 3's live `sdlc merge` lock in the shared git common dir.
- [x] Full `make test`, apart from the environmental failures above.
