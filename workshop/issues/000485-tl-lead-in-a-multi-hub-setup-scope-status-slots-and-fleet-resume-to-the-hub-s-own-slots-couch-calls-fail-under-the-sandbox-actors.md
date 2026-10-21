---
id: 000485
status: codecomplete
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-20
estimate_hours: 2.02
collaboration-mode: collaboration
card_mirror: '71a8eda0837d1d3d57d4147defb74db494a43203' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T11:08:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 5.44
---

# tl-lead in a multi-hub setup: scope status slots and fleet resume to the hub's own slots; couch calls fail under the sandbox (actors '?')

## Problem

From introduce-team's live proof (10-19, two hubs: `couch ops ops-echo`):
1. **`tl-lead status slots` isn't scoped to the hub.** In ops-echo:0 it rendered the whole machine's fleet into echo's
   status file, not just the slots its `.hub` claims (echo).
2. **tl-lead's couch calls fail inside the agent sandbox.** `tl-lead status slots` shows `?` in every actor column,
   and `tl-lead fleet resume` lists every slot as "unknown: couch --actors unreadable … dial unix
   /tmp/pair-message-501/….sock: operation not permitted". The sandbox exemption covers a bare `couch` call only, so
   a couch call made from inside tl-lead is blocked. Run outside the sandbox, `fleet resume` classed correctly (one
   slot to resume, the rest running or idle).
3. `fleet resume`'s dry run prints dozens of "ambiguous … unclaimed" rows for dormant repos (robotics, zellij,
   xianxu.dev, …), which buries the one actionable row.
4. **`tl-lead status check` staleness isn't scoped either** (from ops-echo:0, 10-19 23:40): echo's Threads and Slots
   were marked stale by a parley.nvim merge (ccbc2dd5), a repo echo's `.hub` doesn't claim. Staleness should come only
   from the repositories the hub's `.hub` claims, so devinfra landings don't ping an idle team.

## Spec

- `status slots` and `fleet resume` scope to the hub's own slots by default (from `.hub` and Couch's ownership,
  hub_ownership.go rules), with `--all` for the machine view.
- tl-lead reaches Couch the way agents can: either via the exempt bare `couch` invocation (spawned so the sandbox
  rule matches), or by reading the state Couch publishes for this (whichever pair/ariadne already supports); if
  Couch can't be read, say so once at the top, not on every row.
- `fleet resume` hides unclaimed dormant repos unless `--all`, and prints a one-line count instead.

## Done when

- A stateful delayed-resume fake proves no next slot starts before the current one settles, including 10–30-second modeled delays.
- The first refused, failed or uncertain resume stops the batch, returns nonzero, retains request diagnostics and names untouched slots.
- An uncertain attempt triggers only read-only observation; no automatic resend or inferred success from actor presence.
- --all expands visibility, never automatic resume authority; scoped Slots writes cannot silently drop an existing hold.

- `status check` in a two-hub fixture marks a hub stale only for merges in repos its `.hub` claims (test).

- In a two-hub fixture, each hub's `status slots` lists only its own slots; `--all` lists every slot.
- A run where Couch is unreachable prints one note, not N per-row errors (test).
- The dry run on the live machine shows only owned slots plus a count of unclaimed ones.

## Plan

Proposed durable design: [hub scope and serial resume](../plans/000485-hub-scope-resume-plan.md).
Awaiting the owning TL's approval before change-code.

- [x] Share hub scope across Slots, resume planning and freshness; implement own-only defaults and --all visibility.
- [x] Stop serial resume at the first unsuccessful terminal outcome; retain diagnostics and name untouched slots.
- [x] Verify live read-only views (TL fleet dry runs) plus two-hub, delayed resume and subprocess automated tests.
- [x] Update help/atlas and README.

## Log

### 2026-10-20 — boundary fixes
- 2026-10-20: closed — Resolved the sole atlas prose conflict with current main, retaining hub scope and receipt-conformance guidance plus #504 session.side paths. Code merged automatically, with no code conflict. Integrated TL_LEAD_RESUME_CONFORMANCE=1 go test -race ./cmd/tl-lead -count=1 passes (68.312s); staged-file Go lint catalog passes. An established intermittent #504 helper-start timing failure is registered as unclaimed #514 with same-source fail and ten targeted passes, then complete targeted race pass. Recorded final-head suite queued; required pass remains before landing. Two-hub read-only live proof remains logged. (fixed-then-ship); review verdict: SHIP
- 2026-10-20: closed — Full tl-lead race suite including TL_LEAD_RESUME_CONFORMANCE=1 passes; staged-file golint catalog passes. BR-1 stalled-recheck regression reproduced 30-second delay before fix, now bounded at 500ms injected; stopped/untouched and guidance precede optional probe. BR-2 lint annotations, BR-3 README flags, BR-4 isolated real Couch queued/running/terminal JSON and exit conformance complete. TL live two-hub resume dry runs PASS, logged with exact binary revision. Updated recorded full suite queued on pushed head; required pass remains before landing. Atlas/help updated.; review verdict: SHIP

- BR-1: execActors now bounds the whole observation class at 10 seconds plus at most one second pipe cleanup. The 30-second-stalled-recheck regression passes with an injected 500 ms bound, and asserts stopped/untouched output plus recovery guidance precede optional observation.
- BR-2: virtual-time Sleep lines annotated with the supported test-fixed-sleep allowance; staged-file catalog passes. The earlier pre-stage catalog could not see new untracked tests; this was a verification gap, now covered in lessons.
- BR-3: README documents hub-scoped Slots, both --all flags, visibility-only mutation scope and serial stop behavior, linking atlas.
- BR-4: opt-in isolated real-Couch conformance exercises queued → running → succeeded/refused/failed, standalone receipt JSON and actual process exits through our classifier. Real Couch owns reporting/polling; only its scratch broker is fake. No live supervisor, GUI or slot mutation. Cadence documented for resume closes and supported Couch upgrades.

### 2026-10-20 — live proof and boundary round 1

- Authenticated TL delivery e401f088-ae98-4585-859b-05af9fa5da07 reports live read-only proof PASS with the exact unmodified binary at 5bba914916306320fbf216b85f44d6d2d7789970. Inspected output in sibling slot p485-live-proof.txt: ops sees only its own running/idle slots and 5 foreign slots counted; ops-echo sees echo:0–3 and 23 foreign slots counted. --all exposes other-hub slots as excluded and does not add resume candidates. No --apply, real resume or :0 stage occurred. Slots/freshness behavior additionally has two-hub and real-Git automated coverage.
- Boundary round 1: BR-1 Critical unbounded actor recheck; BR-2 virtual-clock sleep lint annotations; BR-3 README flags; BR-4 opt-in real Couch receipt conformance. Queued old-head suite cancelled through sdlc local-check cancel for these code fixes. BR-1 reproduced 30-second delayed recovery with first actors read successful then a stalled subprocess.

### 2026-10-20 — implementation evidence

- Shared scope, own-only Slots/--all, scoped freshness (including exact/nested claims), single actor diagnostic and hidden-hold guard implemented. Stateful two-hub tests proved red for excess issue reads (12/24 versus selected 3/6), unrelated newer landings, nested claims, dangling/unreadable ownership and hidden holds; focused green after fixes.
- Serial resume now requests Couch's existing terminal JSON receipt, applies pure outcome classification, waits synchronously and stops/nonzero at the first unsuccessful attempt. Tests proved old continue-on-error behavior red; stateful synctest time models 10/20/30-second operations, FIFO barriers prove actual subprocess waiting, and partial-completion/uncertain-observation assertions pass. --all never widens mutation authority; operator-started help is asserted.
- Verification: go test ./cmd/tl-lead -count=1 passed; go test -race ./cmd/tl-lead -count=1 passed; targeted pkg/golint TestTrackedGoFilesPassTheCatalog|TestRules passed. Race run initially found an unsynchronized attempt-slice read in the fake fixture; mutex snapshots fixed the root cause and the full race run passed (not a flaky rerun). No production race was implicated.
- Scope source tests and parent tests finished before committing. Full recorded suite, gate review and staged live proof remain pending.


### 2026-10-19
- 10-20 11:00 (TL): add to scope. After the 10-20 reboot, `tl-lead fleet resume --apply` fired `couch --resume` for
  14 slots back to back: 5 came up, 9 ended "context deadline exceeded … outcome uncertain" (Couch's 2 s admission,
  pair#726), and the operator resumed the rest by hand from the switcher. fleet resume should (1) resume one slot,
  then poll `couch --operation-status` for that request until it settles before the next; (2) on an uncertain
  outcome, check the slot's state and stop the batch, naming the remaining slots and the manual path, instead of
  firing the rest into the same timeout. pair#726 fixes the admission itself; this keeps tl-lead honest either way.
- 10-20 11:15 (TL): test spec for the serial resume above (operator: their manual pair:0–8 resume, one at a time,
  tens of seconds each, all succeeded; that's the behavior to match): unit test against a fake Couch whose resume
  takes 10–30 s, asserting serial order, a wait for each slot to settle, and stop-and-name on the first uncertain or
  failed outcome.

### 2026-10-20 — code map and design sketch (ariadne:2 on Claude, handing to Codex in the same slot)

These are the findings from a read-only exploration. Nothing has been built yet. The design has decisions in it, so
it goes to the TL before `change-code`.

**Code map**
- **Source:** `cmd/tl-lead/` in this module.
- **Seams:** `deps` in `main.go:25-43`: `actors: execActors{bin:"couch"}`, `resume: execResume`, `sdlc`,
  `landing: gitNewestLanding`.
- **Couch calls:** `actors.go:26` runs `couch --actors --json` (no timeout). `fleet.go:55-106` runs
  `--resume-candidates` and `--resume`.
- **`status slots`:** `renderedSections` (`status.go:321`) calls `gatherSlots` (`slots.go:141`).
  - Its source is `sdlc fleet inventory --json`, which covers the whole machine.
  - `slotLines` (`slots.go:84`) has **no hub filter**, so it's problem 1.
  - `agentLabel` (`slots.go:67`) prints `?` on every row when actors fail, and drops the error.
- **`fleet resume`:** `fleetResume` (`fleet.go:216`), with `hereHub` (`team_cmd.go:131`), `readHubs` and
  `planResume` (`fleet.go:143`).
  - Problem 3: an unclaimed slot becomes a `classAmbiguous` row (`fleet.go:151-156`).
  - Problem 2: an actors error becomes one `classUnknown` row per slot, each carrying the full error
    (`fleet.go:165-167`).
- **Ownership:** `.hub` parsing and `ownerOf` (`team.go:18-174`) mirror pair's
  `cmd/internal/couchcore/hub_ownership.go`, which can't be imported from here (pair's `internal`). Couch gives
  unclaimed slots to the first hub on its command line, but tl-lead can't see that order, so it reports `Unclaimed`.
- **`status check` staleness:** `gitNewestLanding` (`freshness.go:124-157`) scans **every** sibling repo that has
  `workshop/issues/` and keeps the newest merge machine-wide. No `.hub` is consulted, so it's problem 4.
- **Tests and fakes:**
  - `fakeCouch` (`fleet_test.go:18`), `fleetFixture` (`fleet_test.go:56`, already two hubs: ops and ops-echo),
    `fakeSDLC` (`trains_test.go:16`).
  - `TestGatherSlots` (`trains_test.go:161`, golden output, no hubs), `TestGitNewestLanding`
    (`freshness_test.go:128`, real git repos).
  - Tests that pin the current behaviour and will change: `TestFleetResumeDryRun` (the ambiguous rows) and
    `TestFleetResumeActorsUnreadable` (the per-row unknown).
- **Docs:** `atlas/workflow/tl-skills.md:54-64`, `cmd/tl-lead/help.md` (TEAMS 28-64; status verbs 271-303), and
  `atlas/workflow/sandbox.md`, "Calling an excluded binary".

**Problem 2 is mostly not a code bug.** `.claude/settings.json` already lists `"tl-lead *"` and `"couch *"` in
`sandbox.excludedCommands`, here and in ops, ops-echo and echo. A *bare* `tl-lead …` Bash call should run
unsandboxed, couch child included. Likely causes are a compound call (`cd X && tl-lead …`, a pipe), a Codex hub
(Codex has no per-command exclusion, per the sandbox atlas), or stale settings. Confirm against the 10-19 echo
transcript. Couch publishes no on-disk state that could stand in for the socket (`registry.json` has no slot
addresses or resting state). The in-binary fix is one Couch-unreadable note at the top, plus documenting the
bare-call rule.

**Design sketch**
0. **Serial resume (TL scope, 10-20 above):** resume one slot, poll `--operation-status` until it settles,
   stop and name the rest on an uncertain or failed outcome; fake Couch with 10–30 s resumes.
1. **Scope `status slots`.** Resolve self from `hereHub(dir of the status file)` and the hubs from `readHubs`. Keep
   a slot when `ownerOf(...).sole() == self`.
   - Unclaimed slots are counted, not listed: `(N unclaimed, M other hubs; --all)`.
   - `--all` keeps today's machine view.
   - With no `.hub`, today's behaviour stays.
   - *Decision for the TL:* unclaimed slots go to Couch's main hub, but tl-lead can't tell which hub that is.
     Should they be counted everywhere, or shown in the hub whose name is first alphabetically?
2. **Scope staleness.** Pass self's claimed repo names into `gitNewestLanding` and skip repos the hub doesn't
   claim. With no `.hub`, keep the machine-wide behaviour.
3. **Couch unreadable once.** `gatherSlots` and `fleetResume` print one note at the top
   (`couch unreadable: <err>; is tl-lead being run as a bare call?`). Rows stay `?` or `unknown` without the
   repeated error.
4. **`fleet resume`** hides unclaimed ambiguous rows unless `--all`, and prints a one-line count. A tie that includes
   self stays visible, because it's actionable.
5. **Tests:** a two-hub `status slots` case (own versus `--all`), a claims-scoped `TestGitNewestLanding` case, a
   single note when actors fail, and updates to the two pinned resume tests. `help.md` and the atlas get `--all` and
   the bare-call rule.


### 2026-10-20 — Codex recovery and proposed design

- #489 landed as PR 344; its merge returned exit 0 before this branch was created from main. The long landing/test queue prevented starting #485 earlier; no stacked development.
- Read the prior handoff and TL scope entries from origin/main. Current pair Couch's exchangeSlotOperation already polls operation-status to terminal, at 500 ms intervals with a five-minute budget. tl-lead's three-minute outer timeout can interrupt that wait; it currently discards stdout, calls every later slot after errors and classifies Couch's explicit uncertain error as a refusal. The plan keeps Couch as the polling owner and changes batch control/diagnostics.
- Live read-only baseline: bare tl-lead fleet resume from ops reached Couch successfully in this Codex environment and printed dormant unclaimed repos individually. No --apply, live resume, settings edits or :0 writes were performed.
- Proposed plan written to workshop/plans/000485-hub-scope-resume-plan.md; implementation has not started. Await owning TL approval of count-only unclaimed ownership, --all visibility without mutation authority, stop/nonzero partial completion, CLI-owned polling and the hold-preservation refusal.

### 2026-10-20 18:24 PDT — landing integration recovery

- Landing admission integrated origin/main after #504 and refused one conflict in atlas/workflow/tl-skills.md, then cleanly aborted. Manual merge resolves only that prose conflict: keep hub-scoped freshness/holds and receipt-conformance guidance, adopt session.side/ from #504, and retain its root-session documentation. Code merges automatically; no code conflict or scope change.
- The first integrated tl-lead race run with receipt conformance failed only TestMoveCouchRenameDeadlineUnconfirmed from #504: the helper had not published its effect before its 300 ms deadline. On the identical source/index tree, ten targeted race attempts all passed. Registered this established intermittent failure as unclaimed ariadne#514, preserving fail/pass evidence and the startup-timing hypothesis; reported it to the hub. No test was removed or weakened.
- Staged-file Go lint catalog passed after integration. The complete integrated TL_LEAD_RESUME_CONFORMANCE=1 go test -race ./cmd/tl-lead -count=1 passed (68.312 s); the narrow fixed-then-ship re-close follows. Required recorded full-suite verification on the final landing head still precedes merge; no suite or review gate is waived.

## Revisions

- 2026-10-20 14:06 PDT — Issue Plan checkboxes now cover implementation and completed live proof. Recorded suite, close and landing are mandatory sequential publication gates, recorded in Log as they happen; they are not claimed complete by the implementation checklist. This removes the prior circular close/land checkbox, preserving its obligation without requiring a plan-check waiver on later closes. BR-1 refinement bounds all execActors reads (including recovery) at 10 seconds plus a one-second pipe-drain cap; recovery result/guidance prints before the optional probe.

- 2026-10-20 13:43 PDT — proposed design after Codex handoff: include the TL's 11:00/11:15 serial-resume acceptance in Done when. Reuse Couch's existing terminal-receipt polling instead of adding a second poller (ARCH-DRY). Count unclaimed slots without guessing the main hub; --all is visibility only. Add explicit protection against losing an existing hold when narrowing Slots. Full details and state/recovery table are in the durable plan; approval pending.

- 2026-10-20 13:44 PDT — TL approved the durable plan and authorized change-code. Added constraint: fleet resume remains operator-started, never automatic at startup; preserve --help wording and assert it in tests. Approval is delivery bac6759f-4e2e-4a0d-a272-4c8409efe48e, authenticated as submitted from ops:0 to this incarnation. The earlier awaiting-approval notes are historical.

- 2026-10-20 13:47 PDT — plan gate round 1 blocked on PQ-1 (confirming child-outcome evidence) and PQ-2 (test-plan presentation). Verified Couch already supports --resume ADDR --json and emits terminal receipt status plus normal exit. Plan now names pure classifyResumeResult: only a matching terminal receipt consistent with normal exit confirms completion; every possibly admitted unconfirmed result remains uncertain. Test plan compacted to named functions/strategies; no change to approved visible states or polling ownership.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* Calibration is flagged stale by estimate-source; numbers are provisional.

Three smaller-go-module extensions cover shared scope/Slots, scoped freshness and pure receipt classification: each base design 0.2 × 0.2 thorough-spec discount = 0.04; base implementation 0.5 × 0.40 = 0.20. Serial batch integration uses api-integration (base design 1.5 × 0.2 = 0.30; implementation 1.0 × 0.40 = 0.40). Issue/spec is 0.75 × 0.2 design and 0.2 × 0.40 implementation. Docs is 0.1 × 0.2 design and 0.1 × 0.40 implementation; the single close review is 0.1 × 0.2 design and 0.5 × 0.40 implementation. Existing ownerOf, Cobra, Couch receipt polling and encoding/json cover library needs; no greenfield stack or new external API needs discovery. Familiarity 1.0; thorough-plan design buffer 15%. Queue wait is not implementation effort.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: api-integration design=0.30 impl=0.40
item: issue-spec design=0.15 impl=0.08
item: atlas-docs design=0.02 impl=0.04
item: milestone-review design=0.02 impl=0.20
design-buffer: 0.15
total: 2.02
```

Design 0.61 × 1.15 + implementation 1.32 = 2.0215 hours, rounded to 2.02.

### Estimate allocation clarification

The serial api-integration row includes batch wiring, delayed/subprocess fake coverage, output retention and read-only observation. Smaller-module rows cover scope/freshness/classifier logic plus their focused fixtures. The single milestone-review row includes active suite invocation/result inspection, live-proof coordination and landing checks; idle queue waiting is excluded, while recorded verification execution remains measured by actuals. The estimate is provisional against stale calibration, not a ceiling on the necessary verification.
