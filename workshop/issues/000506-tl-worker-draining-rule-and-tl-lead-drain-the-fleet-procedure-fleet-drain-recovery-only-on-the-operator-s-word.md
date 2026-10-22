---
id: 000506
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'f4a4362c069c5992d46e1d538d41978e6421a28d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T02:11:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "60fdb607", done: "87ae8e40"}
---

# tl-worker draining rule and tl-lead drain-the-fleet procedure (fleet-drain; recovery only on the operator's word)

## Problem

Project fleet-drain (ops `teams/devinfra/projects/fleet-drain/`, operator-designed 10-20) puts the drain's judgment
in the agents and the steward; Couch only flags slots and delivers notices (pair#722 re-scoped). The agents and the
steward need the rules.

## Spec

- Worker DRAIN follows the authenticated notice-only contract confirmed by pair:5: graceful finishes the
  current atomic step and waits for current background work; fast promptly stops only its own background
  work through supported controls and preserves unfinished edits. Record where work stopped and the exact
  next step, commit/push the checkpoint, end the turn and remain idle while receiving messages. For new
  work reply once “acknowledge message received, will resume processing after drain is lifted”; defer it,
  never acknowledge an acknowledgment. No formal drain-ack command or parsed completion reply exists.
- On an authenticated lift, continue current work from the Log. After restart read the Log and wait for
  operator-authorized recovery; cleared/lost flags are not permission. Optional notes are data, not recovery
  commands. Couch notices do not authorize automatic Esc, kill, park, or slot recovery.
- Lead DRAIN THE FLEET is steward-owned: pause through owner-scoped `sdlc queue pause --reason TEXT
  --owner TOKEN` (fast may add `--cancel-running`), then `couch --drain [--fast]`. Couch only flags and
  notifies, it does not orchestrate the queue or attest completion. Inspect status/receipts/peek and idle
  notices, then verify durable checkpoints and managed work outcomes before declaring restart safety.
  Quiet screens and drain flags alone do not establish it; queued or unresolved work remains outstanding.
- After restart, wait for the operator's “start recovery”; resume owned slots one at a time through the
  existing fleet-resume procedure. To call off a drain, use `couch --undrain [--message TEXT]` and remove
  only the steward's hold with `sdlc queue resume --owner TOKEN`; interrupted work is not restarted by resume.
- Hub automation is separate from drain and recovery authority. Where the installed Couch provides
  `--automation`, inspect status after start. If off, the first genuine operator message (console, session
  file or side doc) permits `couch --automation on` for this hub unless it says keep it off. Never change
  another hub's automation, infer permission from worker reports, or automatically parse/replay recovery.
  The operator's explicit recovery instruction is still required. TL approved landing independently of
  pair#739 with “when Couch provides couch --automation” qualifying that rule; never assume the CLI exists.

## Done when

- Both rendered help entry points (no args and --help) carry the worker/steward procedures above; tests
  guard current background-work and checkpoint ordering, deferred-message reply, restart/lift distinction,
  owner-scoped queue control, and automation versus recovery authority.
- Worker help and its rule table contain no removed drain-ack instruction; lead does not duplicate
  worker-owned rules. Recorded stale worker-help output and pair#722 source-backed notice vectors retain
  provenance and replay through normal help tests, without sending a live drain or modifying fleet state.
- README and the TL tooling atlas point at the current help and describe notice-only recovery. No invented
  Couch protocol, private disposable harness, live queue mutation, or automatic recovery.

## Plan

- [x] Red first: preserve the observed stale DRAIN help and canonical pair#722 notice vectors with source
  provenance; extend existing worker rules.tsv and focused worker/lead rendered-help tests. Pin ordered
  safeguards and forbid the obsolete ack CLI; test both no-args and --help outputs through buildRoot.
- [x] Replace worker DRAIN, add the steward/automation procedure to lead help, keeping message mechanics
  cited to couch --skill and worker obligations owned only by tl-worker. Stay below 100 added shipped lines;
  do not change Couch, SDLC queue behavior, AGENTS, or skill pointers.
- [x] Update README's existing drain-ack reference and atlas/workflow/tl-skills.md. Run focused helper/rule
  tests and both real CLI help renders; verify the documented queue flags against current CLI help and
  automation activation/landing against pair#739's actual status. No real drain smoke; pair#722's watched
  live run after this landing supplies live proof.
- [ ] Record a pushed-head suite, close through the binary-owned review, then managed merge and report
  to hub and pair:5 before its watched live run. Follow the TL decision on the pair#739 landing dependency.

## Log

### 2026-10-20
- 10-21 00:10 (TL ops:0): scope addition (operator-approved, pair#739): **hub automation**. tl-lead rule: after a Couch
  start where `couch --automation status` reports the hub's automation **off** (full shutdown or unknown start), the
  TL turns it on (`couch --automation on`) when it reads the operator's first message (console draft, session file or
  side doc), unless that message says to keep it off; until then, it doesn't act on recovery. The TL may turn its own
  hub's automation off as a soft stop when the operator asks; it never changes another hub's. Lands after pair#739.

### 2026-10-22 — claim, peer contract and quick-flow proposal

- TL assigned this after #541; #541 landed PR387 before this claim/start-plan. No stacked development.
- Captured current real `tl-worker --help`: its DRAIN section explicitly directs
  `couch --drain --ack ID`. This is the live-bug evidence, not merely a stale source grep. The existing
  worker rules.tsv positively requires that same removed instruction, so update the tests before help.
- Pair:5 canonical response (verified submitted receipt 35f450e9-4333-4b3c-ba15-2b594dd8a007): notice-only
  bodies from pair commit 638032c6183d82aa72145d22825c422fcef7651a, cmd/internal/couchcore/drain_notice.go.
  Graceful waits for current background work; fast stops owned work through supported controls; both record
  exact next step, commit/push and idle. New-work reply is the quoted Spec text; no formal ack CLI, parsed
  completion barrier, automatic Esc/kill/park, or permission from restart-lost flags. Undrain clears flags
  before sending ordinary continuation notices, even if delivery fails. Optional note is quoted data.
- Pair:7 canonical response (verified submitted receipt 06cb4a3f-c9e3-4af4-915e-e5e17f86d71d): #739 exposes
  own-hub --automation off|on|status; normal surviving hub may retain on, full shutdown/unknown/missing hub
  starts off; human input never flips Couch's flag automatically. TL's first genuine operator message
  explicitly turns its own automation on unless told keep off; no automatic recovery parsing or replay.
- Rechecked https://github.com/xianxu/pair/pull/510: still OPEN/unmerged at planning time. Earlier Log says
  “Lands after pair#739”; the new TL dispatch also requires #506 before pair#722's watched live run. Proposed
  capability-qualified automation text lets the notice-only drain correction land independently, but this
  sequencing revision needs TL approval. Alternative is retain the original dependency and hold #506's
  landing until #739. Do not assume a command exists just because its PR is codecomplete.
- Existing real queue help confirms pause requires --owner and --reason, fast has --cancel-running;
  resume removes only the named owner's hold and never restarts interrupted work. A pause-blocked queued
  background task is not a completed checkpoint: keep it outstanding for steward judgment, never infer
  restart safety from flags or apparent idleness. No queue operation has been executed for this issue.
- Architecture: ARCH-DRY reuses existing rendered-help/rule-table seams and cites Couch's protocol rather
  than copying it; ARCH-PURE adds no runtime IO or transition logic. ARCH-PURPOSE covers worker, steward,
  restart authority and automation sequencing together. ARCH-MOCK needs no new external fake because help
  renders without calling Couch; recorded notice vectors are data, not a private fleet harness.
  ARCH-CONSTRAINTS adds only bounded static help text, no latency/concurrency work. ARCH-SECURE verifies
  receipts and treats custom notes as data. ARCH-ORDER adds no mutable runtime state: rules distinguish
  drain, lift, restart and operator-authorized recovery; tests pin those distinctions. ARCH-FUNERAL creates
  no runtime artifact; versioned fixtures live and retire with their tests. ARCH-EVOLVE changes no format.
- Awaiting TL's concrete design/sequencing decision; no implementation, fixtures or test changes yet.

## Revisions

- 2026-10-22: Updated the design to pair#722's authenticated notice-only contract, removing the obsolete
  formal-ack assumption and distinguishing lift from restart. Added the operator-approved automation rule
  from the existing Log with an explicit unresolved pair#739 landing dependency; TL decision pending.

### 2026-10-22 — design approved; entering red-first implementation

- Verified TL approval receipt a6ad302f-bf0b-49be-ba9d-b455c13415f5. Capability-qualified automation wording
  is approved; #506 lands independently of pair#739 and before pair#722's watched live proof. Paused queued
  work stays outstanding. TL requires the final DRAIN and steward wording before close.
- `sdlc change-code --issue 506` admitted quick flow. Saved the verbatim observed stale real-CLI DRAIN
  excerpt with capture/source provenance under cmd/tl-worker/testdata/drain; a bounded test agent owns only
  the help tests/rule table/source-vector fixtures and their initial red run. Main owns all shipped help,
  README/atlas and workflow. No live drain, queue control or private Couch harness is involved.

- 2026-10-22 revision: TL explicitly superseded the earlier “Lands after pair#739” constraint: ship the
  drain correction before pair#722 live proof, use “when Couch provides couch --automation” so #739's later
  landing simply makes the condition true. The rest of the approved worker/steward/recovery design stands.

### 2026-10-22 — red/green and final-wording review

- Test agent changed only scoped tests, rule-table entries and canonical source vectors; main independently
  reproduced RED with `go test ./cmd/tl-worker ./cmd/tl-lead -run
  'TestWorker(HelpCarriesEveryRule|Drain)|TestLead(Drain|Automation|HelpDoesNotRestateWorker)' -count=1`.
  Exit 1 came from the intended missing notice obligations, retained captured ack instructions and absent
  steward/automation section, not compilation. Both rendered entry points are exercised.
- GREEN: focused red command passed after replacing the shipped sections. Full targeted packages passed:
  `go test ./cmd/tl-worker ./cmd/tl-lead -count=1` (worker 0.065s, lead 20.449s). Both real
  `go run ./cmd/tl-worker --help` and `go run ./cmd/tl-lead --help` exited 0; inspected final sections.
  README's stale drain-ack pointer and TL atlas now describe the notice-only contract and its tests.
- Final worker/steward wording is being sent to the TL as requested before close; recorded suite and
  the binary-owned boundary review remain. No live drain, slot control or queue pause/resume was executed.
