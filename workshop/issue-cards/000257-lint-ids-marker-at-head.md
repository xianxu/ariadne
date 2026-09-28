---
id: 000257
status: codecomplete
created: 2026-09-27
updated: 2026-09-27
estimate_hours:
github_issue:
started: 2026-09-27T23:33:11-07:00
actual_hours: 0.19
tracker:
    version: 1
    handoff:
        token: move-ed46b701b18c
        repository: github.com/xianxu/ariadne
        source_branch: refs/heads/main
        source_base: 5ded00f34bfad5674b8ecc8bf394514741033580
        source_head: 5ded00f34bfad5674b8ecc8bf394514741033580
        source_path: workshop/issues/000257-lint-ids-marker-at-head.md
        source_blob: a70d70bacfdbd72c27e6067287241c0c8550efa7
        destination: workshop/issues/000257-lint-ids-marker-at-head.md
        main_commit: 76c347e6829f88c629a3b20b7e999281dbe26257
    completion:
        token: close-5382178e7d90
        repository: github.com/xianxu/ariadne
        reviewed_head: 79ee4f70a2a65ab5dd3c97a07ab96fe6222cd8cc
        evidence_commit: aad120e70a25d4c2ac40ebb599fb90882cd86b85
---

# lint-ids reads the cutover marker from the checked commit, not the checkout

## Problem
