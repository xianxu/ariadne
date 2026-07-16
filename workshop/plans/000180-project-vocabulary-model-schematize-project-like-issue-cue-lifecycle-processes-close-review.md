# Boundary Review — ariadne#180 (whole-issue close)

| field | value |
|---|---|
| issue | 180 — project vocabulary model |
| boundary | whole-issue close |
| window | `a18dd5e850c351909825d3cc8555ce366fe64604..481878f` |
| reviewer | codex |
| timestamp | 2026-07-16T16:19:20-07:00 |
| verdict | FIX-THEN-SHIP (high confidence) |

## Summary

The implementation fulfills the project-vocabulary purpose. The CUE model
drives lifecycle consumers; project commands enforce modeled transitions; the
board and close paths cover their failure modes; and README plus atlas map the
new surface. No correctness blocker remained. This bounded record replaces the
raw generated review transcript.

## Strengths

- Lifecycle, discovery, scaffold, guards, help, and archive behavior derive
  from `vocab.Project()`.
- Parsing, metadata decoding, estimates, summaries, guards, board computation,
  and ledger transformation remain pure cores behind thin IO seams.
- Close fails closed on unknown guards, incomplete or invalid calibration,
  duplicate logical refs, unavailable peers, and transaction failures.
- Tests cover hostile YAML, traversal, fenced Markdown, aliases, non-finite
  values, rollback, and EOF ledger layouts.
- README and atlas cover commands, bypasses, lifecycle, storage, archive, and
  calibration.

## Findings

No Critical findings.

1. Detailed M1 plan rows remained unchecked despite delivered and reviewed
   work.
2. The M2 review sidecar retained a 334 KB raw prompt, diff, and duplicated
   transcript instead of a bounded decision record.
3. The issue Spec's historical Dogfood paragraph still described #182 as out,
   while the later operator decision and live project put #182 in MVP scope.

Minor follow-ups were deferred to their owning work: the parser tolerates
legacy trailing task prose despite the datatype's terse canonical form, and
#171 owns removal of the brain-era `data/project` close help.

## Remediation

- Reconciled every delivered M1 detailed-plan checkbox.
- Condensed the M2 and whole-close review artifacts to verdict, findings,
  remediation, verification, and architecture evidence.
- Added issue and plan revisions establishing the authoritative scope:
  `ariadne#180`, `ariadne#171`, and `ariadne#182` are included;
  `ariadne#15` is explicitly out.

## Verification

- `go test ./... -count=1`
- `bash construct/vocabulary/vet_test.sh`
- `go run ./cmd/vocabulary check --output construct/generated/vocabulary`
- live project conformance and derived status checks
- `git diff --check`

## Architecture

- **ARCH-DRY:** pass — shared model and parsing sources serve derived consumers.
- **ARCH-PURE:** pass — business rules remain deterministic and IO-free.
- **ARCH-PURPOSE:** pass — the model is enforced across verbs, guards,
  validation, docs, board derivation, and calibrated close.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```
