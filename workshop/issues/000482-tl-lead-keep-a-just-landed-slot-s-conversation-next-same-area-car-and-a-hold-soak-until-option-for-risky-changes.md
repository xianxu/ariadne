---
id: 000482
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '3d40ef3e2a1fc7215b0ed152f2b57a0c782aed89' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-19
