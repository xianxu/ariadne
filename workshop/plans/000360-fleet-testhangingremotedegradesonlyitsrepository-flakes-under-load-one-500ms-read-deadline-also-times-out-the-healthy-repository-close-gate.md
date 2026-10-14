---
gate: boundary-review
issue: 360
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T21:43:22-07:00"
      dispatched: "2026-10-13T21:33:00-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: atlas/workflow/sdlc-binary.md:365 still describes recordsReadDeadline as a whole-read 15s deadline
          detail: Since this change the 15s bounds only remote contact (ls-remote/fetch/push via runRemote); local git work is not counted, and a hanging remote now degrades to stale (unknown only if never fetched). Helptext and the issues.go doc comment were updated; the atlas paragraph was not. It is the only stale restatement in the tree.
          family: docs-restate-changed-contract
          round: 1
        - id: BR-2
          severity: Important
          title: Fleet hanging-remote test still gives the healthy repository's real remote calls only 500ms
          detail: 'recordsdeadline_test.go:108 keeps shortDeadline(500ms). The healthy repository''s ls-remote (and possible fetch) against a local bare repo spawns git plus upload-pack, and under load that can exceed 500ms, turning it stale and flaking the test: the same failure mode, reduced. The Spec asked for a few seconds of headroom. Raise it to about 2s; the 30s hang script and the 5s elapsed bound still separate bounded from unbounded. The other timing tests in the window (gitx remotebudget_test.go) already use wide margins or fake remotes.'
          family: timing-test-margin
          round: 1
        - id: BR-3
          severity: Minor
          title: runRemote duplicates run's context-check, call, and context-error-precedence body (ARCH-DRY)
          detail: trunkfile.go:163-170 against trunkfile.go:209-221. Extract runIn(ctx, env, args...) and have run and runRemote both call it.
          family: shared-helper-not-extracted
          round: 1
        - id: BR-4
          severity: Minor
          title: TestRemoteBudgetBoundsAHangingRemote fails via test-timeout panic when the budget is removed
          detail: hangRemote(t, time.Hour) means a regression shows up as a go test timeout rather than a clean assertion. A delay of about 10s would make the failure crisp.
          family: test-failure-mode
          round: 1
        - id: BR-5
          severity: Minor
          title: BoundRemote doc does not say the budget is consumed for the trunk's lifetime
          detail: remoteDeadline is fixed at the first remote command and never reset, so a reused TrunkFile starts its next operation with the budget spent. Safe for fleet, which builds a fresh Repository per load; worth one clause in the trunkfile.go:144 comment.
          family: docs-restate-changed-contract
          round: 1
      recipe: small-diff-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-13T21:47:53-07:00"
      dispatched: "2026-10-13T21:45:16-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: atlas/workflow/sdlc-binary.md:365-371 now states the remote-only shared 15s budget, local work not counted, stale/unknown degradation.
          round: 2
        - id: BR-2
          disposition: addressed
          note: recordsdeadline_test.go:112 uses shortDeadline(2s); elapsed bound 10s vs 30s hang; reverting BoundRemote turns it red (60.7s).
          round: 2
        - id: BR-3
          disposition: addressed
          note: trunkfile.go:214-223 runIn extracted; run and runRemote both delegate to it.
          round: 2
        - id: BR-4
          disposition: addressed
          note: remotebudget_test.go:87 hangRemote(t, 10s); with runRemote reverted the test fails cleanly at line 97 in 10.2s, no timeout panic.
          round: 2
        - id: BR-5
          disposition: addressed
          note: trunkfile.go:148-150 documents the lifetime budget and calling BoundRemote again to reset; reset implemented by zeroing remoteDeadline.
          round: 2
      recipe: small-diff-review
      isolation: checkout
      reviewed: 45da7ab97efc46037c6ca297e935ecb111f27f63
      blocked: false
---

# Gate ledger — ariadne#360 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T21:43:22-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `docs-restate-changed-contract` atlas/workflow/sdlc-binary.md:365 still describes recordsReadDeadline as a whole-read 15s deadline
  Since this change the 15s bounds only remote contact (ls-remote/fetch/push via runRemote); local git work is not counted, and a hanging remote now degrades to stale (unknown only if never fetched). Helptext and the issues.go doc comment were updated; the atlas paragraph was not. It is the only stale restatement in the tree.
- **BR-2** [Important] `timing-test-margin` Fleet hanging-remote test still gives the healthy repository's real remote calls only 500ms
  recordsdeadline_test.go:108 keeps shortDeadline(500ms). The healthy repository's ls-remote (and possible fetch) against a local bare repo spawns git plus upload-pack, and under load that can exceed 500ms, turning it stale and flaking the test: the same failure mode, reduced. The Spec asked for a few seconds of headroom. Raise it to about 2s; the 30s hang script and the 5s elapsed bound still separate bounded from unbounded. The other timing tests in the window (gitx remotebudget_test.go) already use wide margins or fake remotes.
- **BR-3** [Minor] `shared-helper-not-extracted` runRemote duplicates run's context-check, call, and context-error-precedence body (ARCH-DRY)
  trunkfile.go:163-170 against trunkfile.go:209-221. Extract runIn(ctx, env, args...) and have run and runRemote both call it.
- **BR-4** [Minor] `test-failure-mode` TestRemoteBudgetBoundsAHangingRemote fails via test-timeout panic when the budget is removed
  hangRemote(t, time.Hour) means a regression shows up as a go test timeout rather than a clean assertion. A delay of about 10s would make the failure crisp.
- **BR-5** [Minor] `docs-restate-changed-contract` BoundRemote doc does not say the budget is consumed for the trunk's lifetime
  remoteDeadline is fixed at the first remote command and never reset, so a reused TrunkFile starts its next operation with the budget spent. Safe for fleet, which builds a fresh Repository per load; worth one clause in the trunkfile.go:144 comment.

## Round 2 — 2026-10-13T21:47:53-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — atlas/workflow/sdlc-binary.md:365-371 now states the remote-only shared 15s budget, local work not counted, stale/unknown degradation.
- BR-2 — addressed — recordsdeadline_test.go:112 uses shortDeadline(2s); elapsed bound 10s vs 30s hang; reverting BoundRemote turns it red (60.7s).
- BR-3 — addressed — trunkfile.go:214-223 runIn extracted; run and runRemote both delegate to it.
- BR-4 — addressed — remotebudget_test.go:87 hangRemote(t, 10s); with runRemote reverted the test fails cleanly at line 97 in 10.2s, no timeout panic.
- BR-5 — addressed — trunkfile.go:148-150 documents the lifetime budget and calling BoundRemote again to reset; reset implemented by zeroing remoteDeadline.

## Open findings

(none — every finding has been disposed)
