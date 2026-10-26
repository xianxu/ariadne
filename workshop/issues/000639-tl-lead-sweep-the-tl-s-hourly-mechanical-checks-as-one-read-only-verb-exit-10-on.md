---
id: 000639
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '9c352bbe17d1ec0d2b866521fa95d717a513a9a0' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead sweep: the TL's hourly mechanical checks as one read-only verb, exit 10 on findings

## Problem

## Spec

**The TL's hourly slot sweep as a tl-lead verb (operator 10-24, via ops#10).** The sweep is a timer inside the TL's agent session: a restart loses it, and every run costs the TL's attention even when nothing is wrong. The operator's rule for janitor jobs is that they run on their own and the hub reacts only to an exit code. tl-lead already joins the needed sources for `tl-lead status slots` (sdlc fleet inventory, couch --actors, sdlc issue show, hub scope). The sweep reuses that join, and TL rules live in tl-lead, so the logic belongs here. ops' scheduled task (ops#10) becomes a thin wrapper.

`tl-lead sweep [--hub DIR] [--json]` is read-only: it never nudges, kills or writes. It prints one line per finding with the command to inspect it. It exits 0 when quiet, 10 on any finding, and 1 when a source fails, naming that source (an unreadable source is never read as quiet). Checks and bounds (TL-set):
- **Idle with a working issue:** a slot in this hub's scope whose agent is idle at its prompt (couch --actors Resting) and whose LastActivity is older than 30 min while it owns a working issue. Inspect: couch --peek SLOT --lines 20. Whether it reported is the TL's to read.
- **Stale queue operation or waiter:** a running merge, landing or suite operation, or a waiting ticket, older than 90 min; or an operation whose holder pid is gone. Inspect: sdlc queue status, sdlc landing status --repo R. Needs issue B's fields. Until then this check reports 'unavailable' as a failure, not quiet.
- **Load:** the 1-minute load average over 2x the logical cores (sysctl vm.loadavg, hw.logicalcpu).

## Done when

- Tests over fake sources (actors, inventory, issue show, queue and landing JSON, load) cover each finding, the all-quiet case, and a failing source (exit 1 naming it).
- `tl-lead --help` documents the verb, each bound and what the hub does on exit 10. ops#10 can wrap it.

## Plan

- [ ]

## Log

### 2026-10-25
