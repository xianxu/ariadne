# Boundary Review — ariadne#180 (milestone M2)

| field | value |
|---|---|
| issue | 180 — project vocabulary model |
| boundary | milestone M2 |
| window | `3feac0619757cde0ff51b908db67218360c53422..HEAD` |
| reviewer | codex |
| timestamp | 2026-07-16T12:33:07-07:00 |
| verdict | FIX-THEN-SHIP (high confidence) |

## Summary

M2 delivered the typed project parser, behavior-preserving tick migration,
and noun-table conformance gate. No Critical defect was found. The raw review
transcript was condensed after the whole-issue review identified it as an
unbounded artifact.

## Findings

1. **ARCH-DRY:** issue discovery still defaulted to the literal
   `workshop/issues` instead of `vocab.Issue().Discovery().Home`.
2. **ARCH-PURE:** Markdown section parsing treated fenced example headings as
   structure.
3. **ARCH-PURPOSE:** typed `Doc.Tasks` collected checkboxes outside the real
   `## Breakdown` section.
4. **ARCH-PURPOSE:** the issue and plan disagreed about close-time schema
   validation while legacy `status: active` brain projects still existed.
5. M2's detailed checklist and durable verification evidence lagged the
   implementation.

Minor follow-ups noted that `SetTaskState` accepted arbitrary state bytes and
that one actual-measurement test comment named the old seam.

## Remediation

- The validate gate now derives the issue discovery fallback from the issue
  vocabulary model while preserving explicit overrides.
- Project Markdown spans use a fence-aware scanner. Typed tasks are
  Breakdown-only; the legacy whole-document tick behavior is an explicitly
  named compatibility seam.
- #171 owns activating close-time project validation after migrating the
  grandfathered brain records.
- The issue Log and detailed plan record the bad-status push refusal, live
  project validation, full-suite run, and vocabulary checks.
- The milestone-actual deviation correction is documented as an enabling side
  quest: cumulative claim-to-HEAD time cannot validate milestone increments.

## Verification

- `go test ./cmd/sdlc -run 'Validate|ActualDeviation|Project' -count=1`
- `go test ./cmd/sdlc/internal/project ./pkg/vocab -count=1`
- `go vet ./cmd/sdlc/... ./pkg/vocab/...`
- `sh construct/vocabulary/vet_test.sh`
- live project instance validation
- scratch bad-status push refusal
- vocabulary freshness and `git diff --check`

## Architecture

- **ARCH-DRY:** resolved by model-derived discovery.
- **ARCH-PURE:** resolved by the in-memory fence-aware structural scanner.
- **ARCH-PURPOSE:** resolved by Breakdown-scoped typed tasks and explicit #171
  ownership of legacy close validation.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```
