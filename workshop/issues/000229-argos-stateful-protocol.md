---
id: 000229
status: open
deps: []
github_issue:
created: 2026-09-15
updated: 2026-09-15
estimate_hours:
---

# Design Argos stateful protocol binary

## Problem

The current `sdlc` binary manages repository lifecycle checkpoints, but it does
not manage a long-running backlog execution session. Small fixes need a higher
level process that can select and group suitable issues, assign them to a stable
worktree/Couch slot, advance one task at a time, recover after context loss, and
produce a durable session report.

The model must remain grounded in durable state rather than asking the LLM to
remember which task is current or infer whether a transition is legal. This is a
different layer from repository execution: Argos coordinates the session, while
the admitted task enters the normal `sdlc quick` flow. Argos must not reimplement
issue status, claims, verification, or close semantics.

## Spec

Design Argos as a **stateful protocol binary** and durable workflow coordinator
for LLM sessions. The LLM remains the reasoning and coding participant; Argos
owns the session state machine and chooses the next legal action. This is a
pull-based continuation protocol, not a second autonomous agent or harness.

### Protocol

The normal protocol has three verbs:

```text
argos next
argos done [--note ...] [--evidence ...]
argos fail --kind failed|blocked [--reason ...] [--evidence ...]
```

`argos next` is idempotent. It inspects authoritative session state and returns
the one action currently legal. If the LLM calls it repeatedly without submitting
the requested outcome, it returns the same action. If no session exists, it may
create one from the configured repository/worktree and enter candidate triage.

`argos done` and `argos fail` submit explicit outcome events for the currently
leased task. They validate and record the outcome, then return the next action in
the same response. A later `argos next` remains the recovery path when the LLM
crashes or loses context before sending the outcome. The protocol must not infer
completion merely because files changed or an issue appears edited.

`failed` and `blocked` are separate outcomes. A failed implementation attempt is
different from a task waiting on a dependency or operator decision. Additional
inspection and administrative verbs (`status`, `show-session`, `pause`, `abort`)
may exist, but normal state advancement stays behind `next` plus explicit
outcome events (ARCH-ORDER).

Every response should identify the session, phase, current task when applicable,
required action, expected artifact/evidence, and exact follow-up command. Provide
human-readable output and a stable `--json` envelope for tool composition. The
LLM normally does not need full `sdlc --help`; Argos returns the exact
`sdlc quick --issue N` command when execution begins.

### Session lifecycle and ownership

Argos owns a session state distinct from each issue's SDLC state:

```text
triaging → ready → executing → reporting → complete
    ↘ awaiting-operator       ↘ paused / blocked / aborted
```

`ready` means the session has the requested number of execution-ready tasks; it
does not merely mean that the LLM has named that many issue IDs. A session asking
for ten tasks may inspect many candidates and wait for operator decisions before
it can admit ten. `awaiting-operator` is a normal durable state, not an execution
failure.

Each admitted task has its own dispatch state (`proposed`, `admitted`, `active`,
`done`, `failed`, `blocked`, `skipped`). A session manifest records repository,
worktree/Couch slot, selected issue identities and order, current lease, retry
policy, and terminal outcomes. An append-only event log records transitions and
evidence references; a derived report summarizes them.

Argos should use stable issue identity and an explicit lease, not a free-form
session field as its only ownership mechanism. The issue remains the source of
truth for problem/spec/code and its `sdlc` status. Argos state records which
session selected it, whether an attempt is active, and what happened in this
session. CUE should validate Argos manifests/events and any durable triage
metadata. Do not duplicate the issue lifecycle in Argos (ARCH-DRY).

### Candidate triage and grouping

The human starts a session with a repository/worktree and a limit, for example:

```text
argos next   # session is created; candidate triage is requested
```

Argos returns a compact candidate inventory and a quantified rubric: repository,
issue identity/title, current status, dependencies, estimate when present, and
small-task constraints. The LLM reads relevant issue context, checks dependency
and correlation evidence, and classifies each candidate as suitable, unsuitable,
or needing an operator decision. For the last category, the LLM produces a
clarification queue covering behavior, scope, acceptance, dependencies, risk, and
other choices that would otherwise interrupt execution. The operator resolves
those questions; the LLM records the answers in the issue Spec/decision artifact,
and Argos validates that no required decision remains open.

Each candidate has a separate readiness lifecycle:

```text
unreviewed → inspected → rejected
                    ↘ clarification-needed → clarified → execution-ready → admitted
```

The session target is `ready_count >= requested_count`, not “requested count issue
IDs selected.” Argos locks the proposal only after the selected tasks satisfy the
readiness contract. The contract includes a stated outcome, scope/non-goals,
acceptance evidence, dependency disposition, verification path, and no unresolved
operator decision. The LLM supplies semantic judgment; Argos enforces the shape
and the explicit operator acknowledgement. It must never invent an answer to
move a task into `execution-ready`.

Groups are hypotheses recorded in the session, not merges of issue identity. A
group contains member issues, proposed root-cause relationship, confidence, and
execution order. Argos may split a group when implementation evidence disproves
the correlation.

### Execution and reconciliation

Once selection is locked, `argos next` returns an execution packet such as:

```text
session: 21
task: pair#22
worktree: ../worktree/pair/111
command: sdlc quick --issue 22
```

Argos coordinates admission and sequencing but does not code or reimplement the
`sdlc` flow. `sdlc quick` is the lower-level optimization for bounded work: one
issue, one compact plan, document updates as part of the same task, and one
verification/close boundary rather than unnecessary milestone decomposition.
Its design is separate from this issue but is the expected execution target.

`argos done`/`fail` must reconcile against repository evidence where possible:
issue status, observed revision, test/verification references, and expected
`sdlc` terminal state. A note is context, not proof. A repository artifact is
the work product; an Argos command is the explicit control event. The two must
agree before Argos advances (ARCH-PURE, ARCH-ORDER).

### Artifacts and final report

Use well-known locations for durable work products, while keeping explicit
commands for state transitions. A session may contain:

```text
session manifest     authoritative state and leases
event log            append-only transitions and evidence references
selection proposal   LLM triage and grouping rationale
session report       generated summary for the operator
```

`argos next` returns the expected artifact path and schema when it requests one.
The existence of a file alone does not advance state; `done` or `fail` is the
explicit acknowledgement that Argos validates. After the final task, `next`
returns a report-writing action. The LLM writes the report, signals `done`, and
Argos closes the session only when all admitted tasks have terminal outcomes and
the report is present.

The report should cover completed, failed, blocked, and skipped tasks; evidence
and revisions; unexpected findings; follow-up issues; and operator actions such
as a bug-bash agenda. Reports are derived from the event log rather than being
the sole source of session truth.

## Done when

- Argos has a reviewed state model and CUE-validated session/event schema.
- `next` is idempotent and returns one explicit legal action with a stable human
  and JSON protocol envelope.
- `done` and `fail --kind failed|blocked` record explicit outcomes, validate
  current ownership/evidence, and return the next action without requiring a
  second advancement call.
- Candidate triage is a first-class state machine: the LLM inspects candidates,
  surfaces quantified readiness questions, the operator resolves material product
  decisions, and Argos admits only execution-ready tasks. A request for ten means
  ten ready tasks, not ten unreviewed selections.
- Clarification questions, operator answers, readiness decisions, selection
  proposals, and grouping rationale are durable and validated; unanswered
  decisions keep a task in `awaiting-operator`/`clarification-needed`.
- Argos never duplicates or overrides `sdlc` issue lifecycle semantics; admitted
  execution returns an exact `sdlc quick` packet and reconciliation checks the
  repository's evidence.
- Session state survives LLM context loss, repeated `next` calls, process restart,
  stale leases, and a replaced worktree without fabricating progress.
- Final report generation is an explicit terminal phase with required sections,
  and session completion requires the report plus terminal outcomes for every
  admitted task.
- Protocol, CUE, and integration tests exercise repeated calls, failed/blocked
  outcomes, stale ownership, evidence mismatch, partial batches, and restart
  recovery through the production binary seam (ARCH-MOCK).

-

## Plan

- [ ] Map the existing issue/sdlc state and durable artifact conventions; define
  the exact Argos boundary and avoid duplicating repository lifecycle authority.
- [ ] Author and review the Argos session, task, lease, event, triage, and report
  vocabulary in CUE, including legal states and rejected transitions.
- [ ] Design the `next`/`done`/`fail` protocol and human/JSON response envelopes,
  including idempotent replay and context-loss recovery.
- [ ] Design the candidate/readiness state machine, quantified rubric, clarification
  queue, operator-answer record, and ten-ready-tasks admission rule for small
  unblocked single-repository work.
- [ ] Specify the session artifact locations, event log, evidence reconciliation,
  stable worktree/Couch-slot lease, and final report phase.
- [ ] Implement the smallest protocol slice and stateful fakes/tests before adding
  task execution integration; then connect the execution packet to `sdlc quick`.
- [ ] Add restart, stale-lease, failed/blocked, partial-batch, and evidence-mismatch
  tests; update atlas/help and verify through the SDLC review gates.

## Log

### 2026-09-15

Captured the design discussion for Argos. Core decision: Argos is a stateful
protocol binary whose `next` verb owns continuation and whose explicit `done` and
`fail` events submit the LLM's outcome; those commands return the next action so
the normal loop does not need a separate submit-then-advance call. Durable issue
and repository artifacts ground the session, while Argos owns session leases,
ordering, reconciliation, and reporting. No implementation changes made.

### 2026-09-15 — triage refinement

Refined the session design: triage is a first-class state machine before execution.
Argos must produce execution-ready tasks, not merely select issue IDs. The LLM
surfaces unresolved product decisions, the operator answers them, and the answers
become durable issue/decision context before Argos admits the task. The session
may remain `awaiting-operator` while it prepares the requested ready-task count.

## Revisions

### 2026-09-15 — Human-in-the-loop readiness triage

Reason: “pick ten tasks” should shift operator attention into a deliberate
preflight rather than interrupting execution later.

Delta: replaced the loose selecting/selected phase with `triaging → ready`, added
per-candidate readiness states and an `awaiting-operator` session state, and made
the admission target ten execution-ready tasks. Added clarification queues,
operator-confirmed decisions, and readiness invariants. No implementation changes.
