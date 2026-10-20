---
id: 000488
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
severity: medium
collaboration-mode: collaboration
card_mirror: '0b5899289c90a738130b064f0ded4301c37aa46c' # card fields mirrored from issue-cards; edit via sdlc
---

# Duplicate-ID guard counts numbered .side directories as issue records

## Problem

PR338's merge-check failed in scripts/merge-checks.d/40-duplicate-issue-id.sh (GitHub run 39388400345) after adding a legitimate numbered issue side folder. The checker reported the issue .md and its matching .side directory as two claimants of #487. The .side contained only a measured sandbox transcript, not a second issue record. The local-gates guard passed. #424 explicitly supports document side folders and archives them with the owning document.

The CI diagnostic named workshop/issues/000487-make-testscratchfallbackoverridesinheritedtemp-independent-of-host-sandbox-temp-permissions.md and the same stem ending .side as separate issue paths. #487 avoids the bug by relocating the unchanged transcript into its numbered workshop/plans artifact family; it does not bypass or modify the guard.

## Spec

Make issue-ID collection distinguish canonical issue Markdown records from support directories and non-record artifacts, reusing the repository's existing filename/location rule where possible. Directories such as numbered .side folders must not claim IDs, including their nested support files. Preserve rejection of genuinely duplicated canonical issue records and fail-closed behavior when Git evidence cannot be read.

## Done when

- A real Git fixture with one issue .md and its numbered .side support directory fails against the old collector and passes after correction, through the same CLI seam CI calls.
- Current and supported archive layouts ignore non-record directories/support files while genuine duplicated issue .md records still fail.
- Focused collector/CLI tests and the merge-check wrapper pass; no gate bypass or reduced unknown-evidence handling is introduced.

## Plan

- [ ]

## Log

2026-10-20: Filed unclaimed by ariadne:3 while finishing #487. CI run https://github.com/xianxu/ariadne/actions/runs/39388400345 produced the false duplicate. The fixture code, sandbox profile and measured transcript content are unrelated to the collector defect. No collector or workflow change is included in #487.
