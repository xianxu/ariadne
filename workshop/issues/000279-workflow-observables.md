---
id: 000279
status: codecomplete
deps: [ariadne#277]
github_issue:
created: 2026-10-01
updated: 2026-10-02
estimate_hours: 3.31
card_mirror: '972d3ceb182a97c958edac136e44b30623bd8fca' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-01T23:06:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.59
---

# Expose authoritative workflow observations for agents

## Problem

Agents and humans reconstruct workflow progress through ad hoc issue, Git and filesystem inspection. This makes effect verification and recovery inconsistent and duplicates SDLC’s interpretation of its own state.

## Spec

Project: pair/workshop/projects/cross-slot-work-scheduling.md. Captured for operator review; no implementation is authorized by this issue creation.

Extend existing read-only SDLC inspection commands with a coherent machine-readable contract. Expose claim/assignment, current slot/worktree/branch association, workflow checkpoints, working-tree/recent-commit activity, milestone gate verdict/evidence and publication/merge outcomes. Derive from their existing authorities; do not introduce a second workflow state store.

For each observation expose its source/revision and freshness or collection time, and distinguish absent, stale and failed/unknown reads. Separate authoritative milestones from weak activity signals: file changes are not proof of correct progress. A working card proves claim state, not execution or completion. Support direct local invocation against another worktree, including a stopped/parked slot, without asking that slot’s agent to respond. Do not add Couch RPC proxies for hypothetical multi-machine support.

## Done when

- An agent can answer who owns an issue, where it is assigned, what checkpoint passed and whether it landed through documented SDLC queries.
- Structured outputs identify evidence/source and distinguish unknown/unreadable/stale from absent; queries do not mutate state.
- Tests query another local worktree, including parked/no-agent and conflicting/missing worktree cases.
- Activity versus authoritative progress and tracker freshness semantics are documented and covered by contract tests.

## Plan

Durable plan: `workshop/plans/000279-workflow-observables-plan.md`.

- [x] M1 — the contract and its tracker sections: `internal/observe` types,
      strict JSON and golden; `Assemble` for card, assignment, completion and
      landing; `issue show --json`; parked-slot, stale-tracker and no-mutation
      tests.
- [x] M2 — checkpoints (flow, plan, review verdicts, open blocking),
      workspaces and activity, `--repo`; conflicting, missing and
      other-machine cases; docs.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec             design=0.6 impl=0.05
item: greenfield-go-module   design=0.5 impl=0.22
item: smaller-go-module      design=0.1 impl=0.14
item: smaller-go-module      design=0.2 impl=0.14
item: smaller-go-module      design=0.1 impl=0.14
item: smaller-go-module      design=0.05 impl=0.14
item: cross-cutting-refactor design=0.1 impl=0.14
item: atlas-docs             design=0.1 impl=0.05
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
design-buffer: 0.15
total: 3.31
```

Items, in order:
- issue-spec: decisions plus two plan-quality rounds.
- greenfield-go-module: `internal/observe` (types, strict JSON, Assemble,
  text).
- smaller-go-module ×4:
  1. tracker-section collectors;
  2. checkpoint collectors (sidecars and ledgers at a ref);
  3. workspaces, branch activity and `--repo`;
  4. `issue show --json` and the text view.
- cross-cutting-refactor: lifecycle real-git fixtures.
- atlas-docs, then two milestone reviews.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* The calibration source is flagged stale (#127). For reference, #277 (a similar shape) came in at 2.41h actual against 4.09h estimated.

## Log




- 2026-10-02: closed — issue show --json: versioned observation (schema_version 1, golden-pinned, strict JSON, every enum and set-exactly-when invariant validated with self-naming rejection rows) of card, assignment (relation, owner worktree fate), workspaces activity, branch, checkpoints (flow, plan, per-boundary verdict + scoped open_blocking from committed artifacts; branch before landing, main archive after), completion, landing; read quality distinct from values, stale carries its reason, failed reads degrade; repo anchored on the issues dir; real-git: parked slot, stale tracker, squash landing with branch deleted, milestones through the real gates, worktree fates, two-repo --repo; no local mutation; full sharded suite green (sandbox-only processgroup failure).; review verdict: SHIP
- 2026-10-02: closed M2 — checkpoints/branch/workspaces from committed evidence (branch, then main archive; squash landing mutation-checked); failed reads degrade with reasons; every enum validated with a rejection row that asserts its own refusal (flow/provenance/boundary rows mutation-checked); flow and milestone grammars single-sourced (flow.ValidKind/ValidProvenance, issue.MilestoneTagPattern); artifact names from writers' helpers; paths from given dir/vocab; all git via counted seam (<=20); real milestone test with per-boundary scoping (mutation-checked); TestTickedMilestones; full sharded suite green; actual = measured total minus M1's 0.86h; review verdict: SHIP
- 2026-10-02: closed M1 — observation contract (schema_version 1) pinned by golden; strict JSON (unknown/duplicate keys, all enums, invariants); Assemble table + fuzz; issue show --json/--repo anchored on the issues dir (mutation-checked regression: loose dir outside a repo, given dir inside a tracked repo touches nothing); parked slot observed from another checkout, stale tracker+reason, landed via --repo; atlas Observations; full sharded suite green; actual = sdlc actual (first milestone); --no-project: pair's project tracks ariadne#279 at issue granularity; review verdict: SHIP
### 2026-10-01

Captured from the performance → messaging guarantees → SDLC ownership/observability → recovery discussion. No implementation started.

### 2026-10-01 (implementation session)

- Operator authorized the work ("work on #279"). Claimed in ariadne:1 (the
  claim records the claimant) and ran start-plan. Branch
  `000279-workflow-observables` sits at main, which includes #277, #278 and
  #275. Mapping the existing read-only inspection commands (`state`,
  `issue show/list`, `fleet`, `workspace`, `reclaim` inspect) before
  designing, so this extends them rather than adding a second state store.

- Explorer map of the existing read-only surfaces:
  - `state --json`: workspace, issues, worktrees, recent commits, drift and
    `tracker_stale`. No timestamp in JSON; current repo only.
  - `issue show`/`list`: text only.
  - `fleet inventory --json`: strict JSON with `available` + `error`
    discriminators and per-tree facts (head, dirty_count, ahead/behind), plus
    a branch-prefix issue association.
  - `workspace --json`: `schema_version` 2, the only verb taking an address.
  - `reclaim` inspect: text; its revision is the card blob OID.
  - Nothing exposes the flow record, gate-ledger decisions, verdict
    trailers, the completion binding or landing. No output carries
    observed-at or a source revision.
  - Fleet's JSON conventions (discriminator plus error, strict decode,
    goldens) are the precedent to reuse.
- Operator decisions:
  1. **Surface:** `sdlc issue show N --json`, a versioned observation; the
     text view gains the same sections. `state` and `fleet` are unchanged.
  2. **Freshness:** fetch the tracker, falling back to the last fetch marked
     `stale` with the error. Every answer carries the tracker commit and
     `observed_at`.
  3. **Targeting:** the query is repo-wide from any checkout, finding the
     local worktrees that hold the issue branch through git (including
     parked or no-agent slots). `--repo <path>` queries another repository.
- Wrote the durable plan (M1/M2). Awaiting operator approval.
- change-code plan-quality round 1: three Important findings (landed review evidence source, ledger semantics, state vs outcome) and two Minor. All folded into the plan (see its Revisions).
- Plan-quality cleared (round 2). Estimate: 3.31h (v3.1).
- change-code passed (estimate-quality: reasonable). Starting M1: the `internal/observe` contract.
- M1 step: `internal/observe` is in (types, strict JSON with duplicate-key rejection, `Assemble` for tracker, card, assignment, completion and landing). Scenario table plus round trip and rejection cases. `FuzzAssemble` ran about 700k execs with no invariant break.
- M1 step:
  - The collector (`cmd/sdlc/observe.go`) and `issue show --json` / `--repo`
    are in. The text view appends the observation block.
  - `tracker.Records` now keeps `FetchErr`, so a stale answer carries its
    reason.
  - Real-git tests, querying from another checkout:
    - a parked slot with uncommitted work (owner, owner worktree holding the
      branch, and local state unchanged);
    - an unreachable remote (stale with error, never absent);
    - a landed card (outcome and commits), also through `--repo`.
  - Running on #279 itself showed this workspace as owner, holding the
    branch.

- M1 milestone-close needed an atlas update for the new surface. Added the atlas `issue-tracker.md` Observations section (M1 scope).
- M1 review round 1: **REWORK**.
  - **BR-1 (Critical):** `issueShowRepo` resolved the repository from the
    process cwd instead of the issues dir. Outside a repository the command
    died; with `--issues-dir` it observed the wrong repository.
  - **BR-3:** an existing test therefore fetched the real checkout's tracker
    during `go test`.
  - **BR-2:** the planned golden fixture was missing.
  - Minors: a rev-parse failure was dropped; relation and claimant_worktree
    enums went unvalidated; the "not observed by this build" placeholder
    (removed in M2).
  - Fixing all of them.
- REWORK fixes:
  - BR-1/BR-3: `issueShowRepo` resolves the repository from the issues dir,
    or from `--repo`. An issues dir outside any repository degrades: the
    tracker is absent and identity and worktrees are reported as such.
    `TestObserveAnchorsOnTheIssuesDir` covers the loose dir, and a given
    issues dir while standing in a tracked repo touches nothing there.
    Mutation-checked.
  - BR-2: golden `internal/observe/testdata/observation-v1.golden.json`, with
    strict decode of the golden.
  - Minors: `tracker.ref_error` keeps a rev-parse failure; relation and
    claimant_worktree enums are validated, and claimant_worktree is set
    exactly when a claimant is.
  - Added a lesson on anchoring repositories to the given path.

- M1 closed (SHIP after REWORK; 0.86h measured). Advisories BR-4 (ref-error test), BR-5 (validate every enum including authority, one rejection table) and BR-6 (placeholder sections and golden) are folded into M2's first steps.
- M2 step: pure assembly for branch, workspaces (activity) and checkpoints.
  - Checkpoints cover flow, plan, and reviews in plan order. A boundary the
    record says closed but whose artifact is missing is `unknown`; an
    unreached boundary is omitted.
  - Advisories done: BR-4 (ref-error test); BR-5 (every enum, including each
    section's fixed authority and review boundaries, validated, with a
    rejection table); BR-6 (the placeholder is gone and the golden
    regenerated with full v1 content).
  - Plan correction: `DecideScoped`'s OpenBlocking depends on the round cap
    (it demotes past the cap). So `open_blocking` = OpenBlocking + Demoted,
    computed collector-side with `openScopeFor`.
- M2 step: collectors for the branch (local, else the remote copy), the
  holding worktrees' activity, and evidence. Evidence comes from the branch
  before landing and from main's archive, then main's plans, after done.
  Ledgers are scoped as the gates scope them.
  - The text view shows every section.
  - Real-git tests:
    - the lifecycle from another checkout: in progress; closed SHIP; the
      close artifact lost (unknown); landed by squash with the branch deleted
      (verdict from main's archive, mutation-checked);
    - worktree fates (elsewhere while a third worktree holds the branch,
      missing, other machine);
    - a subprocess bound of at most 20 git commands per query.
- M2 docs: `issue show` help gains OBSERVATIONS (sections, read quality, authority, freshness, reach). Atlas: the Observations section extended plus a verb row. README sentence. Plan Revisions record the open_blocking correction.
- The full suite found `TestGuardScopeCoversEveryPackage`, a real guard:
  `internal/observe` reads milestones but sat outside the plan-item guard's
  scan.
  - Fixed the class: added `internal/observe` to `guardScanDirs`.
  - Replaced my private ticked-milestone regex with `issue.TickedMilestones`,
    which shares `milestonePlanRE` (the regex now captures the box). It is
    added to `planItemMatchers` and exempted like `MilestonesInPlanOrder`.
  - `assembleCheckpoints` filters the plan once through `PlanItemsBody`.
- M2 review round 1: five Important findings, most of them repeat families
  ("fix rules, not instances"):
  - BR-7: read failures dropped to zero values (details parse or show, flow,
    workspace.Resolve).
  - BR-8: enums still unchecked (boundary, verdict, flow). The BR-5 Log line
    overstated what was done.
  - BR-9: the M1-milestone lifecycle test and the trailer cross-check were
    dropped without a Revisions entry.
  - BR-10: evidence paths came from env defaults instead of the given issues
    dir or the authority.
  - BR-11: artifact names were restated instead of using
    `planGateSuffix` / `boundaryGateSuffix` / `sidecarPath` /
    `reviewMilestoneRe`.
  - Minors: Resolve bypassed the counted seam; sidecar rows used the last
    match; the `--repo` test wasn't across two tracker repositories.
  - Fixing each as a rule across the whole collector.
- M2 review fixes (all as rules, applied across the collector):
  - BR-7: no zero-value fallbacks.
  - BR-8: every enum validated against its authority.
  - BR-9: a real milestone test (plan-ledger writer, M1 SHIP, M2 with an
    open Important; scoping mutation-checked). The trailer cross-check is
    withdrawn in Revisions.
  - BR-10: paths come from the given dir or vocab.
  - BR-11: names come from the writers' helpers.
  - Minors: Resolve is counted; first metadata row wins; a two-repo
    `--repo` test.
  - Found while testing: an undisposed plan-quality finding is inherited by
    every boundary (`seedFromPlanGate`). That is real gate behavior, so the
    fixture disposes its plan finding.
- M2 review round 2: BR-8 stayed open, and rightly. My earlier rejection
  rows were never written (an unasserted scripted replace missed after
  gofmt), and the Log claimed them. Worse, the existing table was vacuous:
  its `append([]Review(nil), …)` copy produced a nil collection, so every
  row failed on "collections must be present".
  - Rewrote the table: every row asserts the refusal names its own field.
    Each of the flow kind, provenance and boundary rows was mutation-checked
    and goes red.
  - DRY: `flow.ValidKind` / `flow.ValidProvenance` are shared with
    `flow.Parse`, and `issue.MilestoneTagPattern` builds both
    `milestonePlanRE` and the boundary grammar.
  - BR-12: every collector git call goes through the counted seam.
  - Added a direct `TestTickedMilestones` (repeated rows, `[.]`, lettered
    tags).
  - Added two lessons (table rows assert their own refusal; scripted edits
    assert their anchor).

- M2 closed (SHIP; 0.63h measured). Three advisory Minors (verdict iff present non-plan review; envelope claim scope; plan.go doc-comment attachment) are being fixed before the issue close.
- Fixed the M2 advisories: the verdict is set exactly when a present non-plan review (two rejection rows); the envelope test and the atlas now state the collector-only scope; the plan.go doc comments are reattached, with capture groups 1 and 2 documented.
