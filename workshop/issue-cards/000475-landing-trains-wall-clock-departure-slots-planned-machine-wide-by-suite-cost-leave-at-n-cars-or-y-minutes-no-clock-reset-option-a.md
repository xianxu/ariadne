---
id: 000475
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# landing trains: wall-clock departure slots planned machine-wide by suite cost; leave at N cars or Y minutes; no clock reset (option A)

## Problem

Landing trains (#449) depart per repo on independent clocks: a train leaves `train_interval_minutes` after the previous one,
takes up to `train_max_cars`. Nothing coordinates repos on one machine, so heavy suites collide at random in the 2-slot suite
budget (ariadne's suite takes both slots; measured 7-day medians: ariadne 9.5 min, pair 12.5, parley 2.7; slot waits 11–19 min).
A full train waits for the clock even when N cars are ready, and a lone or granted landing resets the train clock
(`landingtrain.go` ~l.661, `NextDue = now + interval`), pushing the next train a whole interval back. Operator discussion and
decision (option A, 10-19): ops `teams/devinfra/tl-status-xian-xu.side/2026-10-19-11-landing-schedule.md`.
