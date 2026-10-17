---
id: 000416
status: codecomplete
created: 2026-10-16
updated: 2026-10-17
estimate_hours:
github_issue:
started: 2026-10-16T19:36:21-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: N/A
tracker:
    version: 1
    completion:
        token: close-018e9b780761
        repository: github.com/xianxu/ariadne
        reviewed_head: 412b0dc4db97e5531f54ebd596b37627b9b40459
        evidence_commit: cd3fbe49f9dd58c3765e62f6e379aed7dd3871a4
    actual_correction:
        version: 1
        token: actual-correction-c6c9f528ea66
        completion_token: close-018e9b780761
        repository: github.com/xianxu/ariadne
        evidence_commit: cd3fbe49f9dd58c3765e62f6e379aed7dd3871a4
        reviewed_head: 412b0dc4db97e5531f54ebd596b37627b9b40459
        previous_actual: "1.12"
        new_actual: N/A
        reason: Adopted 1.12h used a whole-queue subtraction; the recorded wait cannot be uniquely bound to its owning invocation in the pinned native transcript, so the sample is unverifiable
        intent_reason: Adopted 1.12h used a whole-queue subtraction; the recorded wait cannot be uniquely bound to its owning invocation in the pinned native transcript, so the sample is unverifiable
        measurement_digest: 93927e018783f21fe8fe0919443dc5c5405ae3df36c853abcdd1d2d0c82149b6
        measurement_inputs: '{"version":1,"method":"legacy-v1","repository":"github.com/xianxu/ariadne","code_head":"412b0dc4db97e5531f54ebd596b37627b9b40459","since":"2026-10-17T00:40:58-07:00","until":"2026-10-17T00:40:58-07:00","gap_minutes":30,"commit_weight":0,"include_assistant":false}'
        exclude: true
---

# local-check: a cancelled queued run orphans its test-lock helper, which holds the queue ticket after its watched process is gone

## Problem
