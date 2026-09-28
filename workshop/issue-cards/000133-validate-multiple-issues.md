---
id: '000133'
status: done
started: 2026-06-29T14:52:25-07:00
created: 2026-06-26
updated: 2026-06-29
estimate_hours: 0.27
actual_hours: 0.26
---

# sdlc validate multiple issues

## Problem

`sdlc issue validate` currently supports exactly one issue ID (`--issue N`), exactly one file positional, or `--all`. During pair#72, validating a newly-created set of issues naturally led to:

```sh
sdlc issue validate file1 file2 file3
```

but the command failed with `accepts at most 1 arg(s), received 8`. The same workflow also wants a concise issue-ID form:

```sh
sdlc issue validate --issue 1,2,3,4
```

The single-file/single-ID contract makes batch validation awkward exactly when agents are creating or updating multiple linked issues.
