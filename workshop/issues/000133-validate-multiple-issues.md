---
id: 000133
status: working
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-29
estimate_hours: 0.27
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

- [x] `sdlc issue validate --issue 1,2,3,4` validates all listed issue IDs.
- [x] `sdlc issue validate file1 file2 file3` validates all listed files.
- [x] Existing single-target and `--all` behavior still works.
- [x] Mixed-target and `--all` conflict behavior is documented and tested.
- [x] Tests cover all-targets-pass and one-target-fails exit behavior.
- [x] Help text includes the new examples.

## Design decisions

Resolving the spec's open questions:

- `--issue` switches from cobra `IntVar` → `IntSliceVar`, so `--issue 1,2,3,4`
  parses natively (and `--issue 124` is just a one-element slice). No hand-rolled
  comma splitting.
- Positional files: `Args` relaxes from `MaximumNArgs(1)` → `ArbitraryArgs`; each
  arg is validated.
- **Mixing `--issue` IDs and positional files is rejected** with a clear error.
  Rationale: agents batch-validate via one form or the other; rejecting keeps the
  contract crisp and the precedence unambiguous (no dedup/order surprises).
- **`--all` combined with any explicit target is rejected** (mutual exclusion made
  explicit, not silent precedence).
- All resolution still funnels into the existing per-file `validateIssueFull`
  loop — no duplicated schema/section logic (ARCH-DRY). `resolveValidateTargets`
  stays a thin filesystem seam over the pure validation core (ARCH-PURE).

## Plan

- [x] Change `issueValidateFlags.Issue int` → `Issues []int` (`IntSliceVar`); relax
      `Args` to `ArbitraryArgs`; reuse `locateIssueFile` per ID.
- [x] Rewrite `resolveValidateTargets` to handle multi-file, multi-ID, the
      mixing-rejection, and the `--all` mutual-exclusion errors; keep feeding the
      existing `validateIssueFull` loop.
- [x] Add tests: comma-separated IDs, multiple files, all-pass vs one-fails exit,
      invalid ID/file, mixed-target rejection, `--all`+target rejection.
- [x] Update the command Long help with both multi-target examples.
- [x] Run the focused `sdlc` test suite (`go test ./cmd/sdlc/...`).

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module   design=0.1 impl=0.15
design-buffer: 0.15
total: 0.27
```

Well-specced extend of an existing command (`issue validate`). Design is mostly
pre-resolved by the thorough spec (only a few CLI-contract decisions), so low
design + reduced buffer; impl is the v3.1 40%-scaled mid of the smaller-go-module
range (flag-type change + resolver rewrite + ~6 table-driven tests + help).

## Log

### 2026-06-26

Created from pair#81 retro point 3. The motivating failure was an agent trying to validate #72-#79 in one command and hitting the current one-positional-file limit.

### 2026-06-29

Implemented in `cmd/sdlc/issue.go`. `--issue` is now a cobra `IntSliceVar` (native
comma parsing, no hand-rolled split — ARCH-DRY), `Args` relaxed to `ArbitraryArgs`,
and `resolveValidateTargets` rewritten as a 4-way switch (`--all` exclusion →
mix-rejection → IDs → files) that still funnels into the single `validateIssueFull`
loop. Tooling: `go test ./cmd/sdlc/...`.

Verified against the rebuilt binary (exit codes confirmed directly):
- `validate --issue 133,137` (all conform) → exit 0; `validate a.md b.md` (mix of
  conforming + missing) → per-file diagnostics + exit 1.
- `validate --issue N file.md` → `Error: specify either <file> path(s) or --issue
  ID(s), not both`, exit 1.
- `validate --all --issue N` → `Error: --all is mutually exclusive with explicit
  <file>/--issue targets`, exit 1.
- `validate --all` still globs the dir; `validate --issue 133` single-ID unchanged.
- `validate --help` shows both `--issue 1,2,3,4` and `a.md b.md c.md` examples.

Tests (`issuevalidate_test.go`): added `TestResolveValidateTargets` (comma IDs,
multi-file, `--all`, unknown ID, mix-rejection, `--all`+target rejection),
multi-file all-pass / one-fails subtests in `TestRunIssueValidate`, and
`TestIssueValidateCmdCommaIDs` pinning the end-to-end `IntSliceVar` wiring.
