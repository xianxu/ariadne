---
id: 000135
status: working
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-26
estimate_hours:
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

- [ ] Issue schema/vocabulary accepts `actual_hours` as either a number or the approved N/A sentinel.
- [ ] `sdlc close --no-actual` writes a schema-valid closed issue with `actual_hours: N/A`.
- [ ] `sdlc issue validate` passes on an issue closed with `--no-actual`.
- [ ] Calibration ledger and drift logic skip N/A rows.
- [ ] Invalid arbitrary strings in `actual_hours` still fail validation.
- [ ] Help text documents the N/A close path and its velocity-calibration implication.

## Plan

- [ ] Locate the issue vocabulary/schema definition for `actual_hours`.
- [ ] Update schema/model to permit numeric actuals or the approved N/A sentinel.
- [ ] Update `sdlc close --no-actual` write path.
- [ ] Update validation and close tests for numeric, N/A, missing, and invalid-string actuals.
- [ ] Update calibration ledger handling.
- [ ] Update help text.

## Log

### 2026-06-26

Created from pair#81 retro point 4. The observed bug was `sdlc close --no-actual` marking pair#72 done without `actual_hours`, leaving the issue schema-invalid. Desired direction from the user: allow `actual_hours` to be a number or an explicit string representing N/A.
