# Boundary Review — ariadne#547 (whole-issue close)

| field | value |
|-------|-------|
| issue | 547 — Train car active-time counts conductor-owned queue wait and status polls as focused work |
| repo | ariadne |
| issue file | workshop/issues/000547-train-car-wait-attribution.md |
| boundary | whole-issue close |
| milestone | — |
| window | b99ed24720b23c3d7370f183a5fb0a360706c137..9dcb195ed5fd08d9c38cddd3892657fb922cd3bb |
| command | sdlc close --issue 547 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T05:53:58-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: medium
```

The implementation matches the governing visibility revision: retained train waits produce descriptive coverage without changing activity subtraction, measured hours, or calibration policy. No blocking defects found. README, help, and atlas updates describe the implemented behavior. Verification was targeted; the full suite was not completed during this review.

```findings
{}
```

1. **Strengths**

   - Wait endpoints follow the held-resource lifecycle, with readiness joined before recording the owner’s return (`landingqueue.go:209`).
   - Selection requires immutable-head membership and matching owner/source windows (`actualtrainwait.go:31`).
   - Optional snapshot attachment validates a copy and preserves accounting evidence on overflow (`actualtrainwait.go:139`).
   - Nested snapshot fields use the existing preserving codec, with JSON/YAML round-trip and strict-validation tests (`measurement_codec.go:502`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed both affected internal packages, focused coverage/selection/replay/producer tests, and focused race checks. Tests cover malformed evidence, missing endpoints, duplicate facts, write failures, capacity limits, and unchanged Claude/Codex accounting.

   A broader queue/train run exhausted my three-minute timeout, producing deadline-related archive-resume failures. The affected archive-resume and split-refresh cases passed when rerun separately in 53.2 seconds. Full-suite verification remains separate.

6. **Architectural notes**

   - **ARCH-DRY — pass:** reuses audit retention, immutable-head proof, and snapshot codecs.
   - **ARCH-PURE — pass:** reduction and selection are testable separately from IO.
   - **ARCH-PURPOSE — pass:** delivers the revised visibility contract; makes no idle-subtraction claim.
   - **ARCH-MOCK — pass:** uses existing injected seams and portable fixtures; adds no external dependency.
   - **ARCH-CONSTRAINTS — pass:** explicit audit and snapshot budgets degrade to diagnostics.
   - **ARCH-SECURE — pass:** validates identities, known fields, digests, and pinned source membership.
   - **ARCH-ORDER — pass:** held-resource guards and endpoint ordering preserve incomplete outcomes conservatively.
   - **ARCH-FUNERAL — pass:** reuses 30-day audit retention; polling creates no additional facts.
   - **ARCH-EVOLVE — pass:** snapshot additions preserve unknown fields and retain existing method meanings.

7. **Plan revision recommendations:** None beyond the existing governing revisions.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1a30f2ea-5cf6-445a-9844-d989e088a33b","repository":"github.com/xianxu/ariadne","issue_id":"000547","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/9b54a5a111d9781d","reviewed_head":"9dcb195ed5fd08d9c38cddd3892657fb922cd3bb","slot":"ariadne:4","requested_at":"2026-10-22T12:46:55.305807Z"},"phase":"finished","pid":92136,"started_at":"2026-10-22T12:46:55.403419Z","finished_at":"2026-10-22T12:53:53.521743Z","session_id":"01a163a6-74bc-74b1-9c25-fd6555f932da","outcome":"returned"}] -->
