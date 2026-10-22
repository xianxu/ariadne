---
id: 000547
status: working
created: 2026-10-21
updated: 2026-10-22
estimate_hours: 2.9
github_issue:
started: 2026-10-22T04:59:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
---

# Train car active-time counts conductor-owned queue wait and status polls as focused work

## Problem

pair#778 renewed close on 2026-10-21 adopted 4.41h and identified a 213.9-minute/80% dominant segment from20:18 to23:52. That segment includes the retained car waiter PID51323 from20:37 to23:44 (about187minutes) while train261 was owned by pair:2 and waiting/running its combined suite, main baseline and prefix suite. pair:5 only polled/reported retained state during that interval; branch changes resumed after its waiter exited. The timing report excluded66.57minutes across2 earlier own-suite waits, but did not exclude this conductor-owned waiting. Frequent status polling appears as engagement. Do not treat the full213.9m as pure wait: it also contains work before/after the retained waiter.
