# Boundary Review — ariadne#525 (whole-issue close)

| field | value |
|-------|-------|
| issue | 525 — tl-worker: don't checkpoint/restart while your recorded check is queued or running in your terminal |
| repo | ariadne |
| issue file | workshop/issues/000525-tl-worker-don-t-checkpoint-restart-while-your-recorded-check-is-queued-or-running-in-your-terminal.md |
| boundary | whole-issue close |
| milestone | — |
| window | 401c97fd76a49de49211bbf4068416db056a9036..a9ece645dfbe6f89fff995112ee7678f6aa0eec7 |
| command | sdlc close --issue 525 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T05:54:41-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range fulfills #525’s quick-flow plan: worker help states safe checkpoint timing, covers owned queued/running checks, requires hub-first recovery, and requires post-restart verification. The continuation pointer and README/atlas updates are present. No findings.

```findings
{}
```

1. **Strengths**

   - `cmd/tl-worker/help.md:58–63` states the complete recovery rule concisely.
   - `construct/datatype/continuation.md:59` points to the policy owner without duplicating its rules.
   - `cmd/tl-worker/testdata/rules.tsv:70–79` adds guards exercised through both rendered-help entry points.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Worker and shared-rule tests passed, as did `TestLeadHelpDoesNotRestateWorker`. In a temporary copy, substituting the base help made all ten new guards fail for both `--help` and bare invocation. The temporary copy was removed; the reviewed checkout remains unchanged. The full suite was not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** one policy owner, referenced by the continuation prototype.
   - **ARCH-PURE — pass:** no new runtime logic or IO coupling.
   - **ARCH-PURPOSE — pass:** delivers the committed rule-and-guard scope; detached execution remains an expressly separate extension.
   - **ARCH-EVOLVE — pass:** no persisted format changes.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1cea856a-9384-44e6-9134-2ca2adc29b66","repository":"github.com/xianxu/ariadne","issue_id":"000525","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/cdaf339dbeffe059","reviewed_head":"a9ece645dfbe6f89fff995112ee7678f6aa0eec7","slot":"ariadne:2","requested_at":"2026-10-21T12:53:13.965334Z"},"phase":"finished","pid":50837,"started_at":"2026-10-21T12:53:14.059627Z","finished_at":"2026-10-21T12:54:40.473945Z","session_id":"01a15e85-dfd5-7460-b63a-76addd3d35a9","outcome":"returned"}] -->
