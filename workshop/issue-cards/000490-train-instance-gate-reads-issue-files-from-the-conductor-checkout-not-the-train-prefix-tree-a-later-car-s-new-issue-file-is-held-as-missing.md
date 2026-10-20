---
id: 000490
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
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
