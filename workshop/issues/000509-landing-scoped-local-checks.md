---
id: 000509
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '9a61de2e1e4a1633b8957d7ac2c36b63c4535dd6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T22:27:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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

Durable design: [landing checks plan](../plans/000509-landing-checks-plan.md). Await operator review before implementation and the plan-quality gate.

- [ ] M1 — Versioned declaration, pure selection/proof and separate runner-resource execution; adversarial controls and milestone review.
- [ ] M2 — Shared post-suite train/single-landing gate, full consumer sweep, stateful/private conformance and milestone review.
- [ ] Recorded suite, full close, operator local-gates verification where applicable, and central landing.

## Log

### 2026-10-20

Filed from pair#743 read-only audit. Relevant seams: cmd/sdlc/localchecks.go, localcheck_run.go, landingcar.go productionCarGates, landingtrain.go conductor.build/runSuite/waitCI, cigate.go, internal/landingline/trainstate.go, internal/affected/glob.go and internal/gitx/window.go. Tests: localchecks_test.go, landingtrain_test.go, landingcar_affected_test.go, landingtrain_provenance_test.go, localcheck_signal_unix_test.go. ARCH-DRY: extend the existing post-suite check owner and exact-head runner, rather than wrapping or bypassing landing in pair.


### 2026-10-20 22:45 PDT — Design checkpoint

Claim/start-plan completed on this issue's branch from main. Wrote the durable proposal after auditing local-check codecs, saved evidence, subprocess cancellation, CI/car consumers, pure train transitions and both single-landing shells. ARCH-DRY: one shared post-suite gate and existing execution/log shell; runner-owned lifetime admission remains external, with zero suite allocation or VM lease probing.

Authenticated pair:3 receipt 74652a7c-7ea4-4228-9888-41cf23280c2d confirms make conformance-vm, the direct tests/conformance-paths.json reference, 20m runner budget plus up to10s cleanup, 30m named timeout and at least10s cancellation grace. The proposal uses15s runner grace. No activation or peer workflow changes. A follow-up send was refused with “session listing observations disagree” and no receipt; it was not blindly resent. Read-only inspection confirmed the current path syntax. The contract establishes the intended seam, not completed VM qualification.

Operator tradeoffs are explicit in the plan: v2 unsafe-to-ignore semantics, manual-only disabled landing checks, owner-local evidence rather than unbound saved-status reuse, cleanup grace, terminal check refusal without bisection, and rollout proof/staging. Main-only burn-in excludes --all; selected landing checks still owe real suite proof under --no-ci. No source changes, change-code, test runs or live VM actions during planning. Previous TTY capture remains unavailable; recovery uses the branch, sdlc state and issue evidence.

## Revisions

### 2026-10-20 22:45 PDT — Expand the initial design task into reviewable boundaries

The original single planning row (“Design the narrow schema and train/single-landing selection together, then obtain full-flow approval”) is superseded by the linked durable proposal and two genuine implementation review boundaries. Scope and Done-when remain unchanged; operator approval is pending.

### 2026-10-20 22:51 PDT — Operator approval

Authenticated hub receipt 15f0281c-9941-4e15-aee3-03d074fdd9d0 records operator approval of all seven choices, with external-runner cancellation grace30s instead of15s. Proceed to change-code and M1/M2. Rollout remains capability only; no Pair activation or hosted workflow changes. This supersedes the pending approval state above.
