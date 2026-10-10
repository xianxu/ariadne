# Review verdicts survive late messages, backgrounding, network blocks and timeouts — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A boundary or plan-quality review either yields the verdict the reviewer actually gave, or reads "review did not run" with the cause, never `unknown` from a late postscript and never "passed" from a blocked API (#300, absorbing #271, D2, D3 of the 2026-10-09 evidence pensive).

**Architecture:** The claude reviewer runs with `--output-format stream-json`, and a new pure reader turns the event stream into a run record: every assistant message's text, the run's terminal result and its error flag. The verdict and findings are taken from the latest message that carries each block, so a later block-less message erases nothing. A run with no verdict anywhere is dispatched once more; an API or network failure is a dispatch failure ("review did not run"), never a round; and a round with a protocol error is recorded as blocking. The review timeout scales with the window's added lines.

**Tech Stack:** Go (`cmd/sdlc/internal/judge`, `internal/gatestate`, `cmd/sdlc` close/milestone-close/change-code), recorded stream fixtures under `internal/judge/testdata/`, the opt-in live conformance test pattern of `live_stream_conformance_test.go`.

---

## Non-goals

- Structured capture for codex and gemini: they keep text mode (their final output is their whole answer today). The retry and the network-failure classification apply to them through their text output.
- Making the sandbox allow the API host: that is the operator's sandbox configuration. sdlc names the cause and the next action.
- D4–D9 of the pensive (round-cap demotion, reconcile running unmerged finishers, uncommitted-fix reviews, detached checkouts, ledger files left uncommitted, no-code reviews): separate items.

## Design decisions

- **D1 — Stream capture for claude.** `BuildArgs` adds `--output-format stream-json --verbose` for claude. A pure `ReadStream(stdout) (Run, error)` parses the newline-delimited events: the text parts of each `assistant` message in order, and the terminal `result` event (`subtype`, `is_error`, `result`). A stream that doesn't parse (an older CLI, a crash mid-line) falls back to treating stdout as text, as today, so capture never gets worse.
- **D2 — Latest block wins, per kind.** The verdict is parsed from the latest assistant message that carries a verdict (the existing `ParseVerdict` rules applied per message, newest first); findings from the latest message with a valid `findings` block. A block-less later message, the pair#362 and pair#387 shape, changes neither. The sidecar's review text is all assistant messages joined, so a human sees the whole run.
- **D3 — No verdict anywhere: retry once, then `unknown`.** A completed run with no verdict in any message (#271's "I'll wait for the background run to notify") is dispatched once more, with the same prompt and a one-line notice that the previous attempt ended without a verdict. If the second also has none, today's fail-safe stands: `unknown`, the close doesn't finalize, and the sidecar keeps both runs' text. Only the run that produced the verdict becomes the round.
- **D4 — The reviewer is told it runs unattended.** `ContractPreamble` (shared by every recipe) gains: you run non-interactively and receive no notifications; run commands in the foreground within the time limit; your final message must carry the verdict and findings blocks. Allowed tools stay `Read,Grep,Glob,Bash`: Claude Code can't deny Bash's background parameter alone, so this is the prompt rule plus D2/D3's capture making a late message harmless.
- **D5 — An unreachable API is "review did not run".** A run is a dispatch failure, not a verdict, when its terminal result is an error, or when its output (stream result, or text) matches the API-failure signatures seen in the evidence: `API Error`, `ERR_PROXY_TUNNEL`, `Connection blocked by network allowlist`, `403 … blocked`, `ECONNREFUSED`, `ENOTFOUND`. The error names the cause and the next action: "the reviewer could not reach its API (…); in an agent sandbox, allow `api.anthropic.com` for this command, then rerun". Like today's dispatch errors, nothing is persisted and the round cap isn't spent. It isn't retried: the block would repeat.
- **D6 — A protocol-error round is never "passed".** `persistBoundaryRound` and `stampAndPersist` stamp `Blocked = d.Block || round.ProtocolError != ""`, so a round whose review produced no findings block renders as blocked in the ledger and gate output, matching plan-quality's existing behavior. This is the ledger half of D2: even an unforeseen failure shape can't read as a pass.
- **D7 — The timeout scales with the window.** Default = 30 min for a window of up to 500 added lines, plus 15 min per further 1,000 added lines, capped at 2 h (the existing ceiling). The callers compute the added lines of the review window with `git diff --numstat base..head` (boundary) or the durable plan's line count (plan-quality) and pass `DispatchOptions.Timeout`. `WF_REVIEW_TIMEOUT` still overrides. The heartbeat line shows the limit in use.

ARCH-FUNERAL: creates nothing durable beyond the existing sidecar, which grows by the retried run's text when a retry happens (at most one extra run per round).

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Run` / `ReadStream` | `cmd/sdlc/internal/judge/stream.go` | new |
| `RunVerdict(run) (Verdict, findings)` | `cmd/sdlc/internal/judge/stream.go` | new |
| `APIFailure(run) (cause string, ok bool)` | `cmd/sdlc/internal/judge/apifailure.go` | new |
| `ReviewTimeout(addedLines int) time.Duration` | `cmd/sdlc/internal/judge/dispatch.go` | modified (`reviewTimeout` gains the size input) |
| `ContractPreamble` | `cmd/sdlc/internal/judge/contract.go` | modified |

Each is unit-tested without IO: `ReadStream` and `RunVerdict` against recorded streams in `internal/judge/testdata/stream/` (a clean run; verdict then postscript; findings then block-less message; background-wait only; API error result; garbage falling back to text).

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `BuildArgs` | `internal/judge/dispatch.go` | modified | claude stream flags |
| `Dispatch` / `classifyRunResult` | `internal/judge/dispatch.go` | modified | stream read, API failure, one retry |
| `DispatchOptions.Timeout` | `internal/judge/dispatch.go` | new field | — |
| `dispatchBoundaryReview` | `cmd/sdlc/milestoneclose.go` | modified | window size → timeout |
| plan-quality dispatch | `cmd/sdlc/changecode.go` | modified | plan size → timeout |
| `persistBoundaryRound` / `stampAndPersist` | `cmd/sdlc/boundaryledger.go`, `gatepersist.go` | modified | protocol error blocks |

The fake reviewer in the judge tests (a script printing a recorded stream) is the seam; an opt-in live conformance test (`SDLC_LIVE_AGENT_STREAM_CONFORMANCE=1`) checks that the real `claude` emits the event shapes `ReadStream` reads.

## M1 — Verdict capture (D1–D4)

- [ ] M1 — stream capture, latest block wins, one retry, the unattended rule

### Task 1: `ReadStream` and `RunVerdict`
- [ ] Record fixture streams (a short live `claude -p --output-format stream-json --verbose` run trimmed by hand, plus hand-edited variants); table tests for every shape above, including the text fallback. Implement. Commit.

### Task 2: dispatch through the stream, with one retry
- [ ] Tests with a fake reviewer: verdict-then-postscript parses to the earlier verdict and findings (the Done-when); background-wait-only is retried once and the second run's verdict is used; two verdict-less runs record `unknown`, refuse to finalize, and the sidecar holds both runs. Implement in `BuildArgs`/`Dispatch`; update `FormatCommandLine` goldens. Commit.

### Task 3: the unattended rule
- [ ] Add the rule to `ContractPreamble`; regenerate goldens (`-update-golden`); the output-contract drift check stays green. Commit; `sdlc milestone-close --issue 300 --milestone M1`.

## M2 — Honest failures and a sized timeout (D5–D7)

- [ ] M2 — API failure is "did not run", protocol errors block, timeout scales with the window

### Task 4: API failure
- [ ] `APIFailure` table test over the evidence's signatures and a clean run; a dispatch test where the fake reviewer emits an `API Error … ERR_PROXY_TUNNEL` result → "review did not run" naming the sandbox action, nothing persisted, round cap unspent, no retry. Commit.

### Task 5: protocol errors block
- [ ] A boundary round with no findings block renders as blocked, not "passed", and doesn't count as convergence (`TestProtocolErrorRoundBlocks`). Commit.

### Task 6: sized timeout
- [ ] `ReviewTimeout` table (≤500 → 30m; 1,500 → 45m; huge → 2h; `WF_REVIEW_TIMEOUT` wins); the boundary caller passes the window's added lines (test via the dispatch options a fake records). Help text in `root.md`, `close.md`, `milestone-close.md`, `change-code.md`; atlas (judge section). Commit; `make test`; `sdlc milestone-close --issue 300 --milestone M2`; `sdlc close --issue 300`.
- [ ] Close #271 as absorbed (its Done-when is covered by Tasks 2–3): `sdlc issue set-status` per the lifecycle, with a Log line pointing here; tick its project row.
