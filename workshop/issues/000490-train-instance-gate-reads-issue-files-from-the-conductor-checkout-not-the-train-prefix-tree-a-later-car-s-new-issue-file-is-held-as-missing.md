---
id: 000490
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '21d25956972bf97c14a58ab5b2145c52fa434ec1' # card fields mirrored from issue-cards; edit via sdlc
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

## Spec

The instance gate validates each prefix against that prefix's own tree (paths resolved against the train checkout
at that prefix, not the conductor worktree). Reuse the train's prefix checkout; no second resolver.

## Done when

- A 2-car train fixture where car 2 adds an issue file the conductor lacks: car 2 validates and boards (red before).
- A genuinely malformed car still holds with the file and rule named (control).
- pair's next train with 2+ cars lands both; recorded in test-economy.

## Plan

- [ ]

## Log

### 2026-10-20
