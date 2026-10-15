# Boundary Review — ariadne#387 (whole-issue close)

| field | value |
|-------|-------|
| issue | 387 — Agent: trailer says 'human' (not 'sdlc') for a commit made outside an agent session |
| repo | ariadne |
| issue file | workshop/issues/000387-agent-trailer-says-human-not-sdlc-for-a-commit-made-outside-an-agent-session.md |
| boundary | whole-issue close |
| milestone | — |
| window | 84d3981d4d486611e70c77a20848d5270c24a13c..83cf89adaae2526d9c5d35ae228dda063a84f75a |
| command | sdlc close --issue 387 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-15T15:03:36-07:00 |
| verdict | SHIP |

## Review

The diff is small and the tests pass, so this review comes back as SHIP.

```verdict
verdict: SHIP
confidence: high
```

The change does what the issue asks. When no agent session is found, the trailer now says `Agent: human`, set by the new constant `agentid.Human`. Older history that says `sdlc` still parses (`agentid.Legacy`) to the same empty human identity. I ran the tests in the review checkout and they pass: `go test ./cmd/sdlc/internal/agentid/` and `go test ./cmd/sdlc/ -run 'TestEveryCommitSiteStamps|AgentID'`. All three Done-when clauses are covered:
- **Plain-shell fallback:** the plain-shell cases in `detect_test.go` and `qoder_test.go` now expect `Human`.
- **Parse of both values:** `human` is parsed through the round-trip case in `agentid_test.go:13`, and the old `sdlc` value through the explicit check at `agentid_test.go:23-26`.
- **Commit-site sweep:** it still passes.

The docs gate is met: `AGENTS.base.md` §12, `helptext/agent-id.md`, the atlas `sdlc-binary.md` row and the `detect.go` override comment are all updated. I grepped `cmd/` and `pkg/`, and the Log's claim that only `Parse` and the agentid tests read the value holds. `closeagent.go` and `MissingAgent` only check whether the trailer is present, not what it says.

1. **Strengths**
   - There is one source for each value, and `String` and `Parse` both use the constants (`agentid.go:18-24, 35, 47`).
   - Tests use the constant rather than repeating the literal, so the expected value is defined in one place.
   - Keeping the old `sdlc` value readable avoids rewriting history, and a dedicated assertion covers it.
   - Doc comments were updated next to the code they describe (`Identity`, `OverrideEnv`).

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - The check that old `sdlc` history still parses compares against `Parse(Legacy)`, not the literal `"sdlc"`. If someone edits the constant, the test changes with it and stays green even though old trailers stop parsing. The only instance in this window is `agentid_test.go:24`.

5. **Test coverage notes**
   - Both states are covered: a session found (it gets an agent/model/effort value) and no session (it gets `Human`). Parse is tested for `human`, `sdlc` and malformed input.
   - The `sdlc agent-id` command test still uses a pinned override. The plain-shell output of the command is covered one layer down, in `Detect` and `String`.

6. **Architectural notes**
   - **ARCH-DRY: pass.** Each value is defined once as a constant, and nothing is duplicated.
   - **ARCH-PURE: pass.** The changes are confined to the pure `String`/`Parse` functions and their table tests; no IO changed.
   - **ARCH-PURPOSE: pass.** I checked every reader of the trailer value: `Parse` is the only one that interprets it, and `closeagent` and `MissingAgent` only test for presence. Every place that writes the trailer goes through `Identity.String`, so all of them now produce `human`.

7. **Plan revision recommendations:** none.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: legacy-compat-pinned-by-literal
    title: |
      Legacy-parse test references the Legacy constant rather than the literal "sdlc"
    detail: |
      agentid_test.go:24 calls Parse(Legacy). If someone edits the constant, the test changes with it and passes even though existing trailers in history no longer parse. Asserting Parse("sdlc") pins the history format. This is the only instance in the window.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e6b4e583-9d58-43fa-be54-405be5c59207","repository":"github.com/xianxu/ariadne","issue_id":"000387","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c440b4c4351bdb59","reviewed_head":"83cf89adaae2526d9c5d35ae228dda063a84f75a","slot":"ariadne:4","requested_at":"2026-10-15T22:00:38.937282Z"},"phase":"finished","pid":6587,"started_at":"2026-10-15T22:00:39.429975Z","finished_at":"2026-10-15T22:02:59.007037Z","session_id":"773f0f51-7b19-4c99-971b-18a7a7ade280","outcome":"returned"},{"version":1,"intent":{"disposable_root":true,"id":"0488afdc-9e99-478a-b261-e51655f9586d","repository":"github.com/xianxu/ariadne","issue_id":"000387","recipe":"small-diff-review","round":1,"attempt":2,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c440b4c4351bdb59","reviewed_head":"83cf89adaae2526d9c5d35ae228dda063a84f75a","slot":"ariadne:4","requested_at":"2026-10-15T22:03:00.147758Z"},"phase":"finished","pid":38483,"started_at":"2026-10-15T22:03:00.280015Z","finished_at":"2026-10-15T22:03:35.779504Z","session_id":"daea5e58-2cc2-495e-993f-8c6409485dba","outcome":"returned"}] -->

## Review preparation and protocol diagnostics

Quoted values are diagnostic data, not review verdicts.
Protocol attempt 1
reason: "missing valid findings block"
output: "Tests pass. Next I'm checking the remaining doc mentions and the fixture value.\n\n```verdict\nverdict: SHIP\nconfidence: high\n```\n\nThis change does what ariadne#387 asks. When `sdlc` runs with no agent session, it now writes `Agent: human`. `Parse` accepts both `human` and the old `sdlc` value and maps each to the empty (human) `Identity`. The value is defined once, as the constants `agentid.Human` and `agentid.Legacy`. The help text, AGENTS.base.md §12 and the atlas row all say the same thing. I checked who reads the value: `agentid.Parse` has no callers outside the `agentid` package. Its only other caller there is the `SDLC_AGENT_ID` override in `detect.go:32`. `closeagent.go` only checks whether the trailer is present. So the log's claim that \"the only readers are `Parse` and the agentid tests\" holds. I ran the `agentid` package tests and the matching `cmd/sdlc` tests (Agent/Detect/Qoder/Identity/Stamp/CommitSite, which includes `TestEveryCommitSiteStamps`), and all passed. Nothing blocks SHIP.\n\n1. **Strengths**\n   - There is one place where the value is defined (`cmd/sdlc/internal/agentid/agentid.go:18-25`). `String` and `Parse` both use the constants, and the tests use `Human` rather than the literal string (`detect_test.go:55-63`, `qoder_test.go:39-40`) (ARCH-DRY: pass).\n   - Old trailers keep working: `Parse(Legacy)` is pinned to the human identity (`agentid_test.go:23-26`), so history from #367 still groups correctly.\n   - All five plain-shell, no-session cases in detect and qoder were updated, so each Done-when case is covered:\n     - the fallback renders `human`;\n     - `human` parses (round trip);\n     - legacy `sdlc` parses;\n     - the commit-site sweep passes.\n   - The help text, AGENTS.base.md §12 and the atlas row agree with the code.\n\n2. **Critical:** none.\n3. **Important:** none.\n4. **Minor:** none worth recording. One side effect: `SDLC_AGENT_ID=sdlc` now quietly comes out as `human`. That fits the intent and nothing documents `sdlc` as an override value.\n5. **Test coverage:** `Human` is covered both ways (rendered and parsed) and `Legacy` is covered when parsed. Detect's no-session paths were run with the new value, including the malformed-override fallback, the plain Pair pane and plugin-alias-only environments. The commit-site sweep ran and passed.\n6. **Architecture:**\n   - **ARCH-DRY:** pass. The constants are the only definition.\n   - **ARCH-PURE:** pass. Rendering and parsing stay pure; `detect.go` was changed only in a comment.\n   - **ARCH-PURPOSE:** pass. I checked every place that produces or reads the value: `String` (via `StampCurrent`, which every commit site uses), `Parse`, the override, the docs and the atlas. Each one gets the value from the constants or states it to match. If a future reader groups commits by agent (#365) without calling `Parse`, it will need to treat `sdlc` as `human`. `Parse` already provides that.\n7. **Plan revisions:** none. The plan matches the code.\n\n```findings\n```"

truncated: false
retained-bytes: 0000003225
