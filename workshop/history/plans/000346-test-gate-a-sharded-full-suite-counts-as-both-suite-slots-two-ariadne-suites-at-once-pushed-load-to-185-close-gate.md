---
gate: boundary-review
issue: 346
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T13:55:38-07:00"
      dispatched: "2026-10-13T13:53:44-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: Makefile.workflow comment says WF_TEST_WEIGHT := N; the contract is ?= (env override must win)
          detail: Makefile.workflow:16 says `:=`, while Makefile:10 and the atlas use `?=`, and the issue Log records that `:=` was rejected because it beats an operator's environment override. Change the comment to `?=`. Those are all the places in the window that state the declaration form; only Makefile.workflow:16 is wrong.
          family: doc-claim-drifts-from-code
          round: 1
        - id: BR-2
          severity: Minor
          title: AGENTS.base.md:62 machine test budget still says "at most 2 suites; a third waits", which ignores weights
          detail: 'With WF_TEST_WEIGHT, ariadne''s sharded suite takes both slots: it waits behind any single suite and blocks any second one. This agent-facing summary propagates to dependent repos; add one clause about weighted suites (ARCH-PURPOSE shadow sweep).'
          family: doc-claim-drifts-from-code
          round: 1
        - id: BR-3
          severity: Minor
          title: '"Light suites cannot starve a heavy waiter" overstates it; only the queue head is protected'
          detail: 'The queue is a polled flock, not first-in-first-out, so a heavy waiter that is not yet head races light newcomers for the queue each time it frees. The claim appears in two places: the atlas "One queue head" bullet and the test-lock.py header. Reword both as "the head waiter is never overtaken by light suites".'
          family: doc-claim-drifts-from-code
          round: 1
        - id: BR-4
          severity: Minor
          title: test-lock.test.sh does not unset WF_TEST_WEIGHT, so the make-wiring case fails under an operator's env override
          detail: The script already unsets WF_TEST_LOCK, WF_TEST_LOCK_TIMEOUT and WF_TEST_LOCK_HELD_BY; add WF_TEST_WEIGHT to that line so "make test waits for both slots" does not depend on the caller's environment.
          family: test-env-hermeticity
          round: 1
      recipe: small-diff-review
      isolation: checkout
      reviewed: 5b2e260be969a79d7dbb70263b4c8719ca9633c0
      blocked: false
---

# Gate ledger — ariadne#346 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T13:55:38-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `doc-claim-drifts-from-code` Makefile.workflow comment says WF_TEST_WEIGHT := N; the contract is ?= (env override must win)
  Makefile.workflow:16 says `:=`, while Makefile:10 and the atlas use `?=`, and the issue Log records that `:=` was rejected because it beats an operator's environment override. Change the comment to `?=`. Those are all the places in the window that state the declaration form; only Makefile.workflow:16 is wrong.
- **BR-2** [Minor] `doc-claim-drifts-from-code` AGENTS.base.md:62 machine test budget still says "at most 2 suites; a third waits", which ignores weights
  With WF_TEST_WEIGHT, ariadne's sharded suite takes both slots: it waits behind any single suite and blocks any second one. This agent-facing summary propagates to dependent repos; add one clause about weighted suites (ARCH-PURPOSE shadow sweep).
- **BR-3** [Minor] `doc-claim-drifts-from-code` "Light suites cannot starve a heavy waiter" overstates it; only the queue head is protected
  The queue is a polled flock, not first-in-first-out, so a heavy waiter that is not yet head races light newcomers for the queue each time it frees. The claim appears in two places: the atlas "One queue head" bullet and the test-lock.py header. Reword both as "the head waiter is never overtaken by light suites".
- **BR-4** [Minor] `test-env-hermeticity` test-lock.test.sh does not unset WF_TEST_WEIGHT, so the make-wiring case fails under an operator's env override
  The script already unsets WF_TEST_LOCK, WF_TEST_LOCK_TIMEOUT and WF_TEST_LOCK_HELD_BY; add WF_TEST_WEIGHT to that line so "make test waits for both slots" does not depend on the caller's environment.

## Open findings

- **BR-1** [Minor] `doc-claim-drifts-from-code` Makefile.workflow comment says WF_TEST_WEIGHT := N; the contract is ?= (env override must win)
- **BR-2** [Minor] `doc-claim-drifts-from-code` AGENTS.base.md:62 machine test budget still says "at most 2 suites; a third waits", which ignores weights
- **BR-3** [Minor] `doc-claim-drifts-from-code` "Light suites cannot starve a heavy waiter" overstates it; only the queue head is protected
- **BR-4** [Minor] `test-env-hermeticity` test-lock.test.sh does not unset WF_TEST_WEIGHT, so the make-wiring case fails under an operator's env override
