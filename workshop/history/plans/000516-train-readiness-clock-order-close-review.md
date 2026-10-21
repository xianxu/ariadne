# Boundary Review — ariadne#516 (whole-issue close)

| field | value |
|-------|-------|
| issue | 516 — Train boarding rejects readiness refreshed during conductor checks |
| repo | ariadne |
| issue file | workshop/issues/000516-train-readiness-clock-order.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8544ca6709c2a1fb0609b6e40a8f7e85169478f8..028aa04a505f6a054055a63d1b9ecb11f8a11cac |
| command | sdlc close --issue 516 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T23:37:05-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned change fixes the snapshot/clock ordering without weakening freshness checks or altering the archive fixture. The regression fails with the old ordering and passes at head. No blocking findings.

Done-when: current — the revisions clarify verification sequencing; the existing criteria still cover the implemented behavior.

```findings
{}
```

1. **Strengths**
   - `cmd/sdlc/landingtrain.go:700` samples time after readiness reads and uses it consistently for boarding and departure.
   - `cmd/sdlc/landingtrain_readiness_test.go:22` exercises refreshed, newly expired, and genuinely future vouches.
   - Assertions verify actual train creation, merge results, and both cars’ final states through the existing stateful GitHub fake.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage**
   - Targeted regression, archive, three-car variants, single-car, and grant controls passed in 48.550s.
   - Pure freshness, boarding, admission, and scheduling controls passed.
   - A temporary Go overlay restored the base implementation: all three regression cases failed as expected. Temporary files were removed; the checkout remains clean.
   - Inspected the recorded full-suite summary: all eight shards and other packages passed; its SHA256 matches the issue’s evidence.

6. **Architecture**
   - **ARCH-DRY: pass.** Reuses existing clock, boarding logic, and fixtures; one production `Board` consumer.
   - **ARCH-PURE: pass.** Clock sampling stays in the integration shell; freshness and boarding remain pure.
   - **ARCH-PURPOSE: pass.** Corrects production ordering and preserves archive outcome/cleanup conformance.
   - **ARCH-EVOLVE: pass.** No persisted-format changes.
   - No new architectural or user-facing surface requires atlas or README updates.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b04dcd05-2a95-4e40-8e12-7b774864fb41","repository":"github.com/xianxu/ariadne","issue_id":"000516","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/4ec87a01a2d9cd06","reviewed_head":"028aa04a505f6a054055a63d1b9ecb11f8a11cac","slot":"ariadne:1","requested_at":"2026-10-21T06:35:06.73345Z"},"phase":"finished","pid":20310,"started_at":"2026-10-21T06:35:06.829324Z","finished_at":"2026-10-21T06:37:05.166064Z","session_id":"01a15d2b-b1d2-70c2-8cf3-18b421576834","outcome":"returned"}] -->
