---
gate: boundary-review
issue: 302
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T12:40:11-07:00"
      dispatched: "2026-10-12T12:29:25-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: atlas says every context-bound git runner kills its process group; gitx HistoryStream does not
          detail: atlas/workflow/sdlc-binary.md:1568 claims every context-bound git runner uses processgroup.Terminate, but cmd/sdlc/internal/gitx/trunkfile.go:483 (HistoryStream, read-only git log) uses CommandContext's default Cancel plus a direct Process.Kill. It is harmless since git log has no far end. Narrow the sentence to runners that push or fetch (or spawn hooks or transports). Only instance in the window; weave's acquire/git.go:50 and environment.go:20 are outside the sdlc doc's scope.
          family: doc-claim-overreach
          round: 1
        - id: BR-2
          severity: Minor
          title: 'ARCH-DRY: gitx and cmd/sdlc cancel-ends-receive-side tests are near-identical'
          detail: cmd/sdlc/internal/gitx/cancel_test.go:14-36 and cmd/sdlc/trackerenv_cancel_test.go:15-35 differ only in the runner call. Extract a testfix helper (e.g. AssertCancelEndsReceiveSide(t, push func(ctx, root, origin string) error)) so each new context-bound runner adds one line, as the atlas rule asks.
          family: duplicated-test-scaffold
          round: 1
      recipe: small-diff-review
      isolation: checkout
      reviewed: 6caa63c639fc0859603049140327749f441ee879
      blocked: false
    - "n": 2
      timestamp: "2026-10-12T12:49:15-07:00"
      dispatched: "2026-10-12T12:43:10-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'atlas/workflow/sdlc-binary.md:1568-1572 now scopes the rule to runners that push, fetch, or spawn hooks or transports, and exempts read-only git log. Checked: both such runners use processgroup.Terminate (trackerenv.go:114, trunkfile.go:87); trunkfile.go:483 is the exempt git log.'
          round: 2
        - id: BR-2
          disposition: addressed
          note: Extracted testfix.AssertCancelEndsReceiveSide (blockingorigin.go:56). gitx/cancel_test.go and trackerenv_cancel_test.go are now single closures, and both pass at head.
          round: 2
      recipe: milestone-review
      isolation: checkout
      reviewed: 47a4e4636218dcd3b3dba7a160a0773a03ea48f5
      blocked: false
---

# Gate ledger — ariadne#302 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T12:40:11-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `doc-claim-overreach` atlas says every context-bound git runner kills its process group; gitx HistoryStream does not
  atlas/workflow/sdlc-binary.md:1568 claims every context-bound git runner uses processgroup.Terminate, but cmd/sdlc/internal/gitx/trunkfile.go:483 (HistoryStream, read-only git log) uses CommandContext's default Cancel plus a direct Process.Kill. It is harmless since git log has no far end. Narrow the sentence to runners that push or fetch (or spawn hooks or transports). Only instance in the window; weave's acquire/git.go:50 and environment.go:20 are outside the sdlc doc's scope.
- **BR-2** [Minor] `duplicated-test-scaffold` ARCH-DRY: gitx and cmd/sdlc cancel-ends-receive-side tests are near-identical
  cmd/sdlc/internal/gitx/cancel_test.go:14-36 and cmd/sdlc/trackerenv_cancel_test.go:15-35 differ only in the runner call. Extract a testfix helper (e.g. AssertCancelEndsReceiveSide(t, push func(ctx, root, origin string) error)) so each new context-bound runner adds one line, as the atlas rule asks.

## Round 2 — 2026-10-12T12:49:15-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — atlas/workflow/sdlc-binary.md:1568-1572 now scopes the rule to runners that push, fetch, or spawn hooks or transports, and exempts read-only git log. Checked: both such runners use processgroup.Terminate (trackerenv.go:114, trunkfile.go:87); trunkfile.go:483 is the exempt git log.
- BR-2 — addressed — Extracted testfix.AssertCancelEndsReceiveSide (blockingorigin.go:56). gitx/cancel_test.go and trackerenv_cancel_test.go are now single closures, and both pass at head.

## Open findings

(none — every finding has been disposed)
