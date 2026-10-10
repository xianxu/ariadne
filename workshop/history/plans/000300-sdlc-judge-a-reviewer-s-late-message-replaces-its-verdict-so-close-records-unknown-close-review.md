# Boundary Review — ariadne#300 (whole-issue close)

| field | value |
|-------|-------|
| issue | 300 — sdlc judge: a reviewer's late message replaces its verdict, so close records 'unknown' |
| repo | ariadne |
| issue file | workshop/issues/000300-sdlc-judge-a-reviewer-s-late-message-replaces-its-verdict-so-close-records-unknown.md |
| boundary | whole-issue close |
| milestone | — |
| window | b51060b253657fa9a2c6827e330eb1e4c995f12e..8b812fd64dc6d3a06d68139061ee1ad599c72b62 |
| command | sdlc close --issue 300 |
| reviewer | claude |
| timestamp | 2026-10-09T21:59:10-07:00 |
| verdict | SHIP |

## Review

Reviewing round 7 of #300: I've read the diff and am now running the targeted tests.

```verdict
verdict: SHIP
confidence: high
```

This is round 7, and I found nothing that blocks shipping. I checked the pinned window b51060b2..8b812fd6 in this order: stat, name-status, a full read of the judge, gatestate, milestone-close and gatepersist changes, then the tests themselves. At HEAD 8b812fd6 with a clean tree, `go test` passes on `internal/judge` and `internal/gatestate`. The targeted `cmd/sdlc` close, boundary-ledger and convergence tests also pass. Every test named in the plan's Tasks 1–6 exists under that name. The core design holds up:
- claude's output is read as a stream (`ReadStream`), and the review text is every assistant message joined in order.
- All three parsers take the last verdict block, findings block or `VERDICT:` line, so a postscript written after the verdict no longer replaces it.
- A run with no verdict anywhere is retried once inside the same deadline.
- A run that failed before reviewing becomes a typed "review did not run" error, decided from the run's own channel and never from the review's prose.

All 14 earlier findings stay disposed; I raise one new Minor.

1. **Strengths**
   - `RunFailure` (`internal/judge/apifailure.go:36`) returns nil early when the run already has a verdict. That one rule stops both a late postscript and a late error result from erasing a verdict.
   - `HasVerdict` (`stream.go:95`) is the only retry check, and it covers every verdict form any recipe produces. So plan-quality's CLEAN/INFO/FAILURE and the legacy sentinels are never retried by mistake.
   - Timeout precedence is stated once, in `EffectiveTimeout` (`dispatch.go:363`). The dispatch line and `Dispatch` both read it, so the reported limit can't disagree with the real one.
   - When a retry fails, attempt 1's text is kept under labels (`dispatch.go:211-223`), and all four `Dispatch` callers print it.
   - Only claude's output is parsed as a stream (`classifyRunResult`). A codex or gemini review that quotes an event line can't fail itself.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `gatepersist.go:63`: `Forced` is now stamped from the combined `blocked` value instead of `d.Block`, the gate's actual decision.
     - **When it goes wrong:** the review gives a SHIP verdict but no findings block, and the close runs with the global `--force` for some other gate. The round is recorded as a boundary-gate waiver even though this gate never refused. This is the case BR-42 was fixed to prevent.
     - **Family:** `one-round-outcome-all-readers` (this would be its 3rd finding).
     - **The rule:** `Round.Blocked`/`Forced` should record only what the gate decided. "Produced nothing" is a separate property of the round, and every reader should get it from `Round.ProducedNothing()` rather than from an overloaded `Blocked` flag. The cheapest fix is `forcedRationale(forced, d.Block)`.

5. **Test coverage notes**
   - The fake-reviewer seam drives `Dispatch` from recorded and hand-edited stream fixtures. Postscript, background-wait, verdict-then-error and truncated streams are all covered.
   - Tests for retry-after-deadline and retry-after-error both check that attempt 1 survives.
   - No test covers a findings-less SHIP round combined with `--force` (the Minor above).

6. **Architecture notes**
   - **ARCH-DRY:** pass. One `HasVerdict`, one `ProducedNothing`, one `EffectiveTimeout`.
   - **ARCH-PURE:** pass. `ReadStream`, `RunFailure`, `HasVerdict` and `ReviewTimeout` are pure and unit-tested without IO; process IO stays in `dispatchOnce`.
   - **ARCH-PURPOSE:** pass. All four `Dispatch` callers show retry text, and the docs (README, help, atlas, contract doc) match the code.
   - **ARCH-MOCK:** pass. The fake reviewer script is the seam, and the opt-in live conformance test now runs in stream mode.
   - **ARCH-CONSTRAINTS:** pass. One deadline spans the retry, and the sized timeout is capped at 2h.
   - **ARCH-SECURE:** pass. Stream lines that don't parse are kept as plain text rather than trusted, and a stream is never parsed from a non-claude agent.
   - **ARCH-ORDER:** pass, with the Minor above. The round outcome is one value, though `Forced` reads it a little too broadly.
   - **ARCH-FUNERAL:** pass. Nothing new is durable beyond the existing sidecar, which grows by at most one extra run per round.

7. **Plan revision recommendations:** none. The plan's Revisions already record that D6 is only a stamp on the ledger record and that D7 sizes boundary reviews only.

```findings
findings:
  - id: new
    severity: Minor
    family: one-round-outcome-all-readers
    title: |
      stampAndPersist stamps Forced from the combined blocked flag, recording a waiver for a gate that did not refuse
    detail: |
      gatepersist.go:63 now calls forcedRationale(forced, blocked), where blocked = d.Block || g.Blocked. A findings-less round with a SHIP verdict, closed under the global --force for some other gate, gets a boundary-gate waiver stamped although the decision never blocked (the BR-42 contract). Rule: Blocked and Forced record only the gate's decision; produced-nothing is a separate round property that every reader takes from Round.ProducedNothing(). Cheapest fix: forcedRationale(forced, d.Block), plus a regression test.
```
