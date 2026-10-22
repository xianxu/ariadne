# Boundary Review — ariadne#506 (whole-issue close)

| field | value |
|-------|-------|
| issue | 506 — tl-worker draining rule and tl-lead drain-the-fleet procedure (fleet-drain; recovery only on the operator's word) |
| repo | ariadne |
| issue file | workshop/issues/000506-tl-worker-draining-rule-and-tl-lead-drain-the-fleet-procedure-fleet-drain-recovery-only-on-the-operator-s-word.md |
| boundary | whole-issue close |
| milestone | — |
| window | fce345a89d417394aa6d1b7c09c758b02707066b..27d3c860c3f6fdf2f8f6d35e7ff6b8f21dcc965c |
| command | sdlc close --issue 506 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T03:05:20-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The worker and steward procedures match the issue’s central safety requirements, and both targeted packages pass. One incomplete contract sweep remains: shipped queue help still instructs workers to acknowledge a drain ID, contradicting the new notice-only procedure.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: drain-contract-consistency
    title: |
      Remove the obsolete acknowledgment protocol from remaining drain guidance
    detail: |
      cmd/sdlc/helptext/queue.md:70-71 still instructs workers to "acknowledge the authenticated drain ID"; go run ./cmd/sdlc queue --help exits 0 and renders that instruction. The other remaining instance is atlas/workflow/sdlc-binary.md:2150, which describes a supervisor "notice/ack formatter". These conflict with cmd/tl-worker/help.md:122 and leave the ARCH-DRY/ARCH-PURPOSE contract sweep incomplete. Replace both restatements with pointers to the authoritative worker/Couch procedures, and extend rendered-help regression coverage to reject the removed acknowledgment requirement in queue help. Preserve the deliberately historical fixtures.
```

1. **Strengths**

   - Worker help distinguishes already-running work from queued work and orders checkpointing before idling (`cmd/tl-worker/help.md:109`).
   - Steward help separates queue control, restart-safety judgment, automation, and operator recovery authority (`cmd/tl-lead/help.md:46`).
   - Tests exercise both no-argument and `--help` entry points, including ordering safeguards. README and the TL atlas received corresponding updates.

2. **Critical findings**

   None.

3. **Important findings**

   The contract inconsistency above. Sweep both enumerated locations so agents consulting queue rollout guidance receive the same notice-only contract.

4. **Minor findings**

   None.

5. **Test coverage notes**

   `go test ./cmd/tl-worker ./cmd/tl-lead -count=1` passed. Both real TL CLI help renders exited successfully. The queue help render independently confirmed the stale instruction. Current regression coverage checks obsolete acknowledgment guidance only in the TL help surfaces, leaving this sibling unguarded. No live fleet or queue mutation was performed.

6. **Architectural notes**

   - **ARCH-DRY — flag:** remaining queue and atlas restatements have drifted from the worker-owned contract.
   - **ARCH-PURE — pass:** changes reuse existing rendering seams without adding runtime IO or business logic.
   - **ARCH-PURPOSE — flag:** the main procedures are delivered, but the obsolete protocol remains discoverable through related shipped guidance.
   - **ARCH-EVOLVE — pass:** no persisted production format or rewrite path changes.

   The quick plan has no Core concepts table; that cross-check is inapplicable.

7. **Plan revision recommendations**

   Append a `## Revisions` entry expanding the notice-only sweep to `sdlc queue --help` and the queue-drain atlas passage, with a rendered queue-help regression guard.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f8b854d8-807c-45ba-b436-bd0b1d306113","repository":"github.com/xianxu/ariadne","issue_id":"000506","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/dd0622cbdb164413","reviewed_head":"27d3c860c3f6fdf2f8f6d35e7ff6b8f21dcc965c","slot":"ariadne:5","requested_at":"2026-10-22T10:03:50.544475Z"},"phase":"finished","pid":26491,"started_at":"2026-10-22T10:03:50.655352Z","finished_at":"2026-10-22T10:05:19.893023Z","session_id":"01a16311-273d-7d70-9bfc-10feba40fdc9","outcome":"returned"}] -->
