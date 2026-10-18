# Curated Go lint catalog Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A catalog of "stupid mistake" rules for Go code, each citing its incident, with flag/no-flag tests and reasoned suppressions, enforced on every landing.

**Architecture:** A stdlib-only package `pkg/golint` parses Go files with `go/parser` and runs a list of rules over the AST. A repo test (`pkg/golint/repo_test.go`) runs every rule over the module and fails on any unsuppressed finding. This is the shape `pkg/shelllint` already uses (ARCH-DRY). Because the test is part of `make test`, the recorded suite, and so every landing, enforces it with no new gate.

**Tech Stack:** Go standard library (`go/ast`, `go/parser`, `go/token`); no new module dependency.

---

## Decisions to confirm (deviations from the issue's Spec)

1. **Engine: in-repo stdlib analyzers, not golangci-lint + ruleguard.** golangci-lint isn't installed on this machine. Installing it is a machine change, and running it at landing or in CI means editing the merge checks or workflows, which are local-gates configuration and need the operator's label. ruleguard adds module dependencies (`go-ruleguard/dsl`, `x/tools`). The stdlib route keeps the same properties the Spec wants: rules in one owned file, a message citing each incident, tests per rule, reasoned suppressions, and enforcement at landing (through the suite). Its cost: no off-the-shelf linters (usetesting, thelper, forbidigo, noctx, bodyclose, errorlint). Those become a separate follow-up issue to adopt golangci-lint, which needs the operator because of the install.
2. **pair adoption is a follow-up, not this issue.** Done-when names ariadne and pair. pair is a peer repo, so the TL routes it. The package is stdlib-only and import-free so pair can take it as base-layer files (weave) or as a module dependency; the TL chooses which.
3. **Strict wire decode is deferred.** A syntactic rule can't tell a wire message from a local config: `DisallowUnknownFields` is right for local files and wrong for cross-version peers. Doing it properly needs an annotation convention (for example `// wire:` on the type), which is design work of its own. It goes in the follow-up list with pair#594 cited.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Rule` | `pkg/golint/rule.go` | new |
| `Finding` | `pkg/golint/rule.go` | new |
| `Check` (file → findings, suppressions applied) | `pkg/golint/check.go` | new |
| `suppressions` (parse `//lint:allow <rule> — <reason>`) | `pkg/golint/suppress.go` | new |
| rules `test-shrunk-deadline`, `tmp-literal`, `test-go-build`, `test-fixed-sleep` | `pkg/golint/rules.go` | new |

- **Rule:** `{Name, Incident, Message string; TestOnly bool; Match func(*ast.File, *token.FileSet) []token.Pos}`. Each rule is one pure function over a parsed file. Every message names its incident.
- **Check:** parses one file's source, runs the rules (only `TestOnly` rules on `_test.go` files), and drops findings a valid suppression covers. Suppression problems are findings too: a missing reason, an unknown rule name, or a suppression no finding uses (that last one is nolintlint's job, done here).
- **Suppression:** a `//lint:allow <rule> — <reason>` comment on the flagged line or the line above it. The reason is required.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `repo_test.go` (walk the module, run Check) | `pkg/golint/repo_test.go` | new | filesystem |

The repo test is the only IO: it walks `cmd/` and `pkg/`, skips `testdata/` and generated files, and reports `file:line rule: message (incident)`. It runs in the suite, so landing enforces it. ARCH-MOCK, ARCH-SECURE: N/A. It reads only the repository's own Go sources and calls no external binary or service.

## The seed rules

| Rule | Flags | Incident | Existing hits |
|------|-------|----------|---------------|
| `test-shrunk-deadline` | in a `_test.go` file, assigning a constant duration under 1s to a package-level variable whose name contains `Timeout`, `Grace` or `Deadline` (`fooTimeout = 50 * time.Millisecond`) | ariadne#434: a clock deadline raced fixture readiness under load | 8 (some only bound a hung call and get a reasoned suppression; any that wait on readiness are fixed to `testfix.DeadlineOnCue`) |
| `tmp-literal` | a string literal `"/tmp"` or one starting `"/tmp/"` passed to a filesystem call (`os.*`, `filepath.Join`, `os.MkdirTemp`, `os.CreateTemp`) | ariadne#342 / parley#342: macOS `TMPDIR`; the sandbox denies `/tmp` | 1 production hit (`cmd/sdlc/close.go:242`, the close warmup file) to fix. Test strings that only use `/tmp` as data are not filesystem calls, so they aren't flagged |
| `test-go-build` | in a `_test.go` file, `exec.Command`/`CommandContext` with arguments `"go", "build"` | pair#602: per-test builds | 6. Each gets a reasoned suppression and one follow-up issue to share a binary built once per package |
| `test-fixed-sleep` | in a `_test.go` file, `time.Sleep` with a constant argument from 100ms to under 1 minute | 2026-10-15 flakes: fixed waits instead of events | 0 today (polling ticks under 100ms and the block-until-killed `time.Hour` idiom are allowed); it keeps the class from coming back |

## Operating envelope (ARCH-CONSTRAINTS)

- **Workload:** one batch pass inside the test suite, over about 1,140 Go files (`find cmd pkg -name '*.go' | wc -l` at plan time).
- **Budget:** under 3 s, measured in the repo test's timing line. The cost is one parse per file and one AST walk per rule, so it is O(files × rules), with no type-checking.
- **If exceeded:** the test still passes, and its log line shows the time.
- ARCH-FUNERAL: creates nothing durable. It reads sources and writes nothing.
- ARCH-ORDER: holds no state between events. Each file is checked independently.

## Plan

- [ ] Task 1: `pkg/golint` core: `Rule`, `Finding`, `Check`, and the suppression parser, with table tests (a valid suppression, a missing reason, an unknown rule, an unused suppression, `TestOnly` scoping).
- [ ] Task 2: the four rules, each with flag/no-flag cases in `rules_test.go`. Each test asserts the incident appears in the message.
- [ ] Task 3: `repo_test.go`: walk the module, run `Check`, report every finding with its incident.
- [ ] Task 4: sweep existing hits. Fix `close.go`'s `/tmp` warmup path to use `os.TempDir()`. Review the 8 shrunk deadlines: convert any that race readiness, and suppress any that only bound a call, with a reason. Suppress the 6 per-test builds with reasons and file their follow-up.
- [ ] Task 5: process rule. tl-worker and tl-lead help say "a fix for a mechanical mistake class adds its rule to pkg/golint", guarded by a phrase in `rules.tsv` (cited scope). The atlas gets an entry beside shelllint, and lessons.md points at the catalog.
- [ ] Task 6: file follow-ups. golangci-lint adoption, with the operator, for the install. pair adoption, routed through the TL. The wire-decode annotation rule. Sharing a built binary across the 6 per-test builds.
