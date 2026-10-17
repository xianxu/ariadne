# Boundary Review — ariadne#423 (whole-issue close)

| field | value |
|-------|-------|
| issue | 423 — sdlc issue show reports the first close review's verdict, not the latest (REWORK shown after a SHIP) |
| repo | ariadne |
| issue file | workshop/issues/000423-sdlc-issue-show-reports-the-first-close-review-s-verdict-not-the-latest-rework-shown-after-a-ship.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2bfbc4fa27ed30da2455ad38d4eaea0ddf82a774..4bea66dbd7c3275d35560124f274834b1505e16b |
| command | sdlc close --issue 423 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T04:24:25-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The main REWORK→SHIP selection and archived-history paths work, but stale tracker data exposes a blocking edge case: an issue without a completion binding loses its latest verdict. A temporary regression probe reproduced this without changing repository files.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: binding-presence-independent-of-read-quality
    title: |
      Stale cards without completion bindings suppress the latest review
    detail: |
      cmd/sdlc/internal/observe/checkpoints.go:186 uses valued(comp.Read) as evidence that a completion binding exists. assembleCompletion preserves State=stale when a stale card has no binding (assemble.go:198), leaving ReviewedHead empty. A working issue with a valid latest SHIP consequently becomes unknown with an empty verdict and error "no review entry matches the completion reviewed head ". The complete family in this window is this selection guard and its shared text/JSON consumers. Check actual binding presence separately from read quality; preserve latest-review selection when absent and completion-bound selection when present. Add regression coverage for both stale states.
```

1. **Strengths**
   - Reuses `markdoc.StripFenced` to exclude fenced examples.
   - Tests completion selection, malformed latest evidence, ledger numbering, and duplicate timestamps.
   - The lifecycle test verifies REWORK→SHIP through archival and branch deletion.

2. **Critical findings**
   - The binding-presence defect above blocks shipment. Check binding existence before entering completion matching.

3. **Important findings**
   - None.

4. **Minor findings**
   - None.

5. **Test coverage notes**
   - Observer package passed.
   - Targeted lifecycle and sidecar create/append tests passed.
   - Temporary overlay test for a stale working card without completion failed: expected latest SHIP; received unknown.
   - `git diff --check` passed.

6. **Architectural notes**
   - **ARCH-DRY: pass.** Fence handling reuses the existing helper.
   - **ARCH-PURE: pass.** Selection remains deterministic; ledger collection stays at the IO boundary.
   - **ARCH-PURPOSE: flag.** The required “latest review when there is no completion” behavior fails for stale cards.
   - Atlas documents the changed readout. No new command, flag, or configuration surface requires a README usage update.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry distinguishing completion-binding presence from tracker freshness, with fresh/stale × present/absent regression coverage.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e2b7f2f1-a74d-4196-937d-412659227373","repository":"github.com/xianxu/ariadne","issue_id":"000423","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/8975df82cc57282c","reviewed_head":"4bea66dbd7c3275d35560124f274834b1505e16b","slot":"ariadne:5","requested_at":"2026-10-17T11:23:17.606778Z"},"phase":"finished","pid":88350,"started_at":"2026-10-17T11:23:17.763941Z","finished_at":"2026-10-17T11:24:25.184821Z","session_id":"01a1499a-1936-71b1-bd81-4532363ae12c","outcome":"returned"}] -->
