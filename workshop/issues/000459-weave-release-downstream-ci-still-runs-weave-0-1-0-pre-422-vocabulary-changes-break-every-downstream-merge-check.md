---
id: 000459
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '2fd39ea987862783e54a6698dbec9f124eaac5da' # card fields mirrored from issue-cards; edit via sdlc
---

# weave release: downstream CI still runs weave 0.1.0 (pre-#422); vocabulary changes break every downstream merge-check

## Problem

Downstream repos' merge-checks install the published weave release 0.1.0, which predates ariadne#422 (ship and sidecar ownership). On 2026-10-19 ariadne#394 changed construct/vocabulary/issue.cue; weave 0.1.0 then treated parley's regenerated construct/generated/vocabulary/issue.json as an "authored replacement" and failed every parley PR's merge-check (parley.nvim#396 regenerates to unblock). Any later vocabulary or schema change will trip the old released compiler the same way, in every downstream repo whose CI pins the release.

## Spec

Either cut a new weave release that includes #422 and point downstream CI at it, or make downstream merge-checks build weave from a pinned ariadne commit instead of the old release, whichever keeps CI reproducible. Then add a guard: a vocabulary or schema change in ariadne that the currently pinned downstream compiler would reject fails in ariadne's own CI (a cross-check against the pinned version), not after it lands. Publishing a release is outward-facing, so the operator approves the publish step. Workflow changes in downstream repos are local-gates changes needing the operator's label.

## Done when

- Downstream CI uses a weave that includes #422; a test vocabulary change no longer trips it.
- ariadne CI catches a change the pinned downstream compiler would reject.

## Plan

See [the local implementation plan](../plans/000459-downstream-weave-compiler-plan.md).

- [x] Compare the release and pinned-source CI routes and prepare a recommendation.
- [x] Trace the cold shipped-output failure with the actual old/new compilers.
- [ ] Obtain hub/operator approval of the plan, exact pin and publication limits.
- [ ] After approval, establish normal SDLC ownership and implement the upstream
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
