---
id: 000482
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1263fa7fcae561d06ef9778a245d263116d6caf1' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T22:48:28-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# tl-lead: keep a just-landed slot's conversation (next same-area car), and a 'hold: soak until' option for risky changes

## Problem

When a landed change later breaks something, the slot whose conversation wrote it has the freshest context, but
tl-lead has no rule about keeping it: slots get prepared with a fresh conversation, rebooted, or dispatched into
unrelated code right after landing. Operator idea, 10-19 (ops side doc
`teams/devinfra/tl-status-xian-xu.side/2026-10-19-16-landing-queue-status.md`).

## Spec

Two tl-lead rules (text in `tl-lead --help`, DISPATCH or LANDING section):
1. **Default: keep the conversation, not an idle slot.** Don't reboot, or prepare with a fresh conversation, a slot
   that just landed. Give it the next car in the same code area (the train rule already prefers that). Start it
   fresh only when its context is unrelated or nearly full. Note: `couch --prepare` reloads a fresh agent
   conversation today, so the rule says when to prepare without reload, or that preparing a just-landed slot for a
   same-area car skips the reload (if prepare supports that; else file the prepare option in pair).
2. **TL judgment, for risky changes** (recovery or landing code, data formats, anything hard to roll back):
   "hold: soak until X". The slot stays reserved and idle until the change has soaked (the next train lands, or the
   next main suite after it passes, or a stated time). The TL may open an extra slot so the hold doesn't cost
   throughput. The hold shows in the status file's Slots line.

## Done when

- `tl-lead --help` carries both rules, in a few lines, with the soak exit conditions.
- If `couch --prepare` can't skip the conversation reload, a pair issue for that option is filed and referenced.

## Plan

- [x] tl-lead help, DISPATCH: keep a just-landed slot's conversation for the next same-area car. Skip `--prepare`, since the worker's `start-plan` refreshes the substrate and `couch --relaunch` keeps the conversation. Start it fresh only for unrelated or nearly full context.
- [x] tl-lead help, LANDING: "hold: soak until X" for risky changes, with three exits (the next train lands, the next main suite after the change passes, a stated time). It shows in `## Slots`, and an extra slot can absorb the throughput.
- [x] Help tests pin both rules.

## Log

### 2026-10-19
- 2026-10-19: closed — Round-1 fixes: soak holds live in the rendered Slots section (tl-lead status slots FILE --hold repo:N=until X / --release), carried by every render from the section read at write time; a held slot is never on the free line; help free-for-new-work excludes held slots (TestStatusSlotsSoakHold). Atlas pointer added. Fake race in fakeSDLC.IssueShow fixed (-race clean). Rules as before, pinned by help tests. Recorded suite running on ce8324a5b.; review verdict: SHIP
- 2026-10-19: flow upgraded quick → full — 121 added lines in code files (limit 100)
- **The couch --prepare question (Spec 1):** `couch --prepare` is `weave refresh` followed by a fresh conversation, per `couch --skill`. No pair option is needed: the conversation-keeping path already exists.
  - The worker's own `sdlc start-plan` catches its slot up (it runs `weave refresh` and rebuilds the slot's sdlc; seen on #468).
  - `tl-worker --help` is read anew per item.
  - `couch --relaunch repo:N --confirm` restarts a slot onto a new Pair binary and keeps its conversation.
  So the rule is to skip prepare for a just-landed slot taking a same-area car, and relaunch only for a Pair binary change. No pair issue was filed.
- **Close round 1 (FIX-THEN-SHIP):**
  - BR-1 was right. `## Slots` is binary-rendered, so a hand-written hold would be overwritten, and the held slot listed as free.
    - The hold now lives in the section: `tl-lead status slots FILE --hold repo:N='until X'` / `--release repo:N`. Each render reads the holds back from the section being replaced, at write time inside the compare-and-swap, so a retry re-reads them.
    - A held slot gets its own line and never appears on the free line.
    - The help's "free for new work" excludes held slots.
    - `TestStatusSlotsSoakHold` covers set, carry, release and refusals.
  - Minor: the atlas pointer has been added.
  - Side fix: a race in the test fake `fakeSDLC.IssueShow` (an unguarded `calls++` under `boundedEach`) made `TestGatherTrains` flaky. It is now guarded, and `-race` is clean.
