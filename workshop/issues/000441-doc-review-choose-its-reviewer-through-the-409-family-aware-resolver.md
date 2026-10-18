---
id: 000441
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours: 0.735
collaboration-mode: collaboration
card_mirror: '737a8db96a9943e0b1a506a742f9dd5408b3bc73' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T09:10:46-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# doc-review: choose its reviewer through the #409 family-aware resolver

## Problem

`cmd/doc-review` hard-codes its reviewer (`DefaultAgent = AgentCodex`, "a different vendor from the usual co-authoring agent"). #409 gives sdlc one family-aware reviewer resolver (cross-family by default, visible fallback); doc-review does not use it, so a codex-authored document is reviewed by codex, and an unavailable codex has no visible fallback.

## Spec

**Why it can't simply call the resolver:** #409's resolver lives in `cmd/sdlc/internal/judge`, and Go's `internal` rule
keeps `cmd/doc-review` out. Both proper routes move code, so this item is outside the quick-flow shell.

**Proposed design (awaiting the hub):**
- **Extract the pure policy into `pkg/reviewpolicy`:** `Family`/`FamilyOf`, `ChangeKind`, the default orders,
  `Mode`/`DefaultMode`, `Policy` (`Validate`, `Reviewer`, orders, models), `Candidate`/`Skip`/`Availability`, `Plan`,
  and `Load(root)` (the `workshop/review.json` loader, moved from cmd/sdlc), plus `AuthorFamilies(trailerValues)`.
  `judge` keeps type aliases (`ReviewerPolicy = reviewpolicy.Policy`, …), so no sdlc caller changes. It also keeps the
  run-shaped parts (`Unavailable`, `LocalAvailability` with its containment rule, `DispatchChoice`). ARCH-DRY: one
  policy for both binaries.
- **doc-review plans its reviewer:**
  - The document's authors are the `Agent:` trailers of the commits touching the file (`git log -- <file>`) plus the
    session's agent.
  - The kind is design-heavy, since a document is a design.
  - Availability is the CLI on PATH.
  - An explicit agent argument still wins.
  - doc-review exists to get a *second vendor*, so a repo's `same` (or the default) counts as `other` for it; we never
    regress to self-review. `explicit` pairs are honoured.
- **Visible degradation:** a candidate whose CLI is missing or fails to launch is announced and skipped. A codex crash (a non-zero run) keeps
  today's hint; only a missing CLI or a failed launch moves on (plan Revisions). With no other family available, it refuses and names each
  candidate. The report's frontmatter records `reviewer:` as #409's record (`agent (family; author …)[; degraded: …]`).
- **Not in scope:** quota detection for doc-review's runs. Its run shape differs (codex `-o` capture), and a failure
  there keeps today's message.

## Done when

- doc-review chooses its reviewer through the #409 resolver (the document author family when known, else the policy order), with the same visible degradation; or a decision records why it stays fixed.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: cross-cutting-refactor design=0.05 impl=0.14
item: smaller-go-module      design=0.05 impl=0.14
item: atlas-docs             design=0.0 impl=0.06
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
design-buffer: 0.15
total: 0.735
```

Derivation: design settled by two plan-quality rounds. impl= is 40% of the v2 midpoints (v3.1): the two-package move
with aliases (cross-cutting refactor, 0.35 → 0.14); doc-review's choice, fallback and record (smaller module, 0.35 →
0.14); help and atlas (0.15 → 0.06); two milestone reviews (0.35 → 0.14 each). The calibration doc is flagged stale.

## Plan

Durable plan: `workshop/plans/000441-doc-review-choose-its-reviewer-through-the-409-family-aware-resolver-plan.md` (full flow, per the hub).

- [x] M1 — extract `pkg/reviewpolicy` (+ `Resolve`) and `pkg/agentenv` (env, trailer log parsing); judge/agentid keep aliases; every existing test passes unchanged
- [ ] M2 — doc-review plans, falls back and records its reviewer through it; never self-review; help and atlas

## Log

### 2026-10-18
- Claimed by ariadne:2. The `internal` boundary makes this a move plus wiring, not a one-liner, so the plan goes to the hub before code.
- Hub approved the extraction (full flow) and the never-self-review decision for doc-review (same and the default count as other, explicit honored, stated in its help). Quota detection out of scope.
- M1 implemented: `pkg/agentenv` (Env, EnvFrom, Known, Main, InSession, the agent names, TrailerKey, Commit,
  LogFormat, ParseLog; `agentid` re-exports them) and `pkg/reviewpolicy` (Agent, Family, FamilyOf, ChangeKind, orders,
  Mode, DefaultMode, Policy with Validate/Reviewer/EffectiveMode, Candidate, Skip, Availability, Plan, AuthorFamilies,
  Choice with First/Refusal, Resolve, NoReviewerError, Record, ConfigPath, Load). `judge` aliases them
  (`AgentCLI = reviewpolicy.Agent`, `ReviewerPolicy`, `ReviewChoice`, …) and wraps `ReviewPlan`, `ResolveReview` and
  `ReviewerRecord`. `DispatchChoice`'s refusal still matches both `NoReviewerError` and `ErrReviewDidNotRun`. cmd/sdlc's
  `loadReviewerPolicy` and `authorFamiliesFromLog` delegate. The pure policy tests moved to
  `pkg/reviewpolicy/policy_test.go`; the run-shaped ones stay in judge. judge, cmd/sdlc (policy, boundary, per-site,
  loader, context guard), agentid, gitx and golint tests pass.
