---
id: 000481
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '695b9162170d7c0612c40b1bd599b52948d802e1' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc queue pause/resume (+ --cancel-running) and the tl-worker drain rule, for couch --drain

## Problem

A fleet drain before a machine or Couch restart (pair counterpart: `couch --drain [--fast]`) needs the machine test
queue and the landing lines to stop admitting work, and workers need a rule for the drain notice. Operator item 15
(10-19), discussion: ops `teams/devinfra/tl-status-xian-xu.side/2026-10-19-15-fleet-drain.md`.

## Spec

- `sdlc queue pause [--reason TEXT] [--cancel-running]` / `sdlc queue resume` (operator shell or a :0 hub only, the
  same rule as grants): while paused, the machine test queue and every landing line admit nothing new; waiters stay
  queued and print the pause and its reason. Running suites finish, or, with `--cancel-running` (used by
  `couch --drain --fast`), are cancelled and recorded as interrupted, rerunnable on the same head, never as failures.
- A merge interrupted by the pause resumes on `sdlc merge --yes` after resume, with no new suite needed when its
  head is unchanged and the suite record exists.
- tl-worker rule: on a drain notice, graceful means finish the current atomic step, push, write a Log line saying
  where you stopped, report "drained" to the hub and stay idle; fast means stop now, write the Log line, report.


## Contract (frozen by pair#722's approved plan, 10-19; copied verbatim from pair `workshop/plans/000722-fleet-drain-plan.md` at ca8135bb)

This supersedes the Spec's first bullet where they differ: pause and resume take `--owner TOKEN` (owner-scoped), and `status --json` follows the schema below. Build against it; conformance with pair#722's fake executable adapter is the acceptance.

### Notice text (v1, owned by pair#722's formatter)

The dynamic ID is a canonical UUID; render optional reason as quoted data on its own labelled line. No reason text becomes an argv or an instruction. The formatter is the single production text owner; skill/help describe the contract rather than duplicating the full template.

Graceful:

> Couch drain v1 {ID}: checkpoint and stop. Finish the current atomic step; commit and push the checkpoint; write where you stopped in the issue Log. Start nothing new. Run couch --drain --ack {ID} after the checkpoint is durable, then end your turn and stay idle. Do not send a separate drained report.

Fast:

> Couch drain v1 {ID} (fast): stop now and checkpoint. Do not finish the current step. Write where you stopped in the issue Log; commit and push the checkpoint without discarding unfinished working-tree edits. Start nothing new. Run couch --drain --ack {ID} after the checkpoint is durable, then end your turn and stay idle. Do not send a separate drained report.

These are supervisor-origin notices through Broker.SendSystem with authenticated receipt lookup. A delivery receipt is not an ack. Old or unsupported wrappers remain stuck with a relaunch-needed reason; drain does not relaunch them itself. Exactly one notice is admitted per drain/current incarnation; an uncertain transport result is inspected, not blindly pasted again.

### The sdlc seam (frozen)

Run argv arrays, never a shell. The queue adapter runs from the verified steward workspace and does not manufacture another slot's environment/authority. Owner token is `couch-drain:{ID}`; create and persist it before the first external pause attempt.

```
["sdlc", "queue", "pause", "--reason", "Couch drain {ID}: {reason}", "--owner", "couch-drain:{ID}"]
# fast appends one argument:
"--cancel-running"
["sdlc", "queue", "resume", "--owner", "couch-drain:{ID}"]
["sdlc", "queue", "status", "--json"]
```

Pause/resume need only conventional exit status; every successful or uncertain mutation is reconciled through status. Pause idempotently adds that owner; resume idempotently removes that owner and no other. Multiple owners compose: admission stays paused while any owner remains. Fast cancellation records interrupted work, not failed work; owner removal never restarts it. There is no TTL that silently resumes a live pause. The #481 worker rule uses the notice text and ack action above.

Exact required status schema, lower_snake_case JSON keys:

```json
{
  "schema_version": 1,
  "state": "paused",
  "complete": true,
  "pause_owners": [
    {"owner": "couch-drain:00000000-0000-4000-8000-000000000001", "reason": "Couch drain ...", "cancel_running": false}
  ],
  "active": [
    {"id": "opaque-stable-operation-id", "workspace": "pair:5", "kind": "suite", "state": "running", "head": "optional-git-sha"}
  ],
  "detail": ""
}
```

`state` is `running`, `paused` or `unknown`; `kind` is `suite`, `merge` or `landing`; active entry `state` is `running`, `interrupting` or `unknown`. All fields shown are required except `head`. `complete:false` or `state:unknown` means absence cannot be inferred. `detail` explains incomplete/unknown observations. `running` requires zero pause owners; `paused` requires at least one. Entries with unknown workspace use `workspace:""` and block global quiescence. Waiters held before admission are not active. An operation may own several distinct kinds, each represented; do not identify work by PID alone. A complete empty `active` list is authoritative quiescence only at the captured observation; Couch refreshes it while the board is active.

Decoder rejects duplicate keys, malformed required types, unsupported versions, invalid enum combinations and oversized/truncated output. Tolerate unknown additive fields read-only. Limits proposed: 1MiB output, 128 pause owners, 4096 active entries, bounded text fields; overflow becomes unknown/incomplete, never truncated success. Couch performs no PID killing or registry-file edits.

`FakeDrainQueue` is a stateful implementation of the same pause/resume/status interface: owner map, FIFO waiters and exact operation IDs with workspace/kind/head/state; controllable finish, interruption and observation failure. Pause blocks admission, graceful retains running work, fast marks interrupting then explicitly interrupted/removed, resume releases only one owner's hold. Lost replies can occur after an effect; status still sees the applied state. Include a fake executable adapter test that validates exact argv/JSON, in addition to in-process model tests. Live conformance runs in #481's isolated portable queue state root, never against the operator's machine queue during tests.

## Done when

- Tests: pause blocks admission in both the queue and the landing line; resume releases in FIFO order;
  `--cancel-running` records interrupted (not failed) and a same-head rerun is accepted.
- `tl-worker --help` carries the drain rule.

## Plan

- [ ]

## Log

### 2026-10-19
