---
id: 000509
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'eb1f7e2cad4e9fb4d73e042db905f40d691683b7' # card fields mirrored from issue-cards; edit via sdlc
---

# Landing scoped local checks

## Problem

pair#743 moves native conformance from capacity-constrained GitHub macOS runners to a machine-wide Tart runner. The operator requires one conditional conformance run per landing train, after its suite, with nightly burn-in before activation. Current required local checks also gate each car's readiness; declaring conformance required would run it per car. The schema has no scope/path selection and every local run acquires a full-suite slot.

## Spec

- Add a narrow opt-in landing scope and conditional path-list reference to declared local checks, preserving existing per-head defaults and strict config validation. Read authoritative config/path selection from the train's configured main base.
- Exclude landing-scoped evidence from individual-car readiness. Run selected checks once on the exact integration head after a passing suite and before merge, including train-of-one/single-landing fallback. Failure blocks landing without suite bisection or automatic retries.
- Select from the union of each included car's touched paths, including deleted/renamed paths; a later car reverting an earlier touch must not hide eligibility. Recompute for the prefix actually landing and after reintegration. Unavailable path evidence refuses conservatively.
- Reuse existing local-check execution, retained logs and exact-head status/provenance. Code changes invalidate evidence. Explicitly define any docs-only reuse through the existing equivalence proof.
- Permit a VM-backed local check to use its own runner-owned resource instead of consuming/waiting for a suite slot. Do not reinterpret weight (suite slots) as VM capacity or add a second ticket queue. pair owns the two fixed VM positions.
- Leave landing activation off during nightly-only burn-in; an optional failed train status is still red CI, so burn-in must not run on train heads. External scheduling invokes the named check on main; no scheduler is required in SDLC.
- Pair consumes the capability only after this lands; no shared-contract implementation in pair.

## Done when

- Stateful train tests prove post-suite order, exactly one eligible execution per final integration head, no execution for unrelated paths, and no per-car execution.
- Cases cover any-car union/later revert, excluded car/good prefix, train-of-one, failure without bisection, code reintegration, and stale/absent evidence.
- Runner tests prove resource distinction, cancellation/timeout cleanup, retained failure logs, and compatible existing configuration.
- Operator verifies local-gates changes and applies the label; workers never self-label.

## Plan

- [ ] Design the narrow schema and train/single-landing selection together, then obtain full-flow approval.

## Log

### 2026-10-20

Filed from pair#743 read-only audit. Relevant seams: cmd/sdlc/localchecks.go, localcheck_run.go, landingcar.go productionCarGates, landingtrain.go conductor.build/runSuite/waitCI, cigate.go, internal/landingline/trainstate.go, internal/affected/glob.go and internal/gitx/window.go. Tests: localchecks_test.go, landingtrain_test.go, landingcar_affected_test.go, landingtrain_provenance_test.go, localcheck_signal_unix_test.go. ARCH-DRY: extend the existing post-suite check owner and exact-head runner, rather than wrapping or bypassing landing in pair.
