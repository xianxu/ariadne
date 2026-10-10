# Boundary Review — ariadne#300 (milestone M1)

| field | value |
|-------|-------|
| issue | 300 — sdlc judge: a reviewer's late message replaces its verdict, so close records 'unknown' |
| repo | ariadne |
| issue file | workshop/issues/000300-sdlc-judge-a-reviewer-s-late-message-replaces-its-verdict-so-close-records-unknown.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | b51060b253657fa9a2c6827e330eb1e4c995f12e..c71732d979813be4b1c3057c2476795da09889ba |
| command | sdlc milestone-close --issue 300 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-09T20:50:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Inspected the dispatch path; now checking whether the close-level sidecar promise (Done-when #2) has a test, then running the judge tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 does what it set out to do. The claude reviewer now runs with `--output-format stream-json --verbose`. The pure `ReadStream` turns every assistant message into one joined text, so the existing parsers (each takes the last block) find the latest verdict and findings. `ParseVerdictToken` is now last-match. `Dispatch` retries a run with no verdict once, inside one deadline. `UnattendedRule` reaches all 8 goldens, and those golden diffs contain the rule and nothing else. `go test ./internal/judge/` passes.

Two cheap correctness gaps remain in the new failure classification, and neither blocks the boundary:
- **A verdict is thrown away when the result is an error.** A run whose terminal `result` has `is_error: true` is classified "review did not run" before the code checks for a verdict. A run that wrote its verdict and then failed on a later turn loses that verdict, and the error tells the operator to fix a sandbox the run never hit.
- **Codex and gemini output is parsed as a stream.** `ReadStream` is applied to every agent. A text-mode reviewer that quotes a stream-event JSON line (likely when reviewing this very diff) is misread as a stream, and the quoted line can trip `APIFailure`.

**Strengths**
- `Dispatch` still returns a plain string and reuses the parsers that already take the last block, so no per-message parser was added (ARCH-DRY; `dispatch.go:191-212`).
- `HasVerdict` (`stream.go:99`) matches everything any parser reads, including the legacy sentinels and the bare boundary token. The existing test (`"No DRY violations found."`) shows why that breadth is needed.
- The retry shares the deadline with the first attempt, and when both attempts fail the returned text keeps both runs, labelled.
- `ReadStream` falls back gracefully: a truncated stream keeps its unparsed tail, and output with no events is read as plain text (`stream.go:46-79`). All the fixtures are tested.
- The live conformance test now runs in stream mode and checks that the result is not an error.

**Critical:** none.

**Important**
1. `apifailure.go:25`: an `is_error` result wins over a verdict that is already present. It also covers more than API failures: in claude's stream, `error_during_execution` and `error_max_turns` set it too. **Fix:** check `HasVerdict` first, and when a verdict is present, record the error as a note rather than a failure. Otherwise keep D5's channel rule, but say "the run reported an error" instead of naming the sandbox, unless a signature such as `ERR_PROXY_TUNNEL` actually matches.
2. `dispatch.go:287`: `ReadStream` runs for every agent, but D1 says codex and gemini stay in text mode. Any standalone JSON line with a `type` key in their prose sets `Stream=true` and is dropped from `Text()`. A line like `{"type":"result","is_error":true,…}` (the fixtures' shape) produces a false `ErrAPIUnreachable`. That breaks D5's promise that the review's prose is never scanned (ARCH-SECURE). **Fix:** stream-parse only when the agent is claude; otherwise use `AgentRun{Tail: stdout}`. Add a test with a codex output that quotes `api_error.jsonl`.

**Minor**
- The error's remedy text "allow api.anthropic.com" is hard-coded for all agents (`dispatch.go:290`); codex and gemini use other hosts.
- If the retry hits the deadline, the error drops the first attempt's text, so the sidecar has nothing to read.
- The last-match change to `ParseVerdictToken` also affects the pre-merge `Classify` judges. A `VERDICT:` line quoted later in a findings section now overrides the leading one. This is acceptable given the contract, but it is a behavior change.
- Plan Task 2 says the tests drive "the close end to end … with its sidecar". The new tests stop at `Dispatch`. Done-when #2 (the sidecar points at the full output) holds only by following the code (`rr.Output = output`); no close-level test checks it.
- A stderr signature without a verdict is classified as a failure even after a clean exit. A transient retry warning on a run that then waited on a background job would skip the intended retry.

**Test coverage**
- The postscript, retry, double-failure, API-error and timeout-precedence cases are covered against the `Run` seam fake (ARCH-MOCK: pass).
- Two cases are missing: an `is_error` result after a verdict, and a text-mode agent quoting a stream line. Both match Important findings 1 and 2.
- No test checks that the retry's prompt carries `retryNotice`. A one-line check in the fake would cover it.

**Architecture**
- **ARCH-DRY: pass.** The parsers are reused. `HasVerdict` repeats `Classify`'s sentinel checks, but on purpose, to serve as a predicate.
- **ARCH-PURE: pass.** `ReadStream`, `HasVerdict` and `APIFailure` are pure, and `classifyRunResult` stays thin.
- **ARCH-PURPOSE: pass.** Codex and gemini are an explicitly declared non-goal; the sized timeout and the ledger fix are scheduled for M2.
- **ARCH-MOCK: pass.** The tests use the process-seam fake and the live check is updated.
- **ARCH-CONSTRAINTS: pass.** The retry is bounded to one attempt within one deadline.
- **ARCH-SECURE: flag.** See Important 2: text-mode prose is read on the stream channel.
- **ARCH-ORDER: pass.** The retry runs in sequence and no state is carried between events.
- **ARCH-FUNERAL: pass.** Nothing new is durable; the sidecar grows by at most one retry's text.

**Architectural notes for upcoming work**
- M2's D6 ledger fix and D7 sizing should plumb `DispatchOptions.Timeout` from `milestoneclose.go:640` and `planningreview.go:116`.
- `judge.go:177` (the pre-merge path) wraps `ErrAPIUnreachable` as "dispatch failed", which is fine.

**Plan revisions**
- Add a Revisions note that Task 2's tests run at the `Dispatch` level only. Either add a close-level sidecar test for Done-when #2 in M2, or reword the Task 2 claim.
- D5: record that a run carrying a verdict is not reclassified because of `is_error` (if adopted), and that stream parsing is limited to claude.

```findings
findings:
  - id: new
    severity: Important
    family: late-signal-erases-verdict
    title: |
      APIFailure treats any is_error result as "review did not run" even when the run already carries a verdict
    detail: |
      apifailure.go:25 checks is_error before HasVerdict, so a reviewer that wrote its verdict and then hit error_during_execution/max_turns/an overload loses it, and the error misdirects to a sandbox fix. Check HasVerdict first, or narrow the message when no signature matches; add a fixture test.
  - id: new
    severity: Important
    family: channel-vs-prose-separation
    title: |
      ReadStream runs on codex/gemini text output, so a quoted stream-event JSON line can trip ErrAPIUnreachable
    detail: |
      dispatch.go:287 stream-parses every agent's stdout; a standalone {"type":"result","is_error":true} line in text-mode prose sets Stream=true, is dropped from Text(), and fails the dispatch. Gate stream parsing on AgentClaude and add a test.
  - id: new
    severity: Minor
    family: agent-specific-remedy
    title: |
      The "allow api.anthropic.com" remedy is hard-coded for every agent
  - id: new
    severity: Minor
    family: fail-safe-keeps-evidence
    title: |
      A retry that hits the deadline drops the first attempt's text from the error/sidecar
  - id: new
    severity: Minor
    family: plan-claims-match-tests
    title: |
      Task 2 claims close-level end-to-end and sidecar tests; the tests stop at Dispatch
  - id: new
    severity: Minor
    family: late-signal-erases-verdict
    title: |
      A stderr API signature on a clean-exit, verdict-less run suppresses the intended retry
```

---

## Re-review — 2026-10-09T21:06:47-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 300 — sdlc judge: a reviewer's late message replaces its verdict, so close records 'unknown' |
| repo | ariadne |
| issue file | workshop/issues/000300-sdlc-judge-a-reviewer-s-late-message-replaces-its-verdict-so-close-records-unknown.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | b51060b253657fa9a2c6827e330eb1e4c995f12e..3476268f6010d6f12550ba91ba065c8a11c2366e |
| command | sdlc milestone-close --issue 300 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-09T21:06:47-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still reviewing: prior findings check out on the diff so far. Next I'm checking how `ParseVerdict` handles the joined multi-message text, then running the tests.

Tests pass and the three fix mutations each turn red. Last step: checking whether the atlas and contract doc kept up with the round-1 fix commit.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** All six round-1 findings are fixed, and I checked each against the code. Three of the fixes are behavior changes: putting `HasVerdict` first in `RunFailure`, reading only claude's output as a stream, and requiring a non-zero exit before a stderr signature counts. I reverted each one in a scratch worktree, and each time a test went red (`TestRunFailure` for the first and third, `TestDispatchParsesOnlyClaude` for the second). The judge package and the two new close-level tests pass at HEAD. The remaining problem is that the fix round changed the code's contract but not the prose that describes it. The atlas paragraph, the contract doc and the plan's Core concepts table and D5 still describe the code before round 1. That is cheap to fix and is the only Important finding. One Minor finding in the evidence family is still open.

1. **Strengths**
   - `RunFailure` (`cmd/sdlc/internal/judge/apifailure.go:38`) decides failure from the run's channel (the result event, the exit code, stderr) and never from the review's prose. It also never overrides a verdict the run already gave, so a late error result is handled the same way as a late postscript. That is the correct rule for the `late-signal-erases-verdict` family.
   - `HasVerdict` (`stream.go`) is one retry predicate built from the existing parsers, so every recipe shares one definition of "has a verdict" (ARCH-DRY).
   - `ReadStream` falls back gracefully. A truncated stream keeps its unparsed lines as the tail, and output with no stream events is read as plain text.
   - One deadline covers both attempts, and a retry that runs out of time still returns attempt 1 (`dispatch.go`, Dispatch's retry branch).
   - The tests run at two levels: `Dispatch` through the `Run` seam, and the close command end to end through `stubJudgeSeq`. The live conformance test now checks stream mode.

2. **Critical findings:** none.

3. **Important findings**
   - **The atlas, the contract doc and the plan describe the code as it was before round 1.**
     - `atlas/workflow/sdlc-binary.md:892-895` says any `is_error` result is `ErrAPIUnreachable` and names "the sandbox fix".
     - `construct/judge-output-contract.md` says "a run whose stream reports an error … is 'review did not run', never a verdict".
     - Both are now wrong in three ways:
       - A run that already gave its verdict keeps it.
       - A non-network error is `ErrReviewDidNotRun`.
       - Only claude's output is read as a stream.
     - The plan's Core concepts table lists `APIFailure(run, stderr, exitErr) (cause string, ok bool)`, but the code has `RunFailure(agent, run, stderr, nonZeroExit) error`. The table places `HasVerdict` in `classify.go`; it is in `stream.go`.
     - D5 still hard-codes the `api.anthropic.com` remedy for every agent.
     - This is the 2nd finding in family `plan-claims-match-tests`. **The rule:** a commit that changes a contract also changes every artifact that describes that contract (the plan's decisions and Core concepts, the atlas, the contract doc), in the same commit. Fix all three now, not one at a time.

4. **Minor findings**
   - **A retry that fails for any reason except a timeout still drops attempt 1.**
     - In `Dispatch`, `if err != nil || HasVerdict(second) { return second, err }` returns only the retry's error.
     - The cases are `ErrAPIUnreachable`, `ErrReviewDidNotRun`, an interrupt, and a launch failure on attempt 2.
     - `milestoneclose.go:641` discards the output whenever there is an error, so attempt 1's text is lost.
     - This is the 2nd finding in family `fail-safe-keeps-evidence`. **The rule:** once a retry has started, nothing that comes back without a verdict may drop attempt 1; any error from attempt 2 carries attempt 1's text.
   - The "## Attempt 1 (ended without a verdict)" label is built twice in `Dispatch`. A small helper would remove the duplication.
   - `RunFailure` returns `ErrAPIUnreachable` if stderr matches a network signature even when the cause came from the result event, for example a max-turns run with a stray `ENOTFOUND` from a tool. This is unlikely to matter.

5. **Test coverage notes.** The Done-when shapes are pinned at both the `Dispatch` level and the close level: verdict then postscript, and two runs without a verdict that keep the fail-safe with both runs in the sidecar. The round-1 fixes are mutation-verified. Nothing covers attempt 2 failing with an error (see the first Minor).

6. **Architectural notes for upcoming work.** These cover the 8 ARCH principles; seven pass and one is flagged.
   - **ARCH-DRY: pass.** The only nit is the duplicated label (Minor).
   - **ARCH-PURE: pass.** `ReadStream`, `RunFailure` and `HasVerdict` are pure.
   - **ARCH-PURPOSE: pass for M1.**
   - **ARCH-MOCK: pass.** The tests use the `Run` seam, and the live conformance test runs in stream mode.
   - **ARCH-CONSTRAINTS: pass.** One deadline covers the retry.
   - **ARCH-SECURE: pass.** Stream input degrades to plain text when it doesn't parse, and since BR-2 only claude's stdout is treated as a stream.
   - **ARCH-ORDER: flagged at Minor.** The retry is lexically bounded, but a failed attempt 2 drops attempt 1.
   - **ARCH-FUNERAL: pass.** The sidecar grows by at most one extra run per round.
   - **Note for M2 (D6).** When the retry runs out of time it now returns a nil error with no verdict. Until D6 lands, the boundary ledger will stamp that round `blocked: false`, so make sure D6 covers that path too.

7. **Plan revision recommendations.** Append to `## Revisions`: "2026-10-09 (M1 review round 1): `APIFailure` became `RunFailure(agent, run, stderr, nonZeroExit) error`, which returns nil when the run carries a verdict. A non-network error is `ErrReviewDidNotRun`. The remedy names each agent's own API host. `HasVerdict` lives in `stream.go`. Only claude's stdout is read as a stream. A stderr signature counts only with a non-zero exit." Update the Core concepts rows to match.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      RunFailure checks HasVerdict first; verdict_then_error.jsonl and maxTurns cases in TestRunFailure; reverting the guard turns the test red.
  - id: BR-2
    disposition: addressed
    note: |
      classifyRunResult reads only claude stdout as a stream; TestDispatchParsesOnlyClaudeAsAStream goes red when the gate is removed.
  - id: BR-3
    disposition: addressed
    note: |
      apiHosts maps each agent to its host; TestRunFailure asserts the codex and gemini remedies.
  - id: BR-4
    disposition: addressed
    note: |
      The retry's DeadlineExceeded branch returns attempt 1 labelled; TestDispatchRetryDeadlineKeepsTheFirstAttempt pins it.
  - id: BR-5
    disposition: addressed
    note: |
      TestCloseRecordsTheVerdictBeforeAPostscript and TestCloseWithoutAVerdictKeepsBothRuns exist in cmd/sdlc/closestream_test.go and pass.
  - id: BR-6
    disposition: addressed
    note: |
      The stderr case requires nonZeroExit; dropping that condition turns TestRunFailure red.
findings:
  - id: new
    severity: Important
    family: plan-claims-match-tests
    title: |
      Atlas, contract doc and plan Core concepts/D5 still describe the code before round 1
    detail: |
      atlas sdlc-binary.md:892 and construct/judge-output-contract.md say any is_error result is ErrAPIUnreachable with the sandbox fix. The plan table lists APIFailure(...)(cause, ok) with HasVerdict in classify.go, and D5 hard-codes api.anthropic.com. The code now keeps a verdict before an error, has ErrReviewDidNotRun, reads only claude as a stream, and names a host per agent. 2nd in family. Rule: a commit that changes a contract also updates every artifact describing it (plan, atlas, contract doc) in the same commit.
  - id: new
    severity: Minor
    family: fail-safe-keeps-evidence
    title: |
      A retry that fails with an error (not a timeout) drops attempt 1's text
    detail: |
      Dispatch returns (second, err) for ErrAPIUnreachable, ErrReviewDidNotRun, an interrupt or a launch failure on attempt 2, and milestoneclose.go:641 discards the output on any error. 2nd in family. Rule: once a retry has started, nothing that comes back without a verdict may drop attempt 1; any error from attempt 2 carries attempt 1's text.
```

---

## Re-review — 2026-10-09T21:19:42-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 300 — sdlc judge: a reviewer's late message replaces its verdict, so close records 'unknown' |
| repo | ariadne |
| issue file | workshop/issues/000300-sdlc-judge-a-reviewer-s-late-message-replaces-its-verdict-so-close-records-unknown.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | b51060b253657fa9a2c6827e330eb1e4c995f12e..fb54b7f5c20dc04766de599dbb0e3dba4ecde1df |
| command | sdlc milestone-close --issue 300 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-09T21:19:42-07:00 |
| verdict | SHIP |

## Review

Focus on the round-2 commit first.

```verdict
verdict: SHIP
confidence: high
```

Both open findings from round 2 are fixed, and I found nothing new that blocks the boundary. The round-2 commit (`fb54b7f5`) makes `Dispatch` keep attempt 1's text for any error from attempt 2, not only a timeout. All four callers (`judge.go:177`, `milestoneclose.go:640`, and the two plan-review paths through `planningreview.go:116`) now show that text when dispatch fails. The atlas paragraph, `construct/judge-output-contract.md`, the plan's D5 and Core concepts table, and a new `## Revisions` entry now describe the code as it stands: `RunFailure`, `ErrReviewDidNotRun` vs `ErrAPIUnreachable`, a host per agent, stream parsing for claude only, and the retry keeping its evidence. The judge package tests and the close-level stream tests pass at HEAD.

1. **Strengths**
   - `dispatch.go`'s retry branch now covers the whole class "an error once a retry has started", not just the timeout instance. Only `DeadlineExceeded` turns into the fail-safe with a nil error; every other error is returned alongside the kept text. This is the rule BR-8 stated, applied to the class.
   - All four `Dispatch` callers were swept, including `dispatchPlanningReview`, which used to return `""` on error (ARCH-PURPOSE: the class, not the one site named).
   - D5 was rewritten with the original wording kept underneath it, and Revisions records the delta (BR-7's rule followed).
   - `RunFailure` never overrides a verdict the run already gave, and it decides only from the run's channel (result event, exit code, stderr), never from the review's prose.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings** (all three are notes only; none is filed as a finding)
   - The callers that print the kept text (`milestoneclose.go:642`, `changecode.go:652/890`, `judge.go:179`) have no test. `TestDispatchRetryErrorKeepsTheFirstAttempt` pins the `Dispatch` half only; reverting the old `(second, err)` return turns it red. A caller that went back to dropping the output would pass every test.
   - The atlas limits the stderr-signature rule to "codex and gemini". In the code, the stderr branch (`apifailure.go`, `nonZeroExit && apiFailureRE.Match(stderr)`) also applies to a claude run that exits non-zero with no result event. The difference is small; fix the wording when the text is next edited.
   - Plan Task 4 is ticked and still names `APIFailure`. Revisions explains the rename, so this is cosmetic.

5. **Test coverage notes.** Covered:
   - Dispatch-level: verdict then postscript; two runs without a verdict; a retry that times out; a retry that errors; an error after a verdict; prose that quotes the failure signatures.
   - Close-level: `closestream_test.go`.

   Not covered: how callers present the text on error (the first Minor).

6. **Architectural notes.**
   - **ARCH-DRY: pass.** The "## Attempt 1" label is still built twice in `Dispatch`. A small helper would clean that up.
   - **ARCH-PURE: pass.** `ReadStream`, `HasVerdict` and `RunFailure` are pure.
   - **ARCH-PURPOSE: pass.** All callers were swept.
   - **ARCH-MOCK: pass.** The tests use the `Run` seam, and the live conformance test now runs in stream mode.
   - **ARCH-CONSTRAINTS: pass.** One deadline covers the retry.
   - **ARCH-SECURE: pass.** Only claude's stdout is read as a stream, and output that isn't a stream degrades to plain text.
   - **ARCH-ORDER: pass.** The retry is lexically bounded, and every outcome of attempt 2 keeps attempt 1.
   - **ARCH-FUNERAL: pass.** The sidecar gains at most one extra run per round.
   - **For M2 (D6):** a retry that times out returns a nil error without a verdict, so make sure the boundary ledger stamps that round as blocked.

7. **Plan revision recommendations:** none needed. When Task 4 is next touched, the `APIFailure` name could be updated in place.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      Atlas sdlc-binary.md (#300 paragraph), judge-output-contract.md, plan D5 and the Core concepts row now describe RunFailure, ErrReviewDidNotRun, per-agent hosts, claude-only stream and the verdict-first guard; Revisions records the change. Checked against apifailure.go/dispatch.go.
  - id: BR-8
    disposition: addressed
    note: |
      Dispatch returns attempt 1's labelled text with any error from attempt 2; TestDispatchRetryErrorKeepsTheFirstAttempt fails under the old (second, err) return; all four callers print the output on error.
```
