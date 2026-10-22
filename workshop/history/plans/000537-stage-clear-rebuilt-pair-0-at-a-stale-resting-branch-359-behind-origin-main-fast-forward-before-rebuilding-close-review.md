# Boundary Review — ariadne#537 (whole-issue close)

| field | value |
|-------|-------|
| issue | 537 — stage --clear rebuilt pair:0 at a stale resting branch (359 behind origin/main); fast-forward before rebuilding |
| repo | ariadne |
| issue file | workshop/issues/000537-stage-clear-rebuilt-pair-0-at-a-stale-resting-branch-359-behind-origin-main-fast-forward-before-rebuilding.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8827a5b19a3180775ed57c4da9fb36f1d5da4880..89efa47cbf96720740ec42ad7b981a8a41003013 |
| command | sdlc close --issue 537 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T21:06:30-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned changes fulfill #537’s Spec and Plan. Both detached clear and already-resting retry fetch main, reject unsafe ancestry, fast-forward, and rebuild against the same immutable pin. Targeted tests passed; no blocking defects or documentation gaps were found.

```findings
{}
```

1. **Strengths**
   - The stage lock spans observation through rebuild, preserving existing ownership checks (`cmd/sdlc/stage.go:148`).
   - Ancestry refusal precedes checkout; checkout and build verification use the fetched pin (`cmd/sdlc/stage.go:181–241`).
   - Real Git fixtures cover remote advancement, failed-build retry, numbered slots, file collisions, and intermediate resting attributes.
   - README, help, recovery catalog, and atlas consistently describe the changed clear and dry-run behavior.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes:** Independently passed:
   - Stage and stage-rule tests: `go test ./cmd/sdlc -run 'TestStage|TestCheckStage|TestCheckUnstage' -count=1` (52.545s).
   - Selected recovery and Git snapshot tests.
   - Pinned-range `git diff --check`.

   Coverage exercises both detached and resting success paths, online/offline dry runs, refusal states, and preservation controls. The full suite was not independently rerun.

6. **Architectural notes**
   - **ARCH-DRY — pass:** Both clear paths share the sequence and reuse snapshot, asset, and build helpers.
   - **ARCH-PURE — pass:** The change remains small INTEGRATION orchestration; no pure entity acquires IO.
   - **ARCH-PURPOSE — pass:** Freshness is enforced before rebuilding, including retries.
   - **ARCH-EVOLVE — pass:** No persisted format or field meaning changes.

7. **Plan revision recommendations:** None. The recorded dry-run revision matches the implementation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d0b534d6-cb77-4a46-926f-a5cbf7a43b24","repository":"github.com/xianxu/ariadne","issue_id":"000537","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/786dd4cb3ed41565","reviewed_head":"89efa47cbf96720740ec42ad7b981a8a41003013","slot":"ariadne:2","requested_at":"2026-10-22T04:04:45.179755Z"},"phase":"finished","pid":17994,"started_at":"2026-10-22T04:04:45.30094Z","finished_at":"2026-10-22T04:06:30.035865Z","session_id":"01a161c8-658b-7a20-a99a-f4609576ff0b","outcome":"returned"}] -->
