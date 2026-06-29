---
id: 000133
status: working
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-29
estimate_hours:
started: 2026-06-29T14:52:25-07:00
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

## Spec

Extend `sdlc issue validate` to accept multiple targets.

Required CLI forms:

```sh
sdlc issue validate --issue 1,2,3,4
sdlc issue validate file1 file2 file3
```

Compatibility requirements:

- Preserve current single-target forms:
  - `sdlc issue validate --issue 124`
  - `sdlc issue validate path/to/x.md`
  - `sdlc issue validate --all`
- `--issue` should accept a comma-separated list of numeric IDs.
- Positional file validation should accept one or more paths.
- Mixing `--issue` and positional files should either be supported deliberately or rejected with a clear error; choose and document the behavior.
- `--all` should remain mutually exclusive with explicit targets.
- Output should preserve per-file diagnostics and exit non-zero if any target fails.
- Help text should show both multi-target examples.

Implementation should reuse the existing per-file validation path rather than duplicating schema/section checks (`ARCH-DRY`).

## Done when

- [ ] `sdlc issue validate --issue 1,2,3,4` validates all listed issue IDs.
- [ ] `sdlc issue validate file1 file2 file3` validates all listed files.
- [ ] Existing single-target and `--all` behavior still works.
- [ ] Mixed-target and `--all` conflict behavior is documented and tested.
- [ ] Tests cover all-targets-pass and one-target-fails exit behavior.
- [ ] Help text includes the new examples.

## Plan

- [ ] Update argument parsing for `issue validate`.
- [ ] Reuse the existing validator loop over a resolved list of files.
- [ ] Add tests for comma-separated IDs, multiple files, invalid IDs/files, and partial failure.
- [ ] Update help text and usage examples.
- [ ] Run the focused `sdlc` test suite.

## Log

### 2026-06-26

Created from pair#81 retro point 3. The motivating failure was an agent trying to validate #72-#79 in one command and hitting the current one-positional-file limit.
