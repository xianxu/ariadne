---
id: 000135
status: working
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-26
estimate_hours: 2.65
started: 2026-06-26T18:16:39-07:00
---

# sdlc close actual not applicable

## Problem

`sdlc close --no-actual` can currently produce an invalid closed issue: it flips `status: done` while omitting `actual_hours`, but the issue schema requires `actual_hours` for done issues. This happened during pair#72 and was caught by the boundary review.

There are legitimate closes where focused dev-hours are not applicable or cannot be measured: roadmap-only records, wontfix/punt-like administrative closes, or cases where telemetry is unavailable and recording a made-up number would pollute velocity calibration.

The schema and close command need an explicit, valid "not applicable" representation instead of relying on a missing field.

## Spec

Allow the close record to represent actual time as either:

- a numeric hour value, as today; or
- an explicit string sentinel meaning not applicable, e.g. `actual_hours: N/A`.

Expected behavior:

- The issue schema/vocabulary accepts `actual_hours` as number-or-N/A for closed issues.
- `sdlc close --no-actual` writes the N/A sentinel instead of omitting `actual_hours`.
- The close log/audit message still records that velocity calibration is skipped.
- Calibration ledger logic excludes N/A rows.
- Active-time sanity checks are bypassed only for the explicit N/A sentinel, not for arbitrary strings.
- Help text explains when `--no-actual` is appropriate and what `actual_hours: N/A` means.
- Existing numeric actual behavior remains unchanged.

Prefer a small closed set of valid sentinel spellings. `N/A` is the requested spelling unless implementation finds a stronger reason to use a different single token.

## Done when

- [x] Issue schema/vocabulary accepts `actual_hours` as either a number or the approved N/A sentinel.
- [x] `sdlc close --no-actual` writes a schema-valid closed issue with `actual_hours: N/A`.
- [x] `sdlc issue validate` passes on an issue closed with `--no-actual`.
- [x] Calibration ledger and drift logic skip N/A rows.
- [x] Invalid arbitrary strings in `actual_hours` still fail validation.
- [x] Help text documents the N/A close path and its velocity-calibration implication.

## Plan

- [x] Locate the issue vocabulary/schema definition for `actual_hours`.
- [x] Update schema/model to permit numeric actuals or the approved N/A sentinel.
- [x] Update `sdlc close --no-actual` write path.
- [x] Update validation and close tests for numeric, N/A, missing, and invalid-string actuals.
- [x] Update calibration ledger handling.
- [x] Update help text.

## Estimate

```estimate
model: estimate-logic-v2
item: smaller-go-module design=0.2 impl=0.5
item: smaller-go-module design=0.1 impl=0.5
item: smaller-go-module design=0.1 impl=0.4
item: atlas-docs design=0.1 impl=0.2
item: milestone-review design=0.0 impl=0.4
design-buffer: 0.30
total: 2.65
```

Produced via `brain/data/life/42shots/velocity/estimate-logic-v2.md` against `baseline-v2.md`. Method A only; the calibration source is currently marked stale by `sdlc estimate-source`, so the estimate is provisional.

## Log

### 2026-06-26

- Implemented `actual_hours: N/A`: updated the CUE issue schema, close write path, calibration ledger parser/guard, CLI help, and atlas docs. Verification: targeted vocabulary/sdlc/estimate tests pass; synthetic `actual_hours: N/A` issue validates through both `vocabulary validate-instance` and `sdlc issue validate`, while `actual_hours: unknown` is rejected; `go test ./...` passes. `sdlc issue validate --all` still fails on unrelated pre-existing issue-structure gaps, but #135 conforms.

- Planning: claimed the issue, ran `sdlc start-plan --issue 135`, read `construct/vocabulary/issue.cue`, and wrote durable plan `workshop/plans/000135-close-actual-na-plan.md`. Architecture notes: use one N/A sentinel contract (`ARCH-DRY`), keep sentinel parsing/checking pure with close/validation as IO seams (`ARCH-PURE`), and cover schema, close, ledger, validation, and help/docs rather than only the writer (`ARCH-PURPOSE`).

Created from pair#81 retro point 4. The observed bug was `sdlc close --no-actual` marking pair#72 done without `actual_hours`, leaving the issue schema-invalid. Desired direction from the user: allow `actual_hours` to be a number or an explicit string representing N/A.
