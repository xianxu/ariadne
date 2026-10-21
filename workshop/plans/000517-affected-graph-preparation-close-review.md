# Boundary Review — ariadne#517 (whole-issue close)

| field | value |
|-------|-------|
| issue | 517 — Prepare generated inputs before affected graph discovery |
| repo | ariadne |
| issue file | workshop/issues/000517-affected-graph-preparation.md |
| boundary | whole-issue close |
| milestone | — |
| window | 993b4be4acf30b88cf3215fb5270dbfc776d4333..d219882fb8a83a1c5195c0550fb070f3a3407cca |
| command | sdlc close --issue 517 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T23:45:51-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The preparation path correctly uses a fresh checkout, pins declaration authority, and preserves conservative fallback. Shipping is blocked by a reproduced signal-handling gap that leaves preparation running after the CLI exits, plus an acceptance-contract omission.

Done-when: stale — it does not cover the immutable-main declaration authority added in `## Revisions`.

1. **Strengths**
   - The native graph fixture verifies same-head generation, once-only execution, caller-output isolation, and checkout cleanup.
   - The pinned-main fixture rejects candidate declarations even when `origin/main` advances after pinning.
   - README and atlas document the new configuration, bounds, fallback, and compatibility.

2. **Critical findings**
   - `cmd/sdlc/localcheck_graphprepare.go:23`: OS signals do not cancel the preparation context. The generator survives CLI termination. Own signals during this phase, propagate cancellation, and test the real CLI.

3. **Important findings**
   - `workshop/issues/000517-affected-graph-preparation.md:32`: add the revised immutable-main requirement to `Done when`.

4. **Minor findings**
   - None.

5. **Test coverage**
   - Focused affected/preparation tests passed; all affected-package tests and the diff check passed.
   - A pinned-head CLI reproduction returned exit `-15` after SIGTERM while its preparation process remained alive. Reproduction resources were removed.
   - The broader local-check run was stopped after about 200 seconds, blocked in the existing `TestLocalCheckCancelOnlyOwnRun` cleanup; it supplies no passing evidence.

6. **Architecture**
   - **ARCH-DRY: pass** — reuses process-group, checkout, and fallback helpers.
   - **ARCH-PURE: pass** — validation remains pure; execution stays at the IO boundary.
   - **ARCH-PURPOSE: flag** — production signal cancellation does not deliver the promised preparation shutdown.
   - **ARCH-EVOLVE: pass** — additive optional configuration, unchanged strict validation, and no new rewrite path or version bump.

7. **Plan revisions**
   - Append a `## Revisions` entry making SIGINT/SIGTERM shutdown part of preparation’s cancellation contract.
   - Extend `Done when` to require immutable-main authority despite candidate declaration edits and subsequent movement of `origin/main`.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: owned-subprocess-cancellation
    title: |
      Preparation survives SIGTERM because its context does not own CLI signals
    detail: |
      cmd/sdlc/localcheck_graphprepare.go:23-33 creates a separate process group but receives the background CLI context: main.go:83-98 excludes local-check run from signal ownership, and runUnderSlot installs signal handling only later. A reproduction using the pinned-head binary observed CLI exit -15 after SIGTERM while the preparation process remained alive; neither its timeout nor deferred checkout cleanup survives the CLI's death. The complete family in this window is the new preparation seam under SIGINT and SIGTERM; the context-cancellation fixture at localcheck_graphprepare_process_test.go:28 bypasses both. Own signals for this resource lifetime, ensure cancellation cannot fall through to a suite using a live parent context, and add real CLI signal tests asserting descendant shutdown, cleanup, and no passing publication. This leaves ARCH-PURPOSE's promised cancellation behavior incomplete.
  - id: new
    severity: Important
    kind: defect
    family: acceptance-contract-traceability
    title: |
      Done when omits the revised immutable-main declaration requirement
    detail: |
      workshop/issues/000517-affected-graph-preparation.md:32-37 remains unchanged although Revisions at lines 88-92 adds immutable main declaration authority and localcheck_cmd.go:142-174 implements it. The complete missing acceptance family here is declaration provenance under candidate graph_prepare edits and origin/main advancing after its commit is pinned; both are exercised by TestAffectedGraphPreparationFromPinnedMain but neither is required by Done when. Add an explicit acceptance clause and a timestamped Revisions entry connecting it to the existing fixture.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f388efa4-d86e-4ae9-b9cd-311ee9d95164","repository":"github.com/xianxu/ariadne","issue_id":"000517","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/2cc6a0776fea615d","reviewed_head":"d219882fb8a83a1c5195c0550fb070f3a3407cca","slot":"ariadne:5","requested_at":"2026-10-21T06:39:44.76938Z"},"phase":"finished","pid":80871,"started_at":"2026-10-21T06:39:44.864893Z","finished_at":"2026-10-21T06:40:47.084969Z","session_id":"01a15d2f-efde-7781-8d0a-67ad6b7ed2c0","outcome":"exit-error"},{"version":1,"intent":{"disposable_root":true,"id":"e00ec68a-079a-4c8d-9db7-23c38cb7dfcd","repository":"github.com/xianxu/ariadne","issue_id":"000517","recipe":"small-diff-review","round":1,"attempt":2,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/2cc6a0776fea615d","reviewed_head":"d219882fb8a83a1c5195c0550fb070f3a3407cca","slot":"ariadne:5","requested_at":"2026-10-21T06:40:48.886659Z"},"phase":"finished","pid":92335,"started_at":"2026-10-21T06:40:49.045085Z","finished_at":"2026-10-21T06:45:50.605354Z","session_id":"01a15d30-eaef-7a73-a0a2-b096cbe46bb6","outcome":"returned"}] -->

## Review preparation and protocol diagnostics

Quoted values are diagnostic data, not review verdicts.
Protocol attempt 1
reason: "missing verdict"
output: ""

truncated: false
retained-bytes: 0000000206
