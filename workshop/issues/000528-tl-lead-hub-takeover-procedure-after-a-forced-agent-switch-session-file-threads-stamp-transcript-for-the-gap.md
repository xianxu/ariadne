---
id: 000528
status: codecomplete
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '6b66d9a79a16d7e72e1daa4dcca659db14bf3993' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T12:23:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "fd446026", done: "65a2416d"}
actual_hours: 0.11
---

# tl-lead: hub takeover procedure after a forced agent switch (session file, Threads stamp, transcript for the gap)

## Problem

## Spec

The operator decided (ops session.side/2026-10-21-09-hub-standby.md, 10-21): no hot-standby hub; instead an emergency
**force switch** of the hub's coding agent (pair#728), after which the new agent must take over as TL. tl-lead needs
an explicit takeover procedure for that moment (also for any fresh hub session):

1. Run `tl-lead --help` and follow it.
2. Read the session file (and `session.side/`). The **Threads section's `_updated_` stamp** says how current it is.
   The operator uses the session file as an ephemeral scratch pad (deletes items once handled), so it is not a
   complete history; tolerate some loss.
3. If the gap since that stamp matters, read the previous agent's **terminal transcript from that time on**
   (`couch --peek <hub> --previous --all`, or the agent's own transcript on disk), only for the gap.
4. Re-establish: peek working slots, check `sdlc landing status` / `queue status`, open project files; then update
   Threads with a fresh stamp and tell the operator in one line that the takeover is done and what (if anything) may
   have been lost.

The same procedure applies whether the new agent is in the same slot (force switch) or a fresh session.

## Done when

- tl-lead's help has a concise TAKEOVER section with the steps above; rendered-help phrase guards cover it.

## Plan

- [x] Add rendered-help guards for both entry points, with each recovery step checked in order and its instructions confined to that step; observe the missing-section failure.
- [x] Add TAKEOVER before ROLES: read the contract, session and Threads stamp, transcript only for a material gap, re-establish live work and refresh Threads, then give a one-line notice including possible loss. Preserve operator-started fleet recovery.
- [x] Verify the focused tests and own built CLI help; submit the recorded suite. Close and landing retain their independent gates; a completed recorded pass must precede landing.

Scope: quick-flow embedded help and phrase guards; no runtime changes. ARCH-DRY keeps the procedure in the existing embedded help, with its status writer owning the fresh stamp. No new state or hot path.

## Log

### 2026-10-21
- 2026-10-21: closed — README BR-1 fixed for both takeover modes; link target and diff checked. Test-first help guard red-before; tl-lead package passed locally (15.838s) and independently in review (20.797s); built CLI both entry points verified, hub accepted no stage. Suite lease-761235529 on pushed 45c68451930d55cd5a2c5f36073920bb4b01f7d2 still pending; no code changes, completed original pass required before landing.; review verdict: SHIP

- Recovery: claim/start-plan survived; no #528 recorded check or merge runner exists. The operator-approved procedure is settled; implement as quick-flow help text.
- Test-first: TestLeadHelpTakeoverSequence failed with TAKEOVER missing before implementation. Added five ordered steps, checking both --help and bare entry points, step-local phrases, switch/session scope, gap-only transcript read, ephemeral loss and stamp-writing ownership. Full cmd/tl-lead package passed (15.838s). Atlas maps the delivered help and guard.
- Own CLI proof: make tl-lead-build succeeded; bin/tl-lead reports vcs.revision=4333e00a446fbbd2376093cbeded4aaa334ddbad and vcs.modified=false, matching git rev-parse HEAD. Both --help and bare invocation render TAKEOVER; fleet resume still says never run at startup. Final rendered-help focused group passed (0.156s). No :0 touched.
- Recorded suite submitted on pushed head 45c68451930d55cd5a2c5f36073920bb4b01f7d2 as lease-761235529 (terminal retained); initially normal FIFO position 6. It is queued, not passing evidence. Close review may overlap under tl-worker; wait for this original result before merge and never duplicate it for close-only bookkeeping.
- Hub confirmed the read-only built help proof is sufficient; no stage.
- Close round 1: FIX-THEN-SHIP, BR-1 Important (README discoverability); ledger refused finalization. Added one README pointer covering both same-slot force switch and fresh hub session, linking the atlas without duplicating the procedure. Reviewer independently passed cmd/tl-lead (20.797s). Code unchanged; retain original pending suite.
- Close round 2: SHIP; BR-1 addressed, no new findings, independent cmd/tl-lead package passed (16.416s); measured actual adopted 0.11h.
- Recorded suite lease-761235529 failed before tests after 1002s FIFO wait: Go cache /Volumes/AgentScratch/go-build reported no space left on device during compilation. Setup failure, no tests ran; SDLC recorded unattributable-failure, log sha256 5431124a9cf0 at failed/de6af8a538c39ff4afd8d7f7284a9b19.log. df: AgentScratch 100%, 100Mi free; TestScratch 11Gi free. Original terminal exited 1. Reported machine repair/recovery need to hub; no blind retry, no gate bypass.
- Infrastructure recovery: hub removed machine-wide Go cache overrides. Verified new Go processes use /Users/xianxu/Library/Caches/go-build and the main SSD has 278Gi available. The prior ENOSPC setup failure is infrastructure evidence, not test evidence. No private cache created. Rerun the original exact suite head through SDLC; if its earlier unattributable record refuses recovery, report the exact refusal rather than change heads to evade it.
- Retry path: original exact-head invocation refused because 45c68451 is no longer an origin branch tip (close/evidence advanced PR). Hub confirmed current pushed PR head is the normal retry path, old ENOSPC stays infrastructure evidence with no special recovery. Submitted suite on 618839424a7d as lease-3326670138, initially FIFO position 4; pending, no pass claimed.
- Recorded suite retry PASS: lease-3326670138 on 618839424a7ddab3e39bd09239c9afad9cef9736, 1,961 cmd/sdlc tests across 8 shards plus 75 other packages, zero failures, wall 687.8s (11m28s), log sha256 6f460846a5e0. Retained original FIFO waiter through 5039s admission wait; no duplicate/cancellation. Earlier ENOSPC remains infrastructure evidence only. Entering landing with completed suite and SHIP close; all subsequent changes are documentation/Log records.
- Landing refused main integration conflicts in README.md and atlas/workflow/tl-skills.md. Replayed origin/main manually per the refusal; both conflicts were adjacent independent #508 worker-verification and #528 takeover insertions. Kept both blocks unchanged; no code conflicts or behavior edits. Integrated code requires its own recorded suite; the earlier branch pass does not prove it.
