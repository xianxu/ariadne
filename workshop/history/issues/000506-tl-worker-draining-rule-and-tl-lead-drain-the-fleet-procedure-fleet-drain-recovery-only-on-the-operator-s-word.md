---
id: 000506
status: done
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '148d66dfdaa8fe909c661c381bb3192d43b7021b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T02:11:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "60fdb607", done: "87ae8e40"}
actual_hours: 0.95
---

# tl-worker draining rule and tl-lead drain-the-fleet procedure (fleet-drain; recovery only on the operator's word)

## Problem

Project fleet-drain (ops `teams/devinfra/projects/fleet-drain/`, operator-designed 10-20) puts the drain's judgment
in the agents and the steward; Couch only flags slots and delivers notices (pair#722 re-scoped). The agents and the
steward need the rules.

## Spec

- Worker DRAIN follows the authenticated notice-only contract confirmed by pair:5 and the TL's clarification:
  graceful finishes the current atomic step and waits only for background work already running; leave queued,
  not-started work queued under the steward's pause and note it in the Log as outstanding. Fast promptly stops only its own background
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
  guard already-running background-work/checkpoint ordering, queued work left queued and logged outstanding,
  deferred-message reply, restart/lift distinction,
  owner-scoped queue control, and automation versus recovery authority. General suite/targeted-test commands
  live in TESTS, outside DRAIN.
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
- [x] Record a pushed-head suite, close through the binary-owned review, then managed merge and report
  to hub and pair:5 before its watched live run. Follow the TL decision on the pair#739 landing dependency.

## Log

### 2026-10-20
- 10-21 00:10 (TL ops:0): scope addition (operator-approved, pair#739): **hub automation**. tl-lead rule: after a Couch
  start where `couch --automation status` reports the hub's automation **off** (full shutdown or unknown start), the
  TL turns it on (`couch --automation on`) when it reads the operator's first message (console draft, session file or
  side doc), unless that message says to keep it off; until then, it doesn't act on recovery. The TL may turn its own
  hub's automation off as a soft stop when the operator asks; it never changes another hub's. Lands after pair#739.

### 2026-10-22 — claim, peer contract and quick-flow proposal
- 2026-10-22: closed — BR-1 consistency sweep RED→GREEN on both real-root queue-help entry points, with zero store/authority/live-inspection calls; frozen protocol, authority and root render controls PASS 0.474s; TL drain/automation/ownership guards PASS; real queue CLI exit0 shows authoritative owner pointers. TL-approved worker/steward wording unchanged; README/atlases synchronized. Earlier suite745b PASS 8 shards/2135 tests/76 packages; corrected-head452d suite lease-1208786352 is QUEUED, not claimed passing. Current tl-worker PERMISSIONS allows close overlap; final-head recorded pass still required before landing, and code fixes require a final-head rerun. No live drain, queue pause/resume or stage.; review verdict: SHIP

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

### 2026-10-22 — design approved; entering red-first implementation

- Verified TL approval receipt a6ad302f-bf0b-49be-ba9d-b455c13415f5. Capability-qualified automation wording
  is approved; #506 lands independently of pair#739 and before pair#722's watched live proof. Paused queued
  work stays outstanding. TL requires the final DRAIN and steward wording before close.
- `sdlc change-code --issue 506` admitted quick flow. Saved the verbatim observed stale real-CLI DRAIN
  excerpt with capture/source provenance under cmd/tl-worker/testdata/drain; a bounded test agent owns only
  the help tests/rule table/source-vector fixtures and their initial red run. Main owns all shipped help,
  README/atlas and workflow. No live drain, queue control or private Couch harness is involved.

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

### 2026-10-22 — TL final wording correction

- TL reviewed the exact rendered sections at 485ead4d8 (outbound receipt
  d030a2c1-92e1-4934-82d7-b5fa19e818b4 submitted). Verified reply
  07154afa-0199-4587-9e03-a3c2b592173e approves steward/automation unchanged and requires two worker edits:
  graceful waits only for already-running background work, leaves queued work queued and logs it outstanding;
  general suite/targeted command advice moves from DRAIN into TESTS. Approved proceeding to suite/close after edits.
- This closes the paused-queue waiting ambiguity without treating queued work as completed or restart-safe.
  Pair:5 notified via submitted receipt 500c1056-df14-4c64-9a12-e3fe18eb044f. Verified peer reply
  17f85fca-9b21-45a4-a6f9-476865a0958b: its old “current background” source is ambiguous, not conflicting;
  peer proposed the same running/queued distinction to the hub before changing its reviewed #722 candidate.
  Original recorded/source fixtures remain verbatim; clarification gets its own provenance rather than
  silently rewriting historical evidence. Worker tests are being updated red-first before help changes.

### 2026-10-22 — clarified worker rules verified

- Independently reproduced clarification RED on both help entry points: old help retained the broad wait,
  lacked queued/outstanding safeguards, and kept general verification guidance in DRAIN instead of TESTS.
  After the two TL-required edits, focused worker RED→GREEN passed (0.133s). Full TL packages passed again:
  worker 0.075s, lead 25.207s. The corrected real worker CLI render exited 0; steward/automation is unchanged.
- Pair:5 supplied the TL-approved revised graceful source vector (verified receipt
  2d574326-5cb3-41de-99a2-d2fb341fd3e8); current fixtures use it, while the original vectors remain verbatim
  in notices.before-running-clarification.json. Provenance distinguishes the historical source commit from
  peer/TL-approved revised text awaiting peer source patch/re-close; no invented new commit identity.
- Moved later dated evidence back under Log and retained all scope revision paragraphs under Revisions;
  no historical note content was discarded. Recorded suite is next on the pushed corrected head.

### 2026-10-22 — recorded suite passed; entering close

- PR #390: https://github.com/xianxu/ariadne/pull/390. Final-wording corrections at
  745b0f7dcc27c1f7d6daa02fdcc316b831396797 were pushed before the recorded check; no code changed since.
- `sdlc local-check run --name suite` passed on that exact head: 8/8 shards, 2135 cmd/sdlc tests,
  76 other packages, zero failures; wall 758.1s (12m38s). Lease lease-798942241 waited 1139s before
  starting, then completed normally. Saved log `.git/sdlc/local-checks/745b0f7dcc27c1f7d6daa02fdcc316b831396797/suite.log`,
  SHA256 prefix b461107aceb1. Current GitHub checks are green; no local-gates paths changed.
- Pair:5's source formatter and expected-body tests now agree byte-for-byte with our current graceful
  fixture (verified receipt 06d83c76-b5db-4ab2-9c53-5d8e81a75c3a). Independently measured 537 UTF-8 bytes,
  SHA256 612bc49844512b2c992db2b3e973b117244c659a37411d2b2d06febf36d5fe92. No vector change/requeue needed.
- TL final wording approval and both requested corrections are fulfilled. The final plan checkbox records
  verified suite completion and entry into the binary-owned acceptance/landing sequence, not a claimed
  review verdict or landing. Invoking close next; pair#722's operator-watched live run remains held until landing.

### 2026-10-22 — close BR-1 consistency sweep

- Close round 1 returned FIX-THEN-SHIP with BR-1 Important (drain-contract-consistency), keeping close
  unfinalized: cmd/sdlc/helptext/queue.md still requires acknowledging an authenticated drain ID, and
  atlas/workflow/sdlc-binary.md still calls Couch's formatter “notice/ack”. Independently rendered real
  `go run ./cmd/sdlc queue --help` (exit 0) and confirmed the stale instruction.
- Enumerated current consumers with a repository-wide acknowledgment/drain grep, excluding history,
  issue/review evidence and deliberately historical fixtures. The two stale shipped prose sites above
  are the complete remaining class; current TL help/atlas explicitly reject a formal ack, and test
  negatives/historical captures remain evidence, not instructions.
- Correction replaces these two restatements with pointers to authoritative worker/Couch procedures.
  No worker/steward/automation wording changes and no queue runtime behavior changes. Add rendered queue
  help checks for both entry points and a zero-store-access guard before changing the shipped help.
  A fresh exact-head recorded suite and another close are due after this executable help correction.

### 2026-10-22 — BR-1 red/green correction committed

- Independently reproduced `TestQueueHelpUsesNoticeOnlyDrainContract` RED: both queue --help and no-args
  lacked authoritative pointers and retained acknowledgment instructions; the injected queue store,
  authority and live-inspection traps were not called. Replaced the queue help and atlas restatements
  with worker/Couch owner pointers; historical captures remain unchanged.
- GREEN: queue-help plus frozen protocol, authority and root-render wiring tests passed (0.474s).
  Worker/lead drain, automation, ownership and narrative guards passed (0.067s/0.085s).
  Real `go run ./cmd/sdlc queue --help` exited 0 and displays both owner pointers. `git diff --check` clean.
  Diff confirms no change to the exact TL-approved worker/steward/automation sections.
- Close round 1 was not finalized, so the binary explicitly requires another full close for BR-1
  disposition after this fix; no finalized close generation is being bypassed. Fresh recorded suite
  on the corrected pushed head is next. Earlier passing suite remains attributed to its original head.

### 2026-10-22 — corrected-head review overlaps the queued suite

- Fresh recorded suite on 452d20e0dc40 is retained as lease-1208786352 (attempt
  d26c2327b97e6c200e84d6be09cef61e); it is queued, not yet passing. Earlier suite proof stays on 745b0f7d.
- Current tl-worker PERMISSIONS explicitly permits a close review to overlap a queued recorded suite,
  provided a final-head pass precedes landing and code fixes requeue the check on the final head.
  Using that existing allowance for BR-1 disposition; targeted queue/TL controls already passed.
- The final checkbox records entry into the acceptance/landing sequence, not a claim that this queued
  suite passed or that landing finished. No merge integration or publication until its required proof
  is complete; if this review changes executable code, the queued check will be replaced through its
  supported owner-scoped control. No runner duplication or queue reordering.

### 2026-10-22 — corrected-head suite passed; entering managed landing

- Close round 2 returned SHIP/high with BR-1 addressed and zero new/open findings, bound to evidence
  4b775cbfc2c33c59b913ea6f0413e4d9c3206060. Reviewer independently passed both TL packages, queue controls,
  all six real CLI help renders and Go overlays restoring old help, which made all three regressions red.
- Corrected-head suite on 452d20e0dc40355d4069d300b773ef3f1caa0e54 completed PASS: 8/8 shards,
  2136 cmd/sdlc tests, 76 other packages, zero failures; wall 710.4s (11m51s), after 1710s queue wait.
  Lease lease-1208786352 completed normally. Saved log `.git/sdlc/local-checks/452d20e0dc40355d4069d300b773ef3f1caa0e54/suite.log`,
  SHA256 prefix 1d1760b31905. No executable code changed between that tested head and the accepted close.
- Required review-overlap condition is now fulfilled: the corrected code has a recorded pass before
  landing. Entering sdlc merge --yes next. TL-owned ops/fleet-drain.md remains for the hub to update after
  the verified landing report; pair#722's watched live remains held until this item lands.

### 2026-10-22 — retained landing recovery and independent priority work

- The first retained merge reached admission after #507/#548 but refused because the final documentation-only
  evidence commit had not been pushed. Updated PR390 through sdlc pr, then re-entered the managed line.
- A temporary additional worktree briefly occupied main-slot5 while preparing independent ariadne#553,
  causing the next runner's resting-branch refusal. Removed that clean temporary worktree and followed
  the printed sdlc merge recovery. The independent task now uses an ordinary clone within ariadne-slot5;
  this issue's source checkout and executable changes stay intact.
- Current retained runner is pid44224, terminal18055, waiting since06:35:58 behind #547. No duplicate runner,
  failed suite or pending operator decision. The hub confirmed keeping it intact (receipt aa81d80b-89b8-49cf-b014-2279db49ed71).
- Hub explicitly dispatched #553 while this item waits. Its private clone has a separate landing line,
  so the hub approved completing its suite/close there, then importing its pushed branch here only after
  #506 lands (receipt 1e45c411-1ba8-4ffc-b17d-ce7fb0d5b577). No separate-line landing or admission bypass.
  The hub revoked #553's temporary grant so #506 can retain normal admission. #722's watched live proof
  remains after this landing and the hub's binary pickup.

## Revisions

- 2026-10-22: Updated the design to pair#722's authenticated notice-only contract, removing the obsolete
  formal-ack assumption and distinguishing lift from restart. Added the operator-approved automation rule
  from the existing Log with an explicit unresolved pair#739 landing dependency; TL decision pending.

- 2026-10-22 revision: TL explicitly superseded the earlier “Lands after pair#739” constraint: ship the
  drain correction before pair#722 live proof, use “when Couch provides couch --automation” so #739's later
  landing simply makes the condition true. The rest of the approved worker/steward/recovery design stands.

- 2026-10-22 revision: TL final-wording review narrows graceful waiting to already-running work and keeps
  queued work outstanding in the Log; test-command advice moves into TESTS. Steward and automation stay as approved.

- 2026-10-22 revision: BR-1 completes the approved notice-only contract sweep through the queue help and
  queue-drain atlas consumer. Remove obsolete restatements in favor of owner pointers; add the missing
  rendered queue-help guard. The approved worker/steward procedure and all command behavior stay unchanged.
