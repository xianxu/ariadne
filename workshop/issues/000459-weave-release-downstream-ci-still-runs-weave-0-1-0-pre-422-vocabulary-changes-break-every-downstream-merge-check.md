---
id: 000459
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours: 2.72
collaboration-mode: collaboration
card_mirror: 'a29382afb6c151a9b64d7bdba8337e14c0cd5d4b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T01:11:31-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# weave release: downstream CI still runs weave 0.1.0 (pre-#422); vocabulary changes break every downstream merge-check

## Problem

Downstream repos' merge-checks install the published weave release 0.1.0, which predates ariadne#422 (ship and sidecar ownership). On 2026-10-19 ariadne#394 changed construct/vocabulary/issue.cue; weave 0.1.0 then treated parley's regenerated construct/generated/vocabulary/issue.json as an "authored replacement" and failed every parley PR's merge-check (parley.nvim#396 regenerates to unblock). Any later vocabulary or schema change will trip the old released compiler the same way, in every downstream repo whose CI pins the release.

## Spec

Either cut a new weave release that includes #422 and point downstream CI at it, or make downstream merge-checks build weave from a pinned ariadne commit instead of the old release, whichever keeps CI reproducible. Then add a guard: a vocabulary or schema change in ariadne that the currently pinned downstream compiler would reject fails in ariadne's own CI (a cross-check against the pinned version), not after it lands. Publishing a release is outward-facing, so the operator approves the publish step. Workflow changes in downstream repos are local-gates changes needing the operator's label.

## Done when

- Ariadne declares the approved full post-#422 compiler pin and its CI executes
  that exact compiler against candidate vocabulary exports in cold consumers;
  an incompatible changed-output control fails, while the approved pin succeeds.
- Authored-output/sidecar safety controls remain strict and actual CI gateway
  workflow tests prevent candidate/PATH compiler substitution.
- A verified active GitHub consumer inventory links one unclaimed local-gates
  migration issue per repository. Consumer adoption and real downstream CI
  regeneration are accepted in those issues after the guard lands, owned by
  their repository slots and the operator.

## Plan

See [the local implementation plan](../plans/000459-downstream-weave-compiler-plan.md).

- [x] Compare the release and pinned-source CI routes and prepare a recommendation.
- [x] Trace the cold shipped-output failure with the actual old/new compilers.
- [x] Obtain hub approval of option B, exact initial pin and publication limits.
- [x] After approval, establish normal SDLC ownership and implement the upstream
  guard plus operator-verified, independently owned downstream migrations.


## Log

### 2026-10-19

### Local preparation (2026-10-19)

- Explicit local-only dispatch overrides normal claim/start-plan publication for
  this preparation. Neither command ran: the card remains open/unclaimed, and
  this slot holds a local unpublished issue branch from fetched main. No PR,
  tag, release, tap update, workflow dispatch, downstream edit or label was made.
- Source evidence: published weave-v0.1.0 peels to
  `4de31b2de32c8143d61abe17bbe7f7654503cd42`; #422 merged at
  `664ff7938b85a79b04c80cf1e1c90279f06342bc`; proposed immutable CI source is
  `63d6bfd86509d97649e10d7a5dd1e9a06db17d21`. All identities came from Git.
- The shared merge-check workflow is refreshed by seed on compile, but that does not commit a consumer workflow update or change the workflow already executing in GitHub.
  Bootstrap's capability probe accepts 0.1.0 and does not pin the tap version.
  Read-only primary-checkout inventory found 20 downstream workflows; remote
  deployment identity and every shipped-output declaration still need rollout
  audit. No remote-fleet completeness is claimed from local files.
- Existing `go test ./cmd/weave/internal/plan -run
  '^Test(ShippedPaths|ShipStaged|ShippedOutputOwnership)$' -count=1` passed
  (2.759s). Both compiler snapshots built successfully with cached dependencies,
  `GOPROXY=off`, `GOTOOLCHAIN=local`, CGO disabled and `-trimpath`.
- Actual scratch process comparison used independent temporary Git owner/leaf
  fixtures, a tracked generated SKILL.md plus matching sidecar, and no runtime
  ownership inventory. 0.1.0 exited 1 with "preserving authored replacement";
  the proposed pin exited 0, regenerated expected bytes and the matching digest.
  Both binaries refused the tampered-output control and left its bytes unchanged.
  This is an observed minimal ownership reproduction, not yet the proposed
  whole candidate vocabulary/schema corpus guard or a real downstream CI run.
- Scratch sources, logs and observed-results.json remain at
  `/var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/weave459-rehearsal.xrRzR4`
  for local review. They are not release candidates and were not uploaded.

## Revisions

### 2026-10-19 — plan review corrections

A fresh advisory reader identified seed semantics, corpus scope and the need for
an independently exercised changed-output control. Source inspection confirms
seed refreshes workflow bytes; seed-once currently performs Makefile adoption,
so the plan uses an authored pin outside the manifest and no new seed mechanism.
The guard covers real top-level vocabulary exports and their producer
declarations, with instance fixtures validated separately, and explicitly runs
the old compiler through compile rather than failing only on pin membership.

### 2026-10-19 — approval and hand-off boundary

The hub approved option B and the full proposed pin, permitted normal SDLC
publication, and asked this worker to claim/proceed. Claim/start-plan now record
this workspace normally. The hub then explicitly authorized filing unclaimed
migration issues in verified active GitHub consumer :0 checkouts, and prohibited
isolated-checkout downstream PRs: TL dispatches consumer-owned work after this
guard lands. Current Done when is synchronized to that boundary. The original
"Downstream CI uses a weave that includes #422; a test vocabulary change no longer
trips it" criterion remains required fleet acceptance in those peer issues;
this issue does not claim those future consumer migrations are already done.

- Plan-quality round 1 (PQ-1) requested named function boundaries and one
  adversarial-input strategy per risky function. Added signatures and mechanical
  oracles to the durable plan, separating pure parsing/comparison from process
  IO; no code implementation preceded the gate.

- Verified inventory: [remote consumer inventory](../plans/000459-downstream-compiler-inventory.md),
  17 active unarchived GitHub workflows, 15 ordinary SDLC filing targets.
- Operator decision pending: ops has opted out of SDLC/seeded CI but its remote
  CI workflow is active; emma-private has no issue tracker. Hub explicitly said
  not to touch either repo and will bring both to the operator in the morning.
  These exceptions are recorded, not silently skipped as completed migrations.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md`
against `baseline-v3.1.md`. Method A only. Calibration is provisional/stale.

Scope is the upstream compiler pin, guard and workflow, regression/conformance
checks, documentation, review and fifteen issue hand-offs; downstream migration
implementation and operator waiting are excluded. Python subprocess/tempfile,
Git and existing weave/vocabulary helpers cover the stack: no novel library or
service is needed. Two well-specified small modules map to the tooling primitive;
the resolved plan applies the 0.2 design discount and 15% design buffer. Impl
values below already apply v3.1's 40% scale once, with familiar-stack 1.0.
Cross-repo filing uses the large-refactor coordination primitive at its upper
implementation bound; there is no vendor propagation in this issue, so no
additional vendor multiplier. The GitHub inventory is one API-discovery item.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.20 impl=0.12
item: smaller-go-module design=0.08 impl=0.32
item: smaller-go-module design=0.08 impl=0.32
item: cross-cutting-refactor design=0.04 impl=0.20
item: cross-repo-refactor-large design=0.10 impl=0.60
item: atlas-docs design=0.04 impl=0.08
item: milestone-review design=0.02 impl=0.20
item: real-api-discovery design=0.00 impl=0.24
design-buffer: 0.15
total: 2.72
```

Design 0.56 × 1.15 + implementation 2.08 = 2.724 hours, rounded to 2.72.

- Estimate-quality returned INFO (no blocking finding): each tooling allocation
  includes two bounded responsibilities rather than a single table-sized module:
  exact compiler acquisition/identity plus workflow process tests, and corpus
  construction/comparison plus cold-state conformance. These include targeted
  tests, observed old/new compiler runs and resulting fixes; the review item
  includes the recorded suite's active orchestration (queue/idle excluded).
  The two 0.08/0.32 allocations intentionally exceed one smaller-module range
  because each covers that paired responsibility. Corrected a prose/subtotal
  addition error to 2.08 implementation hours and a reconciled 2.72 total.
- First migration hand-off filed from 42shots:0 as 42shots#1, unclaimed,
  depending on ariadne#459; remaining filings are sequential and independently
  verified before the hub receives their complete list.

- Echo filing exception: its legacy path published echo#1 remotely but also made
  a local issue-sync commit, leaving echo:0 main ahead/behind. The hub explicitly
  ordered it left exactly as-is because the operator has unrelated uncommitted
  echo-control-board.md and .side work; hub/operator reconcile in the morning.
  Remaining repositories were checked for local tracker configuration and remote
  issue-tracker branch before filing; legacy paths would be listed, not invoked.

- Completed verified unclaimed migration hand-offs: 42shots#1, astro#1, ducks#12, echo#1, kaggle#11, kbench#29, metis#70, nous#54, pair#699, parley.nvim#398, parli#1, robotics#1, tools#84, xianxu.dev#5, you-decide#16. All depend on ariadne#459; repository owners implement after TL dispatch. No other legacy exceptions were found.

### Implementation verification (2026-10-19)

- Added exact source-pin acquisition/provenance checks, isolated candidate vs
  downstream compilers, and the every-PR/main real compatibility guard. The pin
  is authored outside the manifest, and CI helper changes require operator
  local-gates verification.
- Red/green: actual workflow tests first refused missing pinned-toolchain
  acquisition; matcher tests refused unprotected CI helpers. Added a negative
  old-compiler control that prints the ownership diagnostic without the
  preserved-baseline marker: it failed before the second oracle was enforced.
- Parent verification: portable workflow blocks pass; local-gates matcher
  passes all 17 patterns and near misses. Guard tests pass with pure grammar,
  linear corpus comparison, stateful cold transitions, generator failure before
  compiler invocation, zero-exit/no-output refusal, timeout and SIGTERM cleanup.
- Focused Go vocabulary instance/export/resolution and shipped ownership tests
  pass; real TestCompileRegeneratesShippedOutputCold passes.
- Parent real conformance: approved pinned compiler passes seven cold fixtures
  across six actual top-level noun exports. Real 0.1.0 exits 1 at compile with
  preserving-authored-replacement plus COLD-REFUSAL-PRESERVED, after checking
  unchanged output and sidecar bytes. Both paths use the candidate's actual
  producer declaration/marker and vocabulary generator, with no live peers.
  Evidence remains under
  /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/weave459-verification.RaEfPU.
- PR319 is the local-gates hand-off. Full recorded suite, close review and
  landing remain outstanding pending operator verification; agents never apply
  local-gates-verified. No release, tag, tap or downstream workflow was edited.

### Authorized landing continuation (2026-10-19)

- The operator applied local-gates-verified to PR319; the GitHub label and successful local-gates-guard/verdict were read before continuation. The original branch remains unchanged in code.
- A separate ordinary clone at ariadne-slot3/ariadne-459 resolves through the Git-verified environment host as the existing ariadne:3 owner, using the same GitHub origin; no reclaim or owner change was made. The original slot checkout remains on #465.
- The recorded suite is queued on the pushed implementation head. Per tl-worker, close review may overlap that queued run; a recorded pass on the final code must precede landing. Upstream implementation and all fifteen hand-offs are complete; consumer rollout remains separately owned.
