---
id: 000505
status: done
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
started: 2026-10-20T17:05:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.03
tracker:
    version: 1
    completion:
        token: close-b0eea5515189
        repository: github.com/xianxu/ariadne
        reviewed_head: f090762e0082b0815e7a553be42d438cb4c49ed5
        evidence_commit: ef5c1d3f3371efd06b7b9e0b0f614a1fe98d7a9b
        landed_commit: 8d68f1933299a2837ec1fda55a90af6a7c9bdc6a
---

# tl-lead: designs of fleet-wide, failure-sensitive workflows (drain, recovery, lifecycle, queue, scheduling) go to the operator for review before code

## Problem

The TL approved pair#722's fleet drain through several revisions (acks, a queue barrier, a message hold, persisted
state). Its first live run on 10-20 wedged the whole fleet: refused acks, all messages held, a cancel that never
confirmed, state that survived a reboot. The operator's own model, written in ten lines afterwards, was far simpler
and robust (steward-driven, an in-memory flag, agents keep accepting messages). Operator, 10-20: "this also shed
light of where operator needs to pay more attention: those type product/system design that needing assess risks of
various event and design workflow that's robust ... those are the design that should be reviewed by the operator, a
lot of different ways things can be done, and AI tends to both over engineer and under engineer."
