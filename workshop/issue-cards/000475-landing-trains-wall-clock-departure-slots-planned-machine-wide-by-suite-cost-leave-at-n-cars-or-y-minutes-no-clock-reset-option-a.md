---
id: 000475
status: done
created: 2026-10-19
updated: 2026-10-20
estimate_hours: 2.67
github_issue:
started: 2026-10-20T03:03:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.90
tracker:
    version: 1
    completion:
        token: close-9659c67503da
        repository: github.com/xianxu/ariadne
        reviewed_head: fc32559e438281ea63d14ab4fa4905e0c299e0b7
        evidence_commit: cef25b965e8d203db878db7ac96b256d0192dab3
        landed_commit: f5e742f33ddbcca69869e808279e47a63c3678c3
---

# landing trains: wall-clock departure slots planned machine-wide by suite cost; leave at N cars or Y minutes; no clock reset (option A)

## Problem

Landing trains (#449) depart per repo on independent clocks: a train leaves `train_interval_minutes` after the previous one,
takes up to `train_max_cars`. Nothing coordinates repos on one machine, so heavy suites collide at random in the 2-slot suite
budget (ariadne's suite takes both slots; measured 7-day medians: ariadne 9.5 min, pair 12.5, parley 2.7; slot waits 11–19 min).
A full train waits for the clock even when N cars are ready, and a lone or granted landing resets the train clock
(`landingtrain.go` ~l.661, `NextDue = now + interval`), pushing the next train a whole interval back. Operator discussion and
decision (option A, 10-19): ops `teams/devinfra/tl-status-xian-xu.side/2026-10-19-11-landing-schedule.md`.
