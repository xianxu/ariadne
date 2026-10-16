---
id: 000404
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-16
estimate_hours:
card_mirror: 'bf9703714c7f6ce3b1ff137c57818f977e50b541' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T09:16:33-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "89e3531d", done: "6fb4de37"}
---

# tl-lead/tl-worker: a third design delta on one mechanism triggers a re-plan; invariants weigh a lock-shaped design first

## Problem

2026-10-15: the TL approved eight design deltas to pair#584, one at a time, each reasonable alone (argv for every OS
process, legacy rows, unbound launches, ledger caps, codec speed, concurrency), without asking why one guard needed
eight. The guard (pair#556) proves "no live writer" by forensics over historical state, about 1,700 lines with ~45
refusal reasons; a wrapper-held flock that dies with its process does the same in one step with two outcomes
(pair#595). The operator: "the same old over engineering." #556's plan never weighed a lock.

## Spec

- tl-lead's plan-review rules gain a **delta-count tripwire:** when a slot asks for a third design delta on the same
  mechanism, stop approving deltas; ask for a re-plan that names the simplest alternative and why it fails.
- tl-lead's and tl-worker's plan rules gain one question for any invariant ("no two X", "never while Y"): **what
  single resource, held for the owner's lifetime (a lock or lease), would make this a one-step check?** A plan that
  proves absence by observation must say why that resource can't work.
- Phrase-guarded in the help tests.

## Done when

- `tl-lead --help` and `tl-worker --help` carry both rules; help tests guard the phrases.

## Plan

- [x] Extend existing help phrase guards for both commands; observe failure.
- [ ] Add role-specific planning prose for the third-delta tripwire and the
  owner-lifetime lock/lease question; run targeted guards and the full suite.
- [ ] Close with the boundary review and land on SHIP.

## Log

### 2026-10-15

### 2026-10-16 — design

Quick-flow help-text change. ARCH-DRY: retain the existing embedded help and
phrase-test mechanisms; each role states its action (TL stops approving, worker
stops requesting incremental approval). ARCH-PURE: no runtime IO changes.
ARCH-PURPOSE: guard both rules on both help surfaces, including why a lock/lease
cannot work before absence-by-observation designs. ARCH-MOCK, ARCH-CONSTRAINTS,
ARCH-SECURE: no external dependency, workload or trust-boundary changes.
ARCH-ORDER: help holds no state between events; it instructs actors about the
third-delta boundary. ARCH-FUNERAL: creates no new durable artifact family;
issue and review records follow existing archive lifecycle.

Plan sent to ops:0 for approval; #406's requested Revision published separately
before implementation. No operator smoke needed for help-only wording.

### 2026-10-16 — implementation verification

ops:0 approved the short plan via verified delivery
7218f7f2-b1b4-45fc-9d1b-96d8c3474575. Existing help tests failed on all four
new phrases on both default and --help paths before implementation. After adding
the planning section, `go test ./cmd/tl-worker ./cmd/tl-lead ./pkg/tlrules
-count=1` passed; `git diff --check` passed. Full `make test` running in the
background; log /tmp/ariadne-404-suite.log. Atlas maps the new planning rules.
