# Boundary Review Semantic Output Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist and parse only a reviewer's semantic final response while keeping harness diagnostics visible on the operator's stderr.

**Architecture:** Replace the judge subprocess seam's combined byte stream with a small `ProcessOutput` value containing distinct stdout and stderr bytes. Both synchronous and heartbeat dispatch paths pass that value through one completion function, which forwards stderr to the configured diagnostic sink, preserves the existing exit-code policy, and returns stdout as the only semantic review output. The existing boundary-review path continues to parse, print, and persist the returned string, so its sidecar and gate behavior become correct without a second filtering rule.

**Tech Stack:** Go 1.26, `os/exec`, stdlib writers/buffers, existing sdlc judge/boundary-review tests.

**ARCH alignment.** `ARCH-DRY`: synchronous and heartbeat dispatch use the same diagnostic-routing and exit-classification function. `ARCH-PURE`: stream routing and exit classification operate on a value and injected writer; the real subprocess adapter only captures bytes. `ARCH-PURPOSE`: `Run` owns process streams, `Dispatch` owns the semantic/diagnostic boundary, and review sidecars remain durable review records rather than process logs. `ARCH-MOCK`: fake `Run` values test every adapter and boundary path, while one real subprocess test checks the OS stream split.

---

## Decisions and acceptance assumptions

1. Agent CLI stdout is the semantic final-response channel; stderr is diagnostics/progress. This is one contract at the `Dispatch` boundary for Claude, Codex, and Gemini.
2. `Dispatch` returns stdout only. If `DispatchOptions.Stderr` is non-nil, it receives captured process stderr plus heartbeat lines; if it is nil, diagnostics have no caller-provided destination and are not joined into the semantic result.
3. A subprocess non-zero exit remains review output, not a launch failure: stderr is forwarded, stdout is returned, and downstream classification decides whether the response is usable. A launch failure remains an error and forwards any captured diagnostic bytes first.
4. `*-review.md` persists boundary metadata plus the semantic final response. `*-gate.md` remains the structured finding/disposition ledger parsed from that response.
5. Existing review artifacts are not rewritten by ariadne. The oversized pair#146 artifact is normalized separately before retrying that close.
6. Prompt argv transport, review-window sizing, reviewer checkout isolation, and external-finding lifecycle are outside #201 and remain tracked by #162, #204, and #202/#195.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `ProcessOutput` | `cmd/sdlc/internal/judge/dispatch.go` | new; tagged process result with `Stdout` and `Stderr` byte slices |
| `classifyRunResult` | `cmd/sdlc/internal/judge/dispatch.go` | modified; forwards diagnostics and classifies one `ProcessOutput` |
| review-sidecar body contract | `cmd/sdlc/reviewsidecar.go` | unchanged implementation; body is now explicitly semantic review output |

`ProcessOutput` prevents an agent adapter or test fake from representing two process channels as one ambiguous byte slice. `classifyRunResult` is the sole transition from process output to semantic output: it writes diagnostics, applies the existing `*exec.ExitError` policy, and returns only stdout.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `Run` | `cmd/sdlc/internal/judge/dispatch.go` | modified | `exec.CommandContext`, separate stdout/stderr buffers, start callback, wait |
| `Dispatch` | `cmd/sdlc/internal/judge/dispatch.go` | modified | agent argv, optional heartbeat, shared process-result completion |
| `dispatchBoundaryReview` | `cmd/sdlc/milestoneclose.go` | behavior clarified | semantic output printing, verdict/findings parsing, sidecar persistence |
| boundary-review regression | `cmd/sdlc/closereview_test.go` | extended | fake judge process through persisted sidecar and terminal writers |
| artifact documentation | `atlas/workflow/sdlc-binary.md`, `atlas/workflow/ledger-landscape.md` | modified | durable review/gate ledger contract |

---

## Chunk 1: Separate the process streams at the judge boundary

### Task 1.1: Capture real subprocess stdout and stderr independently

**Files:**
- Modify: `cmd/sdlc/internal/judge/heartbeat_test.go`
- Modify: `cmd/sdlc/internal/judge/dispatch.go`
- Modify mechanically: judge `Run` fakes in `cmd/sdlc/**/*_test.go`

- [ ] Change `TestRun_RealSubprocess` first to require `ProcessOutput.Stdout` to contain only `to-stdout` and `ProcessOutput.Stderr` to contain only `to-stderr`. Keep the PID assertion and require stderr from a non-zero process to survive alongside `*exec.ExitError`.
- [ ] Run `go test ./cmd/sdlc/internal/judge -run TestRun_RealSubprocess -count=1`; verify FAIL because `Run` still returns one combined byte slice.
- [ ] Add `ProcessOutput { Stdout, Stderr []byte }`. Change production `Run` to use two buffers and return the value after `Start`/`Wait`, preserving the owner-bin PATH and `onStart` behavior.
- [ ] Mechanically update all replaceable `judge.Run` test fakes to return `ProcessOutput`, without changing their assertions or canned semantic output. This is a signature migration required to restore compilation, not new behavior.
- [ ] Re-run `go test ./cmd/sdlc/internal/judge -run TestRun_RealSubprocess -count=1`; expect PASS.

### Task 1.2: Route diagnostics once in synchronous and heartbeat dispatch

**Files:**
- Modify: `cmd/sdlc/internal/judge/judge_test.go`
- Modify: `cmd/sdlc/internal/judge/heartbeat_test.go`
- Modify: `cmd/sdlc/internal/judge/dispatch.go`

- [ ] Add a table test over Claude, Codex, and Gemini. Each fake returns a valid verdict on stdout and a distinct diagnostic on stderr. Assert `Dispatch` returns only the verdict, writes only the diagnostic to `opts.Stderr`, and still builds the expected adapter command.
- [ ] Extend the synchronous-path test (`opts.Stderr == nil`) to prove stderr is absent from the returned semantic output. Extend the deterministic heartbeat test so heartbeat lines and captured process diagnostics share the stderr sink while the returned output remains stdout-only.
- [ ] Extend launch-error and non-zero-exit tests: captured stderr reaches the sink in both cases; non-zero exit returns stdout with nil dispatch error; launch failure returns the existing diagnosable error.
- [ ] Run `go test ./cmd/sdlc/internal/judge -run 'TestDispatch|TestRun_RealSubprocess' -count=1`; verify the new stream-routing assertions FAIL.
- [ ] Change the shared completion function to accept `ProcessOutput`, the run error, binary name, and diagnostic writer. Forward stderr before applying the existing exit policy, then return only stdout. Route both synchronous and heartbeat branches through this one function.
- [ ] Re-run `go test ./cmd/sdlc/internal/judge -count=1`; expect PASS.

---

## Chunk 2: Prove and document the durable artifact contract

### Task 2.1: Persist only semantic output at a real boundary-review seam

**Files:**
- Modify: `cmd/sdlc/closereview_test.go`
- Modify: `cmd/sdlc/milestoneclose.go`

- [ ] Add a boundary-review test whose fake `judge.Run` returns a valid `VERDICT` plus structured findings fence on stdout and a trust-dialog/progress diagnostic on stderr. Use separate stdout/stderr buffers and a real temporary plans directory.
- [ ] Assert the command stdout, `reviewResult.Output`, verdict, and parsed findings use the semantic response; assert command stderr contains the diagnostic; assert the written `*-review.md` contains the verdict/findings and excludes the diagnostic. Keep the existing `*-gate.md` round assertions where the close fixture exercises gate persistence.
- [ ] Run `go test ./cmd/sdlc -run 'TestDispatchBoundaryReview.*Semantic|TestClose.*Sidecar' -count=1`; verify FAIL before the routing implementation is complete (or prove PASS only after Task 1.2 supplies the intended behavior).
- [ ] Update misleading `milestoneclose.go` comments from “full transcript” to “final review response.” Do not add filtering or stream knowledge to sidecar code.
- [ ] Re-run the targeted close/boundary tests; expect PASS.

### Task 2.2: Update the durable-ledger contract

**Files:**
- Modify: `atlas/workflow/sdlc-binary.md`
- Modify: `atlas/workflow/ledger-landscape.md`
- Modify: `workshop/issues/000201-boundary-review-artifacts-capture-the-harness-trust-dialog-preamble.md`

- [ ] Replace “full transcript” claims with the precise split: `*-review.md` is boundary metadata plus final review response; `*-gate.md` is structured finding/disposition state; process diagnostics/progress remain terminal output. Preserve the local agent transcript as the fallback only when no sidecar exists.
- [ ] Mark the four issue plan boxes complete and append a dated `## Log` entry naming the stream boundary, regression coverage, and issue separations (#162/#204).
- [ ] Run `rg -n 'full review transcript|full transcript' cmd/sdlc atlas/workflow` and inspect every remaining match; historical references may remain only when they intentionally describe old behavior.

### Task 2.3: Verify the complete change

- [ ] Run `gofmt -w` on modified Go files.
- [ ] Run `go test ./cmd/sdlc/internal/judge -count=1`.
- [ ] Run `go test ./cmd/sdlc -count=1`.
- [ ] Run `go test ./... -count=1`.
- [ ] Run the repository's standard generated/help checks identified by `make help` or the existing test target, then run `git diff --check`.
- [ ] Inspect `git diff --stat` and `git diff` to confirm no prompt transport, reviewer isolation, or historical sidecar rewrite entered #201.
- [ ] Close with `sdlc close --issue 201 --verified '<exact commands and results>'`; fix every Critical/Important finding from the gate-owned fresh review before retrying.

