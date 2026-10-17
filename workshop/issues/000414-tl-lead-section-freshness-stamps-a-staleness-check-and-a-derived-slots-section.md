---
id: 000414
status: codecomplete
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: 'a166b790d34835a550127adb56a4fceeedaa7c3d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T16:53:29-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.17
---

# tl-lead: section freshness stamps, a staleness check, and a derived Slots section

## Problem

The TL session file (`teams/<team>/tl-status-<operator>.md` in ops, written only through `tl-lead status`) goes stale:
on 2026-10-16 its Threads section was ~20 hours old while Needs you stayed current. Nothing tells the TL a section is
stale: updating Threads is a synthesis no single event asks for, and writes blocked by the operator's lock get queued
and forgotten. Operator, 10-16: move this into the hard shell; long-term, only Operator instructions (chat) and Needs
you need to be prose in the doc; the rest becomes command output, a "flexible cockpit", some of it eventually a pure
operations dashboard.

## Spec

1. **Freshness stamp inside each section (operator's design: no scattered state).** Every `tl-lead status` write of a
   section puts one line right after its `## ` heading, e.g. `_updated 2026-10-16 17:05_`. Tolerant: a section whose
   stamp line the operator deleted is fine (treated as unknown age) and gets the stamp back on the next write. The
   stamp never counts as operator content (the operator-section guard ignores it).
2. **Staleness check.** `tl-lead status check` (light) reports sections older than their budget (default: Threads and
   Slots older than 2 h, or older than the newest landing the TL recorded); the session-sync ping to the TL carries the
   stale list so it survives compactions.
3. **Derived sections.** `tl-lead status slots --refresh` writes Slots from `couch --list` (slot, agent kind, live or
   not) plus each slot's claimed issue and branch from `sdlc` (title, status); no TL prose. Design it as the first of
   several derived sections (a "landed since the last update" per project from the project ticks is the next).
4. TL prose keeps only judgement lines; (2) flags them when stale.

## Done when

- Tests: a write stamps its section; a deleted stamp is restored on the next write and doesn't trip the operator guard;
  `status check` flags a stale section and not a fresh one; `slots --refresh` renders a fixture fleet exactly.
- Live: the ops session file shows stamps, `slots --refresh` produces today's Slots, and a stale Threads shows up in
  the TL's ping.

## Plan

Design (sent to the TL, ops:0, for approval before code). `ARCH-*` noted where they shaped a choice.

**Found in the code:** `tl-lead status slots FILE` already exists (`cmd/tl-lead/slots.go`, a row in
`renderedSections`) and renders from `sdlc fleet inventory` + `sdlc issue show` + the checkout's issue file. The ops
file's Slots is still hand-written, because the renderer lacks the agent kind and the inventory takes 10–45 s. So (3)
extends the existing renderer; it adds no new one (ARCH-DRY).

**Stamp format.** The first line after the `## ` heading: `_updated YYYY-MM-DD HH:MM_` in local time, the same format
as the file's header line. One owner: `writeSection` (the single write path) removes any stamp line at the top of the
current body before `compute` sees it, and puts a fresh stamp at the top of the new body. So `section`, `thread` and
every rendered verb stamp, and a body file or thread entry that contains a stamp can't double it. A missing stamp has
unknown age and comes back on the next write. `## Operator instructions` is never stamped (the operator owns it; it
has no TL freshness). `instructions list/link` skip a stamp line if the operator adds one, so it never counts as a
bullet (the "operator guard"). Parse rule: line 1 of the body matches `^_updated \d{4}-\d\d-\d\d \d\d:\d\d_$`;
anything else is content.

**Staleness rule (`tl-lead status check FILE`, light: one couch read plus local git).** A table in code, one row per
section: `Threads` and `Slots` have a 2 h budget and the landing rule; `Needs you` has no budget (event-driven; flagged only when unstamped, TL 10-16); `Quota` has 12 h;
`Operator instructions` has no check. A section is stale when (a) its stamp is older than its budget, (b) for the
landing-rule sections, its stamp is older than the newest landing, or (c) it has no stamp ("unstamped"). Newest
landing is the newest committer time of `refs/remotes/origin/main` across the fleet's repos, read from local refs with
no fetch. The repos are the sibling directories of the file's repo with a `construct/` (the peer convention). That
avoids `sdlc fleet inventory`, which took 44 s today. Output: one line per stale section, e.g. `stale: ## Threads
(updated 20:41, 3h12m; ariadne landed 22:10)`, or `fresh`; exit 0 either way (it reports, it doesn't refuse).

**Where the stale list reaches the TL.** The spec says the session-sync ping carries it. That ping is couch's (pair),
not tl-lead's. In ariadne: every `tl-lead status` verb the TL runs on a ping (`submit`, `owed`, each write) ends with
the `check` lines, so each event's handling shows what's stale, after any compaction. A couch-side ping that carries
it, or a timer that wakes the TL when nothing else does, is a pair change: file it as a pair issue (recommended), not
in this one.

**Slots renderer inputs.** Keep `sdlc fleet inventory` (slot → checkout, branch, resting branch) and per busy slot
`sdlc issue show` (status) + the issue file (bar). Add `couch --actors --json` (light): `Binding.Agent` (claude, codex,
qoder) and `Resting` (idle or working) per slot address. Line: `pair:2  pair#618  started  ▓▓░ 2/3  claude·working
slug`. A failed actors read renders `?` for agent and liveness and never blocks the section (the same rule as a failed
issue read). A new `CouchActors` seam with a stateful fake (as for `SessionSync`). `--refresh` per the spec is the
existing `status slots FILE`; I'd keep the verb and not add a flag. The registry row stays the extension point for
"landed since the last update" (next).

**Durable state (ARCH-CONSTRAINTS at-plan):** the stamp lives in the section it describes, one line per section,
replaced on each write; it creates no file, row or cache. `check` creates nothing.

- [x] Stamp in `writeSection`: strip on read, prepend on write; `instructions` skips a stamp line. Tests: a write
      stamps; a deleted stamp comes back; a stamp in a body file isn't doubled; a thread upsert keeps one stamp;
      instructions list ignores it.
- [x] `status check FILE`: budget table, landing reader (seam + fake), report; appended to submit/owed/write output.
      Tests: stale by age, stale by landing, unstamped, fresh, Operator instructions skipped.
- [x] Slots: `CouchActors` seam + fake; agent and liveness columns; `?` on failure. Test: fixture fleet + actors renders
      exactly.
- [x] help.md: stamps, `check`, slots inputs; atlas entry.
- [x] Live: run on the ops file (stamps appear, `slots` renders today's fleet, `check` flags a stale Threads).

## Revisions

- 2026-10-16 (close review): `check` does one session-sync read per budgeted section (4) plus one local git call per
  sibling SDLC repository, not "one couch read"; still inventory-free. The Slots agent label is `agent` (working) or
  `agent·idle`, `none` where no agent runs, and the free line shows each free slot's agent. The landing source is SDLC
  siblings only (`workshop/issues/`, no `.brain/`).

## Log

### 2026-10-16
- 2026-10-16: closed — Merged origin/main; resolved the atlas/workflow/tl-skills.md conflict keeping both the #414 Freshness entry and #412 Operator instructions entry; help.md auto-merged and still documents check; go test ./cmd/tl-lead green; full suite green. (fixed-then-ship; full suite: make test green after merging origin/main 63b1a010: 12 shards, 1522 cmd/sdlc tests + 66 packages, 0 failed, 821 s); review verdict: SHIP
- 2026-10-16: closed — Round-1 SHIP minors fixed: go test ./cmd/tl-lead green incl. TestStatusCheckEdges; TL_LEAD_LIVE=1 conformance (session-sync missing-section text, couch --actors shape) passes live. Earlier: make test green (12 shards, 0 failed); TL live run of status slots and check matched.; review verdict: SHIP
- 2026-10-16: closed — make test green (12 shards, 1519 cmd/sdlc tests + 66 packages, 0 failed). New tests: TestWriteStamps (a write stamps; a deleted stamp returns; a stamp in a body is not doubled; Operator instructions never stamped), TestStampIsNotABullet, TestSplitStamp, TestStatusCheck (stale by age, by landing, unstamped; Needs you only unstamped; fresh), TestGitNewestLanding (brain and non-SDLC excluded), TestGatherSlots (exact fixture render plus the ? degraded render), TestDecodeActors. Live (TL ops:0, branch build): status slots wrote a stamped Slots with agents and idle/working and printed the stale list; status check matched.; review verdict: SHIP
- 2026-10-16: flow upgraded quick → full — 352 added lines in code files (limit 100)
- TL approved the plan (ops:0): keep `check` light with the sibling-repo local refs; the couch-side ping goes to pair
  (draft its text in the report; the TL files it); tl-lead runs unsandboxed and `?` is the right degraded render.
  Change: `Needs you` has no time budget, flagged only when unstamped.
- Implemented. Landing source narrowed during the live check: siblings with `workshop/issues/` and no `.brain/`. Brain
  and ops auto-commit to origin/main many times a day and would keep Threads permanently stale; this is the same line
  sdlc's spine guard draws (test: `TestGitNewestLanding`). The missing-section read is matched on couch's real refusal
  text ("heading line, found 0", pair `sessionsync/file.go`); the fake now returns the same text.
- Live, read-only: `tl-lead status check` on the ops file → all four budgeted sections `unstamped` (correct: nothing has
  stamped yet). The live writes (stamps, `slots`) are the TL's: the ops file is ops:0's, so I don't write it from
  ariadne:4.
- Live (TL, ops:0, branch build, unsandboxed): `status slots` wrote Slots stamped `_updated 2026-10-16 17:02_`, agent
  kinds and idle or working, 12 working slots and a `free:` line, then printed `stale:` for Needs you, Quota and
  Threads (unstamped); `status check` printed the same three. 37 s wall, mostly the inventory; the per-event check stays
  inventory-free.
- Close review (round 1): SHIP, 4 Minor, all fixed in-round: the landing compare truncates to the stamp's minute
  (`TestStatusCheckEdges`); couch's missing-section text is a named constant pinned by `TestSessionSyncConformance`,
  and `TestCouchActorsConformance` pins the `--actors` shape (both pass live, TL_LEAD_LIVE=1); main wires
  `gitNewestLanding` explicitly and an empty seam means no landing; this Revisions entry fixes the plan text. Also added:
  the landing-unknown and section-unread lines are tested.
