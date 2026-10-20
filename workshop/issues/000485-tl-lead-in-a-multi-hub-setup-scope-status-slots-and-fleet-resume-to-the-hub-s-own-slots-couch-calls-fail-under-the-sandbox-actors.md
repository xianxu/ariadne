---
id: 000485
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '84c92ba5c46cc12d06b82a5fb58a7f279aaa82c6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T11:08:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] Share hub scope across Slots, resume planning and freshness; implement own-only defaults and --all visibility.
- [ ] Stop serial resume at the first unsuccessful terminal outcome; retain diagnostics and name untouched slots.
- [ ] Verify two-hub behavior, stateful delayed resumes, subprocess contracts and live read-only views.
- [ ] Update help/atlas, record the suite, close and land.

## Log

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

## Revisions

- 2026-10-20 13:43 PDT — proposed design after Codex handoff: include the TL's 11:00/11:15 serial-resume acceptance in Done when. Reuse Couch's existing terminal-receipt polling instead of adding a second poller (ARCH-DRY). Count unclaimed slots without guessing the main hub; --all is visibility only. Add explicit protection against losing an existing hold when narrowing Slots. Full details and state/recovery table are in the durable plan; approval pending.
