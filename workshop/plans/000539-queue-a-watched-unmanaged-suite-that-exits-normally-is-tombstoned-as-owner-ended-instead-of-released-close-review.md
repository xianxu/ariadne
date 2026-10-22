# Boundary Review — ariadne#539 (whole-issue close)

| field | value |
|-------|-------|
| issue | 539 — queue: a watched unmanaged suite that exits normally is tombstoned as 'owner ended' instead of released |
| repo | ariadne |
| issue file | workshop/issues/000539-queue-a-watched-unmanaged-suite-that-exits-normally-is-tombstoned-as-owner-ended-instead-of-released.md |
| boundary | whole-issue close |
| milestone | — |
| window | fd00ccdbe7b25459c35fc820c5320aab9e43cea8..68f2de6564778342dc6e1bdc0db630cefa29d0d5 |
| command | sdlc close --issue 539 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T23:21:22-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation meets the issue’s approved narrow contract: positively observed exit 0 and an empty private process group permit retirement; uncertain endings remain unresolved. Real Darwin process tests and compatibility checks passed. No blocking findings.

```findings
{}
```

1. **Strengths**

   - One pure decision function gates completion on independent positive evidence (`scripts/test-lock.py:110`).
   - Observer registration precedes acquisition acknowledgment, with incarnation checks and bounded readiness.
   - Completion audit durability precedes finished-phase publication; audit failures preserve observed facts and retain uncertainty.
   - Historical fixtures remain unresolved, and Go rewrites preserve additive evidence and nested unknown fields.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Independently passed:

   - Python decision/observation/audit tests: 11.
   - Real-process completion tests: 10, including survivors, cancellation, watcher loss, rapid completion and supervisor exit.
   - `go test ./cmd/sdlc/internal/queuegate -count=1`.
   - Targeted Go real-helper and priority integration tests.
   - `bash scripts/test/test-lock.test.sh`.
   - Pinned-range `git diff --check`.

   The shell suite exercised FIFO fallback because this checkout lacks `bin/sdlc`; the current-provider Go integration harness separately passed. The full suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Existing queue mutex, codec and persistence helpers are reused.
   - **ARCH-PURE — pass:** Completion policy is pure; observation and persistence stay in the IO shell.
   - **ARCH-PURPOSE — pass:** The actual unmanaged completion path is fixed; historical uncertainty is preserved.
   - **ARCH-MOCK — pass:** Stateful observation doubles share production seams, supplemented by native process conformance.
   - **ARCH-CONSTRAINTS — pass:** Readiness, snapshot capacity and survivor output are bounded; enumeration occurs at exit.
   - **ARCH-SECURE — pass:** Exact incarnation/private-group checks fence completion; tests use private storage.
   - **ARCH-ORDER — pass:** Final cancellation checks and retirement serialize under the mutex; crash cuts distinguish committed completion from uncertainty.
   - **ARCH-FUNERAL — pass:** Watcher resources have exit paths; permanent audit growth is explicitly bounded per completion and documented.
   - **ARCH-EVOLVE — pass:** Schema changes are additive, with strict Python validation and Go preservation coverage.

   The atlas documents the new flow and audit artifact. No new command, flag or configuration key requires a README update.

7. **Plan revision recommendations:** None. The remaining close/merge checkbox appropriately awaits this boundary’s outcome.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"45526231-506e-473e-b121-6569b2645344","repository":"github.com/xianxu/ariadne","issue_id":"000539","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/7f290ea38fcd646a","reviewed_head":"68f2de6564778342dc6e1bdc0db630cefa29d0d5","slot":"ariadne:5","requested_at":"2026-10-22T06:18:56.851709Z"},"phase":"finished","pid":458,"started_at":"2026-10-22T06:18:57.06997Z","finished_at":"2026-10-22T06:21:22.02208Z","session_id":"01a16243-4309-7712-b7fc-7d10a2200e9c","outcome":"returned"}] -->
