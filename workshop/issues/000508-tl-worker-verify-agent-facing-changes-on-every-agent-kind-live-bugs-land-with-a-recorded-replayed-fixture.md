---
id: 000508
status: codecomplete
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'fadcb69b89743263760eacf58f8bfa6152870122' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T01:53:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "5290755a", done: "a7c5224f"}
actual_hours: 0.33
---

# tl-worker: verify agent-facing changes on every agent kind; live bugs land with a recorded replayed fixture

## Problem

Live-only bugs on 10-20 (pair#727, #735, #742, the drain) were each found by hand, and pair#717's live check passed
on the agent it was built for and failed on Codex. Operator decisions (ops side doc
`2026-10-20-19-conformance-structure.md`): every agent kind, every time; a live bug's fix lands with its recorded
capture as a replayed fixture.

## Spec

Add to `tl-worker --help` (verification section):
- An agent-facing change is verified against **every agent kind** (Claude and Codex; Qoder when it has credits), in
  its tests and in any live check.
- A bug found live is fixed together with a **recorded fixture** of the real capture (terminal bytes, agent screen,
  tool reply) replayed in the normal suite, so it can't regress unnoticed.
- Live tests use the shared disposable-Couch harness (pair#744), never the live fleet.

## Done when

- `tl-worker --help` carries the three rules; a test pins the section.

## Plan

Quick flow, operator-reviewed Spec unchanged. Put all three rules at the start
of worker TESTS: every agent kind in tests/live checks (Claude, Codex, Qoder
when credited), real live capture replayed in the normal suite with the fix,
and live tests isolated through the shared disposable-Couch harness. Extend the
existing worker rule table and rendered section tests at --help/no-args. The
lead continues citing worker ownership, rather than duplicating these rules.

ARCH-DRY/PURE: existing embedded help, composition and rule-table tests remain
the only mechanisms; no new IO. ARCH-PURPOSE: all three approved rules and
all named capture forms stay explicit. Other architecture lenses introduce
nothing here: no runtime state, external calls, credentials, new persisted
formats, growing artifact family or hot-path work; static help dies with its
command invocation. No state table or new process fixture is needed for a
rule-only change. The future bug fixes own their recorded captures.

- [x] Red rendered-section regression and worker rule-table rows.
- [x] Three rules in worker help; atlas map and focused verification.

## Log

- 2026-10-21: the first recorded branch suite passed all 1,888 SDLC tests and
  75 other packages with zero failures. Landing then found an append conflict
  in the worker rule table with main's checkpoint/restart rules (#525).
  Resolved by retaining both complete row sets; the rendered help and atlas
  merged without conflicts. All worker, lead and shared rule controls pass
  after resolution (0.134s, 15.364s, 0.127s). A fresh recorded suite and narrow
  re-close will verify the integrated generation before another landing entry.

- 2026-10-21: closed — Red first: both rendered entry points lacked all three TESTS rules (0.177s). Worker help now requires every agent kind in tests and live checks (Claude/Codex, Qoder when credited), a recorded real live capture replayed with its fix in the normal suite, and reuse of the shared disposable-Couch harness instead of live fleet/private Couch fixtures. All worker tests pass (latest 0.127s), all lead tests pass (17.095s), shared tlrules tests pass (0.124s); rule-table ownership guard remains green. Atlas updated, diff clean. Pair#744 freshly verified codecomplete/not-landed; exact TL-approved placeholder retained. Restart audit found no own queue/landing/run lease; PR 372, pushed head and local HEAD agree. First recorded suite requested under authorized review overlap; final pass remains required for landing. This changes help instructions only and drives no live fleet or harness process. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
### 2026-10-20

- 2026-10-21: claimed after #517 landed; existing Spec is the operator-reviewed
  design. Current TESTS and LIVE CHECKS do not state the three conformance
  requirements. Small plan posted to hub, proceeding unless objected. No change
  to the agent inventory, Couch harness implementation or live fleet.

- 2026-10-21: red first on all missing rule-table rows and TESTS clauses at
  both entry points (0.177s). After help update, all tl-worker tests pass
  (0.132s), tl-lead tests pass (17.095s), and tlrules tests pass (0.124s).
  Lead ownership guard remains green. Atlas map updated; diff check clean.
  This is a rendered instruction contract, not a runtime fixture implementation;
  no new mechanical API misuse class or Go lint rule is introduced.

## Revisions

- 2026-10-21 01:55 PDT — TL clarified harness provenance: while pair#744 is
  unlanded, use the exact shared disposable-Couch harness (pair#744) wording
  and avoid private Couch fixtures. Its tracker was working/not-landed when
  checked. Name cmd/internal/testcouch and make test-couch-fleet after adoption
  lands; no unlanded command is advertised by this patch.
