---
id: 000487
status: done
created: 2026-10-19
updated: 2026-10-20
estimate_hours:
github_issue:
severity: high
started: 2026-10-20T00:23:30-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.30
tracker:
    version: 1
    completion:
        token: close-ac0a8408c71c
        repository: github.com/xianxu/ariadne
        reviewed_head: 8513558178988a1fe9af9ce1bc32f484fb7da941
        evidence_commit: e4e50f4c47053274093d01bb96af4471cc229a0d
        landed_commit: 0ccd3ab9f5786b53cb27f017650ca70e09a85b70
---

# Make TestScratchFallbackOverridesInheritedTemp independent of host sandbox temp permissions

## Problem

Follow-up to ariadne#477. ariadne:5 reports that #467 PR325's integrated recorded suite attempt c3ec77ef failed exactly cmd/sdlc.TestScratchFallbackOverridesInheritedTemp (suiteenvironment_test.go:100). The real normal-temp probes under /var/folders/.../T and /tmp both returned operation not permitted inside the agent sandbox. A main-baseline diagnostic is queued; its result is not yet known. #476's unsandboxed recorded full suite passed, so this is an environment-dependent test assumption, not evidence of a #467 weave defect.

The test sets inherited TMPDIR and GOTMPDIR below a regular file, then calls the production runLocalCheck path. That deliberately invalidates its fixture temp root too. prepareScratch subsequently calls scratchdisk.NormalTempDir, which searches the host user-temp directory and /tmp. The child-write assertions therefore depend on host sandbox permissions rather than a controlled fallback root.
