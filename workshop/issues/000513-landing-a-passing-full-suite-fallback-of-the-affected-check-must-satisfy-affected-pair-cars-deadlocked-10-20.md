---
id: 000513
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'f211b8fb282547add4a3cfa348845220918c850e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T20:07:43-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "db4bd15b", done: "260845ad"}
---

# landing: a passing full-suite fallback of the affected check must satisfy 'affected' (pair cars deadlocked 10-20)

## Problem

On 10-20 pair's affected check fell back to the full suite on fresh checkouts (the runtime bundle isn't generated,
so the Go graph can't load; pair#738), and the fallback **passed** but posted only `sdlc-local/suite`. `sdlc merge`
then refuses readiness forever: "PR #489 needs recorded local checks on its head: affected". pair#732 (and pair#706)
deadlocked with all checks green; no pair landing for over two hours.

## Spec

The readiness gate already accepts a passing required suite before considering its affected stand-in. Preserve that behavior and its same-head, failure and docs-only ancestry safety. Clarify a missing affected record as waiting for affected OR its full-suite fallback, with a conditional instruction to wait if affected fell back. The gate has no evidence of a queued run or its provenance, so it must not assert that fallback actually occurred. No new statuses, records or runtime policy.

## Done when

- Production car controls pin that a passing suite on the same head satisfies readiness, a failed suite does not, and a suite before a code change does not.
- Missing-record diagnostics name both affected and its full-suite fallback, explain when to wait, and retain the affected command for a missing run.
- Existing fallback execution continues recording suite rather than a fabricated affected result.

## Plan

- [ ] Pin current readiness behavior and red-test the misleading diagnostic.
- [ ] Clarify only the message and map it in the atlas.
- [ ] Run focused controls and prepare recorded-suite/close requests; require passing results before landing.

## Log

### 2026-10-20

Claimed and started after #505 landed. Source confirms carCIGateFacts maps required suite to affected; decideCIGate first checks the suite’s own passing record, so full-suite success already satisfies the car. Existing TestDecideCIGateLocalChecks pins that case. runAffectedCheck intentionally runs/records suite for full selection. Quick message-only plan posted to hub, proceeding unless objected. ARCH-PURPOSE: diagnose the actual missing-evidence wait rather than alter a correct gate; ARCH-DRY: keep the diagnostic in ciVerdict.refusal.

## Revisions

### 2026-10-20T20:09-07:00 — correct the reported deadlock to a diagnostic issue

TL correction (Couch 7f902972-6ced-43f5-8f92-9b080625666b, previously authenticated) reports pair#732/#706 became ready when the same-head fallback passed. Code inspection agrees: the missing-affected text was shown while the suite was queued/running, not after its pass. Scope shrinks to truthful diagnostics and controls; retain existing gate and record semantics. No fallback-in-progress assertion without provenance; the message must be conditional. Original acceptance/design follows for the record:

> ## Spec
>
> - When the affected check falls back to the full suite and that suite passes on the same head, record the affected
>   check as satisfied (with "fallback: full suite" in its evidence): the full suite is a superset of any affected set.
> - A failing fallback records affected as failed, as now.
> - `sdlc landing status` names the fallback when it applies.
>
> ## Done when
>
> - Tests pin: fallback pass satisfies the affected requirement on that head; fallback fail does not; a suite on a
>   different head does not.
> - pair#732-style car (fresh checkout, fallback) becomes ready without manual recovery.
