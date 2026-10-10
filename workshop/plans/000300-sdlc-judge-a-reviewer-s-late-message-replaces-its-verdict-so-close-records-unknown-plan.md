# Review verdicts survive late messages, backgrounding, network blocks and timeouts — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A boundary or plan-quality review either yields the verdict the reviewer actually gave, or reads "review did not run" with the cause, never `unknown` from a late postscript and never "passed" from a blocked API (#300, absorbing #271, D2, D3 of the 2026-10-09 evidence pensive).

**Architecture:** The claude reviewer runs with `--output-format stream-json`, and a new pure reader turns the event stream into a run record: every assistant message's text, the run's terminal result and its error flag. `Dispatch` keeps returning one text, now all assistant messages joined in order, so the existing last-block-wins parsers take the latest verdict and findings and a later block-less message erases nothing. Inside `Dispatch`, a run with no verdict anywhere is dispatched once more, and a run whose stream reports an API or network error is a dispatch failure ("review did not run"), never a round. On the boundary path a round with a protocol error is recorded as blocking. The review timeout scales with the window's added lines.

**Tech Stack:** Go (`cmd/sdlc/internal/judge`, `internal/gatestate`, `cmd/sdlc` close/milestone-close/change-code), recorded stream fixtures under `internal/judge/testdata/`, the opt-in live conformance test pattern of `live_stream_conformance_test.go`.

---

## Non-goals

- Structured capture for codex and gemini: they keep text mode (their final output is their whole answer today). The retry and the network-failure classification apply to them through their text output.
- Making the sandbox allow the API host: that is the operator's sandbox configuration. sdlc names the cause and the next action.
- D4–D9 of the pensive (round-cap demotion, reconcile running unmerged finishers, uncommitted-fix reviews, detached checkouts, ledger files left uncommitted, no-code reviews): separate items.

## Design decisions

- **D1 — Stream capture for claude.** `BuildArgs` (`internal/judge/dispatch.go:130`, today `claude -p` in text mode, which prints only the final message) adds `--output-format stream-json --verbose` for claude. A pure `ReadStream(stdout) Run` parses the newline-delimited events: the text parts of each `assistant` message in order (decoded from JSON, so fenced blocks arrive intact), and the terminal `result` event (`subtype`, `is_error`, `result`). Lines that don't parse are kept as raw text after the messages parsed so far, so a truncated stream loses nothing; only when no line parses at all is the whole stdout treated as text, as today.
- **D2 — Latest block wins, through the existing parsers (ARCH-DRY).** `Dispatch`'s return contract is unchanged in type: one string, now the run's assistant messages joined in order (`Run.Text()`). The existing parsers already take the last block of each kind (`ParseVerdictBlock`, `classify.go:182-188`; `gatestate.ParseFindingsBlock`, `internal/gatestate/parse.go:28`), so a block-less later message, the pair#362 and pair#387 shape, changes neither, and no per-message parser or judge→gatestate import is added. The one parser that takes the first match, the `VERDICT:` line fallback (`ParseVerdictToken`, `classify.go:62-74`), becomes last-match so all three agree on "latest wins". The sidecar's review text is that same joined text.
- **D3 — No verdict anywhere: retry once, then `unknown`. `Dispatch` owns the retry.** The predicate is a new pure `HasVerdict(text)`: a valid ```verdict block (`ParseVerdictBlock`, `internal/judge/classify.go:182`, which accepts every emitted token) or a `VERDICT:` line (`ParseVerdictToken`, `classify.go:62`, whose pattern covers CLEAN, INFO, FAILURE, SHIP, FIX-THEN-SHIP, REWORK and BLOCK). It deliberately isn't `ParseVerdict != VerdictUnknown`: that maps only the boundary set (`classify.go:148-153`), so every plan-quality round would read verdict-less. One predicate for every recipe, owned by `Dispatch`, so no caller decides it. A completed run without a verdict (#271's "I'll wait for the background run to notify") is dispatched once more, with the same prompt and a one-line notice that the previous attempt ended without a verdict. One deadline covers both attempts (the retry gets what is left of D7's limit), so the worst case stays the sized limit, not double. If the second run also has none, today's fail-safe stands: `unknown`, the close doesn't finalize, and the returned text holds both runs, separated and labelled, so the sidecar keeps them.
- **D4 — The reviewer is told it runs unattended.** `ContractPreamble` (shared by every recipe) gains: you run non-interactively and receive no notifications; run commands in the foreground within the time limit; your final message must carry the verdict and findings blocks. Allowed tools stay `Read,Grep,Glob,Bash`: Claude Code can't deny Bash's background parameter alone, so this is the prompt rule plus D2/D3's capture making a late message harmless.
- **D5 — An unreachable API is "review did not run", judged on the channel, not the prose.** The review's own text can quote these strings (this issue's diff will), so assistant messages are never scanned. A run is a dispatch failure when the stream's terminal `result` event has `is_error: true` (its `result` text names the cause), or, for text-mode agents (codex, gemini) and a claude run whose stream didn't parse, only when `HasVerdict` is false AND the exit was non-zero (today swallowed in `classifyRunResult`, `dispatch.go:240-263`) or stderr matches the API-failure signatures from the evidence (`API Error`, `ERR_PROXY_TUNNEL`, `Connection blocked by network allowlist`, `ECONNREFUSED`, `ENOTFOUND`). The error names the cause and the next action: "the reviewer could not reach its API (…); in an agent sandbox, allow `api.anthropic.com` for this command, then rerun". Like today's dispatch errors, nothing is persisted and the round cap isn't spent. It isn't retried: the block would repeat.
- **D6 — A boundary protocol-error round is never "passed".** On the boundary path only: `persistBoundaryRound` writes `Blocked: true` for a round without a findings block (`boundaryledger.go:167`), and `stampAndPersist` then overwrites it with `d.Block` (`gatepersist.go:50`). The boundary caller stamps `Blocked = d.Block || round.ProtocolError != ""`, so such a round renders as blocked and isn't convergence. `stampAndPersist` stays shared and unchanged in contract (it feeds `PassesUnchanged`); plan-quality keeps its own rule (blocked when its verdict can't be classified, `changecode.go:673`), which this issue doesn't change. This is the ledger half of D2.
- **D7 — The timeout scales with the window.** Today a fixed 30 min, 1s–2h accepted (`reviewTimeout`, `dispatch.go:293`). Default = 30 min for a window of up to 500 added lines, plus 15 min per further 1,000 added lines, capped at 2 h (the existing ceiling). The callers compute the added lines of the review window with `git diff --numstat base..head` (boundary) or the durable plan's line count (plan-quality) and pass `DispatchOptions.Timeout`. `WF_REVIEW_TIMEOUT` still overrides. The heartbeat line shows the limit in use.

ARCH-FUNERAL: creates nothing durable beyond the existing sidecar, which grows by the retried run's text when a retry happens (at most one extra run per round).

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Run` / `ReadStream` | `cmd/sdlc/internal/judge/stream.go` | new |
| `HasVerdict(text) bool` | `cmd/sdlc/internal/judge/classify.go` | new |
| `APIFailure(run, stderr, exitErr) (cause string, ok bool)` | `cmd/sdlc/internal/judge/apifailure.go` | new |
| `ParseVerdictToken` | `cmd/sdlc/internal/judge/classify.go` | modified (last match) |
| `ReviewTimeout(addedLines int) time.Duration` | `cmd/sdlc/internal/judge/dispatch.go` | modified (`reviewTimeout` gains the size input) |
| `ContractPreamble` | `cmd/sdlc/internal/judge/contract.go` | modified |

Each is unit-tested without IO; `ReadStream` against recorded streams in `internal/judge/testdata/stream/`.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `BuildArgs` | `internal/judge/dispatch.go` | modified | claude stream flags |
| `Dispatch` / `classifyRunResult` | `internal/judge/dispatch.go` | modified | stream read, API failure, one retry |
| `DispatchOptions.Timeout` | `internal/judge/dispatch.go` | new field | — |
| `dispatchBoundaryReview` | `cmd/sdlc/milestoneclose.go` | modified | window size → timeout |
| plan-quality dispatch | `cmd/sdlc/changecode.go` | modified | plan size → timeout |
| `persistBoundaryRound` / `stampAndPersist` | `cmd/sdlc/boundaryledger.go`, `gatepersist.go` | modified | protocol error blocks |

The fake reviewer in the judge tests (a script printing a recorded stream) is the seam. The existing opt-in live conformance test (`live_stream_conformance_test.go:45`, which asserts claude's stdout is exactly `STREAM_OK`) moves to stream mode: it asserts `ReadStream` recovers `STREAM_OK` as the run's text and a non-error result, and that stderr stays separate.

## M1 — Verdict capture (D1–D4)

- [ ] M1 — stream capture, latest block wins, one retry, the unattended rule

### Task 1: `ReadStream`
- [x] Record a short live stream and hand-edited variants as fixtures; a table test over each shape the decisions name (clean, postscript, background-wait, error result, truncated, not a stream). Implement; `ParseVerdictToken` last-match with its test. Commit.

### Task 2: dispatch through the stream, with one retry
- [x] Fake-reviewer tests drive `Dispatch` and the close end to end: the Done-when shape (verdict, then postscript), the retry path, and the two-failures fail-safe with its sidecar. Implement in `BuildArgs`/`Dispatch`; update `FormatCommandLine` goldens and the live conformance test. Commit.

### Task 3: the unattended rule
- [x] Add the rule to `ContractPreamble`; regenerate goldens (`-update-golden`); the output-contract drift check stays green. Commit; `sdlc milestone-close --issue 300 --milestone M1`.

## M2 — Honest failures and a sized timeout (D5–D7)

- [ ] M2 — API failure is "did not run", protocol errors block, timeout scales with the window

### Task 4: API failure
- [x] `APIFailure` table test, including a review whose prose quotes the signatures (must not trip); a fake-reviewer dispatch with an error result reads "review did not run" with the sandbox action, persists nothing and doesn't retry. Commit.

### Task 5: protocol errors block
- [ ] A boundary round with no findings block renders as blocked, not "passed", and doesn't count as convergence (`TestProtocolErrorRoundBlocks`). Commit.

### Task 6: sized timeout
- [ ] `ReviewTimeout` table test, and a test that the boundary caller passes the window's size and one deadline spans the retry. Help text in `root.md`, `close.md`, `milestone-close.md`, `change-code.md`; atlas (judge section). Commit; `make test`; `sdlc milestone-close --issue 300 --milestone M2`; `sdlc close --issue 300`.
- [ ] Close #271 as absorbed (its Done-when is covered by Tasks 2–3): `sdlc issue set-status` per the lifecycle, with a Log line pointing here; tick its project row.

## Revisions

- 2026-10-09 (plan-quality round 1): `Dispatch` keeps a string contract (the joined messages), so the existing last-block-wins parsers do D2 and `RunVerdict` is dropped; `Dispatch` owns the retry with the shared `ParseVerdict` predicate, and `ParseVerdictToken` becomes last-match (PQ-1). D6 is scoped to the boundary path; plan-quality's rule is unchanged (PQ-2). D5 judges the stream's error flag and stderr, never the review's prose (PQ-3). One deadline spans the retry (PQ-4). A truncated stream keeps what parsed, and the live conformance test moves to stream mode (PQ-5). Test prose compressed (PQ-6).
- 2026-10-09 (plan-quality round 2, PQ-7): the retry predicate is a new `HasVerdict` (a valid verdict block or a `VERDICT:` line, so plan-quality's CLEAN/INFO/FAILURE count), not `ParseVerdict != VerdictUnknown`, which knows only the boundary set. Every existing-behavior claim in the decisions now carries its file:line.
- 2026-10-09 (M1 build): the type is `AgentRun` (the package already had a `Run` process seam). `HasVerdict` also accepts the legacy sentinels `Classify` reads and `ParseVerdict`'s bare boundary token: an existing dispatch test showed that 'No DRY violations found.' would otherwise be retried. Task 4 (API failure) and `dispatchTimeout`'s precedence landed in M1, since they live in the same dispatch function; M2 keeps the boundary ledger (D6) and the callers' sized timeout (D7).
