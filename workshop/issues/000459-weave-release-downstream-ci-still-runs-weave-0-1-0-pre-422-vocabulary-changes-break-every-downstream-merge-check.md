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

- [ ]

## Log

### 2026-10-19
