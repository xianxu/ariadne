---
id: 000567
status: codecomplete
created: 2026-10-22
updated: 2026-10-23
estimate_hours:
github_issue:
flake_tests: [github.com/xianxu/ariadne/cmd/tl-lead.TestOperatingCancellationStopsGitHookDescendants]
started: 2026-10-23T05:17:26-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.35
tracker:
    version: 1
    completion:
        token: close-8b65c838c39c
        repository: github.com/xianxu/ariadne
        reviewed_head: 839fc18ae652343c0777a8a515ac943476fc32d3
        evidence_commit: 843549fcb391e21e27b8cd37ee66e4b1830c6b90
---

# flake: TestOperatingCancellationStopsGitHookDescendants

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/tl-lead.TestOperatingCancellationStopsGitHookDescendants: Recorded #563 suite attempt 2cb913014a018f68422fdf9ef52180cf on Git-verified head 6a0b35faea2b351b5b1ec19a310f6ea1d19f1948 failed at doc_operating_test.go:419: Git descendant remains runnable: ?E (0.87s). Unchanged exact detached checkout targeted command go test ./cmd/tl-lead -run ^TestOperatingCancellationStopsGitHookDescendants$ -count=1 -timeout 2m returned exit0 PASS in0.986s. Captured original JSON events and targeted result are retained in common test-stats/evidence/000563/cancellation-failure.jsonl and cancellation-targeted-pass.txt. Full original suite is still running; no full retry or source edit. Investigate whether transient Darwin exiting state E is misclassified as runnable, without weakening cleanup proof.
