# Boundary Review — ariadne#570 (whole-issue close)

| field | value |
|-------|-------|
| issue | 570 — Suite queue deadlocks when old and new helpers mix (grant priority vs grant aging); fall back to FIFO |
| repo | ariadne |
| issue file | workshop/issues/000570-suite-queue-deadlocks-when-old-and-new-helpers-mix-grant-priority-vs-grant-aging-fall-back-to-fifo.md |
| boundary | whole-issue close |
| milestone | — |
| window | ad8f8e62ee8e1096788a6164f088205541072e0e..401723f03420ca4d9e774917d64001bc3a60e5d5 |
| command | sdlc close --issue 570 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T03:02:26-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned changes satisfy the issue’s revised guarded-rollout contract: unsafe new grants fail closed, status exposes existing mixed-policy risks, and revoke remains independent of inspection. Admission ordering is unchanged. No blocking defects found. The recorded full-suite result remains required before landing.

```findings
{}
```

1. **Strengths**

   - Grant preflight runs after authorization, uses a three-second deadline, and checks cancellation before mutation (`cmd/sdlc/landingcmd.go:124`).
   - Grant refusal and status share one pure compatibility assessment (`cmd/sdlc/queue_suite_policy.go:24`).
   - The native regression reproduces historical/current admission disagreement and verifies progress after private revoke (`scripts/test/test-lock-priority.test.py:36`).
   - README, atlas, and operator guidance describe the retained later-join risk without claiming universal deadlock prevention.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed:

   - Targeted suite-queue and landing grant/revoke tests, including the native Python helper harness: **56.013s**.
   - Targeted compatibility, decoding, risk, and grant lifecycle tests under the race detector: **2.235s**.
   - TL guidance test: **0.164s**.
   - Range whitespace check.

   Independently compared every captured historical function/class against its cited Git revision; all matched. Tests cover malformed policy metadata, duplicate keys, unavailable inspection, cancellation, metadata preservation, and frozen status JSON. No repository edits were made. The full suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — Pass:** One assessment serves both new consumers.
   - **ARCH-PURE — Pass:** Compatibility logic is pure; inspection stays behind an injected boundary.
   - **ARCH-PURPOSE — Pass:** Implements the explicitly revised acceptance contract and documents its limitations.
   - **ARCH-MOCK — Pass:** Private stateful grant logs and native locks exercise the production boundary without live queue mutation.
   - **ARCH-CONSTRAINTS — Pass:** Bounded inspection and counted linear assessment; no additional admission scan.
   - **ARCH-SECURE — Pass:** Unknown metadata refuses grants; duplicate-key parsing prevents ambiguous policy authority.
   - **ARCH-ORDER — Pass:** Observation precedes mutation; cancellation and revoke recovery are tested. Later joins remain an explicit operational risk.
   - **ARCH-FUNERAL — Pass:** No new durable artifact family; existing ticket ownership and cleanup govern the added field.
   - **ARCH-EVOLVE — Pass:** Additive ticket metadata survives holder publication with nested extensions; frozen drain JSON remains unchanged.

7. **Plan revision recommendations:** None. The core-concept classifications and delivered implementation agree with the revised plan.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cb31a4da-3241-44ae-a725-067260c9648b","repository":"github.com/xianxu/ariadne","issue_id":"000570","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/8420b570c5c245a2","reviewed_head":"401723f03420ca4d9e774917d64001bc3a60e5d5","slot":"ariadne:3","requested_at":"2026-10-23T10:00:28.483888Z"},"phase":"finished","pid":27165,"started_at":"2026-10-23T10:00:28.659543Z","finished_at":"2026-10-23T10:02:26.192486Z","session_id":"01a16834-6e7e-7840-b4f8-8a6b4a35cd8b","outcome":"returned"}] -->
