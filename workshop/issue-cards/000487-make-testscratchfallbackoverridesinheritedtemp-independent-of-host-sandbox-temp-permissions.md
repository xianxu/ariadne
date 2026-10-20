---
id: 000487
status: open
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
---

# Make TestScratchFallbackOverridesInheritedTemp independent of host sandbox temp permissions

## Problem

Follow-up to ariadne#477. ariadne:5 reports that #467 PR325's integrated recorded suite attempt c3ec77ef failed exactly cmd/sdlc.TestScratchFallbackOverridesInheritedTemp (suiteenvironment_test.go:100). The real normal-temp probes under /var/folders/.../T and /tmp both returned operation not permitted inside the agent sandbox. A main-baseline diagnostic is queued; its result is not yet known. #476's unsandboxed recorded full suite passed, so this is an environment-dependent test assumption, not evidence of a #467 weave defect.

The test sets inherited TMPDIR and GOTMPDIR below a regular file, then calls the production runLocalCheck path. That deliberately invalidates its fixture temp root too. prepareScratch subsequently calls scratchdisk.NormalTempDir, which searches the host user-temp directory and /tmp. The child-write assertions therefore depend on host sandbox permissions rather than a controlled fallback root.
