---
id: 000490
status: done
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
started: 2026-10-20T03:05:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.09
tracker:
    version: 1
    completion:
        token: close-043775f70226
        repository: github.com/xianxu/ariadne
        reviewed_head: 63346f514919ad5b093b8e3184c068222e28a49f
        evidence_commit: 645c637a28d90b361cd6c32b764809d0c801b7af
        landed_commit: a70a26a195fdefd190dc03b8f184dda5bfe52bf6
---

# Train instance gate reads issue files from the conductor checkout, not the train-prefix tree: a later car's new issue file is held as missing

## Problem

pair's live train 174 (10-20 ~03:10, after ariadne#486) boarded pair#720 (conductor, pair:7) and pair#723. The
workflow provider now works (suite queued on c2c6b927), but the instance gate **held #723** with:

```
workshop/issues/000723-622-cpu-replay.md (frontmatter): vocabulary: read that file: no such file or directory
```

The file exists on #723's branch and in the train prefix, but not in the conductor's own checkout (#720's branch).
So the instance validation reads issue files relative to the **conductor checkout**, not the **train-prefix tree** it
is validating. Every multi-car train where a later car adds or changes an issue file the conductor lacks will
wrongly hold that car. (The same gate's earlier "1 nonconforming instance" in train 167 was probably this too.)
