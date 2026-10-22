---
id: 000548
status: done
created: 2026-10-22
updated: 2026-10-22
estimate_hours: 1.93
github_issue:
severity: critical
started: 2026-10-22T01:34:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.54
tracker:
    version: 1
    completion:
        token: close-60c2a7c53889
        repository: github.com/xianxu/ariadne
        reviewed_head: f4d418291df2bf170ea38006a4a55052526ce60b
        evidence_commit: ab0c8051f9cedb50b7382e29a2f8b72ccf007d9c
        landed_commit: 80602233df36a670023b60af8dabb0b6f3c2c071
---

# Actual queue evidence must belong to the measured issue head

## Problem

Echo#9 actual preview refuses because another issue's recorded suite attempt lies in its measurement period while echo:0 holds more than one claim. The suite attempt belongs to echo#3, not echo#9. This blocks echo#9's fix and therefore echo#10, #3 and #5 suites/landings.
