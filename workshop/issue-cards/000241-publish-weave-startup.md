---
id: 000241
status: codecomplete
created: 2026-09-20
updated: 2026-09-28
estimate_hours: 1.23
github_issue:
started: 2026-09-28T18:52:27-07:00
actual_hours: 0.85
tracker:
    version: 1
    completion:
        token: close-b764aef79e1b
        repository: github.com/xianxu/ariadne
        reviewed_head: 54e9cc5ec34688b5b2bfab7d878662321ef2614d
        evidence_commit: 133b6ac6f6b78fad73015767c715be0b9548cc18
---

# Publish weave and cut over startup

## Problem

Weave must be available as the standalone entrypoint for ariadne-style repos.
The startup implementation in #239 needs to merge before a public release and
consumer cutover can be verified. Tracking delivery here avoids making #239's
close depend on a release that itself depends on #239 being merged.
