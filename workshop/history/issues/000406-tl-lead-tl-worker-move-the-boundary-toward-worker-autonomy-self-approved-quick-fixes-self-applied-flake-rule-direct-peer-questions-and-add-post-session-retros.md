---
id: 000406
status: done
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-16
estimate_hours:
card_mirror: '3e8e1efde793b84117cf20ef9db629e770e6cb65' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T10:17:23-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "b9f55260", done: "a044b6ed"}
actual_hours: 0.56
---

# tl-lead/tl-worker: move the boundary toward worker autonomy (self-approved quick fixes, self-applied flake rule, direct peer questions) and add post-session retros

## Problem

The TL is a router for messages that don't need a decision: small bug-fix plans, mechanical flake-rule application,
and worker-to-worker questions all wait on ops:0. The operator (2026-10-15): evolve the tl-lead / tl-worker split, hold
regular retros after long sessions, and agree to move the boundary so workers are more autonomous.

## Spec

- **Boundary move 1 (tl-worker):** quick-flow fixes inside the shell proceed after posting the plan to the hub
  ("proceeding unless you object"); only full-flow designs, design changes (a new state, refusal or mechanism) and
  scope revisions wait for approval. The third-delta rule (ariadne#404) still applies.
- **Boundary move 2 (tl-worker):** the flake rule is self-applied: file the flake issue with evidence, one unchanged
  rerun, report; a second unrelated flake on the same head means stop and ask.
- **Boundary move 3 (both):** peer questions (consumer review of an export, an interface question) go worker to worker
  directly, with the hub copied only on the outcome.
- **Retro (tl-lead):** after a long session (or at the operator's request), the TL writes a short retro to the
  operator: what routed through the TL that needn't have, what went wrong, and proposed boundary moves. Agreed moves
  are filed against tl-lead/tl-worker like these. The boundary lives only in tl-lead/tl-worker help text.
- Phrase-guarded in the help tests.

## Done when

- `tl-worker --help` states moves 1–3; `tl-lead --help` states the retro and that the boundary is changed only through
  agreed retro items; help tests guard the phrases.

- Worker help states the revised quick-fix approval and issue-backed flake
  rules, including required design/scope approval and second-flake escalation.
- Cross-repo findings route to the TL as one-line proposals for filing/folding;
  same-repo filings stay direct. The landing line owns landing order; reported
  hashes come only from Git output.
- Lead help dispatches design-heavy work (an invariant, a new mechanism,
  user-visible state) to Claude, and bounded execution against a fixed design
  to Codex. Phrase guards cover both cases.
- Guards cover required approval/escalation/agreement actions on both default
  and --help output; retros change the boundary only through agreed items.

## Plan

- [x] Add worker/lead phrase guards plus negative guards for retired approval
  and landing-order wording; observe expected failures.
- [x] Update both help surfaces and Atlas for the approved boundary; run targeted
  tests, push the implementation, and record the full-suite check.
- [x] Prepare close with verification evidence and the boundary review.

## Log

### 2026-10-16 — final recorded suite

The recorded suite on the final guard-test revision passed: 1,508 SDLC tests
across 12 shards plus 66 packages, zero failures, 906.2 seconds. Saved log:
.git/sdlc/local-checks/3c49114ab3713d6aaa13c4c3d82ac2ca6dc20be5/suite.log.
Close round 3 returned SHIP with both findings disposed. No help or test code
changed after this verified revision; later commits are acceptance/review records.


### 2026-10-16 — BR-1 verification
- 2026-10-16: closed — BR-2: Done when explicitly includes both approved dispatch cases and their phrase guards plus revised filing/landing/hash/action criteria; documentation only. BR-1 four independent reversals each fail default and --help guards; targeted packages pass. Recorded full suite passed 1508 SDLC tests and 66 packages; current guard-test revision has a recorded suite queued.; review verdict: SHIP

Independently reversed each of worker approval, TL approval, second-flake
escalation and retro agreement in the actual help source. Each mutation was
rejected by its corresponding guard on both default and --help paths. Restored
the unchanged help text; all tl-worker/tl-lead/tlrules tests and diff checks pass.
The correction changes tests only, not executable help behavior.


### 2026-10-16 — verification

Targeted `go test ./cmd/tl-worker ./cmd/tl-lead ./pkg/tlrules -count=1`
passed, including all new default/--help phrase guards and the retired central
approval/order guard. `git diff --check` passed. Recorded full suite via
`sdlc local-check run --name suite` passed: 1,508 SDLC tests across 12 shards
plus 66 other packages, zero failures, 812.2 seconds execution after 617 seconds
queue wait. PR #268; no operator smoke needed for these help-only changes.


### 2026-10-16 — implementation

All new worker/lead phrase guards failed before implementation on both default
and --help output. The negative guard also caught all three retired central
approval/order phrases. Updated both sides of the boundary, preserving the
shared planning source. The hash rule already existed; strengthen its wording
and add a phrase guard instead of duplicating it. Atlas maps the revised roles.


### 2026-10-15

## Revisions

### 2026-10-16 — operator-approved boundary, relayed by ops:0

- Approve self-applied flake handling and self-approved small fix plans within
  the quick-flow shell; retain escalation for full-flow designs, design changes
  and scope revisions.
- Landing order belongs to the landing line; workers follow its admission order.
- Replace move 3: cross-repo findings go to the TL as a one-line proposal. The TL
  files or folds them to retain the cross-repo picture. Same-repo filings stay
  direct. This supersedes the original peer-question boundary move above; workers
  do not file peer-repo issues directly.
- Report commit hashes only from Git output, never typed from memory; retain and
  guard the existing worker rule using `git rev-parse`. Two overnight typed-hash
  slips motivated this explicit requirement.
- Preserve the post-session retro requirement. Source: verified Couch delivery
  a98cba8c-06ec-4136-af9e-f797d7bd5418 from ops:0.

### 2026-10-16 — implementation design and acceptance clarification

Use the existing embedded command help and phrase-test mechanisms; shared
planning prose continues to derive from pkg/tlrules (ARCH-DRY). Worker-owned
procedures stay in worker help; lead help cites those and states the TL's side.
No runtime state, IO, trust boundary, dependency or durable artifact family is
introduced (ARCH-PURE, ARCH-MOCK, ARCH-SECURE, ARCH-ORDER, ARCH-FUNERAL).
Workload is help output with unchanged size order (ARCH-CONSTRAINTS).
ARCH-PURPOSE: update contradictory TL execution/plan-review wording as well as
worker permissions, tests, filing and landing text. Keep operator smoke and
full-flow/design/scope approvals; the third-delta planning tripwire stays.

The Done-when phrase "moves 1–3" now means the operator-approved Revision:
self-approved small fix plans after posting to the hub; self-applied issue-backed
flake handling with one unchanged rerun and escalation on a second unrelated
flake on the same head; landing-line ownership of order; cross-repo findings as
one-line proposals for the TL to file or fold, with same-repo filings direct.
Guard Git-output hashes (the existing git rev-parse rule), session retros and
agreed-rule-change ownership. Do not enable direct peer-repo filings or retain
the superseded original peer-question boundary move.

### 2026-10-16 — dispatch by kind of work

Operator decision relayed in verified ops:0 delivery
3114a443-08d3-4a92-a549-7f6a3edd839d: tl-lead dispatches design-heavy items
(an invariant, a new mechanism, user-visible state) to a Claude slot; bounded
execution against a fixed design goes to Codex. Add phrase guards alongside the
boundary rules. This extends the acceptance contract for the lead help.

### 2026-10-16 — BR-1: guard required actions

The first boundary review found the same test gap in four clauses: worker
approval, TL approval, second-flake escalation, and retro agreement. Extend
existing phrase guards to include the required actions/restrictions, and mutate
each clause independently to prove removal or reversal fails both invocation
paths. The help behavior and approved scope stay unchanged.

### 2026-10-16 — BR-2: synchronize acceptance criteria

Add the already-approved Revision requirements explicitly to Done when,
including both work-kind dispatch cases and their phrase guards. Also spell
out the revised filing, landing, hash and required-action criteria rather than
leaving them implicit behind the old "moves 1–3" reference. This changes only
acceptance documentation; help behavior and passing regression guards are
unchanged. BR-1 is disposed as addressed by round 2.
